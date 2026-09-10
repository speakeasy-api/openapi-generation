registerReadmeSection(
  "jackson",
  () => true,
  (_sdk: SDK) => templateString("readme/jackson.stmpl", {}),
);
