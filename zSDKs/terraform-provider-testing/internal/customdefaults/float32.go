package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom float32 default implementation.
func Float32() defaults.Float32 {
	return Float32Default{}
}

type Float32Default struct{}

func (d Float32Default) Description(ctx context.Context) string {
	return "Custom float32 default"
}

func (d Float32Default) MarkdownDescription(ctx context.Context) string {
	return "Custom float32 default"
}

func (d Float32Default) DefaultFloat32(ctx context.Context, req defaults.Float32Request, resp *defaults.Float32Response) {
	resp.PlanValue = types.Float32Value(1.23)
}
