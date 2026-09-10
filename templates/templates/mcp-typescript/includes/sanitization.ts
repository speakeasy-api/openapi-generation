const reservedMCPCLIFlags = new Set([
  "auth",
  "env",
  "h",
  "help",
  "host-name",
  "host",
  "hostname",
  "log-level",
  "port",
  "prompt",
  "prompts",
  "resource",
  "resources",
  "scope",
  "token",
  "tool",
  "tools",
  "transport",
  "mode",
  "v",
  "version",
]);

function sanitizeAcceptEnumKey(acceptType: string): string {
  return `${caser().ToCamel(
    sanitizeName(sanitizeAcceptEnumValue(acceptType)),
  )}Accept`;
}

registerTemplateFunc("sanitizeAcceptEnumKey", sanitizeAcceptEnumKey);

function sanitizeAcceptEnumName(operation: Operation): string {
  return `${caser().ToPascal(sanitizeName(operation.GetID()))}AcceptEnum`;
}

registerTemplateFunc("sanitizeAcceptEnumName", sanitizeAcceptEnumName);

function sanitizeAcceptEnumValue(acceptType: string): string {
  return `${acceptType.split(";")[0]}`;
}

registerTemplateFunc("sanitizeAcceptEnumValue", sanitizeAcceptEnumValue);

function deduplicateAcceptTypes(acceptTypes: string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];

  for (const acceptType of acceptTypes) {
    const sanitizedValue = sanitizeAcceptEnumValue(acceptType);
    if (!seen.has(sanitizedValue)) {
      seen.add(sanitizedValue);
      result.push(acceptType);
    }
  }

  return result;
}

registerTemplateFunc("deduplicateAcceptTypes", deduplicateAcceptTypes);

function sanitizeAccessor(
  field: string,
  property: string,
  optional?: boolean,
): string {
  const isValidIdent = typescriptIdentifierRegex.test(property);
  const sanitized = isValidIdent ? property : `["${property}"]`;

  if (optional) {
    return `${field}?.${sanitized}`;
  }

  return isValidIdent ? `${field}.${sanitized}` : `${field}${sanitized}`;
}

registerTemplateFunc("sanitizeAccessor", sanitizeAccessor);

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
): string {
  let className = sanitizeClassName(typeDef.Name);

  const nsPrefix = getNamespacePrefix(typeDef.OutputLocation, usageLocation);
  if (nsPrefix) {
    return `${nsPrefix}_${className}`;
  }

  let prefix = "";
  if (!definition) {
    const modelLocation = getModelsLocation(typeDef.OutputLocation);
    if (modelLocation !== usageLocation && usageLocation !== "usage") {
      const parts = modelLocation.split("/");

      prefix = `${parts[parts.length - 1]}.`;
    }
  }

  return `${prefix}${className}`;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

//@ts-ignore
function sanitizeClassName(name: string) {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (typescriptReservedTypeNames.has(name)) {
    name += "T";
  }

  return caser().ToPascal(name);
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

function shouldTemplateConstValue(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): boolean {
  return true;
}

function sanitizeComments(comment: string | undefined): string {
  if (!comment) {
    return "";
  }

  return comment
    .replaceAll("*/", "* /")
    .replaceAll("\r\n", "\n")
    .replaceAll("\uFEFF", "")
    .replaceAll(/\u00A0/g, " ")
    .replaceAll(/\u200b/g, " ")
    .replaceAll("{{", `{{"{{"}}`);
}

function sanitizeConstName(name: string, prefix: string) {
  name = sanitizeName(name);
  return caser().ToPascal(`${prefix}_${name}`);
}

registerTemplateFunc("sanitizeConstName", sanitizeConstName);

function sanitizeCoreSDKName(sdk?: SDK) {
  const s = sdk ?? context.Global.AST.MainSDK;
  const name = sanitizeName(s.Type.Name);
  return caser().ToPascal(name + "Core");
}

registerTemplateFunc("sanitizeCoreSDKName", sanitizeCoreSDKName);

// @ts-ignore
function sanitizeEnumName(value: string): string {
  let name = value.trim();

  if (name === "") {
    name = "Unknown";
  }

  name = sanitizeName(name);

  return caser().ToPascal(name);
}

// @ts-ignore
function sanitizeEnumValue(value: any, type: GojaEnum<DataType>): string {
  switch (type.toString()) {
    case "string":
      return `"${value}"`;
    case "int32":
    case "integer":
      return `${value}`;
    default:
      throw new Error(`Unknown enum type: ${type}`);
  }
}

// @ts-ignore
function sanitizeFieldName(name) {
  let result = sanitizeName(name);

  if (!typescriptIdentifierRegex.test(name)) {
    result = caser().ToCamel(result);
  }

  if (typescriptReservedTypeNames.has(result)) {
    result += "T";
  }

  return result;
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

function truncateName(
  name: string,
  maxLength: number,
  separator: string = "_",
): string {
  if (name.length <= maxLength) return name;
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = ((hash << 5) - hash + name.charCodeAt(i)) | 0;
  }
  const hashStr = Math.abs(hash).toString(16).padStart(8, "0").substring(0, 8);
  const suffixLen = separator.length + 8;
  return name.substring(0, maxLength - suffixLen) + separator + hashStr;
}

//@ts-ignore
function sanitizeFileName(name: string) {
  name = sanitizeFile(name, "");
  name = name.toLowerCase();

  name = truncateName(name, 240);

  return name;
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function sanitizeFromJSONFuncName(typeDef: TypeDef): string {
  return sanitizeFieldName(typeDef.Name) + "FromJSON";
}

registerTemplateFunc("sanitizeFromJSONFuncName", sanitizeFromJSONFuncName);

// @ts-ignore
function sanitizeFuncFilename(operation: Operation): string {
  let name = sanitizeFuncName(operation);
  return truncateName(name, 240);
}

// @ts-ignore
function sanitizeFuncName(operation: Operation): string {
  const group = operation.OwningSDK.Group;
  let prefix = "";
  // We only qualify sub-sdk functions, not the root SDKs functions.
  if (group) {
    prefix = group.replaceAll(".", "_") + "_";
  }

  let name = sanitizeName(`${prefix}${operation.GetID()}`);
  name = caser().ToCamel(name);

  if (typescriptReservedTypeNames.has(name)) {
    name += "T";
  }

  return name;
}

registerTemplateFunc("sanitizeFuncName", sanitizeFuncName);

// @ts-ignore
function sanitizeKey(input: string): string {
  return typescriptIdentifierRegex.test(input) ? input : `"${input}"`;
}

registerTemplateFunc("sanitizeKey", sanitizeKey);

// @ts-ignore
function sanitizeMap(typeDef: TypeDef, usageLocation: string): string {
  return `{ [k: string]: ${sanitizeType(
    typeDef.ItemType,
    false,
    usageLocation,
  )} }`;
}

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  let id = operation.GetID();
  return caser().ToCamel(sanitizeName(id));
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

// @ts-ignore
function sanitizePrivateMethodName(name: string): string {
  return caser().ToCamel(sanitizeName(name));
}

registerTemplateFunc("sanitizePrivateMethodName", sanitizePrivateMethodName);

// Required from common.
// @ts-ignore
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName);
}

// @ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional?: boolean,
  usageLocation: string = "",
): string {
  switch (typeDef.Type.toString()) {
    case "union":
    case "enum":
    case "error":
    case "class":
      return sanitizeClass(typeDef, usageLocation, false);
    case "date":
    case "date-time":
    case "string":
      return "string";
    case "integer":
    case "int32":
    case "number":
    case "float32":
      return "number";
    case "boolean":
      return "boolean";
    case "bytes":
      return "Uint8Array";
    case "map":
      return sanitizeMap(typeDef, usageLocation);
    case "array":
      return sanitizeType(typeDef.ItemType, false, usageLocation) + "[]";
    case "event-stream":
      const eventType = sanitizeType(typeDef.ItemType, false, usageLocation);
      return `EventStream<${eventType}>`;
    case "jsonl":
      const jsonlEventType = sanitizeType(
        typeDef.ItemType,
        false,
        usageLocation,
      );
      return `JsonLStream<${jsonlEventType}>`;
    case "any":
      return "any";
    case "response":
      return "Response";
    case "request":
      return "Request";
    case "decimal":
      return "Decimal";
    case "bigint":
      return "BigInt";
    case "request-stream":
    case "response-stream":
      return "ReadableStream<Uint8Array>";
    case "function":
      return `() => ${sanitizeType(typeDef.ItemType, false, usageLocation)}`;
    case "async":
      return `Promise<${sanitizeType(typeDef.ItemType, false, usageLocation)}>`;
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}

registerTemplateFunc("sanitizeType", sanitizeType);

// @ts-ignore
function sanitizeVariableName(name: string, fallback?: string) {
  let result = sanitizeFieldName(name);
  if (typescriptReservedVariableKeywords.has(result)) {
    if (fallback) return fallback;
    return result + "$";
  }
  return result;
}

registerTemplateFunc("sanitizeVariableName", sanitizeVariableName);

// @ts-ignore
function sanitizeZod(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
) {
  return sanitizeClass(typeDef, usageLocation, definition) + "$zodSchema";
}

// @ts-ignore
function sanitizeZodName(name: string) {
  return sanitizeClassName(name) + "$zodSchema";
}

registerTemplateFunc("sanitizeZodName", sanitizeZodName);

// @ts-ignore
function sanitizeZodRef(typeDef: TypeDef, usageLocation: string): string {
  return sanitizeZod(typeDef, usageLocation, true);
}

// @ts-ignore
function getEnumName(value) {
  let name = value.trim();

  if (name === "") {
    name = "Unknown";
  }

  name = sanitizeName(name);

  return caser().ToPascal(name);
}

registerTemplateFunc("getEnumName", getEnumName);

// @ts-ignore
function getEnumNames(t: TypeDef): string[] {
  if (t.Enum?.Names.length > 0) {
    if (context.Global.Config.FixEnumNameSanitization === true) {
      return t.Enum.Names.map((n) => sanitizeName(n.trim() || "Unknown"));
    }
    return t.Enum.Names.map((n) => getEnumName(n));
  }

  return getEnumNamesFromValues(t.Enum?.Values);
}

registerTemplateFunc("getEnumNames", getEnumNames);

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  additionalContext?: TemplateValueContext,
) {
  // The `mcp-typescript` only uses const objects with named properties
  // (enumFormat == "union") so we always return `Enum.PropertyName` style
  const names = getEnumNames(fieldDef.Type);
  const enumClassName = sanitizeClassName(fieldDef.Type.Name);

  // Add import - use test import if in test context
  if (additionalContext?.isTest) {
    const resolved = resolveTypeDefImport(fieldDef.Type, "usage");
    if (resolved) {
      addTestImport(resolved.path, enumClassName, resolved.type);
    }
  } else {
    addUsageImportForType(fieldDef.Type);
  }

  return `${enumClassName}.${names[idx]}`;
}

registerTemplateFunc("templateEnumValue", templateEnumValue);

// @ts-ignore
function templateMapValue(key: string, val: string): string {
  const trimmed = val.trim();
  if (trimmed === "" || trimmed === "{}" || trimmed === "{\n}") {
    return "";
  }
  return `"${key}": ${val}`;
}

registerTemplateFunc("templateMapValue", templateMapValue);

// @ts-ignore
function processMapValues(values: any[]): string[] {
  return values
    .filter((v) => v !== "")
    .map((v) => v + templateFieldDelimiter());
}

registerTemplateFunc("processMapValues", processMapValues);

// @ts-ignore
function templateFieldDelimiter(): string {
  return ",";
}

registerTemplateFunc("templateFieldDelimiter", templateFieldDelimiter);

// @ts-ignore
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => {
      segment = sanitizeName(segment);
      return caser().ToCamel(segment);
    })
    .join("/");
}

// Converts a package name (e.g. "@acme/my-sdk") into a valid
// UPPER_SNAKE_CASE identifier suitable for use as an environment variable or
// Cloudflare binding name (e.g. "ACME_MY_SDK").
function sanitizeEnvVarName(name: string): string {
  return name
    .replace(/^@/, "")
    .replace(/[^a-zA-Z0-9]/g, "_")
    .replace(/_+/g, "_")
    .replace(/^_|_$/g, "")
    .toUpperCase();
}

registerTemplateFunc("sanitizeEnvVarName", sanitizeEnvVarName);

// @ts-ignore
function languageSpecificEnvVarWrapping(
  envVar: EnvVar,
  additionalContext?: TemplateValueContext,
): string {
  return `process.env["${envVar.name}"] ?? "${envVar.defaultValue}"`;
}
