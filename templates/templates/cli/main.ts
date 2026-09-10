require("includes/includes.ts");
require("features.ts");
require("includes/tests.ts");
require("mockserver.ts");

// @ts-ignore
function runJobs(jobs: Job[]) {
  for (const job of jobs) {
    runJob(job);
  }
}

// @ts-ignore
function runJob(job: Job) {
  if (!handleCommonJobs(job)) {
    switch (job.ID) {
      case "goreleaser-config":
        writeGoreleaserConfig();
        break;
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}

// @ts-ignore
function getJobs(): Job[] {
  const jobs: Job[] = [];

  // @ts-ignore
  const goDefaultConfig = getTargetDefaultTemplateConfig("go");

  goDefaultConfig.PackageName = getGolangPackage() + "/internal/sdk";
  goDefaultConfig.SDKVersion = context.Global.Config.SDKVersion;
  goDefaultConfig.WrapperName = "speakeasy-sdk/go";

  goDefaultConfig.IdiomaticMethodCollisionNames =
    context.Global.Config.IdiomaticMethodCollisionNames;

  // Save test-related AST fields before intertemplate call, because
  // ExecuteInterTemplateTarget clears them when tests: false is passed.
  const savedTestGroup = context.Global.AST.MainSDK.TestGroup;
  const savedOutputTests = context.Global.AST.MainSDK.OutputTests;

  interoptTemplateTarget(
    "go",
    "internal/sdk",
    context.Global.AST,
    goDefaultConfig,
    {
      documentation: false,
      module: false,
      tests: false,
      internalModules: false,
    },
  );

  // Restore test-related AST fields for CLI's own test generation.
  context.Global.AST.MainSDK.TestGroup = savedTestGroup;
  context.Global.AST.MainSDK.OutputTests = savedOutputTests;

  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  jobs.push(...getCommandJobs(context.Global.AST.MainSDK));

  // Template tests directory
  if (isTemplateFeatureEnabled("tests")) {
    if (
      context.Global.AST.MainSDK.OutputTests ||
      sdkHasTests(context.Global.AST)
    ) {
      const testDirectory = getTestDirectory();

      // Template common test files individually (templateDirectory creates
      // separate output directories like "testscommon/" instead of merging
      // into "tests/", which breaks Go package resolution).
      const commonFiles = getDirectoryFiles(`tests/common`);
      for (const file of commonFiles) {
        if (file.endsWith(".skip")) continue;
        const baseName = file.substring(file.lastIndexOf("/") + 1);
        const outFile = `${testDirectory}/${baseName}`;
        if (file.endsWith(".stmpl")) {
          jobs.push(
            createTemplateFileJob(file, outFile.replace(".stmpl", ""), {}),
          );
        } else {
          copy(file, outFile);
        }
      }

      // Template the CLI test harness (needs dynamic package import)
      jobs.push(
        createTemplateFileJob(
          "cli_harness_test.go.stmpl",
          `${testDirectory}/cli_harness_test.go`,
          {},
        ),
      );
      if (isAnyInteractiveEnabled()) {
        jobs.push(
          createTemplateFileJob(
            "interactive_policy_test.go.stmpl",
            "internal/interactive/policy_test.go",
            {},
          ),
        );
      }
      jobs.push(
        createTemplateFileJob(
          "usage_test.go.stmpl",
          `${testDirectory}/usage_test.go`,
          {},
        ),
      );
      jobs.push(
        createTemplateFileJob(
          "dryrun_response_test.go.stmpl",
          `${testDirectory}/dryrun_response_test.go`,
          {},
        ),
      );
      jobs.push(
        createTemplateFileJob(
          "dryrun_matrix_test.go.stmpl",
          `${testDirectory}/dryrun_matrix_test.go`,
          {},
        ),
      );
      // README examples are a contract: every `<cli> ...` line in a fenced
      // bash block runs under --dry-run (skips itself when no README exists).
      if (isTemplateFeatureEnabled("documentation")) {
        jobs.push(
          createTemplateFileJob(
            "readme_examples_test.go.stmpl",
            `${testDirectory}/readme_examples_test.go`,
            {},
          ),
        );
      }
      // Artifact-output contract tests render only when a declared command
      // carries output.artifact (the template body is gated; an empty render
      // emits nothing).
      jobs.push(
        createTemplateFileJob(
          "artifact_test.go.stmpl",
          `${testDirectory}/artifact_test.go`,
          {},
        ),
      );
      if (templateAsyncTestEnabled()) {
        jobs.push(
          createTemplateFileJob(
            "async_test.go.stmpl",
            `${testDirectory}/async_test.go`,
            {},
          ),
        );
      }
      // Streamed-projection proof for declared intents (only when the
      // manifest declares output.stream.select over an SSE operation).
      if (templateIntentStreamTestEnabled()) {
        jobs.push(
          createTemplateFileJob(
            "intentstream_test.go.stmpl",
            `${testDirectory}/intentstream_test.go`,
            {},
          ),
        );
      }
      if (templateOperationStreamTestEnabled()) {
        jobs.push(
          createTemplateFileJob(
            "opstream_test.go.stmpl",
            `${testDirectory}/opstream_test.go`,
            {},
          ),
        );
      }
      // Spaced-boolean hint on variadic intent commands (only when the
      // manifest declares intents; the primary group also ships its own
      // intents_test.go, so this file has a distinct name).
      if (hasCliIntents()) {
        jobs.push(
          createTemplateFileJob(
            "intenthint_test.go.stmpl",
            `${testDirectory}/intenthint_test.go`,
            {},
          ),
        );
        // Unit test for declared-command ordering against cobra's late-added
        // builtin help/completion commands (package cli, next to intents.go).
        jobs.push(
          createTemplateFileJob(
            "intents_order_test.go.stmpl",
            "internal/cli/intents_order_test.go",
            {},
          ),
        );
      }

      // Template group-specific test files individually into tests/ directory.
      // When TestGroup is empty (e.g., review variant), groupTestPath resolves
      // to "tests/" which would recurse into all subdirectories. Skip in that case.
      const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
      if (groupTestPath !== "tests/" && directoryExists(groupTestPath)) {
        const groupFiles = getDirectoryFiles(groupTestPath);
        for (const file of groupFiles) {
          if (file.endsWith(".skip")) continue;
          const baseName = file.substring(file.lastIndexOf("/") + 1);
          const outFile = `${testDirectory}/${baseName}`;
          if (file.endsWith(".stmpl")) {
            jobs.push(
              createTemplateFileJob(file, outFile.replace(".stmpl", ""), {}),
            );
          } else {
            copy(file, outFile);
          }
        }
      }
    }

    jobs.push(...getTestJobs(context.Global.AST, "go"));
    jobs.push(...getTestHelperJobs(context.Global.AST, "go"));
  }

  if (isTemplateFeatureEnabled("documentation")) {
    // Only include readme and usage jobs; skip model docs (CLI doesn't need per-type documentation)
    jobs.push(
      ...getReadmeJobs().filter(
        (j: Job) => j.ID === "readme" || j.ID === "usage",
      ),
    );
  }

  jobs.push(...getGitFilesJobs("go"));

  if (shouldGenerateReleaseFiles()) {
    jobs.push(...getReleaseJobs());
  }

  return jobs;
}

// @ts-ignore
function getCommandJobs(sdk: SDK): Job[] {
  const jobs: Job[] = [];

  jobs.push(
    createTemplateFileJob(
      "root.go.stmpl",
      "internal/cli/root.go",
      context.Global.AST.MainSDK,
    ),
  );

  // Custom command registration hook: scaffolded once, never overwritten —
  // hand-written commands live in internal/cli/custom and survive
  // regeneration. A patch file for the scaffold overrides that contract:
  // the patch pins the file to generator+patch ownership, so the scaffold
  // is emitted deterministically for the patch to apply to, and hermetic
  // (fresh-tree) and incremental regeneration produce identical output.
  if (isCustomCommandsEnabled()) {
    const registerFile = "internal/cli/custom/register.go";
    if (!readFile(registerFile) || patchFileExists(registerFile)) {
      jobs.push(
        createTemplateFileJob("custom_register.go.stmpl", registerFile, {}),
      );
    }
  }

  // Catalog commands declared in the OpenAPI document via x-speakeasy-cli-catalog
  const catalogs = collectCliCatalogs();
  if (catalogs.length > 0) {
    jobs.push(
      createTemplateFileJob("catalog.go.stmpl", "internal/cli/catalog.go", {
        Catalogs: catalogs,
      }),
    );
  }

  // Declarative intent commands (x-speakeasy-cli-commands)
  const intentManifest = collectIntentManifest();
  if (
    intentManifest.Bound.length > 0 ||
    intentManifest.Planned.length > 0 ||
    intentManifest.GroupTags.length > 0
  ) {
    for (const bound of intentManifest.Bound) {
      const dir = bound.PkgPath
        ? `internal/cli/${bound.PkgPath}`
        : "internal/cli";
      jobs.push(
        createTemplateFileJob(
          "intentcmd.go.stmpl",
          `${dir}/intent_${bound.ID}.go`,
          bound,
        ),
      );
    }
    jobs.push(
      createTemplateFileJob(
        "intents.go.stmpl",
        "internal/cli/intents.go",
        intentManifest,
      ),
    );
  }

  // Configure command is always generated (preferences are always configurable)
  jobs.push(
    createTemplateFileJob(
      "configure.go.stmpl",
      "internal/cli/configure.go",
      {},
    ),
  );

  // Whoami and masking require security or global parameters
  if (hasConfigurableSettings()) {
    jobs.push(
      createTemplateFileJob("whoami.go.stmpl", "internal/cli/whoami.go", {}),
    );
    jobs.push(
      createTemplateFileJob("masking.go.stmpl", "internal/cli/masking.go", {}),
    );
  }

  // Add auth commands if interactive auth is enabled and security fields exist
  if (isInteractiveAuthEnabled() && hasGlobalSecurity()) {
    jobs.push(
      createTemplateFileJob("auth.go.stmpl", "internal/cli/auth.go", {}),
    );
  }

  // Add version command (always present)
  jobs.push(
    createTemplateFileJob("version.go.stmpl", "internal/cli/version.go", {}),
  );

  // Usage schema emitter (always present)
  jobs.push(
    createTemplateFileJob("usage.go.stmpl", "internal/usage/schema.go", {}),
  );

  // Exact request-body JSON Schemas (--schema surface)
  const bodySchemaEntries = collectBodySchemaEntries();
  if (bodySchemaEntries.length > 0) {
    jobs.push(
      createTemplateFileJob(
        "bodyschemas.go.stmpl",
        "internal/usage/bodyschemas.go",
        { Entries: bodySchemaEntries },
      ),
    );
  }

  // Compute aliases for root-level sibling commands (sub-groups + root operations)
  const rootSiblingNames: string[] = [];
  for (const subSDK of sdk.SubSDKs) {
    rootSiblingNames.push(sanitizeCLICommand(getSDKGroupName(subSDK)));
  }
  for (const op of sdk.Operations) {
    const stutterKind = getStutterKind("", op.GetID());
    if (stutterKind !== "exact") {
      rootSiblingNames.push(sanitizeCLICommand(op.GetID()));
    }
  }
  const rootAliases = computeCommandAliases(rootSiblingNames);

  jobs.push(
    ...getOperationCommandJobs(sdk, "cli", "internal/cli/", "", rootAliases),
  );

  for (const subSDK of sdk.SubSDKs) {
    jobs.push(
      ...getSubCommandJobs(
        subSDK,
        `internal/cli/${sanitizeCLIPkgName(subSDK.Type.Name)}`,
        "",
        "", // no parent group for top-level sub-groups
        rootAliases,
      ),
    );
  }

  return jobs;
}

function getSubCommandJobs(
  subSDK: SDK,
  loc: string,
  parentAccessor: string,
  parentGroupName: string,
  parentAliases: Map<string, string[]>,
): Job[] {
  const jobs: Job[] = [];

  if (subSDK.Type.Name === "") return jobs;

  const sdkGroupName = getSDKGroupName(subSDK);

  // Build the current subSDK's accessor by appending its field name to the parent's accessor
  // Must use sanitizeSDKFieldName (matches Go SDK struct field names, e.g. "First" not "SDKFirst")
  const currentAccessor = `${parentAccessor}.${sanitizeSDKFieldName(
    subSDK.FieldName,
  )}`;

  // Look up aliases for this group command from the parent level
  const groupCmdName = getDeStutteredCommandName(parentGroupName, sdkGroupName);
  const groupAliases = parentAliases.get(groupCmdName) || [];

  // Compute aliases for children of this group (operations + nested sub-groups)
  const childNames: string[] = [];
  for (const childSDK of subSDK.SubSDKs) {
    childNames.push(
      getDeStutteredCommandName(sdkGroupName, getSDKGroupName(childSDK)),
    );
  }
  for (const op of subSDK.Operations) {
    const stutterKind = hasNameOverride(op)
      ? "none"
      : getStutterKind(sdkGroupName, op.GetID());
    if (stutterKind !== "exact") {
      const opName =
        stutterKind === "prefix" || stutterKind === "suffix"
          ? getDeStutteredCommandName(sdkGroupName, op.GetID())
          : sanitizeCLICommand(op.GetID());
      childNames.push(opName);
    }
  }
  const childAliases = computeCommandAliases(childNames);

  jobs.push(
    createTemplateFileJob(`subroot.go.stmpl`, `${loc}/root.go`, {
      Root: subSDK,
      Loc: loc,
      ParentGroupName: parentGroupName,
      SubSDKAccessor: parentAccessor,
      Aliases: groupAliases,
    }),
  );

  // Operations on this subSDK use the current accessor (includes this subSDK's name)
  jobs.push(
    ...getOperationCommandJobs(
      subSDK,
      `${sanitizeCLIPkgName(subSDK.Type.Name)}`,
      loc,
      currentAccessor,
      childAliases,
    ),
  );

  // Child subSDKs receive the current accessor as their parent
  for (const sdk of subSDK.SubSDKs) {
    jobs.push(
      ...getSubCommandJobs(
        sdk,
        `${loc}/${sanitizeCLIPkgName(sdk.Type.Name)}`,
        currentAccessor,
        sdkGroupName, // pass current group as parent for children
        childAliases,
      ),
    );
  }

  return jobs;
}

function getOperationCommandJobs(
  sdk: SDK,
  pkgName: string,
  loc: string,
  subSDKAccessor: string,
  aliases: Map<string, string[]>,
): Job[] {
  const jobs: Job[] = [];
  const sdkGroupName = getSDKGroupName(sdk);
  const groupCommandName = sdkGroupName ? sanitizeCLICommand(sdkGroupName) : "";

  for (const operation of sdk.Operations) {
    const stutterKind =
      groupCommandName && !hasNameOverride(operation)
        ? getStutterKind(sdkGroupName, operation.GetID())
        : "none";

    // Compute the command name, stripping group prefix/suffix for stutter matches
    const opCommandName =
      stutterKind === "prefix" || stutterKind === "suffix"
        ? getDeStutteredCommandName(sdkGroupName, operation.GetID())
        : sanitizeCLICommand(operation.GetID());

    // Exact-match operations are promoted: they register RunE on the parent
    // group command instead of creating a separate subcommand.
    const promotedToParent = stutterKind === "exact";

    // Look up aliases for this operation command
    const opAliases = promotedToParent ? [] : aliases.get(opCommandName) || [];

    jobs.push(
      createTemplateFileJob(
        "opcmd.go.stmpl",
        `${loc}/${sanitizeCLIOpFileName(operation.GetID())}.go`,
        {
          Op: operation,
          Package: pkgName,
          SubSDKAccessor: subSDKAccessor,
          GroupCommandName: groupCommandName,
          OpCommandName: opCommandName,
          PromotedToParent: promotedToParent,
          Aliases: opAliases,
        },
      ),
    );
  }

  return jobs;
}
