function flattenSecurityObject(
  typeDef: TypeDef,
  parentIsOptional: boolean = false,
): FieldDef[] {
  const results: FieldDef[] = [];
  if (!typeDef || !typeDef.Fields || !typeDef.Fields.length) {
    return [];
  }
  for (const field of typeDef.Fields) {
    if (field.Const) {
      continue;
    }

    if (field.Type.Type.toString() === "class") {
      const subResults = flattenSecurityObject(field.Type, field.Optional);
      subResults.map((fieldDef) => {
        fieldDef.Optional = parentIsOptional || fieldDef.Optional;
      });
      results.push(...subResults);
      continue;
    }

    results.push({
      ...field,
      Optional: parentIsOptional || field.Optional,
    });
  }

  return results;
}

function getProviderSecurityAttributes(security: TypeDef): ProviderAttributes {
  const attributes: ProviderAttributes = {};

  if (!security) {
    return attributes;
  }

  const fieldDefs = flattenSecurityObject(security);

  for (const fieldDef of fieldDefs) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    const environmentVariable = getEnvironmentVariable(attributeName);
    const optional =
      fieldDef.Optional || fieldDef.Default?.Value || environmentVariable;
    const attribute = ProviderAttributeFromFieldDef(fieldDef, {
      markdownDescription: getSecurityAttributeDescription(
        fieldDef,
        environmentVariable,
      ),
      optional: optional,
      required: !optional,
      sensitive: true,
    });

    attributes[attributeName] = attribute;
  }

  return attributes;
}

/**
 * Returns the description for a security attribute.
 */
function getSecurityAttributeDescription(
  fieldDef: FieldDef,
  environmentVariable: string | undefined,
): string {
  const description: string[] = [];

  if (fieldDef.Comments?.Description) {
    description.push(fieldDef.Comments.Description);
  } else if (fieldDef.Type.Comments?.Description) {
    description.push(fieldDef.Type.Comments.Description);
  }

  if (environmentVariable) {
    description.push(
      "Configurable via environment variable `" + environmentVariable + "`",
    );
  }

  return description.join(". ") + ".";
}

/**
 * Returns true if the provider has security attributes.
 */
function hasProviderSecurityAttributes(
  providerContext: ProviderContext,
): boolean {
  const securityAttributes = getProviderSecurityAttributes(
    providerContext.AST.MainSDK.Security?.Type,
  );

  return Object.keys(securityAttributes).length > 0;
}

/**
 * Templates the provider Configure() method with the Go SDK security
 * configuration based on provider configuration and environment variables.
 */
function templateProviderConfigureSecurity(security: FieldDef) {
  if (!security?.Type?.Fields.length) {
    return "";
  }

  const symbolManager = makeSymbolMananger();
  const templateSecurityType = templateType(security.Type);

  addGenImport(`${getSDKPackage()}/models/${security.Type.Scope}`);
  const result: string[] = [`security := ${templateSecurityType}{}`];

  for (const originalFieldDef of security.Type.Fields) {
    const fieldDef = {
      ...originalFieldDef,
      Optional: security.Optional || originalFieldDef.Optional,
    };

    if (fieldDef.Const) {
      continue;
    }

    const securityAnno = fieldDef.Annotations?.Get(
      "security",
    ) as SecurityAnnotation;

    if (!securityAnno) {
      logger().Warn(
        `Security annotation not found inside security ${fieldDef.Name} field. Skipping field configuration in provider.Configure() method.`,
      );

      continue;
    }

    // Handle various security data situations:
    //  - apiKey security type
    //  - bearer security type
    //  - hoisted single security (e.g. username and password if only basic auth)
    if (fieldDef.Type.Type.toString() === "string") {
      result.push(
        templateProviderConfigureSecurityAttribute(
          symbolManager,
          fieldDef,
          `security`,
          fieldDef.Optional,
        ),
      );

      continue;
    }

    const fieldName = sanitizeFieldName(fieldDef.Name);

    // Handle class-based security
    switch (securityAnno.SecType) {
      case "http":
        switch (securityAnno.SubType) {
          case "basic":
            const basicAccessor = getPluralizedVarSymbolName(
              symbolManager,
              fieldName,
            );

            result.push(
              ``,
              `${basicAccessor} := ${
                fieldDef.Optional ? "&" : ""
              }${templateType(fieldDef.Type)}{}`,
            );

            for (const basicFieldDef of fieldDef.Type.Fields) {
              result.push(
                templateProviderConfigureSecurityAttribute(
                  symbolManager,
                  basicFieldDef,
                  basicAccessor,
                  true, // Do not require both username and password to be set.
                ),
              );
            }

            result.push(
              ``,
              `if ${basicAccessor}.Username != "" || ${basicAccessor}.Password != "" {`,
              `security.${fieldName} = ${basicAccessor}`,
              `}`,
            );

            break;
          case "custom":
            const customAccessor = getPluralizedVarSymbolName(
              symbolManager,
              fieldName,
            );

            result.push(
              ``,
              `${customAccessor} := ${
                fieldDef.Optional ? "&" : ""
              }${templateType(fieldDef.Type)}{}`,
            );

            for (const customFieldDef of fieldDef.Type.Fields) {
              result.push(
                templateProviderConfigureSecurityAttribute(
                  symbolManager,
                  customFieldDef,
                  customAccessor,
                  // Safest to introduce as optional for now (where errors are
                  // raised in the custom security logic, for example). The
                  // required-ness could potentially be determined based on the
                  // custom security schema and if this scheme is the only
                  // option.
                  true,
                ),
              );
            }

            // Always let the custom security logic determine behavior with
            // empty/missing fields.
            result.push(``, `security.${fieldName} = ${customAccessor}`);

            break;
          default:
            logger().Warn(
              `Unsupported security type ${securityAnno.SecType} subtype ${securityAnno.SubType} for ${fieldDef.Name} field. Skipping configuration in provider.Configure() method.`,
            );
        }

        break;
      case "oauth2":
        switch (securityAnno.SubType) {
          case "client_credentials":
            const clientCredentialsAccessor = getPluralizedVarSymbolName(
              symbolManager,
              fieldName,
            );

            result.push(
              ``,
              `${clientCredentialsAccessor} := ${
                fieldDef.Optional ? "&" : ""
              }${templateType(fieldDef.Type)}{}`,
            );

            for (const oauthFieldDef of fieldDef.Type.Fields) {
              result.push(
                templateProviderConfigureSecurityAttribute(
                  symbolManager,
                  oauthFieldDef,
                  clientCredentialsAccessor,
                  fieldDef.Optional || oauthFieldDef.Optional,
                ),
              );
            }

            result.push(
              ``,
              `if ${clientCredentialsAccessor}.ClientID != "" && ${clientCredentialsAccessor}.ClientSecret != "" {`,
              `security.${fieldName} = ${clientCredentialsAccessor}`,
              `}`,
            );

            break;
          default:
            logger().Warn(
              `Unsupported security type ${securityAnno.SecType} subtype ${securityAnno.SubType} for ${fieldDef.Name} field. Skipping configuration in provider.Configure() method.`,
            );
        }

        break;
      default:
        logger().Warn(
          `Unsupported security type ${securityAnno.SecType} for ${fieldDef.Name} field. Skipping configuration in provider.Configure() method.`,
        );
    }
  }

  return result.join("\n");
}

registerTemplateFunc(
  "templateProviderConfigureSecurity",
  templateProviderConfigureSecurity,
);

/**
 * Templates the provider Configure() method for individual security fields from
 * Terraform to the Go SDK.
 */
function templateProviderConfigureSecurityAttribute(
  symbolManager: Record<string, boolean>,
  fieldDef: FieldDef,
  sdkAccessor: string,
  attributeIsOptional: boolean,
): string {
  const attributeName = sanitizeTFStateName(fieldDef.Name);
  const envVariable = getEnvironmentVariable(attributeName);
  const fieldName = sanitizeFieldName(fieldDef.Name);
  const frameworkType = FrameworkTypeFromFieldDef(fieldDef);
  const result: string[] = [];
  sdkAccessor += `.${fieldName}`;

  result.push(
    ``,
    frameworkType.templateTerraformToSDK(
      symbolManager,
      fieldName,
      fieldDef.Type,
      fieldDef.Type.Type.toString() === "enum",
      fieldDef.Optional,
      sdkAccessor,
      `data.${fieldName}`,
      false,
      "resp.Diagnostics",
    ),
  );

  if (envVariable) {
    addGenImport("os");

    const check = fieldDef.Optional
      ? `${sdkAccessor} == nil`
      : `${sdkAccessor} == ""`;
    const variableName = `${sanitizeVariableName(fieldDef.Name)}EnvVar`;

    result.push(
      ``,
      `if ${variableName} := os.Getenv("${envVariable}"); ${check} && ${variableName} != "" {`,
      `${sdkAccessor} = ${fieldDef.Optional ? "&" : ""}${variableName}`,
      `}`,
    );
  }

  if (!attributeIsOptional) {
    const check = fieldDef.Optional
      ? `${sdkAccessor} == nil || *${sdkAccessor} == ""`
      : `${sdkAccessor} == ""`;
    const details = envVariable
      ? `"Either the environment variable ${envVariable} or provider configuration ${attributeName} attribute must be configured.",`
      : `"Provider configuration ${attributeName} attribute must be configured.",`;

    result.push(
      ``,
      `if ${check} {`,
      `resp.Diagnostics.AddError(`,
      `"Missing Provider Security Configuration",`,
      details,
      `)`,
      `}`,
    );
  }

  return result.join("\n");
}

/**
 * Templates operation security with the Go SDK security configuration based on
 * schema configuration.
 */
function templateOperationSecurity(
  security: FieldDef,
  securityAccessor: string,
  symbolManager: Record<string, boolean>,
) {
  if (!security?.Type?.Fields.length) {
    return "";
  }

  const templateSecurityType = templateType(security.Type);

  addGenImport(`${getSDKPackage()}/models/${security.Type.Scope}`);
  const result: string[] = [`${securityAccessor} := ${templateSecurityType}{}`];

  for (const originalFieldDef of security.Type.Fields) {
    const fieldDef = {
      ...originalFieldDef,
      Optional: security.Optional || originalFieldDef.Optional,
    };

    if (fieldDef.Const) {
      continue;
    }

    const securityAnno = fieldDef.Annotations?.Get(
      "security",
    ) as SecurityAnnotation;

    if (!securityAnno) {
      logger().Warn(
        `Security annotation not found inside security ${fieldDef.Name} field. Skipping field configuration in operation logic.`,
      );

      continue;
    }

    // Handle various security data situations:
    //  - apiKey security type
    //  - bearer security type
    //  - hoisted single security (e.g. username and password if only basic auth)
    if (fieldDef.Type.Type.toString() === "string") {
      result.push(
        templateOperationSecurityAttribute(
          symbolManager,
          fieldDef,
          securityAccessor,
          fieldDef.Optional,
        ),
      );

      continue;
    }

    const fieldName = sanitizeFieldName(fieldDef.Name);

    // Handle class-based security
    switch (securityAnno.SecType) {
      case "http":
        switch (securityAnno.SubType) {
          case "basic":
            const basicAccessor = getPluralizedVarSymbolName(
              symbolManager,
              fieldName,
            );

            addGenImport(`${getSDKPackage()}/models/${fieldDef.Type.Scope}`);

            result.push(
              ``,
              `${basicAccessor} := ${
                fieldDef.Optional ? "&" : ""
              }${templateType(fieldDef.Type)}{}`,
            );

            for (const basicFieldDef of fieldDef.Type.Fields) {
              result.push(
                templateOperationSecurityAttribute(
                  symbolManager,
                  basicFieldDef,
                  basicAccessor,
                  true, // Do not require both username and password to be set.
                ),
              );
            }

            result.push(
              ``,
              `if ${basicAccessor}.Username != "" || ${basicAccessor}.Password != "" {`,
              `${securityAccessor}.${fieldName} = ${basicAccessor}`,
              `}`,
            );

            break;
          case "custom":
            const customAccessor = getPluralizedVarSymbolName(
              symbolManager,
              fieldName,
            );

            addGenImport(`${getSDKPackage()}/models/${fieldDef.Type.Scope}`);

            result.push(
              ``,
              `${customAccessor} := ${
                fieldDef.Optional ? "&" : ""
              }${templateType(fieldDef.Type)}{}`,
            );

            for (const customFieldDef of fieldDef.Type.Fields) {
              result.push(
                templateOperationSecurityAttribute(
                  symbolManager,
                  customFieldDef,
                  customAccessor,
                  // Safest to introduce as optional for now (where errors are
                  // raised in the custom security logic, for example). The
                  // required-ness could potentially be determined based on the
                  // custom security schema and if this scheme is the only
                  // option.
                  true,
                ),
              );
            }

            // Always let the custom security logic determine behavior with
            // empty/missing fields.
            result.push(
              ``,
              `${securityAccessor}.${fieldName} = ${customAccessor}`,
            );

            break;
          default:
            logger().Warn(
              `Unsupported security type ${securityAnno.SecType} subtype ${securityAnno.SubType} for ${fieldDef.Name} field. Skipping configuration in provider.Configure() method.`,
            );
        }

        break;
      case "oauth2":
        switch (securityAnno.SubType) {
          case "client_credentials":
            const clientCredentialsAccessor = getPluralizedVarSymbolName(
              symbolManager,
              fieldName,
            );

            addGenImport(`${getSDKPackage()}/models/${fieldDef.Type.Scope}`);

            result.push(
              ``,
              `${clientCredentialsAccessor} := ${
                fieldDef.Optional ? "&" : ""
              }${templateType(fieldDef.Type)}{}`,
            );

            for (const oauthFieldDef of fieldDef.Type.Fields) {
              result.push(
                templateOperationSecurityAttribute(
                  symbolManager,
                  oauthFieldDef,
                  clientCredentialsAccessor,
                  fieldDef.Optional || oauthFieldDef.Optional,
                ),
              );
            }

            result.push(
              ``,
              `if ${clientCredentialsAccessor}.ClientID != "" && ${clientCredentialsAccessor}.ClientSecret != "" {`,
              `${securityAccessor}.${fieldName} = ${clientCredentialsAccessor}`,
              `}`,
            );

            break;
          default:
            logger().Warn(
              `Unsupported security type ${securityAnno.SecType} subtype ${securityAnno.SubType} for ${fieldDef.Name} field. Skipping configuration in provider.Configure() method.`,
            );
        }

        break;
      default:
        logger().Warn(
          `Unsupported security type ${securityAnno.SecType} for ${fieldDef.Name} field. Skipping configuration in provider.Configure() method.`,
        );
    }
  }

  return result.join("\n");
}

/**
 * Templates the operation security logic for individual security fields from
 * Terraform to the Go SDK.
 */
function templateOperationSecurityAttribute(
  symbolManager: Record<string, boolean>,
  fieldDef: FieldDef,
  sdkAccessor: string,
  attributeIsOptional: boolean,
): string {
  const attributeName = sanitizeTFStateName(fieldDef.Name);
  // const envVariable = getEnvironmentVariable(attributeName);
  const fieldName = sanitizeFieldName(fieldDef.Name);
  const frameworkType = FrameworkTypeFromFieldDef(fieldDef);
  const result: string[] = [];
  sdkAccessor += `.${fieldName}`;

  result.push(
    ``,
    frameworkType.templateTerraformToSDK(
      symbolManager,
      fieldName,
      fieldDef.Type,
      fieldDef.Type.Type.toString() === "enum",
      fieldDef.Optional,
      sdkAccessor,
      `data.${fieldName}`,
      false,
      "resp.Diagnostics",
    ),
  );

  // if (envVariable) {
  //   addGenImport("os");

  //   const check = fieldDef.Optional
  //     ? `${sdkAccessor} == nil`
  //     : `${sdkAccessor} == ""`;
  //   const variableName = `${sanitizeVariableName(fieldDef.Name)}EnvVar`;

  //   result.push(
  //     ``,
  //     `if ${variableName} := os.Getenv("${envVariable}"); ${check} && ${variableName} != "" {`,
  //     `${sdkAccessor} = ${fieldDef.Optional ? "&" : ""}${variableName}`,
  //     `}`,
  //   );
  // }

  if (!attributeIsOptional) {
    const check = fieldDef.Optional
      ? `${sdkAccessor} == nil || *${sdkAccessor} == ""`
      : `${sdkAccessor} == ""`;
    // const details = envVariable
    //   ? `"Either the environment variable ${envVariable} or resource configuration ${attributeName} attribute must be configured.",`
    const details = `"Provider configuration ${attributeName} attribute must be configured.",`;

    result.push(
      ``,
      `if ${check} {`,
      `resp.Diagnostics.AddError(`,
      `"Missing Provider Security Configuration",`,
      details,
      `)`,
      `}`,
    );
  }

  return result.join("\n");
}

function getSensitiveHeaders(): Record<string, boolean> {
  const primitives = flattenSecurityObject(
    context.Global.AST.MainSDK.Security?.Type,
  );
  const builder: Record<string, boolean> = {};
  for (const primitive of primitives) {
    const fieldAnno = primitive.Annotations.Get(
      "security",
    ) as SecurityAnnotation;
    if (!fieldAnno || !fieldAnno.FieldName) {
      continue;
    }
    switch (fieldAnno.SecType) {
      case "apiKey":
        // FieldName is the apiKey security name value.
        builder[fieldAnno.FieldName] = true;
        break;
      case "http":
      case "oauth2":
      case "openIdConnect":
        // These always use the HTTP Authorization header.
        builder["Authorization"] = true;
        break;
    }
  }

  // Include any additional sensitive headers from generation configuration.
  // This supports custom security schemes where headers are set in hook code
  // and not automatically detected from the OpenAPI security definition.
  const debugLogging = context.Global.Config.DebugLogging as
    | { additionalSensitiveHttpHeaders?: string[] }
    | undefined;
  const additionalHeaders = debugLogging?.additionalSensitiveHttpHeaders;
  if (additionalHeaders) {
    for (const header of additionalHeaders) {
      builder[header] = true;
    }
  }

  return builder;
}

function redactSecurityHeaders(httpResponseVar: string): string {
  let sensitiveHeaders = getSensitiveHeaders();
  let txtResponse = "";

  for (const key of Object.keys(sensitiveHeaders).sort()) {
    txtResponse += `${templateIndent(
      1,
    )}if v := ${httpResponseVar}.Request.Header.Get(${JSON.stringify(
      key,
    )}); v != "" {\n`;
    txtResponse += `${templateIndent(
      2,
    )}response.Request.Header.Set(${JSON.stringify(key)}, "(sensitive)")\n`;
    txtResponse += `${templateIndent(1)}}\n`;
  }
  return txtResponse;
}
registerTemplateFunc("redactSecurityHeaders", redactSecurityHeaders);

function redactSecurityFields(symbolName: string): string {
  let sensitiveHeaders = getSensitiveHeaders();
  let txtResponse = "";

  for (const key of Object.keys(sensitiveHeaders).sort()) {
    txtResponse += `${templateIndent(
      1,
    )}if _, ok := ${symbolName}[${JSON.stringify(key)}]; ok {\n`;
    txtResponse += `${templateIndent(2)}${symbolName}[${JSON.stringify(
      key,
    )}] = "(sensitive)"\n`;
    txtResponse += `${templateIndent(1)}}\n`;
  }
  return txtResponse;
}
registerTemplateFunc("redactSecurityFields", redactSecurityFields);

// Returns a Go slice literal containing all sensitive header names.
// Used by utils_test.go to verify header redaction.
function sensitiveHeadersList(): string {
  const sensitiveHeaders = getSensitiveHeaders();
  const headers = Object.keys(sensitiveHeaders).sort();
  if (headers.length === 0) {
    return "[]string{}";
  }
  const items = headers.map((h) => JSON.stringify(h)).join(", ");
  return `[]string{${items}}`;
}
registerTemplateFunc("sensitiveHeadersList", sensitiveHeadersList);
