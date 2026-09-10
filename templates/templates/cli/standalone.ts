require("includes/includes.ts");
require("includes/tests.ts");

if (context.Global.Config.Usage?.Standalone) {
  generateStandaloneUsage("#", "sh");
}

if (context.Global.Config.Readme?.Standalone) {
  generateStandaloneReadme();
}
