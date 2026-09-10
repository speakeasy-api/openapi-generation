// @ts-ignore
function getHooksLocation() {
  return `${getSDKTopLevelFolder()}/${getScopePath("hooks")}`;
}

// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];

  const hooksPath = getHooksLocation();
  const registrationFile = `${hooksPath}/HookRegistration.php`;

  // If we don't already have a registration file, create it
  if (!readFile(registrationFile) && accountHasFeatureAccess("sdkHooks")) {
    jobs.push(
      createTemplateFileJob(
        "hooks/HookRegistration.php.stmpl.skip",
        registrationFile,
        {},
      ),
    );
  }

  const oauth2Config = collectAvailableOAuth2Scopes();
  // Generate one file per OAuth2 scheme to comply with PSR-4 autoloading
  for (const [schemeKey, config] of Object.entries(oauth2Config)) {
    if (config.AvailableScopes.length > 0) {
      const enumName = getOAuth2ScopeClassName(schemeKey);
      jobs.push(
        createTemplateFileJob(
          "hooks/OAuth2Scope.php.stmpl.skip",
          `${hooksPath}/${enumName}.php`,
          { schemeKey, config },
        ),
      );
    }
  }

  jobs.push(getTemplateDirectoryJob("hooks", `${hooksPath}`));

  return jobs;
}

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return "null";
  }

  return `[${scopes.map((s) => `'${s}'`).join(", ")}]`;
}

registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

function templateClientCredentialsMethodName(id: string): string {
  if (id === "") {
    id = "global";
  }
  return `getCredentials${caser().ToPascal(sanitizeName(id))}`;
}

// @ts-ignore
function templateClientCredentialsSecurityAccess(
  clientCredentialsAccess: ClientCredentialsSecurityAccess,
): string {
  const methodCall = (id: string) => {
    const methodName = templateClientCredentialsMethodName(id);
    return `return $this->${methodName}($context->securitySource);`;
  };

  if (clientCredentialsAccess.every((a) => a.id === "")) {
    return methodCall("global");
  }

  let hasGlobalOAuth2 = false;
  const lines = [`switch ($context->operationID) {`];

  for (const { id } of clientCredentialsAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
    } else {
      lines.push(indentString(`case '${id}':`, 1));
      lines.push(indentString(methodCall(id), 2));
    }
  }

  lines.push(indentString("default:", 1));
  lines.push(
    indentString(hasGlobalOAuth2 ? methodCall("global") : "return null;", 2),
  );
  lines.push("}");

  return lines.join("\n");
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccess",
  templateClientCredentialsSecurityAccess,
);

// @ts-ignore
function templateClientCredentialsSecurityAccessFunction(
  clientCredentialsSecurity: ClientCredentialsSecurity,
): string {
  const { id, securityType, optional } = clientCredentialsSecurity;
  const securityFields = getSecurityFields(securityType);
  let fieldName = `$security`;
  let securitySchemeType = securityType;

  let fieldSetPredicate = "";
  let nullSecurityCheck = "";
  if (optional) {
    nullSecurityCheck = `

    if (${fieldName} == null) {
        return null;
    }`;
  }

  if (!hasClientCredentialsOnly(securityFields)) {
    const parts = getOAuth2FlowFields("client_credentials", securityFields);
    if (!parts.length) {
      throw new Error("No client credentials field found");
    }

    fieldSetPredicate =
      parts
        .filter((f) => f.Type.Type.toString() === "class")
        .map((f) => {
          securitySchemeType = f.Type;
          fieldName += `->${sanitizeFieldName(f.Name)}`;
          return `${fieldName} == null`;
        })
        .join(" || ") + " || ";
  }

  fieldSetPredicate += `${fieldName}->clientID == null || ${fieldName}->clientSecret == null || ${fieldName}->tokenURL == null`;

  const scopes = hasOverridableOAuth2Scopes(securitySchemeType)
    ? `${fieldName}->scopes`
    : "null";

  const methodName = templateClientCredentialsMethodName(id);
  const typeName = sanitizeClass(
    securityType,
    "hooks",
    true,
    securityType.OutputLocation === ""
      ? Qualification.ANNOTATION
      : Qualification.TYPE,
  );

  return `/**
 * @param  \\Closure  $securitySource
 * @return Credentials|null
 */
private function ${methodName}(\\Closure $securitySource): ?Credentials
{
    $security = $securitySource();${nullSecurityCheck}

    if (! $security instanceof ${typeName}) {
        throw new \\InvalidArgumentException('Invalid security type: expected ${typeName}, got '.get_class($security));
    }

    if (${fieldSetPredicate}) {
        return null;
    }

    return new Credentials(
        ${fieldName}->clientID,
        ${fieldName}->clientSecret,
        ${fieldName}->tokenURL,
        ${scopes}
    );
}`;
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);
