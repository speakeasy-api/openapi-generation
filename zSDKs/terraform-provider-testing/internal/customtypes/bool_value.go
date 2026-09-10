package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.BoolValuable = BoolValue{}

type BoolValue struct {
	basetypes.BoolValue
}

func (v BoolValue) Equal(o attr.Value) bool {
	other, ok := o.(BoolValue)

	if !ok {
		return false
	}

	return v.BoolValue.Equal(other.BoolValue)
}

func (v BoolValue) Type(ctx context.Context) attr.Type {
	return BoolType{}
}
