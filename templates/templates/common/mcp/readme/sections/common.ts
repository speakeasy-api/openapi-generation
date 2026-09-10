registerReadmeSection(
  "summary",
  () => context.Global.AST.MainSDK.Comments != undefined,
  (_sdk: SDK) => templateString("readme/summary.stmpl", {}),
  MCP_AVAILABLE_README_SECTIONS,
);

registerReadmeSection(
  "toc",
  () => true,
  (_sdk: SDK) => "", // generated separately based on final README content
  MCP_AVAILABLE_README_SECTIONS,
);

registerReadmeSection(
  "installation",
  () => true,
  (_sdk: SDK) => templateString("readme/installation.stmpl", {}),
  MCP_AVAILABLE_README_SECTIONS,
);

registerReadmeSection(
  "dynamic-mode",
  () => true,
  (sdk: SDK) => {
    const scopes = new Set<string>();
    const queue: SDK[] = [sdk];
    while (queue.length > 0) {
      const s = queue.shift()!;
      queue.push(...s.SubSDKs);
      for (const op of s.Operations) {
        if (op.Extensions?.MCP?.Scopes) {
          op.Extensions.MCP.Scopes.forEach((scope: string) =>
            scopes.add(scope),
          );
        }
      }
    }
    const sortedScopes = Array.from(scopes).sort();
    return templateString("readme/dynamic-mode.stmpl", {
      HasScopes: sortedScopes.length > 0,
      FirstScope: sortedScopes.length > 0 ? sortedScopes[0] : "",
    });
  },
  MCP_AVAILABLE_README_SECTIONS,
);
