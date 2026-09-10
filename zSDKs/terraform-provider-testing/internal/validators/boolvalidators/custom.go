package boolvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Bool = BoolCustomValidator{}

type BoolCustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v BoolCustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v BoolCustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v BoolCustomValidator) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	// let it through
}

func Custom() validator.Bool {
	return BoolCustomValidator{}
}
