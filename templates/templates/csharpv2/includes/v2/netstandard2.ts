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
    // Must match the version in System.Text.Json's netstandard2.0 dependency group.
    "Microsoft.Bcl.AsyncInterfaces": {
      name: "Microsoft.Bcl.AsyncInterfaces",
      version: "10.0.8",
      cpe: "cpe:2.3:a:microsoft:bcl.asyncinterfaces:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
      condition: "netstandard2.0 && dotnetVersion == net10.0",
    },
    "Microsoft.Bcl.AsyncInterfaces@net8": {
      name: "Microsoft.Bcl.AsyncInterfaces",
      version: "8.0.0",
      cpe: "cpe:2.3:a:microsoft:bcl.asyncinterfaces:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
      condition: "netstandard2.0 && dotnetVersion == net8.0",
    },
  };
}

// @ts-ignore
function getNetstandard2Dependency(name: string) {
  const ns2Deps = getNetstandard2Dependencies();
  return context.Global.Config.DotnetVersion === "net8.0"
    ? ns2Deps[`${name}@net8`]
    : ns2Deps[name];
}

if (typeof module !== "undefined") {
  module.exports = { getNetstandard2Dependencies };
}
