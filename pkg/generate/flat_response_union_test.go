package generate

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const flatResponseUnionSpec = `openapi: 3.1.0
info:
  title: Flat response unions
  version: 1.0.0
paths:
  /image:
    get:
      operationId: getImage
      responses:
        "200":
          description: OK
          content:
            image/png:
              schema:
                type: string
            image/svg+xml:
              schema:
                type: string
  /report:
    get:
      operationId: getReport
      responses:
        "200":
          description: OK
          content:
            text/csv:
              schema:
                type: string
            application/json:
              schema:
                type: object
                properties:
                  rows:
                    type: integer
`

func resolveFlatResponseTypes(t *testing.T, target string) map[string]*ast.TypeDef {
	t.Helper()

	_, astTree, err := TestResolveAST(TestResolveASTInput{
		OpenAPIContents:  []byte(flatResponseUnionSpec),
		Target:           target,
		AdditionalConfig: &map[string]any{"responseFormat": "flat"},
	})
	require.NoError(t, err)

	types := map[string]*ast.TypeDef{}
	err = astTree.WalkSDK(astTree.MainSDK, nil, func(node ast.Node, _ []ast.Node, _ *ast.AST) error {
		if ast.GetNodeType(node) == ast.NodeTypeOperation {
			op := node.(*ast.Operation)
			types[op.ID] = op.Response.Type
		}
		return nil
	}, map[string]bool{})
	require.NoError(t, err)

	return types
}

func TestFlatResponseCollapsesEqualContentTypesForGo(t *testing.T) {
	t.Parallel()

	types := resolveFlatResponseTypes(t, "go")

	image := types["getImage"]
	require.NotNil(t, image)
	assert.Equal(t, ast.DataTypeString, image.Type)

	report := types["getReport"]
	require.NotNil(t, report)
	require.Equal(t, ast.DataTypeUnion, report.Type)
	assert.Len(t, report.AssociatedTypes, 2)
}

func TestFlatResponseKeepsUnionForOtherTargets(t *testing.T) {
	t.Parallel()

	types := resolveFlatResponseTypes(t, "typescript")

	image := types["getImage"]
	require.NotNil(t, image)
	require.Equal(t, ast.DataTypeUnion, image.Type)
	assert.Len(t, image.AssociatedTypes, 2)
}
