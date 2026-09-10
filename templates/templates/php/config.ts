/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies
    php: {
      name: "php",
      version: ">=8.2",
      cpe: "cpe:2.3:a:php:php:*:*:*:*:*:*:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    "galbar/jsonpath": {
      name: "galbar/jsonpath",
      version: ">=3.0",
      cpe: "cpe:2.3:a:galbar:jsonpath:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    "guzzlehttp/guzzle": {
      name: "guzzlehttp/guzzle",
      version: "^7.15.2",
      cpe: "cpe:2.3:a:guzzlephp:guzzle:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    "speakeasy/serializer": {
      name: "speakeasy/serializer",
      version: "^4.0.3",
      cpe: "cpe:2.3:a:speakeasy:serializer:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    "brick/date-time": {
      name: "brick/date-time",
      version: ">=0.7.0",
      cpe: "cpe:2.3:a:brick:date-time:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    "phpdocumentor/type-resolver": {
      name: "phpdocumentor/type-resolver",
      version: ">=1.8",
      cpe: "cpe:2.3:a:phpdocumentor:type-resolver:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    "brick/math": {
      name: "brick/math",
      version: ">=0.12.1",
      cpe: "cpe:2.3:a:brick:math:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "runtime",
    },
    // Dev dependencies
    "laravel/pint": {
      name: "laravel/pint",
      version: "1.29.0",
      cpe: "cpe:2.3:a:laravel:pint:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "dev",
    },
    "phpstan/phpstan": {
      name: "phpstan/phpstan",
      version: "2.1.44",
      cpe: "cpe:2.3:a:phpstan:phpstan:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "dev",
    },
    "phpunit/phpunit": {
      name: "phpunit/phpunit",
      version: "^11.5.50 || ^12.5.8 || >=13.0.0",
      cpe: "cpe:2.3:a:phpunit:phpunit:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "dev",
    },
    "roave/security-advisories": {
      name: "roave/security-advisories",
      version: "dev-latest",
      cpe: "cpe:2.3:a:roave:security-advisories:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "dev",
    },
    "orchestra/testbench": {
      name: "orchestra/testbench",
      version: ">=9.6",
      cpe: "cpe:2.3:a:orchestra:testbench:*:*:*:*:*:php:*:*",
      ecosystem: "Packagist",
      category: "dev",
      condition: "laravelServiceProvider",
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
      DefaultValue: "openapi/openapi",
      Description:
        "The name of the composer package. https://getcomposer.org/doc/04-schema.md#name",
      ValidationRegex:
        /^[a-z0-9]([_.-]?[a-z0-9]+)*\/[a-z0-9](([_.]?|-{0,2})[a-z0-9]+)*/
          .source,
      ValidationMessage:
        "Package name must match the pattern given here https://getcomposer.org/doc/04-schema.md#name",
    },
    namespace: {
      Name: "namespace",
      Required: true,
      DefaultValue: "OpenAPI\\OpenAPI",
      Description:
        "https://www.php.net/manual/en/language.namespaces.rationale.php",
      ValidationRegex:
        /^([a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*\\)*[a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*$/
          .source,
      ValidationMessage:
        "Each part of namespace must match the pattern given here https://www.php.net/manual/en/language.variables.basics.php",
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
    flattenGlobalSecurity: {
      Name: "flattenGlobalSecurity",
      Required: false,
      DefaultValue: true,
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
          shared: newSDK ? "Models/Components" : "Models/Shared",
          operations: "Models/Operations",
          errors: "Models/Errors",
          callbacks: "Models/Callbacks",
          webhooks: "Models/Webhooks",
        },
      },
      Description: "Configuration for model import structure",
    },
    laravelServiceProvider: {
      Name: "laravelServiceProvider",
      Required: false,
      DefaultValue: {
        enabled: false,
        svcName: "openapi",
      },
      Description:
        "This determined whether a laravelServiceProvider should be generated.  In order to enable generation, set the `enabled` flag to true, and provide a useful value for `svcName`.",
    },
    methodArguments: {
      Name: "methodArguments",
      Required: false,
      DefaultValue: "infer-optional-args",
      Description:
        "Determines how arguments for SDK methods are generated.  PHP only supports `infer-optional-args` - this configuration option is only here for consistency.",
      ValidationRegex: /^infer-optional-args$/.source,
      ValidationMessage: '"infer-optional-args"',
    },
    envVarPrefix: {
      Name: "envVarPrefix",
      Required: false,
      Description:
        "The environment variable prefix for laravel service provider env variable overrides. If empty these overrides will not be prefixed",
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
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: {
        autoload: {},
        "autoload-dev": {},
        require: {},
        "require-dev": {},
      },
      Description:
        "Specify additional dependencies to include in the generated composer.json file.",
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
function getPhpCommandDependencies(): RunnerCommandDependencies {
  return [
    {
      command: "php",
      version: {
        args: ["--version"],
        regex: `(?m)PHP.*?(\\d+\\.\\d+\\.\\d+).*?`,
        minVersion: "8.2.0",
      },
      installDocumentation: `Install PHP by following the instructions at https://www.php.net/manual/en/install.php.`,
    },
    {
      command: "composer",
      version: {
        args: ["--version"],
        regex: `(?m)Composer version.*?(\\d+\\.\\d+\\.\\d+).*?`,
        minVersion: "2.5.0",
      },
      installDocumentation: `Install Composer by following the instructions at https://getcomposer.org/download/.`,
    },
  ];
}

// @ts-ignore
function getCompileConfiguration(): CompileConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "composer",
          args: ["update"],
        },
      ],
      dependencies: {
        commands: getPhpCommandDependencies(),
      },
    },
  };
}

// @ts-ignore
function getLintConfiguration(): CompileConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "vendor/bin/phpstan",
          args: "analyse src --level 7 --memory-limit 1G --no-progress --error-format=table".split(
            " ",
          ),
        },
        {
          command: "php",
          args: "-d memory_limit=-1 vendor/bin/pint src --test -vvv".split(" "),
        },
      ],
      dependencies: {
        commands: getPhpCommandDependencies(),
      },
    },
  };
}

// @ts-ignore
function getTestingConfiguration(): TestingConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "composer",
          args: ["run", "test"],
        },
      ],
      dependencies: {
        commands: getPhpCommandDependencies(),
      },
    },
  };
}

// @ts-ignore
function validateConfig(cfg: Record<string, any>): string {
  if ("additionalDependencies" in cfg) {
    const additionalDependencies: Record<string, any> =
      cfg.additionalDependencies;

    for (const [type, deps] of Object.entries(additionalDependencies)) {
      if (typeof deps !== "object") {
        return `additionalDependencies.${type} must be an object`;
      }

      for (const [name, version] of Object.entries(deps)) {
        if (typeof name !== "string") {
          return `additionalDependencies.${type}.${name} must be a string`;
        }

        if (typeof version !== "string") {
          return `additionalDependencies.${type}.${name} must be a string`;
        }
      }
    }
  }

  return "";
}
