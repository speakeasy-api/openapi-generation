// @ts-ignore
function templateResponseHandler(
  op: Operation,
  subResponse?: SubResponse,
  content?: ResponseBodyContent,
): string {
  const pagination = op.Extensions.Pagination ? true : false;
  if (!op.Response.Type) {
    return "return;";
  }

  if (content == null || subResponse == null) {
    const responseFormat = getResponseFormat();

    const responseType = sanitizeType(
      op.Response.Type,
      false,
      false,
      "",
      Qualification.TYPE,
    );

    return `return ${templateEmptyResponseBuilder(
      op,
      responseFormat,
      responseType,
    )};`;
  }
  switch (content.SerializationMethod) {
    case "json":
      return templateJSONHandler(op, subResponse, content, pagination);
    case "raw":
    case "string":
      return templateBodyHandler(op, subResponse, content);
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
  content?: ResponseBodyContent,
  pagination: boolean = false,
): string {
  const responseFormat = getResponseFormat();
  const responseType = sanitizeType(
    op.Response.Type,
    false,
    false,
    "",
    Qualification.TYPE,
  );

  const deserializedType = sanitizeType(
    content.Content.Type,
    false,
    false,
    "",
    Qualification.ANNOTATION,
  );
  addImportInline("Speakeasy\\Serializer\\DeserializationContext");
  let handler = `$serializer = Utils\\JSON::createSerializer();\n$responseData = (string) $httpResponse->getBody();\n$obj = $serializer->deserialize($responseData, '${deserializedType}', 'json', DeserializationContext::create()->setRequireAllRequiredProperties(true));`;

  if (subResponse.Error) {
    if (hasRawResponseField(content.Content.Type)) {
      switch (responseFormat) {
        case "envelope":
          handler += `\n$obj->rawResponse = $httpResponse;`;
          break;
        case "envelope-http":
          handler += `\n$obj->httpMeta = new ${getHTTPMetadata()}(
    response: $httpResponse,
    request: $httpRequest
);`;
          break;
      }
    }
    return handler + `\nthrow $obj->toException();`;
  }
  if (responseFormat != "flat") {
    handler += `\n$response = ${templateResponseBuilder(
      subResponse,
      responseFormat,
      responseType,
      content,
    )}`;
    if (pagination) {
      handler += templatePagination(op);
    }
    handler += `\n\nreturn $response;`;
  } else if (op.Response.Type.ResponseEnvelope) {
    handler += `\n$result = $obj;`;
    handler += `\n$response = ${templateResponseBuilder(
      subResponse,
      responseFormat,
      responseType,
      content,
    )}`;
    if (pagination) {
      handler += templatePagination(op);
    }
    handler += `\n\nreturn $response;`;
  } else {
    handler += `\n\nreturn $obj;`;
  }
  return handler;
}

// @ts-ignore
function templateBodyHandler(
  op: Operation,
  subResponse: SubResponse,
  content: ResponseBodyContent,
  pagination: boolean = false,
): string {
  const responseFormat = getResponseFormat();
  const responseType = sanitizeType(
    op.Response.Type,
    false,
    false,
    "",
    Qualification.TYPE,
  );

  function bodyAccessor(): string {
    switch (true) {
      case content.SerializationMethod.toString() === "raw" &&
        content.Content.Type.Type.toString() === "stream":
        return `$httpResponse->getBody()`;
      case content.SerializationMethod.toString() === "raw" &&
        content.Content.Type.Type.toString() !== "stream":
        return `$httpResponse->getBody()->getContents()`;
      case content.SerializationMethod.toString() === "string":
        return `$httpResponse->getBody()->getContents()`;
      default:
        return ".content";
    }
  }

  switch (true) {
    case responseFormat == "envelope":
    case responseFormat == "envelope-http":
    case responseFormat == "flat" && op.Response.Type.ResponseEnvelope:
      let handler = `$obj = ${bodyAccessor()};\n\n`;
      handler += `return ${templateResponseBuilder(
        subResponse,
        responseFormat,
        responseType,
        content,
      )}`;
      return handler;
    case responseFormat == "flat":
      return `return ${bodyAccessor()};`;
    default:
      throw new Error(`Unexpected SSE response format: ${responseFormat}`);
  }
}

function templateResponseBuilder(
  subResponse: SubResponse,
  responseFormat: string,
  responseType: string,
  content?: ResponseBodyContent,
): string {
  const params = [];
  switch (responseFormat) {
    case "envelope":
      params.push("statusCode: $statusCode");
      params.push("contentType: $contentType");
      params.push("rawResponse: $httpResponse");
      if (subResponse.Headers) {
        params.push("headers: $httpResponse->getHeaders()");
      }
      if (content) {
        params.push(`${sanitizeFieldName(content.Content.Name)}: $obj`);
      }
      return `new ${responseType}(
    ${params.join(",\n    ")});`;

    case "envelope-http":
      params.push(`httpMeta: new ${getHTTPMetadata()}(
        response: $httpResponse,
        request: $httpRequest
    )`);
      if (content) {
        params.push(`    ${sanitizeFieldName(content.Content.Name)}: $obj`);
      }

      return `new ${responseType}(
    ${params.join(",\n")}
);`;
    case "flat":
      params.push(`result: $result`);

      return `new ${responseType}(
    ${params.join(",\n    ")}
);`;
  }

  return ``;
}

function templateEmptyResponseBuilder(
  op: Operation,
  responseFormat: string,
  responseType: string,
): string {
  const params = [];
  switch (responseFormat) {
    case "envelope":
      params.push("statusCode: $statusCode");
      params.push("contentType: $contentType");
      params.push("rawResponse: $httpResponse");
      return `new ${responseType}(
    ${params.join(",\n    ")}
)`;
    case "envelope-http":
      params.push(`httpMeta: new ${getHTTPMetadata()}(
        response: $httpResponse,
        request: $httpRequest
    )`);
      return `new ${responseType}(
    ${params.join(",\n")}
)`;
    case "flat":
      const resultField = getResultField(op);
      if (resultField) {
        params.push(`result: null`);
      } else {
        return `new ${responseType}()`;
      }

      return `new ${responseType}(
    ${params.join(",\n    ")}
)`;
  }

  return "";
}

//@ts-ignore
function getHTTPMetadata(): string {
  const access = getScopeNamespace("shared", false);
  const parts = access.split("\\");
  const lastPart = parts[parts.length - 1];
  return access ? `${lastPart}\\HTTPMetadata` : "HTTPMetadata";
}

function templatePagination(op: Operation): string {
  const indent = "    ";

  const paginationExt = op.Extensions.Pagination;
  let args = op.Arguments;
  const canSimplify = op.Extensions?.Pagination?.Outputs?.CanUseDotNotation;
  let inputVar = "$request";

  let uses = new Set<string>();
  uses.add("sdk");

  let response = `\n$sdk = $this;\n\n$response->next = function () use (**ALL_USES**): ?${sanitizeType(
    op.Response.Type,
    false,
    false,
    "",
    Qualification.TYPE,
  )} {\n`;
  const limitInput = getPaginationInput(paginationExt, "limit");
  const pageInput = getPaginationInput(paginationExt, "page");
  const offsetInput = getPaginationInput(paginationExt, "offset");
  const cursorInput = getPaginationInput(paginationExt, "cursor");
  const _hasPaginationURL = hasPaginationURL(op);

  if (op.Extensions.Pagination.Type == "offsetLimit") {
    if (pageInput) {
      uses.add(inputVar.substring(1));
      response += `${indent}$page = ${
        templatePaginationInputAccessor(op, inputVar, pageInput, uses) || 0
      };\n`;
      response += `${indent}$nextPage = $page + 1;\n`;

      if (op.Extensions.Pagination.Outputs.NumPages) {
        if (canSimplify && !op.Extensions.Pagination.Outputs.NumPagesDot) {
          response += `${indent}$numPages = $responseData;\n`;
        } else {
          response += `${indent}$jsonObject = new \\JsonPath\\JsonObject($responseData);\n`;
          response += `${indent}$numPages = $jsonObject->get('${sanitizeJsonPath(
            op.Extensions.Pagination.Outputs.NumPages,
          )}');\n`;
        }
        response += `${indent}if ($numPages == null || $numPages[0] <= $page) {
        return null;
    }\n`;
      }
    }

    if (offsetInput) {
      uses.add(inputVar.substring(1));
      response += `${indent}$offset = ${templatePaginationInputAccessor(
        op,
        inputVar,
        offsetInput,
        uses,
      )};\n`;
    }
    uses.add("responseData");
    response += `${indent}if (! $responseData) {
        return null;
    }\n`;
  } else if (_hasPaginationURL) {
    if (canSimplify && !op.Extensions.Pagination.Outputs.NextURLDot) {
      uses.add("responseData");
      response += `${indent}$nextURL = $responseData;\n`;
    } else {
      uses.add("responseData");
      response += `${indent}$jsonObject = new \\JsonPath\\JsonObject($responseData);\n`;
      response += `${indent}$nextURL = $jsonObject->get('${sanitizeJsonPath(
        op.Extensions.Pagination.Outputs.NextURL,
      )}');\n`;
    }
    response += `${indent}if ($nextURL == null) {
        return null;
    } else {
        $nextURL = $nextURL[0];
    }\n`;
    response += `${indent}if (str_starts_with($nextURL, '/')) {\n`;
    uses.add("baseUrl");
    response += `${indent}${indent}$url = $this->getUrl($baseUrl, []);\n
        $nextURL = $url.$nextURL;\n
    }\n`;
    response += `${indent}if (filter_var($nextURL, FILTER_VALIDATE_URL) === false) {
        return null;
    }\n`;
  } else {
    if (canSimplify && !op.Extensions.Pagination.Outputs.NextCursorDot) {
      uses.add("responseData");
      response += `${indent}$nextCursor = $responseData;\n`;
    } else {
      uses.add("responseData");
      response += `${indent}$jsonObject = new \\JsonPath\\JsonObject($responseData);\n`;
      response += `${indent}$nextCursor = $jsonObject->get('${sanitizeJsonPath(
        op.Extensions.Pagination.Outputs.NextCursor,
      )}');\n`;
    }
    response += `${indent}if ($nextCursor == null) {
        return null;
    } else {
        $nextCursor = $nextCursor[0];
        if ($nextCursor == null || (is_string($nextCursor) && trim($nextCursor) === '')) {
            return null;
        }
    }\n`;
  }
  if (op.Extensions.Pagination.Outputs.Results) {
    if (canSimplify && !op.Extensions.Pagination.Outputs.ResultsDot) {
      uses.add("responseData");
      response += `${indent}$results = json_decode($responseData, true);
    if ($results == null || ! is_array($results)) {
        return null;
    }\n`;
    } else {
      uses.add("responseData");
      response += `${indent}$jsonObject = new \\JsonPath\\JsonObject($responseData);\n`;
      response += `${indent}$results = $jsonObject->get('${sanitizeJsonPath(
        op.Extensions.Pagination.Outputs.Results,
      )}');\n
    if (is_array($results)) {
        $results = $results[0];
    }\n`;
    }
    response += `${indent}if (count($results) === 0) {
        return null;
    }\n`;
    if (limitInput) {
      uses.add(inputVar.substring(1));
      response += `${indent}$limit = ${
        templatePaginationInputAccessor(op, inputVar, limitInput, uses) || 0
      };\n`;
      response += `${indent}if (count($results) < $limit) {
        return null;
    }\n`;
    }
    if (offsetInput) {
      response += `${indent}$nextOffset = $offset + count($results);\n`;
    }
  }
  /* Setup Complete */
  let request = "";
  let prerequisites = "";
  request += `\n${indent}return $sdk->${sanitizeMethodName(op)}Individual(\n`;
  if (["all", "body"].includes(op.Arguments.Flattening)) {
    for (const field of args.Sorted) {
      let fieldName = sanitizeFieldName(field.Name);
      if (field.Name === pageInput?.Name) {
        request += `${indent}${fieldName}: $nextPage,\n`;
      } else if (field.Name == offsetInput?.Name) {
        request += `${indent}${fieldName}: $nextOffset,\n`;
      } else if (field.Name == cursorInput?.Name) {
        request += `${indent}${fieldName}: $nextCursor,\n`;
      } else if (op.Security && field === op.Security) {
        uses.add(fieldName);
        request += `${indent}${fieldName}: $${fieldName},\n`;
      } else {
        uses.add("request");
        request += `${indent}${fieldName}: $request != null ? $request->${fieldName} : ${templateDefaultValue(
          field,
          "",
        )},\n`;
      }
    }
  } else if (op.Arguments.Flattening.toString() === "params") {
    for (const field of args.Sorted) {
      if (op.Arguments.BodyField === field) {
        continue;
      }
      let fieldName = sanitizeFieldName(field.Name);
      const inFieldName = resolveFieldNameFromInput(op, field, "parameters");
      if (inFieldName !== undefined) {
        request += `${indent}${indent}${fieldName}: $${inFieldName},\n`;
      } else if (op.Security && field === op.Security) {
        uses.add(fieldName);
        request += `${indent}${indent}${fieldName}: $${fieldName},\n`;
      } else {
        uses.add("request");
        request += `${indent}${indent}${fieldName}: $request != null ? $request->${fieldName} : ${templateDefaultValue(
          field,
          "",
        )},\n`;
      }
    }

    if (op.Arguments.BodyField) {
      let bodyFieldName = sanitizeFieldName(op.Arguments.BodyField.Name);
      if (op.Arguments.BodyField.Type.Type.toString() === "class") {
        request += `${indent}${indent}${bodyFieldName}: new ${sanitizeClass(
          op.Arguments.BodyField.Type,
          "",
          false,
          Qualification.TYPE,
        )}(\n`;
        for (const f of args.BodyField.Type.Fields) {
          const fieldName = sanitizeFieldName(f.Name);
          const inFieldName = resolveFieldNameFromInput(op, f, "requestBody");
          if (inFieldName !== undefined) {
            request += `${indent}${indent}${indent}${fieldName}: $${inFieldName},\n`;
          } else {
            uses.add("request");
            request += `${indent}${indent}${indent}${fieldName}: $request != null && $request->${sanitizeFieldName(
              op.Arguments.BodyField.Name,
            )} != null ? $request->${sanitizeFieldName(
              op.Arguments.BodyField.Name,
            )}->${fieldName} : ${templateDefaultValue(f, "")},\n`;
          }
        }
        request += `${indent}${indent}),\n`;
      } else {
        uses.add("request");
        request += `${indent}${indent}${indent}$${bodyFieldName}: $request != null ? $request->${bodyFieldName} : ${templateDefaultValue(
          op.Arguments.BodyField,
          "",
        )},\n`;
      }
    }
  } else {
    let globalNamespace = context.Global.Config.Namespace;

    if (op.Request) {
      const modelNamespace = getModelNamespace(
        op.Request.Field.Type.OutputLocation,
      );
      if (globalNamespace != modelNamespace) {
        addImportInline(modelNamespace);
      }
      request += `${indent}${indent}request: new ${sanitizeClass(
        op.Request.Field.Type,
        "",
        false,
        Qualification.TYPE,
      )}(\n`;

      for (const field of op.Request.Field.Type.Fields) {
        if (field.Name === pageInput?.Name) {
          request += `${indent}${indent}${indent}${sanitizeFieldName(
            pageInput.Name,
          )}: $nextPage,\n`;
        } else if (field.Name === offsetInput?.Name) {
          request += `${indent}${indent}${indent}${sanitizeFieldName(
            offsetInput.Name,
          )}: $nextOffset,\n`;
        } else if (field.Name === cursorInput?.Name) {
          request += `${indent}${indent}${indent}${cursorInput.Name}: $nextCursor,\n`;
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
            const modelNamespace = getModelNamespace(field.Type.OutputLocation);
            if (globalNamespace != modelNamespace) {
              addImportInline(modelNamespace);
            }
            prerequisites += `${indent}$${sanitizeFieldName(
              field.Name,
            )} = new ${sanitizeClass(
              field.Type,
              "",
              false,
              Qualification.TYPE,
            )}(\n`;
            for (const reqField of field.Type.Fields) {
              const inFieldName = resolveFieldNameFromInput(
                op,
                reqField,
                "requestBody",
              );
              if (inFieldName !== undefined) {
                prerequisites += `${indent}${indent}${sanitizeFieldName(
                  reqField.Name,
                )}: $${inFieldName},\n`;
              } else {
                uses.add("request");
                prerequisites += `${indent}${indent}${sanitizeFieldName(
                  reqField.Name,
                )}: $request != null && $request->${sanitizeFieldName(
                  field.Name,
                )} != null ? $request->${sanitizeFieldName(
                  field.Name,
                )}->${sanitizeFieldName(
                  reqField.Name,
                )} : ${templateDefaultValue(reqField, "")},\n`;
              }
            }
            prerequisites += `${indent});\n`;
            request += `${indent}${indent}${indent}${sanitizeFieldName(
              field.Name,
            )}: $${sanitizeFieldName(field.Name)},\n`;
          } else {
            uses.add("request");
            request += `${indent}${indent}${indent}${sanitizeFieldName(
              field.Name,
            )}: $request != null ? $request->${sanitizeFieldName(
              field.Name,
            )} : ${templateDefaultValue(field, "")},\n`;
          }
        }
      }

      request += `${indent}${indent}),\n`;
    }
    if (op.Security) {
      uses.add("security");
      request += `${indent}${indent}security: $security,\n`;
    }
  }
  if (op.Servers) {
    uses.add("serverURL");
    request += `${indent}${indent}serverURL: $serverURL,\n`;
  }
  if (hasPaginationURL(op)) {
    request += `${indent}${indent}urlOverride: $nextURL,\n`;
  }

  response += prerequisites;
  response += request;

  response += `${indent});\n`;
  response += `};\n`;

  return response.replace(
    "**ALL_USES**",
    [...uses].map((u) => `$${u}`).join(", "),
  );
}

registerTemplateFunc("templatePagination", templatePagination);

function templatePaginationInputAccessor(
  op: Operation,
  inputVar: string,
  field: PaginationInputs,
  uses: Set<string>,
): string {
  const reqField = op.Request?.Field;
  if (!reqField) {
    throw new Error(
      "Operation does not have a request to read pagination input from",
    );
  }

  const usesRequestWraper = hasAnnotation(reqField, "requestWrapper");

  const inLoc = field.In.toString();

  // fall back to the per-type default; a 0 page would re-fetch page 1 forever
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
    uses.add(inputVar.substring(1));
    // guard each level: reading a property off a null wrapped body warns
    const bodyAccess = `${inputVar}->${sanitizeFieldName(bodyFieldName)}`;
    const access = `${bodyAccess}->${sanitizeFieldName(field.Name)}`;
    return `${inputVar} != null && ${bodyAccess} != null && ${access} != null ? ${access} : ${fallback}`;
  } else if (inLoc === "requestBody" && !usesRequestWraper) {
    uses.add(inputVar.substring(1));
    const access = `${inputVar}->${sanitizeFieldName(field.Name)}`;
    return `${inputVar} != null && ${access} != null ? ${access} : ${fallback}`;
  } else if (inLoc === "parameters") {
    uses.add(inputVar.substring(1));
    const access = `${inputVar}->${sanitizeFieldName(field.Name)}`;
    return `${inputVar} != null && ${access} != null ? ${access} : ${fallback}`;
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
    return "nextPage";
  } else if (
    offsetInput?.In.toString() === requestPart &&
    f.Name === offsetInput.Name
  ) {
    return "nextOffset";
  } else if (
    cursorInput?.In.toString() === requestPart &&
    f.Name === cursorInput.Name
  ) {
    return "nextCursor";
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
      return `$request->${field}`;
    case "requestBody":
      if (!op.Request.RequestBody) {
        throw new Error(
          `${op.ID}: ${input.Name}: attempted to access pagination field in request body which is not defined`,
        );
      }
      if (op.Request.IsRequestBody) {
        return `$request->${field}`;
      }

      return `$request->${sanitizeFieldName(
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
