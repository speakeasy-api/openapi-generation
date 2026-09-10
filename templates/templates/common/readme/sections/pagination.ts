function needsPaginationSection(): boolean {
  const scope: UsageExampleScope = {
    OpFilter: "pagination",
    Feature: "",
    IsGlobal: false,
  };
  const paginationOps = selectExampleOperations(
    context.Global.AST.MainSDK,
    1,
    [scope],
    false,
  );
  return paginationOps.length > 0;
}

registerReadmeSection("pagination", needsPaginationSection, (_sdk: SDK) =>
  templateString("readme/pagination.stmpl", {}),
);
