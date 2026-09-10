registerReadmeSection(
  "requirements",
  () => true,
  (sdk: SDK) => templateString("readme/runtimes.stmpl", {}),
);
