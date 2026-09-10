package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom int32 default implementation.
func Int32() defaults.Int32 {
	return Int32Default{}
}

type Int32Default struct{}

func (d Int32Default) Description(ctx context.Context) string {
	return "Custom int32 default"
}

func (d Int32Default) MarkdownDescription(ctx context.Context) string {
	return "Custom int32 default"
}

func (d Int32Default) DefaultInt32(ctx context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
	resp.PlanValue = types.Int32Value(123)
}
