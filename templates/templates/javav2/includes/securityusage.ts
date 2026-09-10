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
    let optionalPrefix = templateOptionalSymbol();
    return (
      indentLines(
        [
          `${templateFieldDeclaration(field)}${optionalPrefix}${templateType(
            field.Type,
            example,
            { indent: 1 },
          )}${templateBracket(field.Type, true)}`,
          templateIndent(1) + templateBracket(field.Type, false),
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
      return templateOption(sec, 0, example, additionalContext);
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
      0,
      example,
      additionalContext,
    );
  }

  return "";
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
  }

  const lines: string[] = [];
  lines.push(
    `${templateFieldDeclaration(fieldDef)}${optionalPrefix}${templateType(
      fieldDef.Type,
      example,
      { indent: 1 },
    )}`,
  );

  for (const field of fieldDef.Type.Fields) {
    if (field.Optional) {
      const fieldExample = findFieldExample(example, field);

      lines.push(
        `${templateIndent(indent + 1)}${templateSchemeField(
          field,
          schemeAnno,
          false,
          fieldExample,
          additionalContext,
        )}`,
      );
    }
  }

  lines.push(
    templateIndent(indent + 1) + templateBracket(fieldDef.Type, false),
  );

  return indentLines(lines, indent) + templateFieldDelimiter();
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
    templateSecurityValue(example, fieldDef, optional, additionalContext);

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
      switch (schemeAnno.SubType.toString()) {
        case "password":
          // is a oneOf String | Credentials
          // so need to surround with sec field type like `Oauth2Input.of("<YOUR_OAUTH2_HERE>")`
          return `${importedClass(fieldDef.Type)}.of(${value()})`;
        default:
          return value();
      }
    case "openIdConnect":
      return value();
    default:
      throw new Error(`unknown type ${schemeAnno.SecType.toString()}`);
  }
}

// @ts-ignore
function templateFlattenedSecurityUsage(
  sec: TypeDef,
  indent: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let schemeAnno = null;
  const secField = sec.Fields[0];

  for (const anno of secField.Annotations as unknown as Array<SecurityAnnotation>) {
    if (anno.IsType("security")) {
      if (anno.Scheme) {
        schemeAnno = anno;
      }
    }
  }
  const fieldExample = findFieldExample(example, secField);

  const security = templateSecurityFieldValue(
    secField,
    schemeAnno,
    false,
    fieldExample,
    additionalContext,
  );
  const builderName = templateSDKBuilderName(secField.Name);

  return indentLines([`.${builderName}(${security}\)`], indent).trimStart();
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
  let prefix = "";
  if (!valueOnly) {
    const fakeFieldDef: FieldDef = {
      Clone: (): FieldDef => {
        return { ...fakeFieldDef };
      },
      GetID: () => "Security",
      Name: "Security",
      // Nothing else used in templateFieldDeclaration
      OriginalName: "security",
      Nullable: false,
      Optional: false,
      ErrorMessage: false,
      IsAdditionalProperties: false,
      IsResponseHeaders: false,
      IsResponseMetadata: false,
    };
    prefix = `${templateFieldDeclaration(fakeFieldDef)}`;
  }

  return indentLines(
    [
      `.security\(${prefix}${builderOf(sec)}`,
      indentLines(
        [
          `${templateSecurity(
            sec,
            0,
            secFieldIdx,
            example,
            additionalContext,
          )}`,
          `.build\(\)\)`,
        ],
        1,
      ),
    ],
    indent,
  );
}

// @ts-ignore
function supportsEnvVarExamples(): boolean {
  // trick the example generation to use env vars
  return true;
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
  // Tests don't yet support env vars, so we stick to asserting literal example values
  if (envVar === undefined) {
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
