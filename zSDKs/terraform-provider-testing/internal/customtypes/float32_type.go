package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.Float32Typable = Float32Type{}

type Float32Type struct {
	basetypes.Float32Type
}

func (t Float32Type) Equal(o attr.Type) bool {
	other, ok := o.(Float32Type)

	if !ok {
		return false
	}

	return t.Float32Type.Equal(other.Float32Type)
}

func (t Float32Type) String() string {
	return "Float32Type"
}

func (t Float32Type) ValueFromFloat32(ctx context.Context, in basetypes.Float32Value) (basetypes.Float32Valuable, diag.Diagnostics) {
	value := Float32Value{
		Float32Value: in,
	}

	return value, nil
}

func (t Float32Type) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.Float32Type.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	float32Value, ok := attrValue.(basetypes.Float32Value)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	float32Valuable, diags := t.ValueFromFloat32(ctx, float32Value)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting Float32Value to Float32Valuable: %v", diags)
	}

	return float32Valuable, nil
}

func (t Float32Type) ValueType(ctx context.Context) attr.Value {
	return Float32Value{}
}
