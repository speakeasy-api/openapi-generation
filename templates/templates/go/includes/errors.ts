// @ts-ignore
function sanitizeErrorType(typeDef: TypeDef): string {
  return sanitizeClass(typeDef, "usage", false);
}

// @ts-ignore
function templateDefaultError(full: boolean = true): string {
  const className = getDefaultErrorClassName();

  if (full) {
    return `${getAccessNamespace("usage", "errors")}${className}`;
  }

  return className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);

function templateErrorReturn(op: Operation, error: string): string {
  if (!op.Response.Type) {
    return `return ${error}`;
  }

  return `return nil, ${error}`;
}

registerTemplateFunc("templateErrorReturn", templateErrorReturn);

function templateDefaultErrorReturn(op: Operation, message: string): string {
  const errorsNamespace = getAccessNamespace("", "errors");
  const newError = `New${getDefaultErrorClassName()}`;

  return templateErrorReturn(
    op,
    `${errorsNamespace}${newError}(${message}, httpRes.StatusCode, string(rawBody), httpRes)`,
  );
}
registerTemplateFunc("templateDefaultErrorReturn", templateDefaultErrorReturn);

// @ts-ignore
function templateErrorStatusCodesCheck(response: ResponseDef): string {
  const errorCodes = response.GetErrorStatusCodes();

  const templateErrorCodes = (codes: string[]) =>
    `[]string{${codes.map((s) => `"${s}"`).join(",")}}`;

  if (errorCodes.includes("default")) {
    const nonErrorCodes = getNonErrorStatusCodes(response);
    return `!utils.MatchStatusCodes(${templateErrorCodes(
      nonErrorCodes,
    )}, httpRes.StatusCode)`;
  }

  return `utils.MatchStatusCodes(${templateErrorCodes(
    sanitizeStatusCodes(errorCodes),
  )}, httpRes.StatusCode)`;
}
registerTemplateFunc(
  "templateErrorStatusCodesCheck",
  templateErrorStatusCodesCheck,
);
