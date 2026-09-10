// @ts-ignore
function getTestDirectory(): string {
  // Ensure config.ts is updated with any updates.
  return "Tests";
}

// @ts-ignore
function refNeedsNullForgive(
  ref: ExampleReferenceValue,
  targetField?: FieldDef,
): boolean {
  if (targetField && (targetField.Nullable || targetField.Optional)) {
    return false;
  }

  if (ref.source && canFieldBeNull(ref.source)) {
    return true;
  }

  return ref.path.indexOf("?.") !== -1 || ref.path.indexOf("?[") !== -1;
}

// @ts-ignore
function templateExampleReferenceValue(
  exampleReferenceValue: ExampleReferenceValue,
  targetField?: FieldDef,
  _additionalContext?: TemplateValueContext,
): string {
  const path = exampleReferenceValue.path;
  return refNeedsNullForgive(exampleReferenceValue, targetField)
    ? `${path}!`
    : path;
}

// Collect refs that may be null at the source but feed non-nullable targets.
// @ts-ignore
function collectNullableRefPaths(
  value: any,
  field: FieldDef | undefined,
  paths: Set<string>,
): void {
  if (value == null) return;
  if (field && (field.Nullable || field.Optional)) return;
  if (isExampleReferenceValue(value)) {
    const ref = value as ExampleReferenceValue;
    if (refNeedsNullForgive(ref, field)) {
      paths.add(ref.path);
    }
    return;
  }
  if (Array.isArray(value)) {
    const itemField = field?.Type?.ItemType
      ? typeDefToFieldDef(field.Type.ItemType, field)
      : undefined;
    value.forEach((v) => collectNullableRefPaths(v, itemField, paths));
    return;
  }
  if (typeof value === "object") {
    const subFields: FieldDef[] = field?.Type?.Fields ?? [];
    if (subFields.length > 0) {
      for (const sub of subFields) {
        const v =
          value[sub.OriginalName] ??
          value[originalFieldName(sub)] ??
          value[sub.Name];
        collectNullableRefPaths(v, sub, paths);
      }
    } else {
      for (const k of Object.keys(value)) {
        collectNullableRefPaths(value[k], undefined, paths);
      }
    }
  }
}

// Collects nullable cross-step refs used by an operation step's call args.
// @ts-ignore
function operationStepRefPaths(local: UsageContext): Set<string> {
  const paths = new Set<string>();
  const operation = local?.Operation;
  if (!operation) return paths;

  if (operation.Request) {
    if (operationParametersFlattened(operation)) {
      for (const field of operation.Request.Field.Type.Fields ?? []) {
        collectNullableRefPaths(
          getOperationMethodFieldExample(local, field),
          field,
          paths,
        );
      }
    } else {
      collectNullableRefPaths(
        getOperationMethodFieldExample(local, operation.Request.Field),
        operation.Request.Field,
        paths,
      );
    }
  }

  for (const field of operation.Arguments?.Sorted ?? []) {
    if (operation.Request && field === operation.Request.Field) continue;
    collectNullableRefPaths(
      getOperationMethodFieldExample(local, field),
      field,
      paths,
    );
  }

  return paths;
}

// Emits Assert.NotNull once per nullable ref before its first dependent step.
// @ts-ignore
function templatePreCallAssertions(
  workflow: ArazzoWorkflow,
  stepIdx: number,
): string {
  const steps = workflow?.Steps ?? [];
  if (stepIdx < 0 || stepIdx >= steps.length) return "";
  const current = steps[stepIdx];
  if (current.Type !== "operation") return "";

  const alreadyAsserted = new Set<string>();
  for (let i = 0; i < stepIdx; i++) {
    const prior = steps[i];
    if (prior.Type !== "operation") continue;
    for (const p of operationStepRefPaths(prior.UsageContext)) {
      alreadyAsserted.add(p);
    }
  }

  const lines: string[] = [];
  for (const p of operationStepRefPaths(current.UsageContext)) {
    if (alreadyAsserted.has(p)) continue;
    lines.push(`Assert.NotNull(${p});`);
  }
  return lines.join("\n");
}
registerTemplateFunc("templatePreCallAssertions", templatePreCallAssertions);

// @ts-ignore
function getTestFileName(testGroupName: string): string {
  return `${getTestDirectory()}/${getTestSimpleClassName(testGroupName)}.cs`;
}

// @ts-ignore
function getTestHelpersFileName(): string {
  return `${getTestDirectory()}/TestHelpers.cs`;
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
  const lines = [];

  const resVariableName = getResponseVariableName(usageContext.StepID);
  lines.push(`Assert.NotNull(${resVariableName});`);

  const statusCode = assertion.Value as string;
  const responseFormat = getResponseFormat();
  let statusCodeAccessor = "";
  switch (responseFormat) {
    case "envelope":
      statusCodeAccessor = `${resVariableName}.StatusCode`;
      break;
    case "envelope-http":
      statusCodeAccessor = `(int)${resVariableName}.HttpMeta.Response.StatusCode`;
      break;
    case "flat":
      return [];
    default:
      throw new Error(`Unknown response format: ${responseFormat}`);
  }

  const response = assertion.Target as ResponseDef;
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
        `Assert.DoesNotContain(${statusCodeAccessor}, new int[] { ${statusCodesToNotMatch.join(
          ", ",
        )} });`,
      );
      break;
    case statusCode.toLowerCase().endsWith("xx"):
      lines.push(
        `Assert.Contains(${statusCodeAccessor}, new int[] { ${getStatusCodesForRange(
          statusCode,
        ).join(", ")} });`,
      );
      break;
    default:
      lines.push(`Assert.Equal(${statusCode}, ${statusCodeAccessor});`);
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
  lines.push(`Assert.Null(${path});`);
  return lines;
}

// @ts-ignore
function addDefinedAssertion(
  lines: string[],
  path: string,
  resField?: FieldDef,
): string[] {
  if (resField && !resField.Nullable && !isReferenceType(resField.Type)) {
    return lines;
  }
  lines.push(`Assert.NotNull(${path});`);
  return lines;
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  lines.push(`Assert.NotNull(${path});`);
  return lines;
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  lines.push(`Assert.NotEmpty(${path}?.ToString() ?? "");`);
  return lines;
}

// @ts-ignore
// C# classes don't implement structural equality (IEquatable/Equals), but
// templateAssertion uses NormalizeJson for class/array/map/union types which provides
// structural comparison via JSON serialization. We only explode assertions when the
// common framework requires it (test-ignore extensions or child directives), avoiding
// issues where schema-level example values (e.g. enum defaults) are incorrectly asserted
// for optional fields not present in the test response.
function explodeAssertions(typeDef: TypeDef): boolean {
  // Don't explode types with additional properties since the example extraction doesn't
  // handle them correctly (additional props are at the same level as regular fields).
  if (typeDef.Fields?.some((f: FieldDef) => f.IsAdditionalProperties)) {
    return false;
  }
  // Check for x-speakeasy-test-ignore extension on the type or any child
  if (`x-speakeasy-test-ignore` in (typeDef.Extensions?.All ?? {})) {
    return true;
  }
  // Check if any child field has x-speakeasy-test-internal-directives
  function hasDirectives(td: TypeDef, visited: TypeDef[] = []): boolean {
    if (td.Type.toString() !== "class") return false;
    if (visited.includes(td)) return false;
    visited.push(td);
    return (td.Fields || [])
      .filter((f: FieldDef) => !visited.includes(f.Type))
      .some((f: FieldDef) => {
        if (
          `x-speakeasy-test-internal-directives` in
          (f.Type.Extensions?.All ?? {})
        )
          return true;
        if (`x-speakeasy-test-ignore` in (f.Type.Extensions?.All ?? {}))
          return true;
        return hasDirectives(f.Type, visited);
      });
  }
  function hasIgnoredFields(td: TypeDef, visited: TypeDef[] = []): boolean {
    if (`x-speakeasy-test-ignore` in (td.Extensions?.All ?? {})) return true;
    if (td.Type.toString() !== "class") return false;
    if (visited.includes(td)) return false;
    visited.push(td);
    return (td.Fields || [])
      .filter((f: FieldDef) => !visited.includes(f.Type))
      .some((f: FieldDef) => hasIgnoredFields(f.Type, visited));
  }
  return hasIgnoredFields(typeDef) || hasDirectives(typeDef);
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
  const expected = templateValue(field, example, false, {
    usageContext: usageContext,
    operation: usageContext?.Operation,
    test: usageContext.Test,
    isTest: true,
    templateDefaultValue: true,
    isResponse: true,
    targetingMockServer: usageContext.UsingMockServer,
  });
  const actual = assertionValue;

  // C# classes don't implement structural equality (IEquatable/Equals),
  // so arrays, maps, and non-exploded classes need normalized JSON comparison.
  // This sorts keys and normalizes DateTimes to UTC to avoid format differences.
  // Leaf primitives (string, int, DateTime, decimal, etc.) use Assert.Equal directly.
  const needsJsonComparison =
    typ === "map" || typ === "array" || typ === "class" || typ === "union";

  if (needsJsonComparison) {
    lines.push(
      `Assert.Equal(\n        CommonHelpers.NormalizeJson(${expected}),\n        CommonHelpers.NormalizeJson(${actual}));`,
    );
  } else if (typ === "string") {
    // Normalize DateTime trailing zeros for string comparisons (C# uses 7 digits, Go uses minimal)
    lines.push(
      `Assert.Equal(\n        CommonHelpers.NormalizeDateTimeStrings(${expected}),\n        CommonHelpers.NormalizeDateTimeStrings(${actual}));`,
    );
  } else {
    lines.push(`Assert.Equal(\n        ${expected},\n        ${actual});`);
  }

  return lines;
}

// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  return `${path}${fieldAccessor(classField, parentField)}`;
}

// null-conditional operator
function nullCond(field: FieldDef): string {
  if (!field) {
    return "";
  }

  if (
    field.Type.OutputLocation === "tests" &&
    (field.Type.Name.endsWith("_inputs") ||
      field.Type.Name.endsWith("_outputs"))
  ) {
    return "";
  }

  return canFieldBeNull(field) ? "?" : "";
}

//@ts-ignore
function fieldAccessor(field: FieldDef, parent: FieldDef): string {
  return `${nullCond(parent)}.${sanitizeFieldName(field.Name)}`;
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
      // Normalize URL encoding case before sorting (.NET 6 uses lowercase hex,
      // .NET 8 uses uppercase; delimiters like %2C need consistent casing)
      value = `CommonHelpers.NormalizeUrlEncoding(${value})`;
      value = `CommonHelpers.SortQueryParameters(${value})`;
    }

    const sortSerializedMapsDirectives = directives.filter(
      (d) => "sortSerializedMaps" in d,
    );

    for (const directive of sortSerializedMapsDirectives) {
      const sortSerializedMaps = directive["sortSerializedMaps"];
      value = `CommonHelpers.SortSerializedMaps(${value}, "${escapeCSharpString(
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
      value = `CommonHelpers.SortJSONObjectKeys(${value}${flds})`;
    }
  }

  return value;
}

function escapeCSharpString(s: string): string {
  return s.replaceAll("\\", "\\\\").replaceAll('"', '\\"');
}

// @ts-ignore
function templateTestID(test: Test): string {
  const id = getTestInternalID(test);

  if (!id) {
    return "";
  }

  return `        CommonHelpers.RecordTest("${id}");\n`;
}
registerTemplateFunc("templateTestID", templateTestID);

const csharpTestGenerationTestsEnabled = true;

// Tests skipped by name (for tests without internal IDs)
const testsSkippedByName = [
  // C# SDK usage snippet includes auto-pagination while loop, making extra requests
  // that the mock server doesn't expect
  "listTest1",
  // C# SDK RequestBodySerializer.SerializeMultipart does not handle File types
  "postFile",
  "postFileWithEncoding",
  // C# SDK uses [JsonProperty("additionalProperties")] instead of
  // [JsonExtensionData], so additional properties in JSON responses are not deserialized
  // into the dictionary. Un-skip these once the wire format is fixed.
  "getUser-testWithResponseBody",
  "getUser-testWithResponseBodyFields",
  // C# SDK pagination: while loop leaves res=null, then assertions access res (Java also skips this test)
  "getUnionErrors",
];

function templateTestSkipAnnotation(test: Test): string {
  if (test.Incomplete && test.Incomplete.length > 0) {
    const msg =
      `incomplete test found please make sure to address the following errors: [${test.Incomplete.map(
        (r: string) => `\`${r}\``,
      ).join(", ")}]`.replaceAll('"', '\\"');
    return `(Skip = "${msg}")`;
  }

  const incompatibilityReason = isTestIncompatibleWithConfig(test);
  if (incompatibilityReason != null) {
    return `(Skip = "test marked as skipped for csharp: ${incompatibilityReason}")`;
  }

  if (isTestGenerationSkipped(test)) {
    return `(Skip = "test marked as skipped for csharp: generated body would not compile")`;
  }

  if (testsSkippedByName.includes(test.Name)) {
    return `(Skip = "test marked as skipped for csharp or generated unit tests not production ready yet")`;
  }
  const id = getTestInternalID(test);
  if (
    id == undefined ||
    isTestSkipped(id) ||
    !csharpTestGenerationTestsEnabled
  ) {
    return `(Skip = "test marked as skipped for csharp or generated unit tests not production ready yet")`;
  } else {
    return "";
  }
}
registerTemplateFunc("templateTestSkipAnnotation", templateTestSkipAnnotation);

// Tests that cannot have their body generated due to compile errors or template issues.
// Matches on test Name (not InternalID, which may be empty for Arazzo tests).
const testsSkippedFromGeneration: string[] = [
  // Flat format + union response + string assertion: the AST narrows responseType to
  // "string" but C# returns the union class, causing NormalizeDateTimeStrings(res) to
  // fail compilation (expects string, gets union class).
  "responseBodyOptionalGet-override",
];

function isTestIncompatibleWithConfig(test: Test): string | null {
  // smart-union-* tests validate the `populated-fields` disambiguation feature.
  const internalID = getTestInternalID(test);
  if (
    internalID != undefined &&
    internalID.startsWith("smart-union-") &&
    context.Global.Config.UnionStrategy !== "populated-fields"
  ) {
    return "smart-union test requires unionStrategy: populated-fields";
  }

  return null;
}

function isTestGenerationSkipped(test: Test): boolean {
  return testsSkippedFromGeneration.includes(test.Name);
}
registerTemplateFunc("isTestGenerationSkipped", isTestGenerationSkipped);

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

  const lines: string[] = [];
  if (envVars.length != 0) {
    lines.push("");
  }
  lines.push(
    ...envVars.map(
      (envVar) =>
        `System.Environment.SetEnvironmentVariable("${envVar.Name}", "${envVar.Value}");\n`,
    ),
  );
  return indentLines(lines, 2);
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
  if (getResponseFormat() == "flat") {
    if (fieldDef.Type.Type.toString() == "union") {
      // For top-level flat union response, res IS the union directly - no accessor
      // needed. NormalizeJson handles the comparison correctly.
      if (!context?.topLevel) {
        let type = fieldDef.Type;

        if (example !== undefined) {
          const subType = matchTypeWithExample(fieldDef.Type, example)?.type;

          if (subType) {
            type = subType;
          }
        }

        value += `.${sanitizeFieldName(type.Name)}`;
      }
    } else if (context?.topLevel && responseType?.Type?.toString() == "union") {
      // Flat format, top-level, response is a union but the assertion targets a
      // specific variant type (e.g., asserting a string from a union response).
      // Access the matching variant property on the union.
      const variant = (responseType.Fields || []).find(
        (f: FieldDef) =>
          f.Type.Name === fieldDef.Type.Name ||
          f.Type.Type.toString() === fieldDef.Type.Type.toString(),
      );
      if (variant) {
        value += `.${sanitizeFieldName(variant.Name)}`;
      }
    }
  }

  switch (fieldDef.Type.Type.toString()) {
    case "response-stream":
      return `CommonHelpers.ReadStreamToBytes(${value})`;
    default:
      return value;
  }
}

// @ts-ignore
function templateTestFieldAccessor(
  parent: FieldDef,
  field: FieldDef,
  _additionalContext?: TemplateValueContext,
): string {
  return fieldAccessor(field, parent);
}

// @ts-ignore
function templateTestIndexAccessor(parent: FieldDef, idx: string): string {
  return `${nullCond(parent)}[${idx}]`;
}

// @ts-ignore
function templateTestKeyAccessor(parent: FieldDef, key: string): string {
  return `${nullCond(parent)}["${key}"]`;
}

// @ts-ignore
function templateTestAdditionalPropertiesAccessor(
  parent: FieldDef,
  additionalPropertiesField: FieldDef,
  fieldName: string,
): string {
  // `?.` based on receiver (parent) nullability;
  // `?[` based on the additionalProperties field's nullability since the
  // indexer applies to the result of `.Field`.
  const dotCond = nullCond(parent);
  const idxCond = nullCond(additionalPropertiesField);
  return `${dotCond}.${sanitizeFieldName(
    additionalPropertiesField.Name,
  )}${idxCond}["${fieldName}"]`;
}

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  return sanitizePrivateFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}res`,
  );
}
registerTemplateFunc("getResponseVariableName", getResponseVariableName);

// @ts-ignore
function getRequestVariableName(stepID: string): string {
  return sanitizePrivateFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}req`,
  );
}
registerTemplateFunc("getRequestVariableName", getRequestVariableName);

// @ts-ignore
function getOutputVariableName(stepID: string, accessor?: boolean): string {
  return sanitizePrivateFieldName(
    `${stepID && stepID !== "test" ? stepID + "_" : ""}outputs`,
  );
}
registerTemplateFunc("getOutputVariableName", getOutputVariableName);

function sanitizeHelperMethodName(name: string): string {
  return caser().ToPascal(sanitizeName(name));
}
registerTemplateFunc("sanitizeHelperMethodName", sanitizeHelperMethodName);

// @ts-ignore
function templateWorkflow(
  parent: ArazzoWorkflow,
  workflowStep: ArazzoWorkflowStep,
): string {
  const workflow = resolveWorkflowStepTarget(workflowStep);
  if (!workflow) {
    return "";
  }

  let postCallAssertions = "";
  let preCallAssertions = "";
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

    // templateValue generates "new ClassName() { ... }" - we need to qualify with TestHelpers namespace
    const inputValue = templateValue(
      workflow.Inputs,
      inputExample,
      false,
      context,
    );
    // "new ClassName(...)" -> "new TestHelpers.ClassName(...)"
    inputs = inputValue.replace(/^new /, "new TestHelpers.");

    const refPaths = new Set<string>();
    for (const field of workflow.Inputs.Type.Fields) {
      collectNullableRefPaths(
        inputExample[field.OriginalName],
        field,
        refPaths,
      );
    }
    if (refPaths.size > 0) {
      preCallAssertions =
        Array.from(refPaths)
          .map((p) => `Assert.NotNull(${p});`)
          .join("\n") + "\n";
    }
  }

  if (workflow.Outputs) {
    const outputVariableName = getOutputVariableName(workflowStep.StepID);
    const className = `TestHelpers.${sanitizeClassName(
      workflow.Outputs.Type.Name,
    )}`;
    outputVariable = `${className}? ${outputVariableName} = `;
    postCallAssertions = `\nAssert.NotNull(${outputVariableName});`;
  }

  return `\n${preCallAssertions}${outputVariable}await TestHelpers.${sanitizeHelperMethodName(
    workflow.Name,
  )}(${inputs});${postCallAssertions}`;
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

  const usageContext = getWorkflowOutputsUsageContext(workflow);
  const paths = new Set<string>();
  if (usageContext) {
    const ctx = { usageContext, test: workflow };
    for (const field of workflow.Outputs.Type.Fields) {
      const fieldExample = findExampleByName(
        field.Type.Examples ?? [],
        workflow.Name,
      );
      if (fieldExample) {
        collectNullableRefPaths(
          getExampleValue(fieldExample, ctx),
          field,
          paths,
        );
      }
    }
  }

  const asserts = Array.from(paths)
    .map((p) => `Assert.NotNull(${p});`)
    .join("\n");
  const prefix = asserts ? `\n${asserts}` : "";

  return `${prefix}\nreturn ${value};`;
}
registerTemplateFunc("templateHelperOutputs", templateHelperOutputs);

// @ts-ignore
function joinAssertions(lines: string[]): string {
  return [...lines].join("\n");
}

/**
 * Generates import statements for test files without relying on recurse.
 * Collects all model namespaces used by operations in the test group.
 */
function genTestFileImports(testGroup: any): string {
  const imports = new Set<string>();
  const aliases = new Set<string>();

  imports.add(getSDKNamespace());

  // Add all known model scopes
  const scopes = ["models", "shared", "operations", "errors"];
  for (const scope of scopes) {
    try {
      const fqn = getScopeNamespace(scope, true);
      if (fqn) {
        imports.add(fqn);
        if (scope === "models" && !usingGlobalImports()) {
          aliases.add(`using Models = ${fqn};`);
        }
      }
    } catch (e) {
      // scope might not exist
    }
  }

  // Add utils namespace
  try {
    const utilsNs = getScopeNamespace("utils", true);
    if (utilsNs) {
      imports.add(utilsNs);
    }
  } catch (e) {
    // ignore
  }

  // Add NodaTime import if the SDK uses NodaTime for date types
  if (useNodatime()) {
    imports.add("NodaTime");
  }

  // Collect actual output locations from BucketedTypes so we only import
  // namespaces that have generated types (e.g. skip Webhooks if no webhooks exist)
  const actualOutputLocations = new Set<string>();
  if (context.Global.AST?.BucketedTypes) {
    for (const [outputLocation] of sequencedMapEntries(
      context.Global.AST.BucketedTypes,
    )) {
      if (outputLocation) {
        try {
          const fqn = getModelNamespace(outputLocation, true);
          if (fqn) {
            imports.add(fqn);
            actualOutputLocations.add(fqn);
          }
        } catch (e) {
          // ignore
        }
      }
    }
  }

  // Add configured import paths only if they have actual generated types
  if (context.Global.Config.Imports?.Paths) {
    for (const scope in context.Global.Config.Imports.Paths) {
      const path = context.Global.Config.Imports.Paths[scope];
      if (path) {
        const ns = `${getSDKNamespace()}.${path.replace(/\//g, ".")}`;
        if (actualOutputLocations.has(ns)) {
          imports.add(ns);
        }
      }
    }
  }

  const importLines = Array.from(imports)
    .sort()
    .map((i) => `using ${i};`);

  return [...importLines, ...Array.from(aliases).sort()].join("\n");
}
registerTemplateFunc("genTestFileImports", genTestFileImports);

// @ts-ignore
function languageSpecificEnvVarWrapping(
  envVar: EnvVar,
  additionalContext?: TemplateValueContext,
  fieldDef?: FieldDef,
): string {
  const isArray = fieldDef?.Type?.Type?.toString() === "array";
  const envVarCall = `Environment.GetEnvironmentVariable("${
    envVar.name
  }") ?? "${envVar.defaultValue || ""}"`;

  if (isArray) {
    return `new List<string> { ${envVarCall} }`;
  }
  return envVarCall;
}
