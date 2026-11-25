# csharpv2 template

`csharpv2` is template version `v2` of the `csharp` target. It generates C# SDKs that serialize JSON with
System.Text.Json source generation and are trimming- and NativeAOT-safe on net8.0 and later. It is not a full
template: it is an overlay on the `csharp` (v1) template and only contains the files that differ.

This README is for contributors. Part A explains how a template inherits from a base template and how `csharpv2`
uses that. Part B covers AOT rules for template changes and what changes for an SDK that moves from v1 to v2.

- [Part A — Template inheritance](#part-a--template-inheritance)
  - [Base templates](#base-templates)
  - [How csharpv2 builds on csharp](#how-csharpv2-builds-on-csharp)
  - [Shared templates and `isCsharpV2`](#shared-templates-and-iscsharpv2)
  - [Tests](#tests)
  - [What is emitted](#what-is-emitted)
  - [Checks](#checks)
- [Part B — AOT](#part-b--aot)
  - [B1. Rules for template changes](#b1-rules-for-template-changes)
  - [B2. Migrating an SDK from v1](#b2-migrating-an-sdk-from-v1)

---

## Part A — Template inheritance

### Base templates

A template directory `templates/templates/<template>/` can declare a `template.yaml`:

```yaml
baseTemplate: csharp           # overlay on another template
```

```yaml
additionalSearchPaths:         # extra lookup locations only
  - go
```

Without `template.yaml` the template is standalone. A `baseTemplate` or search path that does not exist fails
template loading. The loader and the resolution rules live in
[`internal/templatedir/templatedir.go`](../../../internal/templatedir/templatedir.go) (tests in
`templatedir_test.go` next to it).

With `baseTemplate`, the template (the primary) is a sparse overlay: a file at the same relative path as a base file
replaces it, and a file the primary does not have falls back to the base. Each lookup applies this rule as follows:

| Lookup | Resolution |
| --- | --- |
| Template engine: `templateFile`, `templateString`, `require`, the `main.ts` entry script | first match in: primary, base, each `additionalSearchPaths` entry, `templates/common`, then the path as given |
| `getDirectoryFiles(dir)` and `templateDirectory(dir, out)` | the files of `dir` in the base and in the primary, merged; on the same relative path the primary file wins |
| `directoryExists(dir)` | true if `dir` exists in the primary or in the base |
| `copy(src, out)` | `src` from the primary if it exists there, else from the base |
| Executor scripts (`config.ts`, `features.ts`, `exclusions.ts`) | the primary file if present, else the base file (one file, never merged) |
| Changelogs (`changelogs/`) | the template's own changelog folders, then those of its base template |

`require` in template scripts first resolves the path against the locations above. Only if nothing matches does it
resolve the path relative to the requiring script, using the path that script was required by. This has four
consequences:

- A root-relative `require` (`includes/includes.ts`, `hooks/hooks.ts`, `features.ts`) picks the primary file when one
  exists at that path.
- A relative `require` inside a base file resolves to the base directory. `csharp/includes/includes.ts` is loaded as
  `../csharp/includes/includes.ts`, so its `require("./models.ts")` loads `csharp/includes/models.ts` even when
  `csharpv2/includes/models.ts` exists. A same-name file next to the base file does not override it.
- Because the root locations are tried first, a top-level file named like a relative include is loaded in its place
  (`require("./exclusions.ts")` in `csharp/includes/includes.ts` loads the top-level `csharp/exclusions.ts`). Do not
  name a top-level `.ts` file of the primary after a base include. `csharpv2/config.ts` is not affected: the base
  chain requires `../config.ts`, which resolves to `csharp/config.ts`.
- A script reached through a relative `require` is known by the path as written (`./v2/models.ts`). A relative
  `require` inside it resolves against `v2/`, not `includes/v2/`, and fails: require every include from
  `includes/includes.ts`.

All template scripts share one JavaScript scope. A function declared again by a script that runs later replaces the
earlier declaration, also for callers in base files. Declarations are hoisted: a function declared in the script
that requires the base chain exists before the base chain runs and is then replaced by the base declaration, so
overrides go in a script required after the base chain. Go template functions are different: `registerTemplateFunc`
stores the function value it is given and fails if the name is taken. A script that changes a template function calls
`unregisterTemplateFunc(name)` and then `registerTemplateFunc(name, fn)`.

Test and other overlay directories are inherited through the merged directory walks: a target that templates
`tests/Common` with `templateDirectory` gets the base files plus the primary's, the primary winning on the same path.
A template whose output renders empty (or whitespace only) is not written, which is how a base file can be left out
of the primary's output.

`additionalSearchPaths` (the `cli` target lists `go`) is different. It only adds locations to the template engine
lookup, so `cli` can `require` helpers and read templates from `go` by path. `getDirectoryFiles`, `directoryExists`,
`copy` and the executor ignore it: no `.stmpl`, copied asset or test directory of `go` reaches `cli` output, and `cli`
owns every file it generates.

### How csharpv2 builds on csharp

Selection. The target is still `csharp` and its `gen.yaml` key is still `csharp`. `templateVersion: v2` in that
block selects the `csharpv2` template; `-l csharpv2` on `cmd/generate` selects it directly. Templates test the
version with `isCsharpV2()` (JavaScript) or `isCsharpV2` (Go templates), registered in
`../csharp/includes/templateVersion.ts`. It reads `templateVersion` from `gen.yaml`: set `templateVersion: v2` in
`gen.yaml` also when generating with `-l csharpv2`, otherwise shared templates render their v1 branches. Every
csharpv2 test config sets it.

`template.yaml` contains `baseTemplate: csharp`.

Entry point and include chain. `csharpv2` has no `main.ts`: `csharp/main.ts` runs, and its
`require("includes/includes.ts")` resolves to [`includes/includes.ts`](includes/includes.ts) here. That file first
requires the whole base chain (`../csharp/includes/includes.ts`), then `includes/auxiliary.ts` and the
`includes/v2/*.ts` files. `csharp/main.ts` adds the v2-only jobs (`getCsharpV2Jobs`, `includes/v2/main.ts`) when
`isCsharpV2()` is true.

Overrides live in `includes/v2/`, one file per concern, and replace only what differs:

- a JavaScript function is redefined under the same name, for example `getSDKDependencies` in
  `includes/v2/dependencies.ts` and `templateAuxiliaryFiles` in `includes/auxiliary.ts`;
- a Go template function is re-registered with `unregisterTemplateFunc` + `registerTemplateFunc`, for example
  `getModelsLocation` in `includes/v2/namespaces.ts`.

Do not add a file with the name of a base include to `includes/`: the base chain keeps loading the base file (see
above), and requiring the new file in its place would mean reimplementing every helper of the base file. Besides
`includes/includes.ts`, which replaces the base file through the root-relative `require` in `csharp/main.ts`, the
only same-path file, `includes/auxiliary.ts`, is loaded in addition to `csharp/includes/auxiliary.ts` (explicitly, by
`includes/includes.ts`, after the base chain) and wins by redefining `templateAuxiliaryFiles`.

`.stmpl` templates override by path: `enum.cs.stmpl`, `union.cs.stmpl`, `pagination.cs.stmpl` and the other
top-level templates that also exist in `csharp` replace the base files; `union-*.cs.stmpl` and `json-*.cs.stmpl` exist
only here; every template not present here comes from `csharp`.

`config.ts` replaces `csharp/config.ts` as a whole, because the executor loads one entry script. It requires the base
file as a module (`require("../csharp/config.ts")`) and declares every entry function again (`getConfigFields`,
`resolveConfig`, `getTemplateDependencies`, ...), delegating to the base where v2 behaves the same. Template scripts
do not see this file: the include chain loads `csharp/config.ts`, so a function templates call
(`getTemplateDependencies`, for example) is the base one during templating. Code both sides need goes into an include
that both require, as `includes/v2/netstandard2.ts` does.

Changelogs: there is no `changelogs/csharpv2/`; csharpv2 uses the entries in `changelogs/csharp/`.

Auxiliary files. `includes/auxiliary.ts` redefines `templateAuxiliaryFiles`. It renders every file under
`auxiliary/v2/` (the path is rendered as a template; an empty path is skipped), then renders a fixed list of base
files that are serializer-agnostic: SDK `Utils/` files (`Constants`, `Retries/*`, the default HTTP client and, when
server-sent events are used, `Sse/SseEvent`, `Sse/SseParseException`, `Sse/SseStreamParser`), `global.json`, the
base and default error classes and `ResponseValidationException`, `OptionalNullable`, `IOpenUnion` (open unions only)
and `NotNullIfNotNullAttribute` (netstandard2.0 mode, when a union has nullable implicit conversions). The base
`auxiliary/` tree is not walked, so a base auxiliary file reaches v2 output only through that list. To share one more
base file, add a `templateFile(...)` call there and gate any differing lines in the base file with `isCsharpV2`.

### Shared templates and `isCsharpV2`

Files under `../csharp/` render for both versions. When v2 needs different output:

- Gate in the shared file when the difference is small: a namespace, one expression, an attribute, a few lines of an
  otherwise shared function (`{{if isCsharpV2}}…{{else}}…{{end}}` in `.stmpl`, `if (isCsharpV2())` in `.ts`).
- Copy the file into `csharpv2/` when most of it differs (another serializer, another type hierarchy), so neither
  version reads as two programs interleaved.
- The v1 branch renders exactly what v1 rendered before: shipped v1 SDKs must not change.
- A shared `.stmpl` can only name template functions that are registered for both versions. A function registered in
  `includes/v2/` cannot appear in a shared template, not even inside an `isCsharpV2` branch: the template fails to
  parse for v1.
- `{{- if …}}` trims the newline before it. At the start of a file use `{{- if not isCsharpV2 -}}`.
- A file that renders empty is not written: wrapping a whole shared file in a gate removes it from one version's
  output.

### Tests

`csharp/main.ts` templates `tests/Common` and `tests/<test group>` into `Tests/`. Both walks merge `../csharp/tests/`
with `tests/` here, so every shared test also runs for v2. The rules for shared, v1-only and v2-only tests, and for
the `Compat` members that absorb v1/v2 differences, are in [`../csharp/tests/README.md`](../csharp/tests/README.md).
The v2 `Compat` implementation is `tests/Common/Compat.cs.stmpl`.

### What is emitted

Only the walked directories are copied into generated output: `tests/Common/` and `tests/<test group>/` (into
`Tests/`), `testproject/` (into the usage test project, when one is generated), `testusage/` (into
`testusages/<template>/<test group>/`, in test-group builds) and `auxiliary/v2/` (into the SDK root, path rendered).
Every file in them is emitted, including `.md` files
(`AnalyzerReleases.*.md` under `auxiliary/v2/` is emitted on purpose). Files at the top level of `csharpv2/` and in
`includes/`, such as this README, are not.

### Checks

- A change under `../csharp/` must leave v1 output unchanged: run `TARGET=review make test-csharp`, then
  `git status --porcelain zSDKs/sdk-csharp` must be empty. Also run `make check-template-csharp`.
- `make check-template-csharpv2` type-checks the TypeScript templates; `npm run format` formats them.
- `TARGET=<variant> make test-csharpv2` generates and tests a variant; `TARGET=<variant> make test-csharpv2-aot`
  also publishes the usage test project with NativeAOT.
- `make build-csharpv2-netstandard2` builds variants with `targetFramework: netstandard2.0` for every target
  framework.
- After adding, renaming or removing a file under `templates/`, run `go generate ./...` (updates
  `templates/perms.go`).

---

## Part B — AOT

### B1. Rules for template changes

- No reflection-based serialization. Every JSON read or write goes through a `JsonTypeInfo` from the SDK's
  source-generated `JsonSerializerContext`s, chained in `Utils/Json/JsonOptions.cs` (`json-options.cs.stmpl`). A type
  that is serialized at run time must be declared with `[JsonSerializable]` in one of those contexts;
  [`includes/v2/jsoncontext.ts`](includes/v2/jsoncontext.ts) collects them (models, operation request and response
  shapes, collection wrappers of union variants, JSON form and multipart parts). A missing entry compiles and then
  throws `NotSupportedException` at run time.
- A converter named in `[JsonConverter]` or in `JsonSourceGenerationOptions.Converters` must be template output, that
  is ordinary source of the SDK project. The System.Text.Json source generator does not see the output of other
  source generators: a converter emitted by a Roslyn generator is reported as SYSLIB1220, the attribute is dropped and
  serialization fails at run time. Enum converters, enum extensions and the date, decimal and big-integer converters
  are therefore templates.
- The `<Package>.Generators` project (`auxiliary/v2/{{getGeneratorsTopLevelFolder}}/`) holds the union,
  serialization and security source generators. The SDK project references it as an analyzer
  (`ReferenceOutputAssembly="false"`): it runs at build time only and is not part of the package.
- On net8.0 and later the SDK project sets `IsAotCompatible`, `IsTrimmable` and the trim and AOT analyzers, and
  `TreatWarningsAsErrors` exempts only CS0618 and NU1901–NU1904. Any IL2xxx / IL3xxx warning fails the SDK build.
- Target frameworks. `dotnetVersion` is `net8.0` or `net10.0`. By default the SDK targets `dotnetVersion` alone and
  the templates emit no `#if`. `targetFramework: net8.0` with `dotnetVersion: net10.0` targets `net8.0` with the
  `LangVersion` of `dotnetVersion` (14.0), as in v1. With `targetFramework: netstandard2.0` it targets
  `netstandard2.0;<dotnetVersion>` with `LangVersion` 10.0. Code that must differ per target framework is wrapped in
  `#if NET8_0_OR_GREATER` … `#else` … `#endif`, emitted only in that mode (`{{- if isNetstandard2}}`, see
  `auxiliary/v2/{{getSDKTopLevelFolder}}/Utils/Sse/SseStreamExtensions.cs.stmpl`). Gate only where netstandard2.0
  lacks an API or would raise a warning; the net8.0+ build must keep its full behaviour.
- CS0618 from `deprecated: true` stays a warning and is not suppressed. The only exception is a file-scoped
  `#pragma warning disable 0618` around the SDK's own converter and extension of a deprecated enum, after the enum
  declaration.
- Keep types internal unless SDK users need them. Converter types named on public models stay public, so an
  application can declare its own `JsonSerializerContext` over SDK models.
- Comments in emitted code (`.stmpl` output) are read by SDK users: XML docs describe what a member does, how to use
  it and what it throws. Implementation notes (AOT, source generation, design) belong in the `.ts` templates.
- Run `TARGET=<variant> make test-csharpv2-aot` for changes that can affect trimming or AOT.

### B2. Migrating an SDK from v1

Examples use the repository's review SDK (`tests/specs/review.yaml`, package `Speakeasy.OpenAPI`, generated with
every `imports.paths` entry set to `""`, so v1 models live in the root namespace). `<Package>` stands for the
configured `packageName`. Snippets without `using` directives assume those shown under
[Model namespaces](#model-namespaces).

#### Switching to v2

Add `templateVersion: v2` to the `csharp` block of `.speakeasy/gen.yaml` and regenerate:

```yaml
csharp:
  templateVersion: v2
  dotnetVersion: net10.0   # net8.0 or net10.0
```

- Regeneration removes files v1 generated and v2 does not (for example `Utils/Sse/EventStream.cs`,
  `Utils/AnyDeserializer.cs`). `Hooks/HookRegistration.cs` and files you added yourself are kept.
- The build needs the .NET SDK named by `dotnetVersion` (pinned in `global.json`, as in v1) and NuGet access for the
  Roslyn packages of the new `<Package>.Generators` project (see [Project layout](#project-layout)).

Version bump. Moving an existing SDK from v1 to v2 is a breaking change: bump the SDK's major version yourself. The
generator treats the `gen.yaml` change as a patch bump (for example `0.0.1` → `0.0.2`) and only logs a warning that
suggests a major bump. Set the `version` key of the `csharp` block, or run the generation with
`SPEAKEASY_BUMP_OVERRIDE=major`.

#### Target frameworks

What changed:
- `dotnetVersion` accepts `net8.0` or `net10.0` only. `net5.0` and `net6.0` fail validation.
- `targetFramework` accepts `netstandard2.0`, `net8.0`, `net10.0` or `dotnetVersion`, and must not exceed
  `dotnetVersion`. `net5.0` and `net6.0` fail validation. `net8.0` and `net10.0` behave as in v1: the SDK targets
  that framework, with the `LangVersion` of `dotnetVersion` when the two differ.
- `targetFramework: netstandard2.0` builds one package for `netstandard2.0;<dotnetVersion>` (v1 targeted
  `netstandard2.0` alone). The SDK project compiles with `LangVersion` 10.0 for both targets (v1 used the language
  version of `dotnetVersion`, e.g. 14.0 for `net10.0`).

```yaml
# v1 and v2
csharp:
  templateVersion: v2        # v2 only
  dotnetVersion: net10.0
  targetFramework: net8.0    # net8.0 SDK, LangVersion 14.0
```

What to do: replace `targetFramework: net5.0` or `net6.0` with `net8.0`; keep `targetFramework: netstandard2.0` if
you need .NET Framework or other netstandard2.0 consumers. In netstandard2.0 mode, code you add to the SDK project
(hooks, partial classes) must compile as C# 10.

#### JSON serialization: Newtonsoft.Json → System.Text.Json

What changed:
- Models carry `System.Text.Json` attributes (`[JsonPropertyName]`, `[JsonRequired]`, `[JsonConverter]`) and are
  serialized through source-generated contracts. The `Newtonsoft.Json` package is no longer referenced.
- The public Newtonsoft helpers are gone: `Utilities.GetDefaultJsonSerializerSettings()`,
  `GetDefaultJsonDeserializerSettings()`, `GetDefaultJsonSerializers()`, `GetDefaultJsonDeserializers()`,
  `GetJsonSerializers()`, `GetJsonDeserializers()`, `SerializeJSON()`, `ParseRawToken()`, the `AnyDeserializer`
  converter, and the nested union converters (`Animal.AnimalConverter`, …). Unions now reference a public
  `<Union>ConverterFactory`; enums a public `<Enum>JsonConverter` / `<Enum>OpenEnumConverter`. The other public
  `Utilities` helpers that served the v1 serializers are gone too: `IsEnum()`, `IsPrimitive()`, `IsString()`,
  `IsSameOrSubclass()`, `IsUnionWrapper()`, `UnwrapValue()`.
- `<Package>.Utils.Json.JsonSerialization` serializes and deserializes SDK models with the SDK's converters and
  contracts. It is trimming- and NativeAOT-safe.

```csharp
// v1
using Newtonsoft.Json;
using Speakeasy.OpenAPI.Foo;
using Speakeasy.OpenAPI.Utils;

var pet = new Pet { Id = "1", Name = "Rex", Species = "dog" };
string json = JsonConvert.SerializeObject(pet, Utilities.GetDefaultJsonSerializerSettings());
Pet? copy = JsonConvert.DeserializeObject<Pet>(json, Utilities.GetDefaultJsonDeserializerSettings());
```

```csharp
// v2
using Speakeasy.OpenAPI.Models.Foo;
using Speakeasy.OpenAPI.Utils.Json;

var pet = new Pet { Id = "1", Name = "Rex", Species = "dog" };
string json = JsonSerialization.Serialize(pet);
Pet? copy = JsonSerialization.Deserialize<Pet>(json);
```

What to do: replace `JsonConvert` calls on SDK models with `JsonSerialization.Serialize` / `Deserialize<T>`.
`Deserialize<T>` always validates strictly, also when the SDK uses `schemaValidation: lenient` (a missing required
property throws `System.Text.Json.JsonException`); an unsupported `T` throws `NotSupportedException`; error exceptions
are not supported, serialize their payload model. `System.Text.Json.JsonSerializer` with your own options or context
applies the converters named on the models, but not what the SDK registers on its own options: the handling of
free-form `object` values, the rejection of explicit `null` for non-nullable fields, reading numbers from JSON strings,
and the NodaTime `LocalDate` converter (`date` fields throw `JsonException` unless you add
`<Package>.Utils.LocalDateJsonConverter` to your options). Without a source-generated context it is not AOT-safe.
If your code uses `Newtonsoft.Json` for other purposes, add your own package reference: it no longer comes with the
SDK.

#### Model namespaces

What changed: v2 always places models under `<Package>.Models`. v1 used the `imports.paths` value as given, so SDKs
whose paths are empty (all models in the root namespace) or do not start with `Models/` see their model and error
namespaces change. With the default paths (`Models/Components`, `Models/Shared`, `Models/Errors`, …) model namespaces
are unchanged. Independently of `imports`, `OptionalNullable<T>`, `IOptionalNullable` and `IOpenUnion` move from
`<Package>.Utils` to `<Package>.Models`.

```csharp
// v1
using Speakeasy.OpenAPI;          // Emoji, ExhaustiveObject, …
using Speakeasy.OpenAPI.Foo;      // Pet
using Speakeasy.OpenAPI.Utils;    // OptionalNullable<T>

var pet = new Pet { Id = "1", Name = "Rex", Species = "dog" };
Emoji emoji = Emoji.Wave;
OptionalNullable<string?> nickname = OptionalNullable<string?>.Absent;
```

```csharp
// v2
using Speakeasy.OpenAPI;          // SDK, sub-SDKs, hooks
using Speakeasy.OpenAPI.Models;   // Emoji, ExhaustiveObject, OptionalNullable<T>, …
using Speakeasy.OpenAPI.Models.Foo;

var pet = new Pet { Id = "1", Name = "Rex", Species = "dog" };
Emoji emoji = Emoji.Wave;
OptionalNullable<string?> nickname = OptionalNullable<string?>.Absent;
```

What to do: update `using` directives and fully qualified names; move `partial` class extensions of models to the new
namespace.

#### Server-sent events: `EventStream<T>` → `IAsyncEnumerable<T>`

What changed: streaming response properties are `IAsyncEnumerable<T>`. `EventStream<T>`, its `Next()` method,
`IsClosed`, `Dispose()` / `DisposeAsync()`, and `SseStreamExtensions.AsServerSentEventStream<T>()` are gone.
Enumeration disposes the response stream when it completes or exits through `break`, cancellation or an exception;
the sequence can be enumerated once. The `CancellationToken` passed to the operation also cancels the enumeration
(see [Exceptions](#exceptions)).

```csharp
// v1
var res = await sdk.ChatAsync(ChatRequest.CreateChatModelRequest(new ChatModelRequest { Prompt = "Hello" }));
using var stream = res.ChatStream!;
ChatStream? ev;
while ((ev = await stream.Next()) != null)
{
    Console.WriteLine(ev.Data);
}
```

```csharp
// v2
var res = await sdk.ChatAsync(ChatRequest.CreateChatModelRequest(new ChatModelRequest { Prompt = "Hello" }), cancellationToken);
if (res.ChatStream != null)
{
    await foreach (ChatStream ev in res.ChatStream)
    {
        Console.WriteLine(ev.Data);
    }
}
```

What to do: replace `Next()` loops with `await foreach`; drop `using` / `Dispose` on the stream. Code that already
used `await foreach` (as the v1 README showed) compiles unchanged.

#### Open unions: raw unknown payload is a `JsonElement`

What changed: `UnknownRaw` is `System.Text.Json.JsonElement?` (was `Newtonsoft.Json.Linq.JToken?`) and
`CreateUnknown` takes a `JsonElement`. `IOpenUnion` gains `GetUnknownRaw()`.

```csharp
// v1
using Newtonsoft.Json.Linq;
using Speakeasy.OpenAPI.Foo;

var res = await sdk.NamespaceTests.Types.GetNamespaceAnimalAsync();
if (res.Animal is { } animal && animal.IsUnknown())
{
    string? raw = animal.UnknownRaw?.ToString(Newtonsoft.Json.Formatting.None);
}
Animal bird = Animal.CreateUnknown(JToken.Parse("{\"animalType\":\"bird\"}"));
```

```csharp
// v2
using System.Text.Json;
using Speakeasy.OpenAPI.Models.Foo;

var res = await sdk.NamespaceTests.Types.GetNamespaceAnimalAsync();
if (res.Animal is { } animal && animal.IsUnknown())
{
    string? raw = animal.UnknownRaw?.GetRawText();
}
using var doc = JsonDocument.Parse("{\"animalType\":\"bird\"}");
Animal bird = Animal.CreateUnknown(doc.RootElement.Clone());
```

What to do: read the raw payload through the `JsonElement` API; pass a cloned `JsonElement` to `CreateUnknown`.

#### Open enums: `IOpenEnum` removed

What changed: open enum classes no longer implement `<Package>.Utils.IOpenEnum` / `IOpenEnum<T>`, and
`Utilities.IsOpenEnum()` is gone. The class members (`Value`, `Of()`, `Values()`, `IsKnown()`, implicit conversions)
are unchanged.

```csharp
// v1
using Speakeasy.OpenAPI.Bar;
using Speakeasy.OpenAPI.Utils;

static bool IsKnownValue(object value) => value is IOpenEnum e && e.IsKnown();
```

```csharp
// v2
using Speakeasy.OpenAPI.Models.Bar;

static bool IsKnownValue(PetStatus value) => value.IsKnown();
```

What to do: call `IsKnown()` on the concrete open enum type.

#### Free-form `object` values

What changed: a property typed `object` (schema without a type) holds plain .NET values. v1 produced a `JObject`
for a JSON object and a `List<object>` of each element's JSON text for an array. v2 produces
`Dictionary<string, object?>` for objects, `List<object?>` for arrays, and `long`, `double`, `string`, `bool` or
`null` for scalars, recursively.

```csharp
// v1
using Newtonsoft.Json.Linq;

var anything = (JObject)payload.Anything;
string? name = (string?)anything["name"];
```

```csharp
// v2
using System.Collections.Generic;

var anything = (Dictionary<string, object?>)payload.Anything;
string? name = anything["name"] as string;
```

What to do: replace `JObject` / `JArray` casts and JSON-text parsing of array items with dictionary / list access.

#### Additional properties

What changed: a schema with free-form `additionalProperties` (`true` or `{}`) generates
`Dictionary<string, JsonElement>? AdditionalProperties` (was `Dictionary<string, object>`). A typed
`additionalProperties` schema keeps its `Dictionary<string, T>?` type. Entries are still read from and written to the
model's own JSON object, next to the declared properties:

- a key equal to a declared property's JSON name, ignoring case, is never an additional property: on read it is
  dropped (v1 bound a case-variant key such as `NormalField` to the declared `normalField`), on write the declared
  property wins;
- each value is read and written with the JSON contract of `T`: a value that does not match `T` fails the read, and a
  JSON `null` is kept for reference and `Nullable<T>` types and rejected for other value types; a free-form `null` is
  a `JsonElement` of kind `Null`;
- the property is `[JsonIgnore]` on the model: `JsonSerialization` and the SDK's requests and responses handle it,
  a `JsonSerializer` call with your own options and resolver does not;
- a `default(JsonElement)` entry cannot be written.

```csharp
// v1
var metadata = new Metadata { AdditionalProperties = new Dictionary<string, object> { ["count"] = 1 } };
long count = (long)metadata.AdditionalProperties["count"];
```

```csharp
// v2
using System.Text.Json;

using var count1 = JsonDocument.Parse("1");
var metadata = new Metadata
{
    AdditionalProperties = new Dictionary<string, JsonElement> { ["count"] = count1.RootElement.Clone() },
};
long count = metadata.AdditionalProperties["count"].GetInt64();
```

What to do: build free-form values as `JsonElement` and read them through the `JsonElement` API.

Known limitation: when a union variant declares typed additional properties, union selection counts each additional
key as one undeclared property without scoring its value against `T`. A value that does not match `T` still fails
when that variant is deserialized.

#### `presenceAwareJsonSerialization` is always on

What changed: v2 forces `presenceAwareJsonSerialization: true`. SDKs generated with `false` (the default for SDKs
created before the option existed) change in two ways:
- Optional nullable fields become `OptionalNullable<T>` instead of `T?`. Assigning `null` now sends an explicit JSON
  `null`; v1 omitted the field. To omit it, leave the default or assign `OptionalNullable<T>.Absent`.
- Responses are validated for presence: a missing required field or an explicit `null` for a non-nullable field
  raises `ResponseValidationException`. v1 accepted both.

Example from the repository's quaternary SDK (`Company.Product.Feature.Subnamespace`):

```csharp
// v1 (presenceAwareJsonSerialization: false)
using Company.Product.Feature.Subnamespace.Models.Shared;

string? value = model.DefaultStrNullable;
model.DefaultStrNullable = null;                              // field omitted
```

```csharp
// v2
using Company.Product.Feature.Subnamespace.Models;
using Company.Product.Feature.Subnamespace.Models.Shared;

string? value = model.DefaultStrNullable.ValueOrDefault;
model.DefaultStrNullable = OptionalNullable<string?>.Absent;  // field omitted
model.DefaultStrNullable = null;                              // sent as JSON null
```

What to do: read wrapped fields through `ValueOrDefault` (or `IsSet` / `Value`), use `Absent` to omit a field, and set
`schemaValidation: lenient` if the API returns responses that do not match the schema.

#### `optionalNullableScope: writable-models` is replaced by `all`

What changed: v2 wraps optional nullable fields in `OptionalNullable<T>` on every model and logs a warning when
`writable-models` is configured. Response-only models that had plain `T?` fields now use `OptionalNullable<T>`, and the
implicit conversion from `OptionalNullable<T>` to `T?` that `writable-models` added is gone.

```csharp
// v1 (optionalNullableScope: writable-models)
string? value = response.SomeOptionalNullableField;
```

```csharp
// v2
string? value = response.SomeOptionalNullableField.ValueOrDefault;
```

What to do: read those fields through `ValueOrDefault`.

#### `unionStrategy: left-to-right` is replaced by `populated-fields`

What changed: undiscriminated unions are always resolved by scoring which member matches the fields present in the
payload; v2 logs a warning when `left-to-right` is configured. A payload that v1 assigned to the first member that
deserialized may now resolve to a different member.

```yaml
# v1
csharp:
  unionStrategy: left-to-right
```

```yaml
# v2: the key is removed from gen.yaml after a warning
csharp:
  templateVersion: v2
```

What to do: check code that branches on which member of an undiscriminated union is set.

#### Exceptions

The exception types thrown by SDK methods and their messages are unchanged, apart from model namespaces in messages.
Three cases differ.

`ResponseValidationException.InnerException`: when a response body cannot be deserialized, the inner exception is a
`System.Text.Json.JsonException` (v1: `Newtonsoft.Json.JsonException` subclasses, or `Exception` for an unexpected
`null` body).

```csharp
// v1
catch (ResponseValidationException e) when (e.InnerException is Newtonsoft.Json.JsonException)
```

```csharp
// v2
catch (ResponseValidationException e) when (e.InnerException is System.Text.Json.JsonException)
```

Cancelling a server-sent event stream: the operation's `CancellationToken` (or one passed with `WithCancellation`)
cancels the enumeration.
- net8.0 and later: cancellation interrupts a pending read and throws `TaskCanceledException`; when the next line
  has already been received it throws `OperationCanceledException`. `TaskCanceledException` derives from
  `OperationCanceledException`.
- netstandard2.0 (in `targetFramework: netstandard2.0` mode, for consumers that load that asset): the token is
  checked when the next line arrives (`OperationCanceledException`); a pending read is not interrupted, and when the
  stream ends first the enumeration completes normally.
- v1 checked the enumeration token only between events (`OperationCanceledException`) and never interrupted a
  pending read; `Next()` ignored its token.

Read errors on a server-sent event stream: an I/O error while the stream is read, such as a connection reset, is
thrown from the enumeration as an `IOException`. v1 ended the stream without an exception.

What to do: update filters on the inner exception type; catch `OperationCanceledException`, not an exact type, around
stream enumeration, and catch `IOException` there if your code relied on a dropped connection ending the stream.

#### Hooks and custom code in the SDK project

What changed: `Hooks/HookRegistration.cs` and your own hook files are not regenerated, so they keep any
`Newtonsoft.Json` usage, which no longer compiles. The hook interfaces (`IBeforeRequestHook`, `IAfterSuccessHook`, …)
are unchanged. On net8.0+ the SDK project enables the trimming and AOT analyzers and treats their warnings as errors,
so reflection-based serialization in custom code fails the build with IL2026 / IL3050. Helpers custom code may have
called (`Utilities.SerializeJSON`, `ResponseBodyDeserializer`, `RequestBodySerializer.Serialize`) changed or were
removed.

```csharp
// v1
using Newtonsoft.Json.Linq;
using Speakeasy.OpenAPI;

public class TraceIdHook : IAfterSuccessHook
{
    public string? LastTraceId { get; private set; }

    public async Task<HttpResponseMessage> AfterSuccessAsync(AfterSuccessContext hookCtx, HttpResponseMessage response)
    {
        var body = await response.Content.ReadAsStringAsync();
        LastTraceId = (string?)JObject.Parse(body)["traceId"];
        return response;
    }
}
```

```csharp
// v2
using System.Text.Json;
using Speakeasy.OpenAPI;

public class TraceIdHook : IAfterSuccessHook
{
    public string? LastTraceId { get; private set; }

    public async Task<HttpResponseMessage> AfterSuccessAsync(AfterSuccessContext hookCtx, HttpResponseMessage response)
    {
        var body = await response.Content.ReadAsStringAsync();
        using var doc = JsonDocument.Parse(body);
        LastTraceId = doc.RootElement.TryGetProperty("traceId", out var id) ? id.GetString() : null;
        return response;
    }
}
```

What to do: port custom hooks and partial classes to `System.Text.Json` (`JsonDocument`, a `JsonSerializerContext`,
or `JsonSerialization` for SDK models); the generated client-credentials hook shows the source-generated pattern.

#### Project layout

What changed: v2 adds a `<Package>.Generators` project next to the SDK project and to the solution. It holds Roslyn
source generators (union, serialization and security code) that the SDK project references as an analyzer. The
generators run at build time only: the NuGet package still contains one SDK assembly per target framework and no new
dependency.

```text
v1                                   v2
src/Speakeasy/OpenAPI/               src/Speakeasy/OpenAPI/
  Speakeasy.OpenAPI.csproj             Speakeasy.OpenAPI.csproj   (analyzer reference ↓)
                                     src/Speakeasy/OpenAPI.Generators/
                                       Speakeasy.OpenAPI.Generators.csproj
```

What to do: if you reference the SDK project by path or copy its sources, keep the `Generators` folder next to it;
build pipelines need NuGet access for `Microsoft.CodeAnalysis.CSharp` 4.3.0. Package consumers need no change.

#### Configuration keys

| Key | v1 | v2 |
| --- | --- | --- |
| `templateVersion` | absent | `v2` selects this generator |
| `dotnetVersion` | `net5.0`, `net6.0`, `net8.0`, `net10.0` | `net8.0` or `net10.0`; other values are rejected |
| `targetFramework` | any framework up to `dotnetVersion` | `net8.0` or `net10.0` up to `dotnetVersion`, as in v1; `netstandard2.0` builds `netstandard2.0;<dotnetVersion>`; other values are rejected |
| `presenceAwareJsonSerialization` | default `true` for new SDKs | fixed `true`: assigning `null` to an optional nullable field sends JSON `null` (see [above](#presenceawarejsonserialization-is-always-on)) |
| `optionalNullableScope` | `all` or `writable-models` | fixed `all`: response-only models also get `OptionalNullable<T>` fields (see [above](#optionalnullablescope-writable-models-is-replaced-by-all)) |
| `unionStrategy` | `left-to-right` or `populated-fields` | fixed `populated-fields` |
| `partialTypes` | default `false` | fixed `true` |
| `unionImplicitConversions` | default `false` | fixed `true` |
| `enableFormatting` | default `true` for new SDKs; `false` runs `dotnet format` at build | fixed `true` |
| `preserveUnknownFields` | — | new, default `false`: keeps undeclared JSON fields in an `UnknownFields` dictionary per model |
| `forwardCompatibleUnionsByDefault: true` | rewritten to `tagged-and-untagged` | kept as `true`, same behaviour |

The six fixed keys support one value each in v2 and are not v2 options: a new `gen.yaml` does not list them and the
configuration documentation does not describe them. A fixed key set to another value logs one warning, for example
`csharp: unionStrategy "left-to-right" is not supported with templateVersion: v2. Using "populated-fields".`; a key
set to the fixed value, or left empty, is accepted without a warning. Generation uses the fixed value and saves
`gen.yaml` without the key, so the warning appears on the first successful generation after the switch only.

Switching back to v1 does not restore removed values: v1 applies its defaults for an existing SDK
(`presenceAwareJsonSerialization`, `partialTypes`, `unionImplicitConversions` and `enableFormatting` off,
`unionStrategy: left-to-right`, `optionalNullableScope: all`). Add the values you need back to `gen.yaml` before
regenerating with v1.

Every other `csharp` key keeps its v1 meaning.

#### NuGet dependencies

| Package | v1 | v2 |
| --- | --- | --- |
| `Newtonsoft.Json` | always | removed |
| `System.Text.Json` | — | netstandard2.0 target only (in-box on net8.0+) |
| `Microsoft.Bcl.AsyncInterfaces` | netstandard2.0 SDKs | netstandard2.0 target only |
| `NodaTime` | when `useNodatime` or netstandard2.0 / net5.0 | when `useNodatime` or netstandard2.0, and only when the SDK has `date` fields |

What to do: add direct package references for any of these your own code uses.

#### No action needed

- SDK class, sub-SDKs, `SDKBuilder`, `SDKConfig`, operation method signatures, request and response envelope types,
  `HTTPMetadata`: unchanged apart from model namespaces.
- Error classes (`<baseErrorName>`, `<defaultErrorName>`, `ResponseValidationException`, typed errors): same members;
  messages name types with their new namespace.
- Hook interfaces and contexts, `SDKHooks`, `RetryConfig` / `BackoffStrategy`, `ISpeakeasyHttpClient` /
  `SpeakeasyHttpClient`, `Constants`, `SseEvent`, `SseStreamParser`, `SseParseException`: unchanged.
- Closed enums: `<Enum>Extension.Value()` / `ToEnum()` unchanged; integer enums gain the same extension class.
- Unions: `Create*` factories, implicit conversions, `IsUnknown()`, the `Type` property: unchanged.
- `OptionalNullable<T>`: same members (`Absent`, `Set`, `IsSet`, `IsNull`, `Value`, `ValueOrDefault`); only the
  namespace moves.
- `schemaValidation: lenient`, `responseFormat`, `maxMethodParams`, `flatteningOrder`, `enableCancellationToken`,
  `bodyVariantOverloads`, `multipartArrayFormat`, `useNodatime` and the other `csharp` keys behave as in v1.
- `global.json` is still written from `dotnetVersion`.
- The generated client-credentials hook deserializes its token with a source-generated context; it is regenerated on
  every run, as in v1. `TokenResponse` properties gain public setters.
- Newtonsoft conditional-serialization methods (`ShouldSerialize<Property>()`) are no longer generated; they existed
  for Newtonsoft and were not meant to be called.
- Types that were `internal` in v1 (`URLBuilder`, `RequestBodySerializer`, `HeaderSerializer`, `SpeakeasyMetadata`,
  `ResponseBodyDeserializer`, `EnumConverter`, …) are replaced by other internal types.
- New public types and members (`JsonSerialization`, `SseStreamExtensions.AsAsyncEnumerable<T>()`,
  `Utilities.UnknownPrimitiveToString()`, and the converters below) are additions.
- Converters named in `[JsonConverter]` attributes on models are public, so that a `JsonSerializerContext` in your own
  assembly applies them: `<Union>ConverterFactory`, `<Enum>JsonConverter`, `<Enum>OpenEnumConverter`,
  `OptionalNullableJsonConverter<T>` and the `Utils` factories `BigIntegerJsonConverterFactory`,
  `BigIntegerStringJsonConverterFactory`, `DecimalJsonConverterFactory`, `DecimalStringJsonConverterFactory` and
  `DateTimeJsonConverterFactory`. `Utils.LocalDateJsonConverter` is public so you can add it to your own options. The
  converters these factories create are internal.
- On net8.0 and later the SDK is marked trimming- and NativeAOT-compatible.
