package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.Int32Valuable = Int32Value{}

type Int32Value struct {
	basetypes.Int32Value
}

func (v Int32Value) Equal(o attr.Value) bool {
	other, ok := o.(Int32Value)

	if !ok {
		return false
	}

	return v.Int32Value.Equal(other.Int32Value)
}

func (v Int32Value) Type(ctx context.Context) attr.Type {
	return Int32Type{}
}
