const SRC_ROOT = "src/";

// @ts-ignore
function sortFields(fields: FieldDef[]): FieldDef[] {
  const hasDataclassDefaultField = (field: FieldDef) => {
    if (field.Optional) {
      return true;
    }
    if (field.Const) {
      return true;
    }
    if (field.Default) {
      return true;
    }

    return false;
  };

  fields = fields.slice();

  const sorted = fields.sort((a, b) => {
    if (hasDataclassDefaultField(a) && !hasDataclassDefaultField(b)) {
      return 1;
    }
    if (!hasDataclassDefaultField(a) && hasDataclassDefaultField(b)) {
      return -1;
    }
    return 0;
  });

  return sorted;
}
registerTemplateFunc("sortFields", sortFields);

/**
 * Returns the moduleName configuration or packageName configuration after
 * filename sanitization and with path separators instead of period (.)
 * characters.
 */
function getModuleDirectory(): string {
  const moduleName: string = context.Global.Config.ModuleName;

  if (moduleName) {
    return moduleName
      .split(".")
      .map((part) => sanitizeFileName(part))
      .join("/");
  }

  return sanitizeFileName(context.Global.Config.PackageName);
}

registerTemplateFunc("getModuleDirectory", getModuleDirectory);

/**
 * Returns the root directory for the moduleName configuration or packageName
 * configuration.
 */
function getModuleRootDirectory(): string {
  const moduleName: string = context.Global.Config.ModuleName;

  if (moduleName) {
    return sanitizeFileName(moduleName.split(".")[0]);
  }

  return sanitizeFileName(context.Global.Config.PackageName);
}

registerTemplateFunc("getModuleRootDirectory", getModuleRootDirectory);

/**
 * Returns the moduleName configuration or packageName configuration after
 * filename sanitization.
 *
 * If subModuleImport is provided, it is joined with a period (.) character.
 */
function getModuleImport(subModuleImport?: string): string {
  const pkg = subModuleImport ?? "";

  if (useRelativeImports()) {
    const rel = relativeInternalImport(pkg);
    if (rel !== null) {
      return rel;
    }
    // File outside src/<pkg>/ (tests, docs, etc.) — relative imports would
    // escape the package; emit absolute instead.
    return getAbsoluteModuleImport(pkg);
  }

  const moduleImport = getModuleDirectory().replaceAll("/", ".");
  return pkg ? `${moduleImport}.${pkg}` : moduleImport;
}

/**
 * Computes the Python relative-import path from the currently-templated file
 * to `pkg` (a subpath under the SDK root). Returns null if the current file
 * lives outside `src/<packageName>/` — in which case the caller should emit
 * an absolute import instead.
 */
function relativeInternalImport(pkg: string): string | null {
  const outFile: string = context.OutFile || "";
  if (!outFile.startsWith(`${getSourcePath()}/`)) {
    return null;
  }
  // Drop the filename and convert directory separators to dots.
  const currentFilePkg = outFile
    .slice(SRC_ROOT.length)
    .split("/")
    .slice(0, -1)
    .join(".");

  const root = getModuleDirectory().replaceAll("/", ".");
  const targetAbs = pkg ? `${root}.${pkg}` : root;
  return pythonRelativeImport(currentFilePkg, targetAbs);
}

registerTemplateFunc("getModuleImport", getModuleImport);

/**
 * Returns true when the SDK is configured to emit relative imports for
 * SDK-internal references (`imports.relative: true` in gen.yaml).
 */
function useRelativeImports(): boolean {
  return context.Global.Config.Imports?.Relative === true;
}
registerTemplateFunc("useRelativeImports", useRelativeImports);

/** Module path for runtime `importlib.import_module` call sites. */
function getDynamicImportPath(module: string): string {
  if (useRelativeImports()) {
    return `.${module}`;
  }
  return getAbsoluteModuleImport(module);
}
registerTemplateFunc("getDynamicImportPath", getDynamicImportPath);

/**
 * Like getModuleImport but always returns an absolute dotted module path,
 * bypassing relative-imports emission. Required for runtime call sites like
 * `_sub_sdk_map` that feed `importlib.import_module` — relative names there
 * would need a `package=` argument which we don't want to add for every
 * customer.
 */
function getAbsoluteModuleImport(subModuleImport?: string): string {
  const root = getModuleDirectory().replaceAll("/", ".");
  return subModuleImport ? `${root}.${subModuleImport}` : root;
}
registerTemplateFunc("getAbsoluteModuleImport", getAbsoluteModuleImport);

/**
 * Returns the source path for the moduleName configuration or packageName
 * configuration.
 */
function getSourcePath(): string {
  return `${SRC_ROOT}${getModuleDirectory()}`;
}

registerTemplateFunc("getSourcePath", getSourcePath);

// @ts-ignore
function inSameModel(
  modelName: string,
  outputLocation: string,
  typeDef: TypeDef,
): boolean {
  if (!typeDef.IsCustomType()) {
    return true;
  }

  if (!typeDef.ResolvedModel) {
    throw new Error(`Type ${typeDef.Name} is not resolved`);
  }

  return (
    `${outputLocation}${modelName}` ===
    `${getModelsLocation(typeDef.OutputLocation)}${typeDef.ResolvedModel}`
  );
}

function hasAdditionalPropertiesField(typeDef: TypeDef): boolean {
  return typeDef.Fields.some((f) => {
    return f.IsAdditionalProperties;
  });
}
registerTemplateFunc(
  "hasAdditionalPropertiesField",
  hasAdditionalPropertiesField,
);

function hasAdditionalPropertiesFieldRecursive(
  typeDef: TypeDef,
  visited: Record<string, boolean> = {},
): boolean {
  if (!typeDef || !typeDef.IsCustomType) {
    return false;
  }
  if (typeDef.IsCustomType()) {
    const uniqueID = getUniqueID(typeDef);

    if (typeDef.Truncated || visited[uniqueID]) {
      return true;
    }

    visited[uniqueID] = true;
  }

  return (
    typeDef.Fields?.some((f) => {
      return (
        f.IsAdditionalProperties ||
        hasAdditionalPropertiesFieldRecursive(f.Type, visited)
      );
    }) ||
    typeDef.AssociatedTypes?.some((t) =>
      hasAdditionalPropertiesFieldRecursive(t, visited),
    )
  );
}

// @ts-ignore
function transformAdditionalContext(
  fieldDef: FieldDef,
  additionalContext?: TemplateValueContext,
): TemplateValueContext {
  if (hasAdditionalPropertiesFieldRecursive(fieldDef.Type)) {
    return {
      ...(additionalContext ?? {}),
      typedDictDisabled: true,
    };
  }

  return additionalContext;
}

function getAdditionalPropertiesField(typeDef: TypeDef): FieldDef | undefined {
  return typeDef.Fields.find((f) => {
    return f.IsAdditionalProperties;
  });
}
registerTemplateFunc(
  "getAdditionalPropertiesField",
  getAdditionalPropertiesField,
);

function getAdditionalPropertiesFieldName(typeDef: TypeDef): string {
  return (
    typeDef.Fields.find((f) => {
      return f.IsAdditionalProperties;
    })?.Name ?? ""
  );
}
registerTemplateFunc(
  "getAdditionalPropertiesFieldName",
  getAdditionalPropertiesFieldName,
);

function hasOpenEnumField(typeDef: TypeDef): boolean {
  return typeDef.Fields.some((f) => f.Type.Enum?.Open);
}
registerTemplateFunc("hasOpenEnumField", hasOpenEnumField);

function openEnums(typeDef: TypeDef): FieldDef[] {
  const seen = new Set<string>();
  return typeDef.Fields.filter((f) => {
    if (!f.Type.Enum?.Open) {
      return false;
    }
    // Deduplicate by sanitized field name to avoid generating duplicate serializer methods
    // This can happen when allOf wraps a single ref to an open enum
    const fieldName = sanitizeFieldName(f.Name);
    if (seen.has(fieldName)) {
      return false;
    }
    seen.add(fieldName);
    return true;
  });
}
registerTemplateFunc("openEnums", openEnums);

function hasBase64FileInputField(typeDef: TypeDef): boolean {
  return typeDef.Fields.some(isBase64FileInputField);
}
registerTemplateFunc("hasBase64FileInputField", hasBase64FileInputField);

function base64FileInputFields(typeDef: TypeDef): FieldDef[] {
  const seen = new Set<string>();
  return typeDef.Fields.filter((f) => {
    if (!isBase64FileInputField(f)) {
      return false;
    }
    const fieldName = sanitizeFieldName(f.Name);
    if (seen.has(fieldName)) {
      return false;
    }
    seen.add(fieldName);
    return true;
  });
}
registerTemplateFunc("base64FileInputFields", base64FileInputFields);

// Returns a safe method name for the @model_serializer method.
// Normally "serialize_model", but if an open enum field would generate a
// @field_serializer method named "serialize_model" (i.e. a field named
// "model"), we fall back to "_serialize_model" to avoid a collision.
function sanitizeSerializeMethodName(typeDef: TypeDef): string {
  const defaultName = "serialize_model";
  const hasCollision = openEnums(typeDef).some(
    (f) => !f.Const && sanitizeFieldName(f.Name) === "model",
  );
  if (hasCollision) {
    return "_serialize_model";
  }
  return defaultName;
}
registerTemplateFunc(
  "sanitizeSerializeMethodName",
  sanitizeSerializeMethodName,
);

function needsTypedDictPassStatement(
  typeDef: TypeDef,
  comments: string,
): boolean {
  if (comments) {
    return false;
  }

  if (sanitizeClassName(typeDef.Name).startsWith("Test")) {
    return false;
  }

  const includedFields = typeDef.Fields.filter((f) => {
    return !f.IsAdditionalProperties;
  });

  return includedFields.length == 0;
}
registerTemplateFunc(
  "needsTypedDictPassStatement",
  needsTypedDictPassStatement,
);

function typeIncludesPydanticModel(
  typeDef?: TypeDef,
  visited: Record<string, boolean> = {},
): boolean {
  if (!typeDef) {
    return false;
  }

  // For custom types, check if we've already visited this type to prevent infinite recursion
  if (typeDef.IsCustomType()) {
    const uniqueID = typeDef.GetRegistrationID();
    if (visited[uniqueID]) {
      return false;
    }
    visited[uniqueID] = true;
  }

  switch (typeDef.Type.toString()) {
    case "class":
      return true;
    case "union":
      return typeDef.AssociatedTypes.some((t) =>
        typeIncludesPydanticModel(t, visited),
      );
    case "array":
    case "map":
      return typeIncludesPydanticModel(typeDef.ItemType, visited);
    default:
      return false;
  }
}

registerTemplateFunc("typeIncludesPydanticModel", typeIncludesPydanticModel);

function typeIncludesKind(
  typeDef: TypeDef | undefined,
  kind: "array" | "map",
  visited: Record<string, boolean> = {},
): boolean {
  if (!typeDef) {
    return false;
  }

  if (typeDef.IsCustomType()) {
    const uniqueID = typeDef.GetRegistrationID();
    if (visited[uniqueID]) {
      return false;
    }
    visited[uniqueID] = true;
  }

  switch (typeDef.Type.toString()) {
    case kind:
      return true;
    case "union":
      return typeDef.AssociatedTypes.some((t) =>
        typeIncludesKind(t, kind, visited),
      );
    case "array":
    case "map":
      return typeIncludesKind(typeDef.ItemType, kind, visited);
    case "class":
      return typeDef.Fields.some((field) =>
        typeIncludesKind(field.Type, kind, visited),
      );
    default:
      return false;
  }
}

function typeIncludesArray(
  typeDef?: TypeDef,
  visited: Record<string, boolean> = {},
): boolean {
  return typeIncludesKind(typeDef, "array", visited);
}

registerTemplateFunc("typeIncludesArray", typeIncludesArray);

function typeIncludesMap(
  typeDef?: TypeDef,
  visited: Record<string, boolean> = {},
): boolean {
  return typeIncludesKind(typeDef, "map", visited);
}

registerTemplateFunc("typeIncludesMap", typeIncludesMap);

// @ts-ignore
function sortModelTypes(types: PythonV2Type[]): PythonV2Type[] {
  if (useConflictResistantModelImports()) {
    // preserve order from TopologicalSortTypeDefs(types, true)
    return types;
  }

  types = types.slice();
  // Sort types so that union members are defined before the union itself.
  // Python evaluates Union members immediately (not lazily), so member classes
  // must be defined before the union type alias that references them.
  return types.sort((a, b) => {
    const aIsUnion = a.Type.Type.toString() == "union";
    const bIsUnion = b.Type.Type.toString() == "union";

    // If a union depends on a non-union, the non-union should come first
    if (aIsUnion && !bIsUnion && isTypeDependentOn(a.Type, b.Type)) {
      return 1; // move union (a) after non-union (b)
    }
    if (bIsUnion && !aIsUnion && isTypeDependentOn(b.Type, a.Type)) {
      return -1; // move union (b) after non-union (a)
    }

    return 0;
  });
}
registerTemplateFunc("sortModelTypes", sortModelTypes);

function isTypeDependentOn(parent: TypeDef, target: TypeDef): boolean {
  switch (true) {
    case parent.Type.toString() == "class" || parent.Type.toString() == "error":
      for (const field of parent.Fields) {
        let typeToCheck = field.Type;

        switch (typeToCheck.Type.toString()) {
          case "array":
          case "map":
            typeToCheck = typeToCheck.ItemType;
            break;
        }

        if (typeToCheck.IsEqualType(target)) {
          return true;
        }
      }
      break;
    case parent.Type.toString() == "union":
      for (const at of parent.AssociatedTypes) {
        let typeToCheck = at;

        switch (typeToCheck.Type.toString()) {
          case "array":
          case "map":
            typeToCheck = typeToCheck.ItemType;
            break;
        }

        if (typeToCheck.IsEqualType(target)) {
          return true;
        }
      }
      break;
    case parent.Type.toString() == "enum":
      break;
    default:
      throw new Error(`Unexpected type: ${parent.Type.toString()}`);
  }

  return false;
}

function identifyCircularTypes(types: TypeDef[]): TypeDef[] {
  const circularTypes = new Set<TypeDef>();

  const numTypes = types.length;
  // Compare each type with every other type
  for (let left = 0; left < numTypes; left++) {
    for (let right = left + 1; right < numTypes; right++) {
      if (areCircular(types[left], types[right])) {
        circularTypes.add(types[left]);
        circularTypes.add(types[right]);
      }
    }
  }
  return Array.from(circularTypes);
}

function sortBodyParamFields(fields: FieldDef[]): FieldDef[] {
  fields = fields.slice();
  // Sort additional properties to the end
  return fields.sort((a, b) => {
    if (a.IsAdditionalProperties && !b.IsAdditionalProperties) {
      return 1;
    }
    if (!a.IsAdditionalProperties && b.IsAdditionalProperties) {
      return -1;
    }
    return 0;
  });
}
registerTemplateFunc("sortBodyParamFields", sortBodyParamFields);

function isReturnTypeOptional(
  response: ResponseDef,
  pagination?: PaginationConfig,
): boolean {
  const responseFormat = getResponseFormat();

  if (pagination) {
    return true;
  }

  if (responseFormat !== "flat" || response.Type.ResponseEnvelope) {
    return false;
  }

  let numSuccessfulResponses = 0;
  let numEmptyResponses = 0;

  for (const subRes of response.Responses) {
    if (subRes.Error) {
      continue;
    }

    numSuccessfulResponses++;
    if (subRes.Content.length == 0) {
      numEmptyResponses++;
    }
  }

  if (numSuccessfulResponses > 1 && numEmptyResponses > 0) {
    return true;
  }

  return false;
}
registerTemplateFunc("isReturnTypeOptional", isReturnTypeOptional);

function containsStreamingResponse(response: ResponseDef): boolean {
  return response.Responses.some((r) =>
    r.Content.some(
      (c) =>
        c.SerializationMethod == "eventstream" ||
        c.SerializationMethod == "jsonl" ||
        c.Content.Type.Type.toString() == "response-stream",
    ),
  );
}
registerTemplateFunc("containsStreamingResponse", containsStreamingResponse);

function containsEventStreamResponse(response: ResponseDef): boolean {
  return response.Responses.some((r) =>
    r.Content.some((c) => c.SerializationMethod == "eventstream"),
  );
}

//@ts-ignore
function shouldTemplateConstValue(
  field: FieldDef,
  additionalContext?: TemplateValueContext,
): boolean {
  if (!field.Const) {
    return false;
  }

  if (additionalContext?.typedDictDisabled) {
    return false;
  }

  return true;
}

function includeAsyncProp(): boolean {
  return context.Global.Config.AsyncMode == "both";
}
registerTemplateFunc("includeAsyncProp", includeAsyncProp);

function isSplitMode() {
  return context.Global.Config.AsyncMode == "split";
}
registerTemplateFunc("isSplitMode", isSplitMode);

function raiseError(msg: string) {
  throw new Error(msg);
}
registerTemplateFunc("raiseError", raiseError);

// @ts-ignore
function needsModelRebuild(typeDef: TypeDef): boolean {
  const typeType = typeDef.Type.toString();
  if (typeType === "enum" || typeType === "error" || typeType === "union") {
    return false;
  }

  return typeDef.Fields.some((field) => {
    if (field.Const) {
      return true;
    }

    if (!field.Annotations?.Has("json")) {
      return false;
    }

    const jsonAnnotation = field.Annotations.Get("json") as JSONAnnotation;
    const sanitizedFieldName = sanitizeFieldName(field.Name);

    return (
      jsonAnnotation.FieldName !== "-" &&
      sanitizedFieldName !== jsonAnnotation.FieldName
    );
  });
}
registerTemplateFunc("needsModelRebuild", needsModelRebuild);
