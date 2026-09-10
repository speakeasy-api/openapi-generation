package ast_post_processing

import (
	"iter"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

func OperationRequestTypes(operation *ast.Operation) iter.Seq[*ast.TypeDef] {
	return func(yield func(*ast.TypeDef) bool) {
		if operation.Request == nil {
			return
		}

		// Yield request field type
		if operation.Request.Field != nil && operation.Request.Field.Type != nil {
			if !yield(operation.Request.Field.Type) {
				return
			}
		}

		// Yield argument types
		for _, arg := range operation.Arguments.Sorted {
			if arg.Type != nil && !yield(arg.Type) {
				return
			}
		}

		// Yield request body type
		if operation.Request.RequestBody != nil && operation.Request.RequestBody.Type != nil {
			if !yield(operation.Request.RequestBody.Type) {
				return
			}
		}

		// Yield parameter types
		if operation.Request.Params != nil {
			for typeDef := range walkParams(operation.Request.Params) {
				if !yield(typeDef) {
					return
				}
			}
		}
	}
}

func walkParams(params *ast.RequestParams) iter.Seq[*ast.TypeDef] {
	return func(yield func(*ast.TypeDef) bool) {
		for _, param := range params.QueryParams {
			if param.Field != nil && param.Field.Type != nil && !yield(param.Field.Type) {
				return
			}
		}
		for _, param := range params.PathParams {
			if param.Field != nil && param.Field.Type != nil && !yield(param.Field.Type) {
				return
			}
		}
		for _, param := range params.HeaderParams {
			if param.Field != nil && param.Field.Type != nil && !yield(param.Field.Type) {
				return
			}
		}
	}
}
