package ast

// TerraformSDKMethod describes a deduplicated SDK conversion method for a
// Terraform entity's data model type. Contains the resolved method name and
// operation context needed for method body rendering.
type TerraformSDKMethod struct {
	// MethodName is the fully sanitized method name, e.g.
	// "ToSharedCreateThingRequest" or "RefreshFromSharedGetThingResponse".
	MethodName string

	// Operation is the first TerraformOperation that contributed this method
	// during deduplication (first-wins). Provides operation context such as
	// patch style, pagination, and entity operation string.
	Operation *TerraformOperation

	// Optional indicates whether the SDK parameter for response methods
	// should be a pointer type.
	Optional bool

	// Target is the underlying SDK method target, containing the TypeDef
	// and metadata needed for method body rendering.
	Target TerraformSDKMethodTarget
}

// terraformSDKMethodName computes the fully sanitized Go method name for an SDK
// conversion method by combining a prefix (e.g. "To_", "RefreshFrom_") with the
// scope-qualified type name.
//
// Most targets are class or union TypeDefs, but array/set TypeDefs may appear
// in paths from FindEntitySDKMethodTargets (when traversing ItemType) or as
// IsArrayWrapper response targets. For arrays/sets, the ItemType is used for
// naming since the array TypeDef itself has an empty Name.
//
// Scope handling: the default scope is operations. If td.Scope matches,
// it swaps to shared to prevent self-prefixing. When scopes differ, the
// TypeDef's scope is used as a package prefix
// (e.g. "shared.Thing" or "operations.CreateThingRequest").
//
// Reserved keyword collision is not possible since the prefix always makes
// the result longer than any Go SDK reserved keyword.
func terraformSDKMethodName(prefix string, td *TypeDef) string {
	// Resolve array/set types to their ItemType for naming. Array TypeDefs
	// have empty names; the meaningful name comes from the item type.
	if (td.Type == DataTypeArray || td.Type == DataTypeSet) && td.ItemType != nil {
		td = td.ItemType
	}

	scope := ScopeOperations
	if td.Scope == scope {
		scope = ScopeShared
	}

	typeName := TerraformGoTypeName(td.Name)

	if td.Scope != scope {
		typeName = string(td.Scope) + "." + typeName
	}

	return TerraformGoTypeName(prefix + typeName)
}
