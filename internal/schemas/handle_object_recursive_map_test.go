package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func handleTestComponent(t *testing.T, schemasYAML, schemaName string) *ast.FieldDef {
	t.Helper()

	fullYAML := testutils.CreateOpenAPIDoc(schemasYAML)

	common, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
		OpenAPIYAML:  fullYAML,
		MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
	})
	require.NoError(t, err)

	schemas := &Schemas{
		Config:    common.Config,
		Target:    common.Target,
		Subsystem: common.Subsystem,
		Namer:     common.Namer,
	}

	testSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get(schemaName)
	require.True(t, exists)
	require.NotNil(t, testSchema)

	result, err := schemas.HandleSchema(context.Background(), CreateTestParams(testSchema, common.DocInfo))
	require.NoError(t, err)
	require.NotNil(t, result)

	return result
}

func TestHandleObject_SelfRecursiveAdditionalPropertiesMap(t *testing.T) {
	schemasYAML := `FieldMap:
  type: object
  additionalProperties:
    anyOf:
      - $ref: "#/components/schemas/TextField"
      - $ref: "#/components/schemas/ObjectField"
TextField:
  type: object
  required:
    - type
  properties:
    type:
      type: string
      enum:
        - text
    value:
      type: string
ObjectField:
  type: object
  required:
    - type
  properties:
    type:
      type: string
      enum:
        - object
    properties:
      $ref: "#/components/schemas/FieldMap"`

	result := handleTestComponent(t, schemasYAML, "FieldMap")

	require.Equal(t, ast.DataTypeMap, result.Type.Type)
	union := result.Type.ItemType
	require.NotNil(t, union)
	require.Equal(t, ast.DataTypeUnion, union.Type)
	assert.False(t, union.Truncated)
	require.Len(t, union.AssociatedTypes, 2)

	var objectField *ast.TypeDef
	for _, member := range union.AssociatedTypes {
		if member.Name == "ObjectField" {
			objectField = member
		}
	}
	require.NotNil(t, objectField)

	var propertiesField *ast.FieldDef
	for _, field := range objectField.Fields {
		if field.OriginalName == "properties" {
			propertiesField = field
		}
	}
	require.NotNil(t, propertiesField)

	require.Equal(t, ast.DataTypeMap, propertiesField.Type.Type)
	innerUnion := propertiesField.Type.ItemType
	require.NotNil(t, innerUnion)
	require.Equal(t, ast.DataTypeUnion, innerUnion.Type)
	assert.False(t, innerUnion.Truncated)
	assert.Len(t, innerUnion.AssociatedTypes, 2)
}

func TestHandleObject_PureSelfRecursiveAdditionalPropertiesMapTerminates(t *testing.T) {
	schemasYAML := `SelfMap:
  type: object
  additionalProperties:
    $ref: "#/components/schemas/SelfMap"`

	result := handleTestComponent(t, schemasYAML, "SelfMap")

	require.Equal(t, ast.DataTypeMap, result.Type.Type)
	require.NotNil(t, result.Type.ItemType)

	inner := result.Type.ItemType
	require.Equal(t, ast.DataTypeMap, inner.Type)
	require.NotNil(t, inner.ItemType)

	terminalMap := inner.ItemType
	require.Equal(t, ast.DataTypeMap, terminalMap.Type)
	require.NotNil(t, terminalMap.ItemType)

	terminal := terminalMap.ItemType
	assert.Equal(t, ast.DataTypeClass, terminal.Type)
	assert.Equal(t, "SelfMap", terminal.Name)
	assert.False(t, terminal.Truncated)
	assert.Empty(t, terminal.Fields)
}
