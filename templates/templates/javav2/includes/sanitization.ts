// @ts-ignore
const javaReservedKeywords = [
  "package",
  "return",
  "null",
  "public",
  "default",
  "private",
  "class",
  "new",
  "interface",
  "static",
  "import",
  "protected",
  "native",
  "case",
  "boolean",
  "short",
  "long",
  "enum",
  "transient",
  "continue",
  "break",
  "switch",
  "abstract",
  "super",
  "final",
  "int",
  "void",
  "else",
  "if",
  "char",
  "do",
  "goto",
  "equals",
  "hashCode",
  "toString",
  "clone",
  "wait",
  "notify",
  "notifyAll",
];

// @ts-ignore
function sanitizeSDKAccess(sdk: SDK): string {
  const group = sdk.Group;

  if (!group) {
    return "";
  }

  let formattedName = "";
  if (group.includes(".") && ".".repeat(group.length) !== group) {
    const pieces = group.split(".");
    formattedName = sanitizeFieldName(pieces[0]) + "()";
    for (const i of pieces.slice(1)) {
      formattedName += "." + sanitizeFieldName(i) + "()";
    }
  } else {
    formattedName = sanitizeFieldName(group) + "()";
  }

  if (formattedName != "") {
    formattedName = `.${formattedName}`;
  }

  return formattedName;
}

registerTemplateFunc("sanitizeSDKAccess", sanitizeSDKAccess);

// @ts-ignore
function sanitizeFieldName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToCamel(name);
  return sanitizeJavaWord(name);
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// @ts-ignore
function sanitizeJavaBeanFieldName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToCamel(name);

  // Convert consecutive uppercase letters to proper camelCase
  // This ensures Spring Boot generates proper kebab-case YAML properties
  // e.g., clientID -> clientId, tokenURL -> tokenUrl
  name = name.replace(/([A-Z])([A-Z]+)/g, (match, first, rest) => {
    return first + rest.toLowerCase();
  });

  return sanitizeJavaWord(name);
}

registerTemplateFunc("sanitizeJavaBeanFieldName", sanitizeJavaBeanFieldName);

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  let name = sanitizeName(operation.GetID());
  name = caser().ToCamel(name);
  return sanitizeJavaWord(name);
}

// @ts-ignore
function sanitizePrivateMethodName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToCamel(name);
  return sanitizeJavaWord(name);
}
registerTemplateFunc("sanitizePrivateMethodName", sanitizePrivateMethodName);

function sanitizeEnumMember(name: string): string {
  return sanitizeJavaWord(templateConstName(sanitizeName(name), ""));
}

function sanitizeJavaWord(name: string): string {
  if (javaReservedKeywords.includes(name)) {
    return name + "_";
  } else {
    return name;
  }
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

function sanitizeMethodNameDirect(operation: Operation): string {
  let name = sanitizeName(operation.GetID());
  name = caser().ToCamel(name);

  if (methodParameters(operation, "operations").length == 0) {
    name += "Direct";
  }

  return sanitizeJavaWord(name);
}

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

// @ts-ignore
function sanitizeClassName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  // Truncate extremely long class names to stay within OS filename limits.
  // Java requires class names to match filenames, and suffixes like
  // "RequestBuilder" (14 chars) + ".java" (5 chars) may be appended.
  // No underscore separator to ensure sanitizeFile consistency between
  // class name and filename.
  name = truncateName(name, 230, "");

  return name;
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

// @ts-ignore
function sanitizeFileName(name: string): string {
  name = sanitizeFile(sanitizeClassName(name), "");

  name = truncateName(name, 250);

  return name;
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => sanitizeName(segment).toLowerCase())
    .join("/");
}

// @ts-ignore
function templateConstName(name: string, prefix: string): string {
  name = sanitizeName(name);
  return caser().ToSNAKE(`${prefix}_${name}`);
}

registerTemplateFunc("templateConstName", templateConstName);

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  scope: string,
  definition: boolean,
): string {
  const className = sanitizeClassName(typeDef.Name);

  if (typeDef.Scope.toString() != scope.toString() && !definition) {
    return getModelNamespace(true, "", typeDef.OutputLocation, className);
  }

  return className;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

function importedClass(typeDef: TypeDef): string {
  return javaImport(fullClassName(typeDef));
}

registerTemplateFunc("importedClass", importedClass);

// ignores optional and nullable properties
function fullClassName(typeDef: TypeDef, defaultOutputLocation: string = "") {
  if (!typeDef.Name) {
    throw new Error("no name: " + JSON.stringify(typeDef, null, 2));
  }
  const simpleName = sanitizeClassName(typeDef.Name);
  let loc = typeDef.OutputLocation;
  if (typeDef.OutputLocation == "") {
    loc = defaultOutputLocation;
  }
  return getModelNamespace(true, "", loc, simpleName);
}
registerTemplateFunc("fullClassName", fullClassName);

function isOpenEnum(typeDef: TypeDef): boolean {
  return (
    typeDef.Enum?.Open ||
    (context.Global.Config.ForwardCompatibleEnumsByDefault &&
      typeDef.UsedInResponse)
  );
}
registerTemplateFunc("isOpenEnum", isOpenEnum);

function enumSimpleClassName(typeDef: TypeDef) {
  const simpleName = sanitizeClassName(typeDef.Name);
  if (isOpenEnum(typeDef)) {
    return simpleName + "Enum";
  } else {
    return simpleName;
  }
}
registerTemplateFunc("enumSimpleClassName", enumSimpleClassName);

// @ts-ignore
function sanitizeTypeBase(
  typeDef: TypeDef,
  mandatory: boolean,
  isAsync: boolean = false,
): string {
  let typeDefString = typeDef.Type.toString();
  switch (typeDefString) {
    case "union":
      // this is a bug fix for AST in that when a union is in AssociatedTypes
      // it does not have the correct OutputLocation
      const defaultOutputLocation = "models/shared";
      return fullClassName(typeDef, defaultOutputLocation);
    case "enum":
    case "error":
    case "class":
      return fullClassName(typeDef);
    case "string":
      return "java.lang.String";
    case "date":
      return "java.time.LocalDate";
    case "date-time":
      return "java.time.OffsetDateTime";
    case "integer":
      return mandatory ? "long" : "java.lang.Long";
    case "bigint":
      return "java.math.BigInteger";
    case "decimal":
      return "java.math.BigDecimal";
    case "int32":
      return mandatory ? "int" : "java.lang.Integer";
    case "number":
      return mandatory ? "double" : "java.lang.Double";
    case "float32":
      return mandatory ? "float" : "java.lang.Float";
    case "boolean":
      return mandatory ? "boolean" : "java.lang.Boolean";
    case "bytes":
      return "byte[]";
    case "request-stream":
      return useBlob() ? `${templatePackageName()}.utils.Blob` : `byte[]`;
    case "response-stream":
      // The reactive (JDK arm) async body is a Blob over a Publisher; the
      // okhttp arm's async body is a plain InputStream, as in sync.
      return isAsync && !useOkHttp()
        ? `${templatePackageName()}.utils.Blob`
        : "java.io.InputStream";
    case "map":
      return sanitizeMap(typeDef);
    case "array":
      return sanitizeList(typeDef);
    case "any":
      return "java.lang.Object";
    case "response":
      return javaRawResponseType(isAsync);
    case "event-stream":
      return sanitizeEventStream(typeDef);
    case "jsonl":
      return sanitizeJsonLStream(typeDef);
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}
registerTemplateFunc("sanitizeTypeBase", sanitizeTypeBase);

function sanitizeEventStream(typeDef: TypeDef): string {
  return `${javaImportLocal("utils.EventStream")}<${javaImport(
    sanitizeEventStreamItemType(typeDef),
  )}>`;
}
function sanitizeJsonLStream(typeDef: TypeDef): string {
  return `${javaImportLocal("utils.JsonLStream")}<${javaImport(
    sanitizeJsonLStreamItemType(typeDef),
  )}>`;
}

function sanitizeEventStreamItemType(typeDef: TypeDef) {
  return sanitizeTypeMandatory(typeDef.ItemType);
}

function sanitizeJsonLStreamItemType(typeDef: TypeDef) {
  return sanitizeTypeMandatory(typeDef.ItemType);
}

const knownFinalClasses = new Set<string>([
  "java.lang.String",
  "java.lang.Integer",
  "java.lang.Short",
  "java.lang.Long",
  "java.lang.Float",
  "java.lang.Double",
  "java.lang.Byte",
  "java.lang.Boolean",
  "java.util.Optional",
  "java.time.LocalDate",
  "java.time.OffsetDateTime",
  "byte[]",
]);

function simpleClassName(className: string): string {
  const i = className.lastIndexOf(".");
  if (i == -1) {
    return className;
  } else {
    return className.substring(i + 1);
  }
}
registerTemplateFunc("simpleClassName", simpleClassName);

function unqualifyType(javaType: string): string {
  return JavaTypeParser.convert(javaType, (x) => simpleClassName(x));
}
registerTemplateFunc("unqualifyType", unqualifyType);

// returns a fully qualified java type from a TypeDef.
// for example an optional map type might be returned
// as `java.util.Optional<java.util.Map<java.lang.String, java.lang.String>>`
//
// Note that such a fully qualified java type can be passed to
// the `JavaTypeParser.convert` function for manipulation. The function
// `unqualifyType` uses that function so that the example above would
// be converted to
//   `Optional<Map<String, String>>`.
// @ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional?: boolean,
  nullable?: boolean,
  useWildcard: boolean = true,
  isAsync: boolean = false,
): string {
  let typ = sanitizeTypeBase(typeDef, !optional && !nullable, isAsync);
  const t =
    useWildcard && !knownFinalClasses.has(typ) ? `? extends ${typ}` : typ;
  if (optional && nullable) {
    return `org.openapitools.jackson.nullable.JsonNullable<${t}>`;
  }
  if (optional || nullable) {
    return `java.util.Optional<${t}>`;
  }
  return typ;
}

registerTemplateFunc("sanitizeType", sanitizeType);

function needToCastToRemoveWildcard(field: FieldDef): boolean {
  const withoutWildcard = sanitizeType(
    field.Type,
    field.Optional,
    field.Nullable,
    false,
  );
  const withWildcard = sanitizeType(
    field.Type,
    field.Optional,
    field.Nullable,
    true,
  );
  return withWildcard !== withoutWildcard;
}
registerTemplateFunc("needToCastToRemoveWildcard", needToCastToRemoveWildcard);

function castToRemoveWildcard(field: FieldDef, isAsync: boolean): string {
  const withoutWildcard = sanitizeType(
    field.Type,
    field.Optional,
    field.Nullable,
    false,
    isAsync,
  );
  const withWildcard = sanitizeType(
    field.Type,
    field.Optional,
    field.Nullable,
    true,
    isAsync,
  );
  if (withWildcard == withoutWildcard) {
    // no cast necessary
    return "";
  } else {
    // drop wildcard for use in cast
    return `(${javaImport(withoutWildcard)}) `;
  }
}

registerTemplateFunc("castToRemoveWildcard", castToRemoveWildcard);

// @ts-ignore
function templateAuxPath(ctx: Context): string {
  return templatePackageName().replaceAll(".", "/");
}
registerTemplateFunc("templateAuxPath", templateAuxPath);

function sanitizeRequiredRequestParams(fields: FieldDef[]): string {
  return fields
    .filter((f) => !f.Optional)
    .map((f) => sanitizeFieldName(f.Name))
    .join(", ");
}
registerTemplateFunc(
  "sanitizeRequiredRequestParams",
  sanitizeRequiredRequestParams,
);

// @ts-ignore
function sanitizeMap(typeDef: TypeDef): string {
  return `java.util.Map<java.lang.String, ${toNonPrimitive(
    sanitizeTypeMandatory(typeDef.ItemType),
  )}>`;
}

function sanitizeList(typeDef: TypeDef): string {
  // be a bit defensive with fully qualified List type, List may not be an uncommon
  // client-defined type out there. TODO confirm
  return `java.util.List<${toNonPrimitive(
    sanitizeTypeMandatory(typeDef.ItemType),
  )}>`;
}

// @ts-ignore
function sanitizeComments(comment) {
  if (comment.length > 1 && comment[0] == "@") {
    if (comment[1] != " ") {
    }
  }

  // First, extract and preserve {@code ...} blocks
  const codeBlocks = [];
  let processedComment = comment.replace(
    /{@code ([^}]*)}/g,
    (match, content) => {
      const placeholder = `__CODE_BLOCK_${codeBlocks.length}__`;
      codeBlocks.push(`{@code ${content}}`);
      return placeholder;
    },
  );

  // Extract and preserve Javadoc-compatible HTML tags
  const htmlBlocks = [];

  // Preserve <a href="...">...</a> tags
  processedComment = processedComment.replace(
    /<a\s+href="([^"]*)"[^>]*>([^<]*)<\/a>/gi,
    (match, href, text) => {
      const placeholder = `__HTML_BLOCK_${htmlBlocks.length}__`;
      htmlBlocks.push(`<a href="${href}">${text}</a>`);
      return placeholder;
    },
  );

  // Preserve other common Javadoc HTML tags: <p>, <ul>, <ol>, <li>, <pre>, <code>, <em>, <strong>, <b>, <i>
  const javadocTags = [
    "p",
    "ul",
    "ol",
    "li",
    "pre",
    "code",
    "em",
    "strong",
    "b",
    "i",
  ];
  javadocTags.forEach((tag) => {
    // Handle both self-closing and paired tags
    const regex = new RegExp(
      `<(${tag})([^>]*)>([^<]*)<\\/${tag}>|<(${tag})([^>]*)\\/?>`,
      "gi",
    );
    processedComment = processedComment.replace(regex, (match) => {
      const placeholder = `__HTML_BLOCK_${htmlBlocks.length}__`;
      htmlBlocks.push(match);
      return placeholder;
    });
  });

  // Convert <br> tags to newlines before other HTML escaping
  processedComment = processedComment.replace(/<br\s*\/?>/gi, "\n");

  // Apply normal sanitization to the comment with placeholders
  processedComment = processedComment
    .replaceAll(/{@link (.*?)}/g, "$1")
    .replaceAll("&", "&amp;")
    .replaceAll(/^@/g, "&#64;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll("\r\n", "\n")
    .replaceAll("*/", "* /")
    .replaceAll("\\u", "\\ u");

  // Restore the preserved HTML blocks
  htmlBlocks.forEach((htmlBlock, index) => {
    processedComment = processedComment.replace(
      `__HTML_BLOCK_${index}__`,
      htmlBlock,
    );
  });

  // Restore the {@code ...} blocks
  codeBlocks.forEach((codeBlock, index) => {
    processedComment = processedComment.replace(
      `__CODE_BLOCK_${index}__`,
      codeBlock,
    );
  });

  return processedComment;
}

registerTemplateFunc("sanitizeComments", sanitizeComments);

// @ts-ignore
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return `.${sanitizeFieldName(fieldDef.Name)}(`;
}

// @ts-ignore
function templateFieldDelimiter(): string {
  // using builders now to specify fields so this\
  // close bracket will terminate the setting of a value
  return ")";
}

function joinWithComma(values: any[]) {
  // every value but the last gets a comma appended
  const len = values.length;
  return values.map((x, index) => {
    return index + 1 == len ? x : x + ",";
  });
}

// @ts-ignore
function templateArrayValue(val: any): string {
  return `${val.trimStart()}`;
}

// @ts-ignore
function templateMapValue(key: string, val: any): string {
  return `${templateJdkCall("Map.entry")}("${key}", ${val})`;
}

// @ts-ignore
function templateMapFieldDelimiter(): string {
  return ",";
}

// @ts-ignore
function templateOptionalSymbol(): string {
  return "";
}

// @ts-ignore
function templateBracket(
  typeDef: TypeDef,
  opening: boolean,
  additionalContext?: TemplateValueContext,
): string {
  // because we use a builder we don't need {{, }} wrapping field setters
  const typ = typeDef.Type.toString();
  if (typ === "map" || typ === "array") {
    return opening ? "" : ")";
  } else {
    return opening ? "" : ".build()";
  }
}

// @ts-ignore
function templateIndent(indent: number): string {
  return "    ".repeat(indent);
}

// @ts-ignore
function templateStringValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (val.includes(".getOrDefault(")) {
    return val;
  }
  return quote(val);
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  _additionalContext?: TemplateValueContext,
): string {
  const enumNames = getEnumNames(fieldDef.Type);
  return `${javaImport(fullClassName(fieldDef.Type))}.${enumNames[idx]}`;
}

// @ts-ignore
function templateBoolValue(
  val: boolean,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return val ? "true" : "false";
}

// @ts-ignore
function templateByteValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `"${val.replaceAll(/"/g, '\\"')}".getBytes(${javaImport(
    "java.nio.charset.StandardCharsets",
  )}.UTF_8)`;
}

// @ts-ignore
function templateDateValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${javaImport("java.time.LocalDate")}.parse("${val}")`;
}

// @ts-ignore
function templateDateTimeValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${javaImport("java.time.OffsetDateTime")}.parse("${val}")`;
}

// @ts-ignore
function templateIntValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (fieldDef.Type.Type.toString() === "int32") {
    return `${val}`;
  } else if (fieldDef.Type.Type.toString() == "bigint") {
    return `new ${javaImport("java.math.BigInteger")}("${val}")`;
  }
  return `${val}L`;
}

// @ts-ignore
function templateFloatValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let t = fieldDef.Type.Type.toString();
  if (t == "any") {
    t = "number";
  }
  if (t == "number") {
    let valueAsString = `${val}`;
    if (!valueAsString.includes(".")) {
      valueAsString = `${valueAsString}d`;
    }

    return `${valueAsString}`;
  } else if (t === "decimal") {
    return `new ${javaImport("java.math.BigDecimal")}("${val}")`;
  } else {
    return `${val}f`;
  }
}

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  example?: any,
  additionalContext?: TemplateValueContext,
): string {
  if (typeDef.Type.toString() === "array") {
    return `${templateJdkCall("List.of")}(`;
  } else if (typeDef.Type.toString() === "map") {
    return `${templateJdkCall("Map.ofEntries")}(`;
  } else {
    let type: string = sanitizeTypeMandatory(typeDef);
    const requiredFields = nonConstFields(typeDef.Fields)
      .filter((field) => !field.Optional)
      .map((field) => {
        let fieldExample = findFieldExample(example, field);

        return `\n${templateIndent(
          additionalContext?.indent,
        )}.${sanitizeFieldName(field.Name)}(${indentLines(
          [templateValue(field, fieldExample, false, additionalContext)],
          1,
        ).trimStart()})`;
      })
      .join("");
    return `${javaImport(
      type,
      false,
      additionalContext,
    )}.builder()${requiredFields}`;
  }
}

function findFieldExample(example: any, field: FieldDef): any {
  if (example == undefined) {
    return undefined;
  }
  const k = Object.keys(example).find(
    (key) => key.toLowerCase() === originalFieldName(field).toLowerCase(),
  );
  if (k == undefined) {
    return undefined;
  }
  return example?.[k];
}

// @ts-ignore
function getEnumNamesFromValues(values: string[]): string[] {
  let enumNames = [];

  let names = {};
  for (const value of values) {
    let name = getEnumName(value);
    if (!names[name]) {
      names[name] = 0;
    }

    names[name] += 1;
  }

  let seen = {};
  for (const value of values) {
    let name = getEnumName(value);
    if (names[name] > 1) {
      let candidate = `${name}${caser().ToPascal(getCasing(value))}`;
      if (seen[candidate]) {
        let suffix = seen[candidate];
        seen[candidate] += 1;
        candidate = `${candidate}${suffix}`;
      } else {
        seen[candidate] = 1;
      }
      name = candidate;
    }

    enumNames.push(sanitizeEnumMember(name));
  }

  return enumNames;
}

// Java-specific scope dedup that produces PascalCase names
// (Java OAuth2 scope enums use PascalCase, not UPPER_SNAKE_CASE)
// @ts-ignore
function getDeduplicatedScopeClassNames(scopes: OAuth2Scope[]): string[] {
  const values = scopes.map((s) => s.Name);

  let counts = {};
  for (const value of values) {
    let name = getEnumName(value);
    if (!counts[name]) {
      counts[name] = 0;
    }
    counts[name] += 1;
  }

  let result = [];
  let seen = {};
  for (const value of values) {
    let name = getEnumName(value);
    if (counts[name] > 1) {
      let candidate = `${name}${caser().ToPascal(getCasing(value))}`;
      if (seen[candidate]) {
        let suffix = seen[candidate];
        seen[candidate] += 1;
        candidate = `${candidate}${suffix}`;
      } else {
        seen[candidate] = 1;
      }
      name = candidate;
    }
    result.push(name);
  }

  return result;
}
registerTemplateFunc(
  "getDeduplicatedScopeClassNames",
  getDeduplicatedScopeClassNames,
);

// @ts-ignore
function getEnumName(value: string): string {
  let name = value.trim();

  if (name === "") {
    name = "Unknown";
  }

  name = sanitizeName(name);

  return caser().ToPascal(name);
}

// @ts-ignore
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName);
}

// @ts-ignore
function sanitizeJsonPath(jsonpath: string): string {
  /*
   * Support for pagination is specified by the client through a `x-speakeasy-pagination` entry in the spec file.
   * This uses a JSONPath expression to specify the path of pagination fields in the response.
   */
  const path = jsonpath
    // Note (alexa): couldn't find a library that correctly parses this expression
    // This is not a long-term solution
    .replaceAll(/\(\@\.length ?- ?1\)/gi, "-1");
  return path;
}

registerTemplateFunc("sanitizeJsonPath", sanitizeJsonPath);

function castValue(value: string, type: TypeDef): string {
  if (type.Type.toString() == "number" || type.Type.toString() == "integer") {
    return `${javaImport("java.lang.Long")}.valueOf(${value})`;
  } else if (type.Type.toString() == "string") {
    return `${value}`;
  }
  return type.Type.toString();
}

registerTemplateFunc("castValue", castValue);

function sanitizeEnumOption(name: string): string {
  return caser().ToSNAKE(sanitizeFieldName(name));
}

function bigType(type: TypeDef): BigType | undefined {
  const typ = type.Type.toString();
  const fmt = type.Format;
  if (typ == "bigint") {
    if (fmt == "string") {
      return BigType.BigIntegerString;
    } else {
      return BigType.BigInteger;
    }
  } else if (typ == "decimal") {
    if (fmt == "string") {
      return BigType.BigDecimalString;
    } else {
      return BigType.BigDecimal;
    }
  } else if (type.ItemType) {
    return bigType(type.ItemType);
  } else {
    return undefined;
  }
}

function isBigint(type: TypeDef): boolean {
  return type.Type.toString() == "bigint";
}

registerTemplateFunc("isBigint", isBigint);

function isDecimal(type: TypeDef): boolean {
  return type.Type.toString() == "decimal";
}

registerTemplateFunc("isDecimal", isDecimal);

function isResponse(type: TypeDef): boolean {
  return type.Fields.some((f) => f.IsResponseMetadata);
}
registerTemplateFunc("isResponse", isResponse);

// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  if (context.Global.Config.NullFriendlyParameters) {
    return "null";
  }
  if (fieldDef.Optional && fieldDef.Nullable) {
    return `${javaImportJsonNullable()}.of(null)`;
  } else {
    return `${javaImportOptional()}.empty()`;
  }
}

// @ts-ignore
function sanitizeUnionTypeName(typeDef: TypeDef): string {
  if (typeDef.Name) {
    return typeDef.Name;
  }

  switch (typeDef.Type.toString()) {
    case "string":
      return "str";
    case "map":
    case "array":
      return (
        typeDef.Type.toString() +
        "Of" +
        caser().ToPascal(sanitizeUnionTypeName(typeDef.ItemType))
      );
    default:
      return typeDef.Type.toString();
  }
}

// @ts-ignore
function languageSpecificEnvVarWrapping(
  envVar: EnvVar,
  additionalContext?: TemplateValueContext,
  fieldDef?: FieldDef,
): string {
  const isArray = fieldDef?.Type?.Type?.toString() === "array";

  if (additionalContext?.isTest) {
    const envVarCall = `${javaImportLocal(
      "utils.Utils",
    )}.environmentVariable("${envVar.name}", "${envVar.defaultValue}")`;

    if (isArray) {
      return `${templateJdkCall("List.of")}(${envVarCall})`;
    }
    return envVarCall;
  }

  const envVarCall = `System.getenv().getOrDefault("${envVar.name}", "${
    envVar.defaultValue || ""
  }")`;

  if (isArray) {
    return `${templateJdkCall("List.of")}(${envVarCall})`;
  }
  return envVarCall;
}

// Function to strip generics from a Java class name
// For example: "MyClass<String, Integer>" becomes "MyClass"
function stripGenerics(className: string): string {
  const genericStart = className.indexOf("<");
  if (genericStart === -1) {
    return className;
  }
  return className.substring(0, genericStart);
}

registerTemplateFunc("stripGenerics", stripGenerics);

function* reservedSdkNames(sdk: SDK): Generator<string> {
  // Operation method names
  for (const op of sdk.Operations) {
    yield sanitizeMethodName(op);
  }

  // Sub-SDK accessor names
  for (const subSDK of sdk.SubSDKs) {
    if (subSDK.Type.Name !== "") {
      yield sanitizeFieldName(subSDK.FieldName);
    }
  }
}

function sanitizeModeMethodName(mode: string, sdk: SDK): string {
  const prefixed = !!context.Global.Config.PrefixModeMethodNames;
  let name: string;

  if (prefixed) {
    name = "to" + caser().ToPascal(sanitizeName(mode)); // toSync, toAsync
  } else {
    name = caser().ToCamel(sanitizeName(mode)); // sync, async
  }

  for (const reservedName of reservedSdkNames(sdk)) {
    if (reservedName === name) {
      if (prefixed) {
        return name + "SDK"; // toSyncSDK, toAsyncSDK
      }
      return "_" + name; // _sync, _async (legacy)
    }
  }

  return name;
}

registerTemplateFunc("sanitizeModeMethodName", sanitizeModeMethodName);

function modeMethodBaseName(mode: string): string {
  if (context.Global.Config.PrefixModeMethodNames) {
    return "to" + caser().ToPascal(sanitizeName(mode));
  }
  return caser().ToCamel(sanitizeName(mode));
}
registerTemplateFunc("modeMethodBaseName", modeMethodBaseName);
