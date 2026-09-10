/** Returns true if the SDK has an absolute server URL. */
function hasServerURL(sdk: SDK): boolean {
  return sdk.Servers?.HasAbsoluteURL() || false;
}

/** Returns a generic description for the server URL provider attribute. */
function serverURLDescription(sdk: SDK): string {
  if (hasServerURL(sdk)) {
    return `Server URL (defaults to ${sdk.Servers.Servers[0].URL})`;
  }
  return "Server URL";
}

/** Returns the description for a server URL variable provider attribute. */
function serverURLVariableDescription(variable: ServerVariable): string {
  let description = variable.Type.Comments?.Description || "";

  if (variable.Default) {
    description += ` (defaults to ${variable.Default})`;
  }

  return description;
}

/** Returns the terraform.serverVariableProviderAttributes generation
 *  configuration. */
function providerServerVariableOverrides(): Record<string, string> {
  return context.Global.Config.ServerVariableProviderAttributes || {};
}

/** Returns true if the OAS server variable has an explicitly configured
 *  provider attribute name. */
function hasProviderServerVariableOverride(variableName: string): boolean {
  return Object.prototype.hasOwnProperty.call(
    providerServerVariableOverrides(),
    variableName,
  );
}

/** Returns the source name for an OAS server variable's provider attribute
 *  and data model field: the configured override, otherwise the variable
 *  name. */
function providerServerVariableSourceName(variableName: string): string {
  if (hasProviderServerVariableOverride(variableName)) {
    return providerServerVariableOverrides()[variableName] || "";
  }

  return variableName;
}

/** Returns the provider attribute name for an OAS server variable. */
function providerServerVariableAttributeName(variableName: string): string {
  return sanitizeTFStateName(providerServerVariableSourceName(variableName));
}

/** Throws if an explicitly configured server variable provider attribute name
 *  collides with a reserved provider attribute name or with another server
 *  variable's provider attribute name. Server variables without a configured
 *  override are never rejected. */
function assertServerVariableProviderAttributeNames(
  serverVariables: ServerVariable[],
  reservedAttributeNames: Set<string>,
): void {
  const takenAttributeNames = new Set(reservedAttributeNames);

  serverVariables.forEach((variable) => {
    if (!hasProviderServerVariableOverride(variable.Name)) {
      takenAttributeNames.add(
        providerServerVariableAttributeName(variable.Name),
      );
    }
  });

  serverVariables.forEach((variable) => {
    if (!hasProviderServerVariableOverride(variable.Name)) {
      return;
    }

    const attributeName = providerServerVariableAttributeName(variable.Name);

    if (!attributeName) {
      throw new Error(
        `terraform.serverVariableProviderAttributes maps server variable ${JSON.stringify(
          variable.Name,
        )} to ${JSON.stringify(
          providerServerVariableSourceName(variable.Name),
        )}, which sanitizes to an empty provider attribute name`,
      );
    }

    if (takenAttributeNames.has(attributeName)) {
      throw new Error(
        `terraform.serverVariableProviderAttributes maps server variable ${JSON.stringify(
          variable.Name,
        )} to provider attribute ${JSON.stringify(
          attributeName,
        )}, which is already used by another provider attribute`,
      );
    }

    takenAttributeNames.add(attributeName);
  });
}

/**
 * Templates server URL handling logic for the Provider Configure method.
 *
 * Server URL configuration preference order:
 *    1. Generation configuration baseServerUrl (no other value if set)
 *    2. server_url provider attribute value
 *    3. Generation configuration terraform.environmentVariables["server_url"]
 *       environment variable value
 *    3. First SDK server URL
 */
function templateProviderServerURLConfiguration(
  context: ProviderContext,
): string {
  if (context.Config.BaseServerURL) {
    return `serverUrl := "${context.Config.BaseServerURL}"`;
  }

  const result: string[] = [`serverUrl := data.ServerURL.ValueString()`];
  const serverVariables = context.AST.MainSDK.Servers?.GetVariables();

  if (getEnvironmentVariable("server_url")) {
    const envKey = JSON.stringify(getEnvironmentVariable("server_url"));
    addGenImport("os");
    result.push(
      ``,
      `if serverUrl == "" && os.Getenv(${envKey}) != "" {`,
      `serverUrl = os.Getenv(${envKey})`,
      `}`,
    );
  }

  if (!hasServerURL(context.AST.MainSDK)) {
    result.push(
      ``,
      `if serverUrl == "" {`,
      `resp.Diagnostics.AddError("server_url is required", "The server_url attribute must be provided in the provider configuration.")`,
      `return`,
      `}`,
    );
  } else {
    result.push(
      ``,
      `if serverUrl == "" {`,
      `serverUrl = "${context.AST.MainSDK.Servers.Servers[0].URL}"`,
      `}`,
    );
  }

  if (serverVariables && serverVariables.length > 0) {
    result.push(``, `serverUrlParams := make(map[string]string)`);

    serverVariables.forEach((variable) => {
      const attributeName = providerServerVariableAttributeName(variable.Name);
      const envVariable = getEnvironmentVariable(attributeName);
      const fieldName = sanitizeFieldName(
        providerServerVariableSourceName(variable.Name),
      );
      // NOTE: serverUrlParams keys must equal the OAS server variable names
      // as they are directly used in the server URL template.
      const paramKey = JSON.stringify(variable.Name);

      result.push(
        ``,
        `if data.${fieldName}.ValueString() != "" {`,
        `serverUrlParams[${paramKey}] = data.${fieldName}.ValueString()`,
        `}`,
      );

      if (envVariable) {
        const envKey = JSON.stringify(envVariable);
        addGenImport("os");
        result.push(
          ``,
          `if _, ok := serverUrlParams[${paramKey}]; !ok && os.Getenv(${envKey}) != "" {`,
          `serverUrlParams[${paramKey}] = os.Getenv(${envKey})`,
          `}`,
        );
      }

      if (variable.Default) {
        result.push(
          ``,
          `if _, ok := serverUrlParams[${paramKey}]; !ok {`,
          `serverUrlParams[${paramKey}] = "${variable.Default}"`,
          `}`,
        );
      }
    });
  }

  return result.join(`\n`);
}

registerTemplateFunc(
  "templateProviderServerURLConfiguration",
  templateProviderServerURLConfiguration,
);
