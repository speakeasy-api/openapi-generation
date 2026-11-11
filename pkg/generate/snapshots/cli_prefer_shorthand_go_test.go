package snapshots

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

const cliPreferShorthandGenYaml = `cli:
  packageName: github.com/example/pet-cli
  cliName: pet-cli
  envVarPrefix: PET
`

const cliPreferShorthandBodySpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  operations:
    CreatePet:
      preferShorthand: [%s]
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Two optional body fields share an initial letter; preferShorthand keeps the first letter on the pinned flag.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                names:
                  type: array
                  items:
                    type: string
                nick_name:
                  type: string
      responses:
        '200':
          description: OK
`

const cliPreferShorthandIntentSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  categories: [Create]
  operations:
    CreatePet:
      preferShorthand: [sort]
  commands:
    create:
      category: Create
      summary: Create a pet
      op: CreatePet
      args:
        name:
          to: $.name
          variadic: true
          summary: Pet name
      flags:
        summary:
          to: $.summary
          shorthand: s
          summary: Pet summary
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Two query parameters share an initial letter; the pinned one keeps it, so an intent flag cannot reuse it.
      tags: [pets]
      security: []
      parameters:
        - name: sort
          in: query
          schema:
            type: string
        - name: size
          in: query
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
                summary:
                  type: string
      responses:
        '200':
          description: OK
`

const cliPreferShorthandSplitSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  operations:
    CreatePet:
      preferShorthand: [%s]
paths:
  /pets:
    post:
      operationId: CreatePet
      description: Request content types with different fields; a pin may name a flag that only one of them generates.
      tags: [pets]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                bark: {type: boolean}
                bool: {type: boolean}
          multipart/form-data:
            schema:
              type: object
              properties:
                num: {type: number}
      responses:
        '200':
          description: OK
`

func TestSnapCLIPreferShorthand(t *testing.T) {
	t.Parallel()

	t.Run("pinned flag keeps its shorthand", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          fmt.Sprintf(cliPreferShorthandBodySpec, "names"),
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			AfterGenerate: func(t *testing.T, outputDir string) {
				t.Helper()
				command, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "pets", "createpet.go"))
				require.NoError(t, err)
				require.Contains(t, string(command), `FlagName: "names", Shorthand: "n"`)
				require.Contains(t, string(command), `FlagName: "nick-name"`)
				require.NotContains(t, string(command), `FlagName: "nick-name", Shorthand:`)

				usage, err := os.ReadFile(filepath.Join(outputDir, "internal", "usage", "schema.go"))
				require.NoError(t, err)
				require.Contains(t, string(usage), `-n --names <names...>`)
				require.Contains(t, string(usage), `--nick-name <nick_name>`)
				require.NotContains(t, string(usage), `-n --nick-name`)
			},
		})
	})

	t.Run("unknown flag fails generation", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          fmt.Sprintf(cliPreferShorthandBodySpec, "missing"),
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			ExpectErrors: []string{
				`operation "CreatePet" preferShorthand "missing" does not match any generated flag on that operation`,
			},
		})
	})

	t.Run("intent shorthand collides with pinned flag", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          cliPreferShorthandIntentSpec,
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			ExpectErrors: []string{
				`declares shorthand "-s" on flag "--summary", which collides with the auto-assigned shorthand of generated flag "--sort"`,
			},
		})
	})

	t.Run("pin on a flag of one content-type variant", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          fmt.Sprintf(cliPreferShorthandSplitSpec, "bool"),
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			AfterGenerate: func(t *testing.T, outputDir string) {
				t.Helper()
				command, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "pets", "createpet.go"))
				require.NoError(t, err)
				require.Contains(t, string(command), `FlagName: "bool", Shorthand: "b"`)
				require.NotContains(t, string(command), `FlagName: "bark", Shorthand:`)

				multipart, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "pets", "createpetmultipart.go"))
				require.NoError(t, err)
				require.Contains(t, string(multipart), `FlagName: "num", Shorthand: "n"`)
			},
		})
	})

	t.Run("pin on no content-type variant fails generation", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          fmt.Sprintf(cliPreferShorthandSplitSpec, "missing"),
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			ExpectErrors: []string{
				`operation "CreatePet" preferShorthand "missing" does not match any generated flag on that operation`,
			},
		})
	})
}
