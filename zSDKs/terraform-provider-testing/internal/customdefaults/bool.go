package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom bool default implementation.
func Bool() defaults.Bool {
	return BoolDefault{}
}

type BoolDefault struct{}

func (d BoolDefault) Description(ctx context.Context) string {
	return "Custom bool default"
}

func (d BoolDefault) MarkdownDescription(ctx context.Context) string {
	return "Custom bool default"
}

func (d BoolDefault) DefaultBool(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(true)
}
