type OperationStateTS = {
  funcName: string;
  usageLocation: string;
  operation: Operation;
  signature: {
    returnType: string;
    unwrappedReturnType: string;
  };
  valueType: string;
  errorType: string;
  resultType: string;
  inboundResponse: ResolvedTypes;
  responseFormat: string;
  importErrorTypes: () => "";
  flatOptions: { needsEnvelope: boolean; returnTypeOptional: boolean };
};

function getBuiltinErrors() {
  if (isNoZod()) {
    // SDKValidationError + ResponseValidationError are unreachable in no-zod
    // (no schema validation, no parse-error wrapping). Excluded so the funcs'
    // Result error union doesn't reference types that nothing constructs.
    return [
      getBaseErrorClassName(),
      "ConnectionError",
      "RequestAbortedError",
      "RequestTimeoutError",
      "InvalidRequestError",
      "UnexpectedClientError",
    ] as const;
  }
  return [
    getBaseErrorClassName(),
    "ResponseValidationError",
    "ConnectionError",
    "RequestAbortedError",
    "RequestTimeoutError",
    "InvalidRequestError",
    "UnexpectedClientError",
    "SDKValidationError",
  ] as const;
}

type BuiltinError = string;
