// =============================================================================
// CLI Test Assertions & Type Coercion
// Implements the common test framework abstract methods for JSON-based CLI output,
// plus type coercion to match Go's serialization behavior.
// =============================================================================

// ============================================================================
// Common Test Framework Abstract Method Implementations
// These implement the methods defined in common/common/tests.ts to generate
// JSON-based assertions for CLI output.
// ============================================================================

// @ts-ignore
function getResponseVariableName(stepID: string): string {
  return "output";
}
registerTemplateFunc("getResponseVariableName", getResponseVariableName);

// @ts-ignore
function getOutputVariableName(stepID: string): string {
  return "output";
}
registerTemplateFunc("getOutputVariableName", getOutputVariableName);

// @ts-ignore
// CLI extracts result content and outputs it directly (stripping the response
// field wrapper). Raw JSON passthrough also outputs the API body directly.
// Override the core framework's path generation to NOT add the content field
// prefix (e.g., ".User"), so assertion paths are relative to the extracted content.
function getResponseContentVariablePath(
  usageContext: UsageContext,
  content: ResponseBodyContent,
): string {
  return getResponseVariableName(usageContext.StepID);
}

// @ts-ignore
function addRequiredTestImports() {
  addImport("github.com/stretchr/testify/assert");
}

// Returns the JSON field accessor for navigating CLI JSON output.
// For model fields (with json tags), uses OriginalName (e.g., "email", "first_name").
// For envelope fields (no json tags), uses the Go field name (e.g., "User").
// @ts-ignore
function templateTestFieldAccessor(
  parent: FieldDef,
  field: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  // Use originalFieldName which returns OriginalName || Name
  // For model fields: OriginalName is the json tag name (snake_case)
  // For envelope fields: Name is already the Go field name (PascalCase)
  return `.${originalFieldName(field)}`;
}

// Array index accessor for JSON paths: ".0", ".1", etc.
// @ts-ignore
function templateTestIndexAccessor(parent: FieldDef, idx: string): string {
  return `.${idx}`;
}

// Map key accessor for JSON paths: ".keyName"
// @ts-ignore
function templateTestKeyAccessor(parent: FieldDef, key: string): string {
  return `.${key}`;
}

// Additional properties accessor for JSON output.
// In JSON serialization, additional properties are spread into the parent object
// (the field has json:"-" and additionalProperties:"true"), so we access them
// directly by key name without the intermediate field name.
// @ts-ignore
function templateTestAdditionalPropertiesAccessor(
  parent: FieldDef,
  additionalPropertiesField: FieldDef,
  fieldName: string,
): string {
  return `.${fieldName}`;
}

// No pointer unwrapping needed for JSON navigation
// @ts-ignore
function templateTestFieldAccess(
  parent: FieldDef,
  currentPath: string,
): string[] {
  return [];
}

// Build field path for class fields using JSON key names
// @ts-ignore
function getTestResponseClassFieldPath(
  path: string,
  parentField: FieldDef,
  classField: FieldDef,
): string {
  return `${path}.${originalFieldName(classField)}`;
}

// Status codes are checked implicitly via require.NoError (SDK returns errors for non-success)
// @ts-ignore
function templateStatusCodeAssertion(
  usageContext: UsageContext,
  assertion: Assertion,
): string[] {
  return [];
}

// @ts-ignore
function addDefinedAssertion(lines: string[], path: string): string[] {
  const jsonPath = stripOutputPrefix(path);
  if (jsonPath === "") {
    lines.push(`assert.NotEmpty(t, output)`);
  } else {
    lines.push(
      `assert.True(t, jsonExists(output, "${escapeGoString(jsonPath)}"))`,
    );
  }
  return lines;
}

// @ts-ignore
function addNotNullAssertion(lines: string[], path: string): string[] {
  const jsonPath = stripOutputPrefix(path);
  if (jsonPath === "") {
    lines.push(`assert.NotEmpty(t, output)`);
  } else {
    lines.push(
      `assert.NotNil(t, jsonGet(output, "${escapeGoString(jsonPath)}"))`,
    );
  }
  return lines;
}

// @ts-ignore
function addNotEmptyAssertion(lines: string[], path: string): string[] {
  const jsonPath = stripOutputPrefix(path);
  if (jsonPath === "") {
    lines.push(`assert.NotEmpty(t, output)`);
  } else {
    lines.push(
      `assert.NotEmpty(t, jsonGetString(output, "${escapeGoString(
        jsonPath,
      )}"))`,
    );
  }
  return lines;
}

// @ts-ignore
function addNullableOptionalAssertions(
  lines: string[],
  path: string,
  field: FieldDef,
  example: any,
): string[] {
  const jsonPath = stripOutputPrefix(path);
  switch (true) {
    case field.Optional && example === undefined:
    case field.Nullable && example === null:
      lines.push(
        `assert.Nil(t, jsonGet(output, "${escapeGoString(jsonPath)}"))`,
      );
      break;
  }
  return lines;
}

// No mutation needed for CLI JSON assertions
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

// No directive handling needed for CLI JSON assertions
// @ts-ignore
function sanitizeValuesWithDirectives(
  extensions: Record<string, any>,
  value: string,
): string {
  return value;
}

// =============================================================================
// Type Coercion
// =============================================================================

// Recursively walk an untyped (any) value and coerce numbers that exceed
// float32 precision. The mock server may round-trip these through typed Go
// structs with float32 fields, producing fewer significant digits.
function coerceAnyValueForFloat32(value: any): any {
  if (value === null || value === undefined) return value;
  if (typeof value === "number") {
    const f32 = Math.fround(value);
    if (f32 !== value) {
      return parseFloat(f32.toPrecision(7));
    }
    return value;
  }
  if (Array.isArray(value)) {
    return value.map((v: any) => coerceAnyValueForFloat32(v));
  }
  if (typeof value === "object") {
    const result: Record<string, any> = {};
    for (const [k, v] of Object.entries(value)) {
      result[k] = coerceAnyValueForFloat32(v);
    }
    return result;
  }
  return value;
}

// Coerce example values to match Go's JSON serialization behavior.
// For example, Go's `*string` serializes 94110 as "94110", and
// `map[string]string` serializes {height: 182} as {"height":"182"}.
function coerceExampleToGoTypes(example: any, typeDef: TypeDef): any {
  if (example === null || example === undefined || !typeDef) return example;

  const typeStr = typeDef.Type?.toString() || "";

  switch (typeStr) {
    case "date-time":
    case "date": {
      // Go's time.Time.MarshalJSON() uses time.RFC3339Nano format which:
      // - Normalizes "+00:00" offset to "Z"
      // - Strips trailing zeros from fractional seconds (".850" → ".85")
      // - Removes fractional part entirely if all zeros (".000" → "")
      if (typeof example === "string") {
        return normalizeGoDateTime(example);
      }
      return example;
    }
    case "string":
      // Go serializes string types as JSON strings
      if (typeof example === "number") return String(example);
      if (typeof example === "boolean") return String(example);
      return example;
    case "boolean":
      // Coerce string booleans from form examples to proper booleans
      if (typeof example === "string") {
        if (example === "true") return true;
        if (example === "false") return false;
      }
      return example;
    case "integer":
    case "int32":
    case "int64":
      // Coerce string numbers from form examples to proper integers
      if (typeof example === "string") {
        const n = parseInt(example, 10);
        if (!isNaN(n)) return n;
      }
      return example;
    case "number":
      // Coerce string numbers from form examples to proper floats
      if (typeof example === "string") {
        const n = parseFloat(example);
        if (!isNaN(n)) return n;
      }
      return example;
    case "float32":
      // Round to float32 precision to match Go's JSON serialization.
      // Go uses strconv.FormatFloat(float64(f32), 'f', -1, 32) which finds the
      // shortest decimal representation. toPrecision(7) approximates this since
      // float32 has ~7.2 significant decimal digits.
      if (typeof example === "number") {
        return parseFloat(Math.fround(example).toPrecision(7));
      }
      return example;
    case "any":
      // For any-typed fields, the inner structure is unknown at template time.
      // The mock server may round-trip through typed Go structs, so numbers
      // that exceed float32 precision need coercion. Recursively walk the value.
      return coerceAnyValueForFloat32(example);
    case "union": {
      // For union types, find the best-matching variant and coerce through it.
      // The mock server deserializes union values into typed Go structs,
      // so float32 truncation etc. occurs based on the matched variant.
      // Discriminated unions use Discriminator.Mapping; non-discriminated use AssociatedTypes.
      const variantTypes: TypeDef[] = [];
      if (typeDef.Discriminator?.Mapping?.length > 0) {
        for (const m of typeDef.Discriminator.Mapping) {
          if (m.Type) variantTypes.push(m.Type);
        }
      } else if (typeDef.AssociatedTypes?.length > 0) {
        for (const t of typeDef.AssociatedTypes) {
          variantTypes.push(t);
        }
      }
      if (variantTypes.length > 0) {
        const exType = typeof example;
        for (const vt of variantTypes) {
          const variantTypeStr = vt.Type?.toString() || "";
          if (
            exType === "object" &&
            !Array.isArray(example) &&
            variantTypeStr === "class"
          ) {
            return coerceExampleToGoTypes(example, vt);
          }
          if (exType === "string" && variantTypeStr === "string") {
            return coerceExampleToGoTypes(example, vt);
          }
          if (
            exType === "number" &&
            (variantTypeStr === "number" ||
              variantTypeStr === "float32" ||
              variantTypeStr === "integer" ||
              variantTypeStr === "int32")
          ) {
            return coerceExampleToGoTypes(example, vt);
          }
        }
      }
      return example;
    }
    case "class": {
      if (typeof example !== "object" || Array.isArray(example)) return example;
      const result = { ...example };
      for (const field of typeDef.Fields || []) {
        const key = originalFieldName(field);
        if (key in result) {
          result[key] = coerceExampleToGoTypes(result[key], field.Type);
        }
      }
      // Handle AdditionalProperties fields (json:"-" spread into parent)
      const apField = (typeDef.Fields || []).find(
        (f) => f.IsAdditionalProperties,
      );
      const declaredKeys = new Set(
        (typeDef.Fields || []).map((f) => originalFieldName(f)),
      );
      if (apField) {
        // Keys not matching any declared field belong to AdditionalProperties
        for (const key of Object.keys(result)) {
          if (!declaredKeys.has(key)) {
            result[key] = coerceExampleToGoTypes(
              result[key],
              apField.Type?.ItemType,
            );
          }
        }
      } else {
        // Handle unmatched keys (e.g., mock server wrappers like {"json": <body>}).
        // Recurse with the same type so inner values get type-driven coercion.
        for (const key of Object.keys(result)) {
          if (
            !declaredKeys.has(key) &&
            typeof result[key] === "object" &&
            result[key] !== null
          ) {
            result[key] = coerceExampleToGoTypes(result[key], typeDef);
          }
        }
      }
      return result;
    }
    case "array":
    case "set":
      if (!Array.isArray(example)) return example;
      if (!typeDef.ItemType) return example;
      return example.map((item: any) =>
        coerceExampleToGoTypes(item, typeDef.ItemType),
      );
    case "map":
      if (typeof example !== "object" || Array.isArray(example)) return example;
      if (!typeDef.ItemType) return example;
      const mapResult: Record<string, any> = {};
      for (const [k, v] of Object.entries(example)) {
        mapResult[k] = coerceExampleToGoTypes(v, typeDef.ItemType);
      }
      return mapResult;
    default:
      return example;
  }
}

// Get the result type from a response envelope type.
// Mirrors the runtime logic in output.extractResultContent().
function getCLIResultType(responseType: TypeDef): TypeDef | null {
  if (!responseType || !responseType.Fields) return null;

  const resultField = getResultField({
    Response: { Type: responseType },
  } as any);
  if (resultField) {
    return resultField.Type;
  }

  // Fallback: find first non-envelope field
  for (const field of responseType.Fields) {
    if (RESPONSE_ENVELOPE_FIELDS_NO_HEADERS.has(field.Name)) continue;
    return field.Type;
  }
  return null;
}

// =============================================================================
// Core Assertion Generator
// =============================================================================

// Core assertion generator for CLI JSON output.
// Generates type-appropriate assertions comparing expected values
// against values extracted from the JSON output string.
// @ts-ignore
function templateAssertion(
  lines: string[],
  assertionValue: string,
  field: FieldDef,
  example: any,
  usageContext: UsageContext,
): string[] {
  const jsonPath = stripOutputPrefix(assertionValue);

  if (example === undefined) {
    return lines;
  }

  // For unresolved placeholder examples in response assertions, use relaxed assertions.
  // Placeholders like "oneOf[1]", "anyOf[0]", "..." come from the spec but the mock server
  // returns resolved values that may differ from what calculateExample generates.
  if (isExamplePlaceholder(example)) {
    if (jsonPath === "") {
      lines.push(`assert.NotEmpty(t, output)`);
      return lines;
    }
    lines.push(
      `assert.NotNil(t, jsonGet(output, "${escapeGoString(jsonPath)}"))`,
    );
    return lines;
  }

  // For objects containing placeholder values anywhere in nested fields, filter them out
  // and use jsonContains (subset assertion) so resolved fields don't break matching
  if (typeof example === "object" && example !== null) {
    if (hasNestedPlaceholders(example)) {
      const filtered = filterPlaceholders(example);
      if (
        filtered !== undefined &&
        (typeof filtered !== "object" || Object.keys(filtered).length > 0)
      ) {
        if (jsonPath === "") {
          const jsonStr = JSON.stringify(filtered);
          lines.push(
            `jsonContains(t, ${goJSONLiteral(
              jsonStr,
            )}, strings.TrimSpace(output))`,
          );
          addImport("strings");
        } else {
          const jsonStr = JSON.stringify(filtered);
          lines.push(
            `jsonContains(t, ${goJSONLiteral(
              jsonStr,
            )}, toJSON(jsonGet(output, "${escapeGoString(jsonPath)}")))`,
          );
        }
      } else {
        lines.push(`assert.NotEmpty(t, output)`);
      }
      return lines;
    }
  }

  // Check if this is a non-JSON response (text, binary, XML, HTML)
  const isNonJSON = hasNonJSONResponse(usageContext.Operation);

  // When jsonPath is empty, assertions are on the entire response output
  if (jsonPath === "") {
    if (example === null) {
      lines.push(`assert.Empty(t, output)`);
    } else if (isNonJSON && typeof example === "string") {
      // Non-JSON response in RunRaw mode: compare raw text output directly
      lines.push(
        `assert.Equal(t, ${goStringLiteral(
          example,
        )}, strings.TrimSpace(output))`,
      );
      addImport("strings");
    } else if (typeof example === "object") {
      // Coerce example values to match the raw API response format.
      // Use the field's type directly (e.g., User type) since the CLI
      // strips the response envelope in output.
      let coercedExample = example;
      if (field?.Type) {
        coercedExample = coerceExampleToGoTypes(example, field.Type);
        // Filter null values for optional fields — mock server uses omitempty
        coercedExample = filterOmitemptyNulls(coercedExample, field.Type);
      }

      // For split operations or multiple content types, filter out null variant
      // fields from the expected data (nil maps/pointers are omitted by omitzero)
      const op = usageContext.Operation;
      const isSplitOp = op.OriginalID && op.OriginalID !== op.ID;
      const hasMultiContent = hasMultipleSuccessContentTypes(op);
      if (isSplitOp || hasMultiContent) {
        coercedExample = filterEmptyResponseFields(coercedExample);
      }

      // Check for AdditionalProperties — use subset assertion (check nested types too)
      const hasAdditionalProps = hasOperationAdditionalProps(op);

      const jsonStr = JSON.stringify(coercedExample);
      if (hasAdditionalProps) {
        lines.push(
          `jsonContains(t, ${goJSONLiteral(
            jsonStr,
          )}, strings.TrimSpace(output))`,
        );
      } else {
        lines.push(
          `assert.JSONEq(t, ${goJSONLiteral(
            jsonStr,
          )}, strings.TrimSpace(output))`,
        );
      }
      addImport("strings");
    } else if (typeof example === "string") {
      const fieldTypeStr = field?.Type?.Type?.toString() || "";
      if (
        fieldTypeStr === "string" ||
        fieldTypeStr === "decimal" ||
        fieldTypeStr === "bigint"
      ) {
        // CLI outputs raw JSON, so string/decimal/bigint responses are JSON-encoded: "test" -> "\"test\""
        const jsonStr = JSON.stringify(example);
        lines.push(
          `assert.JSONEq(t, ${goJSONLiteral(
            jsonStr,
          )}, strings.TrimSpace(output))`,
        );
      } else {
        // Binary/bytes or other non-JSON response — compare raw output
        lines.push(
          `assert.Equal(t, ${goStringLiteral(
            example,
          )}, strings.TrimSpace(output))`,
        );
      }
      addImport("strings");
    }
    return lines;
  }

  if (example === null) {
    lines.push(`assert.Nil(t, jsonGet(output, "${escapeGoString(jsonPath)}"))`);
    return lines;
  }

  // Use the field's Go type to determine the correct assertion method,
  // not typeof example (which reflects the Arazzo YAML type, not the Go type).
  const fieldType = field.Type?.Type?.toString() || "";

  switch (fieldType) {
    case "string": {
      const strValue = typeof example === "string" ? example : String(example);
      lines.push(
        `assert.Equal(t, ${goStringLiteral(
          strValue,
        )}, jsonGetString(output, "${escapeGoString(jsonPath)}"))`,
      );
      break;
    }
    case "integer":
    case "int32":
    case "number":
    case "float32": {
      const numValue =
        typeof example === "number" ? example : parseFloat(String(example));
      if (!isNaN(numValue)) {
        lines.push(
          `assert.Equal(t, float64(${numValue}), jsonGetFloat64(output, "${escapeGoString(
            jsonPath,
          )}"))`,
        );
      }
      break;
    }
    case "boolean":
      lines.push(
        `assert.Equal(t, ${example}, jsonGetBool(output, "${escapeGoString(
          jsonPath,
        )}"))`,
      );
      break;
    default:
      // Fall back to typeof-based assertion for unknown/complex types
      switch (typeof example) {
        case "string":
          lines.push(
            `assert.Equal(t, ${goStringLiteral(
              example,
            )}, jsonGetString(output, "${escapeGoString(jsonPath)}"))`,
          );
          break;
        case "number":
          lines.push(
            `assert.Equal(t, float64(${example}), jsonGetFloat64(output, "${escapeGoString(
              jsonPath,
            )}"))`,
          );
          break;
        case "boolean":
          lines.push(
            `assert.Equal(t, ${example}, jsonGetBool(output, "${escapeGoString(
              jsonPath,
            )}"))`,
          );
          break;
        case "object": {
          const filtered = filterOmitemptyNulls(example, field?.Type);
          const jsonStr = JSON.stringify(filtered);
          lines.push(
            `assert.JSONEq(t, ${goJSONLiteral(
              jsonStr,
            )}, toJSON(jsonGet(output, "${escapeGoString(jsonPath)}")))`,
          );
          break;
        }
      }
  }

  return lines;
}

// =============================================================================
// Assertion Helpers
// =============================================================================

// Helper: strip the "output" prefix from assertion paths to get the JSON navigation path.
// "output.User.email" -> "User.email"
// "output" -> "" (empty = entire response)
function stripOutputPrefix(path: string): string {
  if (path === "output") {
    return "";
  }
  if (path.startsWith("output.")) {
    return path.substring("output.".length);
  }
  return path;
}

// Like goStringLiteral but for JSON strings that may be long.
// Uses backtick raw strings when safe (no backticks in content and no {{/}}),
// falls back to double-quoted strings with escaping otherwise.
function goJSONLiteral(jsonStr: string): string {
  if (
    !jsonStr.includes("`") &&
    !jsonStr.includes("{{") &&
    !jsonStr.includes("}}")
  ) {
    return "`" + jsonStr + "`";
  }
  // Fall back to double-quoted string with escaping
  return goStringLiteral(jsonStr);
}

// Normalize a date-time string to match Go's time.Time JSON serialization.
// Go's time.RFC3339Nano:
//   - Converts "+00:00" to "Z"
//   - Uses minimal fractional seconds (strips trailing zeros)
//   - Removes fractional part entirely if all zeros
function normalizeGoDateTime(dateStr: string): string {
  // Replace "+00:00" offset with "Z"
  let normalized = dateStr.replace(/\+00:00$/, "Z");
  // Strip trailing zeros from fractional seconds: .850Z → .85Z, .500Z → .5Z
  normalized = normalized.replace(/(\.\d*?)0+(Z|[+-])/, "$1$2");
  // If all fractional digits were zeros, the dot remains: ".Z" → "Z"
  normalized = normalized.replace(/\.(Z|[+-])/, "$1");
  return normalized;
}

// Coerce a request body example to match Go's type expectations.
// E.g., postal_code: 94110 -> "94110" when Go expects *string.
// Handles both object values and JSON string values.
function coerceRequestBodyExample(
  exampleValue: any,
  operation: Operation,
): any {
  const bodyType = operation.Request?.RequestBody?.Type;
  if (!bodyType) return exampleValue;

  if (typeof exampleValue === "string") {
    // Try to parse JSON string, coerce, and re-stringify
    try {
      const parsed = JSON.parse(exampleValue);
      if (typeof parsed === "object" && parsed !== null) {
        return JSON.stringify(coerceExampleToGoTypes(parsed, bodyType));
      }
    } catch (_e) {
      // Not valid JSON, return as-is
    }
    return exampleValue;
  }

  if (typeof exampleValue === "object" && exampleValue !== null) {
    return coerceExampleToGoTypes(exampleValue, bodyType);
  }

  return exampleValue;
}
