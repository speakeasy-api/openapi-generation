package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.SetValuable = SetValue{}

type SetValue struct {
	basetypes.SetValue
}

func (v SetValue) Equal(o attr.Value) bool {
	other, ok := o.(SetValue)

	if !ok {
		return false
	}

	return v.SetValue.Equal(other.SetValue)
}

func (v SetValue) Type(ctx context.Context) attr.Type {
	return SetType{
		SetType: v.SetValue.Type(ctx).(basetypes.SetType),
	}
}
