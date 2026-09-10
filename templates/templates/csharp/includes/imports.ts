// @ts-ignore
function addImport(
  imports: Set<string>,
  imp: string,
  local: boolean,
): Set<string> {
  if (local) {
    const namespace = getSDKNamespace();
    imp = imp != "" ? `${namespace}.${imp}` : namespace;
  }

  imports.add(imp);
  return imports;
}

// @ts-ignore
function addUsageImport(namespace: string, local: boolean = false): string {
  const imports: Set<string> =
    context.RecursiveComputed?.Imports || new Set<string>();

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }
  context.RecursiveComputed.Imports = addImport(imports, namespace, local);

  return "";
}
registerTemplateFunc("addUsageImport", addUsageImport);

// @ts-ignore
function addImportForType(imports: Set<string>, type: TypeDef): Set<string> {
  return addImport(
    imports,
    getModelNamespace(type.OutputLocation, false),
    true,
  );
}

// @ts-ignore
function addUsageImportForType(type: TypeDef): string {
  return addUsageImport(getModelNamespace(type.OutputLocation, false), true);
}

registerTemplateFunc("addUsageImportForType", addUsageImportForType);

// @ts-ignore
function templateImports(imports: Set<string>, indent: number): string {
  return Array.from(imports)
    .sort()
    .map((i) => {
      return `${templateIndent(indent)}using ${i};`;
    })
    .join("\n");
}
registerTemplateFunc("templateImports", templateImports);

// @ts-ignore
function templateModelImports(model: CSharpModel): string {
  let imports = new Set<string>();

  imports = addImport(imports, "Utils", true);

  if (model.Type.Type == "error") {
    imports = addImport(imports, "System", false);
    imports = addImport(imports, "System.Net.Http", false);
    // Only import errors namespace if the error type is not already in it
    const errorsNs = getScopeNamespace("errors", true);
    const modelNs = getModelNamespace(model.Type.OutputLocation, true);
    if (errorsNs !== modelNs) {
      imports = addImport(imports, getScopeNamespace("errors"), true);
    }
  }
  if (model.Type.Type == "union") {
    imports = addImport(imports, "Newtonsoft.Json", false);
    imports = addImport(imports, "Newtonsoft.Json.Linq", false);
    imports = addImport(imports, "System", false);
    imports = addImport(imports, "System.Collections.Generic", false);
    imports = addImport(imports, "System.Numerics", false);
    imports = addImport(imports, "System.Reflection", false);
    imports = addImport(imports, "Utils", true);
  }
  if (model.Type.Extensions.Pagination) {
    imports = addImport(imports, "System.Threading.Tasks", false);
    imports = addImport(imports, "System", false);
  }

  for (const fieldDef of model.Type.Fields) {
    imports = getTypeImports(imports, fieldDef.Type, "", false);

    if (fieldDef.Type.Type == "union") {
      imports = addImport(imports, getScopeNamespace("shared"), true);
    }

    if (containsSpeakeasyMetadataAnnotation(fieldDef)) {
      imports = addImport(imports, "Utils", true);
    }

    if (fieldDef.Type.Type == "response") {
      imports = addImport(imports, "System", false);
    }

    for (const anno of fieldDef.Annotations) {
      if (anno.IsType("json")) {
        imports = addImport(imports, "Newtonsoft.Json", false);
      }
    }
  }
  // Add imports for event stream models
  if (getEventStreamField(model.Type)) {
    imports = addImport(imports, "Utils.Sse", true);
    imports = addImport(imports, "System.Threading.Tasks", false);
  }

  for (const associated of model.Type.AssociatedTypes) {
    imports = getTypeImports(imports, associated, "", false);
  }

  if (model.Type.Comments?.Deprecated) {
    imports = addImport(imports, "System", false);
  }

  imports = getTypeImports(imports, model.Type, model.OutputLocation, true);

  if (imports.size == 0) {
    return "\n";
  }

  return `\n${templateImports(imports, 1)}\n`;
}
registerTemplateFunc("templateModelImports", templateModelImports);

// @ts-ignore
function getTypeImports(
  imports: Set<string>,
  typeDef: TypeDef,
  usageLocation: string,
  topLevel: boolean,
  visited: Record<string, boolean> = {},
): Set<string> {
  if (typeDef.IsCustomType()) {
    if (typeDef.Truncated || visited[getUniqueID(typeDef)]) {
      return imports;
    }

    visited[getUniqueID(typeDef)] = true;
  }

  switch (typeDef.Type.toString()) {
    case "enum":
      if (typeDef.OutputLocation != usageLocation && !topLevel) {
        imports = addImportForType(imports, typeDef);
      }

      if (
        (topLevel && typeDef.Enum.Type.Type == "string") ||
        typeDef.Enum.Open
      ) {
        imports = addImport(imports, "Newtonsoft.Json", false);
        imports = addImport(imports, "System", false);
      }

      if (typeDef.Enum.Open) {
        imports = addImport(imports, "System.Collections.Generic", false);
        imports = addImport(imports, "System.Collections.Concurrent", false);
        imports = addImport(imports, "System.Linq", false);
      }
      break;
    case "class":
    case "union":
      if (typeDef.OutputLocation != usageLocation && !topLevel) {
        imports = addImportForType(imports, typeDef);
      }

      if (topLevel) {
        for (const fieldDef of typeDef.Fields) {
          if (fieldDef.Comments?.Deprecated) {
            imports = addImport(imports, "System", false);
          }

          if (
            fieldDef.Annotations?.Has("json") &&
            needsCustomJsonConverter(fieldDef)
          ) {
            imports = addImport(imports, "Utils", true);
          }

          imports = getTypeImports(
            imports,
            fieldDef.Type,
            usageLocation,
            false,
            visited,
          );
        }
      }
      break;
    case "error":
      imports = addImport(imports, "System", false);
      if (typeDef.OutputLocation != usageLocation && !topLevel) {
        imports = addImportForType(imports, typeDef);
      }

      if (topLevel) {
        for (const fieldDef of typeDef.Fields) {
          if (fieldDef.Comments?.Deprecated) {
            imports = addImport(imports, "System", false);
          }

          if (
            fieldDef.Annotations?.Has("json") &&
            needsCustomJsonConverter(fieldDef)
          ) {
            imports = addImport(imports, "Utils", true);
          }

          imports = getTypeImports(
            imports,
            fieldDef.Type,
            usageLocation,
            false,
            visited,
          );
        }
      }
      break;
    case "array":
    case "map":
      imports = addImport(imports, "System.Collections.Generic", false);

      imports = getTypeImports(
        imports,
        typeDef.ItemType,
        usageLocation,
        false,
        visited,
      );
      break;
    case "date-time":
      imports = addImport(imports, "System", false);
      break;
    case "date":
      if (useNodatime()) {
        imports = addImport(imports, "NodaTime", false);
      } else {
        imports = addImport(imports, "System", false);
      }
      break;
    case "response":
      imports = addImport(imports, "System.Net.Http", false);
      break;
    case "response-stream":
      imports = addImport(imports, "Utils", true);
      break;
    case "bigint":
      imports = addImport(imports, "System.Numerics", false);
      break;
    case "event-stream":
      imports = addImport(imports, "Utils.Sse", true);
      imports = addImport(imports, "System.Collections.Generic", false);
      imports = getTypeImports(
        imports,
        typeDef.ItemType,
        usageLocation,
        false,
        visited,
      );
      break;
  }

  return imports;
}

// @ts-ignore
function templateSDKImports(sdk: SDK, main: boolean): string {
  let imports = new Set<string>();

  imports = addImport(imports, "System", false);
  imports = addImport(imports, "Utils", true);
  imports = addImport(imports, getScopeNamespace("errors"), true);

  if (main || sdk.Operations.length > 0) {
    imports = addImport(imports, "System.Threading.Tasks", false);
    if (context.Global.Config.EnableCancellationToken) {
      imports = addImport(imports, "System.Threading", false);
    }
    imports = addImport(imports, "System.Net.Http", false);
    imports = addImport(imports, getScopeNamespace("hooks"), true);
    imports = addImport(imports, "Utils.Retries", true);
  }

  if (sdkHasGlobalSecurity()) {
    imports = addImportForType(
      imports,
      context.Global.AST.MainSDK.Security.Type,
    );
  }

  if (sdk.Servers) {
    imports = addImport(imports, "System.Collections.Generic", false);
    imports = addImport(imports, "Newtonsoft.Json", false);
  }

  for (var op of sdk.Operations) {
    if (op.Request) {
      if (operationParametersFlattened(op)) {
        imports = getTypeImports(imports, op.Request.Field.Type, "", true);
      }

      imports = getTypeImports(imports, op.Request.Field.Type, "", false);

      imports = addImport(imports, "System.Net.Http.Headers", false);
    }

    if (op.Security) {
      imports = getTypeImports(imports, op.Security.Type, "", false);
    }

    if (op.Response.Type) {
      imports = getTypeImports(imports, op.Response.Type, "", false);
    }

    if (op.Extensions.Pagination) {
      imports = addImport(imports, "System.Linq", false);
      imports = addImport(imports, "Newtonsoft.Json.Linq", false);
    }

    if (op.Response.Responses.length > 0) {
      imports = addImport(imports, "System.Collections.Generic", false);

      for (var response of op.Response.Responses) {
        if (response.Headers) {
          imports = addImport(imports, "Utils", true);
        }

        for (var content of response.Content) {
          if (content.SerializationMethod == "json") {
            imports = addImport(imports, "Newtonsoft.Json", false);
            imports = getTypeImports(imports, content.Content.Type, "", false);
            continue;
          }
          if (content.SerializationMethod == "eventstream") {
            imports = addImport(imports, "Utils.Sse", true);
            imports = getTypeImports(
              imports,
              content.Content.Type.ItemType,
              "",
              false,
            );
            continue;
          }
        }
      }
    }

    if (op.Servers) {
      imports = addImport(imports, "System.Collections.Generic", false);
    }
  }

  return templateImports(imports, 1);
}

registerTemplateFunc("templateSDKImports", templateSDKImports);

// @ts-ignore
function templateSDKConfigImports(sdk: SDK): string {
  let imports = new Set<string>();

  imports = addImport(imports, "System", false);
  imports = addImport(imports, "System.Linq", false);
  imports = addImport(imports, "Utils", true);

  imports = addImport(imports, getScopeNamespace("hooks"), true);
  imports = addImport(imports, "Utils.Retries", true);

  if (sdkHasGlobalSecurity()) {
    imports = addImportForType(
      imports,
      context.Global.AST.MainSDK.Security.Type,
    );
  }

  if (sdk.Servers) {
    imports = addImport(imports, "System.Collections.Generic", false);
  }

  return templateImports(imports, 1);
}

registerTemplateFunc("templateSDKConfigImports", templateSDKConfigImports);

// @ts-ignore
function usingGlobalImports(): boolean {
  const modelsPath = getScopePath("models");

  for (const scope in context.Global.Config.Imports.Paths) {
    const path = context.Global.Config.Imports.Paths[scope];
    if (path !== "" && path !== modelsPath) {
      return false;
    }
  }

  return true;
}

// @ts-ignore
function templateUsageImports(_usageContext: UsageContext): string {
  let imports: Set<string> =
    context.RecursiveComputed?.Imports || new Set<string>();
  imports = addImport(imports, getSDKNamespace(), false);

  return templateImports(imports, 0);
}

registerTemplateFunc("templateUsageImports", templateUsageImports);

function genUsageImports(): string {
  return "{{ templateUsageImports .Local }}";
}

registerTemplateFunc("genUsageImports", genUsageImports);
