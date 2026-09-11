// =============================================================================
// CLI Operation Helpers
// Command inspection, security, SDK call args, global flags, field expansion.
// Descriptions are in descriptions.ts, metadata in metadata.ts.
// =============================================================================

// @ts-ignore
function templateReadmeTitle(): string {
  return sanitizeCliName();
}
unregisterTemplateFunc("templateReadmeTitle");
registerTemplateFunc("templateReadmeTitle", templateReadmeTitle);

// @ts-ignore
function templateReadmeForeword(): string {
  let apiName =
    (context.Global.AST as any).OpenAPIDocument?.Info?.Title ||
    templateReadmeTitle();
  // Avoid "Foo API API" when the title already ends with "API"
  apiName = apiName.replace(/\s+API$/i, "");
  return `Command-line interface for the *${apiName}* API.`;
}
unregisterTemplateFunc("templateReadmeForeword");
registerTemplateFunc("templateReadmeForeword", templateReadmeForeword);

/**
 * Check if a field should be expanded into nested flags.
 * Returns true for class types that can be flattened into dot-notation flags.
 * JSON and form request bodies are always expanded — complex fields within
 * them get individual FlagKindJSON entries rather than blocking all expansion.
 */
function shouldExpandNestedField(field: FieldDef): boolean {
  const typeDef = field.Type;
  if (typeDef.Type.toString() !== "class") {
    return false;
  }

  // Check if the type has fields to expand
  const nestedFields = typeDef.Fields || [];
  if (nestedFields.length === 0) {
    return false;
  }

  // Don't expand MultipartRequestBody - it has its own handling for file fields
  const inputType = getInputClassType(field);
  if (inputType === "MultipartRequestBody") {
    return false;
  }

  // JSON and form request bodies are always expanded.
  // collectMetadataFromFields classifies each field individually:
  // simple fields → typed flags, complex fields (maps, nested classes) → FlagKindJSON.
  if (inputType === "JSONRequestBody" || inputType === "FormRequestBody") {
    return true;
  }

  // For other class types (params, nested objects), don't expand.
  // Params use FlagKindJSON (single JSON flag) which matches test generation.
  return false;
}
registerTemplateFunc("shouldExpandNestedField", shouldExpandNestedField);

/**
 * Determine the flag name prefix for body fields in a mixed param+body operation.
 * Returns an empty prefix (flattened) when no body field name conflicts with a
 * sibling param flag name. Returns the body field's sanitized name as prefix
 * when there are conflicts, to avoid ambiguity.
 */
function getBodyFlagPrefix(
  bodyField: FieldDef,
  siblingFields: FieldDef[],
  parentPrefix: string,
): string {
  const bodyFieldNames = new Set(
    (bodyField.Type.Fields || [])
      .filter((f: FieldDef) => !f.Const)
      .map((f: FieldDef) => sanitizeFlagNameWithReserved(f.Name)),
  );
  // Compare in the same namespace — both use bare sanitized names
  // since siblings and body fields are at the same nesting level
  const siblingNames = siblingFields
    .filter((f: FieldDef) => f !== bodyField && !f.Const)
    .map((f: FieldDef) => sanitizeFlagNameWithReserved(f.Name));
  const hasConflict = siblingNames.some((n: string) => bodyFieldNames.has(n));
  if (hasConflict) {
    return parentPrefix
      ? `${parentPrefix}.${sanitizeFlagNameWithReserved(bodyField.Name)}`
      : sanitizeFlagNameWithReserved(bodyField.Name);
  }
  return parentPrefix; // Flatten: use parent prefix (empty at top level)
}
registerTemplateFunc("getBodyFlagPrefix", getBodyFlagPrefix);

/**
 * Get a user-friendly flag name for a request body.
 * Uses type name for shared/component models, param name for operation-specific models.
 */
function getRequestBodyFlagName(field: FieldDef): string {
  const typeName = field.Type?.Name || "";
  const fieldName = field.Name || "";

  // Check if it's a shared/component model (not operation-specific)
  const scope = field.Type?.Scope?.toString() || "";
  const isComponentModel = scope !== "" && scope !== "operations";

  if (isComponentModel && typeName) {
    // Use the type name for component models (e.g., BaseUser -> user)
    return sanitizeFlagName(typeName);
  }

  // For operations models or when no type name, use the field/param name
  if (fieldName) {
    return sanitizeFlagName(fieldName);
  }

  // Fallback
  return typeName ? sanitizeFlagName(typeName) : "data";
}
registerTemplateFunc("getRequestBodyFlagName", getRequestBodyFlagName);

/**
 * Generate request body description for flag help text.
 * Uses the type name to create a user-friendly description.
 */
function templateRequestBodyDescription(field: FieldDef): string {
  let desc = "";

  // Use the field's own description if available
  if (field.Comments?.Description) {
    desc = field.Comments.Description.trim();
  }

  // For class types, show field names as a hint
  if (!desc && field.Type.Type?.toString() === "class" && field.Type.Fields) {
    const fieldNames = field.Type.Fields.filter((f: FieldDef) => !f.Const).map(
      (f: FieldDef) => f.Name,
    );
    if (fieldNames.length > 0) {
      const shown = fieldNames.slice(0, 6);
      const suffix = fieldNames.length > 6 ? ", ..." : "";
      desc = `JSON object with fields: ${shown.join(", ")}${suffix}`;
    }
  }

  if (!desc) {
    desc = "Request body as JSON";
  }

  if (!field.Optional) {
    desc += " [required]";
  }
  desc +=
    ". Can also be provided via stdin; @path reads a file, @- reads stdin to EOF.";

  return escapeGoString(desc);
}
registerTemplateFunc(
  "templateRequestBodyDescription",
  templateRequestBodyDescription,
);

/**
 * Check if an operation has any flags to register.
 * Used to conditionally generate the register function.
 */
function operationHasFlags(op: Operation): boolean {
  // Check for per-operation security
  if (op.Security) {
    return true;
  }

  // Check for request flags
  if (op.Request) {
    if (op.Request.IsRequestBody) {
      // Request body operations have a request-body flag
      return true;
    }
    // Check if there are any non-const fields
    const fields = op.Request.Field?.Type?.Fields || [];
    return fields.some((f: FieldDef) => !f.Const);
  }

  return false;
}
registerTemplateFunc("operationHasFlags", operationHasFlags);

/**
 * Check if an operation has per-operation security.
 */
function operationHasSecurity(op: Operation): boolean {
  return op.Security != undefined && op.Security != null;
}
registerTemplateFunc("operationHasSecurity", operationHasSecurity);

/**
 * Check if an operation has operation-level server URLs.
 * When true, the global --server-url needs to be passed as a per-request option.
 */
function operationHasServers(op: Operation): boolean {
  return (op.Servers?.Servers?.length ?? 0) > 0;
}
registerTemplateFunc("operationHasServers", operationHasServers);

/**
 * Check if any response type in an operation has a jq transform.
 * When true, we must NOT use WithSkipDeserialization() so that the SDK's
 * UnmarshalJSON() runs and applies the transform.
 */
function hasResponseTransform(op: Operation): boolean {
  if (!isFeatureUsed("transformJq")) return false;
  // Check if the request body type has a transform extension.
  // Request + response transforms typically come as a pair, so detecting
  // a request transform is a reliable proxy for response transforms too.
  if (hasTransformExtension(op.Request?.RequestBody?.Type)) return true;
  // Also check the response content types (envelope → content field → children)
  const responseType = op.Response?.Type;
  if (!responseType?.Fields) return false;
  for (const field of responseType.Fields) {
    const name = field.Name;
    if (RESPONSE_ENVELOPE_FIELDS_NO_HEADERS.has(name)) continue;
    // Content field — check it and its immediate children
    if (hasTransformExtension(field.Type)) return true;
    for (const subField of field.Type?.Fields || []) {
      if (hasTransformExtension(subField.Type)) return true;
    }
  }
  return false;
}

// Check if a type has a jq transform extension.
// Uses truthiness checks on the TransformFromAPI/TransformToAPI objects
// because the Type field is a Go enum proxy that doesn't compare with ===.
// The isFeatureUsed("transformJq") guard ensures only jq transforms are present.
function hasTransformExtension(typeDef: TypeDef | undefined): boolean {
  if (!typeDef?.Extensions) return false;
  return Boolean(
    typeDef.Extensions.TransformFromAPI || typeDef.Extensions.TransformToAPI,
  );
}
registerTemplateFunc("hasResponseTransform", hasResponseTransform);

/**
 * Walk a security field tree, visiting each leaf field (string or union string variant).
 * Handles class recursion and union flattening in one place.
 */
interface SecurityLeafInfo {
  name: string;
  description: string;
}

function walkSecurityLeafFields(
  fields: FieldDef[],
  visitor: (leaf: SecurityLeafInfo) => void,
): void {
  for (const field of fields || []) {
    if (field.Const) continue;
    const typeStr = field.Type.Type.toString();
    if (typeStr === "class") {
      walkSecurityLeafFields(field.Type.Fields || [], visitor);
    } else if (typeStr === "union") {
      for (const assocType of field.Type.AssociatedTypes || []) {
        if (assocType.Type.toString() === "class") {
          walkSecurityLeafFields(assocType.Fields || [], visitor);
        } else if (assocType.Type.toString() === "string") {
          visitor({
            name: assocType.Name,
            description:
              assocType.Comments?.Description || "Security credential",
          });
        }
      }
    } else if (typeStr === "string") {
      visitor({
        name: field.Name,
        description: field.Comments?.Description || "Security credential",
      });
    }
  }
}

/**
 * Generate security flag registration code.
 */
function templateSecurityFlagRegistration(op: Operation): string {
  if (!op.Security) return "";

  const lines: string[] = [];
  const seen = new Set<string>();
  walkSecurityLeafFields(op.Security.Type.Fields || [], (leaf) => {
    const flagName = sanitizeFlagNameWithReserved(leaf.name);
    if (!seen.has(flagName)) {
      seen.add(flagName);
      lines.push(
        `cmd.Flags().String("${flagName}", "", "${escapeGoString(
          leaf.description,
        )}")`,
      );
    }
  });
  return lines.join("\n    ");
}
registerTemplateFunc(
  "templateSecurityFlagRegistration",
  templateSecurityFlagRegistration,
);

/**
 * Generate security parsing code.
 */
function templateSecurityParsing(op: Operation): string {
  if (!op.Security) return "";

  addImport("internal/flagutil", true);
  const lines: string[] = [];
  const seen = new Set<string>();
  walkSecurityLeafFields(op.Security.Type.Fields || [], (leaf) => {
    const flagName = sanitizeFlagNameWithReserved(leaf.name);
    const varName = sanitizePrivateFieldName(leaf.name);
    if (!seen.has(varName)) {
      seen.add(varName);
      lines.push(`${varName}, _ := flagutil.GetStringFlag(cmd, "${flagName}")`);
    }
  });
  return lines.join("\n    ");
}
registerTemplateFunc("templateSecurityParsing", templateSecurityParsing);

/**
 * Generate security object construction code.
 * Recursively handles nested class types (e.g., Option structs containing BasicAuth).
 */
function templateSecurityConstruction(op: Operation): string {
  if (!op.Security) {
    return "";
  }

  // Note: sanitizeType automatically adds the correct import for the security type
  const securityTypeName = sanitizeType(op.Security.Type, false, "");
  const assignments = buildSecurityFieldAssignmentsWithChanged(
    op.Security.Type.Fields || [],
    "security",
  );

  return `var security ${securityTypeName}\n    ${assignments.join("\n    ")}`;
}

/**
 * Collect all leaf flag names from a security field tree.
 * Used to build Changed() conditions for nested class fields.
 */
function collectSecurityLeafFlagNames(fields: FieldDef[]): string[] {
  const flags: string[] = [];
  walkSecurityLeafFields(fields, (leaf) => {
    flags.push(sanitizeFlagNameWithReserved(leaf.name));
  });
  return flags;
}

/**
 * Build security field assignment expressions with Changed() guards.
 * Only sets security fields when their corresponding flags were explicitly provided.
 * This prevents the SDK from picking the wrong security option when multiple are available.
 */
function buildSecurityFieldAssignmentsWithChanged(
  fields: FieldDef[],
  parentVar: string,
): string[] {
  const result: string[] = [];
  for (const field of fields) {
    if (field.Const) continue;
    const fieldName = sanitizeFieldName(field.Name);
    const typeStr = field.Type.Type.toString();

    if (typeStr === "class") {
      // Collect all leaf flag names for the Changed() condition
      const leafFlags = collectSecurityLeafFlagNames(field.Type.Fields || []);
      if (leafFlags.length === 0) continue;
      const condition = leafFlags
        .map((f) => `cmd.Flags().Changed("${f}")`)
        .join(" || ");

      const nestedTypeName = sanitizeType(field.Type, false, "");
      const nestedAssignments = buildSecurityFieldAssignmentsInline(
        field.Type.Fields || [],
      );
      const construction = `${nestedTypeName}{${nestedAssignments.join(", ")}}`;

      if (field.Optional) {
        result.push(`if ${condition} {
        ${parentVar}.${fieldName} = &${construction}
    }`);
      } else {
        result.push(`if ${condition} {
        ${parentVar}.${fieldName} = ${construction}
    }`);
      }
    } else if (typeStr === "union") {
      // Delegate to the union security construction logic (token > credentials priority)
      result.push(templateUnionSecurityConstruction(field, parentVar));
    } else if (typeStr === "string") {
      const varName = sanitizePrivateFieldName(field.Name);
      const flagName = sanitizeFlagNameWithReserved(field.Name);
      if (field.Optional) {
        result.push(`if cmd.Flags().Changed("${flagName}") {
        ${parentVar}.${fieldName} = &${varName}
    }`);
      } else {
        result.push(`if cmd.Flags().Changed("${flagName}") {
        ${parentVar}.${fieldName} = ${varName}
    }`);
      }
    }
  }
  return result;
}

/**
 * Build inline security field assignments (for nested struct construction).
 * These don't need Changed() guards since the parent already checks.
 */
function buildSecurityFieldAssignmentsInline(fields: FieldDef[]): string[] {
  const result: string[] = [];
  for (const field of fields) {
    if (field.Const) continue;
    const fieldName = sanitizeFieldName(field.Name);
    const typeStr = field.Type.Type.toString();

    if (typeStr === "class") {
      const nestedTypeName = sanitizeType(field.Type, false, "");
      const nestedFields = buildSecurityFieldAssignmentsInline(
        field.Type.Fields || [],
      );
      const construction = `${nestedTypeName}{${nestedFields.join(", ")}}`;
      if (field.Optional) {
        result.push(`${fieldName}: &${construction}`);
      } else {
        result.push(`${fieldName}: ${construction}`);
      }
    } else if (typeStr === "string") {
      const varName = sanitizePrivateFieldName(field.Name);
      if (field.Optional) {
        result.push(`${fieldName}: &${varName}`);
      } else {
        result.push(`${fieldName}: ${varName}`);
      }
    }
  }
  return result;
}
registerTemplateFunc(
  "templateSecurityConstruction",
  templateSecurityConstruction,
);

/**
 * Generate the SDK method call argument list for an operation.
 * Uses getArguments() to determine correct ordering of security vs request args,
 * and handles flattened operations by extracting individual fields from the request struct.
 *
 * For infer-optional-args (CLI default):
 *   - Required security → security before request
 *   - Optional security → request before security (security is pointer)
 *
 * For flattened operations (per-operation x-speakeasy-max-method-params):
 *   - Individual struct fields are passed instead of the request struct
 */
function templateSDKCallArgsWithContext(
  op: Operation,
  reqVarName: string,
  contextExpr: string,
): string {
  const args = getArguments(op);
  const flattened = args.flattenedRequest;

  const parts: string[] = [contextExpr];

  for (const field of args.merged) {
    const isSecurity = args.security.some((s) => s === field);

    if (isSecurity) {
      if (field.Optional) {
        parts.push("&security");
      } else {
        parts.push("security");
      }
    } else {
      // Request field
      if (flattened) {
        // Flattened: pass individual fields from the built request struct
        parts.push(`${reqVarName}.${sanitizeFieldName(field.Name)}`);
      } else {
        // Non-flattened: pass the whole request struct
        if (sdkMethodTakesPointerRequest(op)) {
          parts.push(reqVarName);
        } else if (isOptionalReferenceTypeBody(op)) {
          // Optional reference-type body (bytes, array, map): BuildRequestBody
          // returns *T (nil when not provided) but SDK takes T (value type).
          // Use DerefOrZero for a nil-safe dereference.
          parts.push(`flagutil.DerefOrZero(${reqVarName})`);
        } else {
          parts.push(`*${reqVarName}`);
        }
      }
    }
  }

  parts.push("sdkOpts...");
  return parts.join(", ");
}

function templateSDKCallArgs(op: Operation, reqVarName: string): string {
  return templateSDKCallArgsWithContext(op, reqVarName, "cmd.Context()");
}
registerTemplateFunc("templateSDKCallArgs", templateSDKCallArgs);
registerTemplateFunc(
  "templateSDKCallArgsWithContext",
  templateSDKCallArgsWithContext,
);

/**
 * Check if the SDK method takes the request as a pointer.
 * This is true when the request field is Optional.
 */
function sdkMethodTakesPointerRequest(op: Operation): boolean {
  if (!op.Request) {
    return false;
  }
  // For IsRequestBody operations, check if the SDK method uses a pointer.
  // Reference types in Go (array→[]T, map, bytes→[]byte) never need pointers.
  // Nullable+optional with wrapper → OptionalNullable[T] value type, no pointer.
  // All other types (string, class, union, date, etc.) use pointers when optional/nullable.
  if (op.Request.IsRequestBody) {
    const bodyField = op.Request.RequestBody;
    const bodyType = bodyField.Type.Type.toString();
    if (bodyType === "array" || bodyType === "map" || bodyType === "bytes")
      return false;
    // When both nullable+optional and wrapper is enabled, the Go SDK uses
    // OptionalNullable[T] (value type) instead of *T
    if (
      bodyField.Nullable &&
      bodyField.Optional &&
      context.Global.Config.NullableOptionalWrapper
    )
      return false;
    return bodyField.Optional === true || bodyField.Nullable === true;
  }
  // When the request field is Optional, the SDK method takes a pointer
  return op.Request.Field.Optional === true;
}
registerTemplateFunc(
  "sdkMethodTakesPointerRequest",
  sdkMethodTakesPointerRequest,
);

/**
 * Check if an operation has an optional request body with a reference type
 * (bytes, array, map). These types don't use pointer params in the SDK method,
 * but BuildRequestBody returns *T which can be nil when the body isn't provided.
 */
function isOptionalReferenceTypeBody(op: Operation): boolean {
  if (!op.Request?.IsRequestBody) return false;
  const bodyField = op.Request.RequestBody;
  if (!bodyField.Optional) return false;
  const bodyType = bodyField.Type.Type.toString();
  return bodyType === "bytes" || bodyType === "array" || bodyType === "map";
}

// =============================================================================
// Multipart Form Data Support
// =============================================================================

/**
 * Check if a field in a multipart request body is a JSON-encoded field.
 * This is for complex object fields that are NOT file fields.
 */
function isMultipartJSONField(field: FieldDef): boolean {
  // Check if the field type is a class (object) but not a file type
  if (field.Type.Type.toString() !== "class") {
    return false;
  }
  // Use the common isMultipartFileField function (from common/fieldDefs.ts) to check if it's a file
  return !isMultipartFileField(field);
}
registerTemplateFunc("isMultipartJSONField", isMultipartJSONField);

// =============================================================================
// Global Parameter Helpers
// =============================================================================

/**
 * Check if the SDK has absolute server URLs defined.
 * Controls whether global server URL and selection options can be applied
 * while constructing the SDK client.
 */
function sdkHasServers(): boolean {
  return context.Global.AST.MainSDK.Servers?.HasAbsoluteURL() || false;
}
registerTemplateFunc("sdkHasServers", sdkHasServers);

/**
 * Check if the SDK's New() function takes only ...SDKOption (true) or
 * requires a serverURL string as its first argument (false).
 * Matches the Go SDK's $hasAnyServers logic: absolute global servers OR operation-level servers.
 */
function sdkNewTakesOnlyOptions(): boolean {
  return (
    sdkHasServers() ||
    context.Global.AST.MainSDK.HasAnyOperationServers() ||
    false
  );
}
registerTemplateFunc("sdkNewTakesOnlyOptions", sdkNewTakesOnlyOptions);

/**
 * Get the generated top-level Go SDK type name.
 */
function templateMainSDKTypeName(): string {
  return sanitizeClassName(context.Global.AST.MainSDK.Type.Name);
}
registerTemplateFunc("templateMainSDKTypeName", templateMainSDKTypeName);

// =============================================================================
// Catalog Commands (x-speakeasy-cli-catalog)
// =============================================================================

interface CliCatalogValue {
  Value: string;
  Description: string;
  IsDefault: boolean;
}

interface CliCatalog {
  Command: string;
  FuncName: string;
  Summary: string;
  Description: string;
  Values: CliCatalogValue[];
}

/**
 * Collect enum types annotated with x-speakeasy-cli-catalog. Each becomes a
 * generated listing command (e.g. a `models` command rendered from the model
 * enum, its descriptions, and its declared default) — the OpenAPI document is
 * the machine-readable manifest.
 */
function collectCliCatalogs(): CliCatalog[] {
  const catalogs: CliCatalog[] = [];
  const seen = new Set<string>();
  // Distinct declared commands that normalize to the same CLI command or the
  // same Go function identifier would redeclare symbols; each namespace is
  // validated independently and a collision fails generation by name.
  const byCommand = new Map<string, string>();
  const byFuncName = new Map<string, string>();
  let collisionError = "";
  try {
    const buckets = context.Global.AST.BucketedTypes;
    for (const [, models] of sequencedMapEntries(buckets)) {
      for (const [, types] of sequencedMapEntries(models)) {
        for (const t of types as TypeDef[]) {
          const ext: any = t.Extensions?.All?.["x-speakeasy-cli-catalog"];
          if (!ext || typeof ext !== "object" || !ext.command) continue;
          if (t.Type?.toString() !== "enum" || !t.Enum) continue;
          if (seen.has(ext.command)) continue;
          seen.add(ext.command);
          const commandName = sanitizeCLICommand(`${ext.command}`);
          const funcName = sanitizeClassName(`${ext.command}`);
          const priorCommand = byCommand.get(commandName);
          const priorFunc = byFuncName.get(funcName);
          if (priorCommand !== undefined && priorCommand !== `${ext.command}`) {
            collisionError = `x-speakeasy-cli-catalog: commands "${priorCommand}" and "${ext.command}" normalize to the same CLI command "${commandName}"; rename one`;
            break;
          }
          if (priorFunc !== undefined && priorFunc !== `${ext.command}`) {
            collisionError = `x-speakeasy-cli-catalog: commands "${priorFunc}" and "${ext.command}" normalize to the same generated identifier "${funcName}"; rename one`;
            break;
          }
          byCommand.set(commandName, `${ext.command}`);
          byFuncName.set(funcName, `${ext.command}`);
          const defaultValue =
            ext.default !== undefined ? `${ext.default}` : "";
          const values: CliCatalogValue[] = (t.Enum.Values || []).map(
            (v: string) => {
              const description = t.Enum.Descriptions?.[v];
              return {
                Value: v,
                Description: typeof description === "string" ? description : "",
                IsDefault: defaultValue !== "" && v === defaultValue,
              };
            },
          );
          if (values.length === 0) continue;
          catalogs.push({
            Command: sanitizeCLICommand(`${ext.command}`),
            FuncName: sanitizeClassName(`${ext.command}`),
            Summary: `${ext.summary || `List available ${ext.command}`}`,
            Description: `${ext.description || ""}`,
            Values: values,
          });
        }
      }
    }
  } catch (e) {
    // Catalog rendering is best-effort; a malformed extension never breaks generation.
  }
  if (collisionError) {
    // Identifier collisions would emit uncompilable Go; unlike malformed
    // extensions they must fail generation with the collision named.
    throw new Error(collisionError);
  }
  return catalogs;
}
registerTemplateFunc(
  "hasCatalogCommands",
  () => collectCliCatalogs().length > 0,
);

/**
 * Check if the API has global parameters defined.
 */
function hasGlobals(): boolean {
  const globals = context.Global.AST.MainSDK.Globals;
  return globals && globals.Fields && globals.Fields.length > 0;
}
registerTemplateFunc("hasGlobals", hasGlobals);

/**
 * Compute the SDK option function name for a global parameter.
 * Mirrors the Go SDK's `templateSDKOptionName` to produce matching names.
 */
function cliSDKOptionName(name: string): string {
  const reservedOptions = [
    "WithServerURL",
    "WithTemplatedServerURL",
    "WithClient",
    "WithSecurity",
    "WithSecuritySource",
    "WithServerIndex",
    "WithServer",
    "WithRetryConfig",
    "WithTimeout",
  ];

  let optionName = `With${sanitizeClassName(name)}`;
  if (reservedOptions.includes(optionName)) {
    optionName = `WithGlobal${sanitizeClassName(name)}`;
  }
  return optionName;
}

/**
 * Generate persistent flag registration code for all global parameters.
 * Called from root.go.stmpl to register globals as persistent flags.
 */
function templateGlobalFlagRegistration(): string {
  if (!hasGlobals()) return "";

  const globals = context.Global.AST.MainSDK.Globals;
  const lines: string[] = [];

  for (const field of globals.Fields) {
    const flagName = sanitizeFlagNameWithReserved(field.Name);
    const desc =
      field.Comments?.Summary?.trim() ||
      field.Comments?.Description?.trim()?.split(/[.\n]/)[0]?.trim() ||
      `Global ${sanitizeFlagName(field.Name)} parameter`;

    const envVar = context.Global.Config.EnvVarPrefix
      ? ` (env: ${templateGlobalEnvVars(field)})`
      : "";

    const typeStr = field.Type?.Type?.toString() || "string";
    // Schema defaults become flag defaults so zero-config invocations work
    // and cobra surfaces the default in help output.
    const defaultValue = field.Default?.Value ?? undefined;
    switch (typeStr) {
      case "integer":
      case "int32": {
        const def = defaultValue !== undefined ? `${defaultValue}` : "0";
        lines.push(
          `rootCmd.PersistentFlags().Int64("${flagName}", ${def}, "${escapeGoString(
            desc,
          )}${envVar}")`,
        );
        break;
      }
      case "number":
      case "float32": {
        const def = defaultValue !== undefined ? `${defaultValue}` : "0";
        lines.push(
          `rootCmd.PersistentFlags().Float64("${flagName}", ${def}, "${escapeGoString(
            desc,
          )}${envVar}")`,
        );
        break;
      }
      case "boolean": {
        const def = defaultValue !== undefined ? `${defaultValue}` : "false";
        lines.push(
          `rootCmd.PersistentFlags().Bool("${flagName}", ${def}, "${escapeGoString(
            desc,
          )}${envVar}")`,
        );
        break;
      }
      default: {
        // string, enum, etc.
        const def =
          defaultValue !== undefined
            ? `"${escapeGoString(`${defaultValue}`)}"`
            : `""`;
        lines.push(
          `rootCmd.PersistentFlags().String("${flagName}", ${def}, "${escapeGoString(
            desc,
          )}${envVar}")`,
        );
        break;
      }
    }
    lines.push(
      `_ = rootCmd.PersistentFlags().SetAnnotation("${flagName}", "speakeasy:group", []string{"API Parameters"})`,
    );
  }

  return lines.join("\n    ");
}
registerTemplateFunc(
  "templateGlobalFlagRegistration",
  templateGlobalFlagRegistration,
);

/**
 * Generate the body of buildGlobalOptions() in client.go.stmpl.
 * Reads global parameter flags and returns SDK options.
 * Priority: flag > env var > config file.
 */
function templateGlobalOptionsBuilding(): string {
  if (!hasGlobals()) return "";

  const globals = context.Global.AST.MainSDK.Globals;
  const lines: string[] = [];

  // Check if any globals have non-string types that need strconv
  const needsStrconv = globals.Fields.some((field: FieldDef) => {
    const t = field.Type?.Type?.toString() || "string";
    return (
      t === "integer" || t === "int32" || t === "number" || t === "float32"
    );
  });
  if (needsStrconv) {
    addImport("strconv");
  }

  for (const field of globals.Fields) {
    const flagName = sanitizeFlagNameWithReserved(field.Name);
    const optionName = cliSDKOptionName(field.Name);
    const typeStr = field.Type?.Type?.toString() || "string";
    const hasSchemaDefault = field.Default?.Value != null;

    switch (typeStr) {
      case "integer":
      case "int32":
        lines.push(`if flagutil.FlagChanged(cmd, "${flagName}") {`);
        lines.push(`    val, _ := flagutil.GetInt64Flag(cmd, "${flagName}")`);
        lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        lines.push(
          `} else if val := config.GetString("${flagName}"); val != "" {`,
        );
        lines.push(
          `    if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {`,
        );
        lines.push(`        opts = append(opts, sdk.${optionName}(intVal))`);
        lines.push(`    }`);
        if (hasSchemaDefault) {
          // Cobra holds the schema default even though the flag is unchanged;
          // pass it after explicit flag, environment, and config sources.
          lines.push(`} else {`);
          lines.push(`    val, _ := flagutil.GetInt64Flag(cmd, "${flagName}")`);
          lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        }
        lines.push(`}`);
        break;
      case "number":
      case "float32":
        lines.push(`if flagutil.FlagChanged(cmd, "${flagName}") {`);
        lines.push(`    val, _ := flagutil.GetFloat64Flag(cmd, "${flagName}")`);
        lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        lines.push(
          `} else if val := config.GetString("${flagName}"); val != "" {`,
        );
        lines.push(
          `    if floatVal, err := strconv.ParseFloat(val, 64); err == nil {`,
        );
        lines.push(`        opts = append(opts, sdk.${optionName}(floatVal))`);
        lines.push(`    }`);
        if (hasSchemaDefault) {
          lines.push(`} else {`);
          lines.push(
            `    val, _ := flagutil.GetFloat64Flag(cmd, "${flagName}")`,
          );
          lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        }
        lines.push(`}`);
        break;
      case "boolean":
        lines.push(`if flagutil.FlagChanged(cmd, "${flagName}") {`);
        lines.push(`    val, _ := flagutil.GetBoolFlag(cmd, "${flagName}")`);
        lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        lines.push(
          `} else if val := config.GetString("${flagName}"); val != "" {`,
        );
        lines.push(`    opts = append(opts, sdk.${optionName}(val == "true"))`);
        if (hasSchemaDefault) {
          lines.push(`} else {`);
          lines.push(`    val, _ := flagutil.GetBoolFlag(cmd, "${flagName}")`);
          lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        }
        lines.push(`}`);
        break;
      case "enum": {
        // Enum types are named string types that need a type conversion
        const goType = sanitizeType(field.Type, false, "");
        lines.push(`if flagutil.FlagChanged(cmd, "${flagName}") {`);
        lines.push(`    val, _ := flagutil.GetStringFlag(cmd, "${flagName}")`);
        lines.push(
          `    opts = append(opts, sdk.${optionName}(${goType}(val)))`,
        );
        lines.push(
          `} else if val := config.GetString("${flagName}"); val != "" {`,
        );
        lines.push(
          `    opts = append(opts, sdk.${optionName}(${goType}(val)))`,
        );
        if (hasSchemaDefault) {
          lines.push(`} else {`);
          lines.push(
            `    val, _ := flagutil.GetStringFlag(cmd, "${flagName}")`,
          );
          lines.push(
            `    opts = append(opts, sdk.${optionName}(${goType}(val)))`,
          );
        }
        lines.push(`}`);
        break;
      }
      default: // string
        lines.push(`if flagutil.FlagChanged(cmd, "${flagName}") {`);
        lines.push(`    val, _ := flagutil.GetStringFlag(cmd, "${flagName}")`);
        lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        lines.push(
          `} else if val := config.GetString("${flagName}"); val != "" {`,
        );
        lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        if (hasSchemaDefault) {
          lines.push(`} else {`);
          lines.push(
            `    val, _ := flagutil.GetStringFlag(cmd, "${flagName}")`,
          );
          lines.push(`    opts = append(opts, sdk.${optionName}(val))`);
        }
        lines.push(`}`);
        break;
    }
  }

  return lines.join("\n    ");
}
registerTemplateFunc(
  "templateGlobalOptionsBuilding",
  templateGlobalOptionsBuilding,
);

// =============================================================================
// Server Selection Helpers
// =============================================================================

/**
 * Check if the global server list is map-based (named servers) vs array-based (indexed).
 * Map-based: `var ServerList = map[string]string{...}` with `WithServer(name)`
 * Array-based: `var ServerList = []string{...}` with `WithServerIndex(idx)`
 */
function sdkServerMap(): boolean {
  return context.Global.AST.MainSDK.Servers?.ServerMap || false;
}
registerTemplateFunc("sdkServerMap", sdkServerMap);

/**
 * Check if the SDK has global server variables (e.g., {hostname}, {port}, {protocol}).
 */
function sdkHasServerVariables(): boolean {
  const vars = context.Global.AST.MainSDK.Servers?.GetVariables() || [];
  return vars.length > 0;
}
registerTemplateFunc("sdkHasServerVariables", sdkHasServerVariables);

/**
 * Generate persistent flag registration for global server variable flags.
 * Called from root.go.stmpl to register flags like --hostname, --port, --protocol.
 */
function templateServerVariableFlags(): string {
  const servers = context.Global.AST.MainSDK.Servers;
  if (!servers) return "";

  const variables = servers.GetVariables();
  if (variables.length === 0) return "";

  const lines: string[] = [];
  lines.push("// Server template variable flags");

  for (const v of variables) {
    const flagName = sanitizeFlagName(v.Name);
    const desc = `Server template variable: ${v.Name}`;
    lines.push(
      `rootCmd.PersistentFlags().String("${flagName}", "", "${escapeGoString(
        desc,
      )}")`,
    );
    lines.push(
      `_ = rootCmd.PersistentFlags().SetAnnotation("${flagName}", "speakeasy:group", []string{"Server"})`,
    );
  }

  return lines.join("\n    ");
}
registerTemplateFunc(
  "templateServerVariableFlags",
  templateServerVariableFlags,
);

/**
 * Generate an example flag snippet for server variables in the README.
 * E.g., "--hostname api.example.com --port 8080"
 */
function templateServerVariableFlagExample(): string {
  const servers = context.Global.AST.MainSDK.Servers;
  if (!servers) return "";

  const variables = servers.GetVariables();
  if (variables.length === 0) return "";

  return variables
    .map((v: any) => {
      const flagName = sanitizeFlagName(v.Name);
      const defaultVal = String(v.Default?.Value || "value");
      // The README example contract test executes these lines: a default
      // containing whitespace or shell metacharacters must stay one argument.
      const quoted = /^[A-Za-z0-9._\/:@%+=-]+$/.test(defaultVal)
        ? defaultVal
        : `'${defaultVal.replaceAll("'", `'\\''`)}'`;
      return `--${flagName} ${quoted}`;
    })
    .join(" ");
}
registerTemplateFunc(
  "templateServerVariableFlagExample",
  templateServerVariableFlagExample,
);

/**
 * Generate SDK option calls for global server variable flags in client.go.
 * Reads each server variable flag and calls the corresponding sdk.With*() option.
 */
function templateServerVariableOptions(): string {
  const servers = context.Global.AST.MainSDK.Servers;
  if (!servers) return "";

  const variables = servers.GetVariables();
  if (variables.length === 0) return "";

  const lines: string[] = [];

  for (const v of variables) {
    const flagName = sanitizeFlagName(v.Name);
    const optionName = cliSDKOptionName(v.Name);
    const typeStr = v.Type?.Type?.toString() || "string";

    // Check if variable type is an enum — needs sdk. prefix since it's defined in the sdk package
    if (typeStr === "enum") {
      const goType = sanitizeClassName(v.Type.Name);
      lines.push(
        `if v, _ := flagutil.GetStringFlag(cmd, "${flagName}"); v != "" {`,
      );
      lines.push(
        `    sdkOpts = append(sdkOpts, sdk.${optionName}(sdk.${goType}(v)))`,
      );
      lines.push(`}`);
    } else {
      lines.push(
        `if v, _ := flagutil.GetStringFlag(cmd, "${flagName}"); v != "" {`,
      );
      lines.push(`    sdkOpts = append(sdkOpts, sdk.${optionName}(v))`);
      lines.push(`}`);
    }
  }

  return lines.join("\n    ");
}
registerTemplateFunc(
  "templateServerVariableOptions",
  templateServerVariableOptions,
);

/**
 * Generate operation-level --server resolution code.
 * Resolves the --server flag against the operation's own server list,
 * then applies server template variables from flags.
 */
function templateOperationServerResolution(
  op: Operation,
  returnPrefix = "",
): string {
  if (!op.Servers || !op.Servers.Servers || op.Servers.Servers.length === 0) {
    return "";
  }

  const modelName = sanitizeFieldName(getOperationModelName(op));
  const isMap = op.Servers.ServerMap;
  const variables = op.Servers.GetVariables();

  addImport("operations");
  addImport("internal/flagutil", true);
  addImport("fmt");

  const lines: string[] = [];

  if (isMap) {
    lines.push(
      `serverURL, ok := operations.${modelName}ServerList[serverFlag]`,
    );
    lines.push(`if !ok {`);
    lines.push(
      `    return ${returnPrefix}fmt.Errorf("unknown server %q for this operation", serverFlag)`,
    );
    lines.push(`}`);
  } else {
    addImport("strconv");
    lines.push(`serverIdx, err := strconv.Atoi(serverFlag)`);
    lines.push(`if err != nil {`);
    lines.push(
      `    return ${returnPrefix}fmt.Errorf("invalid server index %q: must be an integer", serverFlag)`,
    );
    lines.push(`}`);
    lines.push(
      `if serverIdx < 0 || serverIdx >= len(operations.${modelName}ServerList) {`,
    );
    lines.push(
      `    return ${returnPrefix}fmt.Errorf("server index %d out of range (0-%d)", serverIdx, len(operations.${modelName}ServerList)-1)`,
    );
    lines.push(`}`);
    lines.push(`serverURL := operations.${modelName}ServerList[serverIdx]`);
  }

  // Build template variable params map
  if (variables.length > 0) {
    lines.push(`params := map[string]string{}`);
    for (const v of variables) {
      const flagName = sanitizeFlagName(v.Name);
      lines.push(
        `if v, _ := flagutil.GetStringFlag(cmd, "${flagName}"); v != "" { params["${v.Name}"] = v }`,
      );
    }
    lines.push(
      `sdkOpts = append(sdkOpts, operations.WithTemplatedServerURL(serverURL, params))`,
    );
  } else {
    lines.push(
      `sdkOpts = append(sdkOpts, operations.WithServerURL(serverURL))`,
    );
  }

  return lines.join("\n        ");
}
registerTemplateFunc(
  "templateOperationServerResolution",
  templateOperationServerResolution,
);

/**
 * Generate validation code for --server against the global server list.
 * Used in opcmd.go.stmpl for operations that do NOT have their own servers,
 * so that invalid --server values produce an error rather than silently using the default.
 * Delegates to output.ValidateGlobalServerIndex / output.ValidateGlobalServerName helpers.
 */
function templateGlobalServerValidation(returnPrefix = ""): string {
  if (serverSelectionFlag() === "none") {
    return "";
  }

  const servers = context.Global.AST.MainSDK.Servers;
  if (!servers || !servers.Servers || servers.Servers.length === 0) {
    return "";
  }

  addImport("internal/sdk", true);

  const lines: string[] = [];

  if (servers.ServerMap) {
    lines.push(
      `if err := output.ValidateGlobalServerName(cmd, sdk.ServerList); err != nil {`,
    );
  } else {
    lines.push(
      `if err := output.ValidateGlobalServerIndex(cmd, len(sdk.ServerList)); err != nil {`,
    );
  }
  lines.push(`    return ${returnPrefix}err`);
  lines.push(`}`);

  return lines.join("\n    ");
}
registerTemplateFunc(
  "templateGlobalServerValidation",
  templateGlobalServerValidation,
);

// @ts-ignore
function hasPagination(op: Operation): boolean {
  return Boolean(op.Extensions?.Pagination);
}
registerTemplateFunc("hasPagination", hasPagination);

// @ts-ignore
function getPaginationFieldNames(op: Operation): {
  contentFieldName: string;
  resultsFieldName: string;
} | null {
  const pagination = op.Extensions?.Pagination;
  if (!pagination) return null;

  const responseType = op.Response?.Type;
  if (!responseType?.Fields) return null;

  // Find the content field by skipping envelope fields — mirrors extractResultContent()
  // in output.go.stmpl at runtime.
  let contentField: FieldDef | undefined;
  for (const field of responseType.Fields) {
    const goName = sanitizeFieldName(field.Name);
    if (RESPONSE_ENVELOPE_FIELDS.has(goName)) continue;
    contentField = field;
    break;
  }
  if (!contentField) return null;

  const contentFieldName = sanitizeFieldName(contentField.Name);

  // If ResultsDot is available, resolve the JSON path to Go field names.
  // When empty (cursor/URL pagination without explicit results output),
  // resultsFieldName stays empty and PaginatedResult will output the whole content per page.
  let resultsFieldName = "";
  const resultsDot = pagination.Outputs?.ResultsDot;
  if (resultsDot) {
    const contentFieldType = contentField.Type;
    const parts = resultsDot.split(".");
    const goResultParts: string[] = [];
    let currentType = contentFieldType;

    for (const jsonPart of parts) {
      if (!currentType?.Fields) break;
      const matchField = currentType.Fields.find(
        (f: FieldDef) => originalFieldName(f) === jsonPart,
      );
      if (!matchField) break;
      goResultParts.push(sanitizeFieldName(matchField.Name));
      currentType = matchField.Type;
    }
    if (goResultParts.length === parts.length) {
      resultsFieldName = goResultParts.join(".");
    }
  }

  return { contentFieldName, resultsFieldName };
}

// @ts-ignore
function getPaginationContentFieldName(op: Operation): string {
  return getPaginationFieldNames(op)?.contentFieldName || "";
}
registerTemplateFunc(
  "getPaginationContentFieldName",
  getPaginationContentFieldName,
);

// @ts-ignore
function getPaginationResultsFieldName(op: Operation): string {
  return getPaginationFieldNames(op)?.resultsFieldName || "";
}
registerTemplateFunc(
  "getPaginationResultsFieldName",
  getPaginationResultsFieldName,
);

interface PaginationProbeContext {
  Type: string;
  CursorKind: string;
  NextCursor: string;
  NextURL: string;
  Results: string;
  HasLimit: boolean;
}

// Mirror the Go SDK's findPaginationCursorField lookup so the guard uses the
// request type that Next() actually coerces its response cursor into.
function findPaginationProbeCursorField(
  requestType: TypeDef | undefined,
  fieldName: string | undefined,
): FieldDef | undefined {
  if (!requestType?.Fields || !fieldName) return undefined;

  const direct = requestType.Fields.find((field) => field.Name === fieldName);
  if (direct) return direct;

  for (const field of requestType.Fields) {
    const nested = field.Type?.Fields?.find(
      (candidate: FieldDef) => candidate.Name === fieldName,
    );
    if (nested) return nested;
  }

  return undefined;
}

/**
 * Return the pagination metadata needed by the CLI's non-consuming raw-body
 * probe. Offset/limit pagination is intentionally represented without a
 * continuation expression: a full result page is not proof that another page
 * exists, so the CLI only emits truthful hints for cursor and URL pagination.
 */
// @ts-ignore
function getPaginationProbe(op: Operation): PaginationProbeContext {
  const pagination = op.Extensions?.Pagination;
  if (!pagination) {
    return {
      Type: "",
      CursorKind: "",
      NextCursor: "",
      NextURL: "",
      Results: "",
      HasLimit: false,
    };
  }

  const requestType = op.Request?.Field?.Type;
  const cursorInput = getPaginationInput(pagination, "cursor");
  const cursorField = findPaginationProbeCursorField(
    requestType,
    cursorInput?.Name,
  );
  const rawCursorKind = cursorField?.Type?.Type?.toString() || "";
  const cursorKind = ["string", "integer", "number"].includes(rawCursorKind)
    ? rawCursorKind
    : "";

  const limitInput = getPaginationInput(pagination, "limit");
  const limitField =
    requestType && limitInput?.Name
      ? findPaginationFieldDeep(requestType, limitInput.Name)
      : undefined;

  return {
    Type: pagination.Type?.toString() || "",
    CursorKind: cursorKind,
    NextCursor: pagination.Outputs?.NextCursor || "",
    NextURL: pagination.Outputs?.NextURL || "",
    Results: pagination.Outputs?.Results || "",
    // Next() applies its result-length check whenever a limit input exists
    // unless that field is provably const. Treat an unresolved field as
    // mutable too, matching the Go template's conservative condition.
    HasLimit: Boolean(limitInput && !(limitField && limitField.Const)),
  };
}
registerTemplateFunc("getPaginationProbe", getPaginationProbe);

// @ts-ignore
function paginationProbeHasContinuation(
  probe: PaginationProbeContext,
): boolean {
  return (
    (probe.Type === "cursor" && probe.NextCursor !== "") ||
    (probe.Type === "url" && probe.NextURL !== "")
  );
}
registerTemplateFunc(
  "paginationProbeHasContinuation",
  paginationProbeHasContinuation,
);

/**
 * Return the manual cursor flag for a truthful pagination hint.
 * Offset/limit inputs are intentionally excluded because they do not prove
 * that another page exists.
 */
// @ts-ignore
function getPaginationFlagHint(op: Operation): string {
  const inputs = op.Extensions?.Pagination?.Inputs;
  if (!inputs || inputs.length === 0) return "";

  const cursor = inputs.find(
    (input) => input.Type?.toString() === "cursor" && input.Name,
  );
  if (!cursor?.Name) return "";
  return `--${sanitizeFlagNameWithReserved(cursor.Name)}`;
}
registerTemplateFunc("getPaginationFlagHint", getPaginationFlagHint);

// =============================================================================
// Streaming Detection Helpers
// =============================================================================

/**
 * Check if an operation has a streaming response (SSE EventStream or JSONL/NDJSON).
 * Uses the response body content's SerializationMethod to detect streaming types.
 */
// @ts-ignore
function hasStreamingResponse(op: Operation): boolean {
  if (!op.Response?.Responses) return false;
  for (const resp of op.Response.Responses) {
    if (resp.Error) continue;
    for (const content of resp.Content || []) {
      const sm = content.SerializationMethod?.toString();
      if (sm === "eventstream" || sm === "jsonl") return true;
    }
  }
  return false;
}
registerTemplateFunc("hasStreamingResponse", hasStreamingResponse);

/**
 * Check if an operation's response body can outlive the SDK method call:
 * event streams, JSONL, and raw response-stream bodies are consumed by the
 * caller after return, so an implicit per-operation timeout must not be
 * applied to them (its deferred cancel would kill the body mid-read).
 * Buffered []byte bodies are fully read inside the SDK and stay bounded.
 */
// @ts-ignore
function operationHasDeferredBody(op: Operation): boolean {
  if (hasStreamingResponse(op)) return true;
  if (!op.Response?.Responses) return false;
  for (const resp of op.Response.Responses) {
    if (resp.Error) continue;
    for (const content of resp.Content || []) {
      const typeStr = content.Content?.Type?.Type?.toString();
      if (typeStr === "response-stream") return true;
    }
  }
  return false;
}
registerTemplateFunc("operationHasDeferredBody", operationHasDeferredBody);

/**
 * Get the Go field name of the streaming field on the response type.
 * For SSE operations, this is the EventStream field; for JSONL, the JsonLStream field.
 * Returns "" if no streaming field is found.
 */
// @ts-ignore
function getStreamingFieldName(op: Operation): string {
  const responseType = op.Response?.Type;
  if (!responseType?.Fields) return "";
  for (const field of responseType.Fields) {
    const goName = sanitizeFieldName(field.Name);
    if (RESPONSE_ENVELOPE_FIELDS.has(goName)) continue;
    const typeStr = field.Type?.Type?.toString();
    if (typeStr === "event-stream" || typeStr === "jsonl") {
      return goName;
    }
  }
  return "";
}
registerTemplateFunc("getStreamingFieldName", getStreamingFieldName);

// =============================================================================
// Binary Response Detection Helpers
// =============================================================================

/**
 * Check if an operation has any binary response content (io.ReadCloser or []byte).
 * Binary responses use SerializationMethod "raw" or have Type.Type "response-stream" or "bytes".
 */
// @ts-ignore
function hasBinaryResponse(op: Operation): boolean {
  if (!op.Response?.Responses) return false;
  for (const resp of op.Response.Responses) {
    if (resp.Error) continue;
    for (const content of resp.Content || []) {
      const sm = content.SerializationMethod?.toString();
      if (sm === "raw") return true;
      const typeStr = content.Content?.Type?.Type?.toString();
      if (typeStr === "response-stream" || typeStr === "bytes") return true;
    }
  }
  return false;
}
registerTemplateFunc("hasBinaryResponse", hasBinaryResponse);

/**
 * Check if ALL success response content types are binary (no JSON/text alternatives).
 * When true, --output-file is enforced on TTY. When false (mixed), it's optional.
 */
// @ts-ignore
function isOnlyBinaryResponse(op: Operation): boolean {
  if (!op.Response?.Responses) return false;
  let hasBinary = false;
  let hasNonBinary = false;
  for (const resp of op.Response.Responses) {
    if (resp.Error) continue;
    for (const content of resp.Content || []) {
      const sm = content.SerializationMethod?.toString();
      const typeStr = content.Content?.Type?.Type?.toString();
      if (
        sm === "raw" ||
        typeStr === "response-stream" ||
        typeStr === "bytes"
      ) {
        hasBinary = true;
      } else {
        hasNonBinary = true;
      }
    }
  }
  return hasBinary && !hasNonBinary;
}
registerTemplateFunc("isOnlyBinaryResponse", isOnlyBinaryResponse);

// =============================================================================
// Global Operation Predicate Helpers
// Used by auxiliary templates (e.g., output.go.stmpl) to conditionally generate
// functions that are only needed when at least one operation uses a feature.
// =============================================================================

/**
 * Recursively check if any operation in the SDK tree satisfies a predicate.
 */
function anyOperation(
  sdk: SDK,
  predicate: (op: Operation) => boolean,
): boolean {
  for (const op of sdk.Operations) {
    if (predicate(op)) return true;
  }
  for (const sub of sdk.SubSDKs) {
    if (anyOperation(sub, predicate)) return true;
  }
  return false;
}

// @ts-ignore
function anyOperationHasPagination(): boolean {
  return anyOperation(context.Global.AST.MainSDK, hasPagination);
}
registerTemplateFunc("anyOperationHasPagination", anyOperationHasPagination);

// @ts-ignore
function anyOperationHasStreaming(): boolean {
  return anyOperation(context.Global.AST.MainSDK, hasStreamingResponse);
}
registerTemplateFunc("anyOperationHasStreaming", anyOperationHasStreaming);

// @ts-ignore
function anyOperationHasTransform(): boolean {
  return anyOperation(context.Global.AST.MainSDK, hasResponseTransform);
}
registerTemplateFunc("anyOperationHasTransform", anyOperationHasTransform);

// @ts-ignore
function anyOperationHasBinary(): boolean {
  return anyOperation(context.Global.AST.MainSDK, hasBinaryResponse);
}
registerTemplateFunc("anyOperationHasBinary", anyOperationHasBinary);
