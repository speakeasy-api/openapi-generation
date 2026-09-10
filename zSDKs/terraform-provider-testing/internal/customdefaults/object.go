package customdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Example custom object default implementation.
func Object(value types.Object) defaults.Object {
	return ObjectDefault{
		value: value,
	}
}

type ObjectDefault struct {
	value types.Object
}

func (d ObjectDefault) Description(ctx context.Context) string {
	return "Custom object default"
}

func (d ObjectDefault) MarkdownDescription(ctx context.Context) string {
	return "Custom object default"
}

func (d ObjectDefault) DefaultObject(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
	resp.PlanValue = d.value
}
