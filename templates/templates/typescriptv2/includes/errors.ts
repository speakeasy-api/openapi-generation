// @ts-ignore
function templateDefaultError(full: boolean = true): string {
  if (full) {
    return `${getAccessNamespace("errors")}${getDefaultErrorClassName()}`;
  }

  return getDefaultErrorClassName();
}

registerTemplateFunc("templateDefaultError", templateDefaultError);

function templateHttpClientErrorsTable(): string {
  const contents = [["HTTP Client Error", "Description"]];

  const httpClientErrors: [string, string][] = [
    ["RequestAbortedError", "HTTP request was aborted by the client"],
    [
      "RequestTimeoutError",
      "HTTP request timed out due to an AbortSignal signal",
    ],
    ["ConnectionError", "HTTP client was unable to make a request to a server"],
    ["InvalidRequestError", "Any input used to create a request is invalid"],
    ["UnexpectedClientError", "Unrecognised or unexpected error"],
  ];

  httpClientErrors.forEach(([error, description]) => {
    contents.push([error, description]);
  });

  return createMarkdownTable(contents);
}

registerTemplateFunc(
  "templateHttpClientErrorsTable",
  templateHttpClientErrorsTable,
);

function templateBaseErrorClassTable(): string {
  const contents = [["Property", "Type", "Description"]];

  contents.push(["`error.message`", "`string`", "Error message"]);
  if (isErrorEnvelope()) {
    contents.push([
      "`error." + getConstants().errorFields.httpMeta + ".response`",
      "`Response`",
      "HTTP response. Access to headers and more.",
    ]);
    contents.push([
      "`error." + getConstants().errorFields.httpMeta + ".request`",
      "`Request`",
      "HTTP request. Access to headers and more.",
    ]);
  } else {
    contents.push([
      "`error.statusCode`",
      "`number`",
      "HTTP response status code eg `404`",
    ]);
    contents.push(["`error.headers`", "`Headers`", "HTTP response headers"]);
    contents.push([
      "`error.body`",
      "`string`",
      "HTTP body. Can be empty string if no body is returned.",
    ]);
    contents.push(["`error.rawResponse`", "`Response`", "Raw HTTP response"]);
  }

  if (sdkHasCustomResponseErrors()) {
    contents.push([
      "`error.data$`",
      "",
      "Optional. Some errors may contain structured data. [See Error Classes](#error-classes).",
    ]);
  }

  return createMarkdownTable(contents, false);
}
registerTemplateFunc(
  "templateBaseErrorClassTable",
  templateBaseErrorClassTable,
);

function getErrorMessageAccessorString(type: TypeDef): string | undefined {
  const fields = getOrInferErrorMessageFieldDeep(type);
  if (fields) {
    return fields.map((x) => sanitizeFieldName(x.Name)).join("?.");
  }
}

registerTemplateFunc(
  "getErrorMessageAccessorString",
  getErrorMessageAccessorString,
);

function isErrorEnvelope(): boolean {
  return getResponseFormat() === "envelope-http";
}
registerTemplateFunc("isErrorEnvelope", isErrorEnvelope);

// @ts-ignore
function templateErrorGroupMarkdownLink(group: any): string {
  const typeName = sanitizeType(group.Type, false, "usage");
  return templateErrorMarkdownLink(typeName);
}
registerTemplateFunc(
  "templateErrorGroupMarkdownLink",
  templateErrorGroupMarkdownLink,
);

// @ts-ignore
function templateErrorMarkdownLink(className: string): string {
  let location = `./src/${getModelsLocation(getErrorsLocation())}/`;

  switch (className) {
    case "UnexpectedClientError":
    case "InvalidRequestError":
    case "RequestAbortedError":
    case "RequestTimeoutError":
    case "ConnectionError":
      location += `${getHttpClientErrorsFileName()}.ts`;
      break;
    default:
      location += `${sanitizeFileName(className)}.ts`;
      break;
  }

  return `[\`${className}\`](${location})`;
}
registerTemplateFunc("templateErrorMarkdownLink", templateErrorMarkdownLink);

// @ts-ignore
function templateErrorStatusCodesCheck(response: ResponseDef): string {
  addInternalImport("http", "matchStatusCode", "funcs", "typeImport");

  const errorCodes = response.GetErrorStatusCodes();

  const templateErrorCodes = (codes: string[]) =>
    `[${codes.map((s) => `"${s}"`).join(", ")}]`;

  if (errorCodes.includes("default")) {
    const nonErrorCodes = getNonErrorStatusCodes(response);
    return `(statusCode: number) => !matchStatusCode({ status: statusCode } as Response, ${templateErrorCodes(
      nonErrorCodes,
    )})`;
  }

  return `(statusCode: number) => matchStatusCode({ status: statusCode } as Response, ${templateErrorCodes(
    sanitizeStatusCodes(errorCodes),
  )})`;
}
registerTemplateFunc(
  "templateErrorStatusCodesCheck",
  templateErrorStatusCodesCheck,
);
