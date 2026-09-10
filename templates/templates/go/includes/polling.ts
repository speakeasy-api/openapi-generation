/**
 * Templates polling success criteria response error assertions logic, which
 * runs before the response error is returned. This templated logic will return
 * early without error (success) if the polling success criteria is dependent on
 * error status codes and those assertions match. The later success criteria
 * assertions are always still templated after this to cover non-error status
 * code success criteria such as explicitly ignored response error codes.
 */
function templatePollingSuccessCriteriaErrorAssertions(
  assertions: Assertions,
  operation: Operation,
): string {
  if (!assertions || assertions.length === 0) {
    return "";
  }

  // For now, status code must be the first condition if used.
  if (assertions[0].TargetType.toString() !== "statusCode") {
    return "";
  }

  const statusCodeAssertion = assertions[0];

  // For now, only equal and regex are supported to prevent future enhancement
  // edge cases.
  switch (statusCodeAssertion.Type.toString()) {
    case "equal":
    case "regex":
      break;
    default:
      return "";
  }

  const operationErrorStatusCodes = operation.Response.GetErrorStatusCodes();
  const statusCode = statusCodeAssertion.Value as string;

  // Check if statusCode matches any error status codes, including wildcard patterns like "4XX" or "5XX"
  const matchesErrorStatusCode = operationErrorStatusCodes.some((errorCode) => {
    switch (statusCodeAssertion.Type.toString()) {
      case "equal":
        if (errorCode === statusCode) {
          return true;
        }

        // Check for wildcard patterns like "4XX" or "5XX"
        if (/^[45]XX$/.test(errorCode)) {
          const prefix = errorCode.charAt(0);

          return statusCode.startsWith(prefix);
        }

        break;
      case "regex":
        const statusCodeRegex = new RegExp(statusCode);

        if (statusCodeRegex.test(errorCode)) {
          return true;
        }

        // Check for wildcard patterns like "4XX" or "5XX"
        if (/^[45]XX$/.test(errorCode)) {
          const prefix = errorCode.charAt(0);

          // Return true if statusCodeRegex matches any prefixed status code.
          // For example, if prefix is "4", check if regex matches "400", "401", ..., "499".
          for (let i = 0; i <= 99; i++) {
            const testStatusCode = `${prefix}${i.toString().padStart(2, "0")}`;

            if (statusCodeRegex.test(testStatusCode)) {
              return true;
            }
          }
        }

        break;
    }

    return false;
  });

  if (!matchesErrorStatusCode) {
    return "";
  }

  const statusCodeAssertionExpression =
    templatePollingStatusCodeErrorAssertionExpression(
      statusCodeAssertion,
      "apiErr",
    );

  if (!statusCodeAssertionExpression) {
    return "";
  }

  addImport("apierrors");
  addImport("errors");

  return [
    ``,
    ``,
    `var apiErr *${templateDefaultError()}`,
    ``,
    `if err != nil && errors.As(err, &apiErr) && ${statusCodeAssertionExpression} {`,
    `return res, nil`,
    `}`,
  ].join("\n");
}

registerTemplateFunc(
  "templatePollingSuccessCriteriaErrorAssertions",
  templatePollingSuccessCriteriaErrorAssertions,
);

/**
 * Templates polling failure criteria assertions, which are logically AND
 * criterion, and therefore must all be satisfied for the polling to be
 * considered failed and immediately return a FailureCriteriaError.
 */
function templatePollingFailureCriteriaAssertions(
  assertions: Assertions,
  operation: Operation,
): string {
  if (assertions.length === 0) {
    return "";
  }

  addImport("strings");

  const result: string[] = [
    `failureCriteriaMessage := make([]string, 0, ${assertions.length})`,
    `failureCriteriaMet := true`,
  ];

  for (const assertion of assertions) {
    switch (assertion.TargetType.toString()) {
      case "statusCode":
        const statusCodeAssertion =
          templatePollingFailureStatusCodeAssertion(assertion);
        if (statusCodeAssertion) {
          result.push(...statusCodeAssertion);
        }
        break;
      case "responseBody":
        const responseBodyAssertion =
          templatePollingFailureResponseBodyAssertion(assertion, operation);
        if (responseBodyAssertion) {
          result.push(...responseBodyAssertion);
        }
        break;
    }
  }

  result.push(
    ``,
    `if failureCriteriaMet {`,
    `return res, &polling.FailureCriteriaError{Message: strings.Join(failureCriteriaMessage, " and ")}`,
    `}`,
  );

  return result.join("\n");
}

registerTemplateFunc(
  "templatePollingFailureCriteriaAssertions",
  templatePollingFailureCriteriaAssertions,
);

/**
 * Templates polling success criteria assertions, which are logically AND
 * criterion, and therefore must all be satisfied for the polling to be
 * considered successful and immediately return.
 */
function templatePollingSuccessCriteriaAssertions(
  assertions: Assertions,
  operation: Operation,
): string {
  if (assertions.length === 0) {
    return "";
  }

  const result: string[] = [`successCriteriaMet := true`];

  for (const assertion of assertions) {
    switch (assertion.TargetType.toString()) {
      case "statusCode":
        const statusCodeAssertion =
          templatePollingSuccessStatusCodeAssertion(assertion);
        if (statusCodeAssertion) {
          result.push(...statusCodeAssertion);
        }
        break;
      case "responseBody":
        const responseBodyAssertion =
          templatePollingSuccessResponseBodyAssertion(assertion, operation);
        if (responseBodyAssertion) {
          result.push(...responseBodyAssertion);
        }
        break;
    }
  }

  result.push(``, `if successCriteriaMet {`, `return res, nil`, `}`);

  return result.join("\n");
}

registerTemplateFunc(
  "templatePollingSuccessCriteriaAssertions",
  templatePollingSuccessCriteriaAssertions,
);

/**
 *
 */
function templatePollingFailureStatusCodeAssertion(
  assertion: Assertion,
): string[] {
  const statusCode = assertion.Value as string;
  const statusCodeAssertionExpression =
    templatePollingStatusCodeAssertionExpression(assertion);

  if (!statusCodeAssertionExpression) {
    return [];
  }

  let assertionTypeMessage = "";

  switch (assertion.Type.toString()) {
    case "equal":
      assertionTypeMessage = "was";
      break;
    case "notEqual":
      assertionTypeMessage = "was not";
      break;
    case "regex":
      assertionTypeMessage = "matched regular expression";
      break;
    default:
      throw new Error(`Unsupported assertion type: ${assertion.Type}`);
  }

  const message = `HTTP status code ${assertionTypeMessage} ${statusCode}`;

  return [
    ``,
    `if failureCriteriaMet {`,
    `  failureCriteriaMet = ${statusCodeAssertionExpression}`,
    `  failureCriteriaMessage = append(failureCriteriaMessage, "${message}")`,
    `}`,
  ];
}

/**
 *
 */
function templatePollingFailureResponseBodyAssertion(
  assertion: Assertion,
  operation: Operation,
): string[] {
  const responseBodyAssertion = assertion.Value as ResponseBodyAssertion;
  const responseBodyFieldDef = responseBodyAssertion.Content.Content;
  const responseFormat = getResponseFormat();
  const assertionTargetFieldDef = !isFieldDef(assertion.Target)
    ? typeDefToFieldDef(assertion.Target)
    : (assertion.Target as FieldDef);
  let responseBodyAccessor = "res";

  switch (responseFormat) {
    case "envelope":
    case "envelope-http":
      responseBodyAccessor += `${fieldAccessor(responseBodyFieldDef)}`;
      break;
    case "flat":
      if (operation.Response.Type?.ResponseEnvelope) {
        const resultField = getResultField(operation);

        if (resultField) {
          responseBodyAccessor += `${fieldAccessor(resultField)}`;
        }
      }
      break;
    default:
      throw new Error(`Unknown response format: ${responseFormat}`);
  }

  let { path: accessor } = templateJSONPointerPath(
    responseBodyFieldDef,
    responseBodyAccessor,
    responseBodyAssertion.Path,
    (partIdx, parent, currentPath) => {
      return;
    },
    {},
  );

  let assertionValue = undefined;

  if (responseBodyAssertion.Value) {
    assertionValue = getExampleValue(responseBodyAssertion.Value, {});
  }

  if (
    (!responseBodyAssertion.Path || responseBodyAssertion.Path === "/") &&
    assertion.Type.toString() === "notEqual" &&
    assertionValue === undefined
  ) {
    if (assertionTargetFieldDef.Nullable) {
      return [
        ``,
        `if failureCriteriaMet {`,
        `  failureCriteriaMet = ${accessor} != nil`,
        `}`,
      ];
    }

    switch (assertionTargetFieldDef.Type.Type.toString()) {
      case "array":
      case "map":
      case "string":
        return [
          ``,
          `if failureCriteriaMet {`,
          `  failureCriteriaMet = len(${accessor}) > 0`,
          `}`,
        ];
    }

    return [];
  }

  const supportedTypes = [
    "boolean",
    "enum",
    "float",
    "float32",
    "integer",
    "int32",
    "number",
    "string",
  ];

  if (!supportedTypes.includes(assertionTargetFieldDef.Type.Type.toString())) {
    throw new Error(
      `Unsupported polling success criteria response body assertion type: ${assertionTargetFieldDef.Type.Type}`,
    );
  }

  const result: string[] = [];

  if (assertionTargetFieldDef.Optional || assertionTargetFieldDef.Nullable) {
    result.push(
      ``,
      `if failureCriteriaMet {`,
      `  failureCriteriaMet = ${accessor} != nil`,
      `}`,
    );
    accessor = `*${accessor}`;
  }

  let assertionTypeCheck = "";
  let assertionTypeMessage = "";

  switch (assertion.Type.toString()) {
    case "equal":
      assertionTypeCheck = `${accessor} == ${assertionValue}`;
      assertionTypeMessage = "was";
      break;
    case "notEqual":
      assertionTypeCheck = `${accessor} != ${assertionValue}`;
      assertionTypeMessage = "was not";
      break;
    case "regex":
      addImport("regexp");

      if (assertionTargetFieldDef.Type.Type.toString() === "enum") {
        accessor = `string(${accessor})`;
      }

      assertionTypeCheck = `regexp.MustCompile(${templateRegexString(
        assertionValue,
      )}).MatchString(${accessor})`;
      assertionTypeMessage = "matched regular expression";

      break;
    default:
      throw new Error(`Unsupported assertion type: ${assertion.Type}`);
  }

  const message = `Response body at ${
    responseBodyAssertion.Path
  } ${assertionTypeMessage} ${escapeString(assertionValue)}`;

  result.push(
    ``,
    `if failureCriteriaMet {`,
    `  failureCriteriaMet = ${assertionTypeCheck}`,
    `  failureCriteriaMessage = append(failureCriteriaMessage, "${message}")`,
    `}`,
  );

  return result;
}

/**
 * Templates a code condition/expression for a polling status code assertion.
 *
 * For example: res.HTTPMeta.Response.StatusCode == 200
 */
function templatePollingStatusCodeAssertionExpression(
  assertion: Assertion,
): string {
  const response = assertion.Target as ResponseDef;
  const responseFormat = getResponseFormat();
  const statusCode = assertion.Value as string;
  const statusCodeAccessor = getStatusCodeAccessorFromResponse(
    response,
    "res",
    responseFormat,
    statusCode,
  );

  if (!statusCodeAccessor) {
    return "";
  }

  switch (assertion.Type.toString()) {
    case "equal":
      return responseFormat === "flat"
        ? `${statusCodeAccessor} != nil`
        : `${statusCodeAccessor} == ${statusCode}`;
    case "notEqual":
      return responseFormat === "flat"
        ? `${statusCodeAccessor} == nil`
        : `${statusCodeAccessor} != ${statusCode}`;
    case "regex":
      if (responseFormat === "flat") {
        return `${statusCodeAccessor} != nil`;
      }

      addImport("regexp");
      addImport("strconv");

      return `regexp.MustCompile(${templateRegexString(
        statusCode,
      )}).MatchString(strconv.Itoa(${statusCodeAccessor}))`;
    default:
      throw new Error(`Unsupported assertion type: ${assertion.Type}`);
  }
}

/**
 * Templates a code condition/expression for a polling status code assertion
 * from an error return.
 *
 * For example: apiErr.StatusCode == 200
 */
function templatePollingStatusCodeErrorAssertionExpression(
  assertion: Assertion,
  accessor: string,
): string {
  const statusCode = assertion.Value as string;
  const statusCodeAccessor = `${accessor}.StatusCode`;

  if (!statusCode) {
    return "";
  }

  switch (assertion.Type.toString()) {
    case "equal":
      return `${statusCodeAccessor} == ${statusCode}`;
    case "notEqual":
      return `${statusCodeAccessor} != ${statusCode}`;
    case "regex":
      addImport("regexp");
      addImport("strconv");

      return `regexp.MustCompile(${templateRegexString(
        statusCode,
      )}).MatchString(strconv.Itoa(${statusCodeAccessor}))`;
    default:
      throw new Error(`Unsupported assertion type: ${assertion.Type}`);
  }
}

/**
 *
 */
function templatePollingSuccessStatusCodeAssertion(
  assertion: Assertion,
): string[] {
  const statusCodeAssertionExpression =
    templatePollingStatusCodeAssertionExpression(assertion);

  return !statusCodeAssertionExpression
    ? []
    : [
        ``,
        `if successCriteriaMet {`,
        `  successCriteriaMet = ${statusCodeAssertionExpression}`,
        `}`,
      ];
}

/**
 *
 */
function templatePollingSuccessResponseBodyAssertion(
  assertion: Assertion,
  operation: Operation,
): string[] {
  const responseBodyAssertion = assertion.Value as ResponseBodyAssertion;
  const responseBodyFieldDef = responseBodyAssertion.Content.Content;
  const responseFormat = getResponseFormat();
  const assertionTargetFieldDef = !isFieldDef(assertion.Target)
    ? typeDefToFieldDef(assertion.Target)
    : (assertion.Target as FieldDef);
  let responseBodyAccessor = "res";

  switch (responseFormat) {
    case "envelope":
    case "envelope-http":
      responseBodyAccessor += `${fieldAccessor(responseBodyFieldDef)}`;
      break;
    case "flat":
      if (operation.Response.Type?.ResponseEnvelope) {
        const resultField = getResultField(operation);

        if (resultField) {
          responseBodyAccessor += `${fieldAccessor(resultField)}`;
        }
      }
      break;
    default:
      throw new Error(`Unknown response format: ${responseFormat}`);
  }

  let { path: accessor } = templateJSONPointerPath(
    responseBodyFieldDef,
    responseBodyAccessor,
    responseBodyAssertion.Path,
    (partIdx, parent, currentPath) => {
      return;
    },
    {},
  );

  let assertionValue = undefined;

  if (responseBodyAssertion.Value) {
    assertionValue = getExampleValue(responseBodyAssertion.Value, {});
  }

  if (
    (!responseBodyAssertion.Path || responseBodyAssertion.Path === "/") &&
    assertion.Type.toString() === "notEqual" &&
    assertionValue === undefined
  ) {
    if (assertionTargetFieldDef.Nullable) {
      return [
        ``,
        `if successCriteriaMet {`,
        `  successCriteriaMet = ${accessor} != nil`,
        `}`,
      ];
    }

    switch (assertionTargetFieldDef.Type.Type.toString()) {
      case "array":
      case "map":
      case "string":
        return [
          ``,
          `if successCriteriaMet {`,
          `  successCriteriaMet = len(${accessor}) > 0`,
          `}`,
        ];
    }

    return [];
  }

  const supportedTypes = [
    "boolean",
    "enum",
    "float",
    "float32",
    "integer",
    "int32",
    "number",
    "string",
  ];

  if (!supportedTypes.includes(assertionTargetFieldDef.Type.Type.toString())) {
    throw new Error(
      `Unsupported polling success criteria response body assertion type: ${assertionTargetFieldDef.Type.Type}`,
    );
  }

  const result: string[] = [];

  if (assertionTargetFieldDef.Optional || assertionTargetFieldDef.Nullable) {
    result.push(
      ``,
      `if successCriteriaMet {`,
      `  successCriteriaMet = ${accessor} != nil`,
      `}`,
    );
    accessor = `*${accessor}`;
  }

  let assertionTypeCheck = "";

  switch (assertion.Type.toString()) {
    case "equal":
      assertionTypeCheck = `${accessor} == ${assertionValue}`;
      break;
    case "notEqual":
      assertionTypeCheck = `${accessor} != ${assertionValue}`;
      break;
    case "regex":
      addImport("regexp");

      if (assertionTargetFieldDef.Type.Type.toString() === "enum") {
        accessor = `string(${accessor})`;
      }

      assertionTypeCheck = `regexp.MustCompile(${templateRegexString(
        assertionValue,
      )}).MatchString(${accessor})`;

      break;
    default:
      throw new Error(`Unsupported assertion type: ${assertion.Type}`);
  }

  result.push(
    ``,
    `if successCriteriaMet {`,
    `  successCriteriaMet = ${assertionTypeCheck}`,
    `}`,
  );

  return result;
}

/**
 * Returns the response status code accessor based on the response format. For
 * example with envelope format, it would be "res.StatusCode".
 */
function getStatusCodeAccessorFromResponse(
  response: ResponseDef,
  responseAccessor: string,
  responseFormat: "envelope" | "envelope-http" | "flat",
  statusCode: string,
): string {
  switch (responseFormat) {
    case "envelope":
      return `${responseAccessor}.StatusCode`;
    case "envelope-http":
      return `${responseAccessor}.HTTPMeta.Response.StatusCode`;
    case "flat":
      if (
        response.Responses.find((r) =>
          doesSubResponseContainStatusCode(r, statusCode),
        )?.Content.length > 0 ||
        response.Responses.find((r) => r.Headers) !== undefined
      ) {
        return responseAccessor;
      }

      // No response type accessor for flat responses without content or headers
      return ``;
    default:
      throw new Error(`Unknown response format: ${responseFormat}`);
  }
}
