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
  );

  if (hasClientCredentials()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/clientcredentials.ts.stmpl",
        "src/hooks/clientcredentials.ts",
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

  const oauth2Config = collectAvailableOAuth2Scopes();
  if (hasAvailableOAuth2Scopes(oauth2Config)) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2scopes.ts.stmpl",
        "src/hooks/oauth2scopes.ts",
        oauth2Config,
      ),
    );
  }

  return jobs;
}

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

  const lines = [`switch (hookCtx.operationID) {`];

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

function templateClientCredentialsTokenAuthentication(
  credentialAccess: Array<{
    id: string;
    securityType: TypeDef;
  }>,
): string {
  const authMechanismsInUse: Record<string, boolean> = {};

  for (const { securityType } of credentialAccess) {
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
    lines.push(`formData.append("client_id", credentials.clientID);`);
    lines.push(`formData.append("client_secret", credentials.clientSecret);`);
  } else if (authMechanismsInUse["client_secret_basic"]) {
    // old code paths might give us 0 code paths
    addInternalImport("base64", "stringToBase64", "hooks");
    lines.push(
      `headers["Authorization"] = \`Basic \$\{stringToBase64(\`\$\{credentials.clientID\}:\$\{credentials.clientSecret\}\`)\}\``,
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

function templateClientCredentialsSecurityAccessFunction(
  securityType: TypeDef,
): string {
  let securityFields = getSecurityFields(securityType);
  const securityTypes = resolveOutbound({
    usageLocation: "hooks",
    typeDef: securityType,
    rootTypeDef: securityType,
    optional: false,
    nullable: false,
    constValue: null,
    defaultValue: null,
  });
  securityTypes.importZodTypes();

  let accessor = "out";
  let securitySchemeType = securityType;

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
    const parts = getOAuth2FlowFields(
      "client_credentials",
      securityType.Fields,
    );
    if (!parts.length) {
      throw new Error("No client credentials field found.");
    }
    accessor = parts.reduce(
      (acc, f) =>
        sanitizeAccessor(acc, sanitizeFieldName(originalFieldName(f)), true),
      accessor,
    );

    securitySchemeType = parts[parts.length - 1].Type;
    securityFields = securitySchemeType.Fields;
    // Don't need to inject token url field here because it's not been filtered
    // out as with getGlobalSecurityFields.
  }

  const coreFields = securityFields.filter((f) =>
    ["clientid", "clientsecret", "tokenurl", "scopes"].includes(
      f.Name.replace(/_/g, "").toLowerCase(),
    ),
  );

  const kvs = coreFields.map((f) => {
    const k = sanitizeFieldName(originalFieldName(f));
    let v = sanitizeAccessor(
      accessor,
      sanitizeFieldName(originalFieldName(f)),
      true,
    );
    if (context.Global.Config.EnvVarPrefix) {
      addInternalImport("env", "env", "hooks");
      v += " ?? env()." + templateSecurityEnvVars(f);
    }
    const fallback = isScopesField(f) ? "undefined" : '""';
    return `${k}: ${v} ?? ${fallback}`;
  });

  if (!hasScopesField(coreFields)) {
    kvs.push("scopes: undefined");
  }

  const additionalPropsLogic = `
const additionalProperties: Record<string, string> = {};
for (const [key, value] of Object.entries(${accessor} ?? {})) {
  if (typeof value === "string" && !["clientID", "clientSecret", "tokenURL", "scopes"].includes(key)) {
    additionalProperties[key] = value;
  }
}`;

  return `
const out = parse(
  security,
  (val) => ${securityTypes.zod}.parse(val),
  "unexpected security type",
);
${additionalPropsLogic}

return {${kvs.join(", ")}, additionalProperties};`;
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);

function hasScopesField(fields: FieldDef[]): boolean {
  return fields.some(isScopesField);
}

function isScopesField(f: FieldDef): boolean {
  return f.Name.replace(/_/g, "").toLowerCase() === "scopes";
}
