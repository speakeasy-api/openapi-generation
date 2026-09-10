package snapshots

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSnapMCPPaginatedResponseValidationKey guards against a regression where a
// paginated MCP operation generated with `responseFormat: flat` validated the
// raw response body against the wrong key.
//
// In flat mode a paginated response body is wrapped in a `{ Result: ... }`
// envelope, so the generated zod schema validates a `Result` property. The MCP
// response matcher must therefore place the parsed body under `Result`; if it
// instead keys by the body schema's own name the runtime validation fails with
// `Response validation failed ... path ["Result"]`. A non-paginated sibling has
// no envelope and must continue to key by the body name.
func TestSnapMCPPaginatedResponseValidationKey(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Pagination MCP API
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - apiKey: []
paths:
  /servers:
    get:
      operationId: serversList
      x-speakeasy-pagination:
        type: offsetLimit
        inputs:
          - name: page
            in: parameters
            type: page
        outputs:
          results: $.data
      parameters:
        - name: page
          in: query
          schema:
            type: integer
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ServerList"
  /virtual_machines:
    get:
      operationId: virtualMachinesList
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ServerList"
components:
  securitySchemes:
    apiKey:
      type: apiKey
      name: x-api-key
      in: header
  schemas:
    ServerList:
      type: object
      properties:
        data:
          type: array
          items:
            type: object
            properties:
              id:
                type: string
`

	genYaml := `mcp-typescript:
  packageName: pagination-mcp
  responseFormat: flat
  validateResponse: true
`

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:    spec,
		GenYaml: genYaml,
		// Assertions live in AfterGenerate; no file snapshot needed.
		Expected: "",
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			// Paginated op: response is wrapped in a `{ Result: ... }` envelope,
			// so the matcher must key by `Result`.
			paginated, err := os.ReadFile(
				filepath.Join(tempDir, "src/funcs/serversList.ts"),
			)
			require.NoError(t, err)
			model, err := os.ReadFile(
				filepath.Join(tempDir, "src/models/serverslistop.ts"),
			)
			require.NoError(t, err)
			assert.Contains(t, string(model), "Result:",
				"paginated response type should carry a Result envelope field")
			assert.Contains(t, string(paginated), `{ key: "Result" }`,
				"matcher key must match the Result envelope field validated by the schema")

			// Non-paginated sibling: no envelope, so the matcher keys by the
			// body schema name and must not gain a phantom Result key.
			plain, err := os.ReadFile(
				filepath.Join(tempDir, "src/funcs/virtualMachinesList.ts"),
			)
			require.NoError(t, err)
			assert.Contains(t, string(plain), `{ key: "ServerList" }`,
				"non-enveloped response should key by the body schema name")
			assert.NotContains(t, string(plain), `key: "Result"`,
				"non-enveloped response must not key by Result")
		},
	})
}
