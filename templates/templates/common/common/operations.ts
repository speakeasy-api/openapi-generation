registerTemplateFunc("getOperationModelName", getOperationModelName);

function getResponseFormat() {
  const value = context.Global.Config.ResponseFormat;
  if (!isFeatureUsed("responseFormat") || typeof value !== "string") {
    return "envelope";
  }

  switch (value) {
    case "flat":
    case "envelope-http":
      return value;
  }

  return "envelope";
}
registerTemplateFunc("getResponseFormat", getResponseFormat);

function getRequestBody(
  operation: Operation,
):
  | { Type: "request"; Value: RequestDef }
  | { Type: "request-wrapper"; Value: FieldDef }
  | undefined {
  const requestModel = operation.Request;
  if (requestModel == null) {
    return undefined;
  }

  const box = requestModel.Field.Type.Fields.find((fd) => {
    return !!fd.Annotations?.Has("request");
  });

  if (box != null) {
    return { Type: "request-wrapper", Value: box };
  }

  return requestModel.IsRequestBody
    ? { Type: "request", Value: requestModel }
    : undefined;
}
registerTemplateFunc("getRequestBody", getRequestBody);

function getRequestBodyType(op: Operation): TypeDef {
  const requestModel = getRequestBody(op);
  if (!requestModel) {
    return undefined;
  }

  return requestModel.Type == "request-wrapper"
    ? requestModel.Value.Type
    : requestModel.Value.Field.Type;
}
registerTemplateFunc("getRequestBodyType", getRequestBodyType);

function getArguments(op: Operation): MethodArguments {
  const style = context.Global.Config.MethodArguments;

  switch (style) {
    case "infer-optional-args":
    case "positional-path-params":
    case "positional-path-with-extras":
      return getArgumentsInferOptionalArgs(op);
    case "require-security-and-request":
      return getArgumentsDeprecated(op);
    default:
      throw new Error(`Invalid method signature configuration: ${style}`);
  }
}

function getArgumentsInferOptionalArgs(o: Operation): MethodArguments {
  const flattened = operationParametersFlattened(o);

  let requestFieldName = "";
  let requestFields: FieldDef[] = [];

  if (!o.Request) {
    // No request fields to consider
  } else if (flattened) {
    requestFields = sortMethodParams(
      o.Request.Field.Type.Fields.filter((fieldDef) => !fieldDef.Const),
    );
  } else {
    requestFieldName = o.Request.Field.Name;
    requestFields = [o.Request.Field];
  }

  let securityFields: FieldDef[] = [];
  let securityOptional = true;
  if (o.Security) {
    securityFields = [o.Security];
    securityOptional = o.Security.Optional;
  }

  let mergedFields: FieldDef[] = [];

  if (securityOptional) {
    mergedFields = [...requestFields, ...securityFields];
  } else {
    mergedFields = [...securityFields, ...requestFields];
  }

  return {
    merged: mergedFields,
    request: requestFields,
    requestFieldName: requestFieldName,
    flattenedRequest: flattened,
    security: securityFields,
  };
}

function getArgumentsDeprecated(o: Operation): MethodArguments {
  const flattened = operationParametersFlattened(o);

  let requestFieldName = "";
  let requestFields: FieldDef[] = [];

  if (!o.Request) {
    // No request fields to consider
  } else if (flattened) {
    requestFields = o.Request.Field.Type.Fields.filter(
      (fieldDef) => !fieldDef.Const,
    );
  } else {
    requestFieldName = o.Request.Field.Name;
    requestFields = [o.Request.Field];
  }

  let securityFields: FieldDef[] = [];
  if (o.Security) {
    securityFields = [o.Security];
  }

  let mergedFields: FieldDef[] = [];

  if (flattened) {
    mergedFields = [...securityFields, ...requestFields];
  } else {
    mergedFields = [...requestFields, ...securityFields];
  }

  return {
    merged: sortMethodParams(mergedFields),
    request: requestFields,
    requestFieldName: requestFieldName,
    flattenedRequest: flattened,
    security: securityFields,
  };
}

function removeAnySubSDKsWithNoOperations(root: SDK) {
  const queue: Array<SDK> = [root];

  while (queue.length > 0) {
    const sdk = queue.shift();
    sdk.SubSDKs = sdk.SubSDKs.filter((x) => countNestedOperations(x) > 0);
    queue.push(...sdk.SubSDKs);
  }

  return root;
}

function countNestedOperations(sdk: SDK = context.Global.AST.MainSDK) {
  let sum = sdk.Operations.length;
  for (const subSDK of sdk.SubSDKs) {
    sum += countNestedOperations(subSDK);
  }
  return sum;
}

function* iterateOperations(
  root: SDK = context.Global.AST.MainSDK,
  {
    includeWebhooks = false,
    dedupe = true,
  }: { includeWebhooks?: boolean; dedupe?: boolean } = {},
) {
  const queue: Array<SDK> = [root];
  const visited: Set<SDK> = new Set(queue);
  const seenOperations: Set<string> = new Set();
  while (queue.length > 0) {
    const sdk = queue.shift();
    for (const op of sdk.Operations) {
      if (!includeWebhooks && op.Webhook) {
        continue;
      }
      if (dedupe && seenOperations.has(op.ID)) {
        continue;
      }
      seenOperations.add(op.ID);
      yield op;
    }
    for (const subSDK of sdk.SubSDKs) {
      if (!visited.has(subSDK)) {
        visited.add(subSDK);
        queue.push(subSDK);
      }
    }
  }
}

function sdkHasCustomResponseErrors(sdk: SDK = context.Global.AST.MainSDK) {
  for (const op of iterateOperations(sdk)) {
    if (hasCustomResponseErrors(op)) {
      return true;
    }
  }
  return false;
}
registerTemplateFunc("sdkHasCustomResponseErrors", sdkHasCustomResponseErrors);

function hasCustomResponseErrors(op: Operation): boolean {
  return (
    op.Response?.Responses.findIndex((r) => {
      if (!r.Error) {
        return false;
      }

      return r.Content.findIndex((c) => c.Content?.Type.IsCustomType()) >= 0;
    }) >= 0
  );
}

registerTemplateFunc("hasCustomResponseErrors", hasCustomResponseErrors);

function allOperations(
  root: SDK = context.Global.AST.MainSDK,
  {
    includeWebhooks = false,
    dedupe = true,
  }: { includeWebhooks?: boolean; dedupe?: boolean } = {},
): Operation[] {
  const operations: Operation[] = [];

  for (const op of iterateOperations(root, { includeWebhooks, dedupe })) {
    operations.push(op);
  }

  return operations;
}
registerTemplateFunc("allOperations", allOperations);

// @ts-ignore
function getAllowEmptyValueQueryParamNames(op: Operation): string[] {
  const queryParams = op.Request?.Params?.QueryParams ?? [];
  return queryParams
    .filter((p) => p.AllowEmptyValue)
    .map((p) => {
      const paramAnnotation = p.Field?.Annotations?.Get("param") as
        | ParamAnnotation
        | undefined;
      return paramAnnotation?.Name ?? p.Field?.Name ?? "";
    })
    .filter((name) => name !== "");
}
