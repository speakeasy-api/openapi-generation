package customtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.Float32Valuable = Float32Value{}

type Float32Value struct {
	basetypes.Float32Value
}

func (v Float32Value) Equal(o attr.Value) bool {
	other, ok := o.(Float32Value)

	if !ok {
		return false
	}

	return v.Float32Value.Equal(other.Float32Value)
}

func (v Float32Value) Type(ctx context.Context) attr.Type {
	return Float32Type{}
}
