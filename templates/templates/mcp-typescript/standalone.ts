require("includes/includes.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("", "json");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
