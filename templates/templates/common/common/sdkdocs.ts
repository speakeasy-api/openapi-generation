// Language template accessible utility functions for SDK docs

type SDKDocContentType =
  | "class"
  | "enum"
  | "array"
  | "map"
  | "union"
  | "primitive";

type SDKDocModel =
  | { Type: "class"; Content: SDKDocModelContent[] }
  | { Type: "enum"; EnumType: TypeDef }
  | { Type: "array"; ItemType: SDKDocModelContent }
  | { Type: "map"; Content: SDKDocModelContent[] }
  | { Type: "union"; Content: SDKDocModelContent[] }
  | { Type: "primitive"; Content: SDKDocModelContent };

interface SDKDocModelContent {
  Name: string;
  Type: string;
  Optional: boolean;
  Description?: string;
  Example?: string;
  ReferencedContentType?: SDKDocContentType;
}

interface SDKDocType {
  Name: string;
  Type: string;
  Optional: boolean;
  Description?: string;
  Example?: string;
  Reference?: {
    Name: string;
    Type: SDKDocContentType | undefined;
    ImportPath: string;
    InlineContent: boolean;
  };
}

function generateBulkUsageSnippets(
  inputs: { SDK: SDK; Operation: Operation }[],
): Record<string, string> {
  const resultMap: Record<string, string> = {};

  for (const i of inputs) {
    const usageContext = createUsageContext(
      i.SDK,
      i.Operation,
      i.Operation.Extensions.UsageExample,
    );
    resultMap[i.Operation.ID] = templateString(
      "usage/snippet.stmpl",
      usageContext,
    );
  }

  return resultMap;
}

function getParams(): Record<string, SDKDocType[]> {
  const resultMap: Record<string, SDKDocType[]> = {};
  for (const operation of flattenOperationsPerSDK(context.Global.AST.MainSDK)) {
    resultMap[operation.ID] = templateMethodParametersDocsSDKs(operation);
  }
  return resultMap;
}

function getResponses(): Record<string, SDKDocType[]> {
  const resultMap: Record<string, SDKDocType[]> = {};
  for (const operation of flattenOperationsPerSDK(context.Global.AST.MainSDK)) {
    resultMap[operation.ID] = templateMethodResponseDocsSDKs(operation);
  }
  return resultMap;
}
