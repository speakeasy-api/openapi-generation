// Loaded by both config.ts (SBOM) and includes.ts (templating).
// @ts-ignore
function getNetstandard2Dependencies() {
  return {
    "System.Text.Json": {
      name: "System.Text.Json",
      version: "10.0.8",
      cpe: "cpe:2.3:a:microsoft:system.text.json:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
      condition: "netstandard2.0 && dotnetVersion == net10.0",
    },
    "System.Text.Json@net8": {
      name: "System.Text.Json",
      version: "8.0.5",
      cpe: "cpe:2.3:a:microsoft:system.text.json:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
      condition: "netstandard2.0 && dotnetVersion == net8.0",
    },
  };
}

if (typeof module !== "undefined") {
  module.exports = { getNetstandard2Dependencies };
}
