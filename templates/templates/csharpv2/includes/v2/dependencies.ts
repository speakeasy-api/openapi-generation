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

// Multi-target (netstandard2.0;netX) SDK project only: a direct
// System.Text.Json reference on a single netX TFM is pruned and NU1510 fails
// warnings-as-errors builds (e.g. the usage testproject).
function templateNetstandard2Dependencies(): string {
  return renderDependencies([
    {
      package: "System.Text.Json",
      version: getNetstandard2Dependency("System.Text.Json").version,
    },
    // A netX consumer below dotnetVersion binds lib/netstandard2.0 but resolves
    // System.Text.Json's netX group, which has no Microsoft.Bcl.AsyncInterfaces.
    {
      package: "Microsoft.Bcl.AsyncInterfaces",
      version: getNetstandard2Dependency("Microsoft.Bcl.AsyncInterfaces")
        .version,
    },
  ]);
}

registerTemplateFunc(
  "templateNetstandard2Dependencies",
  templateNetstandard2Dependencies,
);
