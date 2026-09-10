// TODO:
// - TODO: Move all this to common if it largely matches the go template imports

// @ts-ignore
function genImports(): string {
  return `{{ templateImports .RecursiveComputed.Imports }}`;
}
registerTemplateFunc("genImports", genImports);

// @ts-ignore
function addImport(
  imp: string,
  local: boolean = false,
  alias: string = "",
): string {
  if (!local) {
    switch (imp) {
      case "operations":
        imp = `internal/sdk/models/operations`;
        local = true;
        break;
      case "shared":
        imp = `internal/sdk/models/components`;
        local = true;
        break;
      case "types":
        imp = `internal/sdk/types`;
        local = true;
        break;
      case "sdkerrors":
        imp = `internal/sdk/models/sdkerrors`;
        local = true;
        break;
      case "utils":
        imp = `internal/sdk/sdkinternal/utils`;
        local = true;
        break;
      case "retry":
        imp = `internal/sdk/${getRetryLocation()}`;
        local = true;
        break;
    }
  }

  if (local) {
    if (!imp) {
      return "";
    }

    // SDK model imports (e.g. "models/components") need the "internal/sdk/" prefix.
    // Skip paths that already have a full prefix like "internal/sdk/" or "internal/cli/".
    const isSDKModelImport =
      imp.includes("models") &&
      !imp.startsWith("internal/sdk/") &&
      !imp.startsWith("internal/cli/") &&
      !imp.startsWith("internal/");
    if (isSDKModelImport) {
      imp = `internal/sdk/${imp}`;
    }

    imp = sanitizeOutputLocation(imp);
    imp = `${getGolangPackage()}/${imp}`;
  }
  imp = `"${imp}"`;
  if (alias) {
    imp = `${alias} ${imp}`;
  }

  let imports = context.RecursiveComputed?.Imports;
  if (!imports) {
    imports = [];
  }

  if (
    !imports.find((i) => {
      return i === imp;
    })
  ) {
    imports.push(imp);
  }

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }

  context.RecursiveComputed.Imports = imports;

  return ""; // Still needs to return a string if used in templates
}
registerTemplateFunc("addImport", addImport);

// @ts-ignore
function templateImports(imports: string[]): string {
  if (!imports || imports.length == 0) {
    return "";
  }

  return `import(\n\t${imports.join("\n\t")}\n)`;
}
registerTemplateFunc("templateImports", templateImports);

// @ts-ignore
function addSDKPackageImport(): string {
  const packageName = getGolangPackage();

  const parts = packageName.split("/");
  if (parts.length > 0 && parts[parts.length - 1] != sanitizeSDKPackageName()) {
    addImport(packageName, false, sanitizeSDKPackageName());
  } else {
    addImport(packageName);
  }

  return ""; // Still needs to return a string if used in templates
}
registerTemplateFunc("addSDKPackageImport", addSDKPackageImport);

// @ts-ignore
function sanitizeModelPackageName(
  outputLocation: string,
  useAlias: boolean,
): string {
  if (!outputLocation) {
    return sanitizeSDKPackageName(useAlias);
  }

  outputLocation = sanitizeOutputLocation(outputLocation);
  let parts = outputLocation.split("/");

  return parts[parts.length - 1];
}
registerTemplateFunc("sanitizeModelPackageName", sanitizeModelPackageName);

// @ts-ignore
function getAccessNamespace(
  currentLocation: string,
  scope: string,
  outputLocation?: string,
): string {
  let scopePath = "";

  switch (scope.toString()) {
    case "errors":
      scopePath = getErrorsLocation();
      break;
    case "shared":
      scopePath = getSharedLocation();
      break;
    case "operations":
      scopePath = getOperationsLocation();
      break;
    case "polling":
      scopePath = getPollingLocation();
      break;
    default:
      throw new Error(`Unexpected scope: '${scope}'`);
  }

  // Check if this is a namespace package (outputLocation is just a simple directory name
  // like "foo" or "bar", not a path like "models/components")
  // Namespace packages are created via x-speakeasy-model-namespace extension
  if (
    outputLocation &&
    outputLocation !== scopePath &&
    !outputLocation.includes("/")
  ) {
    scopePath = outputLocation;
  }

  if (currentLocation == scopePath) {
    return "";
  }

  return sanitizeModelPackageName(scopePath, true) + ".";
}
registerTemplateFunc("getAccessNamespace", getAccessNamespace);

//@ts-ignore
function getErrorsLocation() {
  return context.Global.Config.Imports.GetErrorsPath();
}
registerTemplateFunc("getErrorsLocation", getErrorsLocation);

//@ts-ignore
function getOperationsLocation() {
  return context.Global.Config.Imports.GetOperationsPath();
}
registerTemplateFunc("getOperationsLocation", getOperationsLocation);

//@ts-ignore
function getSharedLocation() {
  return context.Global.Config.Imports.GetSharedPath();
}
registerTemplateFunc("getSharedLocation", getSharedLocation);

//@ts-ignore
function getTypesLocation() {
  return "types";
}
registerTemplateFunc("getTypesLocation", getTypesLocation);

//@ts-ignore
function getUtilsLocation() {
  return "internal/utils";
}
registerTemplateFunc("getUtilsLocation", getUtilsLocation);

//@ts-ignore
function getPollingLocation() {
  return "polling";
}

registerTemplateFunc("getPollingLocation", getPollingLocation);

//@ts-ignore
function getRetryLocation() {
  return "retry";
}
registerTemplateFunc("getRetryLocation", getRetryLocation);

//@ts-ignore
function getNullableLocation() {
  return "optionalnullable";
}
registerTemplateFunc("getNullableLocation", getNullableLocation);
