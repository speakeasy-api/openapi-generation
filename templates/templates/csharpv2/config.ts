// @ts-ignore
const csharpv1 = require("../csharp/config.ts");
// @ts-ignore
const netstandard2 = require("./includes/v2/netstandard2.ts");

const TARGET_FRAMEWORK_PATTERN =
  /^(netstandard2\.0|net8\.0|net10\.0|dotnetVersion)$/;
const TARGET_FRAMEWORK_MESSAGE =
  "netstandard2.0, net8.0, net10.0 or dotnetVersion";

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
  const targetFramework = config.LangCfg["targetFramework"];
  if (targetFramework && !TARGET_FRAMEWORK_PATTERN.test(targetFramework)) {
    throw new Error(
      `targetFramework: ${targetFramework} incompatible with templateVersion: v2. ${TARGET_FRAMEWORK_MESSAGE} required.`,
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

  if (!config.LangCfg["enableFormatting"]) {
    commands.push({
      command: "dotnet",
      args: ["format", "whitespace"],
    });
  }

  // Note: Usage is set to "true" in testprojects/ (see g.compileTestProject())
  commands.push({
    command: "dotnet",
    args: [
      '{{if eq .Usage "true"}}publish{{else}}build{{end}}',
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
  // v1 fix flags select Newtonsoft-specific output.
  const { fixes: _fixes, ...v1Fields } = csharpv1.getConfigFields(
    commonFields,
    newSDK,
  );
  const fixedConfig = getFixedConfig();
  return {
    ...Object.fromEntries(
      Object.entries(v1Fields).filter(([key]) => !(key in fixedConfig)),
    ),
    targetFramework: {
      ...v1Fields.targetFramework,
      ValidationRegex: TARGET_FRAMEWORK_PATTERN.source,
      ValidationMessage: TARGET_FRAMEWORK_MESSAGE,
      Description:
        "The compile target framework moniker emitted as <TargetFramework>. Must not exceed the configured dotnetVersion. netstandard2.0 builds one package for netstandard2.0 and dotnetVersion. Set 'dotnetVersion' to use the same value as dotnetVersion.",
    },
    dotnetVersion: {
      Name: "dotnetVersion",
      Required: false,
      DefaultValue: newSDK ? "net10.0" : "net8.0",
      ValidationRegex: /^(net8\.0|net10\.0)$/.source,
      ValidationMessage: "net8.0 or net10.0 only",
      Description:
        "The version of .NET to target. net10.0 (default) or net8.0. AOT-safety requires net8.0+. https://learn.microsoft.com/en-us/dotnet/standard/frameworks",
    },
    preserveUnknownFields: {
      Name: "preserveUnknownFields",
      Required: false,
      DefaultValue: false,
      Description:
        "Preserve undeclared fields of JSON models: request and response bodies, webhook and callback payloads, nested models, error payloads, server-sent event data and JSON-encoded parameters, including schemas with additionalProperties: false. Each such model gets an UnknownFields dictionary that holds the undeclared fields it received (null when there are none) and writes them back when the model is serialized; a JSON null value is sent as null. Models whose schema declares additionalProperties keep their AdditionalProperties member and get no UnknownFields. Flattened operation arguments accept only declared fields, so undeclared top-level request fields require maxMethodParams: 0; nested models passed as arguments keep theirs. Requests for later pages of a paginated operation rebuild the top-level request body from its declared fields. The member name UnknownFields is reserved in generated models. Disabled by default.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
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
  return { ...rest, ...netstandard2.getNetstandard2Dependencies() };
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  return csharpv1.upgradeConfig(oldVersion, newVersion, cfg, defaults);
}

// @ts-ignore
function getFixedConfig(): Record<string, any> {
  return {
    unionStrategy: "populated-fields",
    optionalNullableScope: "all",
    // csharpv2 targets AOT: System.Text.Json source generation has no reflection
    // fallback, so the tri-state OptionalNullable<T> path is mandatory. Force
    // presenceAwareJsonSerialization on — with it off, optional+nullable fields
    // silently collapse to plain T? and the absent-vs-explicit-null distinction is
    // lost at runtime.
    presenceAwareJsonSerialization: true,
    partialTypes: true,
    unionImplicitConversions: true,
    enableFormatting: true,
  };
}

// @ts-ignore
function resolveConfig(cfg: Record<string, any>) {
  for (const [key, value] of Object.entries(getFixedConfig())) {
    if (cfg[key] != null && cfg[key] !== "" && cfg[key] !== value) {
      console.warn(
        `csharp: ${key} ${JSON.stringify(
          cfg[key],
        )} is not supported with templateVersion: v2. Using ${JSON.stringify(
          value,
        )}.`,
      );
    }
    cfg[key] = value;
  }
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
