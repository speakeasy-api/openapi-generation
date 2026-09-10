type PythonV2Imports = { [key: string]: string[] };

// @ts-ignore
function genImports(): string {
  return `{{ templateImports .RecursiveComputed.Imports }}`;
}
registerTemplateFunc("genImports", genImports);

// @ts-ignore
function genLazyImports(): string {
  return `{{ templateLazyImports .RecursiveComputed.LazyImports }}`;
}
registerTemplateFunc("genLazyImports", genLazyImports);

function genCircularImports(): string {
  return `{{ templateCircularImports .RecursiveComputed.ForwardRefs }}`;
}
registerTemplateFunc("genCircularImports", genCircularImports);

/**
 * Templates import statements.
 *
 * The contents generally look like this:
 *
 * ```python
 * import <package>
 *
 * from .<package> import <type1>, <type2>
 * ```
 *
 * @param imports - A map of package names to arrays of imported items
 * @returns A string containing the formatted import statements
 */
function templateImports(imports: PythonV2Imports): string {
  // Early return if no imports
  if (!imports || Object.keys(imports).length === 0) {
    return "";
  }

  const aliasSet = new Set<string>(
    context.RecursiveComputed?.ModuleAliases ?? [],
  );
  const lines: string[] = [];
  const sortedPackages = Object.keys(imports).sort();

  for (const pkg of sortedPackages) {
    const importedItems = imports[pkg].sort();

    // Handle full package imports (denoted by empty string in imports array)
    if (importedItems.includes("")) {
      // httpx may resolve to an aliased fork depending on httpClientLibrary
      lines.push(pkg === "httpx" ? pythonHttpxImport() : `import ${pkg}`);
      // Filter out empty string for specific imports
      importedItems.splice(importedItems.indexOf(""), 1);
    }

    // Handle specific imports from package
    if (importedItems.length > 0) {
      const rendered = importedItems.map((item) =>
        aliasSet.has(item) ? `${item} as ${item}_` : item,
      );
      lines.push(`from ${pkg} import ${rendered.join(", ")}`);
    }
  }

  return lines.join("\n");
}
registerTemplateFunc("templateImports", templateImports);

function templateLazyImports(typeCheckingImports: PythonV2Imports): string {
  const importLines: string[] = [];
  if (!typeCheckingImports || Object.keys(typeCheckingImports).length === 0) {
    return "";
  }

  const sortedTypeCheckingPackages = Object.keys(typeCheckingImports).sort();
  importLines.push("if TYPE_CHECKING:");
  const indent = templateIndent(1);

  for (const pkg of sortedTypeCheckingPackages) {
    const importedItems = typeCheckingImports[pkg].sort();

    // Handle full package imports (denoted by empty string in imports array)
    if (importedItems.includes("")) {
      importLines.push(`${indent}import ${pkg}`);
      // Filter out empty string for specific imports
      importedItems.splice(importedItems.indexOf(""), 1);
    }

    // Handle specific imports from package
    if (importedItems.length > 0) {
      importLines.push(
        `${indent}from ${pkg} import ${importedItems.join(", ")}`,
      );
    }
  }

  return importLines.join("\n");
}
registerTemplateFunc("templateLazyImports", templateLazyImports);

/**
 * Templates import statements in a `if TYPE_CHECKING` block.
 * This is the recommended way to avoid circular imports, and works in tandem with forward references.
 *
 * See: https://docs.python.org/3/library/typing.html#typing.TYPE_CHECKING
 *
 * The contents generally look like this:
 *
 * ```python
 * if TYPE_CHECKING:
 *     from .<model> import *
 * ```
 *
 * @param imports - A map of package names to arrays of imported items
 * @returns A string containing the formatted forward reference import statements
 */
function templateCircularImports(imports: PythonV2Imports): string {
  // Early return if no imports
  if (!imports || Object.keys(imports).length === 0) {
    return "";
  }

  const lines: string[] = [];
  const sortedPackages = Object.keys(imports).sort();

  lines.push("if TYPE_CHECKING:");

  for (const pkg of sortedPackages) {
    const importedItems = imports[pkg].sort();
    const indent = templateIndent(1);

    // Handle full package imports (denoted by empty string in imports array)
    if (importedItems.includes("")) {
      lines.push(`${indent}import ${pkg}`);
      // Filter out empty string for specific imports
      importedItems.splice(importedItems.indexOf(""), 1);
    }

    // Handle specific imports from package
    if (importedItems.length > 0) {
      lines.push(`${indent}from ${pkg} import ${importedItems.join(", ")}`);
    }
  }

  return lines.join("\n");
}
registerTemplateFunc("templateCircularImports", templateCircularImports);

// @ts-ignore
function addImport(
  pkg: string,
  imp: string,
  internal = false,
  circularReference = false,
): string {
  context.RecursiveComputed ??= {};
  // Add import to appropriate collection based on whether it's a circular reference
  let targetImports =
    context.RecursiveComputed[circularReference ? "ForwardRefs" : "Imports"] ||
    {};

  if (circularReference) {
    addImport("typing", "TYPE_CHECKING");
  }

  // Handle internal package naming
  const packageName = internal ? getModuleImport(pkg) : pkg;

  // Initialize array for package if needed and add import if not already present
  targetImports[packageName] ??= [];
  if (!targetImports[packageName].includes(imp)) {
    targetImports[packageName].push(imp);
  }

  context.RecursiveComputed[circularReference ? "ForwardRefs" : "Imports"] =
    targetImports;

  return "";
}
registerTemplateFunc("addImport", addImport);

function addLazyImport(pkg: string, imp: string): string {
  context.RecursiveComputed ??= {};
  let key = "LazyImports";
  let targetImports = context.RecursiveComputed[key] || {};

  // Lazy importing is done for internal packages
  const packageName = getModuleImport(pkg);
  // Initialize array for package if needed and add import if not already present
  targetImports[packageName] ??= [];
  if (!targetImports[packageName].includes(imp)) {
    targetImports[packageName].push(imp);
  }

  context.RecursiveComputed[key] = targetImports;

  // When doing lazy loading we need "from typing import TYPE_CHECKING"
  // because we will be using TYPE_CHECKING. So add the neccessary
  // import when we have atleast 1 item in LazyImports
  addImport("typing", "TYPE_CHECKING");
  return "";
}
registerTemplateFunc("addLazyImport", addLazyImport);

function addModuleAlias(name: string): string {
  context.RecursiveComputed ??= {};
  context.RecursiveComputed.ModuleAliases ??= [];
  if (!context.RecursiveComputed.ModuleAliases.includes(name)) {
    context.RecursiveComputed.ModuleAliases.push(name);
  }
  return "";
}

function precomputeModuleCollisions(operations: Operation[]): string {
  const allParamNames = new Set<string>();
  const allClassLevelNames = new Set<string>();
  for (const op of operations) {
    // Collect method names (operation IDs become methods on the SDK class)
    allClassLevelNames.add(sanitizeMethodName(op));
    for (const a of op.Arguments.Sorted) {
      allParamNames.add(sanitizeParameterName(a.Name));
    }
    // Include SDK-injected parameters from templateCommonMethodParams
    if (op.Extensions?.Retries) {
      allParamNames.add("retries");
    }
  }
  // Sub-SDK property names (direct children only) can shadow module-level
  // imports at the class body level, e.g. a `utils: Utils2` property makes
  // `utils.RetryConfig` resolve to the sub-SDK instance in type annotations.
  // Only the CURRENT class's direct children are relevant — nested descendants
  // don't appear as properties on this class.
  const localSubSDKs = context?.Local?.SubSDKs;
  if (localSubSDKs) {
    for (const sdk of localSubSDKs) {
      allClassLevelNames.add(sanitizeSubSDKFieldName(sdk.FieldName));
    }
  }
  // Always-present SDK-injected parameters
  allParamNames.add("server_url");
  allParamNames.add("timeout_ms");
  allParamNames.add("http_headers");
  if (context.Global.Config.MethodTimeoutArgument === "timeout") {
    allParamNames.add("timeout");
  }
  if (context.Global.Config.MethodArguments === "positional-path-with-extras") {
    allParamNames.add("extra_headers");
    allParamNames.add("extra_query");
    allParamNames.add("extra_body");
  }
  context.RecursiveComputed ??= {};
  context.RecursiveComputed.AllParamNames = Array.from(allParamNames);
  context.RecursiveComputed.AllMethodNames = Array.from(allClassLevelNames);

  // Precompute async method name collisions for "both" mode.
  // When two operations generate methods where one's base name equals another's
  // base name + "_async", the async variant of the latter collides with the
  // former's sync method. We detect these and store the colliding op IDs so
  // sanitizeMethodName can disambiguate.
  if (context.Global.Config.AsyncMode === "both") {
    const baseNames = new Map<string, string>(); // snake_case name -> op ID
    for (const op of operations) {
      const name = caser().ToSnake(sanitizeName(op.GetID()));
      baseNames.set(name, op.GetID());
    }
    const asyncCollisions = new Set<string>();
    for (const [name, opID] of baseNames) {
      // If this operation's name ends with _async and there's another operation
      // whose name is the prefix (without _async), then this op's sync method
      // collides with the other op's async variant.
      if (name.endsWith("_async")) {
        const prefix = name.slice(0, -"_async".length);
        if (baseNames.has(prefix) && baseNames.get(prefix) !== opID) {
          asyncCollisions.add(opID);
        }
      }
    }
    context.RecursiveComputed.AsyncMethodCollisions =
      Array.from(asyncCollisions);
  }

  return "";
}
registerTemplateFunc("precomputeModuleCollisions", precomputeModuleCollisions);

function addUtilsImport(): string {
  return addImport("", "utils", true);
}
registerTemplateFunc("addUtilsImport", addUtilsImport);

/**
 * Returns the correct module reference for `utils` — either `utils` or `utils_`
 * depending on whether `utils` collides with any operation parameter name.
 * This mirrors the logic in `rewriteMethodLocals` for consistency: if method
 * bodies use `utils_`, the import is aliased `utils as utils_`, so all other
 * references (e.g. in __init__) must also use `utils_`.
 */
function resolveModuleRef(name: string): string {
  const allParamNames = context.RecursiveComputed?.AllParamNames ?? [];
  if (allParamNames.includes(name)) {
    addModuleAlias(name);
    return `${name}_`;
  }
  return name;
}
registerTemplateFunc("resolveModuleRef", resolveModuleRef);

// @ts-ignore
function addHooksImport(type: string): string {
  return addImport("_hooks", type, true);
}
registerTemplateFunc("addHooksImport", addHooksImport);

function addEventStreamingImport(): string {
  if (pythonEventStreamUsesDirectTypeRefs()) {
    addImport("utils.eventstreaming", pythonEventStreamSyncClassName(), true);
    addImport("utils.eventstreaming", pythonEventStreamAsyncClassName(), true);
    return "";
  }
  return addImport("utils", "eventstreaming", true);
}
registerTemplateFunc("addEventStreamingImport", addEventStreamingImport);

function addJsonlImport(): string {
  return addImport("utils", "jsonl", true);
}
registerTemplateFunc("addJsonlImport", addJsonlImport);

function addUnmarshalJsonResponseImport(): string {
  return addImport(
    "utils.unmarshal_json_response",
    "unmarshal_json_response",
    true,
  );
}
registerTemplateFunc(
  "addUnmarshalJsonResponseImport",
  addUnmarshalJsonResponseImport,
);

// @ts-ignore
function addSubSDKImport(subSDK: SDK): string {
  return addImport(
    sanitizeSDKFileName(subSDK.Type.Name),
    sanitizeClassName(subSDK.Type.Name),
    true,
  );
}
registerTemplateFunc("addSubSDKImport", addSubSDKImport);

// @ts-ignore
function addSubSDKImportAsync(subSDK: SDK): string {
  return addImport(
    sanitizeSDKFileName(subSDK.Type.Name),
    "Async" + sanitizeClassName(subSDK.Type.Name),
    true,
  );
}
registerTemplateFunc("addSubSDKImportAsync", addSubSDKImportAsync);

// @ts-ignore
function addLazyImportForSubSDK(subSDK: SDK): string {
  return addLazyImport(
    sanitizeSDKFileName(subSDK.Type.Name),
    sanitizeClassName(subSDK.Type.Name),
  );
}
registerTemplateFunc("addLazyImportForSubSDK", addLazyImportForSubSDK);

// @ts-ignore
function addLazyImportForSubSDKAsync(subSDK: SDK): string {
  return addLazyImport(
    sanitizeSDKFileName(subSDK.Type.Name),
    "Async" + sanitizeClassName(subSDK.Type.Name),
  );
}
registerTemplateFunc(
  "addLazyImportForSubSDKAsync",
  addLazyImportForSubSDKAsync,
);

function templateTypeImportAlias(typeDef: TypeDef): string {
  const modelLocation = getModelsLocation(typeDef.OutputLocation);
  const resolvedModelName = resolveModelName(typeDef);
  const modelParts = modelLocation.split("/");

  return `${modelParts[modelParts.length - 1]}_${resolvedModelName}`;
}

// @ts-ignore
function addTypeImport(
  typeDef: TypeDef,
  usageLocation: string,
  typeSuffix?: string,
  circularReference: boolean = false,
  forceImportAlias: boolean = false,
) {
  typeSuffix ??= "";

  const modelLocation = getModelsLocation(typeDef.OutputLocation);
  const resolvedModelName = resolveModelName(typeDef);

  const wantRelativeFileImport = usageLocation == modelLocation;
  const importLikeUser =
    !usageLocation ||
    usageLocation == "usage" ||
    usageLocation == "tests" ||
    usageLocation == "sdk";
  const internalImport =
    forceImportAlias || (!wantRelativeFileImport && !importLikeUser);

  switch (true) {
    case internalImport:
      {
        const modelParts = modelLocation.split("/");
        addImport(
          modelParts.join("."),
          `${resolvedModelName} as ${templateTypeImportAlias(typeDef)}`,
          true,
          circularReference,
        );
      }
      break;

    case wantRelativeFileImport:
      {
        let importName = sanitizeClassName(typeDef.Name) + typeSuffix;
        // When the suffix is "TypedDict", check for disambiguation overrides.
        if (typeSuffix === DEFAULT_TYPED_DICT_SUFFIX) {
          importName = resolveTypedDictName(
            typeDef,
            sanitizeClassName(typeDef.Name),
          );
        }
        addImport(
          `.${resolvedModelName}`,
          importName,
          false,
          circularReference,
        );
      }
      break;

    case importLikeUser:
      {
        if (typeDef.OutputLocation == "tests") {
          addImport("tests", "test_helpers");
        } else {
          addSDKImportFromOutputLocation(typeDef.OutputLocation);
        }
      }
      break;

    default:
      throw new Error("Unreachable code in addTypeImport");
  }
}

// @ts-ignore
function addSDKImportFromOutputLocation(outputLocation: string) {
  let outputLocationParts = getModelsLocation(outputLocation).split("/");

  const pkg = outputLocationParts.pop();

  addImport(outputLocationParts.join("."), pkg, true);
}

function getScopePath(scope: string): string {
  let scopePath = "";
  switch (scope) {
    case "errors":
      scopePath = context.Global.Config.Imports.GetErrorsPath();
      break;
    case "shared":
      scopePath = context.Global.Config.Imports.GetSharedPath();
      break;
    case "operations":
      scopePath = context.Global.Config.Imports.GetOperationsPath();
      break;
    default:
      throw new Error(`Unexpected scope: ${scope}`);
  }
  if (scopePath == "") {
    scopePath = "models";
  }
  return scopePath;
}

// @ts-ignore
function addSDKImportForAccessNamespace(scope: string): string {
  const parts = getScopePath(scope).split("/");
  if (
    usingGlobalImports() ||
    (useConflictResistantModelImports() && parts[0] === "models")
  ) {
    return addImport("", "models", true);
  }
  const pkg = parts.pop();
  return addImport(parts.join("."), pkg, true);
}
registerTemplateFunc(
  "addSDKImportForAccessNamespace",
  addSDKImportForAccessNamespace,
);

//@ts-ignore
function getErrorsLocation() {
  return getModelsLocation(context.Global.Config.Imports.GetErrorsPath());
}
registerTemplateFunc("getErrorsLocation", getErrorsLocation);

function getPythonInfraDirs(): Set<string> {
  const dirs = new Set<string>(getModuleRefNames());
  // Add Python-specific infrastructure directories
  dirs.add("sdk");
  dirs.add("models");
  dirs.add("types");
  dirs.add("_hooks");
  dirs.add("basesdk");
  return dirs;
}

/**
 * Maps an output location (from the spec or config) to the actual filesystem
 * directory where models are written. Sanitizes the path and redirects
 * single-segment namespace names under models/ to avoid collisions with
 * sub-SDK files at the package root.
 *
 * Examples (assuming default import paths — all empty):
 *   "" → "models"                       (default shared path)
 *   "models/components" → "models/components" (multi-segment, kept as-is)
 *   "pet" → "models/pet"                (single-segment namespace, redirected)
 *   "errors" → "models/errors"          (namespace, not an explicit config path)
 *
 * Examples (with explicitly configured paths like shared="models/components"):
 *   "models/components" → "models/components" (explicit config path, kept as-is)
 *   "pet" → "models/pet"                (still redirected — not in config)
 */
// @ts-ignore
function getModelsLocation(outputLocation: string): string {
  const wasEmpty = outputLocation == "";
  if (wasEmpty) {
    outputLocation = `models`;
  }

  const sanitized = sanitizeOutputLocation(outputLocation);
  // Redirect single-segment namespace names under models/ to prevent collisions
  // with sub-SDK files at the package root. Only explicitly configured import
  // paths (non-empty values in gen.yaml) are kept as-is since the SDK
  // infrastructure depends on them being at their configured locations.
  // A namespace named "models" maps to the root models/ package directly —
  // it would be redundant to redirect it to models/models/.
  if (!wasEmpty && !sanitized.includes("/") && sanitized !== "models") {
    const explicitlyConfiguredPaths = new Set(
      [
        context.Global.Config.Imports.GetSharedPath(),
        context.Global.Config.Imports.GetOperationsPath(),
        context.Global.Config.Imports.GetErrorsPath(),
        context.Global.Config.Imports.GetCallbacksPath(),
        context.Global.Config.Imports.GetWebhooksPath(),
      ]
        .filter(Boolean)
        .map((p) => sanitizeOutputLocation(p)),
    );
    if (!explicitlyConfiguredPaths.has(sanitized)) {
      return `models/${sanitized}`;
    }
  }
  return sanitized;
}

unregisterTemplateFunc("getModelsLocation");
registerTemplateFunc("getModelsLocation", getModelsLocation);

// @ts-ignore
function getAccessNamespace(scope: string): string {
  const parts = getScopePath(scope).split("/");
  if (useConflictResistantModelImports() && parts[0] === "models") {
    return `${parts.join(".")}.`;
  }
  return `${parts[parts.length - 1]}.`;
}

registerTemplateFunc("getAccessNamespace", getAccessNamespace);

// @ts-ignore
function getAccessNamespaceSuffixed(scope: string): string {
  const parts = getScopePath(scope).split("/");
  if (useConflictResistantModelImports() && parts[0] === "models") {
    // Suffix the root module name (e.g. "models" → "models_") so
    // rewriteMethodLocals can detect and resolve the collision.
    const rest = parts.slice(1);
    return rest.length > 0
      ? `${parts[0]}_.${rest.join(".")}.`
      : `${parts[0]}_.`;
  }
  return `${parts[parts.length - 1]}_.`;
}

registerTemplateFunc("getAccessNamespaceSuffixed", getAccessNamespaceSuffixed);

function getModuleRefNames(): string[] {
  const names = new Set<string>();
  for (const scope of ["errors", "shared", "operations"] as const) {
    const parts = getScopePath(scope).split("/");
    names.add(parts[parts.length - 1]);
  }
  names.add("utils");
  // With conflict-resistant imports, access prefixes shift from per-scope
  // names (e.g. "errors_.") to the root "models_." prefix. The rewriter
  // needs "models" in the ref list to recognize and handle "models_."
  // patterns in method bodies.
  if (useConflictResistantModelImports()) {
    names.add("models");
  }
  return Array.from(names);
}
registerTemplateFunc("getModuleRefNames", getModuleRefNames);

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

// @ts-ignore
function templateGlobalImport(): string {
  if (!usingGlobalImports()) {
    return "";
  }

  return `from typing import TYPE_CHECKING
from . import models as _models

if TYPE_CHECKING:
    from .models import *
else:

    def __getattr__(name: str):
        try:
            return getattr(_models, name)
        except AttributeError:
            raise AttributeError(
                f"module {__name__!r} has no attribute {name!r}"
            ) from None

    def __dir__():
        return dir(_models) + list(globals().keys())
`;
}
registerTemplateFunc("templateGlobalImport", templateGlobalImport);

// @ts-ignore
function importBaseError(): string {
  const outputLocation = getErrorsLocation();
  const formattedLocation = outputLocation.replaceAll("/", ".");
  return addImport(formattedLocation, getBaseErrorClassName(), true);
}
registerTemplateFunc("importBaseError", importBaseError);
