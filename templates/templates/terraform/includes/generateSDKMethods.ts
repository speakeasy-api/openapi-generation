/**
 * Adds the Go SDK package import scope for a given TypeDef.
 */
function addImportScope(typedef: TypeDef) {
  const importScope = getImportScope(typedef);

  if (importScope) {
    addGenImport(importScope);
  }
}

/**
 * Returns the Go SDK package import path for the scope of a given TypeDef.
 * For example, shared scope returns the SDK package shared models path.
 */
function getImportScope(typeDef: TypeDef): string | undefined {
  if (typeDef.Scope && String(typeDef.Scope).length) {
    return `${getSDKPackage()}/models/${typeDef.Scope.toString()}`;
  }

  if (
    (typeDef.Type.toString() === "array" ||
      typeDef.Type.toString() === "set") &&
    typeDef.ItemType?.Scope &&
    String(typeDef.ItemType.Scope).length
  ) {
    return `${getSDKPackage()}/models/${typeDef.ItemType.Scope.toString()}`;
  }

  return undefined;
}

function generateSDKMethods(entity: TerraformEntity): string {
  const methods: Record<string, string> = {};
  const modelTypeDef = entity.SchemaTypeDef;
  const modelTypeName = entity.GoDataModelTypeName;

  // Emit all SDK request methods (To_ prefix) from the deduplicated
  // entity-level method list.
  for (const m of entity.SDKRequestMethods) {
    const sdkTypeDef = m.Target.TypeDef;
    const sdkTypeName = templateType(sdkTypeDef);

    methods[m.MethodName] = generateModelToSDKMethod(
      m.MethodName,
      modelTypeName,
      modelTypeDef,
      sdkTypeName,
      sdkTypeDef,
      entity,
      m.Operation,
    );
  }

  // Emit all SDK response methods (RefreshFrom_ / RefreshFromArrayOf_
  // prefix) from the deduplicated entity-level method list.
  for (const m of entity.SDKResponseMethods) {
    const sdkTypeDef = m.Target.TypeDef;
    const sdkTypeName = templateType(sdkTypeDef);

    methods[m.MethodName] = generateSDKToModelMethod(
      m.MethodName,
      modelTypeName,
      modelTypeDef,
      sdkTypeName,
      sdkTypeDef,
      m.Optional,
      entity,
      m.Operation,
    );
  }

  // In the future, all Go SDK operation response methods could be created here
  // which could encapsulate all the boilerplate SDK response handling (empty
  // responses, HTTP status codes, etc) logic that currently is in CRUD logic.

  let out = "";

  if (entity.IncludeSDKMethodOptions) {
    out +=
      `// ${entity.GoDataModelTypeName}Options enables patch sdk method construction.\n` +
      `type ${entity.GoDataModelTypeName}Options struct {\n` +
      `  Config *${entity.GoDataModelTypeName}\n` +
      `  State  *${entity.GoDataModelTypeName}\n` +
      `}\n\n`;
  }

  for (const methodName of Object.keys(methods).sort()) {
    const method = methods[methodName];
    if (method) {
      out += method + "\n\n";
    }
  }

  return out;
}
registerTemplateFunc("generateSDKMethods", generateSDKMethods);

/**
 * Templates a Go SDK response type to entity data model method. These methods
 * are prefixed with RefreshFrom.
 */
function generateSDKToModelMethod(
  methodName: string,
  modelTypeName: string,
  modelTypeDef: TypeDef,
  sdkTypeName: string,
  sdkTypeDef: TypeDef,
  sdkIsOptional: boolean,
  entity: TerraformEntity,
  op: TerraformOperation,
): string {
  // Must always be optional.
  sdkTypeName = sanitizeType(sdkTypeDef, true, "");

  const symbolManager = makeSymbolMananger();
  symbolManager["ctx"] = true;
  symbolManager["diags"] = true;
  symbolManager["r"] = true;
  symbolManager["resp"] = true;

  const methodBody = generateSDKToModelMethodBody(
    symbolManager,
    entity,
    "r",
    modelTypeDef,
    false,
    false,
    sdkIsOptional,
    "",
    "resp",
    sdkTypeDef,
    1,
    `${op.EntityOperation}`,
    op,
  ).trim();

  if (!methodBody) {
    return "";
  }

  addImportScope(sdkTypeDef);
  addGenImport("context");
  addGenImport("github.com/hashicorp/terraform-plugin-framework/diag");

  const result: string[] = [];

  result.push(
    `func (r *${modelTypeName}) ${methodName}(ctx context.Context, resp ${sdkTypeName}) diag.Diagnostics {`,
    `var diags diag.Diagnostics`,
    ``,
    methodBody,
    ``,
    `return diags`,
    `}`,
  );

  return result.join("\n");
}

/**
 * Templates an entity data model to Go SDK request type method. These methods
 * are prefixed with To.
 */
function generateModelToSDKMethod(
  methodName: string,
  modelTypeName: string,
  modelTypeDef: TypeDef,
  sdkTypeName: string,
  sdkTypeDef: TypeDef,
  entity: TerraformEntity,
  op: TerraformOperation,
): string {
  const symbolManager = makeSymbolMananger();
  symbolManager["ctx"] = true;
  symbolManager["diags"] = true;
  symbolManager["r"] = true;

  let extraArgument = "";
  if (entity.IncludeSDKMethodOptions) {
    extraArgument = `, opts *${entity.GoDataModelTypeName}Options`;
    symbolManager["opts"] = true;
  }

  const methodBody = generateModelToSDKMethodBody(
    symbolManager,
    "out",
    sdkTypeDef,
    modelTypeDef,
    false,
    false,
    false,
    modelTypeDef.Scope,
    true,
    "",
    "",
    "r",
    entity,
    1,
    `${op.EntityOperation}`,
    op,
  ).trim();

  if (!methodBody) {
    return "";
  }

  addImportScope(sdkTypeDef);
  addGenImport("context");
  addGenImport("github.com/hashicorp/terraform-plugin-framework/diag");

  const result: string[] = [];

  result.push(
    `func (r *${modelTypeName}) ${methodName}(ctx context.Context${extraArgument}) (*${sdkTypeName}, diag.Diagnostics) {`,
    `var diags diag.Diagnostics`,
    ``,
    methodBody,
    ``,
    `return &out, diags`,
    `}`,
  );

  return result.join("\n");
}
