package ast_post_processing

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// MarkUsedInSecurity marks all types that are used in security by iterating through
// global and operation-level security
func MarkUsedInSecurity(a *ast.AST) {
	visited := map[*ast.TypeDef]bool{}

	// Mark global security types
	if a.MainSDK.Security != nil {
		markTypeAsUsedBySecurity(a.MainSDK.Security.Type, visited)
	}

	// Mark operation security types
	for operation := range a.MainSDK.WalkOperations() {
		// Mark operation-specific security
		if operation.Security != nil {
			markTypeAsUsedBySecurity(operation.Security.Type, visited)
		}

		// Mark global security when referenced by operations
		if operation.GlobalSecurity != nil {
			markTypeAsUsedBySecurity(operation.GlobalSecurity.Type, visited)
		}
	}
}

func markTypeAsUsedBySecurity(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	if typeDef == nil {
		return
	}

	if visited[typeDef] {
		return
	}

	visited[typeDef] = true
	typeDef.UsedInSecurity = true

	for _, child := range typeDef.Children {
		markTypeAsUsedBySecurity(child, visited)
	}
}
