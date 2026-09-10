package snapshots

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/assert"
)

// Pagination on a polled operation is unsupported in Go: the polling
// helpers hand each attempt a request-scoped context, so a Next closure
// created there cannot outlive the call. The generator warns and drops the
// pagination extension, so the output carries the polling wrapper but no
// Next closure or field.
func TestSnapGoPollingDropsPagination(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Polling Pagination API
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - apiKey: []
paths:
  /jobs:
    get:
      operationId: listJobs
      description: |-
        Polled operation that also declares cursor pagination. Pagination is
        dropped for polled operations in Go with a generation warning.
      parameters:
        - name: cursor
          in: query
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
      x-speakeasy-pagination:
        type: cursor
        inputs:
          - name: cursor
            in: parameters
            type: cursor
          - name: limit
            in: parameters
            type: limit
        outputs:
          results: $.items
          nextCursor: $.nextCursor
      x-speakeasy-polling:
        - name: WaitForCompleted
          intervalSeconds: 1
          limitCount: 3
          successCriteria:
            - condition: $statusCode == 201
      responses:
        "200":
          description: Partial results
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Jobs"
        "201":
          description: Completed results
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Jobs"
components:
  securitySchemes:
    apiKey:
      type: apiKey
      in: header
      name: X-API-Key
  schemas:
    Jobs:
      type: object
      properties:
        items:
          type: array
          items:
            type: string
        nextCursor:
          type: string
`

	genYaml := `go:
  packageName: github.com/example/pollingpagination
`

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:    spec,
		GenYaml: genYaml,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			sdk := readGeneratedFile(t, tempDir, "sdk.go")
			assert.Contains(t, sdk, "polling.LimitCountError")
			assert.NotContains(t, sdk, "res.Next")

			response := readGeneratedFile(t, tempDir, "models", "operations", "listjobs.go")
			assert.NotContains(t, response, "Next func()")
		},
	})
}
