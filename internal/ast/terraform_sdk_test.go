package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
)

func TestTypeDef_isTerraformDataModelCompatible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		schema     *TypeDef
		entityName string
		sdkType    *TypeDef
		expected   bool
	}{
		{
			name:       "nil schema returns false",
			schema:     nil,
			entityName: "Thing",
			sdkType:    &TypeDef{Type: DataTypeClass},
			expected:   false,
		},
		{
			name:       "nil sdkType returns false",
			schema:     &TypeDef{Type: DataTypeClass},
			entityName: "Thing",
			sdkType:    nil,
			expected:   false,
		},
		{
			name: "schema with TerraformIgnore.DataModel returns false",
			schema: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					TerraformIgnore: &extensions.TerraformIgnore{
						DataModel: true,
					},
				},
			},
			entityName: "Thing",
			sdkType:    &TypeDef{Type: DataTypeClass},
			expected:   false,
		},
		{
			name:       "sdkType with TerraformIgnore.DataModel returns false",
			schema:     &TypeDef{Type: DataTypeClass},
			entityName: "Thing",
			sdkType: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					TerraformIgnore: &extensions.TerraformIgnore{
						DataModel: true,
					},
				},
			},
			expected: false,
		},
		{
			name:       "incompatible types returns false",
			schema:     &TypeDef{Type: DataTypeString},
			entityName: "Thing",
			sdkType:    &TypeDef{Type: DataTypeClass},
			expected:   false,
		},
		{
			name:       "matching primitive types returns true",
			schema:     &TypeDef{Type: DataTypeString},
			entityName: "Thing",
			sdkType:    &TypeDef{Type: DataTypeString},
			expected:   true,
		},
		{
			name: "primitive with entity extension mismatch returns false",
			schema: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-speakeasy-terraform-in-entity": true,
					},
				},
			},
			entityName: "Thing",
			sdkType:    &TypeDef{Type: DataTypeString},
			expected:   false,
		},
		{
			name: "primitive with matching entity extension returns true",
			schema: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-speakeasy-terraform-in-entity": true,
					},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-speakeasy-terraform-in-entity": true,
					},
				},
			},
			expected: true,
		},
		{
			name: "class with matching field names returns true",
			schema: &TypeDef{
				Name: "Schema",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDK",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: true,
		},
		{
			name: "class with no matching field names and no entity fallback returns false",
			schema: &TypeDef{
				Name: "Schema",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDK",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "unrelated", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: false,
		},
		{
			name: "class with entity containment fallback returns true",
			schema: &TypeDef{
				Name: "Schema",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDK",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{
						Name: "data",
						Type: &TypeDef{
							Name: "ThingData",
							Type: DataTypeClass,
							Extensions: &TypeDefExtensions{
								Entity: &extensions.Entity{Names: []string{"Thing"}},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "class schema with SDK array containing entity returns true",
			schema: &TypeDef{
				Name: "Schema",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "ThingList",
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Name: "ThingItem",
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						Entity: &extensions.Entity{Names: []string{"Thing"}},
					},
				},
			},
			expected: true,
		},
		{
			name: "wrapped extension returns true regardless of field matching",
			schema: &TypeDef{
				Name: "Schema",
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-speakeasy-wrapped-attribute": "data",
					},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDK",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "unrelated", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: true,
		},
		{
			name: "matching array item types returns true",
			schema: &TypeDef{
				Name: "SchemaList",
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Name: "SchemaItem",
					Type: DataTypeClass,
					Fields: []*FieldDef{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDKList",
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Name: "SDKItem",
					Type: DataTypeClass,
					Fields: []*FieldDef{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			expected: true,
		},
		{
			name: "union with matching associated types returns true",
			schema: &TypeDef{
				Name: "SchemaUnion",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name:         "Variant",
						OriginalName: "Variant",
						Type:         DataTypeClass,
						Fields: []*FieldDef{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDKUnion",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name:         "Variant",
						OriginalName: "Variant",
						Type:         DataTypeClass,
						Fields: []*FieldDef{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "numeric coercion int32 to integer is compatible",
			schema: &TypeDef{
				Name: "Schema",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "count", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			entityName: "Thing",
			sdkType: &TypeDef{
				Name: "SDK",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "count", Type: &TypeDef{Type: DataTypeInt32}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.schema.isTerraformDataModelCompatible(tt.entityName, tt.sdkType)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeDef_isTerraformDataModelCompatible_cycleDetection(t *testing.T) {
	t.Parallel()

	// Create a recursive type: Schema has a field "child" of type Schema.
	// Extensions must be non-nil on the recursive type since the compatibility
	// check calls Extensions.Has() on primitive-typed fields.
	schema := &TypeDef{
		Name:       "RecursiveSchema",
		Type:       DataTypeClass,
		Extensions: &TypeDefExtensions{},
	}
	schema.Fields = []*FieldDef{
		{Name: "name", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
		{Name: "child", Type: schema}, // self-reference
	}

	sdkType := &TypeDef{
		Name:       "RecursiveSDK",
		Type:       DataTypeClass,
		Extensions: &TypeDefExtensions{},
	}
	sdkType.Fields = []*FieldDef{
		{Name: "name", Type: &TypeDef{Type: DataTypeString, Extensions: &TypeDefExtensions{}}},
		{Name: "child", Type: sdkType}, // self-reference
	}

	// Should not infinite loop; cycle detection returns true for revisited types.
	result := schema.isTerraformDataModelCompatible("Thing", sdkType)

	assert.True(t, result)
}

func TestTypeDef_hasTerraformIgnoreDataModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		td       *TypeDef
		expected bool
	}{
		{
			name:     "nil Extensions returns false",
			td:       &TypeDef{Type: DataTypeString},
			expected: false,
		},
		{
			name: "nil TerraformIgnore returns false",
			td: &TypeDef{
				Type:       DataTypeString,
				Extensions: &TypeDefExtensions{},
			},
			expected: false,
		},
		{
			name: "TerraformIgnore.DataModel false returns false",
			td: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					TerraformIgnore: &extensions.TerraformIgnore{
						DataModel: false,
					},
				},
			},
			expected: false,
		},
		{
			name: "TerraformIgnore.DataModel true returns true",
			td: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					TerraformIgnore: &extensions.TerraformIgnore{
						DataModel: true,
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.td.hasTerraformIgnoreDataModel())
		})
	}
}

func TestTypeDef_hasWrappedExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		td       *TypeDef
		expected bool
	}{
		{
			name:     "nil Extensions returns false",
			td:       &TypeDef{Type: DataTypeClass},
			expected: false,
		},
		{
			name: "nil All map returns false",
			td: &TypeDef{
				Type:       DataTypeClass,
				Extensions: &TypeDefExtensions{},
			},
			expected: false,
		},
		{
			name: "no wrapped extensions returns false",
			td: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-speakeasy-other": true,
					},
				},
			},
			expected: false,
		},
		{
			name: "wrapped extension returns true",
			td: &TypeDef{
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-speakeasy-wrapped-attribute": "data",
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.td.hasWrappedExtension())
		})
	}
}

func TestTypeDef_terraformCycleDetectionKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		td       *TypeDef
		expected string
	}{
		{
			name: "prefers Symbol extension",
			td: &TypeDef{
				Name: "MyType",
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"Symbol": "MySymbol",
					},
				},
			},
			expected: "MySymbol",
		},
		{
			name: "falls back to Name when no Symbol",
			td: &TypeDef{
				Name: "MyType",
				Type: DataTypeClass,
			},
			expected: "MyType",
		},
		{
			name: "falls back to Type string when no Name",
			td: &TypeDef{
				Type: DataTypeClass,
			},
			expected: "class",
		},
		{
			name: "skips empty string Symbol",
			td: &TypeDef{
				Name: "MyType",
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"Symbol": "",
					},
				},
			},
			expected: "MyType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.td.terraformCycleDetectionKey())
		})
	}
}

func TestTypeDef_TerraformHasNestedWriteOnlyFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schema   *TypeDef
		sdkType  *TypeDef
		expected bool
	}{
		{
			name:     "nil schema returns false",
			schema:   nil,
			sdkType:  &TypeDef{Type: DataTypeClass},
			expected: false,
		},
		{
			name:     "nil sdkType returns false",
			schema:   &TypeDef{Type: DataTypeClass},
			sdkType:  nil,
			expected: false,
		},
		{
			name: "all fields matched returns false",
			schema: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					{Name: "name", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: false,
		},
		{
			name: "leaf write-only field returns true",
			schema: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					{Name: "secret", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "id", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: true,
		},
		{
			name: "nested class write-only field returns true",
			schema: &TypeDef{
				Name: "Parent",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "config", Type: &TypeDef{
						Name: "Config",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "visible", Type: &TypeDef{Type: DataTypeString}},
							{Name: "hidden", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			sdkType: &TypeDef{
				Name: "SDKParent",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "config", Type: &TypeDef{
						Name: "SDKConfig",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "visible", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			expected: true,
		},
		{
			name: "nested class all matched returns false",
			schema: &TypeDef{
				Name: "Parent",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "config", Type: &TypeDef{
						Name: "Config",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "value", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			sdkType: &TypeDef{
				Name: "SDKParent",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "config", Type: &TypeDef{
						Name: "SDKConfig",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "value", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			expected: false,
		},
		{
			name: "entire nested class is write-only returns true",
			schema: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "metadata", Type: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "key", Type: &TypeDef{Type: DataTypeString}},
						},
					}},
				},
			},
			sdkType: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{},
			},
			expected: true,
		},
		{
			name: "array item type with write-only field returns true",
			schema: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						{Name: "secret", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			expected: true,
		},
		{
			name: "class with array field whose items have write-only fields returns true",
			schema: &TypeDef{
				Name: "Variant",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "credentials", Type: &TypeDef{
						Type: DataTypeArray,
						ItemType: &TypeDef{
							Name: "Credentials",
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "client_id", Type: &TypeDef{Type: DataTypeString}},
								{Name: "client_secret", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					}},
				},
			},
			sdkType: &TypeDef{
				Name: "SDKVariant",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "credentials", Type: &TypeDef{
						Type: DataTypeArray,
						ItemType: &TypeDef{
							Name: "SDKCredentials",
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "client_id", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					}},
				},
			},
			expected: true,
		},
		{
			name: "class with array field all matched returns false",
			schema: &TypeDef{
				Name: "Variant",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "credentials", Type: &TypeDef{
						Type: DataTypeArray,
						ItemType: &TypeDef{
							Name: "Credentials",
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "client_id", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					}},
				},
			},
			sdkType: &TypeDef{
				Name: "SDKVariant",
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "credentials", Type: &TypeDef{
						Type: DataTypeArray,
						ItemType: &TypeDef{
							Name: "SDKCredentials",
							Type: DataTypeClass,
							Fields: Fields{
								{Name: "client_id", Type: &TypeDef{Type: DataTypeString}},
							},
						},
					}},
				},
			},
			expected: false,
		},
		{
			name: "array item type all matched returns false",
			schema: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			expected: false,
		},
		{
			name: "map item type with write-only field returns true",
			schema: &TypeDef{
				Type: DataTypeMap,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "value", Type: &TypeDef{Type: DataTypeString}},
						{Name: "write_only", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeMap,
				ItemType: &TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "value", Type: &TypeDef{Type: DataTypeString}},
					},
				},
			},
			expected: true,
		},
		{
			name: "union variant with write-only field returns true",
			schema: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: TypeDefs{
					{
						Name: "VariantA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
							{Name: "secret", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: TypeDefs{
					{
						Name: "VariantA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "union variant all matched returns false",
			schema: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: TypeDefs{
					{
						Name: "VariantA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: TypeDefs{
					{
						Name: "VariantA",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "union variants matched by discriminator mapping returns true",
			schema: &TypeDef{
				Type: DataTypeUnion,
				Discriminator: &Discriminator{
					Mapping: DiscriminatorMappings{
						{Name: "type_a", Type: &TypeDef{Name: "SchemaVariant"}},
					},
				},
				AssociatedTypes: TypeDefs{
					{
						Name: "SchemaVariant",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
							{Name: "secret", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeUnion,
				Discriminator: &Discriminator{
					Mapping: DiscriminatorMappings{
						{Name: "type_a", Type: &TypeDef{Name: "SDKVariant"}},
					},
				},
				AssociatedTypes: TypeDefs{
					{
						Name: "SDKVariant",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "id", Type: &TypeDef{Type: DataTypeString}},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "recursive type does not infinite loop",
			schema: func() *TypeDef {
				td := &TypeDef{
					Name: "Recursive",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				}
				td.Fields = append(td.Fields, &FieldDef{Name: "child", Type: td})
				return td
			}(),
			sdkType: func() *TypeDef {
				td := &TypeDef{
					Name: "Recursive",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				}
				td.Fields = append(td.Fields, &FieldDef{Name: "child", Type: td})
				return td
			}(),
			expected: false,
		},
		{
			name: "field name matching is case-insensitive via sanitization",
			schema: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			sdkType: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "MyField", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: false,
		},
		{
			name: "empty class returns false",
			schema: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{},
			},
			sdkType: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{},
			},
			expected: false,
		},
		{
			name:     "primitive types return false",
			schema:   &TypeDef{Type: DataTypeString},
			sdkType:  &TypeDef{Type: DataTypeString},
			expected: false,
		},
		{
			name: "nil ItemType on one side returns false",
			schema: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeClass, Fields: Fields{{Name: "secret", Type: &TypeDef{Type: DataTypeString}}}},
			},
			sdkType: &TypeDef{
				Type: DataTypeArray,
			},
			expected: false,
		},
		{
			name: "union field with no SDK equivalent returns true",
			schema: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "data", Type: &TypeDef{
						Type:            DataTypeUnion,
						AssociatedTypes: TypeDefs{{Name: "Str", Type: DataTypeString}},
					}},
				},
			},
			sdkType: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.schema.TerraformHasNestedWriteOnlyFields(tt.sdkType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_terraformSanitizeUnionVariantName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		unionType      *TypeDef
		associatedType *TypeDef
		expected       string
	}{
		{
			name:           "nil associated type returns empty",
			unionType:      &TypeDef{Type: DataTypeUnion},
			associatedType: nil,
			expected:       "",
		},
		{
			name: "discriminator mapping name preferred",
			unionType: &TypeDef{
				Type: DataTypeUnion,
				Discriminator: &Discriminator{
					Mapping: DiscriminatorMappings{
						{Name: "my_variant", Type: &TypeDef{Name: "MyVariant"}},
					},
				},
			},
			associatedType: &TypeDef{Name: "MyVariant", Type: DataTypeClass},
			expected:       "MyVariant",
		},
		{
			name:           "original name used when no discriminator",
			unionType:      &TypeDef{Type: DataTypeUnion},
			associatedType: &TypeDef{OriginalName: "original_name", Name: "ResolvedName", Type: DataTypeClass},
			expected:       "OriginalName",
		},
		{
			name:           "name used when no original name",
			unionType:      &TypeDef{Type: DataTypeUnion},
			associatedType: &TypeDef{Name: "VariantName", Type: DataTypeClass},
			expected:       "VariantName",
		},
		{
			name:           "string type returns Str",
			unionType:      &TypeDef{Type: DataTypeUnion},
			associatedType: &TypeDef{Type: DataTypeString},
			expected:       "Str",
		},
		{
			name:      "array type returns array_Of_ pattern",
			unionType: &TypeDef{Type: DataTypeUnion},
			associatedType: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Name: "Item", Type: DataTypeClass},
			},
			expected: "ArrayOfItem",
		},
		{
			name:           "other type returns sanitized type name",
			unionType:      &TypeDef{Type: DataTypeUnion},
			associatedType: &TypeDef{Type: DataTypeInteger},
			expected:       "Integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := terraformSanitizeUnionVariantName(tt.unionType, tt.associatedType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
