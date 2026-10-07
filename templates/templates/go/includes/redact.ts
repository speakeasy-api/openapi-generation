function getRedactLocation() {
  return getTypesLocation().replace(/types$/, "redact");
}
registerTemplateFunc("getRedactLocation", getRedactLocation);

function isSensitiveBodyType(type: TypeDef): boolean {
  const marker = type.Extensions?.All?.["x-speakeasy-param-sensitive"];
  return marker === true || marker === "true" || type.Format === "password";
}

interface RedactNode {
  Sensitive?: boolean;
  // A stream body is masked as a whole when any of its items holds a
  // sensitive value, since a stream is not parsed item by item.
  Stream?: boolean;
  Fields?: Map<string, number>;
  Item?: number;
  Values?: number;
  Variants?: number[];
}

interface RedactGraph {
  Nodes: RedactNode[];
  Roots: number[];
}

function newRedactGraph(): RedactGraph {
  return { Nodes: [{}], Roots: [] };
}

function addRedactRoot(graph: RedactGraph, type: TypeDef | undefined) {
  const walked = new Map<string, number>();
  const walk = (type: TypeDef | undefined): number => {
    if (!type) return 0;
    const typeID = type.IsCustomType() ? getUniqueID(type) : "";
    if (typeID && walked.has(typeID)) return walked.get(typeID)!;
    const id = graph.Nodes.length;
    const node: RedactNode = {};
    graph.Nodes.push(node);
    if (typeID) walked.set(typeID, id);
    if (isSensitiveBodyType(type)) {
      node.Sensitive = true;
      return id;
    }
    switch (type.Type.toString()) {
      case "class":
      case "error":
        node.Fields = new Map();
        for (const field of type.Fields || []) {
          if (field.IsAdditionalProperties)
            node.Values = walk(field.Type.ItemType);
          else
            node.Fields.set(
              originalFieldName(field),
              field.Const ? 0 : walk(field.Type),
            );
        }
        break;
      case "array":
      case "set":
        node.Item = walk(type.ItemType);
        break;
      case "map":
        node.Values = walk(type.ItemType);
        break;
      case "event-stream":
      case "jsonl":
        node.Stream = true;
        node.Item = walk(type.ItemType);
        break;
      case "union":
        node.Variants = (type.AssociatedTypes || []).map(walk);
        break;
    }
    return id;
  };
  graph.Roots.push(walk(type));
}

function renderRedactGraph(graph: RedactGraph): string | undefined {
  const reachable = new Set<number>();
  graph.Nodes.forEach((node, i) => {
    if (node.Sensitive) reachable.add(i);
  });
  let changed = true;
  while (changed) {
    changed = false;
    graph.Nodes.forEach((node, i) => {
      if (reachable.has(i)) return;
      const children = [
        ...(node.Fields?.values() || []),
        ...(node.Variants || []),
        node.Item || 0,
        node.Values || 0,
      ];
      if (children.some((child) => reachable.has(child))) {
        reachable.add(i);
        changed = true;
      }
    });
  }
  const roots = graph.Roots.filter((root) => reachable.has(root));
  if (roots.length === 0) return undefined;
  const indices = [...reachable].sort((a, b) => a - b);
  const remap = new Map(indices.map((id, i) => [id, i + 1]));
  const nodes = indices.map((id) => {
    const node = graph.Nodes[id];
    if (node.Stream) return "{ sensitive: true }";
    const parts: string[] = [];
    if (node.Sensitive) parts.push("sensitive: true");
    const values = node.Values && reachable.has(node.Values) ? node.Values : 0;
    // With sensitive map values, declared fields are listed even when plain so
    // that they never fall back to the map-value schema.
    const fields = [...(node.Fields || [])]
      .filter(([, child]) => values || reachable.has(child))
      .map(
        ([name, child]) => `${JSON.stringify(name)}: ${remap.get(child) || 0}`,
      );
    if (fields.length)
      parts.push(`fields: map[string]int{ ${fields.join(", ")} }`);
    if (node.Item && reachable.has(node.Item))
      parts.push(`item: ${remap.get(node.Item)}`);
    if (values) parts.push(`values: ${remap.get(values)}`);
    const variants = (node.Variants || []).filter((child) =>
      reachable.has(child),
    );
    if (variants.length)
      parts.push(
        `variants: []int{ ${variants
          .map((child) => remap.get(child))
          .join(", ")} }`,
      );
    return `{ ${parts.join(", ")} }`;
  });
  return `&graph{ nodes: []node{ {}, ${nodes.join(
    ", ",
  )} }, roots: []int{ ${roots.map((id) => remap.get(id)).join(", ")} } }`;
}

function redactGraphs(op: Operation): {
  Request?: string;
  Response?: string;
} {
  const request = newRedactGraph();
  const body = getRequestBody(op);
  addRedactRoot(
    request,
    body?.Type === "request" ? body.Value.RequestBody?.Type : body?.Value.Type,
  );
  const response = newRedactGraph();
  for (const sub of op.Response?.Responses || []) {
    for (const content of sub.Content || []) {
      addRedactRoot(response, content.Content?.Type);
    }
  }
  return {
    Request: renderRedactGraph(request),
    Response: renderRedactGraph(response),
  };
}

function hasRedactGraphs(op: Operation): boolean {
  const graphs = redactGraphs(op);
  return !!(graphs.Request || graphs.Response);
}
registerTemplateFunc("hasRedactGraphs", hasRedactGraphs);

function templateRedactGraphs(): {
  OperationID: string;
  Request?: string;
  Response?: string;
}[] {
  const entries: {
    OperationID: string;
    Request?: string;
    Response?: string;
  }[] = [];
  for (const op of allOperations()) {
    const graphs = redactGraphs(op);
    if (graphs.Request || graphs.Response)
      entries.push({ OperationID: JSON.stringify(op.ID), ...graphs });
  }
  return entries.sort((a, b) => a.OperationID.localeCompare(b.OperationID));
}
registerTemplateFunc("templateRedactGraphs", templateRedactGraphs);

function hasAnyRedactGraphs(): boolean {
  return templateRedactGraphs().length > 0;
}
registerTemplateFunc("hasAnyRedactGraphs", hasAnyRedactGraphs);
