package snapshots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

const cliStickyShorthandSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Optional body fields whose initials can collide with a flag that already shipped with a shorthand.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
%s
      responses:
        '200':
          description: OK
`

const cliStickyGenYaml = `cli:
  packageName: github.com/example/pet-cli
  cliName: pet-cli
  envVarPrefix: PET
`

const (
	cliStickyNamesField    = "                names:\n                  type: array\n                  items:\n                    type: string\n"
	cliStickyTagField      = "                tag:\n                  type: string\n"
	cliStickyNickNameField = "                nick_name:\n                  type: string\n"
)

type cliStickyShorthandEntry struct {
	Flags   map[string]string `yaml:"flags"`
	Retired map[string]struct {
		Flag string `yaml:"flag"`
	} `yaml:"retired"`
}

type cliStickyShorthandLock struct {
	CLIShorthands map[string]map[string]cliStickyShorthandEntry `yaml:"cliShorthands"`
}

func cliStickyShorthandSpecWith(fields ...string) string {
	return fmt.Sprintf(cliStickyShorthandSpec, strings.TrimRight(strings.Join(fields, ""), "\n"))
}

func cliStickyReadFile(t *testing.T, outputDir string, elem ...string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(append([]string{outputDir}, elem...)...))
	require.NoError(t, err)
	return string(content)
}

// cliStickyUsageCommand returns the usage schema entry of one command, keyed
// by its command path, so assertions cannot match another command's flags.
func cliStickyUsageCommand(t *testing.T, outputDir, commandPath string) string {
	t.Helper()
	prefix := strconv.Quote(commandPath) + ":"
	for _, line := range strings.Split(cliStickyReadFile(t, outputDir, "internal", "usage", "schema.go"), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return line
		}
	}
	require.FailNow(t, "usage entry not found", commandPath)
	return ""
}

func cliStickyReadLock(t *testing.T, outputDir string) cliStickyShorthandLock {
	t.Helper()
	var lock cliStickyShorthandLock
	require.NoError(t, yaml.Unmarshal([]byte(cliStickyReadFile(t, outputDir, ".speakeasy", "gen.lock")), &lock))
	return lock
}

func cliStickyLockFlags(t *testing.T, outputDir string) map[string]map[string]map[string]string {
	t.Helper()
	flags := map[string]map[string]map[string]string{}
	for operationID, methods := range cliStickyReadLock(t, outputDir).CLIShorthands {
		flags[operationID] = map[string]map[string]string{}
		for method, entry := range methods {
			require.Empty(t, entry.Retired, "%s %s", operationID, method)
			flags[operationID][method] = entry.Flags
		}
	}
	return flags
}

func cliStickyDeleteRetired(t *testing.T, outputDir, operationID, method, letter string) {
	t.Helper()
	lockPath := filepath.Join(outputDir, ".speakeasy", "gen.lock")
	var lock map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(cliStickyReadFile(t, outputDir, ".speakeasy", "gen.lock")), &lock))
	entry := lock["cliShorthands"].(map[string]any)[operationID].(map[string]any)[method].(map[string]any)
	retired := entry["retired"].(map[string]any)
	require.Contains(t, retired, letter)
	delete(retired, letter)
	content, err := yaml.Marshal(lock)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(lockPath, content, 0o644))
}

func cliStickyRegenerate(t *testing.T, outputDir, spec string, opts ...generate.GeneratorOptions) []error {
	t.Helper()
	generator, err := generate.New(append([]generate.GeneratorOptions{generate.WithLogger(logging.NewLogger(zapcore.WarnLevel))}, opts...)...)
	require.NoError(t, err)
	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	require.Empty(t, generator.Generate(ctx, []byte(spec), "test-schema.yaml", "cli", outputDir, false, false))
	return generator.GetWarnings()
}

func TestSnapCLIStickyShorthands(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyTagField),
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()

			readCommand := func() string {
				t.Helper()
				return cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go")
			}
			readUsage := func() string {
				t.Helper()
				return cliStickyReadFile(t, outputDir, "internal", "usage", "schema.go")
			}
			readLock := func() cliStickyShorthandLock {
				t.Helper()
				return cliStickyReadLock(t, outputDir)
			}
			regenerate := func(spec string) {
				t.Helper()
				cliStickyRegenerate(t, outputDir, spec)
			}

			require.Contains(t, readCommand(), `FlagName: "names", Shorthand: "n"`)
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, readLock().CLIShorthands["CreatePet"]["json"].Flags)

			regenerate(cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyNickNameField, cliStickyTagField))
			command := readCommand()
			require.Contains(t, command, `FlagName: "names", Shorthand: "n"`)
			require.Contains(t, command, `FlagName: "nick-name"`)
			require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
			usage := readUsage()
			require.Contains(t, usage, `-n --names <names...>`)
			require.NotContains(t, usage, `-n --nick-name`)
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, readLock().CLIShorthands["CreatePet"]["json"].Flags)

			regenerate(cliStickyShorthandSpecWith(cliStickyNickNameField, cliStickyTagField))
			command = readCommand()
			require.Contains(t, command, `FlagName: "nick-name"`)
			require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
			entry := readLock().CLIShorthands["CreatePet"]["json"]
			require.Equal(t, map[string]string{"tag": "t"}, entry.Flags)
			require.Equal(t, "names", entry.Retired["n"].Flag)

			cliStickyDeleteRetired(t, outputDir, "CreatePet", "json", "n")
			regenerate(cliStickyShorthandSpecWith(cliStickyNickNameField, cliStickyTagField))
			require.Contains(t, readCommand(), `FlagName: "nick-name", Shorthand: "n"`)
			require.Contains(t, readUsage(), `-n --nick-name <nick_name>`)
			entry = readLock().CLIShorthands["CreatePet"]["json"]
			require.Equal(t, map[string]string{"nick-name": "n", "tag": "t"}, entry.Flags)
			require.Empty(t, entry.Retired)
		},
	})
}

const cliStickySplitSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Request content types with different fields render as separate commands that each keep their own shorthands.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                bool: {type: boolean}
                num: {type: number}
          multipart/form-data:
            schema:
              type: object
              properties:
                bool2: {type: boolean}
                num2: {type: number}
      responses:
        '200':
          description: OK
`

func TestSnapCLIStickyShorthandsSplitOperation(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          cliStickySplitSpec,
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()

			readOutputs := func() []string {
				t.Helper()
				return []string{
					cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go"),
					cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpetmultipart.go"),
					cliStickyReadFile(t, outputDir, "internal", "usage", "schema.go"),
				}
			}
			wantLock := map[string]map[string]map[string]string{
				"CreatePet": {
					"json":      {"bool": "b", "num": "n"},
					"multipart": {"bool2": "b", "num2": "n"},
				},
			}

			first := readOutputs()
			require.Contains(t, first[0], `FlagName: "bool", Shorthand: "b"`)
			require.Contains(t, first[0], `FlagName: "num", Shorthand: "n"`)
			require.Contains(t, first[1], `FlagName: "bool2", Shorthand: "b"`)
			require.Contains(t, first[1], `FlagName: "num2", Shorthand: "n"`)
			require.Contains(t, first[2], `-b --bool2`)
			require.Contains(t, first[2], `-n --num2`)
			require.Equal(t, wantLock, cliStickyLockFlags(t, outputDir))

			cliStickyRegenerate(t, outputDir, cliStickySplitSpec)
			require.Equal(t, first, readOutputs())
			require.Equal(t, wantLock, cliStickyLockFlags(t, outputDir))
		},
	})
}

const cliStickyListSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    get:
      operationId: ListPets
      description: Query parameters that can all be removed, leaving the operation without inputs.
      tags: [pets]
      security: []
%s      responses:
        '200':
          description: OK
`

func cliStickyListSpecWith(param string) string {
	if param == "" {
		return fmt.Sprintf(cliStickyListSpec, "")
	}
	return fmt.Sprintf(cliStickyListSpec, "      parameters:\n        - name: "+param+"\n          in: query\n          schema:\n            type: string\n")
}

const cliStickyWholeBodySpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    post:
      operationId: CreatePet
      description: An array request body is sent as one whole-body JSON flag without per-field flags.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: array
              items:
                type: string
      responses:
        '200':
          description: OK
`

func TestSnapCLIStickyShorthandsLostInputs(t *testing.T) {
	t.Parallel()

	t.Run("operation loses its last parameter", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          cliStickyListSpecWith("names"),
			GenYaml:       cliStickyGenYaml,
			ShouldCompile: &shouldCompile,
			AfterGenerate: func(t *testing.T, outputDir string) {
				t.Helper()
				require.Equal(t, map[string]string{"names": "n"}, cliStickyReadLock(t, outputDir).CLIShorthands["ListPets"]["json"].Flags)

				cliStickyRegenerate(t, outputDir, cliStickyListSpecWith(""))
				entry := cliStickyReadLock(t, outputDir).CLIShorthands["ListPets"]["json"]
				require.Empty(t, entry.Flags)
				require.Equal(t, "names", entry.Retired["n"].Flag)

				cliStickyRegenerate(t, outputDir, cliStickyListSpecWith("nick_name"))
				command := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "listpets.go")
				require.Contains(t, command, `FlagName: "nick-name"`)
				require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
				entry = cliStickyReadLock(t, outputDir).CLIShorthands["ListPets"]["json"]
				require.Empty(t, entry.Flags)
				require.Equal(t, "names", entry.Retired["n"].Flag)
			},
		})
	})

	t.Run("body becomes a whole-body JSON flag", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyTagField),
			GenYaml:       cliStickyGenYaml,
			ShouldCompile: &shouldCompile,
			AfterGenerate: func(t *testing.T, outputDir string) {
				t.Helper()
				require.Equal(t, map[string]string{"names": "n", "tag": "t"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"]["json"].Flags)

				cliStickyRegenerate(t, outputDir, cliStickyWholeBodySpec)
				entry := cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"]["json"]
				require.Empty(t, entry.Flags)
				require.Equal(t, "names", entry.Retired["n"].Flag)
				require.Equal(t, "tag", entry.Retired["t"].Flag)
			},
		})
	})
}

func TestSnapCLIStickyShorthandsPrototypeOperationID(t *testing.T) {
	t.Parallel()

	spec := strings.Replace(cliStickyListSpecWith("names"), "operationId: ListPets", "operationId: constructor", 1) + `  /pets/count:
    get:
      operationId: toString
      description: An operation without inputs whose ID is an inherited object property.
      tags: [pets]
      security: []
      responses:
        '200':
          description: OK
`
	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          spec,
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()
			require.Equal(t, map[string]string{"names": "n"}, cliStickyReadLock(t, outputDir).CLIShorthands["constructor"]["json"].Flags)

			cliStickyRegenerate(t, outputDir, spec)
			lock := cliStickyReadLock(t, outputDir).CLIShorthands
			require.Equal(t, map[string]string{"names": "n"}, lock["constructor"]["json"].Flags)
			require.NotContains(t, lock, "toString")
		},
	})
}

const cliStickyContentTypeSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Request content types are added and removed while each one keeps its recorded shorthands.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
%s      responses:
        '200':
          description: OK
`

func cliStickyRequestContent(contentTypes map[string][]string) string {
	var b strings.Builder
	for _, contentType := range []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"} {
		fields, ok := contentTypes[contentType]
		if !ok {
			continue
		}
		b.WriteString("          " + contentType + ":\n            schema:\n              type: object\n              properties:\n")
		for _, field := range fields {
			b.WriteString("                " + field + ": {type: string}\n")
		}
	}
	return b.String()
}

func cliStickyContentTypeSpecWith(contentTypes map[string][]string) string {
	return fmt.Sprintf(cliStickyContentTypeSpec, cliStickyRequestContent(contentTypes))
}

func TestSnapCLIStickyShorthandsContentTypeChanges(t *testing.T) {
	t.Parallel()

	type generation struct {
		contentTypes map[string][]string
		usage        []string
		lock         map[string]map[string]map[string]string
	}
	sequences := map[string][]generation{
		"multipart operation gains and loses a JSON body": {
			{
				contentTypes: map[string][]string{"multipart/form-data": {"names", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]map[string]string{"CreatePet": {"multipart": {"names": "n", "tag": "t"}}},
			},
			{
				contentTypes: map[string][]string{
					"application/json":    {"nova", "title"},
					"multipart/form-data": {"names", "nick_name", "tag"},
				},
				usage: []string{`-n --names <names>`, `-t --tag <tag>`, `-n --nova <nova>`, `-t --title <title>`},
				lock: map[string]map[string]map[string]string{"CreatePet": {
					"json":      {"nova": "n", "title": "t"},
					"multipart": {"names": "n", "tag": "t"},
				}},
			},
			{
				contentTypes: map[string][]string{"multipart/form-data": {"names", "nick_name", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]map[string]string{"CreatePet": {"multipart": {"names": "n", "tag": "t"}}},
			},
		},
		"JSON operation gains and loses a multipart body": {
			{
				contentTypes: map[string][]string{"application/json": {"names", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]map[string]string{"CreatePet": {"json": {"names": "n", "tag": "t"}}},
			},
			{
				contentTypes: map[string][]string{
					"application/json":    {"names", "nick_name", "tag"},
					"multipart/form-data": {"nova", "title"},
				},
				usage: []string{`-n --names <names>`, `-t --tag <tag>`, `-n --nova <nova>`, `-t --title <title>`},
				lock: map[string]map[string]map[string]string{"CreatePet": {
					"json":      {"names": "n", "tag": "t"},
					"multipart": {"nova": "n", "title": "t"},
				}},
			},
			{
				contentTypes: map[string][]string{"application/json": {"names", "nick_name", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]map[string]string{"CreatePet": {"json": {"names": "n", "tag": "t"}}},
			},
		},
	}

	for _, nameResolutionFeb2025 := range []bool{true, false} {
		for name, generations := range sequences {
			t.Run(fmt.Sprintf("%s nameResolutionFeb2025=%t", name, nameResolutionFeb2025), func(t *testing.T) {
				t.Parallel()
				shouldCompile := false
				snaptest.DoTestSnapshot(t, snaptest.Options{
					Spec:          cliStickyContentTypeSpecWith(generations[0].contentTypes),
					GenYaml:       cliStickyGenYaml + fmt.Sprintf("generation:\n  fixes:\n    nameResolutionFeb2025: %t\n", nameResolutionFeb2025),
					ShouldCompile: &shouldCompile,
					AfterGenerate: func(t *testing.T, outputDir string) {
						t.Helper()
						for i, gen := range generations {
							if i > 0 {
								cliStickyRegenerate(t, outputDir, cliStickyContentTypeSpecWith(gen.contentTypes))
							}
							usage := cliStickyReadFile(t, outputDir, "internal", "usage", "schema.go")
							for _, line := range gen.usage {
								require.Contains(t, usage, line, "generation %d", i+1)
							}
							require.NotContains(t, usage, `-n --nick-name`, "generation %d", i+1)
							require.Equal(t, gen.lock, cliStickyLockFlags(t, outputDir), "generation %d", i+1)
						}
					},
				})
			})
		}
	}
}

func TestSnapCLIStickyShorthandsPinDisplacesRecordedFlag(t *testing.T) {
	t.Parallel()

	pinnedSpec := strings.Replace(
		cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyNickNameField, cliStickyTagField),
		"paths:\n",
		"x-speakeasy-cli-commands:\n  version: 1\n  operations:\n    CreatePet:\n      preferShorthand: [names]\npaths:\n",
		1,
	)
	countWarnings := func(warnings []error) int {
		count := 0
		for _, warning := range warnings {
			if strings.Contains(warning.Error(), `preferShorthand "names" takes -n from "nick-name"`) {
				count++
			}
		}
		return count
	}

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          cliStickyShorthandSpecWith(cliStickyNickNameField, cliStickyTagField),
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()
			require.Equal(t, map[string]string{"nick-name": "n", "tag": "t"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"]["json"].Flags)

			require.Equal(t, 1, countWarnings(cliStickyRegenerate(t, outputDir, pinnedSpec)))
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"]["json"].Flags)

			require.Zero(t, countWarnings(cliStickyRegenerate(t, outputDir, pinnedSpec)))
		},
	})
}

const cliStickySuffixedOperationIDSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    post:
      operationId: CreatePet
      description: A form request body on an operation whose ID another operation repeats with a serialization suffix.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                names:
                  type: array
                  items:
                    type: string
      responses:
        '200':
          description: OK
  /notes:
    post:
      operationId: CreatePet_form
      description: A JSON request body on an operation whose ID ends with another operation's serialization suffix.
      tags: [notes]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                nick_name:
                  type: string
      responses:
        '200':
          description: OK
`

func TestSnapCLIStickyShorthandsSuffixedOperationID(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          cliStickySuffixedOperationIDSpec,
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()

			readOutputs := func() []string {
				t.Helper()
				return []string{
					cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go"),
					cliStickyReadFile(t, outputDir, "internal", "cli", "notes", "createpetform.go"),
					cliStickyReadFile(t, outputDir, "internal", "usage", "schema.go"),
				}
			}
			wantLock := map[string]map[string]map[string]string{
				"CreatePet":      {"form": {"names": "n"}},
				"CreatePet_form": {"json": {"nick-name": "n"}},
			}

			first := readOutputs()
			require.Contains(t, first[0], `FlagName: "names", Shorthand: "n"`)
			require.Contains(t, first[1], `FlagName: "nick-name", Shorthand: "n"`)
			require.Contains(t, first[2], `-n --names <names...>`)
			require.Contains(t, first[2], `-n --nick-name <nick_name>`)
			require.Equal(t, wantLock, cliStickyLockFlags(t, outputDir))

			cliStickyRegenerate(t, outputDir, cliStickySuffixedOperationIDSpec)
			require.Equal(t, first, readOutputs())
			require.Equal(t, wantLock, cliStickyLockFlags(t, outputDir))
		},
	})
}

const cliStickyOptionalBodySpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /pets:
    post:
      operationId: CreatePet
      description: A command with a query parameter that gains and loses an optional request body with a colliding field.
      tags: [pets]
      security: []
      parameters:
        - name: names
          in: query
          schema:
            type: string
%s      responses:
        '200':
          description: OK
`

func cliStickyOptionalBodySpecWith(contentType string) string {
	if contentType == "" {
		return fmt.Sprintf(cliStickyOptionalBodySpec, "")
	}
	return fmt.Sprintf(cliStickyOptionalBodySpec, "      requestBody:\n        content:\n          "+contentType+":\n            schema:\n              type: object\n              properties:\n                nick_name:\n                  type: string\n")
}

func TestSnapCLIStickyShorthandsRequestBodySerializationChanges(t *testing.T) {
	t.Parallel()

	type generation struct {
		contentType string
		method      string
	}
	generations := []generation{
		{contentType: "", method: "json"},
		{contentType: "multipart/form-data", method: "multipart"},
		{contentType: "multipart/form-data", method: "multipart"},
		{contentType: "application/x-www-form-urlencoded", method: "form"},
		{contentType: "", method: "json"},
	}

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          cliStickyOptionalBodySpecWith(generations[0].contentType),
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()
			for i, gen := range generations {
				if i > 0 {
					cliStickyRegenerate(t, outputDir, cliStickyOptionalBodySpecWith(gen.contentType))
				}
				command := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go")
				require.Contains(t, command, `FlagName: "names", Shorthand: "n"`, "generation %d", i+1)
				require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`, "generation %d", i+1)
				usage := cliStickyReadFile(t, outputDir, "internal", "usage", "schema.go")
				require.Contains(t, usage, `-n --names <names>`, "generation %d", i+1)
				require.NotContains(t, usage, `-n --nick-name`, "generation %d", i+1)
				require.Equal(t, map[string]map[string]map[string]string{
					"CreatePet": {gen.method: {"names": "n"}},
				}, cliStickyLockFlags(t, outputDir), "generation %d", i+1)
			}
		},
	})
}

func TestSnapCLIStickyShorthandsSplitEndsWithoutShorthands(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: cliStickyContentTypeSpecWith(map[string][]string{
			"application/json":    {"names", "nick_name"},
			"multipart/form-data": {"names", "file"},
		}),
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()
			require.Equal(t, map[string]map[string]map[string]string{"CreatePet": {
				"json":      nil,
				"multipart": {"file": "f", "names": "n"},
			}}, cliStickyLockFlags(t, outputDir))

			cliStickyRegenerate(t, outputDir, cliStickyContentTypeSpecWith(map[string][]string{
				"application/json": {"names", "nick_name", "filter"},
			}))
			command := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go")
			require.Contains(t, command, `FlagName: "filter", Shorthand: "f"`)
			require.NotContains(t, command, `FlagName: "names", Shorthand:`)
			require.Equal(t, map[string]map[string]map[string]string{
				"CreatePet": {"json": {"filter": "f"}},
			}, cliStickyLockFlags(t, outputDir))
		},
	})
}

const cliStickyCollidingIDSpecHeader = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
%spaths:
`

const cliStickySplitPathTemplate = `  /pets:
    post:
      operationId: CreatePet
      description: Request content types render as separate commands; the multipart command's generated ID equals another operation's operationId.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
%s      responses:
        '200':
          description: OK
`

const cliStickySuffixedPathTemplate = `  /notes:
    post:
      operationId: CreatePet_multipart
      description: An operation whose operationId equals the generated ID of another operation's multipart command.
      tags: [%s]
      security: []
      requestBody:
        required: true
        content:
%s      responses:
        '200':
          description: OK
`

// cliStickyCollidingIDOrders places the suffixed operation before or after
// the split one, in both path and group order.
var cliStickyCollidingIDOrders = map[string]struct {
	suffixedFirst bool
	group         string
}{
	"suffixed operation first": {suffixedFirst: true, group: "animals"},
	"suffixed operation last":  {suffixedFirst: false, group: "notes"},
}

var cliStickySuffixedContent = map[string][]string{"application/json": {"names", "nick_name"}}

func cliStickyCollidingIDSpec(suffixedFirst bool, group, manifest string, contentTypes, suffixedContentTypes map[string][]string) string {
	split := fmt.Sprintf(cliStickySplitPathTemplate, cliStickyRequestContent(contentTypes))
	suffixed := fmt.Sprintf(cliStickySuffixedPathTemplate, group, cliStickyRequestContent(suffixedContentTypes))
	paths := split + suffixed
	if suffixedFirst {
		paths = suffixed + split
	}
	return fmt.Sprintf(cliStickyCollidingIDSpecHeader, manifest) + paths
}

func TestSnapCLIStickyShorthandsSplitOperationWithCollidingID(t *testing.T) {
	t.Parallel()

	type generation struct {
		contentTypes map[string][]string
		lock         map[string]map[string]string
	}
	sequences := map[string][]generation{
		"multipart command removed": {
			{
				contentTypes: map[string][]string{
					"application/json":    {"names", "nick_name"},
					"multipart/form-data": {"names"},
				},
				lock: map[string]map[string]string{"json": nil, "multipart": {"names": "n"}},
			},
			{
				contentTypes: map[string][]string{"application/json": {"names", "nick_name"}},
				lock:         nil,
			},
		},
		"JSON command added": {
			{
				contentTypes: map[string][]string{"multipart/form-data": {"names"}},
				lock:         map[string]map[string]string{"multipart": {"names": "n"}},
			},
			{
				contentTypes: map[string][]string{
					"application/json":    {"names", "nick_name"},
					"multipart/form-data": {"names"},
				},
				lock: map[string]map[string]string{"json": nil, "multipart": {"names": "n"}},
			},
		},
	}

	for orderName, order := range cliStickyCollidingIDOrders {
		for name, generations := range sequences {
			t.Run(orderName+" "+name, func(t *testing.T) {
				t.Parallel()
				spec := func(gen generation) string {
					return cliStickyCollidingIDSpec(order.suffixedFirst, order.group, "", gen.contentTypes, cliStickySuffixedContent)
				}
				shouldCompile := false
				snaptest.DoTestSnapshot(t, snaptest.Options{
					Spec:          spec(generations[0]),
					GenYaml:       cliStickyGenYaml,
					ShouldCompile: &shouldCompile,
					AfterGenerate: func(t *testing.T, outputDir string) {
						t.Helper()
						for i, gen := range generations {
							if i > 0 {
								cliStickyRegenerate(t, outputDir, spec(gen))
							}
							if _, ok := gen.contentTypes["application/json"]; ok {
								command := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go")
								require.NotContains(t, command, `FlagName: "names", Shorthand:`, "generation %d", i+1)
								require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`, "generation %d", i+1)
							}
							lock := cliStickyLockFlags(t, outputDir)
							require.Equal(t, gen.lock, lock["CreatePet"], "generation %d", i+1)
							require.NotContains(t, lock, "CreatePet_multipart", "generation %d", i+1)
						}
					},
				})
			})
		}
	}
}

const cliStickyCollidingIDPin = "x-speakeasy-cli-commands:\n  version: 1\n  operations:\n    CreatePet_multipart:\n      preferShorthand: [%s]\n"

func TestSnapCLIStickyShorthandsPinOnCollidingID(t *testing.T) {
	t.Parallel()

	contentTypes := func(multipart ...string) map[string][]string {
		return map[string][]string{
			"application/json":    {"bool"},
			"multipart/form-data": multipart,
		}
	}

	for orderName, order := range cliStickyCollidingIDOrders {
		t.Run(orderName+" pin applies only to the declared operation", func(t *testing.T) {
			t.Parallel()
			shouldCompile := false
			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec:          cliStickyCollidingIDSpec(order.suffixedFirst, order.group, "", contentTypes("nick_name"), cliStickySuffixedContent),
				GenYaml:       cliStickyGenYaml,
				ShouldCompile: &shouldCompile,
				AfterGenerate: func(t *testing.T, outputDir string) {
					t.Helper()
					require.Equal(t, map[string]string{"nick-name": "n"}, cliStickyLockFlags(t, outputDir)["CreatePet"]["multipart"])

					pinned := cliStickyCollidingIDSpec(order.suffixedFirst, order.group, fmt.Sprintf(cliStickyCollidingIDPin, "names"), contentTypes("names", "nick_name"), cliStickySuffixedContent)
					warnings := cliStickyRegenerate(t, outputDir, pinned)
					for _, warning := range warnings {
						require.NotContains(t, warning.Error(), `operation "CreatePet" (multipart request)`)
					}

					split := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpetmultipart.go")
					require.Contains(t, split, `FlagName: "nick-name", Shorthand: "n"`)
					require.NotContains(t, split, `FlagName: "names", Shorthand:`)
					suffixed := cliStickyReadFile(t, outputDir, "internal", "cli", order.group, "createpetmultipart.go")
					require.Contains(t, suffixed, `FlagName: "names", Shorthand: "n"`)
					require.NotContains(t, suffixed, `FlagName: "nick-name", Shorthand:`)

					splitUsage := cliStickyUsageCommand(t, outputDir, "pets create-pet-multipart")
					require.Contains(t, splitUsage, `-n --nick-name <nick_name>`)
					require.Contains(t, splitUsage, `--names <names>`)
					require.NotContains(t, splitUsage, `-n --names`)
					suffixedUsage := cliStickyUsageCommand(t, outputDir, order.group+" create-pet-multipart")
					require.Contains(t, suffixedUsage, `-n --names <names>`)
					require.Contains(t, suffixedUsage, `--nick-name <nick_name>`)
					require.NotContains(t, suffixedUsage, `-n --nick-name`)

					lock := cliStickyLockFlags(t, outputDir)
					require.Equal(t, map[string]string{"nick-name": "n"}, lock["CreatePet"]["multipart"])
					require.Equal(t, map[string]string{"names": "n"}, lock["CreatePet_multipart"]["json"])
				},
			})
		})

		t.Run(orderName+" pin on a flag of the other operation fails generation", func(t *testing.T) {
			t.Parallel()
			shouldCompile := false
			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec:          cliStickyCollidingIDSpec(order.suffixedFirst, order.group, fmt.Sprintf(cliStickyCollidingIDPin, "tag"), contentTypes("tag"), cliStickySuffixedContent),
				GenYaml:       cliStickyGenYaml,
				ShouldCompile: &shouldCompile,
				ExpectErrors: []string{
					`operation "CreatePet_multipart" preferShorthand "tag" does not match any generated flag on that operation`,
				},
			})
		})
	}
}

func TestSnapCLIStickyShorthandsCollidingIDSerializationChange(t *testing.T) {
	t.Parallel()

	splitContent := map[string][]string{
		"application/json":    {"bool"},
		"multipart/form-data": {"num"},
	}

	for orderName, order := range cliStickyCollidingIDOrders {
		t.Run(orderName, func(t *testing.T) {
			t.Parallel()
			shouldCompile := false
			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec:          cliStickyCollidingIDSpec(order.suffixedFirst, order.group, "", splitContent, map[string][]string{"application/json": {"names"}}),
				GenYaml:       cliStickyGenYaml,
				ShouldCompile: &shouldCompile,
				AfterGenerate: func(t *testing.T, outputDir string) {
					t.Helper()
					require.Equal(t, map[string]map[string]string{"json": {"names": "n"}}, cliStickyLockFlags(t, outputDir)["CreatePet_multipart"])

					cliStickyRegenerate(t, outputDir, cliStickyCollidingIDSpec(order.suffixedFirst, order.group, "", splitContent, map[string][]string{
						"application/x-www-form-urlencoded": {"names", "nick_name"},
					}))
					command := cliStickyReadFile(t, outputDir, "internal", "cli", order.group, "createpetmultipart.go")
					require.Contains(t, command, `FlagName: "names", Shorthand: "n"`)
					require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
					usage := cliStickyUsageCommand(t, outputDir, order.group+" create-pet-multipart")
					require.Contains(t, usage, `-n --names <names>`)
					require.NotContains(t, usage, `-n --nick-name`)
					require.Equal(t, map[string]map[string]string{"form": {"names": "n"}}, cliStickyLockFlags(t, outputDir)["CreatePet_multipart"])
				},
			})
		})
	}
}

func TestSnapCLIStickyShorthandsUsageGeneration(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:          cliStickyShorthandSpecWith(cliStickyNamesField),
		GenYaml:       cliStickyGenYaml,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()
			require.Equal(t, map[string]string{"names": "n"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"]["json"].Flags)

			cliStickyRegenerate(t, outputDir, cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyNickNameField, cliStickyTagField), generate.WithOutputUsageGroup("all"))
			command := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "createpet.go")
			require.Contains(t, command, `FlagName: "names", Shorthand: "n"`)
			require.Contains(t, command, `FlagName: "tag", Shorthand: "t"`)
			require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
			recorded := cliStickyReadLock(t, outputDir).CLIShorthands
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, recorded["CreatePet"]["json"].Flags)

			cliStickyRegenerate(t, outputDir, cliStickyShorthandSpecWith(cliStickyNickNameField), generate.WithUsageSnippetArgsGenerateAll())
			require.Equal(t, recorded, cliStickyReadLock(t, outputDir).CLIShorthands)
		},
	})
}
