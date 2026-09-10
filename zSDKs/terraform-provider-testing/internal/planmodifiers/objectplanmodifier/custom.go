package objectplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Object = ObjectCustomPlanModifier{}

type ObjectCustomPlanModifier struct{}

// Description describes the plan modification in plain text formatting.
func (v ObjectCustomPlanModifier) Description(_ context.Context) string {
	return "This string has been intentionally modified from the boilerplate."
}

// MarkdownDescription describes the plan modification in Markdown formatting.
func (v ObjectCustomPlanModifier) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// Validate performs the plan modification.
func (v ObjectCustomPlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	resp.Diagnostics.AddAttributeError(
		req.Path,
		"TODO: implement plan modifier Custom logic",
		req.Path.String()+": "+v.Description(ctx),
	)
}

func Custom() planmodifier.Object {
	return ObjectCustomPlanModifier{}
}
