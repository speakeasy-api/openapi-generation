package ast_post_processing

import (
	"iter"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

func OperationResponseTypes(operation *ast.Operation) iter.Seq[*ast.TypeDef] {
	return func(yield func(*ast.TypeDef) bool) {
		if operation.Response == nil {
			return
		}

		// Yield the response type if it exists
		if operation.Response.Type != nil {
			if !yield(operation.Response.Type) {
				return
			}
		}

		// Yield all response content types
		for _, response := range operation.Response.Responses {
			for _, content := range response.Content {
				if content.Content != nil && content.Content.Type != nil {
					if !yield(content.Content.Type) {
						return
					}
				}
			}
		}
	}
}
