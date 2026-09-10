package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.Int64Valuable = Int64Value{}

type Int64Value struct {
	basetypes.Int64Value
}

func (v Int64Value) Equal(o attr.Value) bool {
	other, ok := o.(Int64Value)

	if !ok {
		return false
	}

	return v.Int64Value.Equal(other.Int64Value)
}

func (v Int64Value) Type(ctx context.Context) attr.Type {
	return Int64Type{}
}
