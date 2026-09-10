function selectErrorOperation(): UsageContext {
  const scopes: UsageExampleScope[] = [
    {
      OpFilter: "errors",
      Feature: "errors",
      IsGlobal: false,
    },
  ];
  const contexts: UsageContext[] = selectExampleOperations(
    context.Global.AST.MainSDK,
    1,
    scopes,
    false,
  );
  if (contexts.length > 0) {
    return contexts[0];
  } else {
    const scopes: UsageExampleScope[] = [
      {
        OpFilter: "all",
        Feature: "errors",
        IsGlobal: false,
      },
    ];
    return selectExampleOperations(
      context.Global.AST.MainSDK,
      1,
      scopes,
      false,
    )[0];
  }
}

registerReadmeSection(
  "errors",
  () => true,
  (sdk: SDK) =>
    templateString("readme/errors.stmpl", {
      UsageContext: selectErrorOperation(),
      SDK: sdk,
    }),
);
