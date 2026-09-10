package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom int64 default implementation.
func Int64() defaults.Int64 {
	return Int64Default{}
}

type Int64Default struct{}

func (d Int64Default) Description(ctx context.Context) string {
	return "Custom int64 default"
}

func (d Int64Default) MarkdownDescription(ctx context.Context) string {
	return "Custom int64 default"
}

func (d Int64Default) DefaultInt64(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(123)
}
