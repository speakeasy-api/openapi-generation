package listvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.List = ListCustomValidator{}

type ListCustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v ListCustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v ListCustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v ListCustomValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	// let it through
}

func Custom() validator.List {
	return ListCustomValidator{}
}
