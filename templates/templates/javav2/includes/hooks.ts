// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];
  const registrationFile = `${getSourceDirectory()}/hooks/${hooksInitializationClass()}.java`;

  // If no hooks are registered (handles no content transparently), create/overwrite the file
  if (!hasRegisteredHooks() && useHooks()) {
    jobs.push(
      createTemplateFileJob("SDKHooks.java.stmpl", registrationFile, {}),
    );
  }

  if (hasClientCredentials()) {
    jobs.push(
      createTemplateFileJob(
        "ClientCredentialsHook.java.stmpl",
        `${getSourceDirectory()}/hooks/ClientCredentialsHook.java`,
        {},
      ),
    );
  }
  if (hasOAuth2PasswordFlow()) {
    jobs.push(
      createTemplateFileJob(
        "OAuth2PasswordHook.java.stmpl",
        `${getSourceDirectory()}/hooks/OAuth2PasswordHook.java`,
        {},
      ),
    );
  }
  const oauth2Config = collectAvailableOAuth2Scopes();
  if (hasAvailableOAuth2Scopes(oauth2Config)) {
    jobs.push(
      createTemplateFileJob(
        "OAuth2Scopes.java.stmpl",
        `${getSourceDirectory()}/hooks/OAuth2Scopes.java`,
        oauth2Config,
      ),
    );
  }
  return jobs;
}

function hooksInitializationClass() {
  return "SDKHooks";
}

function useHooks(): boolean {
  return accountHasFeatureAccess("sdkHooks");
}

registerTemplateFunc("useHooks", useHooks);

// @ts-ignore
function templateOAuth2Scopes(op: Operation): string {
  const scopes = getRequiredOAuth2Scopes(op);
  if (scopes === null) {
    return `${javaImportOptional()}.empty()`;
  }

  return `${javaImportOptional()}.of(${templateJdkCall("List.of")}(${scopes
    .map((s) => `"${s}"`)
    .join(", ")}))`;
}

registerTemplateFunc("templateOAuth2Scopes", templateOAuth2Scopes);

function templateOAuth2PasswordSecurityField(
  sourceKey: string,
  keyName: string,
  mandatory: boolean,
): string {
  const securityFields = getGlobalSecurityFields();

  if (hasOneOAuth2GrantTypeOnly("password", securityFields)) {
    const mapper = mandatory ? "map" : "flatMap";
    const indent = "                        ";
    return `${sourceKey}.get().getSecurity()\n${indent}.oauth2()\n${indent}.map(x -> (OAuth2PasswordFlow) x.value())\n${indent}.${mapper}(x -> x.${keyName}())`;
  }

  const parts = getOAuth2FlowFields("password", securityFields);
  if (!parts.length) {
    throw new Error("No oauth2:password field found");
  }

  return `${sourceKey}.get().getSecurity().${parts
    .map((f) => sanitizeFieldName(f.Name) + "()")
    .join(".")}.map(x -> x.value().${keyName}())`;
}

registerTemplateFunc(
  "templateOAuth2PasswordSecurityField",
  templateOAuth2PasswordSecurityField,
);

function templateClientCredentialsMethodName(id: string): string {
  if (id === "") {
    id = "global";
  }
  return sanitizePrivateMethodName(`getCredentials${id}`);
}

// @ts-ignore
function templateClientCredentialsSecurityAccess(
  clientCredentialsAccess: ClientCredentialsSecurityAccess,
): string {
  const methodCall = (sec?: ClientCredentialsSecurity) => {
    if (sec !== undefined) {
      const methodName = templateClientCredentialsMethodName(sec.id);
      const className = sanitizeClassName(sec.securityType.Name);
      return `return ${methodName}((${className}) security);`;
    }

    return `return ${templateClientCredentialsMethodName(
      "global",
    )}((Security) security);`;
  };

  if (clientCredentialsAccess.every((a) => a.id === "")) {
    return methodCall();
  }

  let hasGlobalOAuth2 = false;
  const lines = [`switch(context.operationId()) {`];
  for (const sec of clientCredentialsAccess) {
    if (sec.id === "") {
      hasGlobalOAuth2 = true;
    } else {
      lines.push(`case "${sec.id}":`);
      lines.push(indentString(methodCall(sec), 1));
    }
  }

  lines.push("default:");
  lines.push(
    indentString(
      hasGlobalOAuth2 ? methodCall() : "return Optional.empty();",
      1,
    ),
  );
  lines.push("}");

  return lines.join("\n");
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccess",
  templateClientCredentialsSecurityAccess,
);

// Optional.isEmpty() is Java 11+; Java 8 uses !isPresent().
function optionalIsEmptyExpr(expr: string): string {
  return isJava8() ? `!${expr}.isPresent()` : `${expr}.isEmpty()`;
}

// @ts-ignore
function templateClientCredentialsSecurityAccessFunction(
  clientCredentialsSecurity: ClientCredentialsSecurity,
): string {
  const { id, securityType } = clientCredentialsSecurity;
  const securityFields = getSecurityFields(securityType);
  let fieldName = "security";
  let fieldSetCheck = "";
  let optionalCredentials = false;
  let securitySchemeType = securityType;

  if (hasClientCredentialsOnly(securityFields)) {
    optionalCredentials = securityFields.every((f) => f.Optional);
  } else {
    const parts = getOAuth2FlowFields("client_credentials", securityFields);
    if (!parts.length) {
      throw new Error("No client credentials field found");
    }

    parts
      .filter((f) => f.Type.Type.toString() === "class")
      .forEach((f) => {
        fieldName += `.${sanitizeFieldName(f.Name)}()`;
        securitySchemeType = f.Type;

        if (f.Optional) {
          fieldSetCheck += `

if (${optionalIsEmptyExpr(fieldName)}) {
    return Optional.empty();
}`;

          fieldName += ".get()";
        }
      });
  }

  const emptyCheck = (expr: string) =>
    optionalCredentials ? optionalIsEmptyExpr(expr) : `${expr} == null`;
  const fieldGet = optionalCredentials ? ".get()" : "";
  fieldSetCheck += `

if (${emptyCheck(`${fieldName}.clientID()`)} || ${emptyCheck(
    `${fieldName}.clientSecret()`,
  )} || ${emptyCheck(`${fieldName}.tokenURL()`)}) {
      return Optional.empty();
}`;

  const scopes = hasOverridableOAuth2Scopes(securitySchemeType)
    ? `${fieldName}.scopes()${fieldGet}`
    : "null";

  const methodName = templateClientCredentialsMethodName(id);
  const className = sanitizeClassName(securityType.Name);

  return `private static Optional < Credentials > ${methodName} (${className} security) {

${fieldSetCheck}

return Optional.of(new Credentials(
      ${fieldName}.clientID()${fieldGet},
      ${fieldName}.clientSecret()${fieldGet},
      ${fieldName}.tokenURL()${fieldGet},
      ${scopes}));
}`;
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);

/**
 * Detects if hooks are registered in the existing SDKHooks.java file
 * @returns true if hooks are found, false otherwise
 */
function hasRegisteredHooks(): boolean {
  const registrationFile = `${getSourceDirectory()}/hooks/${hooksInitializationClass()}.java`;

  // Look for hook registration patterns
  const hookPatterns = [
    /hooks\.registerBeforeRequest\(/,
    /hooks\.registerAfterSuccess\(/,
    /hooks\.registerAfterError\(/,
  ];

  try {
    const content = readFile(registrationFile);
    if (!content) {
      return false; // File doesn't exist
    }
    // Split content into lines and check each line individually
    const lines = content.split("\n");

    for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
      const line = lines[lineIndex].trim();

      // Skip empty lines and commented lines (both // and /* */ style comments)
      if (
        !line ||
        line.startsWith("//") ||
        line.startsWith("/*") ||
        line.startsWith("*")
      ) {
        continue;
      }

      // Check if any hook registration pattern is found in this non-commented line
      for (const pattern of hookPatterns) {
        if (pattern.test(line)) {
          return true;
        }
      }
    }
  } catch (error) {}

  return false;
}
