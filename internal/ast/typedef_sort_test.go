package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type runOptions struct {
	hoistCyclicUnions bool
}

func assertTopologicalSort(t *testing.T, types TypeDefs, hoistCyclicUnions bool, expected []string) {
	t.Helper()
	types = TopologicalSortTypeDefs(types, hoistCyclicUnions)
	names := make([]string, len(types))
	for i, typ := range types {
		names[i] = typ.Name
	}
	assert.Equal(t, expected, names)
}

func run(t *testing.T, fn func(t *testing.T, opt runOptions)) {
	t.Helper()

	for _, opt := range []runOptions{
		{hoistCyclicUnions: false},
		{hoistCyclicUnions: true},
	} {
		t.Run(fmt.Sprintf("%+v", opt), func(t *testing.T) {
			fn(t, opt)
		})
	}
}

func TestTopologicalSortTypeDefs_NoDepedencies(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		type1 := &TypeDef{Name: "Type1", Type: DataTypeString}
		type2 := &TypeDef{Name: "Type2", Type: DataTypeInteger}
		type3 := &TypeDef{Name: "Type3", Type: DataTypeBoolean}

		types := TypeDefs{type1, type2, type3}
		sorted := TopologicalSortTypeDefs(types, opt.hoistCyclicUnions)

		assert.Equal(t, types, sorted, "Types with no dependencies should maintain original order")
	})
}

func TestTopologicalSortTypeDefs_UnionDependency(t *testing.T) {
	stringType := &TypeDef{Name: "StringType", Type: DataTypeString}
	intType := &TypeDef{Name: "IntType", Type: DataTypeInteger}
	unionType := &TypeDef{
		Name:            "UnionType",
		Type:            DataTypeUnion,
		AssociatedTypes: TypeDefs{stringType, intType},
	}

	types := TypeDefs{unionType, stringType, intType}
	sorted := TopologicalSortTypeDefs(types, false)

	unionIndex := -1
	stringIndex := -1
	intIndex := -1
	for i, typ := range sorted {
		switch typ.Name {
		case "UnionType":
			unionIndex = i
		case "StringType":
			stringIndex = i
		case "IntType":
			intIndex = i
		}
	}

	assert.Less(t, stringIndex, unionIndex, "StringType should come before UnionType")
	assert.Less(t, intIndex, unionIndex, "IntType should come before UnionType")
}

func TestTopologicalSortTypeDefs_FieldsDependency(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		fieldType := &TypeDef{Name: "FieldType", Type: DataTypeString}
		classType := &TypeDef{
			Name: "ClassType",
			Type: DataTypeClass,
			Fields: []*FieldDef{
				{Name: "field1", Type: fieldType},
			},
		}

		types := TypeDefs{classType, fieldType}
		sorted := TopologicalSortTypeDefs(types, opt.hoistCyclicUnions)

		classIndex := -1
		fieldIndex := -1
		for i, typ := range sorted {
			switch typ.Name {
			case "ClassType":
				classIndex = i
			case "FieldType":
				fieldIndex = i
			}
		}

		assert.Less(t, fieldIndex, classIndex, "FieldType should come before ClassType")
	})
}

func TestTopologicalSortTypeDefs_ItemTypeDependency(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		itemType := &TypeDef{Name: "ItemType", Type: DataTypeString}
		arrayType := &TypeDef{
			Name:     "ArrayType",
			Type:     DataTypeArray,
			ItemType: itemType,
		}
		mapType := &TypeDef{
			Name:     "MapType",
			Type:     DataTypeMap,
			ItemType: itemType,
		}

		types := TypeDefs{arrayType, mapType, itemType}
		sorted := TopologicalSortTypeDefs(types, opt.hoistCyclicUnions)

		arrayIndex := -1
		mapIndex := -1
		itemIndex := -1
		for i, typ := range sorted {
			switch typ.Name {
			case "ArrayType":
				arrayIndex = i
			case "MapType":
				mapIndex = i
			case "ItemType":
				itemIndex = i
			}
		}

		assert.Less(t, itemIndex, arrayIndex, "ItemType should come before ArrayType")
		assert.Less(t, itemIndex, mapIndex, "ItemType should come before MapType")
	})
}

func TestTopologicalSortTypeDefs_CircularDependency(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		type1 := &TypeDef{Name: "Type1", Type: DataTypeClass}
		type2 := &TypeDef{Name: "Type2", Type: DataTypeClass}

		type1.Fields = []*FieldDef{{Name: "field", Type: type2}}
		type2.Fields = []*FieldDef{{Name: "field", Type: type1}}

		types := TypeDefs{type1, type2}
		sorted := TopologicalSortTypeDefs(types, opt.hoistCyclicUnions)

		assert.Len(t, sorted, 2, "Should return all types even with circular dependency")

		type1Found := false
		type2Found := false
		for _, typ := range sorted {
			if typ.Name == "Type1" {
				type1Found = true
			}
			if typ.Name == "Type2" {
				type2Found = true
			}
		}
		assert.True(t, type1Found, "Type1 should be in result")
		assert.True(t, type2Found, "Type2 should be in result")
	})
}

func TestTopologicalSortTypeDefs_ComplexDependencies(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		stringType := &TypeDef{Name: "StringType", Type: DataTypeString}
		numberType := &TypeDef{Name: "NumberType", Type: DataTypeNumber}
		enumType := &TypeDef{Name: "EnumType", Type: DataTypeEnum, Enum: &Enum{
			Type: stringType,
		}}
		unionType := &TypeDef{
			Name:            "UnionType",
			Type:            DataTypeUnion,
			AssociatedTypes: TypeDefs{stringType, numberType},
		}
		arrayType := &TypeDef{
			Name:     "ArrayType",
			Type:     DataTypeArray,
			ItemType: unionType,
		}
		classType := &TypeDef{
			Name: "ClassType",
			Type: DataTypeClass,
			Fields: []*FieldDef{
				{Name: "arrayField", Type: arrayType},
				{Name: "unionField", Type: unionType},
			},
		}

		types := TypeDefs{classType, stringType, numberType, arrayType, unionType, enumType}
		sorted := TopologicalSortTypeDefs(types, opt.hoistCyclicUnions)

		indices := make(map[string]int)
		for i, typ := range sorted {
			indices[typ.Name] = i
		}

		assert.Less(t, indices["StringType"], indices["UnionType"], "StringType should come before UnionType")
		assert.Less(t, indices["NumberType"], indices["UnionType"], "NumberType should come before UnionType")
		assert.Less(t, indices["UnionType"], indices["ArrayType"], "UnionType should come before ArrayType")
		assert.Less(t, indices["ArrayType"], indices["ClassType"], "ArrayType should come before ClassType")
		assert.Less(t, indices["UnionType"], indices["ClassType"], "UnionType should come before ClassType")
	})
}

func TestTopologicalSortTypeDefs_EmptyInput(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		sorted := TopologicalSortTypeDefs(TypeDefs{}, opt.hoistCyclicUnions)
		assert.Empty(t, sorted, "Empty input should return empty output")

		sorted = TopologicalSortTypeDefs(nil, opt.hoistCyclicUnions)
		assert.Nil(t, sorted, "Nil input should return nil")
	})
}

func TestTopologicalSortTypeDefs_HoistCyclicUnions(t *testing.T) {
	firstType := &TypeDef{Name: "FirstType", Type: DataTypeBoolean}

	// non-cyclic union type
	stringVariant := &TypeDef{Name: "StringVariant", Type: DataTypeString}
	intVariant := &TypeDef{Name: "IntVariant", Type: DataTypeInteger}
	plainUnion := &TypeDef{
		Name:            "PlainUnion",
		Type:            DataTypeUnion,
		AssociatedTypes: TypeDefs{stringVariant, intVariant},
	}

	// cyclic union: RecursiveUnion -> ArrayVariant -> RecursiveUnion
	textVariant := &TypeDef{Name: "TextVariant", Type: DataTypeClass}
	recursiveUnion := &TypeDef{Name: "RecursiveUnion", Type: DataTypeUnion}
	arrayVariant := &TypeDef{
		Name: "ArrayVariant",
		Type: DataTypeClass,
		Fields: []*FieldDef{
			{Name: "itemType", Type: recursiveUnion},
		},
	}
	recursiveUnion.AssociatedTypes = TypeDefs{textVariant, arrayVariant}

	types := TypeDefs{
		firstType,
		arrayVariant,
		recursiveUnion,
		textVariant,
		stringVariant,
		plainUnion,
		intVariant,
	}

	t.Run("hoistCyclicUnions:true only hoists cyclic unions", func(t *testing.T) {
		// RecursiveUnion is hoisted to be DFS root, which forces its variants
		// (TextVariant, ArrayVariant) to be placed before the union itself.
		// PlainUnion is non-cyclic so it does not get moved ahead of First.
		assertTopologicalSort(t, types, true, []string{
			"TextVariant",
			"ArrayVariant",
			"RecursiveUnion",
			"FirstType",
			"StringVariant",
			"IntVariant",
			"PlainUnion",
		})
	})

	t.Run("hoistCyclicUnions:false places cyclic union before its variant", func(t *testing.T) {
		// Default DFS recurses into RecursiveUnion by rooting from ArrayVariant (input order)
		// so RecursiveUnion is placed before the ArrayVariant class.
		// None of the unions are hoisted ahead of FirstType.
		assertTopologicalSort(t, types, false, []string{
			"FirstType",
			"TextVariant",
			"RecursiveUnion",
			"ArrayVariant",
			"StringVariant",
			"IntVariant",
			"PlainUnion",
		})
	})
}

func TestTopologicalSortTypeDefs_HoistDeepRecursiveUnions(t *testing.T) {
	// # Indirect cyclic union
	// components:
	//   classVariant:
	//     type: object
	//     properties:
	//       items:
	//         $ref: arrayWrapper
	//   unionType:
	//     oneOf:
	//       - $ref: classVariant
	//       - $ref: stringVariant
	//   arrayWrapper:
	//     type: array
	//     items:
	//        $ref: unionType
	unionType := &TypeDef{Name: "DeepRecursiveUnion", Type: DataTypeUnion}
	arrayWrapper := &TypeDef{
		Name:     "ArrayWrapper",
		Type:     DataTypeArray,
		ItemType: unionType,
	}
	classVariant := &TypeDef{
		Name: "ClassVariant",
		Type: DataTypeClass,
		Fields: []*FieldDef{
			{Name: "items", Type: arrayWrapper},
		},
	}
	stringVariant := &TypeDef{Name: "StringVariant", Type: DataTypeString}
	unionType.AssociatedTypes = TypeDefs{classVariant, stringVariant}

	types := TypeDefs{stringVariant, classVariant, unionType}
	assertTopologicalSort(t, types, true, []string{"ClassVariant", "StringVariant", "DeepRecursiveUnion"})
}

func TestTopologicalSortTypeDefs_OnlyExternalTypes(t *testing.T) {
	run(t, func(t *testing.T, opt runOptions) {
		t.Helper()
		externalType := &TypeDef{Name: "ExternalType", Type: DataTypeString}
		classType := &TypeDef{
			Name: "ClassType",
			Type: DataTypeClass,
			Fields: []*FieldDef{
				{Name: "field", Type: externalType},
			},
		}

		types := TypeDefs{classType}
		assertTopologicalSort(t, types, opt.hoistCyclicUnions, []string{"ClassType"})
	})
}
