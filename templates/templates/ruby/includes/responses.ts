// @ts-ignore
function templateResponseHandler(
  op: Operation,
  subResponse?: SubResponse,
  content?: ResponseBodyContent,
): string {
  const pagination = op.Extensions.Pagination ? true : false;
  if (!op.Response.Type) {
    return "return";
  }
  if (content == null || subResponse == null) {
    const responseFormat = getResponseFormat();

    const responseType = sanitizeType(op.Response.Type, false, false, "");

    return `return ${templateEmptyResponseBuilder(
      op,
      responseFormat,
      responseType,
    )}`;
  }
  switch (content.SerializationMethod) {
    case "json":
      return templateJSONHandler(op, subResponse, content, pagination);
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

function getHTTPMetadata(): string {
  return `${getScopeNamespace("shared")}::HTTPMetadata`;
}

function templateEmptyResponseBuilder(
  op: Operation,
  responseFormat: string,
  responseType: string,
): string {
  const params = [];
  switch (responseFormat) {
    case "envelope":
      params.push("status_code: http_response.status");
      params.push("content_type: content_type");
      params.push("raw_response: http_response");
      break;
    case "envelope-http":
      params.push(`http_meta: ${getHTTPMetadata()}.new(
    response: http_response,
    request: ${sorbetMust("http_request")}
  )`);
      break;
    case "flat":
      const resultField = getResultField(op);
      if (resultField) {
        params.push(`result: nil`);
      }
      break;
  }
  if (op.Response.Responses.some((r) => r.Headers)) {
    // If any response has headers, it will be required.
    params.push("headers: {}");
  }
  return `${responseType}.new${
    params &&
    `(
  ${params.join(",\n  ")}
)`
  }`;
}

// @ts-ignore
function templateJSONHandler(
  op: Operation,
  subResponse: SubResponse,
  content?: ResponseBodyContent,
  pagination: boolean = false,
): string {
  const responseFormat = getResponseFormat();
  const responseType = sanitizeType(op.Response.Type, false, false, "");

  const deserializedType = sanitizeCrystallineType(
    content.Content.Type,
    false,
    false,
    "",
  );
  let handler = `response_data = http_response.env.response_body\nobj = Crystalline.unmarshal_json(JSON.parse(response_data), ${deserializedType})`;

  if (subResponse.Error) {
    const params = [];
    if (hasRawResponseField(content.Content.Type)) {
      switch (responseFormat) {
        case "envelope":
          handler += `\nobj.raw_response = http_response`;
          break;
        case "envelope-http":
          params.push(`\nobj.http_meta = ${getHTTPMetadata()}.new(
  response: http_response,
  request: ${sorbetMust("http_request")}
)`);
          if (subResponse.Headers) {
            params.push("headers: http_response.headers");
          } else if (op.Response.Responses.some((r) => r.Headers)) {
            // If any response has headers, it will be required.
            params.push("headers: {}");
          }
          handler += `${params.join(",\n  ")}`;
          break;
      }
    }
    return handler + `\nraise obj`;
  }
  if (responseFormat != "flat") {
    handler += `\nresponse = ${templateResponseBuilder(
      op,
      subResponse,
      responseFormat,
      responseType,
      content,
    )}`;
    if (pagination) {
      handler += templatePagination(op);
    }
    handler += `\n\nreturn response`;
  } else if (op.Response.Type.ResponseEnvelope) {
    handler += `\nresult = obj`;
    handler += `\nresponse = ${templateResponseBuilder(
      op,
      subResponse,
      responseFormat,
      responseType,
      content,
    )}`;
    if (pagination) {
      handler += templatePagination(op);
    }
    handler += `\n\nreturn response`;
  } else {
    handler += `\n\nreturn obj`;
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
  const responseType = sanitizeType(op.Response.Type, false, false, "");

  function bodyAccessor(): string {
    switch (true) {
      case content.SerializationMethod.toString() === "raw" &&
        content.Content.Type.Type.toString() === "stream":
        return `http_response.env.body`;
      case content.SerializationMethod.toString() === "raw" &&
        content.Content.Type.Type.toString() !== "stream":
        return `http_response.env.body.force_encoding('UTF-8')`;
      case content.SerializationMethod.toString() === "string":
        return `http_response.env.body.force_encoding('UTF-8')`;
      default:
        return "";
    }
  }

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      let handler = `obj = ${bodyAccessor()}\n\n`;
      handler += `return ${templateResponseBuilder(
        op,
        subResponse,
        responseFormat,
        responseType,
        content,
      )}`;
      return handler;
    case responseFormat == "flat":
      return `return ${bodyAccessor()}`;
    default:
      throw new Error(`Unexpected SSE response format: ${responseFormat}`);
  }
}

// @ts-ignore
function templateEventStreamHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();
  const moduleName = sanitizeModuleName(context.Global.Config.Module);
  const responseType = sanitizeType(op.Response.Type, false, false, "");

  const itemCrystallineType = sanitizeCrystallineType(
    content.Content.Type.ItemType,
    false,
    false,
    "",
  );

  const dataAccessor = getFlattenedEventStreamField(
    content.Content.Type.ItemType,
  )
    ? ".data"
    : "";

  const sentinelArg = content.SSESentinel
    ? `, sentinel: '${content.SSESentinel.replaceAll("'", "\\'")}'`
    : "";

  const sseDataRequired = isSSEDataRequired(content.Content.Type.ItemType);
  const dataRequiredArg = sseDataRequired ? "" : ", data_required: false";

  let handler = `decoder = lambda { |raw|\n  Crystalline.unmarshal_json(JSON.parse(raw), ${itemCrystallineType})${dataAccessor}\n}`;
  handler += `\nobj = ${moduleName}::Utils::EventStream.new(http_response, decoder${sentinelArg}, sdk_ref: self${dataRequiredArg})`;

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
      handler += `\nresponse = ${templateResponseBuilder(
        op,
        subResponse,
        responseFormat,
        responseType,
        content,
      )}`;
      handler += `\n\nreturn response`;
      return handler;
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      handler += `\nresult = obj`;
      handler += `\nresponse = ${templateResponseBuilder(
        op,
        subResponse,
        responseFormat,
        responseType,
        content,
      )}`;
      handler += `\n\nreturn response`;
      return handler;
    case responseFormat == "flat":
      handler += `\n\nreturn obj`;
      return handler;
    default:
      throw new Error(`Unexpected SSE response format: ${responseFormat}`);
  }
}
registerTemplateFunc("templateEventStreamHandler", templateEventStreamHandler);

// @ts-ignore
function templateJsonLHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
): string {
  const responseFormat = getResponseFormat();
  const moduleName = sanitizeModuleName(context.Global.Config.Module);
  const responseType = sanitizeType(op.Response.Type, false, false, "");

  const itemCrystallineType = sanitizeCrystallineType(
    content.Content.Type.ItemType,
    false,
    false,
    "",
  );

  let handler = `decoder = lambda { |raw|\n  Crystalline.unmarshal_json(JSON.parse(raw), ${itemCrystallineType})\n}`;
  handler += `\nobj = ${moduleName}::Utils::JsonLStream.new(http_response, decoder, sdk_ref: self)`;

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
      handler += `\nresponse = ${templateResponseBuilder(
        op,
        subResponse,
        responseFormat,
        responseType,
        content,
      )}`;
      handler += `\n\nreturn response`;
      return handler;
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      handler += `\nresult = obj`;
      handler += `\nresponse = ${templateResponseBuilder(
        op,
        subResponse,
        responseFormat,
        responseType,
        content,
      )}`;
      handler += `\n\nreturn response`;
      return handler;
    case responseFormat == "flat":
      handler += `\n\nreturn obj`;
      return handler;
    default:
      throw new Error(`Unexpected JSONL response format: ${responseFormat}`);
  }
}
registerTemplateFunc("templateJsonLHandler", templateJsonLHandler);

function templateResponseBuilder(
  op: Operation,
  subResponse: SubResponse,
  responseFormat: string,
  responseType: string,
  content?: ResponseBodyContent,
): string {
  const params = [];
  switch (responseFormat) {
    case "envelope":
      params.push("status_code: http_response.status");
      params.push("content_type: content_type");
      params.push("raw_response: http_response");
      if (subResponse.Headers) {
        params.push("headers: http_response.headers");
      } else if (op.Response.Responses.some((r) => r.Headers)) {
        // If any response has headers, it will be required.
        params.push("headers: {}");
      }
      if (content) {
        const val = `${sorbetUnsafe("obj")}`;
        params.push(`${sanitizeFieldName(content.Content.Name)}: ${val}`);
      }
      return `${responseType}.new(
  ${params.join(",\n  ")}
)`;

    case "envelope-http":
      params.push(`http_meta: ${getHTTPMetadata()}.new(
    response: http_response,
    request: ${sorbetMust("http_request")}
  )`);
      if (subResponse.Headers) {
        params.push("headers: http_response.headers");
      } else if (op.Response.Responses.some((r) => r.Headers)) {
        // If any response has headers, it will be required.
        params.push("headers: {}");
      }
      if (content) {
        const val = `${sorbetUnsafe("obj")}`;

        params.push(`${sanitizeFieldName(content.Content.Name)}: ${val}`);
      }

      return `${responseType}.new(
  ${params.join(",\n  ")}
)`;
    case "flat":
      params.push(`result: result`);
      if (subResponse.Headers) {
        params.push("headers: http_response.headers");
      } else if (op.Response.Responses.some((r) => r.Headers)) {
        // If any response has headers, it will be required.
        params.push("headers: {}");
      }
      return `${responseType}.new(
  ${params.join(",\n  ")}
)`;
  }

  return ``;
}

function anyResponseUsesContentType(responses: SubResponse[]): boolean {
  return responses.some((response) => response.Content.length > 0);
}

registerTemplateFunc("anyResponseUsesContentType", anyResponseUsesContentType);

function templatePagination(op: Operation): string {
  const indent = "  ";

  const paginationExt = op.Extensions.Pagination;
  let args = op.Arguments;
  const canSimplify = op.Extensions?.Pagination?.Outputs?.CanUseDotNotation;
  let inputVar = "request";

  let response = `\nsdk = self\n\nresponse.next_page = proc do \n`;
  const limitInput = getPaginationInput(paginationExt, "limit");
  const pageInput = getPaginationInput(paginationExt, "page");
  const offsetInput = getPaginationInput(paginationExt, "offset");
  const cursorInput = getPaginationInput(paginationExt, "cursor");
  const _hasPaginationURL = hasPaginationURL(op);

  if (op.Extensions.Pagination.Type == "offsetLimit") {
    if (pageInput) {
      response += `${indent}request_page = ${sorbetMust(
        templatePaginationInputAccessor(op, inputVar, pageInput),
      )}\n`;
      response += `${indent}next_page = request_page + 1\n`;

      if (op.Extensions.Pagination.Outputs.NumPages) {
        if (canSimplify && !op.Extensions.Pagination.Outputs.NumPagesDot) {
          response += `${indent}num_pages = response_data\n`;
        } else {
          response += `${indent}num_pages = Janeway.enum_for('${sanitizeJsonPath(
            op.Extensions.Pagination.Outputs.NumPages,
          )}', JSON.parse(response_data)).search\n`;
        }
        response += `${indent}if num_pages.nil? || num_pages[0] <= request_page
    next nil
  end\n`;
      }
    }

    if (offsetInput) {
      response += `${indent}request_offset = ${templatePaginationInputAccessor(
        op,
        inputVar,
        offsetInput,
      )}\n`;
    }
    response += `${indent}if !response_data
    next nil
  end\n`;
  } else if (_hasPaginationURL) {
    if (canSimplify && !op.Extensions.Pagination.Outputs.NextURLDot) {
      response += `${indent}next_url = response_data\n`;
    } else {
      response += `${indent}next_url = Janeway.enum_for('${sanitizeJsonPath(
        op.Extensions.Pagination.Outputs.NextURL,
      )}', JSON.parse(response_data)).search\n`;
    }
    response += `${indent}if next_url.nil? || next_url.empty?
    next nil
  else
    next_url = next_url[0]
  end\n`;
    response += `${indent}if next_url.start_with? '/' \n`;
    response += `${indent}${indent}url = get_url(base_url: base_url)\n
    next_url = "#{url}#{next_url}"
\n
  end\n`;
    response += `${indent}if next_url !~ URI::RFC2396_PARSER.make_regexp
    next nil
  end\n`;
  } else {
    if (canSimplify && !op.Extensions.Pagination.Outputs.NextCursorDot) {
      response += `${indent}next_cursor = response_data\n`;
    } else {
      response += `${indent}next_cursor = Janeway.enum_for('${sanitizeJsonPath(
        op.Extensions.Pagination.Outputs.NextCursor,
      )}', JSON.parse(response_data)).search\n`;
    }
    response += `${indent}if next_cursor.nil?
    next nil
  else
    next_cursor = next_cursor[0]
    if next_cursor.nil?
      next nil
    end
  end\n`;
  }
  if (op.Extensions.Pagination.Outputs.Results) {
    if (canSimplify && !op.Extensions.Pagination.Outputs.ResultsDot) {
      response += `${indent}results = JSON.parse(response_data)
  if results.nil?
    next nil
  end\n`;
    } else {
      response += `${indent}results = Janeway.enum_for('${sanitizeJsonPath(
        op.Extensions.Pagination.Outputs.Results,
      )}', JSON.parse(response_data)).search\n`;
      response += `${indent}if results.is_a? Array
    results = results[0]
  end\n`;
    }
    response += `${indent}if results.count.zero?
    next nil
  end\n`;
    if (limitInput) {
      response += `${indent}request_limit = ${
        templatePaginationInputAccessor(op, inputVar, limitInput) || 0
      }\n`;
      response += `${indent}if results.count < request_limit
    next nil
  end\n`;
    }
    if (offsetInput) {
      response += `${indent}next_offset = request_offset + results.count\n`;
    }
  }
  /* Setup Complete */
  let request = "";
  let requestParams = [];
  let prerequisites = "";
  request += `\n${indent}sdk.${sanitizeMethodName(op)}(\n`;
  if (["all", "body"].includes(op.Arguments.Flattening)) {
    for (const field of args.Sorted) {
      let fieldName = sanitizeFieldName(field.Name);
      let argName = sanitizeMethodArgumentName(field.Name);
      if (field.Name === pageInput?.Name) {
        requestParams.push(`${indent}${argName}: next_page`);
      } else if (field.Name == offsetInput?.Name) {
        requestParams.push(`${indent}${argName}: next_offset`);
      } else if (field.Name == cursorInput?.Name) {
        requestParams.push(`${indent}${argName}: next_cursor`);
      } else if (op.Security && field === op.Security) {
        requestParams.push(`${indent}${argName}: ${argName}`);
      } else {
        requestParams.push(`${indent}${argName}: request.${fieldName}`);
      }
    }
  } else if (op.Arguments.Flattening.toString() === "params") {
    for (const field of args.Sorted) {
      if (op.Arguments.BodyField === field) {
        continue;
      }
      let fieldName = sanitizeFieldName(field.Name);
      let argName = sanitizeMethodArgumentName(field.Name);
      const inFieldName = resolveFieldNameFromInput(op, field, "parameters");
      if (inFieldName !== undefined) {
        requestParams.push(`${indent}${indent}${argName}: ${inFieldName}`);
      } else if (op.Security && field === op.Security) {
        requestParams.push(`${indent}${indent}${argName}: ${argName}`);
      } else {
        requestParams.push(
          `${indent}${indent}${argName}: request.${fieldName}`,
        );
      }
    }

    if (op.Arguments.BodyField) {
      let innerRequest = "";
      const innerRequestParams = [];
      let bodyFieldName = sanitizeFieldName(op.Arguments.BodyField.Name);
      let bodyArgName = sanitizeMethodArgumentName(op.Arguments.BodyField.Name);
      if (op.Arguments.BodyField.Type.Type.toString() === "class") {
        innerRequest += `${indent}${indent}${bodyArgName}: ${sanitizeClass(
          op.Arguments.BodyField.Type,
          "",
          false,
        )}.new(\n`;
        for (const f of args.BodyField.Type.Fields) {
          const fieldName = sanitizeFieldName(f.Name);
          const inFieldName = resolveFieldNameFromInput(op, f, "requestBody");
          if (inFieldName !== undefined) {
            innerRequestParams.push(
              `${indent}${indent}${indent}${fieldName}: ${inFieldName}`,
            );
          } else {
            innerRequestParams.push(
              `${indent}${indent}${indent}${fieldName}: !request.${bodyFieldName}.nil? ? request.${bodyFieldName}.${fieldName} : ${templateDefaultValue(
                f,
                "",
              )}`,
            );
          }
        }
        innerRequest += innerRequestParams.join(",\n");
        innerRequest += `\n${indent}${indent})`;
        requestParams.push(innerRequest);
      } else {
        innerRequestParams.push(
          `${indent}${indent}${indent}${bodyArgName}: request.${bodyFieldName}`,
        );
      }
    }
  } else {
    if (op.Request) {
      let innerRequest = "";
      const innerRequestParams = [];
      innerRequest += `${indent}${indent}request: ${sanitizeClass(
        op.Request.Field.Type,
        "",
        false,
      )}.new(\n`;

      for (const field of op.Request.Field.Type.Fields) {
        if (field.Name === pageInput?.Name) {
          innerRequestParams.push(
            `${indent}${indent}${indent}${sanitizeFieldName(
              pageInput.Name,
            )}: next_page`,
          );
        } else if (field.Name === offsetInput?.Name) {
          innerRequestParams.push(
            `${indent}${indent}${indent}${sanitizeFieldName(
              offsetInput.Name,
            )}: next_offset`,
          );
        } else if (field.Name === cursorInput?.Name) {
          innerRequestParams.push(
            `${indent}${indent}${indent}${sanitizeFieldName(
              cursorInput.Name,
            )}: next_cursor`,
          );
        } else {
          const needsPrerequisite =
            field.Type.Fields.filter(
              (f) =>
                (pageInput?.In == "requestBody" && f.Name == pageInput?.Name) ||
                (offsetInput?.In == "requestBody" &&
                  f.Name == offsetInput?.Name) ||
                (cursorInput?.In == "requestBody" &&
                  f.Name == cursorInput?.Name),
            ).length > 0;
          if (needsPrerequisite) {
            prerequisites += `${indent}${sanitizeFieldName(
              field.Name,
            )} = ${sanitizeClass(field.Type, "", false)}.new(\n`;
            const prerequisitesParams = [];
            for (const reqField of field.Type.Fields) {
              const inFieldName = resolveFieldNameFromInput(
                op,
                reqField,
                "requestBody",
              );
              if (inFieldName !== undefined) {
                prerequisitesParams.push(
                  `${indent}${indent}${sanitizeFieldName(
                    reqField.Name,
                  )}: ${inFieldName}`,
                );
              } else {
                prerequisitesParams.push(
                  `${indent}${indent}${sanitizeFieldName(
                    reqField.Name,
                  )}: !request.${sanitizeFieldName(
                    field.Name,
                  )}.nil? ? request.${sanitizeFieldName(
                    field.Name,
                  )}.${sanitizeFieldName(
                    reqField.Name,
                  )} : ${templateDefaultValue(reqField, "")}`,
                );
              }
            }
            prerequisites += prerequisitesParams.join(",\n");
            prerequisites += `\n${indent})\n`;
            innerRequestParams.push(
              `${indent}${indent}${indent}${sanitizeFieldName(
                field.Name,
              )}: ${sanitizeFieldName(field.Name)}`,
            );
          } else if (op.Request.Field.Nullable) {
            // a field the nil request never carried is copied as nil
            innerRequestParams.push(
              `${indent}${indent}${indent}${sanitizeFieldName(
                field.Name,
              )}: !request.nil? ? request.${sanitizeFieldName(
                field.Name,
              )} : nil`,
            );
          } else {
            innerRequestParams.push(
              `${indent}${indent}${indent}${sanitizeFieldName(
                field.Name,
              )}: request.${sanitizeFieldName(field.Name)}`,
            );
          }
        }
      }
      innerRequest += innerRequestParams.join(",\n");
      innerRequest += `\n${indent}${indent})`;
      requestParams.push(innerRequest);
    }
    if (op.Security) {
      requestParams.push(`${indent}${indent}security: security`);
    }
  }
  if (op.Servers) {
    requestParams.push(`${indent}${indent}server_url: server_url`);
  }
  if (hasPaginationURL(op)) {
    requestParams.push(`${indent}${indent}url_override: next_url`);
  }
  requestParams.push(`${indent}${indent}http_headers: http_headers`);
  request += requestParams.join(",\n");
  request += `\n${indent})\n`;

  response += prerequisites;
  response += request;
  response += `end\n`;

  return response;
}

registerTemplateFunc("templatePagination", templatePagination);

function templatePaginationInputAccessor(
  op: Operation,
  inputVar: string,
  field: PaginationInputs,
): string {
  const reqField = op.Request?.Field;
  if (!reqField) {
    throw new Error(
      "Operation does not have a request to read pagination input from",
    );
  }

  const usesRequestWraper = hasAnnotation(reqField, "requestWrapper");

  const inLoc = field.In.toString();

  // a nullable request arrives as nil, so guard before reading inputs off it
  const nilGuard = reqField.Nullable ? `!${inputVar}.nil? && ` : "";

  const fallback = getPaginationDefaults(op)[field.Type.toString()] ?? 0;

  if (inLoc === "requestBody" && usesRequestWraper) {
    const bodyFieldName = reqField.Type?.Fields.find((f) =>
      hasAnnotation(f, "request"),
    )?.Name;
    if (!bodyFieldName) {
      throw new Error(
        "Expected request wrapper does not have a request body field",
      );
    }
    return `${nilGuard}!${inputVar}.${sanitizeFieldName(
      bodyFieldName,
    )}.nil? && !${inputVar}.${sanitizeFieldName(
      bodyFieldName,
    )}.${sanitizeFieldName(field.Name)}.nil? ? ${inputVar}.${sanitizeFieldName(
      bodyFieldName,
    )}.${sanitizeFieldName(field.Name)} : ${fallback}`;
  } else if (inLoc === "requestBody" && !usesRequestWraper) {
    return `${nilGuard}!${inputVar}.${sanitizeFieldName(
      field.Name,
    )}.nil? ? ${inputVar}.${sanitizeFieldName(field.Name)} : ${fallback}`;
  } else if (inLoc === "parameters") {
    return `${nilGuard}!${inputVar}.${sanitizeFieldName(
      field.Name,
    )}.nil? ? ${inputVar}.${sanitizeFieldName(field.Name)} : ${fallback}`;
  } else {
    throw new Error(
      `Could not template pagination input accessor for operation "${op.ID}": "${field.Name}" in ${inLoc}`,
    );
  }
}
registerTemplateFunc(
  "templatePaginationInputAccessor",
  templatePaginationInputAccessor,
);

function resolveFieldNameFromInput(
  op: Operation,
  f: FieldDef,
  requestPart: string,
) {
  const paginationExt = op.Extensions.Pagination;

  const pageInput = getPaginationInput(paginationExt, "page");
  const offsetInput = getPaginationInput(paginationExt, "offset");
  const cursorInput = getPaginationInput(paginationExt, "cursor");
  if (pageInput?.In.toString() === requestPart && f.Name === pageInput.Name) {
    return "next_page";
  } else if (
    offsetInput?.In.toString() === requestPart &&
    f.Name === offsetInput.Name
  ) {
    return "next_offset";
  } else if (
    cursorInput?.In.toString() === requestPart &&
    f.Name === cursorInput.Name
  ) {
    return "next_cursor";
  }
  return undefined;
}

function sanitizeFuncName(operation: Operation): string {
  const group = operation.OwningSDK.Group;
  let prefix = "";
  // We only qualify sub-sdk functions, not the root SDKs functions.
  if (group) {
    prefix = group.replaceAll(".", "_") + "_";
  }

  return sanitizeFieldName(`${prefix}${operation.GetID()}`);
}
registerTemplateFunc("sanitizeFuncName", sanitizeFuncName);

function sanitizePaginationAccess(
  op: Operation,
  input: PaginationInputs,
): string {
  if (!op.Request) {
    throw new Error(
      `${op.ID}: ${input.Name}: attempted to access pagination field but operation has no request`,
    );
  }

  const field = sanitizeFieldName(input.Name);
  switch (input.In.toString()) {
    case "parameters":
      return `request.${field}`;
    case "requestBody":
      if (!op.Request.RequestBody) {
        throw new Error(
          `${op.ID}: ${input.Name}: attempted to access pagination field in request body which is not defined`,
        );
      }
      if (op.Request.IsRequestBody) {
        return `request.${field}`;
      }

      return `request.${sanitizeFieldName(
        op.Request.RequestBody.Name,
      )}.${field}`;
    default:
      throw new Error(
        `${op.ID}: ${
          input.Name
        }: unknown pagination input location "${input.In.toString()}"]`,
      );
  }
}
registerTemplateFunc("sanitizePaginationAccess", sanitizePaginationAccess);
