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
    case "hooks":
      return "Hooks";
    case "retry":
      return "Utils\\Retry";
    default:
      throw new Error(`Unexpected scope: ${scope}`);
  }
}
registerTemplateFunc("getScopePath", getScopePath);

// @ts-ignore
function getSDKNamespace(): string {
  return context.Global.Config.Namespace;
}

registerTemplateFunc("getSDKNamespace", getSDKNamespace);
// @ts-ignore
function getAccessNamespace(
  outputLocation: string,
  full: boolean = false,
  utils: boolean = false,
): string {
  if (utils) {
    return `${getSDKNamespace()}\\${outputLocation}`;
  }

  const parts = outputLocation.split("/");

  if (full) {
    parts.unshift(getSDKNamespace());
  }

  // remove empty parts
  return parts.filter((p) => p.length > 0).join("\\");
}

registerTemplateFunc("getAccessNamespace", getAccessNamespace);

// @ts-ignore
function getScopeNamespace(scope: string, full: boolean = false): string {
  return getAccessNamespace(getScopePath(scope), full, scope === "utils");
}

registerTemplateFunc("getScopeNamespace", getScopeNamespace);

// @ts-ignore
function getModelNamespace(
  outputLocation: string,
  full = true,
  className = "",
): string {
  const parts = [];

  if (full) {
    parts.push(context.Global.Config.Namespace);
  }
  if (outputLocation) {
    outputLocation = sanitizeOutputLocation(outputLocation);
    parts.push(outputLocation.replaceAll("/", "\\"));
  }
  if (className) {
    parts.push(className);
  }

  return parts.join("\\");
}

registerTemplateFunc("getModelNamespace", getModelNamespace);

// @ts-ignore
function getSDKTopLevelFolder(): string {
  return "src";
}

registerTemplateFunc("getSDKTopLevelFolder", getSDKTopLevelFolder);
