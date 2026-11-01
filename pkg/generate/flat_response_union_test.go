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
  /note:
    get:
      operationId: getNote
      responses:
        "200":
          description: OK
          content:
            text/plain:
              schema:
                type: string
            text/markdown:
              schema:
                type:
                  - string
                  - "null"
  /level:
    get:
      operationId: getLevel
      responses:
        "200":
          description: OK
          content:
            text/plain:
              schema:
                type: string
            application/json:
              schema:
                type: string
                enum:
                  - low
                  - high
  /mixed:
    get:
      operationId: getMixed
      responses:
        "200":
          description: OK
          content:
            text/plain:
              schema:
                type: string
            text/csv:
              schema:
                type: string
            application/json:
              schema:
                type: object
                properties:
                  rows:
                    type: integer
  /thumbnail:
    get:
      operationId: getThumbnail
      responses:
        "200":
          description: OK
          content:
            image/png:
              schema:
                type: string
                format: binary
            image/jpeg:
              schema:
                type: string
                format: binary
  /token:
    get:
      operationId: getToken
      responses:
        "200":
          description: OK
          content:
            text/plain:
              schema:
                $ref: "#/components/schemas/Token"
            application/jwt:
              schema:
                $ref: "#/components/schemas/Token"
  /job:
    post:
      operationId: startJob
      responses:
        "200":
          description: Done
          content:
            text/plain:
              schema:
                type: string
        "202":
          description: Accepted
          content:
            text/plain:
              schema:
                type: string
  /receipt:
    get:
      operationId: getReceipt
      responses:
        "200":
          description: OK
          headers:
            X-Request-Id:
              schema:
                type: string
          content:
            text/plain:
              schema:
                type: string
            text/csv:
              schema:
                type: string
components:
  schemas:
    Token:
      type: string
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

type flatResponseShape struct {
	// scalar, when set, is the collapsed response type, with no union members
	scalar ast.DataType
	// members, when set, are the union members in order
	members []ast.DataType
	// itemType, when set, is the item type of a collapsed array response
	itemType ast.DataType
	// envelope, when true, expects a class whose Result field is the scalar
	envelope bool
}

func TestFlatResponseCollapsesEqualContentTypesForGo(t *testing.T) {
	t.Parallel()

	types := resolveFlatResponseTypes(t, "go")

	tests := map[string]flatResponseShape{
		// two type: string bodies collapse to the string itself
		"getImage": {scalar: ast.DataTypeString},
		// a format difference changes neither the Go type nor the union member name
		"getContact": {scalar: ast.DataTypeString},
		// nullability lives on the field, so a nullable string collapses too
		"getNote": {scalar: ast.DataTypeString},
		// two references to one scalar component collapse
		"getToken":     {scalar: ast.DataTypeString},
		"getThumbnail": {scalar: ast.DataTypeResponseStream},
		// the same scalar on two status codes collapses
		"startJob": {scalar: ast.DataTypeString},
		// headers keep the envelope, whose Result is the collapsed scalar
		"getReceipt": {scalar: ast.DataTypeString, envelope: true},
		"listTags":   {itemType: ast.DataTypeString},
		// two equal strings collapse into one member beside the object
		"getMixed": {members: []ast.DataType{ast.DataTypeString, ast.DataTypeClass}},
		// an inline enum is a named type, so it stays distinct from the plain string
		"getLevel":  {members: []ast.DataType{ast.DataTypeString, ast.DataTypeEnum}},
		"getReport": {members: []ast.DataType{ast.DataTypeString, ast.DataTypeClass}},
		// structurally equal inline objects per status code stay separate members
		"createWidget": {members: []ast.DataType{ast.DataTypeClass, ast.DataTypeClass}},
		"streamEvents": {members: []ast.DataType{ast.DataTypeJsonL, ast.DataTypeJsonL}},
		"listItems":    {members: []ast.DataType{ast.DataTypeArray, ast.DataTypeArray}},
	}
	for op, want := range tests {
		t.Run(op, func(t *testing.T) {
			t.Parallel()
			typ, ok := types[op]
			require.True(t, ok, "operation %s not resolved", op)
			require.NotNil(t, typ)

			if want.envelope {
				require.Equal(t, ast.DataTypeClass, typ.Type)
				var result *ast.FieldDef
				for _, field := range typ.Fields {
					if field.Name == "Result" {
						result = field
					}
				}
				require.NotNil(t, result, "fields: %v", typ.Fields)
				typ = result.Type
			}

			switch {
			case want.members != nil:
				require.Equal(t, ast.DataTypeUnion, typ.Type)
				require.Len(t, typ.AssociatedTypes, len(want.members))
				for i, member := range want.members {
					assert.Equal(t, member, typ.AssociatedTypes[i].Type, "member %d", i)
				}
			case want.itemType != "":
				require.Equal(t, ast.DataTypeArray, typ.Type)
				require.NotNil(t, typ.ItemType)
				assert.Equal(t, want.itemType, typ.ItemType.Type)
			default:
				assert.Equal(t, want.scalar, typ.Type)
				assert.Empty(t, typ.AssociatedTypes)
			}
		})
	}
}

func TestFlatResponseKeepsUnionForOtherTargets(t *testing.T) {
	t.Parallel()

	for _, target := range []string{"typescript", "python"} {
		t.Run(target, func(t *testing.T) {
			t.Parallel()
			types := resolveFlatResponseTypes(t, target)

			image := types["getImage"]
			require.NotNil(t, image)
			require.Equal(t, ast.DataTypeUnion, image.Type)
			assert.Len(t, image.AssociatedTypes, 2)
		})
	}
}
