function* iterateWebhookOperations(
  sdk: SDK = context.Global.AST.MainSDK,
): Generator<Operation> {
  const queue = [sdk];
  while (queue.length) {
    const curr = queue.shift();
    for (const op of curr.Operations) {
      if (op.Webhook) {
        yield op;
      }
    }
    queue.push(...curr.SubSDKs);
  }
}

function getWebhookOperations(ast: AST = context.Global.AST): Operation[] {
  const webhooks: Operation[] = [];
  for (const [_, ops] of sequencedMapEntries(ast.Webhooks)) {
    webhooks.push(...ops);
  }
  return webhooks;
}
registerTemplateFunc("getWebhookOperations", getWebhookOperations);

function firstWebhookOperation(ast: AST = context.Global.AST) {
  return getWebhookOperations(ast)[0];
}
registerTemplateFunc("firstWebhookOperation", firstWebhookOperation);

function getWebhookSecurity(ast: AST = context.Global.AST): WebhookSecurity {
  return firstWebhookOperation(ast)?.Webhook.Security;
}
registerTemplateFunc("getWebhookSecurity", getWebhookSecurity);

function hasWebhookSecurity(ast: AST = context.Global.AST) {
  return Boolean(isFeatureUsed("webhooks") && getWebhookSecurity(ast));
}
registerTemplateFunc("hasWebhookSecurity", hasWebhookSecurity);

function hasCustomWebhookSecurity(ast: AST = context.Global.AST) {
  return Boolean(getWebhookSecurity(ast)?.Type === "custom");
}
registerTemplateFunc("hasCustomWebhookSecurity", hasCustomWebhookSecurity);

function shouldTemplateSecretForWebhookConsumer(ast: AST = context.Global.AST) {
  const security = getWebhookSecurity(ast);
  return Boolean(security && security.ConsumerShouldProvideSecret?.valueOf());
}
registerTemplateFunc(
  "shouldTemplateSecretForWebhookConsumer",
  shouldTemplateSecretForWebhookConsumer,
);
