// @ts-ignore
function getCompileDependencies(
  config: GeneratorInitialConfiguration,
): RunnerCommandDependencies {
  // TODO determine best way to check Unity dependency
  if ("UNITY_PATH" in config.Env) {
    return [
      {
        command: "dotnet",
        version: {
          args: ["--list-sdks"],
          regex: `(?m).*?(5\\.\\d+\\.\\d+).*?`,
          minVersion: `5.0.0`,
        },
        installDocumentation: `Install .NET 5.0 by following the instructions at https://dotnet.microsoft.com/en-us/download/dotnet/5.0`,
      },
    ];
  }

  return [];
}

// @ts-ignore
function getCompileCommands(
  config: GeneratorInitialConfiguration,
): RunnerCommands {
  if ("UNITY_PATH" in config.Env) {
    return [
      {
        command: "dotnet",
        args: ["build"],
      },
    ];
  }

  return [
    {
      command: "echo",
      args: ["UNITY_PATH env var not found! Skipping compilation."],
    },
  ];
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  return cfg;
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
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "Openapi",
      Description:
        "The NuGet package ID, also used as the root namespace. https://learn.microsoft.com/en-us/dotnet/standard/design-guidelines/names-of-namespaces.",
      ValidationRegex: /^[\w\d._]+$/.source,
      ValidationMessage: "Letters, numbers, or ._ only",
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
      Required: true,
      DefaultValue: "Speakeasy",
      Description:
        "The name of the author of the published package. https://learn.microsoft.com/en-us/nuget/create-packages/package-authoring-best-practices#authors",
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
    defaultErrorName: {
      Name: "defaultErrorName",
      Required: false,
      DefaultValue: newSDK ? "APIException" : "SDKException",
      Description:
        "The name of the default exception that is thrown when an API error occurs.",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage: "Must only contain letters and numbers",
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
