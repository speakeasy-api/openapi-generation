// @ts-ignore
function templateModelImports(model: UnityModel): string {
  let imports = [];

  imports = addImport(imports, "System", false);
  imports = addImport(imports, "UnityEngine", false);

  if (model.Type.Type == "union") {
    imports = addImport(imports, "Newtonsoft.Json", false);
    imports = addImport(imports, "Newtonsoft.Json.Linq", false);
    imports = addImport(imports, "System", false);
    imports = addImport(imports, "System.Numerics", false);
    imports = addImport(imports, "Utils", true);
  }

  if (model.Type.Extensions.Pagination) {
    imports = addImport(imports, "System.Threading.Tasks", false);
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
  for (const associated of model.Type.AssociatedTypes) {
    imports = getTypeImports(imports, associated, "", false);
  }
  if (model.Type.Comments?.Deprecated) {
    imports = addImport(imports, "System", false);
  }

  imports = getTypeImports(imports, model.Type, model.OutputLocation, true);

  if (imports.length == 0) {
    return "";
  }

  return (
    "\n" +
    imports
      .sort()
      .map((namespace) => `    ${namespace}`)
      .join("\n")
  );
}
registerTemplateFunc("templateModelImports", templateModelImports);

// @ts-ignore
function getTypeImports(
  imports: string[],
  typeDef: TypeDef,
  usageLocation: string,
  topLevel: boolean,
  visited: Record<string, boolean> = {},
): string[] {
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

      if (topLevel && typeDef.Enum.Type.Type == "string") {
        imports = addImport(imports, "Newtonsoft.Json", false);
        imports = addImport(imports, "System", false);
      }
      break;
    case "class":
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

          if (needsCustomJsonConverter(fieldDef)) {
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
      imports = addImport(imports, "Utils", true);
      break;
    case "response":
      imports = addImport(imports, "UnityEngine.Networking", false);
      break;
    case "response-stream":
      imports = addImport(imports, "Utils", true);
      break;
    case "bigint":
      imports = addImport(imports, "System.Numerics", false);
      break;
    case "union":
      if (typeDef.OutputLocation != usageLocation && !topLevel) {
        imports = addImportForType(imports, typeDef);
      }
      break;
  }

  return imports;
}

// @ts-ignore
function templateSDKImports(sdk: SDK, main: boolean): string {
  var imports = [];

  imports = addImport(imports, "Utils", true);
  imports = addImport(imports, getScopeNamespace("errors"), true);
  imports = addImport(imports, "System.Threading.Tasks", false);
  imports = addImport(imports, "System.Text.RegularExpressions", false);
  imports = addImport(imports, "System", false);

  if (sdkHasGlobalSecurity()) {
    imports = addImportForType(
      imports,
      context.Global.AST.MainSDK.Security.Type,
    );
  }

  if (sdk.Servers) {
    imports = addImport(imports, "System.Collections.Generic", false);
  }

  if (sdk.Operations.length > 0) {
    imports = addImport(imports, "UnityEngine.Networking", false);
  }

  for (var op of sdk.Operations) {
    if (op.Request) {
      imports = getTypeImports(
        imports,
        op.Request.Field.Type,
        "",
        operationParametersFlattened(op),
      );
    }
    imports = getTypeImports(imports, op.Response.Type, "", false);

    if (op.Response.Responses.length > 0) {
      imports = addImport(imports, "System.Collections.Generic", false);

      for (var response of op.Response.Responses) {
        for (var content of response.Content) {
          if (content.SerializationMethod == "json") {
            imports = addImport(imports, "Newtonsoft.Json", false);
            if (op.Extensions.Pagination) {
              imports = addImport(imports, "System.Linq", false);
              imports = addImport(imports, "Newtonsoft.Json.Linq", false);
            }
            imports = getTypeImports(imports, content.Content.Type, "", false);
          }
        }
      }
    }

    if (op.Servers) {
      imports = addImport(imports, "System.Collections.Generic", false);
    }
  }

  return imports
    .sort()
    .map((namespace) => `    ${namespace}`)
    .join("\n");
}

registerTemplateFunc("templateSDKImports", templateSDKImports);

// @ts-ignore
function addImportForType(imports: string[], type: TypeDef): string[] {
  return addImport(
    imports,
    getModelNamespace(type.OutputLocation, false),
    true,
  );
}

// @ts-ignore
function usingGlobalImports(): boolean {
  let allImportsIntendedForGlobal = true;

  for (const scope in context.Global.Config.Imports.Paths) {
    const path = context.Global.Config.Imports.Paths[scope];

    if (path != "" && path != "Models") {
      allImportsIntendedForGlobal = false;
      break;
    }
  }

  return allImportsIntendedForGlobal;
}

// @ts-ignore
function getUsageTypeImports(
  imports: string[],
  typeDef: TypeDef,
  visited: Record<string, boolean> = {},
): string[] {
  if (typeDef.IsCustomType()) {
    if (typeDef.Truncated || visited[getUniqueID(typeDef)]) {
      return imports;
    }

    visited[getUniqueID(typeDef)] = true;
  }

  switch (typeDef.Type.toString()) {
    case "enum":
    case "class":
      imports = addImportForType(imports, typeDef);

      for (const field of typeDef.Fields) {
        imports = getUsageTypeImports(imports, field.Type, visited);
      }
      break;
    case "union":
      imports = addImportForType(imports, typeDef);
      break;
    case "array":
    case "map":
      imports = addImport(imports, "System.Collections.Generic", false);
      imports = getUsageTypeImports(imports, typeDef.ItemType, visited);
      break;
    case "date":
      imports = addImport(imports, "SDK.Utils", false);
      break;
    case "date-time":
      break;
  }

  return imports;
}

// @ts-ignore
function templateUsageImports(usageContext: UsageContext): string {
  const operation = usageContext.Operation;
  if (!operation) {
    return "";
  }

  const types = getOperationModels(operation, true, true);

  const rootNamespace = getSDKNamespace();
  let imports = [`using ${rootNamespace};`];

  if (context.Global.AST.MainSDK.Security && !operation.Security) {
    imports = addImportForType(
      imports,
      context.Global.AST.MainSDK.Security.Type,
    );
  }

  if (usageContext.RenderFeature("errors", false)) {
    imports = addImport(imports, "System", false);
    imports = addImport(imports, getScopeNamespace("errors"), true);
  }

  for (const [, models] of sequencedMapEntries(types)) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const type of types) {
        imports = getUsageTypeImports(imports, type);
      }
    }
  }

  return imports.join("\n");
}

registerTemplateFunc("templateUsageImports", templateUsageImports);

// @ts-ignore
function addImport(imports: string[], imp: string, local: boolean): string[] {
  const rootNamespace = getSDKNamespace();
  if (local) {
    if (imp != "") {
      imp = `${rootNamespace}.${imp}`;
    } else {
      imp = rootNamespace;
    }
  }

  imp = `using ${imp};`;

  if (!imports.includes(imp)) {
    imports.push(imp);
  }

  return imports;
}
