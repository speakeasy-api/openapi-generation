package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.NumberValuable = NumberValue{}

type NumberValue struct {
	basetypes.NumberValue
}

func (v NumberValue) Equal(o attr.Value) bool {
	other, ok := o.(NumberValue)

	if !ok {
		return false
	}

	return v.NumberValue.Equal(other.NumberValue)
}

func (v NumberValue) Type(ctx context.Context) attr.Type {
	return NumberType{}
}
