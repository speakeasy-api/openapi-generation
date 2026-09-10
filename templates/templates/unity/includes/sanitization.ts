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

// @ts-ignore
function sanitizeMap(typeDef: TypeDef, usageLocation: string) {
  return `Dictionary<string, ${sanitizeType(
    typeDef.ItemType,
    false,
    usageLocation,
  )}>`;
}

registerTemplateFunc("sanitizeMap", sanitizeMap);

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
function sanitizeClassName(name: string, isInterface: boolean): string {
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

  if (name == getSDKNamespace()) {
    name = `${name}SDK`;
  }

  return name;
}

registerTemplateFunc("sanitizeSDKName", sanitizeSDKName);

// @ts-ignore
function sanitizeMethodName(operation: any, isAsync: boolean) {
  let id = operation.GetID();
  if (isAsync) {
    id = `${id}Async`;
  }

  return caser().ToPascal(sanitizeName(id));
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

// @ts-ignore
function sanitizeFileName(name: string) {
  name = sanitizeName(name);
  return caser().ToPascal(name);
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function sanitizeFieldName(name: string, className: string = "") {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (className == name) {
    name = `${name}Value`;
  }

  return name;
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// @ts-ignore
function sanitizePrivateFieldName(name: string) {
  name = sanitizeName(name);
  name = caser().ToCamel(name);

  return name;
}

registerTemplateFunc("sanitizePrivateFieldName", sanitizePrivateFieldName);

// @ts-ignore
function sanitizeMethodParamName(name: string) {
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
  optional?: boolean,
  usageLocation: string = "",
  containingClass = "",
): string {
  switch (typeDef.Type.toString()) {
    case "enum":
    case "error":
    case "class":
    case "union":
      let name = sanitizeClassName(typeDef.Name);

      if (
        (doesTypeConflict(typeDef) || name == containingClass) &&
        !typeDef.Truncated
      ) {
        name = `${getModelNamespace(
          typeDef.OutputLocation,
          usageLocation != "usage",
        )}.${name}`;
      }

      return name;
    case "string":
      return "string";
    case "date":
      return "DateOnly";
    case "date-time":
      return "DateTime";
    case "integer":
      return "long";
    case "int32":
      return "int";
    case "bigint":
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
      return sanitizeMap(typeDef, usageLocation);
    case "array":
      return `List<${sanitizeType(typeDef.ItemType, false, usageLocation)}>`;
    case "any":
      // "dynamic" would work _better_ for end-users because properties can be accessed
      // easily, but "dynamic" is harder to serialize than "object"
      return "object";
    case "response":
      return "UnityWebRequest";
    case "response-stream":
      return "MemoryQueueBufferStream";
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}

registerTemplateFunc("sanitizeType", sanitizeType);

// @ts-ignore
function templateStatusCode(statusCodes: string[]) {
  const checks: string[] = [];
  let reduce = statusCodes.length > 1;
  for (const statusCode of statusCodes) {
    if (statusCode.toUpperCase().includes("X")) {
      let codeRange = parseInt(statusCode[0], 10);

      checks.push(
        `httpCode >= ${codeRange}00 && httpCode < ${codeRange + 1}00`,
      );
      reduce = false;
    } else {
      checks.push(`httpCode == ${statusCode}`);
    }
  }
  if (!reduce) {
    return checks.join(" || ");
  } else {
    return `new List<int>{${statusCodes.join(", ")}}.Contains(httpCode)`;
  }
}
registerTemplateFunc("templateStatusCode", templateStatusCode);

// @ts-ignore
function templateStatusCodes(statusCodes: string[]): string {
  if (statusCodes.includes("default")) {
    return "";
  }
  const conditions = statusCodes
    .map((condition) => {
      const wildcardIndex = condition.toUpperCase().indexOf("X");
      if (wildcardIndex > 0) {
        const base = parseInt(condition.substring(0, wildcardIndex));
        const power = Math.pow(10, 3 - wildcardIndex);
        const min = base * power;
        const max = (base + 1) * power;
        return `(httpCode >= ${min} && httpCode < ${max})`;
      }

      return `(httpCode == ${condition})`;
    })
    .join(" || ");
  return `if(${conditions})`;
}

registerTemplateFunc("templateStatusCodes", templateStatusCodes);

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

  values.forEach((value) => {
    let name = getEnumName(value);
    if (names[name] > 1) {
      name = `${name}${caser().ToPascal(getCasing(value))}`;
    }

    enumNames.push(sanitizeFieldName(name));
  });

  return enumNames;
}

registerTemplateFunc("getEnumNamesFromValues", getEnumNamesFromValues);

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
): string {
  let className = sanitizeClassName(typeDef.Name);
  if (usageLocation != typeDef.OutputLocation && !definition) {
    className = `${getModelNamespace(
      typeDef.OutputLocation,
      true,
    )}.${className}`;
  }

  return className;
}

// @ts-ignore
function sanitizeComments(comment: string): string {
  // escape illegal XML characters
  comment = comment
    .replaceAll("\r\n", "\n")
    .replaceAll(/&/g, "&amp;")
    .replaceAll(/</g, "&lt;")
    .replaceAll(/>/g, "&gt;")
    .replaceAll(/"/g, "&quot;")
    .replaceAll(/'/g, "&apos;");

  // convert markdown links to HTML links
  let elements = comment.match(/\[.*?\]\(.*?\)/g);
  if (elements != null && elements.length > 0) {
    for (var el of elements) {
      let txt = el.match(/\[(.*?)\]/)[1]; //get only the txt
      let url = el.match(/\((.*?)\)/)[1]; //get only the link
      comment = comment.replace(el, '<a href="' + url + '">' + txt + "</a>");
    }
  }

  return comment
    .replaceAll("*/", "* /")
    .replaceAll("\uFEFF", "")
    .replaceAll(/\u00A0/g, " ")
    .replaceAll(/\u200b/g, " ")
    .replaceAll(/\n/g, "<br/>\n");
}

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
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => caser().ToPascal(sanitizeName(segment)))
    .join("/");
}

// @ts-ignore
function sanitizeGlobalSecurityFieldName(fieldName: string): string {
  return sanitizePrivateFieldName(fieldName);
}
// @ts-ignore
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName, "Security");
}

function sanitizeDictionaryKey(key: string): string {
  return sanitizeName(key).toLowerCase();
}
registerTemplateFunc("sanitizeDictionaryKey", sanitizeDictionaryKey);

// @ts-ignore
function getSDKReservedKeywords() {
  return [
    getAccessNamespace("shared"),
    getAccessNamespace("operations"),
    getAccessNamespace("utils"),
    "url",
  ];
}

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
        return sanitizeClassName(part);
      })
      .join(".");
  }

  // ValidationRegex for `packageName` used to allow `-`
  return name.replace(/-/g, "_");
}
