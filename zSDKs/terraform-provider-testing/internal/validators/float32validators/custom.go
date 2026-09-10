package float32validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Float32 = Float32CustomValidator{}

type Float32CustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v Float32CustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v Float32CustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v Float32CustomValidator) ValidateFloat32(ctx context.Context, req validator.Float32Request, resp *validator.Float32Response) {
	// let it through
}

func Custom() validator.Float32 {
	return Float32CustomValidator{}
}
