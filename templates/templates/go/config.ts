/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies
    "github.com/spyzhov/ajson": {
      name: "github.com/spyzhov/ajson",
      version: "v0.8.0",
      cpe: "cpe:2.3:a:spyzhov:ajson:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/ericlagergren/decimal": {
      name: "github.com/ericlagergren/decimal",
      version: "v0.0.0-20221120152707-495c53812d05",
      cpe: "cpe:2.3:a:ericlagergren:decimal:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "decimal",
    },
    "github.com/stretchr/testify": {
      name: "github.com/stretchr/testify",
      version: "v1.8.4",
      cpe: "cpe:2.3:a:stretchr:testify:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "dev",
      condition: "tests",
    },
    "golang.org/x/sync": {
      name: "golang.org/x/sync",
      version: "v0.8.0",
      cpe: "cpe:2.3:a:golang:sync:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "oauth2",
    },
    "github.com/itchyny/gojq": {
      name: "github.com/itchyny/gojq",
      version: "v0.12.17",
      cpe: "cpe:2.3:a:itchyny:gojq:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "transformJq",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  if (oldVersion == "") {
    let upgraded = defaults;

    if (cfg.version) {
      upgraded.version = cfg.version;
    }

    if (cfg.packagename) {
      upgraded.packageName = cfg.packagename;
    }

    return upgraded;
  }

  return cfg;
}

/**
 * Resolves and applies target-specific configuration transformations.
 *
 * This function is called after the gen.yaml configuration has been loaded and
 * initialized, providing a late-stage hook for applying target-specific logic
 * and configuration adjustments.
 *
 * @param cfg - The target-specific language configuration object. This corresponds
 *              to the language configuration section for this target (e.g., the
 *              'go' section in gen.yaml).
 *
 * @remarks
 * - Modifications to the cfg object are automatically reflected in the Go code
 *   due to Goja's reference handling of the underlying Go map.
 * - This function should contain target-specific logic such as feature flag
 *   resolution, compatibility checks, and dynamic configuration adjustments.
 * - See pkg/templates/config.go for the Go-side implementation details.
 */
// @ts-ignore
function resolveConfig(cfg: Record<string, any>) {
  normalizeForwardCompatibleUnionsConfig(cfg);
  enablePopulatedFieldsUnionStrategyIfNeeded(cfg);
}

// @ts-ignore
function normalizeForwardCompatibleUnionsConfig(cfg: Record<string, any>) {
  if (cfg.forwardCompatibleUnionsByDefault === true) {
    cfg.forwardCompatibleUnionsByDefault = "tagged-and-untagged";
    return;
  }
  if (cfg.forwardCompatibleUnionsByDefault === false) {
    cfg.forwardCompatibleUnionsByDefault = "false";
  }
}

// @ts-ignore
function enablePopulatedFieldsUnionStrategyIfNeeded(cfg: Record<string, any>) {
  // If forwardCompatibleEnumsByDefault is enabled, always override unionStrategy to "populated-fields"
  if (cfg.forwardCompatibleEnumsByDefault === true) {
    cfg.unionStrategy = "populated-fields";
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
    modulePath: {
      Name: "modulePath",
      Required: true,
      DefaultValue: "",
      Description:
        "Root module path. Use sdkPackageName to configure the package clause for the root module package. https://go.dev/ref/mod#module-path.",
      ValidationRegex: /^([\w\d\-~]([\w\d.\-_\/~]*[\w\d\-~])?)?$/.source,
      ValidationMessage:
        "Letters, numbers, or /.-_~ only. Cannot start or end with slash or dot. For more information: https://go.dev/ref/mod#go-mod-file-ident",
    },
    packageName: {
      Name: "packageName",
      Required: false,
      Description:
        "Legacy combined root module path and SDK package naming. Use sdkPackageAlias to update SDK package import aliases in documentation while preserving major version compatibility, otherwise migrate to modulePath and sdkPackageName. https://go.dev/ref/mod#module-path.",
      ValidationRegex: /^([\w\d\-~]([\w\d.\-_\/~]*[\w\d\-~])?)?$/.source,
      ValidationMessage:
        "Letters, numbers, or /.-_~ only. Cannot start or end with slash or dot. For more information: https://go.dev/ref/mod#go-mod-file-ident",
    },
    sdkPackageAlias: {
      Name: "sdkPackageAlias",
      Required: false,
      Description:
        "Root module package import alias for documentation. Use this to preserve compatibility if the SDK has already had a stable major version release with modulePath, packageName, or sdkPackageName, as the package clause determines package naming in consuming code if the import path does not end in a valid identifier. https://go.dev/ref/spec#Packages.",
      ValidationRegex: /^(\w[\w\d_]*)?$/.source,
      ValidationMessage:
        "Letters, numbers, or underscore (_) only. Must start with letter. For more information: https://go.dev/ref/spec#Packages",
    },
    sdkPackageName: {
      Name: "sdkPackageName",
      Required: true,
      DefaultValue: "",
      Description:
        "Root module package name written in the package clause. Determines the package naming in consuming code if the modulePath does not end with a valid identifier. https://go.dev/ref/spec#Packages.",
      ValidationRegex: /^(\w[\w\d_]*)?$/.source,
      ValidationMessage:
        "Letters, numbers, or underscore (_) only. Must start with letter. For more information: https://go.dev/ref/spec#Packages",
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
    clientServerStatusCodesAsErrors: {
      Name: "clientServerStatusCodesAsErrors",
      Required: false,
      DefaultValue: true,
      Description: "Whether to treat 4xx and 5xx status codes as errors.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
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
    flattenGlobalSecurity: {
      Name: "flattenGlobalSecurity",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Flatten the global security configuration if there is only a single option in the spec",
    },
    imports: {
      Name: "imports",
      Required: false,
      DefaultValue: {
        option: "openapi",
        paths: {
          shared: newSDK ? "models/components" : "pkg/models/shared",
          operations: newSDK ? "models/operations" : "pkg/models/operations",
          errors: newSDK ? "models/apierrors" : "pkg/models/sdkerrors",
          callbacks: newSDK ? "models/callbacks" : "pkg/models/callbacks",
          webhooks: newSDK ? "models/webhooks" : "pkg/models/webhooks",
        },
      },
      Description: "Configuration for model import structure",
    },
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: {},
      Description:
        "Specify additional dependencies to include in the generated go.mod",
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
    enableSkipDeserialization: {
      Name: "enableSkipDeserialization",
      Required: false,
      DefaultValue: false,
      Description:
        "Generate a WithSkipDeserialization() option that skips typed deserialization of successful JSON response bodies; the body is buffered and replayed on the HTTP response so callers can read the raw bytes via HTTPMeta.Response after the call returns. Error responses and non-JSON responses are always deserialized. Requires responseFormat 'envelope-http'.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    methodArguments: {
      Name: "methodArguments",
      Required: false,
      DefaultValue: "require-security-and-request",
      Description: "Determines how arguments for SDK methods are generated",
      ValidationRegex: /^(infer-optional-args|require-security-and-request)$/
        .source,
      ValidationMessage:
        '"infer-optional-args" or "require-security-and-request" only',
    },
    optionalMethodArguments: {
      Name: "optionalMethodArguments",
      Required: false,
      DefaultValue: "pointers",
      Description:
        "Determines how optional SDK method arguments are generated. 'pointers' preserves positional pointer arguments, 'shared-options' uses operation-qualified functional options with the shared per-request option type, and 'method-options' uses operation-specific functional option types. Changing this value for an existing SDK is a breaking API change.",
      ValidationRegex: /^(pointers|shared-options|method-options)$/.source,
      ValidationMessage:
        '"pointers", "shared-options", or "method-options" only',
    },
    envVarPrefix: {
      Name: "envVarPrefix",
      Required: false,
      Description:
        "The environment variable prefix for security and global env variable overrides. If empty these overrides will not be possible",
    },
    retractions: {
      Name: "retractions",
      Required: false,
      Description:
        "Specify Go module retractions to include in the generated go.mod. Each retraction should have a 'version' field and optional 'comment' field.",
      ValidationRegex:
        /^v\d+\.\d+\.\d+(-[a-zA-Z0-9\-_.]+)?(\+[a-zA-Z0-9\-_.]+)?$/.source,
      ValidationMessage:
        "Each version must follow semantic versioning format starting with 'v' (e.g., 'v1.0.0', 'v1.2.3-beta.1'). Example: [{version: 'v1.0.0', comment: 'Published accidentally'}, {version: 'v1.0.1', comment: 'Contains retractions only'}]",
    },
    includeEmptyObjects: {
      Name: "includeEmptyObjects",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Include empty objects in serialized request bodies and parameters. When false, optional fields will be excluded when empty.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    defaultErrorName: {
      Name: "defaultErrorName",
      Required: false,
      DefaultValue: newSDK ? "APIError" : "SDKError",
      Description:
        "The name of the default error type used to represent API errors",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage: "Must only contain letters and numbers",
    },
    respectRequiredFields: {
      Name: "respectRequiredFields",
      Required: false,
      DefaultValue: false,
      Description: "Respect required fields in the spec",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    nullableOptionalWrapper: {
      Name: "nullableOptionalWrapper",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Generate nullable optional wrappers for fields that need to distinguish between null and unset",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    idiomaticMethodCollisionNames: {
      Name: "idiomaticMethodCollisionNames",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Resolve field names that collide with generated struct methods (e.g. Error()) using idiomatic replacements (e.g. ErrorInfo) instead of a trailing underscore (Error_)",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    unionStrategy: {
      Name: "unionStrategy",
      Required: false,
      DefaultValue: newSDK ? "populated-fields" : "left-to-right",
      Description:
        "Strategy for deserializing union types. 'left-to-right' tries each type in order of most required fields first and returns the first valid match. 'populated-fields' tries all types and returns the one with the most matching fields (including optional). Falls back to biggest stringified size and beyond that to left-to-right.",
      ValidationRegex: /^(left-to-right|populated-fields)$/.source,
      ValidationMessage: '"left-to-right" or "populated-fields" only',
    },
    unionGenerics: {
      Name: "unionGenerics",
      Required: false,
      DefaultValue: false,
      Description:
        "Generate a single generic 'New<Union>[T <Union>Member](val T)' constructor per union type instead of per-member 'Create<Union><Member>' factory functions. Falls back to per-member constructors for unions that generics cannot express.",
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
    forwardCompatibleEnumsByDefault: {
      Name: "forwardCompatibleEnumsByDefault",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Generate enums as open (forward-compatible) by default for response types. Open enums can handle unknown values from the API without breaking. Enums with explicit x-speakeasy-unknown-values extension or special patterns (days of week, months) are not affected.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    enableCustomCodeRegions: {
      Name: "enableCustomCodeRegions",
      Required: false,
      DefaultValue: false,
      Description: "Allow custom code to be inserted into the generated SDK.",
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
      ValidationRegex: /^(tagged-and-untagged|tagged-only|false)$/.source,
      ValidationMessage: "tagged-and-untagged, tagged-only, or false only",
    },
  };
}

/** Represents the initial target implementation configuration presented to
 * the generator from the target. Customer configuration keys and values from
 * getConfigFields are fetched prior so those values are available. */
// @ts-ignore
function getGeneratorConfig(
  _genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  const result: TargetInitialConfiguration = {
    compile: getCompileConfiguration(),
    lint: getLintConfiguration(),
    testing: getTestingConfiguration(),
  };

  return result;
}

// @ts-ignore
function getGoCommandDependency(): RunnerCommandDependency {
  return {
    command: "go",
    version: {
      args: ["version"],
      regex: `(?m).*?go version go(\\d+\\.\\d+\\.\\d+).*?`,
      minVersion: "1.22.0",
    },
    installDocumentation: `Install Go by following the instructions at https://golang.org/doc/install.`,
  };
}

// @ts-ignore
function getCompileConfiguration(): CompileConfiguration {
  const commands: RunnerCommands = [
    {
      command: "go",
      args: ["mod", "tidy"],
      environmentVariables: {
        GOROOT: "",
      },
    },
  ];

  if (directoryExists("vendor") && fileExists("vendor/modules.txt")) {
    commands.push({
      command: "go",
      args: ["mod", "vendor"],
      environmentVariables: {
        GOROOT: "",
      },
    });
  }

  commands.push({
    command: "go",
    args: ["build", "./..."],
    environmentVariables: {
      GOROOT: "",
    },
  });

  return {
    runner: {
      commands: commands,
      dependencies: {
        commands: [getGoCommandDependency()],
      },
    },
  };
}

// @ts-ignore
function getLintConfiguration(): LintConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "go",
          args: [
            "run",
            "honnef.co/go/tools/cmd/staticcheck@v0.8.1",
            "-checks=inherit,-SA1019,-SA5008",
            "./...",
          ],
          environmentVariables: {
            GOROOT: "",
          },
        },
      ],
      dependencies: {
        commands: [getGoCommandDependency()],
      },
    },
  };
}

// @ts-ignore
function validateConfig(cfg: Record<string, any>): string {
  const optionalMethodArguments = cfg.optionalMethodArguments;
  if (
    optionalMethodArguments !== undefined &&
    optionalMethodArguments !== null &&
    optionalMethodArguments !== "" &&
    optionalMethodArguments !== "pointers" &&
    optionalMethodArguments !== "shared-options" &&
    optionalMethodArguments !== "method-options"
  ) {
    return 'optionalMethodArguments must be "pointers", "shared-options", or "method-options"';
  }

  return "";
}

// @ts-ignore
function getTestingConfiguration(): TestingConfiguration {
  // Ideally, this would just use the getTestDirectory() function from tests.ts,
  // however the pared down config.ts environment does not include common.d.ts
  // and importing that file has its own issues. So for now, hardcode around
  // these issues.
  // const testDirectory = getTestDirectory();
  const testDirectory = "tests";

  return {
    compile: {
      runner: {
        commands: [
          {
            command: "go",
            // The -run flag is a workaround to compile the tests without running them. https://stackoverflow.com/a/72722257
            args: ["test", "-run=XXX_SHOULD_NEVER_MATCH_XXX", "./..."],
            environmentVariables: {
              GOROOT: "",
            },
          },
        ],
        dependencies: {
          commands: [getGoCommandDependency()],
        },
      },
    },
    runner: {
      commands: [
        {
          command: "go",
          args: ["install", "gotest.tools/gotestsum@latest"],
          environmentVariables: {
            GOROOT: "",
          },
        },
        {
          command: "gotestsum",
          // -count=1 to prevent caching of test results
          args: [
            "--junitfile",
            "./.speakeasy/reports/tests.xml", // TODO: the output folder for the tests report should probably be templated based on whether .speakeasy or .gen is used
            `./...`,
          ],
          environmentVariables: {
            GOROOT: "",
          },
        },
      ],
      dependencies: {
        commands: [getGoCommandDependency()],
      },
    },
  };
}
