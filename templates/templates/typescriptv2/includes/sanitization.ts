// Returns `name` with the configured private-identifier prefix prepended (default "#", e.g. "#hooks").
// @ts-ignore
function prefixPrivateId(name: string): string {
  return `${context.Global.Config.PrivateIdentifierPrefix}${name}`;
}

registerTemplateFunc("prefixPrivateId", prefixPrivateId);

// @ts-ignore
function sanitizeSDKAccess(sdk: SDK): string {
  const group = sdk.Group;

  if (!group) {
    return "";
  }

  const sanitize =
    context.Global.Config.ModelPropertyCasing === "snake"
      ? (s: string) => caser().ToSnake(sanitizeName(s))
      : sanitizeCamelCase;

  let formattedName = "";
  if (group.includes(".") && ".".repeat(group.length) !== group) {
    const pieces = group.split(".");
    formattedName = sanitize(pieces[0]);
    for (const i of pieces.slice(1)) {
      formattedName += "." + sanitize(i);
    }
  } else {
    formattedName = sanitize(group);
  }

  if (formattedName != "") {
    formattedName = `.${formattedName}`;
  }

  return formattedName;
}

registerTemplateFunc("sanitizeSDKAccess", sanitizeSDKAccess);

// Based on gen.yaml `modelPropertyCasing` will pick `camel` or `snake` use `sanitizeCamelCase()` if you want to avoid this
// @ts-ignore
function sanitizeFieldName(name: string): string {
  let result = sanitizeName(name);

  if (context.Global.Config.ModelPropertyCasing === "snake") {
    result = caser().ToSnake(result);
  } else {
    result = caser().ToCamel(result);
  }

  if (typescriptReservedTypeNames.has(result)) {
    result += "T";
  }

  return result;
}

registerTemplateFunc("sanitizeFieldName", sanitizeFieldName);

// Field names that collide with ClientSDK base class properties when prefixed
// with `_`. Sub-SDK lazy-init fields use `private _<fieldName>` pattern, which
// would shadow the inherited `_options` or `_baseURL` fields.
const sdkBaseClassFields = new Set(["options", "baseURL"]);

// @ts-ignore
function sanitizeSubSDKFieldName(name: string): string {
  let result = sanitizeFieldName(name);
  if (sdkBaseClassFields.has(result)) {
    result += "$";
  }
  return result;
}
registerTemplateFunc("sanitizeSubSDKFieldName", sanitizeSubSDKFieldName);

// Helper function to check if we should preserve model field casing
function shouldPreserveCasing(): boolean {
  return context.Global.Config.PreserveModelFieldNames === true;
}

// New functions to be used for objects/error model fields
// Prefer calling `sanitizeModelField` when fieldDef is available
// @ts-ignore
function sanitizeModelField(field: FieldDef): string {
  if (field.Annotations?.Has("needsCasing")) {
    return sanitizeFieldName(field.Name);
  }
  return sanitizeModelFieldName(field.Name);
}
registerTemplateFunc("sanitizeModelField", sanitizeModelField);

// When no field def available
// @ts-ignore
function sanitizeModelFieldName(name: string): string {
  if (shouldPreserveCasing()) {
    return name;
  }
  return sanitizeFieldName(name);
}
registerTemplateFunc("sanitizeModelFieldName", sanitizeModelFieldName);

// Should be called consistently whenever constructing a model field key name
// of an object which may or may not be preserved.
// @ts-ignore
function declareModelField(field: FieldDef): string {
  const name = sanitizeModelField(field);
  return typescriptIdentifierRegex.test(name) ? name : `"${name}"`;
}
registerTemplateFunc("declareModelField", declareModelField);

// Returns the appropriate key name for a security field in an $Outbound type
// or Zod schema. Uses originalFieldName by default, but falls back to
// sanitizeModelField when the original name would collide with another
// security field in the same type (e.g. two schemes both named "bearerAuth").
// @ts-ignore
function outboundSecurityKeyName(
  field: FieldDef,
  allFields: FieldDef[],
): string {
  const origName = originalFieldName(field);

  if (!isSecurityField(field)) {
    return origName;
  }

  const hasDuplicate = allFields.some(
    (f) =>
      f !== field && isSecurityField(f) && originalFieldName(f) === origName,
  );

  if (hasDuplicate) {
    return sanitizeModelField(field);
  }

  return origName;
}
registerTemplateFunc("outboundSecurityKeyName", outboundSecurityKeyName);

// @ts-ignore
function accessModelField(field: FieldDef, includeDot: boolean = true): string {
  const name = sanitizeModelField(field);
  if (typescriptIdentifierRegex.test(name)) {
    return includeDot ? "." + name : name;
  }
  return `["${name}"]`;
}
registerTemplateFunc("accessModelField", accessModelField);

// @ts-ignore
function sanitizeCamelCase(name: string): string {
  let result = sanitizeName(name);
  result = caser().ToCamel(result);

  if (typescriptReservedTypeNames.has(result)) {
    result += "T";
  }

  return result;
}

registerTemplateFunc("sanitizeCamelCase", sanitizeCamelCase);

// @ts-ignore
function sanitizeVariableName(name: string, fallback?: string) {
  let result = sanitizeCamelCase(name);
  if (typescriptReservedVariableKeywords.has(result)) {
    if (fallback) return fallback;
    return result + "$";
  }
  return result;
}

registerTemplateFunc("sanitizeVariableName", sanitizeVariableName);

/**
 * Resolves method name collisions within SDK classes.
 * Similar to resolveConflictingFuncNames but scoped to SDK classes.
 * Methods in the same SDK class that would have the same name get numeric suffixes.
 */
function resolveConflictingMethodNames() {
  if (sanitizedMethodNames) {
    return sanitizedMethodNames;
  }

  // Build initial map of operations by SDK and their generated method names
  const opsBySDKAndMethodName = new Map<string, Map<string, Operation[]>>();

  for (const op of iterateOperations()) {
    const sdkID = op.OwningSDK.Type.Name;
    const methodName = sanitizeMethodNameInner(op);

    let sdkMap = opsBySDKAndMethodName.get(sdkID);
    if (!sdkMap) {
      sdkMap = new Map();
      opsBySDKAndMethodName.set(sdkID, sdkMap);
    }

    const ops = sdkMap.get(methodName) || [];
    if (!ops.find((x) => x.ID === op.ID)) {
      ops.push(op);
    }
    sdkMap.set(methodName, ops);
  }

  const result = new Map<string, string>();

  // Process each SDK's methods
  for (const [, sdkMethods] of opsBySDKAndMethodName) {
    // Group by case-insensitive name to detect both exact duplicates and case collisions
    const groupedByCaseInsensitive = new Map<
      string,
      Map<string, Operation[]>
    >();
    for (const [methodName, ops] of sdkMethods) {
      const lowerName = methodName.toLowerCase();
      let group = groupedByCaseInsensitive.get(lowerName);
      if (!group) {
        group = new Map();
        groupedByCaseInsensitive.set(lowerName, group);
      }
      group.set(methodName, ops);
    }

    // Process each case-insensitive group
    for (const [, variants] of groupedByCaseInsensitive) {
      // Sort variant names for deterministic ordering
      const sortedVariants = Array.from(variants.entries()).sort((a, b) =>
        a[0].localeCompare(b[0]),
      );

      // Track all assigned names to avoid collisions
      const usedNames = new Set<string>();

      for (const [methodName, ops] of sortedVariants) {
        // Handle exact duplicates within this variant
        for (let i = 0; i < ops.length; i++) {
          let finalName = methodName;

          // If this is a duplicate (i > 0) or the base name is already used, add suffix
          if (i > 0 || usedNames.has(finalName.toLowerCase())) {
            let suffix = i > 0 ? i + 1 : 1;
            do {
              finalName = methodName + String(suffix);
              suffix++;
            } while (usedNames.has(finalName.toLowerCase()));
          }

          usedNames.add(finalName.toLowerCase());
          result.set(ops[i].ID, finalName);
        }
      }
    }
  }

  sanitizedMethodNames = result;
  return result;
}

let sanitizedMethodNames: null | Map<string, string> = null;

function sanitizeMethodNameInner(operation: Operation): string {
  let id = operation.GetID();
  return caser().ToCamel(sanitizeName(id));
}

// @ts-ignore
function sanitizeMethodName(operation: Operation): string {
  const names = resolveConflictingMethodNames();
  return names.get(operation.ID) ?? sanitizeMethodNameInner(operation);
}
registerTemplateFunc("sanitizeMethodName", sanitizeMethodName);

// @ts-ignore
function sanitizePrivateMethodName(name: string): string {
  return caser().ToCamel(sanitizeName(name));
}
registerTemplateFunc("sanitizePrivateMethodName", sanitizePrivateMethodName);

const WEBHOOK_HANDLER_NAME = "validateWebhook";

function sanitizeWebhookHandlerMethodName(): string {
  // TODO: Build this dynamically from `Webhook` config
  return sanitizeCamelCase(WEBHOOK_HANDLER_NAME);
}
registerTemplateFunc(
  "sanitizeWebhookHandlerMethodName",
  sanitizeWebhookHandlerMethodName,
);

// @ts-ignore
function sanitizeClassName(name: string) {
  name = sanitizeName(name);
  name = caser().ToPascal(name);

  if (typescriptReservedTypeNames.has(name)) {
    name += "T";
  }

  return caser().ToPascal(name);
}

registerTemplateFunc("sanitizeClassName", sanitizeClassName);

function sanitizeCoreSDKName(sdk?: SDK) {
  const s = sdk ?? context.Global.AST.MainSDK;
  const name = sanitizeName(s.Type.Name);
  return caser().ToPascal(name + "Core");
}

registerTemplateFunc("sanitizeCoreSDKName", sanitizeCoreSDKName);

function sanitizeReactContextName(
  variant: "context" | "provider" | "hook",
  sdk?: SDK,
) {
  const s = sdk ?? context.Global.AST.MainSDK;
  const name = sanitizeName(s.Type.Name);

  switch (variant) {
    case "context":
      return caser().ToPascal(name + "Context");
    case "provider":
      return caser().ToPascal(name + "Provider");
    case "hook":
      return `use${caser().ToPascal(name + "Context")}`;
  }
}

registerTemplateFunc("sanitizeReactContextName", sanitizeReactContextName);

/**
 * In very rare cases you can get conflicting function names.
 *
 * For example:
 *
 * Operation 1
 * "x-speakeasy-name-override": "list",
 * "x-speakeasy-group": "reports-executions"
 * Operation 2
 * "x-speakeasy-name-override": "list",
 * "x-speakeasy-group": "reports.executions",
 *
 * This will result in the following function names:
 *
 * "reports-executions_list"
 * "reports-executions_list1"
 */
function resolveConflictingFuncNames() {
  if (sanitizedFuncNames) {
    return sanitizedFuncNames;
  }

  // Build initial map of operations by their generated function names
  const opsByFuncName = new Map<string, Operation[]>();
  for (const op of iterateOperations()) {
    const funcName = sanitizeFuncNameInner(op);
    const g = opsByFuncName.get(funcName) || [];
    if (!g.find((x) => x.ID === op.ID)) {
      g.push(op);
    }
    opsByFuncName.set(funcName, g);
  }

  // Group by case-insensitive name to detect both exact duplicates and case collisions
  // This handles case-insensitive filesystems (macOS, Windows) and TypeScript imports
  const groupedByCaseInsensitive = new Map<string, Map<string, Operation[]>>();
  for (const [funcName, ops] of opsByFuncName) {
    const lowerName = funcName.toLowerCase();
    let group = groupedByCaseInsensitive.get(lowerName);
    if (!group) {
      group = new Map();
      groupedByCaseInsensitive.set(lowerName, group);
    }
    group.set(funcName, ops);
  }

  const result = new Map<string, string>();

  // Process each case-insensitive group
  for (const [, variants] of groupedByCaseInsensitive) {
    // Sort variant names for deterministic ordering
    const sortedVariants = Array.from(variants.entries()).sort((a, b) =>
      a[0].localeCompare(b[0]),
    );

    // Track all assigned names to avoid collisions
    const usedNames = new Set<string>();

    for (const [funcName, ops] of sortedVariants) {
      // Handle exact duplicates within this variant
      for (let i = 0; i < ops.length; i++) {
        let finalName = funcName;

        // If this is a duplicate (i > 0) or the base name is already used, add suffix
        if (i > 0 || usedNames.has(finalName.toLowerCase())) {
          let suffix = i > 0 ? i + 1 : 1;
          do {
            finalName = funcName + String(suffix);
            suffix++;
          } while (usedNames.has(finalName.toLowerCase()));
        }

        usedNames.add(finalName.toLowerCase());
        result.set(ops[i].ID, finalName);
      }
    }
  }

  sanitizedFuncNames = result;

  return result;
}

let sanitizedFuncNames: null | Map<string, string> = null;

function sanitizeFuncName(operation: Operation): string {
  const names = resolveConflictingFuncNames();
  return names.get(operation.ID) ?? sanitizeFuncNameInner(operation);
}

function sanitizeFuncNameInner(operation: Operation): string {
  const group = operation.OwningSDK.Group;
  let prefix = "";
  // We only qualify sub-sdk functions, not the root SDKs functions.
  if (group) {
    prefix = group.replaceAll(".", "_") + "_";
  }

  let result = sanitizeCamelCase(`${prefix}${operation.GetID()}`);

  // Check for reserved variable keywords that would be invalid as function names
  // Also check for names that collide with the SDK's built-in method parameter
  // names (e.g., "options" is used for RequestOptions in every method signature).
  if (typescriptReservedVariableKeywords.has(result) || result === "options") {
    result += "$";
  }

  return result;
}
registerTemplateFunc("sanitizeFuncName", sanitizeFuncName);

function sanitizeWebhookHandlerFuncName(): string {
  // TODO: Build this dynamically from `Webhook` config
  return sanitizeCamelCase(WEBHOOK_HANDLER_NAME);
}
registerTemplateFunc(
  "sanitizeWebhookHandlerFuncName",
  sanitizeWebhookHandlerFuncName,
);

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

function truncateFileName(name: string): string {
  // Leave room for longest tsc output suffix: .core.d.ts.map (14 chars)
  return truncateName(name, 240);
}

function sanitizeFuncFilename(operation: Operation): string {
  const funcName = sanitizeFuncName(operation);
  let name: string;
  if (!context.Global?.Config?.LegacyFileNaming) {
    name = caser().ToKebab(funcName).toLowerCase();
  } else {
    name = funcName;
  }
  return truncateFileName(name);
}
registerTemplateFunc("sanitizeFuncFilename", sanitizeFuncFilename);

function sanitizeReactQueryHookName(operation: Operation): string {
  const name = operation.Extensions.ReactHook?.Name;
  if (name && name.trim() !== "") {
    return sanitizeFieldName(name);
  }

  return sanitizeFuncName(operation);
}
registerTemplateFunc("sanitizeReactQueryHookName", sanitizeReactQueryHookName);

function sanitizeReactQueryHookFilename(operation: Operation): string {
  const hookName = sanitizeReactQueryHookName(operation);
  let name: string;
  if (!context.Global?.Config?.LegacyFileNaming) {
    name = caser().ToKebab(hookName).toLowerCase();
  } else {
    name = hookName;
  }
  return truncateFileName(name);
}
registerTemplateFunc(
  "sanitizeReactQueryHookFilename",
  sanitizeReactQueryHookFilename,
);

function sanitizeWebhookHandlerFuncFilename(): string {
  const funcName = sanitizeWebhookHandlerFuncName();
  if (!context.Global?.Config?.LegacyFileNaming) {
    return caser().ToKebab(funcName).toLowerCase();
  }
  return funcName;
}
registerTemplateFunc(
  "sanitizeWebhookHandlerFuncFilename",
  sanitizeWebhookHandlerFuncFilename,
);

// @ts-ignore
function sanitizeFileName(name: string) {
  if (!context.Global?.Config?.LegacyFileNaming) {
    // In kebab mode: use sanitizeName to preserve word boundaries (underscores),
    // then convert to kebab-case, then remove any remaining illegal chars
    name = sanitizeName(name);
    name = caser().ToKebab(name).toLowerCase();
    // Remove any remaining illegal filename characters (but hyphens are ok for kebab)
    name = name.replace(/[^a-z0-9\-]/g, "");
  } else {
    // Legacy mode: use sanitizeFile which removes underscores
    name = sanitizeFile(name, "");
    name = name.toLowerCase();
  }

  name = truncateName(name, 240);

  return name;
}

registerTemplateFunc("sanitizeFileName", sanitizeFileName);

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

function templateConstName(name: string, prefix: string) {
  name = sanitizeName(name);
  return caser().ToPascal(`${prefix}_${name}`);
}

registerTemplateFunc("templateConstName", templateConstName);

// @ts-ignore
function sanitizeComments(comment: string): string {
  return comment
    .replaceAll("*/", "* /")
    .replaceAll("\r\n", "\n")
    .replaceAll("\uFEFF", "")
    .replaceAll(/\u00A0/g, " ")
    .replaceAll(/\u200b/g, " ")
    .replaceAll("{{", `{{"{{"}}`);
}

// @ts-ignore
function sanitizeClass(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
): string {
  let className = sanitizeClassName(typeDef.Name);

  let prefix = "";
  if (!definition) {
    const modelLocation = getModelsLocation(typeDef.OutputLocation);
    if (modelLocation !== usageLocation && usageLocation != "usage") {
      const parts = modelLocation.split("/");
      if (hasNamespaceCollisions() && isModelSubdirectory(modelLocation)) {
        const rootAlias = getParentBarrelAlias();
        prefix = `${rootAlias}.${parts.slice(1).join(".")}.`;
      } else {
        prefix = `${parts[parts.length - 1]}.`;
      }
    }
  }

  return `${prefix}${className}`;
}

registerTemplateFunc("sanitizeClass", sanitizeClass);

function sanitizeEnumTSType(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
): string {
  let name = sanitizeClass(typeDef, usageLocation, definition);
  if (getEnumFormat(typeDef) === "enum" && typeDef.Enum?.Open) {
    name += "Open";
  }

  return name;
}

registerTemplateFunc("sanitizeEnumTSType", sanitizeEnumTSType);

const codeRE = /^\d\d\d$/;
function inlineCode(code: string) {
  return codeRE.test(code) ? code : `"${code}"`;
}

// @ts-ignore
function templateStatusCode(statusCodes: string[]) {
  if (statusCodes.length === 0) {
    return "[]";
  }
  if (statusCodes.length === 1) {
    return inlineCode(statusCodes[0] || "");
  }

  let out = "";
  for (const code of statusCodes) {
    out += `${inlineCode(code)}, `;
  }

  return `[${out.slice(0, -2)}]`;
}

registerTemplateFunc("templateStatusCode", templateStatusCode);

function templateErrorStatusCodes(response: ResponseDef): string {
  return `[${response
    .GetErrorStatusCodes()
    .map((s) => `"${s}"`)
    .join(",")}]`;
}
registerTemplateFunc("templateErrorStatusCodes", templateErrorStatusCodes);

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
    case "string":
      return "string";
    case "date":
      return isLaxMode() ? "Date" : "RFCDate";
    case "date-time":
      return "Date";
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
      return eventStreamTypeRef(eventType);
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

function sanitizeLastItemType(typeDef: TypeDef, usageLocation: string): string {
  if (
    typeDef.Type.toString() === "array" ||
    typeDef.Type.toString() === "map"
  ) {
    return sanitizeLastItemType(typeDef.ItemType, usageLocation);
  }

  let prefix = "";
  const modelLocation = getModelsLocation(typeDef.OutputLocation);
  if (modelLocation !== usageLocation) {
    const parts = modelLocation.split("/");
    if (hasNamespaceCollisions() && isModelSubdirectory(modelLocation)) {
      const rootAlias = getParentBarrelAlias();
      prefix = `${rootAlias}.${parts.slice(1).join(".")}.`;
    } else {
      prefix = `${parts[parts.length - 1]}.`;
    }
  }

  return `${prefix}${sanitizeClassName(typeDef.Name)}`;
}

registerTemplateFunc("sanitizeLastItemType", sanitizeLastItemType);

function getLastItemType(typeDef: TypeDef): string {
  if (
    typeDef.Type.toString() === "array" ||
    typeDef.Type.toString() === "map"
  ) {
    return getLastItemType(typeDef.ItemType);
  }

  return typeDef.Type.toString();
}

registerTemplateFunc("getLastItemType", getLastItemType);

function sanitizeTypeFilename(name: string): string {
  name = sanitizeFileName(name);
  if (name === "index") {
    return "indext";
  }

  return name;
}
registerTemplateFunc("sanitizeTypeFilename", sanitizeTypeFilename);

function sanitizeAcceptEnumName(operation: Operation): string {
  return `${caser().ToPascal(sanitizeName(operation.GetID()))}AcceptEnum`;
}

registerTemplateFunc("sanitizeAcceptEnumName", sanitizeAcceptEnumName);

// @ts-ignore
function sanitizeAcceptEnumKey(acceptType: string): string {
  return `${caser().ToCamel(sanitizeName(acceptType.split(";")[0]))}`;
}

registerTemplateFunc("sanitizeAcceptEnumKey", sanitizeAcceptEnumKey);

// @ts-ignore
function sanitizeAcceptEnumValue(acceptType: string): string {
  return `${acceptType.split(";")[0]}`;
}

registerTemplateFunc("sanitizeAcceptEnumValue", sanitizeAcceptEnumValue);

// @ts-ignore
function sanitizeUnion(typeDef: TypeDef, scope: string) {
  if (typeDef.AssociatedTypes.length == 0) {
    return "any";
  }

  let types = typeDef.AssociatedTypes.map((type) =>
    sanitizeType(type, false, scope),
  );

  return types.join(" | ");
}

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

    if (
      additionalPropertiesExample &&
      shouldFlattenAdditionalProperties(additionalPropertiesField)
    ) {
      // When using flat additional properties, spread them directly into the object
      for (const key in additionalPropertiesExample) {
        const itemFieldDef = typeDefToFieldDef(
          additionalPropertiesField.Type.ItemType,
          additionalPropertiesField,
          undefined,
          additionalPropertiesField.Type.ContainsNull,
        );
        const value = templateValue(
          itemFieldDef,
          additionalPropertiesExample[key],
          true,
          additionalContext,
        );
        if (value.trim() !== "") {
          fields.push(
            indentLines([`"${key}": ${value}`], 1) + templateFieldDelimiter(),
          );
        }
      }
    } else {
      // Traditional approach: additional properties as a separate field
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
  }

  let optionalPrefix = "";
  if (fieldDef.Optional || fieldDef.Nullable) {
    optionalPrefix = templateOptionalSymbol();
  }
  if (fields.length == 0) {
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
function templateFieldDeclaration(
  fieldDef: FieldDef,
  parent?: TypeDef,
  additionalContext?: TemplateValueContext,
) {
  return `${declareModelField(fieldDef)}: `;
}

// @ts-ignore
function templateNullValue(fieldDef: FieldDef): string {
  return "null";
}

// @ts-ignore
function templateUnsetValue(): string {
  return "undefined";
}

// @ts-ignore
function templateFieldDelimiter() {
  return ",";
}

// @ts-ignore
function templateArrayValue(value) {
  return value;
}

// @ts-ignore
function templateMapValue(key, val) {
  return `"${key}": ${val}`;
}

// @ts-ignore
function templateOptionalSymbol() {
  return "";
}

// @ts-ignore
function templateBracket(
  typeDef,
  opening,
  additionalContext?: TemplateValueContext,
) {
  if (typeDef.Type.toString() == "array") {
    return opening ? "[" : "]";
  }

  return opening ? "{" : "}";
}

// @ts-ignore
function templateIndent(indent) {
  return "  ".repeat(indent);
}

// @ts-ignore
function templateStringValue(
  val: string,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (val.includes(".env[")) {
    return val;
  }

  return quote(val).replaceAll("{{", `{{"{{"}}`);
}

// @ts-ignore
function templateEnumValue(
  fieldDef: FieldDef,
  idx: number,
  _additionalContext?: TemplateValueContext,
) {
  if (getEnumFormat(fieldDef.Type) === "union") {
    const val = `${fieldDef.Type.Enum.Values[idx]}`;

    return getEnumDataType(fieldDef.Type) === "number" ? val : `"${val}"`;
  }

  let names = getEnumNames(fieldDef.Type);
  addUsageImportForType(fieldDef.Type);
  return `${sanitizeClassName(fieldDef.Type.Name)}.${names[idx]}`;
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
  val: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  return `new TextEncoder().encode("${val.replace(/"/g, '\\"')}")`;
}

// @ts-ignore
function templateDateValue(
  val: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  // No-zod: date fields are plain ISO strings on the wire.
  if (isNoZod()) {
    return `"${val}"`;
  }
  if (isLaxMode()) {
    return `new Date("${val}")`;
  }
  addUsageTypeImport("rfcdate", "RFCDate");
  return `new RFCDate("${val}")`;
}

// @ts-ignore
function templateDateTimeValue(
  val: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  // No-zod: date-time fields are plain ISO datetime strings.
  if (isNoZod()) {
    return `"${val}"`;
  }
  return `new Date("${val}")`;
}

// @ts-ignore
function templateIntValue(
  val: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  if (fieldDef.Type.Type.valueOf() === "bigint" && !isNoZod()) {
    if (!isLaxMode() || fieldDef.Type.Format === "string") {
      return `${val}n`;
    }
  }
  // No-zod: integer-string (format=string) fields ride the wire as strings.
  if (isNoZod() && fieldDef.Type.Format === "string") {
    return `"${val}"`;
  }
  return `${val}`;
}

// @ts-ignore
function templateFloatValue(
  val: any,
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): string {
  switch (fieldDef.Type.Type.toString()) {
    case "decimal":
      // No-zod: decimal fields are plain numbers / strings on the wire; no
      // Decimal class is constructed at runtime, so just emit the literal.
      // For format=string preserve precision by emitting a quoted string.
      if (isNoZod()) {
        if (fieldDef.Type.Format === "string") {
          return `"${val}"`;
        }
        return `${val}`;
      }
      addUsageTypeImport("decimal", "Decimal");
      return `new Decimal("${val}")`;
    default:
      // No-zod: number-string (format=string) fields ride the wire as strings.
      if (isNoZod() && fieldDef.Type.Format === "string") {
        return `"${val}"`;
      }
      return `${val}`;
  }
}

// @ts-ignore
function templateType(
  typeDef: TypeDef,
  additionalContext?: TemplateValueContext,
): string {
  return "";
}

// @ts-ignore
function sanitizeMap(typeDef: TypeDef, usageLocation: string): string {
  return `{ [k: string]: ${sanitizeType(
    typeDef.ItemType,
    false,
    usageLocation,
  )} }`;
}

// @ts-ignore
function getEnumNamesFromValues(values) {
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
      name = `${name}${caser().ToPascal(getCasing(value))}`;
    }

    enumNames.push(name);
  }

  return enumNames;
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

function sanitizeZodName(name: string) {
  return sanitizeClassName(name) + "$";
}
registerTemplateFunc("sanitizeZodName", sanitizeZodName);

function sanitizeZod(
  typeDef: TypeDef,
  usageLocation: string,
  definition: boolean,
) {
  return sanitizeClass(typeDef, usageLocation, definition) + "$";
}

function sanitizeKey(input: string): string {
  return typescriptIdentifierRegex.test(input) ? input : `"${input}"`;
}
registerTemplateFunc("sanitizeKey", sanitizeKey);

function destructureKey(input: string): string {
  const key = sanitizeKey(input);

  return key === input ? key : `${key}: ${sanitizeFieldName(input)}`;
}
registerTemplateFunc("destructureKey", destructureKey);

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
function sanitizeSecurityFieldName(fieldName: string): string {
  return sanitizeFieldName(fieldName);
}

function sanitizeClassRef(typeDef: TypeDef, usageLocation: string): string {
  const disableQualifier =
    typeDef.Truncated || !context.Global.Config.UseIndexModules;
  return sanitizeClass(typeDef, usageLocation, disableQualifier);
}

function sanitizeTSEnumRef(typeDef: TypeDef, usageLocation: string): string {
  const disableQualifier =
    typeDef.Truncated || !context.Global.Config.UseIndexModules;
  return sanitizeEnumTSType(typeDef, usageLocation, disableQualifier);
}

function sanitizeZodRef(typeDef: TypeDef, usageLocation: string): string {
  const disableQualifier =
    typeDef.Truncated || !context.Global.Config.UseIndexModules;
  return sanitizeZod(typeDef, usageLocation, disableQualifier);
}

function sanitizeFromJSONFuncName(typeDef: TypeDef): string {
  return sanitizeCamelCase(typeDef.Name) + "FromJSON";
}
registerTemplateFunc("sanitizeFromJSONFuncName", sanitizeFromJSONFuncName);

function sanitizeToJSONFuncName(typeDef: TypeDef): string {
  return sanitizeCamelCase(typeDef.Name) + "ToJSON";
}
registerTemplateFunc("sanitizeToJSONFuncName", sanitizeToJSONFuncName);

//@ts-ignore
function shouldTemplateConstValue(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): boolean {
  return true;
}

const FUNC_LOCALS = new Set([
  "baseURL",
  "body",
  "client",
  "headers",
  "input",
  "page",
  "params",
  "parsed",
  "path",
  "pathParams",
  "payload",
  "query",
  "raw",
  "req",
  "responseFields",
  "result",
  "recipient",
  "value",
]);

let _funcLocalCollisions: Set<string> = new Set();

function initFuncLocals(state: OperationStateTS): string {
  const externalNames = new Set([
    state.funcName,
    ...state.operation.Arguments.Sorted.map((a) => sanitizeFieldName(a.Name)),
  ]);

  _funcLocalCollisions = new Set<string>();
  for (const local of FUNC_LOCALS) {
    if (externalNames.has(local)) {
      _funcLocalCollisions.add(local);
    }
  }

  return "";
}
registerTemplateFunc("initFuncLocals", initFuncLocals);

function funcLocal(name: string): string {
  if (_funcLocalCollisions.has(name)) {
    return name + "$";
  }
  return name;
}
registerTemplateFunc("funcLocal", funcLocal);

// @ts-ignore
function languageSpecificEnvVarWrapping(
  envVar: EnvVar,
  additionalContext?: TemplateValueContext,
): string {
  return `process.env["${envVar.name}"] ?? "${envVar.defaultValue}"`;
}

function sanitizeReactQueryName(
  operation: Operation,
  variant:
    | "mutation-factory"
    | "mutation-key"
    | "mutation-suspense"
    | "mutation"
    | "query-factory"
    | "query-infinite-factory"
    | "query-infinite-suspense"
    | "query-infinite"
    | "query-invalidate-all"
    | "query-invalidate"
    | "query-key"
    | "query-key-infinite"
    | "query-prefetch"
    | "query-prefetch-infinite"
    | "query-set-data"
    | "query-suspense"
    | "query",
): string {
  const hookName = sanitizeReactQueryHookName(operation);

  switch (variant) {
    case "mutation-factory": {
      return sanitizeCamelCase(`build_${hookName}_mutation`);
    }
    case "mutation-key": {
      return sanitizeCamelCase(`mutation_key_${hookName}`);
    }
    case "mutation-suspense": {
      return sanitizeCamelCase(`use_${hookName}_mutation_suspense`);
    }
    case "mutation": {
      return sanitizeCamelCase(`use_${hookName}_mutation`);
    }
    case "query-factory": {
      return sanitizeCamelCase(`build_${hookName}_query`);
    }
    case "query-infinite-factory": {
      return sanitizeCamelCase(`build_${hookName}_infinite_query`);
    }
    case "query-infinite-suspense": {
      return sanitizeCamelCase(`use_${hookName}_infinite_suspense`);
    }
    case "query-infinite": {
      return sanitizeCamelCase(`use_${hookName}_infinite`);
    }
    case "query-invalidate-all": {
      return sanitizeCamelCase(`invalidate_all_${hookName}`);
    }
    case "query-invalidate": {
      return sanitizeCamelCase(`invalidate_${hookName}`);
    }
    case "query-key": {
      return sanitizeCamelCase(`query_key_${hookName}`);
    }
    case "query-key-infinite": {
      return sanitizeCamelCase(`query_key_${hookName}_infinite`);
    }
    case "query-prefetch": {
      return sanitizeCamelCase(`prefetch_${hookName}`);
    }
    case "query-prefetch-infinite": {
      return sanitizeCamelCase(`prefetch_${hookName}_infinite`);
    }
    case "query-set-data": {
      return sanitizeCamelCase(`set_${hookName}_data`);
    }
    case "query-suspense": {
      return sanitizeCamelCase(`use_${hookName}_suspense`);
    }
    case "query": {
      return sanitizeCamelCase(`use_${hookName}`);
    }
    default: {
      variant satisfies never;
      throw new Error(`Unknown react query name variant: ${variant}`);
    }
  }
}
registerTemplateFunc("sanitizeReactQueryName", sanitizeReactQueryName);

function sanitizeJsonPath(jsonpath: string): string {
  const path = jsonpath
    // Note (alexa): couldn't find a library that correctly parses this expression
    // This is not a long-term solution
    .replaceAll(/\(\@\.length ?- ?1\)/gi, "-1");
  return path;
}

function buildConstants() {
  return {
    defaultUnknownDiscriminatorValue: "UNKNOWN",
    httpClient: sanitizeFieldName("httpClient"),
    serverIdx: sanitizeFieldName("serverIdx"),
    serverURL: sanitizeFieldName("serverURL"),
    security: "security",
    defaultClient: sanitizeFieldName("defaultClient"),
    server: "server",
    userAgent: sanitizeFieldName("userAgent"),
    timeoutMs: sanitizeFieldName("timeoutMs"),
    debugLogger: sanitizeFieldName("debugLogger"),
    retryCodes: sanitizeFieldName("retryCodes"),
    fetchOptions: sanitizeFieldName("fetchOptions"),
    extraQuery: sanitizeFieldName("extraQuery"),
    extraBody: sanitizeFieldName("extraBody"),
    retries: "retries",
    retryConfig: sanitizeFieldName("retryConfig"),
    credentials: {
      username: "username",
      password: "password",
      clientID: sanitizeFieldName("clientID"),
      clientSecret: sanitizeFieldName("clientSecret"),
      tokenURL: sanitizeFieldName("tokenURL"),
      scopes: sanitizeFieldName("scopes"),
      appId: sanitizeFieldName("appId"),
    },
    hookContext: {
      baseURL: sanitizeFieldName("baseURL"),
      operationID: sanitizeFieldName("operationID"),
      oAuth2Scopes: sanitizeFieldName("oAuth2Scopes"),
      webhookRecipient: sanitizeFieldName("webhookRecipient"),
      securitySource: sanitizeFieldName("securitySource"),
      resolvedSecurity: sanitizeFieldName("resolvedSecurity"),
      options: "options",
    },
    errorFields: {
      name: "name",
      message: "message",
      stack: "stack",
      cause: "cause",
      data$: "data$",
      httpMeta: sanitizeFieldName("httpMeta"),
      statusCode: sanitizeFieldName("statusCode"),
      contentType: sanitizeFieldName("contentType"),
      body: "body",
      rawResponse: sanitizeFieldName("rawResponse"),
      headers: "headers",
    },
  } as const;
}

const getConstants = () => {
  if (!constantsInitialized) {
    initConstants();
  }
  return context.GlobalComputed.Constants;
};

// Register constants as a global template variable
const templateConstants = () => getConstants();
registerTemplateFunc("constants", templateConstants);

let constants: ReturnType<typeof buildConstants>;
let constantsInitialized = false;
function initConstants() {
  if (constantsInitialized) {
    return;
  }
  constantsInitialized = true;
  constants = buildConstants();

  // Ensure GlobalComputed exists before assigning
  if (!context.GlobalComputed) {
    context.GlobalComputed = {};
  }
  context.GlobalComputed.Constants = constants;
}
