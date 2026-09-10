package ast

const (
	// sdkRequestMethodPrefix is the method name prefix for request conversion
	// methods (e.g. To_SharedThing).
	sdkRequestMethodPrefix = "To_"

	// sdkResponseMethodPrefix is the method name prefix for response conversion
	// methods (e.g. RefreshFrom_SharedThing).
	sdkResponseMethodPrefix = "RefreshFrom_"

	// sdkResponseArrayMethodPrefix is the method name prefix for response
	// conversion methods that handle array/set wrapper types
	// (e.g. RefreshFromArrayOf_SharedThing).
	sdkResponseArrayMethodPrefix = "RefreshFromArrayOf_"
)

// TerraformSDKMethodTarget represents a TypeDef along the path from an API
// request or response type to a Terraform entity TypeDef that needs an SDK
// conversion method. This includes the entity TypeDef itself and any
// intermediate TypeDefs.
type TerraformSDKMethodTarget struct {
	// TypeDef is the target type that needs a conversion method.
	TypeDef *TypeDef

	// Operation is the TerraformOperation that owns this target. Provides
	// operation context such as patch style, pagination, and entity
	// operation string.
	Operation *TerraformOperation

	// Optional indicates whether this TypeDef was reached through an
	// optional/nullable field. Determines whether the SDK parameter for
	// response methods should be a pointer type.
	Optional bool

	// IsArrayWrapper indicates this target represents an array/set type
	// that directly wraps entity items. When true, the method uses a
	// RefreshFromArrayOf_ prefix instead of RefreshFrom_.
	IsArrayWrapper bool
}

// SDKRequestMethod creates a TerraformSDKMethod for a request conversion
// using the To_ prefix.
func (t TerraformSDKMethodTarget) SDKRequestMethod() *TerraformSDKMethod {
	return &TerraformSDKMethod{
		MethodName: terraformSDKMethodName(sdkRequestMethodPrefix, t.TypeDef),
		Operation:  t.Operation,
		Optional:   t.Optional,
		Target:     t,
	}
}

// SDKResponseMethod creates a TerraformSDKMethod for a response conversion.
// The prefix is RefreshFromArrayOf_ for array wrapper targets, otherwise
// RefreshFrom_.
func (t TerraformSDKMethodTarget) SDKResponseMethod() *TerraformSDKMethod {
	prefix := sdkResponseMethodPrefix
	if t.IsArrayWrapper {
		prefix = sdkResponseArrayMethodPrefix
	}

	return &TerraformSDKMethod{
		MethodName: terraformSDKMethodName(prefix, t.TypeDef),
		Operation:  t.Operation,
		Optional:   t.Optional,
		Target:     t,
	}
}
