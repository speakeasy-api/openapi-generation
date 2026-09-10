// @ts-ignore
function addImport(
  imp: string,
  local: boolean = false,
  alias: string = "",
): string {
  if (!local) {
    switch (imp) {
      case "utils":
        imp = getUtilsLocation();
        local = true;
        break;
      case "operations":
        imp = getOperationsLocation();
        local = true;
        break;
      case "shared":
        imp = getSharedLocation();
        local = true;
        break;
      case "types":
        imp = getTypesLocation();
        local = true;
        break;
      case "optionalnullable":
        imp = getNullableLocation();
        local = true;
        break;
      case "stream":
        imp = getTypesLocation() + "/stream";
        local = true;
        break;
      case "jsonl":
        imp = getTypesLocation() + "/jsonl";
        local = true;
        break;
    }
  }

  if (local) {
    if (!imp) {
      return "";
    }

    switch (imp) {
      case "globals":
        imp = "globals";
        break;
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
function genImports(): string {
  return `{{ templateImports .RecursiveComputed.Imports }}`;
}

registerTemplateFunc("genImports", genImports);

// @ts-ignore
function getAccessNamespace(
  currentLocation: string,
  scope: string,
  outputLocation?: string,
): string {
  let scopePath = "";

  switch (scope.toString()) {
    case "operations":
      scopePath = getOperationsLocation();
      break;
    case "shared":
      scopePath = getSharedLocation();
      break;
    default:
      throw new Error(`Unexpected scope: '${scope}'`);
  }

  // Check if this is a namespace package.
  // Namespace packages are created via x-speakeasy-model-namespace extension.
  // They can appear as simple directory names (e.g., "foo") or with a models prefix
  // (e.g., "models/foo" in mockserver context).
  // We detect namespace packages by checking if the outputLocation differs from
  // the standard scope path and ends with a different package name.
  if (outputLocation && outputLocation !== scopePath) {
    const outputParts = outputLocation.split("/");
    const scopeParts = scopePath.split("/");
    const outputLastPart = outputParts[outputParts.length - 1];
    const scopeLastPart = scopeParts[scopeParts.length - 1];

    // If the final package names differ, this is a namespace package
    if (outputLastPart !== scopeLastPart) {
      scopePath = outputLocation;
    }
  }

  if (currentLocation == scopePath) {
    return "";
  }

  return sanitizeModelPackageName(scopePath) + ".";
}
registerTemplateFunc("getAccessNamespace", getAccessNamespace);

//@ts-ignore
function getOperationsLocation() {
  return "models/operations";
}

//@ts-ignore
function getSharedLocation() {
  return "models/components";
}
registerTemplateFunc("getSharedLocation", getSharedLocation);

//@ts-ignore
function getTypesLocation() {
  return "types";
}
registerTemplateFunc("getTypesLocation", getTypesLocation);

//@ts-ignore
function getNullableLocation() {
  return "optionalnullable";
}
registerTemplateFunc("getNullableLocation", getNullableLocation);

function getUtilsLocation() {
  return "utils";
}

registerTemplateFunc("getUtilsLocation", getUtilsLocation);

function sanitizeModelPackageName(outputLocation: string): string {
  if (!outputLocation) {
    return sanitizeSDKPackageName();
  }

  outputLocation = sanitizeOutputLocation(outputLocation);
  let parts = outputLocation.split("/");

  return parts[parts.length - 1];
}

registerTemplateFunc("sanitizeModelPackageName", sanitizeModelPackageName);

// @ts-ignore
function templateImports(imports: string[]): string {
  if (!imports || imports.length == 0) {
    return "";
  }

  return `import(\n\t${imports.join("\n\t")}\n)`;
}

registerTemplateFunc("templateImports", templateImports);
