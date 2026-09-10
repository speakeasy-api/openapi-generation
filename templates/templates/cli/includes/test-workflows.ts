// =============================================================================
// Multi-Step Test & Workflow Support
// Handles multi-step test generation, cross-step references, workflow calls,
// and helper output generation for CLI tests.
// =============================================================================

// Tests that traverse an overridden generated operation are skipped as a
// unit. Dedicated intent tests own the replacement surface; replaying the old
// generated flags through that surface could silently change request meaning.
function cliWorkflowUsesOverriddenOperation(
  workflow: ArazzoWorkflow,
  seen: Set<any> = new Set<any>(),
): boolean {
  if (!workflow || seen.has(workflow)) return false;
  seen.add(workflow);
  for (const step of workflow.Steps || []) {
    if (
      step.Type === "operation" &&
      step.UsageContext?.Operation &&
      isOperationOverridden(step.UsageContext.Operation)
    ) {
      return true;
    }
    if (step.Type === "workflow") {
      const nested = resolveWorkflowStepTarget(step);
      if (nested && cliWorkflowUsesOverriddenOperation(nested, seen)) {
        return true;
      }
    }
  }
  return false;
}

// Check if a test has multiple steps (needs step-specific output tracking).
// @ts-ignore
function isMultiStepTest(test: Test): boolean {
  return (test.Workflow?.Steps?.length ?? 0) > 1;
}
registerTemplateFunc("isMultiStepTest", isMultiStepTest);

// Get a Go variable name for a step's captured output.
// @ts-ignore
function getStepOutputVarName(stepID: string): string {
  return sanitizePrivateFieldName(`${stepID}Output`);
}
registerTemplateFunc("getStepOutputVarName", getStepOutputVarName);

// Resolve an Example reference to a Go expression string for CLI tests.
// Returns undefined if the example has no reference.
function resolveCLIReference(
  example: Example,
  test: ArazzoWorkflow,
): string | undefined {
  if (!example?.Reference) return undefined;

  const ref = example.Reference;
  const steps = test.Steps || [];

  const refType = ref.Type?.toString() || "";

  switch (refType) {
    case "response.body": {
      const target = ref.Target as ResponseBodyTarget;
      const step = steps[target.StepIdx];
      if (!step) return undefined;

      const stepVarName = getStepOutputVarName(step.StepID);

      if (!target.Path || target.Path === "/" || target.Path === "") {
        // Whole body reference
        addImport("strings");
        return `strings.TrimSpace(${stepVarName})`;
      }

      // Field reference - convert JSON pointer path to dot path
      const dotPath = target.Path.replace(/^\//, "").replace(/\//g, ".");
      return `jsonGetString(${stepVarName}, "${dotPath}")`;
    }
    case "inputs": {
      const target = ref.Target as InputTarget;
      const dotPath = (target.Path || "")
        .replace(/^\//, "")
        .replace(/\//g, ".");
      return `inputs["${dotPath}"]`;
    }
    case "outputs": {
      const target = ref.Target as OutputTarget;
      const step = steps[target.StepIdx];
      if (!step) return undefined;

      const stepVarName = getStepOutputVarName(step.StepID);

      // For workflow steps, outputs are in a map[string]string
      // Path like "/user/id" means: get output "user", then navigate JSON path "id"
      if (step.Type === "workflow") {
        const segments = (target.Path || "").replace(/^\//, "").split("/");
        const outputName = segments[0];
        const remainingPath = segments.slice(1).join(".");

        if (remainingPath) {
          return `jsonGetString(${stepVarName}["${outputName}"], "${remainingPath}")`;
        }
        return `${stepVarName}["${outputName}"]`;
      }

      // For operation steps, extract from JSON output
      if (!target.Path || target.Path === "/" || target.Path === "") {
        addImport("strings");
        return `strings.TrimSpace(${stepVarName})`;
      }

      const dotPath = target.Path.replace(/^\//, "").replace(/\//g, ".");
      return `jsonGetString(${stepVarName}, "${dotPath}")`;
    }
  }

  return undefined;
}

// Look up the field type for a given dot-separated path in a body type.
// E.g., "postal_code" in User type returns the FieldDef for PostalCode.
function getBodyFieldByPath(
  bodyType: TypeDef | undefined,
  fieldPath: string,
): FieldDef | undefined {
  if (!bodyType || !fieldPath) return undefined;
  const pathParts = fieldPath.split(".");
  let currentType = bodyType;
  let targetField: FieldDef | undefined;

  for (const part of pathParts) {
    if (!currentType?.Fields) return undefined;
    targetField = currentType.Fields.find(
      (f) => (f.OriginalName || f.Name) === part,
    );
    if (!targetField) return undefined;
    currentType = targetField.Type;
  }
  return targetField;
}

// Format a replacement value for jsonSet, coerced to the target field's Go type.
// E.g., value 94107 for a *string field -> Go string literal "94107"
//        value "33" for a *float64 field -> float64(33)
function formatCoercedReplacementValue(
  rawValue: any,
  fieldPath: string,
  bodyType: TypeDef | undefined,
): string {
  const targetField = getBodyFieldByPath(bodyType, fieldPath);
  if (!targetField) return formatGoMapLiteral(rawValue);

  const typeStr = targetField.Type?.Type?.toString() || "";
  switch (typeStr) {
    case "string": {
      const strVal = typeof rawValue === "string" ? rawValue : String(rawValue);
      return goStringLiteral(strVal);
    }
    case "integer":
    case "int32":
    case "number":
    case "float32": {
      const numVal =
        typeof rawValue === "number" ? rawValue : parseFloat(String(rawValue));
      if (!isNaN(numVal)) return `float64(${numVal})`;
      return formatGoMapLiteral(rawValue);
    }
    case "boolean":
      return rawValue ? "true" : "false";
    default:
      return formatGoMapLiteral(rawValue);
  }
}

// Format a Go literal for use in map[string]interface{} construction.
function formatGoMapLiteral(value: any): string {
  if (value === null || value === undefined) {
    return "nil";
  }
  if (typeof value === "string") {
    return goStringLiteral(value);
  }
  if (typeof value === "number") {
    return `float64(${value})`;
  }
  if (typeof value === "boolean") {
    return value ? "true" : "false";
  }
  // For objects/arrays, marshal them as JSON
  const jsonStr = JSON.stringify(value);
  // Use json.Unmarshal into interface{} for complex values
  return `jsonGet(${goJSONLiteral(jsonStr)}, "")`;
}

// Get the parameter examples for a field based on its param annotation.
function getParamExamples(
  field: FieldDef,
  op: Operation,
  exampleName: string | undefined,
): Example | undefined {
  if (!field.Annotations?.Has("param")) return undefined;
  const paramAnno = field.Annotations.Get("param") as ParamAnnotation;
  let params: ParamDef[] = [];
  switch (paramAnno.ParamType) {
    case "pathParam":
      params = op.Request?.Params?.PathParams || [];
      break;
    case "queryParam":
      params = op.Request?.Params?.QueryParams || [];
      break;
    case "header":
      params = op.Request?.Params?.HeaderParams || [];
      break;
  }
  for (const param of params) {
    if (param.Field.Name === field.Name) {
      return findExampleWithFallback(param.Examples ?? [], exampleName);
    }
  }
  return undefined;
}

// Generate the full step code block for a multi-step CLI test.
// This includes any pre-step variable setup (for body references with
// replacements) and the args assignment line.
// @ts-ignore
function templateCLIStepCode(usageContext: UsageContext): string {
  const prelude: string[] = [];
  const args: string[] = [];
  let stdinBody = ""; // Non-empty when body should be piped via stdin
  const operation = usageContext.Operation;
  const test = usageContext.Test;
  const exampleName = usageContext.ExampleName;

  // Add command path
  const cmdPath = getCLICommandPath(operation);
  for (const part of cmdPath) {
    args.push(`"${part}"`);
  }

  // Add operation-level security flags from Arazzo test examples
  templateSecurityArgs(usageContext, args);

  if (operation.Request) {
    const serMethod = operation.SerializationMethod?.toString() || "";

    // --- Handle request body ---
    if (
      serMethod === "multipart" &&
      operation.Request.IsRequestBody &&
      operation.Request.RequestBody
    ) {
      // Multipart: delegate to existing logic (no references expected)
      const existingArgs = templateCLIArgs(usageContext);
      const runMethod = isOnlyNonJSONResponse(operation)
        ? "h.RunRaw(args)"
        : "h.Run(args)";
      return `args = []string{${existingArgs}}\nerr = ${runMethod}`;
    }

    let bodyHandled = false;
    // Handle body when IsRequestBody=true (body-only) OR when RequestBody exists
    // alongside params (mixed param+body like updateUser with id param + user body)
    if (operation.Request.RequestBody) {
      bodyHandled = true;
      const reqBodyField = operation.Request.RequestBody;
      const flagName = getRequestBodyFlagName(reqBodyField);
      const bodyIsExpandable = shouldExpandNestedField(reqBodyField);
      const example = findExampleWithFallback(
        operation.Request.Examples ?? [],
        exampleName,
      );

      if (example?.Reference) {
        const ref = resolveCLIReference(example, test);
        if (ref) {
          if (example.Replacements && example.Replacements.length > 0) {
            // Body reference with replacements - generate pre-step code
            const bodyVarName = `${sanitizePrivateFieldName(
              usageContext.StepID || "step",
            )}Body`;
            prelude.push(`${bodyVarName} := ${ref}`);
            for (const repl of example.Replacements) {
              const fieldPath = (repl.Path || "")
                .replace(/^\//, "")
                .replace(/\//g, ".");
              if (repl.Value?.Reference) {
                const replRef = resolveCLIReference(repl.Value, test);
                if (replRef) {
                  // Reference replacement: use jsonGet to preserve type
                  const refForGet = replRef.replace(
                    /^jsonGetString\(/,
                    "jsonGet(",
                  );
                  prelude.push(
                    `${bodyVarName} = jsonSet(${bodyVarName}, "${fieldPath}", ${refForGet})`,
                  );
                }
              } else {
                // Literal replacement value - coerce to match Go model type
                const replValue = repl.Value;
                let rawValue: any;
                if (replValue && "ToJSON" in replValue) {
                  // ToJSON returns a JSON string representation; parse to get typed value
                  try {
                    rawValue = JSON.parse(replValue.ToJSON());
                  } catch (_e) {
                    rawValue = replValue.ToJSON();
                  }
                } else {
                  rawValue = replValue;
                }
                const bodyType = operation.Request?.RequestBody?.Type;
                const goValue = formatCoercedReplacementValue(
                  rawValue,
                  fieldPath,
                  bodyType,
                );
                prelude.push(
                  `${bodyVarName} = jsonSet(${bodyVarName}, "${fieldPath}", ${goValue})`,
                );
              }
            }
            if (bodyIsExpandable) {
              stdinBody = bodyVarName;
            } else {
              pushFlagArg(args, flagName, bodyVarName);
            }
          } else {
            // Simple body reference without replacements
            if (bodyIsExpandable) {
              stdinBody = ref;
            } else {
              pushFlagArg(args, flagName, ref);
            }
          }
        }
      } else if (example !== undefined) {
        // Non-reference body - use existing literal handling
        let exampleValue = getSimpleExampleValue(example);
        if (exampleValue !== undefined) {
          // Coerce values to match Go types (e.g., number -> string for *string fields)
          exampleValue = coerceRequestBodyExample(exampleValue, operation);
          if (bodyIsExpandable) {
            stdinBody = goStringLiteral(JSON.stringify(exampleValue));
          } else {
            args.push(
              `"--${flagName}"`,
              goStringLiteral(JSON.stringify(exampleValue)),
            );
          }
        }
      }
    }

    // --- Handle parameters (including flattened body fields) ---
    if (operation.Request.Field) {
      const reqFields = operation.Request.Field.Type?.Fields || [];

      // Get request body example payload for flattened body fields
      const bodyExamplePayload = extractBodyExamplePayload(
        operation,
        exampleName,
      );

      for (const field of reqFields) {
        if (field.Const) continue;

        const isParam = field.Annotations?.Has("param");

        // When body was already handled via IsRequestBody, only process actual params
        if (bodyHandled && !isParam) continue;

        // Check for reference on parameter fields first
        if (isParam) {
          const paramExample = getParamExamples(field, operation, exampleName);
          if (paramExample?.Reference) {
            const ref = resolveCLIReference(paramExample, test);
            if (ref) {
              const flagName = sanitizeFlagNameWithReserved(field.Name);
              pushFlagArg(args, flagName, ref);
              continue;
            }
          }
        }

        // Existing logic for non-reference fields
        const isClassType =
          field.Type?.Type?.toString() === "class" &&
          (field.Type?.Fields?.length ?? 0) > 0;
        const isMultipart =
          isClassType && getInputClassType(field) === "MultipartRequestBody";

        if (
          !isParam &&
          isClassType &&
          (shouldExpandNestedField(field) || isMultipart) &&
          bodyExamplePayload
        ) {
          const bodyPrefix = getBodyFlagPrefix(field, reqFields, "");
          for (const subField of field.Type.Fields) {
            if (subField.Const) continue;
            const subFlagName = bodyPrefix
              ? `${bodyPrefix}.${sanitizeFlagNameWithReserved(subField.Name)}`
              : sanitizeFlagNameWithReserved(subField.Name);
            const fieldKey = subField.OriginalName || subField.Name;
            const subValue = bodyExamplePayload[fieldKey];
            if (subValue === undefined) continue;
            if (
              typeof subValue === "string" &&
              subValue.startsWith("x-file: ")
            ) {
              const fileName = subValue.substring("x-file: ".length).trim();
              args.push(
                `"--${subFlagName}"`,
                `"../.speakeasy/testfiles/${fileName}"`,
              );
            } else if (fieldNeedsJSONFormat(subField)) {
              args.push(
                `"--${subFlagName}"`,
                formatCLIArgValueAsJSON(subValue),
              );
            } else {
              pushFlagArg(args, subFlagName, formatCLIArgValue(subValue));
            }
          }
          continue;
        }

        if (
          !isParam &&
          isClassType &&
          !shouldExpandNestedField(field) &&
          !isMultipart &&
          bodyExamplePayload
        ) {
          const flagName = sanitizeFlagNameWithReserved(field.Name);
          args.push(
            `"--${flagName}"`,
            goStringLiteral(JSON.stringify(bodyExamplePayload)),
          );
          continue;
        }

        const exampleValue = getCLIFieldExample(usageContext, field);
        if (exampleValue === undefined) continue;

        const flagName = sanitizeFlagNameWithReserved(field.Name);
        pushFlagArg(args, flagName, formatCLIArgValue(exampleValue));
      }
    }
  }

  // Build output: prelude lines + args assignment + run call
  const allLines = [...prelude, `args = []string{${args.join(", ")}}`];
  const isNonJSON = isOnlyNonJSONResponse(operation);
  if (stdinBody) {
    allLines.push(
      isNonJSON
        ? `err = h.RunWithStdinRaw(args, ${stdinBody})`
        : `err = h.RunWithStdin(args, ${stdinBody})`,
    );
  } else {
    allLines.push(isNonJSON ? `err = h.RunRaw(args)` : `err = h.Run(args)`);
  }
  return allLines.join("\n");
}
registerTemplateFunc("templateCLIStepCode", templateCLIStepCode);

// Generate a workflow call for CLI tests (e.g., calling a helper function).
// @ts-ignore
function templateCLIWorkflow(
  parent: ArazzoWorkflow,
  workflowStep: ArazzoWorkflowStep,
): string {
  const workflow = resolveWorkflowStepTarget(workflowStep);
  if (!workflow) {
    return "";
  }

  addImport("github.com/stretchr/testify/assert");
  const funcName = `cli${sanitizeTestName(workflow.Name)}`;
  const stepVarName = getStepOutputVarName(workflowStep.StepID);

  let inputs = "";
  if (workflow.Inputs) {
    // Build inputs map from examples
    const inputEntries: string[] = [];
    const exampleName = `${parent.Name}[${workflowStep.StepIdx}]`;

    for (const field of workflow.Inputs.Type.Fields) {
      const fieldName = field.OriginalName || field.Name;
      const example = findExampleByName(field.Type.Examples ?? [], exampleName);
      if (example?.Reference) {
        const ref = resolveCLIReference(example, parent);
        if (ref) {
          inputEntries.push(`"${fieldName}": ${ref}`);
        }
      } else {
        const value = getSimpleExampleValue(example);
        if (value !== undefined) {
          inputEntries.push(`"${fieldName}": ${formatCLIArgValue(value)}`);
        }
      }
    }
    inputs = `, map[string]string{${inputEntries.join(", ")}}`;
  }

  let result = "";
  if (workflow.Outputs) {
    result += `${stepVarName} := ${funcName}(t${inputs})\n`;
    result += `assert.NotNil(t, ${stepVarName})`;
  } else {
    result += `${funcName}(t${inputs})`;
  }

  return result;
}
registerTemplateFunc("templateCLIWorkflow", templateCLIWorkflow);

// Generate the return statement for a CLI test helper function.
// Returns a map[string]string with the workflow's declared outputs.
// @ts-ignore
function templateCLIHelperOutputs(helper: ArazzoWorkflow): string {
  if (!helper.Outputs || !helper.Outputs.Type?.Fields) {
    return "";
  }

  const entries: string[] = [];
  for (const field of helper.Outputs.Type.Fields) {
    const fieldName = field.OriginalName || field.Name;
    const example = findExampleWithFallback(
      field.Type.Examples ?? [],
      helper.Name,
    );
    if (example?.Reference) {
      const ref = resolveCLIReference(example, helper);
      if (ref) {
        entries.push(`"${fieldName}": ${ref}`);
      }
    }
  }

  if (entries.length === 0) {
    // Outputs declared but no references resolved - return empty map to satisfy return type
    return `\nreturn map[string]string{}`;
  }

  addImport("strings");
  return `\nreturn map[string]string{${entries.join(", ")}}`;
}
registerTemplateFunc("templateCLIHelperOutputs", templateCLIHelperOutputs);
