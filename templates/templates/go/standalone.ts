require("includes/includes.ts");
require("features.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("//", "go");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
