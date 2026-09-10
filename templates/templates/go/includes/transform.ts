function templateTransformFromAPI(dataVar: string, jqExpr: any): string {
  const expr = typeof jqExpr === "string" ? jqExpr : JSON.stringify(jqExpr);
  return [
    '{{- addImport "utils"}}',
    `if out, err := utils.RunJQBytes(${dataVar}, ${JSON.stringify(
      expr,
    )}); err != nil {`,
    `    return err`,
    `} else {`,
    `    ${dataVar} = out`,
    `}`,
  ].join("\n");
}
registerTemplateFunc("templateTransformFromAPI", templateTransformFromAPI);

function templateTransformToAPI(valueVar: string, jqExpr: any): string {
  const expr = typeof jqExpr === "string" ? jqExpr : JSON.stringify(jqExpr);
  return [
    '{{- addImport "utils"}}',
    `jsonBytes, err := utils.MarshalJSON(${valueVar}, "", false)`,
    `if err != nil {`,
    `    return nil, err`,
    `}`,
    `out, err := utils.RunJQBytes(jsonBytes, ${JSON.stringify(expr)})`,
    `if err != nil {`,
    `    return nil, err`,
    `}`,
    `return out, nil`,
  ].join("\n");
}
registerTemplateFunc("templateTransformToAPI", templateTransformToAPI);

function getJQFromAPI(type: TypeDef): string {
  return (
    (type?.Extensions?.TransformFromAPI?.Type == "jq" &&
      type?.Extensions?.TransformFromAPI?.Config) ||
    ""
  );
}
registerTemplateFunc("getJQFromAPI", getJQFromAPI);

function getJQToAPI(type: TypeDef): string {
  return (
    (type?.Extensions?.TransformToAPI?.Type == "jq" &&
      type?.Extensions?.TransformToAPI?.Config) ||
    ""
  );
}
registerTemplateFunc("getJQToAPI", getJQToAPI);
