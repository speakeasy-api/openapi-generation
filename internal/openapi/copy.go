package openapi

import (
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/speakeasy-api/openapi/values"
)

// Copy attempts to do a deep copy of the Schema struct so it can be used for merging
// It is not a perfect copy, but works for its intended use case
//
// Copy complex pointer fields that are modified in merge operations
// These need to be copied to avoid modifying the original schema

// Note: We intentionally do NOT deep copy the JSONSchema pointers themselves,
// as that would require recursive copying of the entire schema tree.
// We only copy the pointer references so modifications to the parent schema
// don't affect which child schemas are referenced.

// The following fields are shallow copied (pointer copied, not deep copied):
// - AdditionalProperties, Contains, If, Else, Then, PropertyNames
// - UnevaluatedItems, UnevaluatedProperties, Items, Not
// - Discriminator, ExternalDocs, XML
// - Default, Const, Example (values.Value)
// - Ref (references.Reference)
// - ExclusiveMaximum, ExclusiveMinimum (already copied via struct copy)

// These are already copied by the initial struct copy (c := *s) and don't need
// additional handling since they are either not modified in merge operations
// or are simple value types that are safe to share.
func Copy(s *oas3.Schema) oas3.Schema {
	c := *s

	// Copy pointer fields to avoid modifying original
	if s.Title != nil {
		title := *s.Title
		c.Title = &title
	}

	if s.Description != nil {
		desc := *s.Description
		c.Description = &desc
	}

	if s.MultipleOf != nil {
		val := *s.MultipleOf
		c.MultipleOf = &val
	}

	if s.Maximum != nil {
		val := *s.Maximum
		c.Maximum = &val
	}

	if s.Minimum != nil {
		val := *s.Minimum
		c.Minimum = &val
	}

	if s.MaxLength != nil {
		val := *s.MaxLength
		c.MaxLength = &val
	}

	if s.MinLength != nil {
		val := *s.MinLength
		c.MinLength = &val
	}

	if s.Pattern != nil {
		val := *s.Pattern
		c.Pattern = &val
	}

	if s.Format != nil {
		val := *s.Format
		c.Format = &val
	}

	if s.MaxItems != nil {
		val := *s.MaxItems
		c.MaxItems = &val
	}

	if s.MinItems != nil {
		val := *s.MinItems
		c.MinItems = &val
	}

	if s.UniqueItems != nil {
		val := *s.UniqueItems
		c.UniqueItems = &val
	}

	if s.MaxProperties != nil {
		val := *s.MaxProperties
		c.MaxProperties = &val
	}

	if s.MinProperties != nil {
		val := *s.MinProperties
		c.MinProperties = &val
	}

	if s.Nullable != nil {
		val := *s.Nullable
		c.Nullable = &val
	}

	if s.ReadOnly != nil {
		val := *s.ReadOnly
		c.ReadOnly = &val
	}

	if s.WriteOnly != nil {
		val := *s.WriteOnly
		c.WriteOnly = &val
	}

	if s.Deprecated != nil {
		val := *s.Deprecated
		c.Deprecated = &val
	}

	if s.Schema != nil {
		val := *s.Schema
		c.Schema = &val
	}

	if s.Anchor != nil {
		val := *s.Anchor
		c.Anchor = &val
	}

	if s.MinContains != nil {
		val := *s.MinContains
		c.MinContains = &val
	}

	if s.MaxContains != nil {
		val := *s.MaxContains
		c.MaxContains = &val
	}

	// Copy slice fields
	if s.AllOf != nil {
		c.AllOf = make([]*oas3.JSONSchema[oas3.Referenceable], len(s.AllOf))
		copy(c.AllOf, s.AllOf)
	}

	if s.AnyOf != nil {
		c.AnyOf = make([]*oas3.JSONSchema[oas3.Referenceable], len(s.AnyOf))
		copy(c.AnyOf, s.AnyOf)
	}

	if s.OneOf != nil {
		c.OneOf = make([]*oas3.JSONSchema[oas3.Referenceable], len(s.OneOf))
		copy(c.OneOf, s.OneOf)
	}

	if s.PrefixItems != nil {
		c.PrefixItems = make([]*oas3.JSONSchema[oas3.Referenceable], len(s.PrefixItems))
		copy(c.PrefixItems, s.PrefixItems)
	}

	if s.Examples != nil {
		c.Examples = make([]values.Value, len(s.Examples))
		copy(c.Examples, s.Examples)
	}

	if s.Enum != nil {
		c.Enum = make([]values.Value, len(s.Enum))
		copy(c.Enum, s.Enum)
	}

	if s.Required != nil {
		c.Required = make([]string, len(s.Required))
		copy(c.Required, s.Required)
	}

	// Copy map fields
	if s.DependentSchemas != nil {
		c.DependentSchemas = sequencedmap.From(s.DependentSchemas.All())
	}

	if s.PatternProperties != nil {
		c.PatternProperties = sequencedmap.From(s.PatternProperties.All())
	}

	if s.Properties != nil {
		c.Properties = sequencedmap.From(s.Properties.All())
	}

	if s.Defs != nil {
		c.Defs = sequencedmap.From(s.Defs.All())
	}

	// Copy Extensions
	if s.Extensions != nil {
		c.Extensions = &extensions.Extensions{
			Map: sequencedmap.From(s.Extensions.All()),
		}

		// If we are making a copy we shouldn't persist x-speakeasy-reference-override as then the derived schema will also be treated as a reference which can cause issues with
		if c.Extensions.Has(xReferenceOverride) {
			c.Extensions.Delete(xReferenceOverride)
		}
	}

	// Copy Type field
	if s.Type != nil {
		if s.Type.IsLeft() {
			c.Type = oas3.NewTypeFromArray(s.Type.LeftValue())
		} else {
			c.Type = oas3.NewTypeFromString(s.Type.RightValue())
		}
	}

	return c
}
