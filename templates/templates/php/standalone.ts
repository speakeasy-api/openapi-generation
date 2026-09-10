require("includes/includes.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("#", "php");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
