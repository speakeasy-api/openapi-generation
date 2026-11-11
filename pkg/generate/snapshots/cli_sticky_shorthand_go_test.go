package snapshots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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

type cliStickyShorthandLock struct {
	CLIShorthands map[string]struct {
		Flags   map[string]string `yaml:"flags"`
		Retired map[string]struct {
			Flag string `yaml:"flag"`
		} `yaml:"retired"`
	} `yaml:"cliShorthands"`
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

func cliStickyReadLock(t *testing.T, outputDir string) cliStickyShorthandLock {
	t.Helper()
	var lock cliStickyShorthandLock
	require.NoError(t, yaml.Unmarshal([]byte(cliStickyReadFile(t, outputDir, ".speakeasy", "gen.lock")), &lock))
	return lock
}

func cliStickyLockFlags(t *testing.T, outputDir string) map[string]map[string]string {
	t.Helper()
	flags := map[string]map[string]string{}
	for key, entry := range cliStickyReadLock(t, outputDir).CLIShorthands {
		require.Empty(t, entry.Retired, key)
		flags[key] = entry.Flags
	}
	return flags
}

func cliStickyDeleteRetired(t *testing.T, outputDir, operationID, letter string) {
	t.Helper()
	lockPath := filepath.Join(outputDir, ".speakeasy", "gen.lock")
	var lock map[string]any
	require.NoError(t, yaml.Unmarshal([]byte(cliStickyReadFile(t, outputDir, ".speakeasy", "gen.lock")), &lock))
	entry := lock["cliShorthands"].(map[string]any)[operationID].(map[string]any)
	retired := entry["retired"].(map[string]any)
	require.Contains(t, retired, letter)
	delete(retired, letter)
	content, err := yaml.Marshal(lock)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(lockPath, content, 0o644))
}

func cliStickyRegenerate(t *testing.T, outputDir, spec string) []error {
	t.Helper()
	generator, err := generate.New(generate.WithLogger(logging.NewLogger(zapcore.WarnLevel)))
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
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, readLock().CLIShorthands["CreatePet"].Flags)

			regenerate(cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyNickNameField, cliStickyTagField))
			command := readCommand()
			require.Contains(t, command, `FlagName: "names", Shorthand: "n"`)
			require.Contains(t, command, `FlagName: "nick-name"`)
			require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
			usage := readUsage()
			require.Contains(t, usage, `-n --names <names...>`)
			require.NotContains(t, usage, `-n --nick-name`)
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, readLock().CLIShorthands["CreatePet"].Flags)

			regenerate(cliStickyShorthandSpecWith(cliStickyNickNameField, cliStickyTagField))
			command = readCommand()
			require.Contains(t, command, `FlagName: "nick-name"`)
			require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
			entry := readLock().CLIShorthands["CreatePet"]
			require.Equal(t, map[string]string{"tag": "t"}, entry.Flags)
			require.Equal(t, "names", entry.Retired["n"].Flag)

			cliStickyDeleteRetired(t, outputDir, "CreatePet", "n")
			regenerate(cliStickyShorthandSpecWith(cliStickyNickNameField, cliStickyTagField))
			require.Contains(t, readCommand(), `FlagName: "nick-name", Shorthand: "n"`)
			require.Contains(t, readUsage(), `-n --nick-name <nick_name>`)
			entry = readLock().CLIShorthands["CreatePet"]
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
			wantLock := map[string]map[string]string{
				"CreatePet":           {"bool": "b", "num": "n"},
				"CreatePet_multipart": {"bool2": "b", "num2": "n"},
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
				require.Equal(t, map[string]string{"names": "n"}, cliStickyReadLock(t, outputDir).CLIShorthands["ListPets"].Flags)

				cliStickyRegenerate(t, outputDir, cliStickyListSpecWith(""))
				entry := cliStickyReadLock(t, outputDir).CLIShorthands["ListPets"]
				require.Empty(t, entry.Flags)
				require.Equal(t, "names", entry.Retired["n"].Flag)

				cliStickyRegenerate(t, outputDir, cliStickyListSpecWith("nick_name"))
				command := cliStickyReadFile(t, outputDir, "internal", "cli", "pets", "listpets.go")
				require.Contains(t, command, `FlagName: "nick-name"`)
				require.NotContains(t, command, `FlagName: "nick-name", Shorthand:`)
				entry = cliStickyReadLock(t, outputDir).CLIShorthands["ListPets"]
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
				require.Equal(t, map[string]string{"names": "n", "tag": "t"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"].Flags)

				cliStickyRegenerate(t, outputDir, cliStickyWholeBodySpec)
				entry := cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"]
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
			require.Equal(t, map[string]string{"names": "n"}, cliStickyReadLock(t, outputDir).CLIShorthands["constructor"].Flags)

			cliStickyRegenerate(t, outputDir, spec)
			lock := cliStickyReadLock(t, outputDir).CLIShorthands
			require.Equal(t, map[string]string{"names": "n"}, lock["constructor"].Flags)
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

func cliStickyContentTypeSpecWith(contentTypes map[string][]string) string {
	var b strings.Builder
	for _, contentType := range []string{"application/json", "multipart/form-data"} {
		fields, ok := contentTypes[contentType]
		if !ok {
			continue
		}
		b.WriteString("          " + contentType + ":\n            schema:\n              type: object\n              properties:\n")
		for _, field := range fields {
			b.WriteString("                " + field + ": {type: string}\n")
		}
	}
	return fmt.Sprintf(cliStickyContentTypeSpec, b.String())
}

func TestSnapCLIStickyShorthandsContentTypeChanges(t *testing.T) {
	t.Parallel()

	type generation struct {
		contentTypes map[string][]string
		usage        []string
		lock         map[string]map[string]string
	}
	sequences := map[string][]generation{
		"multipart operation gains and loses a JSON body": {
			{
				contentTypes: map[string][]string{"multipart/form-data": {"names", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]string{"CreatePet_multipart": {"names": "n", "tag": "t"}},
			},
			{
				contentTypes: map[string][]string{
					"application/json":    {"nova", "title"},
					"multipart/form-data": {"names", "nick_name", "tag"},
				},
				usage: []string{`-n --names <names>`, `-t --tag <tag>`, `-n --nova <nova>`, `-t --title <title>`},
				lock: map[string]map[string]string{
					"CreatePet":           {"nova": "n", "title": "t"},
					"CreatePet_multipart": {"names": "n", "tag": "t"},
				},
			},
			{
				contentTypes: map[string][]string{"multipart/form-data": {"names", "nick_name", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]string{"CreatePet_multipart": {"names": "n", "tag": "t"}},
			},
		},
		"JSON operation gains and loses a multipart body": {
			{
				contentTypes: map[string][]string{"application/json": {"names", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]string{"CreatePet": {"names": "n", "tag": "t"}},
			},
			{
				contentTypes: map[string][]string{
					"application/json":    {"names", "nick_name", "tag"},
					"multipart/form-data": {"nova", "title"},
				},
				usage: []string{`-n --names <names>`, `-t --tag <tag>`, `-n --nova <nova>`, `-t --title <title>`},
				lock: map[string]map[string]string{
					"CreatePet":           {"names": "n", "tag": "t"},
					"CreatePet_multipart": {"nova": "n", "title": "t"},
				},
			},
			{
				contentTypes: map[string][]string{"application/json": {"names", "nick_name", "tag"}},
				usage:        []string{`-n --names <names>`, `-t --tag <tag>`},
				lock:         map[string]map[string]string{"CreatePet": {"names": "n", "tag": "t"}},
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
			require.Equal(t, map[string]string{"nick-name": "n", "tag": "t"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"].Flags)

			require.Equal(t, 1, countWarnings(cliStickyRegenerate(t, outputDir, pinnedSpec)))
			require.Equal(t, map[string]string{"names": "n", "tag": "t"}, cliStickyReadLock(t, outputDir).CLIShorthands["CreatePet"].Flags)

			require.Zero(t, countWarnings(cliStickyRegenerate(t, outputDir, pinnedSpec)))
		},
	})
}
