package int32validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int32 = Int32CustomValidator{}

type Int32CustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v Int32CustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v Int32CustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v Int32CustomValidator) ValidateInt32(ctx context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	// let it through
}

func Custom() validator.Int32 {
	return Int32CustomValidator{}
}
