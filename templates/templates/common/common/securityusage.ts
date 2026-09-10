type UsageSecurityScheme = {
  Type: string;
  Description: string;
  Fields: FieldDef[];
};

type UsageSecurityOption = {
  Name: string;
  Schemes: UsageSecurityScheme[];
};

function sdkHasGlobalSecurity(): boolean {
  return (
    context.Global.AST.MainSDK.Security &&
    context.Global.AST.MainSDK.Security.Type.Fields.length > 0
  );
}

registerTemplateFunc("sdkHasGlobalSecurity", sdkHasGlobalSecurity);

function opUsesGlobalSecurity(operation: Operation) {
  return (
    sdkHasGlobalSecurity() && !operation.Security && operation.GlobalSecurity
  );
}

//@ts-ignore
function opRequiresGlobalSecurity(operation: Operation): bool {
  if (!opUsesGlobalSecurity(operation)) {
    return false;
  }

  if (!context.Global.AST.MainSDK.Security.Optional) {
    return true;
  }

  return (
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason === "env-var"
  );
}

// Returns a remark string for operations whose security was hoisted as a strict subset.
// templateHoistedSecurityRemark is responsible for language-specific formatting and wording.
// @ts-ignore
function getHoistedSecurityRemark(op: Operation): string {
  if (!op.HoistedSecurityConfig || op.HoistedSecurityConfig.Equivalent) {
    return "";
  }

  if (
    !op.HoistedSecurityConfig.Fields ||
    op.HoistedSecurityConfig.Fields.length === 0
  ) {
    throw new Error(
      "Unexpected empty Hoist.Fields for operation with non-equivalent hoist",
    );
  }

  return templateHoistedSecurityRemark(
    op.HoistedSecurityConfig.Fields,
    opRequiresGlobalSecurity(op),
  );
}
registerTemplateFunc("getHoistedSecurityRemark", getHoistedSecurityRemark);

// @ts-ignore
function templateHoistedSecurityRemark(
  _fields: HoistedSecurityField[],
  _required: boolean,
): string {
  throw new Error(`templateHoistedSecurityRemark not implemented by target.`);
}

// @ts-ignore - getGlobalSecurity is defined per-language; not all targets provide it
function getGlobalUsageSecurity(local: UsageContext): string | undefined {
  // @ts-ignore
  return getGlobalSecurity(
    { index: local.Operation?.HoistedSecurityConfig?.Fields?.[0]?.Index },
    {
      isTest: local.Test && true,
      // @ts-ignore
      usageContext: local,
    },
  );
}

function collectSecurityScheme(
  securitySchemeField: FieldDef,
  annotation: SecurityAnnotation,
): UsageSecurityScheme {
  const scheme: UsageSecurityScheme = {
    Type: annotation.SecType.toString(),
    Description: "",
    Fields: [securitySchemeField],
  };

  switch (scheme.Type) {
    case "apiKey":
      scheme.Description = "API key";
      break;
    case "http":
      switch (annotation.SubType.toString()) {
        case "basic":
          scheme.Description = "HTTP Basic";
          break;
        case "bearer":
          scheme.Description = "HTTP Bearer";
          break;
        case "custom":
          scheme.Description = "Custom HTTP"; // TODO try and pull this from the extension
          break;
      }
      break;
    case "oauth2":
      scheme.Description = "OAuth2 token";
      break;
    case "openIdConnect":
      scheme.Description = "OpenID Connect Discovery";
      break;
  }

  return scheme;
}

function isSpecialScheme(security: TypeDef): any {
  const specialSchemes = [
    {
      scheme: "basic",
      fieldNames: ["username", "password"],
    },
    {
      scheme: "client_credentials",
      fieldNames: ["clientID", "clientSecret", "tokenURL", "scopes"],
    },
  ];

  const isCustomScheme = security.Fields.every((f) => {
    for (const anno of f.Annotations) {
      if (anno.IsType("security")) {
        const secAnno = anno as SecurityAnnotation;
        if (
          secAnno.Scheme &&
          secAnno.SecType == "http" &&
          secAnno.SubType == "custom"
        ) {
          return true;
        }
      }
    }
  });

  if (isCustomScheme) {
    return {
      scheme: "custom",
      fieldNames: security.Fields.map((f) => f.OriginalName),
    };
  }

  for (const specialScheme of specialSchemes) {
    if (
      security.Fields.every((f) =>
        specialScheme.fieldNames.includes(f.OriginalName),
      )
    ) {
      return specialScheme;
    }
  }

  return false;
}

// @ts-ignore
function collectSecurityOptions(
  security: TypeDef,
  name: string = "",
): UsageSecurityOption[] {
  const options: UsageSecurityOption[] = [];

  const specialScheme = isSpecialScheme(security);
  if (specialScheme) {
    switch (specialScheme.scheme) {
      case "basic":
        return [
          {
            Name: name,
            Schemes: [
              {
                Type: "http",
                Description: "HTTP Basic",
                Fields: security.Fields,
              },
            ],
          },
        ];
      case "custom":
        return [
          {
            Name: name,
            Schemes: [
              {
                Type: "http",
                Description: "Custom HTTP", // TODO try and pull this from the extension
                Fields: security.Fields,
              },
            ],
          },
        ];
      case "client_credentials":
        return [
          {
            Name: name,
            Schemes: [
              {
                Type: "oauth2",
                Description: "OAuth2 Client Credentials Flow",
                Fields: security.Fields,
              },
            ],
          },
        ];
    }
  }

  const schemes: UsageSecurityScheme[] = [];
  security.Fields.forEach((field) => {
    const annotations = securityAnnotations(field);
    // TODO: I don't think it is actually possible to have multiple security options on a single field
    while (annotations.length > 0) {
      const annotation = annotations.shift();
      if (annotation.Option && field.Type != null) {
        options.push(...collectSecurityOptions(field.Type, field.Name));
        continue;
      }
      schemes.push(collectSecurityScheme(field, annotation));
    }
  });

  if (schemes.length > 0) {
    options.push({ Name: name, Schemes: schemes });
  }
  return options;
}

// @ts-ignore
function collectGlobalSecurityOptions(): UsageSecurityOption[] {
  const options: UsageSecurityOption[] = [];

  if (sdkHasGlobalSecurity()) {
    options.push(
      ...collectSecurityOptions(context.Global.AST.MainSDK.Security.Type),
    );
  }

  return options;
}

registerTemplateFunc(
  "collectGlobalSecurityOptions",
  collectGlobalSecurityOptions,
);

// @ts-ignore
function templateSecurityTable(option: UsageSecurityOption): string {
  if (option.Schemes == undefined) {
    return "";
  }

  const envVarPrefix = context.Global.Config.EnvVarPrefix;

  const headers = ["Name", "Type", "Scheme"];

  if (envVarPrefix) {
    headers.push("Environment Variable");
  }

  const contents: string[][] = [headers];

  option.Schemes.forEach((scheme) => {
    const row = [
      scheme.Fields.map((f) => `\`${sanitizeSecurityFieldName(f.Name)}\``).join(
        "<br/>",
      ),
      scheme.Type,
      scheme.Description,
    ];

    if (envVarPrefix) {
      let envVars = [];

      for (const field of scheme.Fields) {
        envVars.push(`\`${templateSecurityEnvVars(field)}\``);
      }

      if (envVars.length > 0) {
        row.push(envVars.join("<br/>"));
      }
    }

    contents.push(row);
  });

  return createMarkdownTable(contents, false);
}
registerTemplateFunc("templateSecurityTable", templateSecurityTable);

function templateSecurityParametersToSet(scheme: UsageSecurityScheme): string {
  const allOptional = scheme.Fields.every((f) => f.Optional || f.Default);
  const fieldsToSet: FieldDef[] = allOptional
    ? scheme.Fields
    : scheme.Fields.filter((f) => !f.Optional && !f.Default);

  const params = fieldsToSet.map(
    (f) => `\`${sanitizeSecurityFieldName(f.Name)}\``,
  );

  return `${params.join(", ")} ${
    params.length > 1 ? "parameters" : "parameter"
  }`;
}
registerTemplateFunc(
  "templateSecurityParametersToSet",
  templateSecurityParametersToSet,
);

// @ts-ignore
function templateFlattenedSecurityUsage(
  sec: TypeDef,
  indent: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let schemeAnno = null;
  const secField = sec.Fields[0];

  const fieldExample =
    example !== undefined ? example[secField.Name] : undefined;

  for (const anno of secField.Annotations as unknown as Array<SecurityAnnotation>) {
    if (anno.IsType("security")) {
      if (anno.Scheme) {
        schemeAnno = anno;
      }
    }
  }
  const security = templateSecurityFieldValue(
    secField,
    schemeAnno,
    false,
    fieldExample,
    additionalContext,
  );

  return indentLines([`${security}`], indent);
}

// @ts-ignore
function templateUnflattenedSecurityUsage(
  sec: TypeDef,
  indent: number,
  valueOnly: boolean,
  secFieldIdx: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  const security = templateSecurity(
    sec,
    1,
    secFieldIdx,
    example,
    additionalContext,
  );

  let prefix = "";
  if (!valueOnly) {
    const securityFieldDef: FieldDef = {
      Clone: (): FieldDef => {
        return { ...securityFieldDef };
      },
      GetID: () => "Security",
      Name: "Security",
      OriginalName: "security",
      Optional: false,
      Nullable: false,
      ErrorMessage: false,
      IsAdditionalProperties: false,
      IsResponseHeaders: false,
      IsResponseMetadata: false,
    };
    prefix = `${templateFieldDeclaration(securityFieldDef)}`;
  }

  return indentLines(
    [
      `${prefix}${templateType(sec, { security: true })}${templateBracket(
        sec,
        true,
        { security: true },
      )}`,
      security,
      templateBracket(sec, false, { security: true }),
    ],
    indent,
  );
}

// @ts-ignore
function templateSecurityUsage(
  sec: TypeDef,
  indent: number,
  valueOnly: boolean,
  flattenedSecurity: boolean,
  secFieldIdx?: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (sec.Fields.length == 1 && flattenedSecurity) {
    return templateFlattenedSecurityUsage(
      sec,
      indent,
      example,
      additionalContext,
    );
  }

  let index = secFieldIndex(example, sec);
  if (index == undefined) {
    index = secFieldIdx;
  }
  index = index || 0;

  return templateUnflattenedSecurityUsage(
    sec,
    indent,
    valueOnly,
    index,
    example,
    additionalContext,
  );
}

registerTemplateFunc("templateSecurityUsage", templateSecurityUsage);

function secFieldIndex(example: any, sec: TypeDef): number {
  if (example !== undefined) {
    // If we have more than one option we need to figure out the index of the option to use
    // by checking which security object has all of the example fields
    if (getNumberOfSecurityOptions(sec.Fields) > 1) {
      const requiredSchemes = Object.keys(example);

      for (let i = 0; i < sec.Fields.length; i++) {
        const option = sec.Fields[i];

        if (
          requiredSchemes.length == 1 &&
          option.OriginalName == requiredSchemes[0]
        ) {
          return i;
        }

        const foundSchemes = option.Type.Fields.map(
          (f) => f.Annotations?.Get("security")?.SchemeKey,
        );

        if (
          foundSchemes.length > 0 &&
          foundSchemes.every((f) => requiredSchemes.includes(f))
        ) {
          return i;
        }
      }
    }
  }
  return undefined;
}

// @ts-ignore
function templateSecurity(
  sec: TypeDef,
  indent: number,
  secFieldIdx: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let optionAnno = null;
  let schemeAnno = null;
  const field = sec.Fields[secFieldIdx];
  for (const annotation of field.Annotations as unknown as Array<SecurityAnnotation>) {
    if (annotation.IsType("security")) {
      if (annotation.Option) {
        optionAnno = annotation;
      }
      if (annotation.Scheme) {
        schemeAnno = annotation;
      }
    }
  }

  if (optionAnno) {
    let templatedOption = templateOption(
      field.Type,
      1,
      example,
      additionalContext,
    );

    let optionalPrefix = templateOptionalSymbol();

    return (
      indentLines(
        [
          `${templateFieldDeclaration(field)}${optionalPrefix}${templateType(
            field.Type,
            { security: true },
          )}${templateBracket(field.Type, true, { security: true })}`,
          templatedOption,
          templateBracket(field.Type, false, { security: true }),
        ],
        indent,
      ) + templateFieldDelimiter()
    );
  } else if (schemeAnno) {
    const hasSecurityOptions = getNumberOfSecurityOptions(sec.Fields) > 0;

    // Auth with multiple fields is flattened up one level so we need to unwrap the example
    if (
      !hasSecurityOptions &&
      ((schemeAnno.SecType == "http" &&
        ["basic", "custom"].includes(schemeAnno.SubType)) ||
        (schemeAnno.SecType == "oauth2" &&
          schemeAnno.SubType == "client_credentials"))
    ) {
      example = example !== undefined ? Object.values(example)[0] : undefined;
    }

    // Preserve legacy test-generation behavior: tests may override optional
    // fields (for example tokenURL), so avoid narrowing the field set.
    if (additionalContext?.isTest) {
      return templateOption(sec, indent, example, additionalContext);
    }

    // For multi-field schemes (basic, client_credentials, custom) that have
    // been flattened, pass all fields so the usage example includes every
    // required credential. For OR-option selection, pass only the selected
    // field. For AND-combined schemes (no options), pass all fields.
    const specialScheme = isSpecialScheme(sec);
    const selectedFields = specialScheme
      ? sec.Fields.filter(
          (f) => !f.Default || securityAnnotations(f).length > 0,
        )
      : hasSecurityOptions && secFieldIdx < sec.Fields.length
      ? [sec.Fields[secFieldIdx]]
      : sec.Fields;
    return templateOption(
      { Fields: selectedFields } as TypeDef,
      indent,
      example,
      additionalContext,
    );
  }

  return "";
}

// @ts-ignore
function templateOption(
  typeDef: TypeDef,
  indent: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let fields = [];
  let optionalFieldTemplated = false;
  const hasSecurityOptions = getNumberOfSecurityOptions(typeDef.Fields) > 0;

  for (const field of typeDef.Fields) {
    const fieldExample =
      example !== undefined
        ? example[
            Object.keys(example).find(
              (key) => key.toLowerCase() === field.OriginalName.toLowerCase(),
            )
          ]
        : undefined;

    // When generating tests, override tokenURL to point at the mock server's
    // token endpoint so OAuth2 client-credentials flows hit the mock server
    // instead of the real identity provider. This must run before optional-field
    // pruning, since tokenURL is optional and would otherwise be skipped.
    if (
      additionalContext?.isTest &&
      fieldExample === undefined &&
      field.OriginalName === "tokenURL"
    ) {
      const tokenUrlValue = templateStringValue(
        "/_mockserver/oauth2/token",
        field,
        additionalContext,
      );
      fields.push(
        indentLines(
          [`${templateFieldDeclaration(field)}${tokenUrlValue}`],
          indent,
        ) + templateFieldDelimiter(),
      );
      continue;
    }

    if (
      field.Optional &&
      (optionalFieldTemplated ||
        (example !== undefined && fieldExample === undefined))
    ) {
      continue;
    }

    let schemeAnno = null;

    for (const anno of field.Annotations as unknown as Array<SecurityAnnotation>) {
      if (anno.IsType("security")) {
        if (anno.Scheme) {
          schemeAnno = anno;
        }
      }
    }

    if (schemeAnno) {
      let optionalPrefix = "";

      if (field.Optional) {
        if (hasSecurityOptions) {
          optionalFieldTemplated = true;
        }
        optionalPrefix = templateOptionalSymbol();
      }

      fields.push(
        templateScheme(
          field,
          schemeAnno,
          optionalPrefix,
          indent,
          fieldExample,
          { ...additionalContext, schemeParent: typeDef },
        ),
      );
    } else {
      const fieldUsage = templateFieldUsage(
        field,
        indent,
        typeDef,
        fieldExample,
        additionalContext,
      );

      if (fieldUsage == "") {
        continue;
      }

      fields.push(fieldUsage + templateFieldDelimiter());
    }
  }

  return fields.join("\n");
}

// @ts-ignore
function templateScheme(
  fieldDef: FieldDef,
  schemeAnno: SecurityAnnotation,
  optionalPrefix: string,
  indent: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (fieldDef.Type.IsPrimitive()) {
    return indentLines(
      [
        templateSchemeField(
          fieldDef,
          schemeAnno,
          optionalPrefix != "",
          example,
          additionalContext,
        ),
      ],
      indent,
    );
  } else {
    let fields = [];

    for (const field of fieldDef.Type.Fields) {
      const fieldExample =
        example !== undefined ? example[field.OriginalName] : undefined;

      fields.push(
        templateSchemeField(
          field,
          schemeAnno,
          false,
          fieldExample,
          additionalContext,
        ),
      );
    }

    return (
      indentLines(
        [
          `${templateFieldDeclaration(fieldDef)}${optionalPrefix}${templateType(
            fieldDef.Type,
            { security: true },
          )}${templateBracket(fieldDef.Type, true, { security: true })}`,
          indentLines(fields, indent),
          templateBracket(fieldDef.Type, false, { security: true }),
        ],
        indent,
      ) + templateFieldDelimiter()
    );
  }
}

// @ts-ignore
function templateSchemeField(
  fieldDef: FieldDef,
  schemeAnno: SecurityAnnotation,
  optional: boolean,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return `${templateFieldDeclaration(fieldDef)}${templateSecurityFieldValue(
    fieldDef,
    schemeAnno,
    optional,
    example,
    additionalContext,
  )}${templateFieldDelimiter()}`;
}

// @ts-ignore
function parseSecurityExampleDirectives(
  example: any,
  additionalContext?: TemplateValueContext,
  fieldDef?: FieldDef,
): any {
  if (example === undefined || typeof example !== "string") {
    return undefined;
  }

  const envVar = getEnvVarFromDirective(example);

  if (!isFeatureUsed("envVarSecurityUsage") || !envVar) {
    return templateStringValue(example, {
      Optional: false,
      Nullable: false,
    } as FieldDef);
  }

  // This helps us solve problems with initial tests failing because keys are empty by default
  if (additionalContext?.isTest && envVar.defaultValue === "") {
    envVar.defaultValue = "value";
  }

  return languageSpecificEnvVarWrapping(envVar, additionalContext, fieldDef);
}

//@ts-ignore
function templateSecurityFieldValue(
  fieldDef: FieldDef,
  schemeAnno: SecurityAnnotation,
  optional: boolean,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  example = getSecurityExample(
    example,
    fieldDef,
    schemeAnno.SecType,
    schemeAnno.SubType,
  );
  const value = () =>
    `${templateSecurityValue(example, fieldDef, optional, additionalContext)}`;

  switch (schemeAnno.SecType.toString()) {
    case "apiKey":
      return value();
    case "http":
      switch (schemeAnno.SubType.toString()) {
        case "custom":
        case "basic":
        case "bearer":
          return value();
      }
      break;
    case "oauth2":
      return value();
    case "openIdConnect":
      return value();
    default:
      throw new Error(`unknown type ${schemeAnno.SecType.toString()}`);
  }
}

// @ts-ignore
function supportsEnvVarExamples(): boolean {
  return (
    isFeatureUsed("envVarSecurityUsage") && context.Global.Config.EnvVarPrefix
  );
}

// @ts-ignore
function languageSpecificEnvVarWrapping(
  envVar: EnvVar,
  additionalContext?: TemplateValueContext,
  fieldDef?: FieldDef,
): string {
  throw new Error(`unsupported language ${context.Global.Config.Language}`);
}

// @ts-ignore
function templateSecurityValue(
  value: any,
  field: FieldDef,
  optional: boolean,
  additionalContext?: TemplateValueContext,
): string {
  if (isExampleReferenceValue(value)) {
    return templateExampleReferenceValue(
      value as ExampleReferenceValue,
      field,
      additionalContext,
    );
  }

  return parseSecurityExampleDirectives(value, additionalContext, field);
}

// @ts-ignore
function getSecurityExample(
  example: any,
  fieldDef: FieldDef,
  type: string,
  subType: string,
): any {
  const fieldName = fieldDef.Name;

  if (example === undefined && fieldDef.Type.Examples?.length > 0) {
    example = getExampleValue(
      faker.helpers.arrayElement(fieldDef.Type.Examples),
    );
  }

  if (example !== undefined) {
    return example;
  }

  if (supportsEnvVarExamples()) {
    return templateSecurityEnvVars(fieldDef, true);
  }

  switch (type) {
    case "apiKey":
      return "<YOUR_API_KEY_HERE>";
    case "http":
      switch (subType) {
        case "basic":
          return `<YOUR_${caser().ToSNAKE(fieldName)}_HERE>`;
        case "bearer":
          return "<YOUR_BEARER_TOKEN_HERE>";
        case "custom":
          return `<YOUR_${caser().ToSNAKE(fieldName)}_HERE>`;
        default:
          throw new Error(`unknown subType ${subType}`);
      }
    case "oauth2":
      return `<YOUR_${caser().ToSNAKE(fieldName)}_HERE>`;
    case "openIdConnect":
      return `<YOUR_${caser().ToSNAKE(fieldName)}_HERE>`;
    default:
      throw new Error(`unknown type ${type}`);
  }
}
