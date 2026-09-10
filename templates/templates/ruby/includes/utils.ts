function allSubResponsesHaveHeaders(subResponses: SubResponse[]): boolean {
  return subResponses.every((subResponse) => subResponse.Headers);
}

registerTemplateFunc("allSubResponsesHaveHeaders", allSubResponsesHaveHeaders);

function getSourcePath(sdkName: string): string {
  return `lib/${getSDKFolderName(sdkName)}`;
}

function getGlobalModulePath(): string {
  return getSourcePath(context.Global.Config.Module);
}
registerTemplateFunc("getGlobalModulePath", getGlobalModulePath);

function getSDKFolderName(sdkName: string): string {
  return caser().ToSnake(sdkName);
}

registerTemplateFunc("getSourcePath", getSourcePath);

function containsStreamingResponse(response: ResponseDef): boolean {
  return response.Responses.some((r) =>
    r.Content.some(
      (c) =>
        c.SerializationMethod == "eventstream" ||
        c.SerializationMethod == "jsonl" ||
        c.Content.Type.Type.toString() == "response-stream",
    ),
  );
}
registerTemplateFunc("containsStreamingResponse", containsStreamingResponse);

// templateLiveStreamContentTypes renders the Ruby array of response content
// types whose handlers consume the body incrementally (SSE, JSONL and raw
// response streams). A streaming method leaves those bodies live and buffers
// every other content type before decoding it.
function templateLiveStreamContentTypes(response: ResponseDef): string {
  const contentTypes: string[] = [];
  response.Responses.forEach((r) => {
    r.Content.forEach((c) => {
      const live =
        c.SerializationMethod == "eventstream" ||
        c.SerializationMethod == "jsonl" ||
        c.Content.Type.Type.toString() == "response-stream";
      if (live && !contentTypes.includes(c.ContentType)) {
        contentTypes.push(c.ContentType);
      }
    });
  });
  return `[${contentTypes.map((ct) => `'${escapeString(ct)}'`).join(", ")}]`;
}
registerTemplateFunc(
  "templateLiveStreamContentTypes",
  templateLiveStreamContentTypes,
);

function getScopeNamespace(scope: string): string {
  const module = sanitizeModuleName(context.Global.Config.Module);
  return `::${module}::${getNamespaceModuleParts(getScopePath(scope)).join(
    "::",
  )}`;
}
