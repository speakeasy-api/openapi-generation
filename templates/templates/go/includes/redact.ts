function getRedactLocation() {
  return getTypesLocation().replace(/types$/, "redact");
}
registerTemplateFunc("getRedactLocation", getRedactLocation);

function renderRedactGraph(graph: SensitiveBodyGraph | null): string {
  if (!graph) return "";
  const nodes = graph.Nodes.map((node) => {
    const parts: string[] = [];
    if (node.Sensitive) parts.push("sensitive: true");
    if (node.Fields?.length)
      parts.push(
        `fields: map[string]int{ ${node.Fields.map(
          (field) => `${JSON.stringify(field.Name)}: ${field.Node}`,
        ).join(", ")} }`,
      );
    if (node.Item) parts.push(`item: ${node.Item}`);
    if (node.Values) parts.push(`values: ${node.Values}`);
    if (node.Variants?.length)
      parts.push(`variants: []int{ ${node.Variants.join(", ")} }`);
    if (node.PlainObject) parts.push("plainObject: true");
    if (node.PlainArray) parts.push("plainArray: true");
    if (node.PlainString) parts.push("plainString: true");
    if (node.PlainNumber) parts.push("plainNumber: true");
    if (node.PlainBoolean) parts.push("plainBoolean: true");
    return `{ ${parts.join(", ")} }`;
  });
  return `&graph{ nodes: []node{ ${nodes.join(
    ", ",
  )} }, roots: []int{ ${graph.Roots.join(", ")} } }`;
}

function templateRedactGraphs(): {
  OperationID: string;
  Request: string;
  Response: string;
}[] {
  return allOperations()
    .filter((op) => op.HasSensitiveBodies())
    .map((op) => ({
      OperationID: JSON.stringify(op.ID),
      Request: renderRedactGraph(op.SensitiveRequestBodyGraph()),
      Response: renderRedactGraph(op.SensitiveResponseBodyGraph()),
    }))
    .sort((a, b) => a.OperationID.localeCompare(b.OperationID));
}
registerTemplateFunc("templateRedactGraphs", templateRedactGraphs);

function hasRedactGraphs(op: Operation): boolean {
  return op.HasSensitiveBodies();
}
registerTemplateFunc("hasRedactGraphs", hasRedactGraphs);

function hasAnyRedactGraphs(): boolean {
  return allOperations().some((op) => op.HasSensitiveBodies());
}
registerTemplateFunc("hasAnyRedactGraphs", hasAnyRedactGraphs);
