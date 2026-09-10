package ast_post_processing

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// PropagateIncludeExtension propagates the x-speakeasy-include: true extension down the tree
// to everything reachable from types that have this extension set
func PropagateIncludeExtension(types []*ast.TypeDef) {
	visited := map[*ast.TypeDef]bool{}

	// First pass: find all types with x-speakeasy-include: true
	for _, typeDef := range types {
		if hasIncludeExtension(typeDef) {
			propagateIncludeToChildren(typeDef, visited)
		}
	}
}

// hasIncludeExtension checks if a TypeDef has x-speakeasy-include: true
func hasIncludeExtension(typeDef *ast.TypeDef) bool {
	if typeDef == nil || typeDef.Extensions == nil || typeDef.Extensions.All == nil {
		return false
	}

	includeVal, ok := typeDef.Extensions.All["x-speakeasy-include"]
	if !ok {
		return false
	}

	// Check if it's a boolean true value
	if boolVal, ok := includeVal.(bool); ok {
		return boolVal
	}

	return false
}

// propagateIncludeToChildren recursively propagates x-speakeasy-include: true to all children
func propagateIncludeToChildren(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	if typeDef == nil {
		return
	}

	if visited[typeDef] {
		return
	}

	visited[typeDef] = true

	// Ensure the current type has the extension set
	ensureIncludeExtension(typeDef)

	// Propagate to all children using the Children iterator
	for _, child := range typeDef.Children {
		propagateIncludeToChildren(child, visited)
	}
}

// ensureIncludeExtension ensures that x-speakeasy-include: true is set on the TypeDef
func ensureIncludeExtension(typeDef *ast.TypeDef) {
	if typeDef == nil {
		return
	}

	if typeDef.Extensions == nil {
		typeDef.Extensions = &ast.TypeDefExtensions{
			All: make(map[string]any),
		}
	}

	if typeDef.Extensions.All == nil {
		typeDef.Extensions.All = make(map[string]any)
	}

	typeDef.Extensions.All["x-speakeasy-include"] = true
}
