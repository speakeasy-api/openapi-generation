/** Customer-enabled additional provider attributes that are not based on the
 *  SDK security model or part of the API definition. For example, provider
 *  configuration to set HTTP client headers.
 *
 *  These are templated into:
 *   * Provider Schema
 *   * Provider Schema model type
 *   * Provider Configure method
 *
 *  The values are sourced from generation configuration (gen.yaml) terraform
 *  section additionalProviderAttributes property. */
type AdditionalProviderAttributesConfig = {
  /** When non-empty, adds a map attribute with the value as the attribute name
   *  for setting HTTP client headers into all requests. */
  httpHeaders?: string;

  /** When non-empty, adds a boolean attribute with the value as the attribute
   *  name for disabling HTTP client TLS verification (useful for running
   *  against test environments with self-signed certificates). */
  tlsSkipVerify?: string;
};

type ProviderContext = GlobalContext & {
  Actions: TerraformAction[];
  Resources: TerraformManagedResource[];
  DataSources: TerraformDataResource[];
  EphemeralResources: TerraformEphemeralResource[];
};

/**
 * Describes relevant provider context for any Terraform resource type.
 */
type ProviderResourceContext = {
  /**
   * Set to true if the provider context has data to pass through during
   * Configure() in addition to the SDK client, such as globals. If enabled, the
   * data is wrapped in the additional ProviderConfigureData struct type.
   */
  HasConfigureData: boolean;

  /**
   * Go SDK type, such as `*sdk.SDK`.
   */
  SDKType: string;
};

/**
 * Returns all provider attributes for the provider context, including:
 *  - server URL/variables
 *  - security
 *  - globals
 *  - additionalProviderAttributes generation configuration
 */
function getProviderAttributes(
  providerContext: ProviderContext,
): ProviderAttributes {
  const attributes: ProviderAttributes = {
    ...getProviderAdditionalAttributes(
      providerContext.Config.AdditionalProviderAttributes,
    ),
    ...getProviderGlobalAttributes(providerContext.AST.MainSDK.Globals),
    ...getProviderSecurityAttributes(
      providerContext.AST.MainSDK.Security?.Type,
    ),
  };

  return {
    ...attributes,
    ...getProviderServerURLAttributes(
      providerContext,
      new Set(Object.keys(attributes)),
    ),
  };
}

/**
 * Returns any provider attributes defined in the generation configuration
 * additionalProviderAttributes property, such as httpHeaders and tlsSkipVerify.
 */
function getProviderAdditionalAttributes(
  additionalProviderAttributesConfig: AdditionalProviderAttributesConfig,
): ProviderAttributes {
  const attributes: ProviderAttributes = {};

  if (!additionalProviderAttributesConfig) {
    return attributes;
  }

  if (additionalProviderAttributesConfig.httpHeaders) {
    const attributeName = sanitizeTFStateName(
      additionalProviderAttributesConfig.httpHeaders,
    );

    attributes[attributeName] = new ProviderMapAttribute(
      new FrameworkStringType(),
      {
        description: "HTTP headers to include in all requests",
        optional: true,
      },
    );
  }

  if (additionalProviderAttributesConfig.tlsSkipVerify) {
    const attributeName = sanitizeTFStateName(
      additionalProviderAttributesConfig.tlsSkipVerify,
    );

    attributes[attributeName] = new ProviderBoolAttribute({
      description: "Disable TLS verification in HTTP client",
      optional: true,
    });
  }

  return attributes;
}

/**
 * Returns the provider attributes for x-speakeasy-globals.
 */
function getProviderGlobalAttributes(
  globals: TypeDef | undefined,
): ProviderAttributes {
  const attributes: ProviderAttributes = {};

  if (!globals) {
    return attributes;
  }

  for (const fieldDef of globals.Fields) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    // As of 2025-06, the AST will always set Optional to true for globals,
    // but this logic is here to ensure correctness if the AST changes.
    const optional =
      fieldDef.Optional ||
      fieldDef.Default?.Value ||
      getEnvironmentVariable(attributeName);
    const attribute = ProviderAttributeFromFieldDef(fieldDef, {
      deprecationMessage: fieldDef.Comments?.DeprecationMessage,
      description: fieldDef.Comments?.Description,
      optional: optional,
      required: !optional,
    });

    attributes[attributeName] = attribute;
  }

  return attributes;
}

/**
 * Returns the provider attributes for server URL handling.
 */
function getProviderServerURLAttributes(
  providerContext: ProviderContext,
  reservedAttributeNames: Set<string>,
): ProviderAttributes {
  const attributes: ProviderAttributes = {};
  const serverVariables = providerContext.AST.MainSDK.Servers?.GetVariables();

  if (!providerContext.Config.BaseServerURL) {
    attributes["server_url"] = new ProviderStringAttribute({
      description: serverURLDescription(providerContext.AST.MainSDK),
      optional: hasServerURL(providerContext.AST.MainSDK),
      required: !hasServerURL(providerContext.AST.MainSDK),
    });
  }

  if (serverVariables) {
    assertServerVariableProviderAttributeNames(
      serverVariables,
      new Set([...reservedAttributeNames, ...Object.keys(attributes)]),
    );

    serverVariables.forEach((variable) => {
      const attributeName = providerServerVariableAttributeName(variable.Name);

      attributes[attributeName] = ProviderAttributeFromTypeDef(variable.Type, {
        markdownDescription: serverURLVariableDescription(variable),
        optional: true,
      });
    });
  }

  return attributes;
}

/**
 * Returns true if the provider context has data to pass through during
 * Configure() in addition to the SDK client, such as globals. If enabled, the
 * data is wrapped in the additional ProviderConfigureData struct type, so this
 * must be checked in resource Configure() methods to prevent compilation
 * errors.
 */
function hasProviderConfigureData(providerContext: ProviderContext): boolean {
  // NOTE: In the future, this may be extended to be a generation configuration
  // option that is enabled by default with newSDKs (along with other features
  // that warrant passing additional data to Configure() methods).
  return hasProviderGlobals(providerContext);
}

registerTemplateFunc("hasProviderConfigureData", hasProviderConfigureData);

/**
 * Returns true if the provider context has x-speakeasy-globals defined.
 * This is used to determine if the provider Configure() method should pass
 * the x-speakeasy-globals data to resource Configure() methods.
 */
function hasProviderGlobals(providerContext: ProviderContext): boolean {
  return Boolean(providerContext.AST.MainSDK.Globals?.Fields.length);
}

/**
 * Templates ProviderConfigureData field initialization.
 */
function templateProviderConfigureData(context: ProviderContext) {
  const globals = context.AST.MainSDK.Globals;
  const result: string[] = [`SDKClient: client,`];

  if (globals) {
    for (const fieldDef of globals.Fields) {
      const fieldName = sanitizeFieldName(fieldDef.Name);

      result.push(`${fieldName}: data.${fieldName},`);
    }
  }

  return result.sort().join("\n");
}

registerTemplateFunc(
  "templateProviderConfigureData",
  templateProviderConfigureData,
);

/**
 * Templates ProviderConfigureData struct fields. Fields are intentionally
 * exported so that they can be accessed in resource Configure() methods outside
 * of the provider package.
 */
function templateProviderConfigureDataStructFields(context: ProviderContext) {
  const globals = context.AST.MainSDK.Globals;
  const sdkType = `*sdk.${sanitizeClassName(context.AST.MainSDK.Type.Name)}`;
  // Use SDKClient as field name to prevent conflicts with any "Client" global.
  const structFields: string[] = [`SDKClient ${sdkType}`];

  if (globals) {
    for (const fieldDef of globals.Fields) {
      const structField = FrameworkTypeFromFieldDef(
        fieldDef,
      ).templateDataModelStructField(fieldDef.Name);

      structFields.push(structField);
    }
  }

  return structFields.sort().join("\n");
}

registerTemplateFunc(
  "templateProviderConfigureDataStructFields",
  templateProviderConfigureDataStructFields,
);

/**
 * Templates the provider Configure() method with the Go SDK x-speakeasy-globals
 * data based on provider configuration and environment variables.
 */
function templateProviderConfigureGlobals(globals: TypeDef | undefined) {
  if (!globals) {
    return "";
  }

  const result: string[] = [];

  for (const fieldDef of globals.Fields) {
    const attributeName = sanitizeTFStateName(fieldDef.Name);
    const envVariable = getEnvironmentVariable(attributeName);
    const fieldName = sanitizeFieldName(fieldDef.Name);

    // Enum types are unwrapped to their underlying type for environment
    // variable parsing and default value handling, matching the framework
    // type resolution in FrameworkTypeFromTypeDef().
    const isEnum = fieldDef.Type.Type.toString() === "enum";
    const enumValues: string[] =
      isEnum && fieldDef.Type.Enum?.Values ? fieldDef.Type.Enum.Values : [];
    const effectiveType = isEnum
      ? fieldDef.Type.Enum.Type.Type.toString()
      : fieldDef.Type.Type.toString();

    if (envVariable) {
      const variableName = sanitizeVariableName(fieldDef.Name);
      const envVariableName = `${variableName}EnvVar`;

      // Best effort
      switch (effectiveType) {
        case "boolean":
          addGenImport("os");
          addGenImport("strconv");

          result.push(
            ``,
            `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
            `${envVariableName}Bool, err := strconv.ParseBool(${envVariableName})`,
            ``,
            `if err != nil {`,
            `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Unable to parse environment variable as boolean: " + err.Error())`,
            `} else {`,
            `data.${fieldName} = types.BoolValue(${envVariableName}Bool)`,
            `}`,
            `}`,
          );

          break;
        case "float32":
          addGenImport("os");
          addGenImport("strconv");

          result.push(
            ``,
            `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
            `${envVariableName}Float32, err := strconv.ParseFloat(${envVariableName}, 32)`,
            ``,
            `if err != nil {`,
            `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Unable to parse environment variable as float32: " + err.Error())`,
          );

          if (isEnum && enumValues.length) {
            addGenImport("slices");

            result.push(
              `} else if !slices.Contains([]float32{${enumValues.join(
                ", ",
              )}}, float32(${envVariableName}Float32)) {`,
              `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Must be one of: ${enumValues.join(
                ", ",
              )}")`,
            );
          }

          result.push(
            `} else {`,
            `data.${fieldName} = types.Float32Value(float32(${envVariableName}Float32))`,
            `}`,
            `}`,
          );

          break;
        case "int32":
          addGenImport("os");
          addGenImport("strconv");

          result.push(
            ``,
            `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
            `${envVariableName}Int32, err := strconv.ParseInt(${envVariableName}, 10, 32)`,
            ``,
            `if err != nil {`,
            `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Unable to parse environment variable as int32: " + err.Error())`,
          );

          if (isEnum && enumValues.length) {
            addGenImport("slices");

            result.push(
              `} else if !slices.Contains([]int32{${enumValues.join(
                ", ",
              )}}, int32(${envVariableName}Int32)) {`,
              `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Must be one of: ${enumValues.join(
                ", ",
              )}")`,
            );
          }

          result.push(
            `} else {`,
            `data.${fieldName} = types.Int32Value(int32(${envVariableName}Int32))`,
            `}`,
            `}`,
          );

          break;
        case "integer":
          addGenImport("os");
          addGenImport("strconv");

          result.push(
            ``,
            `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
            `${envVariableName}Int64, err := strconv.ParseInt(${envVariableName}, 10, 64)`,
            ``,
            `if err != nil {`,
            `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Unable to parse environment variable as int64: " + err.Error())`,
          );

          if (isEnum && enumValues.length) {
            addGenImport("slices");

            result.push(
              `} else if !slices.Contains([]int64{${enumValues.join(
                ", ",
              )}}, ${envVariableName}Int64) {`,
              `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Must be one of: ${enumValues.join(
                ", ",
              )}")`,
            );
          }

          result.push(
            `} else {`,
            `data.${fieldName} = types.Int64Value(${envVariableName}Int64)`,
            `}`,
            `}`,
          );

          break;
        case "number":
          addGenImport("os");
          addGenImport("strconv");

          result.push(
            ``,
            `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
            `${envVariableName}Float64, err := strconv.ParseFloat(${envVariableName}, 64)`,
            ``,
            `if err != nil {`,
            `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Unable to parse environment variable as float64: " + err.Error())`,
          );

          if (isEnum && enumValues.length) {
            addGenImport("slices");

            result.push(
              `} else if !slices.Contains([]float64{${enumValues.join(
                ", ",
              )}}, ${envVariableName}Float64) {`,
              `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Must be one of: ${enumValues.join(
                ", ",
              )}")`,
            );
          }

          result.push(
            `} else {`,
            `data.${fieldName} = types.Float64Value(${envVariableName}Float64)`,
            `}`,
            `}`,
          );

          break;
        case "string":
          addGenImport("os");

          if (isEnum && enumValues.length) {
            addGenImport("slices");

            const sliceValues = enumValues
              .map((v: string) => `"${v}"`)
              .join(", ");
            const displayValues = enumValues
              .map((v: string) => `\\"${v}\\"`)
              .join(", ");

            result.push(
              ``,
              `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
              `if !slices.Contains([]string{${sliceValues}}, ${envVariableName}) {`,
              `resp.Diagnostics.AddError("Invalid ${envVariable} Environment Variable Value", "Must be one of: ${displayValues}")`,
              `} else {`,
              `data.${fieldName} = types.StringValue(${envVariableName})`,
              `}`,
              `}`,
            );
          } else {
            result.push(
              ``,
              `if ${envVariableName}, ok := os.LookupEnv("${envVariable}"); ok && data.${fieldName}.IsNull() {`,
              `data.${fieldName} = types.StringValue(${envVariableName})`,
              `}`,
            );
          }

          break;
      }
    }

    if (fieldDef.Default) {
      switch (effectiveType) {
        case "boolean":
          result.push(
            ``,
            `if data.${fieldName}.IsNull() {`,
            `data.${fieldName} = types.BoolValue(${fieldDef.Default.Value})`,
            `}`,
          );

          break;
        case "float32":
          result.push(
            ``,
            `if data.${fieldName}.IsNull() {`,
            `data.${fieldName} = types.Float32Value(${fieldDef.Default.Value})`,
            `}`,
          );

          break;
        case "int32":
          result.push(
            ``,
            `if data.${fieldName}.IsNull() {`,
            `data.${fieldName} = types.Int32Value(${fieldDef.Default.Value})`,
            `}`,
          );

          break;
        case "integer":
          result.push(
            ``,
            `if data.${fieldName}.IsNull() {`,
            `data.${fieldName} = types.Int64Value(${fieldDef.Default.Value})`,
            `}`,
          );

          break;
        case "number":
          result.push(
            ``,
            `if data.${fieldName}.IsNull() {`,
            `data.${fieldName} = types.Float64Value(${fieldDef.Default.Value})`,
            `}`,
          );

          break;
        case "string":
          result.push(
            ``,
            `if data.${fieldName}.IsNull() {`,
            `data.${fieldName} = types.StringValue(${templateBuiltinString(
              fieldDef.Default.Value,
            )})`,
            `}`,
          );

          break;
      }
    }
  }

  return result.join("\n");
}

registerTemplateFunc(
  "templateProviderConfigureGlobals",
  templateProviderConfigureGlobals,
);

/**
 * Templates the provider Configure() method with the Go SDK options, such as
 * sdk.WithSecurity(security).
 */
function templateProviderConfigureSDKOptions(sdk: SDK) {
  const result: string[] = [`opts := []sdk.SDKOption{`];
  const serverVariables = sdk.Servers?.GetVariables();
  const symbolManager = makeSymbolMananger();

  if (serverVariables && serverVariables.length > 0) {
    result.push(`sdk.WithTemplatedServerURL(serverUrl, serverUrlParams),`);
  } else {
    result.push(`sdk.WithServerURL(serverUrl),`);
  }

  if (sdk.Security?.Type?.Fields.length) {
    result.push(`sdk.WithSecurity(security),`);
  }

  result.push(`sdk.WithClient(httpClient),`, `}`);

  if (sdk.Globals) {
    for (const fieldDef of sdk.Globals.Fields) {
      const fieldName = sanitizeFieldName(fieldDef.Name);
      const frameworkType = FrameworkTypeFromFieldDef(fieldDef);
      const terraformAccessor = `data.${fieldName}`;

      frameworkType
        .templateTerraformToSDKImports(fieldDef.Type, true, false)
        .forEach((importPath) => {
          addGenImport(importPath);
        });

      let value = frameworkType.templateTerraformToSDKValue(
        symbolManager,
        fieldName,
        fieldDef.Type,
        true,
        false,
        terraformAccessor,
      );

      result.push(
        ``,
        `if !${terraformAccessor}.IsUnknown() && !${terraformAccessor}.IsNull() {`,
        `opts = append(opts, sdk.With${fieldName}(${value}))`,
        `}`,
      );
    }
  }

  return result.join("\n");
}

registerTemplateFunc(
  "templateProviderConfigureSDKOptions",
  templateProviderConfigureSDKOptions,
);

/** Templates additional provider attribute fields into the schema data model
 *  type. */
function templateProviderDataModel(context: ProviderContext) {
  const additionalProviderAttributes: AdditionalProviderAttributesConfig =
    context.Config.AdditionalProviderAttributes;
  const globals = context.AST.MainSDK.Globals;
  const security = context.AST.MainSDK.Security?.Type;
  const serverVariables = context.AST.MainSDK.Servers?.GetVariables();

  // There may be shared attribute naming across features, such as security and
  // server variables. Similar to schema generation, last definition wins.
  // This is a mapping of field names to struct field definitions.
  const structFields: Record<string, string> = {};

  if (additionalProviderAttributes) {
    if (additionalProviderAttributes.httpHeaders) {
      const fieldName = sanitizeFieldName(
        additionalProviderAttributes.httpHeaders,
      );
      const structField = new FrameworkMapType(
        new FrameworkStringType(),
      ).templateDataModelStructField(additionalProviderAttributes.httpHeaders);

      structFields[fieldName] = structField;
    }

    if (additionalProviderAttributes.tlsSkipVerify) {
      const fieldName = sanitizeFieldName(
        additionalProviderAttributes.tlsSkipVerify,
      );
      const structField = new FrameworkBoolType().templateDataModelStructField(
        additionalProviderAttributes.tlsSkipVerify,
      );

      structFields[fieldName] = structField;
    }
  }

  if (globals) {
    for (const fieldDef of globals.Fields) {
      const fieldName = sanitizeFieldName(fieldDef.Name);
      const structField = FrameworkTypeFromFieldDef(
        fieldDef,
      ).templateDataModelStructField(fieldDef.Name);

      structFields[fieldName] = structField;
    }
  }

  if (!context.Config.BaseServerURL) {
    const fieldName = sanitizeFieldName("server_url");
    const structField = new FrameworkStringType().templateDataModelStructField(
      "server_url",
    );

    structFields[fieldName] = structField;
  }

  if (serverVariables) {
    serverVariables.forEach((variable) => {
      const sourceName = providerServerVariableSourceName(variable.Name);
      const fieldName = sanitizeFieldName(sourceName);
      const structField = FrameworkTypeFromTypeDef(
        variable.Type,
      ).templateDataModelStructField(sourceName);

      structFields[fieldName] = structField;
    });
  }

  if (security) {
    flattenSecurityObject(security).forEach((fieldDef) => {
      const fieldName = sanitizeFieldName(fieldDef.Name);
      const structField = FrameworkTypeFromFieldDef(
        fieldDef,
      ).templateDataModelStructField(fieldDef.Name);

      structFields[fieldName] = structField;
    });
  }

  return Object.keys(structFields)
    .sort()
    .map((fieldName) => {
      return structFields[fieldName];
    })
    .join("\n");
}

registerTemplateFunc("templateProviderDataModel", templateProviderDataModel);

/** Templates SDK client HTTP transport headers if additionalProviderAttributes
 *  configuration httpHeaders is set.
 *
 *  TODO: This logic should all be rolled into a single function for templating
 *        the SDK client instantiation to simplify the provider.go template. */
function templateProviderHTTPTransportOptsSetHeaders(
  context: ProviderContext,
  dataModelVariable: string,
  httpTransportVariable: string,
): string {
  const additionalProviderAttributes: AdditionalProviderAttributesConfig =
    context.Config.AdditionalProviderAttributes;

  if (!additionalProviderAttributes?.httpHeaders) {
    return "";
  }

  const httpHeadersFieldName = sanitizeFieldName(
    additionalProviderAttributes.httpHeaders,
  );

  const result: string[] = [];

  result.push(
    `resp.Diagnostics.Append(${dataModelVariable}.${httpHeadersFieldName}.ElementsAs(ctx, &${httpTransportVariable}.SetHeaders, false)...)`,
  );
  result.push(`if resp.Diagnostics.HasError() {`);
  result.push(`return`);
  result.push(`}`);

  return result.join("\n");
}

registerTemplateFunc(
  "templateProviderHTTPTransportOptsSetHeaders",
  templateProviderHTTPTransportOptsSetHeaders,
);

/** Templates SDK client TLS skip verification if additionalProviderAttributes
 *  configuration tlsSkipVerify is set.
 *
 *  TODO: This logic should all be rolled into a single function for templating
 *        the SDK client instantiation to simplify the provider.go template. */
function templateProviderHTTPTransportTlsSkipVerify(
  context: ProviderContext,
  dataModelVariable: string,
  httpTransportVariable: string,
): string {
  const additionalProviderAttributes: AdditionalProviderAttributesConfig =
    context.Config.AdditionalProviderAttributes;

  if (!additionalProviderAttributes?.tlsSkipVerify) {
    return "";
  }

  addGenImport("crypto/tls");
  addGenImport("net/http");

  const tlsSkipVerifyFieldName = sanitizeFieldName(
    additionalProviderAttributes.tlsSkipVerify,
  );

  const result: string[] = [];

  result.push(
    `if transport, ok := ${httpTransportVariable}.Transport.(*http.Transport); ok {`,
  );
  result.push(`if transport.TLSClientConfig == nil {`);
  result.push(`transport.TLSClientConfig = &tls.Config{}`);
  result.push(`}`);
  result.push(
    `transport.TLSClientConfig.InsecureSkipVerify = ${dataModelVariable}.${tlsSkipVerifyFieldName}.ValueBool()`,
  );
  result.push(`}`);

  return result.join("\n");
}

registerTemplateFunc(
  "templateProviderHTTPTransportTlsSkipVerify",
  templateProviderHTTPTransportTlsSkipVerify,
);

/** Templates the provider schema. */
function templateProviderSchema(providerContext: ProviderContext): string {
  const schema = new ProviderSchema({
    attributes: getProviderAttributes(providerContext),
    markdownDescription: providerContext.AST.MainSDK.Comments?.Description,
  });

  schema.imports().forEach((imp) => addGenImport(imp.Path, false, imp.Alias));

  return schema.template();
}

registerTemplateFunc("templateProviderSchema", templateProviderSchema);

/**
 * Returns an HCL-formatted example value for a provider attribute based on its type.
 * For sensitive fields, returns a placeholder. Otherwise delegates to the
 * attribute's templateExampleZeroValue() method.
 */
function getProviderAttributeExampleValue(
  attributeName: string,
  attribute: ProviderAttribute,
): string {
  const isSensitive = attribute.sensitive ?? false;

  // For sensitive fields, use placeholder
  if (isSensitive) {
    return `"<YOUR_${attributeName.toUpperCase()}>"`;
  }

  return attribute.templateExampleZeroValue();
}

/**
 * Returns the set of server URL-related attribute names.
 * These are always shown in the provider example even if optional.
 */
function getServerURLAttributeNames(
  providerContext: ProviderContext,
): Set<string> {
  const names = new Set<string>();
  const serverVariables = providerContext.AST.MainSDK.Servers?.GetVariables();

  if (!providerContext.Config.BaseServerURL) {
    names.add("server_url");
  }

  if (serverVariables) {
    serverVariables.forEach((variable) => {
      names.add(providerServerVariableAttributeName(variable.Name));
    });
  }

  return names;
}

/**
 * Gets the example value for a provider attribute, using OAS examples when available.
 * Uses the attribute's stored typeDef for OAS example generation.
 */
function getProviderAttributeExampleValueWithTypeDef(
  attributeName: string,
  attribute: ProviderAttribute,
): string {
  const isSensitive = attribute.sensitive ?? false;
  const typeDef = attribute.typeDef;

  // If we have a TypeDef, use templateTerraformConfigValue for OAS examples
  if (typeDef) {
    const frameworkType = FrameworkTypeFromTypeDef(typeDef);
    const sensitivePlaceholder = `YOUR_${attributeName.toUpperCase()}`;
    return frameworkType.templateTerraformConfigValue(
      typeDef,
      isSensitive,
      sensitivePlaceholder,
    );
  }

  // Fallback to the simpler approach without TypeDef
  return getProviderAttributeExampleValue(attributeName, attribute);
}

/**
 * Templates the provider example configuration for examples/provider/provider.tf.
 * Shows required provider attributes and server URL variables with example values.
 * Optional attributes are omitted to reduce noise.
 * Uses OAS-defined examples when available via attribute.typeDef.
 */
function templateProviderExample(providerContext: ProviderContext): string {
  const attributes = getProviderAttributes(providerContext);
  const serverURLAttributes = getServerURLAttributeNames(providerContext);
  const lines: string[] = [];

  // Sort attributes alphabetically for consistent output
  const sortedNames = Object.keys(attributes).sort();

  for (const attributeName of sortedNames) {
    const attribute = attributes[attributeName];
    const isServerURLAttribute = serverURLAttributes.has(attributeName);

    // Only show required attributes and server URL variables
    if (!attribute.required && !isServerURLAttribute) {
      continue;
    }

    // Get the example value (using OAS examples when TypeDef is available)
    const exampleValue = getProviderAttributeExampleValueWithTypeDef(
      attributeName,
      attribute,
    );

    // Build the comment
    const commentParts: string[] = [];

    if (attribute.required) {
      commentParts.push("Required");
    } else {
      commentParts.push("Optional");
    }

    // Check for environment variable
    const envVar = getEnvironmentVariable(attributeName);
    if (envVar) {
      commentParts.push(`can use ${envVar} environment variable`);
    }

    const comment = ` # ${commentParts.join(" - ")}`;

    lines.push(`  ${attributeName} = ${exampleValue}${comment}`);
  }

  return lines.join("\n");
}

registerTemplateFunc("templateProviderExample", templateProviderExample);
