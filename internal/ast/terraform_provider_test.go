package ast

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testManagedResourceOperations returns a minimal valid
// TerraformManagedResourceOperations for testing purposes.
func testManagedResourceOperations() *TerraformManagedResourceOperations {
	return &TerraformManagedResourceOperations{
		create: map[int]*TerraformOperation{
			0: {APIOperation: &Operation{}},
		},
	}
}

// testActionOperations returns a minimal valid TerraformActionOperations for
// testing purposes.
func testActionOperations() *TerraformActionOperations {
	return &TerraformActionOperations{
		invoke: map[int]*TerraformOperation{
			0: {APIOperation: &Operation{}},
		},
	}
}

// testDataResourceOperations returns a minimal valid
// TerraformDataResourceOperations for testing purposes.
func testDataResourceOperations() *TerraformDataResourceOperations {
	return &TerraformDataResourceOperations{
		read: map[int]*TerraformOperation{
			0: {APIOperation: &Operation{}},
		},
	}
}

// testEphemeralResourceOperations returns a minimal valid
// TerraformEphemeralResourceOperations for testing purposes.
func testEphemeralResourceOperations() *TerraformEphemeralResourceOperations {
	return &TerraformEphemeralResourceOperations{
		open: map[int]*TerraformOperation{
			0: {APIOperation: &Operation{}},
		},
	}
}

func TestTerraformProvider_AnnotateTerraformSymbols(t *testing.T) {
	t.Parallel()

	t.Run("assigns symbol to class type", func(t *testing.T) {
		t.Parallel()

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "status", Type: &TypeDef{
						Name: "Status",
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "code", Type: &TypeDef{Type: DataTypeInteger}},
						},
					}},
				},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		statusType := provider.ManagedResources[0].SchemaTypeDef.Fields[0].Type
		symbolVal, ok := statusType.Extensions.Get("Symbol")
		require.True(t, ok)
		assert.Equal(t, "Status", symbolVal)
	})

	t.Run("deduplicates structurally identical types across resources", func(t *testing.T) {
		t.Parallel()

		metadataType1 := &TypeDef{
			Name: "Metadata",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "key", Type: &TypeDef{Type: DataTypeString}},
			},
		}
		metadataType2 := &TypeDef{
			Name: "Metadata",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "key", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "meta", Type: metadataType1}},
				},
			},
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "meta", Type: metadataType2}},
				},
			},
		}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym1, _ := metadataType1.Extensions.Get("Symbol")
		sym2, _ := metadataType2.Extensions.Get("Symbol")
		assert.Equal(t, "Metadata", sym1)
		assert.Equal(t, "Metadata", sym2, "structurally identical types should share the same symbol")
	})

	t.Run("numeric suffix on name collision with different structure", func(t *testing.T) {
		t.Parallel()

		type1 := &TypeDef{
			Name: "Config",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "key", Type: &TypeDef{Type: DataTypeString}},
			},
		}
		type2 := &TypeDef{
			Name: "Config",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "value", Type: &TypeDef{Type: DataTypeInteger}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "cfg", Type: type1}},
				},
			},
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "cfg", Type: type2}},
				},
			},
		}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym1, _ := type1.Extensions.Get("Symbol")
		sym2, _ := type2.Extensions.Get("Symbol")
		assert.Equal(t, "Config", sym1)
		assert.Equal(t, "Config1", sym2, "structurally different types with same name should get numeric suffix")
	})

	t.Run("reserved keyword avoidance", func(t *testing.T) {
		t.Parallel()

		// A type named "type" should get a numeric suffix since "type" is reserved.
		typeType := &TypeDef{
			Name: "type",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "name", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{{Name: "t", Type: typeType}},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym, _ := typeType.Extensions.Get("Symbol")
		// "Type" would be the PascalCase name, but "type" is reserved in the
		// symbol manager so the first available suffix is used.
		assert.Equal(t, "Type1", sym)
	})

	t.Run("pre-assigned symbol is not overwritten", func(t *testing.T) {
		t.Parallel()

		preAssigned := &TypeDef{
			Name: "Widget",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "id", Type: &TypeDef{Type: DataTypeString}},
			},
			Extensions: &TypeDefExtensions{
				All: map[string]any{"Symbol": "CustomWidget"},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{{Name: "w", Type: preAssigned}},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym, _ := preAssigned.Extensions.Get("Symbol")
		assert.Equal(t, "CustomWidget", sym)
	})

	t.Run("enableTypeDeduplication deduplicates differently named but structurally identical types", func(t *testing.T) {
		t.Parallel()

		type1 := &TypeDef{
			Name: "Alpha",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "value", Type: &TypeDef{Type: DataTypeString}},
			},
		}
		type2 := &TypeDef{
			Name: "Beta",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "value", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "a", Type: type1}},
				},
			},
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "b", Type: type2}},
				},
			},
		}

		// With dedup enabled, structurally identical types share symbol
		// regardless of name.
		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), true))

		sym1, _ := type1.Extensions.Get("Symbol")
		sym2, _ := type2.Extensions.Get("Symbol")
		assert.Equal(t, sym1, sym2, "with enableTypeDeduplication, structurally identical types should share symbol")
	})

	t.Run("without enableTypeDeduplication differently named types get different symbols", func(t *testing.T) {
		t.Parallel()

		type1 := &TypeDef{
			Name: "Alpha",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "value", Type: &TypeDef{Type: DataTypeString}},
			},
		}
		type2 := &TypeDef{
			Name: "Beta",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "value", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "a", Type: type1}},
				},
			},
			{
				Operations: testManagedResourceOperations(),
				SchemaTypeDef: &TypeDef{
					Type:   DataTypeClass,
					Fields: Fields{{Name: "b", Type: type2}},
				},
			},
		}

		// Without dedup, different names mean different symbols even if
		// structurally identical.
		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym1, _ := type1.Extensions.Get("Symbol")
		sym2, _ := type2.Extensions.Get("Symbol")
		assert.Equal(t, "Alpha", sym1)
		assert.Equal(t, "Beta", sym2)
	})

	t.Run("walks union associated types", func(t *testing.T) {
		t.Parallel()

		variant1 := &TypeDef{
			Name: "VariantA",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "a", Type: &TypeDef{Type: DataTypeString}},
			},
		}
		variant2 := &TypeDef{
			Name: "VariantB",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "b", Type: &TypeDef{Type: DataTypeInteger}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "choice", Type: &TypeDef{
						Name:            "Choice",
						Type:            DataTypeUnion,
						AssociatedTypes: []*TypeDef{variant1, variant2},
					}},
				},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		// The union type itself gets a symbol.
		choiceType := provider.ManagedResources[0].SchemaTypeDef.Fields[0].Type
		choiceSym, ok := choiceType.Extensions.Get("Symbol")
		require.True(t, ok)
		assert.Equal(t, "Choice", choiceSym)

		// Each variant gets a symbol.
		sym1, ok := variant1.Extensions.Get("Symbol")
		require.True(t, ok)
		assert.Equal(t, "VariantA", sym1)

		sym2, ok := variant2.Extensions.Get("Symbol")
		require.True(t, ok)
		assert.Equal(t, "VariantB", sym2)
	})

	t.Run("walks array of class items", func(t *testing.T) {
		t.Parallel()

		itemType := &TypeDef{
			Name: "Item",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "name", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "items", Type: &TypeDef{
						Type:     DataTypeArray,
						ItemType: itemType,
					}},
				},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym, ok := itemType.Extensions.Get("Symbol")
		require.True(t, ok)
		assert.Equal(t, "Item", sym)
	})

	t.Run("walks array of map of class items", func(t *testing.T) {
		t.Parallel()

		itemType := &TypeDef{
			Name: "MapItem",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "key", Type: &TypeDef{Type: DataTypeString}},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "list_map_object", Type: &TypeDef{
						Type: DataTypeArray,
						ItemType: &TypeDef{
							Type:     DataTypeMap,
							ItemType: itemType,
						},
					}},
				},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		sym, ok := itemType.Extensions.Get("Symbol")
		require.True(t, ok)
		assert.Equal(t, "MapItem", sym)
	})

	t.Run("entity ordering: actions before managed resources", func(t *testing.T) {
		t.Parallel()

		actionType := &TypeDef{
			Name: "Config",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "action_field", Type: &TypeDef{Type: DataTypeString}},
			},
		}
		resourceType := &TypeDef{
			Name: "Config",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "resource_field", Type: &TypeDef{Type: DataTypeInteger}},
			},
		}

		provider := NewTerraformProvider()
		provider.Actions = []*TerraformAction{{
			Operations: testActionOperations(),
			SchemaTypeDef: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{{Name: "cfg", Type: actionType}},
			},
		}}
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{{Name: "cfg", Type: resourceType}},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		// Actions are processed first, so the action type gets "Config".
		actionSym, _ := actionType.Extensions.Get("Symbol")
		resourceSym, _ := resourceType.Extensions.Get("Symbol")
		assert.Equal(t, "Config", actionSym)
		assert.Equal(t, "Config1", resourceSym)
	})

	// Table-driven tests for type traversal scenarios covering collection
	// types, custom type config, nil safety, and union hoisted fields.
	traversalTests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "walks set of class items",
			run: func(t *testing.T) {
				t.Helper()
				itemType := &TypeDef{
					Name: "SetItem",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "name", Type: &TypeDef{Type: DataTypeString}},
					},
				}

				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "tags", Type: &TypeDef{
								Type:     DataTypeSet,
								ItemType: itemType,
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				sym, ok := itemType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "SetItem", sym)
			},
		},
		{
			name: "custom type config on array skips recursion",
			run: func(t *testing.T) {
				t.Helper()
				innerType := &TypeDef{
					Name: "ShouldNotGetSymbol",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "val", Type: &TypeDef{Type: DataTypeString}},
					},
				}

				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "custom_list", Type: &TypeDef{
								Type:     DataTypeArray,
								ItemType: innerType,
								Extensions: &TypeDefExtensions{
									All: map[string]any{"x-speakeasy-terraform-custom-type": true},
								},
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				assert.Nil(t, innerType.Extensions, "custom type config should prevent recursion into ItemType")
			},
		},
		{
			name: "custom type config on map skips recursion",
			run: func(t *testing.T) {
				t.Helper()
				innerType := &TypeDef{
					Name: "ShouldNotGetSymbol",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "val", Type: &TypeDef{Type: DataTypeString}},
					},
				}

				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "custom_map", Type: &TypeDef{
								Type:     DataTypeMap,
								ItemType: innerType,
								Extensions: &TypeDefExtensions{
									All: map[string]any{"x-speakeasy-terraform-custom-type": true},
								},
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				assert.Nil(t, innerType.Extensions, "custom type config on map should prevent recursion into ItemType")
			},
		},
		{
			name: "nil ItemType in array is safe",
			run: func(t *testing.T) {
				t.Helper()
				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "empty_list", Type: &TypeDef{
								Type:     DataTypeArray,
								ItemType: nil,
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))
			},
		},
		{
			name: "primitive array skips symbol assignment",
			run: func(t *testing.T) {
				t.Helper()
				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "string_list", Type: &TypeDef{
								Type:     DataTypeArray,
								ItemType: &TypeDef{Type: DataTypeString},
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				itemType := provider.ManagedResources[0].SchemaTypeDef.Fields[0].Type.ItemType
				assert.Nil(t, itemType.Extensions, "primitive array items should not get symbol extensions")
			},
		},
		{
			name: "walks map of class items",
			run: func(t *testing.T) {
				t.Helper()
				itemType := &TypeDef{
					Name: "MapValue",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "id", Type: &TypeDef{Type: DataTypeString}},
					},
				}

				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "lookup", Type: &TypeDef{
								Type:     DataTypeMap,
								ItemType: itemType,
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				sym, ok := itemType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "MapValue", sym)
			},
		},
		{
			name: "union with hoisted fields assigns symbols to both variants and hoisted fields",
			run: func(t *testing.T) {
				t.Helper()
				variant := &TypeDef{
					Name: "OptionA",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "a_val", Type: &TypeDef{Type: DataTypeString}},
					},
				}
				hoistedFieldType := &TypeDef{
					Name: "SharedConfig",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "setting", Type: &TypeDef{Type: DataTypeString}},
					},
				}

				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "choice", Type: &TypeDef{
								Name:            "Choice",
								Type:            DataTypeUnion,
								AssociatedTypes: []*TypeDef{variant},
								Fields: Fields{
									{Name: "shared", Type: hoistedFieldType},
								},
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				variantSym, ok := variant.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "OptionA", variantSym)

				hoistedSym, ok := hoistedFieldType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "SharedConfig", hoistedSym)
			},
		},
		{
			name: "walks set of union items",
			run: func(t *testing.T) {
				t.Helper()
				variant := &TypeDef{
					Name: "SetVariant",
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "val", Type: &TypeDef{Type: DataTypeString}},
					},
				}
				unionItemType := &TypeDef{
					Name:            "SetUnion",
					Type:            DataTypeUnion,
					AssociatedTypes: []*TypeDef{variant},
				}

				provider := NewTerraformProvider()
				provider.ManagedResources = []*TerraformManagedResource{{
					Operations: testManagedResourceOperations(),
					SchemaTypeDef: &TypeDef{
						Type: DataTypeClass,
						Fields: Fields{
							{Name: "items", Type: &TypeDef{
								Type:     DataTypeSet,
								ItemType: unionItemType,
							}},
						},
					},
				}}

				require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

				unionSym, ok := unionItemType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "SetUnion", unionSym)

				variantSym, ok := variant.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "SetVariant", variantSym)
			},
		},
	}

	for _, tt := range traversalTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.run(t)
		})
	}

	t.Run("skips fields with MatchConfig Path", func(t *testing.T) {
		t.Parallel()

		pathStr := "id"
		skippedType := &TypeDef{
			Name: "ShouldSkip",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "val", Type: &TypeDef{Type: DataTypeString}},
			},
			Extensions: &TypeDefExtensions{
				All:         map[string]any{},
				MatchConfig: &extensions.MatchConfig{Path: &pathStr},
			},
		}

		provider := NewTerraformProvider()
		provider.ManagedResources = []*TerraformManagedResource{{
			Operations: testManagedResourceOperations(),
			SchemaTypeDef: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "alias_field", Type: skippedType},
				},
			},
		}}

		require.NoError(t, provider.AnnotateTerraformSymbols(context.Background(), false))

		_, hasSymbol := skippedType.Extensions.Get("Symbol")
		assert.False(t, hasSymbol, "fields with MatchConfig.Path should be skipped")
	})
}
