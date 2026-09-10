registerReadmeSection(
  "server",
  () => context.Global.AST.MainSDK.Servers?.HasAbsoluteURL() || false,
  (_sdk: SDK) =>
    templateString("readme/server.stmpl", getServerReadmeContext()),
);
