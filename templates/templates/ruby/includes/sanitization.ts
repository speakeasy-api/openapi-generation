const rubyReservedKeywords = [
  "__ENCODING__",
  "__LINE__",
  "__FILE__",
  "BEGIN",
  "END",
  "alias",
  "and",
  "begin",
  "break",
  "case",
  "class",
  "def",
  "defined?",
  "do",
  "else",
  "elsif",
  "end",
  "ensure",
  "false",
  "for",
  "if",
  "in",
  "module",
  "next",
  "nil",
  "not",
  "or",
  "redo",
  "rescue",
  "retry",
  "return",
  "self",
  "super",
  "then",
  "true",
  "undef",
  "unless",
  "until",
  "when",
  "while",
  "yield",
  "fields",
  "str", // not technically reserved in ruby, added here for more consistency with python
  "sig", // Sorbet's sig method for type signatures
];

// Names that collide with Sorbet's type system or SDK internals when used as class/module names.
// "T" shadows Sorbet's T module (T::Enum, T.nilable, etc.)
// "Sig" shadows Sorbet's sig method used for type signatures
// "Models" shadows the Models module namespace used by the SDK
const rubySorbetReservedClassNames: Record<string, string> = {
  T: "TT",
  Sig: "SigValue",
  Models: "ModelsModel",
};

const rubyMethodArgumentReservedKeywords = [
  "retries",
  "server_url",
  "timeout_ms",
  "url_override",
  "accept_header_override",
  "http_headers",
  "key",
  "value",
];

// @ts-ignore
function sanitizeFieldName(name: string): string {
  name = sanitizeName(name);
  if (name.endsWith("_")) {
    name = name.substring(0, name.length - 1);
  }
  name = caser().ToSnake(name);

  if (rubyReservedKeywords.includes(name)) {
    name += "_";
  }

  return name;
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// @ts-ignore
function sanitizeMethodArgumentName(name: string): string {
  let sanitized = sanitizeFieldName(name);

  if (rubyMethodArgumentReservedKeywords.includes(sanitized)) {
    sanitized += "_";
  }

  return sanitized;
}

registerTemplateFunc("sanitizeMethodArgumentName", sanitizeMethodArgumentName);

/**
 * Returns a deduplicated sanitized field name for use within a model class.
 *
 * When multiple fields in the same model normalize to the same sanitized name
 * (e.g. `db_pass` and `dbPass` both become `db_pass`), the first field (in array
 * order) keeps the plain name and subsequent colliders get a `_2`, `_3`, ... suffix.
 */
// @ts-ignore
function deduplicatedFieldName(name: string, fields: FieldDef[]): string {
  const sanitized = sanitizeFieldName(name);

  // Build ordered list of original names that collide on this sanitized value.
  const colliders: string[] = [];
  for (const f of fields) {
    if (sanitizeFieldName(f.Name) === sanitized) {
      colliders.push(f.Name);
    }
  }

  // No collision — return as-is.
  if (colliders.length <= 1) {
    return sanitized;
  }

  // Find this field's position among the colliders.
  const idx = colliders.indexOf(name);
  if (idx <= 0) {
    // First occurrence keeps the plain name.
    return sanitized;
  }

  // Subsequent occurrences get a numeric suffix.
  return `${sanitized}_${idx + 1}`;
}

registerTemplateFunc("deduplicatedFieldName", deduplicatedFieldName);

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  return caser().ToSnake(sanitizeName(operation.GetID()));
}

registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

// @ts-ignore
function sanitizePrivateMethodName(name: string): string {
  let sanitized = caser().ToSnake(sanitizeName(name));

  if (rubyReservedKeywords.includes(sanitized)) {
    sanitized += "_";
  }

  return sanitized;
}
registerTemplateFunc("sanitizePrivateMethodName", sanitizePrivateMethodName);

// @ts-ignore
function sanitizeClassName(name: string): string {
  name = name.replace(/\./g, "_");
  name = sanitizeName(name);
  let className = caser().ToPascal(name);

  if (className in rubySorbetReservedClassNames) {
    className = rubySorbetReservedClassNames[className];
  }

  return className;
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

// @ts-ignore
function sanitizeModuleName(name: string): string {
  name = name.replace(/\./g, "_");
  name = sanitizeName(name);
  let moduleName = caser().ToPascal(name);

  if (moduleName in rubySorbetReservedClassNames) {
    moduleName = rubySorbetReservedClassNames[moduleName];
  }

  return moduleName;
}

registerTemplateFunc("sanitizeModuleName", sanitizeModuleName);

// Simple hash function (djb2) for generating short, deterministic suffixes.
function djb2Hash(str: string): string {
  let hash = 5381;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) + hash + str.charCodeAt(i)) & 0xffffffff;
  }
  return (hash >>> 0).toString(16);
}

// Max base filename length (without extension). Ruby gems have a hard limit on
// filenames within .gem archives (~100 chars for the relative path). The path
// prefix (e.g. "lib/sdk_name/models/components/") can consume 30-40 chars, so
// we cap the base name to leave room.
const MAX_FILENAME_BASE_LENGTH = 80;

// @ts-ignore
function sanitizeFileName(name: string): string {
  name = name.replace(/\./g, "_");
  name = sanitizeFile(name, "_").toLowerCase();

  if (rubyReservedKeywords.includes(name)) {
    name += "_";
  }

  if (name.length > MAX_FILENAME_BASE_LENGTH) {
    const hash = djb2Hash(name).slice(0, 6);
    name = name.substring(0, MAX_FILENAME_BASE_LENGTH - 7) + "_" + hash;
  }

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
function sanitizeOutputLocation(outputLocation: string): string {
  if (!outputLocation) return outputLocation;
  return outputLocation
    .split("/")
    .map((segment) => caser().ToSnake(sanitizeName(segment)))
    .join("/");
}

// @ts-ignore
function sanitizeComments(
  comment: string,
  indentDepth: number,
  commentFirstLine: boolean,
): string {
  comment = comment.replaceAll(`"`, '\\"');
  let lines = comment.replaceAll("\r\n", "\n").split("\n");

  lines = lines.map((line, index) => {
    let commentPrefix = "    ".repeat(indentDepth);
    if (index == 0 && !commentFirstLine) {
      commentPrefix = "";
    }

    return commentPrefix + "# " + line.trim();
  });

  return lines.join("\n");
}

registerTemplateFunc("sanitizeComments", sanitizeComments);

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  _scope: string,
  definition: boolean,
): string {
  let className = sanitizeClassName(typeDef.Name);
  const fullNamespace = getModelNamespace(typeDef.OutputLocation);
  if (definition) {
    return className;
  } else if (sanitizeModuleName(typeDef.Scope) === "SDK") {
    return `${getScopeNamespace("serverVariables")}::${className}`;
  }
  return `${fullNamespace}::${className}`;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

// @ts-ignore
function templateStatusCode(statusCodes: string[]): string {
  const checks: string[] = [];
  let reduce = statusCodes.length > 1;
  for (const statusCode of statusCodes) {
    if (statusCode.toUpperCase().includes("X")) {
      let codeRange = parseInt(statusCode[0], 10);

      checks.push(
        `r.status >= ${codeRange}00 && r.status < ${codeRange + 1}00`,
      );
      reduce = false;
    } else {
      checks.push(`r.status == ${statusCode}`);
    }
  }
  if (!reduce) {
    return checks.join(" || ");
  } else {
    return `[${statusCodes.join(", ")}].include?(r.status)`;
  }
}

registerTemplateFunc("templateStatusCode", templateStatusCode);

//@ts-ignore
function sanitizeType(
  typeDef: TypeDef,
  optional?: boolean,
  nullable?: boolean,
  scope?: string,
  typesToShallowEmbed: Set<string> = new Set(),
): string {
  if (!sorbetEnabled()) {
    return sanitizeCrystallineType(
      typeDef,
      optional,
      nullable,
      scope,
      typesToShallowEmbed,
    );
  }

  let type = (() => {
    switch (typeDef.Type.toString()) {
      case "enum":
      case "class":
      case "error":
        return sanitizeClass(typeDef, "", false);
      case "union":
        return sanitizeUnionType(typeDef, typesToShallowEmbed);
      case "string":
        return "::String";
      case "date":
        return "::Date";
      case "date-time":
        return "::DateTime";
      case "integer":
      case "int32":
        return "::Integer";
      case "number":
      case "float32":
        return "::Float";
      case "boolean":
        return "T::Boolean";
      case "bytes":
        return "::String";
      case "map":
        return sanitizeMap(typeDef, scope, typesToShallowEmbed);
      case "array":
        return `T::Array[${sanitizeType(
          typeDef.ItemType,
          false,
          false,
          scope,
          typesToShallowEmbed,
        )}]`;
      case "any":
        return "::Object";
      case "response":
        return "::Faraday::Response";
      case "request":
        return "::Faraday::Request";
      case "event-stream":
        return `::${sanitizeModuleName(
          context.Global.Config.Module,
        )}::Utils::EventStream`;
      case "jsonl":
        return `::${sanitizeModuleName(
          context.Global.Config.Module,
        )}::Utils::JsonLStream`;
      default:
        throw new Error(`Unknown type: "${typeDef.Type.toString()}"`);
    }
  })();
  return templateOptional(type, optional, nullable);
}

registerTemplateFunc("sanitizeType", sanitizeType);

function sanitizeUnionType(
  typeDef: TypeDef,
  typesToShallowEmbed: Set<string>,
): string {
  if (typesToShallowEmbed.has(typeDef.Name)) {
    return "T::Hash[Symbol, Object]";
  }
  const isCircular = detectCircularUnion(typeDef);

  const associatedTypes = typeDef.AssociatedTypes.map(
    (t: TypeDef & { ParentName?: string }) => {
      return sanitizeType(
        t,
        false,
        false,
        "",
        isCircular
          ? typesToShallowEmbed.add(typeDef.Name)
          : typesToShallowEmbed,
      );
    },
  ).filter((value, index, array) => array.indexOf(value) === index);
  Array.from(typesToShallowEmbed.values()).length;
  if (associatedTypes.length > 1) {
    return `T.any(${associatedTypes.join(", ")})`;
  } else {
    return associatedTypes[0];
  }
}

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
// @ts-ignore
function templateAuxPath(ctx: Context): string {
  return caser().ToSnake(ctx.Global.Config.Module);
}

registerTemplateFunc("templateAuxPath", templateAuxPath);

// @ts-ignore
function sanitizeMap(
  typeDef: TypeDef,
  scope: string,
  typesToShallowEmbed: Set<string>,
): string {
  return `T::Hash[Symbol, ${sanitizeType(
    typeDef.ItemType,
    false,
    false,
    scope,
    typesToShallowEmbed,
  )}]`;
}

// @ts-ignore
function templateOptional(
  type: string,
  optional: boolean,
  nullable: boolean,
): string {
  if (optional || nullable) {
    return `T.nilable(${type})`;
  }

  return type;
}

// @ts-ignore
function templateFieldDeclaration(
  fieldDef: TypeDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return `${sanitizeFieldName(fieldDef.Name)} = `;
}

// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  return "nil";
}

// @ts-ignore
function templateFieldDelimiter(): string {
  return ",";
}

// Override common templateObject to strip trailing comma from last field
// Ruby's Style/TrailingCommaInArguments disallows trailing commas in .new() args
// @ts-ignore
function templateObject(
  fieldDef: FieldDef,
  example: any,
  additionalContext?: TemplateValueContext,
): string {
  let fields = [];

  additionalContext = transformAdditionalContext(fieldDef, additionalContext);

  let remainingExample = { ...example };

  let additionalPropertiesField = undefined;
  for (const field of fieldDef.Type.Fields) {
    if (field.IsAdditionalProperties) {
      additionalPropertiesField = field;
      continue;
    }

    let fieldExample = remainingExample[originalFieldName(field)];
    delete remainingExample[originalFieldName(field)];

    const fieldUsage = templateFieldUsage(
      field,
      1,
      fieldDef.Type,
      fieldExample,
      additionalContext,
    );
    if (fieldUsage) {
      fields.push(fieldUsage + templateFieldDelimiter());
    }
  }

  if (additionalPropertiesField) {
    const additionalPropertiesExample =
      typeof remainingExample === "object" &&
      Object.keys(remainingExample).length > 0
        ? remainingExample
        : undefined;

    const fieldUsage = templateFieldUsage(
      additionalPropertiesField,
      1,
      fieldDef.Type,
      additionalPropertiesExample,
      additionalContext,
    );
    if (fieldUsage) {
      fields.push(fieldUsage + templateFieldDelimiter());
    }
  }

  // Strip trailing comma from last field
  if (fields.length > 0) {
    const last = fields[fields.length - 1];
    if (last.endsWith(",")) {
      fields[fields.length - 1] = last.slice(0, -1);
    }
  }

  let optionalPrefix = "";
  if (fieldDef.Optional || fieldDef.Nullable) {
    optionalPrefix = templateOptionalSymbol();
  }
  if (fields.length == 0) {
    // For class types with no fields, omit parens: Model.new instead of Model.new()
    if (fieldDef.Type.Type.toString() === "class") {
      return `${optionalPrefix}${templateType(
        fieldDef.Type,
        additionalContext,
      )}`;
    }
    return [
      `${optionalPrefix}${templateType(
        fieldDef.Type,
        additionalContext,
      )}${templateBracket(
        fieldDef.Type,
        true,
        additionalContext,
      )}${templateBracket(fieldDef.Type, false, additionalContext)}`,
    ].join("\n");
  }

  return [
    `${optionalPrefix}${templateType(
      fieldDef.Type,
      additionalContext,
    )}${templateBracket(fieldDef.Type, true, additionalContext)}`,
    fields.join("\n"),
    templateBracket(fieldDef.Type, false, additionalContext),
  ].join("\n");
}

// @ts-ignore
function templateIndent(indent: number): string {
  return "  ".repeat(indent);
}

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  if (typeDef.Type.toString() != "class") {
    return "";
  }

  let scope = "models/operations";
  if (typeDef.Scope.toString() == scope.toString()) {
    scope = "models/shared";
  }

  return `${sanitizeType(typeDef, false, false, scope)}.new`;
}

// @ts-ignore
function templateBracket(
  typeDef: TypeDef,
  opening: boolean,
  additionalContext?: TemplateValueContext,
): string {
  switch (typeDef.Type.toString()) {
    case "class":
      return opening ? "(" : ")";
    case "array":
      return opening ? "[" : "]";
    case "map":
      return opening ? "{" : "}";
    default:
      throw new Error(`Unknown type: ${typeDef.Type.toString()}`);
  }
}

// @ts-ignore
function templateStringValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  // Use double-quoted strings when value contains newlines or carriage returns
  // so that escape sequences are interpreted correctly by Ruby
  if (/[\n\r]/.test(value)) {
    return `"${value
      .replaceAll(/\\/g, "\\\\")
      .replaceAll(/"/g, '\\"')
      .replaceAll(/\r/g, "\\r")
      .replaceAll(/\n/g, "\\n")
      .replaceAll(/#\{/g, "\\#{")}"`;
  }
  return `'${value.replaceAll(/\\/g, "\\\\").replaceAll(/'/g, "\\'")}'`;
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  _additionalContext?: TemplateValueContext,
): string {
  let enumNames = getEnumNames(fieldDef.Type);

  return `${sanitizeClass(fieldDef.Type, "", false)}::${enumNames[idx]}`;
}

// @ts-ignore
function templateBoolValue(
  value: boolean,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return value ? "true" : "false";
}

// @ts-ignore
function templateByteValue(
  value: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `'${value.replace(/'/g, "\\'")}'.encode`;
}

// @ts-ignore
function templateDateValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `Date.parse('${val}')`;
}

// @ts-ignore
function templateDateTimeValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `DateTime.iso8601('${val}')`;
}
// @ts-ignore
function templateIntValue(
  val: number,
  _fieldDef: FieldDef,
  _additionalContext?: TemplateValueContext,
): string {
  return templateNumericLiteral(val);
}

function templateNumericLiteral(val: number | string): string {
  // https://docs.rubocop.org/rubocop/cops_style.html#stylenumericliterals
  // RuboCop default MinDigits is 5: numbers with 5+ digits need underscore separators.
  // Handle sign prefix and decimal portions separately to avoid incorrect underscore placement.
  const str = val.toString();
  const match = str.match(/^(-?)(\d+)(\..*)?$/);
  if (!match) {
    return str;
  }
  const [, sign, intPart, decPart] = match;
  const addSep = (digits: string) =>
    digits.replace(/\B(?=(\d{3})+(?!\d))/g, "_");

  const fmtInt = intPart.length >= 5 ? addSep(intPart) : intPart;

  let fmtDec = decPart ?? "";
  if (fmtDec.length > 0) {
    // Fractional digits (after the dot) also need separators when >= 5 digits.
    const fracDigits = fmtDec.slice(1); // strip leading "."
    fmtDec =
      fracDigits.length >= 5 ? `.${addSep(fracDigits)}` : `.${fracDigits}`;
  }

  if (fmtInt === intPart && fmtDec === (decPart ?? "")) {
    return str;
  }
  return `${sign}${fmtInt}${fmtDec}`;
}

registerTemplateFunc("templateNumericLiteral", templateNumericLiteral);

// @ts-ignore
function templateFloatLiteral(val: number | string): string {
  const parsed = Number(val);
  if (Number.isFinite(parsed) && Number.isInteger(parsed)) {
    return `${templateNumericLiteral(parsed)}.0`;
  }
  return templateNumericLiteral(val);
}

registerTemplateFunc("templateFloatLiteral", templateFloatLiteral);

// @ts-ignore
function templateFloatValue(
  val: number,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  // When targeting mock server with float32 fields, match Go's float32 JSON precision.
  // Go's encoding/json serializes float32 with the shortest decimal representation,
  // which may have fewer digits than the original spec example.
  if (
    additionalContext?.targetingMockServer &&
    fieldDef.Type?.Type?.toString() === "float32"
  ) {
    const f32 = Math.fround(val);
    if (f32 !== val) {
      for (let p = 1; p <= 9; p++) {
        const rounded = Number(f32.toPrecision(p));
        if (Math.fround(rounded) === f32) {
          val = rounded;
          break;
        }
      }
    }
  }

  if (Number.isInteger(val)) {
    return `${templateNumericLiteral(val)}.0`;
  } else {
    return `${templateNumericLiteral(val)}`;
  }
}

// @ts-ignore
function templateArrayValue(value: any): string {
  return `${value}`;
}

// @ts-ignore
function templateMapValue(key: string, val: string) {
  return `'${key}' => ${val}`;
}

// @ts-ignore
function templateOptionalSymbol(): string {
  return "";
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
      let candidate = `${name}_${getCasing(value).toUpperCase()}`;
      if (seen[candidate]) {
        let suffix = seen[candidate];
        seen[candidate] += 1;
        candidate = `${candidate}_${suffix}`;
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
    name = "unknown";
  }

  name = sanitizeName(name);
  name = caser().ToSNAKE(name);

  if (rubyReservedKeywords.includes(name)) {
    name += "_VALUE";
  }

  return name;
}

registerTemplateFunc("getEnumName", getEnumName);

function sanitizeSDKConstructorSorbet(sdk: SDK): string {
  if (!sorbetEnabled()) {
    return "";
  }
  const params = [
    "client: T.nilable(Faraday::Connection)",
    `retry_config: T.nilable(::${sanitizeModuleName(
      context.Global.Config.Module,
    )}::Utils::RetryConfig)`,
    `timeout_ms: T.nilable(Integer)`,
  ];
  const serverMap = sdk.Servers && sdk.Servers.ServerMap;
  const serverVariables = sdk.Servers && sdk.Servers.GetVariables().length > 0;

  let securityType;
  let securityName;
  if (sdk.Security) {
    securityType = sanitizeType(sdk.Security.Type, true, false, "");
    securityName = sanitizeFieldName(sdk.Security.Type.Name);

    if (
      sdk.Security.Type.Fields.length == 1 &&
      context.Global.Config.FlattenGlobalSecurity
    ) {
      securityType = sanitizeType(
        sdk.Security.Type.Fields[0].Type,
        true,
        false,
        "",
      );
      securityName = sanitizeFieldName(sdk.Security.Type.Fields[0].Name);
    }
    params.push(`${securityName}: ${securityType}`);
    params.push(
      `security_source: T.nilable(T.proc.returns(${sanitizeType(
        sdk.Security.Type,
        false,
        false,
        "",
      )}))`,
    );
  }
  if (sdk.Globals) {
    for (const field of sdk.Globals.Fields) {
      params.push(
        `${templateSDKInitFieldName(field.Name, "global")}: ${sanitizeType(
          field.Type,
          true,
          false,
        )}`,
      );
    }
  }
  if (serverVariables) {
    for (const variable of sdk.Servers.GetVariables()) {
      params.push(
        `${templateSDKInitFieldName(variable.Name, "server")}: ${sanitizeType(
          variable.Type,
          true,
          false,
        )}`,
      );
    }
  }
  if (serverMap) {
    params.push("server: T.nilable(Symbol)");
  } else {
    params.push("server_idx: T.nilable(Integer)");
  }
  params.push("server_url: T.nilable(String)");
  params.push("url_params: T.nilable(T::Hash[Symbol, String])");
  return indentString(
    `sig do
  params(
    ${params.join(",\n    ")}
  ).void
end`,
    2,
  );
}
registerTemplateFunc(
  "sanitizeSDKConstructorSorbet",
  sanitizeSDKConstructorSorbet,
);

function sanitizeSDKConstructorSignature(sdk: SDK): string {
  const params = ["client: nil", "retry_config: nil", "timeout_ms: nil"];
  const serverMap = sdk.Servers && sdk.Servers.ServerMap;
  const serverVariables = sdk.Servers && sdk.Servers.GetVariables().length > 0;

  let securityName;
  if (sdk.Security) {
    if (
      sdk.Security.Type.Fields.length == 1 &&
      context.Global.Config.FlattenGlobalSecurity
    ) {
      securityName = sanitizeFieldName(sdk.Security.Type.Fields[0].Name);
    } else {
      securityName = sanitizeFieldName(sdk.Security.Type.Name);
    }
    params.push(`${securityName}: nil`);
    params.push(`security_source: nil`);
  }
  if (sdk.Globals) {
    for (const field of sdk.Globals.Fields) {
      params.push(`${templateSDKInitFieldName(field.Name, "global")}: nil`);
    }
  }
  if (serverVariables) {
    for (const variable of sdk.Servers.GetVariables()) {
      params.push(`${templateSDKInitFieldName(variable.Name, "server")}: nil`);
    }
  }
  if (serverMap) {
    params.push("server: nil");
  } else {
    params.push("server_idx: nil");
  }
  params.push("server_url: nil");
  params.push("url_params: nil");
  return `    def initialize(${params.join(", ")})`;
}
registerTemplateFunc(
  "sanitizeSDKConstructorSignature",
  sanitizeSDKConstructorSignature,
);

function sanitizeSDKConfigConstructorSignature(sdk: SDK): string {
  const serverMap = sdk.Servers && sdk.Servers.ServerMap;
  const serverVariables = sdk.Servers && sdk.Servers.GetVariables().length > 0;

  let securityName;

  if (sdk.Security) {
    securityName = sanitizeFieldName(sdk.Security.Type.Name);

    if (
      sdk.Security.Type.Fields.length == 1 &&
      context.Global.Config.FlattenGlobalSecurity
    ) {
      securityName = sanitizeFieldName(sdk.Security.Type.Fields[0].Name);
    }
  }
  const params = ["client", "hooks", "retry_config", "timeout_ms"];
  if (sdk.Security) {
    params.push(`${securityName}`);
    params.push("security_source");
  }
  params.push("server_url");
  if (serverMap) {
    params.push("server");
  } else {
    params.push("server_idx");
  }
  if (serverVariables) {
    params.push("server_params");
  }
  if (sdk.Globals) {
    params.push("globals");
  }
  return `def initialize(${params.join(", ")})`;
}
registerTemplateFunc(
  "sanitizeSDKConfigConstructorSignature",
  sanitizeSDKConfigConstructorSignature,
);

function sanitizeSDKConfigConstructorSignatureSorbet(sdk: SDK): string {
  if (!sorbetEnabled()) {
    return "";
  }
  const serverMap = sdk.Servers && sdk.Servers.ServerMap;
  const serverVariables = sdk.Servers && sdk.Servers.GetVariables().length > 0;
  let securityType;
  let securityName;

  if (sdk.Security) {
    securityType = sanitizeType(sdk.Security.Type, true, false, "");
    securityName = sanitizeFieldName(sdk.Security.Type.Name);

    if (
      sdk.Security.Type.Fields.length == 1 &&
      context.Global.Config.FlattenGlobalSecurity
    ) {
      securityType = sanitizeType(
        sdk.Security.Type.Fields[0].Type,
        true,
        false,
        "",
      );
      securityName = sanitizeFieldName(sdk.Security.Type.Fields[0].Name);
    }
  }
  const params = [
    "client: T.nilable(Faraday::Connection)",
    `hooks: ::${sanitizeModuleName(
      context.Global.Config.Module,
    )}::SDKHooks::Hooks`,
    `retry_config: T.nilable(::${sanitizeModuleName(
      context.Global.Config.Module,
    )}::Utils::RetryConfig)`,
    `timeout_ms: T.nilable(Integer)`,
  ];
  if (sdk.Security) {
    params.push(`${securityName}: ${securityType}`);
    params.push(
      `security_source: T.nilable(T.proc.returns(${sanitizeType(
        sdk.Security.Type,
        false,
        false,
        "",
      )}))`,
    );
  }
  params.push("server_url: T.nilable(String)");
  if (serverMap) {
    params.push("server: T.nilable(Symbol)");
  } else {
    params.push("server_idx: T.nilable(Integer)");
  }
  if (serverVariables) {
    if (serverMap) {
      params.push("server_params: T::Hash[Symbol, T::Hash[Symbol, String]]");
    } else {
      params.push("server_params: T::Array[T::Hash[Symbol, String]]");
    }
  }
  if (sdk.Globals) {
    params.push(
      "globals: T.nilable(T::Hash[Symbol, T::Hash[Symbol, T::Hash[Symbol, Object]]])",
    );
  }
  return [
    `sig do`,
    `  params(`,
    `    ${params.join(",\n        ")}`,
    `  ).void`,
    `end`,
  ].join("\n    ");
}

registerTemplateFunc(
  "sanitizeSDKConfigConstructorSignatureSorbet",
  sanitizeSDKConfigConstructorSignatureSorbet,
);

// @ts-ignore
function sanitizeOperationMethodSignature(
  op: Operation,
  flattening: boolean,
): string {
  let params = [];
  for (const field of op.Arguments.Sorted) {
    params.push(
      `${sanitizeMethodArgumentName(field.Name)}:${
        field.Optional ? " nil" : ""
      }`,
    );
  }
  if (op.Extensions.Retries) {
    params.push("retries: nil");
  }
  if (op.Servers) {
    params.push("server_url: nil");
  }
  params.push("timeout_ms: nil");

  if (op.Extensions?.Pagination && hasPaginationURL(op)) {
    params.push("url_override: nil");
  }

  if (op.GetAcceptTypes().length > 1) {
    params.push("accept_header_override: nil");
  }
  params.push("http_headers: nil");

  const methodName = sanitizeMethodName(op);
  if (params.length == 0) {
    return `def ${methodName}`;
  } else {
    return `def ${methodName}(${params.join(", ")})`;
  }
}
registerTemplateFunc(
  "sanitizeOperationMethodSignature",
  sanitizeOperationMethodSignature,
);

// @ts-ignore
function sanitizeSorbetOperationMethodSignature(
  op: Operation,
  flattening: boolean,
): string {
  if (!sorbetEnabled()) {
    return "";
  }
  let params = [];

  for (const field of op.Arguments.Sorted) {
    params.push(
      `${sanitizeMethodArgumentName(field.Name)}: ${sanitizeType(
        field.Type,
        field.Optional,
        field.Nullable,
        undefined,
      )}`,
    );
  }
  if (op.Extensions.Retries) {
    params.push(`retries: T.nilable(Utils::RetryConfig)`);
  }
  if (op.Servers) {
    params.push(`server_url: T.nilable(String)`);
  }
  params.push(`timeout_ms: T.nilable(Integer)`);
  if (op.Extensions?.Pagination && hasPaginationURL(op)) {
    params.push("url_override: T.nilable(String)");
  }
  if (op.GetAcceptTypes().length > 1) {
    params.push(`accept_header_override: T.nilable(String)`);
  }
  params.push(
    `http_headers: T.nilable(T::Hash[T.any(String, Symbol), String])`,
  );

  let res = "void";
  if (op.Response.Type !== null) {
    res = `returns(${sanitizeType(op.Response.Type, false, false, "")})`;
  }
  if (params.length === 0) {
    return `sig { ${res} }`;
  } else {
    return `sig { params(${params.join(", ")}).${res} }`;
  }
}
registerTemplateFunc(
  "sanitizeSorbetOperationMethodSignature",
  sanitizeSorbetOperationMethodSignature,
);

// @ts-ignore
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName);
}

// @ts-ignore
function sanitizeHashKey(key: string): string {
  key = sanitizeName(key);
  return caser().ToSnake(key);
}
registerTemplateFunc("sanitizeHashKey", sanitizeHashKey);

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
function sanitizeJsonPath(jsonpath: string): string {
  /*
   * Support for pagination is specified by the client through a `x-speakeasy-pagination` entry in the spec file.
   * This uses a JSONPath expression to specify the path of pagination fields in the response.
   *
   * For legacy reasons, to access the last value in the `resultArray` field, our documentation encourages the syntax: $.resultArray[(@.length-1)]
   * However, this syntax is not supported by the `JsonPath` ruby library, which is used to parse JSONPath expressions.
   * Instead we run the following replace command to match the expected JsonPath syntax: `$.resultArray[-1:]`
   */
  const path = jsonpath.replaceAll(/\(\@\.length ?- ?1\)/gi, "-1:");
  return path;
}

registerTemplateFunc("sanitizeJsonPath", sanitizeJsonPath);

function templateFieldValue(fieldDef: FieldDef, val: any): string {
  switch (fieldDef.Type.Type.toString()) {
    case "enum":
      return templateEnumValue(
        fieldDef,
        fieldDef.Type.Enum.Values.indexOf(val.toString()),
      );
    case "string":
      return templateStringValue(val, fieldDef, {
        useTypesPackage: true,
      });
    case "date":
      return templateDateValue(val, fieldDef);
    case "date-time":
      return templateDateTimeValue(val, fieldDef);
    case "bigint":
    case "int32":
    case "integer":
      return templateIntValue(val, fieldDef);
    case "number":
    case "float32":
    case "decimal":
      return templateFloatValue(val, fieldDef);
    case "boolean":
      return templateBoolValue(val, fieldDef);
    case "bytes":
    default:
      throw new Error(
        `unsupported const type: ${fieldDef.Type.Type.toString()}`,
      );
  }
}

// @ts-ignore
function templateConstValue(fieldDef: FieldDef): string {
  if (fieldDef.Const.Value === null) {
    return "nil";
  }

  return templateFieldValue(fieldDef, fieldDef.Const.Value);
}

registerTemplateFunc("templateConstValue", templateConstValue);

// @ts-ignore
function templateDefaultFieldValue(
  fieldDef: FieldDef,
  usageLocation: string,
): string {
  if (fieldDef.Default.Value === null) {
    return "nil";
  }

  return templateFieldValue(fieldDef, fieldDef.Default.Value);
}

registerTemplateFunc("templateDefaultFieldValue", templateDefaultFieldValue);

function removeConstFields(fields: FieldDef[]): FieldDef[] {
  return fields.filter((field) => !field.Const);
}

registerTemplateFunc("removeConstFields", removeConstFields);

//@ts-ignore
function sanitizeCrystallineType(
  typeDef: TypeDef,
  optional?: boolean,
  nullable?: boolean,
  scope?: string,
  typesToShallowEmbed: Set<string> = new Set(),
): string {
  let type = (() => {
    switch (typeDef.Type.toString()) {
      case "enum":
      case "class":
      case "error":
        return sanitizeClass(typeDef, "", false);
      case "union":
        return sanitizeCrystallineUnionType(typeDef, typesToShallowEmbed);
      case "string":
        return "::String";
      case "date":
        return "::Date";
      case "date-time":
        return "::DateTime";
      case "integer":
      case "int32":
        return "::Integer";
      case "number":
      case "float32":
        return "::Float";
      case "boolean":
        return "Crystalline::Boolean.new";
      case "bytes":
        return "::String";
      case "map":
        return sanitizeCrystallineMap(typeDef, scope, typesToShallowEmbed);
      case "array":
        return `Crystalline::Array.new(${sanitizeCrystallineType(
          typeDef.ItemType,
          false,
          false,
          scope,
          typesToShallowEmbed,
        )})`;
      case "any":
        return "::Object";
      case "response":
        return "::Faraday::Response";
      case "request":
        return "::Faraday::Request";
      case "event-stream":
        return "::Object";
      case "jsonl":
        return "::Object";
      default:
        throw new Error(`Unknown type: "${typeDef.Type.toString()}"`);
    }
  })();
  return templateCrystallineOptional(type, optional, nullable);
}

registerTemplateFunc("sanitizeCrystallineType", sanitizeCrystallineType);

function sanitizeCrystallineUnionType(
  typeDef: TypeDef,
  typesToShallowEmbed: Set<string>,
): string {
  if (typeDef.IsUnionOpen && typeDef.Discriminator?.Mapping) {
    const discKey = typeDef.Discriminator.TypePropertyName || "type";
    const entries = typeDef.Discriminator.Mapping.map((m) => {
      const typeName = sanitizeCrystallineType(
        m.Type,
        false,
        false,
        "",
        typesToShallowEmbed,
      );
      return `'${m.Name}' => ${typeName}`;
    });
    return `Crystalline::DiscriminatedUnion.new('${discKey}', { ${entries.join(
      ", ",
    )} })`;
  }

  if (typesToShallowEmbed.has(typeDef.Name)) {
    return "Crystalline::Hash.new(Symbol, Object)";
  }
  const isCircular = detectCircularUnion(typeDef);

  const associatedTypes = typeDef.AssociatedTypes.map(
    (t: TypeDef & { ParentName?: string }) => {
      return sanitizeCrystallineType(
        t,
        false,
        false,
        "",
        isCircular
          ? typesToShallowEmbed.add(typeDef.Name)
          : typesToShallowEmbed,
      );
    },
  ).filter((value, index, array) => array.indexOf(value) === index);
  Array.from(typesToShallowEmbed.values()).length;
  if (associatedTypes.length > 1) {
    return `Crystalline::Union.new(${associatedTypes.join(", ")})`;
  } else {
    return associatedTypes[0];
  }
}

// @ts-ignore
function sanitizeCrystallineMap(
  typeDef: TypeDef,
  scope: string,
  typesToShallowEmbed: Set<string>,
): string {
  return `Crystalline::Hash.new(Symbol, ${sanitizeCrystallineType(
    typeDef.ItemType,
    false,
    false,
    scope,
    typesToShallowEmbed,
  )})`;
}

// @ts-ignore
function templateCrystallineOptional(
  type: string,
  optional: boolean,
  nullable: boolean,
): string {
  if (optional || nullable) {
    return `Crystalline::Nilable.new(${type})`;
  }

  return type;
}

/**
 * Returns YARD doc type list syntax string for the given TypeDef.
 *
 * Examples:
 *  - String
 *  - String, nil
 *  - Array<String>
 *  - Hash{Symbol => String}
 *
 * Reference: https://rubydoc.info/gems/yard/file/docs/Tags.md#type-list-conventions
 */
function sanitizeYardocType(
  typeDef: TypeDef,
  optional?: boolean,
  nullable?: boolean,
  scope?: string,
  typesToShallowEmbed: Set<string> = new Set(),
): string {
  let type = (() => {
    switch (typeDef.Type.toString()) {
      case "enum":
      case "class":
      case "error":
        return sanitizeClass(typeDef, "", false);
      case "union":
        return sanitizeYardocUnionType(typeDef, typesToShallowEmbed);
      case "string":
        return "String";
      case "date":
        return "Date";
      case "date-time":
        return "DateTime";
      case "integer":
      case "int32":
        return "Integer";
      case "number":
      case "float32":
        return "Float";
      case "boolean":
        return "Boolean";
      case "bytes":
        return "String";
      case "map":
        return sanitizeYardocMap(typeDef, scope, typesToShallowEmbed);
      case "array":
        return `Array<${sanitizeYardocType(
          typeDef.ItemType,
          false,
          false,
          scope,
          typesToShallowEmbed,
        )}>`;
      case "any":
        return "Object";
      case "response":
        return "Faraday::Response";
      case "request":
        return "Faraday::Request";
      case "event-stream":
        return `${sanitizeModuleName(
          context.Global.Config.Module,
        )}::Utils::EventStream`;
      case "jsonl":
        return `${sanitizeModuleName(
          context.Global.Config.Module,
        )}::Utils::JsonLStream`;
      default:
        throw new Error(`Unknown type: "${typeDef.Type.toString()}"`);
    }
  })();
  return templateYardocOptional(type, optional, nullable);
}

registerTemplateFunc("sanitizeYardocType", sanitizeYardocType);

/**
 * Returns YARD doc type list syntax string for the given union TypeDef.
 */
function sanitizeYardocUnionType(
  typeDef: TypeDef,
  typesToShallowEmbed: Set<string>,
): string {
  if (typesToShallowEmbed.has(typeDef.Name)) {
    return "Hash{Symbol => Object}";
  }
  const isCircular = detectCircularUnion(typeDef);

  const associatedTypes = typeDef.AssociatedTypes.map(
    (t: TypeDef & { ParentName?: string }) => {
      return sanitizeYardocType(
        t,
        false,
        false,
        "",
        isCircular
          ? typesToShallowEmbed.add(typeDef.Name)
          : typesToShallowEmbed,
      );
    },
  ).filter((value, index, array) => array.indexOf(value) === index);
  Array.from(typesToShallowEmbed.values()).length;

  return associatedTypes.join(", ");
}

/**
 * Returns YARD doc type list syntax string for the given map TypeDef.
 */
function sanitizeYardocMap(
  typeDef: TypeDef,
  scope: string,
  typesToShallowEmbed: Set<string>,
): string {
  return `Hash{Symbol => ${sanitizeYardocType(
    typeDef.ItemType,
    false,
    false,
    scope,
    typesToShallowEmbed,
  )}}`;
}

/**
 * Suffixes ", nil" to the type if it is optional or nullable.
 */
function templateYardocOptional(
  type: string,
  optional: boolean,
  nullable: boolean,
): string {
  if (optional || nullable) {
    return `${type}, nil`;
  }

  return type;
}

// @ts-ignore
function languageSpecificEnvVarWrapping(
  envVar: EnvVar,
  additionalContext?: TemplateValueContext,
  fieldDef?: FieldDef,
): string {
  return `ENV.fetch('${envVar.name}', '${envVar.defaultValue}')`;
}
