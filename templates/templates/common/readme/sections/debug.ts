registerReadmeSection(
  "debug",
  () => true,
  (_sdk: SDK) => templateString("readme/debug.stmpl", {}),
);
