// SSE Overload helper functions for Python

interface SSEOverloadInfo {
  jsonContentType: string;
  sseContentType: string;
  jsonResponseType: string;
  sseResponseType: string;
}

function templateSSEOverloadArguments(
  operation: Operation,
  streamType: "false" | "true" | "bool",
): string {
  // Positional operation arguments are rendered separately before the
  // keyword-only separator in method.py.stmpl.
  const keywordOnlyArgs = getPythonMethodKeywordOnlyArgs(operation);

  const lines: string[] = [];

  const streamFieldName = sseStreamFieldName(operation);
  for (const arg of keywordOnlyArgs) {
    const paramName = sanitizeParameterName(arg.Name);

    // Handle stream parameter specially
    if (arg.Name === streamFieldName) {
      lines.push(
        templateStreamParameterOverloadType(
          arg,
          operation,
          paramName,
          streamType,
        ),
      );
      continue;
    }

    // Regular parameter handling - use the same type as in the actual implementation
    const paramType = sanitizeInputParameterType(arg, {
      optional: isArgumentOptional(operation, arg),
      iterableCollections: true,
    });
    const defaultValue = templateInputParamDefault(operation, arg);
    lines.push(`${paramName}: ${paramType}${defaultValue}`);
  }

  return lines.join(",\n        ");
}
registerTemplateFunc(
  "templateSSEOverloadArguments",
  templateSSEOverloadArguments,
);

function templateStreamParameterOverloadType(
  arg: any,
  operation: Operation,
  paramName: string,
  streamType: "false" | "true" | "bool",
): string {
  // Check if stream is nullable and/or optional
  const isOptional = isArgumentOptional(operation, arg);
  const isNullable = arg.Nullable === true;

  addImport("typing", "Literal");

  // For stream=True overload, always use Literal[True] regardless of optional/nullable
  if (streamType === "true") {
    return `${paramName}: Literal[True]`;
  }

  if (streamType === "bool") {
    return `${paramName}: bool`;
  }

  // For stream=False overload, we need to handle the tri-state
  // Check if it's both optional and nullable (OptionalNullable)
  if (isOptional && isNullable) {
    // OptionalNullable fields use UNSET as default
    // For the False overload, we use OptionalNullable type alias which already includes UNSET
    addImport("types", "OptionalNullable", true);
    return `${paramName}: OptionalNullable[Literal[False]] = UNSET`;
  }

  // Check if it's required but nullable (Nullable)
  if (!isOptional && isNullable) {
    addImport("typing", "Union");
    // Required nullable - can be False or None, but no default
    return `${paramName}: Union[Literal[False], None]`;
  }

  // Check if it's optional but not nullable (Optional)
  if (isOptional && !isNullable) {
    addImport("typing", "Union");
    // Optional non-nullable - can be False or omitted (None default)
    return `${paramName}: Union[Literal[False], None] = None`;
  }

  // Required and not nullable - must be False
  return `${paramName}: Literal[False]`;
}

function getSSEResponseTypes(operation: Operation): SSEOverloadInfo | null {
  // Check if SSE overload is enabled
  if (!operation.Extensions?.SSEOverload) {
    return null;
  }

  // Find JSON and event-stream responses from non-error status codes
  let jsonResponse: FieldDef | null = null;
  let sseResponse: FieldDef | null = null;
  let responseCount = 0;

  const jsonContentType = "application/json";
  const sseContentType = "text/event-stream";

  if (!operation.Response?.Responses) {
    return null;
  }

  for (const response of operation.Response.Responses) {
    // Skip error responses
    if (response.Error) {
      continue;
    }

    if (!response.Content) continue;

    for (const content of response.Content) {
      responseCount++;

      if (!content.Content) continue;

      if (content.ContentType === jsonContentType) {
        jsonResponse = content.Content;
      }
      if (content.ContentType === sseContentType) {
        sseResponse = content.Content;
      }
    }
  }

  // Validate we have exactly 2 non-error responses: 1 JSON and 1 event-stream
  if (responseCount !== 2 || !jsonResponse || !sseResponse) {
    return null;
  }

  let jsonResponseType: string;
  let sseResponseType: string;

  if (
    getResponseFormat() === "flat" &&
    !operation.Response.Type.ResponseEnvelope
  ) {
    // Flat responses return the matched response body directly, so overloads can
    // expose the JSON-vs-SSE body type split.
    jsonResponseType = sanitizePydanticType(jsonResponse, {
      optional: isReturnTypeOptional(
        operation.Response,
        operation.Extensions?.Pagination,
      ),
      nullable: false,
      methodReturnType: true,
    });

    // Note: sanitizePydanticType with methodReturnType: true already wraps
    // event-stream types in EventStream/EventStreamAsync.
    sseResponseType = sanitizePydanticType(sseResponse, {
      optional: false,
      nullable: false,
      methodReturnType: true,
    });
  } else {
    // Envelope response formats return the operation response wrapper for every
    // content type. Overloads must use the same envelope type as the
    // implementation or type checkers reject the overload set.
    const responseType = sanitizePydanticType(
      typeDefToFieldDef(operation.Response.Type),
      {
        optional: isReturnTypeOptional(
          operation.Response,
          operation.Extensions?.Pagination,
        ),
        nullable: false,
        methodReturnType: true,
      },
    );
    jsonResponseType = responseType;
    sseResponseType = responseType;
  }

  // Add the eventstreaming import
  addEventStreamingImport();

  return {
    jsonContentType,
    jsonResponseType: jsonResponseType,
    sseContentType,
    sseResponseType: sseResponseType,
  };
}
registerTemplateFunc("getSSEResponseTypes", getSSEResponseTypes);

function formatDefaultAcceptHeader(op: Operation): string {
  return `"${op.GetAcceptTypes().join(", ")}"`;
}
registerTemplateFunc("formatDefaultAcceptHeader", formatDefaultAcceptHeader);

function sseHasStreamKwarg(op: Operation): boolean {
  const name = sseStreamFieldName(op);
  const args = op.Arguments?.Sorted ?? [];
  return args.some((field: any) => field?.Name === name);
}
registerTemplateFunc("sseHasStreamKwarg", sseHasStreamKwarg);

function sseStreamAccessor(op: Operation): string {
  const name = sseStreamFieldName(op);
  const sanitized = sanitizeParameterName(name);
  // body-variant composition can nest the stream flag inside a wrapped body field.
  if (shouldEmitBodyVariantOverloads(op)) {
    const bodyFieldName = op?.Request?.RequestBody?.Name
      ? sanitizeFieldName(op.Request.RequestBody.Name)
      : "";
    if (op?.Request?.IsRequestBody || !bodyFieldName) {
      return `getattr(request_, "${name}", False) is True`;
    }
    return `getattr(getattr(request_, "${bodyFieldName}", None), "${name}", False) is True`;
  }
  if (sseHasStreamKwarg(op)) {
    return `${sanitized} is True`;
  }
  return `getattr(request_, "${name}", False) is True`;
}
registerTemplateFunc("sseStreamAccessor", sseStreamAccessor);

function templateAcceptHeaderValue(
  op: Operation,
  canOverrideAcceptHeader: boolean,
  sseTypes: SSEOverloadInfo | null,
): string {
  // If SSE overload is enabled, dynamically switch based on stream parameter.
  // `stream` lives in different places depending on flattening + body-variant
  // composition, so resolve the correct accessor here.
  if (sseTypes) {
    const accessor = sseStreamAccessor(op);
    return `"${sseTypes.sseContentType}" if ${accessor} else "${sseTypes.jsonContentType}"`;
  }

  // If accept header can be overridden, use the override or default to all accept types
  if (canOverrideAcceptHeader) {
    return `accept_header_override.value if accept_header_override is not None else ${formatDefaultAcceptHeader(
      op,
    )}`;
  }

  return formatDefaultAcceptHeader(op);
}
registerTemplateFunc("templateAcceptHeaderValue", templateAcceptHeaderValue);

function shouldAllowAcceptHeaderOverride(
  op: Operation,
  sseTypes: SSEOverloadInfo | null,
): boolean {
  // Accept header override is only available when:
  // 1. There are multiple accept types, AND
  // 2. SSE overload is not enabled (SSE handles accept header dynamically)
  return op.GetAcceptTypes().length > 1 && !sseTypes;
}
registerTemplateFunc(
  "shouldAllowAcceptHeaderOverride",
  shouldAllowAcceptHeaderOverride,
);

function templateRequestStreamValue(
  op: Operation,
  withRawHelpers: boolean,
): string {
  const streamingMode = `_speakeasy_response_mode == "streaming"`;

  if (getSSEResponseTypes(op)) {
    if (withRawHelpers) {
      return `${sseStreamAccessor(op)} or ${streamingMode}`;
    }
    return sseStreamAccessor(op);
  }

  if (containsStreamingResponse(op.Response)) {
    return "True";
  }

  if (withRawHelpers) {
    return streamingMode;
  }

  return "False";
}
registerTemplateFunc("templateRequestStreamValue", templateRequestStreamValue);

function templateResponseMode(op: Operation): string {
  // spec response is plain JSON / text / binary, no streaming hint.
  const buffered = `"buffered"`;

  // spec response is NDJSON, response-stream,
  // OR SSE-overload op's JSON branch at runtime.
  const rawStream = `"raw_stream"`;

  // spec response is text/event-stream
  // OR SSE-overload op's SSE branch at runtime.
  const eventStream = `"event_stream"`;

  if (getSSEResponseTypes(op)) {
    return `(${eventStream} if ${sseStreamAccessor(op)} else ${buffered})`;
  }

  if (containsEventStreamResponse(op.Response)) {
    return eventStream;
  }
  if (containsStreamingResponse(op.Response)) {
    return rawStream;
  }

  return buffered;
}
registerTemplateFunc("templateResponseMode", templateResponseMode);
