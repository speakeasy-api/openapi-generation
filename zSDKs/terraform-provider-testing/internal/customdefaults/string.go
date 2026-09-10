package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom string default implementation.
func String() defaults.String {
	return StringDefault{}
}

type StringDefault struct{}

func (d StringDefault) Description(ctx context.Context) string {
	return "Custom string default"
}

func (d StringDefault) MarkdownDescription(ctx context.Context) string {
	return "Custom string default"
}

func (d StringDefault) DefaultString(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue("custom default")
}
