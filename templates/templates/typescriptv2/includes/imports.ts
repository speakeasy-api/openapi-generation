// @ts-ignore
const typeImport = "typeImport";
// @ts-ignore
const aliasImport = "aliasImport";
// @ts-ignore
const packageImport = "packageImport";
// @ts-ignore
const namedAliasImport = "namedAliasImport";

type TSImportType =
  | typeof typeImport
  | typeof aliasImport
  | typeof packageImport
  | typeof namedAliasImport;

// @ts-ignore
type TSImports = {
  [key: string]: { import: string; type: string; alias?: string }[];
};

const standardImports = "Imports";
const usageImports = "UsageImports";
const testImports = "TestImports";

// @ts-ignore
type ImportTypeLocation =
  | typeof standardImports
  | typeof usageImports
  | typeof testImports;

// @ts-ignore
function templateSDKClassName(name: string): string {
  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }
  context.RecursiveComputed.DefiningClassName = name;
  return name;
}
registerTemplateFunc("templateSDKClassName", templateSDKClassName);

// @ts-ignore
function genImports(importType: ImportTypeLocation = standardImports): string {
  return `{{ templateImports .RecursiveComputed.${importType} }}`;
}
registerTemplateFunc("genImports", genImports);

// @ts-ignore
function addImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
  importType: ImportTypeLocation = standardImports,
): string {
  let imports: TSImports = context.RecursiveComputed?.[importType];
  if (!imports) {
    imports = {};
  }

  if (!imports[pkg]) {
    imports[pkg] = [];
  }

  let actualType = type;
  const definingClassName = context.RecursiveComputed?.DefiningClassName as
    | string
    | undefined;
  let alias: string | undefined;
  if (type === typeImport && definingClassName === imp) {
    actualType = namedAliasImport;
    alias = `${imp}$Model`;
  }

  if (!imports[pkg].find((i) => i.import == imp && i.type == actualType)) {
    imports[pkg].push({
      import: imp,
      type: actualType,
      ...(alias && { alias }),
    });
  }

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }

  context.RecursiveComputed[importType] = imports;

  return "";
}
registerTemplateFunc("addImport", addImport);

function addTestImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
): string {
  return addImport(pkg, imp, type, testImports);
}
registerTemplateFunc("addTestImport", addTestImport);

function addUsageImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
): string {
  return addImport(pkg, imp, type, usageImports);
}
registerTemplateFunc("addUsageImport", addUsageImport);

function addFuncImport(op: Operation, usageLocation: string) {
  const funcName = sanitizeFuncName(op);
  const funcFilename = sanitizeFuncFilename(op);
  const prefix = getImportPrefix("funcs", usageLocation);
  return addImport(`${prefix}/${funcFilename}.js`, funcName, typeImport);
}
registerTemplateFunc("addFuncImport", addFuncImport);

function addWebhookHandlerFuncImport(usageLocation: string) {
  const funcName = sanitizeWebhookHandlerFuncName();
  const funcFilename = sanitizeWebhookHandlerFuncFilename();
  const prefix = getImportPrefix("funcs", usageLocation);
  return addImport(`${prefix}/${funcFilename}.js`, funcName, typeImport);
}
registerTemplateFunc(
  "addWebhookHandlerFuncImport",
  addWebhookHandlerFuncImport,
);

function addFuncUsageImport(op: Operation) {
  const funcName = sanitizeFuncName(op);
  const funcFilename = sanitizeFuncFilename(op);
  const prefix = context.Global.Config.PackageName;
  const path = `${prefix}/funcs/${funcFilename}.js`;
  return addImport(path, funcName, typeImport, usageImports);
}
registerTemplateFunc("addFuncUsageImport", addFuncUsageImport);

function addFuncAcceptEnumImport(op: Operation, usageLocation: string) {
  const enumName = sanitizeAcceptEnumName(op);
  const funcFilename = sanitizeFuncFilename(op);
  const prefix = getImportPrefix("funcs", usageLocation);
  return addImport(`${prefix}/${funcFilename}.js`, enumName, typeImport);
}
registerTemplateFunc("addFuncAcceptEnumImport", addFuncAcceptEnumImport);

function getFuncImportPath(op: Operation, usageLocation: string) {
  const funcFilename = sanitizeFuncFilename(op);
  const prefix = getImportPrefix("funcs", usageLocation);
  return `${prefix}/${funcFilename}.js`;
}
registerTemplateFunc("getFuncImportPath", getFuncImportPath);

function getFuncUsagePath(op: Operation): string {
  const funcFilename = sanitizeFuncFilename(op);
  return `${context.Global.Config.PackageName}/${funcFilename}.js`;
}
registerTemplateFunc("getFuncUsagePath", getFuncUsagePath);

function addCoreSDKUsageImport() {
  const prefix = context.Global.Config.PackageName;
  const path = `${prefix}/core.js`;
  return addImport(path, sanitizeCoreSDKName(), typeImport, usageImports);
}
registerTemplateFunc("addCoreSDKUsageImport", addCoreSDKUsageImport);

// @ts-ignore
function addJsonLStreamUsageImport(): string {
  const prefix = context.Global.Config.PackageName;
  const path = `${prefix}/lib/jsonl.js`;
  return addImport(path, "JsonLStream", typeImport, usageImports);
}
registerTemplateFunc("addJsonLStreamUsageImport", addJsonLStreamUsageImport);

function addSDKHooksImport(
  pkg: string,
  imp: string,
  type: TSImportType = typeImport,
): string {
  return addImport(`${getImportPrefix("hooks", "")}/${pkg}.js`, imp, type);
}
registerTemplateFunc("addSDKHooksImport", addSDKHooksImport);

function addLibImport(globalSecurity: TypeDef | undefined): string {
  if (globalSecurity != null && !canFlattenGlobalSecurity()) {
    return addTypeDefImport(globalSecurity, "lib");
  }

  return "";
}
registerTemplateFunc("addLibImport", addLibImport);

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

function addTypeImport(
  pkg: string,
  symbol: string,
  usageLocation: string,
  type: TSImportType = typeImport,
): string {
  return addImport(
    `${getImportPrefix(getTypesLocation(), usageLocation)}/${pkg}.js`,
    symbol,
    type,
  );
}
registerTemplateFunc("addTypeImport", addTypeImport);

function resolveTypeDefImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): { path: string; name: string; type: TSImportType } | null {
  if (modelName !== "" && inSameModel(modelName, usageLocation, def)) {
    return null;
  }

  const mfilename = resolveModelName(def);
  const outputLocation = getModelsLocation(def.OutputLocation);
  const relativePath = `${getImportPrefix(outputLocation, usageLocation)}`;
  const className = sanitizeClassRef(def, usageLocation);

  const qualified = context.Global.Config.UseIndexModules;

  // In collision mode, all model subdirectory imports go through the parent barrel
  const useParentBarrel =
    qualified &&
    hasNamespaceCollisions() &&
    isModelSubdirectory(outputLocation);

  if (!useParentBarrel && outputLocation == usageLocation) {
    const path = `./${mfilename}.js`;
    return { path, name: className, type: typeImport };
  }

  const extpath = qualified
    ? useParentBarrel
      ? `${getImportPrefix(getModelsLocation(""), usageLocation)}/index.js`
      : `${relativePath}/index.js`
    : `${relativePath}/${mfilename}.js`;
  const extname = qualified
    ? useParentBarrel
      ? getParentBarrelAlias()
      : outputLocation.split("/").pop()
    : className;
  const exttype = qualified ? aliasImport : typeImport;

  return { path: extpath, name: extname, type: exttype };
}

function addTypeDefImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): string {
  const resolved = resolveTypeDefImport(def, usageLocation, modelName);
  if (!resolved) {
    return "";
  }

  return addImport(resolved.path, resolved.name, resolved.type);
}
registerTemplateFunc("addTypeDefImport", addTypeDefImport);

function addEnumTypeImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): string {
  const resolved = resolveTypeDefImport(def, usageLocation, modelName);
  if (!resolved) {
    return "";
  }

  if (def.Enum?.Open && getEnumFormat(def) === "enum") {
    resolved.name += resolved.type === aliasImport ? "" : "Open";
  }

  return addImport(resolved.path, resolved.name, resolved.type);
}
registerTemplateFunc("addEnumTypeImport", addEnumTypeImport);

function isNoZod(): boolean {
  return context.Global.Config.ZodVersion === "none";
}
registerTemplateFunc("isNoZod", isNoZod);

function addZodImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
  symbol: "inboundSchema" | "outboundSchema" | "Inbound" | "Outbound",
): string {
  // In no-zod there are no $inboundSchema/$outboundSchema/$Outbound
  // constants — drop any attempt to import them.
  if (isNoZod()) {
    if (symbol === "inboundSchema" || symbol === "outboundSchema") {
      return "";
    }
    if (symbol === "Outbound") {
      const t = String((def as { Type?: unknown }).Type ?? "");
      if (t !== "error") return "";
    }
  }

  const name = sanitizeZodRef(def, usageLocation) + symbol;
  if (modelName !== "" && inSameModel(modelName, usageLocation, def)) {
    return "";
  }

  const { outputLocation, mfilename, extpath, extname, exttype } =
    resolveModelImport({ def, usageLocation, name });

  if (outputLocation == usageLocation) {
    const path = `./${mfilename}.js`;
    return addImport(path, name, typeImport);
  }

  return addImport(extpath, extname, exttype);
}

function isZodV4(): boolean {
  const version = context.Global.Config.ZodVersion;
  return version === "v4" || version === "v4-mini";
}
registerTemplateFunc("isZodV4", isZodV4);

function isZodV4Mini(): boolean {
  return context.Global.Config.ZodVersion === "v4-mini";
}
registerTemplateFunc("isZodV4Mini", isZodV4Mini);

function isZodV4NotMini(): boolean {
  return context.Global.Config.ZodVersion === "v4";
}
registerTemplateFunc("isZodV4NotMini", isZodV4NotMini);

function getZodPackagePath(): string {
  if (isZodV4Mini()) {
    return "zod/v4-mini";
  }
  if (isZodV4()) {
    return "zod/v4";
  }
  return "zod/v3";
}

function addZodPackageImport(t: string = "z") {
  // In no-zod mode we never emit a zod import. Callers can still invoke this
  // unconditionally; treating it as a no-op keeps the templating code simple.
  if (isNoZod()) {
    return "";
  }
  if (t === "z") {
    return addImport(getZodPackagePath(), "z", "aliasImport");
  }
  return addImport(getZodPackagePath(), t);
}
registerTemplateFunc("addZodPackageImport", addZodPackageImport);

function addZodCorePackageImport(t: string = "z") {
  return addImport("zod/v4/core", t, "aliasImport");
}
registerTemplateFunc("addZodCorePackageImport", addZodCorePackageImport);

function addFromJSONImport(
  def: TypeDef,
  usageLocation: string,
  modelName = "",
): string {
  const name = sanitizeFromJSONFuncName(def);

  if (modelName !== "" && inSameModel(modelName, usageLocation, def)) {
    return name;
  }

  const { outputLocation, mfilename, extpath, extname, exttype, qualified } =
    resolveModelImport({ def, usageLocation, name });

  if (outputLocation == usageLocation) {
    const path = `./${mfilename}.js`;
    addImport(path, name, typeImport);
  } else {
    addImport(extpath, extname, exttype);
  }

  // We return the qualified name here so it can be used
  return qualified ? `${extname}.${name}` : name;
}
registerTemplateFunc("addFromJSONImport", addFromJSONImport);

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

  const qualified = context.Global.Config.UseIndexModules;

  // In collision mode, all model subdirectory imports go through the parent barrel
  const useParentBarrel =
    qualified &&
    hasNamespaceCollisions() &&
    isModelSubdirectory(outputLocation);

  const extpath = qualified
    ? useParentBarrel
      ? `${getImportPrefix(getModelsLocation(""), usageLocation)}/index.js`
      : `${relativePath}/index.js`
    : `${relativePath}/${mfilename}.js`;
  const extname = qualified
    ? useParentBarrel
      ? getParentBarrelAlias()
      : outputLocation.split("/").pop()
    : name;
  const exttype = qualified ? aliasImport : typeImport;
  return { outputLocation, mfilename, extpath, extname, exttype, qualified };
}

// @ts-ignore
function addUsageImportForType(
  typeDef: TypeDef,
  importType: TSImportType = typeImport,
): string {
  const types: ImportTypeLocation[] = [testImports, usageImports];

  for (const type of types) {
    if (!typeDef.IsCustomType()) {
      continue;
    }

    const outputLocation = getModelsLocation(typeDef.OutputLocation);

    const typeName = sanitizeClassName(typeDef.Name);
    const filename = resolveModelName(typeDef);

    // In collision mode, all model subdirectory aliases use the parent barrel name
    let aliasIdentifier: string;
    if (hasNamespaceCollisions() && isModelSubdirectory(outputLocation)) {
      aliasIdentifier = getParentBarrelAlias();
    } else {
      aliasIdentifier = trimTrailing(
        getAccessNamespace(typeDef.Scope, typeDef.OutputLocation),
        ".",
      );
    }

    addUsageImportInner({
      importTypeLocation: type,
      outputLocation,
      filename,
      typeName,
      importType,
      aliasIdentifier,
    });
  }

  return "";
}
registerTemplateFunc("addUsageImportForType", addUsageImportForType);

function addUsageImportInner({
  importTypeLocation,
  outputLocation,
  filename,
  typeName,
  importType,
  aliasIdentifier,
}: {
  importTypeLocation: ImportTypeLocation;
  importType: TSImportType;
  outputLocation: string;
  filename: string;
  typeName: string;
  aliasIdentifier: string;
}) {
  let pkg = context.Global.Config.PackageName;
  if (importTypeLocation === testImports) {
    pkg = "..";
  }

  const globalImports =
    usingGlobalImports() && context.Global.Config.UseIndexModules;

  // Namespace packages (via x-speakeasy-model-namespace) need their path even with global imports
  const isNamespacePackage = isNamespaceOutputLocation(outputLocation);

  // In collision mode, model subdirectory imports use the parent barrel
  if (
    hasNamespaceCollisions() &&
    isModelSubdirectory(outputLocation) &&
    context.Global.Config.UseIndexModules
  ) {
    pkg += `/${getModelsLocation("")}`;
  } else if (outputLocation && (!globalImports || isNamespacePackage)) {
    pkg += `/${outputLocation}`;
  }

  if (!context.Global.Config.UseIndexModules) {
    pkg += `/${filename}.js`;
    addImport(pkg, typeName, typeImport, importTypeLocation);
    return;
  }

  if (importTypeLocation === testImports) {
    pkg += "/index.js";
  }

  if (importType === aliasImport) {
    addImport(pkg, aliasIdentifier, aliasImport, importTypeLocation);
  } else {
    addImport(pkg, typeName, typeImport, importTypeLocation);
  }
}

function addUsageTypeImport(pkg: string, type: string): string {
  const types: ImportTypeLocation[] = [testImports, usageImports];

  for (const t of types) {
    let prefix = context.Global.Config.PackageName;
    if (t === testImports) {
      prefix = "..";
    }

    let path = `${prefix}/${getTypesLocation()}`;
    if (!context.Global.Config.UseIndexModules) {
      addImport(`${path}/${pkg}.js`, type, typeImport, t);
      continue;
    }

    if (t === testImports) {
      path += "/index.js";
    }

    addImport(path, type, typeImport, t);
  }

  return "";
}

function addUsageErrorsImport(): string {
  return addImport(getErrorsLocation(), "errors", aliasImport, usageImports);
}
registerTemplateFunc("addUsageErrorsImport", addUsageErrorsImport);

function addUsageErrorImport(t: BuiltinError): string {
  const outputLocation = getErrorsLocation();
  const importContexts: ImportTypeLocation[] = [testImports, usageImports];

  for (const importContext of importContexts) {
    const filename = getFileNameForBuiltinError(t);
    const typeName = t;
    const aliasIdentifier = trimTrailing(getAccessNamespace("errors"), ".");

    addUsageImportInner({
      importTypeLocation: importContext,
      outputLocation,
      filename,
      typeName,
      importType: aliasImport,
      aliasIdentifier,
    });
  }

  return "";
}
registerTemplateFunc("addUsageErrorImport", addUsageErrorImport);

function getSDKValidationErrorFileName(): string {
  return sanitizeFileName("SDKValidationError");
}
registerTemplateFunc(
  "getSDKValidationErrorFileName",
  getSDKValidationErrorFileName,
);

function getResponseValidationErrorFileName(): string {
  return sanitizeFileName("ResponseValidationError");
}
registerTemplateFunc(
  "getResponseValidationErrorFileName",
  getResponseValidationErrorFileName,
);

function getHttpClientErrorsFileName(): string {
  return sanitizeFileName("HttpClientErrors");
}
registerTemplateFunc(
  "getHttpClientErrorsFileName",
  getHttpClientErrorsFileName,
);

// Type helper files keep their original casing unless legacyFileNaming is false (kebab mode)
function sanitizeTypeHelperFileName(name: string): string {
  if (!context.Global?.Config?.LegacyFileNaming) {
    return caser().ToKebab(name).toLowerCase();
  }
  return name;
}

function getDefaultToZeroValueFileName(): string {
  return sanitizeTypeHelperFileName("defaultToZeroValue");
}
registerTemplateFunc(
  "getDefaultToZeroValueFileName",
  getDefaultToZeroValueFileName,
);

function getUnrecognizedFileName(): string {
  return sanitizeTypeHelperFileName("unrecognized");
}
registerTemplateFunc("getUnrecognizedFileName", getUnrecognizedFileName);

function getSmartUnionFileName(): string {
  return sanitizeTypeHelperFileName("smartUnion");
}
registerTemplateFunc("getSmartUnionFileName", getSmartUnionFileName);

function getDiscriminatedUnionFileName(): string {
  return sanitizeTypeHelperFileName("discriminatedUnion");
}
registerTemplateFunc(
  "getDiscriminatedUnionFileName",
  getDiscriminatedUnionFileName,
);

function getConstDateTimeFileName(): string {
  if (!context.Global?.Config?.LegacyFileNaming) {
    return "const-date-time";
  }
  return "constdatetime";
}
registerTemplateFunc("getConstDateTimeFileName", getConstDateTimeFileName);

function getClientCredentialsFileName(): string {
  if (!context.Global?.Config?.LegacyFileNaming) {
    return "client-credentials";
  }
  return "clientcredentials";
}
registerTemplateFunc(
  "getClientCredentialsFileName",
  getClientCredentialsFileName,
);

function getOAuth2ScopesFileName(): string {
  if (!context.Global?.Config?.LegacyFileNaming) {
    return "oauth2-scopes";
  }
  return "oauth2scopes";
}
registerTemplateFunc("getOAuth2ScopesFileName", getOAuth2ScopesFileName);

function getTestClientFileName(): string {
  if (!context.Global?.Config?.LegacyFileNaming) {
    return "test-client";
  }
  return "testclient";
}
registerTemplateFunc("getTestClientFileName", getTestClientFileName);

function getFileNameForBuiltinError(errorName: BuiltinError): string {
  switch (errorName) {
    case getBaseErrorClassName():
      return `${getBaseErrorFileName()}.js`;
    case getDefaultErrorClassName():
      return `${getDefaultErrorFileName()}.js`;
    case "SDKValidationError":
      return `${getSDKValidationErrorFileName()}.js`;
    case "ResponseValidationError":
      return `${getResponseValidationErrorFileName()}.js`;
    case "UnexpectedClientError":
    case "InvalidRequestError":
    case "RequestAbortedError":
    case "RequestTimeoutError":
    case "ConnectionError":
      return `${getHttpClientErrorsFileName()}.js`;
    default:
      throw new Error(`Unknown built in error type: ${errorName}`);
  }
}

function getUsageAccessErrors() {
  if (!context.Global.Config.UseIndexModules) {
    return "";
  }
  return getAccessNamespace("errors");
}
registerTemplateFunc("getUsageAccessErrors", getUsageAccessErrors);

function addErrorTypeImport(type: BuiltinError, relativeTo = ""): string {
  const outputLocation = getModelsLocation(getErrorsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);

  switch (type) {
    case getDefaultErrorClassName():
      return addImport(
        `${prefix}/${getDefaultErrorFileName()}.js`,
        type,
        typeImport,
      );
    case getBaseErrorClassName():
      return addImport(
        `${prefix}/${getBaseErrorFileName()}.js`,
        type,
        typeImport,
      );
    case "SDKValidationError":
      return addImport(
        `${prefix}/${getSDKValidationErrorFileName()}.js`,
        type,
        typeImport,
      );
    case "ResponseValidationError":
      return addImport(
        `${prefix}/${getResponseValidationErrorFileName()}.js`,
        type,
        typeImport,
      );
    case "UnexpectedClientError":
    case "InvalidRequestError":
    case "RequestAbortedError":
    case "RequestTimeoutError":
    case "ConnectionError":
      return addImport(
        `${prefix}/${getHttpClientErrorsFileName()}.js`,
        type,
        typeImport,
      );
    default:
      throw new Error(`Unknown error type: ${type}`);
  }
}
registerTemplateFunc("addErrorTypeImport", addErrorTypeImport);

function importBaseError(relativeTo = ""): string {
  const outputLocation = getModelsLocation(getErrorsLocation());
  const prefix = getImportPrefix(outputLocation, relativeTo);
  return addImport(
    `${prefix}/${getBaseErrorFileName()}.js`,
    getBaseErrorClassName(),
    typeImport,
  );
}
registerTemplateFunc("importBaseError", importBaseError);

function addErrorTypeUsageImport(type: BuiltinError): string {
  const outputLocation = getModelsLocation(getErrorsLocation());
  const prefix = `${context.Global.Config.PackageName}/${outputLocation}`;

  switch (type) {
    case getDefaultErrorClassName(): {
      const path = `${prefix}/${getDefaultErrorFileName()}.js`;
      return addImport(path, type, typeImport, usageImports);
    }
    case getBaseErrorClassName(): {
      const path = `${prefix}/${getBaseErrorFileName()}.js`;
      return addImport(path, type, typeImport, usageImports);
    }
    case "SDKValidationError": {
      const path = `${prefix}/${getSDKValidationErrorFileName()}.js`;
      return addImport(path, type, typeImport, usageImports);
    }
    case "ResponseValidationError": {
      const path = `${prefix}/${getResponseValidationErrorFileName()}.js`;
      return addImport(path, type, typeImport, usageImports);
    }
    case "UnexpectedClientError":
    case "InvalidRequestError":
    case "RequestAbortedError":
    case "RequestTimeoutError":
    case "ConnectionError": {
      const path = `${prefix}/${getHttpClientErrorsFileName()}.js`;
      return addImport(path, type, typeImport, usageImports);
    }
    default:
      throw new Error(`Unknown error type: ${type}`);
  }
}
registerTemplateFunc("addErrorTypeUsageImport", addErrorTypeUsageImport);

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

// @ts-ignore
function templateImports(imports: TSImports): string {
  let lines = [];

  let keys = Object.keys(imports ?? {}).sort();

  if (isDebug()) {
    // Group by the identifiers which will be created
    // If we will create duplicate identifiers this will produce an error
    const grouped = new Map<string, string[]>();
    for (const [pkg, imps] of Object.entries(imports || {})) {
      for (const imp of imps) {
        const key = imp.alias || imp.import;
        const group = grouped.get(key) || [];
        grouped.set(key, [...group, pkg]);
      }
    }

    // If any groups have more than one type, throw an error
    for (const [key, value] of grouped.entries()) {
      if (value.length > 1) {
        throw new Error(
          `Tried to import duplicate: "${key}" from ${value
            .map((v) => `"${v}"`)
            .join(", ")}`,
        );
      }
    }
  }

  for (const pkg of keys) {
    let imported = imports[pkg].sort((a, b) =>
      a.import.localeCompare(b.import),
    );

    let aliases = [];
    let types = [];

    for (const imp of imported) {
      switch (imp.type) {
        case aliasImport:
          aliases.push(`* as ${imp.import}`);
          break;
        case packageImport:
          aliases.push(imp.import);
          break;
        case typeImport:
          types.push(imp.import);
          break;
        case namedAliasImport:
          types.push(`${imp.import} as ${imp.alias}`);
          break;
      }
    }

    if (aliases.length > 0) {
      if (aliases.join(", ").length > 80) {
        lines.push(`import {`);
        for (const alias of aliases) {
          lines.push(`  ${alias},`);
        }
        lines.push(`} from "${pkg}";`);
      } else {
        lines.push(`import ${aliases.join(", ")} from "${pkg}";`);
      }
    }
    if (types.length > 0) {
      if (types.join(", ").length > 80) {
        lines.push(`import {`);
        for (const type of types) {
          lines.push(`  ${type},`);
        }
        lines.push(`} from "${pkg}";`);
      } else {
        lines.push(`import { ${types.join(", ")} } from "${pkg}";`);
      }
    }
  }

  return lines.join("\n");
}
registerTemplateFunc("templateImports", templateImports);

// @ts-ignore
function getModelsLocation(outputLocation: string): string {
  if (outputLocation == "") {
    outputLocation = `models`;
  }

  return sanitizeOutputLocation(outputLocation);
}

unregisterTemplateFunc("getModelsLocation");
registerTemplateFunc("getModelsLocation", getModelsLocation);

// @ts-ignore
function getErrorsLocation() {
  return getModelsLocation(context.Global.Config.Imports.GetErrorsPath());
}

registerTemplateFunc("getErrorsLocation", getErrorsLocation);

function getSharedLocation(): string {
  return getModelsLocation(context.Global.Config.Imports.GetSharedPath());
}
registerTemplateFunc("getSharedLocation", getSharedLocation);

function getOperationsLocation(): string {
  return getModelsLocation(context.Global.Config.Imports.GetOperationsPath());
}
registerTemplateFunc("getOperationsLocation", getOperationsLocation);

function getWebhooksLocation(): string {
  return getModelsLocation(context.Global.Config.Imports.GetWebhooksPath());
}

function getCallbacksLocation(): string {
  return getModelsLocation(context.Global.Config.Imports.GetCallbacksPath());
}

function getTypesLocation() {
  if (usingLegacyImports()) {
    return "sdk/types";
  }

  return "types";
}
registerTemplateFunc("getTypesLocation", getTypesLocation);

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

// Returns true if the output location is a subdirectory of the root models
// location (e.g. "models/types", "models/errors"). Used by collision mode to
// determine which imports should go through the parent barrel.
function isModelSubdirectory(outputLocation: string): boolean {
  if (!outputLocation) return false;
  const root = getModelsLocation("");
  const normalized = getModelsLocation(outputLocation);
  return normalized !== root && normalized.startsWith(root + "/");
}

// Detects whether any custom namespace (x-speakeasy-model-namespace) would
// collide with reserved aliases or other namespaces when used as an import
// alias. When collisions exist, callers switch to importing from the parent
// models barrel and using dotted access (e.g., `models.types.Widget`).
// Non-colliding SDKs are completely unaffected.
function hasNamespaceCollisions(): boolean {
  if (context.GlobalComputed?._hasNsCollisions !== undefined) {
    return context.GlobalComputed._hasNsCollisions;
  }
  if (!context.GlobalComputed) context.GlobalComputed = {};

  // Reserved aliases that custom namespace last-segments could collide with
  const reserved = new Set(["types", "types$"]);
  // Standard scope last-segments
  for (const loc of [
    getErrorsLocation(),
    getSharedLocation(),
    getOperationsLocation(),
    getWebhooksLocation(),
    getCallbacksLocation(),
  ]) {
    if (loc) {
      const parts = loc.split("/");
      reserved.add(parts[parts.length - 1]);
    }
  }
  // Func names (sdk.ts imports both func names and namespace aliases)
  for (const op of iterateOperations()) {
    reserved.add(sanitizeFuncName(op));
  }

  let found = false;
  if (context.Global.AST.BucketedTypes) {
    for (const [rawLoc] of sequencedMapEntries(
      context.Global.AST.BucketedTypes,
    )) {
      const loc = getModelsLocation(sanitizeOutputLocation(rawLoc));
      if (!isNamespaceOutputLocation(rawLoc)) continue;
      const lastSeg = loc.split("/").pop();
      if (reserved.has(lastSeg)) {
        found = true;
        break;
      }
    }
  }
  // Also check inter-namespace collisions (two namespaces with same last segment)
  if (!found && context.Global.AST.BucketedTypes) {
    const seen = new Set<string>();
    for (const [rawLoc] of sequencedMapEntries(
      context.Global.AST.BucketedTypes,
    )) {
      if (!isNamespaceOutputLocation(rawLoc)) continue;
      const lastSeg = getModelsLocation(sanitizeOutputLocation(rawLoc))
        .split("/")
        .pop();
      if (seen.has(lastSeg)) {
        found = true;
        break;
      }
      seen.add(lastSeg);
    }
  }

  context.GlobalComputed._hasNsCollisions = found;
  return found;
}

// Returns the alias to use for the parent models barrel import.
// Normally this is the root models directory name (e.g. "models"), but if
// that collides with a func name, we append "$" to avoid conflicts.
function getParentBarrelAlias(): string {
  if (context.GlobalComputed?._parentBarrelAlias !== undefined) {
    return context.GlobalComputed._parentBarrelAlias;
  }
  if (!context.GlobalComputed) context.GlobalComputed = {};

  const rootName = getModelsLocation("").split("/").pop();

  // Check if root name collides with any func name
  let collides = false;
  for (const op of iterateOperations()) {
    if (sanitizeFuncName(op) === rootName) {
      collides = true;
      break;
    }
  }

  const alias = collides ? `${rootName}$` : rootName;
  context.GlobalComputed._parentBarrelAlias = alias;
  return alias;
}

// @ts-ignore
function getAccessNamespace(scope: string, outputLocation?: string): string {
  let scopePath = "";

  switch (scope.toString()) {
    case "":
      return "";
    case "errors":
      scopePath = context.Global.Config.Imports.GetErrorsPath();
      break;
    case "shared":
      scopePath = context.Global.Config.Imports.GetSharedPath();
      break;
    case "operations":
      scopePath = context.Global.Config.Imports.GetOperationsPath();
      break;
    case "callbacks":
      scopePath = context.Global.Config.Imports.GetCallbacksPath();
      break;
    case "webhooks":
      scopePath = context.Global.Config.Imports.GetWebhooksPath();
      break;
    default:
      throw new Error(`Unexpected scope: ${scope}`);
  }

  if (scopePath == "") {
    scopePath = "models";
  }

  // Namespace packages (via x-speakeasy-model-namespace) override the default scope path
  // e.g. outputLocation="models/identity" vs scopePath="models/errors"
  if (outputLocation && isNamespaceOutputLocation(outputLocation)) {
    scopePath = outputLocation;
  }

  scopePath = sanitizeOutputLocation(scopePath);
  const parts = scopePath.split("/");
  if (hasNamespaceCollisions() && isModelSubdirectory(scopePath)) {
    // Dotted access through parent barrel: "models$.types." or "models.types."
    const rootAlias = getParentBarrelAlias();
    return `${rootAlias}.${parts.slice(1).join(".")}.`;
  }
  return `${parts[parts.length - 1]}.`;
}
registerTemplateFunc("getAccessNamespace", getAccessNamespace);

// @ts-ignore
function usingGlobalImports(): boolean {
  let allImportsIntendedForGlobal = true;

  for (const scope in context.Global.Config.Imports.Paths) {
    // The resources key configures the public-export surface, not a model
    // scope; it must not change how model imports are laid out.
    if (scope === "resources") {
      continue;
    }
    const path = context.Global.Config.Imports.Paths[scope];

    if (path != "" && path != "models") {
      allImportsIntendedForGlobal = false;
      break;
    }
  }

  return allImportsIntendedForGlobal;
}
registerTemplateFunc("usingGlobalImports", usingGlobalImports);

// @ts-ignore
function usingLegacyImports(): boolean {
  let usingLegacyImports = true;

  for (const scope in context.Global.Config.Imports.Paths) {
    const path = context.Global.Config.Imports.Paths[scope];

    switch (scope) {
      case "operations":
        usingLegacyImports =
          usingLegacyImports && path == "sdk/models/operations";
        break;
      case "shared":
        usingLegacyImports = usingLegacyImports && path == "sdk/models/shared";
        break;
      case "errors":
        usingLegacyImports = usingLegacyImports && path == "sdk/models/errors";
        break;
      case "callbacks":
        usingLegacyImports =
          usingLegacyImports && path == "sdk/models/callbacks";
        break;
      case "webhooks":
        usingLegacyImports =
          usingLegacyImports && path == "sdk/models/webhooks";
        break;
    }
  }

  return usingLegacyImports;
}

registerTemplateFunc("usingLegacyImports", usingLegacyImports);
