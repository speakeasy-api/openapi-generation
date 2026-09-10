package snapshots

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

// TestSnapDocumentationNone_TS asserts that `generation.documentation: none`
// suppresses per-operation docs entirely: no `.md` (or mintlify `.mdx`) files
// are written under `docs/`, while the rest of the SDK still generates and
// compiles. This is the first-class opt-out — preferred over
// post-generation deletion.
func TestSnapDocumentationNone_TS(t *testing.T) {
	t.Parallel()
	spec := `openapi: 3.1.0
info:
  title: Doc None API
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /widgets/{id}:
    get:
      operationId: getWidget
      summary: Get a widget by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Widget'
components:
  schemas:
    Widget:
      type: object
      required: [id, name]
      properties:
        id:
          type: string
        name:
          type: string
`

	genYaml := `generation:
  documentation: none
typescript:
  packageName: docnonesnap
`

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		ExcludeGlobs: []string{"docs/**"},
	})
}
