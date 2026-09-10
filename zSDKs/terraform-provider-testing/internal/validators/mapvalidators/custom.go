package mapvalidators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Map = MapCustomValidator{}

type MapCustomValidator struct{}

// Description describes the validation in plain text formatting.
func (v MapCustomValidator) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v MapCustomValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the validation.
func (v MapCustomValidator) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	// let it through
}

func Custom() validator.Map {
	return MapCustomValidator{}
}
