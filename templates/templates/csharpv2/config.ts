// @ts-ignore
const csharpv1 = require("../csharp/config.ts");

// @ts-ignore
function getCompileDependencies(
  config: GeneratorInitialConfiguration,
): RunnerCommandDependencies {
  const majorVersion =
    config.LangCfg["dotnetVersion"].match(/^net(8|10)\.0$/)?.[1];
  if (!majorVersion) {
    throw new Error(
      `dotnetVersion: ${config.LangCfg["dotnetVersion"]} incompatible with templateVersion: v2. net8.0 or net10.0 required.`,
    );
  }
  return [
    {
      command: "dotnet",
      version: {
        args: ["--list-sdks"],
        regex: `(?m).*?(${majorVersion}\\.\\d+\\.\\d+).*?`,
        minVersion: `${majorVersion}.0.0`,
      },
      installDocumentation: `Install .NET by following the instructions at https://dotnet.microsoft.com/download.`,
    },
  ];
}

// @ts-ignore
function getCompileCommands(
  config: GeneratorInitialConfiguration,
): RunnerCommands {
  const commands = [];
  const packageName = config.LangCfg["packageName"];
  const sourceDirectory = config.LangCfg["sourceDirectory"];
  const packagePath = packageName.replaceAll(".", "/");
  const projectFilePath = `${
    sourceDirectory ? `${sourceDirectory}/` : ""
  }${packagePath}/${packageName}.csproj`;

  if (!config.LangCfg["enableFormatting"]) {
    commands.push({
      command: "dotnet",
      args: ["format"],
    });
  }

  // Note: Usage is set to "true" in testprojects/ (see g.compileTestProject())
  commands.push({
    command: "dotnet",
    args: [
      '{{if eq .Usage "true"}}publish{{else}}build{{end}}',
      `{{if ne .Usage "true"}}${projectFilePath}{{end}}`,
      "-c",
      `{{if eq .Usage "true"}}Release{{else}}${
        config.Publish ? "Release" : "Debug"
      }{{end}}`,
      '{{if eq .Usage "true"}}--use-current-runtime{{end}}',
      // `-warnaserror` escalates all MSBuild warnings (CS + IL trim/AOT) to errors, noting that:
      //   - adding <TreatWarningsAsErrors> at the .csproj level does not consistently catch
      //     IL2xxx/IL3xxx in .NET 8 whereas the CLI flag does.
      //   - CS0618/CS1591 are pre-suppressed via NoWarn in testproject.csproj
      '{{if eq .Usage "true"}}-warnaserror{{end}}',
      // BuildInParallel=false serializes project-graph builds. dotnet publish otherwise
      // parallelizes Models.dll writes from multiple reference paths, triggering MSB3026
      // file-lock races.
      '{{if eq .Usage "true"}}-p:BuildInParallel=false{{end}}',
      // -nodeReuse:false prevents MSBuild from leaving long-lived `/nodemode:1` workers
      // after the command completes. Orphans inherit the parent shell's stdout and can
      // hang the parent after the generator finishes. Applied to both build and publish.
      "-nodeReuse:false",
    ],
  });

  return commands;
}

// @ts-ignore
function getConfigFields(
  commonFields: SDKGenConfigFields,
  newSDK: boolean,
): SDKGenConfigFields {
  return {
    ...csharpv1.getConfigFields(commonFields, newSDK),
    dotnetVersion: {
      Name: "dotnetVersion",
      Required: false,
      DefaultValue: newSDK ? "net10.0" : "net8.0",
      ValidationRegex: /^(net8\.0|net10\.0)$/.source,
      ValidationMessage: "net8.0 or net10.0 only",
      Description:
        "The version of .NET to target. net10.0 (default) or net8.0. AOT-safety requires net8.0+. https://learn.microsoft.com/en-us/dotnet/standard/frameworks",
    },
  };
}

// Re-export wrappers — same body as csharp, but top-level decls here become
// globals on the entry-script VM so executor.Run can resolve them.

// @ts-ignore
function getTemplateDependencies() {
  // csharpv2 drops newtonsoft.json: reflection-heavy serializer is not AOT-safe.
  // Models serialize via source-generated System.Text.Json contexts instead.
  const { ["newtonsoft.json"]: _drop, ...rest } =
    csharpv1.getTemplateDependencies();
  return rest;
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  return csharpv1.upgradeConfig(oldVersion, newVersion, cfg, defaults);
}

// @ts-ignore
function getTestingConfiguration(genConfig): TestingConfiguration {
  return csharpv1.getTestingConfiguration(genConfig);
}

// @ts-ignore
function getCompileConfiguration(
  genConfig: GeneratorInitialConfiguration,
): CompileConfiguration {
  return {
    runner: {
      commands: getCompileCommands(genConfig),
      dependencies: {
        commands: getCompileDependencies(genConfig),
      },
    },
  };
}

// @ts-ignore
function getGeneratorConfig(
  genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  return {
    compile: getCompileConfiguration(genConfig),
    testing: getTestingConfiguration(genConfig),
  };
}
