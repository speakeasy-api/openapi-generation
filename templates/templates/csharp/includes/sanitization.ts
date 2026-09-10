//https://learn.microsoft.com/en-us/dotnet/csharp/language-reference/keywords/
// @ts-ignore
const cSharpReservedKeywords = [
  "abstract",
  "as",
  "base",
  "bool",
  "break",
  "byte",
  "case",
  "catch",
  "char",
  "checked",
  "class",
  "const",
  "continue",
  "decimal",
  "default",
  "delegate",
  "do",
  "double",
  "else",
  "enum",
  "event",
  "explicit",
  "extern",
  "false",
  "finally",
  "fixed",
  "float",
  "for",
  "foreach",
  "goto",
  "if",
  "implicit",
  "in",
  "int",
  "interface",
  "internal",
  "is",
  "lock",
  "long",
  "namespace",
  "new",
  "null",
  "object",
  "operator",
  "out",
  "override",
  "params",
  "private",
  "protected",
  "public",
  "readonly",
  "ref",
  "return",
  "sbyte",
  "sealed",
  "short",
  "sizeof",
  "stackalloc",
  "static",
  "string",
  "struct",
  "switch",
  "this",
  "throw",
  "true",
  "try",
  "typeof",
  "uint",
  "ulong",
  "unchecked",
  "unsafe",
  "ushort",
  "using",
  "virtual",
  "void",
  "volatile",
  "while",
  "add",
  "and",
  "alias",
  "ascending",
  "args",
  "async",
  "await",
  "by",
  "descending",
  "dynamic",
  "equals",
  "file",
  "from",
  "get",
  "global",
  "group",
  "init",
  "into",
  "join",
  "let",
  "managed",
  "nameof",
  "nint",
  "not",
  "notnull",
  "nuint",
  "on",
  "or",
  "orderby",
  "partial",
  "record",
  "remove",
  "required",
  "scoped",
  "select",
  "set",
  "unmanaged",
  "value",
  "var",
  "when",
  "where",
  "with",
  "yield",
  // reserved for our serialization/deserialization methods
  "BuildHttpRequestMessage",
];

const cSharpReservedClasses = [
  "Console",
  "Enum",
  "Environment",
  "Exception",
  "Object",
  "Type",
  "Uri",
  "Version",
  "NotImplementedException",
  "SystemException",
  "AccessViolationException",
  "AggregateException",
  "AppDomainUnloadedException",
  "ApplicationException",
  "ArgumentException",
  "ArgumentNullException",
  "ArgumentOutOfRangeException",
  "ArithmeticException",
  "ArrayTypeMismatchException",
  "BadImageFormatException",
  "CannotUnloadAppDomainException",
  "ContextMarshalException",
  "DataMisalignedException",
  "DivideByZeroException",
  "DllNotFoundException",
  "DuplicateWaitObjectException",
  "EntryPointNotFoundException",
  "ExecutionEngineException",
  "FieldAccessException",
  "FormatException",
  "IndexOutOfRangeException",
  "InsufficientMemoryException",
  "InvalidCastException",
  "InvalidOperationException",
  "InvalidProgramException",
  "InvalidTimeZoneException",
  "MemberAccessException",
  "MethodAccessException",
  "MissingFieldException",
  "MissingMemberException",
  "MissingMethodException",
  "MulticastNotSupportedException",
  "NotCancelableException",
  "NotFiniteNumberException",
  "NotImplementedException",
  "NotSupportedException",
  "NullReferenceException",
  "ObjectDisposedException",
  "OperationCanceledException",
  "OutOfMemoryException",
  "OverflowException",
  "PlatformNotSupportedException",
  "RankException",
  "StackOverflowException",
  "SystemException",
  "TimeoutException",
  "TimeZoneNotFoundException",
  "TypeAccessException",
  "TypeInitializationException",
  "TypeLoadException",
  "TypeUnloadedException",
  "UnauthorizedAccessException",
  "UriFormatException",
  "Exception",
  "ConstraintException",
  "DataException",
  "DBConcurrencyException",
  "DeleteRowInaccessibleException",
  "DuplicateNameException",
  "EvaluateException",
  "InRowChangingEventException",
  "InvalidConstraintException",
  "InvalidExpressionException",
  "MissingPrimaryKeyException",
  "NoNullAllowedException",
  "OperationAbortedException",
  "ReadOnlyException",
  "RowNotInTableException",
  "StrongTypingException",
  "SyntaxErrorException",
  "TypedDataSetGeneratorException",
  "VersionNotFoundException",
  "Exception",
  "DirectoryNotFoundException",
  "DriveNotFoundException",
  "EndOfStreamException",
  "File",
  "FileFormatException",
  "FileLoadException",
  "FileNotFoundException",
  "InternalBufferOverflowException",
  "InvalidDataException",
  "IOException",
  "PathTooLongException",
  "PipeException",
  "SseEvent",
  "OptionalNullable",
];

const cSharpAllowedXMLInlineTags = [
  "see",
  "paramref",
  "b",
  "i",
  "u",
  "br/",
  "a",
  "para",
  "list",
  "listheader",
  "term",
  "description",
  "item",
  "c",
  "code",
  "example",
];

// @ts-ignore
function getSDKReservedKeywords() {
  // TODO: implement reservedNamespaces instead
  return [
    getScopeNamespace("shared"),
    getScopeNamespace("operations"),
    getScopeNamespace("hooks"),
    "utils",
    "url",
  ];
}

// @ts-ignore
function sanitizeItemType(
  typeDef: TypeDef,
  usageLocation: string,
  containingClass: string,
  isUsage: boolean,
) {
  return sanitizeType(
    typeDef.ItemType,
    typeDef.ContainsNull,
    usageLocation,
    containingClass,
    isUsage,
  );
}

// @ts-ignore
function sanitizeSDKAccess(sdk: SDK): string {
  const group = sdk.Group;

  if (!group) {
    return "";
  }

  let formattedName = "";
  if (group.includes(".") && ".".repeat(group.length) !== group) {
    const pieces = group.split(".");
    formattedName = sanitizeFieldName(pieces[0]);
    for (const i of pieces.slice(1)) {
      formattedName += "." + sanitizeFieldName(i);
    }
  } else {
    formattedName = sanitizeFieldName(group);
  }

  if (formattedName != "") {
    formattedName = `.${formattedName}`;
  }

  return formattedName;
}

registerTemplateFunc("sanitizeSDKAccess", sanitizeSDKAccess);

// @ts-ignore
function getUnionTypeName(className: string, outputLocation: string): string {
  let suffixes = ["Type", "Union", "OneOf"];

  const allTypes = getAllTypes();

  var findBestName = undefined;
  findBestName = (suffix: string): string => {
    if (suffixes.length == 0) {
      throw new Error(
        "Unable to find a suitable union type name. Contact Speakeasy for support.",
      );
    }

    const newSuffix = suffixes.shift();
    let unionTypeName = `${className}${newSuffix}${suffix}`;

    const foundConflict = allTypes.find((type) => {
      if (type.OutputLocation != outputLocation) {
        return false;
      }

      const otherClassName = sanitizeClassName(type.Name);

      if (otherClassName != unionTypeName) {
        return false;
      }

      return true;
    });

    if (foundConflict) {
      return findBestName(newSuffix);
    }

    return unionTypeName;
  };

  return findBestName("");
}

registerTemplateFunc("getUnionTypeName", getUnionTypeName);

// @ts-ignore
function sanitizeClassName(name: string, isInterface?: boolean): string {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (isInterface) {
    name = `I${name}`;
  }

  return name;
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

// @ts-ignore
function sanitizeSDKName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (name == getSDKNamespace() || getSDKReservedKeywords().includes(name)) {
    name = `${name}SDK`;
  }

  return name;
}

registerTemplateFunc("sanitizeSDKName", sanitizeSDKName);

// @ts-ignore
function sanitizeMethodName(operation: Operation, isAsync?: boolean): string {
  let id = operation.GetID();
  if (isAsync) {
    id = `${id}Async`;
  }

  return caser().ToPascal(sanitizeName(id));
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

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
function sanitizeFileName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  name = truncateName(name, 250);

  return name;
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => caser().ToPascal(sanitizeName(segment)))
    .join("/");
}

// @ts-ignore
function sanitizeFieldName(name: string, className: string = ""): string {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (className == name) {
    name = `${name}Value`;
  }

  return name;
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// @ts-ignore
function sanitizeField(field: FieldDef, parent?: TypeDef): string {
  return sanitizeFieldName(field.Name, sanitizeClassName(parent?.Name ?? ""));
}

registerTemplateFunc("sanitizeField", sanitizeField);

// @ts-ignore
function sanitizePrivateFieldName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToCamel(name);

  return name;
}

registerTemplateFunc("sanitizePrivateFieldName", sanitizePrivateFieldName);

// @ts-ignore
function sanitizeMethodParamName(name: string): string {
  name = sanitizePrivateFieldName(name);

  if (cSharpReservedKeywords.includes(name)) {
    name = `${name}P`;
  }

  return name;
}

registerTemplateFunc("sanitizeMethodParamName", sanitizeMethodParamName);

// @ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional: boolean = false,
  usageLocation: string = "",
  containingClass: string = "",
  isUsage: boolean = false,
): string {
  const typeName = sanitizeTypeName(
    typeDef,
    usageLocation,
    containingClass,
    isUsage,
  );

  return `${typeName}${optional ? "?" : ""}`;
}

registerTemplateFunc("sanitizeType", sanitizeType);

// @ts-ignore
function sanitizeTypeName(
  typeDef: TypeDef,
  usageLocation: string,
  containingClass: string,
  isUsage: boolean = false,
): string {
  switch (typeDef.Type.toString()) {
    case "enum":
    case "error":
    case "class":
    case "union":
      return sanitizeClass(
        typeDef,
        usageLocation,
        false,
        isUsage,
        containingClass,
      );
    case "string":
      return "string";
    case "date":
      if (useNodatime()) {
        addUsageImport("NodaTime");
        return "LocalDate";
      } else {
        addUsageImport("System");
        return "DateOnly";
      }
    case "date-time":
      return "DateTime";
    case "event-stream":
      addUsageImport("Utils.Sse", true);
      return `EventStream<${sanitizeItemType(
        typeDef,
        usageLocation,
        containingClass,
        isUsage,
      )}>`;
    case "integer":
      return "long";
    case "int32":
      return "int";
    case "bigint":
      addUsageImport("System.Numerics");
      return "BigInteger";
    case "number":
      return "double";
    case "float32":
      return "float";
    case "decimal":
      return "decimal";
    case "boolean":
      return "bool";
    case "bytes":
      return "byte[]";
    case "map":
      addUsageImport("System.Collections.Generic");
      return `Dictionary<string, ${sanitizeItemType(
        typeDef,
        usageLocation,
        containingClass,
        isUsage,
      )}>`;
    case "array":
      addUsageImport("System.Collections.Generic");
      return `List<${sanitizeItemType(
        typeDef,
        usageLocation,
        containingClass,
        isUsage,
      )}>`;
    case "any":
      // "dynamic" would work _better_ for end-users because properties can be accessed
      // easily, but "dynamic" is harder to serialize than "object"
      return "object";
    case "response":
      return "HttpResponseMessage";
    case "request":
      return "HttpRequestMessage";
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}

// @ts-ignore
function templateStatusCodeCheck(
  varName: string,
  statusCodes: string[],
  inverted: boolean = false,
): string {
  if (statusCodes.length === 0) {
    return inverted ? "true" : "false";
  }

  statusCodes = sanitizeStatusCodes(
    statusCodes.filter((code) => code !== "default"),
  );

  const checks: string[] = [];
  let reduce = statusCodes.length > 1;
  for (const statusCode of statusCodes) {
    if (statusCode.toUpperCase().includes("X")) {
      const codeRange = parseInt(statusCode[0], 10);
      const lower = `${codeRange}00`;
      const upper = `${codeRange + 1}00`;

      checks.push(
        inverted
          ? `(${varName} < ${lower} || ${varName} >= ${upper})`
          : `${varName} >= ${lower} && ${varName} < ${upper}`,
      );
      reduce = false;
    } else {
      checks.push(`${varName} ${inverted ? "!=" : "=="} ${statusCode}`);
    }
  }

  if (reduce) {
    return `${inverted ? "!" : ""}new List<int>{${statusCodes.join(
      ", ",
    )}}.Contains(${varName})`;
  }

  return checks.join(` ${inverted ? "&&" : "||"} `);
}

registerTemplateFunc("templateStatusCodeCheck", templateStatusCodeCheck);

// @ts-ignore
function getEnumName(value: string): string {
  let name = value.trim();

  if (name === "") {
    name = "Unknown";
  }

  name = sanitizeName(name);
  return caser().ToPascal(name);
}

registerTemplateFunc("getEnumName", getEnumName);

// @ts-ignore
function getEnumNamesFromValues(values: string[]): string[] {
  const enumNames = [];

  const names = {};
  if (!values) return enumNames;
  values.forEach((value) => {
    const name = getEnumName(value);
    if (!names[name]) {
      names[name] = 0;
    }

    names[name] += 1;
  });

  const seen = {};
  values.forEach((value) => {
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

    enumNames.push(sanitizeFieldName(name));
  });

  return enumNames;
}

registerTemplateFunc("getEnumNamesFromValues", getEnumNamesFromValues);

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  _usageLocation: string = "",
  definition: boolean = false, // set to true to use fully qualified namespace
  isUsage: boolean = false,
  containingClass: string = "",
  suffix: string = "",
): string {
  let className = sanitizeClassName(typeDef.Name);
  if (suffix) {
    className += sanitizeClassName(suffix);
  }

  if (definition) {
    return `${getModelNamespace(typeDef.OutputLocation, true)}.${className}`;
  }

  if (
    !typeDef.Truncated &&
    (className == containingClass ||
      doesTypeConflict(className, typeDef.OutputLocation))
  ) {
    return `${getModelNamespace(
      typeDef.OutputLocation,
      isUsage || usingGlobalImports() || hasNamespaceConflict(),
    )}.${className}`;
  }

  addUsageImportForType(typeDef);
  return className;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

// @ts-ignore
function sanitizeXMLAttribute(attribute: string): string {
  return attribute.replace(
    /(\s+\w+)=(["'])(.*?)\2/g,
    (_match, key, quote, value) => {
      const sanitized = value
        .replaceAll(/&/g, "&amp;")
        .replaceAll(/"/g, "&quot;")
        .replaceAll(/'/g, "&apos;")
        .replaceAll(/</g, "&lt;")
        .replaceAll(/>/g, "&gt;");
      return `${key}=${quote}${sanitized}${quote}`;
    },
  );
}

// @ts-ignore
function sanitizeComments(comment: string): string {
  // escape illegal XML characters
  comment = comment
    .replaceAll(/&/g, "&amp;")
    .replaceAll(/</g, "&lt;")
    .replaceAll(/>/g, "&gt;")
    .replace(/&lt;(\/*)(\w+)(.*?)&gt;/g, (_match, slash, tagName, attrs) =>
      cSharpAllowedXMLInlineTags.includes(tagName)
        ? `<${slash}${tagName}${sanitizeXMLAttribute(attrs)}>`
        : `&lt;${slash}${tagName}${attrs}&gt;`,
    );

  // convert markdown links to HTML links
  comment = comment.replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_match, desc, url) => {
    return `<a href="${url}">${desc}</a>`;
  });

  // sanitize whitespace
  comment = comment
    .replaceAll("*/", "* /")
    .replaceAll("\uFEFF", "")
    .replaceAll(/\u00A0/g, " ")
    .replaceAll(/\u200b/g, " ")
    .trim()
    .replaceAll("\r\n", "\n")
    .replaceAll(/<br>/g, "<br/>")
    .replaceAll(/(?<!>)\n/g, "<br/>\n");

  // add missing trailing period
  if (/[\w)]$/.test(comment)) {
    comment += ".";
  }

  return comment;
}

registerTemplateFunc("sanitizeComments", sanitizeComments);

// @ts-ignore
function sanitizeJsonPath(jsonpath: string): string {
  const path = jsonpath
    // Note (alexa): couldn't find a library that correctly parses this expression
    // This is not a long-term solution
    .replaceAll(/\(\@\.length ?- ?1\)/gi, "-1:");
  return path;
}

registerTemplateFunc("sanitizeJsonPath", sanitizeJsonPath);

// @ts-ignore
function sanitizeGlobalSecurityFieldName(fieldName: string): string {
  return sanitizePrivateFieldName(fieldName);
}

// @ts-ignore
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName, "Security");
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

registerTemplateFunc("sanitizeUnionTypeName", sanitizeUnionTypeName);

// @ts-ignore
function sanitizePrivateClassName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToGoCamel(name);

  if (
    cSharpReservedKeywords.includes(name) ||
    getSDKReservedKeywords().includes(name)
  ) {
    name = `${name}T`;
  }

  return name;
}

registerTemplateFunc("sanitizePrivateClassName", sanitizePrivateClassName);

// @ts-ignore
function sanitizeUnionCast(typeDef: TypeDef, outputLocation: string): string {
  switch (typeDef.Type.toString()) {
    case "string":
      return "string";
    case "enum":
      return sanitizeClass(typeDef, outputLocation, false);
    default:
      throw new Error(`Unsupported type: ${typeDef.Type.toString()}`);
  }
}

registerTemplateFunc("sanitizeUnionCast", sanitizeUnionCast);

// @ts-ignore
function sanitizeVariableName(name: string): string {
  name = sanitizeName(name);

  name = caser().ToGoCamel(name);

  if (
    cSharpReservedKeywords.includes(name) ||
    getSDKReservedKeywords().includes(name)
  ) {
    name = `${name}Var`;
  }

  return name;
}

registerTemplateFunc("sanitizeVariableName", sanitizeVariableName);

// @ts-ignore
function isReferenceType(typeDef: TypeDef): boolean {
  return [
    "bytes",
    "map",
    "array",
    "object",
    "union",
    "class",
    "string",
  ].includes(typeDef.Type.toString());
}

registerTemplateFunc("isReferenceType", isReferenceType);

// @ts-ignore
function sanitizePackageId(name: string): string {
  if (!context.Global.Config.DisableNamespacePascalCasingApr2024) {
    return name
      .split(".")
      .map((part) => {
        return sanitizeFileName(part);
      })
      .join(".");
  }

  // ValidationRegex for `packageName` used to allow `-`
  return name.replace(/-/g, "_");
}

function sanitizeDotnetVersion(dotnetVersion: string) {
  // "netX.0" -> "X.0.0"
  return dotnetVersion.replace(/^net/, "") + ".0";
}

registerTemplateFunc("sanitizeDotnetVersion", sanitizeDotnetVersion);
