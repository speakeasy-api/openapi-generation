unregisterTemplateFunc("postProcessBasicValue");
unregisterTemplateFunc("templateValue");

// @ts-ignore
function templateValue(
  fieldDef: FieldDef,
  example?: any,
  parentHasExample?: boolean,
  additionalContext?: TemplateValueContext,
): string {
  const value = templateValueCommon(
    fieldDef,
    example,
    parentHasExample,
    additionalContext,
  );
  if (fieldDef.Optional || fieldDef.Nullable) {
    if (context.Global.Config.NullFriendlyParameters) {
      if (value.endsWith(".get()")) {
        // An always-optional getter yields an Optional that may be empty,
        // so unwrap permissively (absent -> null)
        return `${removeGetter(value)}.orElse(null)`;
      }
      return value;
    }
    // Non-nullFriendly builders have an Optional/JsonNullable overload, so
    // the trailing .get() is unnecessary.
    return removeGetter(value);
  }
  return value;
}
registerTemplateFunc("templateValue", templateValue);

// @ts-ignore
function templateObject(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  // builder and required fields set by templateType
  let declaration = templateType(fieldDef.Type, example, {
    ...(additionalContext ?? {}),
    indent: 1,
  });

  const lines = [declaration];
  const optionalFields = nonConstFields(fieldDef.Type.Fields).filter(
    (field) => field.Optional,
  );
  const inlineFields = optionalFields
    .map((field) => {
      let fieldExample = example[originalFieldName(field)];

      const fieldUsage = templateFieldUsage(
        field,
        1,
        fieldDef.Type,
        fieldExample,
        additionalContext,
      );
      return fieldUsage ? fieldUsage + templateFieldDelimiter() : undefined;
    })
    .filter((field) => field !== undefined);

  let additionalPropertiesField = fieldDef.Type.Fields.find(
    (field) => field.IsAdditionalProperties,
  );
  if (additionalPropertiesField) {
    // we take the examples object and remove all the fields that
    // are explicit fields in the fieldDef.Fields. What's left over
    // will be additionalProperties
    let remainingExample = { ...example };
    fieldDef.Type.Fields.filter(
      (field) => !field.IsAdditionalProperties,
    ).forEach((field) => delete remainingExample?.[originalFieldName(field)]);

    // build fieldUsage using remaining properties
    const fieldUsage = templateFieldUsage(
      additionalPropertiesField,
      1,
      fieldDef.Type,
      typeof remainingExample === "object" &&
        Object.keys(remainingExample).length > 0
        ? remainingExample
        : undefined,
      additionalContext,
    );
    if (fieldUsage) {
      inlineFields.push(fieldUsage + templateFieldDelimiter());
    }
  }

  if (inlineFields.length) {
    lines[0] += templateBracket(fieldDef.Type, true);
    lines.push(...inlineFields);
    lines.push(templateIndent(1) + templateBracket(fieldDef.Type, false));
  }

  if (lines.length == 1) {
    return (
      declaration +
      "\n" +
      templateIndent(1) +
      templateBracket(fieldDef.Type, false)
    );
  }

  return lines.join("\n");
}

// @ts-ignore
function processArrayValues(values: any[]): string[] {
  return joinWithComma(values);
}

// @ts-ignore
function processMapValues(values: any[]): string[] {
  return joinWithComma(values);
}

// @ts-ignore
function templateUnion(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  let { type: typeToUse, example: selectedExample } = selectExampleUnionType(
    fieldDef,
    example,
    undefined,
    additionalContext,
  );

  const v = templateValue(
    typeDefToFieldDef(typeToUse, fieldDef),
    selectedExample,
    false,
    additionalContext,
  );
  if (fieldDef.Type.Discriminator) {
    return v;
  } else {
    // We have to make sure that factory method names are distinctive
    // when the types have the same erasure (like List<Integer>, List<String>).
    // When there is no erasure issue the factory method name `of` will be used
    const methodNames = associatedTypeFactoryMethodNames(fieldDef.Type);
    const typ = javaImportTypeMandatory(fieldDef.Type);
    const t = sanitizeTypeMandatory(typeToUse);
    const methodName = methodNames.get(t);
    return `${typ}.${methodName}(${v})`;
  }
}

// @ts-ignore
function getRequestVariableName(stepID: string): string {
  return sanitizeFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}req`,
  );
}
registerTemplateFunc("getRequestVariableName", getRequestVariableName);

function templateAssignRequest(
  usageContext: UsageContext,
  request: RequestDef,
): string {
  if (omitRequest(request)) {
    return "";
  } else {
    const typeDef = request.Field.Type;

    let fieldExample = getOperationMethodFieldExample(
      usageContext,
      request.Field,
    );
    const additionalContext: TemplateValueContext = {
      usageContext: usageContext,
      operation: usageContext.Operation,
      test: usageContext.Test,
      isTest: usageContext.Test != undefined,
      isRequest: true,
    };
    const renderedValue = templateModelUsage(
      request.Field,
      3,
      fieldExample,
      additionalContext,
    ).trimStart();

    if (renderedValue === "") {
      return "";
    }

    const requestVarName = getRequestVariableName(usageContext.StepID);

    if (context.Global.Config.NullFriendlyParameters) {
      return `${javaImportType(typeDef)} ${requestVarName} = ${renderedValue};`;
    }

    let typ: string;

    if (request.Field.Nullable && request.Field.Optional) {
      typ = `${javaImportJsonNullable()}<${javaImportTypeMandatory(typeDef)}>`;
    } else if (
      (request.Field.Nullable || request.Field.Optional) &&
      renderedValue.endsWith("Optional.empty()")
    ) {
      typ = `${javaImportOptional()}<${javaImportTypeMandatory(typeDef)}>`;
    } else {
      typ = `${javaImportTypeMandatory(typeDef)}`;
    }

    return `${typ} ${requestVarName} = ${renderedValue};`;
  }
}

registerTemplateFunc("templateAssignRequest", templateAssignRequest);

function omitRequest(request: RequestDef): boolean {
  let omit = false;
  let example = undefined;

  // only consider omitting the request if RequestBody present
  if (!request.RequestBody) {
    return false;
  }
  const rb = request.RequestBody;
  const typeDef = rb.Type;
  if (typeDef.Examples?.length > 0) {
    example = getExampleValue(faker.helpers.arrayElement(typeDef.Examples));
  }
  const unset = typeDef.Extensions?.ExampleUnset;

  if (
    (unset && !request.IsRequestBodyRequired) ||
    (example !== undefined && example === null && !(rb.Nullable && rb.Optional))
  ) {
    omit = true;
  }
  return omit;
}

// @ts-ignore
function closingBracketOnSeparateLine(): boolean {
  return false;
}

// @ts-ignore
function postProcessBasicValue(field: FieldDef, value: string): string {
  // we want to ensure that a string example like "hello\nthere" does not
  // turn up in a generated snippet as a string over multiple lines (compile error)

  if (
    value == undefined ||
    value == null ||
    field.Type.Type.toString() != "string"
  ) {
    return value;
  } else {
    return value //
      .toString()
      .replaceAll("{{", `{{"{{"}}`)
      .replaceAll("\n", "\\n");
  }
}
registerTemplateFunc("postProcessBasicValue", postProcessBasicValue);

function testGroupArtifactGradleCoordinates(testGroup: string): string {
  switch (testGroup) {
    case "primary":
      return "org.openapis:openapi:0.0.1";
    case "java25-refresh":
      return "org.openapis:openapi:0.0.1";
    case "secondary":
      return "org.openapis.secondary:openapi:0.0.1";
    case "tertiary":
      return "org.openapis.tertiary:openapi:0.0.1";
    case "quaternary":
      return "org.openapis.quaternary:openapi:0.0.1";
    case "client-credentials":
      return "org.openapis.clientcredentials:openapi:0.0.1";
    case "client-credentials-basic":
      return "org.openapis.clientcredentialsbasic:openapi:0.0.1";
    case "oauth2-password":
      return "org.openapis.oauth2password:openapi:0.0.1";
    case "custom-http":
      return "org.openapis.customhttp:openapi:0.0.1";
    case "basic-http":
      return "org.openapis.basichttp:basicHttp:0.0.1";
    case "security-options":
      return "org.openapis.securityoptions:openapi:0.0.1";
    default: {
      throw new Error("did not recognise testGroup: " + testGroup);
    }
  }
}
registerTemplateFunc(
  "testGroupArtifactGradleCoordinates",
  testGroupArtifactGradleCoordinates,
);
