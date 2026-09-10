function templateResponseMatcher(state: OperationStateTS): string {
  const {
    operation: op,
    inboundResponse,
    responseFormat,
    flatOptions,
    valueType,
    errorType,
  } = state;

  inboundResponse.importOutputTypes();

  const subResponses = op.Response?.Responses || [];

  let assignment = "const [" + funcLocal("result") + "]";
  if (op.Extensions.Pagination) {
    assignment =
      "const [" + funcLocal("result") + ", " + funcLocal("raw") + "]";
  }

  addInternalImport("matchers", "M", "", aliasImport);
  let chain = `${assignment} = await M.match<${valueType}, ${errorType}>(`;
  for (let subRes of subResponses) {
    const codes = templateStatusCode(subRes.Code);

    if (!subRes.Content.length) {
      const emptyClause = templateResponseEmptyClause(state, codes, subRes);
      chain += emptyClause ? `\n${emptyClause}` : "";
    }

    for (const content of subRes.Content) {
      if (!content.Content && subRes.Error && !subRes.Headers) {
        continue;
      }

      const clause = templateResponseClause(state, codes, subRes, content);
      chain += clause ? `\n${clause}` : "";
    }
  }
  chain += ")";

  let matchargs = "response, " + funcLocal("req") + ",";
  if (responseFormat !== "flat" || flatOptions.needsEnvelope) {
    matchargs += "{ extraFields: " + funcLocal("responseFields") + " }, ";
  }

  chain += `(${matchargs})`;

  return chain;
}
registerTemplateFunc("templateResponseMatcher", templateResponseMatcher);

function templateResponseEmptyClause(
  methodState: OperationStateTS,
  codes: string,
  subResponse: SubResponse,
): string {
  if (subResponse.Error) {
    return `  M.fail(${codes}),`;
  }

  let args = "";
  if (subResponse.Headers) {
    args = "{ hdrs: true }";
  }

  methodState.inboundResponse.importZodTypes();
  let schema = methodState.inboundResponse.zod;
  if (schema.endsWith(".inboundSchema")) {
    schema = schema.slice(0, -".inboundSchema".length);
  }

  // If schema contains z. calls, ensure zod is imported
  if (schema.includes("z.")) {
    addZodPackageImport();
  }

  if (isNoZod()) {
    const tsType = stripOutbound(methodState.inboundResponse.outputType);
    return `  M.nil<${tsType}>(${codes}${args ? `, ${args}` : ""}),`;
  }

  return `  M.nil(${codes}, ${schema}, ${args}),`;
}

function templateResponseClause(
  methodState: OperationStateTS,
  codes: string,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const { operation: op, responseFormat, usageLocation } = methodState;
  const hasKeyedResult =
    responseFormat !== "flat" || op.Response.Type?.ResponseEnvelope;
  const resultField = getResultField(op);

  const sm = content.SerializationMethod;

  let key: string | null = null;
  switch (true) {
    case subResponse.Error:
      break;
    case hasKeyedResult && responseFormat === "flat":
      key = originalFieldName(resultField);
      break;
    case hasKeyedResult:
      key = originalFieldName(content.Content);
      break;
  }

  const e = subResponse.Error ? "Err" : "";
  const args: string[] = [
    contentTypeArg(sm, content.ContentType),
    subResponse.Headers ? "hdrs: true" : "",
    key ? `key: "${key}"` : "",
  ].filter(Boolean);
  const argstr = args.length ? `{${args.join(",")}}` : "";

  let subtypes = methodState.inboundResponse;
  if (subResponse.Error) {
    const errTypes = resolveInbound({
      usageLocation,
      typeDef: content.Content.Type,
      rootTypeDef: op.OwningSDK.Type,
      optional: false,
      nullable: false,
    });

    if (isNoZod()) {
      // The matcher receives the class constructor; pull in its import.
      errTypes.importOutputTypes();
    } else {
      errTypes.importZodTypes();
    }
    subtypes = errTypes;
  }

  // In no-zod mode the matcher gets a TS generic, not a schema value — so we
  // skip the importZodTypes calls that would otherwise pull SSE helpers like
  // tryParseJson into the funcs file unused.
  if (!isNoZod()) {
    subtypes.importZodTypes();
  }
  let schema = subtypes.zod;
  if (schema.endsWith(".inboundSchema")) {
    schema = schema.slice(0, -".inboundSchema".length);
  }

  // If schema contains z. calls, ensure zod is imported
  if (schema.includes("z.")) {
    addZodPackageImport();
  }

  // In no-zod mode there is no per-type schema. Errors dispatch through the
  // class constructor (second arg) when the spec defines a single error
  // class; for discriminated-union error responses (a type alias rather
  // than a class) there is no constructor — the matcher just passes the
  // parsed payload through as the error value. Event-streams pass sentinel/
  // flattened/dataRequired options that the matcher uses to wrap the raw
  // byte stream.
  if (isNoZod()) {
    const tsType = stripOutbound(subtypes.outputType);
    // Pass the constructor only when the error response is a single error
    // class (`typeDef.Type === "error"`). For union/object/etc. there is no
    // instantiable class to call `new` on; emit `undefined` so the matcher
    // skips the construction path.
    const isErrorClass =
      subResponse.Error && content.Content.Type.Type.toString() === "error";
    const errArg = subResponse.Error
      ? isErrorClass
        ? `, ${tsType}`
        : `, undefined`
      : "";

    let sseArg = "";
    if (sm === "eventstream") {
      const typeDef = content.Content.Type;
      const sentinel: string = typeDef.EventStreamSentinel || "";
      const flattened = !!getFlattenedEventStreamField(typeDef.ItemType);
      const dataRequired = isSSEDataRequired(typeDef.ItemType);
      const parts: string[] = [];
      if (sentinel) parts.push(`sentinel: ${JSON.stringify(sentinel)}`);
      if (flattened) parts.push(`flattened: true`);
      if (!dataRequired) parts.push(`dataRequired: false`);
      // Always emit the sseArg slot for `M.sse` / `M.sseErr` so the positional
      // argument ordering stays stable; an empty `{}` is a no-op.
      sseArg = `, {${parts.join(", ")}}`;
    }
    const tailArgs = argstr ? `, ${argstr}` : "";

    switch (sm) {
      case "json":
        return `  M.json${e}<${tsType}>(${codes}${errArg}${tailArgs}),`;
      case "form":
        return `  M.form${e}<${tsType}>(${codes}${errArg}${tailArgs}),`;
      case "string":
        return `  M.text${e}<${tsType}>(${codes}${errArg}${tailArgs}),`;
      case "eventstream":
        return `  M.sse${e}<${tsType}>(${codes}${errArg}${sseArg}${tailArgs}),`;
      case "jsonl":
        return `  M.jsonl${e}<${tsType}>(${codes}${errArg}${tailArgs}),`;
      case "raw":
        const ndatatype = content.Content.Type.Type.toString();
        return ndatatype === "response-stream"
          ? `  M.stream${e}<${tsType}>(${codes}${errArg}${tailArgs}),`
          : `  M.bytes${e}<${tsType}>(${codes}${errArg}${tailArgs}),`;
      case "multipart":
        throw new Error("multipart responses are not supported");
      default:
        sm satisfies never;
        throw new Error(`Unsupported response serialization method: ${sm}`);
    }
  }

  switch (sm) {
    case "json":
      return `  M.json${e}(${codes}, ${schema}, ${argstr}),`;
    case "form":
      return `  M.form${e}(${codes}, ${schema}, ${argstr}),`;
    case "string":
      return `  M.text${e}(${codes}, ${schema}, ${argstr}),`;
    case "eventstream":
      return `  M.sse${e}(${codes}, ${schema}, ${argstr}),`;
    case "jsonl":
      return `  M.jsonl${e}(${codes}, ${schema}, ${argstr}),`;
    case "raw":
      const datatype = content.Content.Type.Type.toString();
      return datatype === "response-stream"
        ? `  M.stream${e}(${codes}, ${schema}, ${argstr}),`
        : `  M.bytes${e}(${codes}, ${schema}, ${argstr}),`;
    case "multipart":
      throw new Error("multipart responses are not supported");
    default:
      sm satisfies never;
      throw new Error(`Unsupported response serialization method: ${sm}`);
  }
}

function stripOutbound(s: string): string {
  return s.replace(/\$Outbound\b/g, "");
}

function contentTypeArg(serializationMethod: string, ctype: string): string {
  const sm = serializationMethod;
  switch (true) {
    case sm === "json" && ctype === "application/json":
    case sm === "form" && ctype === "application/x-www-form-urlencoded":
    case sm === "string" && ctype === "text/plain":
    case sm === "eventstream" && ctype === "text/event-stream":
    case sm === "raw" && ctype === "application/octet-stream":
      return "";
  }

  return `ctype: "${ctype}"`;
}
