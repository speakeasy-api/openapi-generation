// @ts-ignore
type TypescriptModel = {
  DocGroup: string;
  Name: string;
  Types: TypeDef[];
  Servers?: any;
  OutputLocation: string;
};

type TypeScriptPublicExportImport = {
  Path: string;
  Specifiers: { Name: string; Alias: string }[];
};

type TypeScriptPublicExportMember = {
  Name: string;
  Intermediate: string;
  CanInterface?: boolean;
  MustInterface?: boolean;
};

type TypeScriptPublicExportNamespace = {
  Declaration: string;
  Name: string;
  Members: TypeScriptPublicExportMember[];
  Namespaces: TypeScriptPublicExportNamespace[];
};

type TypeScriptPublicExportFile = {
  Imports: TypeScriptPublicExportImport[];
  Intermediates: { Name: string; Target: string }[];
  FlatAliases: TypeScriptPublicExportMember[];
  Namespaces: TypeScriptPublicExportNamespace[];
};

// @ts-ignore
function templateLanguageName(lang: string): string {
  return "typescript";
}
unregisterTemplateFunc("templateLanguageName");
registerTemplateFunc("templateLanguageName", templateLanguageName);

function getSDKFunctionsJobs(root: SDK): Job[] {
  const jobs: Job[] = [];

  const queue: Array<SDK> = [root];

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    for (const op of sdk.Operations) {
      const funcFilename = sanitizeFuncFilename(op);
      const funcPath = `src/funcs/${funcFilename}.ts`;

      jobs.push(
        createTemplateFileJob("func.ts.stmpl", funcPath, {
          DocGroup: funcPath,
          Operation: op,
        }),
      );
    }
  }

  return jobs;
}

function examplePackageName(): string {
  return `${context.Global.Config.PackageName}-examples`;
}
registerTemplateFunc("examplePackageName", examplePackageName);

function getEnvVarPlaceholder(fieldName: string): string {
  const placeholderMap: Record<string, string> = {
    ApiKey: "your_api_key_here",
    BearerToken: "your_bearer_token_here",
    Username: "your_username_here",
    Password: "your_password_here",
    Token: "your_token_here",
    ClientID: "your_client_id_here",
    ClientSecret: "your_client_secret_here",
    TokenURL: "https://example.com/oauth/token",
    BearerAuth: "your_bearer_auth_here",
    AppID: "your_app_id_here",
    Secret: "your_secret_here",
    OAuth2: "your_oauth2_token_here",
    ClientCredentials: "your_client_credentials_here",
  };

  // Check if we have a specific placeholder
  if (placeholderMap[fieldName]) {
    return placeholderMap[fieldName];
  }

  // Default placeholder based on field name
  return `your_${fieldName.toLowerCase()}_here`;
}
registerTemplateFunc("getEnvVarPlaceholder", getEnvVarPlaceholder);

function getMainUsageExample(
  root: SDK = context.Global.AST.MainSDK,
): UsageContext {
  const exampleOps = selectExampleOperations(root, 0, [], false);
  if (exampleOps.length === 0) {
    return null;
  }
  return exampleOps[0];
}
registerTemplateFunc("getMainUsageExample", getMainUsageExample);

function getMainUsageExampleOperation(
  root: SDK = context.Global.AST.MainSDK,
): Operation | undefined {
  return getMainUsageExample(root)?.Operation;
}
registerTemplateFunc(
  "getMainUsageExampleOperation",
  getMainUsageExampleOperation,
);

function getExamplesJobs(root: SDK = context.Global.AST.MainSDK): Job[] {
  const jobs: Job[] = [];

  const existingPackageJsonString = readFile("examples/package.json");
  if (existingPackageJsonString) {
    // Detect the case where the user has created an examples directory
    // so it doesn't look to be like the one we generate
    const existingPackageJson = JSON.parse(existingPackageJsonString);
    if (!existingPackageJson.scripts["build:examples"]) {
      context.Global.Config.generateExamples = false;
    }
  }

  // Check if examples generation is enabled
  if (context.Global.Config.generateExamples === false) {
    return jobs;
  }

  // Get the main usage example operation
  const example = getMainUsageExample(root);
  if (!example) {
    return jobs;
  }

  const exampleContext = createUsageContext(
    example.Operation.OwningSDK || example.SDK || root,
    example.Operation,
    example.Config,
  );
  const exampleFileName = sanitizeFuncName(example.Operation);

  // Create the package.json - need to pass a context with Global
  jobs.push(
    createTemplateFileJob(
      "examples/package.json.stmpl",
      "examples/package.json",
      {
        Global: context.Global,
      },
    ),
  );

  // Create the README.md
  jobs.push(
    createTemplateFileJob("examples/README.md.stmpl", "examples/README.md", {
      Global: context.Global,
    }),
  );

  // Create .env.template
  jobs.push(
    createTemplateFileJob(
      "examples/.env.template.stmpl",
      "examples/.env.template",
      {
        Global: context.Global,
      },
    ),
  );

  // Create the example file
  jobs.push(
    createTemplateFileJob(
      "examples/example.ts.stmpl",
      `examples/${exampleFileName}.example.ts`,
      exampleContext,
    ),
  );

  return jobs;
}

function getReactQueryJobs(root: SDK): Job[] {
  const jobs: Job[] = [];

  const seenHookNames: Record<string, string> = {};
  const seenFilePaths: Record<string, string> = {};
  const seenOpIds = new Set<string>();
  const queue: Array<SDK> = [root];

  jobs.push(
    createTemplateFileJob(
      "react-query/index.ts.stmpl",
      "src/react-query/index.ts",
      {},
    ),
    createTemplateFileJob(
      "react-query/REACT_QUERY.md.stmpl",
      "REACT_QUERY.md",
      {},
    ),
    createTemplateFileJob(
      "react-query/_context.tsx.stmpl",
      "src/react-query/_context.tsx",
      {},
    ),
    createTemplateFileJob(
      "react-query/_types.ts.stmpl",
      "src/react-query/_types.ts",
      {},
    ),
  );

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    const ops = sdk.Operations.filter((op) => !op.Webhook);
    for (const op of ops) {
      if (!hasReactQuery(op)) {
        continue;
      }

      if (seenOpIds.has(op.ID)) {
        continue;
      }
      seenOpIds.add(op.ID);

      const queryFilename = sanitizeReactQueryHookFilename(op);

      const hookName = sanitizeReactQueryName(
        op,
        hasQueryHook(op) ? "query" : "mutation",
      );

      if (seenHookNames[hookName]) {
        throw new Error(
          `Duplicate React hook name "${hookName}" for operations: ["${seenHookNames[hookName]}","${op.ID}"]`,
        );
      }
      seenHookNames[hookName] = op.ID;

      const hookPath = `src/react-query/${queryFilename}.ts`;
      if (seenFilePaths[hookPath]) {
        throw new Error(
          `Duplicate React hook file names detected "${hookPath}" for operations: [${seenFilePaths[hookPath]}, ${op.ID}]`,
        );
      }
      seenFilePaths[hookPath] = op.ID;

      switch (true) {
        case hasQueryHook(op): {
          // Generate server-safe core file (no _context.tsx import)
          const corePath = `src/react-query/${queryFilename}.core.ts`;
          jobs.push(
            createTemplateFileJob("react-query/query-core.ts.stmpl", corePath, {
              DocGroup: corePath,
              Operation: op,
            }),
          );

          // Generate hooks file (imports from _context.tsx and re-exports from core)
          jobs.push(
            createTemplateFileJob("react-query/query.ts.stmpl", hookPath, {
              DocGroup: hookPath,
              Operation: op,
            }),
          );
          break;
        }
        case hasMutationHook(op): {
          jobs.push(
            createTemplateFileJob("react-query/mutation.ts.stmpl", hookPath, {
              DocGroup: hookPath,
              Operation: op,
            }),
          );
          break;
        }
      }
    }
  }

  return jobs;
}

function templateSDKFunctionsList(root: SDK) {
  const queue: Array<SDK> = [root];
  const list: string[] = [];

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    for (const op of sdk.Operations) {
      const deprecated = !!op.Comments?.Deprecated;
      let entry = templateFuncMarkdownLink(op);
      if (deprecated) {
        entry = `~~${entry}~~`;
      }
      entry += templateFuncSummary(op);
      list.push(`- ${entry}`);
    }
  }

  return list
    .sort((a: string, b: string) => {
      const aDeprecated = a.includes("~~");
      const bDeprecated = b.includes("~~");
      if (aDeprecated && !bDeprecated) {
        return 1;
      } else if (!aDeprecated && bDeprecated) {
        return -1;
      } else {
        return a.localeCompare(b);
      }
    })
    .join("\n");
}
registerTemplateFunc("templateSDKFunctionsList", templateSDKFunctionsList);

function templateReactQueryHooksList(root: SDK) {
  const queue: Array<SDK> = [root];
  const list: string[] = [];

  while (queue.length > 0) {
    const sdk = queue.shift();
    queue.push(...sdk.SubSDKs);

    for (const op of sdk.Operations) {
      const ext = op.Extensions.ReactHook;
      if (ext?.Disabled) {
        continue;
      }

      const deprecated = !!op.Comments?.Deprecated;
      let entry = templateReactQueryHookMarkdownLink(op);
      if (deprecated) {
        entry = `~~${entry}~~`;
      }
      entry += templateReactQueryHookSummary(op);
      list.push(`- ${entry}`);
    }
  }

  return list
    .sort((a: string, b: string) => {
      const aDeprecated = a.includes("~~");
      const bDeprecated = b.includes("~~");
      if (aDeprecated && !bDeprecated) {
        return 1;
      } else if (!aDeprecated && bDeprecated) {
        return -1;
      } else {
        return a.localeCompare(b);
      }
    })
    .join("\n");
}
registerTemplateFunc(
  "templateReactQueryHooksList",
  templateReactQueryHooksList,
);

function templateFuncSummary(operation: Operation): string {
  let replacementLink = "";
  if (operation.Comments?.Deprecated) {
    const replacementID = operation.Comments.DeprecationReplacement;

    if (replacementID) {
      const replacementOp =
        context.Global.AST.MainSDK.FindOperation(replacementID);

      replacementLink = templateFuncMarkdownLink(replacementOp);
    }
  }

  return templateOperationSummary(operation, replacementLink);
}

function templateReactQueryHookSummary(operation: Operation): string {
  if (!operation.Comments?.Summary && !operation.Comments?.Deprecated) {
    return "";
  }

  let summary = " -";

  if (operation.Comments?.Summary) {
    summary += " " + operation.Comments.Summary;
  }

  if (operation.Comments?.Deprecated) {
    summary += ` :warning: **Deprecated**`;

    const replacementID = operation.Comments.DeprecationReplacement;
    const replacementOp = replacementID
      ? context.Global.AST.MainSDK.FindOperation(replacementID)
      : null;

    summary += replacementOp
      ? ` Use ${templateReactQueryHookMarkdownLink(replacementOp)} instead.`
      : "";
  }

  return summary;
}

function templateFuncMarkdownLink(operation: Operation): string {
  const filename = getSDKReadmeFileName(operation.OwningSDK);
  const link = `${filename}#${sanitizeMarkdownSlug(
    sanitizeMethodName(operation),
  )}`;
  return `[\`${sanitizeFuncName(operation)}\`](${link})`;
}

function templateReactQueryHookMarkdownLink(operation: Operation): string {
  const filename = getSDKReadmeFileName(operation.OwningSDK);
  const name = sanitizeReactQueryName(
    operation,
    hasMutationHook(operation) ? "mutation" : "query",
  );

  const link = `${filename}#${sanitizeMarkdownSlug(
    sanitizeMethodName(operation),
  )}`;
  return `[\`${name}\`](${link})`;
}

function sanitizePublicExportGroupPart(part: string): string {
  return sanitizeClassName(String(part || "").trim());
}

function sanitizePublicExportName(name: string): string {
  return caser().ToPascal(sanitizeName(name));
}

function uniquePublicExportAlias(base: string, used: Set<string>): string {
  let candidate = base;
  let index = 2;
  while (used.has(candidate)) {
    candidate = `${base}${index}`;
    index += 1;
  }
  used.add(candidate);
  return candidate;
}

function publicExportTargetCanMergeNamespace(target: TypeDef): boolean {
  const targetType = target.Type?.toString();
  return targetType === "class" || targetType === "error";
}

// The resources surface location comes from imports.paths.resources in
// gen.yaml. Configuring it also opts the SDK into implicit model-namespace
// exports; explicit x-speakeasy-exports render at the default location
// whether or not it is configured.
function tsConfiguredResourcesPath(): string {
  // Normalize away empty, ".", and ".." segments so values like
  // "./resources" cannot produce malformed output paths or export keys.
  return (context.Global.Config.Imports?.GetResourcesPath?.() ?? "")
    .split("/")
    .filter((segment) => segment !== "" && segment !== "." && segment !== "..")
    .join("/");
}

function tsResourcesModulePath(): string {
  return tsConfiguredResourcesPath() || "resources";
}
registerTemplateFunc("tsResourcesModulePath", tsResourcesModulePath);

// A resources path overlapping a model-scope location (or a core SDK
// directory) would have the export surface and other jobs writing into the
// same place nondeterministically; the surface is skipped instead.
function tsResourcesPathCollides(): boolean {
  const modulePath = tsResourcesModulePath();
  const reserved = [
    getModelsLocation(""),
    getErrorsLocation(),
    getSharedLocation(),
    getOperationsLocation(),
    getWebhooksLocation(),
    getCallbacksLocation(),
    getTypesLocation(),
    "core",
    "funcs",
    "hooks",
    "index",
    "lib",
    "mcp-server",
    "react-query",
    "sdk",
  ].filter((p) => p);
  return reserved.some(
    (p) =>
      modulePath === p ||
      modulePath.startsWith(`${p}/`) ||
      p.startsWith(`${modulePath}/`),
  );
}

// Relative prefix from the resources module to other src-rooted locations,
// accounting for a configured path nested in subdirectories.
function tsResourcesImportPrefix(): string {
  const depth = tsResourcesModulePath().split("/").length - 1;
  return depth === 0 ? "./" : "../".repeat(depth);
}

// The public export surface renders as a single src/resources.ts module so
// SDK wrappers can `export *` (or pick names) from one place.
//
// Namespace members reference non-exported intermediate alias declarations
// rather than import aliases: .d.ts rollup tools (api-extractor) collapse
// import aliases onto their target's declared name, which would turn
// same-named members (`Agents.Agent`) into circular self-references. A real
// intermediate declaration survives the rollup, keeps every member
// resolvable, and surfaces each public name as a top-level declaration.
type TypeScriptPublicExportNamespaceBuilder = {
  Name: string;
  Members: TypeScriptPublicExportMember[];
  Children: Map<string, TypeScriptPublicExportNamespaceBuilder>;
  ChildrenByKey: Map<string, TypeScriptPublicExportNamespaceBuilder>;
  UsedChildNames: Set<string>;
};

function buildTypeScriptPublicExportFile():
  | TypeScriptPublicExportFile
  | undefined {
  if (context.GlobalComputed?._typeScriptPublicExportFile !== undefined) {
    return context.GlobalComputed._typeScriptPublicExportFile;
  }
  context.GlobalComputed ??= {};

  if (tsResourcesPathCollides()) {
    context.GlobalComputed._typeScriptPublicExportFile = undefined;
    return undefined;
  }

  const groups = (context.Global.AST.PublicExports?.Groups || [])
    .map((group) => ({
      group,
      keys: (group.Parts || [])
        .map((part) => String(part || "").trim())
        .filter((part) => part.length > 0),
      parts: (group.Parts || [])
        .map((part) => sanitizePublicExportGroupPart(part))
        .filter((part) => part.length > 0),
    }))
    .filter((entry) => entry.parts.length > 0)
    .sort((a, b) => a.parts.join("/").localeCompare(b.parts.join("/")));

  const includeImplicit = tsConfiguredResourcesPath() !== "";
  const importPrefix = tsResourcesImportPrefix();
  const importsByPath = new Map<string, TypeScriptPublicExportImport>();
  const importBindings = new Map<string, string>();
  const usedIdentifiers = new Set<string>();
  const intermediates: { Name: string; Target: string }[] = [];
  const flatByName = new Map<string, string>();
  const root = {
    Children: new Map<string, TypeScriptPublicExportNamespaceBuilder>(),
    ChildrenByKey: new Map<string, TypeScriptPublicExportNamespaceBuilder>(),
    UsedChildNames: new Set<string>(),
  };

  // In params-object mode, exports that target an operation request model
  // alias the generated <Operation>Params type (plus its streaming variants)
  // so the legacy flat-params names track the method signatures.
  const paramsStatesByRequestName = new Map<string, TSMethodParamsState>();
  if (tsParamsObjectEnabled()) {
    for (const state of tsMethodParamsFileStates()) {
      paramsStatesByRequestName.set(state.RequestName, state);
    }
  }

  const importBinding = (path: string, sourceName: string): string => {
    const key = `${path}#${sourceName}`;
    const existing = importBindings.get(key);
    if (existing !== undefined) {
      return existing;
    }
    const alias = uniquePublicExportAlias(
      `${sourceName}$Import`,
      usedIdentifiers,
    );
    importBindings.set(key, alias);
    let importBucket = importsByPath.get(path);
    if (!importBucket) {
      importBucket = { Path: path, Specifiers: [] };
      importsByPath.set(path, importBucket);
    }
    importBucket.Specifiers.push({ Name: sourceName, Alias: alias });
    return alias;
  };

  const namespaceNode = (
    parent: {
      Children: Map<string, TypeScriptPublicExportNamespaceBuilder>;
      ChildrenByKey: Map<string, TypeScriptPublicExportNamespaceBuilder>;
      UsedChildNames: Set<string>;
    },
    key: string,
    name: string,
  ): TypeScriptPublicExportNamespaceBuilder => {
    const existing = parent.ChildrenByKey.get(key);
    if (existing) {
      return existing;
    }

    const uniqueName = uniquePublicExportAlias(name, parent.UsedChildNames);
    const child = {
      Name: uniqueName,
      Members: [],
      Children: new Map<string, TypeScriptPublicExportNamespaceBuilder>(),
      ChildrenByKey: new Map<string, TypeScriptPublicExportNamespaceBuilder>(),
      UsedChildNames: new Set<string>(),
    };
    parent.ChildrenByKey.set(key, child);
    parent.Children.set(uniqueName, child);
    return child;
  };

  const nodeForParts = (
    keys: string[],
    parts: string[],
  ): TypeScriptPublicExportNamespaceBuilder => {
    let parent = root;
    let node: TypeScriptPublicExportNamespaceBuilder | undefined;
    for (let i = 0; i < parts.length; i++) {
      node = namespaceNode(parent, keys[i] || parts[i], parts[i]);
      parent = node;
    }
    if (!node) {
      throw new Error("public export namespace requires at least one part");
    }
    return node;
  };

  const addAlias = (
    members: TypeScriptPublicExportMember[],
    aliasName: string,
    path: string,
    sourceName: string,
    canInterface = false,
  ) => {
    const intermediate = uniquePublicExportAlias(
      `${aliasName}$`,
      usedIdentifiers,
    );
    intermediates.push({
      Name: intermediate,
      Target: importBinding(path, sourceName),
    });
    members.push({
      Name: aliasName,
      Intermediate: intermediate,
      CanInterface: canInterface,
    });
    if (!flatByName.has(aliasName)) {
      flatByName.set(aliasName, intermediate);
    }
  };

  for (const { group, keys, parts } of groups) {
    const node = nodeForParts(keys, parts);

    for (const publicExport of group.Exports || []) {
      const target = publicExport.Target;
      if (!target) {
        continue;
      }
      if (publicExport.Implicit && !includeImplicit) {
        continue;
      }

      const aliasName = sanitizePublicExportName(publicExport.Name || "");
      if (!aliasName) {
        continue;
      }

      const paramsState = paramsStatesByRequestName.get(
        sanitizeClassName(target.Name),
      );
      if (paramsState) {
        const paramsPath = `${importPrefix}${getOperationsLocation()}/method-params.js`;
        addAlias(node.Members, aliasName, paramsPath, paramsState.ParamsName);
        if (paramsState.SSE && paramsState.BodyVariants.length === 0) {
          const variantNode = namespaceNode(node, aliasName, aliasName);
          addAlias(
            node.Members,
            `${aliasName}NonStreaming`,
            paramsPath,
            `${paramsState.ParamsName}NonStreaming`,
          );
          addAlias(
            variantNode.Members,
            `${aliasName}NonStreaming`,
            paramsPath,
            `${paramsState.ParamsName}NonStreaming`,
          );
          addAlias(
            node.Members,
            `${aliasName}Streaming`,
            paramsPath,
            `${paramsState.ParamsName}Streaming`,
          );
          addAlias(
            variantNode.Members,
            `${aliasName}Streaming`,
            paramsPath,
            `${paramsState.ParamsName}Streaming`,
          );
        }
        continue;
      }

      addAlias(
        node.Members,
        aliasName,
        `${importPrefix}${getModelsLocation(
          target.OutputLocation,
        )}/${resolveModelName(target)}.js`,
        sanitizeClassName(target.Name),
        publicExportTargetCanMergeNamespace(target),
      );
    }
  }

  const finalizeNamespace = (
    node: TypeScriptPublicExportNamespaceBuilder,
    rootNamespace: boolean,
  ): TypeScriptPublicExportNamespace | undefined => {
    const namespaces = [...node.Children.values()]
      .map((child) => finalizeNamespace(child, false))
      .filter(
        (namespace): namespace is TypeScriptPublicExportNamespace =>
          namespace !== undefined,
      )
      .sort((a, b) => a.Name.localeCompare(b.Name));
    const childNames = new Set(namespaces.map((namespace) => namespace.Name));
    const members = node.Members.map((member) => ({
      ...member,
      MustInterface: childNames.has(member.Name) && member.CanInterface,
    })).sort((a, b) => a.Name.localeCompare(b.Name));

    if (members.length === 0 && namespaces.length === 0) {
      return undefined;
    }

    return {
      Declaration: rootNamespace
        ? "export declare namespace"
        : "export namespace",
      Name: node.Name,
      Members: members,
      Namespaces: namespaces,
    };
  };

  const namespaces = [...root.Children.values()]
    .map((node) => finalizeNamespace(node, true))
    .filter(
      (namespace): namespace is TypeScriptPublicExportNamespace =>
        namespace !== undefined,
    )
    .sort((a, b) => a.Name.localeCompare(b.Name));

  let file: TypeScriptPublicExportFile | undefined;
  if (namespaces.length > 0) {
    const rootNamespaceNames = new Set(
      namespaces.map((namespace) => namespace.Name),
    );
    file = {
      Imports: [...importsByPath.values()]
        .map((imp) => ({
          ...imp,
          Specifiers: imp.Specifiers.sort((a, b) =>
            a.Alias.localeCompare(b.Alias),
          ),
        }))
        .sort((a, b) => a.Path.localeCompare(b.Path)),
      Intermediates: intermediates,
      // A namespace declaration cannot merge with a type alias of the same
      // name, so root namespace names win over flat aliases.
      FlatAliases: [...flatByName.entries()]
        .filter(([name]) => !rootNamespaceNames.has(name))
        .map(([name, intermediate]) => ({
          Name: name,
          Intermediate: intermediate,
        }))
        .sort((a, b) => a.Name.localeCompare(b.Name)),
      Namespaces: namespaces,
    };
  }
  context.GlobalComputed._typeScriptPublicExportFile = file;
  return file;
}

function hasPublicExports(): boolean {
  return buildTypeScriptPublicExportFile() !== undefined;
}
registerTemplateFunc("hasPublicExports", hasPublicExports);

// Used by the package index when models are also star-exported (global
// imports + index modules): flat aliases typically share their model's name,
// so the index re-exports only the group namespaces to avoid TS2308
// ambiguity; the flat aliases stay importable from the resources subpath.
function tsPublicExportNamespaceNames(): string[] {
  return (buildTypeScriptPublicExportFile()?.Namespaces ?? []).map(
    (namespace) => namespace.Name,
  );
}
registerTemplateFunc(
  "tsPublicExportNamespaceNames",
  tsPublicExportNamespaceNames,
);

function getTypeScriptPublicExportJobs(path: string): TemplateFileJob[] {
  const file = buildTypeScriptPublicExportFile();
  if (!file) {
    return [];
  }
  return [
    createTemplateFileJob(
      "resources.ts.stmpl",
      `${path}/${tsResourcesModulePath()}.ts`,
      file,
    ),
  ];
}

// @ts-ignore
function getModelJobs(types: BucketedTypes, path: string): TemplateFileJob[] {
  const jobs: TemplateFileJob[] = [];

  let modelsToExport: Record<
    string,
    {
      templatedModels: string[];
      templatedTypes: string[];
      typesByFile: Record<string, string[]>;
      exportsByFile: Record<string, { name: string; typeOnly: boolean }[]>;
    }
  > = {};

  // Pre-add default API error
  modelsToExport[
    getModelsLocation(context.Global.Config.Imports.GetErrorsPath())
  ] = {
    templatedModels: [
      getDefaultErrorFileName(),
      getBaseErrorFileName(),
      getHttpClientErrorsFileName(),
      ...(isNoZod()
        ? []
        : [
            getSDKValidationErrorFileName(),
            getResponseValidationErrorFileName(),
          ]),
    ],
    templatedTypes: [
      getDefaultErrorClassName(),
      getBaseErrorClassName(),
      "HTTPClientError",
      ...(isNoZod() ? [] : ["SDKValidationError", "ResponseValidationError"]),
    ],
    typesByFile: {},
    exportsByFile: {},
  };

  for (const [outputLocation, models] of sequencedMapEntries(types)) {
    for (let [model, types] of sequencedMapEntries(models)) {
      let modelPath = getModelsLocation(outputLocation);

      if (!(modelPath in modelsToExport)) {
        modelsToExport[modelPath] = {
          templatedModels: [],
          templatedTypes: [],
          typesByFile: {},
          exportsByFile: {},
        };
      }

      let docGroup = path;
      if (modelPath) {
        docGroup += "/" + modelPath;
      }

      types = reorderTypesToAvoidUsageBeforeDeclaration(types);

      let servers: any;
      if (
        types.find((t) => t.Scope.toString() == "operations") ||
        types.length == 0
      ) {
        servers = context.Global.AST.OperationServers.Get(model)[0];
      }
      if (types.length === 0 && !servers) {
        continue;
      }

      let modelName = sanitizeTypeFilename(model);

      let ctx: TypescriptModel = {
        DocGroup: docGroup,
        Name: model,
        Types: types,
        OutputLocation: outputLocation,
        Servers: servers,
      };

      modelsToExport[modelPath].templatedModels.push(modelName);

      const fileTypes: string[] = [];
      const fileExports: { name: string; typeOnly: boolean }[] = [];
      for (const type of types) {
        let typeName = sanitizeClassName(type.Name);
        modelsToExport[modelPath].templatedTypes.push(`"${typeName}"`);
        fileTypes.push(typeName);

        // Track all exported identifiers for collision detection.
        // The main type name is type-only for class/union but a value for
        // enum (const+type) and error (class extends Error).
        const isEnum = type.Type === "enum";
        const isError = type.Type === "error";
        const mainIsTypeOnly = !isEnum && !isError;
        fileExports.push({ name: typeName, typeOnly: mainIsTypeOnly });

        // In no-zod nothing schema-related is emitted: no $inboundSchema,
        // $outboundSchema, $Outbound, or FromJSON/ToJSON helpers. The error
        // class itself is exported via the typeName entry above.
        if (!isNoZod()) {
          if (shouldIncludeInboundSchema(type)) {
            fileExports.push({
              name: typeName + "$inboundSchema",
              typeOnly: false,
            });
            if (!isEnum && !isError) {
              fileExports.push({
                name: sanitizeCamelCase(type.Name) + "FromJSON",
                typeOnly: false,
              });
            }
          }
          if (shouldIncludeOutboundSchema(type)) {
            fileExports.push({
              name: typeName + "$outboundSchema",
              typeOnly: false,
            });
            if (!isEnum) {
              fileExports.push({
                name: typeName + "$Outbound",
                typeOnly: true,
              });
            }
            if (!isEnum && !isError) {
              fileExports.push({
                name: sanitizeCamelCase(type.Name) + "ToJSON",
                typeOnly: false,
              });
            }
          }
        }
      }
      modelsToExport[modelPath].typesByFile[modelName] = fileTypes;
      modelsToExport[modelPath].exportsByFile[modelName] = fileExports;

      jobs.push(
        createTemplateFileJob(
          `modelfile.ts.stmpl`,
          `${path}/${modelPath}/${sanitizeTypeFilename(model)}.ts`,
          ctx,
        ),
      );
    }
  }

  // The method-params module is imported by SDK class methods whenever
  // params-object mode is on, so its job must be emitted regardless of the
  // index-modules setting.
  const methodParamsStates = tsMethodParamsFileStates();
  if (methodParamsStates.length > 0) {
    jobs.push(
      createTemplateFileJob(
        "method-params.ts.stmpl",
        `${path}/${getOperationsLocation()}/method-params.ts`,
        { Ops: methodParamsStates },
      ),
    );
  }

  if (context.Global.Config.UseIndexModules) {
    // When collisions exist, ensure the root models directory has an index
    // that re-exports all child namespaces for dotted access (models.types.Widget)
    if (hasNamespaceCollisions()) {
      const rootModels = getModelsLocation("");
      if (!(rootModels in modelsToExport)) {
        modelsToExport[rootModels] = {
          templatedModels: [],
          templatedTypes: [],
          typesByFile: {},
          exportsByFile: {},
        };
      }
    }

    const errorsModelPath = getModelsLocation(
      context.Global.Config.Imports.GetErrorsPath(),
    );

    // Map of built-in error file names to the type names they export
    const builtinErrorTypesByFile: Record<string, string[]> = {};
    builtinErrorTypesByFile[getHttpClientErrorsFileName()] = [
      "HTTPClientError",
      "UnexpectedClientError",
      "InvalidRequestError",
      "RequestAbortedError",
      "RequestTimeoutError",
      "ConnectionError",
    ];
    builtinErrorTypesByFile[getDefaultErrorFileName()] = [
      getDefaultErrorClassName(),
    ];
    builtinErrorTypesByFile[getBaseErrorFileName()] = [getBaseErrorClassName()];
    // SDKValidationError and ResponseValidationError are not emitted in
    // no-zod (no runtime validation throws either of them).
    if (!isNoZod()) {
      builtinErrorTypesByFile[getSDKValidationErrorFileName()] = [
        "SDKValidationError",
      ];
      builtinErrorTypesByFile[getResponseValidationErrorFileName()] = [
        "ResponseValidationError",
      ];
    }

    const builtinFileNames = new Set(Object.keys(builtinErrorTypesByFile));

    if (methodParamsStates.length > 0) {
      const opLocation = getOperationsLocation();
      modelsToExport[opLocation] ??= {
        templatedModels: [],
        templatedTypes: [],
        typesByFile: {},
        exportsByFile: {},
      };
      modelsToExport[opLocation].templatedModels.push("method-params");
    }

    for (var modelPath in modelsToExport) {
      let builtinExportLines: string[] = [];

      if (modelPath === errorsModelPath) {
        // Detect collisions: built-in type names that also have a spec-generated model file
        // or type with the same name. A collision occurs when a non-builtin model file would
        // export a type with the same name as a built-in error type. This can happen either
        // because the file has the same name (e.g. spec generates "requesttimeouterror.ts")
        // or because a differently-named file exports a type with a built-in name (e.g.
        // "get-server-identity.ts" exports RequestTimeoutError).
        const collidingBuiltinTypes = new Set<string>();

        // Collect all type names exported by spec-defined (non-builtin) files
        const specDefinedTypeNames = new Set<string>();
        for (const [fileName, fileTypes] of Object.entries(
          modelsToExport[modelPath].typesByFile,
        )) {
          if (!builtinFileNames.has(fileName)) {
            for (const t of fileTypes) {
              specDefinedTypeNames.add(t);
            }
          }
        }

        for (const typeNames of Object.values(builtinErrorTypesByFile)) {
          for (const typeName of typeNames) {
            // Check 1: Spec creates a file with same name as built-in type
            const sanitized = sanitizeTypeFilename(typeName);
            if (
              modelsToExport[modelPath].templatedModels.includes(sanitized) &&
              !builtinFileNames.has(sanitized)
            ) {
              collidingBuiltinTypes.add(typeName);
            }
            // Check 2: A spec-defined file exports a type matching a built-in name
            if (specDefinedTypeNames.has(typeName)) {
              collidingBuiltinTypes.add(typeName);
            }
          }
        }

        if (collidingBuiltinTypes.size > 0) {
          // Remove built-in files from the models list and add explicit exports
          const filteredModels = modelsToExport[
            modelPath
          ].templatedModels.filter((m) => !builtinFileNames.has(m));
          modelsToExport[modelPath].templatedModels = filteredModels;

          for (const [filename, typeNames] of Object.entries(
            builtinErrorTypesByFile,
          )) {
            const nonColliding = typeNames.filter(
              (t) => !collidingBuiltinTypes.has(t),
            );
            if (nonColliding.length > 0) {
              builtinExportLines.push(
                `export { ${nonColliding.join(
                  ", ",
                )} } from "./${filename}.js";`,
              );
            }
          }
        }
      }

      // Detect files whose exported types are all already provided by other
      // files. When two files export the same type (e.g. a dedicated model file
      // and an operation file that inlines the same type), remove the one whose
      // types are fully subsumed to avoid TS2308 duplicate export errors.
      const typeFirstFile = new Map<string, string>();
      const filesFullySubsumed = new Set<string>();

      for (const model of modelsToExport[modelPath].templatedModels) {
        const fileTypes = modelsToExport[modelPath].typesByFile[model] || [];
        let allSubsumed = fileTypes.length > 0;
        for (const typeName of fileTypes) {
          if (!typeFirstFile.has(typeName)) {
            typeFirstFile.set(typeName, model);
            allSubsumed = false;
          }
        }
        if (allSubsumed && fileTypes.length > 0) {
          filesFullySubsumed.add(model);
        }
      }

      if (filesFullySubsumed.size > 0) {
        modelsToExport[modelPath].templatedModels = modelsToExport[
          modelPath
        ].templatedModels.filter((m) => !filesFullySubsumed.has(m));
      }

      // Detect partial overlaps: exported identifiers that appear in multiple
      // files. When two files export the same identifier via `export *`, TS
      // emits TS2308. We resolve this by adding explicit re-exports from the
      // first file (alphabetically) for the colliding identifiers.
      const identFirstFile = new Map<string, string>();
      const collidingIdents = new Map<
        string,
        { firstFile: string; typeOnly: boolean }
      >();
      const sortedModels = [
        ...modelsToExport[modelPath].templatedModels,
      ].sort();

      for (const model of sortedModels) {
        const fileExports =
          modelsToExport[modelPath].exportsByFile[model] || [];
        for (const exp of fileExports) {
          if (!identFirstFile.has(exp.name)) {
            identFirstFile.set(exp.name, model);
          } else {
            collidingIdents.set(exp.name, {
              firstFile: identFirstFile.get(exp.name)!,
              typeOnly: exp.typeOnly,
            });
          }
        }
      }

      if (collidingIdents.size > 0) {
        // Group colliding identifiers by the file they should be re-exported from
        const reExportsByFile = new Map<
          string,
          { types: string[]; values: string[] }
        >();
        for (const [ident, info] of collidingIdents) {
          if (!reExportsByFile.has(info.firstFile)) {
            reExportsByFile.set(info.firstFile, { types: [], values: [] });
          }
          const bucket = reExportsByFile.get(info.firstFile)!;
          if (info.typeOnly) {
            bucket.types.push(ident);
          } else {
            bucket.values.push(ident);
          }
        }

        for (const [fileName, { types, values }] of reExportsByFile) {
          if (values.length > 0) {
            builtinExportLines.push(
              `export { ${values.sort().join(", ")} } from "./${fileName}.js";`,
            );
          }
          if (types.length > 0) {
            builtinExportLines.push(
              `export type { ${types
                .sort()
                .join(", ")} } from "./${fileName}.js";`,
            );
          }
        }
      }

      const locals: any = {
        Models: modelsToExport[modelPath].templatedModels.sort(),
        BuiltinExportLines: builtinExportLines,
      };

      // When namespace collisions exist, add re-exports to root-level model
      // directories so types can be accessed via dotted paths like models.types.Widget
      if (hasNamespaceCollisions()) {
        const reexports: { Name: string; Path: string }[] = [];
        for (const childPath in modelsToExport) {
          if (
            childPath !== modelPath &&
            childPath.startsWith(modelPath + "/")
          ) {
            const relative = childPath.slice(modelPath.length + 1);
            // Only direct children (no nested slashes)
            if (!relative.includes("/")) {
              reexports.push({ Name: relative, Path: relative });
            }
          }
        }
        if (reexports.length > 0) {
          locals.NamespaceReexports = reexports.sort((a, b) =>
            a.Name.localeCompare(b.Name),
          );
        }
      }

      jobs.push(
        createTemplateFileJob(
          `index.ts.stmpl`,
          `${path}/${modelPath}/index.ts`,
          locals,
        ),
      );
    }
  }

  return jobs;
}

// Do a topological sort of the types to avoid usage before declaration
function reorderTypesToAvoidUsageBeforeDeclaration(
  types: TypeDef[],
): TypeDef[] {
  return topologicalSortTypeDefs(types);
}

// @ts-ignore
function templateGlobalFieldName(name: string): string {
  let reserved = [
    getConstants().security,
    getConstants().defaultClient,
    getConstants().serverURL,
  ];

  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.ServerMap
  ) {
    reserved.push(getConstants().server);
  } else {
    reserved.push(getConstants().serverIdx);
  }

  let fieldName = sanitizeFieldName(name);

  if (reserved.includes(name)) {
    fieldName += "Global";
  }

  return fieldName;
}

registerTemplateFunc("templateGlobalFieldName", templateGlobalFieldName);

// @ts-ignore
function getGlobalSecurity(
  options?: SecurityUsageContext | Example,
  additionalContext?: TemplateValueContext,
): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example, additionalContext);
    }
  }

  if (
    index === undefined &&
    example === undefined &&
    context.Global.AST.MainSDK.Security.Optional &&
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason ==
      "optional-scheme"
  ) {
    return "";
  }

  let securityFieldName = "security";
  const flattenable = canFlattenGlobalSecurity();
  if (flattenable) {
    securityFieldName = sanitizeFieldName(
      context.Global.AST.MainSDK.Security.Type.Fields[0].Name,
    );
  }

  return `${securityFieldName}: ${templateSecurityUsage(
    context.Global.AST.MainSDK.Security.Type,
    0,
    true,
    flattenable,
    index,
    example,
    additionalContext,
  ).trim()},`;
}

// @ts-ignore
function joinSDKOptions(options: string[]): string {
  if (options.length > 0) {
    options.unshift("");

    return "{" + indentLines(options, 1) + "\n}";
  }

  return "";
}

function templateUsageSDKInit(local: UsageContext): string {
  if (context.Global.Config.UsageSdkInit) {
    if (
      context.Global.Config.UsageSdkInitImports &&
      Array.isArray(context.Global.Config.UsageSdkInitImports)
    ) {
      for (const importConfig of context.Global.Config.UsageSdkInitImports) {
        const pkg = importConfig.package || importConfig.Package;
        const imp = importConfig.import || importConfig.Import;
        const type = importConfig.type || importConfig.Type || "typeImport";

        let importType: TSImportType = typeImport;
        if (type === "packageImport") {
          importType = packageImport;
        } else if (type === "aliasImport") {
          importType = aliasImport;
        }

        addUsageImport(pkg, imp, importType);
      }
    }
    return context.Global.Config.UsageSdkInit;
  }

  const sdkName = sanitizeClassName(context.Global.AST.MainSDK.Type.Name);
  return `new ${sdkName}(${templateUsageSDKOptions(local)})`;
}

registerTemplateFunc("templateUsageSDKInit", templateUsageSDKInit);

// @ts-ignore
function templateUsageSDKOptions(local: UsageContext): string {
  const options = [];
  let hasSecurity = local.Operation.Security != undefined;
  let hasAbsoluteServerURL =
    context.Global.AST.MainSDK.Servers?.HasAbsoluteURL() || false;

  const ctx: TemplateValueContext = {
    usageContext: local, // TODO see if this can be decomposed into the required fields below
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
  };

  const hasOperationServers = local.Operation?.Servers?.Servers?.length > 0;

  if (!hasAbsoluteServerURL && !hasOperationServers) {
    const url = getUsageServerUrl(local, undefined, ctx);
    options.push(`${getConstants().serverURL}: ${url},`);
  }

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (scope.IsGlobal && scope.Feature != "") {
        const feature = scope.Feature.toString();

        switch (feature) {
          case "security":
            if (!hasSecurity) {
              const globalSecurityUsage = getGlobalSecurity(scope.Value, ctx);
              if (globalSecurityUsage) {
                options.push(globalSecurityUsage);
                hasSecurity = true;
              }
            }
            break;
          case "server_url": {
            if (hasAbsoluteServerURL && !hasOperationServers) {
              options.push(
                `${getConstants().serverURL}: ${getUsageServerUrl(
                  local,
                  scope,
                  ctx,
                )},`,
              );
            }
            break;
          }
          case "server_selection": {
            const server: UsageGlobalServer = getUsageGlobalServer(
              !!scope.Value,
            );

            if (server.ID) {
              options.push(`${getConstants().server}: "${server.ID}",`);
            } else if (server.Index !== undefined) {
              options.push(`${getConstants().serverIdx}: ${server.Index},`);
            }

            if (server.Variables) {
              server.Variables.forEach((v: ServerVariable) => {
                const fieldName = sanitizeFieldName(v.Name);
                options.push(
                  `${fieldName}: "${getUsageServerVariableValue(v)}",`,
                );
              });
            }

            break;
          }
          case "retries": {
            options.push(
              `${getConstants().retryConfig}: ${templateString(
                "usage/retries.stmpl",
                {},
              )},`,
            );
            break;
          }
          case "http_client": {
            options.push(
              `${getConstants().httpClient}: ${templateHTTPClient(
                scope.Value,
              )},`,
            );
            break;
          }
          case "parameter": {
            const parameter = scope.Value as ParameterUsage;

            const value = templateValueAsRequired(
              parameter.field,
              getExampleValue(parameter.example),
              ctx,
            );

            if (value === "") {
              return;
            }

            options.push(
              `${templateGlobalFieldName(parameter.field.Name)}: ${value},`,
            );

            break;
          }
        }
      }
    });
  }

  if (!hasSecurity && opUsesGlobalSecurity(local.Operation)) {
    const globalSecurityUsage = getGlobalUsageSecurity(local);
    if (globalSecurityUsage) {
      options.push(globalSecurityUsage);
    }
  }

  return joinSDKOptions(options);
}

registerTemplateFunc("templateUsageSDKOptions", templateUsageSDKOptions);

// @ts-ignore
function templateCmsComment(comments: SimpleCommentDef): string[] {
  let lines = [];

  let firstLineParts = [];

  let descriptionLines = [];
  if (comments.Description) {
    descriptionLines = comments.Description.split("\n");
  }

  if (comments.Summary) {
    firstLineParts.push(comments.Summary);
  } else if (descriptionLines.length > 0) {
    firstLineParts.push(descriptionLines.shift());
  }

  const firstLine = firstLineParts.join(" - ");

  if (firstLine) {
    lines.push(...sanitizeComments(firstLine).split("\n"));
  }

  if (descriptionLines.length > 0) {
    lines.push("");
    lines.push("@remarks");
    lines.push(...descriptionLines.map(sanitizeComments));
  }

  return lines;
}

// @ts-ignore
function finalizeComments(lines: string[], indent: number): string {
  if (lines.length == 0) {
    return "";
  }

  lines = lines.map((line) => {
    return ` * ${line}`;
  });

  lines.unshift("/**");
  lines.push(" */");

  return "\n" + indentLines(lines, indent);
}

// @ts-ignore
function templateComments(
  comments: CommentDef | null,
  docgroup: string | null,
  title: string | null,
  indent: number,
  type: "method" | "field" | "class" | "enum" | "const" | "var" = null,
  omitExternalDocs: boolean = false,
  additionalNotes: string = "",
): string {
  const resolved = resolveCommentDef(
    "openapi",
    docgroup,
    type ?? "",
    title,
    comments,
  );
  const cmsComment = templateCmsComment(resolved);

  if (!comments && !additionalNotes) {
    return finalizeComments(cmsComment, indent);
  }

  let lines = cmsComment;

  if (comments?.ExternalDocs && !omitExternalDocs) {
    let externalDocsLines = [];

    if (comments.ExternalDocs.Description) {
      externalDocsLines = sanitizeComments(
        comments.ExternalDocs.Description,
      ).split("\n");
    }

    lines.push("");
    lines.push(
      `@see {@link ${comments.ExternalDocs.URL}}${
        externalDocsLines.length > 0 ? " - " + externalDocsLines.shift() : ""
      }`,
    );
    lines.push(...externalDocsLines);
  }

  if (additionalNotes) {
    if (lines.length > 0) {
      lines.push("");
    }

    lines.push(...additionalNotes.split("\n"));
  }

  if (comments?.Deprecated) {
    if (lines.length > 0) {
      lines.push("");
    }

    let deprecated = `@deprecated ${type}: ${comments.DeprecationMessage}.`;

    if (comments.DeprecationReplacement) {
      const replacement = sanitizeDeprecationReplacement(
        comments.DeprecationReplacement,
        type as "method" | "field" | "class",
      );
      if (replacement) {
        deprecated += ` Use ${replacement} instead.`;
      }
    }

    lines.push(deprecated);
  }

  if (lines.length == 0) {
    return "";
  }

  return finalizeComments(lines, indent);
}

registerTemplateFunc("templateComments", templateComments);

// @ts-ignore
function templateHoistedSecurityRemark(
  fields: HoistedSecurityField[],
  required: boolean,
): string {
  const fieldList = fields.map(
    (f) => `{@link Security.${sanitizeFieldName(f.Name)}}`,
  );

  const remark = (items: string) =>
    required
      ? `This operation requires ${items} to be set on the \`security\` parameter when initializing the SDK.`
      : `If set, this operation will use ${items} from the global security.`;

  if (fieldList.length === 1) {
    return remark(fieldList[0]);
  }

  const last = fieldList.pop();
  if (fieldList.length === 1) {
    return remark(`either ${fieldList[0]} or ${last}`);
  }

  return remark(`one of ${fieldList.join(", ")}, or ${last}`);
}

// @ts-ignore
function templateBuiltinComment(
  docgroup: string | null,
  title: string,
  description: string,
  indent: number,
): string {
  const resolved = resolveComment(
    "builtin",
    docgroup,
    "",
    title,
    "",
    description,
  );
  return finalizeComments(templateCmsComment(resolved), indent);
}

registerTemplateFunc("templateBuiltinComment", templateBuiltinComment);

// @ts-ignore
function getOptionalUsageMethodParameters(local: UsageContext): string {
  if (!local.Scopes) {
    return "";
  }

  const options = local.Scopes.map((scope) => {
    if (scope.IsGlobal) {
      return "";
    }

    switch (scope.Feature.toString()) {
      case "server_url": {
        return `${getConstants().serverURL}: ${getUsageServerUrl(local, scope, {
          isTest: local.Test && true,
        })},`;
      }
      case "retries": {
        return `retries: ${templateString("usage/retries.stmpl", {})},`;
      }
      case "content_type": {
        const enumClass = sanitizeAcceptEnumName(local.Operation);

        addTestImport(
          `../sdk/${sanitizeTypeFilename(local.SDK.Type.Name)}.js`,
          enumClass,
        );
        return `acceptHeaderOverride: ${enumClass}.${sanitizeAcceptEnumKey(
          scope.OpFilter,
        )},`;
      }
    }
  }).filter(Boolean);

  return options.length > 0 ? `{\n${indentLines(options, 1)}\n}` : "";
}

// @ts-ignore
function templateUsageLocalDeclarations(usageContext: UsageContext): string {
  const declarations = [];
  const operation = usageContext.Operation;

  const ctx: TemplateValueContext = {
    usageContext: usageContext,
    operation: usageContext.Operation,
    isTest: usageContext.Test && true,
    test: usageContext.Test,
  };

  for (const field of operation.Arguments.Sorted) {
    if (isSecurityClassField(field)) {
      continue;
    }

    const fieldExample = getOperationMethodFieldExample(usageContext, field);

    const isReferenceWithReplacements =
      fieldExample &&
      isExampleReferenceValue(fieldExample) &&
      fieldExample?.replacements.length > 0;

    if (!isReferenceWithReplacements) {
      continue;
    }

    let additionalCode = "";

    const fieldName = sanitizeFieldName(field.Name);

    for (const replacement of fieldExample.replacements) {
      let basePath = fieldName;

      const path = templateJSONPointerPath(
        field,
        basePath,
        replacement.Path,
        undefined,
      );

      const val = templateValue(
        path.target,
        getExampleValue(replacement.Value, { usageContext: usageContext }),
        false,
        ctx,
      );

      additionalCode += `\n${path.path} = ${val};`;
    }

    const value = templateValue(field, fieldExample, false, ctx);

    if (value === "") {
      continue;
    }

    declarations.push(`const ${fieldName} = ${value};${additionalCode}`);
  }

  if (!declarations.length) {
    return "";
  }

  return "\n" + declarations.join("\n\n") + "\n";
}
registerTemplateFunc(
  "templateUsageLocalDeclarations",
  templateUsageLocalDeclarations,
);

// @ts-ignore
function templateUsageMethodParameters(
  local: UsageContext,
  indent: number,
): string {
  const operation = local.Operation;
  const args = operation.Arguments;
  const methodParams = [];

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
    exampleName: local.ExampleName,
    shouldTemplateConstValue: (field, additionalContext) => {
      if (additionalContext?.withinUnion) {
        return true;
      }

      if (isConstFieldsAlwaysOptional()) return false;
      return !field.Optional;
    },
  };

  // We can't skip templating optional fields, if there will be a later
  // field included in the method call otherwise we'll get argument type mismatch
  // TODO: come back and revisit this as it looks like the code omits parameters after the first optional parameter that isn't included
  let lastIncludedFieldIndex = -1;
  const examples = [];
  for (const [i, field] of args.Sorted.entries()) {
    if (isSecurityClassField(field)) continue;
    const fieldExample = getOperationMethodFieldExample(local, field);
    examples[i] = fieldExample;
    if (includeField(field, fieldExample, false, ctx)) {
      lastIncludedFieldIndex = i;
    }
  }

  for (const [i, field] of args.Sorted.entries()) {
    let value = "";

    if (isSecurityClassField(field)) {
      let securityScope = undefined;

      for (const scope of local.Scopes) {
        if (scope.Feature == "security") {
          securityScope = scope;
          break;
        }
      }

      let securityExample = undefined;
      if (securityScope?.Value) {
        if ("example" in securityScope.Value) {
          securityExample = securityScope.Value.example;
        } else {
          securityExample = getExampleValue(securityScope.Value, {
            usageContext: local,
          });
        }
      }

      if (securityExample === undefined && field.Optional) {
        continue;
      }

      value = templateSecurityUsage(
        operation.Security.Type,
        indent,
        true,
        false,
        undefined,
        securityExample,
        ctx,
      ).trimStart();
    } else {
      const fieldExample = examples[i];

      if (i > lastIncludedFieldIndex) {
        continue;
      }

      value = templateValue(field, fieldExample, false, ctx);
    }

    if (value === "") {
      if (args.Flattening == "params" || args.Flattening == "all") {
        methodParams.push("undefined");
      }

      continue;
    }
    methodParams.push(value);
  }

  const options = getOptionalUsageMethodParameters(local);
  if (options) {
    // When all method params are optional and omitted, options would
    // incorrectly become the first argument. Pad with undefined to
    // ensure options is in the correct argument position.
    while (methodParams.length < args.Sorted.length) {
      methodParams.push("undefined");
    }
    methodParams.push(options);
  }

  // Remove the "undefined" values from the end of the array.
  while (
    methodParams.length > 0 &&
    methodParams[methodParams.length - 1] === "undefined"
  ) {
    methodParams.pop();
  }

  return methodParams.join(", ").trim();
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

// @ts-ignore
function inSameModel(
  modelName: string,
  outputLocation: string,
  typeDef: TypeDef,
): boolean {
  if (!typeDef.IsCustomType()) {
    return true;
  }

  if (!typeDef.ResolvedModel) {
    throw new Error(
      `Type '${typeDef.Name}' is not resolved '${typeDef.GetRegistrationID()}'`,
    );
  }

  return (
    `${outputLocation}${modelName}` ===
    `${getModelsLocation(typeDef.OutputLocation)}${typeDef.ResolvedModel}`
  );
}

function shouldTemplateAcceptHeaderEnum(op: Operation): boolean {
  return (
    context.Global.Config.AcceptHeaderEnum && op.GetAcceptTypes().length > 1
  );
}
registerTemplateFunc(
  "shouldTemplateAcceptHeaderEnum",
  shouldTemplateAcceptHeaderEnum,
);

function formatDefaultAcceptHeader(op: Operation): string {
  return `"${op.GetAcceptTypes().join(", ")}"`;
}
registerTemplateFunc("formatDefaultAcceptHeader", formatDefaultAcceptHeader);

function templateMethodOptions(op: Operation, withURLOverride = true) {
  let out: string[] = [];

  if (shouldTemplateAcceptHeaderEnum(op)) {
    out.push(`acceptHeaderOverride?: ${sanitizeAcceptEnumName(op)};`);
  }

  if (
    withURLOverride &&
    op.Extensions.Pagination != null &&
    op.Extensions.Pagination.Type == "url"
  ) {
    out.push("[URL_OVERRIDE]?: URL;");
  }

  // extraBody can only be merged into JSON object request bodies, so
  // operations without one do not accept it.
  let base = "RequestOptions";
  if (
    context.Global.Config.RequestExtras === true &&
    !/^(application|text)\/([^+]+\+)*json/.test(getRequestMediaType(op))
  ) {
    base = `Omit<RequestOptions, "${sanitizeFieldName("extraBody")}">`;
  }

  return out.length ? `${base} & { ${out.join(" ")} }` : base;
}
registerTemplateFunc("templateMethodOptions", templateMethodOptions);

function templateGitIgnore() {
  const prefixes = new Set();

  prefixes.add(getTypesLocation());
  prefixes.add(getModelsLocation(""));
  prefixes.add(getErrorsLocation());

  return Array.from(prefixes)
    .sort()
    .filter(Boolean)
    .map((v) => `/${v}`)
    .join("\n");
}
registerTemplateFunc("templateGitIgnore", templateGitIgnore);

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return "null";
  }

  return JSON.stringify(scopes);
}
registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

function ensureQuoted(str: string): string {
  // Check if already quoted
  if (
    (str.startsWith('"') && str.endsWith('"')) ||
    (str.startsWith("'") && str.endsWith("'"))
  ) {
    return str;
  }
  return `"${str}"`;
}

function templateFieldRemaps(
  direction: "inbound" | "outbound",
  fields: FieldDef[],
): { mappingObject: string; additionalPropsField: FieldDef } {
  let remaps = "";
  let additionalPropsField = undefined;

  for (const field of fields) {
    const fieldName = declareModelField(field);

    let origName = originalFieldName(field);
    if (isParameterField(field)) {
      origName = field.Name;
    } else if (isSecurityField(field)) {
      origName = outboundSecurityKeyName(field, fields);
    }

    if (field.IsAdditionalProperties) {
      additionalPropsField = field;
      if (!shouldFlattenAdditionalProperties(field)) {
        remaps += direction === "outbound" ? `  ${fieldName}: null,\n` : "";
      }
      continue;
    }

    if (fieldName === origName) {
      continue;
    }

    remaps +=
      direction === "inbound"
        ? `  "${origName}": ${ensureQuoted(fieldName)},\n`
        : `  ${fieldName}: "${origName}",\n`;
  }

  remaps = remaps ? `{\n${remaps}}` : "";

  return { mappingObject: remaps, additionalPropsField };
}

registerTemplateFunc("templateFieldRemaps", templateFieldRemaps);

function templateServerListName(op: Operation, relativeTo = ""): string {
  const model = getOperationModelName(op);
  const outputLocation = getModelsLocation(getOperationsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);
  const listName = `${templateConstName(model, "")}ServerList`;

  addImport(`${prefix}/${sanitizeFileName(model)}.js`, listName, typeImport);

  return listName;
}
registerTemplateFunc("templateServerListName", templateServerListName);

function templateServerID(op: Operation, id: string, relativeTo = ""): string {
  const model = getOperationModelName(op);
  const outputLocation = getModelsLocation(getOperationsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);
  const idconst = templateConstName(id, `${sanitizeFieldName(model)}Server`);

  addImport(`${prefix}/${sanitizeFileName(model)}.js`, idconst, typeImport);

  return idconst;
}
registerTemplateFunc("templateServerID", templateServerID);

function selectFuncExample(): UsageContext {
  const examples = selectExampleOperations(
    context.Global.AST.MainSDK,
    1,
    [],
    true,
  );

  return examples[0];
}

registerTemplateFunc("selectFuncExample", selectFuncExample);

function templateModelSnippet(typeDef: TypeDef): string {
  addUsageImportForType(typeDef);

  if (!canTemplateModelUsage(typeDef)) {
    return "// No examples available for this model";
  }

  const typeName = sanitizeClassName(typeDef.Name);

  seedFaker(typeName);

  const example = templateModelUsage(typeDefToFieldDef(typeDef), 0);

  let code = formatUsageSnippetOutput(
    `let value: ${typeName} = ${example};`,
  ).trimEnd();

  if (typeDef.Type.toString() === "enum" && typeDef.Enum?.Open) {
    const isString = typeDef.Enum?.Type.Type.toString() === "string";
    const tsType = isString ? "string" : "number";
    code += `\n\n// Open enum: unrecognized values are captured as Unrecognized<${tsType}>`;
  }

  return code;
}
registerTemplateFunc("templateModelSnippet", templateModelSnippet);

// @ts-ignore
function templateRequest(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return 'new Request("https://example.com")';
}

// @ts-ignore
function templateResponse(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return 'new Response(\'{"message": "hello world"}\', {headers: {"Content-Type": "application/json"}})';
}

// @ts-ignore
function templateStream(
  fieldDef: FieldDef,
  filePath: string,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (filePath) {
    if (additionalContext?.isTest) {
      filePath = `.speakeasy/testfiles/${filePath}`;

      if (additionalContext?.isResponse) {
        addTestImport("./files.js", "filesToByteArray");
        return `(await filesToByteArray("${filePath}"))`;
      } else {
        addTestImport("./files.js", "filesToStream");
        return `filesToStream("${filePath}")`;
      }
    } else {
      addUsageImport("node:fs", "openAsBlob");

      return `await openAsBlob("${filePath}")`;
    }
  } else {
    if (typeof example !== "string") {
      example = JSON.stringify(example);
    }
    const byteValue = templateByteValue(example, fieldDef, additionalContext);

    if (additionalContext?.isResponse) {
      return byteValue;
    } else {
      addTestImport("./files.js", "bytesToStream");
      return `bytesToStream(${byteValue})`;
    }
  }
}

// @ts-ignore
function templateFileToByteArrayValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    addTestImport("./files.js", "filesToByteArray");
    filePath = `.speakeasy/testfiles/${filePath}`;

    return `await filesToByteArray("${filePath}")`;
  }

  return `new Uint8Array(/* Populate with bytes from file, for example ${filePath} */)`;
}

// @ts-ignore
function templateFileToStringValue(
  fieldDef: FieldDef,
  filePath: string,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  if (additionalContext?.isTest) {
    addTestImport("./files.js", "filesToString");
    filePath = `.speakeasy/testfiles/${filePath}`;
    return `await filesToString("${filePath}")`;
  }

  return `"" // Populate with string from file, for example ${filePath}`;
}

function templatePaginationInputAccessor(
  op: Operation,
  inputVar: string,
  field: PaginationInputs,
): string {
  const reqField = op.Request?.Field;
  if (!reqField) {
    throw new Error(
      "Operation does not have a request to read pagination input from",
    );
  }

  const usesRequestWraper = hasAnnotation(reqField, "requestWrapper");

  const inLoc = field.In.toString();

  if (inLoc === "requestBody" && usesRequestWraper) {
    const bodyField = reqField.Type?.Fields.find((f) =>
      hasAnnotation(f, "request"),
    );
    if (!bodyField) {
      throw new Error(
        "Expected request wrapper does not have a request body field",
      );
    }

    // Find the actual field def for the pagination input field
    const paginationFieldDef = bodyField.Type?.FindFieldByName(field.Name);

    const bodyAccessor = sanitizeFieldName(bodyField.Name);
    const fieldAccessor = paginationFieldDef
      ? accessModelField(paginationFieldDef, false)
      : sanitizeFieldName(field.Name);

    return `${inputVar}?.${bodyAccessor}?.${fieldAccessor}`;
  } else if (inLoc === "requestBody" && !usesRequestWraper) {
    // Find the actual field def for the pagination input field
    const paginationFieldDef = reqField.Type?.FindFieldByName(field.Name);
    const fieldAccessor = paginationFieldDef
      ? accessModelField(paginationFieldDef, false)
      : sanitizeFieldName(field.Name);

    return `${inputVar}?.${fieldAccessor}`;
  } else if (inLoc === "parameters") {
    // Find the parameter field def from query/path/header params
    let paramFieldDef: FieldDef | undefined;
    const allParams = [
      ...(op.Request?.Params?.QueryParams || []),
      ...(op.Request?.Params?.PathParams || []),
      ...(op.Request?.Params?.HeaderParams || []),
    ];

    for (const param of allParams) {
      if (param.Field.Name === field.Name) {
        paramFieldDef = param.Field;
        break;
      }
    }

    const fieldAccessor = paramFieldDef
      ? accessModelField(paramFieldDef, false)
      : sanitizeFieldName(field.Name);

    return `${inputVar}?.${fieldAccessor}`;
  } else {
    throw new Error(
      `Could not template pagination input accessor for operation "${op.ID}": "${field.Name}" in ${inLoc}`,
    );
  }
}
registerTemplateFunc(
  "templatePaginationInputAccessor",
  templatePaginationInputAccessor,
);

// @ts-ignore
function templateHTTPClient(value: HTTPClientUsage): string {
  switch (value.type) {
    case "test":
      return `testHttpClient`;
    default:
      throw new Error(`unsupported http client type: ${value.type}`);
  }
}

function templateNextPageState(
  op: Operation,
): { bareType: string; staticType: string; value: string } | null {
  const pagination = op.Extensions.Pagination;
  const isPaginated = pagination && op.Response?.Responses.length > 0;
  if (!isPaginated) {
    throw new Error(
      `${op.ID}: templateNextPageState should only be called for paginated operations`,
    );
  }

  const hasURL = hasPaginationURL(op);

  const pType = pagination.Type.toString();

  const pageInput = getPaginationInput(pagination, "page");
  const offsetInput = getPaginationInput(pagination, "offset");
  const cursorInput = getPaginationInput(pagination, "cursor");

  if (pType === "offsetLimit" && pageInput) {
    return {
      bareType: "number",
      staticType: "{ page: number }",
      value: "{ page: nextPage }",
    };
  } else if (pType === "offsetLimit" && offsetInput) {
    return {
      bareType: "number",
      staticType: "{ offset: number }",
      value: "{ offset: nextOffset }",
    };
  } else if (pType === "cursor" && cursorInput) {
    const bareType = getCursorType(op);
    return {
      bareType,
      staticType: `{ cursor: ${bareType} }`,
      value: "{ cursor: nextCursor }",
    };
  } else if (hasURL) {
    return {
      bareType: "string",
      staticType: "{ url: string }",
      value: "{ url: nextURL }",
    };
  } else {
    throw new Error("Unsupported pagination type: " + pType);
  }
}
registerTemplateFunc("templateNextPageState", templateNextPageState);

/**
 * This utility function supports templateNextPageState. It is not intended to
 * be used directly in other parts of the target.
 */
function getCursorType(op: Operation): string {
  const pagination = op.Extensions.Pagination;
  const isPaginated = pagination && op.Response?.Responses.length > 0;
  if (!isPaginated) {
    throw new Error(
      `${op.ID}: getCursorType should only be called for paginated operations`,
    );
  }

  const cursorInput = getPaginationInput(pagination, "cursor");
  if (!cursorInput) {
    throw new Error(
      `${op.ID}: getCursorType: Cursor input not found in pagination config.`,
    );
  }

  let found: FieldDef | null = null;
  if (cursorInput.In.toString() === "requestBody") {
    found = getCursorFromBody(op, cursorInput);
  } else {
    found = getCursorFromParams(op, cursorInput);
  }

  if (!found) {
    return "any";
  }

  if (!found.Type.IsPrimitive()) {
    console.log(
      `${op.ID}: getCursorType: falling back to any due to non-primitive cursor type.`,
    );
    return "any";
  }

  const cursorTypes = resolveOutbound({
    rootTypeDef: op.OwningSDK.Type,
    typeDef: found.Type,
    nullable: false,
    optional: false,
    usageLocation: "",
    streamable: false,
  });

  return cursorTypes.inputType;
}

function getCursorFromBody(
  op: Operation,
  cursorInput: PaginationInputs,
): FieldDef | null {
  if (!op.Request?.Field) {
    console.log(`${op.ID}: getCursorFromBody: Request body not defined.`);
    return null;
  }

  let bodyField: FieldDef | null = null;
  if (op.Request?.Field.Annotations?.Has("requestWrapper")) {
    bodyField = op.Request.Field.Type?.Fields?.find(
      (f) => f.Annotations?.Has("request"),
    );
  } else if (op.Request?.Field.Annotations?.Has("request")) {
    bodyField = op.Request.Field;
  }

  if (!bodyField) {
    console.log(
      `${op.ID}: getCursorFromBody: Could not found request body field.`,
    );
    return null;
  }

  const result = bodyField.Type?.FindFieldByName(cursorInput.Name);
  if (!result) {
    console.log(
      `${op.ID}: getCursorFromBody: Could not found cursor field in request body.`,
    );
    return null;
  }

  return result;
}

function getCursorFromParams(
  op: Operation,
  cursorInput: PaginationInputs,
): FieldDef | null {
  const allDefs = [
    op.Request?.Params?.QueryParams,
    op.Request?.Params?.PathParams,
    op.Request?.Params?.HeaderParams,
  ];

  for (const defs of allDefs) {
    if (defs == null) {
      continue;
    }

    for (const param of defs) {
      if (param.Field.Name === cursorInput.Name) {
        return param.Field;
      }
    }
  }

  console.log(
    `${op.ID}: getCursorFromParams: Cursor input not found in path/query/header parameters.`,
  );

  return null;
}

// @ts-ignore
function templateExampleReferenceValue(
  exampleReferenceValue: ExampleReferenceValue,
  targetField?: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  const isAnyPartOfFieldOptional =
    exampleReferenceValue.parents.some((p) => {
      return p.Optional;
    }) || exampleReferenceValue.source.Optional;

  if (
    additionalContext?.isTest &&
    isAnyPartOfFieldOptional &&
    !targetField.Optional
  ) {
    addTestImport("./assertions.js", "assertDefined");
    return `assertDefined(${exampleReferenceValue.path})`;
  }

  return exampleReferenceValue.path;
}

function templateSDKQualifier(usageContext: UsageContext): string {
  if (usageContext.ContextIndex > 0) {
    return "";
  }

  if (!usageContext.Test) {
    return "const ";
  }

  const count = usageContext.Test.Steps.filter(
    (s) => s.Type == "operation" && !s.UsageContext.SkipSDKInstantiation,
  ).length;

  return count == 1 ? "const " : "let ";
}
registerTemplateFunc("templateSDKQualifier", templateSDKQualifier);

function templateJSONPathExpression(
  documentVar: string,
  queryExpr: string,
): string {
  const lib = context.Global.Config.Jsonpath;
  const expr = sanitizeJsonPath(queryExpr);

  if (lib === "rfc9535") {
    addImport("jsonpath-rfc9535", "query as jpQuery");
    addImport("jsonpath-rfc9535", "JsonValue");

    return `jpQuery(${documentVar} as JsonValue, "${expr}")[0]`;
  } else {
    addImport("jsonpath", "jp", "packageImport");

    return `jp.value(${documentVar}, "${expr}")`;
  }
}
registerTemplateFunc("templateJSONPathExpression", templateJSONPathExpression);

// Renders an inline, null-safe accessor for a generation-time dot path over
// the raw (wire-shape) response JSON, replacing the old runtime dlv() helper:
// ("responseData", "meta.paging.next_cursor", op) =>
//   `(responseData as { meta?: { paging: { next_cursor?: unknown } | null } })
//     .meta?.paging?.next_cursor`
// When Zod validation is active the response has already passed inbound
// schema validation by the time pagination outputs are read, so segments the
// schema marks required and non-nullable are accessed with plain `.`;
// optional or nullable segments use `?.`. In no-zod mode — or when the path
// cannot be statically resolved against a single JSON success response type —
// nothing is guaranteed about the body, so every level (including the root)
// is treated as potentially absent.
function templateDotPathExpression(
  documentVar: string,
  dotPath: string,
  op?: Operation,
): string {
  const segments = dotPath ? dotPath.split(".") : [];
  const fields = isNoZod() ? null : resolveResponseDotPathFields(op, segments);

  let castType = "unknown";
  for (let i = segments.length - 1; i >= 0; i--) {
    const key = typescriptIdentifierRegex.test(segments[i])
      ? segments[i]
      : JSON.stringify(segments[i]);
    const optional = fields ? fields[i].Optional : true;
    const nullable = fields ? fields[i].Nullable : false;
    const value = nullable ? `${castType} | null` : castType;
    castType = `{ ${key}${optional ? "?" : ""}: ${value} }`;
  }
  if (!fields) {
    castType = `${castType} | null | undefined`;
  }

  let expr = `(${documentVar} as ${castType})`;
  let nullish = !fields; // an unvalidated body may itself be nullish
  for (let i = 0; i < segments.length; i++) {
    expr += dotPathAccessor(segments[i], nullish);
    nullish = fields ? fields[i].Optional || fields[i].Nullable : true;
  }
  return expr;
}
registerTemplateFunc("templateDotPathExpression", templateDotPathExpression);

function dotPathAccessor(segment: string, nullish: boolean): string {
  if (typescriptIdentifierRegex.test(segment)) {
    return `${nullish ? "?." : "."}${segment}`;
  }
  return `${nullish ? "?." : ""}[${JSON.stringify(segment)}]`;
}

// Walks the operation's JSON success response type along the dot path,
// returning one FieldDef per segment, or null when any segment cannot be
// resolved to a plain object property (unions, maps, multiple success
// responses, unknown fields).
function resolveResponseDotPathFields(
  op: Operation | undefined,
  segments: string[],
): FieldDef[] | null {
  if (!op || !segments.length) {
    return null;
  }

  let typeDef = paginationSuccessResponseType(op);
  const fields: FieldDef[] = [];
  for (const segment of segments) {
    if (!typeDef || typeDef.Type.toString() !== "class") {
      return null;
    }
    const field = (typeDef.Fields || []).find(
      (f) => !f.IsAdditionalProperties && originalFieldName(f) === segment,
    );
    if (!field?.Type) {
      return null;
    }
    fields.push(field);
    typeDef = field.Type;
  }
  return fields;
}

// The single non-error JSON response body type pagination outputs are read
// from; undefined when there is not exactly one candidate.
function paginationSuccessResponseType(op: Operation): TypeDef | undefined {
  let found: TypeDef | undefined;
  for (const sub of op.Response?.Responses || []) {
    if (sub.Error) {
      continue;
    }
    for (const content of sub.Content || []) {
      const typeDef = content.Content?.Type;
      if (!typeDef) {
        continue;
      }
      if (found && found !== typeDef) {
        return undefined;
      }
      found = typeDef;
    }
  }
  return found;
}

function hasSpeakeasyInclude(type: TypeDef): boolean {
  return type.Extensions?.All?.["x-speakeasy-include"] === true;
}

function shouldIncludeInboundAndOutboundSchema(type: TypeDef): boolean {
  return (
    type.UsedInWebhook ||
    type.UsedInCallback ||
    hasSpeakeasyInclude(type) ||
    isMCPServerEnabled() ||
    context.Global.Config.AlwaysIncludeInboundAndOutbound
  );
}

function shouldIncludeInboundSchema(type: TypeDef): boolean {
  return type.UsedInResponse || shouldIncludeInboundAndOutboundSchema(type);
}
registerTemplateFunc("shouldIncludeInboundSchema", shouldIncludeInboundSchema);

function shouldIncludeOutboundSchema(type: TypeDef): boolean {
  return (
    type.UsedInRequest ||
    type.UsedInSecurity ||
    shouldIncludeInboundAndOutboundSchema(type)
  );
}
registerTemplateFunc(
  "shouldIncludeOutboundSchema",
  shouldIncludeOutboundSchema,
);

function shouldIncludeInboundOrOutboundSchema(type: TypeDef): boolean {
  return shouldIncludeInboundSchema(type) || shouldIncludeOutboundSchema(type);
}
registerTemplateFunc(
  "shouldIncludeInboundOrOutboundSchema",
  shouldIncludeInboundOrOutboundSchema,
);

function shouldIncludeSchemaNamespace(type: TypeDef): boolean {
  return (
    context.Global.Config.ExportZodModelNamespace &&
    (shouldIncludeInboundSchema(type) || shouldIncludeOutboundSchema(type))
  );
}
registerTemplateFunc(
  "shouldIncludeSchemaNamespace",
  shouldIncludeSchemaNamespace,
);

// @ts-ignore
function templateAuxiliaryFiles(
  data: unknown,
  sourcePath: string = "auxiliary",
) {
  const files = getDirectoryFiles(sourcePath);

  for (const file of files) {
    if (file.endsWith(".oxlintrc.json") && !context.Global.Config.UseOxlint) {
      continue;
    }
    if (file.endsWith("eslint.config.mjs") && context.Global.Config.UseOxlint) {
      continue;
    }
    // In no-zod mode, zod-backed enum/coercion/unrecognized helpers are not
    // referenced. Open enums inline their primitive fallback type.
    if (
      isNoZod() &&
      (file.endsWith("/enums.ts.stmpl") ||
        file.endsWith(`/${getUnrecognizedFileName()}.ts.stmpl`) ||
        file.endsWith("/{{getUnrecognizedFileName}}.ts.stmpl") ||
        file.endsWith(`/${getDefaultToZeroValueFileName()}.ts.stmpl`) ||
        file.endsWith("/{{getDefaultToZeroValueFileName}}.ts.stmpl") ||
        file.endsWith("/primitives.ts.stmpl"))
    ) {
      continue;
    }

    let outFile = file.replace(`${sourcePath}/`, "");

    outFile = templateStringInput("auxPath:" + outFile, outFile, data);
    if (outFile.trim().length == 0) {
      continue;
    }

    if (file.endsWith(".stmpl")) {
      outFile = outFile.replace(".stmpl", "");
      templateFile(file, outFile, context.Global);
    } else {
      copy(file, outFile);
    }
  }
}
