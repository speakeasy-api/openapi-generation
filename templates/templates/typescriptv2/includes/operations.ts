// @ts-ignore
function opTemplateState(op: Operation, usageLoc?: string): OperationStateTS {
  let flatOptions = { needsEnvelope: false, returnTypeOptional: false };

  const responseFormat = getResponseFormat();

  if (responseFormat === "flat") {
    // operations with pagination will need an envelope
    flatOptions.needsEnvelope ||= Boolean(op.Extensions.Pagination);

    for (let resp of op.Response.Responses) {
      // errors required enveloping
      flatOptions.needsEnvelope ||=
        Boolean(resp.Error) && resp.Content.length > 0;
      // responses which declare headers also need enveloping
      flatOptions.needsEnvelope ||= resp.Headers;

      // look for non-error responses that have no schema defined
      flatOptions.returnTypeOptional ||= !resp.Error && !resp.Content.length;
    }
  }

  const usageLocation = usageLoc ?? op.OwningSDK.Type.OutputLocation;

  const errorImportFuncs: Array<() => void> = [];

  const inboundResponse = resolveInbound({
    usageLocation,
    typeDef: op.Response.Type,
    rootTypeDef: op.OwningSDK.Type,
    optional: flatOptions.returnTypeOptional,
    nullable: false,
  });

  const errorTypes = new Set<string>();
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
      errorTypes.add(re.outputType);
    }
  }
  // Deduplicate: if a spec-generated error has the same name as a built-in,
  // the spec-generated type takes precedence.
  const dedupedBuiltins = getBuiltinErrors().filter((e) => !errorTypes.has(e));
  const allErrors = new Set([...errorTypes, ...dedupedBuiltins]);
  const errorType = [...allErrors].join(" | ");

  const importErrorTypes = (): "" => {
    for (const func of errorImportFuncs) {
      func();
    }
    for (const className of dedupedBuiltins) {
      addErrorTypeImport(className, usageLocation);
    }
    return "";
  };

  const valueType = inboundResponse.outputType;
  const resultType = `Result<${valueType}, ${errorType}>`;
  let returnType = resultType;
  let unwrappedReturnType = inboundResponse.outputType;
  if (op.Extensions.Pagination) {
    const nextPageState = templateNextPageState(op).staticType;
    returnType = `PageIterator<${resultType}, ${nextPageState}>`;
    unwrappedReturnType = `PageIterator<${unwrappedReturnType}, ${nextPageState}>`;
  }

  return {
    funcName: sanitizeFuncName(op),
    usageLocation,
    operation: op,
    signature: { returnType, unwrappedReturnType },
    valueType,
    errorType,
    resultType,
    inboundResponse,
    responseFormat,
    flatOptions,
    importErrorTypes,
  };
}

registerTemplateFunc("opTemplateState", opTemplateState);
