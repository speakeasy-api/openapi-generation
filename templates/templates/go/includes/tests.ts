// @ts-ignore
function getTestDirectory(): string {
  // Ensure config.ts is updated with any updates.
  return "tests";
}

// @ts-ignore
function getTestFileName(testGroupName: string): string {
  const directory = getTestDirectory();
  const fileName = `${sanitizeFileName(testGroupName)}_test.go`;
  return `${directory}/${fileName}`;
}

// @ts-ignore
function getTestHelpersFileName(): string {
  const directory = getTestDirectory();
  return `${directory}/testhelpers_test.go`;
}

// @ts-ignore
function sanitizeTestName(name: string): string {
  return caser().ToGoPascal(sanitizeName(name));
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
      statusCodeAccessor = `${resVariableName}.StatusCode`;
      break;
    case "envelope-http":
      statusCodeAccessor = `${resVariableName}.HTTPMeta.Response.StatusCode`;
      break;
    case "flat":
      if (
        response.Responses.find((r) =>
          doesSubResponseContainStatusCode(r, statusCode),
        )?.Content.length > 0 ||
        response.Responses.find((r) => r.Headers) !== undefined
      ) {
        addImport("github.com/stretchr/testify/assert");
        return [`assert.NotNil(t, ${resVariableName})`];
      }

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

      addImport("github.com/stretchr/testify/assert");
      lines.push(
        `assert.NotContains(t, []any{${statusCodesToNotMatch.join(
          ", ",
        )}}, ${statusCodeAccessor})`,
      );
      break;
    case statusCode.toLowerCase().endsWith("xx"):
      addImport("github.com/stretchr/testify/assert");
      lines.push(
        `assert.Contains(t, []any{${getStatusCodesForRange(statusCode).join(
          ", ",
        )}}, ${statusCodeAccessor})`,
      );
      break;
    default:
      addImport("github.com/stretchr/testify/assert");
      lines.push(`assert.Equal(t, ${statusCode}, ${statusCodeAccessor})`);
      break;
  }

  return lines;
}

// @ts-ignore
function addRequiredTestImports() {
  addImport("github.com/stretchr/testify/assert");
}

// @ts-ignore
function addNullableOptionalAssertions(
  lines: string[],
  path: string,
  field: FieldDef,
  example: any,
): string[] {
  switch (true) {
    case field.Optional && example === undefined:
    case field.Nullable && example === null:
      lines.push(`assert.Nil(t, ${path})`);
      break;
  }
  return lines;
}

// @ts-ignore
function addDefinedAssertion(lines: string[], path: string): string[] {
  lines.push(`assert.NotNil(t, ${path})`);

  return lines;
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  lines.push(`assert.NotNil(t, ${path})`);

  return lines;
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  lines.push(`assert.NotEmpty(t, ${path})`);

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
  lines.push(
    `assert.Equal(t, ${templateValue(field, example, false, {
      usageContext: usageContext,
      operation: usageContext?.Operation,
      test: usageContext.Test,
      isTest: true,
      templateDefaultValue: true,
      isResponse: true,
      targetingMockServer: usageContext.UsingMockServer,
    })}, ${assertionValue})`,
  );
  return lines;
}

// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  return `${path}.${sanitizeFieldName(classField.Name)}`;
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
      value = `sortQueryParameters(${value})`;
    }

    const sortSerializedMapsDirectives = directives.filter(
      (d) => "sortSerializedMaps" in d,
    );

    for (const directive of sortSerializedMapsDirectives) {
      const sortSerializedMaps = directive["sortSerializedMaps"];
      value = `sortSerializedMaps(${value}, \`${sortSerializedMaps.regex}\`, "${sortSerializedMaps.delim}")`;
    }
  }

  return value;
}

// @ts-ignore
function templateTestID(test: Test): string {
  const id = getTestInternalID(test);

  if (!id) {
    return "";
  }

  return `\n    recordTest("${id}")\n`;
}
registerTemplateFunc("templateTestID", templateTestID);

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

  addImport("os");

  return (
    indentLines(
      envVars.map((envVar) => `os.Setenv("${envVar.Name}", "${envVar.Value}")`),
      1,
    ) + "\n"
  );
}
registerTemplateFunc("templateTestEnvVars", templateTestEnvVars);

// @ts-ignore
function mutateAssertionValue(
  responseType: TypeDef,
  value: string,
  fieldDef: FieldDef,
  context?: ResponseAssertionContext,
  example?: any,
): string {
  if (getResponseFormat() == "flat" && context?.totalResponses > 1) {
    value += `.${sanitizeFieldName(sanitizeUnionTypeName(fieldDef.Type))}`;
  }

  switch (fieldDef.Type.Type.toString()) {
    case "response-stream":
      return `readBytes(${value})`;
    default:
      return value;
  }
}

// @ts-ignore
function templateTestFieldAccessor(parent: FieldDef, field: FieldDef): string {
  return `.${sanitizeFieldName(field.Name)}`;
}

// @ts-ignore
function templateTestIndexAccessor(parent: FieldDef, idx: string): string {
  return `[${idx}]`;
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
  )}["${fieldName}"]`;
}

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  return sanitizePrivateFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}res`,
  );
}
registerTemplateFunc("getResponseVariableName", getResponseVariableName);

// @ts-ignore
function getOutputVariableName(stepID: string): string {
  return sanitizePrivateFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}outputs`,
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
  let inputs = "";
  let outputVariable = "";

  if (workflow.Inputs) {
    const inputExample = {};

    const context: TemplateValueContext = {
      test: parent,
      isTest: true,
      scope: "usage",
    };

    const exampleName = `${parent.Name}[${workflowStep.StepIdx}]`;

    for (const field of workflow.Inputs.Type.Fields) {
      const example = findExampleByName(field.Type.Examples ?? [], exampleName);

      inputExample[field.OriginalName] = getExampleValue(example, context);
    }

    inputs = `, ${templateValue(
      workflow.Inputs,
      inputExample,
      false,
      context,
    )}`;
  }

  if (workflow.Outputs) {
    const outputVariableName = getOutputVariableName(workflowStep.StepID);
    outputVariable = `${outputVariableName} := `;
    assertions = `\nassert.NotNil(t, ${outputVariableName})`;
  }

  return `\n${outputVariable}${sanitizePrivateMethodName(
    workflow.Name,
  )}(t${inputs})${assertions}`;
}
registerTemplateFunc("templateWorkflow", templateWorkflow);

// @ts-ignore
function templateHelperOutputs(workflow: ArazzoWorkflow): string {
  const value = templateWorkflowOutputsValue(workflow);
  if (!value) {
    return "";
  }

  return `\nreturn &${value}`;
}
registerTemplateFunc("templateHelperOutputs", templateHelperOutputs);
