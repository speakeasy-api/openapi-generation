// @ts-ignore
function getScopePath(scope: string): string {
  const imports = context.Global.Config.Imports;

  switch (scope) {
    case "errors":
      return imports.GetErrorsPath();
    case "shared":
      return imports.GetSharedPath();
    case "operations":
      return imports.GetOperationsPath();
    case "webhooks":
      return imports.GetWebhooksPath();
    case "callbacks":
      return imports.GetCallbacksPath();
    case "utils":
      return "Utils";
    case "models":
      return "Models";
    case "hooks":
      return "Hooks";
    default:
      throw new Error(`Unexpected scope: ${scope}`);
  }
}

// @ts-ignore
function getBuiltInNamespaces(): string[] {
  const paths = new Set<string>([getScopePath("utils")]);

  if (accountHasFeatureAccess("sdkHooks")) {
    paths.add(getScopePath("hooks"));
  }
  if (hasClientCredentials()) {
    paths.add("ClientCredentials");
  }

  if (usingGlobalImports()) {
    paths.add(getScopePath("models"));
  } else {
    for (const scope in context.Global.Config.Imports.Paths) {
      paths.add(getScopePath(scope));
    }
  }

  return [...paths];
}

// @ts-ignore
function getModelNamespaces(): string[] {
  const namespaces = new Set<string>();

  for (const [outputLocation] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    const ns = getModelNamespace(outputLocation, true);

    if (ns !== "") {
      namespaces.add(ns);
    }
  }
  return [...namespaces];
}

registerTemplateFunc("getModelNamespaces", getModelNamespaces);

// @ts-ignore
function isCustomNamespace(outputLocation: string): boolean {
  if (!outputLocation) {
    return false;
  }

  return !getBuiltInNamespaces().some(
    (p) => outputLocation === p || outputLocation.startsWith(p + "/"),
  );
}

// @ts-ignore
function getAccessNamespace(
  outputLocation: string,
  full: boolean = false,
  utils: boolean = false,
): string {
  if (utils) {
    return `${getSDKNamespace()}.${outputLocation.replace("/", ".")}`;
  }

  if (usingGlobalImports() && !isCustomNamespace(outputLocation)) {
    return full ? getSDKNamespace() : "";
  }

  if (outputLocation == "") {
    return getSDKNamespace();
  }

  const parts = outputLocation.split("/");
  if (full) {
    parts.unshift(getSDKNamespace());
  }

  return parts.join(".");
}

registerTemplateFunc("getAccessNamespace", getAccessNamespace);

// @ts-ignore
function getScopeNamespace(scope: string, full: boolean = false): string {
  return getAccessNamespace(getScopePath(scope), full, scope === "utils");
}

registerTemplateFunc("getScopeNamespace", getScopeNamespace);

// @ts-ignore
function getModelsLocation(outputLocation: string): string {
  return sanitizeOutputLocation(outputLocation) || getScopePath("models");
}

unregisterTemplateFunc("getModelsLocation");
registerTemplateFunc("getModelsLocation", getModelsLocation);

// @ts-ignore
function getModelNamespace(
  outputLocation: string,
  full: boolean = true,
): string {
  return getAccessNamespace(getModelsLocation(outputLocation), full);
}

registerTemplateFunc("getModelNamespace", getModelNamespace);

// @ts-ignore
function getPackageId(): string {
  return sanitizePackageId(context.Global.Config.PackageName);
}

registerTemplateFunc("getPackageId", getPackageId);

// @ts-ignore
function getSDKNamespace(): string {
  return getPackageId();
}

registerTemplateFunc("getSDKNamespace", getSDKNamespace);

// @ts-ignore
function getSDKTopLevelFolder(): string {
  const srcDir = context.Global.Config.SourceDirectory;
  const projectPath = getSDKNamespace().replaceAll(".", "/");
  return srcDir ? `${srcDir}/${projectPath}` : projectPath;
}

registerTemplateFunc("getSDKTopLevelFolder", getSDKTopLevelFolder);

// @ts-ignore
function getProjectFilePath(): string {
  const winPath = getSDKTopLevelFolder().replaceAll("/", "\\");
  return `${winPath}\\${getPackageId()}.csproj`;
}

registerTemplateFunc("getProjectFilePath", getProjectFilePath);

// @ts-ignore
function getUsageSourceFilePath(sourceDir: string): string {
  return `${sourceDir.replaceAll("/", "\\")}\\${getProjectFilePath()}`;
}

registerTemplateFunc("getUsageSourceFilePath", getUsageSourceFilePath);

function getDefaultHttpClient(): string {
  const prefix = sanitizeClassName(
    context.Global.Config.HTTPClientPrefix ||
      context.Global.AST.MainSDK.Type.Name,
  );
  return `${prefix}HttpClient`;
}

registerTemplateFunc("getDefaultHttpClient", getDefaultHttpClient);
