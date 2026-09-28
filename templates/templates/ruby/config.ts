/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies
    base64: {
      name: "base64",
      version: ">= 0.2.0, < 1.0",
      cpe: "cpe:2.3:a:ruby:base64:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "runtime",
    },
    faraday: {
      name: "faraday",
      version: ">= 2.14.3",
      cpe: "cpe:2.3:a:faraday:faraday:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "runtime",
    },
    "faraday-multipart": {
      name: "faraday-multipart",
      version: "~> 1.2.0",
      cpe: "cpe:2.3:a:faraday:faraday-multipart:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "runtime",
    },
    "faraday-retry": {
      name: "faraday-retry",
      version: "~> 2.4.0",
      cpe: "cpe:2.3:a:faraday:faraday-retry:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "runtime",
    },
    "janeway-jsonpath": {
      name: "janeway-jsonpath",
      version: "~> 0.6.0",
      cpe: "cpe:2.3:a:janeway:janeway-jsonpath:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "runtime",
      condition: "pagination",
    },
    "sorbet-runtime": {
      name: "sorbet-runtime",
      version: "~> 0.6.12872",
      cpe: "cpe:2.3:a:sorbet:sorbet-runtime:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "runtime",
      condition: "sorbet",
    },
    // Dev dependencies
    // irb dependency missing from yard.
    // Reference: https://github.com/lsegal/yard/issues/1636
    irb: {
      name: "irb",
      version: "",
      cpe: "cpe:2.3:a:ruby:irb:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
      condition: "sorbet",
    },
    minitest: {
      name: "minitest",
      // NOTE: Ruby 4.0 includes minitest 6 as a bundled gem, so only set the
      // minimum version instead of pessimistic pinning.
      version: ">= 5.27.0",
      cpe: "cpe:2.3:a:minitest:minitest:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
    },
    "minitest-focus": {
      name: "minitest-focus",
      version: "~> 1.4.1",
      cpe: "cpe:2.3:a:minitest:minitest-focus:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
    },
    "minitest-reporters": {
      name: "minitest-reporters",
      version: "~> 1.7.1",
      cpe: "cpe:2.3:a:minitest:minitest-reporters:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
    },
    rubocop: {
      name: "rubocop",
      version: "~> 1.73.2",
      cpe: "cpe:2.3:a:rubocop:rubocop:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
    },
    "rubocop-minitest": {
      name: "rubocop-minitest",
      version: "~> 0.37.1",
      cpe: "cpe:2.3:a:rubocop:rubocop-minitest:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
    },
    rake: {
      name: "rake",
      version: "",
      cpe: "cpe:2.3:a:rake:rake:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
    },
    sorbet: {
      name: "sorbet",
      version: "~> 0.6.12872",
      cpe: "cpe:2.3:a:sorbet:sorbet:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
      condition: "sorbet",
    },
    tapioca: {
      name: "tapioca",
      version: "~> 0.17.10",
      cpe: "cpe:2.3:a:shopify:tapioca:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
      condition: "sorbet",
    },
    // gems/rbs-4.0.0.dev.4/lib/rbs.rb:11: warning: tsort was loaded from the standard library, but will no longer be part of the default gems starting from Ruby 4.1.0
    // You can add tsort to your Gemfile or gemspec to silence this warning.
    tsort: {
      name: "tsort",
      version: "~> 0.2.0",
      cpe: "cpe:2.3:a:ruby:tsort:*:*:*:*:*:ruby:*:*",
      ecosystem: "RubyGems",
      category: "dev",
      condition: "sorbet",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

/**
 * Called after config fields are resolved to apply target-specific overrides.
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
  enablePopulatedFieldsUnionStrategyIfNeeded(cfg);
}

// @ts-ignore
function enablePopulatedFieldsUnionStrategyIfNeeded(cfg: Record<string, any>) {
  // If forwardCompatibleEnumsByDefault is enabled, always override unionStrategy to "populated-fields"
  if (cfg.forwardCompatibleEnumsByDefault === true) {
    cfg.unionStrategy = "populated-fields";
  }
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

    if (cfg.author) {
      upgraded.author = cfg.author;
    }

    if (cfg.description) {
      upgraded.description = cfg.description;
    }

    if (cfg.namespace) {
      upgraded.namespace = cfg.namespace;
    }

    return upgraded;
  }

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
      DefaultValue: "openapi",
      Description:
        "The distribution name of the Ruby Package. https://guides.rubygems.org/name-your-gem/",
      ValidationRegex: /^[\w\d.\-_]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    author: {
      Name: "author",
      Required: true,
      DefaultValue: "Speakeasy",
      Description: "The name of the author of the published package.",
    },
    description: {
      Name: "description",
      Required: true,
      DefaultValue: "Ruby Client SDK Generated by Speakeasy",
    },
    license: {
      Name: "license",
      Required: true,
      DefaultValue: "Apache-2.0",
      Description:
        "The SPDX license identifier for the gem. Sets `s.licenses` in the generated gemspec. https://guides.rubygems.org/specification-reference/#license=",
    },
    module: {
      Name: "module",
      Required: true,
      DefaultValue: "OpenApiSdk",
      Description:
        "The top level module names for your sdk https://ruby-doc.org/3.2.6/syntax/modules_and_classes_rdoc.html",
      ValidationRegex: /^([A-Z][a-z0-9]+)+$/.source,
      ValidationMessage:
        "Module names should be PascalCase and have no special chars",
    },
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: {
        runtime: {},
        development: {},
      },
      Description:
        "Specify additional dependencies as a mapping to include in the generated gemspec file. There are two top level mappings: runtime and development. The runtime mapping is for dependencies that are required for the SDK to function. The development mapping is for dependencies that are only required for development and testing. Each dependency mapping is the name of the gem to the version constraint.",
    },
    maxMethodParams: {
      Name: "maxMethodParams",
      Required: false,
      DefaultValue: 4,
      Description:
        "The maximum number of parameters a method can have before the resulting SDK endpoint is no longer 'flattened' and an input object is created instead. 0 will use input objects always. https://www.speakeasy.com/docs/using-speakeasy/create-client-sdks/customize-sdks/parameters/",
      ValidationRegex: /^\d+$/.source,
      ValidationMessage: "Numbers only",
    },
    flattenGlobalSecurity: {
      Name: "flattenGlobalSecurity",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Flatten the global security configuration if there is only a single option in the spec",
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
          shared: newSDK ? "models/components" : "models/shared",
          operations: "models/operations",
          errors: "models/errors",
          callbacks: "models/callbacks",
          webhooks: "models/webhooks",
        },
      },
      // Known limitation: the last segment of each scope path must be unique
      // across all paths (e.g., "operations/schemas" and "components/schemas"
      // would both create a ::Schemas module). Model references use relative
      // constant paths like Operations::Schemas::Foo, and Sorbet cannot resolve
      // these when multiple parent modules contain a child with the same name.
      Description: "Configuration for model import structure",
    },
    defaultErrorName: {
      Name: "defaultErrorName",
      Required: false,
      DefaultValue: "APIError",
      Description:
        "The name of the default exception that is thrown when an API error occurs.",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage: "Must only contain letters and numbers",
    },
    constFieldsAlwaysOptional: {
      Name: "constFieldsAlwaysOptional",
      Required: false,
      DefaultValue: false,
      Description:
        "Whether const fields should be treated as optional in Ruby types and schemas regardless of OpenAPI spec requirements. When true (legacy behavior), all const fields are optional. When false (new behavior), const fields respect the OpenAPI spec's required array.",
      ValidationRegex: /^(false)$/.source,
      ValidationMessage: "false only",
    },
    typingStrategy: {
      Name: "typingStrategy",
      Required: false,
      DefaultValue: "sorbet",
      Description:
        "Which type checking strategy to employ. 'sorbet' will generate sorbet types and enable sorbet type checking commands during compilation. 'none' will generate no types and will not run a type checker.",
      ValidationRegex: /^(sorbet|none)$/.source,
      ValidationMessage: "sorbet or none only.",
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
    unionStrategy: {
      Name: "unionStrategy",
      Required: false,
      DefaultValue: newSDK ? "populated-fields" : "left-to-right",
      Description:
        "Strategy for deserializing union types. 'left-to-right' tries each type in order of most required fields first and returns the first valid match. 'populated-fields' tries all types and returns the one with the most matching fields (including optional).",
      ValidationRegex: /^(left-to-right|populated-fields)$/.source,
      ValidationMessage: '"left-to-right" or "populated-fields" only',
    },
    enableFormatting: {
      Name: "enableFormatting",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Enable formatting of generated Ruby source files using rubyfmt.",
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
      DefaultValue: newSDK ? "tagged-only" : "false",
      Description:
        "Controls forward compatibility for discriminated unions in responses. 'tagged-only' makes discriminated unions open to unknown discriminator values. 'false' disables forward compatibility.",
      ValidationRegex: /^(tagged-only|false)$/.source,
      ValidationMessage: "tagged-only or false only",
    },
  };
}

// @ts-ignore
function validateConfig(cfg: Record<string, any>): string {
  if ("additionalDependencies" in cfg) {
    const additionalDependencies: Record<string, any> =
      cfg.additionalDependencies;

    for (const [type, deps] of Object.entries(additionalDependencies)) {
      if (type !== "runtime" && type !== "development") {
        return `additionalDependencies.${type} must be either 'runtime' or 'development'`;
      }

      if (typeof deps !== "object") {
        return `additionalDependencies.${type} must be an object`;
      }

      for (const [name, version] of Object.entries(deps)) {
        if (typeof name !== "string") {
          return `additionalDependencies.${type}.${name} must be a string`;
        }

        if (typeof version !== "string" && !Array.isArray(version)) {
          return `additionalDependencies.${type}.${name} must be an array or a string i.e. 'library', ['>=2.2.0', '<3.0'] OR 'library', '-> 2.2'`;
        }
      }
    }
  }

  return "";
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
    lint: getLintConfiguration(),
    testing: getTestingConfiguration(),
  };

  return result;
}

// @ts-ignore
function getRubyCommandDependency(): RunnerCommandDependency {
  return {
    command: "ruby",
    version: {
      args: ["--version"],
      regex: `(?m)ruby.*?(\\d+\\.\\d+\\.\\d+).*?`,
      minVersion: "3.2.0",
    },
    installDocumentation: `Install Ruby by following the instructions at https://www.ruby-lang.org/en/downloads/.`,
  };
}

// @ts-ignore
function getCompileConfiguration(
  genConfig: GeneratorInitialConfiguration,
): CompileConfiguration {
  const ret = {
    runner: {
      commands: [
        {
          command: "gem",
          args: ["build"],
        },
        {
          command: "bundle",
          args: ["install"],
        },
      ],
      dependencies: {
        commands: [getRubyCommandDependency()],
      },
    },
  };

  if (genConfig.LangCfg.typingStrategy === "sorbet") {
    ret.runner.commands = [
      ...ret.runner.commands,
      {
        command: "bundle",
        args: ["exec", "rake", "sorbet:clean"],
      },
      {
        command: "bundle",
        args: ["exec", "tapioca", "annotations"],
      },
      {
        command: "bundle",
        args: ["exec", "tapioca", "gems"],
      },
      {
        command: "bundle",
        args: ["exec", "srb", "tc", "--ignore=/test"],
      },
    ];
  }
  return ret;
}

// @ts-ignore
function getLintConfiguration(): LintConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "bundle",
          args: ["exec", "rake", "rubocop"],
        },
      ],
      dependencies: {
        commands: [getRubyCommandDependency()],
      },
    },
  };
}

// @ts-ignore
function getTestingConfiguration(): TestingConfiguration {
  return {
    runner: {
      commands: [
        // Reference: internal issue reference
        {
          command: "bundle",
          args: ["install"],
        },
        {
          command: "rake",
          args: ["test"],
        },
      ],
      dependencies: {
        commands: [getRubyCommandDependency()],
      },
    },
  };
}
