registerReadmeSection(
  "summary",
  () => context.Global.AST.MainSDK.Comments != undefined,
  (_sdk: SDK) => templateString("readme/summary.stmpl", {}),
);

registerReadmeSection(
  "toc",
  () => true,
  (_sdk: SDK) => "", // generated separately based on final README content
);

registerReadmeSection(
  "installation",
  () => true,
  (_sdk: SDK) => templateString("readme/installation.stmpl", {}),
);

registerReadmeSection(
  "usage",
  () => true,
  (_sdk: SDK) => templateString("readme/usage-container.stmpl", {}),
);

registerReadmeSection("operations", () => true, templateAvailableOperations);

registerReadmeSection(
  "dev-containers",
  () => context.Global.Config.DevContainterSchemaPath != undefined,
  (_sdk: SDK) => templateString("readme/dev-containers.stmpl", {}),
);

const globalsInfo = prepareGlobalParametersSection();
registerReadmeSection(
  "global-parameters",
  () => globalsInfo.NeedsGlobalParameters,
  (_sdk: SDK) => templateString("readme/globals.stmpl", globalsInfo),
);
