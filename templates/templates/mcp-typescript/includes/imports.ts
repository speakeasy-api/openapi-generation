const aliasImport = "aliasImport";
const packageImport = "packageImport";
const typeImport = "typeImport";
const namedAliasImport = "namedAliasImport";

type TSImportType =
  | typeof aliasImport
  | typeof packageImport
  | typeof typeImport
  | typeof namedAliasImport;

type TSImports = {
  [key: string]: { import: string; type: string; alias?: string }[];
};

function addImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
  alias?: string,
): string {
  let imports: TSImports = context.RecursiveComputed?.Imports;

  if (!imports) {
    imports = {};
  }

  if (!imports[pkg]) {
    imports[pkg] = [];
  }

  if (!imports[pkg].find((i) => i.import == imp && i.type == type)) {
    imports[pkg].push({
      import: imp,
      type,
      ...(alias ? { alias } : {}),
    });
  }

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }

  context.RecursiveComputed.Imports = imports;

  return "";
}

registerTemplateFunc("addImport", addImport);

function addTestImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
): string {
  let imports: TSImports = context.RecursiveComputed?.TestImports;
  if (!imports) {
    imports = {};
  }

  if (!imports[pkg]) {
    imports[pkg] = [];
  }

  if (!imports[pkg].find((i) => i.import == imp && i.type == type)) {
    imports[pkg].push({
      import: imp,
      type,
    });
  }

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }

  context.RecursiveComputed.TestImports = imports;

  return "";
}
registerTemplateFunc("addTestImport", addTestImport);

function addUsageImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
): string {
  let imports: TSImports = context.RecursiveComputed?.UsageImports;
  if (!imports) {
    imports = {};
  }

  if (!imports[pkg]) {
    imports[pkg] = [];
  }

  if (!imports[pkg].find((i) => i.import == imp && i.type == type)) {
    imports[pkg].push({
      import: imp,
      type,
    });
  }

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }

  context.RecursiveComputed.UsageImports = imports;

  return "";
}
registerTemplateFunc("addUsageImport", addUsageImport);

function addUsageImportForType(typeDef: TypeDef): string {
  return addTypeDefImport(typeDef, "usage");
}
registerTemplateFunc("addUsageImportForType", addUsageImportForType);

function addUsageTypeImport(pkg: string, type: string): string {
  addTestImport(`../types/${pkg}.js`, type);
  addImport(pkg, type);
  return "";
}
registerTemplateFunc("addUsageTypeImport", addUsageTypeImport);

function genImports(importsKey?: string): string {
  const importType = importsKey || "Imports";
  return `{{ templateImports .RecursiveComputed.${importType} }}`;
}
registerTemplateFunc("genImports", genImports);

function addEnumTypeImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): string {
  const resolved = resolveTypeDefImport(def, usageLocation, modelName);
  if (!resolved) {
    return "";
  }

  return addImport(resolved.path, resolved.name, resolved.type, resolved.alias);
}

function addErrorTypeImport(type: BuiltinError, relativeTo = ""): string {
  const outputLocation = getModelsLocation(getErrorsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);

  switch (type) {
    case getDefaultErrorClassName():
    case getBaseErrorClassName():
    case "SDKValidationError":
    case "ResponseValidationError":
      return addImport(`${prefix}/${type.toLowerCase()}.js`, type, typeImport);
    case "UnexpectedClientError":
    case "InvalidRequestError":
    case "RequestAbortedError":
    case "RequestTimeoutError":
    case "ConnectionError":
      return addImport(`${prefix}/httpclienterrors.js`, type, typeImport);
    default:
      throw new Error(`Unknown error type: ${type}`);
  }
}
registerTemplateFunc("addErrorTypeImport", addErrorTypeImport);

function addFuncImport(op: Operation, usageLocation: string) {
  const funcName = sanitizeFuncName(op);
  const funcFilename = sanitizeFuncFilename(op);
  const prefix = getImportPrefix("funcs", usageLocation);
  return addImport(`${prefix}/${funcFilename}.js`, funcName, typeImport);
}

registerTemplateFunc("addFuncImport", addFuncImport);

function addInternalImport(
  pkg: string,
  imp: string,
  usageLocation: string,
  type: TSImportType = typeImport,
) {
  const prefix =
    usageLocation === "lib" ? "." : getImportPrefix("lib", usageLocation);
  return addImport(`${prefix}/${pkg}.js`, imp, type);
}

registerTemplateFunc("addInternalImport", addInternalImport);

function addLibImport(globalSecurity: TypeDef | undefined): string {
  const isGlobalSecurityFlattened =
    globalSecurity?.Fields.length === 1 &&
    context.Global.Config.FlattenGlobalSecurity;

  if (globalSecurity != null && !isGlobalSecurityFlattened) {
    return addTypeDefImport(globalSecurity, "lib");
  }

  return "";
}

registerTemplateFunc("addLibImport", addLibImport);

function addTypeImport(
  pkg: string,
  symbol: string,
  usageLocation: string,
  type: TSImportType = typeImport,
): string {
  return addImport(
    `${getImportPrefix("types", usageLocation)}/${pkg}.js`,
    symbol,
    type,
  );
}

registerTemplateFunc("addTypeImport", addTypeImport);

function addTypeDefImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): string {
  const resolved = resolveTypeDefImport(def, usageLocation, modelName);
  if (!resolved) {
    return "";
  }

  return addImport(resolved.path, resolved.name, resolved.type, resolved.alias);
}

function addZodImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): string {
  const originalName = sanitizeClassName(def.Name) + "$zodSchema";

  if (modelName !== "" && inSameModel(modelName, usageLocation, def)) {
    return "";
  }

  const { outputLocation, mfilename, extpath } = resolveModelImport({
    def,
    usageLocation,
    name: originalName,
  });

  const nsPrefix = getNamespacePrefix(def.OutputLocation, usageLocation);
  const alias = nsPrefix ? `${nsPrefix}_${originalName}` : undefined;
  const importType = alias ? namedAliasImport : typeImport;

  if (outputLocation == usageLocation) {
    return addImport(`./${mfilename}.js`, originalName, importType, alias);
  }

  return addImport(extpath, originalName, importType, alias);
}

function getErrorsLocation() {
  return context.Global.Config.Imports.GetErrorsPath();
}

function getOperationsLocation() {
  return context.Global.Config.Imports.GetOperationsPath();
}

function getSharedLocation(): string {
  return context.Global.Config.Imports.GetSharedPath();
}

function getWebhooksLocation(): string {
  return context.Global.Config.Imports.GetWebhooksPath();
}

function getCallbacksLocation(): string {
  return context.Global.Config.Imports.GetCallbacksPath();
}

// Returns true if the outputLocation represents a model namespace (created via
// x-speakeasy-model-namespace) rather than a standard scope path like "models/errors".
function isNamespaceOutputLocation(outputLocation: string): boolean {
  if (!outputLocation) {
    return false;
  }
  const normalized = sanitizeOutputLocation(outputLocation);
  const standardPaths = [
    getErrorsLocation(),
    getSharedLocation(),
    getOperationsLocation(),
    getWebhooksLocation(),
    getCallbacksLocation(),
    "models",
  ];
  return !standardPaths.includes(normalized);
}

// Returns a PascalCase namespace prefix for a type that lives in a custom
// namespace (via x-speakeasy-model-namespace) when used from a different
// location.  Returns null when no aliasing is needed.
function getNamespacePrefix(
  outputLocation: string,
  usageLocation: string,
): string | null {
  if (!isNamespaceOutputLocation(outputLocation)) {
    return null;
  }
  const outputParts = (outputLocation || "").split("/");
  const namespace = outputParts[outputParts.length - 1].toLowerCase();
  const usageParts = (usageLocation || "").split("/");
  const usageNamespace = usageParts[usageParts.length - 1].toLowerCase();
  if (usageNamespace === namespace) {
    return null;
  }
  return caser().ToPascal(namespace);
}

function getImportPrefix(target: string, relativeTo: string): string {
  if (relativeTo === "") {
    relativeTo = "sdk";
  }

  const targetParts = target.split("/");
  const relParts = relativeTo.split("/");

  // Track where the two paths diverge
  let i = 0;
  const limit = Math.min(relParts.length, targetParts.length);
  for (; i < limit; i++) {
    if (relParts[i] !== targetParts[i]) {
      break;
    }
  }

  let prefix = "../".repeat(relParts.length - i) || "./";
  prefix = prefix + targetParts.slice(i).join("/");
  // Remove trailing slash - if needed
  return prefix.replace(/\/$/, "");
}

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

function resolveModelImport({
  def,
  usageLocation,
  name,
}: {
  def: TypeDef;
  usageLocation: string;
  name: string;
}): {
  outputLocation: string;
  mfilename: string;
  extpath: string;
  extname: string;
  exttype: TSImportType;
  qualified: boolean;
} {
  const mfilename = resolveModelName(def);
  const outputLocation = getModelsLocation(def.OutputLocation);
  const relativePath = getImportPrefix(outputLocation, usageLocation);

  const qualified = false;
  const extpath = `${relativePath}/${mfilename}.js`;
  const extname = name;
  const exttype = typeImport;
  return { outputLocation, mfilename, extpath, extname, exttype, qualified };
}

function resolveTypeDefImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): { path: string; name: string; type: TSImportType; alias?: string } | null {
  if (modelName !== "" && inSameModel(modelName, usageLocation, def)) {
    return null;
  }

  const mfilename = resolveModelName(def);
  const outputLocation = getModelsLocation(def.OutputLocation);
  const relativePath = `${getImportPrefix(outputLocation, usageLocation)}`;
  const className = sanitizeClassName(def.Name);

  const nsPrefix = getNamespacePrefix(def.OutputLocation, usageLocation);
  const alias = nsPrefix ? `${nsPrefix}_${className}` : undefined;
  const importType = alias ? namedAliasImport : typeImport;

  const extpath = `${relativePath}/${mfilename}.js`;

  return outputLocation == usageLocation
    ? { path: `./${mfilename}.js`, name: className, type: importType, alias }
    : { path: extpath, name: className, type: importType, alias };
}

function formatBracedImport(specifiers: string[], pkg: string): string[] {
  const joined = specifiers.join(", ");
  if (joined.length <= 80) {
    return [`import { ${joined} } from "${pkg}";`];
  }
  return [`import {`, ...specifiers.map((s) => `  ${s},`), `} from "${pkg}";`];
}

function formatBareImport(specifiers: string[], pkg: string): string[] {
  const joined = specifiers.join(", ");
  if (joined.length <= 80) {
    return [`import ${joined} from "${pkg}";`];
  }
  return [`import {`, ...specifiers.map((s) => `  ${s},`), `} from "${pkg}";`];
}

function templateImports(imports: TSImports): string {
  const lines: string[] = [];

  for (const pkg of Object.keys(imports ?? {}).sort()) {
    const sorted = imports[pkg].sort((a, b) =>
      a.import.localeCompare(b.import),
    );

    const bareImports: string[] = [];
    const namedImports: string[] = [];

    for (const imp of sorted) {
      if (imp.type === aliasImport) {
        bareImports.push(`* as ${imp.import}`);
        continue;
      }
      if (imp.type === packageImport) {
        bareImports.push(imp.import);
        continue;
      }
      if (imp.type === namedAliasImport && imp.alias) {
        namedImports.push(`${imp.import} as ${imp.alias}`);
        continue;
      }
      namedImports.push(imp.import);
    }

    if (bareImports.length > 0) {
      lines.push(...formatBareImport(bareImports, pkg));
    }
    if (namedImports.length > 0) {
      lines.push(...formatBracedImport(namedImports, pkg));
    }
  }

  return lines.join("\n");
}

registerTemplateFunc("templateImports", templateImports);
