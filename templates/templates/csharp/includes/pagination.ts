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

  // Pagination inputs feed arithmetic (page + 1) and comparisons (count < limit),
  // so the accessor must yield a non-null value type. OptionalNullable<T> must be
  // unwrapped by reading its raw underlying value.
  const unwrapOptionalNullable = (
    inputField: FieldDef | undefined,
    accessOperator: string,
  ): string =>
    inputField && needsOptionalNullableWrapper(inputField)
      ? `${accessOperator}ValueOrDefault`
      : "";

  const isRequestOptional = requestField.Optional || requestField.Nullable;

  const wrappedRequestField = findWrappedRequestField(op);

  // Check if pagination input exists in wrapped field vs top-level request
  const inputInWrappedField =
    wrappedRequestField?.Type?.Fields?.some((f) => f.Name === input.Name) ??
    false;

  if (inputInWrappedField) {
    const inputInTopLevel =
      requestField.Type?.Fields?.some((f) => f.Name === input.Name) ?? false;
    if (inputInTopLevel) {
      throw new Error(
        `Ambiguous pagination input for operation "${op.ID}": "${input.Name}" exists in both top-level request and wrapped field`,
      );
    }
    const inputField = wrappedRequestField.Type.Fields.find(
      (f) => f.Name === input.Name,
    );
    const wrappedAccessOp = templateAccessOperator(wrappedRequestField);
    return (
      sanitizePrivateFieldName(requestField.Name) +
      templateAccessOperator(requestField) +
      sanitizeFieldName(wrappedRequestField.Name) +
      wrappedAccessOp +
      sanitizeFieldName(input.Name) +
      unwrapOptionalNullable(inputField, wrappedAccessOp) +
      defaultValue(
        isRequestOptional ||
          wrappedRequestField.Optional ||
          wrappedRequestField.Nullable ||
          (inputField ? canFieldBeNull(inputField) : false),
      )
    );
  }

  const inputField = requestField.Type?.Fields?.find(
    (f) => f.Name === input.Name,
  );
  const requestAccessOp = templateAccessOperator(requestField);
  return (
    sanitizePrivateFieldName(requestField.Name) +
    requestAccessOp +
    sanitizeFieldName(input.Name) +
    unwrapOptionalNullable(inputField, requestAccessOp) +
    defaultValue(
      isRequestOptional || (inputField ? canFieldBeNull(inputField) : false),
    )
  );
}
registerTemplateFunc(
  "templatePaginationInputAccessor",
  templatePaginationInputAccessor,
);

// @ts-ignore
function templateCursorExtraction(op: Operation): string {
  const pagination = op.Extensions.Pagination;
  const cursorInput = getPaginationInput(pagination, "cursor");
  if (!cursorInput) {
    return "";
  }

  // Search for the cursor field in top-level request fields and nested body fields
  let cursorField: FieldDef | undefined;
  for (const field of op.Request.Field.Type.Fields) {
    if (field.Name === cursorInput.Name) {
      cursorField = field;
      break;
    }
    // Search nested body type fields
    if (field.Type?.Fields) {
      const nested = field.Type.Fields.find((f) => f.Name === cursorInput.Name);
      if (nested) {
        cursorField = nested;
        break;
      }
    }
  }

  if (!cursorField) {
    return "";
  }

  const cursorType = sanitizeType(cursorField.Type, false);
  if (cursorType === "string") {
    return `var nextCursor = nextCursorToken.Value<${cursorType}>();
    if (string.IsNullOrWhiteSpace(nextCursor))
    {
        return null;
    }`;
  } else if (canFieldBeNull(cursorField)) {
    const nullableType = `${cursorType}${
      cursorField.Optional || cursorField.Nullable ? "?" : ""
    }`;
    return `var nextCursorOrNull = nextCursorToken.Value<${nullableType}>();
    if (nextCursorOrNull == null)
    {
        return null;
    }
    ${cursorType} nextCursor = (${cursorType})nextCursorOrNull;`;
  } else {
    return `var nextCursor = nextCursorToken.Value<${cursorType}>();`;
  }
}
registerTemplateFunc("templateCursorExtraction", templateCursorExtraction);

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
  const parentType = inputAccessField
    ? inputAccessField.Type
    : requestField.Type;
  return (
    templateFieldDeclaration(field, parentType) +
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

  const paginationInputMap: Record<string, string> = {
    page: "newPage",
    offset: "newOffset",
    cursor: "nextCursor",
  };

  if (op.Extensions.Pagination.Outputs.Results && !requestField.Nullable) {
    // a null request must not materialize a limit the caller never sent
    paginationInputMap["limit"] = "limit";
  }

  for (const inputType in paginationInputMap) {
    const input = getPaginationInput(op.Extensions.Pagination, inputType);
    if (
      input?.Name == field.Name &&
      (input?.In == inputIn ||
        (input?.In == "parameters" && op.Arguments.ParamFields.length == 0))
    ) {
      return paginationInputMap[inputType];
    }
  }

  if (inputAccessField) {
    return (
      sanitizePrivateFieldName(requestField.Name) +
      templateAccessOperator(requestField) +
      sanitizeField(inputAccessField, requestField.Type) +
      templateAccessOperator(inputAccessField) +
      sanitizeField(field, inputAccessField.Type)
    );
  }

  return (
    sanitizePrivateFieldName(requestField.Name) +
    templateAccessOperator(requestField) +
    sanitizeField(field, requestField.Type)
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

  const originalRequestFieldDef = op.Request?.Field; // the original composite request field
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

  // For URL pagination, cursor inputs are embedded in the next URL and
  // removed from the method signature by the backend, so skip them.
  const isURLPagination = hasPaginationURL(op);
  const cursorInput = isURLPagination
    ? getPaginationInput(op.Extensions.Pagination, "cursor")
    : undefined;
  const cursorInputLower = cursorInput?.Name?.toLowerCase();
  const shouldSkipField = (f: FieldDef): boolean => {
    return (
      isURLPagination &&
      cursorInputLower != null &&
      (f.Name?.toLowerCase() === cursorInputLower ||
        f.OriginalName?.toLowerCase() === cursorInputLower)
    );
  };

  if (flattened) {
    op.Arguments.ParamFields.forEach((f) => {
      if (shouldSkipField(f)) return;
      // Use the original request field for param fields, not the wrapped body field
      nextMethodArguments.push(
        templateNextPaginationMethodArgument(op, f, originalRequestFieldDef),
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
        if (shouldSkipField(f)) return;
        nextMethodArguments.push(
          templateNextPaginationMethodArgument(op, f, requestBodyFieldDef),
        );
      });
    }
  } else {
    for (const field of op.Arguments.Sorted.filter((f) => f != op.Security)) {
      if (shouldSkipField(field)) continue;
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

  if (op.Extensions.Retries) {
    nextMethodArguments.push("retryConfig: retryConfig");
  }

  if (hasPaginationURL(op)) {
    nextMethodArguments.push("urlOverride: nextURL");
  }

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
