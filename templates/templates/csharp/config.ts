// @ts-ignore
function getCompileDependencies(
  config: GeneratorInitialConfiguration,
): RunnerCommandDependencies {
  const majorVersion =
    config.LangCfg["dotnetVersion"].match(/^net(5|6|8|10)\.0$/)?.[1];
  if (!majorVersion) {
    throw new Error(
      `Unsupported dotnetVersion: ${config.LangCfg["dotnetVersion"]}. Supported versions are net10.0, net8.0, net6.0 and net5.0.`,
    );
  }

  // A net targetFramework must not exceed the dotnetVersion build SDK: runner only installs the dotnetVersion SDK.
  const targetMajor =
    config.LangCfg["targetFramework"]?.match(/^net(\d+)\.0$/)?.[1];
  if (targetMajor && Number(targetMajor) > Number(majorVersion)) {
    throw new Error(
      `Invalid dotnetVersion: ${config.LangCfg["dotnetVersion"]}. Cannot be lower than targetFramework: ${config.LangCfg["targetFramework"]}.`,
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
  let commands = [];

  // `dotnet format` not available in .NET 5.x
  if (
    !config.LangCfg["enableFormatting"] &&
    config.LangCfg["dotnetVersion"] != "net5.0"
  ) {
    commands.push({
      command: "dotnet",
      args: ["format"],
    });
  }

  commands.push({
    command: "dotnet",
    args: ["build", "-c", config.Publish ? "Release" : "Debug"],
  });

  return commands;
}

/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies
    "newtonsoft.json": {
      name: "newtonsoft.json",
      version: "13.0.3",
      cpe: "cpe:2.3:a:newtonsoft:json.net:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
    },
    nodatime: {
      name: "nodatime",
      version: "3.1.9",
      cpe: "cpe:2.3:a:nodatime:nodatime:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
    },
    "Microsoft.Bcl.AsyncInterfaces": {
      name: "Microsoft.Bcl.AsyncInterfaces",
      version: "8.0.0",
      cpe: "cpe:2.3:a:microsoft:bcl.asyncinterfaces:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "runtime",
      condition: "netstandard2.0",
    },
    // Dev dependencies (test)
    "Microsoft.NET.Test.Sdk": {
      name: "Microsoft.NET.Test.Sdk",
      version: "17.11.1",
      cpe: "cpe:2.3:a:microsoft:net_test_sdk:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "dev",
      condition: "tests",
    },
    xunit: {
      name: "xunit",
      version: "2.4.2",
      cpe: "cpe:2.3:a:xunit:xunit:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "dev",
      condition: "tests",
    },
    "xunit.runner.visualstudio": {
      name: "xunit.runner.visualstudio",
      version: "2.4.5",
      cpe: "cpe:2.3:a:xunit:xunit_runner_visualstudio:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "dev",
      condition: "tests",
    },
    "coverlet.collector": {
      name: "coverlet.collector",
      version: "3.1.2",
      cpe: "cpe:2.3:a:coverlet:coverlet_collector:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "dev",
      condition: "tests",
    },
    "JUnitXml.TestLogger": {
      name: "JUnitXml.TestLogger",
      version: "3.1.12",
      cpe: "cpe:2.3:a:junitxml:testlogger:*:*:*:*:*:nuget:*:*",
      ecosystem: "NuGet",
      category: "dev",
      condition: "tests",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

// @ts-ignore
function upgradeConfig(_oldVersion, _newVersion, cfg, _defaults) {
  return cfg;
}

/** Post-processes the resolved language config in-place. Runs after defaults
 * are materialized; writes persist into the generated gen.yaml. */
// @ts-ignore
function resolveConfig(cfg: Record<string, any>) {
  resolveUnionStrategy(cfg);
  normalizeForwardCompatibleUnionsConfig(cfg);
}

// @ts-ignore
function normalizeForwardCompatibleUnionsConfig(cfg: Record<string, any>) {
  const mode = cfg.forwardCompatibleUnionsByDefault;
  if (mode === true || mode === "true") {
    cfg.forwardCompatibleUnionsByDefault = "tagged-and-untagged";
  }
}

/** Smart-union scoring (unionStrategy: populated-fields) relies on the Required
 * JsonProperty attribute and OptionalNullable wrapper type, which are only emitted
 * when presenceAwareJsonSerialization is enabled. */
// @ts-ignore
function resolveUnionStrategy(cfg: Record<string, any>) {
  if (
    cfg.unionStrategy === "populated-fields" &&
    cfg.presenceAwareJsonSerialization !== true
  ) {
    console.warn(
      'csharp: unionStrategy "populated-fields" requires presenceAwareJsonSerialization: true. ' +
        'Falling back to "left-to-right".',
    );
    cfg.unionStrategy = "left-to-right";
  }
}

/** Returns the configuration fields available for customers. This data is
 * fetched early in the generation process so values can be passed back to
 * getGeneratorConfig. */
// @ts-ignore
function getConfigFields(
  commonFields: SDKGenConfigFields,
  newSDK: boolean,
): SDKGenConfigFields {
  return {
    ...commonFields,
    version: {
      Name: "version",
      Required: true,
      DefaultValue: "0.0.1",
      Description: "The current version of the SDK",
      ValidationRegex: /^[\w\d.\-_]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    description: {
      Name: "description",
      Required: false,
      Description:
        "The description of the SDK to use in the .csproj file. If not provided, will use the OpenAPI description or summary.",
    },
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "Openapi",
      Description:
        "The NuGet package ID, also used as the root namespace. https://learn.microsoft.com/en-us/dotnet/standard/design-guidelines/names-of-namespaces.",
      ValidationRegex: /^[\w\d._]+$/.source,
      ValidationMessage: "Letters, numbers, or ._ only",
    },
    sourceDirectory: {
      Name: "sourceDirectory",
      Required: false,
      DefaultValue: newSDK ? "src" : "",
      Description: 'The name of the source directory. Default is "src"',
      ValidationRegex: /^\w*$/.source,
      ValidationMessage: "Letters, numbers and underscores only",
    },
    disableNamespacePascalCasingApr2024: {
      Name: "disableNamespacePascalCasingApr2024",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Whether to disable Pascal Casing sanitization on provided packageName when setting the root namespace and NuGet package ID.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    maxMethodParams: {
      Name: "maxMethodParams",
      Required: false,
      DefaultValue: 4,
      Description:
        "The maximum number of parameters a method can have before the resulting SDK endpoint is no longer 'flattened' and an input object is created instead. 0 will use input objects always. https://www.speakeasy.com/docs/customize-sdks/methods",
      ValidationRegex: /^\d+$/.source,
      ValidationMessage: "Numbers only",
    },
    author: {
      Name: "author",
      Required: false,
      DefaultValue: "Speakeasy",
      Description:
        "The name of the author of the published package. https://learn.microsoft.com/en-us/nuget/create-packages/package-authoring-best-practices#authors",
    },
    dotnetVersion: {
      Name: "dotnetVersion",
      Required: false,
      DefaultValue: newSDK ? "net10.0" : "net8.0",
      ValidationRegex: /^(net5\.0|net6\.0|net8\.0|net10\.0)$/.source,
      ValidationMessage: "net5.0, net6.0, net8.0 or net10.0 only",
      Description:
        "The version of .NET to target. net10.0 (default), net8.0, net6.0 and net5.0 supported. https://learn.microsoft.com/en-us/dotnet/standard/frameworks",
    },
    targetFramework: {
      Name: "targetFramework",
      Required: false,
      DefaultValue: "dotnetVersion",
      ValidationRegex:
        /^(netstandard2\.0|net5\.0|net6\.0|net8\.0|net10\.0|dotnetVersion)$/
          .source,
      ValidationMessage:
        "netstandard2.0, net5.0, net6.0, net8.0, net10.0 or dotnetVersion",
      Description:
        "The compile target framework moniker emitted as <TargetFramework>. Must not exceed the configured dotnetVersion. Set 'dotnetVersion' to use the same value as dotnetVersion.",
    },
    flattenGlobalSecurity: {
      Name: "flattenGlobalSecurity",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Flatten the global security configuration if there is only a single option in the spec",
    },
    respectTitlesForPrimitiveUnionMembers: {
      Name: "respectTitlesForPrimitiveUnionMembers",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Use title or x-speakeasy-name-override for primitive union member field names instead of type-based defaults",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    presenceAwareJsonSerialization: {
      Name: "presenceAwareJsonSerialization",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Distinguish required-nullable from optional fields by wrapping optional+nullable fields in OptionalNullable<T> (preserving absent vs. explicit-null), and validate field presence and null-permissiveness during deserialization.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    forwardCompatibleEnumsByDefault: {
      Name: "forwardCompatibleEnumsByDefault",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Generate enums as open (forward-compatible) by default for response types. Open enums can handle unknown values from the API without breaking. Enums with explicit x-speakeasy-unknown-values extension or special patterns (days of week, months) are not affected.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    forwardCompatibleUnionsByDefault: {
      Name: "forwardCompatibleUnionsByDefault",
      Required: false,
      DefaultValue: newSDK ? "tagged-and-untagged" : false,
      Description:
        "Controls forward compatibility for unions in responses. " +
        "'tagged-and-untagged' makes both discriminated and untagged unions open to unknown values with raw JSON fallback. " +
        "'tagged-only' makes only discriminated unions open. " +
        "'false' disables forward compatibility. " +
        "Individual unions can be overridden with x-speakeasy-unknown-values: allow/disallow.",
      ValidationRegex: /^(tagged-and-untagged|true|tagged-only|false)$/.source,
      ValidationMessage: "tagged-and-untagged, tagged-only or false",
    },
    unionStrategy: {
      Name: "unionStrategy",
      Required: false,
      DefaultValue: newSDK ? "populated-fields" : "left-to-right",
      Description:
        "Strategy for deserializing undiscriminated union types. 'left-to-right' tries each type in order and returns the first valid match. 'populated-fields' scores candidates by field-presence match. 'populated-fields' requires presenceAwareJsonSerialization; with presenceAwareJsonSerialization off it is inert and legacy left-to-right behavior is used.",
      ValidationRegex: /^(left-to-right|populated-fields)$/.source,
      ValidationMessage: '"left-to-right" or "populated-fields" only',
    },
    schemaValidation: {
      Name: "schemaValidation",
      Required: false,
      DefaultValue: "strict",
      Description:
        "Controls validation of response, error and streaming bodies against typed models. 'strict' (default) raises ResponseValidationException on a schema mismatch. 'lenient' constructs typed models without raising: missing required fields are defaulted, wrong-typed known fields are skipped, discriminated-union variants with a known discriminator but invalid payload are still constructed as their typed variant, error bodies are constructed on a best-effort basis as their typed error class with the strict failure recorded on DeserializationException, and malformed streaming frames do not abort the stream.",
      ValidationRegex: /^(strict|lenient)$/.source,
      ValidationMessage: '"strict" or "lenient" only',
    },
    clientServerStatusCodesAsErrors: {
      Name: "clientServerStatusCodesAsErrors",
      Required: false,
      DefaultValue: true,
      Description: "Whether to treat 4xx and 5xx status codes as errors.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    imports: {
      Name: "imports",
      Required: false,
      DefaultValue: {
        option: "openapi",
        paths: {
          shared: newSDK ? "Models/Components" : "Models/Shared",
          operations: newSDK ? "Models/Requests" : "Models/Operations",
          errors: "Models/Errors",
          callbacks: "Models/Callbacks",
          webhooks: "Models/Webhooks",
        },
      },
      Description: "Configuration for model import structure",
    },
    responseFormat: {
      Name: "responseFormat",
      Required: false,
      DefaultValue: newSDK ? "envelope-http" : "envelope",
      Description:
        "Determines the shape of the response envelope that is returned from SDK methods",
      ValidationRegex: /^(envelope|envelope-http|flat)$/.source,
      ValidationMessage: '"envelope-http", "envelope" or "flat" only',
    },
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: [],
      Description:
        "Specify additional dependencies to include in the generated .csproj file",
    },
    packageTags: {
      Name: "packageTags",
      Required: false,
      DefaultValue: "",
      Description:
        "Space-delimited list of tags and keywords used when searching for packages on NuGet.",
    },
    includeDebugSymbols: {
      Name: "includeDebugSymbols",
      Required: false,
      DefaultValue: false,
      Description:
        "Whether to generate .pdb files and publish a .snupkg symbol package to NuGet.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    enableSourceLink: {
      Name: "enableSourceLink",
      Required: false,
      DefaultValue: false,
      Description:
        "Whether to produce and publish the package with Source Link. https://github.com/dotnet/sourcelink",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    methodArguments: {
      Name: "methodArguments",
      Required: false,
      DefaultValue: "infer-optional-args",
      Description: "Determines how arguments for SDK methods are generated",
      ValidationRegex: /^infer-optional-args$/.source,
      ValidationMessage: '"infer-optional-args" is the only available style.',
    },
    flatteningOrder: {
      Name: "flatteningOrder",
      Required: false,
      DefaultValue: newSDK ? "parameters-first" : "",
      Description:
        "When flattening parameters and body fields, determines the ordering of generated method arguments. Leave empty to apply legacy ordering.",
      ValidationRegex: /^(parameters-first|body-first)?$/.source,
      ValidationMessage: "parameters-first or body-first.",
    },
    defaultErrorName: {
      Name: "defaultErrorName",
      Required: false,
      DefaultValue: newSDK ? "APIException" : "SDKException",
      Description:
        "The name of the default exception that is thrown when an API error occurs.",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage: "Must only contain letters and numbers",
    },
    httpClientPrefix: {
      Name: "httpClientPrefix",
      Required: false,
      DefaultValue: newSDK ? "" : "Speakeasy",
      Description:
        "The prefix used to name the HTTP client interface: 'I<Prefix>HttpClient'. If left empty, the SDK name will be used instead.",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage: "Must only contain letters and numbers",
    },
    enableCancellationToken: {
      Name: "enableCancellationToken",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Whether to support HTTP cancellation through CancellationToken: https://learn.microsoft.com/en-us/dotnet/api/system.threading.cancellationtoken",
    },
    useNodatime: {
      Name: "useNodatime",
      Required: false,
      DefaultValue: !newSDK,
      Description:
        "Whether to use NodaTime for date and datetime types, forced on for net5.0 and netstandard2.0 targets.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    multipartArrayFormat: {
      Name: "multipartArrayFormat",
      Required: false,
      DefaultValue: newSDK ? "standard" : "legacy",
      Description:
        "Format for array field names in multipart form data. 'legacy' appends '[]' to array field names (e.g., 'files[]'), which was the previous behavior. 'standard' uses the field name as-is without any suffix. Set to 'standard' for correct multipart/form-data handling.",
      ValidationRegex: /^(legacy|standard)$/.source,
      ValidationMessage: "legacy or standard only",
    },
    enableFormatting: {
      Name: "enableFormatting",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Enable embedded clang-format for generated C# files. When disabled, falls back to dotnet format during compilation.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
  };
}

/** Represents the initial target implementation configuration presented to
 * the generator from the target. Customer configuration keys and values from
 * getConfigFields are fetched prior so those values are available. */
// @ts-ignore
function getGeneratorConfig(
  genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  const result: TargetInitialConfiguration = {
    compile: getCompileConfiguration(genConfig),
    testing: getTestingConfiguration(genConfig),
  };

  return result;
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
function getTestingConfiguration(genConfig): TestingConfiguration {
  const testDirectory = "Tests";

  return {
    runner: {
      commands: [
        {
          command: "dotnet",
          args: [
            "test",
            testDirectory,
            "--logger",
            "junit;LogFilePath=../.speakeasy/reports/tests.xml",
          ],
        },
      ],
      dependencies: {
        commands: getCompileDependencies(genConfig),
      },
    },
  };
}
