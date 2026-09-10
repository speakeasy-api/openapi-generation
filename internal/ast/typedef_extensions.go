package ast

import (
	"fmt"
	"maps"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

// TerraformHoistedSource describes the path to a source field within a oneOf
// associated type that was hoisted to the parent level.
type TerraformHoistedSource struct {
	// AssociatedTypeName is the name of the oneOf associated type containing the source field.
	AssociatedTypeName string `yaml:"associatedTypeName" json:"associatedTypeName"`
	// FieldName is the name of the source field within the associated type.
	FieldName string `yaml:"fieldName" json:"fieldName"`
	// PathPrefix is the path from root to the parent oneOf, used to build absolute paths
	// in plan modifiers. For nested oneOf unions, this contains all intermediate path segments.
	PathPrefix []string `yaml:"pathPrefix,omitempty" json:"pathPrefix,omitempty"`
}

// Describes extensions for a TypeDef.
type TypeDefExtensions struct {
	// Describes x-speakeasy-additional-properties-name extension configuration.
	// TODO: Migrate usage to this field.
	// AdditionalPropertiesName *string `yaml:",omitempty"`

	// All extensions, including those not represented in other fields. Ideally,
	// any TypeDef extensions should be fully validated and typed for templating
	// rather than relying on this catch-all.
	All map[string]any `yaml:",omitempty"`

	// Describes x-speakeasy-allow-empty-value extension configuration.
	AllowEmptyValue bool `yaml:",omitempty"`

	// Describes x-speakeasy-base64-input-mode extension configuration.
	// Set on request-side string schemas with format:byte or contentEncoding:base64.
	Base64InputMode string `yaml:",omitempty"`

	// Describes x-speakeasy-conflicts-with extension configuration.
	// TODO: Migrate usage to this field.
	// ConflictsWith *extensions.ConflictsWith `yaml:",omitempty"`

	// Describes x-speakeasy-deprecation-message extension configuration.
	// TODO: Migrate usage to this field.
	// DeprecationMessage *string `yaml:",omitempty"`

	// Describes x-speakeasy-error-message extension configuration.
	// TODO: Migrate usage to this field.
	// ErrorMessage *bool `yaml:",omitempty"`

	// Describes x-speakeasy-entity extension configuration.
	Entity *extensions.Entity `yaml:",omitempty"`

	// Describes x-speakeasy-entity-description extension configuration.
	EntityDescription *extensions.EntityDescription `yaml:",omitempty"`

	// Describes x-speakeasy-entity-version extension configuration.
	EntityVersion *extensions.EntityVersion `yaml:",omitempty"`

	// Describes x-speakeasy-enums extension configuration.
	// TODO: Migrate usage to this field.
	// Enums any `yaml:",omitempty"`

	// Describes x-speakeasy-enum-format extension configuration.
	// TODO: Migrate usage to this field.
	// EnumFormat *string `yaml:",omitempty"`

	// Describes x-speakeasy-example-unset extension configuration.
	ExampleUnset *bool `yaml:",omitempty"`

	// Describes x-speakeasy-globals-hidden extension configuration.
	// TODO: Migrate usage to this field.
	// GlobalsHidden *bool `yaml:",omitempty"`

	// Describes x-speakeasy-ignore extension configuration.
	Ignore *bool `yaml:",omitempty"`

	// Describes x-speakeasy-include extension configuration.
	// TODO: Migrate usage to this field.
	// Include *bool `yaml:",omitempty"`

	// Describes x-speakeasy-match extension configuration.
	//
	// NOTE: Awkward naming to avoid conflict with existing Match method.
	MatchConfig *extensions.MatchConfig `yaml:",omitempty"`

	// Describes x-speakeasy-model-namespace extension configuration.
	// Determines the namespace folder where the type will be generated.
	// This enables multiple components with the same name to exist in different namespaces.
	ModelNamespace *string `yaml:",omitempty"`

	// Describes x-speakeasy-name-override extension configuration.
	// TODO: Migrate usage to this field.
	// NameOverride *string `yaml:",omitempty"`

	// Describes x-speakeasy-overridable-scopes extension configuration.
	OverridableOAuth2Scopes bool `yaml:",omitempty"`

	// Describes x-speakeasy-pagination extension configuration.
	//
	// NOTE: This is not parsed directly for TypeDef, but instead added to the
	// response TypeDef when configured on the containing operation.
	Pagination *extensions.Pagination `yaml:",omitempty"`

	// Describes x-speakeasy-param-computed extension configuration.
	// TODO: Migrate usage to this field.
	// ParamComputed *bool `yaml:",omitempty"`

	// Describes x-speakeasy-param-force-new extension configuration.
	// TODO: Migrate usage to this field.
	// ParamForceNew *bool `yaml:",omitempty"`

	// Describes x-speakeasy-param-optional extension configuration.
	// TODO: Migrate usage to this field.
	// ParamOptional *bool `yaml:",omitempty"`

	// Describes x-speakeasy-param-read-only extension configuration.
	// TODO: Migrate usage to this field.
	// ParamReadOnly *bool `yaml:",omitempty"`

	// Describes x-speakeasy-param-sensitive extension configuration.
	// TODO: Migrate usage to this field.
	// ParamSensitive *bool `yaml:",omitempty"`

	// Describes x-speakeasy-param-suppress-computed-diff extension
	// configuration.
	// TODO: Migrate usage to this field.
	// ParamSuppressComputedDiff *bool `yaml:",omitempty"`

	// Describes x-speakeasy-pagination extension configuration. Added to the
	// response when configured on the containing operation.
	// TODO: Migrate usage to this field.
	// Pagination *extensions.Pagination `yaml:",omitempty"`

	// Describes x-speakeasy-exports extension configuration.
	// These are language-agnostic public export aliases keyed by SDK group path.
	// Export declarations are source-site specific and are not merged into
	// structurally similar schemas.
	PublicExports []extensions.PublicExport `yaml:",omitempty"`

	// Describes x-speakeasy-plan-modifiers extension configuration.
	// TODO: Migrate usage to this field.
	// PlanModifiers *extensions.PlanModifiers `yaml:",omitempty"`

	// Describes x-speakeasy-plan-validators extension configuration.
	// TODO: Migrate usage to this field.
	// PlanValidators *extensions.PlanValidators `yaml:",omitempty"`

	// Describes x-speakeasy-required-with extension configuration.
	// TODO: Migrate usage to this field.
	// RequiredWith *extensions.RequiredWith `yaml:",omitempty"`

	// Describes x-speakeasy-soft-delete-property extension configuration.
	// TODO: Migrate usage to this field.
	// SoftDeleteProperty *string `yaml:",omitempty"`

	// Describes x-speakeasy-terraform-alias-to extension configuration.
	TerraformAliasTo *string `yaml:",omitempty"`

	// TerraformHoistedFrom tracks source associated types for hoisted oneOf fields.
	// When set, generates UseHoistedValue plan modifier to prevent false drift.
	// This is set during template processing, not from OAS extensions.
	//
	// This is an array because a hoisted field can originate from multiple oneOf
	// variants. For example, if a oneOf has variants A, B, and C, and all three
	// have a common "description" field that gets hoisted to the parent, the
	// plan modifier needs to check all three variants to find the active one
	// and copy its value to the hoisted field.
	TerraformHoistedFrom []TerraformHoistedSource `yaml:",omitempty"`

	// Describes x-speakeasy-terraform-custom-default extension configuration.
	TerraformCustomDefault *extensions.TerraformCustomDefault `yaml:",omitempty"`

	// Describes x-speakeasy-terraform-custom-type extension configuration.
	// TODO: Migrate usage to this field.
	// TerraformCustomType *extensions.TerraformCustomType `yaml:",omitempty"`

	// Describes x-speakeasy-terraform-ignore extension configuration.
	TerraformIgnore *extensions.TerraformIgnore `yaml:",omitempty"`

	// Describes x-speakeasy-terraform-plan-only extension configuration.
	// TODO: Migrate usage to this field.
	// TerraformPlanOnly *bool `yaml:",omitempty"`

	// Describes x-speakeasy-response-filter extension configuration.
	ResponseFilter *bool `yaml:",omitempty"`

	// Describes x-speakeasy-terraform-write-only extension configuration.
	TerraformWriteOnly *bool `yaml:",omitempty"`

	// Describes x-speakeasy-token-endpoint-authentication extension
	// configuration.
	// TODO: Migrate usage to this field.
	// TokenEndpointAuthentication *string `yaml:",omitempty"`

	// Describes x-speakeasy-transform-from-api extension configuration.
	TransformFromAPI *extensions.TransformerConfig `yaml:",omitempty"`

	// Describes x-speakeasy-transform-to-api extension configuration.
	TransformToAPI *extensions.TransformerConfig `yaml:",omitempty"`

	// Describes x-speakeasy-type-override extension configuration.
	// TODO: Migrate usage to this field.
	// TypeOverride *string `yaml:",omitempty"`

	// Describes x-speakeasy-unknown-values extension configuration.
	// TODO: Migrate usage to this field.
	// UnknownValues *string `yaml:",omitempty"`

	// Describes x-speakeasy-wrapped-attribute extension configuration.
	WrappedAttribute *string `yaml:",omitempty"`

	// Describes x-speakeasy-xor-with extension configuration.
	// TODO: Migrate usage to this field.
	// XorWith *extensions.XorWith `yaml:",omitempty"`
}

// Returns a new TypeDefExtensions instance or any decoding errors.
func NewTypeDefExtensions(e *extensions.Extensions, typeDef *TypeDef, oaExtensions extensions.OAExtensions) (*TypeDefExtensions, error) {
	if e == nil || typeDef == nil {
		return nil, nil
	}

	result := &TypeDefExtensions{
		All: make(map[string]any),
	}

	if oaExtensions.Len() == 0 {
		return result, nil
	}

	for oasExtensionKey, oasExtensionValue := range oaExtensions.All() {
		var decodedValue any

		if err := oasExtensionValue.Decode(&decodedValue); err != nil {
			return nil, fmt.Errorf("failed to decode OAS extension %q: %w", oasExtensionKey, err)
		}

		result.All[oasExtensionKey] = decodedValue

		switch oasExtensionKey {
		case extensions.ExtEntity.Name():
			entity, err := e.HandleEntityExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtEntity.Name(), err)
			}

			result.Entity = entity
		case extensions.ExtEntityDescription.Name():
			entityDescription, err := e.HandleEntityDescriptionExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtEntityDescription.Name(), err)
			}

			result.EntityDescription = entityDescription
		case extensions.ExtEntityVersion.Name():
			entityVersion, err := e.HandleEntityVersionExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtEntityVersion.Name(), err)
			}

			result.EntityVersion = entityVersion
		case extensions.ExtExampleUnset.Name():
			exampleUnset, err := e.HandleExampleUnsetExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtExampleUnset.Name(), err)
			}

			result.ExampleUnset = exampleUnset
		case extensions.ExtIgnore.Name():
			ignore, err := e.HandleIgnoreExtension(oaExtensions)

			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtIgnore.Name(), err)
			}

			result.Ignore = ignore
		case extensions.ExtMatch.Name():
			match, err := e.HandleMatchExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtMatch.Name(), err)
			}

			result.MatchConfig = match
		case extensions.ExtModelNamespace.Name():
			modelNamespace, err := e.GetModelNamespace(oaExtensions)

			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtModelNamespace.Name(), err)
			}

			if modelNamespace != "" {
				result.ModelNamespace = &modelNamespace
			}
		case extensions.ExtPublicExports.Name():
			publicExports, err := e.HandlePublicExportsExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtPublicExports.Name(), err)
			}

			result.PublicExports = publicExports
		case extensions.ExtTerraformAliasTo.Name():
			terraformAliasTo, err := e.HandleTerraformAliasToExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtTerraformAliasTo.Name(), err)
			}

			result.TerraformAliasTo = terraformAliasTo
		case extensions.ExtTerraformCustomDefault.Name():
			terraformCustomDefault, err := e.HandleTerraformCustomDefaultExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtTerraformCustomDefault.Name(), err)
			}

			result.TerraformCustomDefault = terraformCustomDefault
		case extensions.ExtTerraformIgnore.Name():
			terraformIgnore, err := e.HandleTerraformIgnoreExtension(oaExtensions)

			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtTerraformIgnore.Name(), err)
			}

			result.TerraformIgnore = terraformIgnore
		case extensions.ExtResponseFilter.Name():
			responseFilter, err := e.HandleResponseFilterExtension(oaExtensions)
			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtResponseFilter.Name(), err)
			}

			result.ResponseFilter = responseFilter
		case extensions.ExtTerraformWriteOnly.Name():
			terraformWriteOnly, err := e.HandleTerraformWriteOnlyExtension(oaExtensions)

			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtTerraformWriteOnly.Name(), err)
			}

			result.TerraformWriteOnly = terraformWriteOnly
		case extensions.ExtTransformFromAPI.Name():
			transformExtension, err := e.HandleTransformExtension(oaExtensions, extensions.ExtTransformFromAPI)

			if err != nil {
				return nil, err
			}
			result.TransformFromAPI = transformExtension
		case extensions.ExtTransformToAPI.Name():
			transformExtension, err := e.HandleTransformExtension(oaExtensions, extensions.ExtTransformToAPI)

			if err != nil {
				return nil, err
			}
			result.TransformToAPI = transformExtension
		case extensions.ExtWrappedAttribute.Name():
			wrappedAttribute, err := e.HandleWrappedAttributeExtension(oaExtensions)

			if err != nil {
				return nil, fmt.Errorf("failed to handle %q extension: %w", extensions.ExtWrappedAttribute.Name(), err)
			}

			result.WrappedAttribute = wrappedAttribute
		}
	}

	return result, nil
}

// Clone creates a deep copy of the TypeDefExtensions
func (e *TypeDefExtensions) Clone() *TypeDefExtensions {
	if e == nil {
		return nil
	}

	// Clone TerraformHoistedFrom slice with deep copy of PathPrefix
	var hoistedFrom []TerraformHoistedSource
	if e.TerraformHoistedFrom != nil {
		hoistedFrom = make([]TerraformHoistedSource, len(e.TerraformHoistedFrom))
		for i, src := range e.TerraformHoistedFrom {
			hoistedFrom[i] = TerraformHoistedSource{
				AssociatedTypeName: src.AssociatedTypeName,
				FieldName:          src.FieldName,
				PathPrefix:         append([]string(nil), src.PathPrefix...),
			}
		}
	}

	return &TypeDefExtensions{
		All:                     maps.Clone(e.All),
		AllowEmptyValue:         e.AllowEmptyValue,
		Base64InputMode:         e.Base64InputMode,
		Entity:                  e.Entity.Clone(),
		EntityDescription:       e.EntityDescription.Clone(),
		EntityVersion:           e.EntityVersion.Clone(),
		ExampleUnset:            clonePtr(e.ExampleUnset),
		Ignore:                  clonePtr(e.Ignore),
		MatchConfig:             clonePtr(e.MatchConfig),
		ModelNamespace:          clonePtr(e.ModelNamespace),
		OverridableOAuth2Scopes: e.OverridableOAuth2Scopes,
		Pagination:              e.Pagination.Clone(),
		PublicExports:           slices.Clone(e.PublicExports),
		TerraformAliasTo:        clonePtr(e.TerraformAliasTo),
		TerraformCustomDefault:  e.TerraformCustomDefault.Clone(),
		TerraformHoistedFrom:    hoistedFrom,
		ResponseFilter:          clonePtr(e.ResponseFilter),
		TerraformIgnore:         e.TerraformIgnore.Clone(),
		TerraformWriteOnly:      clonePtr(e.TerraformWriteOnly),
		WrappedAttribute:        clonePtr(e.WrappedAttribute),
	}
}

// Get returns the extension value and whether it exists in All.
func (e *TypeDefExtensions) Get(name string) (any, bool) {
	if e.All == nil {
		return nil, false
	}

	v, ok := e.All[name]

	return v, ok
}

// Has returns true if the extension exists in All.
func (e *TypeDefExtensions) Has(name string) bool {
	if e.All == nil {
		return false
	}

	_, ok := e.All[name]

	return ok
}

// Remove deletes the extension from All.
func (e *TypeDefExtensions) Remove(name string) {
	if e.All == nil {
		return
	}

	delete(e.All, name)
}

// Rename moves the extension value from one key to another in All,
// only when the value is truthy (not false or nil). Falsy values
// serve as negative markers that must remain under the original key.
func (e *TypeDefExtensions) Rename(from, to string) {
	if e.All == nil {
		return
	}

	value, exists := e.All[from]

	if !exists || value == nil || value == false {
		return
	}

	e.All[to] = value

	delete(e.All, from)
}

// Set unconditionally sets the extension value in All.
func (e *TypeDefExtensions) Set(name string, value any) {
	if e.All == nil {
		e.All = make(map[string]any)
	}

	e.All[name] = value
}

// SetIfAbsent sets the extension value in All only if the key does
// not already exist.
func (e *TypeDefExtensions) SetIfAbsent(name string, value any) {
	if e.All == nil {
		e.All = make(map[string]any)
	}

	if _, exists := e.All[name]; !exists {
		e.All[name] = value
	}
}

func (o *TypeDefExtensions) Match(matchers Matchers) error {
	if matchers.TypeDefExtensions != nil {
		return matchers.TypeDefExtensions(o)
	}

	return nil
}

// Merges the given TypeDefExtensions into this TypeDefExtensions. The algorithm
// adds data from the given TypeDefExtensions to this TypeDefExtensions where it
// is undefined. Where there is conflicting data, this TypeDefExtensions's data
// is preserved. It does not remove any data from this TypeDefExtensions.
func (e *TypeDefExtensions) Merge(other *TypeDefExtensions) {
	if e == nil || other == nil {
		return
	}

	for name, ext := range other.All {
		if _, ok := e.All[name]; !ok {
			e.All[name] = ext
		}
	}

	if !e.AllowEmptyValue {
		e.AllowEmptyValue = other.AllowEmptyValue
	}

	if e.Base64InputMode == "" {
		e.Base64InputMode = other.Base64InputMode
	}

	if e.Entity == nil {
		e.Entity = other.Entity
	} else {
		e.Entity.Merge(other.Entity)
	}

	if e.EntityDescription == nil {
		e.EntityDescription = other.EntityDescription
	} else {
		e.EntityDescription.Merge(other.EntityDescription)
	}

	if e.EntityVersion == nil {
		e.EntityVersion = other.EntityVersion
	} else {
		e.EntityVersion.Merge(other.EntityVersion)
	}

	if e.ExampleUnset == nil {
		e.ExampleUnset = other.ExampleUnset
	}

	if e.Ignore == nil {
		e.Ignore = other.Ignore
	}

	if e.MatchConfig == nil {
		e.MatchConfig = other.MatchConfig
	}

	if e.ModelNamespace == nil {
		e.ModelNamespace = other.ModelNamespace
	}

	if !e.OverridableOAuth2Scopes {
		e.OverridableOAuth2Scopes = other.OverridableOAuth2Scopes
	}

	if e.Pagination == nil {
		e.Pagination = other.Pagination
	}

	if e.TerraformAliasTo == nil {
		e.TerraformAliasTo = other.TerraformAliasTo
	}

	if e.TerraformCustomDefault == nil {
		e.TerraformCustomDefault = other.TerraformCustomDefault
	}

	if e.TerraformHoistedFrom == nil {
		e.TerraformHoistedFrom = other.TerraformHoistedFrom
	}

	if e.ResponseFilter == nil {
		e.ResponseFilter = other.ResponseFilter
	}

	if e.TerraformIgnore == nil {
		e.TerraformIgnore = other.TerraformIgnore
	}

	if e.TerraformWriteOnly == nil {
		e.TerraformWriteOnly = other.TerraformWriteOnly
	}

	if e.WrappedAttribute == nil {
		e.WrappedAttribute = other.WrappedAttribute
	}
}

// Merges the given TypeDefExtensions into this TypeDefExtensions for Terraform
// usage. The algorithm first calls Merge for base "preserve existing"
// semantics, then overrides specific fields where the other TypeDefExtensions
// should take precedence:
//
//   - Ignore: other wins
//   - TerraformIgnore: OR-merges DataModel and Schema
//   - TerraformWriteOnly: other wins
//   - TerraformHoistedFrom: other wins
func (e *TypeDefExtensions) TerraformMerge(other *TypeDefExtensions) {
	if e == nil || other == nil {
		return
	}

	e.Merge(other)

	if other.Ignore != nil {
		e.Ignore = other.Ignore
	}

	if other.TerraformIgnore != nil {
		existingTI := e.TerraformIgnore
		e.TerraformIgnore = &extensions.TerraformIgnore{
			DataModel: (existingTI != nil && existingTI.DataModel) || other.TerraformIgnore.DataModel,
			Schema:    (existingTI != nil && existingTI.Schema) || other.TerraformIgnore.Schema,
		}
	}

	if other.ResponseFilter != nil {
		e.ResponseFilter = other.ResponseFilter
	}

	if other.TerraformWriteOnly != nil {
		e.TerraformWriteOnly = other.TerraformWriteOnly
	}

	if other.TerraformHoistedFrom != nil {
		e.TerraformHoistedFrom = other.TerraformHoistedFrom
	}
}

// Merges extensions from the given OAExtensions into the TypeDefExtensions,
// without overwriting any existing values.
func (e *TypeDefExtensions) MergeWithoutOverwrite(extensions *extensions.Extensions, typeDef *TypeDef, oaExtensions extensions.OAExtensions) error {
	if e == nil || extensions == nil || typeDef == nil || oaExtensions.Len() == 0 {
		return nil
	}

	newTypeDefExtensions, err := NewTypeDefExtensions(extensions, typeDef, oaExtensions)
	if err != nil {
		return err
	}

	for name, ext := range newTypeDefExtensions.All {
		if _, ok := typeDef.Extensions.All[name]; !ok {
			e.All[name] = ext
		}
	}

	if e.Base64InputMode == "" {
		e.Base64InputMode = newTypeDefExtensions.Base64InputMode
	}

	if e.Entity == nil {
		e.Entity = newTypeDefExtensions.Entity
	}

	if e.EntityDescription == nil {
		e.EntityDescription = newTypeDefExtensions.EntityDescription
	}

	if e.EntityVersion == nil {
		e.EntityVersion = newTypeDefExtensions.EntityVersion
	}

	if e.ExampleUnset == nil {
		e.ExampleUnset = newTypeDefExtensions.ExampleUnset
	}

	if e.Ignore == nil {
		e.Ignore = newTypeDefExtensions.Ignore
	}

	if e.MatchConfig == nil {
		e.MatchConfig = newTypeDefExtensions.MatchConfig
	}

	if e.ModelNamespace == nil {
		e.ModelNamespace = newTypeDefExtensions.ModelNamespace
	}

	if e.Pagination == nil {
		e.Pagination = newTypeDefExtensions.Pagination
	}

	if e.PublicExports == nil {
		e.PublicExports = slices.Clone(newTypeDefExtensions.PublicExports)
	}

	if e.TerraformAliasTo == nil {
		e.TerraformAliasTo = newTypeDefExtensions.TerraformAliasTo
	}

	if e.TerraformCustomDefault == nil {
		e.TerraformCustomDefault = newTypeDefExtensions.TerraformCustomDefault
	}

	if e.TerraformHoistedFrom == nil {
		e.TerraformHoistedFrom = newTypeDefExtensions.TerraformHoistedFrom
	}

	if e.ResponseFilter == nil {
		e.ResponseFilter = newTypeDefExtensions.ResponseFilter
	}

	if e.TerraformIgnore == nil {
		e.TerraformIgnore = newTypeDefExtensions.TerraformIgnore
	}

	if e.TerraformWriteOnly == nil {
		e.TerraformWriteOnly = newTypeDefExtensions.TerraformWriteOnly
	}

	if e.WrappedAttribute == nil {
		e.WrappedAttribute = newTypeDefExtensions.WrappedAttribute
	}

	return nil
}
