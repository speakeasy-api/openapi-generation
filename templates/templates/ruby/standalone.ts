require("includes/includes.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("#", "rb");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
