package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom float64 default implementation.
func Float64() defaults.Float64 {
	return Float64Default{}
}

type Float64Default struct{}

func (d Float64Default) Description(ctx context.Context) string {
	return "Custom float64 default"
}

func (d Float64Default) MarkdownDescription(ctx context.Context) string {
	return "Custom float64 default"
}

func (d Float64Default) DefaultFloat64(ctx context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
	resp.PlanValue = types.Float64Value(1.23)
}
