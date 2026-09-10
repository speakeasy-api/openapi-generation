package tests

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/speakeasy-api/openapi/arazzo"
)

type GetOperationsByIDOptions struct {
	IncludeWebhooks bool
}

func GetOperationsByID(a *ast.AST, opts *GetOperationsByIDOptions) map[string][]*ast.Operation {
	operations := map[string][]*ast.Operation{}

	if opts == nil {
		opts = &GetOperationsByIDOptions{}
	}

	_ = a.Walk(func(node ast.Node, parents []ast.Node, a *ast.AST) error {
		return node.Match(ast.Matchers{
			Operation: func(o *ast.Operation) error {
				if !opts.IncludeWebhooks && o.Webhook != nil {
					return nil
				}

				ops, ok := operations[o.OriginalID]
				if !ok {
					ops = []*ast.Operation{}
				}
				ops = append(ops, o)
				operations[o.OriginalID] = ops

				return nil
			},
		})
	})

	return operations
}

func isOpRequestBodyRequired(op *ast.Operation) bool {
	return op.Request != nil && op.Request.IsRequestBodyRequired
}

func IsOperationSupported(op *ast.Operation, cfg *configuration.Config) bool {
	for _, response := range op.Response.Responses {
		if response.Error {
			continue
		}

		for _, content := range response.Content {
			if contenttypes.IsEventStream(content.ContentType) && !cfg.Generation.Tests.SkipResponseBodyAssertions {
				return false
			}
		}
	}

	return true
}

func isTestingDisabledForWorkflow(workflow *arazzo.Workflow) bool {
	if workflow == nil || workflow.Extensions == nil {
		return false
	}

	v, ok := workflow.Extensions.Get("x-speakeasy-test")
	return ok && v.Value == "false"
}
