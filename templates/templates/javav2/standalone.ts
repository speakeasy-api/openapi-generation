require("includes/includes.ts");

if (!context.Global.Config.GroupID?.includes(".")) {
  throw new ValidationError(
    `the groupID in gen.yaml must be in reverse domain name notation, e.g. com.example`,
  );
}

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("//", "java");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
