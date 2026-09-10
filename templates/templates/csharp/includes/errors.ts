// @ts-ignore
function getErrorsLocation() {
  return getScopePath("errors") || getScopePath("models");
}

registerTemplateFunc("getErrorsLocation", getErrorsLocation);

// @ts-ignore
function templateErrorStatusCodesCheck(
  varName: string,
  response: ResponseDef,
): string {
  const errorCodes = response.GetErrorStatusCodes();

  if (errorCodes.includes("default")) {
    return templateStatusCodeCheck(
      varName,
      getNonErrorStatusCodes(response),
      true,
    );
  }

  return templateStatusCodeCheck(varName, errorCodes);
}

registerTemplateFunc(
  "templateErrorStatusCodesCheck",
  templateErrorStatusCodesCheck,
);

// @ts-ignore
function templateDefaultError(full: boolean = true): string {
  const className = getDefaultErrorClassName();

  if (usingGlobalImports()) {
    return className;
  }

  const useFull = full || hasNamespaceConflict();
  const namespace = getScopeNamespace("errors", useFull);
  return namespace ? `${namespace}.${className}` : className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);

// @ts-ignore
function sanitizeErrorType(typeDef: TypeDef): string {
  return sanitizeClass(typeDef, "", !usingGlobalImports());
}

// Returns the C# expression that reads the inferred error-message string
// off the `payload` local variable, or "" when no message field exists.
//
// The error message may be deeply nested (e.g. payload.Error.Message). Under
// presenceAwareJsonSerialization any hop in that chain can be a OptionalNullable<T>
// and must be unwrapped via Utilities.UnwrapValue() before navigating further.
function getErrorMessageAccessorString(
  type: TypeDef,
  className: string,
  usageLocation: string = "",
): string {
  const fields = getOrInferErrorMessageFieldDeep(type);
  if (!fields?.length) {
    return "";
  }
  let expr = "payload";
  let nullable = false;
  fields.forEach((f, i) => {
    const name = sanitizeFieldName(f.Name, i == 0 ? className : "");
    const access = `${expr}${nullable ? "?." : "."}${name}`;
    if (needsOptionalNullableWrapper(f)) {
      const inner =
        i == fields.length - 1
          ? "string"
          : sanitizeType(f.Type, false, usageLocation);
      expr = `((${inner}?)Utilities.UnwrapValue(${access}))`;
      nullable = true;
    } else {
      expr = access;
      // Although not strictly correct the flag-off branch preserves the legacy
      // blanket null-conditional `?.` to avoid churn for existing customers.
      nullable = context.Global.Config.PresenceAwareJSONSerialization
        ? f.Optional || f.Nullable
        : true;
    }
  });
  return expr;
}

registerTemplateFunc(
  "getErrorMessageAccessorString",
  getErrorMessageAccessorString,
);

// @ts-ignore
function isReservedErrorField(fieldName: string): boolean {
  return [
    // Exception
    "Data",
    "Message",
    "InnerException",
    // BaseException
    ...(getResponseFormat() == "envelope-http"
      ? ["Request", "Response"]
      : ["StatusCode", "Headers", "ContentType", "RawResponse"]),
    "Body",
    // Custom Error
    "Payload",
    ...(schemaValidationLenient() ? ["DeserializationException"] : []),
  ].includes(fieldName);
}
registerTemplateFunc("isReservedErrorField", isReservedErrorField);

// @ts-ignore
function templateErrorMarkdownLink(className: string): string {
  const sdkDir = getSDKTopLevelFolder();
  const errorsDir = getScopePath(usingGlobalImports() ? "models" : "errors");
  const fileName = `${sanitizeFileName(className)}.cs`;

  return `[\`${className}\`](./${sdkDir}/${errorsDir}/${fileName})`;
}

registerTemplateFunc("templateErrorMarkdownLink", templateErrorMarkdownLink);

// @ts-ignore
function templateErrorGroupMarkdownLink(group: any): string {
  return templateErrorMarkdownLink(sanitizeClassName(group.Type.Name));
}
registerTemplateFunc(
  "templateErrorGroupMarkdownLink",
  templateErrorGroupMarkdownLink,
);

// @ts-ignore
function schemaValidationLenient(): boolean {
  return context.Global.Config.SchemaValidation === "lenient";
}
registerTemplateFunc("schemaValidationLenient", schemaValidationLenient);

// @ts-ignore
function schemaValidationStrict(): boolean {
  return !schemaValidationLenient();
}
registerTemplateFunc("schemaValidationStrict", schemaValidationStrict);
