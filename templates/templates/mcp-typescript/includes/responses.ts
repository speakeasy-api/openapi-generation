const codeRE = /^\d\d\d$/;

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

function inlineCode(code: string) {
  return codeRE.test(code) ? code : `"${code}"`;
}

function templateResponseClauseStrict(
  methodState: OperationStateTS,
  codes: string,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const { operation: op, usageLocation } = methodState;
  const sm = content.SerializationMethod;
  const resultField = subResponse.Error ? null : getResultField(op);
  const key = resultField
    ? sanitizeFieldName(originalFieldName(resultField))
    : sanitizeFieldName(originalFieldName(content.Content));
  const e = subResponse.Error ? "Err" : "";
  const args: string[] = [
    contentTypeArg(sm, content.ContentType),
    subResponse.Headers ? "hdrs: true" : "",
    key ? `key: "${key}"` : "",
    content.SSESentinel
      ? `sseSentinel: "${content.SSESentinel.replace(/"/g, '"')}"`
      : "",
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

    errTypes.importZodTypes();
    subtypes = errTypes;
  }

  subtypes.importZodTypes();
  let schema = subtypes.zod;
  if (schema.endsWith(".inboundSchema")) {
    schema = schema.slice(0, -".inboundSchema".length);
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

function templateResponseClauseLax(
  methodState: OperationStateTS,
  codes: string,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  if (subResponse.Error) {
    return `  M.fail(${codes}),`;
  }

  const sm = content.SerializationMethod;
  const key = sanitizeFieldName(originalFieldName(content.Content));
  const e = subResponse.Error ? "Err" : "";
  const args: string[] = [
    contentTypeArg(sm, content.ContentType),
    subResponse.Headers ? "hdrs: true" : "",
    key ? `key: "${key}"` : "",
    content.SSESentinel
      ? `sseSentinel: "${content.SSESentinel.replace(/"/g, '"')}"`
      : "",
  ].filter(Boolean);
  const argstr = args.length ? `{${args.join(",")}}` : "";

  addImport("zod", "z", "aliasImport");
  const textMatch = `  M.text${e}(${codes}, z.unknown(), ${argstr}),`;

  switch (sm) {
    case "json":
      return `  M.json${e}(${codes}, z.unknown(), ${argstr}),`;
    case "form":
    case "string":
    case "eventstream":
    case "jsonl":
      return textMatch;
    case "raw":
      const datatype = content.Content.Type.Type.toString();
      return datatype === "response-stream"
        ? `  M.stream${e}(${codes}, z.unknown(), ${argstr}),`
        : `  M.bytes${e}(${codes}, z.unknown(), ${argstr}),`;
    case "multipart":
      throw new Error("multipart responses are not supported");
    default:
      sm satisfies never;
      throw new Error(`Unsupported response serialization method: ${sm}`);
  }
}

function templateResponseClause(
  methodState: OperationStateTS,
  codes: string,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  if (isStrictMCPServer()) {
    return templateResponseClauseStrict(
      methodState,
      codes,
      subResponse,
      content,
    );
  } else {
    return templateResponseClauseLax(methodState, codes, subResponse, content);
  }
}

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

  return `  M.nil(${codes}, ${schema}, ${args}),`;
}

function templateResponseMatcher(state: OperationStateTS): string {
  const { operation: op, valueType, errorType } = state;

  // inboundResponse.importOutputTypes();

  const subResponses = op.Response?.Responses || [];

  addInternalImport("matchers", "M", "", aliasImport);
  let chain = `const [result$] = await M.match<${valueType}, ${errorType}>(`;
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
  chain += ")(response, req$, { extraFields: responseFields$ })";

  return chain;
}

registerTemplateFunc("templateResponseMatcher", templateResponseMatcher);

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
