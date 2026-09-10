package ast_post_processing

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkUsedInRequest marks all types that are used in requests by iterating through operations
func MarkUsedInRequest(a *ast.AST) {
	visited := map[*ast.TypeDef]bool{}
	for operation := range a.MainSDK.WalkOperations() {
		for typeDef := range OperationRequestTypes(operation) {
			markTypeAsUsedByRequest(typeDef, visited)
		}
	}
}

func markTypeAsUsedByRequest(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	if typeDef == nil {
		return
	}

	if visited[typeDef] {
		return
	}

	visited[typeDef] = true
	typeDef.UsedInRequest = true

	for _, child := range typeDef.Children {
		markTypeAsUsedByRequest(child, visited)
	}
}
