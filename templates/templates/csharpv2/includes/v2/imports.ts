// @ts-ignore
function templateSDKImports(sdk: SDK, main: boolean): string {
  let imports = new Set<string>();

  imports = addImport(imports, "System", false);
  imports = addImport(imports, "Utils", true);
  imports = addImport(imports, "Generators.Enum", true);
  imports = addImport(imports, getScopeNamespace("errors", true), false);
  imports = addImport(imports, "System.Text.Json", false);
  imports = addImport(imports, "System.Text.Json.Serialization", false);

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
    }

    if (op.Response.Responses.length > 0) {
      imports = addImport(imports, "System.Collections.Generic", false);

      for (var response of op.Response.Responses) {
        if (response.Headers) {
          imports = addImport(imports, "Utils", true);
        }

        for (var content of response.Content) {
          if (content.SerializationMethod == "json") {
            imports = getTypeImports(imports, content.Content.Type, "", false);
            continue;
          }
          if (content.SerializationMethod == "eventstream") {
            imports = addImport(imports, "System.Collections.Generic", false);
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

unregisterTemplateFunc("templateSDKImports");
registerTemplateFunc("templateSDKImports", templateSDKImports);
