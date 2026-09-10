package setplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Set = SetCustomPlanModifier{}

type SetCustomPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v SetCustomPlanModifier) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v SetCustomPlanModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the plan modification.
func (v SetCustomPlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement plan modifier Custom logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func Custom() planmodifier.Set {
	return SetCustomPlanModifier{}
}
