package float64planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Float64 = Float64CustomPlanModifier{}

type Float64CustomPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v Float64CustomPlanModifier) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v Float64CustomPlanModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the plan modification.
func (v Float64CustomPlanModifier) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement plan modifier Custom logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func Custom() planmodifier.Float64 {
	return Float64CustomPlanModifier{}
}
