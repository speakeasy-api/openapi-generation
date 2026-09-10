// @ts-ignore
function getErrorsLocation() {
  return getScopePath("errors");
}

registerTemplateFunc("getErrorsLocation", getErrorsLocation);

// @ts-ignore
function templateDefaultError(full: boolean = true): string {
  const className = getDefaultErrorClassName();

  if (usingGlobalImports()) {
    return className;
  }

  const namespace = getScopeNamespace("errors", full);
  return namespace ? `${namespace}.${className}` : className;
}

registerTemplateFunc("templateDefaultError", templateDefaultError);
