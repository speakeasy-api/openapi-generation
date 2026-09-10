package ast_post_processing

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkUsedInCallback marks all types that are used in callbacks by iterating through operations with callbacks
func MarkUsedInCallback(a *ast.AST) {
	visited := map[*ast.TypeDef]bool{}
	for operation := range a.MainSDK.WalkOperations() {
		if len(operation.Callbacks) == 0 {
			continue
		}

		// Mark each callback type and all its children
		for _, callback := range operation.Callbacks {
			markTypeAsUsedByCallback(callback, visited)
		}
	}
}

func markTypeAsUsedByCallback(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	if typeDef == nil {
		return
	}

	if visited[typeDef] {
		return
	}

	visited[typeDef] = true
	typeDef.UsedInCallback = true

	for _, child := range typeDef.Children {
		markTypeAsUsedByCallback(child, visited)
	}
}
