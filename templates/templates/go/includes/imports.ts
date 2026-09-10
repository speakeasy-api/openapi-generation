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
      case "apierrors":
        imp = getErrorsLocation();
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
      case "hooks":
        imp = `${getInternalPackageName()}/hooks`;
        local = true;
        break;
      case "polling":
        imp = getPollingLocation();
        local = true;
        break;
      case "retry":
        imp = getRetryLocation();
        local = true;
        break;
      case "config":
        imp = `${getInternalPackageName()}/config`;
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
        imp = `${getInternalPackageName()}/globals`;
        break;
    }

    imp = sanitizeOutputLocation(imp);
    imp = `${getRootModulePath()}/${imp}`;
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

/**
 * Returns the SDK package (root module package) import statement with the
 * following logic:
 *
 * - If the sdkPackageAlias configuration is set, returns the module import path
 *   and that value set as the import alias.
 * - If sanitized SDK package name differs from the final segment of the module
 *   path, returns the module import path and the sanitized SDK package name as
 *   the import alias.
 * - Otherwise, returns module import path with no alias.
 */
function addSDKPackageImport(useAlias: boolean): string {
  const modulePath = getRootModulePath();
  const sdkPackageAlias = context.Global.Config.SdkPackageAlias;

  if (sdkPackageAlias && useAlias) {
    addImport(modulePath, false, sdkPackageAlias);
  } else if (modulePath.includes("/")) {
    const modulePathParts = modulePath.split("/");
    const sanitizedSDKPackageName = sanitizeSDKPackageName(false);

    modulePathParts[modulePathParts.length - 1] != sanitizedSDKPackageName
      ? addImport(modulePath, false, sanitizedSDKPackageName)
      : addImport(modulePath);
  } else {
    addImport(modulePath);
  }

  return ""; // Still needs to return a string if used in templates
}
registerTemplateFunc("addSDKPackageImport", addSDKPackageImport);

function sanitizeModelPackageName(
  outputLocation: string,
  useAlias: boolean,
): string {
  if (!outputLocation) {
    return sanitizeSDKPackageName(useAlias);
  }

  outputLocation = sanitizeOutputLocation(outputLocation);

  if (useAlias) {
    return resolveNamespaceAlias(outputLocation);
  }

  let parts = outputLocation.split("/");
  return parts[parts.length - 1];
}
registerTemplateFunc("sanitizeModelPackageName", sanitizeModelPackageName);

/**
 * Resolves a collision-resistant import alias for a namespace output location.
 *
 * In Go, the package name for an import is its last path segment. When two
 * imports share the same last segment (e.g., "module/types" and
 * "module/models/types"), Go requires one to have an explicit import alias.
 *
 * This function detects collisions against:
 * 1. Reserved internal package names (utils, types, operations, hooks, etc.)
 * 2. Standard scope path last-segments (errors, shared, operations, etc.)
 * 3. Inter-namespace collisions (two custom namespaces producing the same
 *    package name)
 *
 * Non-colliding namespaces return the simple last-segment name unchanged.
 * Colliding namespaces get a camelCase-joined alias (e.g., "models/types" →
 * "modelstypes").
 */
function resolveNamespaceAlias(outputLocation: string): string {
  if (!context.GlobalComputed) {
    context.GlobalComputed = {};
  }

  const cacheKey = `_nsAlias_${outputLocation}`;
  if (context.GlobalComputed[cacheKey] !== undefined) {
    return context.GlobalComputed[cacheKey];
  }

  // Build the collision set lazily
  if (!context.GlobalComputed._nsAliasMap) {
    // Standard scope paths — these are imported via addImport() switch cases
    // (e.g., addImport("operations")) which handle path resolution internally.
    // They must NOT be aliased because their import lines don't go through
    // addModelImport().
    const standardScopePaths = new Set(
      [
        getErrorsLocation(),
        getSharedLocation(),
        getOperationsLocation(),
        getTypesLocation(),
        getUtilsLocation(),
        getPollingLocation(),
        getRetryLocation(),
        getNullableLocation(),
      ]
        .filter(Boolean)
        .map((loc) => sanitizeOutputLocation(loc)),
    );

    // Also include "models/components" — the HTTPMetadata/response wrapper
    // package — which is always imported without alias via addImport(loc, true)
    // from the standard templating code.
    const componentsLocation = sanitizeOutputLocation(
      context.Global.Config.Imports.GetSharedPath(),
    );
    if (componentsLocation) {
      standardScopePaths.add(componentsLocation);
    }

    // Collect all custom namespace output locations from types, excluding
    // standard scope paths
    const allOutputLocations: string[] = [];
    const allTypes = getAllTypes();
    for (const typeDef of allTypes) {
      if (typeDef.OutputLocation) {
        const loc = sanitizeOutputLocation(typeDef.OutputLocation);
        if (
          loc &&
          !allOutputLocations.includes(loc) &&
          !standardScopePaths.has(loc)
        ) {
          allOutputLocations.push(loc);
        }
      }
    }

    // Reserved package names: internal directory names used as import identifiers
    // from addImport() switch cases and globals mapping
    const reservedPkgNames = new Set([
      "utils",
      "operations",
      "shared",
      "types",
      "optionalnullable",
      "apierrors",
      "stream",
      "jsonl",
      "hooks",
      "polling",
      "retry",
      "config",
      "globals",
      "components",
    ]);

    // Also reserve last-segments of standard scope paths
    for (const loc of standardScopePaths) {
      const parts = loc.split("/");
      reservedPkgNames.add(parts[parts.length - 1]);
    }

    // Build alias map: outputLocation -> resolved alias
    const aliasMap: Record<string, string> = {};

    // Count how many output locations produce each default package name
    const pkgNameCounts: Record<string, string[]> = {};
    for (const loc of allOutputLocations) {
      const parts = loc.split("/");
      const defaultPkg = parts[parts.length - 1];
      if (!pkgNameCounts[defaultPkg]) {
        pkgNameCounts[defaultPkg] = [];
      }
      pkgNameCounts[defaultPkg].push(loc);
    }

    for (const loc of allOutputLocations) {
      const parts = loc.split("/");
      const defaultPkg = parts[parts.length - 1];

      const hasCollision =
        reservedPkgNames.has(defaultPkg) ||
        (pkgNameCounts[defaultPkg] && pkgNameCounts[defaultPkg].length > 1);

      if (hasCollision) {
        // Join all path segments: "models/types" -> "modelstypes"
        aliasMap[loc] = parts.join("");
      } else {
        aliasMap[loc] = defaultPkg;
      }
    }

    context.GlobalComputed._nsAliasMap = aliasMap;
  }

  const aliasMap = context.GlobalComputed._nsAliasMap as Record<string, string>;
  const result = aliasMap[outputLocation];
  if (result !== undefined) {
    context.GlobalComputed[cacheKey] = result;
    return result;
  }

  // Fallback for locations not in the alias map (e.g., standard scope paths)
  const parts = outputLocation.split("/");
  const fallback = parts[parts.length - 1];
  context.GlobalComputed[cacheKey] = fallback;
  return fallback;
}

/**
 * Adds an import for a model namespace, automatically applying an alias when
 * the package name collides with an internal SDK package.
 */
function addModelImport(outputLocation: string): string {
  const resolved = resolveNamespaceAlias(
    sanitizeOutputLocation(outputLocation),
  );
  const parts = sanitizeOutputLocation(outputLocation).split("/");
  const defaultPkg = parts[parts.length - 1];

  if (resolved !== defaultPkg) {
    return addImport(outputLocation, true, resolved);
  }
  return addImport(outputLocation, true);
}
registerTemplateFunc("addModelImport", addModelImport);

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

  // Check if this is a namespace package.
  // Namespace packages are created via x-speakeasy-model-namespace extension.
  // They can appear as simple directory names (e.g., "identity") or with a models prefix
  // (e.g., "models/identity").
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
  const sharedPath: string = context.Global.Config.Imports.GetSharedPath();
  if (sharedPath?.includes("pkg/")) {
    return "pkg/types";
  }

  return "types";
}
registerTemplateFunc("getTypesLocation", getTypesLocation);

function getUtilsLocation() {
  const sharedPath: string = context.Global.Config.Imports.GetSharedPath();
  if (sharedPath?.includes("pkg/")) {
    return "pkg/utils";
  }

  return `${getInternalPackageName()}/utils`;
}
registerTemplateFunc("getUtilsLocation", getUtilsLocation);

function getPollingLocation() {
  const sharedPath: string = context.Global.Config.Imports.GetSharedPath();
  if (sharedPath?.includes("pkg/")) {
    return "pkg/polling";
  }

  return "polling";
}

registerTemplateFunc("getPollingLocation", getPollingLocation);

//@ts-ignore
function getRetryLocation() {
  const sharedPath: string = context.Global.Config.Imports.GetSharedPath();
  if (sharedPath?.includes("pkg/")) {
    return "pkg/retry";
  }

  return "retry";
}
registerTemplateFunc("getRetryLocation", getRetryLocation);

//@ts-ignore
function getNullableLocation() {
  const sharedPath: string = context.Global.Config.Imports.GetSharedPath();
  if (sharedPath?.includes("pkg/")) {
    return "pkg/optionalnullable";
  }

  return "optionalnullable";
}
registerTemplateFunc("getNullableLocation", getNullableLocation);

function getInternalPackageName(): string {
  if (isTemplateFeatureEnabled("internalModules")) {
    return "internal";
  }

  return "sdkinternal";
}
