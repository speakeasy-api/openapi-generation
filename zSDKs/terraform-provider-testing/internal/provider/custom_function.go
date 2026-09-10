package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ function.Function = (*CustomFunction)(nil)

func NewCustomFunction() function.Function {
	return &CustomFunction{}
}

// CustomFunction defines a stub custom function type implementation.
// This implementation is intentionally minimal as this is used for verifying
// the inclusion of a custom function in the generated provider.
type CustomFunction struct{}

func (f *CustomFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "custom"
}

func (f *CustomFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "A custom function stub for testing",
		Description: "This custom function implementation intentionally has no real logic. It is used for verifying the inclusion of custom functions in the generated provider.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "input",
				Description: "An input string",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *CustomFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var input string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &input))
	if resp.Error != nil {
		return
	}

	// Simply return the input as-is for this stub implementation
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, input))
}
