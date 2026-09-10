package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = StringValue{}

type StringValue struct {
	basetypes.StringValue
}

func (v StringValue) Equal(o attr.Value) bool {
	other, ok := o.(StringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v StringValue) Type(ctx context.Context) attr.Type {
	return StringType{}
}
