function dataType(dataType: DataType): GojaEnum<DataType> {
  return {
    toString: () => dataType.toString(),
    valueOf: () => dataType,
  };
}

function sliceString(start: number, end: number, val: string): string {
  return val.slice(start, end);
}
registerTemplateFunc("sliceString", sliceString);

function flattenOperationsPerSDK(sdk: SDK): Operation[] {
  let operations: Operation[] = [...sdk.Operations];

  if (sdk.SubSDKs.length == 0) {
    return operations;
  }

  for (const subSDK of sdk.SubSDKs) {
    let localOperations: Operation[] = flattenOperationsPerSDK(subSDK);
    operations = [...operations, ...localOperations];
  }

  return operations;
}

registerTemplateFunc("flattenOperationsPerSDK", flattenOperationsPerSDK);

function operationHasAttemptCountRetries(operation: Operation): boolean {
  return operation.Extensions?.Retries?.Strategy === "attempt-count-backoff";
}
registerTemplateFunc(
  "operationHasAttemptCountRetries",
  operationHasAttemptCountRetries,
);

function hasAttemptCountRetries(
  sdk: SDK = context.Global.AST.MainSDK,
): boolean {
  return flattenOperationsPerSDK(sdk).some(operationHasAttemptCountRetries);
}
registerTemplateFunc("hasAttemptCountRetries", hasAttemptCountRetries);

/**
 * Recursively flattens an SDK and all its nested SubSDKs into a single array.
 */
function flattenSubSDKs(rootSDK: SDK): SDK[] {
  return [rootSDK, ...rootSDK.SubSDKs.flatMap(flattenSubSDKs)];
}

function isSSEFlatResponseEnabled(): boolean {
  const value = context.Global.Config.SseFlatResponse;
  return value === true || value === "true";
}
registerTemplateFunc("isSSEFlatResponseEnabled", isSSEFlatResponseEnabled);

/**
 * Gets the union deserialization strategy from config.
 * Returns either "populated-fields" or "left-to-right".
 * This function provides a unified way to access the unionStrategy config
 * across all SDK targets (Go, Java, Terraform, TypeScript, etc.)
 */
function getUnionStrategy(): string {
  return context.Global.Config.UnionStrategy || "left-to-right";
}
registerTemplateFunc("getUnionStrategy", getUnionStrategy);

/**
 * Checks if the union deserialization strategy is set to "populated-fields".
 * This is a convenience function to make template conditions more readable.
 */
function isUnionStrategyPopulatedFields(): boolean {
  return getUnionStrategy() === "populated-fields";
}
registerTemplateFunc(
  "isUnionStrategyPopulatedFields",
  isUnionStrategyPopulatedFields,
);

/**
 * Gets the data field from an event-stream item type if SSE flat response is enabled.
 * Returns null if flattening is disabled or if no data field exists.
 */
function getFlattenedEventStreamField(itemType: TypeDef): FieldDef | null {
  // Check if SSE flat response is enabled
  if (!isSSEFlatResponseEnabled()) {
    return null;
  }

  // Check if the item type is a class with fields
  if (itemType?.Type?.toString() !== "class" || !itemType?.Fields) {
    return null;
  }

  // Find and return the data field if it exists
  return itemType.Fields.find((f: FieldDef) => f.Name === "data") || null;
}
registerTemplateFunc(
  "getFlattenedEventStreamField",
  getFlattenedEventStreamField,
);

/**
 * Checks if the data field of an SSE event type is required (non-optional).
 * When true, events without data lines should be skipped.
 * When false, data-less events should pass through (e.g. OptionalDataEvent).
 */
function isSSEDataRequired(itemType: TypeDef): boolean {
  if (itemType?.Type?.toString() === "class" && itemType?.Fields) {
    const dataField = itemType.Fields.find((f: FieldDef) => f.Name === "data");
    if (dataField) return !dataField.Optional;
  }
  // For unions: if any variant has required data, be conservative and skip
  if (itemType?.Type?.toString() === "union" && itemType?.AssociatedTypes) {
    return itemType.AssociatedTypes.some((t: TypeDef) => isSSEDataRequired(t));
  }
  return true; // safe default — skip data-less events
}
registerTemplateFunc("isSSEDataRequired", isSSEDataRequired);

// @ts-ignore
function resolveModelName(typeDef: TypeDef): string {
  return sanitizeFileName(typeDef.ResolvedModel);
}

function escapeString(val: string) {
  return val.replaceAll(`"`, `\\"`);
}

registerTemplateFunc("escapeString", escapeString);

function includes(arr: any[], val: any) {
  return arr.includes(val);
}

registerTemplateFunc("includes", includes);

function caseInsensitiveIncludes(arr: string[], val: string) {
  return arr.some((v) => v.toLowerCase() === val.toLowerCase());
}

registerTemplateFunc("caseInsensitiveIncludes", caseInsensitiveIncludes);

function getParameterType(fieldDef: FieldDef): string {
  let anno = fieldDef.Annotations.Get("param") as ParamAnnotation;
  if (anno) {
    return anno.ParamType;
  }

  return "";
}

registerTemplateFunc("getParameterType", getParameterType);

function getVisiblePathParams(op: Operation): ParamDef[] {
  if (op.Request?.Params == null) {
    return [];
  }

  return op.Request.Params.PathParams.filter((p) => !p.Hidden);
}

registerTemplateFunc("getVisiblePathParams", getVisiblePathParams);

function getVisibleQueryParams(op: Operation): ParamDef[] {
  if (op.Request?.Params == null) {
    return [];
  }

  return op.Request.Params.QueryParams.filter((p) => !p.Hidden);
}

registerTemplateFunc("getVisibleQueryParams", getVisibleQueryParams);

function anyRetries(): boolean {
  return Boolean(
    context.Global.AST.MainSDK.Operations.find((op) => {
      if (op.Extensions["Retries"]) {
        return true;
      }
    }) ||
      context.Global.AST.MainSDK.SubSDKs.find((sub) => {
        return flattenOperationsPerSDK(sub).find((op) => {
          if (op.Extensions["Retries"]) {
            return true;
          }
        });
      }),
  );
}

function anyJsonOrStringContents(res: ResponseDef): boolean {
  return Boolean(
    res.Responses.find((sub) => {
      return sub.Content.find((content) => {
        return (
          content.SerializationMethod == "json" ||
          content.SerializationMethod == "string"
        );
      });
    }),
  );
}

registerTemplateFunc("anyJsonOrStringContents", anyJsonOrStringContents);

function getMaxMethodParameters(operation: Operation): number {
  return 0;
}

function operationParametersFlattened(operation: Operation): boolean {
  return (
    !!operation.Request &&
    operation.MaxMethodParams > 0 &&
    operation.Request.Params &&
    operation.Request.Field.Type.Type == "class" &&
    operation.Request.Field.Type.Fields.length <= operation.MaxMethodParams
  );
}

registerTemplateFunc(
  "operationParametersFlattened",
  operationParametersFlattened,
);

function sanitizeDeprecationReplacement(
  deprecationReplacement: string,
  type: "method" | "field" | "class",
): string {
  if (!deprecationReplacement) {
    return "";
  }

  let replacement = sanitizeFieldName(deprecationReplacement);
  if (type == "method") {
    const operation = context.Global.AST.MainSDK.FindOperation(
      deprecationReplacement,
    );
    if (operation) {
      replacement = sanitizeMethodName(operation);
    } else {
      replacement = "";
    }
  }

  return replacement;
}

registerTemplateFunc(
  "sanitizeDeprecationReplacement",
  sanitizeDeprecationReplacement,
);

function getPaginationInput(
  pagination: PaginationConfig,
  type: string,
): PaginationInputs | undefined {
  return pagination.Inputs.find((i) => i.Type.toString() === type);
}

registerTemplateFunc("getPaginationInput", getPaginationInput);

// Finds a pagination field by original (spec) name — checks top-level
// request fields first, then drills into nested body fields marked with
// the "request" annotation. Returns undefined if not found.
// @ts-ignore
function findPaginationFieldDeep(
  requestType: TypeDef,
  fieldName: string,
): FieldDef | undefined {
  if (!requestType?.Fields) return undefined;

  const direct = requestType.Fields.find(
    (f) => originalFieldName(f) === fieldName,
  );
  if (direct) return direct;

  for (const field of requestType.Fields) {
    if (
      field.Type?.Type?.toString() !== "class" ||
      !field.Annotations?.Has("request")
    ) {
      continue;
    }
    const nested = field.Type.Fields?.find(
      (f: FieldDef) => originalFieldName(f) === fieldName,
    );
    if (nested) return nested;
  }

  return undefined;
}
registerTemplateFunc("findPaginationFieldDeep", findPaginationFieldDeep);

function isPaginationInputInBody(op: Operation): boolean {
  if (!op.Extensions?.Pagination) {
    return false;
  }

  if (op.Request.Field?.Type.Type.toString() !== "class") {
    return false;
  }

  for (const inp of op.Extensions.Pagination.Inputs) {
    if (inp.In.toString() === "requestBody") {
      return true;
    }
  }

  return false;
}
registerTemplateFunc("isPaginationInputInBody", isPaginationInputInBody);

function findWrappedRequestField(op: Operation): FieldDef | undefined {
  if (hasAnnotation(op.Request?.Field, "requestWrapper")) {
    return op.Request?.Field.Type?.Fields.find((f) =>
      hasAnnotation(f, "request"),
    );
  }
}

function hasPaginationURL(op: Operation): boolean {
  return op.Extensions.Pagination?.Type.toString() === "url";
}
registerTemplateFunc("hasPaginationURL", hasPaginationURL);

function sdkHasPaginationURL(ops: Operation[]): boolean {
  for (const op of ops) {
    if (opHasPaginationURL(op)) {
      return true;
    }
  }
  return false;
}
registerTemplateFunc("sdkHasPaginationURL", sdkHasPaginationURL);

function opHasPaginationURL(op: Operation): boolean {
  return op.Extensions.Pagination?.Type.toString() === "url";
}
registerTemplateFunc("opHasPaginationURL", opHasPaginationURL);

function sortMethodParams(params: FieldDef[]): FieldDef[] {
  params = params.slice();

  return params.sort((a, b) => {
    if (a.Optional && !b.Optional) {
      return 1;
    }

    if (!a.Optional && b.Optional) {
      return -1;
    }

    return 0;
  });
}

registerTemplateFunc("sortMethodParams", sortMethodParams);

function getErrorMessageField(type: TypeDef): FieldDef | undefined {
  return type.Fields.find((f) => f.ErrorMessage);
}

registerTemplateFunc("getErrorMessageField", getErrorMessageField);

function inferErrorMessageField(type: TypeDef): FieldDef[] | undefined {
  // If `err.message` is a string, return it
  let f = type.Fields.find(
    (f) => f.Name === "message" && f.Type.Type.toString() === "string",
  );
  if (f) return [f];

  // If `err.error` is a class
  const errorField = type.Fields.find(
    (f) => f.Name === "error" && f.Type.Type.toString() === "class",
  );
  if (errorField) {
    // If `err.error.message` is a string, return it
    f = errorField.Type.Fields.find(
      (f) => f.Name === "message" && f.Type.Type.toString() === "string",
    );
    if (f) return [errorField, f];
  }
}
registerTemplateFunc("inferErrorMessageField", inferErrorMessageField);

function getErrorMessageFieldDeep(
  type: TypeDef,
  visited: Record<string, boolean> = {},
): FieldDef[] | undefined {
  if (type.IsCustomType()) {
    if (visited[type.GetRegistrationID()]) return;
    visited[type.GetRegistrationID()] = true;
  }

  for (const subField of type.Fields) {
    if (subField.ErrorMessage) return [subField];

    const result = getErrorMessageFieldDeep(subField.Type, visited);
    if (result) return [subField, ...result];
  }
}
registerTemplateFunc("getErrorMessageFieldDeep", getErrorMessageFieldDeep);

function getOrInferErrorMessageFieldDeep(
  type: TypeDef,
): FieldDef[] | undefined {
  return getErrorMessageFieldDeep(type) ?? inferErrorMessageField(type);
}
registerTemplateFunc(
  "getOrInferErrorMessageFieldDeep",
  getOrInferErrorMessageFieldDeep,
);

function generateStandaloneUsage(
  commentPrefix: string,
  fileExtension: string,
): void {
  for (const usageContext of context.Global.Config.Usage.UsageContexts) {
    const fileName = `${usageContext.Operation.ID}.usage.${fileExtension}`;

    faker.seed(usageContext.Operation.GetExampleSeed());
    let files = [templateString("usage/snippet.stmpl", usageContext)];

    // The heading comment is necessary for the usage snippets tests. We can strip them
    // when generating the standalone usage snippets.
    if (commentPrefix !== "" && fileExtension !== "json") {
      const headingComment = `${commentPrefix} Usage snippet provided for ${usageContext.Operation.ID} (${usageContext.Operation.Method} ${usageContext.Operation.Path})`;
      files.unshift(headingComment);
    }

    let result = files.join("\n");

    // Add three newlines to the end of the file
    result += "\n\n\n";

    writeFile(fileName, result, 0, false);
  }
}

//@ts-ignore
function getTypeRelativeDocsPath(outputLocation: string) {
  const parts = outputLocation.split("/");
  if (parts.length == 0) {
    return ".";
  }

  return `${"../".repeat(parts.length)}${outputLocation}`;
}

function relativePath(source: string, target: string) {
  const sourceParts = source.split("/");
  const targetParts = target.split("/");

  let i = 0;
  for (; i < sourceParts.length; i++) {
    if (sourceParts[i] !== targetParts[i]) {
      break;
    }
  }

  let result = "../".repeat(sourceParts.length - i) || "./";

  result += targetParts.slice(i, targetParts.length - 1).join("/");

  result = result.replace(/\/+/g, "/").replace(/\/$/g, "");

  return result;
}

function findExampleByName(
  examples: Example[],
  name: string,
): Example | undefined {
  return examples.find((e) => !name || e.Name() === name);
}

function getOperationMethodFieldExample(
  usageContext: UsageContext,
  field: FieldDef,
): any | ExampleReferenceValue {
  const exampleName = usageContext.ExampleName;
  const op = usageContext.Operation;

  switch (true) {
    case op.Arguments.Flattening == "none" &&
      !op.Arguments.IsBodyField(field) &&
      !field.Annotations?.Has("param"):
      const exampleObject = {};

      for (const f of field.Type.Fields) {
        // TODO this is going to suffer from parameter name collisions until we make more things aware of this and populate using the field.Name instead of field.OriginalName
        const fieldExample = getOperationMethodFieldExample(usageContext, f);

        if (fieldExample !== undefined) {
          exampleObject[originalFieldName(f)] = fieldExample;
        }
      }

      if (Object.keys(exampleObject).length > 0) {
        return exampleObject;
      }

      return undefined;
    // The field is the request body
    case op.Arguments.IsBodyField(field) || field.Annotations?.Has("request"):
      const example = findExampleByName(
        op.Request?.Examples ?? [],
        exampleName,
      );

      const value = getExampleValue(example, {
        usageContext: usageContext,
        test: usageContext.Test,
      });
      return value;
    // The field is a property of the request body but not an additionalProperties field
    case op.Arguments.IsFieldInBody(field) && !field.IsAdditionalProperties: {
      const requestExample = getExampleValue(
        findExampleByName(op.Request?.Examples ?? [], exampleName),
        { usageContext: usageContext, test: usageContext.Test },
      );
      return requestExample?.[originalFieldName(field)];
    }
    // The field is an additionalProperties field of the request body
    case op.Arguments.IsFieldInBody(field) && field.IsAdditionalProperties: {
      const requestExample = getExampleValue(
        findExampleByName(op.Request?.Examples ?? [], exampleName),
        { usageContext: usageContext, test: usageContext.Test },
      );

      // Copy to avoid mutating the shared AST example data
      const remainingExample = requestExample
        ? { ...requestExample }
        : undefined;

      for (const field of op.Arguments.Sorted) {
        if (field.IsAdditionalProperties) {
          continue;
        }

        delete remainingExample?.[originalFieldName(field)];
      }

      return remainingExample;
    }
    // The field is a parameter field
    case field.Annotations?.Has("param"):
      const paramAnno = field.Annotations.Get("param") as ParamAnnotation;

      let params: ParamDef[] = [];

      switch (paramAnno.ParamType) {
        case "pathParam":
          params = op.Request.Params.PathParams;
          break;
        case "queryParam":
          params = op.Request.Params.QueryParams;
          break;
        case "header":
          params = op.Request.Params.HeaderParams;
          break;
      }

      for (const param of params) {
        if (param.Field.Name === field.Name) {
          const paramExample = getExampleValue(
            findExampleByName(param.Examples ?? [], exampleName),
            { usageContext: usageContext, test: usageContext.Test },
          );

          return paramExample;
        }
      }
  }

  return undefined;
}
registerTemplateFunc(
  "getOperationMethodFieldExample",
  getOperationMethodFieldExample,
);

// @ts-ignore
function isRequestBodyOptional(
  request: RequestDef | null | undefined,
): boolean {
  if (!request?.IsRequestBody) {
    return false;
  }

  return request.RequestBody.Optional || request.RequestBody.Nullable;
}
registerTemplateFunc("isRequestBodyOptional", isRequestBodyOptional);

function isRequestOptional(request?: RequestDef): boolean {
  return request?.Field?.Optional || request?.Field?.Nullable;
}
registerTemplateFunc("isRequestOptional", isRequestOptional);

// @ts-ignore
function hasRawResponseField(
  type: TypeDef,
  visited: Record<string, boolean> = {},
): boolean {
  const uniqueID = getUniqueID(type);
  if (visited[uniqueID]) {
    return false;
  }

  visited[uniqueID] = true;
  switch (type.Type.toString()) {
    case "union":
      if (type.Discriminator) {
        return type.Discriminator.Mapping.some((m) =>
          hasRawResponseField(m.Type, visited),
        );
      } else {
        return type.AssociatedTypes.some((t) =>
          hasRawResponseField(t, visited),
        );
      }
    default:
      // return true if one of the fields is a response
      return type.Fields.some((field) => {
        switch (field.Type.Type.toString()) {
          case DataTypeValue.Response.toString():
            return true;
          case DataTypeValue.Class.toString():
            return hasRawResponseField(field.Type, visited);
          default:
            return false;
        }
      });
  }
}

registerTemplateFunc("hasRawResponseField", hasRawResponseField);

function getResultField(op: Operation): FieldDef | null {
  const fields = op.Response.Type ? op.Response.Type.Fields : [];

  for (const field of fields) {
    for (const annotation of field.Annotations) {
      if (isResponseAnnotation(annotation) && annotation.ResultField) {
        return field;
      }
    }
  }

  return null;
}
registerTemplateFunc("getResultField", getResultField);

// This is exporting a go func
registerTemplateFunc("sortTypeDefFieldCount", sortTypeDefFieldCount);
registerTemplateFunc(
  "sortTypeDefRequiredFieldsDescending",
  sortTypeDefRequiredFieldsDescending,
);

function isUnionOfErrors(typeDef: TypeDef): boolean {
  if (typeDef.Type.toString() !== "union") {
    return false;
  }

  return typeDef.AssociatedTypes.every((t) => {
    if (t.Type.toString() === "error") {
      return true;
    }
    return isUnionOfErrors(t);
  });
}
registerTemplateFunc("isUnionOfErrors", isUnionOfErrors);

// @ts-ignore
function getEnumFormat(typeDef: TypeDef): "enum" | "union" {
  const raw = typeDef.Enum.Format || "enum";

  switch (raw) {
    case "enum":
      return "enum";
    case "union":
      return "union";
    default:
      throw new Error(`Unknown enum format: ${raw}`);
  }
}
registerTemplateFunc("getEnumFormat", getEnumFormat);

function assert(condition: any, message: string): asserts condition {
  if (!condition) {
    throw new AssertionError(message);
  }
}

class AssertionError extends Error {
  constructor(message: string) {
    super(`AssertionError: ${message}`);
  }
}

const statusCodes = {
  1: ["100", "101", "102", "103"],
  2: ["200", "201", "202", "203", "204", "205", "206", "207", "208", "226"],
  3: ["300", "301", "302", "303", "304", "305", "306", "307", "308"],
  4: [
    "400",
    "401",
    "402",
    "403",
    "404",
    "405",
    "406",
    "407",
    "408",
    "409",
    "410",
    "411",
    "412",
    "413",
    "414",
    "415",
    "416",
    "417",
    "418",
    "421",
    "422",
    "423",
    "424",
    "425",
    "426",
    "428",
    "429",
    "431",
    "451",
  ],
  5: [
    "500",
    "501",
    "502",
    "503",
    "504",
    "505",
    "506",
    "507",
    "508",
    "510",
    "511",
  ],
};

function getStatusCodesForRange(statusCode: string): string[] {
  if (!statusCode.toLowerCase().endsWith("xx")) {
    return [statusCode];
  }

  const statusCodeMajor = statusCode[0];

  return statusCodes[statusCodeMajor] ?? [statusCode];
}

function getValidStatusCodes(): string[] {
  const codes: string[] = [];

  for (const major in statusCodes) {
    if (major !== "1") {
      codes.push(...statusCodes[major]);
    }
  }

  return codes;
}

function isNumeric(val: unknown): val is number {
  return typeof val === "string" && !Number.isNaN(Number(val));
}

// @ts-ignore
function isReactQueryEnabled(): boolean {
  return false;
}

// isExampleReferenceValue is a type guard for checking if a value is an ExampleReferenceValue
function isExampleReferenceValue(value: any): value is ExampleReferenceValue {
  return (
    typeof value === "object" &&
    value !== null &&
    ["isExampleReferenceValue", "path", "source"].every((key) => key in value)
  );
}

// isFieldDef is a type guard for checking if a value is a FieldDef
function isFieldDef(value: any): value is FieldDef {
  return (
    typeof value === "object" &&
    value !== null &&
    ["Nullable", "Optional"].every((key) => key in value)
  );
}

type PaginationDefaults = {
  limit?: number;
  offset?: number;
  page?: number;
  cursor?: string | number;
};

const sensiblePaginationDefaults: PaginationDefaults = {
  offset: 0,
  /**
   * This follows what we think is the convention for page-based pagination.
   * What we've observed APIs that use this scheme tend to start at page 1.
   */
  page: 1,
  /**
   * If you trace how the limit field is used in pagination code, you'll see
   * it's not fed into requests if it's nil - the server decides the default.
   * It's only ever used to compare against the size of the result array that is
   *  returned. Generally, the code looks like this:
   *
   *     if len(results) == 0 {
   *       return nil, nil  // stop paginating
   *     }
   *
   *     if len(results) < limit {
   *       return nil, nil  // We got less than we asked for so it must be the last page
   *     }
   *
   * Since the first condition is always hit first, the default limit of 0 is
   * sensible. What will end up happening in the `limit == 0` case is we
   * potentially make an additional, unnecessary, fetch. That's more desirable
   * than guessing an arbitrary default value to use for all customers.
   */
  limit: 0,
};

function getPaginationDefaults(op: Operation): PaginationDefaults {
  const pagination = op.Extensions.Pagination;
  if (pagination == null) {
    throw new Error(
      `${op.ID}: getPaginationDefaults: expected does not have a pagination extension`,
    );
  }

  if (op.Request == null) {
    throw new Error(
      `${op.ID}: getPaginationDefaults: expected operation does not have a request defined`,
    );
  }

  const out: PaginationDefaults = {};

  for (const input of pagination.Inputs) {
    let field: FieldDef | undefined;
    const loc = input.In.toString();
    if (loc === "parameters") {
      if (op.Request.Params == null) {
        throw new Error(
          `${op.ID}: getPaginationDefaults: operation does not have parameters`,
        );
      }

      field = findParamFieldByName(op.Request.Params, input.Name);
    } else if (loc === "requestBody") {
      if (op.Request.RequestBody == null) {
        throw new Error(
          `${op.ID}: getPaginationDefaults: operation does not have a request body`,
        );
      }

      field = findBodyFieldByName(op.Request.RequestBody.Type, input.Name);
    } else {
      throw new Error(
        `${op.ID}: getPaginationDefaults: unrecognized pagination input location: ${loc}`,
      );
    }

    out[input.Type.toString()] =
      field?.Default?.Value ??
      sensiblePaginationDefaults[input.Type.toString()];
  }

  return out;
}

registerTemplateFunc("getPaginationDefaults", getPaginationDefaults);

registerTemplateFunc("stringify", JSON.stringify);

/** Returns the matching request ParamDef by field name. */
function findParamByName(
  reqParams: RequestParams,
  name: string,
): ParamDef | undefined {
  for (const paramDef of [
    reqParams.QueryParams,
    reqParams.PathParams,
    reqParams.HeaderParams,
  ]) {
    for (const param of paramDef) {
      if (originalFieldName(param.Field) === name) {
        return param;
      }
    }
  }
}

/** Returns the matching request parameter FieldDef by field name. */
// @ts-ignore
function findParamFieldByName(
  reqParams: RequestParams,
  name: string,
): FieldDef | undefined {
  return findParamByName(reqParams, name)?.Field;
}

function findBodyFieldByName(
  requestBodyTypeDef: TypeDef,
  name: string,
): FieldDef | undefined {
  switch (requestBodyTypeDef.Type.toString()) {
    case "class":
      for (const field of requestBodyTypeDef.Fields) {
        if (originalFieldName(field) === name) {
          return field;
        }
      }
      break;
    case "union":
      for (const assocType of requestBodyTypeDef.AssociatedTypes) {
        const field = findBodyFieldByName(assocType, name);
        if (field != null) {
          return field;
        }
      }
      break;
    default:
      break;
  }
}

function snake(value: string): string {
  return caser().ToSnake(value);
}
registerTemplateFunc("snake", snake);

// safely handles circular references
// useful for diagnostics
function safeStringify(obj: Object, indent: number = 2) {
  let cache = [];
  const retVal = JSON.stringify(
    obj,
    (key, value) =>
      typeof value === "object" && value !== null
        ? cache.includes(value)
          ? "cached " + key // Duplicate reference found, discard key
          : cache.push(value) && value // Store value in our collection
        : value,
    indent,
  );
  return retVal;
}

// Like JSON.stringify but with a max depth and max elements per object/array, as well as excluding specified keys
// Useful for printing out FieldDefs and TypeDefs in a valid JSON format without overwhelming the output
function maxRecursionStringify(
  obj: any,
  maxDepth: number = 4,
  maxElements: number = 10,
  currentDepth: number = 0,
): string {
  const excludedKeys = new Set<string>([
    "ContextStack",
    "Annotations",
    "Location",
    "Comments",
    "Examples",
  ]);
  if (currentDepth >= maxDepth) {
    return typeof obj === "object" && obj !== null
      ? '"[Object too deep]"'
      : String(obj);
  }

  if (obj === null) return "null";
  if (obj === undefined) return "undefined";

  const type = typeof obj;
  if (type !== "object") {
    return JSON.stringify(obj);
  }

  if (Array.isArray(obj)) {
    const elements = obj
      .slice(0, maxElements)
      .map((item) =>
        maxRecursionStringify(item, maxDepth, maxElements, currentDepth + 1),
      );
    const truncated =
      obj.length > maxElements
        ? `, "... ${obj.length - maxElements} more"`
        : "";
    return `[${elements.join(", ")}${truncated}]`;
  }

  const keys = Object.keys(obj).slice(0, maxElements);
  const pairs = keys.map((key) => {
    try {
      if (excludedKeys.has(key)) {
        return `"${key}": "[Excluded]"`;
      }
      const value = maxRecursionStringify(
        obj[key],
        maxDepth,
        maxElements,
        currentDepth + 1,
      );
      if (
        value === null ||
        value === undefined ||
        value[0] === '"' ||
        value[0] === "{" ||
        value[0] === "["
      ) {
        return `"${key}": ${value}`;
      }
      return `"${key}": "${value}"`;
    } catch (e) {
      return `"${key}": "[Error: ${e instanceof Error ? e.message : e}]"`;
    }
  });

  const totalKeys = Object.keys(obj).length;
  const truncated =
    totalKeys > maxElements
      ? `, "...": "${totalKeys - maxElements} more properties"`
      : "";
  return `{${pairs.join(", ")}${truncated}}`;
}

function collectFromCallbackIterator<T, U>(
  iter: CallbackIterator<T, U>,
): [T, U][] {
  const result: [T, U][] = [];
  iter((a, b) => {
    result.push([a, b]);
    return true;
  });
  return result;
}

function sequencedMapEntries<T, U>(map: SequencedMap<T, U>): [T, U][] {
  if (!map) {
    return [];
  }
  return collectFromCallbackIterator(map.All());
}

function trimTrailing(str: string, char: string): string {
  if (str.endsWith(char)) {
    return str.slice(0, -1);
  }
  return str;
}

function singleLine(str: string): string {
  return str.replace(/\n/g, " ").trim();
}

function trailingDot(str: string): string {
  if (!str.endsWith(".")) {
    return str + ".";
  }
  return str;
}

registerTemplateFunc("isDebug", isDebug);

function matchContentType(contentType: string, pattern: string): boolean {
  if (pattern === "*" || pattern === "*/*") {
    return true;
  }

  const idx = contentType.split(";").findIndex((raw) => {
    const ctype = raw.trim();
    if (ctype === pattern) {
      return true;
    }

    const parts = ctype.split("/");
    if (parts.length !== 2) {
      return false;
    }

    return `${parts[0]}/*` === pattern || `*/${parts[1]}` === pattern;
  });

  return idx >= 0;
}
