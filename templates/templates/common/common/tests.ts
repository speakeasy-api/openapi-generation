// @ts-ignore
// TODO: deprecated remove once all templates are using the job system
function templateTests(ast: AST, ext: string) {
  const jobs = getTestJobs(ast, ext);

  for (const job of jobs) {
    templateFileJob(job);
  }
}

// @ts-ignore
function getTestJobs(ast: AST, ext: string): TemplateFileJob[] {
  const testGroups = ast.Tests?.TestGroups ?? [];

  const jobs = [];

  if (testGroups.length == 0) {
    return jobs;
  }

  if (ast.Tests?.GenerateExampleFile || true) {
    const exampleFile = ".speakeasy/testfiles/example.file"; // TODO we will need to deal with .gen eventually
    addUntrackedPattern("^\\.speakeasy/testfiles/example\\.file");

    let exampleFileData = readFile(exampleFile);

    if (!exampleFileData) {
      jobs.push(createTemplateFileJob("example.file.stmpl", exampleFile, {}));
    }
  }

  for (const testGroup of testGroups) {
    if (testGroup.Tests?.length == 0) {
      continue;
    }

    jobs.push(
      createTemplateFileJob(
        `testfile.${ext}.stmpl`,
        getTestFileName(testGroup.Name),
        {
          Name: testGroup.Name,
          Tests: testGroup.Tests,
        },
      ),
    );
  }

  return jobs;
}

// @ts-ignore
// TODO: deprecated remove once all templates are using the job system
function templateTestHelpers(ast: AST, ext: string) {
  const jobs = getTestHelperJobs(ast, ext);

  for (const job of jobs) {
    templateFileJob(job);
  }
}

/**
 * Resolves the workflow definition targeted by a workflow step.
 *
 * Prefers the embedded resolved workflow when present and otherwise falls back
 * to global workflow lookup by WorkflowID.
 */
// @ts-ignore
function resolveWorkflowStepTarget(
  workflowStep: ArazzoWorkflowStep,
): ArazzoWorkflow | undefined {
  if (workflowStep.Workflow) {
    return workflowStep.Workflow;
  }

  // Contract: workflowStep.WorkflowID maps to ArazzoWorkflow.Name.
  return context.Global.AST.Arazzo?.Workflows?.find(
    (workflow) => workflow.Name === workflowStep.WorkflowID,
  );
}

// @ts-ignore
function getTestHelperJobs(ast: AST, ext: string): TemplateFileJob[] {
  const testGroups = ast.Tests?.TestGroups ?? [];

  if (!testGroups.length) {
    return [];
  }

  const helpersByName = new Map<string, ArazzoWorkflow>();
  const queue: ArazzoWorkflow[] = [];

  const enqueueWorkflow = (workflow: ArazzoWorkflow | undefined) => {
    if (!workflow || helpersByName.has(workflow.Name)) {
      return;
    }

    helpersByName.set(workflow.Name, workflow);
    queue.push(workflow);
  };

  for (const testGroup of testGroups) {
    if (!testGroup.Tests?.length) {
      continue;
    }

    for (const test of testGroup.Tests) {
      const steps =
        test.Workflow?.Steps?.filter(
          (s): s is ArazzoWorkflowStep => s.Type == "workflow",
        ) || [];

      for (const step of steps) {
        enqueueWorkflow(resolveWorkflowStepTarget(step));
      }
    }
  }

  while (queue.length > 0) {
    const workflow = queue.shift();
    if (!workflow) {
      continue;
    }

    const nestedSteps = workflow.Steps?.filter(
      (s): s is ArazzoWorkflowStep => s.Type == "workflow",
    );

    for (const nestedStep of nestedSteps ?? []) {
      enqueueWorkflow(resolveWorkflowStepTarget(nestedStep));
    }
  }

  const helpers = [...helpersByName.values()];

  if (helpers.length == 0) {
    return [];
  }

  return [
    createTemplateFileJob(
      `testhelpers.${ext}.stmpl`,
      getTestHelpersFileName(),
      {
        Helpers: helpers,
      },
    ),
  ];
}

// @ts-ignore
function getTestFileName(testGroupName: string): string {
  throw new Error("getTestFileName not implemented by target");
}

// @ts-ignore
function getTestHelpersFileName(): string {
  throw new Error("getTestHelpersFileName not implemented by target");
}

// @ts-ignore
function templateStatusCodeAssertion(
  usageContext: UsageContext,
  assertion: Assertion,
): string[] {
  throw new Error("templateStatusCodeAssertion not implemented by target");
}

// @ts-ignore
function templateAssertions(usageContext: UsageContext): string {
  if (!usageContext?.Operation) {
    return "";
  }

  const lines = [];

  const assertions = usageContext.Assertions ?? [];

  for (const assertion of assertions) {
    switch (assertion.TargetType.toString()) {
      case "statusCode":
        const statusCodeAssertion = templateStatusCodeAssertion(
          usageContext,
          assertion,
        );
        if (statusCodeAssertion) {
          lines.push(...statusCodeAssertion);
        }
        break;
      case "responseBody":
        if (usageContext?.SkipResponseBodyAssertions) {
          continue;
        }
        const responseBodyAssertion = templateResponseBodyAssertion(
          usageContext,
          assertion,
        );
        if (responseBodyAssertion) {
          lines.push(...responseBodyAssertion);
        }
        break;
    }
  }
  return joinAssertions(lines);
}

registerTemplateFunc("templateAssertions", templateAssertions);

// @ts-ignore
function joinAssertions(lines: string[]): string {
  // a Set will remove duplicate statements and is safe
  // as long as statements don't run across lines (in java
  // they do)
  return [...new Set(lines)].join("\n");
}

type ResponseAssertionContext = {
  topLevel: boolean;
  totalResponses: number;
};

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
      test: usageContext.Test,
    },
  );
  const topLevel = !responseAssertion.Path || responseAssertion.Path === "/";

  let example = undefined;
  if (assertion.Type.toString() != "notEqual" || responseAssertion.Value) {
    example = getExampleValue(responseAssertion.Value, {
      usageContext: usageContext,
      test: usageContext.Test,
    });
  }

  lines.push(
    ...templateResponseAssertions(
      topLevelResponseType,
      target,
      path.path,
      usageContext,
      assertion.Type,
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

// @ts-ignore
function templateResponseAssertions(
  responseType: TypeDef,
  resField: FieldDef,
  path: string,
  usageContext: UsageContext,
  type: AssertionType,
  context?: ResponseAssertionContext,
  example?: any,
): string[] {
  let lines = [];

  addRequiredTestImports();

  if (
    `x-speakeasy-test-ignore` in (resField.Type.Extensions?.All ?? {}) &&
    resField.Type.Extensions?.All[`x-speakeasy-test-ignore`]
  ) {
    return lines;
  }

  if (type.toString() == "notEqual" && example === undefined) {
    switch (true) {
      case resField.Nullable:
        lines = addNotNullAssertion(lines, path);
        break;
      case resField.Type.Type.toString() == "array":
      case resField.Type.Type.toString() == "map":
      case resField.Type.Type.toString() == "string":
        lines = addNotEmptyAssertion(lines, path);
        break;
    }
    return lines;
  }

  if (example === undefined && resField.Type.Examples?.length > 0) {
    example = getExampleValue(
      faker.helpers.arrayElement(resField.Type.Examples),
      { usageContext: usageContext, test: usageContext.Test },
    );
  }

  if (
    !context?.topLevel &&
    ((resField.Optional && example === undefined) ||
      (example === null && resField.Nullable))
  ) {
    return addNullableOptionalAssertions(lines, path, resField, example);
  }

  // TODO might need to check if path.includes(".") this was done in python for some reason
  if (resField.Optional) {
    lines = addDefinedAssertion(lines, path, resField);
  } else if (resField.Nullable) {
    lines = addNotNullAssertion(lines, path);
  }

  let assertionValue = sanitizeValuesWithDirectives(
    resField.Type.Extensions.All,
    path,
  );

  assertionValue = mutateAssertionValue(
    responseType,
    assertionValue,
    resField,
    context,
    example,
  );

  switch (resField.Type.Type.toString()) {
    case "class":
      if (
        resField.Type.Fields.length == 0 ||
        !explodeAssertions(resField.Type)
      ) {
        lines = templateAssertion(
          lines,
          assertionValue,
          resField,
          example,
          usageContext,
        );
      } else {
        for (const field of resField.Type.Fields) {
          let fieldExample = example?.[field.OriginalName];

          lines.push(
            ...templateResponseAssertions(
              responseType,
              field,
              getTestResponseClassFieldPath(path, resField, field),
              usageContext,
              type,
              undefined,
              fieldExample,
            ),
          );
        }
      }
      break;
    default:
      lines = templateAssertion(
        lines,
        assertionValue,
        resField,
        example,
        usageContext,
      );
      break;
  }

  return lines;
}

//@ts-ignore
function fieldAccessor(field: FieldDef): string {
  //@ts-ignore
  return `.${sanitizeFieldName(field?.Name)}`;
}

function sdkHasTests(ast: AST): boolean {
  return (
    context.Global.AST.MainSDK.OutputTests ||
    ast.Tests?.TestGroups?.flatMap((group) => group.Tests)?.length > 0
  );
}

registerTemplateFunc("sdkHasTests", sdkHasTests);

// @ts-ignore
function explodeAssertions(typeDef: TypeDef): boolean {
  return doesTypeIgnoreFields(typeDef) || doesClassHaveChildDirectives(typeDef);
}

function doesTypeIgnoreFields(
  typeDef: TypeDef,
  visited: TypeDef[] = [],
): boolean {
  if (`x-speakeasy-test-ignore` in (typeDef.Extensions?.All ?? {})) {
    return true;
  }
  if (typeDef.Type.toString() !== "class") {
    return false;
  }
  if (!visited.includes(typeDef)) {
    visited.push(typeDef);
  }

  return typeDef.Fields.filter((f) => !visited.includes(f.Type)).some((f) =>
    doesTypeIgnoreFields(f.Type, visited),
  );
}

function doesClassHaveChildDirectives(
  typeDef: TypeDef,
  visited: TypeDef[] = [],
): boolean {
  if (typeDef.Type.toString() !== "class") {
    return false;
  }
  if (!visited.includes(typeDef)) {
    visited.push(typeDef);
  }

  return typeDef.Fields.filter((f) => !visited.includes(f.Type)).some((f) => {
    if (
      `x-speakeasy-test-internal-directives` in (f.Type.Extensions?.All ?? {})
    ) {
      return true;
    }

    return doesClassHaveChildDirectives(f.Type, visited);
  });
}

function templateTestProject(fileNameFunc: (u: UsageContext) => string): void {
  templateDirectory("testproject", context.Global.Config.TestProject.OutputDir);
  context.Global.Config.TestProject.UsageContexts?.forEach((u) => {
    // Skip usage snippets for webhooks
    if (u.Operation.Webhook) return;
    templateFile(
      "usage/test.stmpl",
      `${context.Global.Config.TestProject.OutputDir}/${fileNameFunc(u)}`,
      createUsageContext(
        u.SDK,
        u.Operation,
        u.Operation.Extensions.UsageExample,
      ),
    );
  });
}

function getTestInternalID(test: Test): string {
  return test.InternalID;
}

function templateTestIncompleteMessage(incompleteReasons: string[]): string {
  return `incomplete test found please make sure to address the following errors: [${incompleteReasons
    .map((r) => `\`${r}\``)
    .join(", ")}]`.replaceAll('"', '\\"');
}

registerTemplateFunc(
  "templateTestIncompleteMessage",
  templateTestIncompleteMessage,
);

/** Returns the response associated based on test status code or example name. */
// TODO this may need to take in the content type provided in the assertions as well
function getUsageContextSubResponse(
  usageContext: UsageContext,
  statusCode: string,
): SubResponse {
  return usageContext.Operation.Response.Responses.find((r) => {
    if (doesSubResponseContainStatusCode(r, statusCode)) {
      return true;
    }

    // TODO do we actually need to check more than just the status code?
    if (usageContext.ExampleName.startsWith("speakeasy-default")) {
      return !r.Error; // TODO we prob want to have an order of precedence on status codes
    } else {
      return r.Content.some((c) =>
        c.Examples.some((e) => e.Name() == usageContext.ExampleName),
      );
    }
  });
}

function getResponseBodyContentFromStatusCode(
  usageContext: UsageContext,
  statusCode: string,
): ResponseBodyContent | undefined {
  const subResponse = usageContext.Operation.Response.Responses.find((r) =>
    doesSubResponseContainStatusCode(r, statusCode),
  );
  if (!subResponse) {
    return undefined;
  }
  return getResponseBodyContent(usageContext, subResponse);
}

function doesSubResponseContainStatusCode(
  subResponse: SubResponse,
  statusCode: string,
): boolean {
  if (!statusCode) {
    return false;
  }

  const codes = subResponse.Code.map((c) => c.toLowerCase());

  if (codes.includes(statusCode.toLowerCase())) {
    return true;
  }

  const wildcard = statusCode[0] + "xx";
  return subResponse.Code.map((c) => c.toLowerCase()).includes(wildcard);
}

/** Returns the associated response content based on example name. */
function getResponseBodyContent(
  usageContext: UsageContext,
  response: SubResponse,
  responseBodyAssertion?: Assertion,
): ResponseBodyContent | undefined {
  let content: ResponseBodyContent | undefined;

  if (
    usageContext.ExampleName.startsWith("speakeasy-default") &&
    response.Content.length > 0
  ) {
    content = response.Content[0];
  } else {
    content = response.Content.find((c) =>
      c.Examples.some((e) => e.Name() == usageContext.ExampleName),
    );
  }

  if (content) {
    return content;
  }

  if (!responseBodyAssertion && response.Content.length > 0) {
    return response.Content[0];
  }

  if (responseBodyAssertion && context.Global.Config.TestGroup != "") {
    throw new Error(
      "if responseBodyAssertion is provided then we really shouldn't end up here and it should have been handled earlier",
    );
  }
  return undefined;
}

/** Returns the example value based on example name. */
function getResponseContentValue(
  usageContext: UsageContext,
  content: ResponseBodyContent,
): Example {
  if (content.Examples?.length == 0) {
    return undefined;
  }

  const foundExample = content.Examples.find(
    (e) => e.Name() == usageContext.ExampleName,
  );
  if (foundExample) {
    return foundExample;
  } else {
    return content.Examples[0];
  }
}

registerTemplateFunc("getResponseContentValue", getResponseContentValue);

function templateJSONPointerPath(
  field: FieldDef,
  basePath: string,
  jp: string,
  fnOnPart?: (partIdx: number, parent: FieldDef, currentPath: string) => void,
  additionalContext?: TemplateValueContext,
): {
  path: string;
  target: FieldDef;
} {
  if (jp.startsWith("/")) {
    jp = jp.substring(1);
  }

  if (!jp) {
    return { path: basePath, target: field };
  }

  const parts = jp.split("/");

  let path = basePath;

  let currentField = field;

  for (let i = 0; i < parts.length; i++) {
    let part = parts[i];
    part = part.replaceAll("~1", "/").replaceAll("~0", "~");

    const parent = currentField;

    fnOnPart?.(i, parent, path);

    if (isNumeric(part)) {
      if (
        currentField.Type.Type.toString() !== "array" &&
        currentField.Type.Type.toString() !== "set"
      ) {
        // TODO probably need more graceful error handling here
        throw new Error(
          `JSON Pointer path ${jp} is invalid as part ${i} references a non-array type ${currentField.Type.Type}`,
        );
      }

      currentField = typeDefToFieldDef(
        currentField.Type.ItemType,
        currentField,
      );

      path += templateTestIndexAccessor(parent, part);
    } else {
      // TODO deal with unions
      switch (true) {
        case currentField.Type.IsTypeWithFields():
          const additionalPropertiesField = currentField.Type.Fields.find(
            (f) => f.IsAdditionalProperties,
          );

          currentField = parent.Type.Fields.find((f) => f.Name === part);

          if (currentField) {
            path += templateTestFieldAccessor(
              parent,
              currentField,
              additionalContext,
            );
          } else if (additionalPropertiesField) {
            currentField = typeDefToFieldDef(
              additionalPropertiesField.Type.ItemType,
              additionalPropertiesField,
            );
            fnOnPart?.(
              -1, // Special case for additional properties might want to make this i + 1 instead
              additionalPropertiesField,
              path +
                templateTestFieldAccessor(
                  parent,
                  additionalPropertiesField,
                  additionalContext,
                ),
            );
            path += templateTestAdditionalPropertiesAccessor(
              parent,
              additionalPropertiesField,
              part,
            );
          } else {
            throw new Error(
              `Field ${currentField.Name} does not have field ${part} in type ${currentField.Type.Type}`,
            );
          }
          break;
        case currentField.Type.Type.toString() === "map":
          currentField = typeDefToFieldDef(
            currentField.Type.ItemType,
            currentField,
          );
          path += templateTestKeyAccessor(parent, part);
          break;
        default:
          throw new Error(
            `JSON Pointer path ${jp} is invalid as field ${part} references a non-object type ${currentField.Type.Type}`,
          );
      }
    }
  }

  return { path: path, target: currentField };
}

// @ts-ignore
function addRequiredTestImports() {
  throw new Error("addRequiredTestImports not implemented by target");
}

// @ts-ignore
function addNullableOptionalAssertions(
  lines: string[],
  path: string,
  field: FieldDef,
  example: any,
): string[] {
  throw new Error("addNullableOptionalAssertions not implemented by target");
}

// @ts-ignore
function addDefinedAssertion(
  lines: string[],
  path: string,
  resField?: FieldDef,
): string[] {
  throw new Error("addDefinedAssertion not implemented by target");
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  throw new Error("addNotNullAssertion not implemented by target");
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  throw new Error("addNotEmptyAssertion not implemented by target");
}

// @ts-ignore
function mutateAssertionValue(
  responseType: TypeDef,
  value: string,
  fieldDef: FieldDef,
  context?: ResponseAssertionContext,
  example?: any,
): string {
  throw new Error("mutateAssertionValue not implemented by target");
}

// @ts-ignore
function templateAssertion(
  lines: string[],
  assertionValue: string,
  field: FieldDef,
  example: any,
  usageContext?: UsageContext,
): string[] {
  throw new Error("templateAssertion not implemented by target");
}

// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  throw new Error("getTestResponseClassFieldPath not implemented by target");
}

// @ts-ignore
function sanitizeValuesWithDirectives(
  extensions: Record<string, any>,
  value: string,
): string {
  throw new Error("sanitizeValuesWithDirectives not implemented by target");
}

// @ts-ignore
function templateTestFieldAccessor(
  parent: FieldDef,
  field: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  throw new Error("templateTestFieldAccessor not implemented by target");
}

// @ts-ignore
function templateTestIndexAccessor(parent: FieldDef, idx: string): string {
  throw new Error("templateTestIndexAccessor not implemented by target");
}

// @ts-ignore
function templateTestKeyAccessor(parent: FieldDef, key: string): string {
  throw new Error("templateTestKeyAccessor not implemented by target");
}

// @ts-ignore
function templateTestAdditionalPropertiesAccessor(
  parent: FieldDef,
  additionalPropertiesField: FieldDef,
  fieldName: string,
): string {
  throw new Error(
    "templateTestAdditionalPropertiesAccessor not implemented by target",
  );
}

// @ts-ignore
function templateTestFieldAccess(
  parent: FieldDef,
  currentPath: string,
): string[] {
  return [];
}

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  throw new Error("getResponseVariableName not implemented by target");
}

// @ts-ignore
function getOutputVariableName(stepID: string, accessor?: boolean): string {
  throw new Error("getOutputVariableName not implemented by target");
}

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
      const resultField = getResultField(usageContext.Operation);

      if (resultField) {
        basePath += `${fieldAccessor(resultField)}`;
      }
      break;
  }

  return basePath;
}

function templateWorkflowOutputsValue(
  workflow: ArazzoWorkflow,
  additionalContext?: TemplateValueContext,
): string {
  const usageContext = getWorkflowOutputsUsageContext(workflow);
  if (!usageContext) {
    return "";
  }

  const context = {
    ...(additionalContext ?? {}),
    usageContext,
    test: workflow,
  };

  const outputExample = {};

  for (const field of workflow.Outputs.Type.Fields) {
    const fieldExample = findExampleByName(
      field.Type.Examples ?? [],
      workflow.Name,
    );
    if (fieldExample) {
      outputExample[originalFieldName(field)] = getExampleValue(
        fieldExample,
        context,
      );
    }
  }

  return templateValue(workflow.Outputs, outputExample, false, context);
}

// WARNING: This assumes workflow output references can be resolved from the
// first operation step's usage context. That is brittle because a workflow
// output can reference values produced by later steps, which may require that
// step's response variable/path context instead.
function getWorkflowOutputsUsageContext(
  workflow: ArazzoWorkflow,
): UsageContext | undefined {
  if (!workflow.Outputs) {
    return undefined;
  }

  const firstStep = workflow.Steps?.[0];
  if (!firstStep || firstStep.Type !== "operation") {
    return undefined;
  }

  // TODO: Resolve output references against the step that produced them.
  return firstStep.UsageContext;
}
