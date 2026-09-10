package mapplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Map = MapCustomPlanModifier{}

type MapCustomPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v MapCustomPlanModifier) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v MapCustomPlanModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the plan modification.
func (v MapCustomPlanModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement plan modifier Custom logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func Custom() planmodifier.Map {
	return MapCustomPlanModifier{}
}
