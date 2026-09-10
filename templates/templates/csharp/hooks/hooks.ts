// @ts-ignore
function getHooksLocation() {
  return `${getSDKTopLevelFolder()}/${getScopePath("hooks")}`;
}

// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];

  const hooksPath = getHooksLocation();
  const registrationFile = `${hooksPath}/HookRegistration.cs`;

  // If we don't already have a registration file, create it
  if (!readFile(registrationFile) && accountHasFeatureAccess("sdkHooks")) {
    jobs.push(
      createTemplateFileJob(
        "hooks/registration.cs.stmpl",
        registrationFile,
        {},
      ),
    );
  }

  jobs.push(
    createTemplateFileJob(
      "hooks/hooks.cs.stmpl",
      `${hooksPath}/SDKHooks.cs`,
      {},
    ),
  );
  jobs.push(
    createTemplateFileJob(
      "hooks/types.cs.stmpl",
      `${hooksPath}/HookTypes.cs`,
      {},
    ),
  );

  if (hasClientCredentials()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/clientcredentials.cs.stmpl",
        `${hooksPath}/ClientCredentials/ClientCredentials.cs`,
        {},
      ),
    );
  }
  const oauth2Config = collectAvailableOAuth2Scopes();
  if (hasAvailableOAuth2Scopes(oauth2Config)) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2scopes.cs.stmpl",
        `${hooksPath}/OAuth2Scopes.cs`,
        oauth2Config,
      ),
    );
  }

  return jobs;
}

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return "null";
  }

  return `new List<string> { ${scopes.map((s) => `"${s}"`).join(", ")} }`;
}

registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

function templateClientCredentialsMethodName(id: string): string {
  if (id === "") {
    id = "global";
  }
  return `GetCredentials${caser().ToPascal(sanitizeName(id))}`;
}

// @ts-ignore
function templateClientCredentialsSecurityAccess(
  clientCredentialsAccess: ClientCredentialsSecurityAccess,
): string {
  const methodCall = (id: string) => {
    const methodName = templateClientCredentialsMethodName(id);
    return `return ${methodName}(hookCtx.SecuritySource);`;
  };

  if (clientCredentialsAccess.every((a) => a.id === "")) {
    return methodCall("global");
  }

  let hasGlobalOAuth2 = false;
  const lines: string[] = [];
  for (const { id } of clientCredentialsAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    lines.push(`\nif (hookCtx.OperationID == "${id}")\n{`);
    lines.push(indentString(methodCall(id), 1));
    lines.push("}");
  }

  if (hasGlobalOAuth2) {
    lines.push("else\n{");
    lines.push(indentString(methodCall("global"), 1));
    lines.push("}");
  } else {
    lines.push("\nreturn null;");
  }

  return lines.join("\n");
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccess",
  templateClientCredentialsSecurityAccess,
);

function templateClientCredentialsImports(
  clientCredentialsAccess: ClientCredentialsSecurityAccess,
  indent: number = 0,
): string {
  let imports = new Set<string>();
  imports.add("System");
  imports.add("System.Collections.Generic");
  imports.add("System.Collections.Concurrent");
  imports.add("System.Linq");
  imports.add("System.Net.Http");
  imports.add("System.Security.Cryptography");
  imports.add("System.Text");
  imports.add("System.Threading.Tasks");
  imports.add("Newtonsoft.Json");
  imports = addImport(imports, getScopeNamespace("shared"), true);
  imports = addImport(imports, "Utils", true);

  if (context.Global.Config.EnableCancellationToken) {
    imports.add("System.Threading");
  }

  for (const { securityType } of clientCredentialsAccess) {
    imports = getTypeImports(imports, securityType, "", false);
  }

  return templateImports(imports, indent);
}
registerTemplateFunc(
  "templateClientCredentialsImports",
  templateClientCredentialsImports,
);

function templateClientCredentialsSecurityAccessFunction(
  clientCredentialsSecurity: ClientCredentialsSecurity,
): string {
  const { id, securityType } = clientCredentialsSecurity;
  const securityFields = getSecurityFields(securityType);
  let fieldName = "security?";
  let securitySchemeType = securityType;

  if (!hasClientCredentialsOnly(securityFields)) {
    const parts = getOAuth2FlowFields("client_credentials", securityFields);
    if (!parts.length) {
      throw new Error("No client credentials field found");
    }

    parts
      .filter((f) => f.Type.Type.toString() === "class")
      .forEach((f) => {
        fieldName += `.${sanitizeFieldName(f.Name)}${f.Optional ? "?" : ""}`;
        securitySchemeType = f.Type;
      });
  }

  const methodName = templateClientCredentialsMethodName(id);

  const scopes = hasOverridableOAuth2Scopes(securitySchemeType)
    ? `${fieldName}.Scopes`
    : "null";

  return `
/// <summary>
/// Retrieves Client Credentials OAuth2 credentials from the security source.
/// </summary>
/// <param name="securitySource">A callback function that provides the security object.</param>
/// <returns>Credentials if available, null otherwise.</returns>
private Credentials? ${methodName}(Func<object> securitySource)
{
    var security = securitySource() as ${sanitizeType(securityType)};

    if (${fieldName}.ClientID == null || ${fieldName}.ClientSecret == null || ${fieldName}.TokenURL == null)
    {
        return null;
    }

    return new Credentials(
        ${fieldName}.ClientID!,
        ${fieldName}.ClientSecret!,
        ${fieldName}.TokenURL!,
        ${scopes}
    );
}`;
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);

// @ts-ignore
function templateClientCredentialsTokenAuthentication(
  credentialAccess: ClientCredentialsSecurityAccess,
): string {
  const authMechanismsInUse: Record<string, boolean> = {};

  for (let { securityType } of credentialAccess) {
    securityType = peelClientCredentialsSecurityType(securityType);
    const tokenAuthType =
      securityType?.Extensions?.All[
        "x-speakeasy-token-endpoint-authentication"
      ];
    if (tokenAuthType) {
      authMechanismsInUse[tokenAuthType] = true;
    }
  }

  if (authMechanismsInUse["client_secret_basic"]) {
    return `var basicAuth = Convert.ToBase64String(
    Encoding.UTF8.GetBytes($"{credentials.ClientID}:{credentials.ClientSecret}")
);
request.Headers.Add("Authorization", $"Basic {basicAuth}");`;
  }
  return `payload.Add(new PayloadValue("client_id", credentials.ClientID));
payload.Add(new PayloadValue("client_secret", credentials.ClientSecret));`;
}

registerTemplateFunc(
  "templateClientCredentialsTokenAuthentication",
  templateClientCredentialsTokenAuthentication,
);
