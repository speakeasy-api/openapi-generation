// @ts-ignore
function getTestDirectory(): string {
  return "src/__tests__";
}

// @ts-ignore
function getTestFileName(testGroupName: string): string {
  const directory = getTestDirectory();
  const fileName = `${sanitizeFileName(testGroupName)}.test.ts`;
  return `${directory}/${fileName}`;
}

// @ts-ignore
function getTestHelpersFileName(): string {
  const directory = getTestDirectory();
  return `${directory}/testhelpers.ts`;
}

// @ts-ignore
function sanitizeTestName(name: string): string {
  // Format the test name replacing snake_case with spaces and capitalizing each word
  return caser()
    .ToSnake(name)
    .replace(/_/g, " ")
    .replace(
      /\w\S*/g,
      (txt) => txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase(),
    );
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
      statusCodeAccessor = `${resVariableName}.${sanitizeFieldName(
        "statusCode",
      )}`;
      break;
    case "envelope-http":
      statusCodeAccessor = `${resVariableName}.${sanitizeFieldName(
        "httpMeta",
      )}.response.status`;
      break;
    case "flat":
      if (
        response.Responses.find((r) =>
          doesSubResponseContainStatusCode(r, statusCode),
        )?.Content.length > 0 ||
        response.Responses.find((r) => r.Headers) !== undefined
      ) {
        addTestImport("vitest", "expect");
        return [`expect(${resVariableName}).toBeDefined();`];
      }
      return [];
    default:
      throw new Error(`Unknown response format: ${responseFormat}`);
  }

  switch (true) {
    case statusCode == "default":
      addTestImport("vitest", "expect");
      const statusCodesToNotMatch = [];

      for (const subResponse of response.Responses) {
        for (const code of subResponse.Code) {
          if (code !== "default") {
            statusCodesToNotMatch.push(...getStatusCodesForRange(code));
          }
        }
      }

      lines.push(
        `expect([${statusCodesToNotMatch.join(
          ", ",
        )}]).not.toContain(${statusCodeAccessor});`,
      );
      break;
    case statusCode.toLowerCase().endsWith("xx"):
      addTestImport("vitest", "expect");
      lines.push(
        `expect([${getStatusCodesForRange(statusCode).join(
          ", ",
        )}]).toContain(${statusCodeAccessor});`,
      );
      break;
    default:
      addTestImport("vitest", "expect");
      lines.push(`expect(${statusCodeAccessor}).toBe(${statusCode});`);
      break;
  }

  return lines;
}

// @ts-ignore
function addRequiredTestImports() {
  addTestImport("vitest", "expect");
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
      lines.push(`expect(${path}).toBeUndefined();`);
      break;
    case field.Nullable && example === null:
      lines.push(`expect(${path}).toBeNull();`);
      break;
  }
  return lines;
}

// @ts-ignore
function addDefinedAssertion(lines: string[], path: string): string[] {
  lines.push(`expect(${path}).toBeDefined();`);

  return lines;
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  lines.push(`expect(${path}).not.toBeNull();`);

  return lines;
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  lines.push(`expect(${path}).not.toHaveLength(0);`);

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
  const valueExpr = templateValue(field, example, false, {
    usageContext: usageContext,
    operation: usageContext?.Operation,
    test: usageContext.Test,
    isTest: true,
    templateDefaultValue: true,
    shouldTemplateConstValue: shouldTemplateConstValue,
    isResponse: true,
    targetingMockServer: usageContext.UsingMockServer,
  });
  // No-zod: response values pass through unchanged, so the JSON.parse result
  // contains every field the upstream sent (e.g. httpbin's full echo). Use
  // toMatchObject for a partial match against the declared shape instead of
  // toEqual — but only when the expected value is an object/array literal.
  // `toMatchObject` rejects primitives like `false` or `"foo"` at the type
  // level, so for those we fall back to toEqual.
  const isObjectLiteral =
    valueExpr.startsWith("{") || valueExpr.startsWith("[");
  const matcher = isNoZod() && isObjectLiteral ? "toMatchObject" : "toEqual";
  lines.push(`expect(${assertionValue}).${matcher}(${valueExpr});`);
  return lines;
}

// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  return `${path}${parentField.Optional ? "?" : ""}${accessModelField(
    classField,
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
      addTestImport("./helpers.js", "sortQueryParameters");
      value = `sortQueryParameters(${value})`;
    }

    const sortSerializedMapsDirectives = directives.filter(
      (d) => "sortSerializedMaps" in d,
    );

    for (const directive of sortSerializedMapsDirectives) {
      const sortSerializedMaps = directive["sortSerializedMaps"];
      addTestImport("./helpers.js", "sortSerializedMaps");
      value = `sortSerializedMaps(${value}, \`${sortSerializedMaps.regex.replaceAll(
        "\\",
        "\\\\",
      )}\`, "${sortSerializedMaps.delim}")`;
    }

    const sortJSONObjectKeysDirective = directives.find(
      (d) => "sortJSONObjectKeys" in d,
    );

    if (sortJSONObjectKeysDirective) {
      const sortJSONObjectKeys =
        sortJSONObjectKeysDirective["sortJSONObjectKeys"];

      addTestImport("./helpers.js", "sortJSONObjectKeys");
      value = `sortJSONObjectKeys(${value}, ${JSON.stringify(
        sortJSONObjectKeys.fields,
      )})`;
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

  addTestImport("./common_helpers.js", "recordTest");
  return `\n  recordTest("${id}");\n`;
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
      envVars.map(
        (envVar) => `process.env["${envVar.Name}"] = "${envVar.Value}"`,
      ),
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
  let suffix = "";

  if (
    getResponseFormat() === "flat" &&
    responseType.Type.toString() === "union"
  ) {
    for (const assocType of responseType.AssociatedTypes) {
      if (assocType == fieldDef.Type) {
        let typeToImport = assocType;
        if (typeToImport.IsContainer()) {
          typeToImport = typeToImport.ItemType;
        }

        addUsageImportForType(typeToImport);
        suffix = ` as ${sanitizeType(assocType, false, "usage")}`;
        break;
      }
    }
  }

  value = `${value}${suffix}`;

  switch (fieldDef.Type.Type.toString()) {
    case "response-stream":
      addTestImport("./files.js", "streamToByteArray");
      return `new Uint8Array(await streamToByteArray(${value}))`;
    default:
      return value;
  }
}

//@ts-ignore
function fieldAccessor(field: FieldDef): string {
  return `${accessModelField(field)}`;
}

// @ts-ignore
function templateTestFieldAccessor(parent: FieldDef, field: FieldDef): string {
  return `${parent.Optional ? "?" : ""}${accessModelField(field)}`;
}

// @ts-ignore
function templateTestIndexAccessor(parent: FieldDef, idx: string): string {
  return `${parent.Optional ? "?." : ""}[${idx}]`;
}

// @ts-ignore
function templateTestKeyAccessor(parent: FieldDef, key: string): string {
  return `${parent.Optional ? "?." : ""}[${key}]`;
}

// @ts-ignore
function templateTestAdditionalPropertiesAccessor(
  parent: FieldDef,
  additionalPropertiesField: FieldDef,
  fieldName: string,
): string {
  if (shouldFlattenAdditionalProperties(additionalPropertiesField)) {
    // When using flat additional properties, access directly via index signature
    return `${parent.Optional ? "?." : ""}["${fieldName}"]`;
  }
  return `${parent.Optional ? "?" : ""}.${sanitizeFieldName(
    additionalPropertiesField.Name,
  )}${additionalPropertiesField.Optional ? "?." : ""}["${fieldName}"]`;
}

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  return sanitizeFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}result`,
  );
}
registerTemplateFunc("getResponseVariableName", getResponseVariableName);

// @ts-ignore
function getResponseContentVariablePath(
  usageContext: UsageContext,
  content: ResponseBodyContent,
): string {
  let basePath = getResponseVariableName(usageContext.StepID);

  const responseFormat = getResponseFormat();

  switch (true) {
    case responseFormat != "flat":
      basePath += `${fieldAccessor(content.Content)}`;
      break;
    case responseFormat == "flat" &&
      usageContext.Operation.Response.Type?.ResponseEnvelope:
      // In flat response format with an envelope, methods return T | undefined.
      // The envelope result field uses non-optional access (e.g., .result),
      // so add a non-null assertion after the runtime toBeDefined() check.
      basePath += "!";
      const resultField = getResultField(usageContext.Operation);

      if (resultField) {
        basePath += `${fieldAccessor(resultField)}`;
      }
      break;
  }

  return basePath;
}

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
    outputVariable = `const ${outputVariableName} = `;
    assertions = `\nexpect(${outputVariableName}).toBeDefined();`;
  }

  const methodName = sanitizePrivateMethodName(workflow.Name);
  addTestImport("./testhelpers.js", methodName);
  return `\n${outputVariable}await ${sanitizePrivateMethodName(
    workflow.Name,
  )}(${inputs});${assertions}`;
}
registerTemplateFunc("templateWorkflow", templateWorkflow);

// @ts-ignore
function templateHelperOutputs(workflow: ArazzoWorkflow): string {
  const value = templateWorkflowOutputsValue(workflow);
  if (!value) {
    return "";
  }

  return `\nreturn ${value};`;
}
registerTemplateFunc("templateHelperOutputs", templateHelperOutputs);
