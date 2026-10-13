package snapshots

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

const cliArtifactFieldPointerSpec = `openapi: 3.0.3
info:
  title: Render API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-cli-commands:
  version: 1
  categories: [Create]
  commands:
    render:
      category: Create
      summary: Render an image asset to a file
      description: Render an image from a text prompt and write it to a file.
      op: renderAsset
      args:
        prompt:
          to: $.prompt
          variadic: true
          summary: Text prompt to render
      output:
        artifact:
          contentPointer: $.steps[*].image
          kind: image
          defaultPath: "render-{timestamp}-{rand}.{ext}"
paths:
  /render:
    post:
      operationId: renderAsset
      description: Each step carries its image block under a named field rather than inside a content array.
      tags: [assets]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [prompt]
              properties:
                prompt:
                  type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                  status:
                    type: string
                  steps:
                    type: array
                    items:
                      type: object
                      properties:
                        image:
                          type: object
                          required: [type]
                          properties:
                            type:
                              const: image
                            data:
                              type: string
                            mime_type:
                              type: string
`

func TestSnapCLIArtifactFieldPointerMockBody(t *testing.T) {
	t.Parallel()

	shouldCompile := false
	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec: cliArtifactFieldPointerSpec,
		GenYaml: `generation:
  tests:
    generateTests: true
cli:
  packageName: github.com/example/render-cli
  cliName: render-cli
  envVarPrefix: RENDER
`,
		ShouldCompile: &shouldCompile,
		AfterGenerate: func(t *testing.T, outputDir string) {
			t.Helper()
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, filepath.Join(outputDir, "tests", "artifact_test.go"), nil, 0)
			require.NoError(t, err)

			var body string
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Name.Name != "TestArtifact_WritesLastMatchingBlock" {
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					lit, ok := n.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING || body != "" {
						return true
					}
					body, err = strconv.Unquote(lit.Value)
					require.NoError(t, err)
					return false
				})
			}
			require.NotEmpty(t, body, "TestArtifact_WritesLastMatchingBlock must serve a mock body")
			require.JSONEq(t, `{
  "id": "task-123",
  "status": "completed",
  "steps": [
    {"image": {"type": "image", "data": "aW50ZXJtZWRpYXRl", "mime_type": "image/png"}},
    {"image": {"type": "image", "data": "YXJ0aWZhY3QtYnl0ZXM=", "mime_type": "image/png"}}
  ]
}`, body)
		},
	})
}
