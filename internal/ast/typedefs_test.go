package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeDefs_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typedefs TypeDefs
		test     func(t *testing.T, original, cloned []*TypeDef)
	}{
		{
			name:     "nil TypeDefs",
			typedefs: nil,
			test: func(t *testing.T, original, cloned []*TypeDef) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:     "empty TypeDefs",
			typedefs: TypeDefs{},
			test: func(t *testing.T, original, cloned []*TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				assert.Empty(t, cloned)
			},
		},
		{
			name: "TypeDefs with single TypeDef",
			typedefs: TypeDefs{
				{
					Type: DataTypeString,
					Name: "StringType",
				},
			},
			test: func(t *testing.T, original, cloned []*TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 1)

				assert.NotSame(t, original[0], cloned[0])
				assert.Equal(t, original[0].Type, cloned[0].Type)
				assert.Equal(t, original[0].Name, cloned[0].Name)
			},
		},
		{
			name: "TypeDefs with multiple TypeDef elements",
			typedefs: TypeDefs{
				{
					Type: DataTypeString,
					Name: "StringType",
					Validations: &Validations{
						MinLength: ptr(int64(1)),
						MaxLength: ptr(int64(100)),
					},
				},
				{
					Type: DataTypeNumber,
					Name: "NumberType",
					Validations: &Validations{
						Minimum: ptr(0.0),
						Maximum: ptr(100.0),
					},
				},
				{
					Type: DataTypeClass,
					Name: "ClassType",
					Fields: Fields{
						{
							Name: "field1",
							Type: &TypeDef{
								Type: DataTypeString,
							},
						},
					},
				},
				{
					Type: DataTypeArray,
					Name: "ArrayType",
					ItemType: &TypeDef{
						Type: DataTypeInteger,
					},
				},
			},
			test: func(t *testing.T, original, cloned []*TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 4)

				for i := range original {
					assert.NotSame(t, original[i], cloned[i])
					assert.Equal(t, original[i].Type, cloned[i].Type)
					assert.Equal(t, original[i].Name, cloned[i].Name)

					if original[i].Validations != nil {
						require.NotNil(t, cloned[i].Validations)
						assert.NotSame(t, original[i].Validations, cloned[i].Validations)
					}

					if len(original[i].Fields) > 0 {
						assert.NotEmpty(t, cloned[i].Fields)
					}

					if original[i].ItemType != nil {
						require.NotNil(t, cloned[i].ItemType)
						assert.NotSame(t, original[i].ItemType, cloned[i].ItemType)
						assert.Equal(t, original[i].ItemType.Type, cloned[i].ItemType.Type)
					}
				}
			},
		},
		{
			name: "TypeDefs with nested structures",
			typedefs: TypeDefs{
				{
					Type: DataTypeClass,
					Name: "NestedClass",
					Fields: Fields{
						{
							Name: "nested",
							Type: &TypeDef{
								Type: DataTypeClass,
								Fields: Fields{
									{
										Name: "deepField",
										Type: &TypeDef{
											Type: DataTypeString,
										},
									},
								},
							},
						},
					},
				},
			},
			test: func(t *testing.T, original, cloned []*TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 1)

				assert.NotSame(t, original[0], cloned[0])
				assert.Equal(t, original[0].Type, cloned[0].Type)
				assert.Equal(t, original[0].Name, cloned[0].Name)

				assert.NotEmpty(t, cloned[0].Fields)

				origFields := original[0].Fields
				clonedFields := cloned[0].Fields
				require.Len(t, clonedFields, len(origFields))

				origField := origFields[0]
				clonedField := clonedFields[0]
				assert.NotSame(t, origField, clonedField)
				assert.Equal(t, origField.Name, clonedField.Name)

				assert.NotSame(t, origField.Type, clonedField.Type)
				assert.Equal(t, origField.Type.Type, clonedField.Type.Type)

				assert.NotEmpty(t, clonedField.Type.Fields)

				origDeepFields := origField.Type.Fields
				clonedDeepFields := clonedField.Type.Fields
				require.Len(t, clonedDeepFields, len(origDeepFields))

				origDeepField := origDeepFields[0]
				clonedDeepField := clonedDeepFields[0]
				assert.NotSame(t, origDeepField, clonedDeepField)
				assert.Equal(t, origDeepField.Name, clonedDeepField.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.typedefs.Clone()
			tt.test(t, tt.typedefs, cloned)
		})
	}
}

func TestTypeDefs_AllHaveTerraformEqualField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeDefs TypeDefs
		field    *FieldDef
		expected bool
	}{
		{
			name:     "empty TypeDefs returns true",
			typeDefs: TypeDefs{},
			field:    &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			expected: true,
		},
		{
			name: "single TypeDef with matching field",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			field:    &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			expected: true,
		},
		{
			name: "single TypeDef without matching field",
			typeDefs: TypeDefs{
				{
					Type:   DataTypeClass,
					Fields: Fields{},
				},
			},
			field:    &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			expected: false,
		},
		{
			name: "all TypeDefs have matching field",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonField", Type: &TypeDef{Type: DataTypeString}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonField", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			field:    &FieldDef{Name: "commonField", Type: &TypeDef{Type: DataTypeString}},
			expected: true,
		},
		{
			name: "one TypeDef missing field",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonField", Type: &TypeDef{Type: DataTypeString}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "otherField", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			field:    &FieldDef{Name: "commonField", Type: &TypeDef{Type: DataTypeString}},
			expected: false,
		},
		{
			name: "field with different type",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			field:    &FieldDef{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.typeDefs.AllHaveTerraformEqualField(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeDefs_TerraformCommonFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeDefs TypeDefs
		expected Fields
	}{
		{
			name:     "empty TypeDefs returns nil",
			typeDefs: TypeDefs{},
			expected: nil,
		},
		{
			name: "nil TypeDef in list returns nil",
			typeDefs: TypeDefs{
				nil,
				{Type: DataTypeString},
			},
			expected: nil,
		},
		{
			name: "single TypeDef returns its fields",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
						{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			expected: Fields{
				{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
			},
		},
		{
			name: "multiple TypeDefs with common primitive fields",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonString", Type: &TypeDef{Type: DataTypeString}},
						{Name: "commonInt", Type: &TypeDef{Type: DataTypeInteger}},
						{Name: "uniqueToFirst", Type: &TypeDef{Type: DataTypeBoolean}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonString", Type: &TypeDef{Type: DataTypeString}},
						{Name: "commonInt", Type: &TypeDef{Type: DataTypeInteger}},
						{Name: "uniqueToSecond", Type: &TypeDef{Type: DataTypeNumber}},
					},
				},
			},
			expected: Fields{
				{Name: "commonInt", Type: &TypeDef{Type: DataTypeInteger}},
				{Name: "commonString", Type: &TypeDef{Type: DataTypeString}},
			},
		},
		{
			name: "no common fields",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			expected: Fields{},
		},
		{
			name: "excludes non-primitive fields",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonPrimitive", Type: &TypeDef{Type: DataTypeString}},
						{Name: "commonObject", Type: &TypeDef{Type: DataTypeClass}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "commonPrimitive", Type: &TypeDef{Type: DataTypeString}},
						{Name: "commonObject", Type: &TypeDef{Type: DataTypeClass}},
					},
				},
			},
			expected: Fields{
				{Name: "commonPrimitive", Type: &TypeDef{Type: DataTypeString}},
			},
		},
		{
			name: "fields with different types not common",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "field1", Type: &TypeDef{Type: DataTypeInteger}},
					},
				},
			},
			expected: Fields{},
		},
		{
			name: "incompatible types returns nil",
			typeDefs: TypeDefs{
				{Type: DataTypeString},
				{Type: DataTypeClass},
			},
			expected: nil,
		},
		{
			name: "compatible types with enum",
			typeDefs: TypeDefs{
				{
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "enumField",
							Type: &TypeDef{
								Type: DataTypeEnum,
								Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
							},
						},
					},
				},
				{
					Type: DataTypeClass,
					Fields: Fields{
						{
							Name: "enumField",
							Type: &TypeDef{Type: DataTypeString},
						},
					},
				},
			},
			expected: Fields{
				{
					Name: "enumField",
					Type: &TypeDef{
						Type: DataTypeEnum,
						Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.typeDefs.TerraformCommonFields()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Len(t, result, len(tt.expected))

				// Verify each expected field is present with correct properties (by name, not order)
				for _, expectedField := range tt.expected {
					found := false
					for _, resultField := range result {
						if resultField.Name == expectedField.Name {
							assert.True(t, expectedField.Type.IsTerraformEqual(resultField.Type))
							found = true
							break
						}
					}
					assert.True(t, found, "expected field %s not found in result", expectedField.Name)
				}
			}
		})
	}
}
