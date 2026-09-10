// @ts-ignore
function getErrorsLocation() {
  return getScopePath("errors");
}

registerTemplateFunc("getErrorsLocation", getErrorsLocation);

// @ts-ignore
function getErrorsNamespace(full: boolean = false) {
  return getScopeNamespace("errors", full);
}

// @ts-ignore
function templateErrorHandler(message: string, responseFormat: string) {
  const error = `\\${getErrorsNamespace(true)}\\${getDefaultErrorClassName()}`;
  if (responseFormat == "envelope-http") {
    return `throw new ${error}('${message}', $httpRequest, $httpResponse);`;
  } else {
    return `throw new ${error}('${message}', $statusCode, $httpResponse->getBody()->getContents(), $httpResponse);`;
  }
}

registerTemplateFunc("templateErrorHandler", templateErrorHandler);

//@ts-ignore
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

  return templateStatusCodeCheck(varName, sanitizeStatusCodes(errorCodes));
}

registerTemplateFunc(
  "templateErrorStatusCodesCheck",
  templateErrorStatusCodesCheck,
);

// @ts-ignore
function sanitizeErrorType(typeDef: TypeDef): string {
  return sanitizeClass(typeDef, "", false, Qualification.USAGE);
}

// @ts-ignore
function templateDefaultError(full: boolean = true): string {
  const parts = getErrorsNamespace(full).split("\\");
  const className = getDefaultErrorClassName();
  if (parts.length > 0) {
    return `${parts[parts.length - 1]}\\${className}`;
  }

  return className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);
