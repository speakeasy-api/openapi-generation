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

const cliPreferShorthandPathParamSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  operations:
    UpdatePet:
      preferShorthand: [price]
paths:
  /pets/{pet_id}:
    put:
      operationId: UpdatePet
      description: A pinned body field shares its initial with a required path parameter, which otherwise wins the letter.
      tags: [pets]
      security: []
      parameters:
        - name: pet_id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                price:
                  type: number
      responses:
        '200':
          description: OK
`

const cliPreferShorthandPaginatedSpec = `openapi: 3.0.3
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  operations:
    ListPets:
      preferShorthand: [after]
paths:
  /pets:
    get:
      operationId: ListPets
      description: A paginated operation reserves -a for --all, so a pin starting with a cannot take it.
      tags: [pets]
      security: []
      parameters:
        - name: after
          in: query
          schema:
            type: string
      x-speakeasy-pagination:
        type: cursor
        inputs:
          - name: after
            in: parameters
            type: cursor
        outputs:
          nextCursor: $.next
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  next:
                    type: string
                  pets:
                    type: array
                    items:
                      type: string
`

const cliPreferShorthandNameOverrideSpec = `openapi: 3.1.0
info:
  title: Pet API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  operations:
    foo:
      preferShorthand: [after]
paths:
  /pets:
    get:
      operationId: ListPets
      x-speakeasy-name-override: foo
      description: A paginated operation whose method name equals another operation's operationId.
      tags: [pets]
      security: []
      x-speakeasy-pagination:
        type: offsetLimit
        inputs:
          - {name: offset, in: parameters, type: offset}
        outputs:
          results: $.items
      parameters:
        - {name: offset, in: query, schema: {type: integer}}
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  items: {type: array, items: {type: string}}
  /notes:
    get:
      operationId: foo
      description: An operation without pagination that pins a flag starting with a.
      tags: [notes]
      security: []
      parameters:
        - {name: after, in: query, schema: {type: string}}
        - {name: around, in: query, schema: {type: string}}
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

	t.Run("pin beats a required path parameter on the same initial", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          cliPreferShorthandPathParamSpec,
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			AfterGenerate: func(t *testing.T, outputDir string) {
				t.Helper()
				command, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "pets", "updatepet.go"))
				require.NoError(t, err)
				require.Contains(t, string(command), `FlagName: "price", Shorthand: "p"`)
				require.Contains(t, string(command), `FlagName: "pet-id"`)
				require.NotContains(t, string(command), `FlagName: "pet-id", Shorthand:`)

				usage, err := os.ReadFile(filepath.Join(outputDir, "internal", "usage", "schema.go"))
				require.NoError(t, err)
				require.Contains(t, string(usage), `-p --price <price>`)
				require.NotContains(t, string(usage), `-p --pet-id`)
			},
		})
	})

	t.Run("pin applies to the operationId, not a name override matching it", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          cliPreferShorthandNameOverrideSpec,
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			AfterGenerate: func(t *testing.T, outputDir string) {
				t.Helper()
				command, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "notes", "foo.go"))
				require.NoError(t, err)
				require.Contains(t, string(command), `FlagName: "after", Shorthand: "a"`)
				require.NotContains(t, string(command), `FlagName: "around", Shorthand:`)
			},
		})
	})

	t.Run("pin starting with a on a paginated operation fails generation", func(t *testing.T) {
		t.Parallel()
		shouldCompile := false
		snaptest.DoTestSnapshot(t, snaptest.Options{
			Spec:          cliPreferShorthandPaginatedSpec,
			GenYaml:       cliPreferShorthandGenYaml,
			ShouldCompile: &shouldCompile,
			ExpectErrors: []string{
				`operation "ListPets" preferShorthand "after" cannot take -a`,
			},
		})
	})
}
