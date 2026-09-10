registerReadmeSection(
  "async-support",
  () => asyncEnabled(),
  (_sdk: SDK) => templateString("readme/async-support.stmpl", {}),
);
