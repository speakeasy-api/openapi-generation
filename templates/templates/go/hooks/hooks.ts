// @ts-ignore
function getHooksJobs(): Job[] {
  const jobs: Job[] = [];

  const registrationFile = `${getInternalPackageName()}/hooks/registration.go`;
  let registrationData = readFile(registrationFile);

  // If we don't already have a registration file, create it
  if (!registrationData && accountHasFeatureAccess("sdkHooks")) {
    jobs.push(
      createTemplateFileJob(
        "hooks/registration.go.stmpl",
        registrationFile,
        {},
      ),
    );
  }

  jobs.push(
    createTemplateFileJob(
      "hooks/hooks.go.stmpl",
      `${getInternalPackageName()}/hooks/hooks.go`,
      {},
    ),
  );

  if (hasClientCredentials()) {
    if (getSharedLocation() == "") {
      // Note: an empty 'shared' location implies that all other import paths
      // are flattened, otherwise circular import errors would occur.
      throw new Error(
        "OAuth2 client credentials flow is not supported with fully flattened import paths. Please update your configuration or contact us for further assistance.",
      );
    }
    jobs.push(
      createTemplateFileJob(
        "hooks/clientcredentials.go.stmpl",
        `${getInternalPackageName()}/hooks/clientcredentials.go`,
        {},
      ),
    );
  }

  if (hasOAuth2PasswordFlow()) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2_password.go.stmpl",
        `${getInternalPackageName()}/hooks/oauth2_password.go`,
        {},
      ),
    );
  }

  const oauth2Config = collectAvailableOAuth2Scopes();
  if (hasAvailableOAuth2Scopes(oauth2Config)) {
    jobs.push(
      createTemplateFileJob(
        "hooks/oauth2scopes.go.stmpl",
        `${getInternalPackageName()}/hooks/oauth2scopes.go`,
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
    return `return c.getCredentialsGlobal(sec)`;
  }

  let hasGlobalOAuth2 = false;
  const lines = [`switch ctx.OperationID {`];

  for (const { id } of credentialAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    const methodName = sanitizePrivateMethodName(`getCredentials_${id}`);
    lines.push(`case "${id}":`);
    lines.push(indentString(`return c.${methodName}(sec)`, 1));
  }

  if (hasGlobalOAuth2) {
    lines.push(`default:`);
    lines.push(indentString(`return c.getCredentialsGlobal(sec)`, 1));
    lines.push(`}`);
  } else {
    lines.push(`}`);
    lines.push(`return nil, nil`);
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
  isOptionalSecurity?: boolean,
): string {
  let ptrSymbol = "";
  let fieldName = "security";
  let fieldSetCheck = "";
  let securitySchemeType = securityType;

  const fields = getSecurityFields(securityType);

  if (hasClientCredentialsOnly(fields)) {
    if (fields.every((f) => f.Optional)) {
      ptrSymbol = "*";
      fieldSetCheck = `${fieldName}.ClientID == nil || ${fieldName}.ClientSecret == nil`;
    }
  } else {
    const parts = getOAuth2FlowFields("client_credentials", fields);
    if (!parts.length) {
      throw new Error("No client credentials field found");
    }

    // Only traverse through class-type (container) fields — skip the leaf
    // OAuth2 scheme field which is a string, not a struct.
    const containerParts = parts.filter(
      (f) => f.Type.Type.toString() === "class",
    );

    fieldSetCheck = containerParts
      .filter((f) => f.Optional)
      .map((f) => {
        return `${fieldName}.${sanitizeFieldName(f.Name)} == nil`;
      })
      .join(" || ");

    if (containerParts.length > 0) {
      fieldName +=
        "." + containerParts.map((f) => sanitizeFieldName(f.Name)).join(".");
      securitySchemeType = containerParts[containerParts.length - 1].Type;
    }
  }

  const scopes = hasOverridableOAuth2Scopes(securitySchemeType)
    ? `${fieldName}.Scopes`
    : "nil";

  const typeName = sanitizeType(securityType, false, "");

  // When security is optional, the SDK method passes a pointer (*SecurityType).
  // We need to handle both pointer and value type assertions since the hook
  // receives `any` and different callers may pass either form.
  let securityAccess: string;
  if (isOptionalSecurity) {
    securityAccess = `secPtr, ptrOk := sec.(*${typeName})
var security ${typeName}
if ptrOk {
    if secPtr == nil {
        return nil, nil
    }
    security = *secPtr
} else {
    var ok bool
    security, ok = sec.(${typeName})
    if !ok {
        return nil, fmt.Errorf("unexpected security type: %T", sec)
    }
}`;
  } else {
    securityAccess = `security, ok := sec.(${typeName})

if !ok {
    return nil, fmt.Errorf("unexpected security type: %T", sec)
}`;
  }

  if (fieldSetCheck) {
    securityAccess += `

if (${fieldSetCheck}) {
    return nil, nil
}`;
  }

  addImport("reflect");
  securityAccess += `
  secType := reflect.TypeOf(${fieldName})
  if secType.Kind() == reflect.Ptr {
    secType = secType.Elem()
  }
  secValue := reflect.ValueOf(${fieldName})
	if secValue.Kind() == reflect.Ptr {
		secValue = secValue.Elem()
	}
if ${fieldName}.TokenURL == ${ptrSymbol == "*" ? "nil" : `""`} {

    tokenURLField, ok := secType.FieldByName("TokenURL")
    if !ok {
      return nil, fmt.Errorf("TokenURL is required for security type %s", secType.Name())
    }
    tokenURLDefault := tokenURLField.Tag.Get("default")
    ${fieldName}.TokenURL = ${ptrSymbol == "*" ? "&" : ""}tokenURLDefault
}

additionalProperties := make(map[string]string)
for i := 0; i < secType.NumField(); i++ {
		field := secType.Field(i)
		if field.Name != "TokenURL" && field.Name != "ClientID" && field.Name != "ClientSecret" && field.Name != "Scopes" {
			// Get the field value using reflection
			fieldValue := secValue.Field(i)
			if fieldValue.IsValid() {
				tag := field.Tag.Get("security")
				parts := strings.Split(tag, ",")
				for _, part := range parts {
					if strings.HasPrefix(part, "name=") {
						additionalProperties[strings.TrimPrefix(part, "name=")] = fieldValue.String()
						break
					}
				}
			}
		}
	}

return &credentials{
    ClientID:     ${ptrSymbol}${fieldName}.ClientID,
    ClientSecret: ${ptrSymbol}${fieldName}.ClientSecret,
    TokenURL:     ${ptrSymbol}${fieldName}.TokenURL,
    Scopes:       ${scopes},
    AdditionalProperties: additionalProperties,
}, nil`;

  return securityAccess;
}

registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunction",
  templateClientCredentialsSecurityAccessFunction,
);

// Wrapper that takes a ClientCredentialsSecurity object and extracts
// the optional flag to pass to the core function.
// @ts-ignore
function templateClientCredentialsSecurityAccessFunctionFromAccess(
  access: ClientCredentialsSecurity,
): string {
  return templateClientCredentialsSecurityAccessFunction(
    access.securityType,
    access.optional === true,
  );
}
registerTemplateFunc(
  "templateClientCredentialsSecurityAccessFunctionFromAccess",
  templateClientCredentialsSecurityAccessFunctionFromAccess,
);

// @ts-ignore
function templateOAuth2PasswordSecurityAccess(
  credentialAccess: OAuth2PasswordSecurityAccess,
): string {
  if (credentialAccess.every((a) => a.id === "")) {
    return `return h.getCredentialsGlobal(ctx, sec)`;
  }

  let hasGlobalOAuth2 = false;
  const lines = [`switch ctx.OperationID {`];

  for (const { id } of credentialAccess) {
    if (id === "") {
      hasGlobalOAuth2 = true;
      continue;
    }

    const methodName = sanitizePrivateMethodName(`getCredentials_${id}`);
    lines.push(`case "${id}":`);
    lines.push(indentString(`return h.${methodName}(ctx, sec)`, 1));
  }

  if (hasGlobalOAuth2) {
    lines.push(`default:`);
    lines.push(indentString(`return h.getCredentialsGlobal(ctx, sec)`, 1));
    lines.push(`}`);
  } else {
    lines.push(`}`);
    lines.push(`return nil, nil`);
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

  const credentialsFieldName = sanitizeFieldName(
    unionField.Type.AssociatedTypes.find(
      (at) => at.OriginalName === "Credentials",
    ).Name,
  );
  const credentialsFieldAccessor = [
    unionFieldAccessor,
    credentialsFieldName,
  ].join(".");

  const tokenFieldName = sanitizeFieldName(
    unionField.Type.AssociatedTypes.find((at) => at.OriginalName === "Token")
      .Name,
  );
  const tokenFieldAccessor = [unionFieldAccessor, tokenFieldName].join(".");

  let result = "";

  if (flattenedSecurity) {
    result = `security, ok := sec.(${sanitizeType(securityType, true, "")})\n`;
  } else {
    result = `security, ok := sec.(${sanitizeType(securityType, false, "")})\n`;
  }

  result += `
if !ok {
    return nil, fmt.Errorf("unexpected security type: %T", sec)
}`;

  if (flattenedSecurity) {
    result += `

if security == nil {
  return nil, nil
}
`;
  }

  if (unionField.Optional) {
    result += `

if ${unionFieldAccessor} == nil {
    return nil, nil
}
`;
  }

  addImport("reflect");
  result += `
if ${tokenFieldAccessor} != nil {
    return utils.NewOAuth2AccessTokenCredentials(*${tokenFieldAccessor}), nil
}

if ${credentialsFieldAccessor}.TokenURL == "" {
    secType := reflect.TypeOf(${credentialsFieldAccessor})

    if secType.Kind() == reflect.Ptr {
      secType = secType.Elem()
    }

    tokenURLField, ok := secType.FieldByName("TokenURL")

    if !ok {
      return nil, fmt.Errorf("TokenURL is required for security type %s", secType.Name())
    }

    ${credentialsFieldAccessor}.TokenURL = tokenURLField.Tag.Get("default")
}

tokenEndpoint, err := h.tokenEndpoint(ctx.BaseURL, ${credentialsFieldAccessor}.TokenURL)

if err != nil {
  return nil, err
}

return utils.NewOAuth2ResourceOwnerPasswordCredentials(
  tokenEndpoint,
  ${credentialsFieldAccessor}.Username,
  ${credentialsFieldAccessor}.Password,
  ${credentialsFieldAccessor}.ClientID,
  ${credentialsFieldAccessor}.ClientSecret,
), nil`;

  return result;
}

registerTemplateFunc(
  "templateOAuth2PasswordSecurityAccessFunction",
  templateOAuth2PasswordSecurityAccessFunction,
);
