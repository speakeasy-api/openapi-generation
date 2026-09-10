function templateWebhookHandlerMethodIfNeeded({
  usageLocation,
  docGroup,
}: {
  usageLocation: string;
  docGroup: string;
}): string {
  const webhookOps = getWebhookOperations();
  if (webhookOps.length === 0) {
    return "";
  }

  const returnType = webhookHandlerReturnType(usageLocation);

  templateWebhookHandlerFunc();

  return templateString("validate-webhook-method.ts.stmpl", {
    WebhookHandlerReturnType: returnType,
    UsageLocation: usageLocation,
  });
}
registerTemplateFunc(
  "templateWebhookHandlerMethodIfNeeded",
  templateWebhookHandlerMethodIfNeeded,
);

function templateWebhookHandlerFunc() {
  const funcFilename = sanitizeWebhookHandlerFuncFilename();
  const funcPath = `src/funcs/${funcFilename}.ts`;
  templateFile("validate-webhook-func-source.ts.stmpl", funcPath, {});
}

function webhookHandlerInboundTypes(usageLocation: string) {
  const webhookOps = getWebhookOperations();
  if (!webhookOps.length) {
    return [];
  }

  assert(usageLocation, "usageLocation missing for webhookHandlerInboundTypes");

  const ret = webhookOps
    .filter((op) => op.Request?.Field.Type)
    .map((op) => op.Request?.Field.Type);
  return ret;
}
registerTemplateFunc("webhookHandlerInboundTypes", webhookHandlerInboundTypes);

function webhookHandlerOutboundTypes(usageLocation: string) {
  const mainSDK = context.Global.AST.MainSDK;
  const webhookOps = getWebhookOperations();

  return webhookOps
    .filter((op) => op.Request?.Field)
    .map((op) =>
      resolveOutbound({
        usageLocation,
        rootTypeDef: mainSDK.Type,
        typeDef: op.Request?.Field.Type,
        optional: op.Request?.Field.Optional,
        nullable: op.Request?.Field.Nullable,
      }),
    );
}

registerTemplateFunc(
  "webhookHandlerOutboundTypes",
  webhookHandlerOutboundTypes,
);

function webhookHandlerReturnType(usageLocation: string): string {
  const returnType = webhookHandlerOutboundTypes(usageLocation)
    .map((x) => x.inputType)
    .join(" | ");

  for (const x of webhookHandlerOutboundTypes(usageLocation)) {
    x.importInputTypes();
  }

  // Clean duplicate types eg if multiple are optional
  return [...new Set(returnType.split(" | "))].join(" | ");
}
registerTemplateFunc("webhookHandlerReturnType", webhookHandlerReturnType);
