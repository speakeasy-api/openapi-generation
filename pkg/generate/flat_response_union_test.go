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
  /widgets:
    post:
      operationId: createWidget
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
        "201":
          description: Created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
  /events:
    get:
      operationId: streamEvents
      responses:
        "200":
          description: OK
          content:
            application/jsonl:
              schema:
                type: object
                properties:
                  id:
                    type: string
            application/x-ndjson:
              schema:
                type: object
                properties:
                  count:
                    type: integer
  /tags:
    get:
      operationId: listTags
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  type: string
            application/vnd.acme+json:
              schema:
                type: array
                items:
                  type: string
  /items:
    get:
      operationId: listItems
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
            application/vnd.acme+json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
  /contact:
    get:
      operationId: getContact
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: string
            application/vnd.acme+json:
              schema:
                type: string
                format: email
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

	widget := types["createWidget"]
	require.NotNil(t, widget)
	require.Equal(t, ast.DataTypeUnion, widget.Type)
	assert.Len(t, widget.AssociatedTypes, 2)

	events := types["streamEvents"]
	require.NotNil(t, events)
	require.Equal(t, ast.DataTypeUnion, events.Type)
	assert.Len(t, events.AssociatedTypes, 2)
	for _, member := range events.AssociatedTypes {
		assert.Equal(t, ast.DataTypeJsonL, member.Type)
	}

	tags := types["listTags"]
	require.NotNil(t, tags)
	require.Equal(t, ast.DataTypeArray, tags.Type)
	require.NotNil(t, tags.ItemType)
	assert.Equal(t, ast.DataTypeString, tags.ItemType.Type)

	items := types["listItems"]
	require.NotNil(t, items)
	require.Equal(t, ast.DataTypeUnion, items.Type)
	assert.Len(t, items.AssociatedTypes, 2)

	contact := types["getContact"]
	require.NotNil(t, contact)
	require.Equal(t, ast.DataTypeUnion, contact.Type)
	require.Len(t, contact.AssociatedTypes, 2)
	assert.NotEqual(t, contact.AssociatedTypes[0].Format, contact.AssociatedTypes[1].Format)
}

func TestFlatResponseKeepsUnionForOtherTargets(t *testing.T) {
	t.Parallel()

	types := resolveFlatResponseTypes(t, "typescript")

	image := types["getImage"]
	require.NotNil(t, image)
	require.Equal(t, ast.DataTypeUnion, image.Type)
	assert.Len(t, image.AssociatedTypes, 2)
}
