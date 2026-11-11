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

const (
	cliStickyNamesField    = "                names:\n                  type: array\n                  items:\n                    type: string\n"
	cliStickyTagField      = "                tag:\n                  type: string\n"
	cliStickyNickNameField = "                nick_name:\n                  type: string\n"
)

type cliStickyShorthandLock struct {
	CLIShorthands map[string]struct {
		Flags   map[string]string `yaml:"flags"`
		Retired map[string]struct {
			Flag    string `yaml:"flag"`
			FreedIn string `yaml:"freedIn"`
		} `yaml:"retired"`
	} `yaml:"cliShorthands"`
}

func cliStickyShorthandSpecWith(fields ...string) string {
	return fmt.Sprintf(cliStickyShorthandSpec, strings.TrimRight(strings.Join(fields, ""), "\n"))
}

func TestSnapCLIStickyShorthands(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: cliStickyShorthandSpecWith(cliStickyNamesField, cliStickyTagField),
		GenYaml: `cli:
  packageName: github.com/example/pet-cli
  cliName: pet-cli
  envVarPrefix: PET
`,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()

			readCommand := func() string {
				t.Helper()
				content, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "pets", "createpet.go"))
				require.NoError(t, err)
				return string(content)
			}
			readUsage := func() string {
				t.Helper()
				content, err := os.ReadFile(filepath.Join(outputDir, "internal", "usage", "schema.go"))
				require.NoError(t, err)
				return string(content)
			}
			readLock := func() cliStickyShorthandLock {
				t.Helper()
				content, err := os.ReadFile(filepath.Join(outputDir, ".speakeasy", "gen.lock"))
				require.NoError(t, err)
				var lock cliStickyShorthandLock
				require.NoError(t, yaml.Unmarshal(content, &lock))
				return lock
			}
			regenerate := func(spec string) {
				t.Helper()
				generator, err := generate.New(generate.WithLogger(logging.NewLogger(zapcore.WarnLevel)))
				require.NoError(t, err)
				ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
				errs := generator.Generate(ctx, []byte(spec), "test-schema.yaml", "cli", outputDir, false, false)
				require.Empty(t, errs)
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
			require.NotEmpty(t, entry.Retired["n"].FreedIn)
		},
	})
}
