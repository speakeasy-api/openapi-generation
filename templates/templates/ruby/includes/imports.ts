type RubyAutoloads = { module: string; path: string }[];
type RubyImports = { pkg: string; relative: boolean }[];
type Models = Set<string>;

function getGemImports(): string {
  const importPath = caser().ToSnake(context.Global.Config.Module);
  let imports: RubyImports = [];

  // Hardcoded imports
  imports = addImport(imports, `${importPath}/utils/utils`, true);
  imports = addImport(imports, `${importPath}/utils/request_bodies`, true);
  imports = addImport(imports, `${importPath}/utils/query_params`, true);
  imports = addImport(imports, `${importPath}/utils/forms`, true);
  imports = addImport(imports, `${importPath}/utils/headers`, true);
  imports = addImport(imports, `${importPath}/utils/url`, true);
  imports = addImport(imports, `${importPath}/utils/security`, true);
  imports = addImport(imports, `crystalline`, true);
  imports = addImport(imports, `${importPath}/sdkconfiguration`, true);

  if (isFeatureUsed("serverEvents")) {
    imports = addImport(imports, `${importPath}/utils/eventstreaming`, true);
  }
  if (isFeatureUsed("jsonlResponses")) {
    imports = addImport(imports, `${importPath}/utils/jsonl`, true);
  }

  return templateImports(imports);
}

function getSubSDKAutoloads(sdk: SDK, indentation: number): RubyAutoloads {
  const importPath = caser().ToSnake(context.Global.Config.Module);

  let autoloads: RubyAutoloads = [];
  autoloads.push({
    module: sanitizeClassName(sdk.Type.Name),
    path: `${importPath}/${sanitizeFileName(sdk.Type.Name)}`,
  });
  for (const subSDK of sdk.SubSDKs) {
    autoloads = [...autoloads, ...getSubSDKAutoloads(subSDK, indentation + 1)];
  }
  return autoloads;
}
function getGemAutoloads(sdk: SDK, indentation: number): string {
  let autoloads: RubyAutoloads = [];
  for (const subSDK of sdk.SubSDKs) {
    autoloads = [...autoloads, ...getSubSDKAutoloads(subSDK, indentation)];
  }
  return templateAutoloads(autoloads, indentation);
}

function getModuleAutoloads(
  pathString: string,
  moduleName: string,
  contents: Models,
  indentation: number,
): string {
  let autoloads: RubyAutoloads = [];
  for (const modelName of contents) {
    autoloads.push({
      module: sanitizeClassName(modelName),
      path: `${pathString}${sanitizeFileName(
        caser().ToSnake(moduleName),
      )}/${sanitizeFileName(modelName)}.rb`,
    });
  }
  return templateAutoloads(autoloads, indentation);
}

// @ts-ignore
function getSDKImports(sdk?: SDK): string {
  let imports: RubyImports = [];

  imports = addImport(imports, "faraday");
  imports = addImport(imports, "faraday/multipart");
  imports = addImport(imports, "faraday/retry");
  if (context.Global.Config.TypingStrategy === "sorbet") {
    imports = addImport(imports, "sorbet-runtime");
  }
  const pagination =
    sdk?.Operations?.some((o) => o.Extensions.Pagination) || false;
  if (pagination) {
    imports = addImport(imports, "janeway");
  }
  imports = addImport(imports, "sdk_hooks/hooks", true);
  imports = addImport(imports, "utils/retries", true);

  return templateImports(imports);
}

// @ts-ignore
function getModelImports(model): string {
  let imports: RubyImports = [];

  if (context.Global.Config.TypingStrategy === "sorbet") {
    imports = addImport(imports, "sorbet-runtime");
  }
  imports = addImport(imports, "faraday");
  for (const t of model.Types) {
    const tt: TypeDef = t.Type;
    for (const field of tt.Fields) {
      if (
        field.Type.Name &&
        !model.Types.map((t) => t.Type.Name).includes(field.Type.Name)
      ) {
        if (model.Scope === field.Type.Scope) {
          imports = addImport(
            imports,
            `./${sanitizeFileName(field.Type.Name)}`,
            true,
          );
        } else {
          if (field.Annotations.Has("security")) {
            imports = addImport(
              imports,
              `../${field.Type.Scope}/security`,
              true,
            );
          } else {
            let fileName = sanitizeFileName(field.Type.Name);
            if (field.Type.Input) {
              const matches = fileName.match(/(.*)_input$/i);
              if (matches && matches.length > 1) {
                fileName = matches[1];
              }
            }
            if (field.Type.Output) {
              const matches = fileName.match(/(.*)_output$/i);
              if (matches && matches.length > 1) {
                fileName = matches[1];
              }
            }
            if (field.Type.Scope.toString().length > 0) {
              imports = addImport(
                imports,
                `../${field.Type.Scope}/${fileName}`,
                true,
              );
            }
          }
        }
      }
    }
  }
  const seen = new Set<string>();
  imports = imports.filter((i) => {
    if (seen.has(i["pkg"])) return false;
    seen.add(i["pkg"]);
    return true;
  });
  return templateImports(imports);
}

function getUsageTypeImports(modelName: string, typeDef: TypeDef): RubyImports {
  let imports: RubyImports = [];

  if (!typeDef) {
    return imports;
  }

  switch (typeDef.Type.toString()) {
    case "date":
      imports = addImport(imports, "date", false);
      break;
    case "date-time":
      imports = addImport(imports, "date", false);
      break;
    case "any":
      break;
    case "response":
      break;
  }

  return imports;
}

function templateUsageImports(operation: Operation): string {
  let imports: RubyImports = [];

  if (operation) {
    try {
      const types = getOperationModels(operation, true, true);

      for (const [, models] of sequencedMapEntries(types)) {
        for (const [model, types] of sequencedMapEntries(models)) {
          for (const type of types) {
            imports = mergeImports(imports, getUsageTypeImports(model, type));
          }
        }
      }
    } catch (_err) {
      // Keep README generation resilient for edge-case operations where
      // operation model extraction fails.
      return templateImports(imports);
    }

    imports.sort((a, b) => a.pkg.localeCompare(b.pkg));
  }

  return templateImports(imports);
}

function templateImports(imports: RubyImports): string {
  const lines = [];
  for (const imp of imports) {
    if ("relative" in imp && imp["relative"]) {
      lines.push(`require_relative '${imp["pkg"]}'`);
    } else {
      lines.push(`require '${imp["pkg"]}'`);
    }
  }

  return lines.join("\n");
}

function templateAutoloads(
  autoloads: RubyAutoloads,
  indentation: number,
): string {
  const lines = [];
  for (const autoload of autoloads) {
    lines.push(
      `${"  ".repeat(indentation)}autoload :${autoload["module"]}, '${
        autoload["path"]
      }'`,
    );
  }
  return lines.join("\n");
}

function addImport(
  imports: RubyImports,
  pkg: string,
  relative: boolean = false,
): RubyImports {
  const toRequire = { pkg: pkg, relative: relative };
  imports.push(toRequire);
  return imports;
}

function mergeImports(
  imports: RubyImports,
  newImports: RubyImports,
): RubyImports {
  return imports.concat(newImports);
}

// genImports support for recursive template rendering (used by test files).
// Templates using `recurse 1` can accumulate imports during the first pass
// via addRecursiveImport, then render them on the second pass via genImports.
// @ts-ignore
function genImports(): string {
  return `{{ templateRecursiveImports .RecursiveComputed.Imports }}`;
}
registerTemplateFunc("genImports", genImports);

// @ts-ignore
function templateRecursiveImports(imports: RubyImports): string {
  if (!imports || imports.length === 0) {
    return "";
  }
  return templateImports(imports);
}
registerTemplateFunc("templateRecursiveImports", templateRecursiveImports);

// @ts-ignore
function addRecursiveImport(pkg: string, relative: boolean = false): string {
  context.RecursiveComputed ??= {};
  context.RecursiveComputed["Imports"] ??= [];

  const existing = context.RecursiveComputed["Imports"] as RubyImports;
  if (!existing.some((imp) => imp.pkg === pkg && imp.relative === relative)) {
    existing.push({ pkg, relative });
  }

  return "";
}
registerTemplateFunc("addRecursiveImport", addRecursiveImport);

// @ts-ignore
function addEventStreamingImport(): void {
  // No-op: streaming imports are handled conditionally by getGemImports
}
registerTemplateFunc("addEventStreamingImport", addEventStreamingImport);

// @ts-ignore
function addJsonlImport(): void {
  // No-op: streaming imports are handled conditionally by getGemImports
}
registerTemplateFunc("addJsonlImport", addJsonlImport);

registerTemplateFunc("getGemImports", getGemImports);
registerTemplateFunc("getGemAutoloads", getGemAutoloads);
registerTemplateFunc("getModuleAutoloads", getModuleAutoloads);
registerTemplateFunc("getSDKImports", getSDKImports);
registerTemplateFunc("getModelImports", getModelImports);
registerTemplateFunc("templateUsageImports", templateUsageImports);
