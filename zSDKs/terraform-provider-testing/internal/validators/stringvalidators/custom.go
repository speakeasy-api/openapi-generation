package stringvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = StringCustomValidator{}

type StringCustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v StringCustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v StringCustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v StringCustomValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// let it through
}

func Custom() validator.String {
	return StringCustomValidator{}
}
