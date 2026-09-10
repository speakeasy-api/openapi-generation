package openapi

import (
	"context"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
)

func Merge(ctx context.Context, baseSchema *oas3.Schema, overridingSchema *oas3.Schema, mergeDescriptiveFields, allOfMerge bool, e speakeasyExtensions, allOfMergeStrategy config.AllOfMergeStrategy, docInfo *document.DocumentInfo) error {
	deepMerge := allOfMerge && allOfMergeStrategy == config.AllOfMergeStrategyDeepMerge

	// Type modifying fields
	if overridingSchema.AdditionalProperties != nil {
		if !deepMerge {
			baseSchema.AdditionalProperties = overridingSchema.AdditionalProperties
		} else {
			var err error
			baseSchema.AdditionalProperties, err = deepMergeJSONSchema(ctx, baseSchema.AdditionalProperties, overridingSchema.AdditionalProperties, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		}
	}

	if len(overridingSchema.AllOf) > 0 {
		baseSchema.AllOf = append(baseSchema.AllOf, overridingSchema.AllOf...)
	}

	if overridingSchema.Anchor != nil {
		baseSchema.Anchor = overridingSchema.Anchor
	}

	if len(overridingSchema.AnyOf) > 0 {
		baseSchema.AnyOf = append(baseSchema.AnyOf, overridingSchema.AnyOf...)
	}

	if overridingSchema.Const != nil {
		baseSchema.Const = overridingSchema.Const
	}

	if overridingSchema.Contains != nil {
		if deepMerge {
			var err error
			baseSchema.Contains, err = deepMergeJSONSchema(ctx, baseSchema.Contains, overridingSchema.Contains, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		} else {
			baseSchema.Contains = overridingSchema.Contains
		}
	}

	if overridingSchema.Default != nil {
		baseSchema.Default = overridingSchema.Default
	}

	if overridingSchema.Discriminator != nil && allOfMerge {
		baseSchema.Discriminator = overridingSchema.Discriminator
	}

	if len(overridingSchema.Enum) > 0 {
		baseSchema.Enum = append(baseSchema.Enum, overridingSchema.Enum...)
	}

	if overridingSchema.ExclusiveMaximum != nil {
		baseSchema.ExclusiveMaximum = overridingSchema.ExclusiveMaximum
	}

	if overridingSchema.ExclusiveMinimum != nil {
		baseSchema.ExclusiveMinimum = overridingSchema.ExclusiveMinimum
	}

	if overridingSchema.GetExtensions().Len() > 0 {
		if baseSchema.Extensions == nil {
			baseSchema.Extensions = extensions.New()
		}

		for extension, node := range overridingSchema.Extensions.All() {
			// There are certain extensions that are mergeable
			if e.IsExtensionMergable(extension) {
				baseSchema.Extensions.Set(extension, node)
			}
		}
	}

	if overridingSchema.Format != nil {
		baseSchema.Format = overridingSchema.Format
	}

	if overridingSchema.Items != nil {
		if deepMerge {
			var err error
			baseSchema.Items, err = deepMergeJSONSchema(ctx, baseSchema.Items, overridingSchema.Items, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		} else {
			baseSchema.Items = overridingSchema.Items
		}
	}

	if overridingSchema.MaxContains != nil {
		baseSchema.MaxContains = overridingSchema.MaxContains
	}

	if overridingSchema.Maximum != nil {
		baseSchema.Maximum = overridingSchema.Maximum
	}

	if overridingSchema.MaxItems != nil {
		baseSchema.MaxItems = overridingSchema.MaxItems
	}

	if overridingSchema.MaxLength != nil {
		baseSchema.MaxLength = overridingSchema.MaxLength
	}

	if overridingSchema.MaxProperties != nil {
		baseSchema.MaxProperties = overridingSchema.MaxProperties
	}

	if overridingSchema.MinContains != nil {
		baseSchema.MinContains = overridingSchema.MinContains
	}

	if overridingSchema.Minimum != nil {
		baseSchema.Minimum = overridingSchema.Minimum
	}

	if overridingSchema.MinItems != nil {
		baseSchema.MinItems = overridingSchema.MinItems
	}

	if overridingSchema.MinLength != nil {
		baseSchema.MinLength = overridingSchema.MinLength
	}

	if overridingSchema.MinProperties != nil {
		baseSchema.MinProperties = overridingSchema.MinProperties
	}

	if overridingSchema.MultipleOf != nil {
		baseSchema.MultipleOf = overridingSchema.MultipleOf
	}

	if overridingSchema.Not != nil {
		if deepMerge {
			var err error
			baseSchema.Not, err = deepMergeJSONSchema(ctx, baseSchema.Not, overridingSchema.Not, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		} else {
			baseSchema.Not = overridingSchema.Not
		}
	}

	if overridingSchema.Nullable != nil {
		baseSchema.Nullable = overridingSchema.Nullable
	}

	if len(overridingSchema.OneOf) > 0 {
		baseSchema.OneOf = append(baseSchema.OneOf, overridingSchema.OneOf...)
	}

	if overridingSchema.Pattern != nil {
		baseSchema.Pattern = overridingSchema.Pattern
	}

	if overridingSchema.PatternProperties != nil {
		if baseSchema.PatternProperties == nil {
			baseSchema.PatternProperties = sequencedmap.New[string, *oas3.JSONSchema[oas3.Referenceable]]()
		}

		for k, v := range overridingSchema.PatternProperties.All() {
			if deepMerge {
				pp, err := deepMergeJSONSchema(ctx, baseSchema.PatternProperties.GetOrZero(k), v, mergeDescriptiveFields, e, docInfo)
				if err != nil {
					return err
				}
				baseSchema.PatternProperties.Set(k, pp)
			} else {
				baseSchema.PatternProperties.Set(k, v)
			}
		}
	}

	if overridingSchema.Properties != nil {
		if baseSchema.Properties == nil {
			baseSchema.Properties = sequencedmap.New[string, *oas3.JSONSchema[oas3.Referenceable]]()
		}

		for k, v := range overridingSchema.Properties.All() {
			if deepMerge {
				p, err := deepMergeJSONSchema(ctx, baseSchema.Properties.GetOrZero(k), v, mergeDescriptiveFields, e, docInfo)
				if err != nil {
					return err
				}
				baseSchema.Properties.Set(k, p)
			} else {
				baseSchema.Properties.Set(k, v)
			}
		}
	}

	if overridingSchema.ReadOnly != nil {
		baseSchema.ReadOnly = overridingSchema.ReadOnly
	}

	if len(overridingSchema.Required) > 0 {
		for _, r := range overridingSchema.Required {
			if !slices.Contains(baseSchema.Required, r) {
				baseSchema.Required = append(baseSchema.Required, r)
			}
		}
	}

	if len(overridingSchema.GetType()) > 0 {
		baseTypes := baseSchema.GetType()

		for _, t := range overridingSchema.GetType() {
			if !slices.Contains(baseTypes, t) {
				baseTypes = append(baseTypes, t)
			}
		}

		baseSchema.Type = oas3.NewTypeFromArray(baseTypes)
	}

	if overridingSchema.UnevaluatedItems != nil {
		if deepMerge {
			var err error
			baseSchema.UnevaluatedItems, err = deepMergeJSONSchema(ctx, baseSchema.UnevaluatedItems, overridingSchema.UnevaluatedItems, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		} else {
			baseSchema.UnevaluatedItems = overridingSchema.UnevaluatedItems
		}
	}

	if overridingSchema.UnevaluatedProperties != nil {
		if !deepMerge {
			baseSchema.UnevaluatedProperties = overridingSchema.UnevaluatedProperties
		} else {
			var err error
			baseSchema.UnevaluatedProperties, err = deepMergeJSONSchema(ctx, baseSchema.UnevaluatedProperties, overridingSchema.UnevaluatedProperties, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		}
	}

	if overridingSchema.UniqueItems != nil {
		baseSchema.UniqueItems = overridingSchema.UniqueItems
	}

	if overridingSchema.WriteOnly != nil {
		baseSchema.WriteOnly = overridingSchema.WriteOnly
	}

	if overridingSchema.XML != nil {
		baseSchema.XML = overridingSchema.XML
	}

	// Descriptive fields
	if mergeDescriptiveFields {
		if overridingSchema.Deprecated != nil {
			baseSchema.Deprecated = overridingSchema.Deprecated
		}

		if overridingSchema.Description != nil {
			baseSchema.Description = overridingSchema.Description
		}

		if overridingSchema.Example != nil {
			baseSchema.Example = overridingSchema.Example
		}

		if len(overridingSchema.Examples) > 0 {
			baseSchema.Examples = append(baseSchema.Examples, overridingSchema.Examples...)
		}

		if overridingSchema.ExternalDocs != nil {
			baseSchema.ExternalDocs = overridingSchema.ExternalDocs
		}

		if overridingSchema.Title != nil {
			baseSchema.Title = overridingSchema.Title
		}
	}

	return nil
}

// deepMergeInto performs a deep merge of schema properties, including recursive merging
// of Properties, PatternProperties, AdditionalProperties where schema proxies are resolved
// and deep merged. References are resolved before merging and cleared after merging.
func deepMergeInto(ctx context.Context, baseSchema, overridingSchema *oas3.Schema, mergeDescriptiveFields bool, e speakeasyExtensions, docInfo *document.DocumentInfo) error {
	if overridingSchema == nil {
		return nil
	}

	// Type modifying fields - deep merge for containers, override for scalars

	// Arrays should be appended (deep merge requirement)
	if len(overridingSchema.AllOf) > 0 {
		baseSchema.AllOf = append(baseSchema.AllOf, overridingSchema.AllOf...)
	}

	if overridingSchema.Anchor != nil {
		baseSchema.Anchor = overridingSchema.Anchor
	}

	// Arrays should be appended
	if len(overridingSchema.AnyOf) > 0 {
		baseSchema.AnyOf = append(baseSchema.AnyOf, overridingSchema.AnyOf...)
	}

	if overridingSchema.Const != nil {
		baseSchema.Const = overridingSchema.Const
	}

	if overridingSchema.Contains != nil {
		baseSchema.Contains = overridingSchema.Contains
	}

	if overridingSchema.Default != nil {
		baseSchema.Default = overridingSchema.Default
	}

	if overridingSchema.Discriminator != nil {
		baseSchema.Discriminator = overridingSchema.Discriminator
	}

	// Arrays should be appended
	if len(overridingSchema.Enum) > 0 {
		baseSchema.Enum = append(baseSchema.Enum, overridingSchema.Enum...)
	}

	if overridingSchema.ExclusiveMaximum != nil {
		baseSchema.ExclusiveMaximum = overridingSchema.ExclusiveMaximum
	}

	if overridingSchema.ExclusiveMinimum != nil {
		baseSchema.ExclusiveMinimum = overridingSchema.ExclusiveMinimum
	}

	// Maps should be deeply merged
	if overridingSchema.Extensions.Len() > 0 {
		if baseSchema.Extensions == nil {
			baseSchema.Extensions = extensions.New()
		}

		for extension, node := range overridingSchema.Extensions.All() {
			// There are certain extensions that are mergable
			if e.IsExtensionMergable(extension) {
				baseSchema.Extensions.Set(extension, node)
			}
		}
	}

	if overridingSchema.Format != nil {
		baseSchema.Format = overridingSchema.Format
	}

	if overridingSchema.Items != nil {
		baseSchema.Items = overridingSchema.Items
	}

	if overridingSchema.UnevaluatedItems != nil {
		newUnevaluatedItems := overridingSchema.UnevaluatedItems
		if baseSchema.UnevaluatedItems != nil {
			var err error
			newUnevaluatedItems, err = deepMergeJSONSchema(ctx, baseSchema.UnevaluatedItems, overridingSchema.UnevaluatedItems, mergeDescriptiveFields, e, docInfo)
			if err != nil {
				return err
			}
		}
		baseSchema.UnevaluatedItems = newUnevaluatedItems
	}

	if overridingSchema.MaxContains != nil {
		baseSchema.MaxContains = overridingSchema.MaxContains
	}

	if overridingSchema.Maximum != nil {
		baseSchema.Maximum = overridingSchema.Maximum
	}

	if overridingSchema.MaxItems != nil {
		baseSchema.MaxItems = overridingSchema.MaxItems
	}

	if overridingSchema.MaxLength != nil {
		baseSchema.MaxLength = overridingSchema.MaxLength
	}

	if overridingSchema.MaxProperties != nil {
		baseSchema.MaxProperties = overridingSchema.MaxProperties
	}

	if overridingSchema.MinContains != nil {
		baseSchema.MinContains = overridingSchema.MinContains
	}

	if overridingSchema.Minimum != nil {
		baseSchema.Minimum = overridingSchema.Minimum
	}

	if overridingSchema.MinItems != nil {
		baseSchema.MinItems = overridingSchema.MinItems
	}

	if overridingSchema.MinLength != nil {
		baseSchema.MinLength = overridingSchema.MinLength
	}

	if overridingSchema.MinProperties != nil {
		baseSchema.MinProperties = overridingSchema.MinProperties
	}

	if overridingSchema.MultipleOf != nil {
		baseSchema.MultipleOf = overridingSchema.MultipleOf
	}

	if overridingSchema.Not != nil {
		baseSchema.Not = overridingSchema.Not
	}

	if overridingSchema.Nullable != nil {
		baseSchema.Nullable = overridingSchema.Nullable
	}

	// Arrays should be appended
	if len(overridingSchema.OneOf) > 0 {
		baseSchema.OneOf = append(baseSchema.OneOf, overridingSchema.OneOf...)
	}

	if overridingSchema.Pattern != nil {
		baseSchema.Pattern = overridingSchema.Pattern
	}

	// Deep merge Properties - resolve schemas and merge recursively
	if overridingSchema.Properties != nil {
		if baseSchema.Properties == nil {
			baseSchema.Properties = sequencedmap.New[string, *oas3.JSONSchema[oas3.Referenceable]]()
		}

		for k, overridingSchema := range overridingSchema.Properties.All() {
			if baseProp, exists := baseSchema.Properties.Get(k); exists {
				// Both base and overriding have this property - deep merge
				var err error
				overridingSchema, err = deepMergeJSONSchema(ctx, baseProp, overridingSchema, mergeDescriptiveFields, e, docInfo)
				if err != nil {
					return err
				}
			}
			baseSchema.Properties.Set(k, overridingSchema)
		}
	}

	// Deep merge PatternProperties - resolve schemas and merge recursively
	if overridingSchema.PatternProperties != nil {
		if baseSchema.PatternProperties == nil {
			baseSchema.PatternProperties = sequencedmap.New[string, *oas3.JSONSchema[oas3.Referenceable]]()
		}

		for k, overridingSchema := range overridingSchema.PatternProperties.All() {
			if baseProp, exists := baseSchema.PatternProperties.Get(k); exists {
				// Both base and overriding have this pattern property - deep merge
				var err error
				overridingSchema, err = deepMergeJSONSchema(ctx, baseProp, overridingSchema, mergeDescriptiveFields, e, docInfo)
				if err != nil {
					return err
				}
			}
			baseSchema.PatternProperties.Set(k, overridingSchema)
		}
	}

	// Deep merge AdditionalProperties if both are schemas
	if overridingSchema.AdditionalProperties != nil {
		mergedSchema, err := deepMergeJSONSchema(ctx, baseSchema.AdditionalProperties, overridingSchema.AdditionalProperties, mergeDescriptiveFields, e, docInfo)
		if err != nil {
			return err
		}
		baseSchema.AdditionalProperties = mergedSchema
	}

	if overridingSchema.UnevaluatedProperties != nil {
		mergedSchema, err := deepMergeJSONSchema(ctx, baseSchema.UnevaluatedProperties, overridingSchema.UnevaluatedProperties, mergeDescriptiveFields, e, docInfo)
		if err != nil {
			return err
		}
		baseSchema.UnevaluatedProperties = mergedSchema
	}

	if overridingSchema.ReadOnly != nil {
		baseSchema.ReadOnly = overridingSchema.ReadOnly
	}

	// Arrays should be appended, but avoid duplicates
	for _, r := range overridingSchema.Required {
		if !slices.Contains(baseSchema.Required, r) {
			baseSchema.Required = append(baseSchema.Required, r)
		}
	}

	// Arrays should be appended, but avoid duplicates
	for _, t := range overridingSchema.GetType() {
		if !slices.Contains(baseSchema.GetType(), t) {
			if baseSchema.Type.IsLeft() {
				*baseSchema.Type.Left = append(*baseSchema.Type.Left, t)
			} else {
				baseSchema.Type = oas3.NewTypeFromArray(append([]oas3.SchemaType{*baseSchema.Type.Right}, t))
			}
		}
	}

	if overridingSchema.UniqueItems != nil {
		baseSchema.UniqueItems = overridingSchema.UniqueItems
	}

	if overridingSchema.WriteOnly != nil {
		baseSchema.WriteOnly = overridingSchema.WriteOnly
	}

	if overridingSchema.XML != nil {
		baseSchema.XML = overridingSchema.XML
	}

	// Descriptive fields - same as original merge
	if mergeDescriptiveFields {
		if overridingSchema.Deprecated != nil {
			baseSchema.Deprecated = overridingSchema.Deprecated
		}

		if overridingSchema.Description != nil {
			baseSchema.Description = overridingSchema.Description
		}

		if overridingSchema.Example != nil {
			baseSchema.Example = overridingSchema.Example
		}

		// Arrays should be appended
		if len(overridingSchema.Examples) > 0 {
			baseSchema.Examples = append(baseSchema.Examples, overridingSchema.Examples...)
		}

		if overridingSchema.ExternalDocs != nil {
			baseSchema.ExternalDocs = overridingSchema.ExternalDocs
		}

		if overridingSchema.Title != nil {
			baseSchema.Title = overridingSchema.Title
		}
	}

	return nil
}

// deepMergeSchemaProxies resolves two json schemas, deep merges them,
// clears references, and returns a new json schema with the merged result
func deepMergeJSONSchema(ctx context.Context, baseSchema, overridingSchema *oas3.JSONSchema[oas3.Referenceable], mergeDescriptiveFields bool, e speakeasyExtensions, docInfo *document.DocumentInfo) (*oas3.JSONSchema[oas3.Referenceable], error) {
	if baseSchema == nil {
		return overridingSchema, nil
	}

	if overridingSchema == nil {
		return baseSchema, nil
	}

	// Resolve references to get actual schemas
	resolvedBase, err := resolution.Resolve(ctx, baseSchema, docInfo)
	if err != nil {
		return nil, err
	}

	resolvedOverriding, err := resolution.Resolve(ctx, overridingSchema, docInfo)
	if err != nil {
		return nil, err
	}

	if resolvedBase == nil {
		return overridingSchema, nil
	}

	if resolvedOverriding == nil {
		return baseSchema, nil
	}

	if resolvedBase.IsBool() || resolvedOverriding.IsBool() {
		return overridingSchema, nil
	}

	base := resolvedBase.GetSchema()

	// Deep merge the overriding schema into the base
	err = deepMergeInto(ctx, base, resolvedOverriding.GetSchema(), mergeDescriptiveFields, e, docInfo)
	if err != nil {
		return nil, err
	}

	// Create a new schema proxy with the merged schema
	return oas3.NewJSONSchemaFromSchema[oas3.Referenceable](base), nil
}
