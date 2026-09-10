# Generator-Target Interface

The high level "plugin" architecture of the generator is that each target is represented by a directory of TypeScript files that contain all the target-specific configuration and templating logic.

There are a few main entrypoints that represent the interface between the generator and the target:

- `config.ts`: Target configuration functions, including user-facing SDK configuration and generator configuration.
- `examples.ts`: Target example pre-calculation logic.
- `main.ts`: The main logic for generating target code, including models, operations, utilities, and testing.
- `standalone.ts`: Logic for generating target code that is interopt templated from another target, such as the `terraform` target to generate `go` target code internally, or usage snippets.

There are also a few legacy entrypoints and magic files that will be migrated to `config.ts`:

- `exclusions.ts`: Filetracking configuration.
- `features.ts`: Features, generated code headers, and README configuration.
- `HIDDEN`: Present if the target should not be listed as a generation target. For example, typescriptv2 is hidden in preference of showing typescript.
- `SUNSET`: Present if the target has been removed and should not be directly used.

## Configuration

Target configuration must be handled in a `config.ts` file. User-facing SDK configuration values are loaded early in the generator logic, then the generator configuration is loaded with those values available.

### User-facing SDK Configuration

User-facing SDK configuration is made available in the target-specific `gen.yaml` configuration. For example:

```yaml
configVersion: 2.0.0
generation:
  ...
TARGET:
  ... CUSTOMER CONFIG ...
```

This configuration must be handled via a `getConfigFields()` function in `config.ts` with the following signature:

```typescript
function getConfigFields(
  commonFields: SDKGenConfigFields,
  newSDK: boolean,
): SDKGenConfigFields
```

Its implementation typically involves including common fields and the generator currently requires an `imports` field. As a minimal example:

```typescript
return {
  ...commonFields,
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
}
```

The `SDKGenConfigFields` type in TypeScript has all the available configuration options. The generator `internal/configuration` package contains the Go functionality for handling this data.

### Generator Configuration

This configuration is between the target and the generator itself. After loading user-facing SDK configuration, the generator loads target configuration. This includes information such as target lifecycle (sunset), visibility (hidden), compile commands, feature implementations, and more.

This configuration must be handled via a `getGeneratorConfig()` function in `config.ts` with the following signature:

```typescript
function getGeneratorConfig(
  genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration
```

The `TargetInitialConfiguration` type in TypeScript has all the available configuration options. The generator `internal/targetconfig` package contains the Go functionality for handling this data.

## Compilation

Enable code compilation for the target by implementing the optional `compile` property of `TargetInitialConfiguration`.

In this example, the target inline defines compilation commands and dependencies:

```typescript
function getGeneratorConfig(
  genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  let result: TargetInitialConfiguration = {
    // ... other configuration ...
    compile: {
      runner: {
        commands: [
          {
            command: "go",
            args: ["build", "./..."],
          },
        ],
        dependencies: {
          commands: [
            {
              command: "go",
              version: {
                args: ["version"],
                regex: `(?m).*?go version go(\\d+\\.\\d+\\.\\d+).*?`,
                minVersion: "1.20.0",
              },
              installDocumentation: `Install Go by following the instructions at https://golang.org/doc/install.`,
            },
          ],
        },
      },
    },
  };

  return result;
}
```

Some targets with additional logic may opt to construct this information dynamically via functions instead.
