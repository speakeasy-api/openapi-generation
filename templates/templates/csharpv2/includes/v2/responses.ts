// @ts-ignore
function templateResponseHandler(
  op: Operation,
  nullValueHandling: string,
  subResponse?: SubResponse,
  content?: ResponseBodyContent,
): string {
  if (!op.Response.Type) {
    return "\nreturn;";
  }

  if (content == null || subResponse == null) {
    const responseFormat = getResponseFormat();
    const responseType = sanitizeType(op.Response.Type, false, "");
    const pagination = op.Extensions.Pagination ? true : false;
    return `\nreturn ${templateEmptyResponseBuilder(
      op,
      responseFormat,
      pagination,
      responseType,
    )}`;
  }

  switch (content.SerializationMethod) {
    case "json":
      return templateJSONHandler(op, subResponse, nullValueHandling, content);
    case "raw":
    case "string":
      return templateBodyHandler(op, subResponse, content);
    case "eventstream":
      return templateEventStreamHandler(op, subResponse, content);
    default:
      throw new Error(
        `Unsupported response serialization method: ${content.SerializationMethod}`,
      );
  }
}

// @ts-ignore
function templateJSONHandler(
  op: Operation,
  subResponse: SubResponse,
  _nullValueHandling: string,
  content?: ResponseBodyContent,
): string {
  const pagination = op.Extensions.Pagination ? true : false;
  const responseFormat = getResponseFormat();
  const responseType = sanitizeType(
    op.Response.Type,
    false,
    content.Content.Type.OutputLocation,
  );

  const resultField = getResultField(op);
  const resultType = sanitizeType(
    content.Content.Type,
    false,
    content.Content.Type.OutputLocation,
  );

  const buildResponse = (extra?: ResponseBodyContent) =>
    templateResponseBuilder(
      subResponse,
      responseFormat,
      pagination,
      responseType,
      extra,
    );

  const httpMetaArgs =
    responseFormat == "envelope-http"
      ? "httpRequest, httpResponse, httpResponseBody"
      : "httpResponse, httpResponseBody";

  function templateJSONDeserialization(
    varName: string,
    targetType: string,
    typeDef: TypeDef,
  ): string {
    const errorMessage = `Failed to deserialize response body into ${targetType}.`;
    const typeInfoRef = jsonTypeInfoRef(typeDef);

    return `var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
${targetType} ${varName};
try
{
    ${varName} = JsonSerializer.Deserialize(httpResponseBody, ${typeInfoRef})!;
}
catch (Exception ex)
{
    throw new ResponseValidationException("${errorMessage}", ${httpMetaArgs}, ex);
}\n`;
  }

  if (subResponse.Error) {
    if (content.Content.Type.Type.toString() == "union") {
      return templateErrorUnion(
        content.Content.Type,
        responseFormat,
        templateJSONDeserialization("obj", resultType, content.Content.Type),
      );
    }

    let handler = templateJSONDeserialization(
      "payload",
      `${resultType}Payload`,
      content.Content.Type,
    );
    if (hasRawResponseField(content.Content.Type)) {
      switch (responseFormat) {
        case "envelope":
          handler += `\npayload.RawResponse = httpResponse;`;
          break;
        case "envelope-http":
          handler += `\npayload.HttpMeta = new ${getHTTPMetadata()}()
{
    Response = httpResponse,
    Request = httpRequest
};\n`;
          break;
      }
    }

    return handler + `\nthrow new ${resultType}(payload, ${httpMetaArgs});`;
  }

  let handler = templateJSONDeserialization(
    "obj",
    resultType,
    content.Content.Type,
  );
  if (responseFormat != "flat") {
    handler += `\nvar response = ${buildResponse()}`;
    if (content) {
      handler += `\nresponse.${sanitizeField(
        content.Content,
        op.Response.Type,
      )} = obj;`;
    }
    handler += `\nreturn response;`;
    return handler;
  }
  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    if (op.Response.Type == content.Content.Type) {
      handler += `\nreturn obj!;`;
      return handler;
    }

    if (
      op.Extensions.Pagination &&
      content.Content.Type.Type.toString() == "union"
    ) {
      handler += `\nvar result = obj!;`;
    } else {
      handler += `\nvar result = ${getResponseUnionName(
        resultField?.Type ?? op.Response.Type,
        content,
      )}(obj!);`;
    }

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nvar response = ${buildResponse(content)}\nreturn response;`;
    } else {
      handler += `\nreturn result!;`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nvar result = obj!;`;
      handler += `\nvar response = ${buildResponse()}\nreturn response;`;
    } else {
      handler += `\nreturn obj!;`;
    }
  }
  return handler;
}

unregisterTemplateFunc("templateResponseHandler");
registerTemplateFunc("templateResponseHandler", templateResponseHandler);

// @ts-ignore
function templateEventStreamHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();
  const owningClassName = sanitizeClassName(op.OwningSDK.Type.Name);
  // csharp sanitizeType signature includes containingClass + isUsage at runtime
  const sanitizeTypeRuntime = sanitizeType as unknown as (
    typeDef: TypeDef,
    optional: boolean,
    usageLocation: string,
    containingClass: string,
  ) => string;
  const responseType = sanitizeTypeRuntime(
    op.Response.Type,
    false,
    content.Content.Type.OutputLocation,
    owningClassName,
  );

  const resultField = getResultField(op);
  const resultType = sanitizeTypeRuntime(
    content.Content.Type.ItemType,
    false,
    content.Content.Type.OutputLocation,
    owningClassName,
  );

  const sentinel = content.SSESentinel ? `"${content.SSESentinel}"` : "null";
  const sseDataRequired = isSSEDataRequired(content.Content.Type.ItemType);
  const dataRequiredArg = sseDataRequired ? "" : ", dataRequired: false";
  const typeInfoRef = jsonTypeInfoRef(content.Content.Type.ItemType);
  const eventStreamInit = `httpResponse.AsAsyncEnumerable<${resultType}>(${typeInfoRef}, ${sentinel}${dataRequiredArg});`;

  const buildResponse = (extra?: ResponseBodyContent) =>
    templateResponseBuilder(
      subResponse,
      responseFormat,
      false,
      responseType,
      extra,
    );

  let handler = `var eventStream = ${eventStreamInit}\n`;
  if (responseFormat != "flat") {
    handler += `var response = ${buildResponse()}`;
    if (content) {
      const sanitizeFieldNameRuntime = sanitizeFieldName as (
        name: string,
        className?: string,
      ) => string;
      handler += `\nresponse.${sanitizeFieldNameRuntime(
        content.Content.Name,
        sanitizeClassName(op.Response.Type.Name),
      )} = eventStream;`;
    }
    handler += `\nreturn response;`;
    return handler;
  }

  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    if (op.Response.Type == content.Content.Type) {
      handler += `return eventStream!;`;
      return handler;
    }

    handler += `var result = ${getResponseUnionName(
      resultField?.Type ?? op.Response.Type,
      content,
    )}(eventStream!);`;

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nvar response = ${buildResponse(content)}\nreturn response;`;
    } else {
      handler += `\nreturn result!;`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      handler += `var result = eventStream!;`;
      handler += `\nvar response = ${buildResponse()}\nreturn response;`;
      return handler;
    }

    handler += `return eventStream!;`;
  }

  return handler;
}
