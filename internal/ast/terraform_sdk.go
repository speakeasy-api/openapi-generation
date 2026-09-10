package ast

import (
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
)

// isTerraformDataModelCompatible checks whether the receiver (schema) TypeDef
// and an SDK TypeDef can participate in a Terraform data model conversion
// (SDK-to-model or model-to-SDK). A conversion is incompatible when types are
// ignored, structurally incompatible, or have no matching fields that would
// produce conversion code.
//
// The entityName is used for entity containment checks — a conversion is
// considered compatible when an SDK field's type tree leads to the entity,
// even without a direct field name match in the schema type.
//
// Used to filter ResponseSDKMethodTargets during assembly so that only targets
// with meaningful conversions are included.
func (t *TypeDef) isTerraformDataModelCompatible(entityName string, sdkType *TypeDef) bool {
	return t.isTerraformDataModelCompatibleWithVisited(entityName, sdkType, nil)
}

func (t *TypeDef) isTerraformDataModelCompatibleWithVisited(entityName string, sdkType *TypeDef, visited map[string]bool) bool {
	if t == nil || sdkType == nil {
		return false
	}

	// Check TerraformIgnore.DataModel on either type.
	if t.hasTerraformIgnoreDataModel() || sdkType.hasTerraformIgnoreDataModel() {
		return false
	}

	// Check type compatibility using TerraformMergeDataType which handles
	// same type, class↔union, enum unwrap, numeric coercion, class↔any.
	if _, err := t.TerraformMergeDataType(sdkType); err != nil {
		// Special case: when the schema is a class/union and the SDK type is an
		// array/collection containing the entity, the conversion is still
		// compatible. The generated code extracts entity items from the array
		// via a RefreshFrom_ subcall. The entity name may be on the array
		// type itself or on its item type.
		if (t.Type == DataTypeClass || t.Type == DataTypeUnion) &&
			sdkType.ItemType != nil &&
			sdkType.FindEntityTypeDef(entityName) != nil {
			return true
		}

		return false
	}

	// For primitive types, check entity extension mismatch.
	if t.IsTerraformPrimitiveType() {
		if t.Extensions != nil && t.Extensions.Has("x-speakeasy-terraform-in-entity") &&
			(sdkType.Extensions == nil || !sdkType.Extensions.Has("x-speakeasy-terraform-in-entity")) {
			return false
		}

		return true
	}

	// Cycle detection for recursive type references.
	if visited == nil {
		visited = make(map[string]bool)
	}

	key := t.terraformCycleDetectionKey()
	if visited[key] {
		// Already checking this type, assume compatible to avoid infinite recursion.
		return true
	}

	visited[key] = true

	// For class/union types, check if any SDK field matches a schema field
	// (by sanitized name) where the matched pair is also compatible.
	if t.Type == DataTypeClass || t.Type == DataTypeUnion {
		// Wrapped attribute handling: if the schema has any x-speakeasy-wrapped-*
		// extension, field conversion is redirected through a wrapped field.
		// This always produces conversion code, so treat as compatible.
		if t.hasWrappedExtension() {
			return true
		}

		for _, sdkField := range sdkType.Fields {
			if sdkField.Type == nil {
				continue
			}

			// Direct field name match: if the SDK field has an equivalent
			// schema field and the pair is recursively compatible.
			equivalent, _ := sdkField.FindTerraformEquivalentField(t.Fields, false)
			if equivalent != nil && equivalent.Type != nil {
				if equivalent.Type.isTerraformDataModelCompatibleWithVisited(entityName, sdkField.Type, visited) {
					return true
				}
			}

			// Entity containment fallback: SDK field's type leads to the
			// entity even without a direct name match. Unmatched fields
			// that contain the entity still generate RefreshFrom_ subcalls
			// in the conversion code.
			if sdkField.Type.FindEntityTypeDef(entityName) != nil {
				return true
			}
		}

		// Check union associated types.
		for _, sdkAssoc := range sdkType.AssociatedTypes {
			for _, schemaAssoc := range t.AssociatedTypes {
				if schemaAssoc.isTerraformDataModelCompatibleWithVisited(entityName, sdkAssoc, visited) {
					return true
				}
			}
		}

		return false
	}

	// For container types (array/set/map), check ItemType.
	if t.ItemType != nil && sdkType.ItemType != nil {
		return t.ItemType.isTerraformDataModelCompatibleWithVisited(entityName, sdkType.ItemType, visited)
	}

	return true
}

// hasTerraformIgnoreDataModel returns true if the TypeDef has the
// TerraformIgnore.DataModel extension set.
func (t *TypeDef) hasTerraformIgnoreDataModel() bool {
	return t.Extensions != nil &&
		t.Extensions.TerraformIgnore != nil &&
		t.Extensions.TerraformIgnore.DataModel
}

// hasWrappedExtension returns true if the TypeDef has any
// x-speakeasy-wrapped-* extension set. This indicates the schema uses
// wrapped attribute handling where field conversion is redirected through
// a wrapped field.
func (t *TypeDef) hasWrappedExtension() bool {
	if t.Extensions == nil || t.Extensions.All == nil {
		return false
	}

	for key := range t.Extensions.All {
		if strings.HasPrefix(key, "x-speakeasy-wrapped-") {
			return true
		}
	}

	return false
}

// terraformCycleDetectionKey returns a string key for cycle detection, using the Symbol extension,
// Name, or Type string in that order of preference.
func (t *TypeDef) terraformCycleDetectionKey() string {
	if t.Extensions != nil {
		if symbol, ok := t.Extensions.Get("Symbol"); ok {
			if s, ok := symbol.(string); ok && s != "" {
				return s
			}
		}
	}

	if t.Name != "" {
		return t.Name
	}

	return string(t.Type)
}

// TerraformHasNestedWriteOnlyFields checks whether the receiver (schema)
// TypeDef has any write-only fields at any nesting depth relative to the given
// SDK TypeDef. A field is "write-only" when it exists in the schema type but
// has no name-equivalent in the SDK type.
//
// Used to decide whether to capture prior state data before reallocating a
// parent object so that write-only values from the previous state are not lost.
func (t *TypeDef) TerraformHasNestedWriteOnlyFields(sdkType *TypeDef) bool {
	return t.terraformHasNestedWriteOnlyFieldsWithVisited(sdkType, nil)
}

// terraformHasNestedWriteOnlyFieldsWithVisited is the recursive implementation
// of TerraformHasNestedWriteOnlyFields. It walks class fields, container item
// types, and union associated types, using visited to prevent infinite
// recursion on circular type references.
func (t *TypeDef) terraformHasNestedWriteOnlyFieldsWithVisited(sdkType *TypeDef, visited map[string]bool) bool {
	if t == nil || sdkType == nil {
		return false
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	key := t.terraformCycleDetectionKey()
	if visited[key] {
		return false
	}

	visited[key] = true

	// Class types: check each schema field against the SDK type's fields.
	if t.Type == DataTypeClass && len(t.Fields) > 0 {
		for _, toField := range t.Fields {
			if toField.Type == nil {
				continue
			}

			equivalent, _ := toField.FindTerraformEquivalentField(sdkType.Fields, false)

			if toField.Type.Type == DataTypeClass || toField.Type.Type == DataTypeUnion ||
				toField.Type.IsContainer() {
				if equivalent == nil {
					// Entire nested object is write-only.
					return true
				}

				if toField.Type.terraformHasNestedWriteOnlyFieldsWithVisited(equivalent.Type, visited) {
					return true
				}
			} else if equivalent == nil {
				// Leaf write-only field.
				return true
			}
		}
	}

	// Container types (array/set/map): check ItemType.
	if t.ItemType != nil && sdkType.ItemType != nil {
		if t.ItemType.terraformHasNestedWriteOnlyFieldsWithVisited(sdkType.ItemType, visited) {
			return true
		}
	}

	// Union types: match associated types by sanitized union variant name.
	if len(t.AssociatedTypes) > 0 && len(sdkType.AssociatedTypes) > 0 {
		// Pre-build SDK variant lookup to avoid recomputing names per schema variant.
		sdkVariants := make(map[string]*TypeDef, len(sdkType.AssociatedTypes))
		for _, fromAssoc := range sdkType.AssociatedTypes {
			sdkVariants[terraformSanitizeUnionVariantName(sdkType, fromAssoc)] = fromAssoc
		}

		for _, toAssoc := range t.AssociatedTypes {
			fromAssoc, ok := sdkVariants[terraformSanitizeUnionVariantName(t, toAssoc)]
			if !ok {
				continue
			}

			if toAssoc.terraformHasNestedWriteOnlyFieldsWithVisited(fromAssoc, visited) {
				return true
			}
		}
	}

	return false
}

// terraformSanitizeUnionVariantName returns a comparable name for a union
// variant (associated type) relative to its parent union type. The name is
// derived from (in priority order): the discriminator mapping name, the
// OriginalName, the Name, or a type-based fallback. Container types
// (array/set/map) recurse into their ItemType to build a compound name.
// The result is PascalCase-sanitized so that variants from different type
// trees (schema vs SDK) can be matched regardless of casing differences.
func terraformSanitizeUnionVariantName(unionType, associatedType *TypeDef) string {
	if associatedType == nil {
		return ""
	}

	sanitize := func(s string) string {
		return strcase.ToPascal(sanitization.SanitizeName(s))
	}

	// Prefer discriminator mapping name.
	if unionType != nil && unionType.Discriminator != nil {
		for _, mapping := range unionType.Discriminator.Mapping {
			if mapping.Type == nil {
				continue
			}

			if associatedType.Name == mapping.Type.Name || (associatedType.OriginalName != "" && associatedType.OriginalName == mapping.Type.OriginalName) {
				return sanitize(mapping.Name)
			}
		}
	}

	if associatedType.OriginalName != "" {
		return sanitize(associatedType.OriginalName)
	}

	if associatedType.Name != "" {
		return sanitize(associatedType.Name)
	}

	switch associatedType.Type {
	case DataTypeString:
		return "Str"
	case DataTypeMap, DataTypeArray, DataTypeSet:
		return sanitize(
			string(associatedType.Type) + "_Of_" + terraformSanitizeUnionVariantName(associatedType, associatedType.ItemType),
		)
	default:
		return sanitize(string(associatedType.Type))
	}
}
