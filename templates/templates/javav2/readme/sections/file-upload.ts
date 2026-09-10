// @ts-ignore
function needsFileUploadSection(): boolean {
  if (!useBlob()) {
    return false;
  }

  const scope: UsageExampleScope = {
    OpFilter: "file-upload",
    Feature: "",
    IsGlobal: false,
  };

  return (
    selectExampleOperations(context.Global.AST.MainSDK, 1, [scope], false)
      .length > 0
  );
}

registerReadmeSection("file-upload", needsFileUploadSection, (_sdk: SDK) =>
  templateString("readme/file_upload.stmpl", {}),
);
