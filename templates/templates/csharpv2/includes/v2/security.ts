// @ts-ignore
function templateApplyOpSecurity(op: Operation): string {
  const typeName = sanitizeClassName(op.Security.Type.Name);
  const ns = getModelNamespace(op.Security.Type.OutputLocation, true);
  return `global::${ns}.${typeName}Applier.Apply(() => security, httpRequest)`;
}

unregisterTemplateFunc("templateApplyOpSecurity");
registerTemplateFunc("templateApplyOpSecurity", templateApplyOpSecurity);

// @ts-ignore
function templateApplyGlobalSecurity(op: Operation): string {
  const secType = context.Global.AST.MainSDK.Security.Type;
  const typeName = sanitizeClassName(secType.Name);
  const ns = getModelNamespace(secType.OutputLocation, true);

  const hoisted = op.HoistedSecurityConfig;
  const allowedArg =
    hoisted && !hoisted.Equivalent
      ? `, new System.Collections.Generic.List<string> { ${hoisted.Fields.map(
          (f) => `"${sanitizeFieldName(f.Name)}"`,
        ).join(", ")} }`
      : "";

  return `
if (SDKConfiguration.SecuritySource != null)
{
    httpRequest = global::${ns}.${typeName}Applier.Apply(SDKConfiguration.SecuritySource, httpRequest${allowedArg});
}`;
}

unregisterTemplateFunc("templateApplyGlobalSecurity");
registerTemplateFunc(
  "templateApplyGlobalSecurity",
  templateApplyGlobalSecurity,
);

// @ts-ignore
function getHTTPMetadata(): string {
  // In v2, `using Models.Shared;` brings HTTPMetadata directly into scope — no namespace prefix needed.
  return "HTTPMetadata";
}

unregisterTemplateFunc("getHTTPMetadata");
registerTemplateFunc("getHTTPMetadata", getHTTPMetadata);
