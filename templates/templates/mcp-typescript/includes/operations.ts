function opTemplateState(op: Operation, usageLoc?: string): OperationStateTS {
  const usageLocation = usageLoc ?? op.OwningSDK.Type.OutputLocation;
  const errorImportFuncs: Array<() => void> = [];

  let inboundResponse: ResolvedTypes = {
    inputType: "unknown",
    importInputTypes: () => "",
    outputType: "Response",
    importOutputTypes: () => "",
    zod: "",
    importZodTypes: () => "",
  };
  if (isStrictMCPServer()) {
    inboundResponse = resolveInbound({
      usageLocation,
      typeDef: op.Response.Type,
      rootTypeDef: op.OwningSDK.Type,
      optional: false,
      nullable: false,
    });
  }

  let errorType = "";
  if (isStrictMCPServer()) {
    for (const sub of op.Response.Responses) {
      if (!sub.Error) {
        continue;
      }
      for (const content of sub.Content) {
        if (!content.Content) {
          continue;
        }
        const re = resolveInbound({
          usageLocation,
          typeDef: content.Content.Type,
          rootTypeDef: op.OwningSDK.Type,
          optional: false,
          nullable: false,
        });

        errorImportFuncs.push(re.importOutputTypes);
        errorType += `${re.outputType} | `;
      }
    }
  }
  const errUnion = BUILTIN_ERRORS.join(" | ");
  errorType = errorType ? errorType.slice(0, -" | ".length) : "";
  errorType = errorType ? `${errorType} | ${errUnion}` : errUnion;

  const importErrorTypes = (): "" => {
    for (const func of errorImportFuncs) {
      func();
    }
    BUILTIN_ERRORS.forEach((className) => {
      addErrorTypeImport(className, usageLocation);
    });
    return "";
  };

  const valueType = inboundResponse.outputType;
  const resultType = `Result<${valueType}, ${errorType}>`;
  const returnType = resultType;
  const unwrappedReturnType = inboundResponse.outputType;

  return {
    funcName: sanitizeFuncName(op),
    usageLocation,
    operation: op,
    signature: { returnType, unwrappedReturnType },
    valueType,
    errorType,
    resultType,
    inboundResponse,
    importErrorTypes,
  };
}

registerTemplateFunc("opTemplateState", opTemplateState);
