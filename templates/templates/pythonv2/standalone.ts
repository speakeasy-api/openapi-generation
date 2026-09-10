require("includes/includes.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("#", "py");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
