// @ts-ignore
function templatePaginationInputAccessor(
  op: Operation,
  input: PaginationInputs,
): string {
  const requestField = op.Request?.Field;
  if (!requestField) {
    throw new Error(
      "Operation does not have a request to read pagination input from",
    );
  }
  const inputLocation = input.In?.toString();
  if (!["requestBody", "parameters"].includes(inputLocation)) {
    throw new Error(
      `Invalid pagination input for operation "${op.ID}": "${input.Name}" in ${inputLocation}`,
    );
  }

  const defaultValue = (optional: Boolean) =>
    optional ? ` ?? ${getPaginationDefaults(op)[input.Type.toString()]}` : "";

  const isRequestOptional = requestField.Optional || requestField.Nullable;

  const wrappedRequestField = findWrappedRequestField(op);
  if (wrappedRequestField) {
    const isWrappedOptional =
      wrappedRequestField.Optional || wrappedRequestField.Nullable;
    const needsNullCheck = isRequestOptional || isWrappedOptional;

    if (needsNullCheck) {
      return (
        sanitizePrivateFieldName(requestField.Name) +
        "?." +
        sanitizeFieldName(wrappedRequestField.Name) +
        "?." +
        sanitizeFieldName(input.Name) +
        defaultValue(true)
      );
    }

    return (
      sanitizePrivateFieldName(requestField.Name) +
      templateAccessOperator(requestField) +
      sanitizeFieldName(wrappedRequestField.Name) +
      templateAccessOperator(wrappedRequestField) +
      sanitizeFieldName(input.Name) +
      defaultValue(
        isRequestOptional ||
          wrappedRequestField.Optional ||
          wrappedRequestField.Nullable,
      )
    );
  }

  return (
    sanitizePrivateFieldName(requestField.Name) +
    templateAccessOperator(requestField) +
    sanitizeFieldName(input.Name) +
    defaultValue(isRequestOptional)
  );
}
registerTemplateFunc(
  "templatePaginationInputAccessor",
  templatePaginationInputAccessor,
);

function templateNextPaginationMethodArgument(
  op: Operation,
  field: FieldDef,
  requestField: FieldDef,
): string {
  return (
    templateMethodArgumentDeclaration(field) +
    templateNextPaginationParameterValue(op, field, "parameters", requestField)
  );
}

function templateNextPaginationRequestBodyField(
  op: Operation,
  field: FieldDef,
  requestField: FieldDef,
  inputAccessField?: FieldDef,
): string {
  return (
    templateFieldDeclaration(field) +
    templateNextPaginationParameterValue(
      op,
      field,
      "requestBody",
      requestField,
      inputAccessField,
    )
  );
}

function templateNextPaginationParameterName(paramName: string): string {
  return `new${sanitizeFieldName(paramName)}`;
}

function templateNextPaginationParameterValue(
  op: Operation,
  field: FieldDef,
  inputIn: string,
  requestField: FieldDef,
  inputAccessField?: FieldDef,
): string {
  if (field == requestField || field == inputAccessField) {
    return templateNextPaginationParameterName(field.Name);
  }

  const paramName = sanitizeMethodParamName(field.Name);

  const paginationInputMap: Record<string, string> = {
    page: "newPage",
    offset: "newOffset",
    cursor: "nextCursor",
  };

  if (op.Extensions.Pagination.Outputs.Results) {
    paginationInputMap["limit"] = "limit";
  }

  for (const inputType in paginationInputMap) {
    const input = getPaginationInput(op.Extensions.Pagination, inputType);
    if (
      input?.Name == paramName &&
      (input?.In == inputIn ||
        (input?.In == "parameters" && op.Arguments.ParamFields.length == 0))
    ) {
      return paginationInputMap[inputType];
    }
  }

  if (op.Arguments.Flattening != "none") {
    return paramName;
  }

  if (inputAccessField) {
    const isRequestOptional = requestField.Optional || requestField.Nullable;
    const isInputAccessOptional =
      inputAccessField.Optional || inputAccessField.Nullable;
    const needsNullCheck = isRequestOptional || isInputAccessOptional;

    if (needsNullCheck) {
      return (
        sanitizePrivateFieldName(requestField.Name) +
        "?." +
        sanitizeFieldName(inputAccessField.Name) +
        "?." +
        sanitizeFieldName(field.Name)
      );
    }

    return (
      sanitizePrivateFieldName(requestField.Name) +
      templateAccessOperator(requestField) +
      sanitizeFieldName(inputAccessField.Name) +
      templateAccessOperator(inputAccessField) +
      sanitizeFieldName(field.Name)
    );
  }

  const isRequestOptional = requestField.Optional || requestField.Nullable;
  const accessor = isRequestOptional
    ? "?."
    : templateAccessOperator(requestField);

  return (
    sanitizePrivateFieldName(requestField.Name) +
    accessor +
    sanitizeFieldName(field.Name)
  );
}

function getNextPaginationRequestBodyFields(
  op: Operation,
  requestBodyFieldDef: FieldDef,
): [FieldDef, string[]] {
  let nextRequestBodyFields: string[] = [];

  const wrappedRequestField = findWrappedRequestField(op);
  if (wrappedRequestField) {
    const nextWrappedRequestFields = wrappedRequestField.Type.Fields.map(
      (f) => {
        return templateNextPaginationRequestBodyField(
          op,
          f,
          requestBodyFieldDef,
          wrappedRequestField,
        );
      },
    );

    if (op.Arguments.Flattening !== "none") {
      // when pagination input is wrapped inside a request field and the request
      // is flattened we return the wrapped field so it can be updated.
      return [wrappedRequestField, nextWrappedRequestFields];
    }

    nextRequestBodyFields = op.Request?.Field.Type.Fields.filter(
      (f) => f.Const?.Value == undefined,
    ).map((f) => {
      if (f == wrappedRequestField) {
        return (
          templateFieldDeclaration(wrappedRequestField) +
          `new ${sanitizeType(wrappedRequestField.Type)}
    {
        ${nextWrappedRequestFields.join(",\n        ")}
    }`
        );
      } else {
        return templateNextPaginationRequestBodyField(
          op,
          f,
          requestBodyFieldDef,
        );
      }
    });
  } else {
    nextRequestBodyFields = op.Request?.Field.Type.Fields.filter(
      (f) => f.Const?.Value == undefined,
    ).map((f) => {
      return templateNextPaginationRequestBodyField(op, f, requestBodyFieldDef);
    });
  }

  return [requestBodyFieldDef, nextRequestBodyFields];
}

// @ts-ignore
function templatePaginationCall(op: Operation): string {
  const nextMethodArguments: string[] = []; // the arguments passed to the callback

  let requestBodyFieldDef = op.Request?.Field; // the requestBody field to be updated
  let nextRequestBodyFields: string[] = []; // the updated requestBody fields

  const flattened = op.Arguments.Flattening != "none";

  const isInputInRequest =
    isPaginationInputInBody(op) || op.Arguments.ParamFields.length == 0;

  if (isInputInRequest) {
    // if the pagination input is wrapped inside a request field and the request
    // is flattened we need to update the wrapped field instead of the request.
    [requestBodyFieldDef, nextRequestBodyFields] =
      getNextPaginationRequestBodyFields(op, requestBodyFieldDef);
  }

  if (flattened) {
    op.Arguments.ParamFields.forEach((f) => {
      nextMethodArguments.push(
        templateNextPaginationMethodArgument(op, f, requestBodyFieldDef),
      );
    });

    if (op.Arguments.BodyField) {
      nextMethodArguments.push(
        templateNextPaginationMethodArgument(
          op,
          op.Arguments.BodyField,
          requestBodyFieldDef,
        ),
      );
    } else {
      op.Arguments.BodyFields.forEach((f) => {
        nextMethodArguments.push(
          templateNextPaginationMethodArgument(op, f, requestBodyFieldDef),
        );
      });
    }
  } else {
    for (const field of op.Arguments.Sorted.filter((f) => f != op.Security)) {
      nextMethodArguments.push(
        templateNextPaginationMethodArgument(op, field, requestBodyFieldDef),
      );
    }
  }

  if (op.Security) {
    nextMethodArguments.push(
      `security: security${op.Security.Optional ? "" : "!"}`,
    );
  }

  if (op.Servers) {
    nextMethodArguments.push("serverUrl: serverUrl");
  }

  // Note: Unity doesn't support retries, so we don't add retryConfig here
  // unlike the C# version

  let output = "";

  if (isInputInRequest) {
    output += `
var ${templateNextPaginationParameterName(
      requestBodyFieldDef.Name,
    )} = new ${sanitizeType(requestBodyFieldDef.Type)}
{
    ${nextRequestBodyFields.join(",\n    ")}
};
`;
  }
  output += `
return await ${sanitizeMethodName(op, true)} (
    ${nextMethodArguments.join(",\n    ")}
);`;

  return output;
}

registerTemplateFunc("templatePaginationCall", templatePaginationCall);
