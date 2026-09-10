function needsRetriesSection(): boolean {
  const scope: UsageExampleScope = {
    OpFilter: "retries",
    Feature: "",
    IsGlobal: false,
  };
  const retriesOps = selectExampleOperations(
    context.Global.AST.MainSDK,
    1,
    [scope],
    false,
  );
  return retriesOps.length > 0;
}

registerReadmeSection("retries", needsRetriesSection, (_sdk: SDK) =>
  templateString("readme/retries.stmpl", {}),
);
