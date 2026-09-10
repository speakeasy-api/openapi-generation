package ast_post_processing

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkUsedInResponse marks all types that are used in responses by iterating through operations
func MarkUsedInResponse(a *ast.AST) {

	visited := map[*ast.TypeDef]bool{}

	for operation := range a.MainSDK.WalkOperations() {
		markOperationResponseTypes(operation, visited)
	}
}

// markOperationResponseTypes marks all types used in an operation's response
func markOperationResponseTypes(operation *ast.Operation, visited map[*ast.TypeDef]bool) {
	for typeDef := range OperationResponseTypes(operation) {
		markTypeAsUsedByResponse(typeDef, visited)
	}
}

func markTypeAsUsedByResponse(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	if typeDef == nil {
		return
	}

	if visited[typeDef] {
		return
	}

	visited[typeDef] = true
	typeDef.UsedInResponse = true

	for _, child := range typeDef.Children {
		markTypeAsUsedByResponse(child, visited)
	}
}
