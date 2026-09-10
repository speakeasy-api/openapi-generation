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
  importErrorTypes: () => "";
};

const BUILTIN_ERRORS = [
  getDefaultErrorClassName(),
  "SDKValidationError",
  "UnexpectedClientError",
  "InvalidRequestError",
  "RequestAbortedError",
  "RequestTimeoutError",
  "ConnectionError",
] as const;

type BuiltinError = (typeof BUILTIN_ERRORS)[number];
