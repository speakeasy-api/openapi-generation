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
  if (getResponseFormat() === "flat") {
    return [];
  }

  const statusCode = assertion.Value as string;
  const resVariableName = getResponseVariableName(usageContext.StepID);
  const statusCodeAccessor = isStrictMCPServer()
    ? `${resVariableName}.value.StatusCode`
    : `${resVariableName}.value.status`;

  const lines = [];

  switch (true) {
    case statusCode == "default":
      addTestImport("vitest", "expect");
      const statusCodesToNotMatch = [];

      const response = assertion.Target as ResponseDef;
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
  lines.push(
    `expect(${assertionValue}).toEqual(${templateValue(field, example, false, {
      usageContext: usageContext,
      operation: usageContext?.Operation,
      test: usageContext.Test,
      isTest: true,
      templateDefaultValue: true,
      shouldTemplateConstValue: shouldTemplateConstValue,
      isResponse: true,
      targetingMockServer: usageContext.UsingMockServer,
    })});`,
  );
  return lines;
}

// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  return `${path}${parentField.Optional ? "?" : ""}.${sanitizeFieldName(
    classField.Name,
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

  if (responseType.Type.toString() === "union") {
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

// @ts-ignore
function templateTestFieldAccessor(parent: FieldDef, field: FieldDef): string {
  return `${parent.Optional ? "?" : ""}.${sanitizeFieldName(field.Name)}`;
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
      if (isStrictMCPServer()) {
        basePath += `.value${fieldAccessor(content.Content)}`;
      } else {
        basePath = `(${basePath}.value as any)${fieldAccessor(
          content.Content,
        )}`;
      }
      break;
    case responseFormat == "flat" &&
      usageContext.Operation.Response.Type?.ResponseEnvelope:
      const resultField = getResultField(usageContext.Operation);

      if (resultField) {
        basePath += `.value${fieldAccessor(resultField)}`;
      }
      break;
  }

  return basePath;
}
registerTemplateFunc(
  "getResponseContentVariablePath",
  getResponseContentVariablePath,
);

// @ts-ignore
function templateUsageSDKOptions(local: UsageContext): string {
  const options = [];
  let hasSecurity = local.Operation.Security != undefined;

  const ctx: TemplateValueContext = {
    usageContext: local,
    operation: local.Operation,
    isTest: local.Test && true,
    test: local.Test,
  };

  if (local.Scopes != undefined) {
    local.Scopes.forEach((scope) => {
      if (scope.IsGlobal && scope.Feature != "") {
        const feature = scope.Feature.toString();

        switch (feature) {
          case "security":
            if (!hasSecurity) {
              const globalSecurityUsage = getGlobalSecurity(scope.Value, ctx);
              if (globalSecurityUsage) {
                options.push(globalSecurityUsage);
                hasSecurity = true;
              }
            }
            break;
          case "server_url": {
            const hasOperationServers =
              local.Operation?.Servers?.Servers?.length > 0;
            if (!hasOperationServers) {
              options.push(
                `serverURL: ${getUsageServerUrl(local, scope, ctx)},`,
              );
            }
            break;
          }
          case "parameter": {
            const parameter = scope.Value as ParameterUsage;

            const value = templateValueAsRequired(
              parameter.field,
              getExampleValue(parameter.example),
              ctx,
            );

            if (value === "") {
              return;
            }

            options.push(
              `${templateGlobalFieldName(parameter.field.Name)}: ${value},`,
            );

            break;
          }
        }
      }
    });
  }

  return joinSDKOptions(options);
}
registerTemplateFunc("templateUsageSDKOptions", templateUsageSDKOptions);

// @ts-ignore
function getGlobalSecurity(
  options?: SecurityUsageContext | Example,
  additionalContext?: TemplateValueContext,
): string {
  if (!context.Global.AST.MainSDK.Security) {
    return "";
  }

  let index = undefined;
  let example = undefined;

  if (options) {
    if ("index" in options || "example" in options) {
      ({ index, example } = options);
    } else {
      example = getExampleValue(options as Example, additionalContext);
    }
  }

  if (
    index === undefined &&
    example === undefined &&
    context.Global.AST.MainSDK.Security.Optional &&
    context.Global.AST.MainSDK.SecurityConfig.OptionalityReason ==
      "optional-scheme"
  ) {
    return "";
  }

  let securityFieldName = "security";
  if (
    context.Global.AST.MainSDK.Security.Type.Fields.length == 1 &&
    context.Global.Config.FlattenGlobalSecurity == true
  ) {
    securityFieldName = sanitizeFieldName(
      context.Global.AST.MainSDK.Security.Type.Fields[0].Name,
    );
  }

  return `${securityFieldName}: ${templateSecurityUsage(
    context.Global.AST.MainSDK.Security.Type,
    0,
    true,
    context.Global.Config.FlattenGlobalSecurity,
    index,
    example,
    additionalContext,
  ).trim()},`;
}

// @ts-ignore
function joinSDKOptions(options: string[]): string {
  if (options.length > 0) {
    options.unshift("");

    return "{" + indentLines(options, 1) + "\n}";
  }

  return "";
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

// @ts-ignore
function templateMCPAssertions(usageContext: UsageContext): string {
  if (!isStrictMCPServer()) {
    usageContext = { ...usageContext, SkipResponseBodyAssertions: true };
  }
  return templateAssertions(usageContext);
}
registerTemplateFunc("templateMCPAssertions", templateMCPAssertions);
