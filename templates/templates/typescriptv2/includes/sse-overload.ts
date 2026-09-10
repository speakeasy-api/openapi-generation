interface SSEOverloadInfo {
  jsonContentType: string;
  sseContentType: string;
  jsonReturnType: string;
  sseReturnType: string;
}

const sseJSONContentType = "application/json";
const sseStreamContentType = "text/event-stream";

// Finds the JSON and event-stream response pair of an SSE-overload
// operation. Pure detection with no import registration: callers that only
// need to know whether an operation has SSE overloads (e.g. params-object
// state computation, which iterates every operation while an unrelated file
// renders) must not pollute the rendering file's import table.
function findSSEOverloadResponses(
  op: Operation,
): { jsonResponse: FieldDef; sseResponse: FieldDef } | null {
  if (!op.Extensions?.SSEOverload) {
    return null;
  }

  // Find JSON and event-stream responses from non-error status codes
  let jsonResponse: FieldDef | null = null;
  let sseResponse: FieldDef | null = null;
  let responseCount = 0;

  if (!op.Response?.Responses) {
    return null;
  }

  for (const response of op.Response.Responses) {
    // Skip error responses
    if (response.Error) {
      continue;
    }

    if (!response.Content) continue;

    for (const content of response.Content) {
      responseCount++;

      if (!content.Content) continue;

      if (content.ContentType === sseJSONContentType) {
        jsonResponse = content.Content;
      }
      if (content.ContentType === sseStreamContentType) {
        sseResponse = content.Content;
      }
    }
  }

  // Validate we have exactly 2 non-error responses: 1 JSON and 1 event-stream
  if (responseCount !== 2 || !jsonResponse || !sseResponse) {
    return null;
  }

  return { jsonResponse, sseResponse };
}

function opHasSSEOverload(op: Operation): boolean {
  return findSSEOverloadResponses(op) !== null;
}

function getSSEResponseTypes(
  opState: OperationStateTS,
): SSEOverloadInfo | null {
  const { usageLocation, operation: op, flatOptions } = opState;

  const sseResponses = findSSEOverloadResponses(op);
  if (!sseResponses) {
    return null;
  }
  const { jsonResponse, sseResponse } = sseResponses;

  const jsonContentType = sseJSONContentType;
  const sseContentType = sseStreamContentType;

  let jsonReturnType: string;
  let sseReturnType: string;

  if (
    opState.responseFormat !== "flat" ||
    Boolean(op.Response?.Type?.ResponseEnvelope)
  ) {
    // Envelope response formats return the operation response wrapper for every
    // content type. Overloads must use that same wrapper type so the overload
    // signatures remain compatible with the implementation signature.
    opState.inboundResponse.importOutputTypes();
    jsonReturnType = opState.inboundResponse.outputType;
    sseReturnType = opState.inboundResponse.outputType;
  } else {
    const inboundJSONResponse = resolveInbound({
      usageLocation,
      typeDef: jsonResponse.Type,
      rootTypeDef: op.OwningSDK.Type,
      optional: flatOptions.returnTypeOptional,
      nullable: false,
    });
    inboundJSONResponse.importOutputTypes();

    const inboundSSEResponse = resolveInbound({
      usageLocation,
      typeDef: sseResponse.Type,
      rootTypeDef: op.OwningSDK.Type,
      optional: flatOptions.returnTypeOptional,
      nullable: false,
    });
    inboundSSEResponse.importOutputTypes();

    jsonReturnType = inboundJSONResponse.outputType;
    sseReturnType = inboundSSEResponse.outputType;
  }

  return {
    jsonContentType,
    jsonReturnType,
    sseContentType,
    sseReturnType,
  };
}

registerTemplateFunc("getSSEResponseTypes", getSSEResponseTypes);

function findTopLevelSSEStreamField(requestWrapper: FieldDef): FieldDef | null {
  if (!hasAnnotation(requestWrapper, "requestWrapper")) {
    return null;
  }

  return (
    requestWrapper.Type?.Fields?.find(
      (field) => field?.Name && sanitizeFieldName(field.Name) === "stream",
    ) ?? null
  );
}

function getSSEStreamFieldOverride(
  field: FieldDef,
  streamProperty: string,
): string {
  if (hasAnnotation(field, "requestWrapper")) {
    const streamField = findTopLevelSSEStreamField(field);
    if (streamField) {
      return `{ ${streamProperty} }`;
    }

    const requestBodyField = findRequestBodyField(field);
    if (requestBodyField) {
      const requestBodyFieldName = sanitizeFieldName(requestBodyField.Name);
      return `{ ${requestBodyFieldName}: { ${streamProperty} } }`;
    }
  }

  return `{ ${streamProperty} }`;
}

registerTemplateFunc("getSSEStreamFieldOverride", getSSEStreamFieldOverride);

function getSSEStreamAccessor(op: Operation, inputVar: string): string {
  const streamFieldName = sseStreamFieldName(op);
  // Operations without a wrapper request field expose stream directly as a
  // call-site argument, so the accessor is just the bare field name.
  if (!op.Request?.Field) {
    return streamFieldName;
  }

  if (hasAnnotation(op.Request.Field, "requestWrapper")) {
    const streamField = findTopLevelSSEStreamField(op.Request.Field);
    if (streamField) {
      return `${inputVar}?.${sanitizeFieldName(streamField.Name)}`;
    }

    const requestBodyField = findRequestBodyField(op.Request.Field);
    if (requestBodyField) {
      const requestBodyFieldName = sanitizeFieldName(requestBodyField.Name);
      return `${inputVar}?.${requestBodyFieldName}?.${streamFieldName}`;
    }

    return `${inputVar}?.${streamFieldName}`;
  }

  if (hasAnnotation(op.Request.Field, "request")) {
    return `${inputVar}?.${streamFieldName}`;
  }

  return streamFieldName;
}

registerTemplateFunc("getSSEStreamAccessor", getSSEStreamAccessor);
