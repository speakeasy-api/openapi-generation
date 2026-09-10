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
    const owningClassName = sanitizeClassName(op.OwningSDK.Type.Name);
    const responseType = sanitizeType(
      op.Response.Type,
      false,
      "",
      owningClassName,
    );
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
registerTemplateFunc("templateResponseHandler", templateResponseHandler);

// @ts-ignore
function templateErrorUnion(
  typeDef: TypeDef,
  responseFormat: string,
  output: string,
): string {
  const hasRawField = hasRawResponseField(typeDef);

  if (hasRawField && responseFormat == "envelope-http") {
    output += `
var httpMeta = new ${getHTTPMetadata()}()
{
    Response = httpResponse,
    Request = httpRequest
};`;
  }

  function throwException(typeDef: TypeDef, name: string) {
    const fieldName = sanitizeFieldName(sanitizeUnionTypeName(typeDef));
    const payload = `${sanitizePrivateFieldName(name)}Payload`;
    const httpMetaArgs =
      responseFormat == "envelope-http"
        ? "httpRequest, httpResponse, httpResponseBody"
        : "httpResponse, httpResponseBody";

    const message = `"Failed to deserialize error payload."`;
    const innerException = `new NullReferenceException("${fieldName}Payload is null.")`;
    const exception = schemaValidationLenient()
      ? templateDefaultError(false)
      : "ResponseValidationException";
    const nullPayloadThrow = `throw new ${exception}(${message}, ${httpMetaArgs}, ${innerException});`;

    let output = `
        var ${payload} = obj.${fieldName}Payload;
        if (${payload} == null)
        {
            ${nullPayloadThrow}
        }`;

    if (hasRawField) {
      switch (responseFormat) {
        case "envelope":
          output += `
        ${payload}!.RawResponse = httpResponse;`;
          break;
        case "envelope-http":
          output += `
        ${payload}!.HttpMeta = httpMeta;`;
          break;
      }
    }

    if (schemaValidationLenient()) {
      // The union converter kept a known-discriminator variant despite a strict
      // validation failure; surface that failure on the typed error's
      // DeserializationException to honor the lenient error contract.
      output += `
        ResponseValidationException? ${payload}DeserException = obj.ValidationFailure != null
            ? new ResponseValidationException("Failed to deserialize error payload into ${fieldName}.", ${httpMetaArgs}, obj.ValidationFailure)
            : null;
        throw new ${fieldName}(${payload}, ${httpMetaArgs}, ${payload}DeserException);`;
    } else {
      output += `
        throw new ${fieldName}(${payload}, ${httpMetaArgs});`;
    }

    return output;
  }

  output += `
switch (obj.Type.ToString()) {`;
  if (typeDef.Discriminator) {
    typeDef.Discriminator.Mapping.forEach((map) => {
      const enumVal = sanitizeEnumValue(map.Name, dataType("string"));
      output += `
    case ${enumVal}:${throwException(map.Type, map.Name)}`;
    });
  } else {
    typeDef.AssociatedTypes.forEach((typ) => {
      const name = sanitizeUnionTypeName(typ);
      const enumVal = sanitizeEnumValue(name, dataType("string"));
      output += `
    case ${enumVal}:${throwException(typ, name)}`;
    });
  }

  output += `
    default:
        throw new InvalidOperationException("Unknown error type.");
};`;

  return output;
}

// @ts-ignore
function templateJSONHandler(
  op: Operation,
  subResponse: SubResponse,
  nullValueHandling: string,
  content?: ResponseBodyContent,
): string {
  const pagination = op.Extensions.Pagination ? true : false;
  const responseFormat = getResponseFormat();
  const owningClassName = sanitizeClassName(op.OwningSDK.Type.Name);
  const responseType = sanitizeType(
    op.Response.Type,
    false, // response body nullability is not yet supported;
    content.Content.Type.OutputLocation,
    owningClassName,
  );

  const resultField = getResultField(op);
  const resultType = sanitizeType(
    content.Content.Type,
    false,
    content.Content.Type.OutputLocation,
    owningClassName,
  );

  const httpMetaArgs =
    responseFormat == "envelope-http"
      ? "httpRequest, httpResponse, httpResponseBody"
      : "httpResponse, httpResponseBody";

  function templateJSONDeserialization(
    varName: string,
    targetType: string,
    bodyType: "default" | "error" | "error-union" = "default",
  ): string {
    const errorMessage = `Failed to deserialize response body into ${targetType}.`;

    let preamble = "";
    let catchClause: string;
    if (schemaValidationLenient()) {
      // Lenient mode never raises ResponseValidationException; it degrades.
      //
      // "error-union" bodies are already reconstructed per-variant by the
      // union's converter, which keeps a known-discriminator variant typed.
      // The only failure reaching here is a body the union cannot dispatch at
      // all, handled like the non-JSON fall-through below.
      //
      // "default" and "error" bodies are constructed on a best-effort basis:
      // missing required fields are defaulted and wrong-typed fields are skipped.
      //
      // On "error" responses, a deserialization failure also gets stored on the
      // error's DeserializationException so callers can detect the degraded payload.
      //
      // An unsalvageable body (not parseable JSON, a root that cannot map to the target
      // model, or a value-type target) falls through to the SDK default error, with the
      // raw body preserved and the underlying failure kept as its InnerException.
      const defaultError = templateDefaultError(false);
      if (bodyType === "error") {
        preamble = `ResponseValidationException? deserializationException = null;\n`;
        catchClause = `catch (Exception ex)
{
    deserializationException = new ResponseValidationException("${errorMessage}", ${httpMetaArgs}, ex);
    if (!ResponseBodyDeserializer.TryConstructUnvalidated<${targetType}>(httpResponseBody, out ${varName}))
    {
        throw new ${defaultError}("${errorMessage}", ${httpMetaArgs}, ex);
    }
}`;
      } else if (bodyType === "error-union") {
        catchClause = `catch (Exception ex)
{
    throw new ${defaultError}("${errorMessage}", ${httpMetaArgs}, ex);
}`;
      } else {
        catchClause = `catch (Exception ex)
{
    if (!ResponseBodyDeserializer.TryConstructUnvalidated<${targetType}>(httpResponseBody, out ${varName}))
    {
        throw new ${defaultError}("${errorMessage}", ${httpMetaArgs}, ex);
    }
}`;
      }
    } else {
      catchClause = `catch (Exception ex)
{
    throw new ResponseValidationException("${errorMessage}", ${httpMetaArgs}, ex);
}`;
    }

    return `var httpResponseBody = await httpResponse.Content.ReadAsStringAsync();
${targetType} ${varName};
${preamble}try
{
    ${varName} = ResponseBodyDeserializer.DeserializeNotNull<${targetType}>(httpResponseBody, ${nullValueHandling});
}
${catchClause}\n`;
  }

  /* Error Response Start */
  if (subResponse.Error) {
    if (content.Content.Type.Type.toString() == "union") {
      return templateErrorUnion(
        content.Content.Type,
        responseFormat,
        templateJSONDeserialization("obj", resultType, "error-union"),
      );
    }

    let handler = templateJSONDeserialization(
      "payload",
      `${resultType}Payload`,
      "error",
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

    const deserExcArg = schemaValidationLenient()
      ? ", deserializationException"
      : "";
    return (
      handler +
      `\nthrow new ${resultType}(payload, ${httpMetaArgs}${deserExcArg});`
    );
  }
  /* Error Response End */

  let handler = templateJSONDeserialization("obj", resultType);
  if (responseFormat != "flat") {
    handler += `\nvar response = ${templateResponseBuilder(
      subResponse,
      responseFormat,
      pagination,
      responseType,
    )}`;
    if (content) {
      handler += `\nresponse.${sanitizeFieldName(
        content.Content.Name,
        sanitizeClassName(op.Response.Type.Name),
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
      handler += `\nvar response = ${templateResponseBuilder(
        subResponse,
        responseFormat,
        pagination,
        responseType,
        content,
      )}\nreturn response;`;
    } else {
      handler += `\nreturn result!;`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nvar result = obj!;`;
      handler += `\nvar response = ${templateResponseBuilder(
        subResponse,
        responseFormat,
        pagination,
        responseType,
      )}\nreturn response;`;
    } else {
      handler += `\nreturn obj!;`;
    }
  }
  return handler;
}

// @ts-ignore
function templateBodyHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();
  const owningClassName = sanitizeClassName(op.OwningSDK.Type.Name);
  const responseType = sanitizeType(
    op.Response.Type,
    false,
    "",
    owningClassName,
  );

  function bodyAccessor(): string {
    switch (true) {
      case content.SerializationMethod == "raw" &&
        content.Content.Type.Type.toString() == "response-stream":
        return `downloadHandler.Stream`;
      case content.SerializationMethod == "raw" &&
        content.Content.Type.Type.toString() != "response-stream":
        return `await httpResponse.Content.ReadAsByteArrayAsync()`;
      case content.SerializationMethod == "string":
        return `await httpResponse.Content.ReadAsStringAsync()`;
      default:
        return ".content";
    }
  }

  if (op.Response.Type.ResponseEnvelope) {
    const response = templateResponseBuilder(
      subResponse,
      responseFormat,
      false,
      responseType,
    );
    return `var response = ${response}
response.${sanitizeFieldName(content.Content.Name)} = ${bodyAccessor()};
return response;`;
  }

  if (op.Response.Type.Type == "union") {
    return `${getResponseUnionName(
      op.Response.Type,
      content,
    )}(${bodyAccessor()});`;
  }

  return `return ${bodyAccessor()};`;
}

// @ts-ignore
function templateEventStreamHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();
  const owningClassName = sanitizeClassName(op.OwningSDK.Type.Name);
  const responseType = sanitizeType(
    op.Response.Type,
    false,
    content.Content.Type.OutputLocation,
    owningClassName,
  );

  const resultField = getResultField(op);
  const resultType = sanitizeType(
    content.Content.Type.ItemType, // ItemType is mapped to the data field within the event stream
    false,
    content.Content.Type.OutputLocation,
    owningClassName,
  );

  const sentinel = content.SSESentinel ? `"${content.SSESentinel}"` : "null";
  const sseDataRequired = isSSEDataRequired(content.Content.Type.ItemType);
  const dataRequiredArg = sseDataRequired ? "" : ", dataRequired: false";
  const eventStreamInit = `httpResponse.AsServerSentEventStream<${resultType}>(${sentinel}${dataRequiredArg});`;

  let handler = `var eventStream = await ${eventStreamInit};\n`;
  if (responseFormat != "flat") {
    handler += `var response = ${templateResponseBuilder(
      subResponse,
      responseFormat,
      false,
      responseType,
    )}`;
    if (content) {
      handler += `\nresponse.${sanitizeFieldName(
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
      handler += `\nvar response = ${templateResponseBuilder(
        subResponse,
        responseFormat,
        false,
        responseType,
        content,
      )}\nreturn response;`;
    } else {
      handler += `\nreturn result!;`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      handler += `var result = eventStream!;`;
      handler += `\nvar response = ${templateResponseBuilder(
        subResponse,
        responseFormat,
        false,
        responseType,
      )}\nreturn response;`;
      return handler;
    }

    handler += `return eventStream!;`;
  }

  return handler;
}

function templateResponseObject(
  responseType: string,
  params: string[],
): string {
  if (params.length == 0) {
    if (responseType == "byte[]") {
      return "Array.Empty<byte>();";
    }

    return `new ${responseType}() {};`;
  }

  return `new ${responseType}()
{
${params.map((p) => indentString(p, 1)).join(",\n")}
};`;
}

function templateResponseBuilder(
  subResponse: SubResponse,
  responseFormat: string,
  pagination: boolean,
  responseType: string,
  content?: ResponseBodyContent,
): string {
  const params: string[] = [];

  switch (responseFormat) {
    case "envelope":
      params.push("StatusCode = responseStatusCode");
      params.push("ContentType = contentType");
      params.push("RawResponse = httpResponse");
      if (content) {
        params.push(`${content.Content.Name} = obj`);
      }

      break;
    case "envelope-http":
      params.push(`HttpMeta = new ${getHTTPMetadata()}()
{
    Response = httpResponse,
    Request = httpRequest
}`);
      if (content) {
        params.push(`${content.Content.Name} = obj`);
      }
      break;
    case "flat":
      params.push(`Result = result`);
      break;
  }

  if (pagination) {
    params.push(`Next = nextFunc`);
  }

  if (subResponse.Headers) {
    params.push(`Headers = Utilities.CollectHeaders(httpResponse.Headers)`);
  }

  return templateResponseObject(responseType, params);
}

function templateEmptyResponseBuilder(
  op: Operation,
  responseFormat: string,
  pagination: boolean,
  responseType: string,
): string {
  const params = [];
  switch (responseFormat) {
    case "envelope":
      params.push("StatusCode = responseStatusCode");
      params.push("ContentType = contentType");
      params.push("RawResponse = httpResponse");
      break;
    case "envelope-http":
      params.push(`HttpMeta = new ${getHTTPMetadata()}()
{
    Response = httpResponse,
    Request = httpRequest
}`);
      break;
    case "flat":
      const resultField = getResultField(op);
      if (
        op.Response.Type.Type == "union" ||
        resultField?.Type.Type == "union"
      ) {
        const unionString = `${getResponseNullUnionName(
          resultField?.Type ?? op.Response.Type,
        )}`;
        if (!resultField) {
          return `${unionString}`;
        }
        params.push(`Result = ${unionString}`);
      } else if (resultField) {
        params.push(`Result = null`);
      }
      break;
  }

  if (pagination) {
    params.push(`Next = nextFunc`);
  }

  if (op.Response.Responses.some((r) => r.Headers)) {
    params.push(`Headers = Utilities.CollectHeaders(httpResponse.Headers)`);
  }

  return templateResponseObject(responseType, params);
}

//@ts-ignore
function getHTTPMetadata(): string {
  const useFull = hasNamespaceConflict();
  const access = getScopeNamespace("shared", useFull);
  return access ? `${access}.HTTPMetadata` : "HTTPMetadata";
}

registerTemplateFunc("getHTTPMetadata", getHTTPMetadata);
