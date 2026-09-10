// @ts-ignore
function getErrorsLocation() {
  const imports = context.Global.Config.Imports;
  return imports.GetErrorsPath();
}

registerTemplateFunc("getErrorsLocation", getErrorsLocation);
// @ts-ignore
function getErrorsNamespace() {
  return getScopeNamespace("errors");
}
registerTemplateFunc("getErrorsNamespace", getErrorsNamespace);

function getErrorsScope() {
  const imports = context.Global.Config.Imports;
  return imports.GetErrorsScope();
}

// @ts-ignore
function templateErrorHandler(message: string, responseFormat: string) {
  const error = `${getErrorsNamespace()}::${getDefaultErrorClassName()}`;
  if (responseFormat == "envelope-http") {
    return `raise ${error}.new(status_code: http_response.status, request: ${sorbetMust(
      "http_request",
    )}, response: http_response), '${message}'`;
  } else {
    return `raise ${error}.new(status_code: http_response.status, body: http_response.env.response_body, raw_response: http_response), '${message}'`;
  }
}

registerTemplateFunc("templateErrorHandler", templateErrorHandler);

// @ts-ignore
function templateDefaultError(): string {
  const parts = getErrorsNamespace().split("::");
  const className = getDefaultErrorClassName();
  if (parts.length > 0) {
    return `${parts[parts.length - 1]}::${className}`;
  }

  return className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);
