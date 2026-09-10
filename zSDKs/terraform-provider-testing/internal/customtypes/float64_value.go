package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.Float64Valuable = Float64Value{}

type Float64Value struct {
	basetypes.Float64Value
}

func (v Float64Value) Equal(o attr.Value) bool {
	other, ok := o.(Float64Value)

	if !ok {
		return false
	}

	return v.Float64Value.Equal(other.Float64Value)
}

func (v Float64Value) Type(ctx context.Context) attr.Type {
	return Float64Type{}
}
