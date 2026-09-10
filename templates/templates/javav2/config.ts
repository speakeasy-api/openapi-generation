/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies() {
  return {
    // Runtime dependencies (api/implementation scope)
    "com.fasterxml.jackson.core:jackson-annotations": {
      name: "com.fasterxml.jackson.core:jackson-annotations",
      version: "2.22",
      cpe: "cpe:2.3:a:fasterxml:jackson-annotations:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
    },
    "com.fasterxml.jackson.core:jackson-databind": {
      name: "com.fasterxml.jackson.core:jackson-databind",
      version: "2.22.1",
      cpe: "cpe:2.3:a:fasterxml:jackson-databind:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
    },
    "com.fasterxml.jackson.datatype:jackson-datatype-jsr310": {
      name: "com.fasterxml.jackson.datatype:jackson-datatype-jsr310",
      version: "2.22.0",
      cpe: "cpe:2.3:a:fasterxml:jackson-datatype-jsr310:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
    },
    "com.fasterxml.jackson.datatype:jackson-datatype-jdk8": {
      name: "com.fasterxml.jackson.datatype:jackson-datatype-jdk8",
      version: "2.22.0",
      cpe: "cpe:2.3:a:fasterxml:jackson-datatype-jdk8:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
    },
    "org.openapitools:jackson-databind-nullable": {
      name: "org.openapitools:jackson-databind-nullable",
      version: "0.2.6",
      cpe: "cpe:2.3:a:openapitools:jackson-databind-nullable:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
    },
    "commons-io:commons-io": {
      name: "commons-io:commons-io",
      version: "2.18.0",
      cpe: "cpe:2.3:a:apache:commons_io:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
    },
    "com.squareup.okhttp3:okhttp": {
      name: "com.squareup.okhttp3:okhttp",
      version: "4.12.0",
      cpe: "cpe:2.3:a:squareup:okhttp:*:*:*:*:*:*:*:*",
      ecosystem: "Maven",
      category: "runtime",
      condition: "useOkHttp",
    },
    "jakarta.annotation:jakarta.annotation-api": {
      name: "jakarta.annotation:jakarta.annotation-api",
      version: "3.0.0",
      cpe: "cpe:2.3:a:eclipse:jakarta_annotation_api:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
      condition: "languageVersion >= 11",
    },
    "jakarta.annotation:jakarta.annotation-api@java8": {
      name: "jakarta.annotation:jakarta.annotation-api",
      version: "2.1.1",
      cpe: "cpe:2.3:a:eclipse:jakarta_annotation_api:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
      condition: "languageVersion == 8",
    },
    "org.slf4j:slf4j-api": {
      name: "org.slf4j:slf4j-api",
      version: "2.0.9",
      cpe: "cpe:2.3:a:qos:slf4j:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
      condition: "slf4jLogging",
    },
    "com.jayway.jsonpath:json-path": {
      name: "com.jayway.jsonpath:json-path",
      version: "2.10.0",
      cpe: "cpe:2.3:a:jayway:jsonpath:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
      condition: "pagination",
    },
    "org.reactivestreams:reactive-streams": {
      name: "org.reactivestreams:reactive-streams",
      version: "1.0.4",
      cpe: "cpe:2.3:a:reactivestreams:reactive-streams:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "runtime",
      condition: "async && httpClient == jdk",
    },
    // Dev dependencies (testImplementation scope)
    "org.junit.jupiter:junit-jupiter": {
      name: "org.junit.jupiter:junit-jupiter",
      version: "5.11.4",
      cpe: "cpe:2.3:a:junit:junit:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "tests",
    },
    "org.mockito:mockito-core": {
      name: "org.mockito:mockito-core",
      version: "5.14.2",
      cpe: "cpe:2.3:a:mockito:mockito:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "tests",
    },
    "org.junit.platform:junit-platform-launcher": {
      name: "org.junit.platform:junit-platform-launcher",
      version: "1.11.4",
      cpe: "cpe:2.3:a:junit:junit-platform-launcher:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "tests",
    },
    "org.apache.ant:ant-junit": {
      name: "org.apache.ant:ant-junit",
      version: "1.9.7",
      cpe: "cpe:2.3:a:apache:ant:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "tests",
    },
    "org.slf4j:slf4j-simple": {
      name: "org.slf4j:slf4j-simple",
      version: "2.0.9",
      cpe: "cpe:2.3:a:qos:slf4j:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "slf4jLogging",
    },
    "io.projectreactor:reactor-core": {
      name: "io.projectreactor:reactor-core",
      version: "3.6.2",
      cpe: "cpe:2.3:a:vmware:reactor-core:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "internalTestSnippets",
    },
    "io.projectreactor:reactor-test": {
      name: "io.projectreactor:reactor-test",
      version: "3.6.2",
      cpe: "cpe:2.3:a:vmware:reactor-test:*:*:*:*:*:java:*:*",
      ecosystem: "Maven",
      category: "dev",
      condition: "internalTestSnippets",
    },
  } as const satisfies Record<string, TemplateDependency>;
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
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
 *              'java' section in gen.yaml).
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
  enableAsyncIfPossible(cfg);
  enablePopulatedFieldsUnionStrategyIfNeeded(cfg);
  resolveJava8Compat(cfg);
  pruneFlags(cfg);
}

// Java 8 has no `java.net.http` and no reactive Spring Boot 3.x.
// Forces OkHttp transport and disables the Spring Boot starter.
function resolveJava8Compat(cfg: Record<string, any>) {
  const languageVersion = Number(cfg.languageVersion);
  if (!Number.isFinite(languageVersion) || languageVersion >= 11) {
    return;
  }

  if (cfg.httpClient !== "okhttp") {
    console.warn(
      `httpClient: jdk is incompatible with languageVersion: ${languageVersion}. Overriding configuration with httpClient: okhttp.`,
    );
    cfg.httpClient = "okhttp";
  }

  if (cfg.generateSpringBootStarter === true) {
    console.warn(
      `languageVersion ${languageVersion} does not support the Spring Boot starter. Overriding configuration with generateSpringBootStarter: false.`,
    );
    cfg.generateSpringBootStarter = false;
  }
}

// Normalizes the forwardCompatibleUnionsByDefault boolean flag and preserves the
// truthy-string legacy behavior. String values are legacy artifacts of PR#3819 where
// "tagged-and-untagged" was the materialized form of true and "false" was the normalized
// default value. However a bug induced "false" to behave as ENABLED, meaning existing SDKs
// have shipped open unions since PR#3819. Normalizing "false" to true is therefore intentional
// here. From now on, only a hand-written boolean false genuinely disables open unions.
// @ts-ignore
function normalizeForwardCompatibleUnionsConfig(cfg: Record<string, any>) {
  const mode = cfg.forwardCompatibleUnionsByDefault;
  if (mode === "true" || mode === "tagged-and-untagged" || mode === "false") {
    cfg.forwardCompatibleUnionsByDefault = true;
  }
}

function pruneFlags(cfg: Record<string, any>) {
  if (cfg.enableAsync !== undefined) {
    delete cfg.enableAsync;
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
    templateVersion: {
      Name: "templateVersion",
      Required: false,
      DefaultValue: "v2",
      Description: "The template version to use",
      ValidationRegex: /v2/.source,
      ValidationMessage: "Template version must be v2",
    },
    projectName: {
      Name: "projectName",
      Required: true,
      DefaultValue: "openapi",
      Description:
        "Assigns Gradle rootProject.name, which gives a name to the Gradle build. https://docs.gradle.org/current/userguide/multi_project_builds.html#naming_recommendations",
      ValidationRegex: /^[\w\d.\-_]+$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
    },
    groupID: {
      Name: "groupID",
      Required: true,
      DefaultValue: "org.openapis",
      Description:
        "The groupID to use for namespacing the package. This is usually the reversed domain name of your organization. If publishing is enabled, it will also be used as the artifact's groupId (e.g. <groupId>.my-artifact).",
      ValidationRegex: /^\w[\w\d.\-_]*$/.source,
      ValidationMessage:
        "Letters, numbers, or .-_ only. Must start with a letter",
    },
    artifactID: {
      Name: "artifactID",
      Required: true,
      DefaultValue: "openapi",
      Description:
        "The artifactID is used as the final simple name in the base package if packageName is not specified. This is usually the name of your project. If publishing is enabled, it will also be used as the artifactId (e.g. com.yourorg.<artifactId>).",
      ValidationRegex: /^[\w\d\/._-]+$/.source,
      ValidationMessage:
        "Letters, numbers, or .-_/ only, must start with a letter",
    },
    packageName: {
      Name: "packageName",
      Required: false,
      DefaultValue: newSDK ? "org.openapis.openapi" : undefined,
      Description:
        "The base package name for generated classes. If not present <groupId>.<artifactId> will be used.",
      ValidationRegex: "^[a-z_][a-z_0-9]*(\\.[a-z_][a-z_0-9]*)*$",
      ValidationMessage: "Lowercase letters, numbers, or ._ only",
    },
    ossrhURL: {
      Name: "ossrhURL",
      Required: false,
      RequiredForPublishing: false,
      Description:
        "The URL of the staging repository to publish the SDK artifact to.",
    },
    githubURL: {
      Name: "githubURL",
      Required: false,
      RequiredForPublishing: true,
      DefaultValue: "github.com/owner/repo",
      Description:
        "The github URL where the artifact is hosted. Sets metadata required by Maven.",
      ValidationRegex: /github\.com\/[a-zA-z\d_-]+?\/.+/.source,
      ValidationMessage:
        "Must be a valid github.com URL, including an owner and repo (e.g. github.com/owner/repo)",
    },
    companyName: {
      Name: "companyName",
      Required: false,
      RequiredForPublishing: true,
      DefaultValue: "My Company",
      Description: "The name of your company. Sets metadata required by Maven.",
    },
    companyURL: {
      Name: "companyURL",
      Required: false,
      RequiredForPublishing: true,
      DefaultValue: "www.mycompany.com",
      Description:
        "Your company's homepage URL. Sets metadata required by Maven.",
    },
    companyEmail: {
      Name: "companyEmail",
      Required: false,
      RequiredForPublishing: true,
      DefaultValue: "info@mycompany.com",
      Description:
        "A support email address for your company. Sets metadata required by Maven.",
    },
    description: {
      Name: "description",
      Required: false,
      Description:
        "The description to use in the Maven POM file. If not provided, defaults to 'SDK enabling Java developers to easily integrate with the {CompanyName} API.'",
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
      Description: "Configuration for model import structure",
    },
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: [],
      Description:
        "Specify additional dependencies to include in build.gradle. Format is an array of 'scope:groupId:artifactId:version' strings. For example, 'implementation:com.fasterxml.jackson.core:jackson-databind:2.16.2'.",
    },
    additionalPlugins: {
      Name: "additionalPlugins",
      Required: false,
      DefaultValue: [],
      Description:
        'Specify additional plugins to include in build.gradle. Format is an array of strings. For example, \'id("org.jetbrains.kotlin.jvm") version "1.9.0"\'.',
    },
    license: {
      Name: "license",
      Required: false,
      DefaultValue: {
        name: "The MIT License (MIT)",
        url: "https://mit-license.org/",
        shortName: "MIT",
      },
      Description:
        "License information. Will default to MIT license if not otherwise specified.",
    },
    clientServerStatusCodesAsErrors: {
      Name: "clientServerStatusCodesAsErrors",
      Required: false,
      DefaultValue: true,
      Description: "Whether to treat 4xx and 5xx status codes as errors.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
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
    enableCustomCodeRegions: {
      Name: "enableCustomCodeRegions",
      Required: false,
      DefaultValue: false,
      Description: "Allow custom code to be inserted into the generated SDK.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    languageVersion: {
      Name: "languageVersion",
      Required: false,
      DefaultValue: 11,
      Description:
        "The Java version for the generated SDK (8, or 11-21). This is used to set the source and target compatibility in the Gradle build file. Version 8 requires `httpClient: okhttp` and no Spring Boot starter; its async surface is CompletableFuture-based.",
      ValidationRegex: /^(8|11|12|13|14|15|16|17|18|19|20|21)$/.source,
      ValidationMessage: "Must be 8, or between 11 and 21, inclusive.",
    },
    httpClient: {
      Name: "httpClient",
      Required: false,
      DefaultValue: "jdk",
      Description:
        "The HTTP client backend for the generated SDK. `jdk` uses `java.net.http` (Java 11+) and backs the reactive (Publisher-based) async surface. `okhttp` uses OkHttp, is required for `languageVersion: 8`, and backs a CompletableFuture-based async surface.",
      ValidationRegex: /^(jdk|okhttp)$/.source,
      ValidationMessage: "jdk or okhttp only",
    },
    nullFriendlyParameters: {
      Name: "nullFriendlyParameters",
      Required: false,
      DefaultValue: newSDK,
      Description: "Whether to use null-friendly parameters.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    getterStyle: {
      Name: "getterStyle",
      Required: false,
      DefaultValue: "presence-aware",
      Description:
        "Controls model getter return types. 'presence-aware' (default): required fields return T, optional fields return Optional<T>, nullable fields return JsonNullable<T>. 'raw': optional and nullable fields return @Nullable T (null when absent or explicit null). 'always-optional': every getter returns Optional<T>, including required fields.",
      ValidationRegex: /^(presence-aware|raw|always-optional)$/.source,
      ValidationMessage: "one of: presence-aware, raw, always-optional",
    },
    asyncMode: {
      Name: "asyncMode",
      Required: false,
      Description:
        "Whether to generate async API methods. With `httpClient: okhttp` these are CompletableFuture-based; with `httpClient: jdk` they are reactive (Publisher-based, non-blocking). Automatically enabled for new SDKs and existing SDKs without registered hooks (for compatibility).",
      ValidationRegex: /^(enabled|disabled)$/.source,
      ValidationMessage: "`enabled` or `disabled` only",
    },
    enableStreamingUploads: {
      Name: "enableStreamingUploads",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Whether to enable streaming uploads for sync SDKs, allowing users to pass various stream-oriented types (e.g. Blob) for bytes fields instead of byte arrays. This option is automatically enabled when `nullFriendlyParameters: true` is set.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    generateSpringBootStarter: {
      Name: "generateSpringBootStarter",
      Required: false,
      DefaultValue: true,
      Description: "Generate spring boot starters for the Java SDK.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    baseErrorName: {
      Name: "baseErrorName",
      Required: false,
      // The default is handled in `fixBuiltInErrorNameConflicts()`
      DefaultValue: "",
      Description:
        "The name of the base error class used for HTTP error responses",
      ValidationRegex: /^[a-zA-Z0-9]*$/.source,
      ValidationMessage:
        "Must start with a capital letter and contain only letters and numbers",
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
    forwardCompatibleUnionsByDefault: {
      Name: "forwardCompatibleUnionsByDefault",
      Required: false,
      DefaultValue: true,
      Description:
        "Enable forward compatibility for unions. When true, discriminated unions gain an Unknown* fallback type for unrecognized discriminator values, and untagged unions expose an asJson() accessor for payloads that match no known variant.",
      // The regex stays intentionally broader than the message to preserve support for the legacy "tagged-and-untagged" value (c.f. normalizeForwardCompatibleUnionsConfig).
      ValidationRegex: /^(true|false|tagged-and-untagged)$/.source,
      ValidationMessage: "true or false",
    },
    generateOptionalUnionAccessors: {
      Name: "generateOptionalUnionAccessors",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Generate optional accessor methods for non-discriminated union types instead of the generic value() method. When enabled, generates Option<SubType> subType() accessors for each union member.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    generateUnionDocs: {
      Name: "generateUnionDocs",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Generate Supported Types documentation for union model pages. When enabled, non-discriminated unions show factory method examples and discriminated unions show a discriminator-value-to-type mapping table.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    unionStrategy: {
      Name: "unionStrategy",
      Required: false,
      DefaultValue: "populated-fields",
      Description:
        "Strategy for deserializing union types. 'left-to-right' tries each type in order and returns the first valid match. 'populated-fields' uses sophisticated candidate scoring based on mapped fields, enum fields, and JSON size to select the best matching union member.",
      ValidationRegex: /^(left-to-right|populated-fields)$/.source,
      ValidationMessage: '"left-to-right" or "populated-fields" only',
    },
    enableSlf4jLogging: {
      Name: "enableSlf4jLogging",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Enable SLF4j logging integration in the Java SDK. When enabled, the SDK will use SLF4j for structured logging. Users can configure their preferred SLF4j implementation (Logback, Log4j2, etc.).",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    operationScopedParams: {
      Name: "operationScopedParams",
      Required: false,
      DefaultValue: true,
      Description:
        "When falling back to global parameters, only applies those declared at the operation level.",
    },
    forwardCompatibleEnumsByDefault: {
      Name: "forwardCompatibleEnumsByDefault",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Make enums forward compatible by default. When enabled, provides an API to access the raw value when the enum value fails to map to any of the known enum members. See `x-speakeasy-unknown-values: allow`",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    enableFormatting: {
      Name: "enableFormatting",
      Required: false,
      DefaultValue: false,
      Description:
        "Enable automatic code formatting for generated Java files using dprint.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    showSetterGetterTypesInDocs: {
      Name: "showSetterGetterTypesInDocs",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Show separate Setter Type and Getter Type columns in model documentation instead of a single Type column. Only takes effect when nullFriendlyParameters is also enabled.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    explicitDocImports: {
      Name: "explicitDocImports",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Use explicit imports instead of wildcard imports in doc usage examples to avoid ambiguous class references.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    prefixModeMethodNames: {
      Name: "prefixModeMethodNames",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Use 'to' prefix for sync/async mode-switching methods (toSync/toAsync) instead of bare names. Follows RxJava/Mutiny conventions and avoids collisions with SubSDK accessor names.",
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
  _genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  const languageVersion = Number(_genConfig.LangCfg?.languageVersion);
  const result: TargetInitialConfiguration = {
    compile: getCompileConfiguration(languageVersion),
    testing: getTestingConfiguration(languageVersion),
  };

  return result;
}

// @ts-ignore
function getJavaCommandDependencies(
  languageVersion: number,
): RunnerCommandDependencies {
  // Java 8 reports `javac 1.8.0_xxx`; anything >= 11 reports `javac 21.0.2`.
  // The floor must match the requested languageVersion or javac 8 is rejected.
  const javacMinVersion = languageVersion === 8 ? "1.8.0" : "11.0.0";
  return [
    {
      command: "javac",
      version: {
        args: ["-version"],
        regex: `(?m)^\\w+ (\\d+\\.*\\d*\\.*\\d*).*?`,
        minVersion: javacMinVersion,
      },
      installDocumentation: `Install Java by following the instructions at https://openjdk.org/install`,
    },
    {
      command: "gradle",
      version: {
        args: ["--version"],
        regex: `(?m)Gradle.*?(\\d+\\.\\d+).*?`,
        minVersion: "8.0",
      },
      installDocumentation: `Install Gradle by following the instructions at https://gradle.org/install.`,
    },
  ];
}

// @ts-ignore
function getCompileConfiguration(
  languageVersion: number,
): CompileConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "gradle",
          args: ["build", "-x", "test"],
        },
        {
          command: "gradle",
          args: ["javadoc"],
        },
      ],
      dependencies: {
        commands: getJavaCommandDependencies(languageVersion),
      },
    },
  };
}

// @ts-ignore
function getTestingConfiguration(
  languageVersion: number,
): TestingConfiguration {
  return {
    runner: {
      commands: [
        {
          command: "gradle",
          args: ["test", "publishToMavenLocal", "-Pskip.signing=true"],
        },
      ],
      dependencies: {
        commands: getJavaCommandDependencies(languageVersion),
      },
    },
  };
}

/**
 * Detects if hooks are registered in the existing SDKHooks.java file
 * @returns true if hooks are found, false otherwise
 */
// @ts-ignore
function hasRegisteredHooks(cfg: Record<string, any>): boolean {
  try {
    // Infer the path to SDKHooks.java from the configuration
    const sdkHooksPath = getHooksFilePath(cfg);

    if (!sdkHooksPath) {
      return false; // Cannot determine SDKHooks.java path
    }

    const content = readFile(sdkHooksPath);
    if (!content) {
      return false; // File doesn't exist or is empty
    }

    // Check if the content contains hook registrations
    return checkForHookRegistrations(content);
  } catch (error) {
    // Handle any errors gracefully
    return false;
  }
}

function enableAsyncIfPossible(cfg: Record<string, any>) {
  // if flag is already set - bail out
  if (cfg.asyncMode !== undefined) {
    return;
  }

  if (!hasRegisteredHooks(cfg)) {
    cfg.asyncMode = "enabled";
  }
}

function enablePopulatedFieldsUnionStrategyIfNeeded(cfg: Record<string, any>) {
  // If forwardCompatibleEnumsByDefault is enabled, always override unionStrategy to "populated-fields"
  if (cfg.forwardCompatibleEnumsByDefault === true) {
    cfg.unionStrategy = "populated-fields";
  }
}

/**
 * Infers the path to SDKHooks.java based on the configuration
 */
// @ts-ignore
function getHooksFilePath(cfg: Record<string, any>): string | null {
  try {
    // Get package name using the same logic as templatePackageName
    let packageName: string;
    if (cfg.packageName) {
      packageName = cfg.packageName;
    } else {
      packageName = `${cfg.groupID.replaceAll(
        "-",
        "_",
      )}.${cfg.artifactID.replaceAll("-", "_")}`;
    }

    // Convert package name to directory path
    const packagePath = packageName.replace(/\./g, "/");

    // Construct the expected path to SDKHooks.java
    // Typical structure: src/main/java/{packagePath}/SDKHooks.java
    const sdkHooksPath = `src/main/java/${packagePath}/hooks/SDKHooks.java`;

    return sdkHooksPath;
  } catch (error) {
    return null;
  }
}

/**
 * Checks if the given content contains hook registrations
 */
// @ts-ignore
function checkForHookRegistrations(content: string): boolean {
  // Look for hook registration patterns
  const hookPatterns = [
    /hooks\.registerBeforeRequest\(/,
    /hooks\.registerAfterSuccess\(/,
    /hooks\.registerAfterError\(/,
  ];

  // Split content into lines and check each line individually
  const lines = content.split("\n");

  for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
    const line = lines[lineIndex].trim();

    // Skip empty lines and commented lines (both // and /* */ style comments)
    if (
      !line ||
      line.startsWith("//") ||
      line.startsWith("/*") ||
      line.startsWith("*")
    ) {
      continue;
    }

    // Check if any hook registration pattern is found in this non-commented line
    for (const pattern of hookPatterns) {
      if (pattern.test(line)) {
        return true;
      }
    }
  }

  return false;
}

// @ts-ignore
function validateConfig(cfg: Record<string, any>): string {
  return "";
}
