// Populate Security as postman auth
function createAuth(security?: FieldDef) {
  if (!security) return undefined;
  const securityAnno = security.Type.Fields[0]
    .Annotations[0] as SecurityAnnotation;
  const SecurityType = securityAnno.SecType.toLowerCase();
  const subType = securityAnno.SubType.toLowerCase();
  const SecurityValue = [];

  switch (SecurityType) {
    // API Key auth
    case "apikey":
      const AccessTokenField = security.Type.Fields.find(
        (f) => f.Name === "accessToken",
      );
      const AccessToken = AccessTokenField?.Annotations.Has("security")
        ? AccessTokenField?.Annotations.Get("security").FieldName
        : "Authorization";
      SecurityValue.push({
        key: "key",
        value: AccessToken,
        type: "string",
      });
      SecurityValue.push({
        key: "value",
        value: "{{apikey}}",
        type: "string",
      });

      break;

    // Basic Auth
    case "http":
      switch (subType) {
        case "basic":
          SecurityValue.push({
            key: "password",
            value: "{{password}}",
            type: "string",
          });
          SecurityValue.push({
            key: "username",
            value: "{{username}}",
            type: "string",
          });
          break;
        // Bearer Auth
        case "bearer":
          SecurityValue.push({
            key: "token",
            value: "{{bearerToken}}",
            type: "string",
          });
          break;
      }
      break;

    // OAuth2
    case "oauth2":
      const TokenURL = security.Type.Fields.find((f) => f.Name === "TokenURL");
      if (!TokenURL) {
        return undefined; // TokenURL not always available
      }

      SecurityValue.push({
        key: "accessTokenUrl",
        value: TokenURL.Const || TokenURL.Default || "",
        type: "string",
      });

      break;

    // OpenID Connect
    case "openidconnect":
      const OpenIDURL = security.Type.Fields.find(
        (f) => f.Name === "openIdConnect",
      );
      if (!OpenIDURL) {
        throw new Error("OpenIDURL not found in openIdConnect security");
      }
      SecurityValue.push({
        key: "accessTokenUrl",
        value: OpenIDURL.Const || OpenIDURL.Default || "",
        type: "string",
      });

      break;

    case "":
      return undefined;

    default:
      console.warn(`Unsupported security type: ${SecurityType}, skipping.`);
      return undefined;
  }

  return {
    type: SecurityType,
    [SecurityType]: SecurityValue,
  };
}
