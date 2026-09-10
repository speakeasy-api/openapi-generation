// Cached results — computed once per generation.
let _collisionCheckResult: boolean | null = null;
let _knownSubDirs: Set<string> | null = null;

// Detects whether x-speakeasy-model-namespace creates a child module that
// shadows an existing module under the same shared import path, e.g. if
// shared is "models/components" and a namespace creates "models/models", the
// resulting Models::Models shadows the parent Models module.
function hasModelNamespaceCollisions(): boolean {
  if (_collisionCheckResult !== null) {
    return _collisionCheckResult;
  }

  const sharedRoot = getScopeRoot("shared").toLowerCase();
  _collisionCheckResult = false;

  for (const [outputLocation] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    const parts = outputLocation.split("/");
    if (
      parts.length >= 2 &&
      parts[0].toLowerCase() === sharedRoot &&
      parts[1].toLowerCase() === sharedRoot
    ) {
      _collisionCheckResult = true;
      break;
    }
  }

  return _collisionCheckResult;
}

// Known root directories for scope/utility paths — cached once per generation.
function getKnownSubDirs(): Set<string> {
  if (_knownSubDirs !== null) return _knownSubDirs;
  _knownSubDirs = new Set<string>(["utils", "hooks"]);
  const scopes = ["shared", "operations", "errors", "callbacks", "webhooks"];
  for (const key of scopes) {
    const root = getScopeRoot(key);
    if (root) _knownSubDirs.add(root.toLowerCase());
  }
  return _knownSubDirs;
}

// @ts-ignore
function getModelNamespace(outputLocation: string): string {
  if (outputLocation) {
    const parts = outputLocation.split("/");
    const namespace = parts
      .map((l) => caser().ToPascal(sanitizeName(l)))
      .join("::");

    // For custom namespaces (not under a known scope/utility path),
    // we need to prefix with the root module to ensure proper resolution.
    if (!getKnownSubDirs().has(parts[0].toLowerCase()) && parts[0] !== "") {
      // Custom namespace - prefix with root module for proper constant resolution
      const module = sanitizeModuleName(context.Global.Config.Module);
      return `::${module}::${namespace}`;
    }

    // When a namespace creates a child module that shadows its parent
    // (e.g., models/models -> Models::Models), use fully-qualified paths
    // to avoid Ruby's relative constant resolution ambiguity.
    if (hasModelNamespaceCollisions()) {
      const module = sanitizeModuleName(context.Global.Config.Module);
      return `::${module}::${namespace}`;
    }

    return namespace;
  }
}

registerTemplateFunc("getModelNamespace", getModelNamespace);

// Returns an array of module names for nested module declarations
// e.g., "models/bar" -> ["Models", "Bar"]
// @ts-ignore
function getNamespaceModuleParts(outputLocation: string): string[] {
  if (outputLocation) {
    const parts = outputLocation.split("/");
    return parts.map((l) => caser().ToPascal(l));
  }
  return [];
}

registerTemplateFunc("getNamespaceModuleParts", getNamespaceModuleParts);

// Gets the full qualified class path for namespaces in RBI files
// e.g., "models/components" -> "SDK::Models::Components"
// @ts-ignore
function getModuleFQN(outputLocation: string): string {
  if (outputLocation) {
    const module = sanitizeModuleName(context.Global.Config.Module);
    return `${module}::${getNamespaceModuleParts(outputLocation).join("::")}`;
  }
  return "";
}

registerTemplateFunc("getModuleFQN", getModuleFQN);

function getServerVariablesPath(): string {
  const sharedRoot = getScopeRoot("shared");
  return sharedRoot ? `${sharedRoot}/server_variables` : "server_variables";
}

function parentDir(path: string): string {
  return path.split("/").slice(0, -1).join("/");
}

// @ts-ignore
function getScopePath(scope: string): string {
  const imports = context.Global.Config.Imports;
  switch (scope.toString()) {
    case "errors":
      return imports.GetErrorsPath();
    case "shared":
      return imports.GetSharedPath();
    case "operations":
      return imports.GetOperationsPath();
    case "callbacks":
      return imports.GetCallbacksPath();
    case "webhooks":
      return imports.GetWebhooksPath();
    case "serverVariables":
      return getServerVariablesPath();
    default:
      throw new Error(`Unexpected scope: ${scope}`);
  }
}
registerTemplateFunc("getScopePath", getScopePath);

// @ts-ignore
function getScopeRoot(scope: string): string {
  return getScopePath(scope).split("/")[0];
}

// @ts-ignore
function getScopeName(scope: string): string {
  return caser().ToPascal(getScopePath(scope).split("/").pop());
}
registerTemplateFunc("getScopeName", getScopeName);

// @ts-ignore
function templateModuleDeclarations(
  outputLocation: string,
  indent: number,
): string {
  const parts = getNamespaceModuleParts(outputLocation);
  return parts
    .map((p, i) => "  ".repeat(indent + i) + `module ${p}`)
    .join("\n");
}
registerTemplateFunc("templateModuleDeclarations", templateModuleDeclarations);

// @ts-ignore
function templateModuleClosures(
  outputLocation: string,
  indent: number,
): string {
  const parts = getNamespaceModuleParts(outputLocation);
  return parts
    .map((_, i) => "  ".repeat(indent + parts.length - 1 - i) + "end")
    .join("\n");
}
registerTemplateFunc("templateModuleClosures", templateModuleClosures);

// Inserts an autoload entry into the module tree at the correct nesting level.
function insertAutoload(
  root: AutoloadNode,
  moduleParts: string[],
  autoloadPath: string,
): void {
  let node = root;
  for (let i = 0; i < moduleParts.length - 1; i++) {
    if (!node.children.has(moduleParts[i])) {
      node.children.set(moduleParts[i], {
        children: new Map(),
        autoloads: [],
      });
    }
    node = node.children.get(moduleParts[i])!;
  }
  const leaf = moduleParts[moduleParts.length - 1];
  node.autoloads.push({ name: leaf, path: autoloadPath });
}

interface AutoloadNode {
  children: Map<string, AutoloadNode>;
  autoloads: { name: string; path: string }[];
}

// @ts-ignore
function templateAllScopeAutoloads(auxPath: string, indent: number): string {
  const scopes = ["shared", "operations", "errors", "callbacks"];
  const bucketedPaths = new Set<string>();
  for (const [loc] of sequencedMapEntries(context.Global.AST.BucketedTypes)) {
    bucketedPaths.add(loc.toLowerCase());
  }
  if (bucketedPaths.has(getScopePath("webhooks").toLowerCase())) {
    scopes.push("webhooks");
  }
  if (
    context.Global.AST.MainSDK.Servers &&
    context.Global.AST.MainSDK.Servers.GetVariables().length > 0
  ) {
    scopes.push("serverVariables");
  }
  const scopePathSet = new Set<string>();
  const root: AutoloadNode = { children: new Map(), autoloads: [] };

  for (const scope of scopes) {
    const path = getScopePath(scope);
    scopePathSet.add(path.toLowerCase());
    const moduleParts = getNamespaceModuleParts(path);
    if (moduleParts.length === 0) continue;
    insertAutoload(root, moduleParts, `${auxPath}/${path}`);
  }

  // Add custom namespace autoloads into the same tree
  const seenModules = new Set<string>();

  for (const [outputLocation] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    const isCustomNamespace =
      outputLocation !== "" && !scopePathSet.has(outputLocation.toLowerCase());

    if (isCustomNamespace) {
      const parts = outputLocation.split("/");
      const moduleName = caser().ToPascal(parts[parts.length - 1]);
      if (!seenModules.has(moduleName)) {
        seenModules.add(moduleName);
        const sanitizedPath = parts
          .map((p) => caser().ToSnake(sanitizeName(p)))
          .join("/");
        const moduleParts = getNamespaceModuleParts(outputLocation);
        insertAutoload(root, moduleParts, `${auxPath}/${sanitizedPath}`);
      }
    }
  }

  // Render the tree
  const lines: string[] = [];
  function render(node: AutoloadNode, indent: number): void {
    for (const al of node.autoloads) {
      lines.push("  ".repeat(indent) + `autoload :${al.name}, '${al.path}'`);
    }
    for (const [moduleName, child] of node.children) {
      lines.push("  ".repeat(indent) + `module ${moduleName}`);
      render(child, indent + 1);
      lines.push("  ".repeat(indent) + "end");
    }
  }
  render(root, indent);
  return lines.join("\n");
}
registerTemplateFunc("templateAllScopeAutoloads", templateAllScopeAutoloads);
