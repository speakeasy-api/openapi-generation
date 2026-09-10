/** Returns directory path for SDK hooks files. */
function getHooksLocation() {
  return `${getSourcePath(context.Global.Config.Module)}/sdk_hooks`;
}

/** Templates SDK hooks files. */
// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];
  const hooksPath = getHooksLocation();

  const registrationFile = `${hooksPath}/registration.rb`;

  let registrationData = readFile(registrationFile);

  // If we don't already have a registration file, create it
  if (!registrationData && accountHasFeatureAccess("sdkHooks")) {
    jobs.push(
      createTemplateFileJob(
        "hooks/registration.rb.stmpl",
        registrationFile,
        {},
      ),
    );
  }

  const oauth2Config = collectAvailableOAuth2Scopes();
  const hasAvailableScopes = hasAvailableOAuth2Scopes(oauth2Config);

  jobs.push(
    createTemplateFileJob("hooks/hooks.rb.stmpl", `${hooksPath}/hooks.rb`, {
      HasAvailableOAuth2Scopes: hasAvailableScopes,
    }),
    createTemplateFileJob("hooks/types.rb.stmpl", `${hooksPath}/types.rb`, {}),
  );

  if (hasClientCredentials()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/clientcredentials.rb.stmpl",
        `${hooksPath}/clientcredentials.rb`,
        {},
      ),
    );
  }

  if (hasOAuth2PasswordFlow()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2password.rb.stmpl",
        `${hooksPath}/oauth2password.rb`,
        {},
      ),
    );
  }

  if (hasAvailableScopes) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2scopes.rb.stmpl",
        `${hooksPath}/oauth2scopes.rb`,
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
    return `return get_credentials_global(security)`;
  }

  let hasGlobalOAuth2 = false;
  const lines: string[] = [];

  for (const { id } of credentialAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    const methodName = sanitizePrivateMethodName(`get_credentials_${id}`);
    lines.push(`when '${id}'`);
    lines.push(indentString(`return ${methodName}(security)`, 1));
  }

  if (lines.length > 0) {
    lines.unshift(`case hook_ctx.operation_id`);

    if (hasGlobalOAuth2) {
      lines.push(`else`);
      lines.push(indentString(`return get_credentials_global(security)`, 1));
    }

    lines.push(`end`);

    if (!hasGlobalOAuth2) {
      lines.push(``);
      lines.push(`nil`);
    }
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
  let fieldSetCheck = "security.nil?";

  if (hasClientCredentialsOnly(fields)) {
    if (fields.every((f) => f.Optional)) {
      fieldSetCheck += ` || ${fieldName}.client_id.nil? || ${fieldName}.client_secret.nil?`;
    }
  } else {
    const parts = getOAuth2FlowFields("client_credentials", fields);
    if (!parts.length) {
      throw new Error("No client credentials field found");
    }

    const optionalNilChecks = parts
      .filter((f) => f.Type.Type.toString() === "class" && f.Optional)
      .map((f) => {
        return `${fieldName}.${sanitizeFieldName(f.Name)}.nil?`;
      })
      .join(" || ");

    if (optionalNilChecks.length > 0) {
      fieldSetCheck += " || " + optionalNilChecks;
    }

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
    : "nil";

  return `return nil if ${fieldSetCheck}

# Extract additional properties from security object
additional_properties = {}
${fieldName}.class.fields.each do |f|
  name = f.name.to_s
  next if %w[client_id client_secret token_url scopes].include?(name)

  metadata = f.metadata&.dig(:security)
  next if metadata.nil?

  field_name = metadata[:field_name] || name
  value = ${fieldName}.send(f.name)
  additional_properties[field_name] = value.to_s unless value.nil?
end

Credentials.new(
  client_id: ${fieldName}.client_id,
  client_secret: ${fieldName}.client_secret,
  token_url: ${fieldName}.token_url,
  scopes: ${scopes},
  additional_properties: additional_properties
)`;
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);

// @ts-ignore
function templateOAuth2PasswordSecurityAccess(
  credentialAccess: OAuth2PasswordSecurityAccess,
): string {
  if (credentialAccess.every((a) => a.id === "")) {
    return `return get_credentials_global(hook_ctx, security)`;
  }

  let hasGlobalOAuth2 = false;
  const lines: string[] = [];

  for (const { id } of credentialAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    const methodName = sanitizePrivateMethodName(`get_credentials_${id}`);
    lines.push(`when '${id}'`);
    lines.push(indentString(`return ${methodName}(hook_ctx, security)`, 1));
  }

  if (lines.length > 0) {
    lines.unshift(`case hook_ctx.operation_id`);

    if (hasGlobalOAuth2) {
      lines.push(`else`);
      lines.push(
        indentString(`return get_credentials_global(hook_ctx, security)`, 1),
      );
    }

    lines.push(`end`);

    if (!hasGlobalOAuth2) {
      lines.push(``);
      lines.push(`nil`);
    }
  }

  return lines.join("\n");
}
registerTemplateFunc(
  "templateOAuth2PasswordSecurityAccess",
  templateOAuth2PasswordSecurityAccess,
);

// @ts-ignore
function templateOAuth2PasswordSecurityAccessFunction(
  securityType: TypeDef,
): string {
  const flattenedSecurity =
    securityType.Fields.length === 1 &&
    context.Global.Config.FlattenGlobalSecurity;
  const securityFields = getSecurityFields(securityType);
  const oAuth2PasswordFields = getOAuth2FlowFields("password", securityFields);

  if (!oAuth2PasswordFields.length) {
    throw new Error("No OAuth2 Resource Owner Password fields found");
  }

  if (oAuth2PasswordFields.length > 1) {
    throw new Error("Multiple OAuth2 Resource Owner Password fields found");
  }

  const unionField = oAuth2PasswordFields[0];
  const unionFieldName = sanitizeFieldName(unionField.Name);
  const unionFieldAccessor = ["security", unionFieldName].join(".");

  // In Ruby, the union type is flattened to T.any(CredentialsType, String).
  // We use is_a? checks to distinguish the union members instead of accessing
  // sub-fields that don't exist on the flat union.

  let result = "";

  if (flattenedSecurity) {
    result += `return nil if security.nil?\n\n`;
  }

  if (unionField.Optional) {
    result += `return nil if ${unionFieldAccessor}.nil?\n\n`;
  }

  result += `if ${unionFieldAccessor}.is_a?(::String)
  return OAuth2AccessTokenCredentials.new(access_token: ${unionFieldAccessor})
end

creds = ${unionFieldAccessor}
token_url = creds.token_url
if token_url.nil? || token_url.empty?
  # Try to get default token_url from field metadata
  creds.class.fields.each do |f|
    next unless f.name.to_s == 'token_url'
    default_val = f.metadata&.dig(:security, :default) || f.metadata&.dig(:default)
    unless default_val.nil? || default_val.empty?
      token_url = default_val
    end
  end
end

token_endpoint = resolve_token_endpoint(hook_ctx.base_url, token_url)

OAuth2PasswordCredentials.new(
  token_endpoint: token_endpoint,
  username: creds.username,
  password: creds.password,
  client_id: creds.respond_to?(:client_id) ? creds.client_id : nil,
  client_secret: creds.respond_to?(:client_secret) ? creds.client_secret : nil
)`;

  return result;
}
registerTemplateFunc(
  "templateOAuth2PasswordSecurityAccessFunction",
  templateOAuth2PasswordSecurityAccessFunction,
);
