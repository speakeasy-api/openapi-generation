require("includes/includes.ts");

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
      case "testProject":
        templateTestProject(
          (u: UsageContext) =>
            `src/${sanitizeFileName(
              sanitizeClassName(u.Operation.OwningSDK.Type.Name),
            )}/${sanitizeFileName(sanitizeClassName(u.Operation.ID))}.ts`,
        );
        break;
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}

/**
 * Marks discriminated unions that contain error types as closed (not open).
 * TypeScript templating cannot properly handle open discriminated unions
 * where members are error classes, so we close them to use standard union handling.
 */
function markDiscriminatedUnionOfErrorsClosed(): void {
  for (const t of getAllTypes()) {
    if (!t.IsUnionOpen) continue;
    for (const member of t.AssociatedTypes) {
      if (member.Type.toString() === "error") {
        t.IsUnionOpen = false;
        break;
      }
    }
  }
}

// @ts-ignore
function getJobs(): Job[] {
  markDiscriminatedUnionOfErrorsClosed();

  const jobs: Job[] = [];

  jobs.push(...getSDKFunctionsJobs(context.Global.AST.MainSDK));

  if (
    context.Global.Config.EnableReactQuery &&
    !accountHasFeatureAccess("reactQueryHooks")
  ) {
    console.log(
      "React Query hooks are enabled on business and enterprise tiers. Please reach out to support@speakeasy.com to upgrade or if there was an issue.",
    );
  }

  if (isReactQueryEnabled()) {
    jobs.push(...getReactQueryJobs(context.Global.AST.MainSDK));
  }

  if (isMCPServerEnabled()) {
    jobs.push(...getMCPServerJobs(context.Global.AST.MainSDK));
  }

  jobs.push(...getSDKJobs(context.Global.AST.MainSDK));

  jobs.push(...getExamplesJobs(context.Global.AST.MainSDK));

  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Template tests directory
  if (context.Global.AST.MainSDK.OutputTests) {
    const testDirectory = getTestDirectory();
    templateDirectory("tests/common", testDirectory);

    const groupTestPath = `tests/${context.Global.AST.MainSDK.TestGroup}`;
    if (directoryExists(groupTestPath)) {
      templateDirectory(groupTestPath, testDirectory);
    }

    templateDirectory("tests/lib", `${testDirectory}/lib`);
    templateFile("tests/vitest.config.mts", "vitest.config.mts", {});
  }

  jobs.push(...getTestJobs(context.Global.AST, "ts"));
  jobs.push(...getTestHelperJobs(context.Global.AST, "ts"));

  // Template usage snippets inside test project
  if (context.Global.Config.TestProject?.OutputDir) {
    jobs.push({
      ID: "testProject",
      FileName: "testProject",
      Context: null,
    });
  }

  jobs.push(...getModelJobs(context.Global.AST.BucketedTypes, "src"));
  jobs.push(...getTypeScriptPublicExportJobs("src"));
  jobs.push(...getReadmeJobs(true)); // Single threaded model docs for typescript due to a non-deterministic bug with seeding faker and the examples being generated with different values based on the number of CPUs available
  jobs.push(...getGitFilesJobs("ts"));

  // Generated Dev Containers Templating
  if (context.Global.Config.hasOwnProperty("DevContainerSchemaPath")) {
    jobs.push(
      ...getDevContainerJobs({
        FileSuffix: "ts",
        Language: "typescript",
        FileExecution: "ts-node",
        SchemaPath: context.Global.Config.DevContainerSchemaPath,
      }),
    );
  }

  jobs.push(...getHooksJobs());

  const contributingFileJob = getTemplateContributingFileJob();
  if (contributingFileJob) {
    jobs.push(contributingFileJob);
  }

  // Has a dependency on context.Global so best to just template it while they are available
  templateFile("package.json.stmpl", "package.json", context.Global);

  return jobs;
}

function getSDKJobs(sdk: SDK): Job[] {
  const jobs: Job[] = [];

  jobs.push(createTemplateFileJob("sdk.ts.stmpl", "src/sdk/sdk.ts", sdk));

  for (const subSDK of sdk.SubSDKs) {
    jobs.push(...getSubSDKJobs(subSDK));
  }

  return jobs;
}

function getSubSDKJobs(subSDK: SDK): Job[] {
  const jobs: Job[] = [];

  if (subSDK.Type.Name === "") return jobs;

  jobs.push(
    createTemplateFileJob(
      "subsdk.ts.stmpl",
      `src/sdk/${sanitizeTypeFilename(subSDK.Type.Name)}.ts`,
      subSDK,
    ),
  );

  for (const sdk of subSDK.SubSDKs) {
    jobs.push(...getSubSDKJobs(sdk));
  }

  return jobs;
}
