type TargetDependency = {
  package: string;
  version: string;
  excludeAssets?: string;
  includeAssets?: string;
  privateAssets?: string;
};

// @ts-ignore
const deps = getTemplateDependencies();

function getSDKDependencies(): TargetDependency[] {
  const sdkDeps: TargetDependency[] = [
    {
      package: "newtonsoft.json",
      version: deps["newtonsoft.json"].version,
    },
    {
      package: "nodatime",
      version: deps.nodatime.version,
    },
  ];

  if (isNetstandard2() && isFeatureUsed("serverEvents")) {
    sdkDeps.push({
      package: "Microsoft.Bcl.AsyncInterfaces",
      version: deps["Microsoft.Bcl.AsyncInterfaces"].version,
    });
  }

  return sdkDeps;
}

function getTestDependencies(): TargetDependency[] {
  return [
    {
      package: "Microsoft.NET.Test.Sdk",
      version: deps["Microsoft.NET.Test.Sdk"].version,
    },
    {
      package: "xunit",
      version: deps.xunit.version,
    },
    {
      package: "xunit.runner.visualstudio",
      version: deps["xunit.runner.visualstudio"].version,
      includeAssets:
        "runtime; build; native; contentfiles; analyzers; buildtransitive",
      privateAssets: "all",
    },
    {
      package: "coverlet.collector",
      version: deps["coverlet.collector"].version,
      includeAssets:
        "runtime; build; native; contentfiles; analyzers; buildtransitive",
      privateAssets: "all",
    },
    {
      package: "JUnitXml.TestLogger",
      version: deps["JUnitXml.TestLogger"].version,
    },
  ];
}

//@ts-ignore
function renderDependency(dependency: TargetDependency): string {
  const keys = ["includeAssets", "excludeAssets", "privateAssets"];

  if (keys.some((key) => dependency[key])) {
    let output = `<PackageReference Include="${dependency.package}" Version="${dependency.version}">\n`;

    keys.forEach((key) => {
      if (dependency[key]) {
        const Key = caser().ToPascal(key);
        output += `  <${Key}>${dependency[key]}</${Key}>\n`;
      }
    });

    return output + `</PackageReference>`;
  }

  return `<PackageReference Include="${dependency.package}" Version="${dependency.version}" />`;
}

function renderDependencies(dependencies: TargetDependency[]): string {
  return indentLines(
    dependencies.map((dep) => renderDependency(dep)),
    1,
  );
}

//@ts-ignore
function templateSDKDependencies(): string {
  return renderDependencies([
    ...getSDKDependencies(),
    ...context.Global.Config.AdditionalDependencies,
  ]);
}

registerTemplateFunc("templateSDKDependencies", templateSDKDependencies);

//@ts-ignore
function templateDependencies(): string {
  const hasTests =
    context.Global.AST.MainSDK.OutputTests || sdkHasTests(context.Global.AST);

  return renderDependencies([
    ...getSDKDependencies(),
    ...(hasTests ? getTestDependencies() : []),
    ...context.Global.Config.AdditionalDependencies,
  ]);
}

registerTemplateFunc("templateDependencies", templateDependencies);

//@ts-ignore
function templateTestDependencies(): string {
  return renderDependencies(getTestDependencies());
}

registerTemplateFunc("templateTestDependencies", templateTestDependencies);
