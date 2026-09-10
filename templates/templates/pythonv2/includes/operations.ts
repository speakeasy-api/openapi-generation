// @ts-ignore
function methodTemplateState(op: Operation) {
  return {
    operation: op,
  };
}

registerTemplateFunc("methodTemplateState", methodTemplateState);

// @ts-ignore
function templateAllowEmptyValue(op: Operation): string {
  const paramNames = getAllowEmptyValueQueryParamNames(op);

  if (paramNames.length === 0) {
    return "None";
  }

  return `[${paramNames.map((name) => `"${name}"`).join(", ")}]`;
}

registerTemplateFunc("templateAllowEmptyValue", templateAllowEmptyValue);
