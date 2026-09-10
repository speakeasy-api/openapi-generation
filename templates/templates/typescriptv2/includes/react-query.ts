// Eventually, we should not be using this hard mapping and instead we will
// allow API owners to tell us if an operation should be exposed as a query or
// mutation hook.
const defaultQueryHookMethods = new Set(["get", "head", "query"]);
const defaultMutationHookMethods = new Set(["post", "put", "patch", "delete"]);

function reactQueryTupleRoot(): string {
  return JSON.stringify(context.Global.Config.PackageName);
}
registerTemplateFunc("reactQueryTupleRoot", reactQueryTupleRoot);

function reactQueryBodyField(operation: Operation): FieldDef | null {
  if (
    !operation.Extensions.ReactHook?.QueryKey?.IncludeRequestBody ||
    !hasQueryHook(operation)
  ) {
    return null;
  }

  return operation.Request?.RequestBody ?? null;
}

function reactQueryTupleType(
  operation: Operation,
  usageLocation: string,
): {
  /**
   * Inert means that a query/mutation key does not have variable parts to it
   * which come from query and header parameters.
   */
  inert: boolean;
  /**
   * The query/mutation key which contains the static, path and variable
   * components.
   */
  full: string;
  /**
   * The part of the query/mutation key which contains the static and path
   * components.
   */
  prefix: string;
  /**
   * The query/mutation key without the leading startic parts.
   */
  base: string;
  /**
   * Aggregates the import functions for any custom types used in the path
   * components of the query key
   */
  pathImports: () => string;
  /**
   * Aggregates the import functions for any custom types used in the variables
   * component of the query key
   */
  variableImports: () => string;
  baseFields: FieldDef[];
  variableFields: FieldDef[];
  /**
   * True when the request body is part of the query key, as opted into with
   * x-speakeasy-react-hook queryKey.includeRequestBody.
   */
  hasBody: boolean;
} {
  const pkg = reactQueryTupleRoot();
  const sdk = JSON.stringify(operation.OwningSDK.FieldName);
  const op = JSON.stringify(sanitizeMethodName(operation));
  const segments: string[] = operation.OwningSDK.FieldName
    ? [pkg, sdk, op]
    : [pkg, op];
  const prefixLength = segments.length;
  const baseFields: FieldDef[] = [];
  const variableFields: FieldDef[] = [];
  const variableMappings: string[] = [];
  const pathImports: (() => string)[] = [];
  const variableImports: (() => string)[] = [];

  let searchFields = operation.Arguments.ParamFields;
  if (operation.Arguments.Flattening === "none") {
    searchFields = operation.Request?.Field?.Type?.Fields ?? [];
  }

  for (const param of searchFields) {
    const ann = param.Annotations?.Get("param");

    if (!ann) {
      continue;
    }

    switch (ann.ParamType) {
      case "header":
      case "queryParam": {
        const name = sanitizeFieldName(param.Name);
        const outbound = resolveOutbound({
          usageLocation: usageLocation,
          typeDef: param.Type,
          rootTypeDef: operation.OwningSDK.Type,
          optional: param.Optional,
          nullable: param.Nullable,
          streamable: isMultipartFileField(param),
          constValue: param.Const,
          defaultValue: param.Default,
        });

        variableImports.push(outbound.importInputTypes);

        const optionalSigil = param.Optional ? "?" : "";
        variableFields.push(param);
        variableMappings.push(`${name}${optionalSigil}: ${outbound.inputType}`);
        break;
      }
      case "pathParam": {
        const name = sanitizeFieldName(param.Name);
        const outbound = resolveOutbound({
          usageLocation: usageLocation,
          typeDef: param.Type,
          rootTypeDef: operation.OwningSDK.Type,
          optional: param.Optional,
          nullable: param.Nullable,
          streamable: isMultipartFileField(param),
          constValue: param.Const,
          defaultValue: param.Default,
        });

        pathImports.push(outbound.importInputTypes);

        segments.push(`${name}: ${outbound.inputType}`);
        baseFields.push(param);
        break;
      }
      default: {
        const fieldName = originalFieldName(param);
        throw new Error(
          `${operation.ID}: ${fieldName}: Unsupported parameter type: ${ann.ParamType}`,
        );
      }
    }
  }

  if (variableFields.length > 0) {
    segments.push(`parameters: {${variableMappings.join(", ")}}`);
  }

  const bodyField = reactQueryBodyField(operation);
  if (bodyField) {
    const bodyOptional =
      bodyField.Optional || Boolean(operation.Request?.Field?.Optional);
    const outbound = resolveOutbound({
      usageLocation: usageLocation,
      typeDef: bodyField.Type,
      rootTypeDef: operation.OwningSDK.Type,
      optional: bodyOptional,
      nullable: bodyField.Nullable,
      streamable: isMultipartFileField(bodyField),
      constValue: bodyField.Const,
      defaultValue: bodyField.Default,
    });

    variableImports.push(outbound.importInputTypes);

    const optionalSigil = bodyOptional ? "?" : "";
    segments.push(`body${optionalSigil}: ${outbound.inputType}`);
  }

  return {
    inert: segments.slice(prefixLength).length === 0,
    full: `[${segments.join(", ")}]`,
    prefix: `[${segments.slice(0, prefixLength).join(", ")}]`,
    base: `[${segments.slice(prefixLength).join(", ")}]`,
    baseFields,
    variableFields,
    hasBody: Boolean(bodyField),
    pathImports: () => {
      for (const importFunc of pathImports) {
        importFunc();
      }

      return "";
    },
    variableImports: () => {
      for (const importFunc of variableImports) {
        importFunc();
      }

      return "";
    },
  };
}

registerTemplateFunc("reactQueryTupleType", reactQueryTupleType);

function reactQueryFillTupleFromRequest(
  operation: Operation,
  usageLocation: string,
): string {
  const { baseFields, variableFields } = reactQueryTupleType(
    operation,
    usageLocation,
  );

  const requestOptional = operation.Request?.Field?.Optional;
  const bodyField = reactQueryBodyField(operation);

  let values: string[] = [];

  switch (operation.Arguments.Flattening) {
    case "none": {
      values = baseFields.map((field) => {
        return sanitizeAccessor(
          "request",
          sanitizeModelField(field),
          requestOptional,
        );
      });

      const vars = variableFields
        .map((field) => {
          const name = sanitizeFieldName(field.Name);
          return `${name}: ${sanitizeAccessor(
            "request",
            sanitizeModelField(field),
            requestOptional,
          )}`;
        })
        .join(", ");

      if (vars) {
        values.push(`{ ${vars} }`);
      }

      if (bodyField) {
        values.push(
          operation.Request?.IsRequestBody
            ? "request"
            : sanitizeAccessor(
                "request",
                sanitizeModelField(bodyField),
                requestOptional,
              ),
        );
      }

      break;
    }
    case "all":
    case "params": {
      values = baseFields.map((field) => sanitizeFieldName(field.Name));

      const vars = variableFields
        .map((field) => sanitizeFieldName(field.Name))
        .join(", ");

      if (vars) {
        values.push(`{ ${vars} }`);
      }

      if (bodyField) {
        values.push(sanitizeFieldName(bodyField.Name));
      }

      break;
    }
    case "body":
      throw new Error(
        `${operation.ID}: Attempted to access path parameters on an operation that only has a body`,
      );
    default:
      throw new Error(
        `${operation.ID}: Unsupported flattening type: ${operation.Arguments.Flattening}`,
      );
  }

  return values.join(", ");
}
registerTemplateFunc(
  "reactQueryFillTupleFromRequest",
  reactQueryFillTupleFromRequest,
);

function reactQueryOperationArgTypes(
  operation: Operation,
  usageLocation: string,
): string {
  const out = operation.Arguments.Sorted.map((field) => {
    const fieldType = resolveOutbound({
      usageLocation,
      typeDef: field.Type,
      rootTypeDef: operation.OwningSDK.Type,
      optional: field.Optional,
      nullable: field.Nullable,
      streamable: isMultipartFileField(field),
      constValue: field.Const,
      defaultValue: field.Default,
    });

    fieldType.importInputTypes();

    const fieldName = sanitizeFieldName(field.Name);
    const optionalSigil = field.Optional ? "?" : "";
    return `${fieldName}${optionalSigil}: ${fieldType.inputType}`;
  }).join(",\n");

  return out ? `${out},\n` : "";
}
registerTemplateFunc(
  "reactQueryOperationArgTypes",
  reactQueryOperationArgTypes,
);

// @ts-ignore
function isReactQueryEnabled(): boolean {
  return isFeatureUsed("reactQueryHooks");
}
registerTemplateFunc("isReactQueryEnabled", isReactQueryEnabled);

// queryKey.includeRequestBody only applies to query hooks, so opting in
// resolves an inferred hook type to query regardless of the HTTP method.
function hasQueryKeyOptIn(operation: Operation): boolean {
  return Boolean(operation.Extensions.ReactHook?.QueryKey?.IncludeRequestBody);
}

function hasQueryHook(operation: Operation): boolean {
  const type = operation.Extensions.ReactHook?.Type || "infer";
  if (type === "infer") {
    return (
      hasQueryKeyOptIn(operation) ||
      defaultQueryHookMethods.has(operation.Method.toLowerCase())
    );
  }

  return type === "query";
}
registerTemplateFunc("hasQueryHook", hasQueryHook);

function hasMutationHook(operation: Operation): boolean {
  const type = operation.Extensions.ReactHook?.Type || "infer";
  if (type === "infer") {
    return (
      !hasQueryKeyOptIn(operation) &&
      defaultMutationHookMethods.has(operation.Method.toLowerCase())
    );
  }

  return type === "mutation";
}

function hasReactQuery(op: Operation): boolean {
  return (
    !op.Extensions.ReactHook?.Disabled &&
    !op.Webhook &&
    // We do not generate hooks for OPTIONS, TRACE,CONNECT and similar methods
    // so we should make sure that the operation will have a query or mutation
    // hook generated.
    (hasQueryHook(op) || hasMutationHook(op))
  );
}
registerTemplateFunc("hasReactQuery", hasReactQuery);

function selectReactQueryExample(
  variant: "get-request" | "post-request" | "pagination" = "get-request",
): UsageContext {
  let predicate: UsageExampleScope[] = [];
  switch (variant) {
    case "get-request":
    case "post-request":
      predicate = [
        { OpFilter: variant, IsGlobal: true, Feature: "" },
        { OpFilter: "!pagination", IsGlobal: true, Feature: "" },
      ];
      break;
    case "pagination":
      predicate = [
        { OpFilter: "get-request", IsGlobal: true, Feature: "" },
        { OpFilter: "pagination", IsGlobal: true, Feature: "" },
      ];
      break;
  }

  const examples = selectExampleOperations(
    context.Global.AST.MainSDK,
    1,
    predicate,
    true,
  );

  return examples[0];
}

registerTemplateFunc("selectReactQueryExample", selectReactQueryExample);
