# CLI Generator

Generates a fully functional Go CLI from an OpenAPI specification. The generated CLI uses [Cobra](https://github.com/spf13/cobra) for command structure and a generated Go SDK for API calls. Users interact with the API through typed flags, stdin JSON piping, interactive prompts and exploration, machine-readable usage schemas, and multiple output formats.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Quickstart for Contributors](#quickstart-for-contributors)
  - [Definition of Done](#definition-of-done)
- [Architecture Overview](#architecture-overview)
- [Directory Structure](#directory-structure)
- [Generation Flow](#generation-flow)
  - [1. Job Creation (`main.ts`)](#1-job-creation-maints)
  - [2. Command Generation](#2-command-generation)
  - [3. Request Building Paths](#3-request-building-paths)
- [Key Components](#key-components)
  - [Command Generation](#command-generation)
  - [Metadata-Driven Request Building](#metadata-driven-request-building)
  - [Flag Metadata Generation](#flag-metadata-generation)
  - [Union Type Handling](#union-type-handling)
  - [Declarative Intent Commands and Route Dispatch](#declarative-intent-commands-and-route-dispatch)
  - [Output Formatting](#output-formatting)
  - [Interactive Mode](#interactive-mode)
  - [Agent Mode](#agent-mode)
  - [Exit Codes and Error Boundaries](#exit-codes-and-error-boundaries)
  - [Pagination](#pagination)
  - [Streaming (SSE \& JSONL)](#streaming-sse--jsonl)
    - [Streamed-event projection (`x-speakeasy-cli-commands` `output.stream.select`)](#streamed-event-projection-x-speakeasy-cli-commands-outputstreamselect)
  - [Binary Downloads](#binary-downloads)
  - [Bytes / Base64 Request Input](#bytes--base64-request-input)
  - [Retries \& Timeout](#retries--timeout)
  - [Response Headers](#response-headers)
  - [Configuration \& Auth](#configuration--auth)
  - [Usage Schema \& Grouped Help](#usage-schema--grouped-help)
  - [Generated README \& Command Docs](#generated-readme--command-docs)
  - [Server Selection](#server-selection)
  - [Custom HTTP Headers](#custom-http-headers)
  - [JQ Transforms (`transformJq`)](#jq-transforms-transformjq)
  - [Diagnostics](#diagnostics)
  - [Test Framework](#test-framework)
- [Where to Change What](#where-to-change-what)
- [Relationship with Go SDK](#relationship-with-go-sdk)
- [Release Distribution Configuration](#release-distribution-configuration)
  - [Distribution schema](#distribution-schema)
  - [Validation rules](#validation-rules)
  - [Rendering behavior](#rendering-behavior)
- [Forward Compatibility](#forward-compatibility)
- [Body Input Methods](#body-input-methods)
- [Required Validation Semantics](#required-validation-semantics)
  - [Argument \& Value Hygiene](#argument--value-hygiene)
- [Known Limitations](#known-limitations)
  - [Skipped Tests](#skipped-tests)
    - [Skip Summary](#skip-summary)
    - [Review-Only Skips (3)](#review-only-skips-3)
    - [Previously Skipped (Now Resolved)](#previously-skipped-now-resolved)
    - [How Skip Detection Works](#how-skip-detection-works)
  - [Unsupported / Deferred Features](#unsupported--deferred-features)
  - [Known Bugs](#known-bugs)
- [Evolution \& Design Context](#evolution--design-context)
- [Design Decisions](#design-decisions)
  - [Why metadata-driven instead of imperative per-field parsing?](#why-metadata-driven-instead-of-imperative-per-field-parsing)
  - [Why reflection instead of raw body bypass?](#why-reflection-instead-of-raw-body-bypass)
  - [Why `envelope-http` response format?](#why-envelope-http-response-format)
  - [Why `ValidateMeta` at startup?](#why-validatemeta-at-startup)
  - [Why in-process test execution?](#why-in-process-test-execution)
  - [Why have both interactive mode and agent mode?](#why-have-both-interactive-mode-and-agent-mode)
- [Common Patterns \& Pitfalls](#common-patterns--pitfalls)
  - [`IsRequestBody` vs `RequestBody`](#isrequestbody-vs-requestbody)
  - [`sanitizeType()` has side effects](#sanitizetype-has-side-effects)
  - [Go enum types aren't plain strings in TypeScript](#go-enum-types-arent-plain-strings-in-typescript)
  - [`Example.ToJSON()` returns a JSON string](#exampletojson-returns-a-json-string)
  - [Variant `FieldPath` is relative](#variant-fieldpath-is-relative)
  - [`sanitizeCommandName` already appends `Cmd`](#sanitizecommandname-already-appends-cmd)
  - [Config overlay must use accessor](#config-overlay-must-use-accessor)
- [Adding Features](#adding-features)
  - [Adding a new flag type](#adding-a-new-flag-type)
  - [Adding a new output format](#adding-a-new-output-format)
  - [Adding a new global flag](#adding-a-new-global-flag)
  - [Running tests](#running-tests)
  - [Debugging generation](#debugging-generation)
  - [Keeping this README current](#keeping-this-readme-current)

---

## Quickstart for Contributors

**All edits go in `templates/templates/cli/`** — never edit the generated output directly.

```bash
# Generate + compile + test against the review spec
TARGET=review make test-cli
# Output lands in: zSDKs/sdk-cli/

# Generate + compile + test against the primary spec
TARGET=primary make test-cli
# Output lands in: testSDKs/sdk-cli-primary/

# Generate only (for quick inspection without tests)
go run cmd/generate/main.go -s /tmp/spec.yaml -o /tmp/cli-output -l cli
```

**Typical workflow**:

1. Edit template source in `templates/templates/cli/` (TypeScript in `includes/`, Go templates in `*.stmpl`, runtime in `auxiliary/`)
2. Run `TARGET=review make test-cli` to generate, compile, and test
3. Inspect generated Go files in `zSDKs/sdk-cli/` to verify the output looks correct
4. Run both test targets before submitting a PR

### Definition of Done

Before submitting a PR that touches CLI templates:

- [ ] `TARGET=review make test-cli` passes
- [ ] `TARGET=primary make test-cli` passes
- [ ] Inspected generated files in `zSDKs/sdk-cli/` for correctness (metadata arrays, flag names, request building calls)
- [ ] `make lint` passes
- [ ] Generated command docs are valid (the build step runs `go run ./cmd/gendocs` which produces Cobra markdown docs under `docs/`)
- [ ] If the change affects generated SDK output, ran `go run cmd/changelog/main.go`

---

## Architecture Overview

The generator follows a **generated commands + shared runtime** pattern:

```text
┌──────────────────────────────────────────────────────────────┐
│  Generation Time (TypeScript)                                │
│                                                              │
│  main.ts ──► metadata.ts ──► opcmd.go.stmpl ──► user.go    │
│              templating.ts      root.go.stmpl     root.go    │
│              unions.ts          test.go.stmpl     *_test.go  │
│              tests.ts           configure.go.stmpl            │
│              security.ts        whoami.go.stmpl               │
└──────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────────┐
│  Runtime (Generated Go)                                      │
│                                                              │
│  cmd/                                                        │
│    root.go          ─ Root Cobra command + global flags       │
│    users.go         ─ Subcommand group                       │
│    createuser.go    ─ Operation command (metadata + run)     │
│                                                              │
│  internal/                                                   │
│    flagutil/                                                 │
│      metadata.go    ─ RegisterFlags + BuildRequest (reflect) │
│      flags.go       ─ Low-level flag utilities               │
│    client/                                                   │
│      client.go      ─ SDK client construction                │
│      diagnostics.go ─ --dry-run/--debug wrappers + redaction │
│    output/                                                   │
│      output.go      ─ Result/error formatting                │
│      exitcodes.go   ─ Cobra typing + usage-help rendering    │
│      pretty.go      ─ Pretty-print with colors               │
│      jq.go          ─ --jq flag filtering                    │
│      color.go       ─ Terminal color utilities               │
│      agentmode.go   ─ Explicit/auto agent mode + structured errors │
│    clierrors/                                                │
│      clierrors.go   ─ Stable 0/1/2/3 process-exit contract   │
│    config/                                                   │
│      config.go      ─ Config file + credential resolution    │
│      keyring.go     ─ OS keychain integration (go-keyring)   │
│    interactive/                                              │
│      policy.go      ─ Shared prompt/form/explorer policy     │
│      interactive.go ─ Missing-flag prompting via huh forms   │
│    explorer/                                                 │
│      *.go           ─ Interactive command explorer TUI       │
│    testclient/                                               │
│      testclient.go  ─ Test HTTP client injection             │
└──────────────────────────────────────────────────────────────┘
```

**Generated per operation**: ~30-55 lines (metadata array + init/run functions).
**Shared runtime**: ~5 packages handling flag registration, request building, output, config, and auth.

The core insight is **metadata-driven request building**: each operation generates a `[]flagutil.FlagMeta` array that describes every flag's name, Go struct field path, type, constraints, and resolution metadata. A single runtime function (`BuildRequest[T]`) uses Go reflection to register flags, parse values, and populate typed request structs from that metadata.

On top of that core, the generator now also layers:

- **Interactive UX** — auto-prompting for unresolved required flags plus an interactive command explorer
- **Agent UX** — AI-agent-aware defaults (TOON output, structured errors, no TUI)
- **Machine-readable usage output** — `--usage` emits a KDL schema for the selected command

---

## Directory Structure

```text
cli/
├── main.ts                          # Entry point: orchestrates all job generation
├── config.ts                        # Dependencies, build/test/lint commands, config fields
├── features.ts                      # Feature flag support matrix
├── standalone.ts                    # Standalone CLI generation (no SDK embedding)
├── examples.ts                      # Usage example generation
├── mockserver.ts                    # Mock server config for tests
│
├── includes/                        # TypeScript helpers (generation-time logic)
│   ├── includes.ts                  #   Loads all dependencies
│   ├── sanitization.ts              #   Go naming, commands, escapeGoString
│   ├── flags.ts                     #   Flag name policies, reserved words, types
│   ├── descriptions.ts              #   Short/Long descriptions for commands/groups
│   ├── templating.ts                #   Operation helpers, security, SDK call args, pagination
│   ├── imports.ts                   #   Go import management
│   ├── unions.ts                    #   Union type inspection (discriminated, expansion)
│   ├── metadata.ts                  #   FlagMeta generation, RegisterFlags/BuildRequest calls
│   ├── security.ts                  #   Global security flag/parsing generation
│   ├── dependencies.ts              #   go.mod dependency management
│   ├── utils.ts                     #   Type classification utilities
│   ├── tests.ts                     #   CLI test utilities, arg formatting, skip detection
│   ├── test-assertions.ts           #   Test assertions, type coercion for Go output
│   └── test-workflows.ts            #   Multi-step tests, cross-step references, workflows
│
├── opcmd.go.stmpl                   # Template: individual operation command
├── root.go.stmpl                    # Template: root command + global flags
├── subroot.go.stmpl                 # Template: subcommand groups (tag-based)
├── configure.go.stmpl               # Template: `configure` interactive command
├── whoami.go.stmpl                  # Template: `whoami` credential display
├── auth.go.stmpl                    # Template: `auth login/whoami/logout` commands
├── usage.go.stmpl                   # Template: machine-readable KDL usage schema emitter
│
├── test.go.stmpl                    # Template: individual test function
├── testfile.go.stmpl                # Template: test file (groups tests)
├── testhelper.go.stmpl              # Template: reusable test helper function
├── testhelpers.go.stmpl             # Template: helper registration file
├── cli_harness_test.go.stmpl        # Template: in-process test executor
│
├── auxiliary/                       # Runtime library (copied into generated CLI)
│   ├── go.mod.stmpl                 #   Module file
│   ├── main.go.stmpl                #   CLI entry point
│   ├── cmd/gendocs/main.go.stmpl    #   Cobra markdown doc generator
│   └── internal/
│       ├── flagutil/
│       │   ├── metadata.go.stmpl    #   Reflection-based flag/request engine
│       │   └── flags.go.stmpl       #   HasStdinInput, AnyFlagsChanged, etc.
│       ├── client/
│       │   ├── client.go.stmpl     #   SDK client construction
│       │   └── diagnostics.go.stmpl #  --dry-run/--debug HTTP wrappers + redaction
│       ├── output/
│       │   ├── output.go.stmpl      #   Result/Error output routing
│       │   ├── exitcodes.go.stmpl   #   Cobra typing + UsageHelpError
│       │   ├── binary.go.stmpl      #   Binary response: WriteToFile, base64, TTY enforcement
│       │   ├── pretty.go.stmpl      #   Aligned key-value pretty printer
│       │   ├── jq.go.stmpl          #   gojq integration
│       │   ├── color.go.stmpl       #   NO_COLOR, TTY detection
│       │   └── agentmode.go.stmpl   #   Explicit/auto agent mode + structured errors
│       ├── clierrors/
│       │   └── clierrors.go.stmpl   #   Process exit-code contract + help footer
│       ├── config/
│       │   ├── config.go.stmpl     #   Config file + credential resolution
│       │   └── keyring.go.stmpl    #   OS keychain abstraction (go-keyring)
│       ├── interactive/             #   Shared interaction policy + prompting
│       │   ├── policy.go.stmpl
│       │   └── interactive.go.stmpl
│       ├── explorer/                #   Bubble Tea command explorer
│       │   ├── explorer.go.stmpl
│       │   ├── model.go.stmpl
│       │   ├── styles.go.stmpl
│       │   └── tree.go.stmpl
│       └── testclient/testclient.go.stmpl
│
├── readme/                          # README section templates
│   ├── agents.stmpl                 #   "For AI agents" discovery ladder + NDJSON dry-run protocol
│   ├── installation.stmpl
│   ├── globals.stmpl
│   ├── pagination.stmpl
│   ├── security.stmpl
│   ├── server.stmpl
│   └── errors.stmpl
├── readme_examples_test.go.stmpl    # README examples contract test (every `<cli> ...` line under --dry-run)
├── dryrun_matrix_test.go.stmpl      # Keyless human/JSON/caller-jq request-preview matrix
│
├── tests/                          # Hand-written tests (variant-specific)
│   ├── primary/                    #   Primary variant tests (21 files)
│   │   ├── servers_test.go.stmpl   #     Server selection
│   │   ├── unions_additional_test.go.stmpl   #     Union edge cases
│   │   ├── collections_additional_test.go.stmpl  #  Null collection handling
│   │   ├── globals_additional_test.go.stmpl  #     Global param + header tests
│   │   ├── hooks_additional_test.go.stmpl    #     SDK hook tests
│   │   ├── pagination_test.go.stmpl          #     Pagination --all tests
│   │   ├── eventstreaming_test.go.stmpl     #     SSE streaming tests
│   │   ├── jsonlstreaming_test.go.stmpl     #     JSONL/NDJSON streaming tests
│   │   ├── auth_additional_test.go.stmpl     #     Auth tests
│   │   ├── errors_test.go.stmpl              #     Error handling tests
│   │   ├── retries_test.go.stmpl             #     Retry tests
│   │   ├── polling_test.go.stmpl             #     Polling tests
│   │   ├── flattening_test.go.stmpl          #     Flattened request tests
│   │   ├── telemetry_test.go.stmpl           #     Telemetry tests
│   │   ├── requestbodies_additional_test.go.stmpl  # Request body edge cases
│   │   ├── responsebodies_additional_test.go.stmpl # Response body edge cases
│   │   ├── parameter_additional_test.go.stmpl      # Parameter edge cases
│   │   ├── headers_additional_test.go.stmpl        # Response header + --include-headers tests
│   │   ├── methods_additional_test.go.stmpl        # HTTP method tests
│   │   ├── enums_additional_test.go.stmpl          # Enum edge cases
│   │   ├── optionalnullable_additional_test.go.stmpl # Optional/nullable tests
│   │   └── diagnostics_test.go.stmpl               # --dry-run / --debug tests
│   ├── secondary/                  #   Server tests (by-name) + auth + redirects
│   ├── tertiary/                   #   Server tests (by-id)
│   ├── quaternary/                 #   Server tests (by-name-with-templates) + auth + transforms
│   ├── oauth2-password/            #   OAuth2 password hook tests
│   ├── client-credentials/         #   Client credentials hook tests
│   ├── custom-http/                #   Custom HTTP client auth tests
│   └── no-servers/                 #   No-servers variant tests
│
├── usage/snippet.stmpl              # Usage examples for docs
└── .gitignore.stmpl
```

---

## Generation Flow

### 1. Job Creation (`main.ts`)

`getJobs()` creates jobs in this order:

1. **SDK Generation** - Generates the Go SDK via `interoptTemplateTarget("go", ...)` with docs/tests disabled
2. **Auxiliary Files** - Copies runtime library (`auxiliary/`) into the output
3. **Command Jobs** - One job per: root command, each subcommand group, each operation
4. **Test Jobs** - Test harness + one job per Arazzo test workflow
5. **Documentation** - README sections
6. **Git Files** - .gitignore, CHANGELOG

### 2. Command Generation

For each API operation, the generator:

1. **Reads the operation AST** (parameters, request body, response type, security)
2. **Generates flag metadata** (`templating.ts::templateFlagMetadataVar()`) — a Go `[]FlagMeta` literal describing each flag
3. **Renders `opcmd.go.stmpl`** with the metadata, producing an init function (registers command + flags) and a run function (builds request, calls SDK, outputs result)

### 3. Request Building Paths

The generator chooses between three request building strategies:

| Condition                                                     | Strategy                                       | Template Function                |
| ------------------------------------------------------------- | ---------------------------------------------- | -------------------------------- |
| Has path/query/header params (with or without body)           | `BuildRequest[T]` with `bodyFieldPath`         | `templateBuildRequestCall()`     |
| `IsRequestBody` + JSON + expandable fields                    | `BuildRequest[T]` with individual flags        | `templateBuildRequestCall()`     |
| `IsRequestBody` + JSON + non-expandable (maps, deeply nested) | `BuildRequestBody[T]` (single JSON flag)       | `templateBuildRequestBodyCall()` |
| `IsRequestBody` + multipart                                   | `BuildRequest[T]` with `FlagKindFile` metadata | `templateBuildRequestCall()`     |

"Expandable" means the request body's fields are all simple types (strings, numbers, bools, enums, shallow objects) that can each become an individual CLI flag. Complex types (maps, `any`, deeply nested objects) fall back to a single `--request-body` JSON flag.

---

## Key Components

### Command Generation

**File**: `opcmd.go.stmpl`

Each operation generates two functions:

```go
// Init: creates the Cobra command, registers flags, validates metadata
func initCreateUserCmd(parentCmd *cobra.Command) {
    cmd := &cobra.Command{
        Use:   "create-user",
        Short: "Create a new user",
        RunE:  runCreateUserCmd,
    }
    flagutil.RegisterFlags(cmd, createUserMeta)
    flagutil.ValidateMeta[operations.CreateUserRequest](createUserMeta)
    parentCmd.AddCommand(cmd)
}

// Run: builds request from flags, calls SDK, outputs result
func runCreateUserCmd(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()
    security := buildGlobalSecurity(cmd)
    s := client.NewClient(cmd, sdk.WithSecurity(security))

    req, err := flagutil.BuildRequest[operations.CreateUserRequest](cmd, createUserMeta, "User", "body")
    if err != nil { return err }

    res, err := s.Users.Create(ctx, *req)
    if err != nil { return output.Error(cmd, err) }
    return output.Result(cmd, res)
}
```

### Metadata-Driven Request Building

**Runtime file**: `auxiliary/internal/flagutil/metadata.go.stmpl`

This is the architectural core. A `FlagMeta` struct is the single source of truth for both flag registration and request parsing:

```go
type FlagMeta struct {
    FlagName    string      // CLI flag name, e.g. "user.first-name"
    FieldPath   string      // Go struct path, e.g. "User.FirstName"
    Kind        FlagKind    // Type: String, Bool, Int64, Float64, Enum, JSON, Union, File, etc.
    Optional    bool        // Pointer type in Go struct
    Required    bool        // Must be provided by user
    HasDefault  bool        // Has a default value
    EnumValues  []string    // Allowed values for Enum kind
    Annotations string      // Go struct tag for serialization
    Union       *UnionMeta  // Non-nil for union fields
    Description string      // Help text shown in --help
    // Default values by type...
}
```

**`RegisterFlags(cmd, meta[])`** iterates the metadata and calls the appropriate `cmd.Flags().StringVar()` / `BoolVar()` / etc. for each entry.

**`BuildRequest[T](cmd, meta[], bodyFieldPath, bodyFlagName)`** creates a `*T`, then for each metadata entry reads the flag value and uses `reflect` to set the corresponding struct field. It handles:

- `--body` flag: if `bodyFlagName` is non-empty and the flag was set, unmarshal its JSON into the body sub-field (highest priority body input)
- Stdin merge: if stdin has JSON and `--body` was not used, unmarshal into the `bodyFieldPath` sub-field
- Individual flags: explicit `--flag` values override both `--body` and stdin (always highest priority)
- Required validation: respects `Required` flag but relaxes for body-populated fields (via `--body` or stdin)
- Pointer wrapping: auto-wraps values in pointers for `Optional` fields
- Union fields: delegates to `buildUnionField()` for multi-variant handling
- Bytes fields: supports raw UTF-8 input, `file:<path>`, and `b64:<base64>` prefixes for `[]byte` request fields

**`ValidateMeta[T](meta[])`** runs at CLI startup (in `init`) and panics if any `FieldPath` doesn't resolve on the request type. This catches generator bugs immediately rather than at runtime.

**Interactive prompting hook**: when interactive mode is enabled, `NewRootCommand` installs `interactive.Intercept(rootCmd)` once on the fully assembled command tree. The interceptor discovers unresolved required flags and declared positionals for generated and hand-written leaf commands, fills them before Cobra's required-flag validation, and leaves every non-interactive execution path untouched.

### Flag Metadata Generation

**Files**: `includes/metadata.ts` (metadata arrays, registration/build calls) + `includes/templating.ts` (field expansion helpers)

Key functions:

| Function                            | File            | Purpose                                                                  |
| ----------------------------------- | --------------- | ------------------------------------------------------------------------ |
| `templateFlagMetadataVar(op)`       | `metadata.ts`   | Generates the package-level `var xxxMeta = []flagutil.FlagMeta{...}`     |
| `collectMetadataFromFields(fields)` | `metadata.ts`   | Recursively walks request struct fields, emitting FlagMeta entries       |
| `buildMetaEntryForField(field)`     | `metadata.ts`   | Builds a single FlagMeta Go literal for one field                        |
| `collectMultipartMetadata(fields)`  | `metadata.ts`   | Classifies multipart fields as File, JSON, or primitives                 |
| `templateBuildRequestCall(op)`      | `metadata.ts`   | Emits `flagutil.BuildRequest[T](cmd, meta, bodyFieldPath, bodyFlagName)` |
| `templateBuildRequestBodyCall(op)`  | `metadata.ts`   | Emits `flagutil.BuildRequestBody[T](cmd, flagName, annotations)`         |
| `templateRegisterFlagsCall(op)`     | `metadata.ts`   | Emits `flagutil.RegisterFlags(cmd, meta)`                                |
| `templateValidateMetaCall(op)`      | `metadata.ts`   | Emits `flagutil.ValidateMeta[T](meta)`                                   |
| `isRequestBodyExpandable(op)`       | `metadata.ts`   | Determines if a request body can be expanded to individual flags         |
| `getBodyFieldPath(op)`              | `metadata.ts`   | Finds which struct field holds the body (via `request` annotation)       |
| `shouldExpandNestedField(field)`    | `templating.ts` | Decides if a nested class becomes dot-notation flags or a JSON flag      |

### Union Type Handling

**Files**: `includes/unions.ts` (inspection), `includes/metadata.ts` (metadata generation), `auxiliary/internal/flagutil/metadata.go.stmpl` (runtime)

Unions are handled through nested metadata:

```go
// Generated metadata for a discriminated union
{
    FlagName:  "shape-request.shape",
    FieldPath: "ShapeRequest.Shape",
    Kind:      FlagKindUnion,
    Union: &UnionMeta{
        Discriminated:    true,
        DiscriminatorKey: "Type",
        Variants: []UnionVariantMeta{
            {
                DiscriminatorValue: "circle",
                FlagName:           "shape-request.shape.circle",
                FieldName:          "Circle",
                CanExpand:          true,
                Fields: []FlagMeta{
                    {FlagName: "shape-request.shape.circle.radius", FieldPath: "Radius", Kind: FlagKindFloat64},
                },
            },
            {
                DiscriminatorValue: "rectangle",
                FlagName:           "shape-request.shape.rectangle",
                FieldName:          "Rectangle",
                CanExpand:          true,
                Fields: []FlagMeta{
                    {FlagName: "shape-request.shape.rectangle.width", FieldPath: "Width", Kind: FlagKindFloat64},
                    {FlagName: "shape-request.shape.rectangle.height", FieldPath: "Height", Kind: FlagKindFloat64},
                },
            },
        },
    },
}
```

At runtime, `buildUnionField()` implements a 3-priority strategy:

1. If a discriminator value flag is set, use that variant
2. If variant fields are set, detect which variant and inject the discriminator
3. If a JSON flag is provided for the variant, unmarshal it directly

Variant conflicts (flags from multiple variants) are detected and reported as errors.

### Declarative Intent Commands and Route Dispatch

**Files**: `internal/extensions/cli_commands.go` and `cli_commands_link.go` (decode/link), `includes/intents.ts` and `intentcmd.go.stmpl` (render), `auxiliary/internal/flagutil/dispatch.go.stmpl` (runtime), `includes/usage.ts`, `includes/readme-commands.ts`, and `includes/readme-examples.ts` (mirrors)

`x-speakeasy-cli-commands` can expose one operation's undiscriminated `oneOf` request as one command with mutually exclusive, typed flag sets. Route IDs are stable author-facing keys; `label` defaults to the humanized ID and is used in help and errors:

```yaml
x-speakeasy-cli-commands:
  version: 1
  commands:
    jobs run:
      category: Create
      override: true
      routes:
        engine:
          label: Engine
          op: createJob#EngineJobParams
          default: true
          selector: engine
          preset:
            $.output_modalities: image
        pipeline:
          label: Pipeline
          op: createJob#PipelineJobParams
          preset:
            $.token_budget: 512
      preset:
        $.background: true
      args:
        input: {to: $.input, variadic: true, required: true}
      flags:
        engine: {to: $.engine, defaultFrom: schema}
        pipeline: {to: $.pipeline}
        pipeline-config: {to: $.pipeline_config}
        background: {to: $.background, type: bool}
```

The `routes:` map must contain at least two routes. Every route names the same operation ID and a distinct component-schema member of that operation's `oneOf` body. `anyOf`, an OpenAPI discriminator, value discrimination through a shared const/single-value enum, inline members under `override`, multiple operations, nested pointers, and route-specific positionals are rejected at generation time. Declared body inputs are top-level scalar strings, integers, floats, or booleans; structured construction remains available through `--body`.

Each route has one anchor flag. The linker infers it only when exactly one route-exclusive bound flag is schema-required, has no schema default, and is not covered by the effective preset. Otherwise `selector:` must name a route-exclusive flag backed by a required or schema-defaulted distinguishing property. Other optional route-exclusive flags are still selection evidence, but are not advertised as anchors. Cross-route bindings must have the same scalar kind; schema-derived defaults and descriptions must also agree, while enum suggestions are intersected.

At runtime, every explicitly changed declared input and every recognized key in a supplied JSON body narrows the candidate routes by membership intersection. Displayed flag defaults (`default:` and `defaultFrom: schema`) and presets are not changed flags and never count as evidence. Conflicting evidence such as `--engine --pipeline` or `--engine --body '{"pipeline":"..."}'` is a typed `CLI_VALIDATION` error and no request is sent. With one candidate it is selected. With several candidates, the declared default wins only when it remains among them; otherwise the error says which anchor flags can resolve the choice.

There may be at most one explicit `default: true`. For safety, it must match the union member selected by the generated `UnionMeta.DefaultJSON` rule (one unique non-null schema default on a distinguishing field), or its effective command-plus-route preset must supply a distinguishing key. A command with no default reports `pass one of --engine, --pipeline` when the invocation supplies no distinguishing evidence.

Selection happens before route-local presets or schema defaults are applied. After selection, required inputs are checked for that route, declared inputs are merged, then the effective preset is filled shallowly (command preset first, route preset overriding it; user keys always win). Conditionally required route inputs are reported before shared requirements, so optional pipeline-only evidence can produce `required flag --pipeline not set for the Pipeline variant`. Only inputs required on every route are annotated as required for interactive prompting; anchors and conditionally required flags are not prompted because the prompt engine has no one-of primitive. A shared required positional may be prompted, after which the explicit default or later flag evidence selects the route.

`--body` remains the advanced escape and must contain a JSON object for dispatch. Its `@path`, `@-`, and stdin forms are resolved exactly once; the runtime selects from the object's keys, merges inputs and the selected preset, and replays the final object through the generic `--body` flag for the normal generated `BuildRequest` path. A body key unique to an unrouted union member returns a typed error naming the generated-operation escape. Unknown keys remain owned by `BuildRequest`'s normal body-key verification. Dispatch requires the generated operation to have flag metadata and the generic `--body` surface; pure `BuildRequestBody` operations are gated.

Dispatch commands never register backing request-body metadata: no whole-union JSON flag and no expanded body fields. They retain non-body path/query/header metadata, operation security, `--body`, `--schema`, and declared flags. Help puts route-specific flags under `<Label> variant Flags:` (or `A / B variants` for a strict multi-route subset), leaves shared flags ungrouped, and appends the anchor/default summary. The static `--usage` KDL contains exactly the same surface.

`override: true` replaces only the generated registration at the operation's exact canonical command path. The command must target that same operation, cover every component member, and replace a non-promoted leaf. The operation file is still generated so the intent reuses its metadata and run function. Runtime registration, KDL usage, generated README command trees and examples all omit the generated operation entry and insert the intent in manifest order; generated Arazzo tests that still encode the removed flag surface are skipped, while dedicated intent tests cover the replacement. Without `override`, the generated operation remains available as the full-control escape.

### Output Formatting

**Files**: `auxiliary/internal/output/`

Three output paths:

1. **Raw JSON passthrough** (`--output-format json` or `--jq`): The CLI adds `WithSkipDeserialization()` to the SDK call, leaving the HTTP response body unread. `tryReadRawBody()` reads it directly and outputs the original API JSON without a deserialize-reserialize round trip. This preserves field ordering and numeric precision from the API response. Raw passthrough is attempted first; if the body was already consumed (e.g., non-JSON response), it falls through to the typed path below.

2. **Typed output** (default, or fallback when raw body unavailable): The response envelope is unwrapped via `extractResultContent()`, then formatted as:
   - `pretty` (default): Aligned key-value pairs with color, nested indentation
   - `json`: Marshaled from the typed Go struct (note: field order and precision may differ from original API JSON)
   - `yaml`: Via `gopkg.in/yaml.v3`
   - `table`: Aligned columns via `printTable()` — multi-row for slices, vertical key-value for single structs, two-column for maps
   - `toon`: [Token-Oriented Object Notation](https://github.com/toon-format/spec) via `github.com/alpkeskin/gotoon` — compact, line-oriented, 30–60% fewer tokens than JSON

   `pretty`, `yaml` and `toon` never reflect over the typed Go structs directly. `normalizeForOutput()` (`output.go.stmpl`) marshals the value once through the SDK's `utils.MarshalJSON` and decodes it back with `json.Decoder.UseNumber`, and the renderers work from that wire shape (`marshalYAML()`, `encodeTOON()`, `prettyPrint()`, shared by `Result()` and the per-item streaming/pagination path in `outputitems.go.stmpl`). Reflecting encoders would otherwise leak generated union internals — wrapper keys such as `ErrorEvent:`, `null` sibling variants, `Type`/`UnknownRaw` — plus raw struct-tag options as keys (`"error,omitzero":`) and lowercased Go field names in YAML. Numbers stay exact: `json.Number` is converted per renderer (`convertNumbers()` with `yamlNumber()`/`toonNumber()`, `formatJSONNumber()` for pretty) to a native `int64`/`float64` only when that value re-encodes to the identical JSON lexeme (`nativeNumber()`); otherwise the lexeme is kept verbatim — a tagged `!!int`/`!!float` scalar node for yaml, a string for TOON (which has no wider numeric type), the literal for pretty — so bigint/decimal fields never lose digits. TOON's own encoder goes through `float64`, so native integers above 2^53 round there (a gotoon property, unchanged). `--include-headers` output (`outputWithHeaders()`) already worked from the marshalled JSON. Covered by `TestEventStreamUnionEventsRenderWireShape` and `TestEventStreamNumbersKeepPrecision` in `tests/primary/eventstreaming_test.go.stmpl`.

3. **jq filtering** (`--jq`): Applies a [gojq](https://github.com/itchyny/gojq) expression to the JSON output. Overrides `--output-format` since the result is always JSON — except under `--raw-output` (jq `-r` semantics: string results are written as raw text plus a newline, unquoted, unescaped, never colorized; non-string results stay JSON), so `--jq '.data' --raw-output | base64 -d > image.png` works. The flag's default is the `cli.jqRawOutput` gen.yaml key (default `false`); it threads through `outputJqResults` (single results, `--include-headers`, artifacts) and `outputOneItem` (streams, `--all` pagination).

**Color default** (`cli.defaultColor` gen.yaml key: `auto` | `always` | `never`, default `auto`): the default value of `--color`. `never` makes uncolored ("raw") output the default so redirected/piped bytes are always clean; `--color`, `NO_COLOR`/`FORCE_COLOR`, and agent mode (never colorized) still apply.

Error output is written to stderr so it never mixes with result data on stdout. `Classify(cmd, err)` runs once for every error in every configuration; renderers and exit-code policy consume the returned `Classification`, and `ClassificationFrom(err)` retrieves the retained decision. A compatibility-mode CLI error that main must still print uses a non-rendered classification carrier; every error printed by the output layer uses a `Rendered()` sentinel carrying the same value, preventing double printing.

#### Unified error classification

**Files**: `auxiliary/internal/output/classify.go.stmpl`, `errortable.go.stmpl`, `output.go.stmpl`, `agentmode.go.stmpl`, `includes/errors.ts`, `internal/extensions/cli_errors.go`

`classifiedErrors` gates rendering only. With its compatibility default `false`, pretty and TOON API errors retain the unclassified `API Error (HTTP <status>):` body dump and 401/403 configure hint, plain CLI errors remain main-printed, and JSON/`--jq` retain the unclassified envelope (including `_hint` for 401/403). Agent mode always renders the classified envelope. With `classifiedErrors: true`, pretty errors use the classified human diagnostic and JSON, TOON, `--jq`, and agent mode use one classified envelope. Classification and `ClassificationFrom` remain available in all cases.

Classification reads structured reason codes at the build's **reason carriers** — the primary carrier (the `reasonPointer` declared in `x-speakeasy-cli-errors`, default `$.reason`) followed by each declared `probes` entry — every carrier resolved against the response body's error object (the nested `error` member when the body has one, else the body root). The first carrier that yields a reason owns `error_reason` (carriers in declaration order): within a carrier, a typed rule wins (scanning values in document order), then a hints-only rule (which fixes `error_reason` and its guidance while the type falls through), then — only at an **explicitly declared** carrier — the first value promoted verbatim. When the selected reason carries no declared type, a typed rule matched at any declared carrier still supplies the error type (its hints stay with the reason-selecting rule); then HTTP status; then status-less stream/protocol/connection/validation evidence. With only the default `$.reason` carrier, an undeclared body reason is never promoted to `error_reason` — declaring a carrier pointer (`reasonPointer`, or any `probes` entry for its own carrier) is the opt-in to verbatim promotion. Pretty `Details:` strips only the carrier members whose value was rendered on the `Reason:` line; unrendered carrier values and every other body member are preserved as detail. A singleton top-level `[{"error": {...}}]` wrapper is unwrapped in every classified mode only when the document declares `unwrapErrorArray: true`; an `error` value of any other shape is not sufficient. Multiple inline errors and multi-element arrays are ambiguous and never supply a reason. Invalid JSON is never classified from a valid-looking prefix. An API error carrying no reason at any declared carrier omits `error_reason`; an HTTP 400 is never assigned the local `CLI_VALIDATION` reason.

`Classification.Origin` is `stream` for `StreamEventError`; `cli` for `CLIError`, `AgentModeError`, and any status-less validation failure such as request serialization; and `api` for HTTP responses and status-less transport/protocol failures. Origin is retained for exit policy and is not emitted in the JSON envelope.

**Declared error conventions** (`x-speakeasy-cli-errors`): the generator compiles no vendor error conventions — everything beyond HTTP-status classification is declared in the OpenAPI document. The extension declares a reason's type and/or hints per exact key, where the reason codes live, and the body-shape normalizations the API needs. The top-level `reasons` rules match against the **primary carrier**: the optional `reasonPointer`, default `$.reason`, a restricted JSONPath resolved against the error object — `$.name` dot children, `$['name']` bracket children, and `[*]` fan-out over array items (the same grammar family as the artifact content pointer; no indexes, filters, slices, unions, or recursive descent), beginning with a named member and ending at the named member that holds the reason string. Additional carriers are declared as `probes`, an ordered list of `{pointer, reasons}` entries consulted after the primary carrier; each probe has its own rule set, so the same reason value may carry different rules at different carriers (a status-style carrier can type a value the detail carrier only hints). Declaring a carrier's pointer (via `reasonPointer` or a probe) also opts that carrier into surfacing undeclared reasons found there verbatim as `error_reason`; with only the default `$.reason` carrier, only declared rules fire. `unwrapErrorArray: true` opts into normalizing the single-element `[{"error": {...}}]` body wrapper before classification (default off). The `service_disabled`/`billing_disabled` account-state types enter the emitted taxonomy only when declared rules reference them; their wording comes from `types` hints or a neutral generated fallback. The extension is independent of `x-speakeasy-cli-commands`, so a CLI without declared intent commands can still declare its error taxonomy.

```yaml
x-speakeasy-cli-errors:
  version: 1
  reasonPointer: $.details[*].reason # optional; default $.reason
  unwrapErrorArray: true # optional; default off
  reasons:
    CREDENTIAL_INVALID:
      type: authentication_error
      hints:
        - Replace the credential before retrying
    CLI_PROTOCOL:
      hints: Report the response payload and CLI version
  probes:
    - pointer: $.status
      reasons:
        ACCESS_DENIED:
          type: authorization_error
  types:
    not_found:
      hints: Check that the resource identifier is correct
```

The decoder is strict: `version: 1` is required; at least one of `reasonPointer`, `reasons`, `probes`, `types`, or `unwrapErrorArray` must be declared (a carrier alone opts into verbatim reason surfacing); unknown or duplicate keys are errors; every carrier pointer must parse under the carrier grammar above (targeted errors mirror the artifact/async pointer diagnostics); `type` must be in the closed emitted taxonomy; and `hints` is a non-empty string or string list without edge whitespace or control characters. Reason keys are exact, case-sensitive strings with the same whitespace/control restrictions. Arbitrary server reasons are allowed. A probe may not duplicate the primary carrier's pointer or another probe's (equivalent spellings collide), and a reason value declared with a type at more than one carrier must map to a single error type. The runtime-owned `CLI_VALIDATION`, `CLI_CONNECTION`, and `CLI_PROTOCOL` reasons may declare hints only, and only under the top-level `reasons` — never inside a probe.

Hint order is server hints, then one taxonomy source (reason hints, declared type hints, or built-in type hints), then `CLIHints()`, then command-specific manifest hints. Exact duplicates are removed in first-occurrence order; malformed server hints never suppress local guidance.

Classified pretty output is `Error (<type>): <message>`, followed by `Reason: <reason> (HTTP <status>)` or the applicable reason/status-only line, then `Fix:` bullets. When the SDK fallback error string ends with its raw body, that suffix is removed from `message`; the unchanged string remains in the fallback envelope's compatibility `error` field. `Details:` appears only for residual response data and unwraps a sole residual `error` object. Classified envelopes add generator-owned `error_type`, `message`, `hints`, optional `error_reason`, and optional `status_code`; generator values win on collision and `_hint` is removed. Object bodies keep their fields, while arrays, primitives, and invalid JSON use the compatibility `error`/`body` fallback envelope.

### Interactive Mode

**Files**: `root.go.stmpl`, `opcmd.go.stmpl`, `intentcmd.go.stmpl`, `auxiliary/internal/flagutil/metadata.go.stmpl`, `configure.go.stmpl`, `auth.go.stmpl`, `auxiliary/internal/interactive/policy.go.stmpl`, `auxiliary/internal/interactive/interactive.go.stmpl`, `auxiliary/internal/explorer/`

Interactive mode is a generator feature, enabled by default via `interactiveMode: true` in `config.ts`. It adds richer terminal UX on top of the standard flag-driven CLI:

- **Tree-wide auto-prompting for missing required inputs** on every runnable leaf, including generated operations and intents, catalog/builtin commands that declare inputs, `custom.Register` commands, and commands preserved through the root custom-code region
- **`explore` command** for browsing commands and launching them from a TUI
- **Auto-launch explorer** when the CLI is run with no subcommand from a TTY and `interactiveByDefault` is enabled
- **Rich `configure` / `auth login` forms** when interactive auth is enabled, governed by the same runtime policy as required-input prompts
- **`--interactive`** opt-in for required-input prompts and guided forms in non-interactive-by-default builds
- **`--no-interactive`** escape hatch to force fully non-interactive behavior

```bash
# Launch the explorer explicitly
cli explore

# Prompt for required fields if the configured default is enabled
cli users create

# Explicitly opt into ordinary-command prompts
cli users create --interactive

# Explicitly open the configure form
cli configure --interactive

# Disable all TUI/prompt behavior
cli users create --no-interactive
```

**Prompting semantics**:

- `interactive.Intercept` is installed after generated intents, both custom-command registration surfaces, Cobra's default help/completion commands, and declared command ordering. `usage.Intercept` is installed immediately afterward as the outer wrapper; `--usage`, `--schema`, and any other flag annotated as a documentation surface never open a prompt.
- `interactive.Resolve` is the single runtime gate for required-input prompts, `configure` / `auth login` forms, and direct explorer validation. `interactive.IsInteractive` remains a compatibility wrapper for hand-written commands. The compatibility-only bare-root auto-launch retains `shouldAutoExplore` so its process-entry behavior is unchanged. `interactiveByDefault` selects the generated default of `--interactive`; `--no-interactive` and active agent mode suppress prompts and forms.
- Required-input prompting needs both stdin and stdout to be TTYs unless a context-injected test prompter supplies terminal capability. Guided forms use the rich TUI on a terminal pair and huh's accessible line prompts off-TTY. Consequently, compatibility builds (`interactiveByDefault: true`) retain their existing off-TTY accessible forms, while non-interactive-by-default builds never read stdin for a form unless `--interactive` is passed.
- Direct `explore` is itself an explicit request, but it rejects `--no-interactive` / `--interactive=false` and requires both stdin and stdout to be TTYs before Bubble Tea is called. Agent mode retains the existing behavior of showing root help instead.
- Interception is additive. When prompting is suppressed or no required promptable input is missing, the command keeps its existing behavior exactly: operation request builders report missing flags, bare required-input intents return typed usage help errors, and hand-written commands retain their own errors/help/exit codes. A `speakeasy:prompt-direct` field participates only after a missing required input has already triggered a form.
- Required flags are prompted when unresolved from **all** declared sources: an explicit/default-resolving flag, environment variable, config value, or changed whole-body flag for a body-derived field. Resolution checks are pure; environment and config values remain for the request builder to apply with its existing precedence. Piped stdin is deliberately not part of this check because a pipe makes the session non-interactive.
- Required positionals are declared with `interactive.Declare`. They may be satisfied by existing args or alternative flags such as `--file`, `--stdin`, `--body`, or an operation's whole-body flag. Required positionals are collected before required flags; optional positionals are asked directly afterward and may be left empty; `speakeasy:prompt-direct` flags come next; flags explicitly annotated for optional prompting retain the existing “fill in optional fields?” offer. Optional intent-declared flags are not annotated and therefore do not change the tier-1 prompt flow. Artifact intents use the direct class for an optional `Output file` (`flag:out`) field whose description names the concrete `defaultPath` and says a directory is accepted.
- Cobra validates positionals before lifecycle hooks, so the Args wrapper may defer an arity error only when an interactive prompt can supply a missing declared required positional. The PreRunE wrapper applies answers and always re-runs a deferred validator. Because Cobra computed its argument slice once, RunE and PostRunE wrappers receive the effective injected args through per-command execution state.
- Positional prompting requires an explicit `interactive.Declare` contract; commands without `speakeasy_prompt_args` retain their original `Args` function unchanged. In v1, expanded union variant flags carry `speakeasy:union-member` and are never prompted independently, because requiring every variant at once would be incorrect. Callers select/populate unions through their existing top-level or variant surfaces.
- Route-dispatch intents annotate only inputs required by every route. Anchors and conditionally required inputs are deliberately not marked required: the current prompt plan has no one-of choice, so prompting both anchors (e.g. `--engine` and `--pipeline`) would create an invalid request. The shared prompt runs first; route selection or a typed conditional-required error follows in `RunE`.
- Declared intent commands (`x-speakeasy-cli-commands`) keep their presets under a partial body: when the caller passes `--body` (or the operation's whole-body-field flag, e.g. the union flag of a mixed operation) the intent merges its presets into that JSON per key — caller keys win, presets fill the gaps including the key that selects the pinned request variant, a known discriminator value is filled in — instead of handing the body to the operation untouched, where the union's default variant would silently replace the pinned one. A body that names another variant (a foreign selector key, or the discriminator with a different value) is a typed `CLI_VALIDATION` usage error naming the intent, the variant, and the full-control escape command. The selector knowledge (`CLIVariantSelectors`: own/foreign distinguishing keys, discriminator key + pinned value) is derived by the decoder's link step (`variantSelectors` in `cli_commands_link.go`) and rendered into `flagutil.PresetMerge` (`intentcmd.go.stmpl`, runtime in `auxiliary/internal/flagutil/preset.go.stmpl`).

**Annotation contract**:

| Annotation                                           | Values                                                                                      | Meaning                                                                                                                   |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `speakeasy:required`                                 | `["true"]`                                                                                  | Required promptable flag. Cobra's `cobra.BashCompOneRequiredFlag` annotation from `MarkFlagRequired` is accepted too.     |
| `speakeasy:env` / `speakeasy:config`                 | one env/config key                                                                          | Treat an existing value as resolved without mutating the flag.                                                            |
| `speakeasy:default-resolves`                         | `["true"]`                                                                                  | The registered default satisfies the input.                                                                               |
| `speakeasy:body-field`                               | whole-body flag names                                                                       | Any changed listed flag satisfies this body-derived field.                                                                |
| `speakeasy:body-flag`                                | `["true"]`                                                                                  | Identifies a whole-body JSON surface.                                                                                     |
| `speakeasy:doc-surface`                              | `["true"]`                                                                                  | A changed truthy flag suppresses prompting (`--schema` uses this).                                                        |
| `speakeasy:prompt`                                   | `["true"]`                                                                                  | Offer an unresolved optional flag after required fields.                                                                  |
| `speakeasy:prompt-direct`                            | `["true"]`                                                                                  | Once a missing required input triggers a form, prompt this unresolved optional flag directly before the optional-fields confirmation gate.|
| `speakeasy:prompt-label`                             | one display label                                                                           | Overrides the humanized flag name in the form (artifact `--out` uses `Output file`).                                      |
| `speakeasy:prompt-kind`                              | `string`, `bool`, `int64`, `float64`, `duration`, `enum`, `json`, `file`, or `string-array` | Widget/value parser hint. Missing hints fall back to `pflag.Flag.Value.Type()`, including plain hand-written Cobra flags. |
| `speakeasy:prompt-values` / `speakeasy:prompt-order` | enum values / integer                                                                       | Options and stable field ordering.                                                                                        |
| `speakeasy:union-member`                             | top-level union flag name                                                                   | Excludes expanded variant fields from prompting in v1.                                                                    |

`interactive.Declare(cmd, interactive.CommandSpec{Args: []interactive.ArgSpec{...}})` serializes one command annotation named `speakeasy_prompt_args` with this versioned shape:

```json
{"version":1,"args":[{"name":"target","summary":"Target to process","required":true,"variadic":false,"body_key":"input","satisfied_by":["file","stdin","body"]}]}
```

Omitted zero-value fields are absent in the serialized form. `Declare` also populates `speakeasy_args_help` only when the command has no existing argument help and no rendered `Arguments:` section. Both static intent KDL and the live `--usage` fallback render descriptor entries as `arg` nodes.

**Custom command authoring**:

Register the command normally in `internal/cli/custom.Register` or the root `#region custom-commands`, then use one of the standard contracts. `cmd.MarkFlagRequired("project")` is enough for a plain required flag; `flagutil.AnnotatePromptFlag` adds resolution, ordering, optional-form, or widget metadata. Use `interactive.Declare` for positionals, including alternatives:

```go
cmd.Flags().String("file", "", "Read input from a file")
_ = interactive.Declare(cmd, interactive.CommandSpec{Args: []interactive.ArgSpec{
    {Name: "text", Summary: "Text to process", Required: true, Variadic: true, SatisfiedBy: []string{"file", "stdin"}},
}})
```

Tests should use `interactive.WithPrompter(ctx, p)` (or the generated harness's `RunInteractive(args, p)`). The injected prompter substitutes only for the TTY check; it does not bypass a false `--interactive` value, `--no-interactive`, agent mode, usage, or documentation surfaces. The seam exposes stable field IDs (`arg:<name>`, `flag:<name>`), field order, applied values, and exported `PromptField.Direct` without driving Bubble Tea through a pipe.

**Known startup cost**: huh imports Bubble Tea, whose current package initialization performs a terminal background-color query that can wait about five seconds on PTYs which never answer. The non-interactive policy prevents prompt/TUI execution but cannot avoid that initialization while the Charm packages are linked; eliminating the stall requires a dependency-level or process-boundary change.

**Generator config knobs**:

| Config key                  | Default          | Effect                                                                                                        |
| --------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------- |
| `interactiveMode`           | `true`           | Generates explorer + operation prompting support                                                              |
| `interactiveByDefault`      | `true`           | Enables implicit prompts/forms and bare-root explorer launch; set false to require `--interactive`            |
| `interactiveAuth`           | `true`           | Generates rich auth/configure forms and auth command group                                                    |
| `interactiveTheme`          | built-in palette | Controls accent/dimmed/subtle/error/success colors for forms                                                  |
| `agentEnvironmentDetection` | `true`           | Includes compatibility detection for known agent environments; set false to require explicit `--agent-mode`   |
| `classifiedErrors`          | `false`          | Enables classified pretty output and classified JSON, TOON, and `--jq` envelopes; classification always runs |

### Agent Mode

**Files**: `root.go.stmpl`, `configure.go.stmpl`, `auth.go.stmpl`, `auxiliary/internal/output/agentmode.go.stmpl`, `auxiliary/internal/output/output.go.stmpl`

Agent mode is a runtime behavior optimized for AI coding agents and other non-human terminal drivers.

```bash
# Explicitly enable agent mode
cli users list --agent-mode

# Disable compatibility auto-detection when it is generated
cli users list --agent-mode=false
```

**What agent mode changes**:

- **Default output format becomes `toon`** unless `--output-format` or config overrides it
- **Errors become structured JSON** on stderr, with machine-readable fields like `error_type`, `exit_code`, `message`, and `hints`
- **Interactive surfaces are blocked or suppressed**:
  - explorer will not auto-launch
  - `explore` falls back to help text
  - `configure` and `auth login` use their flag-only storage path instead of launching forms; missing flags become the ordinary structured usage error
- **Compatibility auto-detection** is included when `agentEnvironmentDetection` is `true`; setting it to `false` omits the environment-variable allowlist from the generated runtime and requires explicit `--agent-mode`

**Precedence when detection is generated**: explicit `--agent-mode` flag > environment auto-detection

**Classified error rendering** (`classifiedErrors`, default `false`): `Classify` and `ClassificationFrom` are unconditional; this key selects only the renderer. When enabled, pretty output uses the classified diagnostic and explicit JSON, TOON, and `--jq` use the classified envelope. When disabled, non-agent rendering keeps the unclassified pretty/TOON and JSON/jq behavior (`renderUnclassifiedAPIError`). Agent mode always renders the classified envelope.

Dry-run does not depend on agent environment detection. Successful previews use the human stderr protocol unless the caller explicitly selects `--output-format json` or `--jq`, while validation failures compose with `classifiedErrors` and remain classified envelopes on stderr with empty stdout.

**Error classification** (`classify.go.stmpl`): the first carrier yielding a reason owns it (typed rule, then hints-only rule, then verbatim promotion for declared carriers), a typed rule at any carrier backfills the type, then HTTP status and status-less evidence. Singleton array envelopes are unwrapped only when the document declares `unwrapErrorArray` and only for classified rendering; ambiguous multi-error bodies never contribute a reason. Precise transport and protocol evidence remains distinct from local validation; a status-less error with no transport, protocol, or typed evidence is `runtime_error` / `CLI_RUNTIME`, never guessed to be usage. `Classification.Origin` records stream events, local validation, or API/transport failures for exit policy.

**CLI-originated errors** (`ExecuteRoot` in `root.go.stmpl`, `CLIError` in `agentmode.go.stmpl`): every CLI error is classified, including compatibility-mode errors that main still prints. Flag/argument/JSON parsing failures, Cobra usage errors, invalid option values, and local input/config failures are typed as `CLI_VALIDATION` at their structural boundaries (`flagutil.WithCLIValidation`); an error carrying a `clierrors.ExitCoder` classifies from its code (`ExitAuth` → `CLI_AUTHENTICATION`, `ExitUsage` → `CLI_VALIDATION`). `AgentModeError(cmd, message, hints)` is `validation_error` / `CLI_VALIDATION`. Runtimes may attach `CLIReason()` and `CLIHints()`; command hints append after them. `output.InstallErrorHandling` composes final-tree Cobra flag/Args handlers and pre-validates required/flag-group contracts, while `ExecuteRoot` uses the already-computed structural `root.Find` failure to type root unknown commands without recreating Cobra's message. `output.UsageHelpError` is the reusable exported helper for incomplete hand-written/generated commands: plain mode writes full help to stderr then `Error: ...`; agent mode, explicit JSON or `--jq`, and explicit TOON when `classifiedErrors` is enabled return one envelope with a command-specific `--help` hint. `ExecuteRoot` snapshots rendering flags before Cobra can reject an early command. Rendered diagnostics return one sentinel containing the classification, so `main.go` never prints a second copy.

The closed reason namespace is `CLI_VALIDATION`, `CLI_CONNECTION`, `CLI_PROTOCOL`, `CLI_RUNTIME`, `CLI_UNAVAILABLE`, and `CLI_AUTHENTICATION`; it mirrors `CLIRuntimeHintReasons` in `internal/extensions/cli_commands.go`. Interactive prompt cancellation and TUI/explorer failures deliberately remain untyped and therefore exit as runtime failures.

### Exit Codes and Error Boundaries

**Files**: `auxiliary/internal/clierrors/clierrors.go.stmpl`, `auxiliary/internal/output/exitcodes.go.stmpl`, `auxiliary/internal/output/classify.go.stmpl`, `auxiliary/internal/output/output.go.stmpl`, `root.go.stmpl`

Every generated CLI advertises and implements one process contract:

| Code | Meaning | Representative types/reasons |
| ---- | ------- | ---------------------------- |
| `0` | success | help, version, successful request or dry-run |
| `1` | runtime/API | `runtime_error`, `unsupported_error`, connection/protocol/server/rate-limit/not-found/service/billing failures; `CLI_RUNTIME`, `CLI_UNAVAILABLE`, `CLI_CONNECTION`, `CLI_PROTOCOL` |
| `2` | usage/input | `validation_error`, Cobra parsing/Args/required failures, invalid request flags/body/jq expression; `CLI_VALIDATION` |
| `3` | auth | `authentication_error`, `authorization_error`; `CLI_AUTHENTICATION` |

The exit code is derived from the final `Classification.Type` (`clierrors.ErrorTypeExitCode`) and attached by the output layer's rendered/classified sentinels, so the code the process exits with is always the code shown as `exit_code` in the envelope. `clierrors.CodedError` preserves `errors.Is`/`errors.As`; the outermost classification boundary wins, invalid codes normalize to runtime, and only a nil error is code 0. All structured error envelopes include the numeric `exit_code`; agent mode additionally provides the descriptive classification fields. `clierrors.HelpFooter` is the single footer constant shared by Cobra help and generated command documentation.

Boundary rule: type predictable caller-correctable failures where the runtime still has structural knowledge (Cobra validators, request builders, explicit flag/config parsers, jq compilation, and local input-file reads). Leave output-file I/O, marshaling, hooks, SDK serialization after validation, interactive cancellation, and other unknown failures untyped; **untyped means runtime (1)**. Do not classify by matching arbitrary error text.

### Pagination

**Files**: `opcmd.go.stmpl` (flag registration + branching), `auxiliary/internal/output/paginated.go.stmpl` (pagination runtime), `auxiliary/internal/output/output.go.stmpl` (raw-body peeking + success-diagnostic policy), `includes/templating.ts` (AST helpers)

Operations with `op.Extensions?.Pagination` get an `--all` flag that streams through all pages using the SDK's generated `Next()` closures:

```bash
# Single page (default)
cli list-items --page 1

# All pages, streamed as NDJSON
cli list-items --all --output-format json

# Limit to 5 pages
cli list-items --all --max-pages 5
```

**How it works**:

1. **Template time**: `hasPagination(op)` checks `op.Extensions?.Pagination`. If true, `opcmd.go.stmpl` registers `--all` and `--max-pages` flags and generates a branch that calls `output.PaginatedResult()` instead of `output.Result()`
2. **Field name resolution**: `getPaginationContentFieldName(op)` finds the content field in the response envelope (skipping `HTTPMeta`, `ContentType`, etc.), and `getPaginationResultsFieldName(op)` resolves `Pagination.Outputs.ResultsDot` from JSON names to Go field names
3. **Truthful continuation probe**: `getPaginationProbe(op)` passes the pagination type, raw `nextCursor` expression, cursor input kind, declared results expression, and whether a mutable limit exists to the runtime. `HasMorePages()` evaluates the response with `spyzhov/ajson`, matching the Go SDK's generated `Next()` semantics: numeric cursors are coerced to the request cursor type, cursor strings must be non-blank after trimming, and a missing, non-array, or empty declared results value is terminal. The raw response body is restored after probing. Any probe failure returns false. Offset/limit pagination never guesses from result count. The runtime can also probe a next URL when that pagination type is supported by the target generator.
4. **Runtime**: `PaginatedResult()` uses `callNext()` (reflection-based `Next()` invocation) to loop through pages, extracting items via `extractFieldByPath()` and streaming each item via `outputOneItem()`

**Two output modes**:

| `ResultsDot` available?                                          | Behavior                                                         |
| ---------------------------------------------------------------- | ---------------------------------------------------------------- |
| Yes (e.g., offset/limit with `resultArray`)                      | Extracts items from the results array, streams one item per line |
| No (e.g., cursor pagination without explicit results output)     | Outputs the entire content object per page                       |

**Design choices**:

- `--all` branch is placed **before** `WithSkipDeserialization()` — the SDK needs full deserialization to set up `Next()`
- `--all` branch is placed **before** `DrainRawBody()` — `Next()` closures need the response body
- Results are streamed (not buffered) — safe for large result sets
- `ctx.Done()` is checked each iteration for clean cancellation
- For supported non-polling pagination, `--timeout` applies per page: `Next()` runs with the caller's context, not the first request's timed-out context
- `--max-pages` is valid only with `--all`; negative values fail before client construction, and `0` means unlimited
- Reaching `--max-pages` never calls `Next()` just to inspect another page. A human-only diagnostic is emitted only when the current cursor probe proves a continuation exists
- Cursor continuation keys are tracked across the whole run. Repetition or a cycle stops before another request with a `CLI_PROTOCOL` error; offset/limit pagination has no continuation-key guard. The same runtime guard applies to next URLs when that pagination type is supported by the target generator
- Cursor pagination that declares both a results expression and a mutable limit never emits a continuation hint because the probe cannot reproduce the SDK's runtime limit comparison. The cycle guard still rejects a repeated cursor alongside a non-empty short page; although the SDK's limit check might otherwise stop, the repeated server cursor is treated as a protocol violation
- Success hints and max-page diagnostics are written only for non-agent `pretty` / `table` output without `--jq`. JSON, YAML, TOON, jq, and agent-mode stderr remain silent on success

Without `--all`, a human-oriented single-page response with a usable cursor prints:

```text
Hint: more pages available. Use --all to fetch all results, or --cursor for manual pagination.
```

The manual flag is included only for cursor pagination. Offset/limit pagination emits no hint; target generators that support next-URL pagination can use the same probe without a manual flag.

**Hand-written tests**: `tests/primary/pagination_test.go.stmpl` covers `--all` with offset/limit, cursor, NDJSON, jq, pretty format, truthful hints, structured/agent silence, exact request caps, continuation cycles (including cursor coercion and empty-results termination), and operation-option propagation.

### Streaming (SSE & JSONL)

**Files**: `opcmd.go.stmpl` (branching), `auxiliary/internal/output/output.go.stmpl` (`StreamResult`), `includes/templating.ts` (`hasStreamingResponse`, `getStreamingFieldName`)

Operations that return SSE event streams (`text/event-stream`) or JSONL/NDJSON streams (`application/jsonl`, `application/x-ndjson`) are detected at template time and generate a streaming code path. Events are output incrementally as NDJSON (one JSON object per line), identical to how `--all` pagination works.

```bash
# SSE: each event appears as a separate JSON line
cli eventstreams chat --prompt "Hello"
{"data":{"content":"Hello"}}
{"data":{"content":" world"}}
{"data":{"content":"!"}}

# JSONL: each JSON object on its own line
cli jsonl jsonl-stream
{"name":"Peter","skills":["Go","Python"]}
{"name":"John","skills":["Go","Rust"]}

# --jq filter applied per-event
cli eventstreams chat --prompt "Hello" --jq '.data.content'
"Hello"
" world"
"!"

# YAML output with document separators
cli jsonl jsonl-stream --output-format yaml
name: Peter
skills:
  - Go
  - Python
---
name: John
skills:
  - Go
  - Rust
```

**How it works**:

1. **Template time**: `hasStreamingResponse(op)` checks `SerializationMethod` on response body content for `"eventstream"` or `"jsonl"`. If true, `opcmd.go.stmpl` generates a branch that calls `output.StreamResult()` instead of `output.Result()`
2. **Field name resolution**: `getStreamingFieldName(op)` finds the stream field on the response type by checking `Type.Type.toString()` for `"event-stream"` or `"jsonl"` (skipping envelope fields like `HTTPMeta`, `ContentType`, etc.)
3. **Runtime**: `StreamResult(cmd, res, streamFieldName)` uses reflection to extract the stream field from the response envelope, validates it has the streaming interface (`Next()`, `Value()`, `Err()`, `Close()`), then iterates events via `outputOneItem()`

**Key design decisions**:

- **No special flags needed** — streaming is the natural behavior for these endpoints (unlike pagination's `--all`)
- **No `WithSkipDeserialization()`** — the SDK needs full deserialization to create the stream iterator
- **Handles both SSE and JSONL** — SSE `Value()` returns `*T` (1 value), JSONL `Value()` returns `(T, error)` (2 values). `StreamResult` handles both via `len(valueResults)` check
- **Error routing** — all stream errors go through `Error(cmd, err)` for consistent `--output-format json` behavior
- **In-band SSE error events are errors, not data** — before rendering each SSE item, `StreamResult` runs `streamErrorEvent` (`streaming.go.stmpl`): if the item is a generated discriminated union (a struct with `union:"member"` fields), or its `Data` member is one, and the union's resolved `Type` (the wire discriminator value) is exactly `error` or ends in `.error`/`_error`/`-error` (`isErrorDiscriminator`, case-sensitive so PascalCase variant type names of undiscriminated unions such as `ErrorEvent` never match), the event is wrapped in `output.StreamEventError` (Body = the union value's JSON) and routed through `Error()`. The command exits non-zero, pretty mode prints the event body under `API Error:`, and json/agent mode emits the usual envelope with reason-first classification (`error.status`/`error.details[].reason`); when nothing classifies, a status-less stream error event is `api_error` (`classifyNoStatusError` in `agentmode.go.stmpl`), never `connection_error`. Events streamed before the error event are still output. JSONL items are not inspected. Covered by `tests/primary/eventstreaming_test.go.stmpl` against the CLI-only ops added by `tests/overlays/primary/cli/overlay.yaml` (`eventstreams discriminated-events` / `undiscriminated-events`).
- **Fallback to `Result()`** — if stream field is missing or doesn't have the streaming interface, falls back gracefully
- **Context cancellation** — checks `ctx.Done()` each iteration for clean Ctrl+C handling

**Mock server gap**: The generated mock server doesn't produce SSE/JSONL handlers. Streaming tests connect directly to the API test service via `getAPITestServiceURL()` and `RunWithServer()`.

**Hand-written tests**: `tests/primary/eventstreaming_test.go.stmpl` (14 SSE tests + cross-cutting jq/yaml) and `tests/primary/jsonlstreaming_test.go.stmpl` (5 JSONL tests + cross-cutting jq).

#### Streamed-event projection (`x-speakeasy-cli-commands` `output.stream.select`)

**Files**: `auxiliary/internal/output/streamproject.go.stmpl` (`streamProjector`), `streaming.go.stmpl` (dispatch), `intentcmd.go.stmpl` / `opcmd.go.stmpl` (`speakeasy_stream_select` annotation), `includes/intents.ts`, `intentstream_test.go.stmpl` / `opstream_test.go.stmpl` (timing proofs)

A declared intent over a streaming operation can select the field of each event whose **string value is written raw to stdout as the event arrives**:

```yaml
generate:
  op: CreateInteraction#CreateModelInteractionParams
  flags:
    stream: {to: $.stream, type: bool}
  output:
    stream:
      select: $.data.delta.text   # singular JSONPath from the event root (what a per-event --jq sees)
```

```bash
cli generate "tell me a story" --stream
Once upon a time…            # bytes land as each event arrives; one Write per event, no framing
```

Contract (settled in design review):

- The path is decoded strictly (singular JSONPath subset, same as binds) and linked against every event shape of the operation's streaming success response (union arms, `allOf`, array items). Absent in every arm is a generation error with a did-you-mean; a provably non-string leaf is a generation error; unverifiable schemas (open, untyped, cycles) warn.
- Per event: string → written raw (no quoting, no separator, no newline beyond the content), immediately, then `Flush()` if the writer has one; missing/null → skipped (streams are polymorphic); a non-string at runtime is an error naming the pointer and kind (schema drift), never a mixed raw/JSON stream. Bytes are the UTF-8 of the decoded string (JSON decoding already resolves escapes; lone surrogates become U+FFFD).
- On clean completion, one LF is appended iff something was written and it did not end with LF (the shell prompt lands on its own line, the payload is never altered).
- Applies only when the user has not chosen a rendering: an explicit `--output-format` (flag, env, or config — `outputFormatExplicit`) or a user `--jq` keeps the existing per-event rendering (NDJSON under `json`, jq results, YAML/TOON). Implicit defaults — pretty, or agent mode's TOON — yield to the projection; a manifest `output.jq` default is not a user choice and also yields in stream mode.
- The write path is unbuffered end to end: cobra `OutOrStdout()` → `os.Stdout` (a `write(2)` per event); the SDK `EventStream` scanner returns each event as soon as its boundary is read. `--debug` no longer captures streaming response bodies (`isStreamingResponse` in `diagnostics.go.stmpl`), which previously held every event until the server closed the stream.

**Timing proof**: `intentstream_test.go.stmpl` (emitted when the manifest declares an SSE projection whose inputs are satisfiable from one positional) runs a mock SSE server that pauses 300ms between events and, between events, waits until the CLI's stdout writer has observed the previous event (a causal handshake — a batch-at-end implementation deadlocks it and fails, verified by mutation). It asserts one byte-exact `Write` per event, ≥200ms between consecutive event writes, exact concatenation plus one LF, and the explicit `--output-format json` / `--jq` / `--output-format pretty` opt-outs.

The document-level manifest also accepts an `operations:` map, sibling to `commands:`. This is opt-in: specs without it retain their existing operation commands. It adds the same raw stream projection to the generated operation command and can add scalar body flags needed by non-expandable union bodies:

```yaml
x-speakeasy-cli-commands:
  version: 1
  operations:
    CreateInteraction:
      output:
        stream:
          select: $.data.delta.text
      flags:
        stream:
          to: $.stream
          type: bool
          defaultFrom: schema
          summary: Stream events as they arrive
```

`operations` is a strict map keyed by resolvable `operationId`; each value permits only `output.stream.select` and `flags`. Unknown/duplicate keys, nonexistent operations, non-streaming projections, and flag collisions are generation errors. Operation flags use the same `to`, `type`, `shorthand`, `summary`, `default`, and `defaultFrom: schema` keys as intent flags, but cannot be `required`, variadic, or bind outside a single top-level body property. For a union request body the property must exist as the same scalar type in every variant. `defaultFrom: schema` requires the same actual default in every variant; the value appears in help while the generated SDK remains responsible for materializing that schema default on the wire.

Operation-declared names are checked against generated parameter/body/security flags, inherited/root flags, `--schema`, `--body`, the operation whole-body flag, pagination (`--all/-a`, `--max-pages`), and binary output (`--output-file`, `--output-b64`). Each registered flag carries `speakeasy:op-declared-input=<body-key>`. The shared operation run function merges only flags with that annotation on the invoking Cobra command, so an intent that reuses the run function and independently declares `--stream` is unaffected. A changed operation flag merges into `--body`, the union whole-body flag, or stdin and produces the same duplicate-source error if that JSON already contains the key; `@path` and `@-` body values are resolved before this merge for both operation and intent inputs. With no body, flags synthesize a minimal JSON object only when the OpenAPI request body is optional; a required request body keeps its existing CLI-side missing-body error.

The projection annotation is attached only to the operation command. SSE-overload operations need no decode-time branch: `StreamResult` projects a real stream and already falls back to normal `Result` output when `--stream=false` produces JSON (including completed-interaction replay APIs). Generated long help explains raw projection, `--stream=false`, and `-o json`; static KDL and live `--usage` include operation-declared flags, and Cobra-generated docs inherit the same flags/help.

Ordinary typed `Result` extraction never treats a value with `Next` and `Value` methods as renderable response content; stream wrappers belong exclusively to `StreamResult`. This also keeps a streaming dry-run silent in implicit pretty mode when the dry-run client creates an empty synthetic stream. Explicit JSON dry-runs print the `{"dry_run":true,"request":{...}}` preview object like every other operation.

#### Artifact intent output (`output.artifact`)

The extracted content-block shape is declaration-driven. `block:` binds the member names the runtime reads on each candidate block — `typeField` (the const discriminating the block's kind), `dataField` (base64 payload), `mimeTypeField`, `uriField` — and `identity:` binds the response-root members enriching the reported envelope and not-ready diagnostics — `idField`, `statusField`, plus `terminalStatus`, the status value at which content is expected. Every binding defaults to the v1 shape (`type`/`data`/`mime_type`/`uri` blocks; `id`/`status` roots with terminal `completed`), and default bindings are omitted from the `speakeasy_artifact` annotation, so declarations without a `block:`/`identity:` render byte-identical output. The linker always validates the content pointer and the effective block shape against the response schema; declaring `identity:` (any member) opts that group into strict linking too — each bound member must be a string at the response root in every variant, and a closed status enum must contain `terminalStatus`. A defaulted identity stays the v1 best-effort runtime enrichment and is deliberately not linked, so existing declarations cannot start failing generation.

```yaml
output:
  artifact:
    contentPointer: $.outputs[*]
    kind: image
    defaultPath: "render-{timestamp}-{rand}.{ext}"
    block: # optional; defaults are the v1 shape
      typeField: block_kind
      dataField: b64
      mimeTypeField: content_type
      uriField: content_uri
    identity: # optional; declaring it opts into strict linking
      idField: job_ref
      statusField: phase
      terminalStatus: done
```

Artifact intents register `--out` and `--raw-response`. When a missing required input triggers the interactive form, an unresolved `--out` is offered directly after required inputs and optional positionals as `Output file`, before the ordinary optional-fields confirmation. An argv-complete invocation does not open a form just for `--out`. Empty input uses the declared `defaultPath`; an existing directory or a path ending in a separator places that generated default name inside the directory. `--out` supplied on argv, `--no-interactive`, and agent mode do not prompt.

The returned MIME type controls a file destination's extension. A matching extension stays unchanged; a missing extension is appended silently; a conflicting extension is rewritten and stderr receives one line such as `Note: wrote image/png as photo.png (requested photo.jpg)`. Like the human-only `Wrote image to ...` status, the note is suppressed in machine modes (agent mode, `--jq`, non-pretty/table formats): the envelope already carries the final path and success keeps stderr silent there. `--out` is output routing only and never participates in request-body source/merge rules.

### Binary Downloads

**Files**: `opcmd.go.stmpl` (flag registration + branching), `auxiliary/internal/output/output.go.stmpl` (binary detection in `Result()`), `auxiliary/internal/output/binary.go.stmpl` (`WriteToFile`, TTY detection), `includes/templating.ts` (`hasBinaryResponse`, `isOnlyBinaryResponse`, `anyOperationHasBinary`)

Operations that return binary responses (`application/octet-stream`, file downloads, etc.) get special handling to prevent binary data from being corrupted by JSON formatting or dumped to an interactive terminal.

```bash
# Save binary response to a file
cli files download --id file-123 --output-file ./report.pdf

# Pipe binary to another command (non-interactive terminal)
cli files download --id file-123 | sha256sum

# Binary-only operations enforce --output-file on interactive terminals
cli files download --id file-123
# Error: binary response cannot be written to a terminal; use --output-file <path> or pipe to another command
```

**How it works**:

1. **Template time**: `hasBinaryResponse(op)` checks response content for `SerializationMethod "raw"` or `Type.Type` of `"response-stream"` / `"bytes"`. If true, `opcmd.go.stmpl` registers `--output-file` and generates the binary handling branch
2. **Skip deserialization excluded**: Binary ops do NOT get `WithSkipDeserialization()` — the SDK needs full deserialization to populate `io.ReadCloser` / `[]byte` fields from the HTTP response body
3. **`--output-file` flag**: `WriteToFile()` in `binary.go.stmpl` handles `io.ReadCloser` (streaming copy), `[]byte` (direct write), and non-binary fallback (JSON marshal). Byte count is reported to stderr
4. **TTY enforcement**: For binary-only operations (`isOnlyBinaryResponse`), if stdout is an interactive terminal, an error is returned directing the user to `--output-file` or piping. Mixed binary+JSON operations skip TTY enforcement — if the actual response is JSON, `Result()` handles it normally
5. **`Result()` binary check**: Binary type assertions (`io.ReadCloser`, `[]byte`) are checked before the raw JSON passthrough block in `Result()`, so `--output-format json` is silently ignored for binary responses

**Detection types**:

| AST indicator                 | Go type         | Example                                       |
| ----------------------------- | --------------- | --------------------------------------------- |
| `SerializationMethod "raw"`   | `io.ReadCloser` | File download with `application/octet-stream` |
| `Type.Type "response-stream"` | `io.ReadCloser` | Streaming binary response                     |
| `Type.Type "bytes"`           | `[]byte`        | Buffered binary response                      |

**Design choices**:

- **Mixed-response operations** (both binary and JSON content types): `--output-file` is available but TTY enforcement is skipped. If the actual response is binary and stdout is a terminal, the binary bytes stream directly (the user opted into non-TTY piping or knows the response type). This is an intentional tradeoff — enforcing `--output-file` on mixed ops would break JSON responses
- **`--output-format json` + binary**: Binary check runs first, streaming bytes directly. The format flag is silently ignored rather than producing an error
- **`--jq` + binary**: Same behavior — binary check is before the jq pipeline
- **`--output-file` + non-binary**: `WriteToFile()` handles this gracefully by marshaling to JSON and writing to the file
- **Test harness safety**: `cmd.SetOut(*bytes.Buffer)` lacks an `Fd()` method, so `isInteractiveTTY()` returns false in tests — TTY enforcement is never triggered during testing

**Extra output mode**: binary operations also support `--output-b64`, which base64-encodes the response and prints it to stdout. For streamed binary responses (`io.ReadCloser`), encoding is performed incrementally so the entire payload is not buffered in memory.

```bash
# Print a binary response as base64 instead of raw bytes
cli files download --id file-123 --output-b64
```

**Hand-written tests**: `tests/primary/responsebodies_additional_test.go.stmpl` — binary content validation (byte-level comparison) and `--output-file` write + verify.

### Bytes / Base64 Request Input

**Files**: `includes/flags.ts`, `auxiliary/internal/flagutil/metadata.go.stmpl`

Request-side `bytes` fields are treated specially. The generated help text advertises them as:

```text
binary data (file:<path> or b64:<base64>)
```

At runtime, `FlagKindBytes` supports three input forms:

| Input form               | Behavior                               |
| ------------------------ | -------------------------------------- |
| `file:/path/to/data.bin` | Reads bytes from disk                  |
| `b64:SGVsbG8=`           | Base64-decodes the value               |
| `hello`                  | Uses the raw UTF-8 bytes of the string |

**Base64 variants accepted**:

- standard padded base64
- raw standard base64 (no padding)
- URL-safe base64
- raw URL-safe base64

This is specifically for **request fields typed as `[]byte`**. Multipart file uploads still use the separate file-field logic, and binary responses use the `--output-file` / `--output-b64` flow documented above.

### Retries & Timeout

**Files**: `root.go.stmpl` (flag registration), `auxiliary/internal/client/client.go.stmpl` (retry config building), `readme/retries.stmpl` (generated docs), `includes/security.ts` (config struct generation), `pkg/generate/paths.go` (method policy)

Operations with `x-speakeasy-retries` in the OpenAPI spec get automatic retry behavior from the Go SDK. The CLI exposes flags to control this behavior globally. Since each CLI invocation runs exactly one operation, global SDK-level config (`sdk.WithRetryConfig()`, `sdk.WithTimeout()`) is equivalent to per-operation config.

```bash
# Disable all retries
cli users list --no-retries

# Bound the whole operation, including retry sleeps
cli users list --timeout 30s

# Cap total retry time
cli users list --retry-max-elapsed-time 5s

# Retry on connection errors (EOF, reset, etc.)
cli users list --retry-connection-errors

# Full retry config as JSON (escape hatch for exact control)
cli users list --retry-config '{"strategy":"backoff","backoff":{"initialInterval":500,"maxInterval":60000,"exponent":1.5,"maxElapsedTime":300000}}'
```

**Flags** (persistent on root command, all gated on `{{if isFeatureUsed "retries"}}` except `--timeout`):

| Flag                        | Type                 | Gate    | Description                                     |
| --------------------------- | -------------------- | ------- | ----------------------------------------------- |
| `--timeout`                 | string (Go duration) | always  | Whole-operation timeout (e.g., `30s`, `100ms`)  |
| `--no-retries`              | bool                 | retries | Disable all retries (overrides spec defaults)   |
| `--retry-max-elapsed-time`  | string (Go duration) | retries | Cap total retry time (e.g., `5s`, `1m`)         |
| `--retry-connection-errors` | bool                 | retries | Retry on connection errors (EOF, reset)         |
| `--retry-config`            | string (JSON)        | retries | Full retry config JSON escape hatch             |

**Precedence**: `--no-retries` > `--retry-config` JSON > individual flags > config file > spec defaults

**Visibility** (`cli.retryFlagsVisibility` gen.yaml key, default `visible`): controls the four retry flags (never `--timeout`):

- `visible` — registered and shown in help (default)
- `hidden` — registered but omitted from help; the flags still parse, and config-file/env behavior is unchanged. The `--usage` KDL keeps them, marked `hide=#true`
- `none` — not registered at all (parsing `--no-retries` fails with unknown flag); omitted from the `--usage` KDL. Config-file retry keys keep working (`buildRetryConfig` reads flags nil-safely)

**Method policy** (`cli.retryMethodPolicy` gen.yaml key, default `spec`):

- `spec` — root retry policies apply to every operation that inherits them
- `safe-methods` — root policies are inherited only by GET, HEAD, OPTIONS, and TRACE; explicit operation retry policies still win. Filtered operations have no generated retry loop, so runtime retry config cannot enable them

**Config file support** (`~/.config/<cli>/config.yaml`):

```yaml
timeout: 30s
no_retries: true
retry_max_elapsed_time: 5s
retry_connection_errors: true
retry_config: '{"strategy":"backoff","backoff":{"maxElapsedTime":5000}}'
```

**Key implementation details**:

- `resolveStringFlag(cmd, name)` in `client.go.stmpl` resolves flag > config precedence for string flags
- `buildRetryConfig(cmd)` returns `(*retry.Config, error)` — invalid JSON/duration values surface as errors, not silently swallowed
- Boolean flags use `flagutil.GetBoolFlag(cmd, name)` with `changed` semantics to distinguish `false` from unset
- Individual flags (e.g., `--retry-max-elapsed-time`) **replace** the entire retry config with CLI defaults (InitialInterval=500ms, MaxInterval=60s, Exponent=1.5, MaxElapsedTime=5m) and override the specific field. This is because the SDK takes a complete `retry.Config`, not a partial merge. Use `--retry-config` JSON for exact control over all parameters
- `Retry-After` and `retry-after-ms` override computed waits. Backoff rejects a server-directed wait beyond the remaining MaxElapsedTime budget; attempt-count retries ignore MaxElapsedTime and are bounded by MaxRetries or the operation timeout
- Config struct fields (`Timeout`, `NoRetries`, etc.) are generated by `templateConfigStructFields()` and `templateConfigGetStringCases()` in `security.ts`
- `--timeout` bounds a **whole streaming response**, not just the connect: the Go SDK method template (`templates/go/method.go.stmpl`) hands the timeout context's cancel to the returned `EventStream`/`JsonLStream` (`stream.WithCancel`/`jsonl.WithCancel`, released when the stream ends or is closed) instead of deferring it, so `--timeout 30s` streams to completion while a stream that stalls past the deadline still ends with the deadline error. Covered by `TestEventStreamWithTimeoutStreamsToCompletion` / `TestEventStreamTimeoutStillBoundsSlowStream` in `tests/primary/eventstreaming_test.go.stmpl`

**Hand-written tests**: `tests/primary/retries_test.go.stmpl` covers retry success, retry with body, Retry-After header, connect error, global config disable/success/timeout, request timeout, and connection error retry (reject + reset). `tests/tertiary/retry_method_policy_additional_test.go.stmpl` covers inherited safe/mutating methods and explicit operation precedence.

### Response Headers

**Files**: `root.go.stmpl` (flag registration), `auxiliary/internal/output/output.go.stmpl` (header extraction + injection), `includes/templating.ts` (envelope name sets)

The `--include-headers` flag adds HTTP response headers to the output. Without it, response headers are never shown (clean default). When enabled, headers appear in all output formats.

```bash
# Include response headers in JSON output
cli response-headers response-headers --status-code 200 --include-headers --output-format json

# Extract a specific header with jq
cli response-headers response-headers --status-code 200 --include-headers --jq '._response_headers."X-Required-Header"'

# Pretty format shows a "Response Headers:" section after the body
cli response-headers response-headers --status-code 200 --include-headers --output-format pretty
```

**Output contract**:

| Body type                      | JSON output shape                                                         |
| ------------------------------ | ------------------------------------------------------------------------- |
| Object `{...}`                 | `{"field": "value", "_response_headers": {...}}` (sibling key)            |
| Null / empty                   | `{"_response_headers": {...}}`                                            |
| Non-JSON raw body              | `{"_result_raw": "...", "_response_headers": {...}}`                      |
| Typed non-object (array, etc.) | `{"_result": [...], "_response_headers": {...}}`                          |
| Pretty format                  | Body first, then `Response Headers:` section with aligned key-value pairs |
| YAML format                    | `_response_headers:` as sibling key                                       |

**Header normalization**: Headers use Go's canonical HTTP casing (e.g., `X-Required-Header`). Single-value headers are flattened to strings; multi-value headers remain as arrays. JSON output keys are sorted by `json.Marshal`; pretty output sorts explicitly.

**How it works**:

1. `wantsHeaders(cmd)` checks the `--include-headers` flag
2. `extractResponseHeaders(res)` gets `http.Header` from the response's `HTTPMeta.Response.Header` via reflection
3. `extractErrorResponseHeaders(err)` does the same for error types (via `RawResponse` field)
4. `flattenHeaders(headers)` converts `http.Header` to `map[string]interface{}` with single-value flattening
5. `injectHeaders(data, headers)` merges flattened headers as `_response_headers` key into the output
6. `outputWithHeaders()` unifies header injection for json, yaml, and jq output modes
7. `printResponseHeadersPretty()` prints sorted, aligned headers for pretty mode

**Bug fix**: `isEnvelopeField()` now skips the `Headers map[string][]string` field on response structs. Previously, `extractResultContent()` would return the headers map instead of the body content because `Headers` came before `HTTPMeta` in struct field order. The same fix was applied to `envelopeNames` sets in `getPaginationFieldNames()` and `getStreamingFieldName()` in `templating.ts`.

**Error path**: `Error()` also supports `--include-headers` — it extracts headers from the error's `RawResponse` field and injects them into JSON error output or appends a `Response Headers:` section in pretty mode.

**Hand-written tests**: `tests/primary/headers_additional_test.go.stmpl` — 12 tests covering: 5 functional tests matching auto-generated test IDs, 5 cross-cutting format tests (json, yaml, pretty, jq, no-flag regression), 1 core bug regression test (body extracted over headers), and 1 `--header` request injection test.

### Configuration & Auth

**Config file**: `~/.config/<cli-name>/config.yaml`

**Priority chain** (two tiers depending on field type):

| Field type           | Resolution function           | Priority chain                                     |
| -------------------- | ----------------------------- | -------------------------------------------------- |
| Security credentials | `ResolveSecurityCredential()` | CLI flag > env var > **OS keychain** > config file |
| Global parameters    | `ResolveCredential()`         | CLI flag > env var > config file                   |

**Commands**:

- `configure` - Interactive prompt for credentials and global parameters. When the OS keychain is available (via `go-keyring`), security credentials are stored in the keychain instead of the config file. Falls back to config file on headless/CI environments
- `whoami` - Displays current credential values and global parameter settings with their sources (flag/env/keyring/config/unset)

The `configure` and `whoami` commands are generated whenever the API has global security schemes and/or global parameters (controlled by `hasConfigurableSettings()` in `security.ts`). When both are present, the configure command shows separate "Authentication" and "Global Parameters" sections.

**OS keychain integration** (`keyring.go.stmpl`): Uses [`go-keyring`](https://github.com/zalando/go-keyring) to store security credentials in the OS keychain (macOS Keychain, Linux Secret Service/kwallet, Windows Credential Manager). The keychain is probed at startup via a read-only `Get()` call — `ErrNotFound` means the backend is working. The result is cached for the process lifetime. A `KeyringBackend` interface allows test injection via `SetKeyringBackend()`.

When a secret is stored in the keychain it is **not** written to the config file, so the YAML `security` block stays empty for that field — `whoami` (which resolves keyring-first) is the way to confirm it is set. Two behaviors keep this from looking like the credential was lost:

- **Form placeholders are keychain-aware.** Secret fields in the `configure` and `auth login` forms render their placeholder from `config.GetStoredSecret(flag, configFallback)`, which prefers the keychain value over the config file. Re-running `configure` shows the existing (masked) credential instead of a blank field. The plain-text prompt fallback (`interactiveAuth: false`) uses the same read for its `[masked]` hint and for the required-field "keep existing value on empty Enter" check, so a keychain-stored secret is not re-prompted as if unset.
- **The "stored in OS keychain" message reflects reality.** Both `configure` and `auth login` track a `keychainStored` flag that is set only when a `SetKeyringValue` call actually succeeds during that invocation. The message prints only when at least one secret was written to the keychain in that run — not merely when the keychain is _available_. So if every keychain write fails and the values fall back to the config file, the message is suppressed. Because the flag reflects writes performed this run, re-running `configure` and pressing Enter to keep existing values (no new write) does not reprint the message.

**Global security** is generated as a shared function in the root command package. It reads all security schemes from the OpenAPI spec and maps them to flags/env vars/keychain/config keys.

**Cross-scheme ranking** (`templateRankedSecurityConstruction` in `security.ts`, `config.PickCredential` in `config.go.stmpl`): when the spec declares several OR alternatives, `buildGlobalSecurity` resolves every credential with its source, builds one candidate per alternative (per variant for union-typed alternatives), and populates exactly the one `config.PickCredential` picks: only alternatives the operation accepts (`client.NewClient(cmd, allowedFields...)`, mirroring the SDK's hoisted `PopulateSecurity(..., fields...)` restriction) in the operation's order; the most explicit source wins (flag > env > keyring > config), so `--api-key K` is not shadowed by an access token in the environment whatever the declared order; among equally explicit candidates a complete alternative beats a partial one; remaining ties keep declared order. Single-alternative CLIs keep the original construction.

**Global parameters** are also stored in the config file (with `global_` prefix on YAML keys to avoid collisions with security fields). The `buildGlobalOptions()` function in `client.go.stmpl` reads globals with priority: flag > env var > config file. Non-string types (integer, float, boolean) are stored as strings in the config and parsed when read.

**Interactive auth command group** (`auth.go.stmpl`): when `interactiveAuth` is enabled and the API has global security, the generator also emits:

- `auth login` — interactive credential setup focused on auth only
- `auth whoami` — auth-scoped view of current credential status
- `auth logout` — clears stored credentials from keyring and config

`configure` remains the broader entrypoint for auth + globals + preferences.

### Usage Schema & Grouped Help

**Files**: `usage.go.stmpl`, `includes/usage.ts`, `root.go.stmpl`

The generated CLI exposes two documentation-oriented interfaces beyond normal Cobra help:

1. **Grouped help output** — operation flags are split into Required vs Optional and can also be grouped by object/field category. In full style, persistent flags are organized into Output, Authentication, API Parameters, Server, Network, and Diagnostics; compact style prints those groups once through the root-only `--help-global` flag.
2. **Machine-readable usage schema** — `--usage` prints the selected command's usage schema in KDL format.

```bash
# Human-facing grouped help
cli users create --help

# Machine-readable command schema
cli users create --usage
```

**Help style resolution** (`cli.helpStyle` in `gen.yaml`): `auto` is the default and resolves to `compact` only when the document declares a non-empty `x-speakeasy-cli-commands` manifest; otherwise it resolves to `full`. Explicit `compact` and `full` override that decision. Full style compile-gates every compact runtime addition: it does not register `--help-global`, emit compact annotations or alter the previous Cobra help template, preserving generated output compatibility.

Compact command help keeps the Cobra order but renames the runtime examples heading: description → Usage → Aliases → `Just works:` → subcommands → optional `Defaults:` → local Flags → footer. The footer is assembled from the command's merged local and inherited flag set, in this order: `--json` or `--output-format json`, `--transform <dot.path>` or `--jq <expr>`, optional `--raw`, optional `--dry-run`, and optional `--usage`. It then points to `<cli> --help-global`. Exit codes are intentionally absent until the runtime implements that contract. Root help hides its local Flags block and ends with the same machine-interface line, the globals pointer, an auth Setup line only when a concrete primary credential environment variable exists, and Cobra's per-command help pointer.

`--help-global` is a root-local flag: `<cli> --help-global` prints the persistent Output, Authentication, API Parameters, Server, Network, and Diagnostics groups, while `<cli> subcommand --help-global` is an unknown flag. Its root hooks return before enum validation and config initialization, so global help still works with an unreadable config file. `includes/usage.ts` mirrors the flag as a root-local KDL node; it is not marked `global=#true`.

Declared commands may add a nested `help:` map with the closed keys `defaults`, `learn`, and `escalate`; each value must be a non-empty single line. Explicit `help.defaults` wins. Otherwise bound intents derive defaults from scalar presets (scalar arrays joined with commas) in authored order, followed by declared flags using `defaultFrom: schema`, without repeating a pointer already covered by a preset. Raw JSON keys identify presets and flag names identify flag-derived values. Bound intents derive escalation from their backing operation unless explicitly overridden. At runtime these values use the Cobra annotations `speakeasy_help_defaults`, `speakeasy_help_learn`, and `speakeasy_help_escalate`; any command, including a hand-written one, can set them. `speakeasy_help_footer: "false"` opts a command out of the generated footer.

Compact operation examples suppress only generator-synthesized invocations containing its own angle-bracket value placeholders or fallback `{"key": "value"}` body. Spec examples, manifest examples, and hand-written `Example` text are never filtered. An intent with no authored example gets a bare invocation only when it has no unresolved required positional or declared flag.

The compact layout and `Just works:` heading are runtime help contracts. Cobra markdown generation keeps its existing `Examples` heading and operation-doc footer, and the explorer continues to read `Short`, `Long`, and `Example` without a heading rewrite. Because synthesized-operation filtering happens while `templateCmdExample` builds the command's `Example`, a rejected synthesized example is absent from docs and the explorer too; authored examples remain unchanged everywhere. Custom commands that currently append Defaults/Machine/Globals prose to `Example` should migrate that prose to the annotations above to avoid a duplicate footer, set `speakeasy_help_footer: "false"` while retaining their own footer, or choose `helpStyle: full` during migration.

The KDL schema is built at generation time from the command tree in `includes/usage.ts` and emitted at runtime by `usage.EmitSchema()`. It includes command names, aliases, help text, flags, defaults, config metadata, and the config file location.

Declared commands (`x-speakeasy-cli-commands`) are part of the schema: bound intents mount into the KDL tree under their declared parent path with the same flag surface the runtime registers (declared positional `arg` nodes, operation security flags, the body flag, `--schema`, and declared flags — plus ordinary backing-operation flag metadata for pinned intents, or non-body parameter metadata only for route dispatch; `buildIntentUsageCommand` in `includes/usage.ts`), and planned placeholders appear with their summary. For commands added after static schema generation, live schema emission reads `speakeasy_prompt_args` and emits the same `arg` nodes. Once Cobra validation succeeds, `--usage` on an intent or planned command emits KDL and wins over the command's other runtime surfaces, including `--schema` and the bare-invocation usage error (`runIntent*Cmd` in `intentcmd.go.stmpl` and `newPlannedCmd` in `intents.go.stmpl` check `usage.UsageRequested` first, matching operation commands).

**`--usage` is one root-level contract**: `NewRootCommand` creates cobra's default `help` and `completion` commands (`InitDefaultHelpCmd`/`InitDefaultCompletionCmd`, which cobra would otherwise add lazily inside `Execute`), applies declared command order, calls `interactive.Intercept(rootCmd)` when interactive mode is enabled, and then calls `usage.Intercept(rootCmd)` as the outer/last interceptor. Both see every generated, intent, catalog, auth and hand-written (`custom.Register` and root custom region) command. The usage interceptor walks the tree and (a) wraps each command's `Run`/`RunE` — non-runnable groups such as `auth` get a `RunE` that falls back to help — so `--usage` emits the command's KDL schema and nothing runs, and (b) folds every lifecycle hook (`PersistentPreRun(E)`, `PreRun(E)`, `PostRun(E)`, `PersistentPostRun(E)`) into a no-op under `--usage`, so a custom pre-run cannot mutate state either. Cobra's validation still surrounds those hooks: unknown flags, `--help`, non-deferred `Args` validation (`op bogus --usage` is a `NoArgs` error), and Cobra's post-PreRun required-flag/flag-group validators can win. `output.InstallErrorHandling` skips its eager typed copy of the latter two checks under `--usage`, but does not suppress Cobra's originals; such failures retain Cobra's existing untyped/runtime classification. The interactive Args wrapper explicitly does not defer under `--usage`, and the interactive PreRunE wrapper also checks the usage/doc-surface gates. Per-command `usage.UsageRequested` checks remain (harmless) as the documented precedence. Commands absent from the static KDL table use live Cobra schema emission, including declared positional `arg` nodes. The KDL root tree carries the built-in `auth` group (`login`/`whoami`/`logout`, unless a generated command already owns `auth` by name or alias — cobra dispatches to the first registered), `explore`, catalog commands (`x-speakeasy-cli-catalog`; a catalog whose exact name is a generated group lends the group its summary, one that only matches an alias is unreachable and skipped), and cobra's `help` and `completion` (`bash`/`zsh`/`fish`/`powershell`). `cmd/gendocs` skips `help`/`completion` (utility surface, same filter as the explorer) so the docs set is unchanged.

### Generated README & Command Docs

**Files**: `includes/includes.ts` (section registration), `includes/readme-commands.ts`, `includes/readme-examples.ts`, `readme/*.stmpl`, `readme_examples_test.go.stmpl`, `auxiliary/cmd/gendocs/main.go.stmpl`, `root.go.stmpl` (`groupedUsageTemplate`)

The generated README is assembled from registered sections (`registerReadmeSection` in `includes/includes.ts`; block IDs are stable so existing READMEs update in place). Sections that matter for the command surface and for agents:

- **CLI Example Usage** (`usage`) — leads with a `### Quick start` block of the first three declared intents' first examples (with their `# summary`), then the spec's usage examples (`readme/usage-container.stmpl`).
- **For AI agents** (`agents`, weight 6, `readme/agents.stmpl`) — the discovery ladder `--help` → `--usage` (KDL) → `--schema` → keyless `--dry-run`; human previews use stderr with empty stdout, while `--output-format json` or caller-explicit `--jq` emits one compact preview object per request as NDJSON on stdout (jq is not applied). It also documents the agent-mode error envelope and capability-gated streaming/artifact examples.
- **Commands** (`operations`, `templateCLICommandsSection` in `includes/readme-commands.ts`) — mirrors root `--help`: with an `x-speakeasy-cli-commands` manifest it renders the declared categories in manifest order (bound intents as bullets with an indented ```` ```bash ```` block of their manifest examples; planned commands as `— _not in this build_: <note>`; group tags and planned names owned by a real command render the existing group / built-in under the category), then everything untagged under `### Additional commands`. Without a manifest it renders the plain operation tree in a `<details>` block as before.
- **Overrides** — an exact-path route-dispatch override removes the generated operation entry before the command tree and example pickers run. The intent therefore appears once, in manifest order, and the removed operation cannot be selected for showcase, body, pagination, streaming, or upload examples.
- **Request Body Input / Output Formats / Pagination / Streaming / Retries / Diagnostics / Server Selection / Authentication / Configuration / File Uploads / Error Handling** — no `<command>` placeholders: every example is a real invocation picked from the generated tree by `readmeExampleContext()` (`includes/readme-examples.ts`). Pickers are deterministic (score, then command-path tie-break) and capability-gated: the *showcase* command is a read-only operation with no required inputs (list operations first); *pagination* is the first paginated operation; *streaming* is a declared intent that streams by construction (preset `stream: true` or a declared `--stream` flag) else the first SSE/JSONL operation; *body* is a JSON-body operation with a spec example that has no `<placeholder>` values and no credential-looking keys, preferring operations with expandable field flags (bools render as `--flag=value`); *upload* is the first multipart operation with a file field (its block is preceded by `<!-- readme-examples: skip (needs a local file) -->`). When a picker has no candidate the section drops the example (or the whole illustration) rather than inventing one; the showcase falls back to the first intent example, then the first operation rendered with example values, then `<cli> version`.
- Global-parameter examples use the parameter's default → example → first enum value → `value`, and the showcase command path (never a bare operation ID, which is not invocable inside a group). Security flag examples reference the credential's env var (`--api-key "$ACME_API_KEY"`, first three fields) instead of `<value>`.

`templateCLIUsageCommand` (`includes/tests.ts`) unescapes the `\x7b\x7b` / `\x7d\x7d` sequences `goStringLiteral` emits (Go-template protection for `{{`/`}}`) when rendering shell text, so `}}` inside a JSON example never leaks into README/USAGE output.

**README examples are a contract** (`readme_examples_test.go.stmpl` → `tests/readme_examples_test.go`, generated when tests and documentation are enabled): the test parses every fenced `bash`/`sh`/`shell` block of `README.md`, and every line that starts with the CLI name (after an optional `KEY=value` env prefix or `echo '<json>' |` stdin feed; anything after an unquoted pipe/redirect is dropped) runs in-process through the harness with `--dry-run` appended, asserting exit 0. `configure`, `auth`, `explore`, `completion`, `help` lines and blocks preceded by `<!-- readme-examples: skip ... -->` are not run. `--dry-run` therefore has to succeed keyless on every documented command: `DryRunClient` (`auxiliary/internal/client/diagnostics.go.stmpl`) shapes its synthetic 200 from the request's `Accept` header (JSON `{}` when JSON or `*/*` is accepted, otherwise the first accepted media type with an empty body) so binary/text-only operations dry-run cleanly, and `opcmd.go.stmpl` skips the binary-to-TTY guard under dry-run.

**Command docs** (`cmd/gendocs`) append a `### Machine interface` section to every page whose command carries the `speakeasy_operation` cobra annotation: `<cmd> --usage`, `<cmd> --schema` (only when registered), `<cmd> --dry-run`, and the JSON/jq output note. Every generated command page then ends with `clierrors.HelpFooter`; groups, built-ins, and planned placeholders receive only that exit-code footer. Root `--help` alone carries the one-line "Machine interface:" summary, while root and every per-command help page end with the same exit-code footer (`groupedUsageTemplate` in `root.go.stmpl`). This generated-doc contract is independent of compact runtime help: compact annotations and `Just works:` do not rewrite Cobra markdown pages or the explorer.

### Server Selection

**Files**: `config.ts` (generator option), `includes/dependencies.ts` (resolved visibility), `root.go.stmpl` (flag registration), `includes/usage.ts` (KDL usage), `client.go.stmpl` (global resolution), `opcmd.go.stmpl` (operation-level resolution), `includes/templating.ts` (template functions), `readme/server.stmpl` (generated documentation)

The CLI supports server selection via `--server` (when enabled) and `--server-url` flags, plus server template variable flags (`--hostname`, `--port`, `--protocol`, etc.):

```bash
# Override server URL directly
cli users list --server-url http://localhost:8080

# Select server by index (array-based SDKs)
cli users list --server 0

# Select server by name (map-based SDKs)
cli users list --server "production"

# Server with template variables
cli users list --server "TEMPLATED" --hostname api.example.com --port 443
```

**Precedence**: `--server-url` > `--server` > default

**Visibility** (`cli.serverSelectionFlag` gen.yaml key, default `visible`): controls only `--server`; `--server-url` and dedicated server variable flags are unaffected:

- `visible` (the default) — registered and shown in help; existing `--server` invocations keep working across regeneration
- `auto` — resolves to `visible` when the document offers a choice (more than one global server, or more than one server on any operation); otherwise resolves to `hidden`, so the flag stays out of help but keeps parsing
- `hidden` — registered but omitted from help and the generated README; the flag still parses, and the `--usage` KDL keeps it marked `hide=#true`
- `none` — not registered (parsing `--server` fails with unknown flag) and omitted from the `--usage` KDL

A single named or indexed server is not a choice, even when it has template variables. `auto` therefore hides a flag with no alternative value while keeping it parseable, so scripts that passed `--server 0` to a single-server CLI keep working; only an explicit `none` removes the flag.

**Two-layer validation**:

1. **Global (`client.go.stmpl`)**: `NewClient` reads `--server` and silently applies it if it matches a global server. Unrecognized values are silently skipped — they may be intended for operation-level servers.
2. **Operation-level (`opcmd.go.stmpl`)**: Operations WITH their own server list validate `--server` against that list. Operations WITHOUT their own servers validate against the global server list. Invalid values produce clear errors.

**Array-based vs map-based**: Generated SDKs store servers as either `[]string` (indexed by number) or `map[string]string` (keyed by name), depending on the `ServerMap` AST property. Template functions in `templating.ts` (`sdkServerMap()`, `templateOperationServerResolution()`, `templateGlobalServerValidation()`) generate the appropriate resolution code for each type.

**Server template variables**: Some servers have URL templates like `http://{hostname}:{port}`. The template variables are registered as persistent flags on the root command and passed as `sdk.With*()` options at the global level, or as `operations.WithTemplatedServerURL(url, params)` at the operation level.

**Hand-written tests**: Server tests are hand-written (not auto-generated from Arazzo) in `tests/<variant>/servers_test.go.stmpl`. They use `RunWithServer()` for success cases (echo server) and `RunBare()` for failure cases (broken servers, invalid indices).

### Custom HTTP Headers

The CLI supports adding custom HTTP headers to any request via the `--header` flag:

```bash
# Add a single custom header
cli users list --header "X-Request-ID: abc-123"

# Add multiple custom headers
cli users list --header "X-Request-ID: abc-123" --header "X-Trace-ID: xyz"
```

**Implementation**: The `--header` flag is registered as a `StringArray` persistent flag on the root command (`root.go.stmpl`). Each operation's run function (`opcmd.go.stmpl`) reads the flag and converts headers to `operations.WithSetHeaders(headerMap)`. Headers coexist with other flags — global headers, security headers, and custom `--header` values are all sent together.

### JQ Transforms (`transformJq`)

Operations with SDK-level jq transforms (`x-speakeasy-transform-to-api` / `x-speakeasy-transform-from-api`) are supported. The transforms are applied at the Go SDK model level in `MarshalJSON()` / `UnmarshalJSON()` via `utils.RunJQBytes()`.

**Key behavior**: When an operation has a response transform, `hasResponseTransform(op)` returns `true` and:

1. `WithSkipDeserialization()` is NOT added — the SDK must deserialize the response to apply the transform
2. `output.DrainRawBody(res)` is called after the SDK call — this consumes the raw HTTP body so `output.Result()` falls through to the typed (transformed) output path instead of raw JSON passthrough

This means the CLI outputs the transformed response, not the raw API response.

### Diagnostics

**Files**: `root.go.stmpl` (flag registration), `auxiliary/internal/client/diagnostics.go.stmpl` (HTTP wrappers + redaction), `opcmd.go.stmpl` (skip-deser guard)

Two global flags provide request/response diagnostics without modifying any operation command:

```bash
# Preview request without making a network call
cli users list --dry-run

# Log request/response diagnostics while running normally
cli users list --debug
```

**How it works**:

1. **Root flags**: `--dry-run` and `--debug` are registered as persistent `Bool` flags on the root command (`root.go.stmpl`)
2. **HTTP client wrappers** (`diagnostics.go.stmpl`): Two wrapper types implement the SDK `HTTPClient` interface:
   - `DryRunClient` — emits a deterministic redacted request preview without calling the inner client: human `[DRY-RUN]` blocks on stderr, or compact NDJSON on stdout for resolved JSON / caller-explicit `--jq`; it then returns a synthetic `200` response
   - `DebugClient` — logs redacted request to stderr, delegates to inner client, logs redacted response to stderr, restores the response body stream
3. **Wrapper composition** (`client.go.stmpl`): `WrapClientForDiagnostics(cmd, inner)` composes the appropriate wrapper based on flag state. Applied after test client injection so both layers compose cleanly
4. **Skip deserialization** (`opcmd.go.stmpl`): When `--dry-run` is active, `WithSkipDeserialization()` is appended to SDK call options. This prevents deserialization panics on the synthetic empty response body. The guard is placed before all three execution branches (pagination, streaming, normal)

**Precedence**: `--dry-run` wins if both flags are set. `IsDebug(cmd)` returns `false` when dry-run is also active.

**Redaction and keyless resolution**:

- **Headers**: Case-insensitive denylist (`authorization`, `cookie`, `set-cookie`, `x-api-key`, `x-session-token`, `*-secret`, `*-token` suffixes) → `[REDACTED]`
- **URL query**: Credential parameters (`api_key`, `token`, `client_secret`, `signature`, etc.) are redacted without reordering the query
- **JSON body**: Separator/case-normalized sensitive keys and canonical base64 strings (encoded length ≥128, excluding all-hex IDs) are redacted; numbers retain their exact lexemes and depth overflow becomes an explicit marker
- **Other bodies**: Binary media becomes `<bytes:N>` and multipart payloads become a deterministic inventory; human dry-run is not truncated. Debug applies the same redactor, then caps the rendered preview at 4 KiB
- **Credentials**: Generated request construction uses request-scoped resolvers. Dry-run reads flags, environment and config, but never probes the OS keychain; whoami/configuration use the introspection resolvers and retain keychain access

**Dry-run branch handling**: Pagination (`--all`) and streaming branches are skipped in dry-run mode since the synthetic response has no real paginated/streaming data. Control falls through to the normal SDK call path with skip-deser.

**Generated tests**: `dryrun_response_test.go.stmpl` pins response shaping plus redaction, number preservation, non-HTML escaping, determinism and the zero-value human client; `dryrun_matrix_test.go.stmpl` selects a keyless input-free operation from the document and checks human, JSON and caller-jq protocols (including zero keyring reads); artifact and fixture-specific tests cover filesystem and command-kind behavior.

### Test Framework

**Test harness**: `cli_harness_test.go.stmpl` provides `CLITestHarness` which executes CLI commands in-process:

```go
h := NewCLITestHarness(t)
h.WithTestName("CreateUser")
err := h.Run([]string{"create-user", "--name", "John"})
// or with stdin:
err := h.RunWithStdin([]string{"create-user"}, `{"name": "John"}`)

stdout := h.GetStdout()
name := jsonGetString(t, stdout, "name")
assert.Equal(t, "John", name)
```

Tests are generated from **Arazzo workflows** (the OpenAPI test format). Each workflow step becomes assertions in a test function. Multi-step tests carry output variables between steps.

**Test helpers** handle setup/teardown operations (creating prerequisite resources, cleaning up after tests).

**Test skip detection**: `isTestSkipped()` in `features.ts` does exact-match filtering to skip individual tests for unsupported or inapplicable features (streaming, retries, headers, webhooks, etc.). `getCLITestSkipMessage()` in `tests.ts` also walks operation and nested workflow steps; if one uses an operation whose generated command registration was replaced by an intent override, the generated test is skipped rather than replaying removed generated flags against a semantically different command. Dedicated intent tests own that replacement surface.

**Harness methods**:

- `Run(args)` — injects `--server-url` pointing to the mock server (standard usage)
- `RunWithServer(url, args)` — injects a custom `--server-url` (e.g., echo server)
- `RunBare(args)` — no `--server-url` injection (allows `--server` flag to take effect, or tests broken servers)
- `RunWithStdin(args, stdin)` — runs with stdin JSON input
- `RunRaw(args)` — runs expecting raw (non-JSON) response

**Current test status**: All 11 variants passing. Run `TARGET=<variant> make test-cli` to see current counts. Test counts grow as new Arazzo tests and hand-written tests are added.

---

## Where to Change What

| I want to change...                                                 | Edit this file                                                                                                                                                                                                                                                                                                                      |
| ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| How a command is structured (init/run functions)                    | `opcmd.go.stmpl`                                                                                                                                                                                                                                                                                                                    |
| Declarative intents, route dispatch, and overrides                  | `includes/intents.ts`, `intentcmd.go.stmpl`, `auxiliary/internal/flagutil/dispatch.go.stmpl`, `root.go.stmpl`, `subroot.go.stmpl`, `includes/usage.ts`, `includes/readme-commands.ts`                                                                                                                                               |
| How flag metadata is generated from the AST                         | `includes/metadata.ts`                                                                                                                                                                                                                                                                                                              |
| Field expansion decisions (nested flags vs JSON)                    | `includes/templating.ts`                                                                                                                                                                                                                                                                                                            |
| How flags are registered or requests built at runtime               | `auxiliary/internal/flagutil/metadata.go.stmpl`                                                                                                                                                                                                                                                                                     |
| Command/group Short/Long descriptions                               | `includes/descriptions.ts`                                                                                                                                                                                                                                                                                                          |
| Union type inspection / expansion decisions                         | `includes/unions.ts`                                                                                                                                                                                                                                                                                                                |
| Output formatting (JSON, YAML, pretty, jq, toon)                    | `auxiliary/internal/output/output.go.stmpl`, `pretty.go.stmpl`, `jq.go.stmpl`, `agentmode.go.stmpl`                                                                                                                                                                                                                                 |
| Error taxonomy, extension data, hints, and rendering                | `internal/extensions/cli_errors.go`, `includes/errors.ts`, `auxiliary/internal/output/classify.go.stmpl`, `errortable.go.stmpl`, `output.go.stmpl`                                                                                                                                                                                  |
| Interactive prompting for missing flags                             | `opcmd.go.stmpl`, `auxiliary/internal/interactive/interactive.go.stmpl`                                                                                                                                                                                                                                                             |
| Explorer TUI / auto-launch on empty invocation                      | `root.go.stmpl`, `auxiliary/internal/explorer/`                                                                                                                                                                                                                                                                                     |
| Agent mode behavior and structured errors                           | `root.go.stmpl`, `configure.go.stmpl`, `auth.go.stmpl`, `auxiliary/internal/output/agentmode.go.stmpl`, `classify.go.stmpl`, `output.go.stmpl`                                                                                                                                                                                       |
| Pagination (`--all` flag, streaming output)                         | `opcmd.go.stmpl` (branching), `output/output.go.stmpl` (`PaginatedResult`), `includes/templating.ts` (`hasPagination`, field name helpers)                                                                                                                                                                                          |
| Streaming (SSE/JSONL output)                                        | `opcmd.go.stmpl` (branching), `output/output.go.stmpl` (`StreamResult`), `includes/templating.ts` (`hasStreamingResponse`, `getStreamingFieldName`)                                                                                                                                                                                 |
| Binary downloads (`--output-file`, `--output-b64`, TTY enforcement) | `opcmd.go.stmpl` (branching), `output/binary.go.stmpl` (`WriteToFile`, `WriteBase64`, `IsBinaryTTYBlocked`), `output/output.go.stmpl` (binary check in `Result`), `includes/templating.ts` (`hasBinaryResponse`, `isOnlyBinaryResponse`)                                                                                            |
| Bytes/base64 request flag parsing                                   | `includes/flags.ts`, `auxiliary/internal/flagutil/metadata.go.stmpl` (`FlagKindBytes`, `buildBytesField`)                                                                                                                                                                                                                           |
| Auth / security flag generation                                     | `includes/security.ts`                                                                                                                                                                                                                                                                                                              |
| Config file behavior (priority chain, persistence)                  | `auxiliary/internal/config/config.go.stmpl`                                                                                                                                                                                                                                                                                         |
| OS keychain integration (credential storage)                        | `auxiliary/internal/config/keyring.go.stmpl`                                                                                                                                                                                                                                                                                        |
| Credential resolution (flag > env > keyring > config)               | `auxiliary/internal/config/config.go.stmpl` (`ResolveSecurityCredential`, `ResolveCredential`)                                                                                                                                                                                                                                      |
| SDK client construction                                             | `auxiliary/internal/client/client.go.stmpl`                                                                                                                                                                                                                                                                                         |
| Root command, global flags, subgroups                               | `root.go.stmpl`, `subroot.go.stmpl`                                                                                                                                                                                                                                                                                                 |
| Server selection (`--server` flag, template variables)              | `config.ts` (generator option), `includes/dependencies.ts` (visibility resolver), `root.go.stmpl` (flag registration), `includes/usage.ts` (KDL gating), `client.go.stmpl` (global resolution), `opcmd.go.stmpl` (operation-level resolution), `includes/templating.ts` (template functions), `readme/server.stmpl` (README gating) |
| Test CLI arg formatting / skip detection                            | `includes/tests.ts`                                                                                                                                                                                                                                                                                                                 |
| Test assertions and type coercion                                   | `includes/test-assertions.ts`                                                                                                                                                                                                                                                                                                       |
| Multi-step tests and workflow calls                                 | `includes/test-workflows.ts`                                                                                                                                                                                                                                                                                                        |
| Test templates (Go output)                                          | `test.go.stmpl`, `testfile.go.stmpl`                                                                                                                                                                                                                                                                                                |
| Test harness (in-process executor)                                  | `cli_harness_test.go.stmpl`                                                                                                                                                                                                                                                                                                         |
| Flag naming / reserved words                                        | `includes/flags.ts`                                                                                                                                                                                                                                                                                                                 |
| Go naming conventions (commands, functions)                         | `includes/sanitization.ts`                                                                                                                                                                                                                                                                                                          |
| String escaping for Go literals                                     | `includes/sanitization.ts` (`escapeGoString`)                                                                                                                                                                                                                                                                                       |
| Dependencies (go.mod)                                               | `includes/dependencies.ts`, `config.ts`                                                                                                                                                                                                                                                                                             |
| Build / lint / test commands                                        | `config.ts`                                                                                                                                                                                                                                                                                                                         |
| `configure` / `whoami` commands                                     | `configure.go.stmpl`, `whoami.go.stmpl`                                                                                                                                                                                                                                                                                             |
| `auth login` / `auth whoami` / `auth logout` commands               | `auth.go.stmpl`, `includes/security.ts`                                                                                                                                                                                                                                                                                             |
| Usage schema / `--usage` output                                     | `usage.go.stmpl`, `includes/usage.ts`, `root.go.stmpl`                                                                                                                                                                                                                                                                              |
| Custom HTTP headers (`--header` flag)                               | `root.go.stmpl` (flag registration), `opcmd.go.stmpl` (header injection)                                                                                                                                                                                                                                                            |
| JQ transforms (`hasResponseTransform`)                              | `includes/templating.ts` (detection), `opcmd.go.stmpl` (skip deser + drain)                                                                                                                                                                                                                                                         |
| Retry/timeout flags and config                                      | `root.go.stmpl` (flag registration), `client/client.go.stmpl` (`buildRetryConfig`, `resolveStringFlag`), `includes/security.ts` (config struct fields)                                                                                                                                                                              |
| Diagnostics (`--dry-run`, `--debug`)                                | `root.go.stmpl` (flag registration), `client/diagnostics.go.stmpl` (HTTP wrappers + redaction), `opcmd.go.stmpl` (skip-deser guard), `client/client.go.stmpl` (wrapper composition)                                                                                                                                                 |

---

## Relationship with Go SDK

The CLI generator depends on the Go SDK generator (`templates/templates/go/`). Key interactions:

- **SDK generation**: `main.ts` calls `interoptTemplateTarget("go", ...)` to generate the Go SDK as a library dependency, with docs/tests/modules disabled. The embedded SDK is generated from the Go target's *default* config (computed as a new SDK), not the user's `cli` gen.yaml section — options that must respect the user's config have to be forwarded explicitly in `getJobs()` (e.g. `PackageName`, `SDKVersion`, `IdiomaticMethodCollisionNames`)
- **`SDKVersion` and `WrapperName`**: `cli.version` is forwarded as the embedded SDK's `SDKVersion`, so `config.SDKConfiguration.SDKVersion` and the default `User-Agent` carry the CLI version. `WrapperName` is pinned to `"speakeasy-sdk/go"` (the Go default UA prefix) because a non-empty wrapper name stops the Go templates' `getRootModulePath()` from appending `/v<major>` — the CLI's `getGolangPackage()` already adds that suffix to `PackageName`, and without the pin a CLI at major version ≥ 2 would import `…/v2/internal/sdk/v2`
- **`idiomaticMethodCollisionNames`** (default: on for new SDKs, off for existing): declared in the CLI config and forwarded to the embedded Go SDK config so error struct fields that collide with generated methods (e.g. a field named `error` vs `Error()`) are renamed idiomatically (`ErrorInfo`) instead of underscore-suffixed (`Error_`)
- **`maxMethodParams: 0`**: The CLI config sets this so all SDK methods take a single request struct (containing both params and body), which simplifies the metadata-driven approach
- **Response format**: CLI uses `responseFormat: "envelope-http"` instead of the Go SDK's default `"flat"`. This wraps responses in a struct with `HTTPMeta components.HTTPMetadata` containing `Response *http.Response`, needed for raw JSON passthrough
- **`WithSkipDeserialization()`**: A Go SDK option (gated on `enableSkipDeserialization` config flag) that leaves the HTTP response body unread. CLI uses this for raw JSON output
- **Test generation**: CLI overrides `getResponseContentVariablePath()` to NOT add the content field prefix (e.g., `.User`), since CLI output strips the envelope via `extractResultContent()`. Without this, test assertions would reference `"User.metadata.allergies"` instead of `"metadata.allergies"`
- **`errorUnions` feature**: Do NOT add this feature to the CLI config. It causes compilation errors due to type mismatches between `components.Error` and `sdkerrors.Error`. The `GetUnionErrors` test stays skipped
- **`transformJq` feature**: CLI supports SDK-level jq transforms. The Go SDK generates `RunJQBytes()` in `utils/jq.go`, and models apply transforms in their `MarshalJSON()` / `UnmarshalJSON()`. The CLI template's `hasResponseTransform(op)` detects operations with response transforms and skips `WithSkipDeserialization()` so the SDK can deserialize and apply the transform
- **`sliceUnions` / `openEnums` features**: These are AST-level feature gates in `internal/schemas/schemas.go`. Without `sliceUnions`, array-typed union members get flattened to single structs. Without `openEnums`, enums get strict `UnmarshalJSON` validators that reject unknown values (breaking union variant resolution). Both are required for correct union behavior

---

## Release Distribution Configuration

Release/distribution generation is controlled in `config.ts` using two layers:

- `generateRelease` (default: `true`)
- nested `distribution` channel settings:
  - `distribution.homebrew.*`
  - `distribution.winget.*`
  - `distribution.nfpm.*`

When `generateRelease` is disabled, the CLI generator does not emit:

- `.goreleaser.yaml`
- `.github/workflows/release.yaml`
- install scripts in `scripts/`

### Distribution schema

```yaml
cli:
  generateRelease: true
  distribution:
    homebrew:
      enabled: false
      tap: "owner/homebrew-repo"
    winget:
      enabled: false
      publisher: ""
      publisherUrl: ""
      repositoryOwner: ""
      packageIdentifier: ""
      license: ""
    nfpm:
      enabled: false
      formats: "deb,rpm"
      maintainer: ""
      license: ""
```

### Validation rules

`validateConfig()` in `config.ts` enforces:

- channel `enabled: true` requires `generateRelease: true`
- Homebrew:
  - `tap` required
  - `tap` format must be `owner/repo`
  - repo segment must start with `homebrew-`
- WinGet:
  - requires `publisher`, `publisherUrl`, `repositoryOwner`, `packageIdentifier`, `license`
  - `publisherUrl` must be absolute `http`/`https`
  - `repositoryOwner` must match GitHub owner-name pattern
  - `packageIdentifier` must be dot-separated segments (e.g. `Example.Petstore`)
- nFPM:
  - requires `maintainer`, `license`, `formats`
  - `formats` supports only: `deb`, `rpm`, `apk`

### Rendering behavior

Release helpers in `includes/release.ts` drive channel-specific output:

- `.goreleaser.yaml`: optional `brews:`, `winget:`, `nfpms:` sections
- release workflow env: conditional channel tokens
  - `HOMEBREW_TAP_GITHUB_TOKEN`
  - `WINGET_GITHUB_TOKEN`
- README installation snippets:
  - Homebrew install command (`brew install owner/tap/cli` with `homebrew-` prefix stripped)
  - WinGet install command (`winget install <packageIdentifier>`)
  - Linux package install snippets for enabled nFPM formats

---

## Forward Compatibility

The CLI is **always forward-compatible** with API evolution. Unlike SDKs (where users might want strict validation), a CLI is a pure consumer of API responses and should never break when the API adds new enum values or union discriminator variants.

This is enforced via `getConfigOverlay()` in `config.ts` — no gen.yaml settings are exposed:

| Setting                            | Value                   | Effect                                                                                                                                                                                                             |
| ---------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `forwardCompatibleEnumsByDefault`  | `true`                  | Response enums generate `IsExact()` instead of strict `UnmarshalJSON` validation. Unknown values are preserved, not rejected.                                                                                      |
| `forwardCompatibleUnionsByDefault` | `"tagged-and-untagged"` | Both discriminated and untagged response unions add an `UnknownRaw json.RawMessage` fallback field. Unknown discriminator values are stored as raw JSON instead of causing errors.                                 |
| `unionStrategy`                    | `"populated-fields"`    | Required companion to open enums. Uses `PickBestUnionCandidate()` which scores candidates by matched/inexact/unmatched fields, instead of left-to-right which fails when unknown enum values produce 0 candidates. |

**What this means in practice:**

- API adds `status: "archived"` to an enum → CLI outputs it without error
- API adds `"spaceship"` vehicle type to a discriminated union → CLI outputs the raw JSON payload, no crash
- Request-side enum flags still validate against known `EnumValues` — only response enums are open
- Raw JSON passthrough (`--output-format json`) bypasses typed deserialization entirely and was already forward-compatible
- `resolveConfig()` normalizes boolean `true`/`false` values to the string equivalents expected by `MarkUnionsOpen()` and `MarkOpenEnums()`

---

## Body Input Methods

The CLI supports three ways to provide request body data, with a clear priority chain:

```bash
# 1. Individual flags (highest priority — override everything)
cli create-user --name "Jane" --age 30

# 2. --body flag (whole-body JSON, overridden by individual flags)
cli create-user --body '{"name": "John", "age": 30}' --name "Jane"
# Result: {name: "Jane", age: 30}

# 3. Stdin (lowest priority body input)
echo '{"name": "John", "age": 30}' | cli create-user --name "Jane"
# Result: {name: "Jane", age: 30}
```

**Priority**: Individual flags > `--body` flag > stdin

The `--body` flag is registered on all metadata-driven operations that have a body (excluding multipart). It accepts the entire request body as a JSON string, providing a flag-based alternative to stdin piping. This is useful when stdin is unavailable (e.g., in scripts that already use stdin for other purposes).

**How it works in `BuildRequest[T]`**:

1. If `--body` flag was set, unmarshal its JSON into the `bodyFieldPath` sub-field of the request struct
2. Else if stdin has data, unmarshal it into the `bodyFieldPath` sub-field
3. For each flag in the metadata, check `cmd.Flags().Changed(flagName)` — only set fields for explicitly-provided flags (overrides both `--body` and stdin)
4. `relaxRequiredForBodyFields()` suppresses required-field errors for body fields when `--body` or stdin populated them, and clears `HasDefault` on those body fields so unchanged flag defaults cannot overwrite body/stdin values. Path/query/header params with schema defaults are still applied when their flags are unchanged.

**Edge cases**:

- Empty/whitespace-only stdin is treated as no stdin
- Stdin with union fields works via `json.Unmarshal` into the union's custom `UnmarshalJSON`
- Params-only operations (`bodyFieldPath == ""`): `--body` is not registered, stdin has no effect
- Non-expandable `IsRequestBody` types (single `--request-body` JSON flag): no `--body` flag, no merge, stdin is a fallback if the flag isn't set
- Multipart operations: no `--body` flag (file uploads can't be expressed as JSON)
- Body fields with schema defaults: when `--body`/stdin is used and the dedicated flag is not set, the body/stdin value wins. The flag default is only applied when building from flags alone (no body/stdin), or when the flag is explicitly passed.

**Request-shape guards** (`verifyBodyKeys` in `flagutil/metadata.go.stmpl`) — a whole-body JSON document is never silently reshaped:

- Key loss: after `--body`, stdin, a whole-body `IsRequestBody` flag, or a union JSON flag is decoded, the request is marshaled back with the SDK serializer and every top-level key the caller wrote must still be present. A dropped key is an error naming the key, with a did-you-mean against the request type's JSON keys (walking union members). This is presence-only at the top level — open (`additionalProperties`) schemas keep their keys, nested keys are not inspected, keys set to `null` are exempt (an absent optional field and a null one serialize identically), and form-only body types (form tags, no JSON tags) are not verified — so API additivity is unaffected and no `DisallowUnknownFields` is involved.
- Selector ambiguity: a whole-body non-discriminated union whose body names two selector keys (`UnionMeta.VariantKeys`, e.g. `model` and `agent`) errors naming both instead of sending whichever variant the decoder picked.
- `--body` and the metadata flag that carries the whole body field (e.g. `--body-param`, or `--request-body` for a non-expanded body) both set: error; neither silently wins. That whole-body metadata flag has `--body`'s precedence over stdin: when it is set, stdin is not read.
- Required union fields: after `BuildRequest` or `BuildRequestBody` has merged every source, a reflection preflight walks the request body through structs, pointers, slices, arrays, and maps. A union value (identified by `union:"member"` fields) with no selected member and no non-empty `union:"unknown"` raw value fails as `*flagutil.MissingRequiredFieldError` / `CLI_VALIDATION`, naming its JSON-tag path (`input`, `nested.input`, `items[0].input`) and its generated union flag when metadata provides one. A missing root union instead says that the request body selects no variant and names the whole-body and union flags. Builder errors still win — notably, selecting an expanded discriminated variant while omitting one of that variant's required flags retains the precise `missing required flag` error. Nil pointer unions remain optional; selected members are traversed for nested required unions; and fields omitted by the SDK's `omitempty`/`omitzero` rules are skipped. Metadata-backed form/multipart bodies with no `--body` flag are included only for fields tagged `,json`, because those are the fields their serializer passes through union JSON marshaling. This runs before dry-run or live SDK invocation, so serializer text is not parsed as a fallback. Intent commands are covered because they synthesize or merge the body flag and then call the backing operation's same request builder.

**Declared intent commands** (`intentcmd.go.stmpl`): a whole request supplied up front — `--body`, the whole-body metadata flag (`BodyParamFlag`, computed from the operation's flag metadata by `wholeBodyFlagName` rather than assumed to be `--body-param`), or piped stdin — takes request control and the declared inputs are not synthesized into a competing body. Stdin is read once with `flagutil.AttachStdinBody` (same mode-aware `ReadStdinBody` contract: nothing read on a TTY, EOF is no body, agent mode bounded by the stdin deadline) and re-attached as the command input so `BuildRequest` reads it exactly like the operation path; piped JSON therefore runs the request instead of landing on the help page, and interactive prompting (TTY only) is untouched. A positional argument or declared flag given alongside a supplied body is merged by `flagutil.MergeInputIntoBody` (typed: bool/int/float/string): if the body lacks the bound key the input is injected, if it already sets it the invocation errors naming both sources.

**`@` forms of the body flag** (`flagutil.ResolveBodyFlagValue`, applied to `--body` and single `--request-body`-style flags in `BuildRequest`/`BuildRequestBody`): `--body @path` reads the file at `path`; `--body @-` reads stdin to EOF with a blocking read (no probe window); `--body @@...` is the literal escape (the value minus one `@`) for a raw body whose first byte must be `@`. Any other value is used verbatim. A supplied intent `--body` flows through `BuildRequest` unchanged, so the forms apply to intent commands too; a synthesized intent body never starts with `@`. A failed local read is explicitly `CLI_VALIDATION` and names both the path and flag; it is never inferred to be a connection problem from the wrapped OS error.

**Implicit stdin body** (`ReadStdinBody`, re-attached by `AttachStdinBody`): a TTY is never read and a regular file is read fully. A pipe or FIFO is read through EOF with mode-dependent bounding. Outside agent mode the read blocks with no deadline — the classic `producer | cli op` filter contract, however slow the producer. In agent mode (the root command arms `flagutil.SetStdinReadDeadline` from `output.IsAgentMode()` after flag parsing; output format alone never changes pipe semantics) the complete body must reach EOF within `stdinReadDeadline` (1 s); a pipe that has not delivered EOF by then fails with a typed `*flagutil.StdinTimeoutError` (`CLI_VALIDATION`, exit 2) naming the command's own body flag and pointing at `--<bodyflag> @-` — never a partial body, and never a silently substituted default (agent runtimes and exec wrappers routinely hand the CLI an open-but-silent stdin; they get a fast, classified error instead of a hang or a wrong request). A closed empty pipe is "no body": required bodies fail with the missing-flag error, optional bodies proceed. `--body @-` always reads to EOF with no deadline in any mode.

---

## Required Validation Semantics

Different flag kinds use different "is this flag set?" checks because zero values can be valid:

| Kind                 | Required check                 | Rationale                                                                    |
| -------------------- | ------------------------------ | ---------------------------------------------------------------------------- |
| String               | `cmd.Flags().Changed()`        | `--flag=` is present; the schema decides whether `""` is valid (`minLength`) |
| JSON, DateTime, Date | `val == ""`                    | Empty string = unset (never a valid value)                                   |
| Bool                 | `cmd.Flags().Changed()`        | `false` is a valid explicit value                                            |
| Int64, Float64       | `cmd.Flags().Changed()`        | `0` is a valid explicit value                                                |
| Enum                 | `cmd.Flags().Changed()`        | `--flag=` is present; valid only if the enum declares `""`                   |
| IntEnum              | `val == ""`                    | Empty = unset                                                                |
| StringArray          | `len(val) == 0`                | Empty array = unset                                                          |
| Union                | Delegated to variant detection | At least one variant flag must be set                                        |
| File                 | `val == ""`                    | Empty path = unset                                                           |

The `Changed()` pattern is critical: Cobra tracks whether a flag was explicitly passed on the command line, regardless of its value. Without this, `--verbose=false` or `--count=0` would be indistinguishable from "flag not provided."

### Argument & Value Hygiene

**Files**: `opcmd.go.stmpl`, `subroot.go.stmpl`, `intentcmd.go.stmpl`, `root.go.stmpl`, `auxiliary/internal/flagutil/flags.go.stmpl`, `auxiliary/internal/flagutil/metadata.go.stmpl`, `includes/metadata.ts`

- **No stray positionals**: operation commands and generated group commands set `Args: cobra.NoArgs`. Operations take flags only, so a token that is not a flag is an error (`unknown command "x" for "cli group op"`) raised by Cobra's `ValidateArgs` before `PersistentPreRun` and before any request is built. This closes the `cli agent delete --id x -- --dry-run` hole, where `--dry-run` after `--` used to become an ignored positional while the live request went out. It also turns `cli group typo` into an error instead of the group help page. Intent commands with a declared positional stay `ArbitraryArgs`.
- **Spaced boolean values**: pflag parses `--flag false` as `--flag` (true) plus a positional `false`. On flags-only commands that is now rejected by `NoArgs`; on a variadic intent command it cannot be rejected structurally, so `flagutil.SpacedBoolValueHint` prints a stderr hint when the trailing positional is `true`/`false` and a boolean flag was set (`boolean flags take their value inline: --dry-run=false`; every changed boolean flag is named when there are several, since argv adjacency is not available). Generated tests (`includes/tests.ts` `pushFlagArg`, `includes/test-workflows.ts`) and command examples (`includes/descriptions.ts` `exampleFlagPart`) always render boolean values inline (`--flag=true`).
- **CLI-owned enums**: the root `PersistentPreRunE` validates `--output-format` (against `output.Formats`) and `--color` (`auto`, `always`, `never`) with `flagutil.ValidateEnumFlag` before any command runs (after the `--usage` short-circuit — documenting a command never validates its rendering flags); a typo errors with the option list and a did-you-mean suggestion (prefix or edit distance ≤ 2). Only the flag value is checked — config/env values are resolved separately in `resolveOutputFormat`.
- **Schema-declared bounds** (`FlagMeta.MinLength`, `HasMinimum`/`Minimum`, `HasMaximum`/`Maximum`, emitted by `includes/metadata.ts` from `TypeDef.Validations`): an explicitly set string shorter than `minLength` (including `--code=` when `minLength ≥ 1`) and an explicitly set number outside `minimum`/`maximum` are rejected before the request is sent. Undeclared bounds stay with the server (it is authoritative; client-side bounds would drift), and an unconstrained string still accepts `""` (some APIs clear fields with it).
- **Enum flags**: an explicitly set empty value (`--model=`) is rejected unless the enum declares `""` — it used to be sent as `"model": ""`. Unknown values keep listing the valid options.
- **Required strings/enums**: presence is `Changed()` (`validateRequiredPresence`), so `--tag=` on a required unconstrained string is present and sends `""`; only an omitted flag is "missing required flag". Non-finite floats (`NaN`, `Inf`, accepted by the flag parser) are rejected before bounds checks.

---

## Known Limitations

### Skipped Tests

Tests are skipped at two levels:

1. **Feature-level** (`isTestSkipped()` in `features.ts`) — skips individual tests by exact-match ID for unsupported features (polling, hooks, retries, webhooks, etc.).

2. **Operation-level** (`getCLITestSkipMessage()` in `tests.ts`) — placeholder for future per-operation skip checks. Currently returns `""` (no skips). All previously-skipped operation-level issues have been resolved.

#### Skip Summary

| Variant             | Notes                                                                                                           |
| ------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Primary**         | All skips resolved, includes hand-written server/union/collection/globals/hooks/pagination/streaming tests      |
| **Secondary**       | Includes hand-written server tests (by-name) + auth tests                                                       |
| **Tertiary**        | Includes hand-written server tests (by-id)                                                                      |
| **Quaternary**      | Includes hand-written server tests (by-name-with-templates) + auth + transform tests                            |
| **OAuth2-Password** | Includes hand-written hook tests (with-credentials, with-token, bad-credentials, not-required, operation-scope) |
| **Review**          | 3 skipped: TestEndpoint (broken Arazzo), Chat (SSE), GetUnionErrors (errorUnions feature)                       |

#### Review-Only Skips (3)

| Test           | Reason                                                                                                                 | Status                 |
| -------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------- |
| TestEndpoint   | Missing required request body & path param in Arazzo spec                                                              | Shared gap with Go SDK |
| Chat           | SSE operation in review Arazzo spec has test data issues                                                               | Shared gap with Go SDK |
| GetUnionErrors | `errorUnions` feature causes CLI compilation errors (type mismatches between `components.Error` and `sdkerrors.Error`) | CLI-specific           |

#### Previously Skipped (Now Resolved)

These categories were skipped in primary and have been fixed:

| Category                                      | Fix                                                                                                                                                                                                                                                                        | Tests Unblocked |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------- |
| Nullable fields with null examples            | `filterOmitemptyNulls()` accounts for `omitempty` difference between mock server types and CLI SDK types                                                                                                                                                                   | 8               |
| Map body with date-time/bigint/decimal values | Fixed `v.CanSet()` guard and map pointer dereference in Go SDK `json.go.stmpl`                                                                                                                                                                                             | 4               |
| bigint/decimal array key casing               | Fixed spec (`bigintArray` → `bigIntArray` in `components.yaml`)                                                                                                                                                                                                            | 1               |
| Deep object query params                      | Added `deepObjectParams: "0.0.0"` feature flag to `features.ts`                                                                                                                                                                                                            | 1               |
| Union of arrays / deeply nested arrays        | Added `sliceUnions: "0.0.0"` feature — needed for array-typed union members to be treated as complex types in AST                                                                                                                                                          | 2               |
| Smart union open enums (variant resolution)   | Added `openEnums: "0.0.0"` feature — without it, strict enum validators rejected unknown values and broke union resolution                                                                                                                                                 | 1               |
| Smart union nested union                      | Hand-written test using `RunBare()` to let operation hit its own test server                                                                                                                                                                                               | 1               |
| OAuth2 password bad credentials               | Hand-written test — the real token endpoint was always available, skip comment was outdated                                                                                                                                                                                | 1               |
| Collections containing null                   | Hand-written test — null values pass through CLI JSON pipeline correctly                                                                                                                                                                                                   | 1               |
| Globals header keeps custom client headers    | `--header` flag now supports custom HTTP headers alongside global params                                                                                                                                                                                                   | 1               |
| Pagination                                    | Implemented `--all` flag with streaming NDJSON output + `--max-pages` guardrail. Un-skipped 20 pagination tests, added 6 hand-written `--all` tests                                                                                                                        | 26              |
| SSE streaming                                 | Implemented `StreamResult()` with reflection-based stream iteration. Un-skipped 12 SSE tests, added 2 cross-cutting tests (jq, yaml)                                                                                                                                       | 14              |
| JSONL/NDJSON streaming                        | Added `jsonlResponses` feature flag, reuses `StreamResult()`. Un-skipped 5 JSONL tests, added 1 cross-cutting test (jq)                                                                                                                                                    | 6               |
| Retries & timeout                             | Implemented `--no-retries`, `--timeout`, `--retry-max-elapsed-time`, `--retry-connection-errors`, `--retry-config` flags with config file support. Un-skipped 7 retry tests, added 7 hand-written tests                                                                    | 7               |
| Response headers                              | Implemented `--include-headers` flag with `_response_headers` injection. Fixed `isEnvelopeField()` bug where `Headers` field was returned instead of body content. Un-skipped 5 header tests, added 12 hand-written tests (functional + cross-cutting format + regression) | 17              |

#### How Skip Detection Works

`isTestSkipped(test)` in `features.ts` skips by exact match. The skip list is organized into three sections:

1. **Go SDK also skips** — tests skipped in the Go SDK (pagination edge cases, flattening edge cases, etc.)
2. **Not applicable to CLI** — SDK-internal behavior tests (request recorder, hook scope constants, operation-level OAuth2, etc.)
3. **Applicable but unsupported/deferred** — features the CLI should eventually support

### Unsupported / Deferred Features

- **Response headers (flat format)**: CLI always uses `envelope-http` response format. The 2 tests using `responseFormat: flat` remain skipped (`headers-*-flat`)
- **OAuth device flow**: Auth is config-file based only (no browser-based login)
- **Compact intent operation flags**: operation-backed intents still list every registered backing-operation flag, including flags not declared by the manifest. Collapsing those undeclared flags behind an operation-help pointer is deferred because it needs template-only filtering that must not hide them from Cobra docs or KDL.

### Known Bugs

- **`Example.Clone()`** does not copy `Reference` or `Replacements` fields (latent bug in the AST layer, not CLI-specific)

---

## Evolution & Design Context

The CLI generator was built iteratively. Understanding this history explains why certain things are the way they are:

| Phase   | What changed                                                                                    | Why it matters now                                                                                   |
| ------- | ----------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| 1 (MVP) | One file per operation, imperative flag parsing, 100-260 lines per command                      | Explains why older patterns may still exist in edge cases                                            |
| 1.5     | Discriminated union dot-notation (`--shape.circle.radius`)                                      | Typed union flag support avoids falling back to untyped request maps                                 |
| 2       | Auth extraction, `configure`/`whoami`, flag > env > config chain                                | Security is a shared function, not per-command                                                       |
| 3       | Test framework: in-process harness, Arazzo workflows, multi-step tests                          | Tests are generated, not hand-written                                                                |
| 4       | README generation, Cobra docs via `cmd/gendocs`                                                 | Docs are auto-generated from templates                                                               |
| 5-6     | Metadata-driven `FlagMeta` + `BuildRequest[T]` with reflection                                  | The architectural core — reduced per-command code 60-80%                                             |
| 7       | Review feedback: distinct pretty output, `ValidateMeta`, `cmd.OutOrStdout()`                    | Pretty != indented JSON; validation catches generator bugs at startup                                |
| 8       | Stdin merge (was mutual exclusion), `relaxRequiredForBodyFields()`                              | Stdin provides base body, flags override individual fields                                           |
| 9       | Raw JSON passthrough via `envelope-http` + `WithSkipDeserialization()`                          | Avoids lossy deserialize-reserialize; fixed config overlay bug                                       |
| 10      | `--header` flag, `transformJq`, `sliceUnions`, `openEnums` feature parity                       | Custom headers, jq transforms, array-typed unions, lenient enums                                     |
| 11      | Pagination: `--all` streaming, `--max-pages`, NDJSON output, README section                     | Leverages SDK `Next()` closures via reflection; streams results without buffering                    |
| 12      | SSE & JSONL streaming: `StreamResult()`, `hasStreamingResponse()`, hand-written tests           | Reuses pagination's `outputOneItem()` pattern; reflection-based stream iteration                     |
| 13      | Retries & timeout: `--no-retries`, `--timeout`, `--retry-config`, config file support           | Global retry/timeout control via `client.go.stmpl`; `buildRetryConfig()` with flag>config precedence |
| 14      | OS keychain: `go-keyring`, 4-tier credential resolution, `configure` keychain storage           | Security creds in keychain (flag > env > keyring > config); graceful fallback on headless            |
| 15      | `--body` flag, fixed examples, reserved `"body"` name                                           | Whole-body JSON via flag (not just stdin); examples now show correct flags                           |
| 16      | Diagnostics: `--dry-run` and `--debug` global flags with HTTP client wrappers                   | Request preview / debug logging without per-command changes; redaction for secrets                   |
| 17      | Binary downloads: `--output-file`, TTY enforcement, binary-before-JSON in `Result()`            | Binary responses stream directly; `--output-format json` silently ignored for binary                 |
| 18      | Interactive mode: explicit/default prompting, explorer TUI, `--interactive`, `--no-interactive` | CLI can guide humans through commands without changing the underlying flag-driven contract           |
| 19      | Agent mode: optional environment detection, TOON-by-default, structured errors                  | AI agents get machine-friendly output and are protected from TUI-only workflows                      |
| 20      | Base64 support: `FlagKindBytes` request input + `--output-b64` for binary responses             | Both request-side byte payloads and response-side binary payloads have script/agent-friendly paths   |
| 21      | Usage schema + grouped help                                                                     | Generated CLIs now expose both human-friendly and machine-readable command documentation             |
| 22      | Interactive auth command group (`auth login/whoami/logout`)                                     | Auth-only workflows no longer need to go through the broader `configure` command                     |

---

## Design Decisions

### Why metadata-driven instead of imperative per-field parsing?

Early versions generated bespoke flag-reading code per operation (read each flag, assign to struct field). This produced 100-260 lines per command. The metadata approach reduces this to ~30-55 lines and centralizes the complex logic (stdin merge, required validation, pointer wrapping, union handling) in one runtime function.

### Why reflection instead of raw body bypass?

An alternative is to bypass typed Go structs entirely and build HTTP requests from `map[string]any`. This would require SDK changes (`WithRawBody()` option) and loses type safety. Our approach keeps the typed SDK contract while using reflection only in the CLI layer.

### Why `envelope-http` response format?

The CLI needs access to the raw `*http.Response` for raw JSON passthrough and error status codes. The `envelope-http` response format wraps every response in a struct with `HTTPMeta` containing the response object. Combined with `WithSkipDeserialization()`, this enables reading the HTTP body directly without deserializing into typed structs.

### Why `ValidateMeta` at startup?

Generator bugs (mismatched field paths, wrong types) would otherwise surface as cryptic reflection errors when a user runs a specific command. `ValidateMeta` catches these at CLI startup by validating every `FieldPath` resolves on the request type, providing clear panic messages that point to the broken metadata.

### Why in-process test execution?

Running the CLI as a subprocess in tests would require building the binary, managing processes, and parsing text output. In-process execution via `CLITestHarness` is faster, captures structured output, and allows direct assertion on command errors.

### Why have both interactive mode and agent mode?

They optimize for different users. Interactive mode improves the human terminal experience with prompts and a command explorer, while agent mode disables those surfaces and favors machine-readable output and structured errors. Keeping both modes in the generator lets a single CLI serve humans, scripts, and coding agents without forking the product.

---

## Common Patterns & Pitfalls

### `IsRequestBody` vs `RequestBody`

`IsRequestBody` is `false` for mixed param+body operations (e.g., `updateUser` with an `id` path param and a `User` body). Always check if `RequestBody` exists (truthy), not `IsRequestBody`:

```typescript
// Wrong: misses mixed param+body operations
if (op.IsRequestBody && op.RequestBody) { ... }

// Correct: catches all operations with a body
if (op.RequestBody) { ... }
```

### `sanitizeType()` has side effects

`sanitizeType()` triggers `addImport` for the type's package. If you only need the type name for a description string, use `typeDef.Name` directly to avoid unnecessary imports.

### Go enum types aren't plain strings in TypeScript

When comparing `op.SerializationMethod` or similar Go-origin enum values, call `.toString()`:

```typescript
// Wrong: compares object reference
if (op.SerializationMethod === "json") { ... }

// Correct: compares string value
if (op.SerializationMethod?.toString() === "json") { ... }
```

### `Example.ToJSON()` returns a JSON string

`ToJSON()` returns a JSON-encoded *string*. To get the actual typed value, parse it:

```typescript
const value = JSON.parse(example.ToJSON());
```

### Variant `FieldPath` is relative

Union variant `FlagMeta.FieldPath` values are relative to the variant struct, not the root request:

```go
// For ShapeRequest.Shape.Circle.Radius:
// Union FlagMeta.FieldPath = "ShapeRequest.Shape"
// Variant FlagMeta.FieldPath = "Radius"  (NOT "ShapeRequest.Shape.Circle.Radius")
```

### `sanitizeCommandName` already appends `Cmd`

Don't add a `Cmd` suffix when using `sanitizeCommandName()` — it already does this.

### Config overlay must use accessor

`AddOverlay()` stores the overlay separately from `Languages[target].Cfg`. Direct reads of `Cfg["key"]` bypass the overlay. Always use `GetLanguageConfigValue("key")` which merges both:

```typescript
// Wrong: bypasses overlay
const val = config.Languages[target].Cfg["responseFormat"];

// Correct: checks overlay first
const val = config.GetLanguageConfigValue("responseFormat");
```

---

## Adding Features

### Adding a new flag type

1. Add a `FlagKind` constant in `metadata.go.stmpl`
2. Add registration logic in `RegisterFlags()` (how to create the Cobra flag)
3. Add building logic in `BuildRequest()` (how to read the flag and set the struct field)
4. Add metadata generation in `metadata.ts::buildMetaEntryForField()` (how to emit the Go literal)
5. Run `TARGET=review make test-cli` to verify

### Adding a new output format

1. Add the format to the `--output-format` enum in `root.go.stmpl`
2. Add formatting logic in `output/output.go.stmpl::Result()`
3. Add streaming/pagination handling in `output/outputitems.go.stmpl::outputOneItem()`
4. Add the library to `config.ts::getTemplateDependencies()` and `includes/dependencies.ts::getDefaultDependencies()` (if it requires a new dependency)
5. Add the format row to the tables in `readme/output.stmpl` and `readme/pagination.stmpl`
6. Add a hand-written test in `tests/primary/` and verify with `TARGET=primary make test-cli`

### Adding a new global flag

1. Add flag registration in `root.go.stmpl`
2. If it affects SDK calls, update `client/client.go.stmpl` to read it
3. If it affects output, update `output/output.go.stmpl`
4. If it changes machine-readable docs, update `includes/usage.ts`
5. Update this README so it remains the source of truth for generated CLI behavior

### Running tests

```bash
# Generate + compile + test against review spec (164 tests)
TARGET=review make test-cli

# Generate + compile + test against primary spec (531 tests)
TARGET=primary make test-cli

# Run all 11 CLI variants
for t in primary secondary tertiary quaternary client-credentials client-credentials-basic oauth2-password custom-http no-servers relative-servers review; do
  TARGET=$t make test-cli
done

# Generate only (for inspection)
go run cmd/generate/main.go -s /tmp/spec.yaml -o /tmp/cli-output -l cli
```

### Debugging generation

The generated output lands in `zSDKs/sdk-cli/` (review) or `testSDKs/sdk-cli-primary/` (primary). Inspect the generated Go files there to verify metadata arrays, flag names, and request building calls match expectations.

### Keeping this README current

This file should track the actual generated CLI behavior, not just the original design. When adding or changing features in `templates/templates/cli/`, update this README in the same change if you touch:

- user-facing flags or commands
- output modes or error behavior
- interactive / agent-mode behavior
- config/env/keyring precedence
- request input or binary/base64 handling
- generated documentation surfaces like `--help` or `--usage`

Treat this README as the maintainer-facing source of truth used to update external documentation later.
