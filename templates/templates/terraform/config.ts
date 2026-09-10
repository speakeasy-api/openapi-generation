/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies (Terraform framework)
    "github.com/hashicorp/go-uuid": {
      name: "github.com/hashicorp/go-uuid",
      version: "v1.0.3",
      cpe: "cpe:2.3:a:hashicorp:go-uuid:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/hashicorp/terraform-plugin-docs": {
      name: "github.com/hashicorp/terraform-plugin-docs",
      version: "v0.25.0",
      cpe: "cpe:2.3:a:hashicorp:terraform-plugin-docs:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "dev",
    },
    "github.com/hashicorp/terraform-plugin-framework-validators": {
      name: "github.com/hashicorp/terraform-plugin-framework-validators",
      version: "v0.19.0",
      cpe: "cpe:2.3:a:hashicorp:terraform-plugin-framework-validators:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/hashicorp/terraform-plugin-framework": {
      name: "github.com/hashicorp/terraform-plugin-framework",
      version: "v1.19.0",
      cpe: "cpe:2.3:a:hashicorp:terraform-plugin-framework:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/hashicorp/terraform-plugin-go": {
      name: "github.com/hashicorp/terraform-plugin-go",
      version: "v0.31.0",
      cpe: "cpe:2.3:a:hashicorp:terraform-plugin-go:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/hashicorp/terraform-plugin-log": {
      name: "github.com/hashicorp/terraform-plugin-log",
      version: "v0.11.0",
      cpe: "cpe:2.3:a:hashicorp:terraform-plugin-log:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    // Security floor pins for vulnerable transitive modules required by the Terraform dependencies.
    "golang.org/x/crypto": {
      name: "golang.org/x/crypto",
      version: "v0.52.0",
      cpe: "cpe:2.3:a:golang:crypto:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "golang.org/x/net": {
      name: "golang.org/x/net",
      version: "v0.56.0",
      cpe: "cpe:2.3:a:golang:networking:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "google.golang.org/grpc": {
      name: "google.golang.org/grpc",
      version: "v1.82.1",
      cpe: "cpe:2.3:a:grpc:grpc:*:*:*:*:*:go:*:*",
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
    "github.com/spyzhov/ajson": {
      name: "github.com/spyzhov/ajson",
      version: "v0.9.0",
      cpe: "cpe:2.3:a:spyzhov:ajson:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "pagination",
    },
    "github.com/itchyny/gojq": {
      name: "github.com/itchyny/gojq",
      version: "v0.12.17",
      cpe: "cpe:2.3:a:itchyny:gojq:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "transformJq",
    },
    "github.com/hashicorp/terraform-plugin-framework-jsontypes": {
      name: "github.com/hashicorp/terraform-plugin-framework-jsontypes",
      version: "v0.2.0",
      cpe: "cpe:2.3:a:hashicorp:terraform-plugin-framework-jsontypes:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "typeOverrides",
    },
    "github.com/speakeasy-api/terraform-plugin-framework-base64types": {
      name: "github.com/speakeasy-api/terraform-plugin-framework-base64types",
      version: "v0.1.0",
      cpe: "cpe:2.3:a:speakeasy-api:terraform-plugin-framework-base64types:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "formatBinary",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

/**
 * Minimum compatible versions for specific additionalDependencies.
 *
 * These minimums prevent compilation errors caused by terraform-plugin-go
 * interface changes that propagate through the dependency tree. When
 * terraform-plugin-go is bumped (currently at v0.31.0 in
 * getTemplateDependencies), downstream modules like terraform-plugin-sdk/v2
 * and terraform-plugin-testing must also meet a minimum version to implement
 * the updated ProviderServer interface.
 *
 * Only dependencies already present in the customer's additionalDependencies
 * configuration are checked — new dependencies are never added automatically.
 *
 * Update these minimums when bumping terraform-plugin-go or
 * terraform-plugin-framework in getTemplateDependencies above. The
 * scripts/bump-terraform-deps.sh script can automate this.
 */
const additionalDependencyMinimumVersions: Record<string, string> = {
  "github.com/hashicorp/terraform-plugin-sdk/v2": "v2.40.1",
  "github.com/hashicorp/terraform-plugin-testing": "v1.16.0",
  // Security floor pins: additionalDependencies merge after the template
  // defaults, so without these minimums a customer explicitly pinning an
  // older version would silently downgrade below the patched releases.
  "golang.org/x/crypto": "v0.52.0",
  "golang.org/x/net": "v0.56.0",
  "google.golang.org/grpc": "v1.82.1",
};

/**
 * Resolves configuration before generation, persisting changes back to
 * gen.yaml. Called automatically by the generation pipeline.
 *
 * Currently enforces minimum versions for specific additionalDependencies
 * to prevent compilation errors from terraform-plugin-go interface changes.
 *
 * - See pkg/templates/config.go for the Go-side implementation details.
 */
// @ts-ignore
function resolveConfig(cfg: Record<string, any>) {
  enforceAdditionalDependencyMinimumVersions(cfg);
}

function enforceAdditionalDependencyMinimumVersions(
  cfg: Record<string, any>,
): void {
  const additionalDeps = cfg.additionalDependencies;
  if (
    additionalDeps === undefined ||
    additionalDeps === null ||
    typeof additionalDeps !== "object"
  ) {
    return;
  }

  for (const [mod, minVersion] of Object.entries(
    additionalDependencyMinimumVersions,
  )) {
    const currentVersion = additionalDeps[mod];
    if (typeof currentVersion !== "string") {
      continue;
    }

    if (compareVersions(currentVersion, minVersion) < 0) {
      additionalDeps[mod] = minVersion;
    }
  }
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  return cfg;
}

// @ts-ignore
function resolveConfig(cfg: Record<string, any>) {
  enforceAdditionalDependencyMinimumVersions(cfg);
  enablePopulatedFieldsUnionStrategyIfNeeded(cfg);
}

// @ts-ignore
function enablePopulatedFieldsUnionStrategyIfNeeded(cfg: Record<string, any>) {
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
    version: {
      Name: "version",
      Required: true,
      DefaultValue: "0.0.1",
      Description: "Current version of the Terraform Provider.",
      ValidationRegex: /^[\w\d.\-_]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    environmentVariables: {
      Name: "environmentVariables",
      Required: false,
      DefaultValue: [],
      Description:
        "A list of objects with [env: string, providerAttribute: string] keys/values to associate environment variables [env] with a provider variable [providerAttribute].",
    },
    serverVariableProviderAttributes: {
      Name: "serverVariableProviderAttributes",
      Required: false,
      DefaultValue: {},
      Description:
        "An object mapping OAS server variable names to explicit provider attribute names, for variables whose sanitized name collides with a Terraform reserved provider-schema name (for example version).",
    },
    additionalProviderAttributes: {
      Name: "additionalProviderAttributes",
      Required: false,
      DefaultValue: {
        httpHeaders: "",
        tlsSkipVerify: "",
      },
      Description:
        "An object with specific properties to configure additional provider attributes, such as custom HTTP client headers.",
    },
    enableOperationSecurity: {
      Name: "enableOperationSecurity",
      Required: false,
      DefaultValue: false,
      Description:
        "Enables configurable resource security support for OAS operation-based `security` configuration. When operation security is detected, a configurable attributes for the security scheme are added to the resource schema and the provider security configuration is overridden with the resource configuration when the API operation is called.",
    },
    enableOperationServers: {
      Name: "enableOperationServers",
      Required: false,
      DefaultValue: false,
      Description:
        "Enables configurable resource server URL support for OAS path-based and operation-based `servers` configuration. When an operation server is detected, a configurable server URL attribute is added to the resource schema and the provider server configuration is overridden with the resource configuration value when the API operation is called.",
    },
    enableTypeDeduplication: {
      Name: "enableTypeDeduplication",
      Required: false,
      DefaultValue: !newSDK,
      Description: "Enables deduplication of terraform value types",
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
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: {},
      Description:
        "Specify additional dependencies to include in the generated go.mod",
    },
    additionalResources: {
      Name: "additionalResources",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of { importLocation?: string, importAlias?: string, resource: string } objects. Each `resource` will be inserted into the provider managed resource type list. Use importLocation / importAlias to customize the provider.go import statement when the implementation is outside the internal/provider Go package.",
    },
    additionalDataSources: {
      Name: "additionalDataSources",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of { importLocation?: string, importAlias?: string, datasource: string } objects. Each `datasource` will be inserted into the provider data resource type list. Use importLocation / importAlias to customize the provider.go import statement when the implementation is outside the internal/provider Go package.",
    },
    additionalEphemeralResources: {
      Name: "additionalEphemeralResources",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of { importLocation?: string, importAlias?: string, resource: string } objects. Each `resource` will be inserted into the provider ephemeral resource type list. Use importLocation / importAlias to customize the provider.go import statement when the implementation is outside the internal/provider Go package.",
    },
    additionalListResources: {
      Name: "additionalListResources",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of { importLocation?: string, importAlias?: string, resource: string } objects. Each `resource` will be inserted into the provider list resource type list. Use importLocation / importAlias to customize the provider.go import statement when the implementation is outside the internal/provider Go package.",
    },
    additionalFunctions: {
      Name: "additionalFunctions",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of { importLocation?: string, importAlias?: string, function: string } objects. Each `function` will be inserted into the provider Go package. Use importLocation / importAlias to customize the provider.go import statement when the implementation is outside the internal/provider Go package.",
    },
    additionalActions: {
      Name: "additionalActions",
      Required: false,
      DefaultValue: [],
      Description:
        "Array of { importLocation?: string, importAlias?: string, action: string } objects. Each `action` will be inserted into the provider action type list. Use importLocation / importAlias to customize the provider.go import statement when the implementation is outside the internal/provider Go package.",
    },
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "terraform",
      Description:
        "Terraform Provider name. Prefixes all resource names unless providerTypeNameOverride is set. For providers published in the public Terraform Registry, this typically matches the suffix after 'terraform-provider-' in the GitHub Repository name.",
      ValidationRegex: /^[\w\d-]+$/.source,
      ValidationMessage: "Letters, numbers, or - only",
    },
    providerTypeNameOverride: {
      Name: "providerTypeNameOverride",
      Required: false,
      DefaultValue: "",
      Description:
        "Overrides the provider type name used for resource naming. When set, this value is used as the prefix for all resource, data source, ephemeral resource, and action type names instead of packageName. This is useful for variant providers that need to use the same resource names as another provider (e.g. a beta provider sharing resource names with the main provider).",
      ValidationRegex: /^[\w\d-]*$/.source,
      ValidationMessage: "Letters, numbers, or - only",
    },
    author: {
      Name: "author",
      Required: true,
      DefaultValue: "speakeasy",
      Description:
        "Terraform Provider namespace. For providers published in the public Terraform Registry, this typically matches the GitHub Organization name.",
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
    excludeEmptyObjectSchemas: {
      Name: "excludeEmptyObjectSchemas",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Exclude empty SingleNestedAttributes in generated terraform schema.",
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
    unionStrategy: {
      Name: "unionStrategy",
      Required: false,
      DefaultValue: "populated-fields",
      Description:
        "Strategy for deserializing union types. 'left-to-right' tries each type in order of most required fields first and returns the first valid match. 'populated-fields' tries all types and returns the one with the most matching fields (including optional). Falls back to biggest stringified size and beyond that to left-to-right.",
      ValidationRegex: /^(left-to-right|populated-fields)$/.source,
      ValidationMessage: '"left-to-right" or "populated-fields" only',
    },
    inferUnionDiscriminators: {
      Name: "inferUnionDiscriminators",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Infer union discriminators for oneOfs missing explicit OpenAPI discriminator mapping",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    preApplyUnionDiscriminators: {
      Name: "preApplyUnionDiscriminators",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "When true, discriminator values will be applied as const fields onto union member types when those types are consistently used with the same discriminator value. This removes the discriminator field from the Terraform schema, since the HCL structure already implies which variant is selected.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    enableCustomCodeRegions: {
      Name: "enableCustomCodeRegions",
      Required: false,
      DefaultValue: false,
      Description:
        "Generate custom code region comments in the Terraform Provider, which allow custom code to persist across regenerations.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    forwardCompatibleEnumsByDefault: {
      Name: "forwardCompatibleEnumsByDefault",
      Required: false,
      DefaultValue: false,
      Description:
        "Generate enums as open (forward-compatible) by default for response types. Open enums can handle unknown values from the API without breaking. Enums with explicit x-speakeasy-unknown-values extension or special patterns (days of week, months) are not affected.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    debugLogging: {
      Name: "debugLogging",
      Required: false,
      DefaultValue: {},
      Description:
        "Configuration options for Terraform debug logging (TF_LOG=DEBUG). Contains settings for controlling what information is redacted or exposed in debug output.",
    },
  };
}

function getConfigOverlay(): Record<string, any> {
  return {
    inputModelSuffix: "input",
    outputModelSuffix: "output",
    flattenGlobalSecurity: false,
    methodArguments: "require-security-and-request",
    maxMethodParams: 0,
    imports: {
      option: "openapi",
      paths: {
        shared: "models/shared",
        callbacks: "models/callbacks",
        errors: "models/errors",
        operations: "models/operations",
        webhooks: "models/webhooks",
      },
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
      minVersion: "1.25.0",
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

  commands.push(
    {
      command: "go",
      args: ["build", "./..."],
      environmentVariables: {
        GOROOT: "",
      },
    },
    {
      command: "go",
      args: ["generate", "./..."],
      environmentVariables: {
        GOROOT: "",
      },
    },
  );

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
function getTestingConfiguration(): TestingConfiguration {
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
  };
}
