// @ts-ignore
declare const templateFileValue:
  | ((path: string, additionalContext?: TemplateValueContext) => string)
  | undefined;

const THRESHOLD_MAX_FIELD_COUNT = 20;

// @ts-ignore
function templateModelUsage(
  fieldDef: FieldDef,
  indent: number,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  const start = Date.now();

  const result = indentLines(
    [
      templateValue(fieldDef, example, false, {
        ...(additionalContext ?? {}),
        isUsage: true,
      }),
    ],
    indent,
  );

  if (isDebug() && Date.now() - start > 100) {
    logger().Debug(
      `templateModelUsage took ${Date.now() - start}ms num_chars ${
        result.length
      } field ${
        fieldDef.Name && fieldDef.Name !== fieldDef.Type.Name
          ? `${fieldDef.Name}: ${fieldDef.Type.Name}`
          : fieldDef.Type.Name
      }`,
    );
  }

  return result;
}
registerTemplateFunc("templateModelUsage", templateModelUsage);

// Combined function that calls getOperationMethodFieldExample and templateModelUsage
// within the same JS context, avoiding the goja→Go map→goja roundtrip that occurs
// when these are chained through .stmpl Go templates (which loses map key ordering).
// @ts-ignore
function templateModelUsageForOperationField(
  usageContext: UsageContext,
  fieldDef: FieldDef,
  indent: number,
  additionalContext?: TemplateValueContext,
): string {
  const example = getOperationMethodFieldExample(usageContext, fieldDef);
  return templateModelUsage(fieldDef, indent, example, additionalContext);
}
registerTemplateFunc(
  "templateModelUsageForOperationField",
  templateModelUsageForOperationField,
);

// @ts-ignore
function templateFieldUsage(
  fieldDef: FieldDef,
  indent: number,
  parent: TypeDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (example === undefined && fieldDef.Optional) {
    return "";
  }

  if (!additionalContext?.templateDefaultValue && fieldDef.Default) {
    if (fieldDef.Default.Value === example) {
      return "";
    }
  }

  if (fieldDef.Const) {
    if (!shouldTemplateFieldConstValue(fieldDef, additionalContext)) {
      return "";
    }
  }

  const value = templateValue(fieldDef, example, true, additionalContext);
  if (value.trim() === "") {
    return "";
  }

  const key = templateFieldDeclaration(fieldDef, parent, additionalContext);
  return indentLines([`${key}${value}`], indent);
}

// @ts-ignore
function templateFieldValue(field: FieldDef, example);

// @ts-ignore
function templateObject(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  let fields = [];

  additionalContext = transformAdditionalContext(fieldDef, additionalContext);

  let remainingExample = { ...example };

  let additionalPropertiesField = undefined;
  for (const field of fieldDef.Type.Fields) {
    if (field.IsAdditionalProperties) {
      additionalPropertiesField = field;
      continue;
    }

    let fieldExample = remainingExample[originalFieldName(field)];
    delete remainingExample[originalFieldName(field)];

    const fieldUsage = templateFieldUsage(
      field,
      1,
      fieldDef.Type,
      fieldExample,
      additionalContext,
    );
    if (fieldUsage) {
      fields.push(fieldUsage + templateFieldDelimiter());
    }
  }

  if (additionalPropertiesField) {
    const additionalPropertiesExample =
      typeof remainingExample === "object" &&
      Object.keys(remainingExample).length > 0
        ? remainingExample
        : undefined;

    const fieldUsage = templateFieldUsage(
      additionalPropertiesField,
      1,
      fieldDef.Type,
      additionalPropertiesExample,
      additionalContext,
    );
    if (fieldUsage) {
      fields.push(fieldUsage + templateFieldDelimiter());
    }
  }

  let optionalPrefix = "";
  if (fieldDef.Optional || fieldDef.Nullable) {
    optionalPrefix = templateOptionalSymbol();
  }
  if (fields.length == 0) {
    return [
      `${optionalPrefix}${templateType(
        fieldDef.Type,
        additionalContext,
      )}${templateBracket(
        fieldDef.Type,
        true,
        additionalContext,
      )}${templateBracket(fieldDef.Type, false, additionalContext)}`,
    ].join("\n");
  }

  return [
    `${optionalPrefix}${templateType(
      fieldDef.Type,
      additionalContext,
    )}${templateBracket(fieldDef.Type, true, additionalContext)}`,
    fields.join("\n"),
    templateBracket(fieldDef.Type, false, additionalContext),
  ].join("\n");
}

// @ts-ignore
function templateEnum(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  let idx = fieldDef.Type.Enum.Values.indexOf(`${example}`);

  if (idx == -1) {
    return ""; // TODO maybe force to the first enum value?
  }

  return templateEnumValue(fieldDef, idx, additionalContext);
}

// @ts-ignore
function templateArray(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  // We shouldn't render empty arrays for optional field if we are targeting the mock server as it will fail to return an empty array for an optional field
  if (
    example.length == 0 &&
    fieldDef.Optional &&
    additionalContext?.targetingMockServer
  ) {
    return "";
  }

  const arrayValues = [];

  for (let i = 0; i < example.length; i++) {
    let exampleValue = example[i];

    arrayValues.push(
      templateArrayValue(
        templateValue(
          typeDefToFieldDef(
            fieldDef.Type.ItemType,
            fieldDef,
            undefined,
            fieldDef.Type.ContainsNull,
          ),
          exampleValue,
          false,
          additionalContext,
        ),
      ),
    );
  }

  const lines = [
    `${templateType(fieldDef.Type, additionalContext)}${templateBracket(
      fieldDef.Type,
      true,
      additionalContext,
    )}`,
  ];

  const content = indentLines(processArrayValues(arrayValues), 1);
  const closingBracket = templateBracket(
    fieldDef.Type,
    false,
    additionalContext,
  );
  if (arrayValues.length == 0) {
    return lines[0].trim() + content.trim() + closingBracket.trim();
  }

  if (closingBracketOnSeparateLine()) {
    lines.push(content);
    lines.push(closingBracket);
  } else {
    lines.push(content + closingBracket);
  }
  return lines.join("\n");
}

// @ts-ignore
function closingBracketOnSeparateLine(): boolean {
  return true;
}

// @ts-ignore
function templateMap(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  let mapValues = [];

  // We shouldn't render empty maps for optional field if we are targeting the mock server as it will fail to return an empty map for an optional field
  // But when asserting on a response, we still need to render the empty map as the expected value
  if (
    Object.keys(example).length == 0 &&
    fieldDef.Optional &&
    additionalContext?.targetingMockServer &&
    !additionalContext?.isResponse
  ) {
    return "";
  }

  for (const key in example) {
    mapValues.push(
      templateMapValue(
        key,
        templateValue(
          typeDefToFieldDef(
            fieldDef.Type.ItemType,
            fieldDef,
            undefined,
            fieldDef.Type.ContainsNull,
          ),
          example[key],
          false,
          additionalContext,
        ),
      ),
    );
  }

  const lines = [
    `${templateType(fieldDef.Type, additionalContext)}${templateBracket(
      fieldDef.Type,
      true,
      additionalContext,
    )}`,
  ];
  const content = indentLines(processMapValues(mapValues), 1);
  const closingBracket = templateBracket(
    fieldDef.Type,
    false,
    additionalContext,
  );
  if (closingBracketOnSeparateLine()) {
    lines.push(content);
    lines.push(closingBracket);
  } else {
    lines.push(content + closingBracket);
  }
  return lines.join("\n");
}

function templateMultipartFileField(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  let path = EXAMPLE_FILE;
  let fileName = undefined;

  const fileDirectives = parseFileDirectiveFromExample(
    example,
    additionalContext,
  );
  example = fileDirectives.example;
  if (fileDirectives.fileDirective) {
    path = `${fileDirectives.fileDirective.path}`;
    if (fileDirectives.fileDirective.overriddenFileName) {
      fileName = `${fileDirectives.fileDirective.overriddenFileName}`;
    }
  }

  if (
    typeof templateFileValue !== "undefined" &&
    !fileName &&
    !additionalContext?.isTest
  ) {
    return templateFileValue(path, additionalContext);
  }

  const exampleObject = {};

  for (const field of fieldDef.Type.Fields) {
    if (field.Annotations.Has("multipartForm")) {
      const anno = field.Annotations.Get(
        "multipartForm",
      ) as MultipartFormAnnotation;
      if (field.Name == "fileName") {
        exampleObject[field.Name] = fileName ?? path.split("/").pop();
      } else if (anno.Content) {
        exampleObject[field.Name] = `${FILE_DIRECTIVE_KEY} ${path}`;
      }
    }
  }

  return templateObject(fieldDef, exampleObject, additionalContext);
}

// @ts-ignore
function templateAnyValue(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  const subType = matchTypeWithExample(fieldDef.Type, example)?.type;

  if (subType && subType != fieldDef.Type) {
    return templateValue(typeDefToFieldDef(subType, fieldDef), example, false, {
      ...(additionalContext ?? {}),
      withinUnion: true,
      templateDefaultValue: true,
    });
  }

  return templateBasicValue(fieldDef, example, additionalContext);
}

/** Describes additional context for value templating, such as whether the
 * value is being used during response handling or testing. */
type TemplateValueContext = CalculateExampleAdditionalContext & {
  /** [javav2 only]: Indentation count to pass to templateIndent. */
  indent?: number;

  /** The test associated with this context. May not be set here and available through the usageContext instead. TODO: get this set whenever possible */
  test?: ArazzoWorkflow;

  /** Enable if the value templating is occurring as part of request handling. */
  isRequest?: boolean;

  /** [csharp/php only]: Enable if the value templating is occurring within a usage snippet,
   * typically needed to adjust namespace qualification for enums in templateEnumValue. */
  isUsage?: boolean;

  /** [go only]: When templating enum values, "inline" if the usage location is
   * the same as the TypeDef.OutputLocation or that AST value for templating the
   * selector and import. */
  outputLocation?: "inline" | string;

  /** [pythonv2 only]: Owning model information. */
  owningModelInfo?: any;

  /** [terraform only]: When templating enum values, prefix TypeDef.Scope as
   * selector. */
  scope?: string;

  /** Enable if the value templating is occurring as part of security
   * handling. */
  security?: boolean;

  /** [pythonv2 only]: Enabled if value templating should not include const
   * values, such as test response assertions. */
  typedDictDisabled?: boolean;

  /** [pythonv2 only]: Usage location information for functions that do not
   *  support passing it. */
  usageLocation?: string;

  /** [go only]: Enable if value templating should use the separate types
   * package or SDK package import. */
  useTypesPackage?: boolean;

  /** [go only]: Disable automatic pointer wrapping when the caller will take
   * the address of the rendered const value itself. */
  skipPointerWrapping?: boolean;

  /** The usage context associated with the current values being templated if applicable. */
  usageContext?: UsageContext;

  /** Enable if the test templating is targeting the mock server. */
  targetingMockServer?: boolean;

  /** [csharp only]: The parent TypeDef of a security scheme field, used by
   * sanitizeFieldName to detect CS0542 name collisions. */
  schemeParent?: TypeDef;
};

// @ts-ignore
function templateValue(
  fieldDef: FieldDef,
  example?: any,
  parentHasExample?: boolean,
  additionalContext?: TemplateValueContext,
): string {
  const value = (() => {
    const paramAnno = fieldDef.Annotations.Get("param") as ParamAnnotation;
    if (paramAnno) {
      // For tests we don't render globals unless they are hidden from an operation, preferring setting at the method level
      if (
        additionalContext?.isTest &&
        paramAnno.IsGlobal &&
        !paramAnno.Hidden
      ) {
        return "";
      }
      // For non-tests we don't render params that have globals at the operation level
      if (!additionalContext?.isTest && paramAnno.HasGlobal) {
        return "";
      }
      // For params that have globals and are hidden, we also don't render them
      if (paramAnno.HasGlobal && paramAnno.Hidden) {
        return "";
      }
    }

    // If the example is null and it isn't a nullable field then we need to generate an example
    if (example === null && !fieldDef.Nullable) {
      example = undefined;
    }

    // If we are missing a value or we come across a value that is the UseAutoGeneratedExample value then we should calculate the example
    if (example === undefined || example === UseAutoGeneratedExample) {
      example = calculateExample(
        fieldDef,
        undefined,
        undefined,
        parentHasExample || false,
        additionalContext ?? {},
      );
      if (example === TRUNCATED) {
        example = undefined;
      }
    }

    if (example === null) {
      if (fieldDef.Nullable) {
        return templateNullValue(fieldDef);
      } else {
        example = undefined;
      }
    }
    if (example === undefined) {
      return templateOmittedValue(fieldDef, additionalContext);
    }

    if (isExampleReferenceValue(example)) {
      return templateExampleReferenceValue(
        example as ExampleReferenceValue,
        fieldDef,
        additionalContext,
      );
    }

    if (isMultipartFileField(fieldDef) || fieldDef.Type.IsMultipartFile) {
      return templateMultipartFileField(fieldDef, example, additionalContext);
    }

    switch (fieldDef.Type.Type.toString()) {
      case "union":
        return templateUnion(fieldDef, example, additionalContext);
      case "error":
      case "class":
        return templateObject(fieldDef, example, additionalContext);
      case "enum":
        return templateEnum(fieldDef, example, additionalContext);
      case "array":
      case "set":
        return templateArray(fieldDef, example, additionalContext);
      case "map":
        return templateMap(fieldDef, example, additionalContext);
      case "any":
        return templateAnyValue(fieldDef, example, additionalContext);
      case "response-stream":
      case "request-stream": {
        let path = undefined;

        const fileDirectives = parseFileDirectiveFromExample(
          example,
          additionalContext,
        );

        example = fileDirectives.example;

        if (fileDirectives.fileDirective) {
          path = `${fileDirectives.fileDirective.path}`;
        }

        if (example === undefined && path === undefined) {
          path = EXAMPLE_FILE;
        }

        return templateStream(fieldDef, path, example, additionalContext);
      }
      case "event-stream":
        return templateEventStream(fieldDef, example, additionalContext);
      case "jsonl":
        return templateJsonL(fieldDef, example, additionalContext);
      case "request":
        return templateRequest(fieldDef, example, additionalContext);
      case "response":
        return templateResponse(fieldDef, example, additionalContext);
      default:
        return templateBasicValue(fieldDef, example, additionalContext);
    }
  })();

  return templateValueWrapper(fieldDef, value, additionalContext);
}
registerTemplateFunc("templateValue", templateValue);
const templateValueCommon = templateValue;

/** Wraps templateValue by ensuring the given FieldDef has Optional set to
 *  false. */
function templateValueAsRequired(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return templateValue(
    {
      ...fieldDef,
      Optional: false,
    } as FieldDef,
    example,
    false,
    additionalContext,
  );
}

// @ts-ignore
function processMapValues(values: any[]): string[] {
  return values.map((v) => v + templateFieldDelimiter());
}

// @ts-ignore
function processArrayValues(values: any[]): string[] {
  return values.filter((v) => v !== "").map((v) => v + ",");
}

// @ts-ignore
function templateEventStream(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return "";
}

// @ts-ignore
function templateJsonL(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return "";
}

// @ts-ignore
function templateRequest(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return "";
}

// @ts-ignore
function templateResponse(
  fieldDef: FieldDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  return "";
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

  additionalContext = {
    ...(additionalContext ?? {}),
    templateDefaultValue: true, // If we are in a union then the value of the union could be used to discriminate it so we need to render the default value
    withinUnion: true,
  };

  return templateValue(
    typeDefToFieldDef(typeToUse, fieldDef),
    selectedExample,
    false,
    additionalContext,
  );
}

// @ts-ignore
function templateExampleReferenceValue(
  exampleReferenceValue: ExampleReferenceValue,
  targetField?: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return exampleReferenceValue.path;
}

// @ts-ignore
function transformAdditionalContext(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): TemplateValueContext {
  return additionalContext;
}

// @ts-ignore
function templateValueWrapper(
  fieldDef: FieldDef,
  value: string,
  additionalContext?: TemplateValueContext,
) {
  return value;
}
