package int64planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Int64 = Int64CustomPlanModifier{}

type Int64CustomPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v Int64CustomPlanModifier) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v Int64CustomPlanModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the plan modification.
func (v Int64CustomPlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement plan modifier Custom logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func Custom() planmodifier.Int64 {
	return Int64CustomPlanModifier{}
}
