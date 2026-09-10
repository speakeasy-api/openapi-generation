/** Returns the structured dependency list for this template.
 * This is the single source of truth for all dependencies.
 * Used by both template rendering and security scanning. */
// @ts-ignore
function getTemplateDependencies(): Record<string, TemplateDependency> {
  return {
    // Runtime dependencies
    "github.com/spyzhov/ajson": {
      name: "github.com/spyzhov/ajson",
      version: "v0.9.6",
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
      version: "v1.11.1",
      cpe: "cpe:2.3:a:stretchr:testify:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "dev",
      condition: "tests",
    },
    "golang.org/x/sync": {
      name: "golang.org/x/sync",
      version: "v0.19.0",
      cpe: "cpe:2.3:a:golang:sync:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "oauth2",
    },
    "github.com/itchyny/gojq": {
      name: "github.com/itchyny/gojq",
      version: "v0.12.18",
      cpe: "cpe:2.3:a:itchyny:gojq:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "transformJq",
    },
    "github.com/zalando/go-keyring": {
      name: "github.com/zalando/go-keyring",
      version: "v0.2.6",
      cpe: "cpe:2.3:a:zalando:go-keyring:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "security",
    },
    "github.com/spf13/cobra": {
      name: "github.com/spf13/cobra",
      version: "v1.10.2",
      cpe: "cpe:2.3:a:spf13:cobra:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "gopkg.in/yaml.v3": {
      name: "gopkg.in/yaml.v3",
      version: "v3.0.1",
      cpe: "cpe:2.3:a:yaml:yaml.v3:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/alpkeskin/gotoon": {
      name: "github.com/alpkeskin/gotoon",
      version: "v0.1.1",
      cpe: "cpe:2.3:a:alpkeskin:gotoon:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "golang.org/x/term": {
      name: "golang.org/x/term",
      version: "v0.40.0",
      cpe: "cpe:2.3:a:golang:x_term:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
    },
    "github.com/charmbracelet/huh": {
      name: "github.com/charmbracelet/huh",
      version: "v0.6.0",
      cpe: "cpe:2.3:a:charmbracelet:huh:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "interactive",
    },
    "github.com/charmbracelet/lipgloss": {
      name: "github.com/charmbracelet/lipgloss",
      version: "v1.1.0",
      cpe: "cpe:2.3:a:charmbracelet:lipgloss:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "interactive",
    },
    "github.com/charmbracelet/bubbletea": {
      name: "github.com/charmbracelet/bubbletea",
      version: "v1.3.4",
      cpe: "cpe:2.3:a:charmbracelet:bubbletea:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "interactiveMode",
    },
    "github.com/charmbracelet/bubbles": {
      name: "github.com/charmbracelet/bubbles",
      version: "v0.20.0",
      cpe: "cpe:2.3:a:charmbracelet:bubbles:*:*:*:*:*:go:*:*",
      ecosystem: "Go",
      category: "runtime",
      condition: "interactiveMode",
    },
  };
}

// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  return cfg;
}

// @ts-ignore
function resolveConfig(cfg: Record<string, any>) {
  normalizeForwardCompatibleUnionsConfig(cfg);
  enablePopulatedFieldsUnionStrategyIfNeeded(cfg);
}

function getDistributionValidationConfig(cfg: Record<string, any>) {
  return cfg.distribution && typeof cfg.distribution === "object"
    ? cfg.distribution
    : {};
}

function getHomebrewValidationConfig(cfg: Record<string, any>) {
  const distribution = getDistributionValidationConfig(cfg);
  return distribution.homebrew && typeof distribution.homebrew === "object"
    ? distribution.homebrew
    : {};
}

function getWingetValidationConfig(cfg: Record<string, any>) {
  const distribution = getDistributionValidationConfig(cfg);
  return distribution.winget && typeof distribution.winget === "object"
    ? distribution.winget
    : {};
}

function getNFPMValidationConfig(cfg: Record<string, any>) {
  const distribution = getDistributionValidationConfig(cfg);
  return distribution.nfpm && typeof distribution.nfpm === "object"
    ? distribution.nfpm
    : {};
}

// @ts-ignore
function validateConfig(cfg: Record<string, any>): string {
  const generateRelease = cfg.generateRelease !== false;
  const repoURL = String(cfg.repoURL || cfg.repoUrl || "").trim();
  const packageName = String(cfg.packageName || "").trim();
  const hasGitHubRepo =
    /(?:^|\/\/)github\.com\/[^/]+\/[^/]+/.test(repoURL) ||
    /^github\.com\/[^/]+\/[^/]+/.test(packageName);

  const homebrew = getHomebrewValidationConfig(cfg);
  if (homebrew.enabled === true) {
    if (!generateRelease) {
      return "distribution.homebrew.enabled requires generateRelease to be true";
    }

    if (typeof homebrew.tap !== "string" || homebrew.tap.trim() === "") {
      return "distribution.homebrew.tap is required when distribution.homebrew.enabled is true";
    }

    if (!/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/.test(homebrew.tap.trim())) {
      return "distribution.homebrew.tap must use the format owner/repo";
    }

    if (
      !/^[A-Za-z0-9_.-]+\/homebrew-[A-Za-z0-9_.-]+$/.test(homebrew.tap.trim())
    ) {
      return "distribution.homebrew.tap repository name must start with 'homebrew-'";
    }

    if (!hasGitHubRepo) {
      return "distribution.homebrew.enabled requires repoURL or packageName to resolve to a GitHub repository";
    }
  }

  const winget = getWingetValidationConfig(cfg);
  if (winget.enabled === true) {
    if (!generateRelease) {
      return "distribution.winget.enabled requires generateRelease to be true";
    }

    if (
      typeof winget.publisher !== "string" ||
      winget.publisher.trim() === ""
    ) {
      return "distribution.winget.publisher is required when distribution.winget.enabled is true";
    }

    if (
      typeof winget.repositoryOwner !== "string" ||
      winget.repositoryOwner.trim() === ""
    ) {
      return "distribution.winget.repositoryOwner is required when distribution.winget.enabled is true";
    }

    if (!/^[A-Za-z0-9_.-]+$/.test(winget.repositoryOwner.trim())) {
      return "distribution.winget.repositoryOwner must be a valid GitHub owner name";
    }

    if (
      typeof winget.publisherUrl !== "string" ||
      winget.publisherUrl.trim() === ""
    ) {
      return "distribution.winget.publisherUrl is required when distribution.winget.enabled is true";
    }

    if (!/^https?:\/\//.test(winget.publisherUrl.trim())) {
      return "distribution.winget.publisherUrl must be an absolute http or https URL";
    }

    if (
      typeof winget.packageIdentifier !== "string" ||
      winget.packageIdentifier.trim() === ""
    ) {
      return "distribution.winget.packageIdentifier is required when distribution.winget.enabled is true";
    }

    if (
      !/^[A-Za-z0-9][A-Za-z0-9-]*(\.[A-Za-z0-9][A-Za-z0-9-]*)+$/.test(
        winget.packageIdentifier.trim(),
      )
    ) {
      return "distribution.winget.packageIdentifier must use dot-separated Publisher.Package segments";
    }

    if (typeof winget.license !== "string" || winget.license.trim() === "") {
      return "distribution.winget.license is required when distribution.winget.enabled is true";
    }

    if (!hasGitHubRepo) {
      return "distribution.winget.enabled requires repoURL or packageName to resolve to a GitHub repository";
    }
  }

  const nfpm = getNFPMValidationConfig(cfg);
  if (nfpm.enabled === true) {
    if (!generateRelease) {
      return "distribution.nfpm.enabled requires generateRelease to be true";
    }

    if (typeof nfpm.maintainer !== "string" || nfpm.maintainer.trim() === "") {
      return "distribution.nfpm.maintainer is required when distribution.nfpm.enabled is true";
    }

    if (typeof nfpm.license !== "string" || nfpm.license.trim() === "") {
      return "distribution.nfpm.license is required when distribution.nfpm.enabled is true";
    }

    if (typeof nfpm.formats !== "string" || nfpm.formats.trim() === "") {
      return "distribution.nfpm.formats is required when distribution.nfpm.enabled is true";
    }

    const allowedFormats = new Set(["deb", "rpm", "apk"]);
    const formats = nfpm.formats
      .split(",")
      .map((format) => format.trim())
      .filter((format) => format !== "");

    if (formats.length === 0) {
      return "distribution.nfpm.formats must include at least one of: deb, rpm, apk";
    }

    for (const format of formats) {
      if (!allowedFormats.has(format)) {
        return `distribution.nfpm.formats contains unsupported format '${format}'. Allowed values: deb, rpm, apk`;
      }
    }
  }

  return "";
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
  if (cfg.forwardCompatibleEnumsByDefault === true) {
    cfg.unionStrategy = "populated-fields";
  }
}

// @ts-ignore
function getConfigFields(
  commonFields: SDKGenConfigFields,
  newSDK: boolean,
): SDKGenConfigFields {
  return {
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "openapi",
      Description:
        "The go module package name. https://go.dev/ref/mod#module-path.",
      ValidationRegex: /^[\w\d.\-_/]+$/.source,
      ValidationMessage: "Letters, numbers, or /.-_ only",
    },
    cliName: {
      Name: "cliName",
      Required: false,
      DefaultValue: "cli",
      Description:
        "The name of the CLI binary and config directory (e.g., 'myapp' for ~/.config/myapp/config.yaml)",
      ValidationRegex: /^[a-z][a-z0-9\-]*$/.source,
      ValidationMessage:
        "Lowercase letters, numbers, and hyphens only. Must start with a letter.",
    },
    envVarPrefix: {
      Name: "envVarPrefix",
      Required: false,
      DefaultValue: "CLI",
      Description:
        "Prefix for environment variables (e.g., 'MYAPP' for MYAPP_API_KEY)",
      ValidationRegex: /^[A-Z][A-Z0-9_]*$/.source,
      ValidationMessage:
        "Uppercase letters, numbers, and underscores only. Must start with a letter.",
    },
    additionalDependencies: {
      Name: "additionalDependencies",
      Required: false,
      DefaultValue: {},
      Description:
        "Specify additional dependencies to include in the generated go.mod",
    },
    enableCustomCodeRegions: {
      Name: "enableCustomCodeRegions",
      Required: false,
      DefaultValue: false,
      Description: "Allow custom code to be inserted into the generated CLI.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    removeStutter: {
      Name: "removeStutter",
      Required: false,
      DefaultValue: true,
      Description:
        "Remove command name stutter. Exact matches are promoted to the parent group; prefix matches have the group prefix stripped.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    generateRelease: {
      Name: "generateRelease",
      Required: false,
      DefaultValue: true,
      Description:
        "Generate goreleaser configuration, GitHub Actions release workflow, and install scripts for binary distribution.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    distribution: {
      Name: "distribution",
      Required: false,
      DefaultValue: {
        homebrew: {
          enabled: false,
          tap: "",
        },
        winget: {
          enabled: false,
          publisher: "",
          repositoryOwner: "",
          publisherUrl: "",
          packageIdentifier: "",
          license: "",
        },
        nfpm: {
          enabled: false,
          formats: "deb,rpm",
          maintainer: "",
          license: "",
        },
      },
      Description:
        "Binary distribution configuration for Homebrew, WinGet, and nFPM release channels.",
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
    idiomaticMethodCollisionNames: {
      Name: "idiomaticMethodCollisionNames",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Resolve field names that collide with generated struct methods (e.g. Error()) using idiomatic replacements (e.g. ErrorInfo) instead of a trailing underscore (Error_)",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    interactiveAuth: {
      Name: "interactiveAuth",
      Required: false,
      DefaultValue: true,
      Description:
        "Generate interactive auth login flows using bubbletea/huh forms.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    interactiveMode: {
      Name: "interactiveMode",
      Required: false,
      DefaultValue: true,
      Description:
        "Generate interactive command explorer and prompting support for missing required inputs.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    interactiveByDefault: {
      Name: "interactiveByDefault",
      Required: false,
      DefaultValue: true,
      Description:
        "Enable implicit required-input prompts and configure/auth forms by default when interactive support is generated. Set false for non-interactive-by-default CLIs; users can opt in with --interactive, while explore remains an explicit workflow.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    agentEnvironmentDetection: {
      Name: "agentEnvironmentDetection",
      Required: false,
      DefaultValue: true,
      Description:
        "Automatically enable agent mode from a built-in allowlist of known agent environment variables. Set false to omit the allowlist and require explicit --agent-mode; recommended for non-interactive-by-default agent images.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    classifiedErrors: {
      Name: "classifiedErrors",
      Required: false,
      DefaultValue: false,
      Description:
        "Add reason-first error_type, error_reason, message, and actionable hints whenever JSON, TOON, or --jq machine output is explicitly requested. Agent mode always enables classified errors regardless of this setting.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    defaultColor: {
      Name: "defaultColor",
      Required: false,
      DefaultValue: "auto",
      Description:
        "Default value of the --color flag: auto (color when stdout is a TTY), always, or never. never makes uncolored output the default so piped/redirected output is byte-clean; --color and NO_COLOR/FORCE_COLOR still apply.",
      ValidationRegex: /^(auto|always|never)$/.source,
      ValidationMessage: "auto, always, or never only",
    },
    defaultTimeout: {
      Name: "defaultTimeout",
      Required: false,
      DefaultValue: "",
      Description:
        "Default network timeout applied to each API operation when --timeout is not given (a Go duration such as 60s). Bounds one HTTP operation including retries; operations whose body outlives the call (event streams, JSONL, raw response streams) are exempt so long-lived reads survive, and polling loops re-arm it per request without capping their own total budget. Empty or zero disables the default.",
      ValidationRegex: /^$|^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$/.source,
      ValidationMessage: "a Go duration such as 60s or 2m, or empty",
    },
    jqRawOutput: {
      Name: "jqRawOutput",
      Required: false,
      DefaultValue: false,
      Description:
        "Default value of the --raw-output flag: when true, --jq results that are strings are written as raw text (no JSON quoting or escaping, like jq -r) so values such as base64 payloads can be piped straight into other tools; non-string results stay JSON.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    helpStyle: {
      Name: "helpStyle",
      Required: false,
      DefaultValue: "auto",
      Description:
        "CLI help layout: auto uses compact help when x-speakeasy-cli-commands declares a curated command manifest and full help otherwise; compact moves inherited global flags behind the root --help-global flag; full preserves the traditional Cobra help output.",
      ValidationRegex: /^(auto|compact|full)$/.source,
      ValidationMessage: "auto, compact, or full only",
    },
    retryFlagsVisibility: {
      Name: "retryFlagsVisibility",
      Required: false,
      DefaultValue: "visible",
      Description:
        "Visibility of the retry control flags (--no-retries, --retry-config, --retry-connection-errors, --retry-max-elapsed-time): visible shows them in help, hidden registers them without help output, none skips registration entirely. Config-file and env retry settings keep working in every mode that registers the flags.",
      ValidationRegex: /^(visible|hidden|none)$/.source,
      ValidationMessage: "visible, hidden, or none only",
    },
    retryMethodPolicy: {
      Name: "retryMethodPolicy",
      Required: false,
      DefaultValue: "spec",
      Description:
        "Controls which HTTP methods inherit a root retry policy. spec follows the OpenAPI retry extension exactly. safe-methods limits inherited retries to GET, HEAD, OPTIONS, and TRACE; explicit operation retry policies still win.",
      ValidationRegex: /^(spec|safe-methods)$/.source,
      ValidationMessage: "spec or safe-methods only",
    },
    serverSelectionFlag: {
      Name: "serverSelectionFlag",
      Required: false,
      DefaultValue: "visible",
      Description:
        "Visibility of the --server selection flag: visible (the default) always shows it, auto shows it when the document offers multiple global or operation-level servers and hides it otherwise (still registered, so existing --server invocations keep parsing), hidden registers it without help output, and none skips registration entirely. The --server-url override and server variable flags are unaffected.",
      ValidationRegex: /^(auto|visible|hidden|none)$/.source,
      ValidationMessage: "auto, visible, hidden, or none only",
    },
    interactiveTheme: {
      Name: "interactiveTheme",
      Required: false,
      DefaultValue: {
        accentColor: "#38BDF8",
        dimmedColor: "#64748B",
        subtleColor: "#475569",
        errorColor: "#F87171",
        successColor: "#4ADE80",
      },
      Description:
        "Color theme for interactive prompts. All values are hex color strings. accentColor: primary accent (borders, titles, selectors). dimmedColor: secondary text (descriptions, blurred fields). subtleColor: muted elements (placeholders, inactive borders). errorColor: validation errors. successColor: checkmarks and confirmations.",
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

  commands.push({
    command: "go",
    args: ["build", "./..."],
    environmentVariables: {
      GOROOT: "",
    },
  });

  // Generate Cobra markdown documentation into docs/
  commands.push({
    command: "go",
    args: ["run", "./cmd/gendocs"],
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
            "-checks=inherit,-SA1019,-SA5008,-S1040,-U1000",
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

// @ts-ignore
function getConfigOverlay(): Record<string, any> {
  // This needs to be kept in sync with go's default imports config
  return {
    inputModelSuffix: "input",
    outputModelSuffix: "output",
    flattenGlobalSecurity: false,
    maxMethodParams: 0,
    methodArguments: "infer-optional-args",
    nullableOptionalWrapper: true,
    responseFormat: "envelope-http",
    enableSkipDeserialization: true,
    clientServerStatusCodesAsErrors: true,
    defaultErrorName: "SDKDefaultError",
    baseErrorName: "SDKError",
    forwardCompatibleEnumsByDefault: true,
    forwardCompatibleUnionsByDefault: "tagged-and-untagged",
    unionStrategy: "populated-fields",
    imports: {
      option: "openapi",
      paths: {
        shared: "models/components",
        operations: "models/operations",
        errors: "models/sdkerrors",
        callbacks: "models/callbacks",
        webhooks: "models/webhooks",
      },
    },
  };
}
