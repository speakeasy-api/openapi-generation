/**
 * Flag generation utilities for CLI code generation.
 * Implements naming policy, nested object flattening, and flag access patterns.
 */

// Reserved words that collide with Cobra or a hardcoded root, operation, or
// intent flag. Conditionally registered names (pagination's --all, binary
// output's --output-file, artifact/async flags) are reserved unconditionally:
// a field-derived flag must never depend on which features an operation
// happens to enable to know its own name.
const reservedFlagNames = [
  "help",
  "version",
  "completion",
  "config",
  "debug",
  "verbose",
  "quiet",
  "output",
  "format",
  "output-format",
  "color",
  "jq",
  "raw-output",
  "body",
  "schema",
  "usage",
  "interactive",
  "no-interactive",
  "dry-run",
  "agent-mode",
  "server",
  "server-url",
  "header",
  "include-headers",
  "timeout",
  "no-retries",
  "retry-config",
  "retry-connection-errors",
  "retry-max-elapsed-time",
  "output-file",
  "output-b64",
  "all",
  "max-pages",
  "out",
  "raw-response",
  "async",
  "poll-interval",
  "poll-timeout",
];

/**
 * Check if a flag name is reserved.
 */
function isReservedFlagName(name: string): boolean {
  return reservedFlagNames.includes(name.toLowerCase());
}
registerTemplateFunc("isReservedFlagName", isReservedFlagName);

/**
 * Sanitize a flag name with reserved word handling.
 * Follows the naming policy from the plan:
 * - kebab-case for all flags
 * - Reserved words get "-param" suffix
 * - Handles camelCase and snake_case input
 */
function sanitizeFlagNameWithReserved(name: string): string {
  const kebabName = sanitizeFlagName(name);
  if (isReservedFlagName(kebabName)) {
    return `${kebabName}-param`;
  }
  return kebabName;
}
registerTemplateFunc(
  "sanitizeFlagNameWithReserved",
  sanitizeFlagNameWithReserved,
);

/**
 * Build a nested flag name with dot notation.
 * Example: buildNestedFlagName("user", "name") => "user.name"
 */
function buildNestedFlagName(prefix: string, fieldName: string): string {
  const sanitizedField = sanitizeFlagNameWithReserved(fieldName);
  if (prefix === "") {
    return sanitizedField;
  }
  return `${prefix}.${sanitizedField}`;
}
registerTemplateFunc("buildNestedFlagName", buildNestedFlagName);

/**
 * Get the maximum depth for nested object flag expansion.
 * Per the plan, default is 3 levels (which covers most real-world cases).
 */
function getMaxFlagDepth(): number {
  return 3;
}
registerTemplateFunc("getMaxFlagDepth", getMaxFlagDepth);

/**
 * Check if a type is a primitive that can be directly represented as a flag.
 */
function isPrimitiveType(typeDef: TypeDef): boolean {
  const primitiveTypes = [
    "string",
    "boolean",
    "integer",
    "int32",
    "number",
    "float32",
    "date",
    "date-time",
  ];
  return primitiveTypes.includes(typeDef.Type.toString());
}
registerTemplateFunc("isPrimitiveType", isPrimitiveType);

/**
 * Check if a type is an enum.
 */
function isEnumType(typeDef: TypeDef): boolean {
  return typeDef.Type.toString() === "enum";
}
registerTemplateFunc("isEnumType", isEnumType);

/**
 * Check if an enum has an integer backing type (int64, int32, etc).
 * This is important for proper type conversion in generated code.
 */
function isIntBackedEnum(typeDef: TypeDef): boolean {
  if (!isEnumType(typeDef) || !typeDef.Enum?.Type) {
    return false;
  }
  const backingType = typeDef.Enum.Type.Type.toString();
  return (
    backingType === "integer" ||
    backingType === "int64" ||
    backingType === "int32"
  );
}
registerTemplateFunc("isIntBackedEnum", isIntBackedEnum);

/**
 * Get the backing type for an enum (string, integer, etc).
 */
function getEnumBackingType(typeDef: TypeDef): string {
  if (!isEnumType(typeDef) || !typeDef.Enum?.Type) {
    return "string";
  }
  return typeDef.Enum.Type.Type.toString();
}
registerTemplateFunc("getEnumBackingType", getEnumBackingType);

/**
 * Check if a type is an array.
 */
function isArrayType(typeDef: TypeDef): boolean {
  return (
    typeDef.Type.toString() === "array" || typeDef.Type.toString() === "set"
  );
}
registerTemplateFunc("isArrayType", isArrayType);

/**
 * Get the Cobra flag type for a field.
 * Enhanced to handle arrays and enums.
 */
function getCobraFlagType(field: FieldDef): string {
  const typeDef = field.Type;

  // Arrays use StringArray
  if (isArrayType(typeDef)) {
    const itemType = typeDef.ItemType?.Type.toString() || "string";
    switch (itemType) {
      case "string":
      case "enum":
        return "StringArray";
      case "integer":
      case "int32":
        return "Int64Slice";
      default:
        // For complex array items, use StringArray and parse as JSON
        return "StringArray";
    }
  }

  // Enums are strings with validation
  if (isEnumType(typeDef)) {
    return "String";
  }

  switch (typeDef.Type.toString()) {
    case "string":
    case "date":
    case "date-time":
      return "String";
    case "boolean":
      return "Bool";
    case "integer":
    case "int32":
      return "Int64";
    case "number":
    case "float32":
      return "Float64";
    case "class":
      // Complex objects use String (JSON/YAML input)
      return "String";
    default:
      return "String";
  }
}
registerTemplateFunc("getCobraFlagType", getCobraFlagType);

/**
 * Get the default value for a flag.
 * Enhanced to handle enums and arrays.
 */
function getFlagDefaultValue(field: FieldDef): string {
  const typeDef = field.Type;

  // Arrays default to empty
  if (isArrayType(typeDef)) {
    return "[]string{}";
  }

  // Handle explicit defaults
  if (field.Default) {
    switch (typeDef.Type.toString()) {
      case "string":
      case "enum":
        // Escape the default value to handle special characters and template syntax
        return `"${escapeGoString(String(field.Default.Value))}"`;
      case "boolean":
        return field.Default.Value.toString();
      case "integer":
      case "int32":
      case "number":
      case "float32":
        return field.Default.Value.toString();
      default:
        return '""';
    }
  }

  // Default values by type
  switch (typeDef.Type.toString()) {
    case "string":
    case "date":
    case "date-time":
    case "class":
    case "enum":
      return '""';
    case "boolean":
      return "false";
    case "integer":
    case "int32":
      return "0";
    case "number":
    case "float32":
      return "0.0";
    default:
      return '""';
  }
}
registerTemplateFunc("getFlagDefaultValue", getFlagDefaultValue);

/**
 * Get enum values for help text.
 * Returns a formatted string of all options.
 */
function getEnumHelpText(typeDef: TypeDef): string {
  if (!isEnumType(typeDef) || !typeDef.Enum?.Values) {
    return "";
  }

  const values = typeDef.Enum.Values;
  if (values.length === 0) {
    return "";
  }

  return `options: ${values.join(", ")}`;
}
registerTemplateFunc("getEnumHelpText", getEnumHelpText);

/**
 * Get a type hint string for a field type, used as fallback description.
 */
function getTypeHint(field: FieldDef): string {
  const typeDef = field.Type;
  if (isEnumType(typeDef)) {
    const enumHelp = getEnumHelpText(typeDef);
    return enumHelp || "enum value";
  }
  if (isArrayType(typeDef)) {
    return "list of values";
  }
  switch (typeDef.Type.toString()) {
    case "string":
      return "string value";
    case "boolean":
      return "boolean flag";
    case "integer":
    case "int32":
      return "integer value";
    case "number":
    case "float32":
      return "number value";
    case "date":
    case "date-time":
      return "date/time value";
    case "class":
      return "JSON object";
    case "bytes":
      return "binary data (file:<path> or b64:<base64>)";
    default:
      return "value";
  }
}

/**
 * Reserved shorthand letters used by built-in persistent flags.
 * These cannot be used by spec-derived flags because Cobra errors
 * when a local shorthand conflicts with an inherited persistent shorthand.
 */
const reservedShorthandLetters = new Set(["o", "H", "q", "d", "h"]);

/**
 * Compute non-conflicting single-letter shorthands for a set of flag names.
 * Each flag's candidate is the first character of its kebab name.
 * If exactly one flag claims a letter, it gets the shorthand.
 * If two or more flags share the same first letter, neither gets one.
 *
 * @param flagNames - kebab-case flag names to assign shorthands for
 * @param extraReserved - additional letters to exclude (e.g., "a" on paginated commands)
 * @returns Map from flag name to its single-letter shorthand
 */
function computeFlagShorthands(
  flagNames: string[],
  extraReserved?: Set<string>,
): Map<string, string> {
  const candidates = new Map<string, string[]>(); // letter → flagNames

  for (const name of flagNames) {
    // Skip dotted nested flags (e.g., "address.city") — shorthands don't make sense
    if (name.includes(".")) continue;

    const letter = name[0];
    if (!letter) continue;
    if (reservedShorthandLetters.has(letter)) continue;
    if (reservedShorthandLetters.has(letter.toUpperCase())) continue;
    if (extraReserved?.has(letter)) continue;

    if (!candidates.has(letter)) candidates.set(letter, []);
    candidates.get(letter)!.push(name);
  }

  const result = new Map<string, string>();
  for (const [letter, names] of candidates) {
    if (names.length === 1) {
      result.set(names[0], letter);
    }
  }
  return result;
}

/**
 * Get the flag description/help text.
 * Uses field description, enum options, and required status.
 */
function getFlagDescription(field: FieldDef): string {
  let desc = "";

  // Start with field description or comments
  if (field.Comments?.Description) {
    desc = field.Comments.Description;
  }

  // Add enum options
  if (isEnumType(field.Type)) {
    const enumHelp = getEnumHelpText(field.Type);
    if (enumHelp) {
      desc = desc ? `${desc} (${enumHelp})` : enumHelp;
    }
  }

  // Add required indicator
  if (!field.Optional) {
    desc = desc ? `${desc} [required]` : "[required]";
  }

  // Fallback: show type hint instead of "No description available"
  return desc || getTypeHint(field);
}
registerTemplateFunc("getFlagDescription", getFlagDescription);
