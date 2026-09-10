package changes

import (
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// diffArguments compares the arguments of two operations using DiffTypeDefs
func diffArguments(a, b *ast.Operation) TypeDefDiffResult {
	// Also compare Arguments structure if present
	if a.Arguments == nil && b.Arguments == nil {
		return TypeDefDiffResult{Equal: true}
	}

	if a.Arguments == nil || b.Arguments == nil {
		return TypeDefDiffResult{
			Equal:      false,
			Reason:     DiffReasonKind,
			IsBreaking: true,
		}
	}

	// For Arguments, we need to construct a synthetic TypeDef that represents
	// the argument structure, then use DiffTypeDefs to compare them
	aArgType := buildArgumentsTypeDef(a.Arguments)
	bArgType := buildArgumentsTypeDef(b.Arguments)

	params := DiffTypeDefsParams{
		A:         aArgType,
		B:         bArgType,
		IsRequest: true,
	}
	result := DiffTypeDefs(params)

	return result
}

// buildArgumentsTypeDef creates a synthetic TypeDef from Arguments for comparison
func buildArgumentsTypeDef(args *ast.Arguments) *ast.TypeDef {
	if args == nil {
		return nil
	}

	// Create a class-like type with all argument fields combined
	return &ast.TypeDef{
		Type:   ast.DataTypeClass,
		Fields: slices.Clone(args.Sorted),
	}
}
