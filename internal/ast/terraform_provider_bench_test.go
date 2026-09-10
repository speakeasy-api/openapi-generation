package ast

import (
	"context"
	"fmt"
	"testing"
)

// buildBenchTypeDef constructs a moderately deep type tree with a mix of
// classes, arrays, maps, and unions. The depth parameter controls nesting
// levels, and idx provides unique naming across entities.
func buildBenchTypeDef(entityName string, idx int, depth int) *TypeDef {
	return &TypeDef{
		Type:   DataTypeClass,
		Fields: buildBenchFields(entityName, idx, depth, 0),
	}
}

// buildBenchFields creates a set of fields for a class TypeDef at the given
// nesting level. Each level produces 5 fields: a string primitive, a nested
// class, an array of classes, a map of classes, and a union of two classes.
func buildBenchFields(entityName string, entityIdx, maxDepth, currentDepth int) Fields {
	prefix := fmt.Sprintf("e%d_d%d", entityIdx, currentDepth)

	fields := make(Fields, 0, 7)
	fields = append(fields,
		&FieldDef{Name: prefix + "_name", Type: &TypeDef{Type: DataTypeString}},
		&FieldDef{Name: prefix + "_count", Type: &TypeDef{Type: DataTypeInteger}},
		&FieldDef{Name: prefix + "_enabled", Type: &TypeDef{Type: DataTypeBoolean}},
	)

	if currentDepth >= maxDepth {
		return fields
	}

	// Nested class with recursive children.
	nestedClass := &TypeDef{
		Name:   fmt.Sprintf("%sNestedClass%d", entityName, currentDepth),
		Type:   DataTypeClass,
		Fields: buildBenchFields(entityName, entityIdx, maxDepth, currentDepth+1),
	}
	fields = append(fields, &FieldDef{
		Name: prefix + "_nested",
		Type: nestedClass,
	})

	// Array of class items.
	arrayItemClass := &TypeDef{
		Name: fmt.Sprintf("%sArrayItem%d", entityName, currentDepth),
		Type: DataTypeClass,
		Fields: Fields{
			{Name: "item_id", Type: &TypeDef{Type: DataTypeString}},
			{Name: "item_value", Type: &TypeDef{Type: DataTypeInteger}},
		},
	}
	// Add deeper nesting inside array items if we have room.
	if currentDepth+1 < maxDepth {
		arrayItemClass.Fields = append(arrayItemClass.Fields, &FieldDef{
			Name: "item_detail",
			Type: &TypeDef{
				Name: fmt.Sprintf("%sArrayItemDetail%d", entityName, currentDepth),
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "detail_key", Type: &TypeDef{Type: DataTypeString}},
					{Name: "detail_val", Type: &TypeDef{Type: DataTypeNumber}},
				},
			},
		})
	}
	fields = append(fields, &FieldDef{
		Name: prefix + "_items",
		Type: &TypeDef{
			Type:     DataTypeArray,
			ItemType: arrayItemClass,
		},
	})

	// Map of class values.
	mapValueClass := &TypeDef{
		Name: fmt.Sprintf("%sMapValue%d", entityName, currentDepth),
		Type: DataTypeClass,
		Fields: Fields{
			{Name: "map_key", Type: &TypeDef{Type: DataTypeString}},
			{Name: "map_data", Type: &TypeDef{Type: DataTypeString}},
		},
	}
	fields = append(fields, &FieldDef{
		Name: prefix + "_lookup",
		Type: &TypeDef{
			Type:     DataTypeMap,
			ItemType: mapValueClass,
		},
	})

	// Union with two class variants.
	variant1 := &TypeDef{
		Name: fmt.Sprintf("%sVariantA%d", entityName, currentDepth),
		Type: DataTypeClass,
		Fields: Fields{
			{Name: "a_field", Type: &TypeDef{Type: DataTypeString}},
		},
	}
	variant2 := &TypeDef{
		Name: fmt.Sprintf("%sVariantB%d", entityName, currentDepth),
		Type: DataTypeClass,
		Fields: Fields{
			{Name: "b_field", Type: &TypeDef{Type: DataTypeInteger}},
		},
	}
	fields = append(fields, &FieldDef{
		Name: prefix + "_choice",
		Type: &TypeDef{
			Name:            fmt.Sprintf("%sUnion%d", entityName, currentDepth),
			Type:            DataTypeUnion,
			AssociatedTypes: []*TypeDef{variant1, variant2},
		},
	})

	return fields
}

// clearSymbolExtensions recursively removes Symbol entries from all TypeDef
// Extensions in the tree, ensuring each benchmark iteration starts fresh.
// assignSymbol is a no-op when a Symbol already exists, so this reset is
// necessary for accurate measurement.
func clearSymbolExtensions(t *TypeDef) {
	if t == nil {
		return
	}

	if t.Extensions != nil {
		t.Extensions.Remove("Symbol")

		// If the All map is now empty, nil out Extensions entirely to match
		// the initial state before AnnotateTerraformSymbols runs.
		if len(t.Extensions.All) == 0 {
			t.Extensions = nil
		}
	}

	for _, f := range t.Fields {
		clearSymbolExtensions(f.Type)
	}

	for _, at := range t.AssociatedTypes {
		clearSymbolExtensions(at)
	}

	clearSymbolExtensions(t.ItemType)
}

// buildBenchProvider constructs a TerraformProvider with n entities distributed
// across all four resource types, each with a type tree of the given depth.
func buildBenchProvider(n, depth int) *TerraformProvider {
	nActions := n / 4
	nManaged := n / 4
	nData := n / 4
	nEphemeral := n - nActions - nManaged - nData

	provider := NewTerraformProvider()

	provider.Actions = make([]*TerraformAction, nActions)
	for i := range nActions {
		provider.Actions[i] = &TerraformAction{
			Name:          fmt.Sprintf("action_%d", i),
			Operations:    testActionOperations(),
			SchemaTypeDef: buildBenchTypeDef(fmt.Sprintf("Action%d", i), i, depth),
		}
	}

	provider.ManagedResources = make([]*TerraformManagedResource, nManaged)
	for i := range nManaged {
		provider.ManagedResources[i] = &TerraformManagedResource{
			Name:          fmt.Sprintf("managed_%d", i),
			Operations:    testManagedResourceOperations(),
			SchemaTypeDef: buildBenchTypeDef(fmt.Sprintf("Managed%d", i), nActions+i, depth),
		}
	}

	provider.DataResources = make([]*TerraformDataResource, nData)
	for i := range nData {
		provider.DataResources[i] = &TerraformDataResource{
			Name:          fmt.Sprintf("data_%d", i),
			Operations:    testDataResourceOperations(),
			SchemaTypeDef: buildBenchTypeDef(fmt.Sprintf("Data%d", i), nActions+nManaged+i, depth),
		}
	}

	provider.EphemeralResources = make([]*TerraformEphemeralResource, nEphemeral)
	for i := range nEphemeral {
		provider.EphemeralResources[i] = &TerraformEphemeralResource{
			Name:          fmt.Sprintf("ephemeral_%d", i),
			Operations:    testEphemeralResourceOperations(),
			SchemaTypeDef: buildBenchTypeDef(fmt.Sprintf("Ephemeral%d", i), nActions+nManaged+nData+i, depth),
		}
	}

	return provider
}

// clearProviderSymbols clears all Symbol extensions from a provider's entities,
// ensuring each benchmark iteration starts fresh.
func clearProviderSymbols(p *TerraformProvider) {
	for _, a := range p.Actions {
		clearSymbolExtensions(a.SchemaTypeDef)
	}
	for _, mr := range p.ManagedResources {
		clearSymbolExtensions(mr.SchemaTypeDef)
	}
	for _, dr := range p.DataResources {
		clearSymbolExtensions(dr.SchemaTypeDef)
	}
	for _, er := range p.EphemeralResources {
		clearSymbolExtensions(er.SchemaTypeDef)
	}
}

func BenchmarkAnnotateTerraformSymbols(b *testing.B) {
	scales := []int{10, 50, 100}

	for _, n := range scales {
		provider := buildBenchProvider(n, 4)
		ctx := context.Background()

		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			for range b.N {
				clearProviderSymbols(provider)

				if err := provider.AnnotateTerraformSymbols(ctx, false); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
