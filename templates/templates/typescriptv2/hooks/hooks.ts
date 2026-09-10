// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];

  const registrationFile = "src/hooks/registration.ts";
  let registrationData = readFile(registrationFile);

  // If we don't already have a registration file, create it
  if (!registrationData && accountHasFeatureAccess("sdkHooks")) {
    jobs.push(
      createTemplateFileJob(
        "hooks/registration.ts.stmpl",
        registrationFile,
        {},
      ),
    );
  }

  jobs.push(
    createTemplateFileJob("hooks/hooks.ts.stmpl", "src/hooks/hooks.ts", {}),
    createTemplateFileJob("hooks/types.ts.stmpl", "src/hooks/types.ts", {}),
    createTemplateFileJob("hooks/index.ts.stmpl", "src/hooks/index.ts", {}),
  );

  if (hasClientCredentials()) {
    const clientCredentialsFileName = getClientCredentialsFileName();
    jobs.push(
      createTemplateFileJob(
        `hooks/{{getClientCredentialsFileName}}.ts.stmpl`,
        `src/hooks/${clientCredentialsFileName}.ts`,
        {},
      ),
    );
  }

  if (hasOAuth2PasswordFlow()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2password.ts.stmpl",
        "src/hooks/oauth2password.ts",
        {},
      ),
    );
  }

  const webhookSecurity = getWebhookSecurity();
  if (webhookSecurity) {
    const isCustom = webhookSecurity.Type === "custom";
    const outLocation = isCustom
      ? "src/hooks/webhook-security-custom.ts"
      : "src/hooks/webhook-security.ts";
    const hookAlreadyExists = fileExists(outLocation);
    const shouldWrite = !(isCustom && hookAlreadyExists);
    if (shouldWrite) {
      jobs.push(
        createTemplateFileJob(
          "hooks/webhook-security.ts.stmpl",
          outLocation,
          {},
        ),
      );
    }
  }

  const oauth2Config = collectAvailableOAuth2Scopes();
  if (hasAvailableOAuth2Scopes(oauth2Config)) {
    const oauth2ScopesFileName = getOAuth2ScopesFileName();
    jobs.push(
      createTemplateFileJob(
        `hooks/{{getOAuth2ScopesFileName}}.ts.stmpl`,
        `src/hooks/${oauth2ScopesFileName}.ts`,
        oauth2Config,
      ),
    );
  }

  return jobs;
}

// @ts-ignore
function templateClientCredentialsSecurityAccess(
  credentialAccess: ClientCredentialsSecurityAccess,
): string {
  let hasGlobalOAuth2 = false;

  const switchMembers = [];

  for (const { id } of credentialAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    switchMembers.push({
      id: id,
      methodName: sanitizePrivateMethodName(`getCredentials_${id}`),
    });
  }

  if (hasGlobalOAuth2 && switchMembers.length == 0) {
    return `return this.getCredentialsGlobal(security)`;
  }

  const lines = [`switch (hookCtx.{{constants.hookContext.operationID}}) {`];

  for (const { id, methodName } of switchMembers) {
    lines.push(`case "${id}":`);
    lines.push(indentString(`return this.${methodName}(security)`, 1));
  }

  if (hasGlobalOAuth2) {
    lines.push(`default:`);
    lines.push(indentString(`return this.getCredentialsGlobal(security)`, 1));
    lines.push(`}`);
  } else {
    lines.push(`}`);
    lines.push(`return null`);
  }

  return lines.join("\n");
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccess",
  templateClientCredentialsSecurityAccess,
);

// @ts-ignore
function templateClientCredentialsTokenAuthentication(
  credentialAccess: Array<{
    id: string;
    securityType: TypeDef;
  }>,
): string {
  const authMechanismsInUse: Record<string, boolean> = {};

  for (let { securityType } of credentialAccess) {
    securityType = peelClientCredentialsSecurityType(securityType);
    if (
      securityType.Extensions?.All["x-speakeasy-token-endpoint-authentication"]
    ) {
      authMechanismsInUse[
        securityType.Extensions.All["x-speakeasy-token-endpoint-authentication"]
      ] = true;
    }
  }

  const lines = [];
  if (
    authMechanismsInUse["client_secret_post"] ||
    Object.keys(authMechanismsInUse).length === 0
  ) {
    lines.push(
      `formData.append("client_id", credentials.{{constants.credentials.clientID}});`,
    );
    lines.push(
      `formData.append("client_secret", credentials.{{constants.credentials.clientSecret}});`,
    );
  } else if (authMechanismsInUse["client_secret_basic"]) {
    // old code paths might give us 0 code paths
    addInternalImport("base64", "stringToBase64", "hooks");
    lines.push(
      `headers["Authorization"] = \`Basic \$\{stringToBase64(\`\$\{credentials.{{constants.credentials.clientID}}\}:\$\{credentials.{{constants.credentials.clientSecret}}\}\`)\}\``,
    );
  } else {
    throw new Error(
      "x-speakeasy-token-endpoint-authentication must be set to client_secret_post or client_secret_basic",
    );
  }

  return lines.join("\n");
}
registerTemplateFunc(
  "templateClientCredentialsTokenAuthentication",
  templateClientCredentialsTokenAuthentication,
);

// @ts-ignore
function templateClientCredentialsSecurityAccessFunction(
  securityType: TypeDef,
): string {
  let securityFields = getSecurityFields(securityType);
  let securitySchemeType = securityType;
  const securityTypes = resolveOutbound({
    usageLocation: "hooks",
    typeDef: securityType,
    rootTypeDef: securityType,
    optional: false,
    nullable: false,
    constValue: null,
    defaultValue: null,
  });
  if (!isNoZod()) {
    securityTypes.importZodTypes();
  } else {
    // No-zod: we cast through the TS type below — make sure that import lands.
    securityTypes.importOutputTypes();
  }

  let accessor = "out";

  if (hasClientCredentialsOnly(securityFields)) {
    // 'tokenURL' (has default) and 'scopes' (is optional) don't get collected
    // by getSecurityFields because they don't have a 'security' annotation.
    // When OAuth2Scopes are overridable, the 'scopes' field is expected to be
    // part of the collected security fields. Otherwise we will append it
    // manually to the list of key-value pairs.
    const extraFields = ["tokenurl"];
    if (hasOverridableOAuth2Scopes(securitySchemeType)) {
      extraFields.push("scopes");
    }

    extraFields.forEach((fieldName) => {
      const field = securityType.Fields.find(
        (f) => f.Name.toLowerCase() === fieldName,
      );
      if (!field) {
        throw new Error(
          `No field named ${fieldName} found in ${securityType.Name}`,
        );
      }
      securityFields.push(field);
    });
  } else {
    let parts = getOAuth2FlowFields("client_credentials", securityType.Fields);
    if (!parts.length) {
      throw new Error("No client credentials field found.");
    }

    // filter parts to retrieve field def where .Type.Fields is not empty
    parts = parts.filter((f) => f.Type.Fields.length > 0);

    accessor = parts.reduce(
      (acc, f) => sanitizeAccessor(acc, f.Name, true),
      accessor,
    );

    securitySchemeType = parts[parts.length - 1].Type;
    securityFields = securitySchemeType.Fields;
    // Don't need to inject the 'tokenURL' field here because it's not been filtered
    // out by getOAuth2FlowFields. Same for 'scopes' when they are overridable.
  }

  const coreFields = securityFields.filter((f) =>
    ["clientid", "clientsecret", "tokenurl", "scopes"].includes(
      f.Name.replace(/_/g, "").toLowerCase(),
    ),
  );

  let defaultsLogic = "";
  const kvs = coreFields.map((f) => {
    const k = sanitizeFieldName(f.Name);
    let v = sanitizeAccessor(accessor, originalFieldName(f), true);
    if (context.Global.Config.EnvVarPrefix) {
      addInternalImport("env", "env", "hooks");
      const envVar = "env()." + templateSecurityEnvVars(f);

      if (f.Default) {
        // Check whether field's value matches the default value. If so we assume
        // it was prefilled by Zod and we prioritize the env var if it is set.
        // In no-zod mode the default is never injected by a Zod schema, so we
        // also treat `undefined` as "should use env var / default".
        const dfltVarName = `DEFAULT_${caser().ToSNAKE(f.Name)}`;
        const envVarName = `env${caser().ToPascal(f.Name)}`;
        defaultsLogic = `
const ${dfltVarName} = ${JSON.stringify(f.Default.Value)};
const ${envVarName} = ${envVar} ?? ${dfltVarName};`;

        if (isNoZod()) {
          return `${k}: (${v} == null || ${v} === ${dfltVarName}) ? ${envVarName} : ${v}`;
        }
        return `${k}: ${v} !== ${dfltVarName} ? ${v} : ${envVarName}`;
      }

      v += " ?? " + envVar;
    }
    const fallback = f.Name.toLowerCase() === "scopes" ? "undefined" : '""';
    return `${k}: ${v} ?? ${fallback}`;
  });

  if (!hasOverridableOAuth2Scopes(securitySchemeType)) {
    kvs.push("scopes: undefined");
  }

  const additionalPropsLogic = `
const additionalProperties: Record<string, string> = {};
for (const [key, value] of Object.entries(${accessor} ?? {})) {
  if (typeof value === "string" && !["{{constants.credentials.clientID}}", "{{constants.credentials.clientSecret}}", "{{constants.credentials.tokenURL}}", "{{constants.credentials.scopes}}"].includes(key)) {
    additionalProperties[key] = value;
  }
}`;

  const outDecl = isNoZod()
    ? `const out = security as ${securityTypes.outputType};`
    : `const out = parse(
  security,
  (val) => ${securityTypes.zod}.parse(val),
  "unexpected security type",
);`;

  return `
${outDecl}
${defaultsLogic}
${additionalPropsLogic}

return {${kvs.join(", ")}, additionalProperties};`;
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);
