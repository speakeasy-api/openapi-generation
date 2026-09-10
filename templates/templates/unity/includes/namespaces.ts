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
    case "utils":
      return "Utils";
    case "models":
      return "Models";
    default:
      throw new Error(`Unexpected scope: ${scope}`);
  }
}

// @ts-ignore
function getAccessNamespace(
  outputLocation: string,
  full: boolean = false,
): string {
  const standardSdkPaths = [
    "Models",
    "Operations",
    "Errors",
    "Webhooks",
    "Callbacks",
    "Hooks",
    "Utils",
    "Types",
    "ClientCredentials",
  ];
  const isCustomNamespace =
    outputLocation !== "" &&
    !standardSdkPaths.some(
      (scope) =>
        outputLocation === scope || outputLocation.startsWith(scope + "/"),
    );

  if (usingGlobalImports() && !isCustomNamespace) {
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
  return getAccessNamespace(getScopePath(scope), full);
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
  return getSDKNamespace().replaceAll(".", "/");
}

registerTemplateFunc("getSDKTopLevelFolder", getSDKTopLevelFolder);

// @ts-ignore
function getProjectFilePath(): string {
  const winPath = getSDKNamespace().replaceAll(".", "\\");
  return `${winPath}\\${getPackageId()}.csproj`;
}

registerTemplateFunc("getProjectFilePath", getProjectFilePath);

// @ts-ignore
function getRootLocationRelativeToPackage(): string {
  let path = getProjectFilePath();
  path = path.substring(0, path.lastIndexOf("\\"));

  const parts = path.split("\\");

  return "..\\".repeat(parts.length);
}
registerTemplateFunc(
  "getRootLocationRelativeToPackage",
  getRootLocationRelativeToPackage,
);

// @ts-ignore
function getUsageSourceFilePath(sourceDir: string): string {
  return `${sourceDir.replaceAll("/", "\\")}\\${getProjectFilePath()}`;
}

registerTemplateFunc("getUsageSourceFilePath", getUsageSourceFilePath);
