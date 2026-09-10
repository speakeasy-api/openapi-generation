// =============================================================================
// Metadata-Driven Request Building
// Generates FlagMeta arrays, RegisterFlags/BuildRequest call expressions,
// and ValidateMeta startup checks from operation AST.
// =============================================================================

/**
 * Get the metadata variable name for a command.
 * sanitizeCommandName already appends "Cmd", so we just add "Meta".
 */
function templateMetaVarName(op: Operation): string {
  const cmdName = sanitizePrivateFieldName(sanitizeCommandName(op.GetID()));
  return `${cmdName}Meta`;
}
registerTemplateFunc("templateMetaVarName", templateMetaVarName);

/**
 * Generate the package-level metadata variable declaration.
 * Handles non-IsRequestBody commands and multipart IsRequestBody commands.
 * JSON IsRequestBody commands don't use metadata (they use BuildRequestBody with a single flag).
 */
/**
 * Check if an IsRequestBody+JSON operation can be expanded to individual flags.
 * JSON and form bodies with at least one non-const field are always expandable.
 * collectMetadataFromFields handles per-field classification: simple fields get
 * typed flags, complex fields (maps, nested classes) get individual FlagKindJSON.
 */
function isRequestBodyExpandable(op: Operation): boolean {
  if (!op.Request || !op.Request.IsRequestBody) return false;
  const serMethod = op.SerializationMethod?.toString() || "";
  if (serMethod !== "json" && serMethod !== "form") return false;
  const bodyType = op.Request.RequestBody.Type;
  // Only class types (structs) can be expanded into individual field flags.
  // Union, map, array, and scalar body types remain as single JSON flags.
  if (bodyType.Type.toString() !== "class") return false;
  const bodyFields = bodyType.Fields || [];
  if (bodyFields.length === 0) return false;
  return bodyFields.some((f) => !f.Const);
}
registerTemplateFunc("isRequestBodyExpandable", isRequestBodyExpandable);

/**
 * Check if an operation will generate flag metadata.
 * Returns false for operations with scalar request bodies or empty request types.
 */
function operationHasFlagMetadata(op: Operation): boolean {
  return templateFlagMetadataVar(op) !== "";
}
registerTemplateFunc("operationHasFlagMetadata", operationHasFlagMetadata);

// buildOperationMetadataEntries constructs the FlagMeta entry strings for an
// operation — the single source for both the emitted metadata var and any
// validation that must mirror it (e.g. auto-shorthand collision checks).
// Returns null when the operation takes the BuildRequestBody path and emits
// no metadata var at all.
function buildOperationMetadataEntries(
  op: Operation,
): { entries: string[]; hadCandidateFields: boolean } | null {
  if (!op.Request) return null;

  const entries: string[] = [];
  let hadCandidateFields = false;

  if (op.Request.IsRequestBody) {
    if (op.SerializationMethod?.toString() === "multipart") {
      // Multipart: generate metadata from request body fields (file, JSON, primitives)
      const fields = op.Request.RequestBody.Type.Fields || [];
      hadCandidateFields = fields.some((f) => !f.Const);
      collectMultipartMetadata(fields, entries);
    } else if (isRequestBodyExpandable(op)) {
      // Expandable JSON body: generate individual field flags for merge support
      const fields = op.Request.RequestBody.Type.Fields || [];
      hadCandidateFields = fields.some((f) => !f.Const);
      collectMetadataFromFields(fields, "", "", entries);
    } else {
      return null; // Complex JSON IsRequestBody uses BuildRequestBody, no metadata var
    }
  } else {
    const fields = op.Request.Field.Type.Fields;
    hadCandidateFields = fields.some((f: FieldDef) => !f.Const);
    collectMetadataFromFields(fields, "", "", entries);
  }
  return { entries, hadCandidateFields };
}

const flagMetaEntryFieldPathRegex = /FieldPath: "([^"]+)"/;

// operationAutoShorthandOwners maps each auto-assigned single-letter
// shorthand to its owning generated flag, mirroring templateFlagMetadataVar's
// post-pass exactly (same entry traversal, same reservations). With
// nonBodyOnly, owners are restricted to the entries flagutil.NonBodyMeta
// keeps — the subset dispatch commands actually register.
function operationAutoShorthandOwners(
  op: Operation,
  nonBodyOnly: boolean,
): Map<string, string> {
  const built = buildOperationMetadataEntries(op);
  if (!built) return new Map();
  const flagNameRegex = /FlagName: "([^"]+)"/;
  const names: string[] = [];
  const fieldPaths = new Map<string, string>();
  for (const entry of built.entries) {
    const name = entry.match(flagNameRegex)?.[1];
    if (!name) continue;
    names.push(name);
    fieldPaths.set(name, entry.match(flagMetaEntryFieldPathRegex)?.[1] || "");
  }
  const extraReserved = hasPagination(op) ? new Set(["a"]) : undefined;
  const bodyFieldPath = templateBodyFieldPath(op);
  const owners = new Map<string, string>();
  for (const [name, letter] of computeFlagShorthands(names, extraReserved)) {
    if (nonBodyOnly) {
      const fieldPath = fieldPaths.get(name) || "";
      const isBody =
        bodyFieldPath === "" ||
        fieldPath === bodyFieldPath ||
        fieldPath.startsWith(`${bodyFieldPath}.`);
      if (isBody) continue;
    }
    owners.set(letter, name);
  }
  return owners;
}

function templateFlagMetadataVar(op: Operation): string {
  const built = buildOperationMetadataEntries(op);
  if (built === null) return "";

  addImport("internal/flagutil", true);

  const metaVarName = templateMetaVarName(op);
  const { entries, hadCandidateFields } = built;

  // Return empty string only when there were genuinely no candidate fields.
  // When fields existed but were all filtered (e.g., global parameters), we still
  // emit an empty metadata var so the operation uses the BuildRequest path
  // (which creates a zero-valued request struct for the SDK to populate via globals).
  if (entries.length === 0 && !hadCandidateFields) return "";

  if (entries.length === 0) {
    return `var ${metaVarName} = []flagutil.FlagMeta{}`;
  }

  // Post-pass: assign single-letter shorthands to non-conflicting flags.
  // Extract flag names from entries, compute shorthands, then inject Shorthand field.
  const flagNameRegex = /FlagName: "([^"]+)"/;
  const flagNames: string[] = [];
  for (const entry of entries) {
    const match = entry.match(flagNameRegex);
    if (match) flagNames.push(match[1]);
  }

  // "-a" is reserved on paginated commands (used by --all)
  const extraReserved = hasPagination(op) ? new Set(["a"]) : undefined;
  const shorthands = computeFlagShorthands(flagNames, extraReserved);

  // Inject Shorthand field into entries that got a shorthand
  const enrichedEntries = entries.map((entry) => {
    const match = entry.match(flagNameRegex);
    if (!match) return entry;
    const name = match[1];
    const shorthand = shorthands.get(name);
    if (!shorthand) return entry;
    // Insert Shorthand right after FlagName
    return entry.replace(
      `FlagName: "${name}"`,
      `FlagName: "${name}", Shorthand: "${shorthand}"`,
    );
  });

  return `var ${metaVarName} = []flagutil.FlagMeta{\n\t${enrichedEntries.join(
    ",\n\t",
  )},\n}`;
}
registerTemplateFunc("templateFlagMetadataVar", templateFlagMetadataVar);

/**
 * Generate the flag registration call using RegisterFlags.
 * Works for non-IsRequestBody and multipart IsRequestBody commands.
 */
function templateRegisterFlagsCall(op: Operation): string {
  if (!op.Request) return "";

  addImport("internal/flagutil", true);

  const metaVarName = templateMetaVarName(op);
  return `flagutil.RegisterFlags(cmd, ${metaVarName})`;
}
registerTemplateFunc("templateRegisterFlagsCall", templateRegisterFlagsCall);

/**
 * Generate the ValidateMeta call expression for startup validation.
 * Returns an error on invalid metadata to allow clean error reporting.
 */
function templateValidateMetaCall(op: Operation): string {
  if (!op.Request) return "";

  addImport("internal/flagutil", true);
  addImport("fmt");

  const metaVarName = templateMetaVarName(op);
  let requestTypeName: string;
  if (op.Request.IsRequestBody) {
    requestTypeName = sanitizeType(op.Request.RequestBody.Type, false, "");
  } else {
    requestTypeName = sanitizeType(op.Request.Field.Type, false, "");
  }
  const cmdName = sanitizeCLICommand(op.GetID());
  return `if err := flagutil.ValidateMeta[${requestTypeName}](${metaVarName}); err != nil {
        return fmt.Errorf("invalid metadata for ${cmdName}: %w", err)
    }`;
}
registerTemplateFunc("templateValidateMetaCall", templateValidateMetaCall);

/**
 * Get the set of flag names that are registered as persistent global parameter
 * flags on the root command. These should be skipped in per-operation FlagMeta
 * to avoid local flags shadowing the global resolution logic (flag > env > config).
 */
function getGlobalFlagNames(): Set<string> {
  const globals = context.Global.AST.MainSDK.Globals;
  if (!globals?.Fields) return new Set<string>();
  return new Set<string>(
    globals.Fields.map((f: FieldDef) => sanitizeFlagNameWithReserved(f.Name)),
  );
}

/**
 * Recursively collect FlagMeta entries from fields.
 * Walks nested objects, emits union entries, and maps types to FlagKind.
 */
function collectMetadataFromFields(
  fields: FieldDef[],
  flagPrefix: string,
  fieldPathPrefix: string,
  entries: string[],
  group: string = "",
): void {
  const globalFlags = getGlobalFlagNames();

  for (const field of fields) {
    if (field.Const) continue;

    const flagName = flagPrefix
      ? `${flagPrefix}.${sanitizeFlagNameWithReserved(field.Name)}`
      : sanitizeFlagNameWithReserved(field.Name);

    // Skip global parameters — they are registered as persistent flags on the
    // root command and resolved via buildGlobalOptions() in client.go.
    // Only check top-level fields (no flagPrefix) since globals are always
    // top-level on the request struct.
    if (!flagPrefix && globalFlags.has(flagName)) continue;
    const fieldPath = fieldPathPrefix
      ? `${fieldPathPrefix}.${sanitizeFieldName(field.Name)}`
      : sanitizeFieldName(field.Name);

    // Union fields → FlagKindUnion with embedded UnionMeta
    if (field.Type.Type.toString() === "union") {
      const unionMeta = templateUnionMetaEntry(field, flagPrefix);
      const unionParts = [
        `FlagName: "${flagName}"`,
        `FieldPath: "${fieldPath}"`,
        `Kind: flagutil.FlagKindUnion`,
        `Union: ${unionMeta}`,
      ];
      if (group) unionParts.push(`Group: "${escapeGoString(group)}"`);
      entries.push(`{${unionParts.join(", ")}}`);
      continue;
    }

    // Fields wrapped in OptionalNullable (nullable+optional with wrapper) can't be
    // expanded — reflection sees a map, not a struct. Treat as JSON flag.
    if (
      field.Nullable &&
      field.Optional &&
      context.Global.Config.NullableOptionalWrapper
    ) {
      entries.push(
        buildMetaEntryForField(
          field,
          flagName,
          fieldPath,
          "FlagKindJSON",
          group,
        ),
      );
      continue;
    }

    // Multipart body fields in mixed param+body operations → delegate to
    // collectMultipartMetadata for proper file/JSON/primitive classification.
    // The fieldPath prefix ensures metadata entries reference the correct nested
    // path on the request struct (e.g., "Body.File" not just "File").
    if (getInputClassType(field) === "MultipartRequestBody") {
      collectMultipartMetadata(field.Type.Fields || [], entries, fieldPath);
      continue;
    }

    // Nested objects that get expanded → recurse
    if (shouldExpandNestedField(field)) {
      let expandFlagPrefix = flagName;

      // For body fields in mixed param+body operations, flatten the prefix
      // (e.g., --expression instead of --body.expression) when safe.
      const isBodyWrapper = field.Annotations?.Has("request");
      if (isBodyWrapper) {
        expandFlagPrefix = getBodyFlagPrefix(field, fields, flagPrefix);
      }

      // Set group for nested fields: body wrapper fields don't create a group
      // (their children are top-level body fields), but non-wrapper expanded
      // objects do (e.g., Address, Profile).
      const childGroup = isBodyWrapper
        ? ""
        : sanitizeFieldName(field.Name).replace(/([a-z])([A-Z])/g, "$1 $2");

      collectMetadataFromFields(
        field.Type.Fields || [],
        expandFlagPrefix,
        fieldPath,
        entries,
        childGroup || group,
      );
      continue;
    }

    // Non-expanded JSON/form body fields (class types not expanded)
    const classType = getInputClassType(field);
    if (classType === "JSONRequestBody" || classType === "FormRequestBody") {
      entries.push(
        buildMetaEntryForField(
          field,
          flagName,
          fieldPath,
          "FlagKindJSON",
          group,
        ),
      );
      continue;
    }

    // Complex types that need JSON flag handling
    const typeStr = field.Type?.Type?.toString() || "";

    // any/map/bigint/decimal → FlagKindJSON (runtime buildJSONField handles via reflection)
    if (
      typeStr === "any" ||
      typeStr === "map" ||
      typeStr === "bigint" ||
      typeStr === "decimal"
    ) {
      entries.push(
        buildMetaEntryForField(
          field,
          flagName,
          fieldPath,
          "FlagKindJSON",
          group,
        ),
      );
      continue;
    }

    // Class types → FlagKindJSON (empty or non-expandable classes both use JSON flag)
    if (typeStr === "class") {
      entries.push(
        buildMetaEntryForField(
          field,
          flagName,
          fieldPath,
          "FlagKindJSON",
          group,
        ),
      );
      continue;
    }

    // Regular leaf fields (primitives, enums, arrays)
    entries.push(
      buildMetaEntryForField(field, flagName, fieldPath, undefined, group),
    );
  }
}

/**
 * Collect FlagMeta entries for multipart request body fields.
 * Classifies fields as file uploads, JSON objects, or primitives.
 */
function collectMultipartMetadata(
  fields: FieldDef[],
  entries: string[],
  fieldPathPrefix: string = "",
): void {
  const globalFlags = getGlobalFlagNames();

  for (const field of fields) {
    if (field.Const) continue;

    const flagName = sanitizeFlagNameWithReserved(field.Name);
    const fieldPath = fieldPathPrefix
      ? `${fieldPathPrefix}.${sanitizeFieldName(field.Name)}`
      : sanitizeFieldName(field.Name);

    // Skip global parameters (same as collectMetadataFromFields)
    if (globalFlags.has(flagName)) continue;

    // File upload fields → FlagKindFile or FlagKindFileArray
    // isMultipartFileField handles single file class; also check array-of-file-class
    const fieldTypeStr = field.Type?.Type?.toString() || "";
    const isFileArray =
      fieldTypeStr === "array" &&
      field.Type?.ItemType?.Type?.toString() === "class" &&
      (() => {
        const ann = field.Annotations?.Get("multipartForm");
        return Boolean(ann && isMultipartFormAnnotation(ann) && ann.File);
      })();
    if (isMultipartFileField(field) || isFileArray) {
      const isArray = fieldTypeStr === "array";
      const desc =
        field.Comments?.Description ||
        (isArray
          ? "Comma-separated paths to files to upload"
          : "Path to file to upload");
      const kind = isArray ? "FlagKindFileArray" : "FlagKindFile";
      const parts: string[] = [
        `FlagName: "${flagName}"`,
        `FieldPath: "${fieldPath}"`,
        `Kind: flagutil.${kind}`,
      ];
      if (field.Optional) {
        parts.push(`Optional: true`);
      } else {
        parts.push(`Required: true`);
      }
      parts.push(
        `Description: "${escapeGoString(desc)}${
          field.Optional ? "" : " [required]"
        }"`,
      );
      entries.push(`{${parts.join(", ")}}`);
      continue;
    }

    // JSON object fields (class but not file) → FlagKindJSON
    if (isMultipartJSONField(field)) {
      const desc =
        field.Comments?.Description ||
        `${caser().ToPascal(field.Name)} as JSON`;
      const parts: string[] = [
        `FlagName: "${flagName}"`,
        `FieldPath: "${fieldPath}"`,
        `Kind: flagutil.FlagKindJSON`,
      ];
      if (field.Optional) {
        parts.push(`Optional: true`);
      } else {
        parts.push(`Required: true`);
      }
      parts.push(`Annotations: ${templateAnnotations(field, true)}`);
      parts.push(
        `Description: "${escapeGoString(desc)}${
          field.Optional ? "" : " [required]"
        }"`,
      );
      entries.push(`{${parts.join(", ")}}`);
      continue;
    }

    // Complex non-class types in multipart → FlagKindJSON
    const ft = field.Type?.Type?.toString() || "";
    if (
      ft === "any" ||
      ft === "map" ||
      ft === "bigint" ||
      ft === "decimal" ||
      ft === "union"
    ) {
      entries.push(
        buildMetaEntryForField(field, flagName, fieldPath, "FlagKindJSON"),
      );
      continue;
    }
    if (ft === "array") {
      const itemType = field.Type?.ItemType?.Type?.toString() || "";
      if (itemType === "class" || itemType === "map" || itemType === "union") {
        entries.push(
          buildMetaEntryForField(field, flagName, fieldPath, "FlagKindJSON"),
        );
        continue;
      }
    }

    // Primitive fields → standard metadata entry
    entries.push(buildMetaEntryForField(field, flagName, fieldPath));
  }
}

/**
 * Build a single FlagMeta Go literal for a field.
 */
function buildMetaEntryForField(
  field: FieldDef,
  flagName: string,
  fieldPath: string,
  kindOverride?: string,
  group?: string,
): string {
  const typeDef = field.Type;
  let kind: string;

  if (kindOverride) {
    kind = kindOverride;
  } else if (isEnumType(typeDef)) {
    kind = isIntBackedEnum(typeDef) ? "FlagKindIntEnum" : "FlagKindEnum";
  } else if (isArrayType(typeDef)) {
    // Only string/enum arrays use FlagKindStringArray (cobra StringArray gives []string).
    // Typed arrays (int, float, bool) use FlagKindJSON since the runtime can't convert
    // string array elements to typed values via reflection.
    const itemTypeStr = typeDef.ItemType?.Type?.toString() || "string";
    if (itemTypeStr === "string" || itemTypeStr === "enum") {
      kind = "FlagKindStringArray";
    } else {
      kind = "FlagKindJSON";
    }
  } else if (typeDef.Type.toString() === "date-time") {
    kind = "FlagKindDateTime";
  } else if (typeDef.Type.toString() === "bytes") {
    kind = "FlagKindBytes";
  } else {
    switch (typeDef.Type.toString()) {
      case "string":
        kind = "FlagKindString";
        break;
      case "date":
        kind = "FlagKindDate";
        break;
      case "boolean":
        kind = "FlagKindBool";
        break;
      case "integer":
      case "int32":
        kind = "FlagKindInt64";
        break;
      case "number":
      case "float32":
        kind = "FlagKindFloat64";
        break;
      default:
        kind = "FlagKindString";
        break;
    }
  }

  const parts: string[] = [
    `FlagName: "${flagName}"`,
    `FieldPath: "${fieldPath}"`,
    `Kind: flagutil.${kind}`,
  ];

  // Behavior flags: Optional encompasses both optional-without-default and optional-with-default.
  // Nullable fields are also treated as optional since null is a valid value —
  // not passing the flag leaves the pointer nil, which serializes as JSON null.
  if (field.Optional || field.Default || field.Nullable) {
    parts.push(`Optional: true`);
  }
  if (!field.Optional && !field.Default && !field.Nullable) {
    parts.push(`Required: true`);
  }
  if (field.Default) {
    parts.push(`HasDefault: true`);
    // Add the actual default value for flag registration
    const defaultValue = field.Default.Value;
    switch (kind) {
      case "FlagKindString":
      case "FlagKindEnum":
      case "FlagKindIntEnum":
      case "FlagKindDateTime":
      case "FlagKindDate":
      case "FlagKindJSON":
        parts.push(`DefaultStr: "${escapeGoString(String(defaultValue))}"`);
        break;
      case "FlagKindBool":
        if (defaultValue === true || defaultValue === "true") {
          parts.push(`DefaultBool: true`);
        }
        break;
      case "FlagKindInt64":
        if (Number(defaultValue) !== 0) {
          parts.push(`DefaultInt: ${Number(defaultValue)}`);
        }
        break;
      case "FlagKindFloat64":
        if (Number(defaultValue) !== 0) {
          parts.push(`DefaultFloat: ${Number(defaultValue)}`);
        }
        break;
    }
  }

  // Schema-declared bounds the CLI can check trivially before sending: a
  // string minLength (so an explicitly empty value is rejected when the
  // schema forbids it) and numeric minimum/maximum. Anything richer stays
  // with the server, which is authoritative. Enforcement is strict only on
  // manifest-declared commands and under explicit agent mode; ordinary
  // operations surface violations as warnings and let the server decide.
  const validations = typeDef.Validations;
  if (validations) {
    // Pointer-valued in the AST: coerce through Number and skip when unset.
    const numeric = (v: any): number | undefined => {
      if (v == null) return undefined;
      const n = Number(v);
      return Number.isFinite(n) ? n : undefined;
    };
    const minLength = numeric(validations.MinLength);
    if (kind === "FlagKindString" && minLength !== undefined && minLength > 0) {
      parts.push(`MinLength: ${minLength}`);
    }
    if (kind === "FlagKindInt64" || kind === "FlagKindFloat64") {
      const minimum = numeric(validations.Minimum);
      const maximum = numeric(validations.Maximum);
      if (minimum !== undefined) {
        parts.push(`HasMinimum: true`, `Minimum: ${minimum}`);
      }
      if (maximum !== undefined) {
        parts.push(`HasMaximum: true`, `Maximum: ${maximum}`);
      }
    }
  }

  // Enum values for validation
  if (
    (kind === "FlagKindEnum" || kind === "FlagKindIntEnum") &&
    typeDef.Enum?.Values
  ) {
    const enumValuesStr = typeDef.Enum.Values.map(
      (v: string) => `"${escapeGoString(v)}"`,
    ).join(", ");
    parts.push(`EnumValues: []string{${enumValuesStr}}`);
  }

  // Annotations for JSON fields
  if (kind === "FlagKindJSON") {
    parts.push(`Annotations: ${templateAnnotations(field, true)}`);
  }

  // Description for registration
  const description = getFlagDescription(field);
  parts.push(`Description: "${escapeGoString(description)}"`);

  // Display group for --help organization
  if (group) {
    parts.push(`Group: "${escapeGoString(group)}"`);
  }

  return `{${parts.join(", ")}}`;
}

/**
 * Find the body field path on a non-IsRequestBody request struct.
 * The body field is the one with a "request" annotation (mediaType=application/json).
 * Returns empty string if no body field found (params-only operation).
 */
function getBodyFieldPath(op: Operation): string {
  if (!op.Request || op.Request.IsRequestBody) return "";
  const fields = op.Request.Field.Type.Fields || [];
  for (const field of fields) {
    if (field.Annotations?.Has("request")) {
      return sanitizeFieldName(field.Name);
    }
  }
  return ""; // No body field (params-only operation)
}

/**
 * Return the body field path used to filter interactive prompt metadata.
 * IsRequestBody operations use the entire request struct as their body.
 */
function templateBodyFieldPath(op: Operation): string {
  if (!op.Request || op.Request.IsRequestBody) return "";
  return getBodyFieldPath(op);
}
registerTemplateFunc("templateBodyFieldPath", templateBodyFieldPath);

/**
 * Flag name of the metadata entry that maps to the whole body field of a mixed
 * param+body operation — a body that is not expanded into field flags (a
 * union, JSON object, map, or any), e.g. "body-param" for a body named "body".
 * Returns "" when the body is expanded, multipart, or absent. Side-effect free
 * (no addImport), so it is safe to call while other files render.
 */
function wholeBodyFlagName(op: Operation): string {
  if (!op.Request || op.Request.IsRequestBody) return "";
  const bodyFieldPath = getBodyFieldPath(op);
  if (!bodyFieldPath) return "";
  const entries: string[] = [];
  collectMetadataFromFields(
    op.Request.Field.Type.Fields || [],
    "",
    "",
    entries,
  );
  // Entries are Go literals that always open with the top-level FlagName and
  // FieldPath (buildMetaEntryForField / the union entry); match that prefix so
  // nested variant fields cannot be mistaken for the body field.
  const re = new RegExp(
    `^\\{FlagName: "([^"]+)", FieldPath: "${bodyFieldPath}"[,}]`,
  );
  for (const entry of entries) {
    const match = entry.match(re);
    if (match) return match[1];
  }
  return "";
}

/**
 * Check if a non-IsRequestBody operation has a multipart body field.
 * Used to detect mixed param+body operations where the body is multipart/form-data.
 */
function isMultipartMixedOp(op: Operation): boolean {
  if (!op.Request || op.Request.IsRequestBody) return false;
  const fields = op.Request.Field.Type.Fields || [];
  return fields.some(
    (f: FieldDef) => getInputClassType(f) === "MultipartRequestBody",
  );
}

/**
 * Check if an operation needs a --body whole-JSON flag.
 * True when the operation uses metadata flags AND has a request body AND is not multipart.
 * Multipart bodies can't be expressed as a single JSON blob (they contain files).
 */
function templateHasBodyFlag(op: Operation): boolean {
  if (!op.Request) return false;
  if (op.SerializationMethod?.toString() === "multipart") return false;
  if (isMultipartMixedOp(op)) return false; // multipart mixed param+body ops
  if (op.Request.IsRequestBody) return true; // expanded IsRequestBody ops
  return getBodyFieldPath(op) !== ""; // mixed param+body ops
}
registerTemplateFunc("templateHasBodyFlag", templateHasBodyFlag);

/**
 * Generate the description for the --body flag.
 */
function templateBodyFlagDescription(op: Operation): string {
  let description =
    "Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF.";
  if (hasBodySchemaForOp(op)) {
    description += " Use --schema to print the exact JSON Schema.";
  }
  return escapeGoString(description);
}
registerTemplateFunc(
  "templateBodyFlagDescription",
  templateBodyFlagDescription,
);

/**
 * Generate the BuildRequest call expression (no `return` prefix).
 * Handles both non-IsRequestBody and multipart IsRequestBody commands.
 * Passes bodyFieldPath to enable stdin merge (flags override stdin body),
 * and bodyFlagName for --body whole-JSON flag support.
 */
function templateBuildRequestCall(op: Operation): string {
  if (!op.Request) return "";

  addImport("internal/flagutil", true);

  const metaVarName = templateMetaVarName(op);
  let requestTypeName: string;
  let bodyFieldPath: string;
  let bodyFlagName: string;
  if (op.Request.IsRequestBody) {
    // Multipart or expanded body: BuildRequest with request body type
    requestTypeName = sanitizeType(op.Request.RequestBody.Type, false, "");
    bodyFieldPath = '""'; // entire struct is body
    bodyFlagName = templateHasBodyFlag(op) ? '"body"' : '""';
  } else {
    requestTypeName = sanitizeType(op.Request.Field.Type, false, "");
    const bfp = getBodyFieldPath(op);
    bodyFieldPath = `"${bfp}"`;
    // Only pass bodyFlagName when there IS a body field AND it's not multipart
    // (multipart bodies don't support --body JSON or stdin merge)
    bodyFlagName = bfp !== "" && !isMultipartMixedOp(op) ? '"body"' : '""';
  }
  return `flagutil.BuildRequest[${requestTypeName}](cmd, ${metaVarName}, ${bodyFieldPath}, ${bodyFlagName})`;
}
registerTemplateFunc("templateBuildRequestCall", templateBuildRequestCall);

/**
 * Generate the BuildRequestBody call expression (no `return` prefix).
 * Used for IsRequestBody+JSON commands (single JSON blob flag).
 */
function templateBuildRequestBodyCall(op: Operation): string {
  if (!op.Request) return "";

  addImport("internal/flagutil", true);

  if (op.Request.IsRequestBody) {
    const bodyField = op.Request.RequestBody;
    let typeName: string;
    // When both nullable+optional and wrapper is enabled, the Go SDK uses
    // OptionalNullable[T] as the parameter type instead of *T
    if (
      bodyField.Nullable &&
      bodyField.Optional &&
      context.Global.Config.NullableOptionalWrapper
    ) {
      const baseType = sanitizeType(bodyField.Type, false, "");
      addImport("internal/sdk/optionalnullable", true);
      typeName = `optionalnullable.OptionalNullable[${baseType}]`;
    } else {
      typeName = sanitizeType(bodyField.Type, false, "");
    }
    const flagName = getRequestBodyFlagName(bodyField);
    const annotations = templateAnnotations(bodyField, true);
    const isRequired = !bodyField.Optional && !bodyField.Nullable;
    return `flagutil.BuildRequestBody[${typeName}](cmd, "${flagName}", ${annotations}, ${isRequired})`;
  }

  // Non-IsRequestBody with no expandable fields (scalar body) — use "body" flag
  const typeName = sanitizeType(op.Request.Field.Type, false, "");
  return "flagutil.BuildRequestBody[" + typeName + '](cmd, "body", ``, true)';
}
registerTemplateFunc(
  "templateBuildRequestBodyCall",
  templateBuildRequestBodyCall,
);

// Whether the single whole-body flag used by BuildRequestBody is required.
// This mirrors the boolean passed to BuildRequestBody so prompt discovery and
// request construction share the same contract.
function templateRequestBodyRequired(op: Operation): boolean {
  if (!op.Request) return false;
  if (op.Request.IsRequestBody) {
    const body = op.Request.RequestBody;
    return !body.Optional && !body.Nullable;
  }
  return true;
}
registerTemplateFunc(
  "templateRequestBodyRequired",
  templateRequestBodyRequired,
);

/**
 * Generate the UnionMeta Go literal for a union field.
 * Used by collectMetadataFromFields to embed in FlagMeta entries.
 */
/**
 * Return a compact type label for use inside field signatures.
 * E.g. "string", "integer", "string[]", "object".
 */
function briefTypeLabel(typeDef: TypeDef): string {
  const t = typeDef.Type?.toString() || "";
  switch (t) {
    case "string":
      return "string";
    case "boolean":
      return "boolean";
    case "integer":
    case "int32":
      return "integer";
    case "number":
    case "float32":
      return "number";
    case "date":
    case "date-time":
      return t;
    case "enum":
      return "string";
    case "array":
    case "set": {
      const itemLabel = typeDef.ItemType
        ? briefTypeLabel(typeDef.ItemType)
        : "any";
      return `${itemLabel}[]`;
    }
    case "map":
      return "object";
    case "class":
      return "object";
    case "union": {
      const variants = typeDef.AssociatedTypes || [];
      if (variants.length > 0) {
        const labels = [
          ...new Set(variants.map((v: TypeDef) => briefTypeLabel(v))),
        ];
        return labels.slice(0, 3).join(" | ");
      }
      return "value";
    }
    default:
      return "value";
  }
}

/**
 * Describe a type for user-facing help text in union descriptions.
 * Returns e.g. "{ translate: string[] }" for class types,
 * "array of { type: string, ranges: object[] }" for arrays, or "boolean" for primitives.
 * Deduplication of identical descriptions is handled by the caller.
 */
function describeTypeForHelp(
  typeDef: TypeDef,
  visited: Set<string> = new Set(),
): string {
  const typeStr = typeDef.Type?.toString() || "";

  if (typeDef.IsCustomType()) {
    const id = getUniqueID(typeDef);
    if (visited.has(id)) {
      return typeDef.Name || typeStr || "any";
    }
    // Copy before mutating to isolate recursive branches.
    visited = new Set(visited);
    visited.add(id);
  }

  // Class types: show the JSON shape a caller must supply. Required and
  // defaulted fields come first (they distinguish union variants and are what
  // a caller provides or relies on); other optional fields collapse into "...".
  if (typeStr === "class" && typeDef.Fields && typeDef.Fields.length > 0) {
    const visibleFields = typeDef.Fields.filter((f: FieldDef) => !f.Const);
    const keyFields = visibleFields.filter(
      (f: FieldDef) =>
        !f.Optional ||
        (f.Default?.Value !== undefined && f.Default?.Value !== null),
    );
    const preferred = keyFields.length > 0 ? keyFields : visibleFields;
    if (preferred.length > 0) {
      const shown = preferred.slice(0, 4).map((f: FieldDef) => {
        const def = f.Default?.Value;
        const defSuffix =
          def !== undefined && def !== null ? ` (default: ${def})` : "";
        return `"${originalFieldName(f)}": ${briefTypeLabel(
          f.Type,
        )}${defSuffix}`;
      });
      const truncated =
        preferred.length > 4 || preferred.length < visibleFields.length;
      const suffix = truncated ? ", ..." : "";
      return `{ ${shown.join(", ")}${suffix} }`;
    }
    return "object";
  }

  // Array types: describe item type for clarity (e.g. "array of strings")
  if (typeStr === "array" || typeStr === "set") {
    const itemType = typeDef.ItemType;
    if (itemType) {
      const itemDesc = describeTypeForHelp(itemType, visited);
      return `array of ${itemDesc}`;
    }
    return "array";
  }

  // Primitive / well-known types: use the type name directly
  switch (typeStr) {
    case "string":
      return "string";
    case "boolean":
      return "boolean";
    case "integer":
    case "int32":
      return "integer";
    case "number":
    case "float32":
      return "number";
    case "date":
    case "date-time":
      return typeStr;
    case "enum":
      return typeDef.Name || "enum";
    case "map":
      return "object";
    case "any":
      return "any";
    case "union":
      // Nested union: summarise its variants
      if (typeDef.AssociatedTypes && typeDef.AssociatedTypes.length > 0) {
        const inner = [
          ...new Set(
            typeDef.AssociatedTypes.map((t: TypeDef) =>
              describeTypeForHelp(t, visited),
            ),
          ),
        ];
        return inner.join(" | ");
      }
      return "value";
  }

  // Named class without fields (opaque object)
  if (typeStr === "class") {
    return typeDef.Name || "object";
  }

  // Final fallback: prefer a meaningful name over "object"
  return typeDef.Name || "object";
}

function templateUnionMetaEntry(field: FieldDef, flagPrefix: string): string {
  const typeDef = field.Type;
  const flagName = flagPrefix
    ? `${flagPrefix}.${sanitizeFlagNameWithReserved(field.Name)}`
    : sanitizeFlagNameWithReserved(field.Name);

  const discriminated = isDiscriminatedUnion(typeDef);

  const parts: string[] = [];
  parts.push(`Discriminated: ${discriminated}`);

  if (discriminated) {
    const discriminatorKey = sanitizeFieldName(getDiscriminatorKey(typeDef));
    parts.push(`DiscriminatorKey: "${discriminatorKey}"`);
  }

  if (field.Optional) {
    parts.push(`Optional: true`);
  }

  if (discriminated) {
    const variants = getUnionVariants(typeDef);
    const variantDescs = variants
      .map((v: any) => `${v.name}: ${describeTypeForHelp(v.type)}`)
      .join(", ");
    parts.push(
      `TypeDescription: "${escapeGoString(
        `JSON value (variants: ${variantDescs})`,
      )}"`,
    );

    const variantEntries = variants.map((variant: any) =>
      templateVariantMetaEntry(variant, typeDef, flagName),
    );
    parts.push(
      `Variants: []flagutil.UnionVariantMeta{\n\t\t${variantEntries.join(
        ",\n\t\t",
      )},\n\t}`,
    );
  } else {
    // Non-discriminated: top-level JSON only — describe each variant by its fields.
    // Deduplicate so identical descriptions (e.g. two array types) aren't repeated.
    const uniqueDescs = [
      ...new Set(
        typeDef.AssociatedTypes.map((t: TypeDef) => describeTypeForHelp(t)),
      ),
    ];
    const variantDescs = uniqueDescs.join(" | ");

    // Zero-config variant defaulting: distinguishing required keys select the
    // variant; when exactly one such key carries a schema default, a body that
    // names no selector gets that default merged in at parse time.
    const variantFieldSets = typeDef.AssociatedTypes.filter(
      (t: TypeDef) => t.Type.toString() === "class",
    ).map((t: TypeDef) =>
      (t.Fields || []).filter(
        (f: FieldDef) =>
          !f.Const &&
          (!f.Optional ||
            (f.Default?.Value !== undefined && f.Default?.Value !== null)),
      ),
    );
    let defaultsSuffix = "";
    if (variantFieldSets.length > 1) {
      const keyCounts = new Map<string, number>();
      for (const fields of variantFieldSets) {
        for (const f of fields) {
          const k = originalFieldName(f);
          keyCounts.set(k, (keyCounts.get(k) || 0) + 1);
        }
      }
      const distinguishing: string[] = [];
      const selectorDefaults: { key: string; value: any }[] = [];
      for (const fields of variantFieldSets) {
        for (const f of fields) {
          const k = originalFieldName(f);
          if (keyCounts.get(k) === variantFieldSets.length) continue; // shared key
          if (!distinguishing.includes(k)) {
            distinguishing.push(k);
            const def = f.Default?.Value;
            if (def !== undefined && def !== null) {
              selectorDefaults.push({ key: k, value: def });
            }
          }
        }
      }
      if (distinguishing.length > 0 && selectorDefaults.length === 1) {
        const defaultObj: Record<string, any> = {
          [selectorDefaults[0].key]: selectorDefaults[0].value,
        };
        const defaultJSON = JSON.stringify(defaultObj);
        parts.push(
          `VariantKeys: []string{${distinguishing
            .map((k) => `"${escapeGoString(k)}"`)
            .join(", ")}}`,
        );
        parts.push(`DefaultJSON: "${escapeGoString(defaultJSON)}"`);
        defaultsSuffix = `; default when neither is named: ${defaultJSON}`;
      }
    }

    parts.push(
      `TypeDescription: "${escapeGoString(
        `JSON value (one of: ${variantDescs}${defaultsSuffix})`,
      )}"`,
    );
  }

  return `&flagutil.UnionMeta{${parts.join(", ")}}`;
}

/**
 * Generate a single UnionVariantMeta Go literal.
 */
function templateVariantMetaEntry(
  variant: any,
  unionTypeDef: TypeDef,
  parentFlagName: string,
): string {
  const variantFlagName = `${parentFlagName}.${variant.flagName}`;
  // Use type Name directly for descriptions to avoid triggering addImport side effects.
  // FieldName must match the Go struct field, which uses sanitizeFieldName(sanitizeUnionTypeName(type))
  // — for named types, that's just sanitizeFieldName(type.Name).
  const variantTypeName = variant.type.Name || sanitizeFieldName(variant.name);
  const variantFieldName = sanitizeFieldName(variantTypeName);
  const expand = canExpandVariant(variant.type);
  const discriminatorKey = getDiscriminatorKey(unionTypeDef);

  const parts: string[] = [];
  parts.push(`DiscriminatorValue: "${variant.name}"`);
  parts.push(`FlagName: "${variantFlagName}"`);
  parts.push(`FieldName: "${variantFieldName}"`);
  parts.push(`CanExpand: ${expand}`);
  parts.push(
    `Description: "${escapeGoString(`${variantTypeName} variant as JSON`)}"`,
  );

  if (expand) {
    const fieldEntries: string[] = [];
    const variantFields = variant.type.Fields || [];
    for (const vf of variantFields) {
      if (vf.Const) continue;
      if (vf.Name === discriminatorKey) continue;

      const vfFlagName = `${variantFlagName}.${sanitizeFlagNameWithReserved(
        vf.Name,
      )}`;
      const vfFieldPath = sanitizeFieldName(vf.Name);
      fieldEntries.push(buildMetaEntryForField(vf, vfFlagName, vfFieldPath));
    }

    if (fieldEntries.length > 0) {
      parts.push(
        `Fields: []flagutil.FlagMeta{\n\t\t\t${fieldEntries.join(
          ",\n\t\t\t",
        )},\n\t\t}`,
      );
    }
  }

  return `{${parts.join(", ")}}`;
}
