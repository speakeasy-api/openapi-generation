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
  return [
    {
      package: "nodatime",
      version: deps.nodatime.version,
    },
  ];
}
