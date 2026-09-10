package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom set default implementation.
func Set(value types.Set) defaults.Set {
	return SetDefault{
		value: value,
	}
}

type SetDefault struct {
	value types.Set
}

func (d SetDefault) Description(ctx context.Context) string {
	return "Custom set default"
}

func (d SetDefault) MarkdownDescription(ctx context.Context) string {
	return "Custom set default"
}

func (d SetDefault) DefaultSet(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
	resp.PlanValue = d.value
}
