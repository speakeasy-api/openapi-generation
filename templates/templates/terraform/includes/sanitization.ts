// @ts-ignore
const reservedGoKeywords = [
  "break",
  "default",
  "func",
  "interface",
  "select",
  "case",
  "defer",
  "go",
  "map",
  "struct",
  "chan",
  "else",
  "goto",
  "package",
  "switch",
  "const",
  "fallthrough",
  "if",
  "range",
  "type",
  "continue",
  "for",
  "import",
  "return",
  "var",
  "strings",
  "string",
  "bytes",
  "bool",
  "uint",
  "uint8",
  "uint32",
  "uint64",
  "int",
  "int8",
  "int32",
  "int64",
  "float",
  "float32",
  "float64",
  "complex64",
  "complex128",
  "any",
  "context",
];

// @ts-ignore
const sdkReservedKeywords = [
  "big", // Go standard library math/big package
  "json", // Go standard library encoding/json package
  "operations", // SDK operations package
  "options", // SDK options package
  "s", // Common SDK receiver variable in generated Go code
  "shared", // SDK shared package
  "time", // Go standard library time package
  "types", // SDK and terraform-plugin-framework types packages
  "utils", // SDK utils package
];

/** Copy of go template sdkReservedKeywords. This must be kept in sync to
 *  prevent naming misalignment for class names. */
const goSDKReservedKeywords = [
  "operations",
  "option",
  "options",
  "s",
  "shared",
  "timeout",
  "url",
  "utils",
];

// @ts-ignore
function sanitizeFieldName(name: string): string {
  name = sanitizeName(name);

  return caser().ToGoPascal(name);
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// @ts-ignore
function sanitizeSDKFieldName(name: string): string {
  const regex = /^[a-zA-Z]+(\.[a-zA-Z]+)+$/;
  if (regex.test(name)) {
    const pieces = name.split(".");
    name = pieces[pieces.length - 1];
  }

  name = sanitizeName(name);

  return caser().ToGoPascal(name);
}

registerTemplateFunc("sanitizeSDKFieldName", sanitizeSDKFieldName);

// @ts-ignore
function sanitizeErrorFieldName(name: string): string {
  name = sanitizeName(name);

  name = caser().ToGoPascal(name);

  if (name == "Error") {
    name = "Error_";
  }

  return name;
}

registerTemplateFunc("sanitizeErrorFieldName", sanitizeErrorFieldName);

// @ts-ignore
function sanitizePrivateFieldName(name: string): string {
  name = sanitizeName(name);

  name = caser().ToGoCamel(name);

  if (reservedGoKeywords.includes(name) || sdkReservedKeywords.includes(name)) {
    name = `${name}_`;
  }

  return name;
}

registerTemplateFunc("sanitizePrivateFieldName", sanitizePrivateFieldName);

// @ts-ignore
function sanitizeVariableName(name: string): string {
  name = sanitizeName(name);

  name = caser().ToGoCamel(name);

  if (reservedGoKeywords.includes(name) || sdkReservedKeywords.includes(name)) {
    name = `${name}Var`;
  }

  return name;
}

registerTemplateFunc("sanitizeVariableName", sanitizeVariableName);

// @ts-ignore
function sanitizeClassName(name: string): string {
  name = sanitizeName(name);

  if (goSDKReservedKeywords.includes(name.toLowerCase())) {
    name = `${name}Obj`;
  }

  return caser().ToGoPascal(name);
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

// @ts-ignore
function sanitizePrivateClassName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToGoCamel(name);

  if (reservedGoKeywords.includes(name) || sdkReservedKeywords.includes(name)) {
    name = `${name}T`;
  }

  return name;
}

registerTemplateFunc("sanitizePrivateClassName", sanitizePrivateClassName);

const getID = (o: Operation): string => {
  if (o.GetID) {
    return o.GetID();
  }

  if (o.Extensions.MethodNameOverride != "") {
    return o.Extensions.MethodNameOverride;
  }

  return o.ID;
};

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  return caser().ToGoPascal(sanitizeName(getID(operation)));
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

// @ts-ignore
function sanitizeFileName(name: string): string {
  name = sanitizeFile(name, "");

  name = name.toLowerCase();

  if (goSDKReservedKeywords.includes(name)) {
    name = `${name}_`;
  }

  return name;
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

// @ts-ignore
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => sanitizeFile(segment, "").toLowerCase())
    .join("/");
}

// @ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional?: boolean,
  scope?: string,
): string {
  let type = (() => {
    switch (typeDef.Type.toString()) {
      case "union":
      case "enum":
      case "error":
      case "class":
        return sanitizeClass(typeDef, scope, false);
      case "string":
        return "string";
      case "date":
        return "types.Date";
      case "date-time":
        return "time.Time";
      case "integer":
        return "int64";
      case "int32":
        return "int";
      case "bigint":
        return (optional ? "" : "*") + "big.Int"; // Always has to be a pointer due to big.Int pointer receiver on the MarshalJSON method which does mean we lose requiredness for this type
      case "number":
        return "float64";
      case "float32":
        return "float32";
      case "decimal":
        return (optional ? "" : "*") + "decimal.Big"; // Always has to be a pointer due to decimal.Big pointer receiver on the MarshalText method which does mean we lose requiredness for this type
      case "boolean":
        return "bool";
      case "bytes":
        return "[]byte";
      case "map":
        return sanitizeMap(typeDef, scope);
      case "array":
      case "set":
        return (
          "[]" + sanitizeType(typeDef.ItemType, typeDef.ContainsNull, scope)
        );
      case "any":
        return "interface{}";
      case "response":
        return (optional ? "" : "*") + "http.Response"; // Always has to be a pointer due to http.Response being returned as a pointer
      default:
        throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
    }
  })();
  return templateOptional(typeDef, optional) + type;
}

registerTemplateFunc("sanitizeType", sanitizeType);

// @ts-ignore
function sanitizeClass(typeDef: TypeDef, scope: string, definition: boolean) {
  let className = sanitizeClassName(typeDef.Name);

  let prefix = "";
  if (!definition) {
    if (typeDef.Scope.toString() != scope.toString()) {
      prefix = sanitizeScopeName(typeDef.Scope) + ".";
    }
  }

  return prefix + className;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

function sanitizeScopeName(scope: string): string {
  switch (scope.toString()) {
    case "errors":
      return getDefaultErrorFileName();
    default:
      return scope;
  }
}

registerTemplateFunc("sanitizeScopeName", sanitizeScopeName);

// @ts-ignore
function templateStatusCode(statusCode: string): string {
  if (statusCode.toUpperCase().includes("X")) {
    let codeRange = parseInt(statusCode[0], 10);

    return `httpRes.StatusCode >= ${codeRange}00 && httpRes.StatusCode < ${
      codeRange + 1
    }00`;
  }

  return `httpRes.StatusCode == ${statusCode}`;
}

registerTemplateFunc("templateStatusCode", templateStatusCode);

// @ts-ignore
function templateConstName(name: string, prefix: string): string {
  name = sanitizeName(name);

  return caser().ToGoPascal(prefix + "_" + name);
}

registerTemplateFunc("templateConstName", templateConstName);

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
    case "set":
      return (
        typeDef.Type.toString() +
        "_Of_" +
        sanitizeUnionTypeName(typeDef.ItemType)
      );
    default:
      return typeDef.Type.toString();
  }
}

registerTemplateFunc("sanitizeUnionTypeName", sanitizeUnionTypeName);

// @ts-ignore
function sanitizeUnionCast(typeDef: TypeDef, scope: string): string {
  switch (typeDef.Type.toString()) {
    case "string":
      return "string";
    case "enum":
      return sanitizeClass(typeDef, scope, false);
    default:
      throw new Error(`Unsupported type: ${typeDef.Type.toString()}`);
  }
}

registerTemplateFunc("sanitizeUnionCast", sanitizeUnionCast);

// @ts-ignore
function sanitizeMap(typeDef: TypeDef, scope: string) {
  return `map[string]${sanitizeType(
    typeDef.ItemType,
    typeDef.ContainsNull,
    scope,
  )}`;
}

// @ts-ignore
function templateOptional(typeDef: TypeDef, optional: boolean) {
  if (!optional) {
    return "";
  }

  switch (typeDef.Type.toString()) {
    case "bytes":
    case "map":
    case "any":
    case "set":
    case "array":
      return "";
  }

  return "*";
}

// @ts-ignore
function templateFieldDeclaration(
  field: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${sanitizeFieldName(field.Name)}: `;
}

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext: TemplateValueContext = { scope: "operations" },
): string {
  let scope = additionalContext.scope;

  if (typeDef.Scope.toString() === scope) {
    scope = "shared";
  }

  return sanitizeType(typeDef, false, scope);
}

// @ts-ignore
function templateBracket(
  typeDef: TypeDef,
  openingBracket: boolean,
  additionalContext?: TemplateValueContext,
): string {
  if (openingBracket) {
    return "{";
  }

  return "}";
}

// @ts-ignore
function templateOptionalSymbol(): string {
  return "&";
}

// @ts-ignore
function templateIndent(indent: number): string {
  return "    ".repeat(indent);
}

// @ts-ignore
function templateFieldDelimiter(): string {
  return ",";
}

// @ts-ignore
function templateConstValue(fieldDef: FieldDef, scope: Scope): string {
  if (fieldDef.Const.Value === null) {
    return "nil";
  }

  switch (fieldDef.Type.Type.toString()) {
    case "enum":
      return templateEnumValue(
        fieldDef,
        fieldDef.Type.Enum.Values.indexOf(fieldDef.Const?.Value.toString()),
        {
          scope:
            fieldDef.Type.Scope.toString() != scope.toString()
              ? fieldDef.Type.Scope
              : "",
        },
      );
    case "string":
      return templateStringValue(fieldDef.Const?.Value, fieldDef, {
        useTypesPackage: true,
      });
    case "date":
      return templateDateValue(fieldDef.Const?.Value, fieldDef);
    case "date-time":
      return templateDateTimeValue(fieldDef.Const?.Value, fieldDef);
    case "bigint":
    case "int32":
    case "integer":
      return templateIntValue(fieldDef.Const?.Value, fieldDef, {
        useTypesPackage: true,
      });
    case "number":
    case "float32":
    case "decimal":
      return templateFloatValue(fieldDef.Const?.Value, fieldDef, {
        useTypesPackage: true,
      });
    case "boolean":
      return templateBoolValue(fieldDef.Const?.Value, fieldDef, {
        useTypesPackage: true,
      });
    case "bytes":
    default:
      throw new Error(
        `unsupported const type: ${fieldDef.Type.Type.toString()}`,
      );
  }
}

registerTemplateFunc("templateConstValue", templateConstValue);

// @ts-ignore
function templateArrayValue(value: any): string {
  return value;
}

// @ts-ignore
function templateMapValue(key: string, value: any): string {
  return `"${key}": ${value}`;
}

function templateBuiltinString(value: string): string {
  value = value
    .trim()
    .replaceAll(/(`+)/g, '` + "$1" +`')
    .replaceAll("\n", '` + "\\n" +\n`')
    .replace(/(({{)|(}}))/g, '{{"$1"}}');
  return "`" + value + "`";
}

// @ts-ignore
function templateStringValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  value = value.replaceAll(/"/g, '\\"').replaceAll("{{", `{{"{{"}}`);
  return fieldDef.Optional
    ? `${
        additionalContext?.useTypesPackage ? "types" : sanitizeSDKPackageName()
      }.String("${value}")`
    : `"${value}"`;
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  additionalContext?: TemplateValueContext,
): string {
  let enumNames = getEnumNames(fieldDef.Type);

  let enumValue = enumNames[idx];

  if (additionalContext?.scope) {
    enumValue = `${additionalContext?.scope}.${enumValue}`;
  }

  if (fieldDef.Optional) {
    enumValue += `.ToPointer()`;
  }

  return enumValue;
}

// @ts-ignore
function templateBoolValue(
  value: boolean,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  const strValue = value ? "true" : "false";

  return fieldDef.Optional
    ? `${
        additionalContext?.useTypesPackage ? "types" : sanitizeSDKPackageName()
      }.Bool(${strValue})`
    : strValue;
}

// @ts-ignore
function templateByteValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `[]byte("${value.replace(/"/g, '\\"')}")`;
}

// @ts-ignore
function templateDateValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return fieldDef.Optional || fieldDef.Nullable
    ? `types.MustNewDateFromString("${value}")`
    : `types.MustDateFromString("${value}")`;
}

// @ts-ignore
function templateDateTimeValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return fieldDef.Optional || fieldDef.Nullable
    ? `types.MustNewTimeFromString("${value}")`
    : `types.MustTimeFromString("${value}")`;
}

// @ts-ignore
function templateIntValue(
  value: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let t = fieldDef.Type.Type.toString();
  if (t == "any") {
    t = "integer";
  }

  switch (t) {
    case "int32":
      return fieldDef.Optional || fieldDef.Nullable
        ? `${
            additionalContext?.useTypesPackage
              ? "types"
              : sanitizeSDKPackageName()
          }.Int(${value})`
        : `${value}`;
    case "integer":
      return fieldDef.Optional || fieldDef.Nullable
        ? `${
            additionalContext?.useTypesPackage
              ? "types"
              : sanitizeSDKPackageName()
          }.Int64(${value})`
        : `${value}`;
    case "bigint":
      let val = `${value}`;
      if (val.length > `999999999999999999`.length) {
        return `types.MustNewBigIntFromString("${value}")`;
      } else {
        return `big.NewInt(${value})`;
      }
  }
}

// @ts-ignore
function templateFloatValue(
  value: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let t = fieldDef.Type.Type.toString();
  if (t == "any") {
    t = "number";
  }

  switch (t) {
    case "float32":
      return fieldDef.Optional
        ? `${
            additionalContext?.useTypesPackage
              ? "types"
              : sanitizeSDKPackageName()
          }.Float32(${value})`
        : `${value}`;
    case "number":
      return fieldDef.Optional
        ? `${
            additionalContext?.useTypesPackage
              ? "types"
              : sanitizeSDKPackageName()
          }.Float64(${value})`
        : `${value}`;
    case "decimal":
      return `types.MustNewDecimalFromString("${value}")`;
  }
}

// @ts-ignore
function templateSecurityValue(
  value: any,
  optional: boolean,
  additionalContext?: TemplateValueContext,
) {
  if (!optional) {
    return `"${value}"`;
  }

  return `${sanitizeSDKPackageName()}.String("${value}")`;
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

  for (const value of values) {
    let name = getEnumName(value);
    if (names[name] > 1) {
      name = `${name}${caser().ToGoPascal(getCasing(value))}`;
    }

    enumNames.push(sanitizeFieldName(name));
  }

  return enumNames;
}

// @ts-ignore
function getEnumName(value: string): string {
  let name = value.trim();

  if (name === "") {
    name = "Unknown";
  }

  name = sanitizeName(name);

  return caser().ToGoPascalIgnoreAcronyms(name);
}

// @ts-ignore
function getEnumNames(type: TypeDef): string[] {
  if (type.Enum.Names.length > 0) {
    return type.Enum.Names.map(
      (n) => `${sanitizeFieldName(type.Name)}${getEnumName(n)}`,
    );
  } else {
    return getEnumNamesFromValues(type.Enum.Values).map(
      (n) => `${sanitizeFieldName(type.Name)}${n}`,
    );
  }
}

// @ts-ignore
function getGolangPackage(): string {
  let versionParts = context.Global.Config.SDKVersion.split(".");

  let majorVersion = parseInt(versionParts[0], 10);

  if (majorVersion > 1) {
    return `${context.Global.Config.PackageName}/v${majorVersion}`;
  } else {
    return `${context.Global.Config.PackageName}`;
  }
}

registerTemplateFunc("getGolangPackage", getGolangPackage);

// @ts-ignore
function sanitizeComments(comment: string): string {
  return comment.replaceAll("\uFEFF", "").replaceAll("\r\n", "\n");
}

// @ts-ignore
function sanitizeAcceptEnumKey(acceptType: string): string {
  return `AcceptHeaderEnum${caser().ToPascal(
    sanitizeName(acceptType.split(";")[0]),
  )}`;
}

registerTemplateFunc("sanitizeAcceptEnumKey", sanitizeAcceptEnumKey);

// @ts-ignore
function sanitizeAcceptEnumValue(acceptType: string): string {
  return `${acceptType.split(";")[0]}`;
}

registerTemplateFunc("sanitizeAcceptEnumValue", sanitizeAcceptEnumValue);

// @ts-ignore
function sanitizeSDKPackageName(): string {
  const parts = context.Global.Config.PackageName.split("/");

  let name = sanitizeName(parts[parts.length - 1]);

  return caser().ToGoPascal(name).toLowerCase();
}

registerTemplateFunc("sanitizeSDKPackageName", sanitizeSDKPackageName);

// @ts-ignore
function getUnionTypeName(className: string, scope: Scope): string {
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
      // TODO: this will need to become bucket aware
      if (type.Scope != scope) {
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
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName);
}

// @ts-ignore
function convertFloatToType(value: string, type: TypeDef): string {
  switch (type.Type.toString()) {
    case "integer":
      return `int64(${value})`;
    case "string":
      return `strconv.FormatFloat(${value}, 'f', 0, 64)`;
  }
  return "";
}
registerTemplateFunc("convertFloatToType", convertFloatToType);

/**
 * Returns a best effort sanitized pagination output property name, which has
 * any leading "$." removed and then the remaining is sanitized as a field name.
 * There may be use cases where the JSON Path contains further path information that would
 * need to be removed, but this is not currently handled until there is a need.
 */
function sanitizePaginationOutputName(
  paginationOutputProperty: string,
): string {
  const result = paginationOutputProperty.replace(/^\$\./, "");

  return sanitizeFieldName(result);
}
