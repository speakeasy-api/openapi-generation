// @ts-ignore
type securityParam = {
  Type: string;
  Name: string;
  Source: string;
  FlattenedFieldName: string;
  Optional: boolean;
};

// @ts-ignore
function getSecurityParam(sdk: SDK): securityParam {
  if (!sdk.Security) {
    return null;
  }

  if (
    sdk.Security.Type.Fields.length == 1 &&
    context.Global.Config.FlattenGlobalSecurity == true
  ) {
    const secField = sdk.Security.Type.Fields[0];
    const fieldName = sanitizeSecurityFieldName(secField.Name);
    const paramName = sanitizePrivateFieldName(fieldName);
    return {
      Type: sanitizeType(secField.Type, false, ""),
      Name: paramName,
      Source: `${paramName}Source`,
      FlattenedFieldName: fieldName,
      Optional: secField.Optional,
    };
  }

  return {
    Type: "Security",
    Name: "security",
    Source: "securitySource",
    FlattenedFieldName: "",
    Optional: sdk.Security.Type.Fields.every((f) => f.Optional),
  };
}

registerTemplateFunc("getSecurityParam", getSecurityParam);

// @ts-ignore
function getGlobalSecurity(
  options?: SecurityUsageContext | Example,
  additionalContext?: TemplateValueContext,
): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example);
    }
  }

  if (
    index === undefined &&
    example === undefined &&
    context.Global.AST.MainSDK.Security.Optional &&
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason ==
      "optional-scheme"
  ) {
    return "";
  }

  let securityFieldName = "security";
  const security = context.Global.AST.MainSDK.Security;

  if (
    security.Type.Fields.length == 1 &&
    context.Global.Config.FlattenGlobalSecurity == true
  ) {
    securityFieldName = sanitizeGlobalSecurityFieldName(
      security.Type.Fields[0].Name,
    );
  }

  return `${securityFieldName}: ${templateSecurityUsage(
    security.Type,
    1,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
    example,
    additionalContext,
  ).trim()}`;
}
