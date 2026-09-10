require("includes/includes.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("//", "ts");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
