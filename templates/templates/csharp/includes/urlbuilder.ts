// @ts-ignore
function templateBuildURL(op: Operation): string {
  const hasPath = op.Request?.Params?.HasPathParams() ?? false;
  const hasQuery = op.Request?.Params?.HasQueryParams() ?? false;
  if (!hasPath && !hasQuery) {
    return `var urlString = baseUrl + "${op.Path}";`;
  }
  return `var urlString = URLBuilder.Build(baseUrl, "${
    op.Path
  }", request, ${templateAllowEmptyValue(op)});`;
}

registerTemplateFunc("templateBuildURL", templateBuildURL);

// @ts-ignore
function templateApplyHeaders(op: Operation): string {
  if (!(op.Request?.Params?.HasHeaderParams() ?? false)) return "";
  return `HeaderSerializer.PopulateHeaders(ref httpRequest, request);`;
}

registerTemplateFunc("templateApplyHeaders", templateApplyHeaders);
