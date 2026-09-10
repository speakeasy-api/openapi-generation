// @ts-ignore
function methodTemplateState(op: Operation) {
  return {
    operation: op,
    arguments: getArguments(op),
  };
}

registerTemplateFunc("methodTemplateState", methodTemplateState);

// @ts-ignore
function positionOfSecurity(op: Operation) {
  const style = context.Global.Config.MethodArguments;

  switch (style) {
    case "infer-optional-args":
      let securityOptional = true;
      if (op.Security) {
        securityOptional = op.Security.Optional;
      }
      if (securityOptional) {
        return -1;
      } else {
        return 0;
      }
    case "require-security-and-request":
      const flattened = operationParametersFlattened(op);

      if (flattened) {
        return 0;
      } else {
        return -1;
      }
  }
}

registerTemplateFunc("positionOfSecurity", positionOfSecurity);

// @ts-ignore
function templateAllowEmptyValue(op: Operation): string {
  const paramNames = getAllowEmptyValueQueryParamNames(op);

  if (paramNames.length === 0) {
    return "nil";
  }

  const mapEntries = paramNames.map((name) => `"${name}": {}`).join(", ");
  return `map[string]struct{}{${mapEntries}}`;
}

registerTemplateFunc("templateAllowEmptyValue", templateAllowEmptyValue);
