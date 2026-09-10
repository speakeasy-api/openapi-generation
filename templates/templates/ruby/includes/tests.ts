// Checks if an example value contains the "<value>" placeholder anywhere.
// This placeholder indicates the arazzo generator could not determine
// the actual response value (e.g. for primitive union types).
function containsPlaceholderValue(example: any): boolean {
  if (example === "<value>") {
    return true;
  }
  if (example !== null && typeof example === "object") {
    for (const key of Object.keys(example)) {
      if (containsPlaceholderValue(example[key])) {
        return true;
      }
    }
  }
  if (Array.isArray(example)) {
    for (const item of example) {
      if (containsPlaceholderValue(item)) {
        return true;
      }
    }
  }
  return false;
}

// Ruby-specific override: handle "<value>" placeholders in response body assertions.
// When the arazzo generator could not determine the actual response value,
// fall back to a notEqual assertion instead of comparing against the placeholder.
// @ts-ignore
function templateResponseBodyAssertion(
  usageContext: UsageContext,
  assertion: Assertion,
): string[] {
  const target = !isFieldDef(assertion.Target)
    ? typeDefToFieldDef(assertion.Target)
    : (assertion.Target as FieldDef);

  const responseAssertion = assertion.Value as ResponseBodyAssertion;
  const content = responseAssertion.Content.Content;

  let basePath = getResponseContentVariablePath(
    usageContext,
    responseAssertion.Content,
  );

  const responseFormat = getResponseFormat();

  let lines = [];

  let topLevelResponseType = usageContext.Operation.Response.Type;

  switch (true) {
    case responseFormat != "flat":
      lines.push(...templateTestFieldAccess(content, basePath));
      break;
    case responseFormat == "flat" &&
      usageContext.Operation.Response.Type?.ResponseEnvelope:
      const resultField = getResultField(usageContext.Operation);

      if (resultField) {
        topLevelResponseType = resultField.Type;
        lines.push(...templateTestFieldAccess(resultField, basePath));
      }
      break;
  }

  const path = templateJSONPointerPath(
    content,
    basePath,
    responseAssertion.Path,
    (partIdx, parent, currentPath) => {
      if (partIdx != 0) {
        lines.push(...templateTestFieldAccess(parent, currentPath));
      }
    },
    {
      usageContext: usageContext,
    },
  );
  const topLevel = !responseAssertion.Path || responseAssertion.Path === "/";

  let example = undefined;
  if (assertion.Type.toString() != "notEqual" || responseAssertion.Value) {
    example = getExampleValue(responseAssertion.Value, {
      usageContext: usageContext,
    });
  }

  // Treat "<value>" placeholders as missing examples since they indicate the
  // arazzo generator could not determine the actual response value (e.g. for
  // primitive union types). Fall back to a notEqual assertion in this case.
  let assertionType = assertion.Type;
  if (containsPlaceholderValue(example)) {
    example = undefined;
    assertionType = "notEqual" as AssertionType;
  }

  lines.push(
    ...templateResponseAssertions(
      topLevelResponseType,
      target,
      path.path,
      usageContext,
      assertionType,
      {
        topLevel: topLevel,
        totalResponses: usageContext.Operation.Response.Responses.reduce(
          (total, r) => total + r.Content.length,
          0,
        ),
      },
      example,
    ),
  );

  return lines;
}

// Strip common leading whitespace from a multi-line string
// @ts-ignore
function dedent(input: string): string {
  const lines = input.split("\n");
  const nonEmptyLines = lines.filter((l) => l.trim().length > 0);
  if (nonEmptyLines.length === 0) return input;
  const minIndent = Math.min(
    ...nonEmptyLines.map((l) => l.match(/^\s*/)[0].length),
  );
  if (minIndent === 0) return input;
  return lines.map((l) => l.slice(minIndent)).join("\n");
}
registerTemplateFunc("dedent", dedent);

// @ts-ignore
function getTestDirectory(): string {
  // Ensure config.ts is updated with any updates.
  return "test";
}

// @ts-ignore
function getTestFileName(testGroupName: string): string {
  const directory = getTestDirectory();
  const fileName = `${sanitizeFileName(testGroupName)}_test.rb`;
  return `${directory}/${fileName}`;
}

// @ts-ignore
function getTestHelpersFileName(): string {
  const directory = getTestDirectory();
  const fileName = `test_helpers.rb`;
  return `${directory}/${fileName}`;
}

// @ts-ignore
function sanitizeTestName(name: string): string {
  return caser().ToSnake(sanitizeName(name));
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
      statusCodeAccessor = `${resVariableName}.status_code`;
      break;
    case "envelope-http":
      lines.push(`refute_nil(${resVariableName}.http_meta)`);
      lines.push(`refute_nil(${resVariableName}.http_meta.response)`);

      statusCodeAccessor = `${resVariableName}.http_meta.response.status`;
      break;
    case "flat":
      if (
        response.Responses.find((r) =>
          doesSubResponseContainStatusCode(r, statusCode),
        )?.Content.length > 0 ||
        response.Responses.find((r) => r.Headers) !== undefined
      ) {
        return [`refute_nil(${resVariableName})`];
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

      lines.push(
        `refute_includes([${statusCodesToNotMatch.join(
          ", ",
        )}], ${statusCodeAccessor})`,
      );
      break;
    case statusCode.toLowerCase().endsWith("xx"):
      lines.push(
        `assert_includes([${getStatusCodesForRange(statusCode).join(
          ", ",
        )}], ${statusCodeAccessor})`,
      );
      break;
    default:
      lines.push(`assert_equal(${statusCode}, ${statusCodeAccessor})`);
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
  switch (true) {
    case field.Optional && example === undefined:
    case field.Nullable && example === null:
      lines.push(`assert_nil(${path})`);
      break;
  }
  return lines;
}

// @ts-ignore
function addDefinedAssertion(lines: string[], path: string): string[] {
  lines.push(`refute_nil(${path})`);

  return lines;
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  lines.push(`refute_nil(${path})`);

  return lines;
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  lines.push(`refute_nil(${path})`);
  lines.push(`refute_empty(${path})`);

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
  const value = templateValue(field, example, false, {
    usageContext: usageContext,
    operation: usageContext?.Operation,
    test: usageContext.Test,
    isTest: true,
    isResponse: true,
    targetingMockServer: usageContext.UsingMockServer,
  });

  if (value === "true") {
    lines.push(`assert(${assertionValue})`);
  } else if (value === "false") {
    lines.push(`refute(${assertionValue})`);
  } else if (value.replace(/\s+/g, "") === "{}") {
    lines.push(`assert_empty(${assertionValue})`);
  } else {
    lines.push(`assert_equal(${value}, ${assertionValue})`);
  }
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
      const json =
        sortQueryParametersDirective["sortQueryParameters"].jsonParams;
      const jsonArg = json ? `, json_params: ${JSON.stringify(json)}` : "";
      value = `sort_query_parameters(${value}${jsonArg})`;
    }

    const sortSerializedMapsDirectives = directives.filter(
      (d) => "sortSerializedMaps" in d,
    );

    for (const directive of sortSerializedMapsDirectives) {
      const sortSerializedMaps = directive["sortSerializedMaps"];
      const regex = sortSerializedMaps.regex as string;
      let regexLiteral: string;
      if (regex.includes("/")) {
        // Use %r{} for regexes containing slashes, and unescape redundant \/
        regexLiteral = `%r{${regex.replace(/\\\//g, "/")}}`;
      } else {
        regexLiteral = `/${regex}/`;
      }
      value = `sort_serialized_maps(${value}, ${regexLiteral}, '${sortSerializedMaps.delim}')`;
    }

    const sortJSONObjectKeysDirective = directives.find(
      (d) => "sortJSONObjectKeys" in d,
    );

    if (sortJSONObjectKeysDirective) {
      const fields = sortJSONObjectKeysDirective["sortJSONObjectKeys"].fields;
      const fieldsArg = fields ? `, ${JSON.stringify(fields)}` : "";
      value = `sort_json_object_keys(${value}${fieldsArg})`;
    }

    const sanitizeTimestampDirective = directives.find(
      (d) => "sanitizeTimestamp" in d,
    );

    if (sanitizeTimestampDirective) {
      value = `sanitize_timestamp(${value})`;
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

  return `\n      record_test('${id}')\n`;
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

  return (
    indentLines(
      envVars.map((envVar) => `ENV["${envVar.Name}"] = "${envVar.Value}"`),
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
  return value;
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
  return `['${key}']`;
}

// @ts-ignore
function templateTestAdditionalPropertiesAccessor(
  parent: FieldDef,
  additionalPropertiesField: FieldDef,
  fieldName: string,
): string {
  return `.${sanitizeFieldName(
    additionalPropertiesField.Name,
  )}['${fieldName}']`;
}

// @ts-ignore
function templateTestFieldAccess(
  parent: FieldDef,
  currentPath: string,
): string[] {
  let lines = [];
  if ((parent.Optional || parent.Nullable) && !parent.IsAdditionalProperties) {
    lines = addDefinedAssertion(lines, currentPath);
  }
  return lines;
}

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  return sanitizeFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}res`,
  );
}
registerTemplateFunc("getResponseVariableName", getResponseVariableName);

// @ts-ignore
function operationHasResponseBody(usageContext: UsageContext): boolean {
  const responseFormat = getResponseFormat();
  if (responseFormat !== "flat") {
    // envelope and envelope-http always return a response object
    return true;
  }
  // In flat mode, the operation returns nil if there's no response type
  return !!usageContext.Operation.Response?.Type;
}
registerTemplateFunc("operationHasResponseBody", operationHasResponseBody);

// @ts-ignore
function getOutputVariableName(stepID: string): string {
  return sanitizeFieldName(
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
  let outputVariable = "";
  let inputs = "";

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

    inputs = templateValue(workflow.Inputs, inputExample, false, context);
  }

  if (workflow.Outputs) {
    const outputVariableName = getOutputVariableName(workflowStep.StepID);
    outputVariable = `${outputVariableName} = `;
    assertions = `\nrefute_nil(${outputVariableName})`;
  }

  const methodName = sanitizePrivateMethodName(workflow.Name);
  const args = inputs ? `(${inputs})` : "";
  return `\n${outputVariable}${methodName}${args}${assertions}`;
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

  return `\nreturn ${value}`;
}
registerTemplateFunc("templateHelperOutputs", templateHelperOutputs);
