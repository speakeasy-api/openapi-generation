// @ts-ignore
function templateResponseHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  switch (content.SerializationMethod) {
    case "json":
      return templateJSONHandler(op, subResponse, content);
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
function templateJSONHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const optional = subResponse.Error ? false : content.Content.Optional;
  const deserializedType = sanitizeType(content.Content.Type, optional, "");

  let handler = `\nvar obj = JsonConvert.DeserializeObject<${deserializedType}>(httpResponse.downloadHandler.text, new JsonSerializerSettings(){ NullValueHandling = NullValueHandling.Ignore, Converters = Utilities.GetDefaultJsonDeserializers() });`;

  if (subResponse.Error) {
    if (hasRawResponseField(content.Content.Type)) {
      handler += `\nobj!.RawResponse = httpResponse;`;
    }

    return handler + `\nthrow obj!;`;
  }

  handler += `\nresponse.${sanitizeFieldName(
    content.Content.Name,
    sanitizeClassName(op.Response.Type.Name, false),
  )} = obj;`;

  return handler;
}

// @ts-ignore
function templateBodyHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  let bodyAccessor = "";

  switch (true) {
    case content.SerializationMethod == "raw" &&
      content.Content.Type.Type.toString() == "response-stream":
      bodyAccessor += `downloadHandler.Stream;`;
      break;
    case content.SerializationMethod == "raw" &&
      content.Content.Type.Type.toString() != "response-stream":
      bodyAccessor += `httpResponse.downloadHandler.data;`;
      break;
    case content.SerializationMethod == "string":
      bodyAccessor += `httpResponse.downloadHandler.text;`;
      break;
    default:
      bodyAccessor += ".content";
      break;
  }

  return `\nresponse.${sanitizeFieldName(
    content.Content.Name,
  )} = ${bodyAccessor}`;
}
