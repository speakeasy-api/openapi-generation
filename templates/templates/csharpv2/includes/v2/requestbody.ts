// @ts-ignore
function templatePreprocessRequestBody(op: Operation): string {
  if (!op.Request || !op.SerializationMethod) return "";

  const method = `${op.SerializationMethod}`;
  if (method === "form") {
    return templateApplyFormBody(op);
  }
  if (method === "multipart") {
    return templateApplyMultipartBody(op);
  }
  return "";
}

// Codegen-time null/required/nullable/optional policy. The emitted block
// ends with a `var serializedBody = ...;` declaration so the outer
// method.cs.stmpl wrapper
// `if (serializedBody != null) httpRequest.Content = serializedBody;`
// works unchanged. `mkSerializeCall(bodyVar)` produces the typed serialize
// expression given the local that holds the (now non-null) body.
function emitBodyDispatch(
  bodyExpr: string,
  mkSerializeCall: (bodyVar: string) => string,
  nullable: boolean,
  optional: boolean,
  jsonNullLiteral: boolean,
  valueTypeBody: boolean,
  unwrapValueType: boolean,
): string {
  const hoist = `var __body = ${bodyExpr};`;
  const guardedBody = unwrapValueType ? "__body.Value" : "__body";
  if (nullable && jsonNullLiteral) {
    return `${hoist}
            var serializedBody = __body == null
                ? RequestBodySerializer.NullJsonBody()
                : ${mkSerializeCall(guardedBody)};`;
  }
  if (optional || nullable) {
    return `${hoist}
            HttpContent? serializedBody = __body == null
                ? null
                : ${mkSerializeCall(guardedBody)};`;
  }
  if (valueTypeBody) {
    // Non-nullable value type: caller can't pass null, runtime check is dead.
    return `${hoist}
            var serializedBody = ${mkSerializeCall("__body")};`;
  }
  return `${hoist}
            if (__body == null) throw new System.ArgumentNullException(nameof(request), "request body is required");
            var serializedBody = ${mkSerializeCall("__body")};`;
}

// @ts-ignore
function templateSerializeRequestBody(op: Operation): string {
  if (!op.Request || !op.SerializationMethod) return "";
  const req = op.Request;
  const fieldName = sanitizeFieldName(
    req.RequestBody!.Name,
    sanitizeClassName(req.Field!.Type!.Name, false),
  );
  const method = `${op.SerializationMethod}`;
  const nullable = !!req.RequestBody!.Nullable;
  const optional = !!req.RequestBody!.Optional;
  const format = getDataTypeFormat(req);

  // Form / multipart: preprocess inlines build code that assigns
  // httpRequest.Content directly. Emit a no-op `serializedBody = null`
  // so the outer wrapper's `if (serializedBody != null)` is a no-op.
  if (method === "form" || method === "multipart") {
    const preprocess =
      method === "form"
        ? templateApplyFormBody(op)
        : templateApplyMultipartBody(op);
    return `${preprocess}
            HttpContent? serializedBody = null;`;
  }

  // Wrapper-vs-direct body access. The request param IS the body when its
  // type matches the body's type — no wrapper class. Otherwise the wrapper
  // exposes the body via the property whose sanitized name equals fieldName.
  const requestType = req.Field?.Type;
  const bodyType = req.RequestBody?.Type;
  const isSameType =
    !!requestType &&
    !!bodyType &&
    requestType.Name === bodyType.Name &&
    (requestType.OutputLocation || "") === (bodyType.OutputLocation || "");
  const hasWrapperField =
    !!requestType &&
    !!fieldName &&
    !isSameType &&
    Array.isArray(requestType.Fields) &&
    requestType.Fields.some(
      (f: any) =>
        sanitizeFieldName(
          f.Name,
          sanitizeClassName(requestType.Name, false),
        ) === fieldName,
    );
  const isRequestNullable =
    !!req.Field && (!!req.Field.Optional || !!req.Field.Nullable);
  const bodyExpr = hasWrapperField
    ? isRequestNullable
      ? `(request == null ? default : request.${fieldName})`
      : `request.${fieldName}`
    : `request`;

  const bodyTypeDef = hasWrapperField ? req.RequestBody!.Type! : requestType!;
  const valueTypeBody = !!bodyTypeDef && !canTypeBeNull(bodyTypeDef);

  if (method === "raw") {
    return emitBodyDispatch(
      bodyExpr,
      (b) => `RequestBodySerializer.SerializeRaw(${b})`,
      nullable,
      optional,
      false,
      valueTypeBody,
      false,
    );
  }
  if (method === "string") {
    return emitBodyDispatch(
      bodyExpr,
      (b) => `RequestBodySerializer.SerializeString(${b})`,
      nullable,
      optional,
      false,
      valueTypeBody,
      false,
    );
  }

  if (method !== "json") {
    throw new Error(
      `Unsupported request serialization method for csharpv2: ${method}`,
    );
  }

  const bodyTypeName = sanitizeType(bodyTypeDef, false);
  const call =
    format === "string"
      ? (b: string) =>
          `RequestBodySerializer.SerializeJsonBodyAsString<${bodyTypeName}>(${b})`
      : (b: string) =>
          `RequestBodySerializer.SerializeJsonBody<${bodyTypeName}>(${b}, ${jsonTypeInfoRef(
            bodyTypeDef,
          )})`;
  return emitBodyDispatch(
    bodyExpr,
    call,
    nullable,
    optional,
    true,
    valueTypeBody,
    isValueType(bodyTypeDef),
  );
}

unregisterTemplateFunc("templateSerializeRequestBody");
registerTemplateFunc(
  "templateSerializeRequestBody",
  templateSerializeRequestBody,
);
