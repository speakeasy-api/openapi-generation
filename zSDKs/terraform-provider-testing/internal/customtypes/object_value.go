package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.ObjectValuable = ObjectValue{}

type ObjectValue struct {
	basetypes.ObjectValue
}

func (v ObjectValue) Equal(o attr.Value) bool {
	other, ok := o.(ObjectValue)

	if !ok {
		return false
	}

	return v.ObjectValue.Equal(other.ObjectValue)
}

func (v ObjectValue) Type(ctx context.Context) attr.Type {
	return ObjectType{
		ObjectType: v.ObjectValue.Type(ctx).(basetypes.ObjectType),
	}
}
