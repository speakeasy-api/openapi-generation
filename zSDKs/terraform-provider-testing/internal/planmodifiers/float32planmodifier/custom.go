package float32planmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Float32 = Float32CustomPlanModifier{}

type Float32CustomPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v Float32CustomPlanModifier) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v Float32CustomPlanModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the plan modification.
func (v Float32CustomPlanModifier) PlanModifyFloat32(ctx context.Context, req planmodifier.Float32Request, resp *planmodifier.Float32Response) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement plan modifier Custom logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func Custom() planmodifier.Float32 {
	return Float32CustomPlanModifier{}
}
