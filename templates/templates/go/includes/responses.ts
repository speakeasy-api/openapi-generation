// @ts-ignore
function isSkipDeserializationEnabled(): boolean {
  return context.Global.Config.EnableSkipDeserialization === true;
}
registerTemplateFunc(
  "isSkipDeserializationEnabled",
  isSkipDeserializationEnabled,
);

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
      return templateRawHandler(op, subResponse, content);
    case "string":
      return templateStringHandler(op, subResponse, content);
    case "eventstream":
      return templateEventStreamHandler(op, subResponse, content);
    case "jsonl":
      return templateJSONLHandler(op, subResponse, content);
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
  const responseFormat = getResponseFormat();

  addImport("bytes");
  addImport("utils");
  const bodyAccess = `bytes.NewBuffer(rawBody)`;
  const getRawBody = `rawBody, err := utils.ConsumeRawBody(httpRes)
  if err != nil {
      ${templateErrorReturn(op, "err")}
  }
      
  `;

  let handler = `${getRawBody}var out ${sanitizeType(
    content.Content.Type,
    false,
    "",
  )}
if err := utils.UnmarshalJsonFromResponseBody(${bodyAccess}, &out, ${templateAnnotations(
    content.Content,
    true,
  )}); err != nil {
    ${templateErrorReturn(op, "err")}
}
`;

  if (subResponse.Error) {
    if (hasRawResponseField(content.Content.Type)) {
      // "flat" is used by the CLI target; other targets use "envelope".
      // Both formats include a RawResponse field on error types.
      if (responseFormat == "envelope" || responseFormat == "flat") {
        handler += `\nout.RawResponse = httpRes`;
      } else {
        handler += `\nout.HTTPMeta = ${getAccessNamespace(
          "",
          "shared",
        )}HTTPMetadata{
    Request: req,
    Response: httpRes,
  }`;
      }
    }

    handler += `\n${templateErrorReturn(op, "&out")}`;
    return handler;
  }

  if (responseFormat != "flat") {
    // Check if the field is OptionalNullable type
    const isOptionalNullable =
      content.Content.Nullable &&
      content.Content.Optional &&
      context.Global.Config.NullableOptionalWrapper;

    if (isOptionalNullable) {
      handler += `\nres.${sanitizeFieldName(
        content.Content.Name,
      )} = ${templateFieldValueInit(content.Content, "out", false)}`;
    } else {
      handler += `\nres.${sanitizeFieldName(content.Content.Name)} = ${
        isPointerType(content.Content.Type) ? "" : "&"
      }out`;
    }

    // When enableSkipDeserialization is set: wrap deserialization + assignment
    // in the skip-deser conditional. With SkipDeserialization set at runtime
    // the body stays unread on httpRes (the caller reads it via
    // HTTPMeta.Response) and typed fields remain nil. The unread body takes
    // ownership of the per-operation timeout cancel, exactly like a stream
    // (see operationOwnsTimeoutCancel): the deadline still bounds the
    // caller's read, and is released on Close or once the read finishes,
    // instead of firing when this method returns and severing the body.
    if (isSkipDeserializationEnabled()) {
      const cancelRef = streamCancelRef(op);
      handler = `if o.SkipDeserialization != nil && *o.SkipDeserialization {
    httpRes.Body = utils.BodyWithCancel(httpRes.Body, ${cancelRef})
    ${cancelRef} = nil
} else {
${handler}
}`;
    }

    return handler;
  }

  const resultField = getResultField(op);

  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    if (op.Response.Type == content.Content.Type) {
      handler += `\n return ${
        isPointerType(content.Content.Type) ? "" : "&"
      }out, nil`;
      return handler;
    }

    if (
      op.Extensions.Pagination &&
      content.Content.Type.Type.toString() == "union"
    ) {
      handler += `\nresult := out;`;
    } else {
      handler += `\nresult := ${getResponseUnionName(
        resultField?.Type ?? op.Response.Type,
        content,
      )}(out)`;
    }

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = result`;
    } else {
      handler += `\nreturn &result, nil`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      const needsAddr =
        resultField &&
        handleAsPointer(resultField) &&
        !isPointerType(content.Content.Type);
      handler += `\nres.${sanitizeFieldName(resultField?.Name)} = ${
        needsAddr ? "&" : ""
      }out`;
    } else {
      handler += `\nreturn ${
        isPointerType(content.Content.Type) ? "" : "&"
      }out, nil`;
    }
  }

  return handler;
}

function templateRawHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  if (content.Content.Type.Type.toString() == "response-stream") {
    return templateStreamHandler(op, subResponse, content);
  }

  const responseFormat = getResponseFormat();

  let handler = "";

  addImport("utils");
  handler += `rawBody, err := utils.ConsumeRawBody(httpRes)
if err != nil {
    ${templateErrorReturn(op, "err")}
}
`;

  if (responseFormat != "flat") {
    handler += `\nres.${sanitizeFieldName(content.Content.Name)} = rawBody`;
    return handler;
  }

  const resultField = getResultField(op);

  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    handler += `\nresult := ${getResponseUnionName(
      resultField?.Type ?? op.Response.Type,
      content,
    )}(rawBody)`;

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = result`;
    } else {
      handler += `\nreturn &result, nil`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = rawBody`;
    } else {
      handler += `\nreturn rawBody, nil`;
    }
  }

  return handler;
}

function templateStreamHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();

  addImport("utils");
  const cancelRef = streamCancelRef(op);
  let handler = `httpRes.Body = utils.BodyWithCancel(httpRes.Body, ${cancelRef})
${cancelRef} = nil
`;

  if (responseFormat != "flat" || op.Response.Type.ResponseEnvelope) {
    handler += `res.${sanitizeFieldName(content.Content.Name)} = httpRes.Body

    return res, nil`;
  } else {
    handler += `return httpRes.Body, nil`;
  }

  return handler;
}

function templateStringHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();

  let handler = "";

  addImport("utils");
  handler += `rawBody, err := utils.ConsumeRawBody(httpRes)
if err != nil {
    ${templateErrorReturn(op, "err")}
}
`;

  if (responseFormat != "flat") {
    handler += `\nout := string(rawBody)
res.${sanitizeFieldName(content.Content.Name)} = &out`;
    return handler;
  }

  const resultField = getResultField(op);

  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    handler += `\nresult := ${getResponseUnionName(
      resultField?.Type ?? op.Response.Type,
      content,
    )}(string(rawBody))`;

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = result`;
    } else {
      handler += `\nreturn &result, nil`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = string(rawBody)`;
    } else {
      handler += `\nres := string(rawBody)\nreturn &res, nil`;
    }
  }

  return handler;
}

// operationHasStreamingResponse reports whether any response of op is an
// SSE event stream, a JSONL stream, or a raw response stream. Such methods
// hand the timeout's cancel to the returned stream instead of deferring it
// (see method.go.stmpl).
// @ts-ignore
function operationHasStreamingResponse(op: Operation): boolean {
  return op.Response.Responses.some((r) =>
    r.Content.some(
      (c) =>
        c.SerializationMethod === "eventstream" ||
        c.SerializationMethod === "jsonl" ||
        c.Content?.Type?.Type?.toString() === "response-stream",
    ),
  );
}
registerTemplateFunc(
  "operationHasStreamingResponse",
  operationHasStreamingResponse,
);

// operationSkipsDeserialization reports whether op can return its JSON body
// unread when the caller sets SkipDeserialization: the option is generated
// for successful JSON responses of envelope-style methods only.
function operationSkipsDeserialization(op: Operation): boolean {
  return (
    isSkipDeserializationEnabled() &&
    getResponseFormat() != "flat" &&
    op.Response.Responses.some(
      (r) =>
        !r.Error && r.Content.some((c) => c.SerializationMethod === "json"),
    )
  );
}

// operationOwnsTimeoutCancel reports whether op may hand the per-operation
// timeout cancel to a body that outlives the method call: a streaming
// response, or a JSON body left unread by SkipDeserialization. Such methods
// keep the cancel in a slot the response handler clears on hand-off instead
// of deferring it unconditionally (see method.go.stmpl).
// @ts-ignore
function operationOwnsTimeoutCancel(op: Operation): boolean {
  return operationHasStreamingResponse(op) || operationSkipsDeserialization(op);
}
registerTemplateFunc("operationOwnsTimeoutCancel", operationOwnsTimeoutCancel);

// streamCancelRef is the Go expression naming the timeout cancel a streaming
// response handler takes ownership of. Polling operations run their response
// handling in a private method that receives a pointer to the public
// method's cancel (see method.go.stmpl); everything else uses the local.
function streamCancelRef(op: Operation): string {
  return op.Extensions?.Polling ? "*streamCancel" : "streamCancel";
}

// @ts-ignore
function templateEventStreamHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  addImport("stream");
  addImport("bytes");
  const responseFormat = getResponseFormat();
  const sentinel = (content.SSESentinel || "").replaceAll('"', '\\"');

  const sseDataRequired = isSSEDataRequired(content.Content.Type.ItemType);
  const dataRequiredOpt = sseDataRequired
    ? ""
    : `, stream.WithDataRequired[${sanitizeType(
        content.Content.Type.ItemType,
        false,
        "",
      )}](false)`;

  let handler = `out := stream.NewEventStream(ctx, httpRes.Body, func(se []byte) (${sanitizeType(
    content.Content.Type.ItemType,
    false,
    "",
  )}, error) {
    var e ${sanitizeType(content.Content.Type.ItemType, false, "")}
    if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(se), &e, ${templateAnnotations(
      content.Content,
      true,
    )}); err != nil {
        return ${sanitizeType(content.Content.Type.ItemType, false, "")}{}, err
    }
    return e, nil
}, "${sentinel}"${dataRequiredOpt}, stream.WithCancel[${sanitizeType(
    content.Content.Type.ItemType,
    false,
    "",
  )}](${streamCancelRef(op)}))
${streamCancelRef(op)} = nil`;

  if (responseFormat != "flat") {
    handler += `\nres.${sanitizeFieldName(content.Content.Name)} = out`;
    return handler;
  }

  const resultField = getResultField(op);

  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    handler += `\nresult := ${getResponseUnionName(
      resultField?.Type ?? op.Response.Type,
      content,
    )}(out)`;

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = result`;
    } else {
      handler += `\nreturn &result, nil`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      const needsAddr =
        resultField &&
        handleAsPointer(resultField) &&
        !isPointerType(content.Content.Type);
      handler += `\nres.${sanitizeFieldName(resultField?.Name)} = ${
        needsAddr ? "&" : ""
      }out`;
    } else {
      handler += `\nreturn out, nil`;
    }
  }

  return handler;
}

// @ts-ignore
function templateJSONLHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  addImport("jsonl");
  addImport("bytes");
  const responseFormat = getResponseFormat();

  let handler = `out := jsonl.NewJsonLStream(httpRes.Body, func(se []byte) (${sanitizeType(
    content.Content.Type.ItemType,
    false,
    "",
  )}, error) {
    var e ${sanitizeType(content.Content.Type.ItemType, false, "")}
    if err := utils.UnmarshalJsonFromResponseBody(bytes.NewBuffer(se), &e, ${templateAnnotations(
      content.Content,
      true,
    )}); err != nil {
        return ${sanitizeType(content.Content.Type.ItemType, false, "")}{}, err
    }
    return e, nil
}, jsonl.WithCancel[${sanitizeType(
    content.Content.Type.ItemType,
    false,
    "",
  )}](${streamCancelRef(op)}))
${streamCancelRef(op)} = nil`;

  if (responseFormat != "flat") {
    handler += `\nres.${sanitizeFieldName(content.Content.Name)} = out`;
    return handler;
  }

  const resultField = getResultField(op);

  if (op.Response.Type.Type == "union" || resultField?.Type.Type == "union") {
    handler += `\nresult := ${getResponseUnionName(
      resultField?.Type ?? op.Response.Type,
      content,
    )}(out)`;

    if (op.Response.Type.ResponseEnvelope) {
      handler += `\nres.${sanitizeFieldName(
        resultField?.Name ?? content.Content.Name,
      )} = result`;
    } else {
      handler += `\nreturn &result, nil`;
    }
  } else {
    if (op.Response.Type.ResponseEnvelope) {
      const needsAddr =
        resultField &&
        handleAsPointer(resultField) &&
        !isPointerType(content.Content.Type);
      handler += `\nres.${sanitizeFieldName(resultField?.Name)} = ${
        needsAddr ? "&" : ""
      }out`;
    } else {
      handler += `\nreturn out, nil`;
    }
  }

  return handler;
}

function getResponseUnionName(
  unionType: TypeDef,
  content: ResponseBodyContent,
): string {
  const suffix = isGenericUnion(unionType)
    ? ""
    : sanitizeClassName(sanitizeUnionTypeName(content.Content.Type));
  return `${getAccessNamespace(
    "",
    unionType.Scope,
  )}${unionConstructorPrefix()}${sanitizeClassName(unionType.Name)}${suffix}`;
}
