package infer_union_discriminator

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// getOriginalFieldName returns the original field name or falls back to the name
func getOriginalFieldName(field *ast.FieldDef) string {
	if field.OriginalName != "" {
		return field.OriginalName
	}
	return field.Name
}

// findFieldByName finds a field by its original name
func findFieldByName(typeDef *ast.TypeDef, name string) *ast.FieldDef {
	for _, field := range typeDef.Fields {
		if field.OriginalName == "" && field.Name == name {
			return field
		} else if field.OriginalName == name {
			return field
		}
	}
	return nil
}

// getPossibleConsts returns the set of possible values for a field
// Returns nil for unconstrained primitives
func getPossibleConsts(field *ast.FieldDef) []*ast.AnyValue {
	// Check for const value
	if field.Const != nil {
		return []*ast.AnyValue{field.Const}
	}

	// Check for enum (single or multi-value)
	if field.Type.Type == ast.DataTypeEnum && field.Type.Enum != nil && len(field.Type.Enum.Values) > 0 {
		values := make([]*ast.AnyValue, len(field.Type.Enum.Values))
		for i, v := range field.Type.Enum.Values {
			values[i] = &ast.AnyValue{Value: v}
		}
		return values
	}

	// Unconstrained primitive
	return nil
}

// hasOverlap checks if two value sets have any common elements
func hasOverlap(valuesA, valuesB []*ast.AnyValue) bool {
	setA := make(map[string]bool, len(valuesA))
	for _, val := range valuesA {
		setA[fmt.Sprint(val.Value)] = true
	}
	for _, val := range valuesB {
		if setA[fmt.Sprint(val.Value)] {
			return true
		}
	}
	return false
}

// Helper functions to identify field characteristics
func isSingleEnum(field *ast.FieldDef) bool {
	return field.Type.Type == ast.DataTypeEnum && field.Type.Enum != nil && len(field.Type.Enum.Values) == 1
}

func isMultiEnum(field *ast.FieldDef) bool {
	return field.Type.Type == ast.DataTypeEnum && field.Type.Enum != nil && len(field.Type.Enum.Values) > 1
}

// parsePrimitiveDataType returns the data type of a field, or empty if not a valid discriminator type
func parsePrimitiveDataType(field *ast.FieldDef) ast.DataType {
	t := field.Type
	if t.Type == ast.DataTypeEnum {
		if t.Enum == nil || t.Enum.Type == nil || t.Enum.Open {
			return ""
		}
		return t.Enum.Type.Type
	}
	if t.IsPrimitive() {
		return t.Type
	}
	return ""
}

// getSpecificity returns the specificity of a field based on its characteristics
func getSpecificity(field *ast.FieldDef) DiscriminatorSpecificity {
	if field.Const != nil {
		return DiscriminatorSpecificityConst
	}

	if isSingleEnum(field) {
		return DiscriminatorSpecificitySingle
	}

	if isMultiEnum(field) {
		return DiscriminatorSpecificityMulti
	}

	return DiscriminatorSpecificityType
}

// better returns the better discriminator based on specificity
// If a is nil, returns b. Otherwise returns the discriminator with lower (better) specificity
// Lower specificity values are more specific (better)
func better(a, b *discriminator) *discriminator {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if b.Specificity < a.Specificity {
		return b
	}
	return a
}
