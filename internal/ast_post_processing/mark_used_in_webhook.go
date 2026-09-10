package ast_post_processing

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkUsedInWebhook marks all types that are used in webhooks by iterating through webhook operations
func MarkUsedInWebhook(a *ast.AST) {
	visited := map[*ast.TypeDef]bool{}

	for _, ops := range a.Webhooks.All() {
		for i := range ops {
			operation := &ops[i]

			for typeDef := range OperationRequestTypes(operation) {
				markTypeAsUsedByWebhook(typeDef, visited)
			}

			for typeDef := range OperationResponseTypes(operation) {
				markTypeAsUsedByWebhook(typeDef, visited)
			}
		}
	}
}

func markTypeAsUsedByWebhook(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	if typeDef == nil {
		return
	}

	if visited[typeDef] {
		return
	}

	visited[typeDef] = true
	typeDef.UsedInWebhook = true

	for _, child := range typeDef.Children {
		markTypeAsUsedByWebhook(child, visited)
	}
}
