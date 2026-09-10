package generate

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoOptionalMethodArgumentsDefault(t *testing.T) {
	t.Parallel()

	cfg, err := GetLanguageConfigDefaults("go", true)
	require.NoError(t, err)
	assert.Equal(t, "pointers", cfg.Cfg["optionalMethodArguments"])
}

func TestGoOptionalMethodArgumentsExtensionIgnoredForNonGoTarget(t *testing.T) {
	t.Parallel()

	spec := []byte(`openapi: 3.1.0
info:
  title: Go extension isolation
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      x-speakeasy-go-optional-method-arguments: method-options
      responses:
        "200":
          description: OK
`)

	_, astTree, err := TestResolveAST(TestResolveASTInput{
		OpenAPIContents: spec,
		Target:          "typescript",
	})
	require.NoError(t, err)

	var operation *ast.Operation
	err = astTree.WalkSDK(astTree.MainSDK, nil, func(node ast.Node, _ []ast.Node, _ *ast.AST) error {
		if ast.GetNodeType(node) == ast.NodeTypeOperation {
			op := node.(*ast.Operation)
			if op.ID == "listPets" {
				operation = op
			}
		}
		return nil
	}, map[string]bool{})
	require.NoError(t, err)
	require.NotNil(t, operation)
	assert.Nil(t, operation.Extensions.GoOptionalMethodArguments)
}
