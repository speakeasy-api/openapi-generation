// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];

  const hooksPath = `${getSourcePath()}/_hooks`;

  const registrationFile = `${hooksPath}/registration.py`;

  let registrationData = readFile(registrationFile);

  // If we don't already have a registration file, create it
  if (!registrationData && accountHasFeatureAccess("sdkHooks")) {
    jobs.push(
      createTemplateFileJob(
        "hooks/registration.py.stmpl",
        registrationFile,
        {},
      ),
    );
  }

  const oauth2Config = collectAvailableOAuth2Scopes();
  const hasAvailableScopes = hasAvailableOAuth2Scopes(oauth2Config);

  jobs.push(
    createTemplateFileJob(
      "hooks/sdkhooks.py.stmpl",
      `${hooksPath}/sdkhooks.py`,
      {},
    ),
    createTemplateFileJob("hooks/types.py.stmpl", `${hooksPath}/types.py`, {}),
    createTemplateFileJob(
      "hooks/__init__.py.stmpl",
      `${hooksPath}/__init__.py`,
      { HasAvailableOAuth2Scopes: hasAvailableScopes },
    ),
  );

  // Generate async hooks infrastructure only when enabled
  if (useAsyncHooks()) {
    // Async registration file (one-time only, like registration.py)
    const asyncRegistrationFile = `${hooksPath}/asyncregistration.py`;
    let asyncRegistrationData = readFile(asyncRegistrationFile);

    // If we don't already have an async registration file, create it
    if (!asyncRegistrationData && accountHasFeatureAccess("sdkHooks")) {
      jobs.push(
        createTemplateFileJob(
          "hooks/asyncregistration.py.stmpl",
          asyncRegistrationFile,
          {},
        ),
      );
    }

    jobs.push(
      createTemplateFileJob(
        "hooks/asynctypes.py.stmpl",
        `${hooksPath}/asynctypes.py`,
        {},
      ),
      createTemplateFileJob(
        "hooks/asyncsdkhooks.py.stmpl",
        `${hooksPath}/asyncsdkhooks.py`,
        {},
      ),
      createTemplateFileJob(
        "hooks/adapters.py.stmpl",
        `${hooksPath}/adapters.py`,
        {},
      ),
    );
  }

  if (hasClientCredentials()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/clientcredentials.py.stmpl",
        `${hooksPath}/clientcredentials.py`,
        {},
      ),
    );
  }

  if (hasAvailableScopes) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2scopes.py.stmpl",
        `${hooksPath}/oauth2scopes.py`,
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
  if (credentialAccess.every((a) => a.id === "")) {
    return `return self.get_credentials_global(security)`;
  }

  let hasGlobalOAuth2 = false;
  const lines: string[] = [];
  for (const { id } of credentialAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    const methodName = sanitizePrivateMethodName(`getCredentials_${id}`);
    lines.push(`if hook_ctx.operation_id == "${id}":`);
    lines.push(indentString(`return self.${methodName}(security)`, 1));
  }

  if (hasGlobalOAuth2) {
    lines.push(`\nreturn self.get_credentials_global(security)`);
  } else {
    lines.push(`\nreturn None`);
  }

  return lines.join("\n");
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccess",
  templateClientCredentialsSecurityAccess,
);

// @ts-ignore
function templateClientCredentialsSecurityAccessFunction(
  securityType: TypeDef,
): string {
  const fields = getSecurityFields(securityType);
  let securitySchemeType = securityType;

  let fieldName = "security";
  let fieldSetCheck = "security is None";
  if (hasClientCredentialsOnly(fields)) {
    if (fields.every((f) => f.Optional)) {
      fieldSetCheck += ` or ${fieldName}.client_id is None or ${fieldName}.client_secret is None`;
    }
  } else {
    const parts = getOAuth2FlowFields("client_credentials", fields);
    if (!parts.length) {
      throw new Error("No client credentials field found");
    }

    fieldSetCheck +=
      " or " +
      parts
        .filter((f) => f.Type.Type.toString() === "class" && f.Optional)
        .map((f) => {
          return `${fieldName}.${sanitizeFieldName(f.Name)} is None`;
        })
        .join(" or ");

    fieldName +=
      "." +
      parts
        .filter((f) => f.Type.Type.toString() === "class")
        .map((f) => {
          securitySchemeType = f.Type;
          return sanitizeFieldName(f.Name);
        })
        .join(".");
  }

  const scopes = hasOverridableOAuth2Scopes(securitySchemeType)
    ? `${fieldName}.scopes`
    : "None";

  return `if ${fieldSetCheck}:
    return None

# Extract additional properties from security object
additional_properties = {}
for key, value in dict(${fieldName}).items():
     if key not in ["client_id", "client_secret", "token_url", "scopes"]:
          additional_properties[key] = value

return Credentials(
    client_id=${fieldName}.client_id,
    client_secret=${fieldName}.client_secret,
    token_url=${fieldName}.token_url,
    scopes=${scopes},
    additional_properties=additional_properties
)`;
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);
