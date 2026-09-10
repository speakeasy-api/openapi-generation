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
    // The AST will not set (ast.SubResponse).Error for any 4XX/5XX status codes
    // without this enabled, which causes (ast.Operation).GetAcceptTypes() to
    // return too much data and therefore break Accept header assertions.
    clientServerStatusCodesAsErrors: {
      Name: "clientServerStatusCodesAsErrors",
      Required: false,
      DefaultValue: true,
      Description: "Whether to treat 4xx and 5xx status codes as errors.",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
    // pkg/generate requires imports for AST TypeDef.OutputLocation resolution.
    imports: {
      Name: "imports",
      Required: false,
      DefaultValue: {
        option: "openapi",
        paths: {
          callbacks: "models/callbacks",
          errors: "models/sdkerrors",
          operations: "models/operations",
          shared: "models/components",
          webhooks: "models/webhooks",
        },
      },
      Description: "Configuration for model import structure",
    },
    // Make the AST match the go target default, otherwise defaults to envelope.
    responseFormat: {
      Name: "responseFormat",
      Required: false,
      DefaultValue: "envelope-http",
      Description:
        "Determines the shape of the response envelope that is returned from SDK methods",
      ValidationRegex: /^(envelope|envelope-http|flat)$/.source,
      ValidationMessage: '"envelope-http", "envelope" or "flat" only',
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
  return {
    runner: {
      commands: [
        {
          command: "go",
          args: ["mod", "tidy"],
          environmentVariables: {
            GOROOT: "",
          },
        },
        {
          command: "go",
          args: ["build", "./..."],
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
