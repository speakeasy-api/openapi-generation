/**
 * Regular expression to match valid TypeScript identifiers. For example, use
 * to detect whether a property key must be quotes or access requires bracket
 * notation.
 */
const typescriptIdentifierRegex = /^[a-zA-Z\$_][a-zA-Z0-9\$_]*$/;

/**
 * TypeScript lanaguage reserved type names (e.g. Array, Record, etc.).
 * Use this to prevent shadowing of language keywords.
 */
const typescriptReservedTypeNames = new Set([
  "Array",
  "Awaited",
  "BigInt",
  "Blob",
  "ConstructorParameters",
  "Date",
  "Error",
  "Exclude",
  "Extract",
  "File",
  "Function",
  "InstanceType",
  "NonNullable",
  "null",
  "Number",
  "Object",
  "Omit",
  "OmitThisParameter",
  "OpenEnum",
  "Parameters",
  "Partial",
  "Pick",
  "Promise",
  "Readonly",
  "Record",
  "Request",
  "Required",
  "Response",
  "ReturnType",
  "ThisParameterType",
  "ThisType",
  "undefined",
]);

/**
 * TypeScript language reserved variable keywords (e.g. class, switch, etc.).
 * Use this to prevent shadowing of language keywords.
 */
const typescriptReservedVariableKeywords = new Set([
  "arguments",
  "boolean",
  "break",
  "case",
  "catch",
  "class",
  "const",
  "continue",
  "debugger",
  "default",
  "delete",
  "do",
  "else",
  "enum",
  "export",
  "extends",
  "false",
  "finally",
  "for",
  "function",
  "get",
  "if",
  "implements",
  "import",
  "in",
  "instanceof",
  "interface",
  "let",
  "module",
  "new",
  "null",
  "package",
  "private",
  "protected",
  "public",
  "require",
  "return",
  "static",
  "super",
  "switch",
  "symbol",
  "this",
  "throw",
  "true",
  "try",
  "typeof",
  "var",
  "void",
  "while",
  "with",
  "yield",
]);
