function needsEventStreamSection(): boolean {
  const scope: UsageExampleScope = {
    OpFilter: "eventstream",
    Feature: "eventstream",
    IsGlobal: false,
  };

  return (
    selectExampleOperations(context.Global.AST.MainSDK, 1, [scope], false)
      .length > 0
  );
}

registerReadmeSection("eventstream", needsEventStreamSection, (_sdk: SDK) =>
  templateString("readme/eventstream.stmpl", {}),
);
