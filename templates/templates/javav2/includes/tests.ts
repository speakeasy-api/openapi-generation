// @ts-ignore
function getTestDirectory(): string {
  // Ensure config.ts is updated with any updates.
  return "tests";
}

// @ts-ignore
function getTestFileName(testGroupName: string): string {
  return `${getTestSourceDirectory()}/${getTestSimpleClassName(
    testGroupName,
  )}.java`;
}

// @ts-ignore
function getTestHelpersFileName(): string {
  return `${getTestSourceDirectory()}/TestHelpers.java`;
}

function getTestSimpleClassName(testGroupName: string): string {
  return `${sanitizeFileName(testGroupName)}Tests`;
}
registerTemplateFunc("getTestSimpleClassName", getTestSimpleClassName);

// @ts-ignore
function sanitizeTestName(name: string): string {
  return caser().ToPascal(sanitizeName(name));
}
registerTemplateFunc("sanitizeTestName", sanitizeTestName);

// @ts-ignore
function templateStatusCodeAssertion(
  usageContext: UsageContext,
  assertion: Assertion,
): string[] {
  const responseFormat = getResponseFormat();

  const statusCode = assertion.Value as string;
  const response = assertion.Target as ResponseDef;

  const lines = [];

  let statusCodeAccessor = "";

  const resVariableName = getResponseVariableName(usageContext.StepID);

  switch (responseFormat) {
    case "envelope":
      //      javaImportAssertionsMethod("assertEquals");
      //      return `    assertEquals(${response.Code[0]}, res.statusCode());`;
      statusCodeAccessor = `${resVariableName}.statusCode()`;
      break;
    case "envelope-http":
      // TODO HTTPMeta?
      //      javaImportAssertionsMethod("assertEquals");
      //      return `    assertEquals(${response.Code[0]}, res.httpMeta().response().statusCode());`;
      statusCodeAccessor = `${resVariableName}.httpMeta().response().statusCode()`;
      break;
    case "flat":
      // java does not have null returns so we skip asserting that the response is non-null
      return [];
    default:
      throw new Error(`Unknown response format: ${responseFormat}`);
  }

  switch (true) {
    case statusCode == "default":
      const statusCodesToNotMatch = [];

      for (const subResponse of response.Responses) {
        for (const code of subResponse.Code) {
          if (code !== "default") {
            statusCodesToNotMatch.push(...getStatusCodesForRange(code));
          }
        }
      }

      javaImportAssertionsMethod("assertFalse");
      lines.push(
        `assertFalse(${templateJdkCall("List.of")}(${statusCodesToNotMatch.join(
          ", ",
        )}).contains(${statusCodeAccessor}));`,
      );
      break;
    case statusCode.toLowerCase().endsWith("xx"):
      javaImportAssertionsMethod("assertTrue");
      lines.push(
        `assertTrue(${templateJdkCall("List.of")}(${getStatusCodesForRange(
          statusCode,
        ).join(", ")}).contains(${statusCodeAccessor}));`,
      );
      break;
    default:
      javaImportAssertionsMethod("assertEquals");
      lines.push(`assertEquals(${statusCode}, ${statusCodeAccessor});`);
      break;
  }

  return lines;
}

// @ts-ignore
function addRequiredTestImports() {}

// @ts-ignore
function addNullableOptionalAssertions(
  lines: string[],
  path: string,
  field: FieldDef,
  example: any,
): string[] {
  if (!field.Optional && !field.Nullable) {
    return lines;
  }

  // Raw getters return null when absent; always-optional getters return an
  // empty Optional for both absent and explicit null.
  if (isRawGetter()) {
    lines.push(`${javaImportAssertionsMethod("assertNull")}(${path});`);
    return lines;
  }
  if (isAlwaysOptionalGetter()) {
    lines.push(
      `${javaImportAssertionsMethod("assertFalse")}(${removeGetter(
        path,
      )}.isPresent());`,
    );
    return lines;
  }

  switch (true) {
    case field.Optional && field.Nullable:
      if (context.Global.Config.NullFriendlyParameters) {
        lines.push(`${javaImportAssertionsMethod("assertNull")}(${path});`);
      } else {
        lines.push(
          `${javaImportAssertionsMethod("assertFalse")}(${removeGetter(
            path,
          )}.isPresent() && ${path} != null);`,
        );
      }
      break;
    case field.Optional || field.Nullable:
      if (context.Global.Config.NullFriendlyParameters) {
        lines.push(`${javaImportAssertionsMethod("assertNull")}(${path});`);
      } else {
        lines.push(
          `${javaImportAssertionsMethod("assertFalse")}(${removeGetter(
            path,
          )}.isPresent());`,
        );
      }
      break;
  }
  return lines;
}

// @ts-ignore
function addDefinedAssertion(lines: string[], path: string): string[] {
  // TODO is there a assertion to add here?
  return lines;
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  // TODO is there a assertion to add here?
  return lines;
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  lines.push(
    `${javaImportAssertionsMethod("assertFalse")}(${path}.isEmpty());`,
  );

  return lines;
}

// @ts-ignore
function templateAssertion(
  lines: string[],
  assertionValue: string,
  field: FieldDef,
  example: any,
  usageContext: UsageContext,
): string[] {
  const typ = field.Type.Type.toString();
  let expected = templateValue(field, example, false, {
    usageContext: usageContext,
    operation: usageContext?.Operation,
    test: usageContext.Test,
    isTest: true,
    templateDefaultValue: true,
    isResponse: true,
    targetingMockServer: usageContext.UsingMockServer,
  });
  const actual = assertionValue;
  switch (typ) {
    case "bytes":
    case "response-stream":
      lines.push(`${javaImportAssertionsMethod("assertArrayEquals")}(
        ${expected},
        ${actual});`);
      break;
    default:
      lines.push(`${javaImportAssertionsMethod("assertEquals")}(
        ${expected},
        ${actual});`);
      break;
  }

  return lines;
}

// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  return `${path}${fieldAccessor(classField)}`;
}

function optionalFieldAccessor(
  optional: boolean,
  nullable: boolean,
  defaultValue: boolean = false,
): string {
  if (isRawGetter() && (optional || nullable)) {
    return "";
  }
  if (isAlwaysOptionalGetter()) {
    return ".get()";
  }

  if (context.Global.Config.NullFriendlyParameters) {
    if (optional || nullable || defaultValue) {
      return ".orElse(null)";
    }
    return "";
  }

  if (optional || nullable) {
    return ".get()";
  }
  return "";
}

//@ts-ignore
function fieldAccessor(field: FieldDef): string {
  return `.${sanitizeFieldName(field.Name)}()${optionalFieldAccessor(
    field.Optional,
    field.Nullable,
    field.Default ? true : false,
  )}`;
}

// @ts-ignore
function sanitizeValuesWithDirectives(
  extensions: Record<string, any>,
  value: string,
): string {
  if (`x-speakeasy-test-internal-directives` in (extensions ?? {})) {
    const directives = extensions[
      `x-speakeasy-test-internal-directives`
    ] as any[];

    const sortQueryParametersDirective = directives.find(
      (d) => "sortQueryParameters" in d,
    );

    if (sortQueryParametersDirective) {
      value = `${javaImportUtils()}.sortQueryParameters(${value})`;
    }

    const sortSerializedMapsDirectives = directives.filter(
      (d) => "sortSerializedMaps" in d,
    );

    for (const directive of sortSerializedMapsDirectives) {
      const sortSerializedMaps = directive["sortSerializedMaps"];
      value = `${javaImportUtils()}.sortSerializedMaps(${value}, "${escapeJavaString(
        sortSerializedMaps.regex,
      )}", "${sortSerializedMaps.delim}")`;
    }

    const sortJSONObjectKeysDirective = directives.find(
      (d) => "sortJSONObjectKeys" in d,
    );

    if (sortJSONObjectKeysDirective) {
      const directive = sortJSONObjectKeysDirective["sortJSONObjectKeys"];
      let flds = "";
      if (directive.fields) {
        flds = directive.fields.map((x) => ", " + JSON.stringify(x)).join("");
      }
      value = `${javaImportUtils()}.sortJSONObjectKeys(${value}${flds})`;
    }
  }

  return value;
}

function escapeJavaString(s: string): string {
  // need to escape double quote and backslash
  return s.replaceAll('"', '\\"').replaceAll("\\", "\\\\");
}

// @ts-ignore
function templateTestID(test: Test): string {
  const id = getTestInternalID(test);

  if (!id) {
    return "";
  }

  return `        ${javaImportUtils()}.recordTest("${id}");\n`;
}
registerTemplateFunc("templateTestID", templateTestID);

const javaTestGenerationTestsEnabled = true;

function templateTestDisabledAnnotation(test: Test): string {
  const id = getTestInternalID(test);
  if (id == undefined || isTestSkipped(id) || !javaTestGenerationTestsEnabled) {
    return `    @${javaImport(
      "org.junit.jupiter.api.Disabled",
    )} // test marked as skipped for java or generated unit tests not production ready yet\n`;
  } else {
    return "";
  }
}
registerTemplateFunc(
  "templateTestDisabledAnnotation",
  templateTestDisabledAnnotation,
);

// @ts-ignore
function templateTestEnvVars(test: Test): string {
  const id = getTestInternalID(test);

  if (!id) {
    return "";
  }

  const envVars = test.InternalEnvVars;
  if (!envVars) {
    return "";
  }

  if (envVars.length == 0) {
    return "";
  }

  // there is no accepted way to set an environment variable from Java (requires dodgy hack to alter the environment variables cached in memory)
  // The appropriate way is to use System properties (which are mutable in the JVM life)
  // TODO use the `env.` prefixed properties instead of environment vars
  const lines: string[] = [];
  if (envVars.length != 0) {
    lines.push("");
  }
  lines.push(
    ...envVars.map(
      (envVar) =>
        `${javaImport("java.lang.System")}.setProperty("env.${envVar.Name}", "${
          envVar.Value
        }");\n`,
    ),
  );
  return indentLines(lines, 2);
}
registerTemplateFunc("templateTestEnvVars", templateTestEnvVars);

function dereferenceAccessor(
  fieldDef: FieldDef,
  forcedPointer: boolean,
  accessor: string,
): string {
  if (!accessor.endsWith(")")) {
    accessor += "()";
  }

  // Mirrors the $isBody condition in model.java.stmpl. Response body getters
  // return Optional<T> regardless of getterStyle (see templateBodyFieldGetter())
  const typ = fieldDef.Type.Type.toString();
  if (
    forcedPointer &&
    isRawGetter() &&
    (fieldDef.Optional || fieldDef.Nullable) &&
    sanitizeFieldName(fieldDef.Name) === "body" &&
    (typ === "bytes" || typ === "response-stream")
  ) {
    return `${accessor}.get()`;
  }

  return accessor;
}

// @ts-ignore
function mutateAssertionValue(
  responseType: TypeDef,
  value: string,
  fieldDef: FieldDef,
  context?: ResponseAssertionContext,
  example?: any,
): string {
  if (
    getResponseFormat() == "flat" &&
    fieldDef.Type.Type.toString() == "union"
  ) {
    let type = fieldDef.Type;

    if (example !== undefined) {
      const subType = matchTypeWithExample(fieldDef.Type, example)?.type;

      if (subType) {
        type = subType;
      }
    }

    // TODO this isn't right, we need to change the access patterns based on discriminated vs non-discriminated unions
    value += `.${sanitizeFieldName(type.Name)}}`;
  }

  let v = dereferenceAccessor(fieldDef, context?.topLevel, value);
  switch (fieldDef.Type.Type.toString()) {
    case "response-stream":
      return `${javaImportUtils()}.readBytesAndClose(${v})`;
    default:
      return v;
  }
}

// @ts-ignore
function templateTestFieldAccessor(parent: FieldDef, field: FieldDef): string {
  return `${fieldAccessor(field)}`;
}

// @ts-ignore
function templateTestIndexAccessor(parent: FieldDef, idx: string): string {
  return `.get(${idx})`;
}

// @ts-ignore
function templateTestKeyAccessor(parent: FieldDef, key: string): string {
  return `["${key}"]`;
}

// @ts-ignore
function templateTestAdditionalPropertiesAccessor(
  parent: FieldDef,
  additionalPropertiesField: FieldDef,
  fieldName: string,
): string {
  return `.${sanitizeFieldName(
    additionalPropertiesField.Name,
  )}().get("${fieldName}")`;
}

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  return sanitizeFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}res`,
  );
}
registerTemplateFunc("getResponseVariableName", getResponseVariableName);

// @ts-ignore
function getOutputVariableName(stepID: string, accessor?: boolean): string {
  return (
    sanitizeFieldName(
      `${stepID && stepID !== "test" ? stepID + "_" : ""}outputs`,
    ) + (accessor ? ".get()" : "")
  );
}
registerTemplateFunc("getOutputVariableName", getOutputVariableName);

// @ts-ignore
function templateWorkflow(
  parent: ArazzoWorkflow,
  workflowStep: ArazzoWorkflowStep,
): string {
  const workflow = resolveWorkflowStepTarget(workflowStep);
  if (!workflow) {
    return "";
  }

  let assertions = "";
  let outputVariable = "";
  let inputs = "";

  if (workflow.Inputs) {
    const inputExample = {};

    const context: TemplateValueContext = {
      test: parent,
      isTest: true,
      outputLocation: "tests",
    };

    const exampleName = `${parent.Name}[${workflowStep.StepIdx}]`;

    for (const field of workflow.Inputs.Type.Fields) {
      const example = findExampleByName(field.Type.Examples ?? [], exampleName);

      inputExample[field.OriginalName] = getExampleValue(example, context);
    }

    inputs = `TestHelpers.${templateValue(
      workflow.Inputs,
      inputExample,
      false,
      context,
    )}`;
  }

  if (workflow.Outputs) {
    const outputVariableName = getOutputVariableName(workflowStep.StepID);
    const className = `TestHelpers.${sanitizeClassName(
      workflow.Outputs.Type.Name,
    )}`;
    outputVariable = `${javaImportOptional()}<${className}> ${outputVariableName} = `;
    assertions = `\n${javaImportAssertionsMethod(
      "assertTrue",
    )}(${outputVariableName}.isPresent());`;
  }

  return `\n${outputVariable}TestHelpers.${sanitizePrivateMethodName(
    workflow.Name,
  )}(${inputs});${assertions}`;
}
registerTemplateFunc("templateWorkflow", templateWorkflow);

// @ts-ignore
function templateHelperOutputs(workflow: ArazzoWorkflow): string {
  const value = templateWorkflowOutputsValue(workflow, {
    outputLocation: "tests",
  });
  if (!value) {
    return "";
  }

  return `\nreturn Optional.of(${value});`;
}
registerTemplateFunc("templateHelperOutputs", templateHelperOutputs);

// @ts-ignore
function joinAssertions(lines: string[]): string {
  // java spreads assertions across multiple lines for readability purposes so
  // wrapping lines with new Set causes a total mess
  return [...lines].join("\n");
}
