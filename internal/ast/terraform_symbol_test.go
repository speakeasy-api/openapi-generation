package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerraformSymbolName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeDef  *TypeDef
		expected string
	}{
		{
			name:     "basic name",
			typeDef:  &TypeDef{Name: "example"},
			expected: "Example",
		},
		{
			name:     "strips Input suffix when Input flag is true",
			typeDef:  &TypeDef{Name: "CreateExampleInput", Input: true},
			expected: "CreateExample",
		},
		{
			name:     "does not strip Input suffix when Input flag is false",
			typeDef:  &TypeDef{Name: "CreateExampleInput", Input: false},
			expected: "CreateExampleInput",
		},
		{
			name:     "strips Output suffix when Output flag is true",
			typeDef:  &TypeDef{Name: "CreateExampleOutput", Output: true},
			expected: "CreateExample",
		},
		{
			name:     "does not strip Output suffix when Output flag is false",
			typeDef:  &TypeDef{Name: "CreateExampleOutput", Output: false},
			expected: "CreateExampleOutput",
		},
		{
			name:     "SDK package name is sanitized without Obj suffix",
			typeDef:  &TypeDef{Name: "operations"},
			expected: "Operations",
		},
		{
			name:     "non-reserved name",
			typeDef:  &TypeDef{Name: "widgets"},
			expected: "Widgets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := TerraformSymbolName(tt.typeDef)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSymbolManager_assignSymbol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		assignedSymbols         map[string]*TypeDef
		enableTypeDeduplication bool
		fieldName               string
		typeDef                 *TypeDef
		expectedSymbol          string
	}{
		{
			name:            "assigns symbol to type",
			assignedSymbols: map[string]*TypeDef{},
			typeDef: &TypeDef{
				Name:   "Example",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "id", Type: &TypeDef{Type: DataTypeString}}},
			},
			expectedSymbol: "Example",
		},
		{
			name:            "no-op when symbol already assigned",
			assignedSymbols: map[string]*TypeDef{},
			typeDef: &TypeDef{
				Name:       "Example",
				Type:       DataTypeClass,
				Extensions: &TypeDefExtensions{All: map[string]any{"Symbol": "PreExisting"}},
			},
			expectedSymbol: "PreExisting",
		},
		{
			name:            "uses fieldName when TypeDef name is empty",
			assignedSymbols: map[string]*TypeDef{},
			fieldName:       "fallback_name",
			typeDef: &TypeDef{
				Type:   DataTypeClass,
				Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
			},
			expectedSymbol: "FallbackName",
		},
		{
			name: "numeric suffix on collision",
			assignedSymbols: map[string]*TypeDef{
				"Config": {Name: "Config", Type: DataTypeClass, Fields: Fields{{Name: "a", Type: &TypeDef{Type: DataTypeString}}}},
			},
			typeDef: &TypeDef{
				Name:   "Config",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "b", Type: &TypeDef{Type: DataTypeInteger}}},
			},
			expectedSymbol: "Config1",
		},
		{
			name: "reuses symbol for structurally identical type with same name",
			assignedSymbols: map[string]*TypeDef{
				"Status": {Name: "Status", Type: DataTypeClass, Fields: Fields{{Name: "code", Type: &TypeDef{Type: DataTypeInteger}}}},
			},
			typeDef: &TypeDef{
				Name:   "Status",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "code", Type: &TypeDef{Type: DataTypeInteger}}},
			},
			expectedSymbol: "Status",
		},
		{
			name: "deduplication reuses symbol for differently named but identical types",
			assignedSymbols: map[string]*TypeDef{
				"Alpha": {Name: "Alpha", Type: DataTypeClass, Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}}},
			},
			enableTypeDeduplication: true,
			typeDef: &TypeDef{
				Name:   "Beta",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
			},
			expectedSymbol: "Alpha",
		},
		{
			name: "without deduplication differently named identical types get separate symbols",
			assignedSymbols: map[string]*TypeDef{
				"Alpha": {Name: "Alpha", Type: DataTypeClass, Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}}},
			},
			typeDef: &TypeDef{
				Name:   "Beta",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
			},
			expectedSymbol: "Beta",
		},
		{
			name:            "reserved name collision gets numeric suffix",
			assignedSymbols: map[string]*TypeDef{"Type": nil},
			typeDef: &TypeDef{
				Name:   "type",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "name", Type: &TypeDef{Type: DataTypeString}}},
			},
			expectedSymbol: "Type1",
		},
		{
			name:            "skips nil entries in assignedSymbols during dedup search",
			assignedSymbols: map[string]*TypeDef{"Reserved": nil},
			typeDef: &TypeDef{
				Name:   "Other",
				Type:   DataTypeClass,
				Fields: Fields{{Name: "id", Type: &TypeDef{Type: DataTypeString}}},
			},
			expectedSymbol: "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sm := newSymbolManager(tt.assignedSymbols, tt.enableTypeDeduplication)

			require.NoError(t, sm.assignSymbol(tt.fieldName, tt.typeDef))

			sym, ok := tt.typeDef.Extensions.Get("Symbol")
			require.True(t, ok)
			assert.Equal(t, tt.expectedSymbol, sym)
		})
	}
}

func TestSymbolManager_annotateRecursive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T, sm *symbolManager)
	}{
		{
			name: "nil TypeDef is safe",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				require.NoError(t, sm.annotateRecursive(nil, "field"))
			},
		},
		{
			name: "string type is not annotated",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				td := &TypeDef{Type: DataTypeString}

				require.NoError(t, sm.annotateRecursive(td, "field"))

				assert.Nil(t, td.Extensions)
			},
		},
		{
			name: "integer type is not annotated",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				td := &TypeDef{Type: DataTypeInteger}

				require.NoError(t, sm.annotateRecursive(td, "field"))

				assert.Nil(t, td.Extensions)
			},
		},
		{
			name: "boolean type is not annotated",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				td := &TypeDef{Type: DataTypeBoolean}

				require.NoError(t, sm.annotateRecursive(td, "field"))

				assert.Nil(t, td.Extensions)
			},
		},
		{
			name: "number type is not annotated",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				td := &TypeDef{Type: DataTypeNumber}

				require.NoError(t, sm.annotateRecursive(td, "field"))

				assert.Nil(t, td.Extensions)
			},
		},
		{
			name: "map recurses into ItemType",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				innerClass := &TypeDef{
					Name:   "Inner",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
				}

				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type:     DataTypeMap,
					ItemType: innerClass,
				}, "lookup"))

				sym, ok := innerClass.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Inner", sym)
			},
		},
		{
			name: "map with custom type config does not recurse",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				innerClass := &TypeDef{
					Name:   "Inner",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
				}

				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type:     DataTypeMap,
					ItemType: innerClass,
					Extensions: &TypeDefExtensions{
						All: map[string]any{"x-speakeasy-terraform-custom-type": true},
					},
				}, "lookup"))

				assert.Nil(t, innerClass.Extensions)
			},
		},
		{
			name: "array with nil ItemType is safe",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type:     DataTypeArray,
					ItemType: nil,
				}, "field"))
			},
		},
		{
			name: "array of primitive skips symbol",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				itemType := &TypeDef{Type: DataTypeString}

				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type:     DataTypeArray,
					ItemType: itemType,
				}, "field"))

				assert.Nil(t, itemType.Extensions)
			},
		},
		{
			name: "array of class assigns symbol and recurses into fields",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				nestedType := &TypeDef{
					Name:   "Nested",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "id", Type: &TypeDef{Type: DataTypeString}}},
				}
				itemType := &TypeDef{
					Name:   "Item",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "nested", Type: nestedType}},
				}

				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type:     DataTypeArray,
					ItemType: itemType,
				}, "items"))

				itemSym, ok := itemType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Item", itemSym)

				nestedSym, ok := nestedType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Nested", nestedSym)
			},
		},
		{
			name: "set of union assigns symbol and recurses into variants",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				variant := &TypeDef{
					Name:   "Var",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
				}
				unionType := &TypeDef{
					Name:            "Union",
					Type:            DataTypeUnion,
					AssociatedTypes: []*TypeDef{variant},
				}

				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type:     DataTypeSet,
					ItemType: unionType,
				}, "items"))

				unionSym, ok := unionType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Union", unionSym)

				variantSym, ok := variant.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Var", variantSym)
			},
		},
		{
			name: "array of array of class recurses through catch-all",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				innerClass := &TypeDef{
					Name:   "Deep",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
				}

				require.NoError(t, sm.annotateRecursive(&TypeDef{
					Type: DataTypeArray,
					ItemType: &TypeDef{
						Type:     DataTypeArray,
						ItemType: innerClass,
					},
				}, "matrix"))

				sym, ok := innerClass.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Deep", sym)
			},
		},
		{
			name: "class with empty fields still gets symbol",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				td := &TypeDef{
					Name:   "Empty",
					Type:   DataTypeClass,
					Fields: Fields{},
				}

				require.NoError(t, sm.annotateRecursive(td, "field"))

				sym, ok := td.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Empty", sym)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sm := newSymbolManager(make(map[string]*TypeDef), false)
			tt.run(t, sm)
		})
	}
}

func TestSymbolManager_annotateClassFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T, sm *symbolManager)
	}{
		{
			name: "recurses into each field",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				type1 := &TypeDef{
					Name:   "TypeA",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
				}
				type2 := &TypeDef{
					Name:   "TypeB",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "num", Type: &TypeDef{Type: DataTypeInteger}}},
				}

				require.NoError(t, sm.annotateClassFields(&TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "a", Type: type1},
						{Name: "b", Type: type2},
					},
				}))

				sym1, ok := type1.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "TypeA", sym1)

				sym2, ok := type2.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "TypeB", sym2)
			},
		},
		{
			name: "skips fields with MatchConfig Path",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				pathStr := "source"
				skippedType := &TypeDef{
					Name:   "Skipped",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
					Extensions: &TypeDefExtensions{
						All:         map[string]any{},
						MatchConfig: &extensions.MatchConfig{Path: &pathStr},
					},
				}

				require.NoError(t, sm.annotateClassFields(&TypeDef{
					Type: DataTypeClass,
					Fields: Fields{
						{Name: "alias", Type: skippedType},
					},
				}))

				_, hasSymbol := skippedType.Extensions.Get("Symbol")
				assert.False(t, hasSymbol)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sm := newSymbolManager(make(map[string]*TypeDef), false)
			tt.run(t, sm)
		})
	}
}

func TestSymbolManager_annotateUnionMembers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T, sm *symbolManager)
	}{
		{
			name: "annotates variants and hoisted fields",
			run: func(t *testing.T, sm *symbolManager) {
				t.Helper()
				variant := &TypeDef{
					Name:   "Opt",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "val", Type: &TypeDef{Type: DataTypeString}}},
				}
				hoistedType := &TypeDef{
					Name:   "Common",
					Type:   DataTypeClass,
					Fields: Fields{{Name: "id", Type: &TypeDef{Type: DataTypeString}}},
				}

				require.NoError(t, sm.annotateUnionMembers(&TypeDef{
					Name:            "MyUnion",
					Type:            DataTypeUnion,
					AssociatedTypes: []*TypeDef{variant},
					Fields: Fields{
						{Name: "common", Type: hoistedType},
					},
				}))

				variantSym, ok := variant.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Opt", variantSym)

				hoistedSym, ok := hoistedType.Extensions.Get("Symbol")
				require.True(t, ok)
				assert.Equal(t, "Common", hoistedSym)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sm := newSymbolManager(make(map[string]*TypeDef), false)
			tt.run(t, sm)
		})
	}
}
