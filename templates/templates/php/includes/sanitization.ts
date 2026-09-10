const reservedKeywords = [
  "abstract",
  "and",
  "as",
  "break",
  "callable",
  "case",
  "catch",
  "class",
  "clone",
  "const",
  "continue",
  "declare",
  "default",
  "do",
  "echo",
  "else",
  "elseif",
  "enddeclare",
  "endfor",
  "endforeach",
  "endif",
  "endswitch",
  "endwhile",
  "extends",
  "final",
  "finally",
  "fn",
  "for",
  "foreach",
  "function",
  "global",
  "goto",
  "if",
  "implements",
  "include",
  "instanceof",
  "insteadof",
  "interface",
  "match",
  "namespace",
  "new",
  "or",
  "print",
  "private",
  "protected",
  "public",
  "readonly",
  "require",
  "return",
  "static",
  "switch",
  "throw",
  "trait",
  "try",
  "use",
  "var",
  "while",
  "xor",
  "yield",
  "yield",
  "from",
  "parent",
  "list",
  "self",
  "empty",
  "object",
  "hooks",
];

// @ts-ignore
function sanitizeFieldName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToCamel(name);

  return name;
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  let name = sanitizeName(operation.GetID());
  name = caser().ToCamel(name);

  return name;
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

// @ts-ignore
function sanitizeClassName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (reservedKeywords.includes(name.toLowerCase())) {
    name = `${name}T`;
  }

  return name;
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

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
  name = sanitizeFile(sanitizeClassName(name), "");

  name = truncateName(name, 250);

  return name;
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function templateConstName(name: string, prefix: string): string {
  name = sanitizeName(name);
  return caser().ToSNAKE(`${prefix}_${name}`);
}

registerTemplateFunc("templateConstName", templateConstName);

// @ts-ignore
function sanitizeMapKey(key: string): string {
  key = sanitizeName(key);
  key = caser().ToSnake(key);

  return key;
}

registerTemplateFunc("sanitizeMapKey", sanitizeMapKey);

// Helper function to check if a type has a custom model namespace extension
function hasCustomModelNamespace(type: TypeDef): boolean {
  // Must be truthy (not null, undefined, or empty string)
  return !!type.Extensions?.ModelNamespace;
}

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
  qualification = Qualification.TYPE,
): string {
  let className = sanitizeClassName(typeDef.Name);

  const fullNameSpace = getModelNamespace(typeDef.OutputLocation);
  const parts = fullNameSpace.split("\\");
  const localPath = parts[parts.length - 1];

  // Check if this is a custom model namespace (using the extension for reliability)
  const isCustomNamespace = hasCustomModelNamespace(typeDef);

  switch (qualification) {
    case Qualification.DOCSTRING:
      return `\\${fullNameSpace}\\${className}`;
    case Qualification.TYPE:
      if (usageLocation === typeDef.OutputLocation) {
        return className;
      }
      // For custom namespaces, use fully qualified names to avoid PHP linter issues
      if (isCustomNamespace) {
        return `\\${fullNameSpace}\\${className}`;
      }
      addImportInline(fullNameSpace);
      return `${localPath}\\${className}`;
    case Qualification.USAGE:
      // For custom namespaces, use fully qualified names to avoid PHP linter issues
      if (isCustomNamespace) {
        return `\\${fullNameSpace}\\${className}`;
      }
      addImportInline(fullNameSpace);
      return `${localPath}\\${className}`;
    case Qualification.ANNOTATION:
      return `\\${fullNameSpace}\\${className}`;
    case Qualification.UNQUALIFIED:
      return className;
    default:
      if (typeDef.OutputLocation != usageLocation && !definition) {
        return `\\${fullNameSpace}\\${className}`;
      }
      break;
  }

  return className;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

// @ts-ignore
function templateStatusCode(statusCodes: string[]): string {
  let results = [];
  for (const statusCode of statusCodes) {
    if (statusCode.toUpperCase().includes("X")) {
      let codeRange = parseInt(statusCode[0], 10);

      results.push(
        `($httpResponse->getStatusCode() >= ${codeRange}00 && $httpResponse->getStatusCode() < ${
          codeRange + 1
        }00)`,
      );
    } else {
      results.push(`$httpResponse->getStatusCode() === ${statusCode}`);
    }
  }
  return results.join(" or ");
}

registerTemplateFunc("templateStatusCode", templateStatusCode);

enum Qualification {
  DOCSTRING = "docstring",
  TYPE = "type",
  USAGE = "usage",
  ANNOTATION = "annotation",
  UNQUALIFIED = "unqualified",
}

// @ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional?: boolean,
  nullable?: boolean,
  usageLocation: string = "",
  qualification = Qualification.TYPE,
  typesToShallowEmbed: Set<string> = new Set(),
): string {
  let type = (() => {
    switch (typeDef.Type.toString()) {
      case "enum":
        return sanitizeEnum(typeDef, usageLocation, qualification);
      case "error":
      case "class":
        return sanitizeClass(typeDef, usageLocation, false, qualification);
      case "union":
        return sanitizeUnionType(
          typeDef,
          usageLocation,
          qualification,
          typesToShallowEmbed,
        );
      case "string":
        return "string";
      case "date":
        if (qualification == Qualification.ANNOTATION) {
          return "Brick\\DateTime\\LocalDate";
        } else {
          addImportInline("Brick\\DateTime\\LocalDate");
          return "LocalDate";
        }
      case "date-time":
        return "\\DateTime";
      case "integer":
      case "int32":
        return "int";
      case "bigint":
        return "\\Brick\\Math\\BigInteger";
      case "decimal":
        return "\\Brick\\Math\\BigDecimal";
      case "number":
      case "float32":
        return "float";
      case "boolean":
        return "bool";
      case "bytes":
        return "string";
      case "map":
      case "array":
        return sanitizeArray(
          typeDef,
          usageLocation,
          qualification,
          typesToShallowEmbed,
        );
      case "any":
        return "mixed";
      case "response":
        return "\\Psr\\Http\\Message\\ResponseInterface";
      case "request":
        return "\\Psr\\Http\\Message\\RequestInterface";
      default:
        throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
    }
  })();
  if ((optional || nullable) && type != "mixed" && typeDef.Type != "union") {
    if (qualification == Qualification.ANNOTATION) {
      type = `${type}|null`;
    } else {
      type = `?${type}`;
    }
  } else if (
    (optional || nullable) &&
    typeDef.Type == "union" &&
    type != "mixed"
  ) {
    // should be ?type if it's a union with a single resolved type, otherwise type1|type2|null
    if (type.includes("|") || qualification == Qualification.ANNOTATION) {
      type = `${type}|null`;
    } else {
      type = `?${type}`;
    }
  }

  return type;
}

registerTemplateFunc("sanitizeType", sanitizeType);

// @ts-ignore
function sanitizeMethodParamName(name: string) {
  return sanitizeFieldName(name);
}

registerTemplateFunc("sanitizeMethodParamName", sanitizeMethodParamName);

// @ts-ignore
function sanitizeUnionType(
  typeDef: TypeDef,
  usageLocation: string,
  qualification = Qualification.TYPE,
  typesToShallowEmbed: Set<string>,
): string {
  if (typesToShallowEmbed.has(typeDef.Name)) {
    return "array<string, mixed>";
  }
  const isCircular = detectCircularUnion(typeDef);
  const associatedTypes = typeDef.AssociatedTypes.map(
    (t: TypeDef & { ParentName?: string }) => {
      return sanitizeType(
        t,
        false,
        false,
        usageLocation,
        qualification,
        isCircular
          ? typesToShallowEmbed.add(typeDef.Name)
          : typesToShallowEmbed,
      );
    },
  ).filter((value, index, array) => array.indexOf(value) === index);
  if (
    associatedTypes.includes("mixed") &&
    qualification !== Qualification.DOCSTRING
  ) {
    return "mixed";
  }
  if (associatedTypes.length > 1) {
    return `${associatedTypes.join("|")}`;
  } else {
    return associatedTypes[0];
  }
}
registerTemplateFunc("sanitizeUnionType", sanitizeUnionType);

function detectCircularUnion(typeDef: TypeDef): boolean {
  const circularReferences = typeDef.AssociatedTypes.filter(
    (t: TypeDef & { ParentName?: string }) => {
      const isCollection =
        t.Type.toString() === "array" || t.Type.toString() === "map";
      if (isCollection) {
        return (
          t.ItemType.Type.toString() === "union" &&
          t.ItemType.Name === typeDef.Name
        );
      }
    },
  );
  return circularReferences.length > 0 ? true : false;
}

function toPascal(name: string): string {
  return caser().ToPascal(name);
}

registerTemplateFunc("toPascal", toPascal);

function sanitizeArray(
  typeDef: TypeDef,
  usageLocation: string,
  qualification: Qualification,
  typesToShallowEmbed: Set<string>,
): string {
  if (
    qualification == Qualification.USAGE ||
    qualification == Qualification.DOCSTRING ||
    qualification == Qualification.ANNOTATION
  ) {
    let globalNamespace = context.Global.Config.Namespace;
    if (typeDef.ItemType.Type.toString() == "union") {
      for (const type of typeDef.ItemType.AssociatedTypes) {
        const modelNamespace = getModelNamespace(type.OutputLocation);
        if (
          globalNamespace != modelNamespace &&
          !typesToShallowEmbed.has(typeDef.Name) &&
          qualification !== Qualification.ANNOTATION &&
          !hasCustomModelNamespace(type)
        ) {
          addImportInline(modelNamespace);
        }
      }
    } else {
      const modelNamespace = getModelNamespace(typeDef.ItemType.OutputLocation);
      if (
        globalNamespace != modelNamespace &&
        qualification !== Qualification.ANNOTATION &&
        !hasCustomModelNamespace(typeDef.ItemType)
      ) {
        addImportInline(modelNamespace);
      }
    }
    const itemType = sanitizeType(
      typeDef.ItemType,
      false,
      typeDef.ContainsNull,
      usageLocation,
      qualification,
      typesToShallowEmbed,
    );

    if (typeDef.Type.toString() == "map") {
      return `array<string, ${itemType}>`;
    } else {
      return `array<${itemType}>`;
    }
  } else {
    return "array";
  }
}

// @ts-ignore
function sanitizeEnum(
  typeDef: TypeDef,
  usageLocation: string,
  qualification: Qualification,
): string {
  let cls = sanitizeClass(typeDef, usageLocation, false, qualification);

  return cls;
}

// @ts-ignore
function sanitizeComments(comment: string): string {
  return comment
    .replaceAll(/(\r)([^\n])/g, "\n$2")
    .replaceAll("\r\n", "\n")
    .replaceAll("*/", "* /")
    .replaceAll("<?php", "< ?php")
    .replaceAll("{{", `{{"{{"}}`); // Need to sanitize these so they don't conflict with the templates
}

// @ts-ignore
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${sanitizeFieldName(fieldDef.Name)}: `;
}

// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  return "null";
}
// @ts-ignore
function templateFieldDelimiter(): string {
  return ",";
}

// @ts-ignore
function templateArrayValue(val: any): string {
  return val;
}

// @ts-ignore
function templateMapValue(key: string, val: any): string {
  return `'${key}' => ${val}`;
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
  switch (typeDef.Type.toString()) {
    case "error":
    case "class":
      return opening ? "(" : ")";
    default:
      return opening ? "[" : "]";
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
  return `'${val
    .replaceAll(/\\/g, "\\\\")
    .replaceAll(/'/g, "\\'")
    .replaceAll("{{", `{{"{{"}}`)
    .replaceAll(/\r/g, "\\r")
    .replaceAll(/\n/g, "\\n")}'`;
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  additionalContext?: TemplateValueContext,
): string {
  let globalNamespace = context.Global.Config.Namespace;
  const modelNamespace = getModelNamespace(fieldDef.Type.OutputLocation);
  let enumNames = getEnumNames(fieldDef.Type);
  if (globalNamespace != modelNamespace) {
    addImportInline(modelNamespace);
  }
  let qualification = Qualification.TYPE;
  if (additionalContext?.isUsage == true) {
    qualification = Qualification.USAGE;
  }
  return `${sanitizeClass(
    fieldDef.Type,
    fieldDef.Type.OutputLocation,
    false,
    qualification,
  )}::${enumNames[idx]}`;
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
  return `'${val.replace(/'/g, "\\'")}'`;
}

// @ts-ignore
function templateDateValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  addImportInline("Brick\\DateTime\\LocalDate");
  return `LocalDate::parse('${val}')`;
}

// @ts-ignore
function templateDateTimeValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  addImportInline(getScopeNamespace("utils", true));
  return `Utils\\Utils::parseDateTime('${val}')`;
}

// @ts-ignore
function templateIntValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (fieldDef.Type.Type.toString() == "bigint") {
    addImportInline("Brick\\Math\\BigInteger");
    return `BigInteger::of('${val}')`;
  }
  return `${val}`;
}

// @ts-ignore
function templateFloatValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (fieldDef.Type.Type.toString() == "decimal") {
    addImportInline("Brick\\Math\\BigDecimal");
    return `BigDecimal::of('${val}')`;
  }
  return `${val}`;
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

    enumNames.push(name);
  }

  return enumNames;
}

// @ts-ignore
function getEnumName(value: string): string {
  let name = value.trim();

  if (name === "") {
    name = "Unknown";
  }

  name = caser().ToPascal(sanitizeName(name));
  if (name === "Class") {
    name = "Class_";
  }
  return name;
}

registerTemplateFunc("getEnumName", getEnumName);

function getErrorCodeField(type: TypeDef): FieldDef | undefined {
  for (const field of type.Fields) {
    if (
      field.Name.toLowerCase() === "code" ||
      field.Name.toLowerCase() === "statuscode"
    ) {
      return field;
    }
  }
}

registerTemplateFunc("getErrorCodeField", getErrorCodeField);

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  switch (typeDef.Type.toString()) {
    case "array":
    case "map":
      return "";
    default:
      let globalNamespace = context.Global.Config.Namespace;
      let type = sanitizeType(typeDef, false, false, "", Qualification.USAGE);

      const modelNamespace = getModelNamespace(typeDef.OutputLocation);
      if (globalNamespace != modelNamespace) {
        addImportInline(modelNamespace);
      }
      return `new ${type}`;
  }
}

// @ts-ignore
function sanitizeSDKAccess(sdk: SDK, prefix: string = "."): string {
  // setting default prefix to "." is to prevent the first character from getting stripped in templateAvailableOperations
  const group = sdk.Group;

  if (!group) {
    return "";
  }

  let formattedName = "";
  if (group.includes(".") && ".".repeat(group.length) !== group) {
    const pieces = group.split(".");
    formattedName = sanitizeFieldName(pieces[0]);
    for (const i of pieces.slice(1)) {
      formattedName += "->" + sanitizeFieldName(i);
    }
  } else {
    formattedName = sanitizeFieldName(group);
  }

  if (formattedName != "") {
    formattedName = `${prefix}${formattedName}`;
  }

  return formattedName;
}
registerTemplateFunc("sanitizeSDKAccess", sanitizeSDKAccess);

// @ts-ignore
function isSecurityFieldOption(field: FieldDef): boolean {
  for (const annotation of field.Annotations) {
    if (annotation.Type().toString() == "security") {
      if ((annotation as SecurityAnnotation).Option) {
        return true;
      }
    }
  }
  return false;
}
registerTemplateFunc("isSecurityFieldOption", isSecurityFieldOption);

// @ts-ignore
function hasSecurityFieldOption(type: TypedDef): boolean {
  for (const field of type.Fields) {
    if (isSecurityFieldOption(field)) {
      return true;
    }
  }
  return false;
}
registerTemplateFunc("hasSecurityFieldOption", hasSecurityFieldOption);

// @ts-ignore
function isSecuritySpecial(field: FieldDef): boolean {
  if (!field) {
    return false;
  }
  let isSpecial = false;
  for (const annotation of field.Annotations) {
    if (annotation.Type().toString() == "security") {
      if ((annotation as SecurityAnnotation).SubType == "basic") {
        isSpecial = true;
        break;
      }
    }
  }
  return isSpecial;
}
registerTemplateFunc("isSecuritySpecial", isSecuritySpecial);

// @ts-ignore
function sanitizeJsonPath(jsonpath: string): string {
  /*
   * Support for pagination is specified by the client through a `x-speakeasy-pagination` entry in the spec file.
   * This uses a JSONPath expression to specify the path of pagination fields in the response.
   *
   * For legacy reasons, to access the last value in the `resultArray` field, our documentation encourages the syntax: $.resultArray[(@.length-1)]
   * However, this syntax is not supported by the `JsonPath-PHP` library, which is used to parse JSONPath expressions.
   * Instead we run the foillowing replace command to match the expected JsonPath syntax: `$.resultArray[-1:]`
   */
  const path = jsonpath.replaceAll(/\(\@\.length ?- ?1\)/gi, "-1:");
  return path;
}

registerTemplateFunc("sanitizeJsonPath", sanitizeJsonPath);

// @ts-ignore
function sanitizeURLPath(path: string): string {
  // Escape backslashes first, then single quotes for PHP single-quoted strings
  return path.replaceAll(/\\/g, "\\\\").replaceAll(/'/g, "\\'");
}

registerTemplateFunc("sanitizeURLPath", sanitizeURLPath);

// @ts-ignore
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => caser().ToPascal(sanitizeName(segment)))
    .join("/");
}

// @ts-ignore
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName);
}
