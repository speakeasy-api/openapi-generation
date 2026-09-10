package float64validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Float64 = Float64CustomValidator{}

type Float64CustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v Float64CustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v Float64CustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v Float64CustomValidator) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	// let it through
}

func Custom() validator.Float64 {
	return Float64CustomValidator{}
}
