type PHPModel = {
  Name: string;
  Type: TypeDef;
  Servers?: Servers;
  OutputLocation: string;
};

function getServiceProviderImports(): string {
  const imports = [];
  return templateImports(imports);
}

registerTemplateFunc("getServiceProviderImports", getServiceProviderImports);

function getSimpleTypeImports(type: TypeDef): string {
  switch (type.Type.toString()) {
    case "class":
    case "enum":
    case "error":
      // Don't add imports for custom namespaces - we use fully qualified names for them
      // Note: Check for truthy value because Go nil pointers become null (not undefined) in JS
      const isCustomNamespace = !!type.Extensions?.ModelNamespace;
      if (isCustomNamespace) {
        return "";
      }
      return `${getModelNamespace(type.OutputLocation)}`;
    case "date":
      return "Brick\\DateTime\\LocalDate";
  }
}

function getFieldImports(fieldDef: FieldDef): string[] {
  const imports = [];

  if (fieldDef.Type.Type.toString() == "union") {
    for (const t of fieldDef.Type.AssociatedTypes) {
      getTypeImports(imports, t, 0);
    }
  } else if (["array", "map"].includes(fieldDef.Type.Type.toString())) {
    if (fieldDef.Type.ItemType.Type.toString() !== "union") {
      getTypeImports(imports, fieldDef.Type.ItemType, 0);
    }
  } else {
    const imp = getSimpleTypeImports(fieldDef.Type);
    if (imp) {
      addImport(imports, imp);
    }
  }
  return imports;
}

//@ts-ignore
function templateModelImports(model: PHPModel): string {
  let imports = context.RecursiveComputed?.Imports || [];
  const localNamespace = getModelNamespace(model.OutputLocation);
  imports = imports.filter((i) => i !== localNamespace);
  for (const fieldDef of model.Type.Fields) {
    const fieldImports = getFieldImports(fieldDef).filter(
      (i) => i !== getModelNamespace(model.OutputLocation),
    );
    imports = mergeImports(imports, fieldImports);
    if (containsSpeakeasyMetadataAnnotation(fieldDef)) {
      imports = addImport(
        imports,
        `${context.Global.Config.Namespace}\\Utils\\SpeakeasyMetadata`,
      );
    }
  }

  return templateImports(imports);
}
registerTemplateFunc("templateModelImports", templateModelImports);

// @ts-ignore
// requires the RecursiveComputed to be populated, so delay this calculation
function getModelImports(): string {
  return "{{ templateModelImports .Local }}";
}
registerTemplateFunc("getModelImports", getModelImports);

// @ts-ignore
// requires the RecursiveComputed to be populated, so delay this calculation
function getSDKImports(): string {
  return "{{ templateSDKImports .Local }}";
}
registerTemplateFunc("getSDKImports", getSDKImports);

// @ts-ignore
function templateSDKImports(sdk: SDK): string {
  let imports = context.RecursiveComputed?.Imports || [];

  for (const operation of sdk.Operations) {
    if (getResponseFormat() == "envelope-http") {
      imports = addImport(imports, getScopeNamespace("shared", true));
    }
    if (operation.Request?.Field.Type) {
      if (operationParametersFlattened(operation)) {
        imports = getTypeImports(imports, operation.Request.Field.Type, 1);
      } else {
        imports = getTypeImports(imports, operation.Request.Field.Type, 0);
      }
    }

    if (operation.Security?.Type) {
      imports = getTypeImports(imports, operation.Security.Type, 0);
    }

    if (operation.Response?.Type) {
      imports = getTypeImports(imports, operation.Response.Type, 0);
    }
  }

  return templateImports(imports);
}
registerTemplateFunc("templateSDKImports", templateSDKImports);

// @ts-ignore
function getTypeImports(
  imports: string[],
  type: TypeDef,
  maxDepth: number = 1,
  visited: Record<string, boolean> = {},
  currentDepth: number = 0,
): string[] {
  let globalNamespace = context.Global.Config.Namespace;
  imports = mergeImports(imports, context.RecursiveComputed?.Imports || []);
  if (type.IsCustomType()) {
    if (type.Truncated || visited[getUniqueID(type)]) {
      return imports;
    }

    visited[getUniqueID(type)] = true;
  }

  switch (type.Type.toString()) {
    case "class":
    case "enum":
      const modelNamespace = getModelNamespace(type.OutputLocation);
      // Skip adding imports for custom namespaces - we use fully qualified names for them
      // Note: Check for truthy value because Go nil pointers become null (not undefined) in JS
      const isCustomNamespace = !!type.Extensions?.ModelNamespace;
      if (globalNamespace != modelNamespace && !isCustomNamespace) {
        imports = addImport(imports, modelNamespace);
      }

      if (currentDepth < maxDepth) {
        for (const fieldDef of type.Fields) {
          imports = getTypeImports(
            imports,
            fieldDef.Type,
            maxDepth,
            visited,
            currentDepth + 1,
          );
        }
      }

      break;
    case "array":
    case "map":
      imports = getTypeImports(
        imports,
        type.ItemType,
        maxDepth,
        visited,
        currentDepth + 1,
      );
      break;
    case "union":
      for (const t of type.AssociatedTypes) {
        imports = getTypeImports(
          imports,
          t,
          maxDepth,
          visited,
          currentDepth + 1,
        );
      }
      break;
    default:
      const imp = getSimpleTypeImports(type);
      if (imp) {
        addImport(imports, imp);
      }
      break;
  }

  return imports;
}

// @ts-ignore
function addImport(imports: string[], imp: string): string[] {
  if (!imports.includes(imp)) {
    imports.push(imp);
  }

  return imports;
}

// @ts-ignore
function mergeImports(imports: string[], imps: string[]): string[] {
  for (const imp of imps) {
    imports = addImport(imports, imp);
  }
  return imports;
}

// @ts-ignore
// requires the RecursiveComputed to be populated, so delay this calculation
function getUsageImports(): string {
  return "{{ templateUsageImports .Local }}";
}
registerTemplateFunc("getUsageImports", getUsageImports);

// @ts-ignore
function templateUsageImports(usageCtx: UsageContext): string {
  let namespace = context.Global.Config.Namespace;

  let imports = context.RecursiveComputed?.Imports || [];

  // force insertion of the project import
  imports = addImport(imports, `${namespace}`);

  return templateImports(imports);
}
registerTemplateFunc("templateUsageImports", templateUsageImports);

// @ts-ignore
function addTypeImportInline(type: TypeDef): string {
  let imports = context.RecursiveComputed?.Imports || [];
  imports = getTypeImports(imports, type);
}
// @ts-ignore
function addImportInline(imp: string, outputLocation: string = null): string {
  let imports = context.RecursiveComputed?.Imports || [];

  if (imp == "utils") {
    if (outputLocation != context.Global.Config.Namespace) {
      imp = `${context.Global.Config.Namespace}\\Utils`;
    } else {
      return "";
    }
  }

  imports = addImport(imports, imp);

  if (!context.RecursiveComputed) {
    context.RecursiveComputed = {};
  }

  context.RecursiveComputed.Imports = imports;
  return ""; // Still needs to return a string if used in templates
}

registerTemplateFunc("addImportInline", addImportInline);

// @ts-ignore
function genImports(): string {
  return `{{ templateImports .RecursiveComputed.Imports }}`;
}
registerTemplateFunc("genImports", genImports);

function genClientCredentialsImports(
  clientCredentialsAccess: ClientCredentialsSecurityAccess,
): string {
  let imports = [];
  imports = addImport(imports, "Brick\\DateTime\\LocalTime");
  imports = addImport(imports, "Brick\\DateTime\\TimeZone");
  imports = addImport(imports, "GuzzleHttp\\ClientInterface");
  imports = addImport(imports, "Psr\\Http\\Message\\RequestInterface");
  imports = addImport(imports, "Psr\\Http\\Message\\ResponseInterface");
  imports = addImport(imports, "Speakeasy\\Serializer\\DeserializationContext");
  imports = addImport(imports, `${getScopeNamespace("utils", true)}`);
  if (context.Global.Config.SDKHooksConfigAccess) {
    imports = addImport(imports, getSDKConfigurationImport());
  }

  for (const { securityType } of clientCredentialsAccess) {
    imports = getTypeImports(imports, securityType, 0);
  }

  return templateImports(imports);
}
registerTemplateFunc(
  "genClientCredentialsImports",
  genClientCredentialsImports,
);

function getSDKConfigurationImport(): string {
  return `${getAccessNamespace("", true, false)}\\SDKConfiguration`;
}

function addSDKConfigurationImport(): string {
  return addImportInline(getSDKConfigurationImport());
}
registerTemplateFunc("addSDKConfigurationImport", addSDKConfigurationImport);

// @ts-ignore
function templateImports(imports: string[]): string {
  if (!imports || imports.length == 0) {
    return "";
  }

  return `${imports
    .sort(function (a, b) {
      a = a.toLowerCase();
      b = b.toLowerCase();
      if (a == b) return 0;
      if (a > b) return 1;
      return -1;
    })
    .map((i) => `use ${i};`)
    .join("\n")}`;
}
registerTemplateFunc("templateImports", templateImports);
