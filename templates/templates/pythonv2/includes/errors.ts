// @ts-ignore
function templateDefaultError(
  full: boolean = true,
  methodBody: boolean = false,
): string {
  const className = getDefaultErrorClassName();
  if (full) {
    const ns = methodBody
      ? getAccessNamespaceSuffixed("errors")
      : getAccessNamespace("errors");
    return `${ns}${className}`;
  }

  return className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);

// @ts-ignore
function templateErrorMarkdownLink(className: string): string {
  const outputLocation = getErrorsLocation();
  const fileName = sanitizeFileName(className);
  return `[\`${className}\`](./${getSourcePath()}/${outputLocation}/${fileName}.py)`;
}

registerTemplateFunc("templateErrorMarkdownLink", templateErrorMarkdownLink);

function templateBaseErrorClassTable(): string {
  const contents = [["Property", "Type", "Description"]];

  contents.push(["`err.message`", "`str`", "Error message"]);
  contents.push([
    "`err.status_code`",
    "`int`",
    "HTTP response status code eg `404`",
  ]);
  contents.push(["`err.headers`", "`httpx.Headers`", "HTTP response headers"]);
  contents.push([
    "`err.body`",
    "`str`",
    "HTTP body. Can be empty string if no body is returned.",
  ]);
  contents.push([
    "`err.raw_response`",
    "`httpx.Response`",
    "Raw HTTP response",
  ]);

  if (sdkHasCustomResponseErrors()) {
    contents.push([
      "`err.data`",
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

// @ts-ignore
function templateErrorGroupMarkdownLink(group: any): string {
  return templateErrorMarkdownLink(sanitizeClassName(group.Type.Name));
}
registerTemplateFunc(
  "templateErrorGroupMarkdownLink",
  templateErrorGroupMarkdownLink,
);

// @ts-ignore
function getNoResponseErrorClassName(): string {
  return "NoResponseError";
}

registerTemplateFunc(
  "getNoResponseErrorClassName",
  getNoResponseErrorClassName,
);

// @ts-ignore
function getNoResponseErrorFileName(): string {
  return "no_response_error";
}

registerTemplateFunc("getNoResponseErrorFileName", getNoResponseErrorFileName);

// @ts-ignore
function templateNoResponseError(full: boolean = true): string {
  const className = getNoResponseErrorClassName();
  if (full) {
    return `${getAccessNamespace("errors")}${className}`;
  }

  return className;
}

registerTemplateFunc("templateNoResponseError", templateNoResponseError);

function getErrorMessageAccessorString(
  type: TypeDef,
  identifier: string,
): string | undefined {
  const fields = getOrInferErrorMessageFieldDeep(type);
  if (fields?.length) {
    let result = identifier;
    let optionalParts = [];
    for (const [i, field] of fields.entries()) {
      result += `.${sanitizeFieldName(field.Name)}`;
      if ((field.Optional || field.Nullable) && i < fields.length - 1) {
        optionalParts.push(result);
      }
    }
    if (optionalParts.length > 0) {
      const safeAccess = optionalParts.join(" and ");
      return `str(${result}) if ${safeAccess} else fallback`;
    }

    return `str(${result}) or fallback`;
  }
}

registerTemplateFunc(
  "getErrorMessageAccessorString",
  getErrorMessageAccessorString,
);

// @ts-ignore
function templateErrorStatusCodesCheck(
  response: ResponseDef,
  operand: string,
): string {
  const errorCodes = response.GetErrorStatusCodes();

  const templateErrorCodes = (codes: string[]) =>
    `[${codes.map((s) => `"${s}"`).join(",")}]`;

  if (errorCodes.includes("default")) {
    const nonErrorCodes = getNonErrorStatusCodes(response);
    return `not utils_.match_status_codes(${templateErrorCodes(
      nonErrorCodes,
    )}, ${operand})`;
  }

  return `utils_.match_status_codes(${templateErrorCodes(
    sanitizeStatusCodes(errorCodes),
  )}, ${operand})`;
}
registerTemplateFunc(
  "templateErrorStatusCodesCheck",
  templateErrorStatusCodesCheck,
);

// @ts-ignore
function errorSchemaValidationEnabled(): boolean {
  return context.Global.Config.ErrorSchemaValidation !== false;
}
registerTemplateFunc(
  "errorSchemaValidationEnabled",
  errorSchemaValidationEnabled,
);

type ResponseSchemaValidationMode =
  // `true`, the default: enforce validation, raising on mismatch.
  | "strict"
  // `"lenient"`: construct without raising, defaulting missing required fields
  // and keeping typed union variants.
  | "lenient"
  // `false`: skip validation, return the raw decoded JSON.
  | "disabled";

function responseSchemaValidationMode(): ResponseSchemaValidationMode {
  const value = context.Global.Config.ResponseSchemaValidation;
  if (value === false) {
    return "disabled";
  }
  if (value === "lenient") {
    return "lenient";
  }
  return "strict";
}

// @ts-ignore
function responseSchemaValidationStrict(): boolean {
  return responseSchemaValidationMode() === "strict";
}
registerTemplateFunc(
  "responseSchemaValidationStrict",
  responseSchemaValidationStrict,
);

// @ts-ignore
function responseSchemaValidationLenient(): boolean {
  return responseSchemaValidationMode() === "lenient";
}
registerTemplateFunc(
  "responseSchemaValidationLenient",
  responseSchemaValidationLenient,
);
