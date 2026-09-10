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
  "error",
];

// @ts-ignore
function getSDKReservedKeywords() {
  return ["components", "operations", "types", "utils"];
}

// @ts-ignore
function sanitizeFieldName(name: string): string {
  name = sanitizeName(name);

  return caser().ToGoPascal(name);
}
registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

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

  if (
    reservedGoKeywords.includes(name) ||
    getSDKReservedKeywords().includes(name)
  ) {
    name = `${name}_`;
  }

  return name;
}

registerTemplateFunc("sanitizePrivateFieldName", sanitizePrivateFieldName);

// @ts-ignore
function sanitizeVariableName(name: string): string {
  name = sanitizeName(name);

  name = caser().ToGoCamel(name);

  if (
    reservedGoKeywords.includes(name) ||
    getSDKReservedKeywords().includes(name)
  ) {
    name = `${name}Var`;
  }

  return name;
}

registerTemplateFunc("sanitizeVariableName", sanitizeVariableName);

// @ts-ignore
function sanitizeClassName(name: string): string {
  name = sanitizeName(name);

  if (getSDKReservedKeywords().includes(name)) {
    name = `${name}Obj`;
  }

  return caser().ToGoPascal(name);
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

// @ts-ignore
function sanitizePrivateClassName(name: string): string {
  name = sanitizeName(name);
  name = caser().ToGoCamel(name);

  if (
    reservedGoKeywords.includes(name) ||
    getSDKReservedKeywords().includes(name)
  ) {
    name = `${name}T`;
  }

  return name;
}

registerTemplateFunc("sanitizePrivateClassName", sanitizePrivateClassName);

// @ts-ignore
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
  name = sanitizeFile(name, "");

  name = name.toLowerCase();

  name = truncateName(name, 250);

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

function sanitizePathHandlerName(name: string): string {
  return caser().ToGoCamel(sanitizeName(name));
}

function sanitizeTestHandlerMethodName(
  operation: Operation,
  testName: string,
  stepIdx: number,
): string {
  return caser().ToGoCamel(
    `test_${sanitizeName(operation.BaseOperation.OriginalID)}_${sanitizeName(
      testName,
    )}${stepIdx}`,
  );
}

// @ts-ignore
function getPointerTypes(): string[] {
  return [
    "bytes",
    "response-stream",
    "map",
    "any",
    "array",
    "set",
    "bigint",
    "decimal",
    "event-stream",
    "jsonl",
  ];
}

// @ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional?: boolean,
  usageLocation: string = "",
  addImports = true,
): string {
  let type = (() => {
    switch (typeDef.Type.toString()) {
      case "union":
      case "enum":
      case "error":
      case "class":
        return sanitizeClass(typeDef, usageLocation, false, addImports);
      case "string":
        return "string";
      case "date":
        addImports && addImport("types");
        return "types.Date";
      case "date-time":
        addImports && addImport("time");
        return "time.Time";
      case "integer":
        return "int64";
      case "int32":
        return "int";
      case "bigint":
        addImports && addImport("math/big");
        return (optional ? "" : "*") + "big.Int"; // Always has to be a pointer due to big.Int pointer receiver on the MarshalJSON method which does mean we lose requiredness for this type
      case "number":
        return "float64";
      case "float32":
        return "float32";
      case "decimal":
        addImports && addImport("github.com/ericlagergren/decimal");
        return (optional ? "" : "*") + "decimal.Big"; // Always has to be a pointer due to decimal.Big pointer receiver on the MarshalText method which does mean we lose requiredness for this type
      case "boolean":
        return "bool";
      case "bytes":
        return "[]byte";
      case "request-stream":
        // NOTE: This is stronger typed than the go (v1) target as this target
        // is self-contained and can own the full implementation.
        addImports && addImport("io");
        return "io.Reader";
      case "response-stream":
        addImports && addImport("io");
        return "io.ReadCloser";
      case "map":
        return sanitizeMap(typeDef, usageLocation, addImports);
      case "set":
      case "array":
        return (
          "[]" +
          sanitizeType(
            typeDef.ItemType,
            typeDef.ContainsNull,
            usageLocation,
            addImports,
          )
        );
      case "any":
        return "any";
      case "request":
        addImports && addImport("net/http");
        return (optional ? "" : "*") + "http.Request"; // Always has to be a pointer due to http.Request being returned as a pointer
      case "response":
        addImports && addImport("net/http");
        return (optional ? "" : "*") + "http.Response"; // Always has to be a pointer due to http.Response being returned as a pointer
      case "event-stream":
        addImports && addImport("stream");
        return (
          (optional ? "" : "*") +
          `stream.EventStream[${sanitizeType(
            typeDef.ItemType,
            false,
            usageLocation,
            addImports,
          )}]`
        );
      case "jsonl":
        addImports && addImport("jsonl");
        return (
          (optional ? "" : "*") +
          `jsonl.JsonLStream[${sanitizeType(
            typeDef.ItemType,
            false,
            usageLocation,
            addImports,
          )}]`
        );
      default:
        throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
    }
  })();
  return templateOptional(typeDef, optional) + type;
}

registerTemplateFunc("sanitizeType", sanitizeType);

// @ts-ignore
function sanitizeFieldType(fieldDef: FieldDef, usageLocation: string): string {
  if (fieldDef.Optional && fieldDef.Nullable) {
    addImport("optionalnullable");
    const innerType = sanitizeType(fieldDef.Type, false, usageLocation);
    return `optionalnullable.OptionalNullable[${innerType}]`;
  }
  return sanitizeType(
    fieldDef.Type,
    fieldDef.Optional || fieldDef.Nullable,
    usageLocation,
  );
}

registerTemplateFunc("sanitizeFieldType", sanitizeFieldType);

// @ts-ignore
function templateValueWrapper(fieldDef: FieldDef, value: string): string {
  if (!(fieldDef.Optional && fieldDef.Nullable)) {
    return value;
  }
  if (value === "nil") {
    // Return nil unchanged — a nil OptionalNullable[T] (nil map) means "unset",
    // which is omitted from JSON by omitempty. Using From[T](nil) would create
    // a non-nil map (set-to-null state) that serializes as "field": null.
    return "nil";
  }
  addImport("optionalnullable");
  // Reference types (map, slice, any, etc.) produce non-pointer values,
  // so we need types.Pointer() to get *T for From[T](*T).
  // Other types (string, bool, int, enum, etc.) already produce pointer
  // values via helpers like types.String() or .ToPointer().
  if (templateOptional(fieldDef.Type, true) === "") {
    addImport("types");
    return `optionalnullable.From(types.Pointer(${value}))`;
  }
  return `optionalnullable.From(${value})`;
}

registerTemplateFunc("templateValueWrapper", templateValueWrapper);

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
  addImports = true,
) {
  let prefix = "";
  if (
    !definition &&
    typeDef.OutputLocation != usageLocation &&
    !typeDef.Truncated
  ) {
    addImports && addImport(typeDef.OutputLocation, true);
    prefix = sanitizeModelPackageName(typeDef.OutputLocation) + ".";
  }

  return prefix + sanitizeClassName(typeDef.Name);
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

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
        "Of" +
        caser().ToGoPascal(sanitizeUnionTypeName(typeDef.ItemType))
      );
    default:
      return typeDef.Type.toString();
  }
}

registerTemplateFunc("sanitizeUnionTypeName", sanitizeUnionTypeName);

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
function sanitizeMap(typeDef: TypeDef, outputLocation: string, addImports) {
  return `map[string]${sanitizeType(
    typeDef.ItemType,
    typeDef.ContainsNull,
    outputLocation,
    addImports,
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
    case "response-stream":
    case "array":
    case "set":
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
  additionalContext?: TemplateValueContext,
): string {
  return sanitizeType(typeDef, false, "usage");
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
function templateNullValue(fieldDef: FieldDef): string {
  return "nil";
}

// @ts-ignore
function templateFieldDelimiter(): string {
  return ",";
}

// @ts-ignore
function templateConstValue(
  fieldDef: FieldDef,
  usageLocation: string,
  skipPointerWrapping = false,
): string {
  if (fieldDef.Const.Value === null) {
    return "nil";
  }

  switch (fieldDef.Type.Type.toString()) {
    case "enum":
      return templateEnumValue(
        fieldDef,
        fieldDef.Type.Enum.Values.indexOf(fieldDef.Const?.Value.toString()),
        {
          skipPointerWrapping,
          outputLocation:
            fieldDef.Type.OutputLocation != usageLocation
              ? fieldDef.Type.OutputLocation
              : "inline",
        },
      );
    case "string":
      return templateStringValue(fieldDef.Const?.Value, fieldDef, {
        skipPointerWrapping,
      });
    case "date":
      return templateDateValue(fieldDef.Const?.Value, fieldDef, {
        skipPointerWrapping,
      });
    case "date-time":
      return templateDateTimeValue(fieldDef.Const?.Value, fieldDef, {
        skipPointerWrapping,
      });
    case "bigint":
    case "int32":
    case "integer":
      return templateIntValue(fieldDef.Const?.Value, fieldDef, {
        skipPointerWrapping,
      });
    case "number":
    case "float32":
    case "decimal":
      return templateFloatValue(fieldDef.Const?.Value, fieldDef, {
        skipPointerWrapping,
      });
    case "boolean":
      return templateBoolValue(fieldDef.Const?.Value, fieldDef, {
        skipPointerWrapping,
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

// @ts-ignore
function templateStringValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  value = quote(value).replaceAll("{{", `{{"{{"}}`);

  const needsPointer =
    (fieldDef.Optional || (fieldDef.Nullable && value !== "nil")) &&
    !additionalContext?.skipPointerWrapping;
  if (needsPointer) {
    addImport(getTypesLocation(), true);
    return `types.String(${value})`;
  }

  return `${value}`;
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  additionalContext?: TemplateValueContext,
): string {
  let enumNames = getEnumNames(fieldDef.Type);

  let enumValue = enumNames[idx];

  if (additionalContext?.outputLocation != "inline") {
    addImport(fieldDef.Type.OutputLocation, true);
    enumValue = `${getAccessNamespace(
      "usage",
      fieldDef.Type.Scope.toString(),
      fieldDef.Type.OutputLocation,
    )}${enumValue}`;
  }

  if (
    (fieldDef.Optional || fieldDef.Nullable) &&
    !additionalContext?.skipPointerWrapping
  ) {
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

  if (
    (fieldDef.Optional || fieldDef.Nullable) &&
    !additionalContext?.skipPointerWrapping
  ) {
    addImport(getTypesLocation());
    return `types.Bool(${strValue})`;
  }

  return strValue;
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
  addImport(getTypesLocation(), true);
  return fieldDef.Optional || fieldDef.Nullable
    ? additionalContext?.skipPointerWrapping
      ? `types.MustDateFromString("${value}")`
      : `types.MustNewDateFromString("${value}")`
    : `types.MustDateFromString("${value}")`;
}

// @ts-ignore
function templateDateTimeValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  addImport(getTypesLocation(), true);
  return fieldDef.Optional || fieldDef.Nullable
    ? additionalContext?.skipPointerWrapping
      ? `types.MustTimeFromString("${value}")`
      : `types.MustNewTimeFromString("${value}")`
    : `types.MustTimeFromString("${value}")`;
}

// @ts-ignore
function templateIntValue(
  value: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  let t = fieldDef.Type.Type.toString();
  if (t == "any") {
    t = "integer";
  }

  switch (t) {
    case "int32":
      if (
        (fieldDef.Optional || fieldDef.Nullable) &&
        !additionalContext?.skipPointerWrapping
      ) {
        addImport(getTypesLocation(), true);
        return `types.Int(${value})`;
      }
      return `${value}`;
    case "integer":
      if (
        (fieldDef.Optional || fieldDef.Nullable) &&
        !additionalContext?.skipPointerWrapping
      ) {
        addImport(getTypesLocation(), true);
        return `types.Int64(${value})`;
      }
      return `${value}`;
    case "bigint":
      const isNumber = typeof value === "number";
      if (isNumber) {
        value = value.toString();
      }

      if (!isNumber || value.length > `999999999999999999`.length) {
        addImport(getTypesLocation(), true);
        return `types.MustNewBigIntFromString("${value}")`;
      } else {
        addImport("math/big");
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
      if (
        (fieldDef.Optional || fieldDef.Nullable) &&
        !additionalContext?.skipPointerWrapping
      ) {
        addImport(getTypesLocation(), true);
        return `types.Float32(${value})`;
      }
      return `${value}`;
    case "number":
      if (
        (fieldDef.Optional || fieldDef.Nullable) &&
        !additionalContext?.skipPointerWrapping
      ) {
        addImport(getTypesLocation(), true);
        return `types.Float64(${value})`;
      }
      return `${value}`;
    case "decimal":
      addImport(getTypesLocation(), true);
      return `types.MustNewDecimalFromString("${value}")`;
  }
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
function getGolangPackage(): string {
  return "mockserver/internal/sdk";
}

registerTemplateFunc("getGolangPackage", getGolangPackage);

// @ts-ignore
function sanitizeComments(comment: string): string {
  return comment
    .replaceAll("\uFEFF", "")
    .replaceAll("\r\n", "\n")
    .replaceAll("{{", `{{"{{"}}`); // Need to sanitize these so they don't conflict with the templates
}

// @ts-ignore
function sanitizeSDKPackageName(): string {
  return "sdk";
}

registerTemplateFunc("sanitizeSDKPackageName", sanitizeSDKPackageName);

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
function resolveEnumNameConflicts(
  baseNames: string[],
  outputLocation: string,
): string[] {
  // Get all existing type names in the same scope for conflict detection
  const allTypes = getAllTypes();
  const existingTypeNames = new Set<string>();

  // Collect all type names in the same output location (scope)
  allTypes.forEach((t) => {
    if (t.OutputLocation === outputLocation) {
      existingTypeNames.add(sanitizeClassName(t.Name));
    }
  });

  // Apply conflict resolution strategy to each base name
  return baseNames.map((baseName) => {
    // Try original name first
    if (!existingTypeNames.has(baseName)) {
      return baseName;
    }

    // Try with "Value" suffix
    const withValue = `${baseName}Value`;
    if (!existingTypeNames.has(withValue)) {
      return withValue;
    }

    // Try with "EnumValue" suffix
    const withEnumValue = `${baseName}EnumValue`;
    if (!existingTypeNames.has(withEnumValue)) {
      return withEnumValue;
    }

    // If all attempts fail, throw an error
    throw new Error(
      `Unable to resolve enum name conflict for "${baseName}" in scope "${outputLocation}". ` +
        `All variants (${baseName}, ${withValue}, ${withEnumValue}) conflict with existing types.`,
    );
  });
}

// @ts-ignore
function getUnionEnumNamesFromDiscriminator(
  unionType: string,
  discriminatorMapping: any[],
  outputLocation: string,
): string[] {
  const baseNames = discriminatorMapping.map(
    (mapping) =>
      `${unionType}${sanitizeClassName(getDiscriminatorDisplayName(mapping))}`,
  );
  return resolveEnumNameConflicts(baseNames, outputLocation);
}

registerTemplateFunc(
  "getUnionEnumNamesFromDiscriminator",
  getUnionEnumNamesFromDiscriminator,
);

// @ts-ignore
function getUnionEnumNamesFromAssociatedTypes(
  unionType: string,
  associatedTypes: any[],
  outputLocation: string,
): string[] {
  const baseNames = associatedTypes.map((type) => {
    const name = sanitizeUnionTypeName(type);
    return `${unionType}${sanitizeClassName(name)}`;
  });
  return resolveEnumNameConflicts(baseNames, outputLocation);
}

registerTemplateFunc(
  "getUnionEnumNamesFromAssociatedTypes",
  getUnionEnumNamesFromAssociatedTypes,
);

// @ts-ignore
function getEnumNames(type: TypeDef): string[] {
  const baseNames =
    type.Enum?.Names.length > 0
      ? type.Enum.Names.map(
          (n) => `${sanitizeFieldName(type.Name)}${getEnumName(n)}`,
        )
      : getEnumNamesFromValues(type.Enum.Values).map(
          (n) => `${sanitizeFieldName(type.Name)}${n}`,
        );

  return resolveEnumNameConflicts(baseNames, type.OutputLocation);
}

registerTemplateFunc("getEnumNames", getEnumNames);
