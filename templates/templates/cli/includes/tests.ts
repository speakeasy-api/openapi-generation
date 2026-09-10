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

// Get the mock server test name for routing
// This is the name used in the x-speakeasy-test-name header
// @ts-ignore
function getMockServerTestName(test: Test): string {
  // The test name is typically in kebab-case from the Arazzo workflow
  return test.Name || "";
}
registerTemplateFunc("getMockServerTestName", getMockServerTestName);

// Check if a CLI test should be skipped due to CLI-specific limitations.
// Returns a skip reason string if the test should be skipped, or empty string if OK.
// Feature-level skipping (pagination, streaming, etc.) is handled by isTestSkipped() in features.ts.
function getCLITestSkipMessage(test: Test): string {
  if (test.Workflow && cliWorkflowUsesOverriddenOperation(test.Workflow)) {
    return "generated operation command is replaced by an x-speakeasy-cli-commands override";
  }
  return "";
}
registerTemplateFunc("getCLITestSkipMessage", getCLITestSkipMessage);

// Check if an operation has a non-JSON response content type (string, raw, XML, HTML, binary).
// These operations should use RunRaw() in tests instead of Run() (which forces JSON output).
// @ts-ignore
function hasNonJSONResponse(op: Operation): boolean {
  if (!op.Response?.Responses) return false;
  for (const subResponse of op.Response.Responses) {
    if (subResponse.Error) continue;
    for (const content of subResponse.Content || []) {
      const respSerMethod = content.SerializationMethod?.toString() || "";
      if (respSerMethod === "string" || respSerMethod === "raw") return true;
    }
  }
  return false;
}
registerTemplateFunc("hasNonJSONResponse", hasNonJSONResponse);

/**
 * Check if a test's first operation step has ONLY non-JSON response content types.
 * Used by test.go.stmpl to decide between h.Run() and h.RunRaw().
 * For mixed content type operations (e.g. JSON + text/plain), we use h.Run()
 * because tests asserting JSON output would break with RunRaw().
 */
function testHasNonJSONResponse(test: Test): boolean {
  const steps = test.Workflow?.Steps;
  if (!steps || steps.length === 0) return false;
  const firstStep = steps[0];
  if (firstStep.Type !== "operation") return false;
  const op = firstStep.UsageContext?.Operation;
  if (!op) return false;
  return isOnlyNonJSONResponse(op);
}
registerTemplateFunc("testHasNonJSONResponse", testHasNonJSONResponse);

// Check if ALL success response content types are non-JSON (raw, string, binary).
// Returns true only when every content type is explicitly non-JSON.
// Unknown/default serialization methods are treated as JSON (safe default).
function isOnlyNonJSONResponse(op: Operation): boolean {
  if (!op.Response?.Responses) return false;
  let hasAny = false;
  for (const subResponse of op.Response.Responses) {
    if (subResponse.Error) continue;
    for (const content of subResponse.Content || []) {
      hasAny = true;
      const sm = content.SerializationMethod?.toString() || "";
      // Only "string" and "raw" are known non-JSON types.
      // Everything else (json, empty/default, unknown) is treated as JSON-compatible.
      if (sm !== "string" && sm !== "raw") return false;
    }
  }
  return hasAny;
}

// Check if an operation has multiple success content types in a single response.
// Used by test-assertions.ts for filtering empty response fields on split-op responses.
function hasMultipleSuccessContentTypes(op: Operation): boolean {
  if (!op.Response?.Responses) return false;
  for (const subResponse of op.Response.Responses) {
    if (subResponse.Error) continue;
    if ((subResponse.Content || []).length > 1) return true;
  }
  return false;
}

// Recursively filter out null and empty arrays [] from an example object.
// Used for split-op and multi-content-type responses where null/nil variant fields
// are omitted by Go's omitzero tag. Empty objects {} are NOT filtered because
// Go's omitzero only drops nil maps, not empty maps (so {} appears in CLI output).
function filterEmptyResponseFields(obj: any): any {
  if (typeof obj !== "object" || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(filterEmptyResponseFields);
  const result: Record<string, any> = {};
  for (const [key, value] of Object.entries(obj)) {
    if (value === null || value === undefined) continue;
    if (Array.isArray(value) && value.length === 0) continue;
    result[key] = filterEmptyResponseFields(value);
  }
  return result;
}

// Filter values that Go's custom MarshalJSON omits from expected response JSON.
// The mock server uses Go's omitempty JSON tag for optional fields and a custom
// isEmpty() check. This drops:
//   - nil pointer values (null)
//   - empty slices/arrays (length 0)
// This function removes those entries from the expected assertion data so CLI
// test assertions match the actual mock server output.
function filterOmitemptyNulls(example: any, typeDef: TypeDef | null): any {
  if (!typeDef || typeof example !== "object" || example === null) {
    return example;
  }

  // Handle arrays: recurse into array elements to filter nested objects
  if (Array.isArray(example)) {
    if (!typeDef.ItemType) return example;
    return example.map((item: any) =>
      filterOmitemptyNulls(item, typeDef.ItemType),
    );
  }

  // Handle union types: find the best-matching variant and filter through it.
  // Union types don't have the variant's fields directly, so we need to resolve.
  const typeStr = typeDef.Type?.toString() || "";
  if (typeStr === "union") {
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
    // Pick the class variant with the most overlapping field names
    const exampleKeys = new Set(Object.keys(example));
    let bestVariant: TypeDef | null = null;
    let bestOverlap = 0;
    for (const vt of variantTypes) {
      if (vt.Type?.toString() !== "class" || !vt.Fields) continue;
      const overlap = vt.Fields.filter((f) =>
        exampleKeys.has(originalFieldName(f)),
      ).length;
      if (overlap > bestOverlap) {
        bestOverlap = overlap;
        bestVariant = vt;
      }
    }
    if (bestVariant) {
      return filterOmitemptyNulls(example, bestVariant);
    }
    return example;
  }

  const fields = typeDef.Fields;
  if (!fields || fields.length === 0) return example;

  const result: Record<string, any> = {};
  for (const [key, value] of Object.entries(example)) {
    const field = fields.find((f) => f.OriginalName === key);

    if (field && field.Optional && !field.Const && !field.Default) {
      // Mock server omits nil optional fields due to omitempty.
      // Fields with const or default tags are always serialized by the
      // custom MarshalJSON regardless of omitempty.
      if (value === null) continue;
      // Mock server's custom MarshalJSON uses isEmpty() which also drops
      // empty slices (length 0) for omitempty fields.
      if (Array.isArray(value) && value.length === 0) continue;
    }

    // Recursively filter nested objects and arrays
    if (field && typeof value === "object" && value !== null) {
      result[key] = filterOmitemptyNulls(value, field.Type);
    } else {
      result[key] = value;
    }
  }
  return result;
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
      envVars.map(
        (envVar) =>
          `os.Setenv("${escapeGoString(envVar.Name)}", "${escapeGoString(
            envVar.Value,
          )}")`,
      ),
      1,
    ) + "\n"
  );
}
registerTemplateFunc("templateTestEnvVars", templateTestEnvVars);

// Resolve the chain of sub-SDKs from the main SDK down to the SDK owning the
// operation (empty for root-level operations). Group commands are named from
// each SDK's source-defined FieldName segment. The command path still comes
// from the SDK tree so parent/child de-stuttering matches runtime registration.
function getOwningSDKChain(operation: Operation): SDK[] {
  const owning = operation.OwningSDK;
  if (!owning || !owning.Group) return [];
  const walk = (sdk: SDK, chain: SDK[]): SDK[] | undefined => {
    for (const child of sdk.SubSDKs || []) {
      const next = [...chain, child];
      if (child.Group === owning.Group) return next;
      const found = walk(child, next);
      if (found) return found;
    }
    return undefined;
  };
  return walk(context.Global.AST.MainSDK, []) || [];
}

// Get the full CLI command path for an operation (e.g., ["tag1", "subgroup", "my-command"])
// Mirrors the runtime registration: group names via getDeStutteredCommandName
// against the parent SDK's group segment, then stutter removal of the operation
// name against the owning SDK's group segment.
// @ts-ignore
function getCLICommandPath(operation: Operation): string[] {
  const path: string[] = [];
  const chain = getOwningSDKChain(operation);

  let parentGroupName = "";
  for (const sdk of chain) {
    const sdkGroupName = getSDKGroupName(sdk);
    path.push(getDeStutteredCommandName(parentGroupName, sdkGroupName));
    parentGroupName = sdkGroupName;
  }

  // Apply stutter removal: for exact matches, the operation IS the group command;
  // for prefix/suffix matches, strip the group prefix/suffix from the operation name.
  // Skip stutter removal entirely when x-speakeasy-name-override is set.
  if (parentGroupName && !hasNameOverride(operation)) {
    const stutterKind = getStutterKind(parentGroupName, operation.GetID());
    if (stutterKind === "exact") {
      // Promoted to parent — no subcommand in the path
      return path;
    }
    if (stutterKind === "prefix" || stutterKind === "suffix") {
      path.push(getDeStutteredCommandName(parentGroupName, operation.GetID()));
      return path;
    }
  }

  // Add the operation command name
  path.push(sanitizeCLICommand(operation.GetID()));

  return path;
}

// Get CLI command name (just the operation, without parent path)
// @ts-ignore
function getCLICommandName(operation: Operation): string {
  return sanitizeCLICommand(operation.GetID());
}
registerTemplateFunc("getCLICommandName", getCLICommandName);

// Push a flag and its formatted (Go-literal) value onto a test args slice.
// Boolean flags take their value inline: pflag parses `--flag true` as
// `--flag` plus a stray positional "true" (and `--flag false` silently sets
// true). Generated commands reject stray positionals, and every flag kind
// accepts `--flag=value`, so literal true/false values are always folded.
function pushFlagArg(
  args: string[],
  flagName: string,
  formatted: string,
): void {
  if (formatted === '"true"' || formatted === '"false"') {
    args.push(`"--${flagName}=${formatted.slice(1, -1)}"`);
    return;
  }
  args.push(`"--${flagName}"`, formatted);
}

// Format a value as a Go string literal for CLI args
// @ts-ignore
function formatCLIArgValue(value: any): string {
  if (value === null || value === undefined) {
    return `""`;
  }
  if (typeof value === "string") {
    return goStringLiteral(value);
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return `"${value}"`;
  }
  // For objects/arrays, convert to JSON string and escape for Go
  return goStringLiteral(JSON.stringify(value));
}

/**
 * Format a value as a CLI argument for FlagKindJSON flags.
 * Always JSON-encodes the value (including strings) so it's valid JSON.
 * e.g., string "any" → "\"any\"" (JSON-encoded), object {a:1} → "{\"a\":1}"
 */
function formatCLIArgValueAsJSON(value: any): string {
  if (value === null || value === undefined) {
    return `"null"`;
  }
  return goStringLiteral(JSON.stringify(value));
}

/**
 * Check if a field type maps to FlagKindJSON in the CLI metadata.
 * These fields expect valid JSON input and need formatCLIArgValueAsJSON.
 */
function fieldNeedsJSONFormat(field: FieldDef): boolean {
  // Nullable+optional fields use OptionalNullable wrapper → FlagKindJSON
  if (field.Nullable && field.Optional) return true;
  const ft = field.Type?.Type?.toString() || "";
  if (
    ft === "any" ||
    ft === "union" ||
    ft === "map" ||
    ft === "bigint" ||
    ft === "decimal"
  ) {
    return true;
  }
  if (ft === "class" && !shouldExpandNestedField(field)) {
    return true;
  }
  if (ft === "array") {
    const itemType = field.Type?.ItemType?.Type?.toString() || "";
    // Only string/enum arrays use FlagKindStringArray — everything else is FlagKindJSON
    if (itemType !== "string" && itemType !== "enum") {
      return true;
    }
  }
  return false;
}

// Helper to find an example, falling back to first available if named one not found
// @ts-ignore
function findExampleWithFallback(
  examples: Example[],
  exampleName: string | undefined,
): Example | undefined {
  if (!examples || examples.length === 0) {
    return undefined;
  }

  // Try to find the named example first
  if (exampleName) {
    const namedExample = findExampleByName(examples, exampleName);
    if (namedExample !== undefined) {
      return namedExample;
    }
  }

  // Fall back to first example if available
  return examples[0];
}

// Get raw example value without resolving cross-step references
// This is a simplified version of getExampleValue for CLI tests
// @ts-ignore
function getSimpleExampleValue(example: Example): any {
  if (!example) {
    return undefined;
  }

  // The example is possible already resolved so just return it
  if (!("ToJSON" in example)) {
    return example;
  }

  // If it's a reference, we can't resolve it in CLI context
  // (would require getResponseVariableName)
  if (example.Reference) {
    return undefined;
  }

  const json = example.ToJSON();
  // ToJSON() returns a JSON *string* representation (e.g., '"test"' for string "test").
  // Parse it to get the typed value so formatCLIArgValue doesn't double-encode.
  try {
    return JSON.parse(json);
  } catch (_e) {
    return json;
  }
}

// Check if a string value is a placeholder that needs resolution
function isExamplePlaceholder(value: any): boolean {
  if (typeof value !== "string") return false;
  if (/(?:anyOf|oneOf)\[\d+\]/.test(value) || value === "...") return true;
  // x-file: directive is a reference to a test file, not a literal string.
  // The mock server reads the file and returns its content, so comparing the
  // directive string against the actual file content would always fail.
  if (value.startsWith("x-file: ")) return true;
  return false;
}

// Recursively check if an example value contains any placeholders at any depth
function hasNestedPlaceholders(value: any): boolean {
  if (isExamplePlaceholder(value)) return true;
  if (typeof value !== "object" || value === null) return false;
  if (Array.isArray(value))
    return value.some((v: any) => hasNestedPlaceholders(v));
  return Object.values(value).some((v: any) => hasNestedPlaceholders(v));
}

// Recursively filter out placeholder values from an example object.
// Removes object keys whose values are placeholders; for arrays, replaces placeholder
// elements with undefined (which JSON.stringify will convert to null).
function filterPlaceholders(value: any): any {
  if (isExamplePlaceholder(value)) return undefined;
  if (typeof value !== "object" || value === null) return value;
  if (Array.isArray(value)) {
    return value.map((v: any) => filterPlaceholders(v));
  }
  const result: Record<string, any> = {};
  for (const [k, v] of Object.entries(value)) {
    if (isExamplePlaceholder(v)) continue; // drop placeholder fields
    const filtered = filterPlaceholders(v);
    if (filtered !== undefined) {
      result[k] = filtered;
    }
  }
  return result;
}

// Recursively check if a type or any of its nested class fields has AdditionalProperties
function hasNestedAdditionalProps(typeDef: TypeDef | undefined): boolean {
  if (!typeDef?.Fields) return false;
  for (const f of typeDef.Fields) {
    if (f.IsAdditionalProperties) return true;
    if (
      f.Type?.Type?.toString() === "class" &&
      hasNestedAdditionalProps(f.Type)
    )
      return true;
  }
  return false;
}

// Check if an operation has AdditionalProperties in request body or response type.
// When the request body has additional props, mock server echoes may include extra fields
// that aren't in the expected data, so we need subset assertions.
function hasOperationAdditionalProps(op: Operation): boolean {
  const resultType = getCLIResultType(op.Response?.Type);
  if (hasNestedAdditionalProps(resultType)) return true;
  // Check request body fields — mock server echoes may include extra fields
  const bodyType = op.Request?.RequestBody?.Type;
  if (hasNestedAdditionalProps(bodyType)) return true;
  return false;
}

// Resolve placeholder example values for body/response fields recursively.
// Handles union placeholders ("oneOf[N]", "anyOf[N]") and auto-generated placeholders ("...")
// by calling calculateExample from common/examples.ts to get concrete JSON values.
// This mirrors what the Go SDK does via templateValue → calculateExample.
function resolveBodyExampleValue(
  bodyFieldDef: FieldDef,
  rawExample: any,
  op: Operation,
): any {
  if (rawExample === undefined || rawExample === null) return rawExample;

  const bodyTypeStr = bodyFieldDef.Type?.Type?.toString() || "";

  // Handle string placeholder for union/any types
  if (
    (bodyTypeStr === "union" || bodyTypeStr === "any") &&
    isExamplePlaceholder(rawExample)
  ) {
    const resolved = calculateExample(
      bodyFieldDef,
      rawExample,
      undefined,
      false,
      { operation: op, templateDefaultValue: true },
    );
    if (resolved !== undefined) return resolved;
    return rawExample;
  }

  // Handle array with placeholder elements
  if (bodyTypeStr === "array" && Array.isArray(rawExample)) {
    const itemType = bodyFieldDef.Type?.ItemType;
    if (!itemType) return rawExample;

    const itemFieldDef = typeDefToFieldDef(itemType, bodyFieldDef);
    return rawExample.map((elem: any) => {
      if (isExamplePlaceholder(elem)) {
        const generated = calculateExample(
          itemFieldDef,
          elem === "..." ? undefined : elem,
          undefined,
          false,
          { operation: op, templateDefaultValue: true },
        );
        return generated !== undefined ? generated : elem;
      }
      if (typeof elem === "object" && elem !== null) {
        return resolveBodyExampleValue(itemFieldDef, elem, op);
      }
      return elem;
    });
  }

  // Handle map with placeholder values
  if (
    bodyTypeStr === "map" &&
    typeof rawExample === "object" &&
    !Array.isArray(rawExample)
  ) {
    const itemType = bodyFieldDef.Type?.ItemType;
    if (!itemType) return rawExample;

    const itemFieldDef = typeDefToFieldDef(itemType, bodyFieldDef);
    const result: Record<string, any> = {};
    for (const [key, value] of Object.entries(rawExample)) {
      if (isExamplePlaceholder(value)) {
        const generated = calculateExample(
          itemFieldDef,
          value === "..." ? undefined : value,
          undefined,
          false,
          { operation: op, templateDefaultValue: true },
        );
        result[key] = generated !== undefined ? generated : value;
      } else if (typeof value === "object" && value !== null) {
        result[key] = resolveBodyExampleValue(itemFieldDef, value, op);
      } else {
        result[key] = value;
      }
    }
    return result;
  }

  // Recurse into class objects to resolve nested placeholders
  if (
    bodyTypeStr === "class" &&
    typeof rawExample === "object" &&
    !Array.isArray(rawExample)
  ) {
    const fields = bodyFieldDef.Type?.Fields || [];
    if (fields.length === 0) return rawExample;

    const result = { ...rawExample };
    for (const field of fields) {
      if (field.Const) continue;
      const key = field.OriginalName || field.Name;
      if (!(key in result)) continue;

      result[key] = resolveBodyExampleValue(field, result[key], op);
    }
    return result;
  }

  return rawExample;
}

// Get example value for a parameter field in CLI tests
// Simplified version that doesn't require response variable name resolution
// @ts-ignore
function getCLIFieldExample(usageContext: UsageContext, field: FieldDef): any {
  const exampleName = usageContext.ExampleName;
  const op = usageContext.Operation;

  // Check if it's a parameter field
  if (field.Annotations?.Has("param")) {
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
        const example = findExampleWithFallback(
          param.Examples ?? [],
          exampleName,
        );
        return getSimpleExampleValue(example);
      }
    }
  }

  // For flattened body fields, extract value from request body example payload
  if (op.Arguments?.BodyFields && op.Arguments.BodyFields.length > 0) {
    const isBodyField = op.Arguments.IsBodyField(field);
    if (isBodyField && op.Request?.Examples) {
      const bodyExample = findExampleWithFallback(
        op.Request.Examples,
        exampleName,
      );
      const rawBody = getSimpleExampleValue(bodyExample);
      // Body examples may come back as JSON string - parse to object
      let bodyValue: any = rawBody;
      if (rawBody && typeof rawBody === "string") {
        try {
          bodyValue = JSON.parse(rawBody);
        } catch (_e) {
          bodyValue = undefined;
        }
      }
      if (bodyValue && typeof bodyValue === "object") {
        const fieldKey = field.OriginalName || field.Name;
        if (fieldKey in bodyValue) {
          return bodyValue[fieldKey];
        }
      }
    }
  }

  // Try to get from field type examples
  const example = findExampleWithFallback(
    field.Type?.Examples ?? [],
    exampleName,
  );
  return getSimpleExampleValue(example);
}

// Generate default security test credentials from the security type definition.
// Used as a fallback when no Arazzo test-level security examples are provided.
// Pushes Go string-literal args (e.g., `"--my-api-key"`, `"test_api_key"`) into args[].
function generateDefaultSecurityArgs(
  securityType: TypeDef,
  args: string[],
): void {
  if (!securityType?.Fields?.length) return;

  // Identify ALL security scheme options (class, string, union, array)
  // with security annotations. Each top-level annotated field is a separate
  // scheme alternative (e.g., apiKeyAuth OR oauth2).
  const schemeOptions: Array<{
    field: FieldDef;
    secType: string;
    secSubType: string;
    isClass: boolean;
  }> = [];

  for (const field of securityType.Fields) {
    if (field.Const) continue;
    const secAnno = field.Annotations?.Get("security") as
      | SecurityAnnotation
      | undefined;
    if (!secAnno) continue;

    schemeOptions.push({
      field,
      secType: secAnno.SecType || "",
      secSubType: secAnno.SubType || "",
      isClass: field.Type.Type.toString() === "class",
    });
  }

  if (schemeOptions.length > 1) {
    // Multi-scheme security: prefer non-OAuth2 to avoid hooks overriding simpler auth
    const preferred =
      schemeOptions.find((o) => o.secType !== "oauth2") || schemeOptions[0];

    if (preferred.isClass) {
      // Class scheme: emit inner fields (e.g., api_key, api_secret)
      for (const innerField of preferred.field.Type.Fields || []) {
        if (innerField.Const) continue;
        if (innerField.Type.Type.toString() !== "string") continue;
        const name = innerField.Name.toLowerCase();
        if (name.includes("tokenurl") || name.includes("token_url")) continue;

        const flagName = sanitizeFlagName(innerField.Name);
        const testValue = getTestValueForSchemeField(
          innerField.Name,
          preferred.secType,
          preferred.secSubType,
        );
        pushFlagArg(args, flagName, `"${testValue}"`);
      }
    } else {
      // Primitive scheme (string/array): emit the single flag directly
      const flagName = sanitizeFlagName(preferred.field.Name);
      const testValue = getTestValueForSchemeField(
        preferred.field.Name,
        preferred.secType,
        preferred.secSubType,
      );
      pushFlagArg(args, flagName, `"${testValue}"`);
    }
  } else {
    // Single scheme: use all leaf fields
    const fields = flattenCLISecurityObject(securityType);
    for (const field of fields) {
      if (isTokenURLField(field)) continue;
      if (field.isArray) continue;
      const testValue = getDefaultSecurityTestValue(field);
      pushFlagArg(args, field.flagName, `"${testValue}"`);
    }
  }
}

// Template CLI args for a test step's usage context
// Returns a string like: "tag1", "get-user", "--id", "test123"
// @ts-ignore
// Helper: extract security example values from the UsageContext's security scope
// and generate CLI flag arguments for operation-level security.
function templateSecurityArgs(
  usageContext: UsageContext,
  args: string[],
): void {
  const op = usageContext.Operation;

  // Determine the effective security type: use operation-level if present,
  // otherwise fall back to global security only if the operation inherits it.
  // Operations with security: [{}] (empty security) define their own security
  // as "none required" — don't fall back to global security for those.
  let effectiveSecurityType = op.Security?.Type;
  if (!effectiveSecurityType && opUsesGlobalSecurity(op)) {
    effectiveSecurityType = context.Global.AST.MainSDK.Security?.Type;
  }
  if (!effectiveSecurityType) return;

  // Find the security scope from Arazzo test data
  let securityExample: any = undefined;
  if (usageContext.Scopes) {
    for (const scope of usageContext.Scopes) {
      if (scope.Feature === "security" && scope.Value) {
        if ("example" in scope.Value) {
          securityExample = scope.Value.example;
        } else {
          securityExample = getSimpleExampleValue(scope.Value);
        }
        break;
      }
    }
  }

  // Also check workflow-level security as fallback
  if (securityExample === undefined && usageContext.Test?.Security) {
    securityExample = getSimpleExampleValue(usageContext.Test.Security);
  }

  if (securityExample === undefined) {
    // When global security is optional due to optional-scheme (operations define
    // security: [{}]), don't emit default credentials — the operation doesn't
    // require auth. This matches the Go SDK behavior where no WithSecurity() is
    // emitted for optional-scheme security without an explicit example.
    const secOptional = context.Global.AST.MainSDK.Security?.Optional;
    const secReason =
      context.Global.AST.MainSDK.SecurityConfig?.OptionalityReason;
    if (secOptional && secReason?.toString() === "optional-scheme") {
      return;
    }
    // No Arazzo security example — generate default test credentials
    // from the effective security type so each test is explicit.
    generateDefaultSecurityArgs(effectiveSecurityType, args);
    return;
  }

  // Parse JSON if needed
  if (typeof securityExample === "string") {
    try {
      securityExample = JSON.parse(securityExample);
    } catch (_e) {
      return;
    }
  }

  if (typeof securityExample !== "object" || securityExample === null) return;

  // Flatten the security type to get leaf string fields with flag names
  const secFields = flattenCLISecurityObject(effectiveSecurityType);

  // Map example values to flag names by matching field names
  for (const secField of secFields) {
    // Try to find the value in the example object by field name (various casings)
    const fieldName = secField.field.Name;
    const originalName = secField.field.OriginalName || fieldName;
    const value =
      securityExample[originalName] ??
      securityExample[fieldName] ??
      securityExample[secField.flagName] ??
      securityExample[secField.configKey];

    if (value === undefined) {
      // Try nested: some security examples use scheme names as keys
      // e.g., {customSchemeAppId: {appId: "...", secret: "..."}}
      for (const key of Object.keys(securityExample)) {
        const nested = securityExample[key];
        if (typeof nested === "object" && nested !== null) {
          const nestedValue =
            nested[originalName] ??
            nested[fieldName] ??
            nested[secField.flagName];
          if (nestedValue !== undefined) {
            args.push(
              `"--${secField.flagName}"`,
              formatCLIArgValue(String(nestedValue)),
            );
            break;
          }
        }
      }
      continue;
    }

    pushFlagArg(args, secField.flagName, formatCLIArgValue(String(value)));
  }
}

function templateCLIArgs(usageContext: UsageContext): string {
  const args: string[] = [];
  const operation = usageContext.Operation;

  // Add the full command path (includes parent groups like "tag1", "subgroup")
  const cmdPath = getCLICommandPath(operation);
  for (const part of cmdPath) {
    args.push(`"${part}"`);
  }

  // Add operation-level security flags from Arazzo test examples
  templateSecurityArgs(usageContext, args);

  // Use example name from the step's usage context
  const exampleName = usageContext.ExampleName;

  // Handle request body operations vs parameter-only operations
  if (operation.Request) {
    const serMethod = operation.SerializationMethod?.toString() || "";

    if (
      serMethod === "multipart" &&
      operation.Request.IsRequestBody &&
      operation.Request.RequestBody
    ) {
      // Multipart operation - pass individual file/field flags
      const example = findExampleWithFallback(
        operation.Request.Examples ?? [],
        exampleName,
      );
      const rawBody = getSimpleExampleValue(example);
      let bodyPayload: any = rawBody;
      if (rawBody && typeof rawBody === "string") {
        try {
          bodyPayload = JSON.parse(rawBody);
        } catch (_e) {
          bodyPayload = undefined;
        }
      }

      if (bodyPayload && typeof bodyPayload === "object") {
        const formFields = operation.Request.RequestBody.Type?.Fields || [];
        for (const formField of formFields) {
          if (formField.Const) continue;
          const flagName = sanitizeFlagName(formField.Name);
          const fieldKey = formField.OriginalName || formField.Name;
          let value = bodyPayload[fieldKey];
          if (value === undefined) continue;
          if (value === null && !(formField.Nullable && formField.Optional))
            continue;

          // Resolve placeholder examples (union "anyOf[N]", array ["..."], etc.)
          value = resolveBodyExampleValue(formField, value, operation);

          // Check for x-file: prefix indicating a test file reference
          if (typeof value === "string" && value.startsWith("x-file: ")) {
            const fileName = value.substring("x-file: ".length).trim();
            args.push(
              `"--${flagName}"`,
              `"../.speakeasy/testfiles/${fileName}"`,
            );
          } else if (
            Array.isArray(value) &&
            value.length > 0 &&
            typeof value[0] === "string" &&
            (value[0] as string).startsWith("x-file: ")
          ) {
            // Array of file references — comma-separated paths with optional ;fileName= overrides
            const filePaths = value.map((v: any) => {
              const fileRef = (v as string).substring("x-file: ".length).trim();
              const semiIdx = fileRef.indexOf(";fileName=");
              if (semiIdx >= 0) {
                const path = fileRef.substring(0, semiIdx);
                const nameOverride = fileRef.substring(
                  semiIdx + ";fileName=".length,
                );
                return `../.speakeasy/testfiles/${path};fileName=${nameOverride}`;
              }
              return `../.speakeasy/testfiles/${fileRef}`;
            });
            pushFlagArg(args, flagName, `"${filePaths.join(",")}"`);
          } else if (fieldNeedsJSONFormat(formField)) {
            pushFlagArg(args, flagName, formatCLIArgValueAsJSON(value));
          } else {
            pushFlagArg(args, flagName, formatCLIArgValue(value));
          }
        }
      }
    } else if (
      operation.Request.IsRequestBody &&
      operation.Request.RequestBody &&
      isRequestBodyExpandable(operation)
    ) {
      // Expandable request body - pass individual field flags (not a single JSON blob)
      const bodyExample = findExampleWithFallback(
        operation.Request.Examples ?? [],
        exampleName,
      );
      let bodyPayload: any = getSimpleExampleValue(bodyExample);
      if (bodyPayload && typeof bodyPayload === "string") {
        try {
          bodyPayload = JSON.parse(bodyPayload);
        } catch (_e) {
          bodyPayload = undefined;
        }
      }

      if (bodyPayload && typeof bodyPayload === "object") {
        // Resolve placeholder examples (union "anyOf[N]", array ["..."], etc.)
        bodyPayload = resolveBodyExampleValue(
          operation.Request.RequestBody,
          bodyPayload,
          operation,
        );
        // Coerce values to match Go types before generating flags
        bodyPayload = coerceRequestBodyExample(bodyPayload, operation);

        const bodyFields = operation.Request.RequestBody.Type?.Fields || [];
        for (const field of bodyFields) {
          if (field.Const) continue;
          // Use sanitizeFlagNameWithReserved(field.Name) to match metadata generation
          const flagName = sanitizeFlagNameWithReserved(field.Name);
          const fieldKey = field.OriginalName || field.Name;
          const value = bodyPayload[fieldKey];
          if (value === undefined) continue;
          // Nullable fields with null examples: for nullable+optional fields (OptionalNullable
          // wrapper), fall through to pass "null" via FlagKindJSON to set the "explicitly null"
          // state. For nullable-only fields, skip — the nil pointer serializes as JSON null.
          if (value === null && !(field.Nullable && field.Optional)) continue;

          // Check if this class field is expanded to nested dot-notation flags
          const fieldTypeStr = field.Type?.Type?.toString() || "";
          if (fieldTypeStr === "class" && shouldExpandNestedField(field)) {
            if (typeof value === "object" && value !== null) {
              for (const subField of field.Type.Fields || []) {
                if (subField.Const) continue;
                const subFlagName = `${flagName}.${sanitizeFlagNameWithReserved(
                  subField.Name,
                )}`;
                const subKey = subField.OriginalName || subField.Name;
                const subValue = value[subKey];
                if (subValue === undefined) continue;
                if (
                  subValue === null &&
                  !(subField.Nullable && subField.Optional)
                )
                  continue;
                if (fieldNeedsJSONFormat(subField)) {
                  args.push(
                    `"--${subFlagName}"`,
                    formatCLIArgValueAsJSON(subValue),
                  );
                } else {
                  pushFlagArg(args, subFlagName, formatCLIArgValue(subValue));
                }
              }
            }
          } else if (fieldNeedsJSONFormat(field)) {
            pushFlagArg(args, flagName, formatCLIArgValueAsJSON(value));
          } else {
            pushFlagArg(args, flagName, formatCLIArgValue(value));
          }
        }
      }
    } else if (
      operation.Request.IsRequestBody &&
      operation.Request.RequestBody
    ) {
      // Non-expandable request body - pass JSON via a single flag
      const reqBodyField = operation.Request.RequestBody;
      const flagName = getRequestBodyFlagName(reqBodyField);

      // Get the example from operation.Request.Examples (not field.Type.Examples)
      const example = findExampleWithFallback(
        operation.Request.Examples ?? [],
        exampleName,
      );

      if (example !== undefined) {
        // Get the raw example value using our simple function
        let exampleValue = getSimpleExampleValue(example);
        if (exampleValue !== undefined) {
          // Resolve placeholder examples (union "oneOf[N]", array ["...", "..."])
          exampleValue = resolveBodyExampleValue(
            reqBodyField,
            exampleValue,
            operation,
          );
          // Coerce values to match Go types (e.g., number -> string for *string fields)
          exampleValue = coerceRequestBodyExample(exampleValue, operation);
          // Always JSON.stringify to ensure valid JSON encoding
          // (getSimpleExampleValue now parses JSON, so string values are raw strings)
          args.push(
            `"--${flagName}"`,
            goStringLiteral(JSON.stringify(exampleValue)),
          );
        }
      } else {
        // No example provided — for empty object bodies (class with no fields),
        // send '{}' so the SDK sends the empty body to the mock server.
        const bodyTypeStr = reqBodyField.Type?.Type?.toString() || "";
        const nonConstFields = (reqBodyField.Type?.Fields || []).filter(
          (f: FieldDef) => !f.Const,
        );
        if (bodyTypeStr === "class" && nonConstFields.length === 0) {
          pushFlagArg(args, flagName, `"{}"`);
        }
      }
    } else if (operation.Request.Field) {
      // Parameter/flattened operation - pass individual flags
      const reqFields = operation.Request.Field.Type?.Fields || [];

      // Get request body example payload for flattened body fields
      const bodyExamplePayload = extractBodyExamplePayload(
        operation,
        exampleName,
      );

      for (const field of reqFields) {
        // Skip const fields - they don't become flags
        if (field.Const) {
          continue;
        }

        // Check if this is a body struct field that gets expanded into nested flags
        // (no param annotation + class type + expanded by CLI command via shouldExpandNestedField)
        // Also handle multipart body fields which are expanded via collectMultipartMetadata
        const isParam = field.Annotations?.Has("param");
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
          // Iterate body sub-fields with prefix matching CLI flag naming.
          // Use getBodyFlagPrefix to flatten the prefix when no naming conflicts with sibling params.
          const bodyPrefix = getBodyFlagPrefix(field, reqFields, "");
          for (const subField of field.Type.Fields) {
            if (subField.Const) continue;
            const subFlagName = bodyPrefix
              ? `${bodyPrefix}.${sanitizeFlagNameWithReserved(subField.Name)}`
              : sanitizeFlagNameWithReserved(subField.Name);
            const fieldKey = subField.OriginalName || subField.Name;
            let subValue = bodyExamplePayload[fieldKey];
            if (subValue === undefined) continue;
            // Resolve placeholder examples (union "anyOf[N]", array ["..."], etc.)
            subValue = resolveBodyExampleValue(subField, subValue, operation);
            // Handle file references for multipart file fields
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
        // Non-expanded body class fields use a single JSON flag
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

        // Non-class body fields (unions, arrays, maps, etc.) use a single JSON flag
        if (
          !isParam &&
          !isClassType &&
          bodyExamplePayload &&
          field.Annotations?.Has("request")
        ) {
          const flagName = sanitizeFlagNameWithReserved(field.Name);
          let payload = resolveBodyExampleValue(
            field,
            bodyExamplePayload,
            operation,
          );
          payload = coerceRequestBodyExample(payload, operation);
          args.push(
            `"--${flagName}"`,
            goStringLiteral(JSON.stringify(payload)),
          );
          continue;
        }

        // Use getCLIFieldExample which handles parameter examples
        const exampleValue = getCLIFieldExample(usageContext, field);

        if (exampleValue === undefined) {
          continue;
        }

        // Convert field name to flag name (kebab-case, matching metadata generation)
        const flagName = sanitizeFlagNameWithReserved(field.Name);

        // Format the value for CLI
        const formattedValue = formatCLIArgValue(exampleValue);

        pushFlagArg(args, flagName, formattedValue);
      }
    }
  }

  return args.join(", ");
}
registerTemplateFunc("templateCLIArgs", templateCLIArgs);

// Generate a shell-friendly CLI usage command string for standalone usage snippets.
// Returns something like: cli tag1 list-test1 --page 1 --query-param1 "test"
// @ts-ignore
function templateCLIUsageCommand(usageContext: UsageContext): string {
  const cliName = sanitizeCliName();

  // Reuse templateCLIArgs which returns Go-quoted comma-separated values
  // e.g.: "tag1", "list-test1", "--page", "1"
  const goArgs = templateCLIArgs(usageContext);
  if (!goArgs) {
    return cliName;
  }

  // Parse Go-quoted args back to raw values
  const parts: string[] = [];
  // Match each "..." segment
  const matches = goArgs.match(/"([^"\\]*(?:\\.[^"\\]*)*)"/g);
  if (matches) {
    for (const m of matches) {
      // Strip outer quotes and decode Go string escapes in one pass —
      // named escapes, and every \xNN hex escape the escape core emits
      // (template delimiters and control characters alike). A single pass
      // keeps a literal backslash (encoded \\) from being re-read as the
      // start of another escape.
      const raw = m.slice(1, -1);
      let val = "";
      for (let i = 0; i < raw.length; i++) {
        const c = raw[i];
        if (c !== "\\") {
          val += c;
          continue;
        }
        const n = raw[++i];
        switch (n) {
          case "n":
            val += "\n";
            break;
          case "r":
            val += "\r";
            break;
          case "t":
            val += "\t";
            break;
          case '"':
            val += '"';
            break;
          case "\\":
            val += "\\";
            break;
          case "x": {
            const hex = raw.slice(i + 1, i + 3);
            val += String.fromCharCode(parseInt(hex, 16));
            i += 2;
            break;
          }
          default:
            val += n ?? "";
        }
      }
      parts.push(val);
    }
  }

  // Build shell command: quote values that contain spaces or special chars
  const shellArgs: string[] = [];
  for (const part of parts) {
    if (part.startsWith("--") || /^[a-zA-Z0-9._\-\/]+$/.test(part)) {
      shellArgs.push(part);
    } else {
      // Shell-escape: wrap in single quotes, escape existing single quotes
      shellArgs.push("'" + part.replace(/'/g, "'\\''") + "'");
    }
  }

  return cliName + " " + shellArgs.join(" ");
}
registerTemplateFunc("templateCLIUsageCommand", templateCLIUsageCommand);

/**
 * Extract the body example payload from an operation's request examples.
 * Parses JSON string bodies into objects. Returns undefined if no body example.
 */
function extractBodyExamplePayload(
  op: Operation,
  exampleName: string | undefined,
): any {
  if (!op.Request?.Examples) return undefined;
  const bodyExample = findExampleWithFallback(op.Request.Examples, exampleName);
  const rawBody = getSimpleExampleValue(bodyExample);
  if (rawBody && typeof rawBody === "string") {
    try {
      return JSON.parse(rawBody);
    } catch (_e) {
      return undefined;
    }
  }
  if (rawBody && typeof rawBody === "object") return rawBody;
  return undefined;
}

/**
 * Check if a test qualifies for a stdin-flavor variant.
 * Only single-step operation tests with an expandable body and available example data.
 */
function testHasExpandableBody(test: Test): boolean {
  if (isMultiStepTest(test)) return false;
  const steps = test.Workflow?.Steps;
  if (!steps || steps.length === 0) return false;

  const step = steps[0];
  if (step.Type !== "operation") return false;

  const uc = step.UsageContext;
  const op = uc?.Operation;
  if (!op?.Request) return false;

  const exampleName = uc.ExampleName;

  // Case 1: IsRequestBody expandable
  if (
    op.Request.IsRequestBody &&
    op.Request.RequestBody &&
    isRequestBodyExpandable(op)
  ) {
    const example = findExampleWithFallback(
      op.Request.Examples ?? [],
      exampleName,
    );
    return getSimpleExampleValue(example) !== undefined;
  }

  // Case 2: Mixed param+body with expanded body field
  if (op.Request.Field) {
    const reqFields = op.Request.Field.Type?.Fields || [];
    for (const field of reqFields) {
      if (field.Const) continue;
      if (
        !field.Annotations?.Has("param") &&
        field.Type?.Type?.toString() === "class" &&
        shouldExpandNestedField(field)
      ) {
        return extractBodyExamplePayload(op, exampleName) !== undefined;
      }
    }
  }

  return false;
}
registerTemplateFunc("testHasExpandableBody", testHasExpandableBody);

/**
 * Generate CLI args for the stdin flavor: command path + security + param-only flags.
 * Body fields are excluded since they'll be passed via stdin.
 */
function templateCLIArgsStdin(usageContext: UsageContext): string {
  const args: string[] = [];
  const operation = usageContext.Operation;

  // Command path
  const cmdPath = getCLICommandPath(operation);
  for (const part of cmdPath) {
    args.push(`"${part}"`);
  }

  // Security args
  templateSecurityArgs(usageContext, args);

  // For IsRequestBody case: all fields are body fields, so no additional flags
  if (operation.Request?.IsRequestBody && isRequestBodyExpandable(operation)) {
    return args.join(", ");
  }

  // For mixed param+body case: only emit param flags
  if (operation.Request?.Field) {
    const reqFields = operation.Request.Field.Type?.Fields || [];
    for (const field of reqFields) {
      if (field.Const) continue;
      if (!field.Annotations?.Has("param")) continue;

      const exampleValue = getCLIFieldExample(usageContext, field);
      if (exampleValue === undefined) continue;

      const flagName = sanitizeFlagNameWithReserved(field.Name);
      pushFlagArg(args, flagName, formatCLIArgValue(exampleValue));
    }
  }

  return args.join(", ");
}
registerTemplateFunc("templateCLIArgsStdin", templateCLIArgsStdin);

/**
 * Generate the stdin body JSON as a Go string literal.
 * Coerces values to match Go types before serializing.
 */
function templateStdinBodyJSON(usageContext: UsageContext): string {
  const operation = usageContext.Operation;
  const exampleName = usageContext.ExampleName;

  // For IsRequestBody case: body example from operation.Request.Examples
  if (
    operation.Request?.IsRequestBody &&
    operation.Request?.RequestBody &&
    isRequestBodyExpandable(operation)
  ) {
    const example = findExampleWithFallback(
      operation.Request.Examples ?? [],
      exampleName,
    );
    let bodyPayload: any = getSimpleExampleValue(example);
    if (bodyPayload && typeof bodyPayload === "string") {
      try {
        bodyPayload = JSON.parse(bodyPayload);
      } catch (_e) {
        return `"{}"`;
      }
    }
    if (bodyPayload && typeof bodyPayload === "object") {
      // Resolve placeholder examples (union "anyOf[N]", array ["..."], etc.)
      bodyPayload = resolveBodyExampleValue(
        operation.Request.RequestBody,
        bodyPayload,
        operation,
      );
      bodyPayload = coerceRequestBodyExample(bodyPayload, operation);
      return goStringLiteral(JSON.stringify(bodyPayload));
    }
    return `"{}"`;
  }

  // For mixed param+body case: extract body example payload
  if (operation.Request?.Field) {
    const bodyPayload = extractBodyExamplePayload(operation, exampleName);
    if (bodyPayload) {
      // Resolve placeholder examples (union "anyOf[N]", array ["..."], etc.)
      const resolved = resolveBodyExampleValue(
        operation.Request.Field,
        bodyPayload,
        operation,
      );
      const coerced = coerceRequestBodyExample(resolved, operation);
      return goStringLiteral(JSON.stringify(coerced));
    }
  }

  return `"{}"`;
}
registerTemplateFunc("templateStdinBodyJSON", templateStdinBodyJSON);

// Load assertion and workflow sub-modules (must be after utility functions above)
require("includes/test-assertions.ts");
require("includes/test-workflows.ts");
