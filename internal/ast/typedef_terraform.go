package ast

import (
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Returns a mutated TypeDef where, if necessary, the TypeDef is wrapped into a
// new single Field TypeDef for Terraform usage when:
//   - The x-speakeasy-wrapped-attribute extension is configured. The Field is
//     named with the extension value.
//   - The TypeDef DataType is not class or union. The Field is named "data".
//
// Clone the TypeDef before calling this method to preserve the original.
func (t *TypeDef) TerraformHandleWrappedAttribute(entityName string, entityOperation string) *TypeDef {
	if t == nil {
		return nil
	}

	// Check for class/union types that don't have the wrapped attribute extension.
	// For class/union types without the extension, return the original TypeDef.
	// For non-class/non-union types (primitives, arrays, etc.), always wrap them.
	if (t.Type == DataTypeClass || t.Type == DataTypeUnion) && (t.Extensions == nil || t.Extensions.WrappedAttribute == nil) {
		return t
	}

	fieldName := "data"

	if t.Extensions != nil && t.Extensions.WrappedAttribute != nil {
		fieldName = *t.Extensions.WrappedAttribute
	}

	wrapperFieldDef := &FieldDef{
		Comments:     t.Comments,
		Name:         fieldName,
		OriginalName: fieldName,
		Type:         t,
	}
	wrapperTypeDef := &TypeDef{
		Extensions: &TypeDefExtensions{
			All: map[string]any{
				("x-speakeasy-wrapped-" + entityOperation): fieldName,
			},
			Entity: &extensions.Entity{Names: []string{entityName}},
		},
		Fields:        Fields{wrapperFieldDef},
		Name:          entityName,
		OriginalName:  entityName,
		ResolvedModel: entityName,
		Type:          DataTypeClass,
	}

	return wrapperTypeDef
}

// Returns a mutated TypeDef where data is hoisted to the top level
// Fields when the x-speakeasy-entity configuration matches the given entity
// name for Terraform usage. Clone the TypeDef before calling this method to
// preserve the original.
//
// When the entity is found within an array's ItemType, the TypeDef is
// transformed to the hoisted item type (e.g., array becomes class).
func (t *TypeDef) TerraformHoistByEntityName(entityName string) error {
	hoisted, err := terraformHoistByEntityName(t, entityName)

	if err != nil {
		return err
	}

	if hoisted != t {
		// When there was an ItemType and we got a different result back, it means
		// we followed the ItemType path and the hoisted result is the item type.
		// In this case, we need to replace the entire TypeDef with the hoisted one,
		// including its Type. This transforms arrays to their hoisted item type.
		hadItemType := t.ItemType != nil

		t.AssociatedTypes = hoisted.AssociatedTypes
		t.Discriminator = hoisted.Discriminator
		t.Extensions.Merge(hoisted.Extensions)
		t.Fields = hoisted.Fields

		// Only copy Type and ItemType when we followed the ItemType path.
		// When hoisting through fields, preserve the original Type.
		if hadItemType {
			t.Type = hoisted.Type
			t.ItemType = hoisted.ItemType
		}
	}

	return nil
}

// terraformHoistByEntityName is the internal implementation that returns a
// potentially modified TypeDef. It returns the hoisted TypeDef if changes were
// made, or the original if no hoisting was performed.
func terraformHoistByEntityName(t *TypeDef, entityName string) (*TypeDef, error) {
	if t == nil {
		return t, nil
	}

	if t.HasEntityName(entityName) {
		return t, nil
	}

	if t.ItemType != nil {
		return terraformHoistByEntityName(t.ItemType, entityName)
	}

	hoistedAssociatedTypes := slices.Clone(t.AssociatedTypes)
	hoistedFields := make(Fields, 0)
	hoistedDiscriminatorMappings := make(map[string]*TypeDef)

	for _, field := range t.Fields {
		if field == nil || field.Type == nil {
			continue
		}

		// If the field is a deepObject style query parameter, do not hoist
		// fields. Instead, use the field as-is including the top level name.
		paramAnnotation := field.getParamAnnotation()

		if paramAnnotation != nil && paramAnnotation.Style == "deepObject" {
			hoistedFields = append(hoistedFields, field)
			continue
		}

		hoisted, err := terraformHoistByEntityName(field.Type, entityName)
		if err != nil {
			return nil, err
		}

		// Check hoisted result for nested structure (no fields or associated types)
		hasNoFieldsOrAssociatedTypes := len(hoisted.Fields) == 0 && len(hoisted.AssociatedTypes) == 0

		if hasNoFieldsOrAssociatedTypes {
			// Push the ORIGINAL field, not the hoisted result
			existing := hoistedFields.GetField(field.Name)

			if existing == nil {
				hoistedFields = append(hoistedFields, field.Clone())
				continue
			}

			// Merge using original field.Type, not hoisted
			if err := existing.Type.TerraformMerge(field.Type); err != nil {
				return nil, err
			}

			continue
		}

		// When there IS nested structure, push fields from hoisted result
		for _, nestedType := range hoisted.AssociatedTypes {
			existing := t.FindAssociatedTypeByTypeDef(nestedType)

			if existing == nil {
				hoistedAssociatedTypes = append(hoistedAssociatedTypes, nestedType)

				if hoisted.Discriminator != nil {
					for _, mapping := range hoisted.Discriminator.Mapping {
						if mapping.Type.Name == nestedType.Name || (mapping.Type.OriginalName != "" && mapping.Type.OriginalName == nestedType.OriginalName) {
							hoistedDiscriminatorMappings[mapping.Name] = nestedType
							break
						}
					}
				}
				continue
			}

			if err := existing.TerraformMerge(nestedType); err != nil {
				return nil, err
			}
		}

		for _, nestedField := range hoisted.Fields {
			existing := hoistedFields.GetField(nestedField.Name)

			if existing == nil {
				hoistedFields = append(hoistedFields, nestedField)
				continue
			}

			if err := existing.Type.TerraformMerge(nestedField.Type); err != nil {
				return nil, err
			}
		}

		// Copy entity extension from hoisted result to parent
		if hoisted != nil && hoisted.Extensions != nil && hoisted.Extensions.Entity != nil {
			t.EnsureExtensions()
			t.Extensions.Entity = hoisted.Extensions.Entity
		}
	}

	result := t.Clone()
	result.AssociatedTypes = hoistedAssociatedTypes
	result.Fields = hoistedFields

	enhancedDiscriminator := result.Discriminator
	if len(hoistedDiscriminatorMappings) > 0 {
		if enhancedDiscriminator == nil {
			enhancedDiscriminator = &Discriminator{
				TypePropertyName: "irrelevant",
				Mapping:          []*DiscriminatorMapping{},
				Inferred:         true,
			}
		}

		for name, typeDef := range hoistedDiscriminatorMappings {
			enhancedDiscriminator.Mapping = append(enhancedDiscriminator.Mapping, &DiscriminatorMapping{
				Name: name,
				Type: typeDef,
			})
		}

		slices.SortFunc(enhancedDiscriminator.Mapping, func(a, b *DiscriminatorMapping) int {
			return strings.Compare(a.Name, b.Name)
		})
	}

	result.Discriminator = enhancedDiscriminator

	return result, nil
}

// Mutates the TypeDef where common AssociatedTypes Fields are hoisted to the
// TypeDef Fields for Terraform usage.
//
// This function also tracks hoisting sources via TerraformHoistedFrom on each
// hoisted field's extensions. This enables plan modifiers to prevent false drift
// detection by copying values from the active oneOf variant.
//
// The pathPrefix parameter is used for nested oneOf unions to build absolute
// paths in plan modifiers. For root-level hoisting, pass nil.
func (t *TypeDef) TerraformHoistAssociatedTypesCommonFields(pathPrefix []string) error {
	if t == nil || len(t.AssociatedTypes) == 0 {
		return nil
	}

	commonFields := t.AssociatedTypes.TerraformCommonFields()

	slices.SortFunc(commonFields, func(a, b *FieldDef) int {
		return strings.Compare(
			strings.ToLower(a.Name),
			strings.ToLower(b.Name),
		)
	})

	// Build map of hoisted sources for each field name (to set after merge).
	hoistedSourcesByField := make(map[string][]TerraformHoistedSource)

	// Set TerraformAliasTo on associated type fields and collect hoisted sources.
	for _, field := range commonFields {
		hoistedSources := make([]TerraformHoistedSource, 0)

		for _, associatedType := range t.AssociatedTypes {
			matchedField := associatedType.Fields.GetField(field.Name)

			if matchedField == nil || matchedField.Type == nil {
				continue
			}

			matchedField.Type.EnsureExtensions()

			aliasTo := "../" + matchedField.Name
			matchedField.Type.Extensions.TerraformAliasTo = &aliasTo

			// Record source for plan modifier generation.
			// Use terraform.AttributeName to match Go struct field names format.
			associatedTypeName := terraform.AttributeName(t.AssociatedTypeName(associatedType))
			hoistedSources = append(hoistedSources, TerraformHoistedSource{
				AssociatedTypeName: associatedTypeName,
				FieldName:          terraform.AttributeName(matchedField.Name),
				PathPrefix:         pathPrefix,
			})
		}

		if len(hoistedSources) > 0 {
			hoistedSourcesByField[field.Name] = hoistedSources
		}
	}

	// Merge common fields into parent TypeDef (this clones the fields).
	if len(commonFields) > 0 {
		commonTypeDef := &TypeDef{
			Fields: commonFields,
			Type:   t.Type,
		}
		if err := t.TerraformMerge(commonTypeDef); err != nil {
			return err
		}
	}

	// Now set TerraformHoistedFrom on the MERGED fields (not the originals).
	// This ensures only the parent's hoisted fields have this extension.
	// Also merge enum values to superset when fields are enums with different values.
	for fieldName, hoistedSources := range hoistedSourcesByField {
		mergedField := t.Fields.GetField(fieldName)
		if mergedField == nil || mergedField.Type == nil {
			continue
		}

		mergedField.Type.EnsureExtensions()
		mergedField.Type.Extensions.TerraformHoistedFrom = hoistedSources

		// For enum fields, merge values from all associated types into a superset.
		if mergedField.Type.Type == DataTypeEnum && mergedField.Type.Enum != nil {
			seenValues := make(map[string]bool)
			seenNames := make(map[string]bool)
			var supersetValues []string
			var supersetNames []string

			for _, associatedType := range t.AssociatedTypes {
				matchedField := associatedType.Fields.GetField(fieldName)
				if matchedField == nil || matchedField.Type == nil || matchedField.Type.Enum == nil {
					continue
				}

				for _, v := range matchedField.Type.Enum.Values {
					if !seenValues[v] {
						seenValues[v] = true
						supersetValues = append(supersetValues, v)
					}
				}
				for _, n := range matchedField.Type.Enum.Names {
					if !seenNames[n] {
						seenNames[n] = true
						supersetNames = append(supersetNames, n)
					}
				}
			}

			mergedField.Type.Enum.Values = supersetValues
			mergedField.Type.Enum.Names = supersetNames
		}
	}

	return nil
}

// SetExtensionIfAbsentRecursive recursively sets an extension on all types in
// the TypeDef tree that don't already have the specified extension. Stops
// recursing into children when the extension is already present.
func (t *TypeDef) SetExtensionIfAbsentRecursive(tagName string, tagValue any) {
	if t == nil {
		return
	}

	t.EnsureExtensions()

	if t.Extensions.Has(tagName) {
		return
	}

	t.Extensions.Set(tagName, tagValue)

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.SetExtensionIfAbsentRecursive(tagName, tagValue)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.SetExtensionIfAbsentRecursive(tagName, tagValue)
		}
	}

	if t.ItemType != nil {
		t.ItemType.SetExtensionIfAbsentRecursive(tagName, tagValue)
	}
}

// SetExtensionRecursive recursively sets an extension on all types in the
// TypeDef tree (bottom-up). Only sets the extension if it doesn't already
// exist. Special handling for x-speakeasy-ignore which sets Extensions.Ignore.
func (t *TypeDef) SetExtensionRecursive(tagName string, tagValue any) {
	if t == nil {
		return
	}

	// Recurse into children first, then handle tag assignment
	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.SetExtensionRecursive(tagName, tagValue)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.SetExtensionRecursive(tagName, tagValue)
		}
	}

	if t.ItemType != nil {
		t.ItemType.SetExtensionRecursive(tagName, tagValue)
	}

	// Then handle tag assignment
	t.EnsureExtensions()

	// Special handling for x-speakeasy-ignore which expects a *bool
	if tagName == extensions.ExtIgnore.Name() {
		if boolVal, ok := tagValue.(bool); ok {
			t.Extensions.Ignore = &boolVal
		}
		return
	}

	t.Extensions.SetIfAbsent(tagName, tagValue)
}

// PromoteResponseFilterFields recursively walks the TypeDef tree and promotes
// any field with ResponseFilter set to true to Optional+Computed in the data
// source schema. This sets x-speakeasy-param-optional and
// x-speakeasy-param-computed, and removes x-speakeasy-param-readonly (which
// would otherwise override Optional to false).
func (t *TypeDef) PromoteResponseFilterFields() {
	if t == nil {
		return
	}

	for _, field := range t.Fields {
		if field == nil || field.Type == nil {
			continue
		}

		if field.Type.Extensions != nil && field.Type.Extensions.ResponseFilter != nil && *field.Type.Extensions.ResponseFilter {
			field.Type.EnsureExtensions()
			field.Type.Extensions.Set("x-speakeasy-param-optional", true)
			field.Type.Extensions.Set("x-speakeasy-param-computed", true)
			// Remove readonly so it doesn't override Optional to false.
			field.Type.Extensions.Remove("x-speakeasy-param-readonly")
		}

		field.Type.PromoteResponseFilterFields()
	}

	for _, associatedType := range t.AssociatedTypes {
		associatedType.PromoteResponseFilterFields()
	}

	if t.ItemType != nil {
		t.ItemType.PromoteResponseFilterFields()
	}
}

// HoistArrayItemResponseFilterFields checks the entity's array fields for item
// types that contain x-speakeasy-response-filter fields and hoists those filter
// fields to the entity level. This supports the "entity-above-array" pattern
// where the entity wraps an array of items that should be filtered by a
// user-provided value. Hoisted fields are Optional-only (not Computed) since
// they exist purely as filter inputs at the entity level.
func (t *TypeDef) HoistArrayItemResponseFilterFields() {
	if t == nil {
		return
	}

	var hoistedFields []*FieldDef

	for _, field := range t.Fields {
		if field == nil || field.Type == nil {
			continue
		}

		if field.Type.Type != DataTypeArray || field.Type.ItemType == nil {
			continue
		}

		for _, itemField := range field.Type.ItemType.Fields {
			if itemField == nil || itemField.Type == nil || itemField.Type.Extensions == nil {
				continue
			}

			if itemField.Type.Extensions.ResponseFilter == nil || !*itemField.Type.Extensions.ResponseFilter {
				continue
			}

			hoisted := itemField.Clone()
			hoisted.Type.EnsureExtensions()
			hoisted.Type.Extensions.Set("x-speakeasy-param-optional", true)
			// Not Computed — this is purely a filter input, not populated from the response.
			hoisted.Type.Extensions.Remove("x-speakeasy-param-computed")
			hoisted.Type.Extensions.Remove("x-speakeasy-param-readonly")
			// Keep ResponseFilter on the hoisted field so the TS codegen can
			// detect it and generate the array filter loop.

			hoistedFields = append(hoistedFields, hoisted)

			// Fully revert the item-level field's promotion by PromoteResponseFilterFields.
			// That function set Optional+Computed and removed ReadOnly. For entity-above-
			// array the filter belongs at the entity level, not inside list items, so
			// restore the item field to its original Computed-only state.
			itemField.Type.Extensions.Remove("x-speakeasy-param-optional")
			itemField.Type.Extensions.Remove("x-speakeasy-param-computed")
			itemField.Type.Extensions.Set("x-speakeasy-param-readonly", true)
		}
	}

	for _, hoisted := range hoistedFields {
		// AddField returns an error if the field already exists (e.g., user
		// explicitly declared it on the wrapper schema). Skip silently.
		if ff, err := t.Fields.AddField(hoisted, false); err == nil {
			t.Fields = ff
		}
	}
}

// TerraformApplyEquivalentFieldProperties recursively copies field-level
// properties (Optional, Nullable, Type.Name, Comments) from matching fields
// in the source TypeDef onto this TypeDef's fields.
//
// Fields are matched using sanitized name comparison via
// FindTerraformEquivalentField. AssociatedTypes are matched using
// AssociatedTypeName comparison. The source's values take precedence when
// merging Comments.
func (t *TypeDef) TerraformApplyEquivalentFieldProperties(source *TypeDef) {
	if t == nil || source == nil || t.Type != source.Type {
		return
	}

	for _, field := range t.Fields {
		equivalent, _ := field.FindTerraformEquivalentField(source.Fields, false)

		if equivalent == nil {
			continue
		}

		field.Optional = equivalent.Optional
		field.Nullable = equivalent.Nullable

		if field.Type != nil && equivalent.Type != nil {
			field.Type.Name = equivalent.Type.Name
		}

		// Merge comments, preferring source (equivalent) values.
		// Clone the equivalent's Comments then merge the field's Comments
		// as fallback, so the source takes precedence.
		if field.Comments == nil && equivalent.Comments != nil {
			field.Comments = equivalent.Comments.Clone()
		} else if field.Comments != nil && equivalent.Comments != nil {
			mergedComments := equivalent.Comments.Clone()
			mergedComments.Merge(field.Comments)
			field.Comments = mergedComments
		}

		if field.Type != nil && equivalent.Type != nil {
			field.Type.TerraformApplyEquivalentFieldProperties(equivalent.Type)
		}
	}

	for _, associated := range t.AssociatedTypes {
		targetName := t.AssociatedTypeName(associated)

		for _, sourceAssociated := range source.AssociatedTypes {
			if source.AssociatedTypeName(sourceAssociated) == targetName {
				associated.TerraformApplyEquivalentFieldProperties(sourceAssociated)

				break
			}
		}
	}

	if t.ItemType != nil && source.ItemType != nil {
		t.ItemType.TerraformApplyEquivalentFieldProperties(source.ItemType)
	}
}

// SetExtensionOnEquivalentTypes recursively sets an extension on this TypeDef
// and on nested types that have equivalents in the other TypeDef.
//
// At each matching level, Extensions.All[extensionName] is set to
// extensionValue if not already present. Fields are matched using
// FindTerraformEquivalentField, AssociatedTypes using AssociatedTypeName
// comparison, and ItemType directly.
func (t *TypeDef) SetExtensionOnEquivalentTypes(other *TypeDef, extensionName string, extensionValue any) {
	t.TerraformWalkEquivalent(other, "", func(_ string, self, _ *TypeDef, _ bool) bool {
		self.EnsureExtensions()
		self.Extensions.SetIfAbsent(extensionName, extensionValue)

		return false
	}, false)
}

// IsTerraformSubsetOf returns true if every field, associated type, and item
// type in t has a structural equivalent in other (and those equivalents are
// themselves subsets). Used to determine whether a read operation must follow
// another operation to capture fields not present in that operation's response.
func (t *TypeDef) IsTerraformSubsetOf(other *TypeDef) bool {
	if t == nil || other == nil {
		return true
	}

	for _, field := range t.Fields {
		equivalent, _ := field.FindTerraformEquivalentField(other.Fields, true)

		if equivalent == nil {
			return false
		}

		if !field.Type.IsTerraformSubsetOf(equivalent.Type) {
			return false
		}
	}

	for _, associated := range t.AssociatedTypes {
		selfName := t.AssociatedTypeName(associated)

		var equivalent *TypeDef

		for _, otherAssociated := range other.AssociatedTypes {
			if other.AssociatedTypeName(otherAssociated) == selfName {
				equivalent = otherAssociated

				break
			}
		}

		if equivalent != nil && !associated.IsTerraformSubsetOf(equivalent) {
			return false
		}
	}

	if t.ItemType != nil && other.ItemType != nil {
		if !t.ItemType.IsTerraformSubsetOf(other.ItemType) {
			return false
		}
	}

	return true
}

// IsTerraformSubsetOfEntity finds the entity TypeDef in both t and other, then
// returns whether the entity in t is a structural subset of the entity in
// other. Returns true (is a subset) when either entity cannot be found.
func (t *TypeDef) IsTerraformSubsetOfEntity(entityName string, other *TypeDef) bool {
	if t == nil || other == nil {
		return true
	}

	entityA := t.FindEntityTypeDef(entityName)
	entityB := other.FindEntityTypeDef(entityName)

	if entityA == nil || entityB == nil {
		return true
	}

	return entityA.IsTerraformSubsetOf(entityB)
}

// OverrideExtensionOnEquivalentTypes is like SetExtensionOnEquivalentTypes but
// forces the value regardless of whether the extension already exists. This is
// useful for correcting previously-set extensions, such as overriding
// x-speakeasy-param-readonly from true to false on fields that appear in
// per-operation request shards.
func (t *TypeDef) OverrideExtensionOnEquivalentTypes(other *TypeDef, extensionName string, extensionValue any) {
	t.TerraformWalkEquivalent(other, "", func(_ string, self, _ *TypeDef, _ bool) bool {
		self.EnsureExtensions()
		self.Extensions.Set(extensionName, extensionValue)

		return false
	}, false)
}

// OverrideExtensionOnAliasMatchedFields walks the source TypeDef fields and for
// each field with an x-speakeasy-match configuration (MatchConfig.Path), resolves
// the target field in the receiver using alias-aware matching, then overrides
// the specified extension on the matched receiver field. This handles cases
// where source fields (e.g., path parameters) alias to receiver fields via
// different names that OverrideExtensionOnEquivalentTypes cannot resolve.
func (t *TypeDef) OverrideExtensionOnAliasMatchedFields(source *TypeDef, extensionName string, extensionValue any) {
	if t == nil || source == nil {
		return
	}

	for _, sourceField := range source.Fields {
		if sourceField.Type == nil || sourceField.Type.Extensions == nil ||
			sourceField.Type.Extensions.MatchConfig == nil ||
			sourceField.Type.Extensions.MatchConfig.Path == nil ||
			*sourceField.Type.Extensions.MatchConfig.Path == "" {
			continue
		}

		matched, _ := sourceField.FindTerraformEquivalentField(t.Fields, true)
		if matched == nil || matched.Type == nil {
			continue
		}

		matched.Type.EnsureExtensions()
		matched.Type.Extensions.Set(extensionName, extensionValue)
	}
}

// SetExtensionOnExclusiveTypes recursively tags the receiver TypeDef at fields
// that are present in the whenIn TypeDef but absent from the butNotIn TypeDef.
//
// For each field in whenIn:
//   - If found in both receiver and butNotIn: recurse deeper to check nested levels.
//   - If found in receiver but NOT in butNotIn: tag the receiver field's entire
//     subtree via SetExtensionRecursive.
//
// The same logic applies to AssociatedTypes and ItemType.
func (t *TypeDef) SetExtensionOnExclusiveTypes(whenIn *TypeDef, butNotIn *TypeDef, extensionName string, extensionValue any) {
	if t == nil || whenIn == nil {
		return
	}

	for _, whenInField := range whenIn.Fields {
		receiverField, _ := whenInField.FindTerraformEquivalentField(t.Fields, false)

		if receiverField == nil || receiverField.Type == nil {
			continue
		}

		var butNotInField *FieldDef

		if butNotIn != nil {
			butNotInField, _ = whenInField.FindTerraformEquivalentField(butNotIn.Fields, false)
		}

		if butNotInField != nil && butNotInField.Type != nil && whenInField.Type != nil {
			// Field exists in both whenIn and butNotIn: recurse to find
			// nested exclusive fields.
			receiverField.Type.SetExtensionOnExclusiveTypes(whenInField.Type, butNotInField.Type, extensionName, extensionValue)
		} else {
			// Field is in whenIn but not in butNotIn: tag the receiver's
			// entire subtree.
			receiverField.Type.SetExtensionRecursive(extensionName, extensionValue)
		}
	}

	for _, whenInAssociated := range whenIn.AssociatedTypes {
		whenInName := whenIn.AssociatedTypeName(whenInAssociated)

		var receiverAssociated *TypeDef

		for _, candidate := range t.AssociatedTypes {
			if t.AssociatedTypeName(candidate) == whenInName {
				receiverAssociated = candidate

				break
			}
		}

		if receiverAssociated == nil {
			continue
		}

		var butNotInAssociated *TypeDef

		if butNotIn != nil {
			for _, candidate := range butNotIn.AssociatedTypes {
				if butNotIn.AssociatedTypeName(candidate) == whenInName {
					butNotInAssociated = candidate

					break
				}
			}
		}

		if butNotInAssociated != nil {
			receiverAssociated.SetExtensionOnExclusiveTypes(whenInAssociated, butNotInAssociated, extensionName, extensionValue)
		} else {
			receiverAssociated.SetExtensionRecursive(extensionName, extensionValue)
		}
	}

	if t.ItemType != nil && whenIn.ItemType != nil {
		if butNotIn != nil && butNotIn.ItemType != nil {
			t.ItemType.SetExtensionOnExclusiveTypes(whenIn.ItemType, butNotIn.ItemType, extensionName, extensionValue)
		}
	}
}

// RemoveExtensionOnEquivalentTypes recursively removes an extension from this
// TypeDef at types that have equivalents in the other TypeDef. It follows the
// same walk structure as SetExtensionOnEquivalentTypes, but deletes the
// extension instead of setting it.
//
// Special handling: x-speakeasy-ignore sets Extensions.Ignore to false instead
// of deleting from the All map.
func (t *TypeDef) RemoveExtensionOnEquivalentTypes(other *TypeDef, extensionName string) {
	t.TerraformWalkEquivalent(other, "", func(_ string, self, _ *TypeDef, _ bool) bool {
		self.EnsureExtensions()

		if extensionName == extensions.ExtIgnore.Name() {
			falseBool := false
			self.Extensions.Ignore = &falseBool
		} else {
			self.Extensions.Remove(extensionName)
		}

		return false
	}, false)
}

// SetTerraformIgnoreOnExclusiveTypes recursively sets TerraformIgnore on the
// receiver TypeDef at fields that are present in the whenIn TypeDef but absent
// from the butNotIn TypeDef.
//
// This follows the same walk structure as SetExtensionOnExclusiveTypes, but
// instead of setting an extension via SetExtensionRecursive, it recursively sets
// TerraformIgnore on exclusive subtrees.
func (t *TypeDef) SetTerraformIgnoreOnExclusiveTypes(whenIn *TypeDef, butNotIn *TypeDef, dataModel, schema bool) {
	if t == nil || whenIn == nil {
		return
	}

	for _, whenInField := range whenIn.Fields {
		receiverField, _ := whenInField.FindTerraformEquivalentField(t.Fields, false)

		if receiverField == nil || receiverField.Type == nil {
			continue
		}

		var butNotInField *FieldDef

		if butNotIn != nil {
			butNotInField, _ = whenInField.FindTerraformEquivalentField(butNotIn.Fields, false)
		}

		if butNotInField != nil && butNotInField.Type != nil && whenInField.Type != nil {
			receiverField.Type.SetTerraformIgnoreOnExclusiveTypes(whenInField.Type, butNotInField.Type, dataModel, schema)
		} else {
			setTerraformIgnoreRecursive(receiverField.Type, dataModel, schema)
		}
	}

	for _, whenInAssociated := range whenIn.AssociatedTypes {
		whenInName := whenIn.AssociatedTypeName(whenInAssociated)

		var receiverAssociated *TypeDef

		for _, candidate := range t.AssociatedTypes {
			if t.AssociatedTypeName(candidate) == whenInName {
				receiverAssociated = candidate

				break
			}
		}

		if receiverAssociated == nil {
			continue
		}

		var butNotInAssociated *TypeDef

		if butNotIn != nil {
			for _, candidate := range butNotIn.AssociatedTypes {
				if butNotIn.AssociatedTypeName(candidate) == whenInName {
					butNotInAssociated = candidate

					break
				}
			}
		}

		if butNotInAssociated != nil {
			receiverAssociated.SetTerraformIgnoreOnExclusiveTypes(whenInAssociated, butNotInAssociated, dataModel, schema)
		} else {
			setTerraformIgnoreRecursive(receiverAssociated, dataModel, schema)
		}
	}

	if t.ItemType != nil && whenIn.ItemType != nil {
		if butNotIn != nil && butNotIn.ItemType != nil {
			t.ItemType.SetTerraformIgnoreOnExclusiveTypes(whenIn.ItemType, butNotIn.ItemType, dataModel, schema)
		}
	}
}

// setTerraformIgnoreRecursive is a helper that recursively sets TerraformIgnore
// on all types in the tree.
func setTerraformIgnoreRecursive(t *TypeDef, dataModel, schema bool) {
	if t == nil {
		return
	}

	t.EnsureExtensions()

	t.Extensions.TerraformIgnore = &extensions.TerraformIgnore{
		DataModel: dataModel,
		Schema:    schema,
	}

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			setTerraformIgnoreRecursive(field.Type, dataModel, schema)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			setTerraformIgnoreRecursive(associatedType, dataModel, schema)
		}
	}

	if t.ItemType != nil {
		setTerraformIgnoreRecursive(t.ItemType, dataModel, schema)
	}
}

// RenameExtensionRecursive recursively renames an extension key throughout the
// TypeDef tree. At each node, if Extensions.All[fromName] exists, its value is
// moved to Extensions.All[toName] and the old key is deleted.
func (t *TypeDef) RenameExtensionRecursive(fromName string, toName string) {
	if t == nil {
		return
	}

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.RenameExtensionRecursive(fromName, toName)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.RenameExtensionRecursive(fromName, toName)
		}
	}

	if t.ItemType != nil {
		t.ItemType.RenameExtensionRecursive(fromName, toName)
	}

	if t.Extensions != nil {
		t.Extensions.Rename(fromName, toName)
	}
}

// RemoveExtensionRecursive recursively removes an extension from all types in
// the TypeDef tree.
func (t *TypeDef) RemoveExtensionRecursive(extensionName string) {
	if t == nil {
		return
	}

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.RemoveExtensionRecursive(extensionName)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.RemoveExtensionRecursive(extensionName)
		}
	}

	if t.ItemType != nil {
		t.ItemType.RemoveExtensionRecursive(extensionName)
	}

	t.EnsureExtensions()
	t.Extensions.Remove(extensionName)
}

// TerraformImportRequiredFields recursively filters the TypeDef's Fields to
// only those required for Terraform import state operations (as determined by
// FieldDef.IsTerraformImportRequired), then sorts remaining fields by name.
// Recurses into remaining Field Types, AssociatedTypes, and ItemType.
// Mutates the receiver in place; clone before calling to preserve the original.
func (t *TypeDef) TerraformImportRequiredFields() {
	if t == nil {
		return
	}

	t.Fields = slices.DeleteFunc(t.Fields, func(f *FieldDef) bool {
		return f == nil || !f.IsTerraformImportRequired()
	})

	t.SortFieldsByName()

	for _, field := range t.Fields {
		if field.Type != nil {
			field.Type.TerraformImportRequiredFields()
		}
	}

	for _, associated := range t.AssociatedTypes {
		associated.TerraformImportRequiredFields()
	}

	if t.ItemType != nil {
		t.ItemType.TerraformImportRequiredFields()
	}
}

// TerraformInvalidImportType describes a TypeDef node with a DataType that is
// not supported for Terraform import state operations.
type TerraformInvalidImportType struct {
	// Hierarchy is the dot-separated path to the node within the TypeDef tree.
	Hierarchy string

	// TypeName is the string representation of the unsupported DataType.
	TypeName string
}

// validTerraformImportDataTypes is the set of DataTypes supported for Terraform import
// state operations.
var validTerraformImportDataTypes = map[DataType]bool{
	DataTypeBoolean: true,
	DataTypeClass:   true,
	DataTypeEnum:    true,
	DataTypeFloat32: true,
	DataTypeInt32:   true,
	DataTypeInteger: true,
	DataTypeNumber:  true,
	DataTypeString:  true,
}

// TerraformInvalidImportTypes walks the TypeDef tree and returns all nodes
// with DataTypes not supported for Terraform import state operations. Returns
// nil if all types are valid.
func (t *TypeDef) TerraformInvalidImportTypes() []*TerraformInvalidImportType {
	if t == nil {
		return nil
	}

	var result []*TerraformInvalidImportType

	t.TerraformWalkTypes("", func(hierarchy string, typedef *TypeDef, _ bool) {
		if typedef == nil {
			return
		}

		if !validTerraformImportDataTypes[typedef.Type] {
			result = append(result, &TerraformInvalidImportType{
				Hierarchy: hierarchy,
				TypeName:  string(typedef.Type),
			})
		}
	}, false)

	if len(result) == 0 {
		return nil
	}

	return result
}

// TerraformHasInvalidImportTypes returns true if any node in the TypeDef tree
// has a DataType not supported for Terraform import state operations.
func (t *TypeDef) TerraformHasInvalidImportTypes() bool {
	return len(t.TerraformInvalidImportTypes()) > 0
}

// TerraformDeleteIgnored recursively removes Fields and AssociatedTypes that
// are marked as ignored for Terraform purposes. A field is removed when its
// Type has Extensions.TerraformIgnore.DataModel set, Extensions.Ignore set to
// true, or the field has a Const value. An associated type is removed when it
// has TerraformIgnore.DataModel or Ignore set. Remaining children are recursed
// into. Mutates the receiver in place.
func (t *TypeDef) TerraformDeleteIgnored() {
	if t == nil {
		return
	}

	t.Fields = slices.DeleteFunc(t.Fields, func(f *FieldDef) bool {
		if f == nil || f.Type == nil {
			return false
		}

		if f.Const != nil {
			return true
		}

		if f.Type.Extensions != nil {
			if f.Type.Extensions.TerraformIgnore != nil && f.Type.Extensions.TerraformIgnore.DataModel {
				return true
			}

			if f.Type.Extensions.Ignore != nil && *f.Type.Extensions.Ignore {
				return true
			}
		}

		f.Type.TerraformDeleteIgnored()

		return false
	})

	t.AssociatedTypes = slices.DeleteFunc(t.AssociatedTypes, func(at *TypeDef) bool {
		if at == nil {
			return false
		}

		if at.Extensions != nil {
			if at.Extensions.TerraformIgnore != nil && at.Extensions.TerraformIgnore.DataModel {
				return true
			}

			if at.Extensions.Ignore != nil && *at.Extensions.Ignore {
				return true
			}
		}

		at.TerraformDeleteIgnored()

		return false
	})

	if t.ItemType != nil {
		t.ItemType.TerraformDeleteIgnored()
	}
}

// TerraformMergeAliasNodes merges only the alias nodes (fields/types with
// x-speakeasy-match configuration) from other into the receiver. The other
// TypeDef is cloned, hoisted by entity name, filtered to only alias nodes,
// then merged via TerraformMerge.
func (t *TypeDef) TerraformMergeAliasNodes(entityName string, other *TypeDef) error {
	if t == nil || other == nil {
		return nil
	}

	hoisted := other.Clone()

	if err := hoisted.TerraformHoistByEntityName(entityName); err != nil {
		return err
	}

	// Filter to only alias nodes: fields/types with MatchConfig.Path set.
	hoisted.Fields = slices.DeleteFunc(hoisted.Fields, func(f *FieldDef) bool {
		if f == nil || f.Type == nil {
			return true
		}

		return f.Type.Extensions == nil ||
			f.Type.Extensions.MatchConfig == nil ||
			f.Type.Extensions.MatchConfig.Path == nil ||
			*f.Type.Extensions.MatchConfig.Path == ""
	})

	hoisted.AssociatedTypes = slices.DeleteFunc(hoisted.AssociatedTypes, func(at *TypeDef) bool {
		if at == nil {
			return true
		}

		return at.Extensions == nil ||
			at.Extensions.MatchConfig == nil ||
			at.Extensions.MatchConfig.Path == nil ||
			*at.Extensions.MatchConfig.Path == ""
	})

	return t.TerraformMerge(hoisted)
}

// TerraformSetConflictsWith walks the TypeDef tree and, for each node that has
// AssociatedTypes (union), sets x-speakeasy-conflicts-with on each branch to
// the names of the other branches. Also sets x-speakeasy-param-computed-override
// to false and defaults x-speakeasy-param-suppress-computed-diff to false on
// each branch.
//
// The nameFunc parameter provides the display name for each associated type,
// allowing the caller to control naming.
func (t *TypeDef) TerraformSetConflictsWith(nameFunc func(parent *TypeDef, associated *TypeDef) string) {
	if t == nil {
		return
	}

	if len(t.AssociatedTypes) > 0 {
		for _, subType := range t.AssociatedTypes {
			if subType == nil {
				continue
			}

			var otherNames []string
			for _, other := range t.AssociatedTypes {
				if other != subType {
					otherNames = append(otherNames, nameFunc(t, other))
				}
			}

			subType.EnsureExtensions()

			if len(otherNames) > 0 {
				// Store as a fresh []any slice with simple assignment (not append).
				// Each branch gets its own newly-allocated slice, avoiding
				// duplication from shared pointers after Clone().
				result := make([]any, len(otherNames))
				for i, name := range otherNames {
					result[i] = name
				}
				subType.Extensions.Set("x-speakeasy-conflicts-with", result)
			}

			subType.Extensions.Set("x-speakeasy-param-computed-override", false)
			subType.Extensions.SetIfAbsent("x-speakeasy-param-suppress-computed-diff", false)
		}
	}

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.TerraformSetConflictsWith(nameFunc)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.TerraformSetConflictsWith(nameFunc)
		}
	}

	if t.ItemType != nil {
		t.ItemType.TerraformSetConflictsWith(nameFunc)
	}
}

// PropagateExtensionValueRecursive propagates an extension value downward
// through the TypeDef tree. If a node has the extension set, its value is
// propagated to all descendants that don't already have the key present.
// Uses key-presence semantics (not truthiness).
func (t *TypeDef) PropagateExtensionValueRecursive(extensionName string) {
	t.propagateExtensionValueRecursive(extensionName, nil, false)
}

func (t *TypeDef) propagateExtensionValueRecursive(extensionName string, propagating any, hasPropagating bool) {
	if t == nil {
		return
	}

	t.EnsureExtensions()

	if !t.Extensions.Has(extensionName) && hasPropagating {
		t.Extensions.Set(extensionName, propagating)
	} else if v, ok := t.Extensions.Get(extensionName); ok && v != nil {
		propagating = v
		hasPropagating = true
	}

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.propagateExtensionValueRecursive(extensionName, propagating, hasPropagating)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.propagateExtensionValueRecursive(extensionName, propagating, hasPropagating)
		}
	}

	if t.ItemType != nil {
		t.ItemType.propagateExtensionValueRecursive(extensionName, propagating, hasPropagating)
	}
}

// TerraformPropagateComputedParent walks the TypeDef tree and, for optional
// container types (map/array/set/class/union) that have x-speakeasy-param-computed
// or x-speakeasy-param-readonly, sets both x-speakeasy-parent-require-to-not-null
// and x-speakeasy-param-computed on the container and all its descendants.
//
// Instead of using SetExtensionRecursive (which modifies entire subtrees and can
// contaminate shared TypeDef pointers), this propagates the "computed parent"
// context through the walk itself. Each node tags itself based on its walk
// position, avoiding cross-path contamination on shared nodes.
func (t *TypeDef) TerraformPropagateComputedParent() {
	visited := make(map[*TypeDef]bool)
	t.terraformPropagateComputedParent(false, false, visited)
}

func (t *TypeDef) terraformPropagateComputedParent(optional bool, inComputedParent bool, visited map[*TypeDef]bool) {
	if t == nil {
		return
	}

	// Skip already-visited nodes to prevent shared TypeDef pointers from
	// being processed multiple times from different paths.
	if visited[t] {
		return
	}
	visited[t] = true

	t.EnsureExtensions()

	// If we're a descendant of a computed parent container, tag this node
	// (matching the recursive tag() behavior from the old JS code).
	if inComputedParent {
		t.Extensions.SetIfAbsent("x-speakeasy-parent-require-to-not-null", true)
		t.Extensions.SetIfAbsent("x-speakeasy-param-computed", true)
	}

	// Check if this node is itself a computed+optional container
	propagateDown := inComputedParent
	isContainer := t.Type == DataTypeMap || t.Type == DataTypeArray || t.Type == DataTypeSet || t.Type == DataTypeClass || t.Type == DataTypeUnion
	if isContainer && optional {
		computed, _ := t.Extensions.All["x-speakeasy-param-computed"].(bool)
		readonly, _ := t.Extensions.All["x-speakeasy-param-readonly"].(bool)
		if computed || readonly {
			t.Extensions.Set("x-speakeasy-parent-require-to-not-null", true)
			t.Extensions.Set("x-speakeasy-param-computed", true)
			propagateDown = true
		}
	}

	for _, field := range t.Fields {
		if field != nil && field.Type != nil {
			field.Type.terraformPropagateComputedParent(field.Optional || field.Nullable, propagateDown, visited)
		}
	}

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.terraformPropagateComputedParent(true, propagateDown, visited)
		}
	}

	if t.ItemType != nil {
		t.ItemType.terraformPropagateComputedParent(true, propagateDown, visited)
	}
}

// RemoveEmptyObjectsRecursive recursively removes Fields whose Type is "class"
// with no Fields and no AssociatedTypes. Recurses into remaining children.
// Mutates the receiver in place.
func (t *TypeDef) RemoveEmptyObjectsRecursive() {
	if t == nil {
		return
	}

	t.Fields = slices.DeleteFunc(t.Fields, func(f *FieldDef) bool {
		if f == nil || f.Type == nil {
			return false
		}

		if f.Type.Type == DataTypeClass && len(f.Type.Fields) == 0 && len(f.Type.AssociatedTypes) == 0 {
			return true
		}

		f.Type.RemoveEmptyObjectsRecursive()

		return false
	})

	for _, associatedType := range t.AssociatedTypes {
		if associatedType != nil {
			associatedType.RemoveEmptyObjectsRecursive()
		}
	}

	if t.ItemType != nil {
		t.ItemType.RemoveEmptyObjectsRecursive()
	}
}

// ValidateMatchConfigTypes walks the requestShard TypeDef and validates that
// each field with an x-speakeasy-match configuration (MatchConfig.Path) can be
// resolved in the receiver TypeDef and that the types are compatible. Returns
// an error containing all validation failures.
func (t *TypeDef) ValidateMatchConfigTypes(requestShard *TypeDef, opString string) error {
	if t == nil || requestShard == nil {
		return nil
	}

	var errors []string

	requestShard.TerraformWalkTypes(opString+".request", func(hierarchy string, typedef *TypeDef, _ bool) {
		if typedef == nil || typedef.Extensions == nil || typedef.Extensions.MatchConfig == nil {
			return
		}

		matchConfig := typedef.Extensions.MatchConfig

		// Skip if only usePriorState is set (no path aliasing)
		if (matchConfig.Path == nil || *matchConfig.Path == "") && matchConfig.UsePriorState {
			return
		}

		if matchConfig.Path == nil || *matchConfig.Path == "" {
			return
		}

		matchPath := *matchConfig.Path

		field, _ := findTerraformMatchPath(t.Fields, matchPath)
		if field == nil {
			errors = append(errors, fmt.Sprintf("x-speakeasy-match error at %q: unable to resolve path %q", hierarchy, matchPath))

			return
		}

		if field.Type == nil {
			return
		}

		if _, err := typedef.TerraformMergeDataType(field.Type); err != nil {
			errors = append(errors, fmt.Sprintf("x-speakeasy-match type mismatch at %q:\n  parameter type: %s\n  match target %q: %s",
				hierarchy, terraformPrettyPrintType(typedef), matchPath, terraformPrettyPrintType(field.Type)))
		}
	}, false)

	if len(errors) > 0 {
		return fmt.Errorf("%s", strings.Join(errors, "\n\n"))
	}

	return nil
}

// TerraformClearReadonlyOnMatchPaths walks the requestShard fields looking for
// x-speakeasy-match paths that contain dots (nested paths). For each such path,
// it resolves intermediate containers in the receiver (schema TypeDef) and
// clears the readonly extension so they become Optional+Computed instead of
// Computed-only. The leaf field is also cleared since it is user-provided
// through the flat request field.
func (t *TypeDef) TerraformClearReadonlyOnMatchPaths(requestShard *TypeDef) {
	if t == nil || requestShard == nil {
		return
	}

	requestShard.TerraformWalkTypes("match-path-clear", func(_ string, typedef *TypeDef, _ bool) {
		if typedef == nil || typedef.Extensions == nil || typedef.Extensions.MatchConfig == nil {
			return
		}

		matchConfig := typedef.Extensions.MatchConfig
		if matchConfig.Path == nil || *matchConfig.Path == "" {
			return
		}

		matchPath := *matchConfig.Path
		if !strings.Contains(matchPath, ".") {
			return
		}

		steps := strings.Split(matchPath, ".")
		currentFields := t.Fields

		for i, step := range steps {
			sanitizedStep := SanitizeFieldName(step)

			var matchField *FieldDef

			for _, field := range currentFields {
				if SanitizeFieldName(field.Name) == sanitizedStep {
					matchField = field
					break
				}
			}

			if matchField == nil || matchField.Type == nil {
				break
			}

			isIntermediate := i < len(steps)-1

			matchField.Type.EnsureExtensions()

			if val, ok := matchField.Type.Extensions.Get("x-speakeasy-param-readonly"); ok && val == true {
				matchField.Type.Extensions.Set("x-speakeasy-param-readonly", false)
				matchField.Type.Extensions.Set("x-speakeasy-param-computed", true)

				if isIntermediate {
					matchField.Type.Extensions.Set("x-speakeasy-param-optional", true)
				}
			}

			currentFields = matchField.Type.Fields
		}
	}, false)
}

// terraformPrettyPrintType returns a human-readable representation of a TypeDef
// for error messages in match config validation.
func terraformPrettyPrintType(t *TypeDef) string {
	if t == nil {
		return "<nil>"
	}

	if t.Enum != nil && t.Enum.Type != nil {
		return string(t.Enum.Type.Type) + "(enum)"
	}

	return string(t.Type)
}

// ValidateAliasNodes walks the given request shard TypeDefs and validates that
// each field with an x-speakeasy-match configuration (alias node) can be
// resolved in the receiver (finalized schema) TypeDef, that the target field
// is not itself an alias (nested aliases not supported), and that the types
// are compatible. Errors are accumulated across all shards, deduplicated, and
// returned as a single error.
func (t *TypeDef) ValidateAliasNodes(shards ...*TypeDef) error {
	if t == nil {
		return nil
	}

	var errors []string

	for _, shard := range shards {
		if shard == nil {
			continue
		}

		shard.TerraformWalkTypes("", func(_ string, typedef *TypeDef, _ bool) {
			if typedef == nil || typedef.Extensions == nil || typedef.Extensions.MatchConfig == nil {
				return
			}

			matchConfig := typedef.Extensions.MatchConfig

			// Skip if only usePriorState is set (no path aliasing).
			if (matchConfig.Path == nil || *matchConfig.Path == "") && matchConfig.UsePriorState {
				return
			}

			if matchConfig.Path == nil || *matchConfig.Path == "" {
				errors = append(errors, "unable to process x-speakeasy-match: configuration must specify either 'path' or 'usePriorState: true'")
				return
			}

			matchPath := *matchConfig.Path

			field, accessor := findTerraformMatchPath(t.Fields, matchPath)
			if field == nil {
				// List available non-alias field names for debugging.
				var available []string
				for _, f := range t.Fields {
					if f.Type != nil && f.Type.Extensions != nil && f.Type.Extensions.MatchConfig != nil {
						continue
					}
					available = append(available, SanitizeFieldName(f.Name))
				}
				errors = append(errors, fmt.Sprintf("unable to resolve x-speakeasy-match path %q: %q not in %q. Any x-speakeasy-match used during update must also be used during creation.",
					matchPath, matchPath, strings.Join(available, ",")))
				return
			}

			// Check for nested aliases — target field also has MatchConfig.
			if field.Type != nil && field.Type.Extensions != nil && field.Type.Extensions.MatchConfig != nil {
				accessorStr := strings.Join(accessor, ".")
				errors = append(errors, fmt.Sprintf("unable to resolve x-speakeasy-match path %q: target field %q also has x-speakeasy-match (nested aliases not supported)",
					matchPath, accessorStr))
				return
			}

			// Check type compatibility.
			if field.Type != nil {
				if _, err := typedef.TerraformMergeDataType(field.Type); err != nil {
					accessorStr := strings.Join(accessor, ".")
					errors = append(errors, fmt.Sprintf("unable to resolve x-speakeasy-match path %q: type mismatch at %q - %s",
						matchPath, accessorStr, err.Error()))
				}
			}
		}, false)
	}

	// Deduplicate errors.
	seen := make(map[string]bool)
	unique := errors[:0]
	for _, e := range errors {
		if !seen[e] {
			seen[e] = true
			unique = append(unique, e)
		}
	}

	if len(unique) > 0 {
		return fmt.Errorf("%s", strings.Join(unique, "\n"))
	}

	return nil
}

// collectSoftDeleteProperties scans the given TypeDef (typically a read
// response shard) for fields with the x-speakeasy-soft-delete-property
// extension set to true. It checks both top-level fields and fields within
// associated types (to account for hoisting).
func collectSoftDeleteProperties(readResponseShard *TypeDef) Fields {
	if readResponseShard == nil {
		return nil
	}

	var result Fields

	for _, associatedType := range readResponseShard.AssociatedTypes {
		for _, field := range associatedType.Fields {
			if field.Type != nil && field.Type.Extensions != nil && field.Type.Extensions.All != nil {
				if v, ok := field.Type.Extensions.All["x-speakeasy-soft-delete-property"]; ok && v == true {
					result = append(result, field)
				}
			}
		}
	}

	for _, field := range readResponseShard.Fields {
		if field.Type != nil && field.Type.Extensions != nil && field.Type.Extensions.All != nil {
			if v, ok := field.Type.Extensions.All["x-speakeasy-soft-delete-property"]; ok && v == true {
				result = append(result, field)
			}
		}
	}

	return result
}

// SortFieldsByName sorts the TypeDef's Fields slice in place by field Name
// using lexicographic comparison.
func (t *TypeDef) SortFieldsByName() {
	if t == nil {
		return
	}

	slices.SortFunc(t.Fields, func(a, b *FieldDef) int {
		return strings.Compare(a.Name, b.Name)
	})
}

// TerraformWalkEquivalent walks the receiver and other TypeDef trees in
// parallel, calling fn on each pair of matched nodes. Only recurses into
// subtrees where the receiver has an equivalent in other (matched via
// FindTerraformEquivalentField for fields, AssociatedTypeName for associated
// types, and directly for ItemType). The callback returns true to stop
// recursion at that branch.
func (t *TypeDef) TerraformWalkEquivalent(other *TypeDef, hierarchy string, fn func(hierarchy string, self *TypeDef, other *TypeDef, optional bool) bool, optional bool) {
	if t == nil || other == nil {
		return
	}

	if fn(hierarchy, t, other, optional) {
		return
	}

	for _, field := range t.Fields {
		equivalent, _ := field.FindTerraformEquivalentField(other.Fields, false)

		if equivalent == nil || field.Type == nil || equivalent.Type == nil {
			continue
		}

		field.Type.TerraformWalkEquivalent(
			equivalent.Type,
			hierarchy+"."+equivalent.Name,
			fn,
			equivalent.Optional || equivalent.Nullable,
		)
	}

	for _, associated := range t.AssociatedTypes {
		targetName := t.AssociatedTypeName(associated)

		for _, otherAssociated := range other.AssociatedTypes {
			if other.AssociatedTypeName(otherAssociated) == targetName {
				associated.TerraformWalkEquivalent(
					otherAssociated,
					hierarchy+"."+targetName,
					fn,
					true,
				)

				break
			}
		}
	}

	if t.ItemType != nil && other.ItemType != nil {
		t.ItemType.TerraformWalkEquivalent(other.ItemType, hierarchy+".[]", fn, false)
	}
}

// TerraformWalkTypes walks the receiver TypeDef tree, calling fn on every
// node with hierarchy tracking. Recurses into Fields (by SanitizeFieldName),
// AssociatedTypes (by AssociatedTypeName), and ItemType.
func (t *TypeDef) TerraformWalkTypes(hierarchy string, fn func(hierarchy string, typedef *TypeDef, optional bool), optional bool) {
	if t == nil {
		return
	}

	fn(hierarchy, t, optional)

	if len(hierarchy) > 10000 {
		return
	}

	for _, field := range t.Fields {
		if field.Type == nil {
			continue
		}

		field.Type.TerraformWalkTypes(
			hierarchy+"."+SanitizeFieldName(field.Name),
			fn,
			field.Optional || field.Nullable,
		)
	}

	for _, associated := range t.AssociatedTypes {
		associated.TerraformWalkTypes(
			hierarchy+"."+t.AssociatedTypeName(associated),
			fn,
			true,
		)
	}

	if t.ItemType != nil {
		t.ItemType.TerraformWalkTypes(hierarchy+".[]", fn, true)
	}

	t.EnsureExtensions()
}

// IsTerraformSymbolEqual reports whether two TypeDefs are structurally equal
// for Terraform symbol deduplication purposes. Unlike IsTerraformEqual, fields
// are matched using sanitized name comparison (SanitizeFieldName) rather than
// exact name matching, which accounts for casing/formatting differences in
// field names from different operation shards.
//
// The TypeDef graph must be acyclic. Circular references are rejected earlier
// by ContainsTruncated checks in generateTerraformAST.
func (t *TypeDef) IsTerraformSymbolEqual(other *TypeDef) bool {
	if t == nil || other == nil {
		return t == other
	}

	if t == other {
		return true
	}

	// Type comparison with enum flexibility: if one has an Enum and the other
	// does not, compare the underlying enum type instead.
	if t.Type != other.Type && t.Enum == nil && other.Enum == nil {
		return false
	}

	if t.Enum != nil || other.Enum != nil {
		tDataType := t.Type
		otherDataType := other.Type

		if t.Enum != nil && t.Enum.Type != nil {
			tDataType = t.Enum.Type.Type
		}

		if other.Enum != nil && other.Enum.Type != nil {
			otherDataType = other.Enum.Type.Type
		}

		if tDataType != otherDataType {
			return false
		}
	}

	// Fields comparison using sanitized name matching.
	if !t.Fields.IsTerraformSymbolEqual(other.Fields) {
		return false
	}

	// AssociatedTypes comparison using AssociatedTypeName-based matching.
	if len(t.AssociatedTypes) != len(other.AssociatedTypes) {
		return false
	}

	for _, associatedType := range t.AssociatedTypes {
		candidateName := t.AssociatedTypeName(associatedType)

		var otherAssociatedType *TypeDef

		for _, otherAT := range other.AssociatedTypes {
			if candidateName == other.AssociatedTypeName(otherAT) {
				otherAssociatedType = otherAT
				break
			}
		}

		if !associatedType.IsTerraformSymbolEqual(otherAssociatedType) {
			return false
		}
	}

	for _, otherAssociatedType := range other.AssociatedTypes {
		candidateName := other.AssociatedTypeName(otherAssociatedType)

		found := false

		for _, selfAT := range t.AssociatedTypes {
			if candidateName == t.AssociatedTypeName(selfAT) {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	// ItemType comparison.
	if (t.ItemType == nil) != (other.ItemType == nil) {
		return false
	}

	if !t.ItemType.IsTerraformSymbolEqual(other.ItemType) {
		return false
	}

	return true
}

// TerraformFinalizeSchema applies the shared post-assembly TypeDef
// modifications that are common to all Terraform entity types (managed
// resource, data resource, ephemeral resource, action). This should be called
// at the end of each entity's AssembleSchemaTypeDef after DeepClone.
//
// The operations are performed in the following order, which must be preserved:
//  1. Set x-speakeasy-root extension
//  2. Set ConflictsWith on union types (TerraformSetConflictsWith)
//  3. Propagate x-speakeasy-param-suppress-computed-diff downward
//  4. Delete ignored fields/types (TerraformDeleteIgnored)
//  5. Propagate computed parent flags (TerraformPropagateComputedParent)
//  6. Sort fields alphabetically (SortFieldsByName)
//  7. Rename manual-readonly back to readonly (RenameExtensionRecursive)
//  8. Conditionally delete empty object schemas (RemoveEmptyObjectsRecursive)
func (t *TypeDef) TerraformFinalizeSchema(excludeEmptyObjectSchemas bool) {
	if t == nil {
		return
	}

	// 1. Mark as root Terraform schema type.
	t.EnsureExtensions()
	t.Extensions.Set("x-speakeasy-root", true)

	// 2. Set ConflictsWith validators on union types. The name function gets
	// the base name via AssociatedTypeName (discriminator mapping →
	// OriginalName → Name → type fallback), then applies PascalCase
	// normalization so the template's sanitizeTFStateName produces correct
	// snake_case attribute names.
	c := casing.New()
	t.TerraformSetConflictsWith(func(parent *TypeDef, associated *TypeDef) string {
		return c.ToGoPascal(parent.AssociatedTypeName(associated))
	})

	// 3. Propagate x-speakeasy-param-suppress-computed-diff values downward
	// through the tree. Must run after TerraformSetConflictsWith because
	// ConflictsWith sets suppress-computed-diff=false on union variants and
	// propagation should not overwrite those explicit values.
	t.PropagateExtensionValueRecursive("x-speakeasy-param-suppress-computed-diff")

	// 4. Delete ignored fields and types (x-speakeasy-ignore, TerraformIgnore,
	// Const fields).
	t.TerraformDeleteIgnored()

	// 5. Propagate computed parent flags to child nodes of optional computed
	// containers.
	t.TerraformPropagateComputedParent()

	// 6. Sort top-level fields alphabetically by name.
	t.SortFieldsByName()

	// 7. Rename the temporary manual-readonly extension back to the standard
	// readonly extension name. This reverses the rename done at the start of
	// assembly to protect manually-set readonly values from being overwritten
	// during merge operations.
	t.RenameExtensionRecursive("x-speakeasy-manual-param-readonly", "x-speakeasy-param-readonly")

	// 8. Conditionally remove empty object schemas. This must run after
	// TerraformDeleteIgnored (step 4) which may create newly-empty objects,
	// and after TerraformPropagateComputedParent (step 5) which should
	// operate on the pre-pruned tree.
	if excludeEmptyObjectSchemas {
		t.RemoveEmptyObjectsRecursive()
	}
}
