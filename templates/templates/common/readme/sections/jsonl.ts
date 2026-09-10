function needsJsonLStreamSection(): boolean {
  const scope: UsageExampleScope = {
    OpFilter: "jsonl",
    Feature: "jsonl",
    IsGlobal: false,
  };

  return (
    selectExampleOperations(context.Global.AST.MainSDK, 1, [scope], false)
      .length > 0
  );
}
registerTemplateFunc("needsJsonLStreamSection", needsJsonLStreamSection);

registerReadmeSection("jsonl", needsJsonLStreamSection, (_sdk: SDK) =>
  templateString("readme/jsonl.stmpl", {}),
);
