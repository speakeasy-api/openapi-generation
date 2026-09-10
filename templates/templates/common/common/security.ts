// A "built-in" flow is an OAuth2 flow that is natively
// supported in Speakeasy SDKs via a predefined hook.
// Currently, only one of these flows can be enabled at a time.
const OAuth2BuiltInFlows: OAuth2BuiltInFlow[] = [
  "client_credentials",
  "password",
];

type ClientCredentialsSecurity = {
  id: string;
  nameId: string;
  securityType: TypeDef;
  optional: boolean;
};
type ClientCredentialsSecurityAccess = Array<ClientCredentialsSecurity>;

type OAuth2PasswordSecurity = {
  id: string;
  nameId: string;
  securityType: TypeDef;
  optional: boolean;
};
type OAuth2PasswordSecurityAccess = Array<OAuth2PasswordSecurity>;

//@ts-ignore
function getGlobalSecurityFields(
  security = context.Global.AST.MainSDK.Security,
): FieldDef[] {
  if (!security) {
    return [];
  }

  return security.Type.Fields.filter((f) => f.Annotations.Get("security"));
}
registerTemplateFunc("getGlobalSecurityFields", getGlobalSecurityFields);

//@ts-ignore
function getSecurityFields(securityType: TypeDef): FieldDef[] {
  if (!securityType) {
    return [];
  }

  return securityType.Fields.filter((f) => f.Annotations.Get("security"));
}

function opHasOAuth2Flow(op: Operation, flow: OAuth2Flow): boolean {
  if (!op.Security) {
    return false;
  }

  return getOAuth2FlowFields(flow, op.Security.Type.Fields).length > 0;
}

function hasOAuth2Flow(flow: OAuth2Flow): boolean {
  const security = context.Global.AST.MainSDK.Security;

  if (!security || !accountHasFeatureAccess("sdkHooks")) {
    return false;
  }

  if (getOAuth2FlowFields(flow, security.Type.Fields).length > 0) {
    return true;
  }

  const checkOps = (sdk: SDK) => {
    return sdk.Operations.some((op) => opHasOAuth2Flow(op, flow));
  };

  const mainSDK = context.Global.AST.MainSDK;
  return checkOps(mainSDK) || mainSDK.SubSDKs.some(checkOps);
}

function hasOAuth2BuiltInFlow(flow: OAuth2BuiltInFlow): boolean {
  return hasOAuth2Flow(flow);
}

function hasClientCredentials(): boolean {
  return hasOAuth2Flow("client_credentials");
}
registerTemplateFunc("hasClientCredentials", hasClientCredentials);

function hasOAuth2PasswordFlow(): boolean {
  return hasOAuth2Flow("password");
}
registerTemplateFunc("hasOAuth2PasswordFlow", hasOAuth2PasswordFlow);

/**
 * Search a security field for any OAuth2 scheme child fields that match the
 * given flow and returns them.
 */
//@ts-ignore
function getOAuth2FlowFields(
  flow: OAuth2Flow,
  securityFields: FieldDef[],
  parts: FieldDef[] = [],
): FieldDef[] {
  for (const field of securityFields) {
    if (isOAuth2FlowField(field, flow)) {
      return [...parts, field];
    } else if (isSecurityOption(field)) {
      const _parts = getOAuth2FlowFields(flow, field.Type.Fields, [
        ...parts,
        field,
      ]);
      if (_parts.length > parts.length) {
        return _parts;
      }
    }
  }

  return [];
}

function getOperationOAuth2FlowFields(
  op: Operation,
  flow: OAuth2Flow,
): FieldDef[] {
  let security = context.Global.AST.MainSDK.Security;
  if (op.Security) {
    security = op.Security;
  }

  if (!security) {
    return [];
  }

  return getOAuth2FlowFields(flow, getSecurityFields(security.Type));
}

/**
 * This is a convenience function for getting one of the allowed variations of
 * an OAuth2 security scheme. This is relevant when a security scheme can either
 * take an access token directly or a credentials object for obtaining access
 * tokens. In this case, the security scheme is expressed as a union.
 */
function getOAuth2UnionMember(
  securityField: FieldDef,
  member: "token" | "credentials",
): TypeDef {
  const ann = securityField.Annotations.Get("security") as SecurityAnnotation;
  if (ann?.SecType !== "oauth2") {
    throw new Error(
      `${securityField.Name}: security field must have a security annotation with type oauth2`,
    );
  }

  if (securityField.Type.Type.toString() !== "union") {
    throw new Error(
      `${securityField.Name}: security field type must be a union of an access token and a credentials object`,
    );
  }

  let originalName = "";
  if (member === "token") {
    originalName = "Token";
  } else if (member === "credentials") {
    originalName = "Credentials";
  } else {
    throw new Error(
      `invalid union member name for oauth2 security field: ${member}`,
    );
  }

  for (const t of securityField.Type.AssociatedTypes) {
    if (t.OriginalName === originalName) {
      return t;
    }
  }

  throw new Error(
    `${securityField.Name}: security field union must contain a ${member} member`,
  );
}

//@ts-ignore
function isOAuth2FlowField(field: FieldDef, flow: OAuth2Flow): boolean {
  const ann = field.Annotations.Get("security") as SecurityAnnotation;
  const type = ann?.SecType || "";
  const sub = ann?.SubType || "";

  return `${type}:${sub}` === `oauth2:${flow}`;
}

//@ts-ignore
function isSecurityOption(field: FieldDef): boolean {
  const secAnno = field.Annotations.Get("security") as SecurityAnnotation;

  return secAnno?.SecurityOption;
}

function getNumberOfSecurityOptions(securityFields: FieldDef[]): number {
  return securityFields.filter((f) => isSecurityOption(f)).length;
}

// @ts-ignore
function hasClientCredentialsOnly(securityFields?: FieldDef[]): boolean {
  return hasOneOAuth2GrantTypeOnly("client_credentials", securityFields);
}
registerTemplateFunc("hasClientCredentialsOnly", hasClientCredentialsOnly);

// @ts-ignore
function isClientCredentialsFlattened(operation?: Operation): boolean {
  if (operation !== undefined) {
    return hasClientCredentialsOnly(getSecurityFields(operation.Security.Type));
  }

  return hasClientCredentialsOnly(getGlobalSecurityFields());
}
registerTemplateFunc(
  "isClientCredentialsFlattened",
  isClientCredentialsFlattened,
);

function hasOneOAuth2GrantTypeOnly(
  flow: OAuth2Flow,
  securityFields?: FieldDef[],
): boolean {
  if (!securityFields) {
    securityFields = getGlobalSecurityFields();
  }

  if (
    securityFields.every((f) => {
      return !isOAuth2FlowField(f, flow);
    })
  ) {
    return false;
  }

  const hasMultipleOptions = getNumberOfSecurityOptions(securityFields) > 1;

  const firstFieldAnno = securityFields[0].Annotations.Get(
    "security",
  ) as SecurityAnnotation;

  const isOneOfMultipleOptions =
    !securityFields.every((f) => {
      const secAnno = f.Annotations.Get("security") as SecurityAnnotation;
      return (
        secAnno?.SecType === firstFieldAnno?.SecType &&
        secAnno?.SubType === firstFieldAnno?.SubType
      );
    }) && securityFields.length > 1;

  return !hasMultipleOptions && !isOneOfMultipleOptions;
}

/**
 * Given a security field containing an OAuth2 flow, this function returns an
 * object containing the environment variables that should be used to populate
 * the security scheme config.
 */
function mapOAuth2EnvVariables(
  securityFieldDef: FieldDef,
  flow: OAuth2Flow,
): {
  /** The environment variable holding an OAuth2 access token. */
  token: string;
  /** The environment variables for the credentials fields that can be used to
   * obtain an OAuth2 access token.
   */
  credentials: Record<string, string>;
} {
  if (!securityFieldDef) {
    throw new Error("No security field definition specified");
  }

  const secfields = securityFieldDef.Annotations?.Has("security")
    ? [securityFieldDef]
    : getGlobalSecurityFields(securityFieldDef);
  const fields = getOAuth2FlowFields(flow, secfields);
  if (!fields?.length) {
    throw new Error(`No ${flow} fields found in global security`);
  }

  if (fields.length > 1) {
    throw new Error(
      `${fields.length} ${flow} fields found in global security but expected only one`,
    );
  }

  const schemeField = fields[0];

  const tokenType = getOAuth2UnionMember(schemeField, "token");
  const tokenField = typeDefToFieldDef(tokenType);
  tokenField.Name = "Token";
  const tokenVar = templateSecurityEnvVars(tokenField);

  const credType = getOAuth2UnionMember(schemeField, "credentials");
  if (!credType) {
    throw new Error(`No credentials type found in ${flow} security`);
  }
  const credDataType = credType.Type.toString();
  if (credDataType !== "class") {
    throw new Error(
      `Unexpected credentials data type found in ${flow} security (got: ${credDataType}, want: class).`,
    );
  }

  const credentialEnvVariables: Record<string, string> = {};
  for (const field of credType.Fields) {
    credentialEnvVariables[field.Name] = templateSecurityEnvVars(field);
  }

  return {
    token: tokenVar,
    credentials: credentialEnvVariables,
  };
}
registerTemplateFunc("mapOAuth2EnvVariables", mapOAuth2EnvVariables);

// @ts-ignore
function getClientCredentialsSecurityAccess(): ClientCredentialsSecurityAccess {
  const globalSecurity = context.Global.AST.MainSDK.Security;
  const clientCredentialsTypes: ClientCredentialsSecurityAccess = [];
  const seenIds = new Set<string>();
  const flow = "client_credentials";

  if (
    globalSecurity?.Type &&
    getOAuth2FlowFields(flow, getSecurityFields(globalSecurity.Type)).length > 0
  ) {
    clientCredentialsTypes.push({
      id: "",
      nameId: "",
      securityType: globalSecurity.Type,
      optional: globalSecurity.Optional,
    });
    seenIds.add("");
  }

  const getOpSecurityType = (sdk: SDK) => {
    for (const op of sdk.Operations) {
      if (!op.Security) {
        continue;
      }

      const opId = op.ID;
      if (seenIds.has(opId)) {
        continue;
      }

      const fields = getOAuth2FlowFields(
        flow,
        getSecurityFields(op.Security.Type),
      );

      if (fields.length > 0) {
        seenIds.add(opId);
        clientCredentialsTypes.push({
          id: opId,
          nameId: op.GetID(),
          securityType: op.Security.Type,
          optional: op.Security.Optional,
        });
      }
    }

    for (const subSDK of sdk.SubSDKs) {
      getOpSecurityType(subSDK);
    }
  };

  getOpSecurityType(context.Global.AST.MainSDK);

  if (clientCredentialsTypes.length == 0) {
    throw new Error("No client credentials security found");
  }

  return clientCredentialsTypes;
}
registerTemplateFunc(
  "getClientCredentialsSecurityAccess",
  getClientCredentialsSecurityAccess,
);

function hasClientCredentialsBasic(): boolean {
  const credentialAccess = getClientCredentialsSecurityAccess();
  const authMechanismsInUse: Record<string, boolean> = {};

  for (let { securityType } of credentialAccess) {
    // Peel composite security types to get the actual OAuth2 scheme
    securityType = peelClientCredentialsSecurityType(securityType);

    if (
      securityType?.Extensions?.All["x-speakeasy-token-endpoint-authentication"]
    ) {
      authMechanismsInUse[
        securityType.Extensions.All["x-speakeasy-token-endpoint-authentication"]
      ] = true;
    }
  }

  return Boolean(authMechanismsInUse["client_secret_basic"]);
}
registerTemplateFunc("hasClientCredentialsBasic", hasClientCredentialsBasic);

function hasOverridableOAuth2Scopes(securityType: TypeDef): boolean {
  return securityType.Extensions?.OverridableOAuth2Scopes === true;
}
registerTemplateFunc("hasOverridableOAuth2Scopes", hasOverridableOAuth2Scopes);

// Returns the list of required scopes to be requested in one of the OAuth2 built-in hooks
// See OAuth2BuiltInFlows for the list of currently supported built-in flows
function getRequiredOAuth2Scopes(op: Operation): string[] | null {
  const oauth2Config =
    op.OAuth2Config || context.Global.AST.MainSDK.SecurityConfig.OAuth2Config;

  if (Object.keys(oauth2Config).length === 0) {
    return null;
  }

  // Multiple OAuth2 built-in flows are not currently supported
  // so we can simply return the first one we find, if any.
  for (const flow of OAuth2BuiltInFlows) {
    const config = Object.values(oauth2Config).find(
      (f) => f.Enabled && f.Flow.toString() === flow.toString(),
    );
    if (config) {
      return config.RequiredScopes;
    }
  }

  return null;
}

function collectAvailableOAuth2Scopes(): OAuth2Config {
  const mainSDK = context.Global.AST.MainSDK;

  const config: OAuth2Config = { ...mainSDK.SecurityConfig.OAuth2Config };

  const collectFromSDK = (sdk: SDK) => {
    for (const op of sdk.Operations) {
      if (op.OAuth2Config) {
        Object.entries(op.OAuth2Config).forEach(([schemeKey, flowConfig]) => {
          if (!(schemeKey in config)) {
            config[schemeKey] = flowConfig;
          }
        });
      }
    }

    for (const subSDK of sdk.SubSDKs) {
      collectFromSDK(subSDK);
    }
  };

  collectFromSDK(mainSDK);

  return config;
}

function hasAvailableOAuth2Scopes(config: OAuth2Config): boolean {
  return Object.values(config).some(
    (flowConfig) => flowConfig.AvailableScopes.length > 0,
  );
}

function getOAuth2ScopeClassName(schemeKey: string): string {
  if (schemeKey.toLowerCase() == "oauth2") {
    return sanitizeClassName("OAuth2Scope");
  }

  return sanitizeClassName(`${schemeKey}OAuth2Scope`);
}

registerTemplateFunc("getOAuth2ScopeClassName", getOAuth2ScopeClassName);

// @ts-ignore
function getOAuth2PasswordSecurityAccess(): OAuth2PasswordSecurityAccess {
  const globalSecurity = context.Global.AST.MainSDK.Security;

  const oAuth2PasswordTypes = [];
  const seenIds = new Set<string>();

  if (
    globalSecurity?.Type &&
    getOAuth2FlowFields("password", getSecurityFields(globalSecurity.Type))
      .length > 0
  ) {
    oAuth2PasswordTypes.push({
      id: "",
      nameId: "",
      securityType: globalSecurity.Type,
      optional: globalSecurity.Optional,
    });
    seenIds.add("");
  }

  const getOpSecurityType = (sdk: SDK) => {
    for (const op of sdk.Operations) {
      if (!op.Security) {
        continue;
      }

      const opId = op.GetID();
      if (seenIds.has(opId)) {
        continue;
      }

      const fields = getOAuth2FlowFields(
        "password",
        getSecurityFields(op.Security.Type),
      );

      if (fields.length > 0) {
        seenIds.add(opId);
        oAuth2PasswordTypes.push({
          id: op.ID,
          nameId: op.GetID(),
          securityType: op.Security.Type,
          optional: op.Security.Optional,
        });
      }
    }

    for (const subSDK of sdk.SubSDKs) {
      getOpSecurityType(subSDK);
    }
  };

  getOpSecurityType(context.Global.AST.MainSDK);

  if (oAuth2PasswordTypes.length == 0) {
    throw new Error("No OAuth2 Resource Owner Password security found");
  }

  return oAuth2PasswordTypes;
}
registerTemplateFunc(
  "getOAuth2PasswordSecurityAccess",
  getOAuth2PasswordSecurityAccess,
);

// @ts-ignore
function collectClientCredentialsAdditionalFields(): FieldDef[] {
  let access: ClientCredentialsSecurityAccess = [];
  try {
    access = getClientCredentialsSecurityAccess();
  } catch (e: any) {
    if ((e?.message || "") === "No client credentials security found") {
      return [];
    }
    // Re-throw unexpected errors
    throw e;
  }
  const byName: Record<string, FieldDef> = {};

  const addField = (f: FieldDef) => {
    // Only include primitive string fields for credentials typing in TS
    if (f?.Type?.Type.toString() !== "string") {
      return;
    }
    const key = sanitizeFieldName(f.Name);
    // Exclude standard credential identifiers; we only want additional fields
    const normalizedKey = f.Name.replace(/_/g, "").toLowerCase();
    if (
      ["clientid", "clientsecret", "tokenurl", "scopes"].includes(normalizedKey)
    ) {
      return;
    }
    if (!byName[key]) {
      byName[key] = f;
    }
  };

  for (const { securityType } of access) {
    const secFields = getSecurityFields(securityType);

    if (hasClientCredentialsOnly(secFields)) {
      // flattened: credentials fields live at the top-level of the security type
      for (const f of securityType.Fields) {
        // Include common credential-related fields (clientID, clientSecret, tokenURL, etc.)
        addField(f);
      }
    } else {
      // nested: find the specific client_credentials scheme field and its class fields
      const parts = getOAuth2FlowFields("client_credentials", secFields);
      if (!parts.length) {
        continue;
      }

      // getOAuth2FlowFields returns the deepest field that matches the flow
      // so we need to get the last part of the array
      const schemeField = parts[parts.length - 1];
      if (schemeField?.Type?.Fields) {
        for (const f of schemeField.Type.Fields) {
          addField(f);
        }
      }
    }
  }

  return Object.values(byName);
}
registerTemplateFunc(
  "collectClientCredentialsAdditionalFields",
  collectClientCredentialsAdditionalFields,
);

function isCompositeSecurityType(securityType: TypeDef): boolean {
  const securityFields = getSecurityFields(securityType);
  return (
    securityFields.length > 0 &&
    securityFields.every((f) => isSecurityOption(f))
  );
}

/**
 * Extracts the actual OAuth2 client credentials scheme from a potentially composite security type.
 * This is necessary to read scheme-specific properties like extensions.
 *
 * @param securityType - The security type definition, which may be composite or direct
 * @returns The unwrapped OAuth2 client credentials scheme type, or the original type if not composite
 *
 * @example
 * // For a security definition like: security: [oauth2ClientCredentials: [], apiKey: {}]
 * // This returns the oauth2ClientCredentials TypeDef
 */
function peelClientCredentialsSecurityType(securityType: TypeDef): TypeDef {
  if (!isCompositeSecurityType(securityType)) {
    return securityType;
  }
  // Use the first OAuth2 flow field found (safe since specs with multiple
  // client credentials schemes are rejected during validation)
  const parts = getOAuth2FlowFields(
    "client_credentials",
    getSecurityFields(securityType),
  );

  return parts[0]?.Type;
}
