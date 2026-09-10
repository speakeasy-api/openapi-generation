// @ts-ignore
function templateCmsComment(
  sdkComments: SimpleCommentDef,
  title: string,
  indent: number,
  keepTitle: boolean = false,
): string {
  let lines = [];

  let firstLineParts: string[] = [];

  if (title !== null) {
    if (/[.]/.test(title)) {
      title = title.replace(/^[^.]*[.]/, "");
    }

    if (title && !/\[/.test(title)) {
      firstLineParts.push(title);
    }
  }

  if (
    (!keepTitle || title === null) &&
    !sdkComments.Description &&
    !sdkComments.Summary
  ) {
    return "";
  }

  let descriptionLines = [];
  if (sdkComments.Description) {
    descriptionLines = sanitizeComments(sdkComments.Description).split("\n");
  }

  if (sdkComments.Summary) {
    firstLineParts.push(sanitizeComments(sdkComments.Summary));
  } else if (descriptionLines.length > 0) {
    firstLineParts.push(descriptionLines.shift());
  }

  let firstLine: string;
  if (firstLineParts.length > 1) {
    const words = firstLineParts[1].split(" ");
    // If the first couple words are the title, we don't need to add the title.
    if (
      (words.length > 0 && words[0] === title) ||
      (words.length > 1 && words[1] === title)
    ) {
      firstLine = firstLineParts[1];
    } else {
      firstLine = firstLineParts.join(" - ");
    }
  } else if (firstLineParts.length == 1) {
    firstLine = firstLineParts[0];
  }

  if (firstLine) {
    lines.push(...sanitizeComments(firstLine).split("\n"));
  }

  if (descriptionLines.length > 0) {
    lines.push(...descriptionLines);
  }

  if (lines.length == 0) {
    return "";
  }

  lines = lines.map((line) => `// ${line}`);

  return indentLines(lines, indent) + "\n";
}

// @ts-ignore
function templateBuiltinComment(
  docgroup: string | null,
  title: string,
  description: string,
  indent: number,
): string {
  const resolved = resolveComment(
    "builtin",
    docgroup,
    "",
    title,
    "",
    title + " " + description,
  );
  return templateCmsComment(resolved, title, indent);
}

registerTemplateFunc("templateBuiltinComment", templateBuiltinComment);

// @ts-ignore
function templateUsageMethodOperationParameters(
  operation: Operation,
  indent,
): string {
  const methodParams = [];

  const initialIndent = new Array(1 + indent * 4).join(" ");
  if (operationParametersFlattened(operation) && operation.Request) {
    for (const field of sortMethodParams(operation.Request.Field.Type.Fields)) {
      methodParams.push(
        `${initialIndent}${sanitizePrivateFieldName(
          field.Name,
        )} := ${templateValue(field)}\n`,
      );
    }
  }

  const requestParameters =
    methodParams.length > 0 ? methodParams.join("") : "";

  const operationSecurity = operation.Security
    ? `${initialIndent}operationSecurity := ${templateSecurityUsage(
        operation.Security.Type,
        indent + 1,
        true,
        false,
      ).trimStart()}\n`
    : "";

  return `${requestParameters}${operationSecurity}`;
}

registerTemplateFunc(
  "templateUsageMethodOperationParameters",
  templateUsageMethodOperationParameters,
);

// @ts-ignore
function templateUsageMethodParameters(
  operation: Operation,
  indent: number,
): string {
  if (operationParametersFlattened(operation)) {
    let methodParams = "";

    if (operation.Security) {
      methodParams += ", operationSecurity";
    }

    if (operation.Request) {
      for (const field of sortMethodParams(
        operation.Request.Field.Type.Fields,
      )) {
        methodParams += `, ${sanitizePrivateFieldName(field.Name)}`;
      }
    }

    return methodParams;
  } else {
    let methodParams = "";

    if (operation.Request) {
      methodParams += `, ${templateModelUsage(
        operation.Request.Field,
        indent,
      ).trimStart()}`;
    }

    if (operation.Security) {
      methodParams += ", operationSecurity";
    }

    return methodParams;
  }
}

registerTemplateFunc(
  "templateUsageMethodParameters",
  templateUsageMethodParameters,
);

// @ts-ignore
function templateZero(
  typeDef: TypeDef,
  optional: boolean,
  scope: string,
): string {
  if (optional) {
    return "nil";
  }

  switch (typeDef.Type.toString()) {
    case "string":
      return '""';
    case "date":
      return "types.Date{}";
    case "date-time":
      return "time.Time{}";
    case "float32":
      return "0.0";
    case "integer":
      return "0";
    case "int32":
      return "0";
    case "bigint":
      return "big.NewInt(0)";
    case "decimal":
      return `new(decimal.Big).SetFloat64(0.0)`;
    case "number":
      return "0.0";
    case "boolean":
      return "false";
    case "any":
      return "nil";
    case "enum":
      return `${sanitizeType(typeDef, false, scope)}(${templateZero(
        typeDef.Enum.Type,
        false,
        scope,
      )})`;
    case "response":
      return "nil";
  }
  return sanitizeType(typeDef, false, scope) + "{}";
}

registerTemplateFunc("templateZero", templateZero);

// @ts-ignore
function templateCodeRegion(
  codeRegionsEnabled: boolean,
  regionName: string,
): string {
  if (!codeRegionsEnabled) return "";
  return `// #region ${regionName}
// #endregion ${regionName}`;
}
registerTemplateFunc("templateCodeRegion", templateCodeRegion);
