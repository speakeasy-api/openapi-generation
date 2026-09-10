package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom list default implementation.
func List(value types.List) defaults.List {
	return ListDefault{
		value: value,
	}
}

type ListDefault struct {
	value types.List
}

func (d ListDefault) Description(ctx context.Context) string {
	return "Custom list default"
}

func (d ListDefault) MarkdownDescription(ctx context.Context) string {
	return "Custom list default"
}

func (d ListDefault) DefaultList(ctx context.Context, req defaults.ListRequest, resp *defaults.ListResponse) {
	resp.PlanValue = d.value
}
