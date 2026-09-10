// @ts-ignore
function sanitizeErrorType(typeDef: TypeDef): string {
  return `${getErrorsLocation()}/${sanitizeClass(typeDef, "errors", false)}`;
}

// @ts-ignore
function templateDefaultError(full: boolean = true): string {
  const className = getDefaultErrorClassName();
  if (full) {
    return `${getErrorsLocation()}/${className}`;
  }

  return className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);

// @ts-ignore
function getErrorsLocation(): string {
  return getScopePath("errors");
}

registerTemplateFunc("getErrorsLocation", getErrorsLocation);

function getErrorsPackage(): string {
  return getModelNamespace(true, "errors", "", "");
}
registerTemplateFunc("getErrorsPackage", getErrorsPackage);

function getBaseErrorFullClassName(): string {
  return `${getErrorsPackage()}.${getBaseErrorClassName()}`;
}
registerTemplateFunc("getBaseErrorFullClassName", getBaseErrorFullClassName);

function templateBaseErrorClassTable(): string {
  const contents = [["Method", "Type", "Description"]];

  contents.push(["`message()`", "`String`", "Error message"]);
  contents.push(["`code()`", "`int`", "HTTP response status code eg `404`"]);
  contents.push([
    "`headers`",
    "`Map<String, List<String>>`",
    "HTTP response headers",
  ]);
  contents.push([
    "`body()`",
    "`byte[]`",
    "HTTP body as a byte array. Can be empty array if no body is returned.",
  ]);
  contents.push([
    "`bodyAsString()`",
    "`String`",
    "HTTP body as a UTF-8 string. Can be empty string if no body is returned.",
  ]);
  contents.push([
    "`rawResponse()`",
    "`HttpResponse<?>`",
    "Raw HTTP response (body already read and not available for re-read)",
  ]);

  if (sdkHasCustomResponseErrors()) {
    // TODO There's not much point in abstracting `data()` method
    // across classes when the type can vary and optionality can vary (required=false on response)
    //    contents.push([
    //      "`data()`",
    //      "",
    //      "Optional. Some errors may contain structured data. [See Error Classes](#error-classes).",
    //    ]);
  }

  return createMarkdownTable(contents, false);
}
registerTemplateFunc(
  "templateBaseErrorClassTable",
  templateBaseErrorClassTable,
);

function getSourcePath(): string {
  return "src/main/java";
}

function linkToErrorTypeFromReadMe(type: string): string {
  const outputLocation = getModelsLocation(getErrorsLocation());
  const sourcePath = getSourcePath();
  const prefix = `./${sourcePath}/${outputLocation}`;
  return `${prefix}/${type}.java`;
}
registerTemplateFunc("linkToErrorTypeFromReadMe", linkToErrorTypeFromReadMe);

// @ts-ignore
function templateErrorGroupMarkdownLink(group: any): string {
  const typeName = sanitizeType(group.Type);
  const location = linkToErrorTypeFromReadMe(typeName);
  return `[\`${typeName}\`](${location})`;
}
registerTemplateFunc(
  "templateErrorGroupMarkdownLink",
  templateErrorGroupMarkdownLink,
);

// @ts-ignore
function templateErrorStatusCodesCheck(
  response: ResponseDef,
  varName: string = "httpRes",
): string {
  const errorCodes = response.GetErrorStatusCodes();

  function templateCheck(codes: string[], inverted: boolean = false) {
    if (codes.length === 0) {
      return inverted ? "true" : "false";
    }
    const codeParams = codes.map((c) => `"${c}"`).join(", ");
    return `${
      inverted ? "!" : ""
    }${javaImportUtils()}.statusCodeMatches(${varName}.statusCode(), ${codeParams})`;
  }

  if (errorCodes.includes("default")) {
    return templateCheck(getNonErrorStatusCodes(response), true);
  }

  return templateCheck(sanitizeStatusCodes(errorCodes));
}
registerTemplateFunc(
  "templateErrorStatusCodesCheck",
  templateErrorStatusCodesCheck,
);
