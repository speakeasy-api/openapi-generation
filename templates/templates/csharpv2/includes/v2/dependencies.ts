type TargetDependency = {
  package: string;
  version: string;
  excludeAssets?: string;
  includeAssets?: string;
  privateAssets?: string;
  annotation?: string;
};

//@ts-ignore
function getSDKDependencies(): TargetDependency[] {
  if (!requiresNodatime()) {
    return [];
  }
  return [
    {
      package: "nodatime",
      version: deps.nodatime.version,
    },
  ];
}

// Multi-target (netstandard2.0;netX) SDK and Models projects only: a direct
// System.Text.Json reference on a single netX TFM is pruned and NU1510 fails
// warnings-as-errors builds (e.g. the usage testproject).
function templateNetstandard2Dependencies(): string {
  const ns2Deps = getNetstandard2Dependencies();
  const stj =
    context.Global.Config.DotnetVersion == "net8.0"
      ? ns2Deps["System.Text.Json@net8"]
      : ns2Deps["System.Text.Json"];
  return renderDependencies([
    { package: "System.Text.Json", version: stj.version },
  ]);
}

registerTemplateFunc(
  "templateNetstandard2Dependencies",
  templateNetstandard2Dependencies,
);
