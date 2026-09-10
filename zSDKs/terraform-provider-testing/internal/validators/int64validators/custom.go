package int64validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int64 = Int64CustomValidator{}

type Int64CustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v Int64CustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v Int64CustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v Int64CustomValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	// let it through
}

func Custom() validator.Int64 {
	return Int64CustomValidator{}
}
