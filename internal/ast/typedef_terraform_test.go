package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeDef_TerraformHoistAssociatedTypesCommonFields(t *testing.T) {
	t.Parallel()

	t.Run("hoists common primitive fields from union types", func(t *testing.T) {
		t.Parallel()

		// Create a union type with two associated types that have common fields
		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name: "Type1",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "commonString",
							Type: &TypeDef{Type: DataTypeString},
						},
						{
							Name: "commonInt",
							Type: &TypeDef{Type: DataTypeInteger},
						},
						{
							Name: "uniqueToType1",
							Type: &TypeDef{Type: DataTypeBoolean},
						},
					},
				},
				{
					Name: "Type2",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "commonString",
							Type: &TypeDef{Type: DataTypeString},
						},
						{
							Name: "commonInt",
							Type: &TypeDef{Type: DataTypeInteger},
						},
						{
							Name: "uniqueToType2",
							Type: &TypeDef{Type: DataTypeNumber},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check that common fields are hoisted to parent
		assert.Len(t, result.Fields, 2)

		// Fields should be sorted by name
		assert.Equal(t, "commonInt", result.Fields[0].Name)
		assert.Equal(t, "commonString", result.Fields[1].Name)

		// Check that TerraformAliasTo is set for matched fields in associated types
		for _, associatedType := range result.AssociatedTypes {
			commonStringField := associatedType.Fields.GetField("commonString")
			assert.NotNil(t, commonStringField)
			assert.NotNil(t, commonStringField.Type.Extensions)
			assert.NotNil(t, commonStringField.Type.Extensions.TerraformAliasTo)
			assert.Equal(t, "../commonString", *commonStringField.Type.Extensions.TerraformAliasTo)

			commonIntField := associatedType.Fields.GetField("commonInt")
			assert.NotNil(t, commonIntField)
			assert.NotNil(t, commonIntField.Type.Extensions)
			assert.NotNil(t, commonIntField.Type.Extensions.TerraformAliasTo)
			assert.Equal(t, "../commonInt", *commonIntField.Type.Extensions.TerraformAliasTo)
		}
	})

	t.Run("sets TerraformHoistedFrom on hoisted fields", func(t *testing.T) {
		t.Parallel()

		// Create a union type with two associated types that have common fields
		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name:         "Type1",
					OriginalName: "Type1",
					Type:         DataTypeClass,
					Fields: Fields{
						{
							Name: "commonField",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
				{
					Name:         "Type2",
					OriginalName: "Type2",
					Type:         DataTypeClass,
					Fields: Fields{
						{
							Name: "commonField",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check that TerraformHoistedFrom is set on the hoisted field
		require.Len(t, result.Fields, 1)
		hoistedField := result.Fields[0]
		require.NotNil(t, hoistedField.Type.Extensions)
		require.Len(t, hoistedField.Type.Extensions.TerraformHoistedFrom, 2)

		// Verify the hoisted sources contain correct information
		sources := hoistedField.Type.Extensions.TerraformHoistedFrom
		assert.Equal(t, "type1", sources[0].AssociatedTypeName)
		assert.Equal(t, "common_field", sources[0].FieldName)
		assert.Nil(t, sources[0].PathPrefix)
		assert.Equal(t, "type2", sources[1].AssociatedTypeName)
		assert.Equal(t, "common_field", sources[1].FieldName)
		assert.Nil(t, sources[1].PathPrefix)
	})

	t.Run("sets TerraformHoistedFrom with pathPrefix", func(t *testing.T) {
		t.Parallel()

		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name:         "Variant1",
					OriginalName: "Variant1",
					Type:         DataTypeClass,
					Fields: Fields{
						{
							Name: "description",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
				{
					Name:         "Variant2",
					OriginalName: "Variant2",
					Type:         DataTypeClass,
					Fields: Fields{
						{
							Name: "description",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		pathPrefix := []string{"parent", "config"}
		err := result.TerraformHoistAssociatedTypesCommonFields(pathPrefix)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check that TerraformHoistedFrom includes the pathPrefix
		require.Len(t, result.Fields, 1)
		hoistedField := result.Fields[0]
		require.NotNil(t, hoistedField.Type.Extensions)
		require.Len(t, hoistedField.Type.Extensions.TerraformHoistedFrom, 2)

		for _, source := range hoistedField.Type.Extensions.TerraformHoistedFrom {
			assert.Equal(t, pathPrefix, source.PathPrefix)
		}
	})

	t.Run("returns empty when no common fields", func(t *testing.T) {
		t.Parallel()

		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name: "Type1",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "field1",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
				{
					Name: "Type2",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "field2",
							Type: &TypeDef{Type: DataTypeInteger},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// No common fields, so nothing should be hoisted
		assert.Empty(t, result.Fields)
	})

	t.Run("skips non-primitive fields", func(t *testing.T) {
		t.Parallel()

		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name: "Type1",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "commonObject",
							Type: &TypeDef{Type: DataTypeClass},
						},
						{
							Name: "commonString",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
				{
					Name: "Type2",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "commonObject",
							Type: &TypeDef{Type: DataTypeClass},
						},
						{
							Name: "commonString",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Only primitive fields should be hoisted
		assert.Len(t, result.Fields, 1)
		assert.Equal(t, "commonString", result.Fields[0].Name)
	})

	t.Run("handles empty associated types", func(t *testing.T) {
		t.Parallel()

		unionType := &TypeDef{
			Name:            "TestUnion",
			Type:            DataTypeUnion,
			AssociatedTypes: []*TypeDef{},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// No associated types, nothing to hoist
		assert.Empty(t, result.Fields)
	})

	t.Run("handles nil TypeDef", func(t *testing.T) {
		t.Parallel()

		var nilType *TypeDef
		result := nilType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("merges enum values to superset when hoisting", func(t *testing.T) {
		t.Parallel()

		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name:         "VariantA",
					OriginalName: "VariantA",
					Type:         DataTypeClass,
					Fields: Fields{
						{
							Name: "state",
							Type: &TypeDef{
								Type: DataTypeEnum,
								Enum: &Enum{
									Type:   &TypeDef{Type: DataTypeString},
									Values: []string{"running", "stopped"},
									Names:  []string{"Running", "Stopped"},
								},
							},
						},
					},
				},
				{
					Name:         "VariantB",
					OriginalName: "VariantB",
					Type:         DataTypeClass,
					Fields: Fields{
						{
							Name: "state",
							Type: &TypeDef{
								Type: DataTypeEnum,
								Enum: &Enum{
									Type:   &TypeDef{Type: DataTypeString},
									Values: []string{"running", "stopped", "pending"},
									Names:  []string{"Running", "Stopped", "Pending"},
								},
							},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// The enum field should be hoisted
		require.Len(t, result.Fields, 1)
		hoistedField := result.Fields[0]
		assert.Equal(t, "state", hoistedField.Name)
		assert.Equal(t, DataTypeEnum, hoistedField.Type.Type)

		// The hoisted field's enum should contain the superset of all values
		require.NotNil(t, hoistedField.Type.Enum)
		assert.Equal(t, []string{"running", "stopped", "pending"}, hoistedField.Type.Enum.Values)
		assert.Equal(t, []string{"Running", "Stopped", "Pending"}, hoistedField.Type.Enum.Names)

		// TerraformHoistedFrom should be set
		require.NotNil(t, hoistedField.Type.Extensions)
		require.Len(t, hoistedField.Type.Extensions.TerraformHoistedFrom, 2)
	})

	t.Run("skips fields with different types", func(t *testing.T) {
		t.Parallel()

		unionType := &TypeDef{
			Name: "TestUnion",
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Name: "Type1",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "field1",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
				{
					Name: "Type2",
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "field1", // Same name but different type
							Type: &TypeDef{Type: DataTypeInteger},
						},
					},
				},
			},
		}

		result := unionType.Clone()
		err := result.TerraformHoistAssociatedTypesCommonFields(nil)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Fields with different types should not be hoisted
		assert.Empty(t, result.Fields)
	})
}

func TestTypeDef_TerraformHandleWrappedAttribute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		typeDef         *TypeDef
		entityName      string
		entityOperation string
		validate        func(t *testing.T, result *TypeDef)
	}{
		{
			name:            "nil TypeDef returns nil",
			typeDef:         nil,
			entityName:      "Entity",
			entityOperation: "create",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "class type without wrapped attribute extension returns unchanged",
			typeDef: &TypeDef{
				Name: "MyClass",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName:      "Entity",
			entityOperation: "create",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Equal(t, DataTypeClass, result.Type)
				assert.Equal(t, "MyClass", result.Name)
				assert.Len(t, result.Fields, 1)
				assert.Equal(t, "field1", result.Fields[0].Name)
			},
		},
		{
			name: "union type without wrapped attribute extension returns unchanged",
			typeDef: &TypeDef{
				Name: "MyUnion",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
				},
			},
			entityName:      "Entity",
			entityOperation: "create",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Equal(t, DataTypeUnion, result.Type)
				assert.Equal(t, "MyUnion", result.Name)
			},
		},
		{
			name: "primitive type gets wrapped with default 'data' field name",
			typeDef: &TypeDef{
				Name: "MyString",
				Type: DataTypeString,
			},
			entityName:      "Entity",
			entityOperation: "create",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Equal(t, DataTypeClass, result.Type)
				assert.Equal(t, "Entity", result.Name)
				assert.Len(t, result.Fields, 1)
				assert.Equal(t, "data", result.Fields[0].Name)
				assert.Equal(t, DataTypeString, result.Fields[0].Type.Type)
			},
		},
		{
			name: "array type gets wrapped with default 'data' field name",
			typeDef: &TypeDef{
				Name:     "MyArray",
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			entityName:      "Entity",
			entityOperation: "create",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Equal(t, DataTypeClass, result.Type)
				assert.Equal(t, "Entity", result.Name)
				assert.Len(t, result.Fields, 1)
				assert.Equal(t, "data", result.Fields[0].Name)
				assert.Equal(t, DataTypeArray, result.Fields[0].Type.Type)
			},
		},
		{
			name: "type with wrapped attribute extension uses custom field name",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					WrappedAttribute: ptr("customField"),
				},
			},
			entityName:      "Entity",
			entityOperation: "create",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Equal(t, DataTypeClass, result.Type)
				assert.Equal(t, "Entity", result.Name)
				assert.Len(t, result.Fields, 1)
				assert.Equal(t, "customField", result.Fields[0].Name)
				assert.Equal(t, DataTypeString, result.Fields[0].Type.Type)
			},
		},
		{
			name: "wrapped type includes entity extension and operation tag",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
			},
			entityName:      "TestEntity",
			entityOperation: "update",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				require.NotNil(t, result.Extensions.Entity)
				assert.Equal(t, []string{"TestEntity"}, result.Extensions.Entity.Names)
				require.NotNil(t, result.Extensions.All)
				assert.Equal(t, "data", result.Extensions.All["x-speakeasy-wrapped-update"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone to avoid mutation
			var testTypeDef *TypeDef
			if tt.typeDef != nil {
				testTypeDef = tt.typeDef.Clone()
			}

			result := testTypeDef.TerraformHandleWrappedAttribute(tt.entityName, tt.entityOperation)
			tt.validate(t, result)
		})
	}
}

func TestTypeDef_SetExtensionIfAbsentRecursive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeDef  *TypeDef
		tagName  string
		tagValue any
		validate func(t *testing.T, result *TypeDef)
	}{
		{
			name:     "nil TypeDef does nothing",
			typeDef:  nil,
			tagName:  "x-test",
			tagValue: "value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "tags simple type without extensions",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
			},
			tagName:  "x-test-tag",
			tagValue: "test-value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				require.NotNil(t, result.Extensions.All)
				assert.Equal(t, "test-value", result.Extensions.All["x-test-tag"])
			},
		},
		{
			name: "does not overwrite existing tag",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-existing": "original-value",
					},
				},
			},
			tagName:  "x-existing",
			tagValue: "new-value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				require.NotNil(t, result.Extensions.All)
				assert.Equal(t, "original-value", result.Extensions.All["x-existing"])
			},
		},
		{
			name: "recursively tags fields",
			typeDef: &TypeDef{
				Name: "MyClass",
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "field1",
						Type: &TypeDef{Type: DataTypeString},
					},
					{
						Name: "field2",
						Type: &TypeDef{Type: DataTypeInteger},
					},
				},
			},
			tagName:  "x-recursive",
			tagValue: true,
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				assert.Equal(t, true, result.Extensions.All["x-recursive"])

				for _, field := range result.Fields {
					require.NotNil(t, field.Type.Extensions)
					assert.Equal(t, true, field.Type.Extensions.All["x-recursive"])
				}
			},
		},
		{
			name: "recursively tags associated types",
			typeDef: &TypeDef{
				Name: "MyUnion",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			tagName:  "x-union-tag",
			tagValue: 42,
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				assert.Equal(t, 42, result.Extensions.All["x-union-tag"])

				for _, assocType := range result.AssociatedTypes {
					require.NotNil(t, assocType.Extensions)
					assert.Equal(t, 42, assocType.Extensions.All["x-union-tag"])
				}
			},
		},
		{
			name: "recursively tags item type",
			typeDef: &TypeDef{
				Name:     "MyArray",
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			tagName:  "x-array-tag",
			tagValue: "array-value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				assert.Equal(t, "array-value", result.Extensions.All["x-array-tag"])
				require.NotNil(t, result.ItemType.Extensions)
				assert.Equal(t, "array-value", result.ItemType.Extensions.All["x-array-tag"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone to avoid mutation
			var testTypeDef *TypeDef
			if tt.typeDef != nil {
				testTypeDef = tt.typeDef.Clone()
			}

			testTypeDef.SetExtensionIfAbsentRecursive(tt.tagName, tt.tagValue)
			tt.validate(t, testTypeDef)
		})
	}
}

func TestTypeDef_SetExtensionRecursive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeDef  *TypeDef
		tagName  string
		tagValue any
		validate func(t *testing.T, result *TypeDef)
	}{
		{
			name:     "nil TypeDef does nothing",
			typeDef:  nil,
			tagName:  "x-test",
			tagValue: "value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "tags simple type",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
			},
			tagName:  "x-test-tag",
			tagValue: "test-value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				require.NotNil(t, result.Extensions.All)
				assert.Equal(t, "test-value", result.Extensions.All["x-test-tag"])
			},
		},
		{
			name: "does not overwrite existing tag",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-existing": "original-value",
					},
				},
			},
			tagName:  "x-existing",
			tagValue: "new-value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				require.NotNil(t, result.Extensions.All)
				assert.Equal(t, "original-value", result.Extensions.All["x-existing"])
			},
		},
		{
			name: "special handling for x-speakeasy-ignore sets Ignore field",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeString,
			},
			tagName:  "x-speakeasy-ignore",
			tagValue: true,
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				require.NotNil(t, result.Extensions.Ignore)
				assert.True(t, *result.Extensions.Ignore)
				// Should not be in All map for x-speakeasy-ignore
				assert.Nil(t, result.Extensions.All["x-speakeasy-ignore"])
			},
		},
		{
			name: "recursively tags fields",
			typeDef: &TypeDef{
				Name: "MyClass",
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "field1",
						Type: &TypeDef{Type: DataTypeString},
					},
					{
						Name: "field2",
						Type: &TypeDef{Type: DataTypeInteger},
					},
				},
			},
			tagName:  "x-recursive",
			tagValue: true,
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				assert.Equal(t, true, result.Extensions.All["x-recursive"])

				for _, field := range result.Fields {
					require.NotNil(t, field.Type.Extensions)
					assert.Equal(t, true, field.Type.Extensions.All["x-recursive"])
				}
			},
		},
		{
			name: "recursively tags associated types",
			typeDef: &TypeDef{
				Name: "MyUnion",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			tagName:  "x-union-tag",
			tagValue: 42,
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				assert.Equal(t, 42, result.Extensions.All["x-union-tag"])

				for _, assocType := range result.AssociatedTypes {
					require.NotNil(t, assocType.Extensions)
					assert.Equal(t, 42, assocType.Extensions.All["x-union-tag"])
				}
			},
		},
		{
			name: "recursively tags item type",
			typeDef: &TypeDef{
				Name:     "MyArray",
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			tagName:  "x-array-tag",
			tagValue: "array-value",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result.Extensions)
				assert.Equal(t, "array-value", result.Extensions.All["x-array-tag"])
				require.NotNil(t, result.ItemType.Extensions)
				assert.Equal(t, "array-value", result.ItemType.Extensions.All["x-array-tag"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone to avoid mutation
			var testTypeDef *TypeDef
			if tt.typeDef != nil {
				testTypeDef = tt.typeDef.Clone()
			}

			testTypeDef.SetExtensionRecursive(tt.tagName, tt.tagValue)
			tt.validate(t, testTypeDef)
		})
	}
}

func TestTypeDef_TerraformHoistByEntityName(t *testing.T) {
	t.Parallel()

	wrappedAttributeVal := "data"

	tests := []struct {
		name       string
		typeDef    *TypeDef
		entityName string
		validate   func(t *testing.T, result *TypeDef, err error)
	}{
		{
			name:       "nil TypeDef returns nil",
			typeDef:    nil,
			entityName: "TestEntity",
			validate: func(t *testing.T, result *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Nil(t, result)
			},
		},
		{
			name: "TypeDef with matching entity returns unchanged",
			typeDef: &TypeDef{
				Name: "MyType",
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
			},
			entityName: "TestEntity",
			validate: func(t *testing.T, result *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "MyType", result.Name)
				assert.Equal(t, DataTypeClass, result.Type)
			},
		},
		{
			name: "hoists through array ItemType",
			typeDef: &TypeDef{
				Name: "MyArray",
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Name: "ItemType",
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						Entity: &extensions.Entity{Names: []string{"TestEntity"}},
					},
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			entityName: "TestEntity",
			validate: func(t *testing.T, result *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				// Should transform to the hoisted item type
				assert.Equal(t, DataTypeClass, result.Type)
				assert.Len(t, result.Fields, 1)
			},
		},
		{
			name: "hoists fields from nested structure",
			typeDef: &TypeDef{
				Name: "Wrapper",
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "nested",
						Type: &TypeDef{
							Name: "NestedType",
							Type: DataTypeClass,
							Extensions: &TypeDefExtensions{
								Entity: &extensions.Entity{Names: []string{"TestEntity"}},
								All: map[string]any{
									"x-speakeasy-wrapped-attribute": wrappedAttributeVal,
								},
								WrappedAttribute: &wrappedAttributeVal,
							},
							Fields: Fields{
								{Name: "entityField", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			entityName: "TestEntity",
			validate: func(t *testing.T, result *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				// Fields should be hoisted to top level
				assert.Len(t, result.Fields, 1)
				assert.Equal(t, "entityField", result.Fields[0].Name)
			},
		},
		{
			name: "preserves non-matching fields",
			typeDef: &TypeDef{
				Name: "Wrapper",
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "nonEntity",
						Type: &TypeDef{
							Name: "NonEntityType",
							Type: DataTypeString,
						},
					},
				},
			},
			entityName: "TestEntity",
			validate: func(t *testing.T, result *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				// Non-matching fields should be preserved
				assert.Len(t, result.Fields, 1)
				assert.Equal(t, "nonEntity", result.Fields[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone to avoid mutation
			var testTypeDef *TypeDef
			if tt.typeDef != nil {
				testTypeDef = tt.typeDef.Clone()
			}

			err := testTypeDef.TerraformHoistByEntityName(tt.entityName)
			tt.validate(t, testTypeDef, err)
		})
	}
}

func TestTypeDef_TerraformApplyEquivalentFieldProperties(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		target   *TypeDef
		source   *TypeDef
		validate func(t *testing.T, target *TypeDef)
	}{
		"nil receiver": {
			target: nil,
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.Nil(t, target)
			},
		},
		"nil source": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test", Type: &TypeDef{Type: DataTypeString}, Optional: false},
				},
			},
			source: nil,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.False(t, target.Fields[0].Optional)
			},
		},
		"type mismatch": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test", Type: &TypeDef{Type: DataTypeString}, Optional: false},
				},
			},
			source: &TypeDef{
				Type: DataTypeString,
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.False(t, target.Fields[0].Optional)
			},
		},
		"matching field copies Optional and Nullable": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}, Optional: false, Nullable: false},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}, Optional: true, Nullable: true},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.True(t, target.Fields[0].Optional)
				assert.True(t, target.Fields[0].Nullable)
			},
		},
		"matching field copies Type.Name": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "status", Type: &TypeDef{Type: DataTypeEnum, Name: "OldName"}},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "status", Type: &TypeDef{Type: DataTypeEnum, Name: "NewName"}},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.Equal(t, "NewName", target.Fields[0].Type.Name)
			},
		},
		"matching field merges Comments preferring source": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "my_field",
						Type: &TypeDef{Type: DataTypeString},
						Comments: &Comment{
							Description: "target description",
							Summary:     "target summary",
						},
					},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "my_field",
						Type: &TypeDef{Type: DataTypeString},
						Comments: &Comment{
							Description: "source description",
						},
					},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				require.NotNil(t, target.Fields[0].Comments)
				assert.Equal(t, "source description", target.Fields[0].Comments.Description)
				assert.Equal(t, "target summary", target.Fields[0].Comments.Summary)
			},
		},
		"nil target Comments adopts source Comments": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "my_field",
						Type: &TypeDef{Type: DataTypeString},
						Comments: &Comment{
							Description: "from source",
						},
					},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				require.NotNil(t, target.Fields[0].Comments)
				assert.Equal(t, "from source", target.Fields[0].Comments.Description)
			},
		},
		"no matching fields": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "alpha", Type: &TypeDef{Type: DataTypeString}, Optional: false},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "beta", Type: &TypeDef{Type: DataTypeString}, Optional: true},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.False(t, target.Fields[0].Optional)
			},
		},
		"recursive through nested class fields": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "inner", Type: &TypeDef{Type: DataTypeString}, Optional: false},
							},
						},
					},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "inner", Type: &TypeDef{Type: DataTypeString}, Optional: true},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.True(t, target.Fields[0].Type.Fields[0].Optional)
			},
		},
		"AssociatedTypes matching and recursion": {
			target: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "TypeA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field_a", Type: &TypeDef{Type: DataTypeString}, Optional: false},
						},
					},
				},
			},
			source: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "TypeA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field_a", Type: &TypeDef{Type: DataTypeString}, Optional: true},
						},
					},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.True(t, target.AssociatedTypes[0].Fields[0].Optional)
			},
		},
		"ItemType recursion": {
			target: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "item_field", Type: &TypeDef{Type: DataTypeString}, Nullable: false},
					},
				},
			},
			source: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "item_field", Type: &TypeDef{Type: DataTypeString}, Nullable: true},
					},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.True(t, target.ItemType.Fields[0].Nullable)
			},
		},
		"sanitized name matching": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test_field", Type: &TypeDef{Type: DataTypeString}, Optional: false},
				},
			},
			source: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "TestField", Type: &TypeDef{Type: DataTypeString}, Optional: true},
				},
			},
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.True(t, target.Fields[0].Optional)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.target.TerraformApplyEquivalentFieldProperties(tc.source)

			tc.validate(t, tc.target)
		})
	}
}

func TestTypeDef_SetExtensionOnEquivalentTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		target         *TypeDef
		other          *TypeDef
		extensionName  string
		extensionValue any
		validate       func(t *testing.T, target *TypeDef)
	}{
		"nil receiver": {
			target:         nil,
			other:          &TypeDef{Type: DataTypeClass},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.Nil(t, target)
			},
		},
		"nil other": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			other:          nil,
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				// Should not set any extension when other is nil
				assert.Nil(t, target.Extensions)
			},
		},
		"sets extension on root TypeDef": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			extensionName:  "x-speakeasy-param-computed",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				require.NotNil(t, target.Extensions)
				assert.Equal(t, true, target.Extensions.All["x-speakeasy-param-computed"])
			},
		},
		"does not overwrite existing extension": {
			target: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-test": "original",
					},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
			},
			extensionName:  "x-test",
			extensionValue: "new",
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.Equal(t, "original", target.Extensions.All["x-test"])
			},
		},
		"recurses through matching fields": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "inner", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "inner", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				// Root tagged
				require.NotNil(t, target.Extensions)
				assert.Equal(t, true, target.Extensions.All["x-test"])
				// Nested field.Type tagged
				nestedType := target.Fields[0].Type
				require.NotNil(t, nestedType.Extensions)
				assert.Equal(t, true, nestedType.Extensions.All["x-test"])
				// Leaf field.Type tagged
				leafType := nestedType.Fields[0].Type
				require.NotNil(t, leafType.Extensions)
				assert.Equal(t, true, leafType.Extensions.All["x-test"])
			},
		},
		"does not recurse into non-matching fields": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "alpha", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "beta", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				// Root is tagged
				require.NotNil(t, target.Extensions)
				assert.Equal(t, true, target.Extensions.All["x-test"])
				// Non-matching field's Type is NOT tagged
				assert.Nil(t, target.Fields[0].Type.Extensions)
			},
		},
		"recurses through matching AssociatedTypes": {
			target: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "TypeA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field_a", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			other: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "TypeA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field_a", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assocType := target.AssociatedTypes[0]
				require.NotNil(t, assocType.Extensions)
				assert.Equal(t, true, assocType.Extensions.All["x-test"])
			},
		},
		"recurses through matching ItemType": {
			target: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "item_field", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			other: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "item_field", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				require.NotNil(t, target.ItemType.Extensions)
				assert.Equal(t, true, target.ItemType.Extensions.All["x-test"])
			},
		},
		"sanitized name matching": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "TestField", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				// Root tagged
				require.NotNil(t, target.Extensions)
				assert.Equal(t, true, target.Extensions.All["x-test"])
				// Matching field's Type tagged via sanitized match
				require.NotNil(t, target.Fields[0].Type.Extensions)
				assert.Equal(t, true, target.Fields[0].Type.Extensions.All["x-test"])
			},
		},
		"sets extension value of false": {
			target: &TypeDef{
				Type: DataTypeClass,
			},
			other: &TypeDef{
				Type: DataTypeClass,
			},
			extensionName:  "x-speakeasy-param-readonly",
			extensionValue: false,
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				require.NotNil(t, target.Extensions)
				assert.Equal(t, false, target.Extensions.All["x-speakeasy-param-readonly"])
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.target.SetExtensionOnEquivalentTypes(tc.other, tc.extensionName, tc.extensionValue)

			tc.validate(t, tc.target)
		})
	}
}

func TestTypeDef_IsTerraformSubsetOf(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		a      *TypeDef
		b      *TypeDef
		expect bool
	}{
		"nil a": {
			a:      nil,
			b:      &TypeDef{Type: DataTypeClass},
			expect: true,
		},
		"nil b": {
			a:      &TypeDef{Type: DataTypeClass},
			b:      nil,
			expect: true,
		},
		"both nil": {
			a:      nil,
			b:      nil,
			expect: true,
		},
		"identical single field": {
			a: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expect: true,
		},
		"missing field in other": {
			a: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expect: false,
		},
		"extra field in other is still subset": {
			a: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			expect: true,
		},
		"nested type difference": {
			a: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "nested", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "inner1", Type: &TypeDef{Type: DataTypeString}},
							{Name: "inner2", Type: &TypeDef{Type: DataTypeInteger}},
						},
					}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "nested", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "inner1", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			expect: false,
		},
		"associated type match": {
			a: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Variant1",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			b: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Variant1",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expect: true,
		},
		"associated type not subset": {
			a: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Variant1",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
							{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
						},
					},
				},
			},
			b: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Variant1",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expect: false,
		},
		"missing associated type in other does not fail": {
			a: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Variant1",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			b: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Variant2",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expect: true,
		},
		"item type subset": {
			a: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			b: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			expect: true,
		},
		"item type not subset": {
			a: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			b: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			expect: false,
		},
		"match config path resolution": {
			a: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "alias_field",
						Type: &TypeDef{
							Type: DataTypeString,
							Extensions: &TypeDefExtensions{
								MatchConfig: &extensions.MatchConfig{
									Path: ptr("nested.original_field"),
								},
							},
						},
					},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "nested",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "original_field", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			expect: true,
		},
		"empty fields are subset": {
			a: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expect: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tc.a.IsTerraformSubsetOf(tc.b)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestTypeDef_IsTerraformSubsetOfEntity(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		a          *TypeDef
		b          *TypeDef
		entityName string
		expect     bool
	}{
		"both have entity - is subset": {
			a: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			entityName: "TestEntity",
			expect:     true,
		},
		"both have entity - not subset": {
			a: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "TestEntity",
			expect:     false,
		},
		"entity not found in a": {
			a: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "TestEntity",
			expect:     true,
		},
		"entity not found in b": {
			a: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"TestEntity"}},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			b: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "TestEntity",
			expect:     true,
		},
		"nil a": {
			a:          nil,
			b:          &TypeDef{Type: DataTypeClass},
			entityName: "TestEntity",
			expect:     true,
		},
		"nil b": {
			a:          &TypeDef{Type: DataTypeClass},
			b:          nil,
			entityName: "TestEntity",
			expect:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tc.a.IsTerraformSubsetOfEntity(tc.entityName, tc.b)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestTypeDef_SetExtensionOnExclusiveTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		receiver       *TypeDef
		include        *TypeDef
		exclude        *TypeDef
		extensionName  string
		extensionValue any
		validate       func(t *testing.T, receiver *TypeDef)
	}{
		"nil receiver": {
			receiver: nil,
			include: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			exclude:        nil,
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				assert.Nil(t, receiver)
			},
		},
		"nil include": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "test", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			include:        nil,
			exclude:        nil,
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				assert.Nil(t, receiver.Extensions)
				assert.Nil(t, receiver.Fields[0].Type.Extensions)
			},
		},
		"nil exclude tags subtree": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "my_field",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "nested", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			include: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			exclude:        nil,
			extensionName:  "x-speakeasy-param-readonly",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				// Receiver root should NOT be tagged (only the exclusive field's subtree)
				assert.Nil(t, receiver.Extensions)
				// Field's Type subtree should be tagged
				fieldType := receiver.Fields[0].Type
				require.NotNil(t, fieldType.Extensions)
				assert.Equal(t, true, fieldType.Extensions.All["x-speakeasy-param-readonly"])
				// Nested field should also be tagged (SetExtensionRecursive)
				nestedType := fieldType.Fields[0].Type
				require.NotNil(t, nestedType.Extensions)
				assert.Equal(t, true, nestedType.Extensions.All["x-speakeasy-param-readonly"])
			},
		},
		"field in include but not exclude tags subtree": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "exclusive_field", Type: &TypeDef{Type: DataTypeString}},
					{Name: "shared_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			include: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "exclusive_field", Type: &TypeDef{Type: DataTypeString}},
					{Name: "shared_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			exclude: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "shared_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				// exclusive_field's Type should be tagged
				require.NotNil(t, receiver.Fields[0].Type.Extensions)
				assert.Equal(t, true, receiver.Fields[0].Type.Extensions.All["x-test"])
				// shared_field's Type should NOT be tagged (it's in exclude, leaf level)
				assert.Nil(t, receiver.Fields[1].Type.Extensions)
			},
		},
		"field in all three recurses to find nested exclusive": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "exclusive_inner", Type: &TypeDef{Type: DataTypeString}},
								{Name: "shared_inner", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			include: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "exclusive_inner", Type: &TypeDef{Type: DataTypeString}},
								{Name: "shared_inner", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			exclude: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "outer",
						Type: &TypeDef{
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "shared_inner", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				outerType := receiver.Fields[0].Type
				// outer itself should NOT be tagged (it's shared, recursion happened)
				assert.Nil(t, outerType.Extensions)
				// exclusive_inner should be tagged
				require.NotNil(t, outerType.Fields[0].Type.Extensions)
				assert.Equal(t, true, outerType.Fields[0].Type.Extensions.All["x-test"])
				// shared_inner should NOT be tagged
				assert.Nil(t, outerType.Fields[1].Type.Extensions)
			},
		},
		"field in include but not in receiver is ignored": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "other", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			include: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "missing_from_receiver", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			exclude:        nil,
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				assert.Nil(t, receiver.Extensions)
				assert.Nil(t, receiver.Fields[0].Type.Extensions)
			},
		},
		"AssociatedTypes exclusive tagging": {
			receiver: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "ExclusiveType",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "field_a", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			include: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "ExclusiveType",
						Type: DataTypeClass,
					},
				},
			},
			exclude: &TypeDef{
				Type:            DataTypeUnion,
				AssociatedTypes: []*TypeDef{},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				assocType := receiver.AssociatedTypes[0]
				require.NotNil(t, assocType.Extensions)
				assert.Equal(t, true, assocType.Extensions.All["x-test"])
				// Nested field should also be tagged (SetExtensionRecursive)
				require.NotNil(t, assocType.Fields[0].Type.Extensions)
				assert.Equal(t, true, assocType.Fields[0].Type.Extensions.All["x-test"])
			},
		},
		"ItemType recursion": {
			receiver: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "exclusive_item", Type: &TypeDef{Type: DataTypeString}},
						{Name: "shared_item", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			include: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "exclusive_item", Type: &TypeDef{Type: DataTypeString}},
						{Name: "shared_item", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			exclude: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "shared_item", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			extensionName:  "x-test",
			extensionValue: true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				itemType := receiver.ItemType
				// exclusive_item should be tagged
				require.NotNil(t, itemType.Fields[0].Type.Extensions)
				assert.Equal(t, true, itemType.Fields[0].Type.Extensions.All["x-test"])
				// shared_item should NOT be tagged
				assert.Nil(t, itemType.Fields[1].Type.Extensions)
			},
		},
		"does not overwrite existing extension in subtree": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name: "my_field",
						Type: &TypeDef{
							Type: DataTypeString,
							Extensions: &TypeDefExtensions{
								All: map[string]any{
									"x-test": "original",
								},
							},
						},
					},
				},
			},
			include: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			exclude:        nil,
			extensionName:  "x-test",
			extensionValue: "new",
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				// SetExtensionRecursive preserves existing values
				assert.Equal(t, "original", receiver.Fields[0].Type.Extensions.All["x-test"])
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.receiver.SetExtensionOnExclusiveTypes(tc.include, tc.exclude, tc.extensionName, tc.extensionValue)

			tc.validate(t, tc.receiver)
		})
	}
}

func TestTypeDef_RemoveExtensionOnEquivalentTypes(t *testing.T) {
	t.Parallel()

	boolPtr := func(b bool) *bool { return &b }

	testCases := map[string]struct {
		target        *TypeDef
		other         *TypeDef
		extensionName string
		validate      func(t *testing.T, target *TypeDef)
	}{
		"nil receiver": {
			target:        nil,
			other:         &TypeDef{Type: DataTypeClass},
			extensionName: "x-test",
			validate:      func(t *testing.T, target *TypeDef) { t.Helper() },
		},
		"nil other": {
			target: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-test": true},
				},
			},
			other:         nil,
			extensionName: "x-test",
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				assert.Equal(t, true, target.Extensions.All["x-test"])
			},
		},
		"removes extension from root when equivalent exists": {
			target: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-test": true},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			extensionName: "x-test",
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				_, exists := target.Extensions.All["x-test"]
				assert.False(t, exists)
			},
		},
		"removes extension recursively on matching fields": {
			target: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{
						Type: DataTypeClass,
						Extensions: &TypeDefExtensions{
							All: map[string]any{"x-test": true},
						},
						Fields: Fields{
							{Name: "nested", Type: &TypeDef{
								Type: DataTypeString,
								Extensions: &TypeDefExtensions{
									All: map[string]any{"x-test": true},
								},
							}},
						},
					}},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "nested", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			extensionName: "x-test",
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				nestedType := target.Fields[0].Type
				_, exists := nestedType.Extensions.All["x-test"]
				assert.False(t, exists)

				leafType := nestedType.Fields[0].Type
				_, leafExists := leafType.Extensions.All["x-test"]
				assert.False(t, leafExists)
			},
		},
		"handles x-speakeasy-ignore by setting Ignore to false": {
			target: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All:    map[string]any{},
					Ignore: boolPtr(true),
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			other: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			extensionName: "x-speakeasy-ignore",
			validate: func(t *testing.T, target *TypeDef) {
				t.Helper()
				require.NotNil(t, target.Extensions.Ignore)
				assert.False(t, *target.Extensions.Ignore)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.target.RemoveExtensionOnEquivalentTypes(tc.other, tc.extensionName)

			tc.validate(t, tc.target)
		})
	}
}

func TestTypeDef_SetTerraformIgnoreOnExclusiveTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		receiver  *TypeDef
		whenIn    *TypeDef
		butNotIn  *TypeDef
		dataModel bool
		schema    bool
		validate  func(t *testing.T, receiver *TypeDef)
	}{
		"nil receiver": {
			receiver:  nil,
			whenIn:    &TypeDef{Type: DataTypeClass},
			butNotIn:  nil,
			dataModel: true,
			schema:    true,
			validate:  func(t *testing.T, receiver *TypeDef) { t.Helper() },
		},
		"sets TerraformIgnore on exclusive field": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "shared", Type: &TypeDef{Type: DataTypeString}},
					{Name: "exclusive", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			whenIn: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "shared", Type: &TypeDef{Type: DataTypeString}},
					{Name: "exclusive", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			butNotIn: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "shared", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			dataModel: false,
			schema:    true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				sharedField := receiver.Fields[0]
				assert.Nil(t, sharedField.Type.Extensions)

				exclusiveField := receiver.Fields[1]
				require.NotNil(t, exclusiveField.Type.Extensions)
				require.NotNil(t, exclusiveField.Type.Extensions.TerraformIgnore)
				assert.False(t, exclusiveField.Type.Extensions.TerraformIgnore.DataModel)
				assert.True(t, exclusiveField.Type.Extensions.TerraformIgnore.Schema)
			},
		},
		"recursively sets TerraformIgnore on exclusive subtree": {
			receiver: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "exclusive", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "child", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			whenIn: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "exclusive", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "child", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			butNotIn: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{},
			},
			dataModel: true,
			schema:    true,
			validate: func(t *testing.T, receiver *TypeDef) {
				t.Helper()
				exclusiveType := receiver.Fields[0].Type
				require.NotNil(t, exclusiveType.Extensions)
				require.NotNil(t, exclusiveType.Extensions.TerraformIgnore)
				assert.True(t, exclusiveType.Extensions.TerraformIgnore.DataModel)

				childType := exclusiveType.Fields[0].Type
				require.NotNil(t, childType.Extensions)
				require.NotNil(t, childType.Extensions.TerraformIgnore)
				assert.True(t, childType.Extensions.TerraformIgnore.DataModel)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.receiver.SetTerraformIgnoreOnExclusiveTypes(tc.whenIn, tc.butNotIn, tc.dataModel, tc.schema)

			tc.validate(t, tc.receiver)
		})
	}
}

func TestTypeDef_RenameExtensionRecursive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typeDef  *TypeDef
		fromName string
		toName   string
		validate func(t *testing.T, result *TypeDef)
	}{
		"nil receiver": {
			typeDef:  nil,
			fromName: "x-old",
			toName:   "x-new",
			validate: func(t *testing.T, result *TypeDef) { t.Helper() },
		},
		"renames extension at root": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-old": "value"},
				},
			},
			fromName: "x-old",
			toName:   "x-new",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				_, oldExists := result.Extensions.All["x-old"]
				assert.False(t, oldExists)
				assert.Equal(t, "value", result.Extensions.All["x-new"])
			},
		},
		"renames extension recursively in fields": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-old": true},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{
						Type: DataTypeString,
						Extensions: &TypeDefExtensions{
							All: map[string]any{"x-old": true},
						},
					}},
				},
			},
			fromName: "x-old",
			toName:   "x-new",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				_, rootOldExists := result.Extensions.All["x-old"]
				assert.False(t, rootOldExists)
				assert.Equal(t, true, result.Extensions.All["x-new"])

				fieldType := result.Fields[0].Type
				_, fieldOldExists := fieldType.Extensions.All["x-old"]
				assert.False(t, fieldOldExists)
				assert.Equal(t, true, fieldType.Extensions.All["x-new"])
			},
		},
		"no-op when extension does not exist": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-other": true},
				},
			},
			fromName: "x-old",
			toName:   "x-new",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Equal(t, true, result.Extensions.All["x-other"])
				_, newExists := result.Extensions.All["x-new"]
				assert.False(t, newExists)
			},
		},
		"skips rename when value is false": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-old": false},
				},
			},
			fromName: "x-old",
			toName:   "x-new",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				// false values should NOT be renamed (they serve as negative
				// markers that block subsequent merge operations)
				assert.Equal(t, false, result.Extensions.All["x-old"])
				_, newExists := result.Extensions.All["x-new"]
				assert.False(t, newExists)
			},
		},
		"skips rename when value is nil": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-old": nil},
				},
			},
			fromName: "x-old",
			toName:   "x-new",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				_, oldExists := result.Extensions.All["x-old"]
				assert.True(t, oldExists)
				_, newExists := result.Extensions.All["x-new"]
				assert.False(t, newExists)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.typeDef.RenameExtensionRecursive(tc.fromName, tc.toName)

			tc.validate(t, tc.typeDef)
		})
	}
}

func TestTypeDef_RemoveExtensionRecursive(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typeDef       *TypeDef
		extensionName string
		validate      func(t *testing.T, result *TypeDef)
	}{
		"nil receiver": {
			typeDef:       nil,
			extensionName: "x-test",
			validate:      func(t *testing.T, result *TypeDef) { t.Helper() },
		},
		"removes extension from root": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-test": true, "x-keep": "value"},
				},
			},
			extensionName: "x-test",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				_, exists := result.Extensions.All["x-test"]
				assert.False(t, exists)
				assert.Equal(t, "value", result.Extensions.All["x-keep"])
			},
		},
		"removes extension recursively": {
			typeDef: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-test": true},
				},
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{
						Type: DataTypeClass,
						Extensions: &TypeDefExtensions{
							All: map[string]any{"x-test": true},
						},
						Fields: Fields{
							{Name: "nested", Type: &TypeDef{
								Type: DataTypeString,
								Extensions: &TypeDefExtensions{
									All: map[string]any{"x-test": true},
								},
							}},
						},
					}},
				},
			},
			extensionName: "x-test",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				_, rootExists := result.Extensions.All["x-test"]
				assert.False(t, rootExists)

				_, fieldExists := result.Fields[0].Type.Extensions.All["x-test"]
				assert.False(t, fieldExists)

				_, nestedExists := result.Fields[0].Type.Fields[0].Type.Extensions.All["x-test"]
				assert.False(t, nestedExists)
			},
		},
		"removes from associated types and item type": {
			typeDef: &TypeDef{
				Type: DataTypeUnion,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-test": true},
				},
				AssociatedTypes: []*TypeDef{
					{
						Type: DataTypeString,
						Extensions: &TypeDefExtensions{
							All: map[string]any{"x-test": true},
						},
					},
				},
				ItemType: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-test": true},
					},
				},
			},
			extensionName: "x-test",
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				_, rootExists := result.Extensions.All["x-test"]
				assert.False(t, rootExists)

				_, assocExists := result.AssociatedTypes[0].Extensions.All["x-test"]
				assert.False(t, assocExists)

				_, itemExists := result.ItemType.Extensions.All["x-test"]
				assert.False(t, itemExists)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tc.typeDef.RemoveExtensionRecursive(tc.extensionName)

			tc.validate(t, tc.typeDef)
		})
	}
}

func TestTypeDef_TerraformDeleteIgnored(t *testing.T) {
	t.Parallel()

	boolPtr := func(b bool) *bool { return &b }

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var td *TypeDef
		td.TerraformDeleteIgnored() // should not panic
	})

	t.Run("removes fields with Const", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Fields: Fields{
				{Name: "kept", Type: &TypeDef{Type: DataTypeString}},
				{Name: "removed", Type: &TypeDef{Type: DataTypeString}, Const: &AnyValue{Value: "constant"}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.Fields, 1)
		assert.Equal(t, "kept", td.Fields[0].Name)
	})

	t.Run("removes fields with TerraformIgnore DataModel", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Fields: Fields{
				{Name: "kept", Type: &TypeDef{Type: DataTypeString}},
				{Name: "removed", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						TerraformIgnore: &extensions.TerraformIgnore{DataModel: true},
					},
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.Fields, 1)
		assert.Equal(t, "kept", td.Fields[0].Name)
	})

	t.Run("removes fields with Ignore true", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Fields: Fields{
				{Name: "kept", Type: &TypeDef{Type: DataTypeString}},
				{Name: "removed", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						Ignore: boolPtr(true),
					},
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.Fields, 1)
		assert.Equal(t, "kept", td.Fields[0].Name)
	})

	t.Run("keeps fields with Ignore false", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Fields: Fields{
				{Name: "kept1", Type: &TypeDef{Type: DataTypeString}},
				{Name: "kept2", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						Ignore: boolPtr(false),
					},
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.Fields, 2)
	})

	t.Run("removes associated types with TerraformIgnore DataModel", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			AssociatedTypes: []*TypeDef{
				{Name: "Kept", Type: DataTypeClass},
				{Name: "Removed", Type: DataTypeClass, Extensions: &TypeDefExtensions{
					TerraformIgnore: &extensions.TerraformIgnore{DataModel: true},
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.AssociatedTypes, 1)
		assert.Equal(t, "Kept", td.AssociatedTypes[0].Name)
	})

	t.Run("removes associated types with Ignore true", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			AssociatedTypes: []*TypeDef{
				{Name: "Kept", Type: DataTypeClass},
				{Name: "Removed", Type: DataTypeClass, Extensions: &TypeDefExtensions{
					Ignore: boolPtr(true),
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.AssociatedTypes, 1)
		assert.Equal(t, "Kept", td.AssociatedTypes[0].Name)
	})

	t.Run("recurses into child fields", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Fields: Fields{
				{Name: "parent", Type: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "child_kept", Type: &TypeDef{Type: DataTypeString}},
						{Name: "child_removed", Type: &TypeDef{
							Type: DataTypeString,
							Extensions: &TypeDefExtensions{
								TerraformIgnore: &extensions.TerraformIgnore{DataModel: true},
							},
						}},
					},
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.Fields, 1)
		require.Len(t, td.Fields[0].Type.Fields, 1)
		assert.Equal(t, "child_kept", td.Fields[0].Type.Fields[0].Name)
	})

	t.Run("recurses into associated types", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			AssociatedTypes: []*TypeDef{
				{
					Name: "KeptType",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "nested_kept", Type: &TypeDef{Type: DataTypeString}},
						{Name: "nested_removed", Type: &TypeDef{
							Type: DataTypeString,
							Extensions: &TypeDefExtensions{
								Ignore: boolPtr(true),
							},
						}},
					},
				},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.AssociatedTypes, 1)
		require.Len(t, td.AssociatedTypes[0].Fields, 1)
		assert.Equal(t, "nested_kept", td.AssociatedTypes[0].Fields[0].Name)
	})

	t.Run("recurses into ItemType", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeArray,
			ItemType: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "item_kept", Type: &TypeDef{Type: DataTypeString}},
					{Name: "item_removed", Type: &TypeDef{
						Type: DataTypeString,
						Extensions: &TypeDefExtensions{
							TerraformIgnore: &extensions.TerraformIgnore{DataModel: true},
						},
					}},
				},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.ItemType.Fields, 1)
		assert.Equal(t, "item_kept", td.ItemType.Fields[0].Name)
	})

	t.Run("does not remove fields with TerraformIgnore Schema only", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Fields: Fields{
				{Name: "kept", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						TerraformIgnore: &extensions.TerraformIgnore{Schema: true},
					},
				}},
			},
		}

		td.TerraformDeleteIgnored()

		require.Len(t, td.Fields, 1)
		assert.Equal(t, "kept", td.Fields[0].Name)
	})
}

func TestTypeDef_TerraformMergeAliasNodes(t *testing.T) {
	t.Parallel()

	strPtr := func(s string) *string { return &s }

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var td *TypeDef
		err := td.TerraformMergeAliasNodes("Entity", &TypeDef{})
		assert.NoError(t, err)
	})

	t.Run("nil other", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{Name: "Target"}
		err := td.TerraformMergeAliasNodes("Entity", nil)
		assert.NoError(t, err)
	})

	t.Run("merges only alias node fields", func(t *testing.T) {
		t.Parallel()

		target := &TypeDef{
			Name: "Target",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "existing", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		other := &TypeDef{
			Name: "Source",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "alias_field", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{Path: strPtr("/some/path")},
					},
				}},
				{Name: "non_alias_field", Type: &TypeDef{
					Type: DataTypeString,
				}},
			},
		}

		err := target.TerraformMergeAliasNodes("Target", other)
		require.NoError(t, err)

		// Should have existing + alias_field, but not non_alias_field
		assert.Len(t, target.Fields, 2)

		fieldNames := make([]string, len(target.Fields))
		for i, f := range target.Fields {
			fieldNames[i] = f.Name
		}
		assert.Contains(t, fieldNames, "existing")
		assert.Contains(t, fieldNames, "alias_field")
	})

	t.Run("filters associated types to alias nodes only", func(t *testing.T) {
		t.Parallel()

		target := &TypeDef{
			Name: "Target",
			Type: DataTypeClass,
		}

		other := &TypeDef{
			Name: "Source",
			Type: DataTypeClass,
			AssociatedTypes: []*TypeDef{
				{
					Name: "AliasType",
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{Path: strPtr("/alias/path")},
					},
				},
				{
					Name: "NonAliasType",
					Type: DataTypeClass,
				},
			},
		}

		err := target.TerraformMergeAliasNodes("Target", other)
		require.NoError(t, err)

		assert.Len(t, target.AssociatedTypes, 1)
		assert.Equal(t, "AliasType", target.AssociatedTypes[0].Name)
	})

	t.Run("does not mutate the original other", func(t *testing.T) {
		t.Parallel()

		target := &TypeDef{
			Name: "Target",
			Type: DataTypeClass,
		}

		other := &TypeDef{
			Name: "Source",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "alias_field", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{Path: strPtr("/path")},
					},
				}},
				{Name: "non_alias", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		err := target.TerraformMergeAliasNodes("Target", other)
		require.NoError(t, err)

		// Original other should still have both fields
		assert.Len(t, other.Fields, 2)
	})

	t.Run("filters out fields with empty path", func(t *testing.T) {
		t.Parallel()

		target := &TypeDef{
			Name: "Target",
			Type: DataTypeClass,
		}

		other := &TypeDef{
			Name: "Source",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "empty_path", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{Path: strPtr("")},
					},
				}},
				{Name: "valid_path", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						MatchConfig: &extensions.MatchConfig{Path: strPtr("/valid")},
					},
				}},
			},
		}

		err := target.TerraformMergeAliasNodes("Target", other)
		require.NoError(t, err)

		assert.Len(t, target.Fields, 1)
		assert.Equal(t, "valid_path", target.Fields[0].Name)
	})
}

func TestTypeDef_TerraformSetConflictsWith(t *testing.T) {
	t.Parallel()

	nameFunc := func(parent *TypeDef, associated *TypeDef) string {
		if associated.OriginalName != "" {
			return associated.OriginalName
		}
		return associated.Name
	}

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var td *TypeDef
		td.TerraformSetConflictsWith(nameFunc) // should not panic
	})

	t.Run("sets conflicts-with on union branches", func(t *testing.T) {
		t.Parallel()

		branchA := &TypeDef{
			Name: "BranchA",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}
		branchB := &TypeDef{
			Name: "BranchB",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}
		branchC := &TypeDef{
			Name: "BranchC",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}

		root := &TypeDef{
			Type: DataTypeUnion,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			AssociatedTypes: []*TypeDef{branchA, branchB, branchC},
		}

		root.TerraformSetConflictsWith(nameFunc)

		// BranchA should conflict with BranchB and BranchC
		conflictsA := branchA.Extensions.All["x-speakeasy-conflicts-with"].([]any)
		assert.Equal(t, []any{"BranchB", "BranchC"}, conflictsA)

		// BranchB should conflict with BranchA and BranchC
		conflictsB := branchB.Extensions.All["x-speakeasy-conflicts-with"].([]any)
		assert.Equal(t, []any{"BranchA", "BranchC"}, conflictsB)

		// Check computed-override and suppress-computed-diff defaults
		assert.Equal(t, false, branchA.Extensions.All["x-speakeasy-param-computed-override"])
		assert.Equal(t, false, branchA.Extensions.All["x-speakeasy-param-suppress-computed-diff"])
	})

	t.Run("preserves existing suppress-computed-diff", func(t *testing.T) {
		t.Parallel()

		branchA := &TypeDef{
			Name: "A",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-param-suppress-computed-diff": true,
				},
			},
		}
		branchB := &TypeDef{
			Name: "B",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}

		root := &TypeDef{
			Type: DataTypeUnion,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			AssociatedTypes: []*TypeDef{branchA, branchB},
		}

		root.TerraformSetConflictsWith(nameFunc)

		// Existing true value should be preserved
		assert.Equal(t, true, branchA.Extensions.All["x-speakeasy-param-suppress-computed-diff"])
		// Default should be false
		assert.Equal(t, false, branchB.Extensions.All["x-speakeasy-param-suppress-computed-diff"])
	})

	t.Run("recurses into fields", func(t *testing.T) {
		t.Parallel()

		innerBranchA := &TypeDef{
			Name: "InnerA",
			Type: DataTypeString,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}
		innerBranchB := &TypeDef{
			Name: "InnerB",
			Type: DataTypeString,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "field1", Type: &TypeDef{
					Type: DataTypeUnion,
					Extensions: &TypeDefExtensions{
						All: make(map[string]any),
					},
					AssociatedTypes: []*TypeDef{innerBranchA, innerBranchB},
				}},
			},
		}

		root.TerraformSetConflictsWith(nameFunc)

		conflictsA := innerBranchA.Extensions.All["x-speakeasy-conflicts-with"].([]any)
		assert.Equal(t, []any{"InnerB"}, conflictsA)
	})

	t.Run("overwrites existing conflicts-with value", func(t *testing.T) {
		t.Parallel()

		branchA := &TypeDef{
			Name: "A",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-conflicts-with": "existing_value",
				},
			},
		}
		branchB := &TypeDef{
			Name: "B",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}

		root := &TypeDef{
			Type: DataTypeUnion,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			AssociatedTypes: []*TypeDef{branchA, branchB},
		}

		root.TerraformSetConflictsWith(nameFunc)

		// Should overwrite existing value with new conflicts
		conflictsA := branchA.Extensions.All["x-speakeasy-conflicts-with"].([]any)
		assert.Equal(t, []any{"B"}, conflictsA)
	})
}

func TestTypeDef_PropagateExtensionValueRecursive(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var td *TypeDef
		td.PropagateExtensionValueRecursive("x-test") // should not panic
	})

	t.Run("propagates value to descendants", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{"x-tag": "hello"},
			},
			Fields: Fields{
				{Name: "child", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						All: make(map[string]any),
					},
				}},
			},
		}

		root.PropagateExtensionValueRecursive("x-tag")

		assert.Equal(t, "hello", root.Fields[0].Type.Extensions.All["x-tag"])
	})

	t.Run("does not overwrite existing values", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{"x-tag": "parent"},
			},
			Fields: Fields{
				{Name: "child", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-tag": "child-value"},
					},
				}},
			},
		}

		root.PropagateExtensionValueRecursive("x-tag")

		// Child's existing value should be preserved
		assert.Equal(t, "child-value", root.Fields[0].Type.Extensions.All["x-tag"])
	})

	t.Run("child value propagates further down", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "mid", Type: &TypeDef{
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-tag": "mid-value"},
					},
					Fields: Fields{
						{Name: "leaf", Type: &TypeDef{
							Type: DataTypeString,
							Extensions: &TypeDefExtensions{
								All: make(map[string]any),
							},
						}},
					},
				}},
			},
		}

		root.PropagateExtensionValueRecursive("x-tag")

		// Root has no value, so nothing propagated to root
		_, rootExists := root.Extensions.All["x-tag"]
		assert.False(t, rootExists)

		// Mid has value, propagated to leaf
		assert.Equal(t, "mid-value", root.Fields[0].Type.Fields[0].Type.Extensions.All["x-tag"])
	})

	t.Run("does not propagate when no value exists", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "child", Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						All: make(map[string]any),
					},
				}},
			},
		}

		root.PropagateExtensionValueRecursive("x-tag")

		_, exists := root.Fields[0].Type.Extensions.All["x-tag"]
		assert.False(t, exists)
	})
}

func TestTypeDef_TerraformPropagateComputedParent(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var td *TypeDef
		td.TerraformPropagateComputedParent() // should not panic
	})

	t.Run("tags optional computed container", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "container", Optional: true, Type: &TypeDef{
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-computed": true},
					},
				}},
			},
		}

		root.TerraformPropagateComputedParent()

		containerExt := root.Fields[0].Type.Extensions.All
		assert.Equal(t, true, containerExt["x-speakeasy-parent-require-to-not-null"])
		assert.Equal(t, true, containerExt["x-speakeasy-param-computed"])
	})

	t.Run("tags optional readonly container", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "container", Nullable: true, Type: &TypeDef{
					Type: DataTypeMap,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-readonly": true},
					},
				}},
			},
		}

		root.TerraformPropagateComputedParent()

		containerExt := root.Fields[0].Type.Extensions.All
		assert.Equal(t, true, containerExt["x-speakeasy-parent-require-to-not-null"])
		assert.Equal(t, true, containerExt["x-speakeasy-param-computed"])
	})

	t.Run("does not tag non-optional container", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "required", Optional: false, Type: &TypeDef{
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-computed": true},
					},
				}},
			},
		}

		root.TerraformPropagateComputedParent()

		containerExt := root.Fields[0].Type.Extensions.All
		_, exists := containerExt["x-speakeasy-parent-require-to-not-null"]
		assert.False(t, exists)
	})

	t.Run("does not tag primitive even if optional and computed", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "prim", Optional: true, Type: &TypeDef{
					Type: DataTypeString,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-computed": true},
					},
				}},
			},
		}

		root.TerraformPropagateComputedParent()

		primExt := root.Fields[0].Type.Extensions.All
		_, exists := primExt["x-speakeasy-parent-require-to-not-null"]
		assert.False(t, exists)
	})

	t.Run("associated types are treated as optional", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeUnion,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			AssociatedTypes: []*TypeDef{
				{
					Type: DataTypeArray,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-readonly": true},
					},
				},
			},
		}

		root.TerraformPropagateComputedParent()

		atExt := root.AssociatedTypes[0].Extensions.All
		assert.Equal(t, true, atExt["x-speakeasy-parent-require-to-not-null"])
		assert.Equal(t, true, atExt["x-speakeasy-param-computed"])
	})

	t.Run("recursively tags children of optional computed container", func(t *testing.T) {
		t.Parallel()

		childField := &TypeDef{
			Type: DataTypeString,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				{Name: "container", Optional: true, Type: &TypeDef{
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-computed": true},
					},
					Fields: Fields{
						{Name: "child", Type: childField},
					},
				}},
			},
		}

		root.TerraformPropagateComputedParent()

		// Container itself should be tagged
		containerExt := root.Fields[0].Type.Extensions.All
		assert.Equal(t, true, containerExt["x-speakeasy-parent-require-to-not-null"])
		assert.Equal(t, true, containerExt["x-speakeasy-param-computed"])

		// Child field inside the container should also be tagged (propagated via walk context)
		childExt := childField.Extensions.All
		assert.Equal(t, true, childExt["x-speakeasy-parent-require-to-not-null"])
		assert.Equal(t, true, childExt["x-speakeasy-param-computed"])
	})

	t.Run("shared pointer is not contaminated across paths", func(t *testing.T) {
		t.Parallel()

		// sharedChild is referenced by two different fields:
		// one through a computed+optional container, one through a non-computed path.
		sharedChild := &TypeDef{
			Type: DataTypeString,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
		}

		root := &TypeDef{
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: make(map[string]any),
			},
			Fields: Fields{
				// Non-computed path references sharedChild first (by field order)
				{Name: "non_computed", Optional: true, Type: &TypeDef{
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						All: make(map[string]any),
					},
					Fields: Fields{
						{Name: "leaf", Type: sharedChild},
					},
				}},
				// Computed path references the same sharedChild
				{Name: "computed_container", Optional: true, Type: &TypeDef{
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-param-computed": true},
					},
					Fields: Fields{
						{Name: "leaf", Type: sharedChild},
					},
				}},
			},
		}

		root.TerraformPropagateComputedParent()

		// The shared child should NOT be tagged with computed because it was
		// first visited through the non-computed path. The visited set prevents
		// re-processing from the computed path.
		_, hasComputed := sharedChild.Extensions.All["x-speakeasy-param-computed"]
		assert.False(t, hasComputed, "shared node should not be contaminated with computed from another path")

		_, hasRequireNotNull := sharedChild.Extensions.All["x-speakeasy-parent-require-to-not-null"]
		assert.False(t, hasRequireNotNull, "shared node should not be contaminated with require-to-not-null from another path")
	})
}

func TestTypeDef_RemoveEmptyObjectsRecursive(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var td *TypeDef
		td.RemoveEmptyObjectsRecursive() // should not panic
	})

	t.Run("removes empty class fields", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "empty", Type: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{},
				}},
				{Name: "keep", Type: &TypeDef{
					Type: DataTypeString,
				}},
			},
		}

		root.RemoveEmptyObjectsRecursive()

		assert.Len(t, root.Fields, 1)
		assert.Equal(t, "keep", root.Fields[0].Name)
	})

	t.Run("keeps class with fields", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "has_fields", Type: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "child", Type: &TypeDef{Type: DataTypeString}},
					},
				}},
			},
		}

		root.RemoveEmptyObjectsRecursive()

		assert.Len(t, root.Fields, 1)
		assert.Equal(t, "has_fields", root.Fields[0].Name)
	})

	t.Run("keeps class with associated types", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "has_assoc", Type: &TypeDef{
					Type:            DataTypeClass,
					AssociatedTypes: []*TypeDef{{Type: DataTypeString}},
				}},
			},
		}

		root.RemoveEmptyObjectsRecursive()

		assert.Len(t, root.Fields, 1)
		assert.Equal(t, "has_assoc", root.Fields[0].Name)
	})

	t.Run("recurses into nested fields", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "parent", Type: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "nested_empty", Type: &TypeDef{
							Type:   DataTypeClass,
							Fields: Fields{},
						}},
						{Name: "nested_keep", Type: &TypeDef{
							Type: DataTypeString,
						}},
					},
				}},
			},
		}

		root.RemoveEmptyObjectsRecursive()

		assert.Len(t, root.Fields, 1)
		assert.Len(t, root.Fields[0].Type.Fields, 1)
		assert.Equal(t, "nested_keep", root.Fields[0].Type.Fields[0].Name)
	})
}

func TestTypeDef_TerraformImportRequiredFields(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var td *TypeDef
		td.TerraformImportRequiredFields() // should not panic
	})

	t.Run("filters optional and nullable fields", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "required_field", Type: &TypeDef{Type: DataTypeString}},
				{Name: "optional_field", Type: &TypeDef{Type: DataTypeString}, Optional: true},
				{Name: "nullable_field", Type: &TypeDef{Type: DataTypeString}, Nullable: true},
				{Name: "also_required", Type: &TypeDef{Type: DataTypeInteger}},
			},
		}

		td.TerraformImportRequiredFields()

		require.Len(t, td.Fields, 2)
		assert.Equal(t, "also_required", td.Fields[0].Name)
		assert.Equal(t, "required_field", td.Fields[1].Name)
	})

	t.Run("keeps annotation-required fields", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{
					Name:     "annotated",
					Type:     &TypeDef{Type: DataTypeString},
					Optional: true,
					Annotations: Annotations{
						&ParamAnnotation{
							RequiredForOperation: true,
						},
					},
				},
				{Name: "optional_field", Type: &TypeDef{Type: DataTypeString}, Optional: true},
			},
		}

		td.TerraformImportRequiredFields()

		require.Len(t, td.Fields, 1)
		assert.Equal(t, "annotated", td.Fields[0].Name)
	})

	t.Run("sorts remaining fields by name", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "zebra", Type: &TypeDef{Type: DataTypeString}},
				{Name: "alpha", Type: &TypeDef{Type: DataTypeString}},
				{Name: "middle", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		td.TerraformImportRequiredFields()

		require.Len(t, td.Fields, 3)
		assert.Equal(t, "alpha", td.Fields[0].Name)
		assert.Equal(t, "middle", td.Fields[1].Name)
		assert.Equal(t, "zebra", td.Fields[2].Name)
	})

	t.Run("recurses into nested types", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{
					Name: "nested",
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "inner_required", Type: &TypeDef{Type: DataTypeString}},
							{Name: "inner_optional", Type: &TypeDef{Type: DataTypeString}, Optional: true},
						},
					},
				},
			},
		}

		td.TerraformImportRequiredFields()

		require.Len(t, td.Fields, 1)
		require.Len(t, td.Fields[0].Type.Fields, 1)
		assert.Equal(t, "inner_required", td.Fields[0].Type.Fields[0].Name)
	})

	t.Run("recurses into associated types", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeUnion,
			AssociatedTypes: []*TypeDef{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "required_in_assoc", Type: &TypeDef{Type: DataTypeString}},
						{Name: "optional_in_assoc", Type: &TypeDef{Type: DataTypeString}, Optional: true},
					},
				},
			},
		}

		td.TerraformImportRequiredFields()

		require.Len(t, td.AssociatedTypes, 1)
		require.Len(t, td.AssociatedTypes[0].Fields, 1)
		assert.Equal(t, "required_in_assoc", td.AssociatedTypes[0].Fields[0].Name)
	})

	t.Run("recurses into item type", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeArray,
			ItemType: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "item_required", Type: &TypeDef{Type: DataTypeString}},
					{Name: "item_optional", Type: &TypeDef{Type: DataTypeString}, Optional: true},
				},
			},
		}

		td.TerraformImportRequiredFields()

		require.NotNil(t, td.ItemType)
		require.Len(t, td.ItemType.Fields, 1)
		assert.Equal(t, "item_required", td.ItemType.Fields[0].Name)
	})
}

func TestTypeDef_TerraformInvalidImportTypes(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var td *TypeDef
		result := td.TerraformInvalidImportTypes()
		assert.Nil(t, result)
	})

	t.Run("all valid types returns nil", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "str_field", Type: &TypeDef{Type: DataTypeString}},
				{Name: "bool_field", Type: &TypeDef{Type: DataTypeBoolean}},
				{Name: "int_field", Type: &TypeDef{Type: DataTypeInteger}},
				{Name: "num_field", Type: &TypeDef{Type: DataTypeNumber}},
				{Name: "enum_field", Type: &TypeDef{Type: DataTypeEnum}},
			},
		}

		result := td.TerraformInvalidImportTypes()
		assert.Nil(t, result)
	})

	t.Run("invalid type at root", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeArray,
		}

		result := td.TerraformInvalidImportTypes()
		require.Len(t, result, 1)
		assert.Empty(t, result[0].Hierarchy)
		assert.Equal(t, string(DataTypeArray), result[0].TypeName)
	})

	t.Run("invalid nested type", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "str_field", Type: &TypeDef{Type: DataTypeString}},
				{
					Name: "nested",
					Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "map_field", Type: &TypeDef{Type: DataTypeMap}},
						},
					},
				},
			},
		}

		result := td.TerraformInvalidImportTypes()
		require.Len(t, result, 1)
		assert.Equal(t, "."+SanitizeFieldName("nested")+"."+SanitizeFieldName("map_field"), result[0].Hierarchy)
		assert.Equal(t, string(DataTypeMap), result[0].TypeName)
	})

	t.Run("multiple invalid types", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "arr_field", Type: &TypeDef{Type: DataTypeArray}},
				{Name: "set_field", Type: &TypeDef{Type: DataTypeSet}},
			},
		}

		result := td.TerraformInvalidImportTypes()
		require.Len(t, result, 2)
	})
}

func TestTypeDef_TerraformHasInvalidImportTypes(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()

		var td *TypeDef
		assert.False(t, td.TerraformHasInvalidImportTypes())
	})

	t.Run("all valid types", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "str_field", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		assert.False(t, td.TerraformHasInvalidImportTypes())
	})

	t.Run("has invalid type", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "arr_field", Type: &TypeDef{Type: DataTypeArray}},
			},
		}

		assert.True(t, td.TerraformHasInvalidImportTypes())
	})
}

func TestTypeDef_SortFieldsByName(t *testing.T) {
	t.Parallel()

	t.Run("nil receiver", func(t *testing.T) {
		t.Parallel()
		var td *TypeDef
		td.SortFieldsByName() // should not panic
	})

	t.Run("sorts fields alphabetically", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "zebra", Type: &TypeDef{Type: DataTypeString}},
				{Name: "apple", Type: &TypeDef{Type: DataTypeString}},
				{Name: "mango", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		root.SortFieldsByName()

		assert.Equal(t, "apple", root.Fields[0].Name)
		assert.Equal(t, "mango", root.Fields[1].Name)
		assert.Equal(t, "zebra", root.Fields[2].Name)
	})

	t.Run("empty fields", func(t *testing.T) {
		t.Parallel()

		root := &TypeDef{
			Type:   DataTypeClass,
			Fields: Fields{},
		}

		root.SortFieldsByName() // should not panic
		assert.Empty(t, root.Fields)
	})
}

func TestTypeDef_IsTerraformSymbolEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		td1      *TypeDef
		td2      *TypeDef
		expected bool
	}{
		{
			name:     "nil TypeDefs are equal",
			td1:      nil,
			td2:      nil,
			expected: true,
		},
		{
			name:     "nil and non-nil TypeDef are not equal",
			td1:      nil,
			td2:      &TypeDef{Type: DataTypeString},
			expected: false,
		},
		{
			name:     "same pointer is equal",
			td1:      &TypeDef{Type: DataTypeString},
			td2:      nil, // overwritten below
			expected: true,
		},
		{
			name:     "same data types are equal",
			td1:      &TypeDef{Type: DataTypeString},
			td2:      &TypeDef{Type: DataTypeString},
			expected: true,
		},
		{
			name:     "different data types are not equal",
			td1:      &TypeDef{Type: DataTypeString},
			td2:      &TypeDef{Type: DataTypeInteger},
			expected: false,
		},
		{
			name: "enum with same underlying type is equal",
			td1: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
			},
			td2:      &TypeDef{Type: DataTypeString},
			expected: true,
		},
		{
			name: "enum with different underlying type is not equal",
			td1: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
			},
			td2:      &TypeDef{Type: DataTypeInteger},
			expected: false,
		},
		{
			name: "fields with same sanitized names are equal",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "myField", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: true,
		},
		{
			name: "fields with different counts are not equal",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					{Name: "field2", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: false,
		},
		{
			name: "fields with different optional are not equal",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}, Optional: true},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}, Optional: false},
				},
			},
			expected: false,
		},
		{
			name: "fields with different nullable are not equal",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}, Nullable: true},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}, Nullable: false},
				},
			},
			expected: false,
		},
		{
			name: "fields with no sanitized name match are not equal",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "alpha", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "beta", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: false,
		},
		{
			name: "equal associated types",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			expected: true,
		},
		{
			name: "different number of associated types not equal",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			expected: false,
		},
		{
			name: "associated types with no match in other are not equal",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "TypeA", Type: DataTypeString},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "TypeB", Type: DataTypeString},
				},
			},
			expected: false,
		},
		{
			name: "equal item types",
			td1: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			td2: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			expected: true,
		},
		{
			name: "different item types not equal",
			td1: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			td2: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeInteger},
			},
			expected: false,
		},
		{
			name: "item type present vs absent not equal",
			td1: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			td2: &TypeDef{
				Type: DataTypeArray,
			},
			expected: false,
		},
		{
			name: "nested class fields recursively compared",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "nested", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "inner", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "nested", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "inner", Type: &TypeDef{Type: DataTypeInteger}},
						},
					}},
				},
			},
			expected: false,
		},
	}

	// Fix up the same-pointer test case.
	for i := range tests {
		if tests[i].name == "same pointer is equal" {
			tests[i].td2 = tests[i].td1
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.td1.IsTerraformSymbolEqual(tt.td2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
