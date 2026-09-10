package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom map default implementation.
func Map(value types.Map) defaults.Map {
	return MapDefault{
		value: value,
	}
}

type MapDefault struct {
	value types.Map
}

func (d MapDefault) Description(ctx context.Context) string {
	return "Custom map default"
}

func (d MapDefault) MarkdownDescription(ctx context.Context) string {
	return "Custom map default"
}

func (d MapDefault) DefaultMap(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
	resp.PlanValue = d.value
}
