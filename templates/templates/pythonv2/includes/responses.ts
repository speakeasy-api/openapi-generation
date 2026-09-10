function indent(levels: number, content: string): string {
  const spaces = " ".repeat(levels * 4);
  return content.replaceAll("\n", "\n" + spaces);
}

function block(levels: number) {
  const level = " ".repeat(4 * levels);
  return function indentLines(
    template: TemplateStringsArray,
    ...params: unknown[]
  ) {
    let result = "";
    for (let i = 0; i < template.length; i++) {
      result += `${params[i - 1] || ""}${template[i]}`;
    }

    return "\n" + level + result.replace(/\n/g, "\n" + level).trimEnd();
  };
}

function templateResponseMatchers(op: Operation) {
  return templateResponseMatchersAtIndent(op, 2);
}
registerTemplateFunc("templateResponseMatchers", templateResponseMatchers);

function templateResponseMatchersAtIndent(op: Operation, levels: number) {
  addSDKImportForAccessNamespace("errors");
  const defaultError = templateDefaultError(true, true);

  const l0 = block(0);
  const l1 = block(1);
  let code = "";

  if (containsJSONErrorResponse(op.Response.Responses)) {
    addImport("typing", "Any");
    code += l0`response_data_: Any = None\n`;
  }

  for (const sub of op.Response.Responses) {
    const isDefault =
      sub.Code.findIndex((c) => c.toLowerCase() === "default") >= 0;

    // If there's only one code then let's drop the list notation
    let status =
      sub.Code.length === 1
        ? JSON.stringify(sub.Code[0])
        : JSON.stringify(sub.Code);

    status = isDefault ? status.toLowerCase() : status.toUpperCase();

    for (const content of sub.Content) {
      const ctype = JSON.stringify(content.ContentType);
      addUtilsImport();
      code += l0`if utils_.match_response(http_res_, ${status}, ${ctype}):`;

      code += l1`${templateResponseHandler(op, sub, content)}`;
    }

    // If the response has no content, then the loop above will have added no
    // code. We'll handle the case of no content below.
    if (!sub.Content.length) {
      addUtilsImport();
      code += l0`if utils_.match_response(http_res_, ${status}, "*"):`;
      if (sub.Error) {
        code += l1`http_res_text_ = ${asyncOnly`await`} utils_.stream_to_text${asyncOnly`_async`}(http_res_)`;
        code += l1`raise ${defaultError}("API error occurred", http_res_, http_res_text_)`;
      } else {
        const headers = templateResponseHeaders(op, sub);

        let bodyField = "";
        const resultField = getResultField(op);

        if (resultField) {
          // TODO: Handle other field types
          bodyField = `${sanitizeFieldName(resultField?.Name)}=None`;
        }

        if (op.Response.Type) {
          code += l1`return ${templateResponseObject(
            op,
            bodyField,
            "",
            headers,
          )}`;
        } else {
          code += l1`return`;
        }
      }
    }
  }

  code += "\n";
  const isStreaming = containsStreamingResponse(op.Response);
  if (isStreaming) {
    code += l0`http_res_text_ = ${asyncOnly`await`} utils_.stream_to_text${asyncOnly`_async`}(http_res_)`;
    code += l0`raise ${defaultError}("Unexpected response received", http_res_, http_res_text_)`;
  } else {
    code += l0`raise ${defaultError}("Unexpected response received", http_res_)`;
  }

  return indent(levels, `\n${code}`);
}
registerTemplateFunc(
  "templateResponseMatchersAtIndent",
  templateResponseMatchersAtIndent,
);

function containsJSONErrorResponse(responses: SubResponse[]): boolean {
  for (const sub of responses) {
    for (const content of sub.Content) {
      if (content.SerializationMethod === "json" && sub.Error) {
        return true;
      }
    }
  }

  return false;
}

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
    case "jsonl":
      return templateJsonLHandler(op, subResponse, content);
    default:
      throw new Error(
        `Unsupported response serialization method: ${content.SerializationMethod}`,
      );
  }
}
registerTemplateFunc("templateResponseHandler", templateResponseHandler);

// @ts-ignore
function templateResponseHeaders(
  op: Operation,
  subResponse: SubResponse,
): string {
  if (getResponseFormat() !== "envelope-http") {
    if (subResponse.Headers) {
      addUtilsImport();
      return `headers=utils_.get_response_headers(http_res_.headers)`;
    } else if (op.Response.Responses.some((r) => r.Headers)) {
      return `headers={}`;
    }
  }

  return "";
}

// @ts-ignore
function templateJSONHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  if (subResponse.Error) {
    return templateJSONErrorHandler(op, subResponse, content);
  }

  const opContainsStreamingResponse = containsStreamingResponse(op.Response);

  addUtilsImport();
  const responseFormat = getResponseFormat();

  let optional = content.Content.Optional;
  if (subResponse.Error || responseFormat == "flat") {
    optional = false;
  }

  let code = "";
  let responseAccessor = "http_res_.text";
  if (opContainsStreamingResponse) {
    code += `http_res_text_ = ${asyncOnly`await`} utils_.stream_to_text${asyncOnly`_async`}(http_res_)\n`;
    responseAccessor = "http_res_text_";
  }

  const model = sanitizePydanticType(content.Content, {
    optional: optional,
    nullable: false,
  });

  // Add validation error handling
  addSDKImportForAccessNamespace("errors");
  addUnmarshalJsonResponseImport();

  const validateArg = responseSchemaValidationStrict()
    ? ``
    : `, validate=False`;
  const value = `unmarshal_json_response(${model}, http_res_ ${
    opContainsStreamingResponse ? `, ${responseAccessor}` : ``
  }${validateArg})\n`;

  let pagination = "";
  if (op.Extensions.Pagination) {
    pagination = "next=next_func";
  }

  const headers = templateResponseHeaders(op, subResponse);

  switch (true) {
    case responseFormat == "envelope":
      code += `return ${templateResponseObject(
        op,
        `${sanitizeFieldName(content.Content.Name)}=${value}`,
        pagination,
        headers,
      )}`;
      return code;
    case responseFormat == "envelope-http":
      code += `return ${templateResponseObject(
        op,
        `${sanitizeFieldName(content.Content.Name)}=${value}`,
        pagination,
        headers,
      )}`;
      return code;
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      const resultField = getResultField(op);

      code += `return ${templateResponseObject(
        op,
        `${sanitizeFieldName(resultField?.Name)}=${value}`,
        pagination,
        headers,
      )}`;
      return code;
    case responseFormat == "flat":
      code += `return ${value}`;
      return code;
    default:
      throw new Error(`Unexpected JSON response format: ${responseFormat}`);
  }
}

function templateJSONErrorHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  addUtilsImport();
  addSDKImportForAccessNamespace("errors");
  const responseFormat = getResponseFormat();
  const opContainsStreamingResponse = containsStreamingResponse(op.Response);

  let responseBodyType = content.Content.Type;
  let suffix = responseBodyType.Type.toString() === "union" ? "Union" : "Data";

  let preamble = "";
  let responseAccessor = `http_res_.text`;
  if (opContainsStreamingResponse) {
    preamble += `http_res_text_ = ${asyncOnly`await`} utils_.stream_to_text${asyncOnly`_async`}(http_res_)\n`;
    responseAccessor = `http_res_text_`;
  }

  const pydanticType = sanitizePydanticType(
    content.Content,
    { optional: false, nullable: false },
    "",
    null,
    true,
    false,
    suffix,
  );

  addUnmarshalJsonResponseImport();
  let body = `response_data_ = unmarshal_json_response(${pydanticType}, http_res_${
    opContainsStreamingResponse ? `, ${responseAccessor}` : ``
  })\n`;

  if (hasRawResponseField(content.Content.Type)) {
    switch (responseFormat) {
      case "envelope":
        body += `response_data_.raw_response = http_res_\n`;
        break;
      case "envelope-http":
        addSDKImportForAccessNamespace("shared");
        body += `response_data_.http_meta = ${getAccessNamespaceSuffixed(
          "shared",
        )}HTTPMetadata(request=req_, response=http_res_)\n`;
        break;
    }
  }

  const errorType = templateSimpleType(content.Content.Type);
  body += `raise ${errorType}(response_data_, http_res_${
    opContainsStreamingResponse ? `, ${responseAccessor}` : ``
  })`;

  if (!errorSchemaValidationEnabled()) {
    const errorsNs = getAccessNamespaceSuffixed("errors");
    const defaultError = templateDefaultError(true, true);
    body =
      `try:${indent(1, `\n${body}`)}\n` +
      `except ${errorsNs}ResponseValidationError as e:\n` +
      `    raise ${defaultError}("Error response body did not match expected schema", http_res_, ${responseAccessor}) from e`;
  }

  return preamble + body;
}

// @ts-ignore
function templateBodyHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();
  const resultField = getResultField(op);
  const opContainsStreamingResponse = containsStreamingResponse(op.Response);

  let bodyAccessor = "http_res_";

  let code = "";

  switch (true) {
    case content.SerializationMethod == "raw" &&
      content.Content.Type.Type.toString() == "response-stream":
      bodyAccessor = "http_res_";
      break;
    case content.SerializationMethod == "string":
      addUtilsImport();
      code += opContainsStreamingResponse
        ? `http_res_text_ = ${asyncOnly`await`} utils_.stream_to_text${asyncOnly`_async`}(http_res_)\n`
        : "";
      bodyAccessor = opContainsStreamingResponse
        ? "http_res_text_"
        : "http_res_.text";
      break;
    default:
      addUtilsImport();
      code += `http_res_bytes = ${asyncOnly`await`} utils_.stream_to_bytes${asyncOnly`_async`}(http_res_)\n`;
      bodyAccessor = "http_res_bytes";
      break;
  }

  const headers = templateResponseHeaders(op, subResponse);

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      code += `return ${templateResponseObject(
        op,
        `${
          responseFormat == "flat"
            ? sanitizeFieldName(resultField?.Name)
            : sanitizeFieldName(content.Content.Name)
        }=${bodyAccessor}`,
        "",
        headers,
      )}`;
      return code;
    case responseFormat == "flat":
      code += `return ${bodyAccessor}`;
      return code;
    default:
      throw new Error(
        `Unexpected string/raw response format: ${responseFormat}`,
      );
  }
}

// @ts-ignore
function templateEventStreamHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  addEventStreamingImport();
  const responseFormat = getResponseFormat();
  const resultField = getResultField(op);

  let responseBodyType = content.Content.Type;

  const model = sanitizePydanticType(
    typeDefToFieldDef(responseBodyType.ItemType),
    { optional: false, nullable: false },
  );

  addUnmarshalJsonResponseImport();

  let validateArg = "";
  if (!responseSchemaValidationStrict()) {
    // Non-strict modes don't raise when encountering a bad frame: `lenient`
    // constructs best-effort model, `disabled` returns raw decoded dict.
    validateArg = `, validate=False`;
  }
  let unmarshal = `unmarshal_json_response(${model}, http_res_, raw${validateArg})`;

  const dataField = getFlattenedEventStreamField(responseBodyType.ItemType);

  if (dataField) {
    // Flat SSE response: yield the inner `.data` payload, not the wrapper event.

    const castItemModel = (target: string) => {
      addImport("typing", "cast");
      const itemModel = sanitizePydanticType(dataField, {
        optional: false,
        nullable: dataField.Nullable,
      });
      return `cast(${itemModel}, ${target})`;
    };

    if (responseSchemaValidationMode() === "disabled") {
      addImport("utils.unmarshal_json_response", "sse_event_data", true);
      unmarshal = castItemModel(`sse_event_data(${unmarshal})`);
    } else if (dataField.Optional || dataField.Nullable) {
      unmarshal = castItemModel(`${unmarshal}.data`);
    } else {
      unmarshal = `${unmarshal}.data`;
    }
  }

  const sseDataRequired = isSSEDataRequired(responseBodyType.ItemType);
  let args = "";
  args += "http_res_";
  args += `, lambda raw: ${unmarshal}`;
  args += content.SSESentinel
    ? `, sentinel="${content.SSESentinel.replaceAll('"', '\\"')}"`
    : "";
  args += ", client_ref=self";
  args += sseDataRequired ? "" : ", data_required=False";
  const value = `${pythonEventStreamMethodTypeRef()}(${args})`;

  const headers = templateResponseHeaders(op, subResponse);

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      return `return ${templateResponseObject(
        op,
        `${
          responseFormat == "flat"
            ? sanitizeFieldName(resultField?.Name)
            : sanitizeFieldName(content.Content.Name)
        }=${value}`,
        "",
        headers,
      )}`;
    case responseFormat == "flat":
      return `return ${value}`;
    default:
      throw new Error(`Unexpected SSE response format: ${responseFormat}`);
  }
}

// @ts-ignore
function templateJsonLHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  addJsonlImport();
  const responseFormat = getResponseFormat();
  const resultField = getResultField(op);

  let responseBodyType = content.Content.Type;

  const model = sanitizePydanticType(
    typeDefToFieldDef(responseBodyType.ItemType),
    { optional: false, nullable: false },
  );

  let args = "";
  args += "http_res_";
  args += `, lambda raw: utils_.unmarshal_json(raw, ${model})`;
  args += ", client_ref=self";
  const value = `jsonl.JsonLStreamAsync(${args})`;

  const headers = templateResponseHeaders(op, subResponse);

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      return `return ${templateResponseObject(
        op,
        `${
          responseFormat == "flat"
            ? sanitizeFieldName(resultField?.Name)
            : sanitizeFieldName(content.Content.Name)
        }=${value}`,
        "",
        headers,
      )}`;
    case responseFormat == "flat":
      return `return ${value}`;
    default:
      throw new Error(`Unexpected JsonL response format: ${responseFormat}`);
  }
}

function templateResponseObject(
  op: Operation,
  bodyField?: string,
  paginationField?: string,
  headersField?: string,
): string {
  addUtilsImport();
  const responseFormat = getResponseFormat();

  if (responseFormat == "flat" && !op.Response.Type.ResponseEnvelope) {
    return "None";
  }

  const responseType = sanitizePydanticType(
    typeDefToFieldDef(op.Response.Type),
    { optional: false, nullable: false },
  );

  const parts = [];
  if (bodyField) {
    parts.push(bodyField);
  }

  switch (true) {
    case responseFormat == "envelope":
      parts.push(
        `status_code=http_res_.status_code, content_type=http_res_.headers.get("Content-Type") or "", raw_response=http_res_`,
      );
      break;
    case responseFormat == "envelope-http":
      addSDKImportForAccessNamespace("shared");
      parts.push(
        `http_meta=${getAccessNamespaceSuffixed(
          "shared",
        )}HTTPMetadata(request=req_, response=http_res_)`,
      );
      break;
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      break;
    default:
      throw new Error(`Unexpected response format: ${responseFormat}`);
  }

  if (paginationField) {
    parts.push(paginationField);
  }
  if (headersField) {
    parts.push(headersField);
  }

  return `${responseType}(${parts.join(", ")})`;
}
