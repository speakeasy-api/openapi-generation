package ast

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"

	generatorErrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/jsonpointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"gopkg.in/yaml.v3"
)

const (
	ErrTypeMismatch               = generatorErrors.Error("type mismatch")
	ErrEnumMismatch               = generatorErrors.Error("enum mismatch")
	ErrFieldMismatch              = generatorErrors.Error("field mismatch")
	ErrAnnotationMismatch         = generatorErrors.Error("annotation mismatch")
	ErrItemTypeMismatch           = generatorErrors.Error("item type mismatch")
	ErrScopeMismatch              = generatorErrors.Error("scope mismatch")
	ErrUnionDiscriminatorMismatch = generatorErrors.Error("union discriminator mismatch")
	ErrUnionTypeMismatch          = generatorErrors.Error("union type mismatch")
)

// DataType is the type of a type definition
type DataType string

const (
	DataTypeString         DataType = "string"
	DataTypeInteger        DataType = "integer"
	DataTypeInt32          DataType = "int32"
	DataTypeBigInt         DataType = "bigint"
	DataTypeNumber         DataType = "number"
	DataTypeFloat32        DataType = "float32"
	DataTypeDecimal        DataType = "decimal"
	DataTypeBoolean        DataType = "boolean"
	DataTypeDate           DataType = "date"
	DataTypeDateTime       DataType = "date-time"
	DataTypeUUID           DataType = "uuid"
	DataTypeDuration       DataType = "duration"
	DataTypeMap            DataType = "map"
	DataTypeArray          DataType = "array"
	DataTypeSet            DataType = "set"
	DataTypeAny            DataType = "any"
	DataTypeBytes          DataType = "bytes"
	DataTypeClass          DataType = "class"
	DataTypeEnum           DataType = "enum"
	DataTypeResponse       DataType = "response"
	DataTypeRequest        DataType = "request"
	DataTypeUnion          DataType = "union"
	DataTypeError          DataType = "error"
	DataTypeRequestStream  DataType = "request-stream"
	DataTypeResponseStream DataType = "response-stream"
	DataTypeEventStream    DataType = "event-stream"
	DataTypeJsonL          DataType = "jsonl"
)

const (
	InternalTypeRequest  = "request"
	InternalTypeResponse = "response"
)

type DiscriminatorMapping struct {
	Name        string   `yaml:",omitempty"`
	DisplayName string   `yaml:",omitempty"`
	Type        *TypeDef `yaml:",omitempty"`
}

// Clone creates a deep copy of the DiscriminatorMapping
func (d *DiscriminatorMapping) Clone() *DiscriminatorMapping {
	if d == nil {
		return nil
	}

	return &DiscriminatorMapping{
		Name:        d.Name,
		DisplayName: d.DisplayName,
		Type:        d.Type.Clone(),
	}
}

func (d *DiscriminatorMapping) Match(matchers Matchers) error {
	if matchers.DiscriminatorMapping != nil {
		return matchers.DiscriminatorMapping(d)
	}

	return nil
}

// Collection of DiscriminatorMapping.
type DiscriminatorMappings []*DiscriminatorMapping

func (d DiscriminatorMappings) Clone() DiscriminatorMappings {
	if d == nil {
		return nil
	}

	cloned := make(DiscriminatorMappings, len(d))

	for i, m := range d {
		cloned[i] = m.Clone()
	}

	return cloned
}

type Discriminator struct {
	TypePropertyName string                `yaml:",omitempty"`
	Mapping          DiscriminatorMappings `yaml:",omitempty"`
	Inferred         bool                  `yaml:",omitempty"`
}

// Clone creates a deep copy of the Discriminator
func (d *Discriminator) Clone() *Discriminator {
	if d == nil {
		return nil
	}

	return &Discriminator{
		Mapping:          d.Mapping.Clone(),
		TypePropertyName: d.TypePropertyName,
		Inferred:         d.Inferred,
	}
}

func (d *Discriminator) Match(matchers Matchers) error {
	if matchers.Discriminator != nil {
		return matchers.Discriminator(d)
	}

	return nil
}

type Enum struct {
	Type         *TypeDef          `yaml:",omitempty"`
	Values       []string          `yaml:",omitempty"`
	Names        []string          `yaml:",omitempty"`
	Open         bool              `yaml:",omitempty"`
	Format       string            `yaml:",omitempty"` // Whether the enum is templated as a native enum or union of literals. If empty use language default
	Descriptions map[string]string `yaml:",omitempty"`
}

// Clone creates a deep copy of the Enum
func (e *Enum) Clone() *Enum {
	if e == nil {
		return nil
	}

	return &Enum{
		Descriptions: maps.Clone(e.Descriptions),
		Format:       e.Format,
		Names:        slices.Clone(e.Names),
		Open:         e.Open,
		Type:         e.Type.Clone(),
		Values:       slices.Clone(e.Values),
	}
}

func (e *Enum) Match(matchers Matchers) error {
	if matchers.Enum != nil {
		return matchers.Enum(e)
	}

	return nil
}

type Validations struct {
	// An array instance is valid against "maxItems" if its size is less than, or equal to, the value of this keyword,
	// defaults to 0
	MinItems *int64 `yaml:",omitempty"`

	// A string instance is valid against this keyword if its length is greater than, or equal to, the value of this keyword.
	// The length of a string instance is defined as the number of its characters as defined by RFC 8259.
	MinLength *int64 `yaml:",omitempty"`

	// If the instance is a number, then this keyword validates only if the instance is greater than or exactly equal to "minimum".
	Minimum *float64 `yaml:",omitempty"`

	// Maximum number of items in an array, no default
	MaxItems *int64 `yaml:",omitempty"`

	// A string instance is valid against this keyword if its length is less than, or equal to, the value of this keyword.
	// The length of a string instance is defined as the number of its characters as defined by RFC 8259.
	MaxLength *int64 `yaml:",omitempty"`

	// If the instance is a number, then this keyword validates only if the instance is less than or exactly equal to "maximum".
	Maximum *float64 `yaml:",omitempty"`

	// A string instance is considered valid if the regular expression matches the instance successfully.
	Pattern *string `yaml:",omitempty"`

	// If this keyword has boolean value false, the instance validates successfully.
	// If it has boolean value true, the instance validates successfully if all of its elements are unique.
	UniqueItems *bool `yaml:",omitempty"`
}

// Clone creates a deep copy of the Validations
func (v *Validations) Clone() *Validations {
	if v == nil {
		return nil
	}

	return &Validations{
		Maximum:     clonePtr(v.Maximum),
		MaxItems:    clonePtr(v.MaxItems),
		MaxLength:   clonePtr(v.MaxLength),
		Minimum:     clonePtr(v.Minimum),
		MinItems:    clonePtr(v.MinItems),
		MinLength:   clonePtr(v.MinLength),
		Pattern:     clonePtr(v.Pattern),
		UniqueItems: clonePtr(v.UniqueItems),
	}
}

func (v *Validations) Match(matchers Matchers) error {
	if matchers.Validations != nil {
		return matchers.Validations(v)
	}

	return nil
}

// Merges the given Validations into this Validations. The algorithm
// adds data from the given Validations to this Validations where it is
// undefined. Where there is conflicting data, the stricter validation value is
// used depending on the validation type. For example with Maximum, the minimum
// of the two values is used. It does not remove any data from this Validations.
//
// Pattern is merged by creating a new pattern that is the alternation of both
// patterns if they differ.
func (v *Validations) Merge(other *Validations) {
	if v == nil || other == nil {
		return
	}

	if v.Maximum == nil {
		v.Maximum = other.Maximum
	} else if other.Maximum != nil && *other.Maximum == math.Min(*v.Maximum, *other.Maximum) {
		v.Maximum = other.Maximum
	}

	if v.MaxItems == nil {
		v.MaxItems = other.MaxItems
	} else if other.MaxItems != nil && *other.MaxItems < *v.MaxItems {
		v.MaxItems = other.MaxItems
	}

	if v.MaxLength == nil {
		v.MaxLength = other.MaxLength
	} else if other.MaxLength != nil && *other.MaxLength < *v.MaxLength {
		v.MaxLength = other.MaxLength
	}

	if v.Minimum == nil {
		v.Minimum = other.Minimum
	} else if other.Minimum != nil && *other.Minimum == math.Max(*v.Minimum, *other.Minimum) {
		v.Minimum = other.Minimum
	}

	if v.MinItems == nil {
		v.MinItems = other.MinItems
	} else if other.MinItems != nil && *other.MinItems > *v.MinItems {
		v.MinItems = other.MinItems
	}

	if v.MinLength == nil {
		v.MinLength = other.MinLength
	} else if other.MinLength != nil && *other.MinLength > *v.MinLength {
		v.MinLength = other.MinLength
	}

	if v.Pattern == nil {
		v.Pattern = other.Pattern
	} else if other.Pattern != nil && *other.Pattern != *v.Pattern {
		combinedPattern := "(" + *v.Pattern + "|" + *other.Pattern + ")"
		v.Pattern = &combinedPattern
	}

	if v.UniqueItems == nil {
		v.UniqueItems = other.UniqueItems
	}
}

type OpenAPILocation struct {
	Node *yaml.Node
}

// Clone creates a deep copy of the OpenAPILocation
func (l *OpenAPILocation) Clone() *OpenAPILocation {
	if l == nil {
		return nil
	}

	return &OpenAPILocation{
		Node: l.Node,
	}
}

// TypeDef is a definition of a type, it can represent a class, enum, container or primitive type
type TypeDef struct {
	// The unresolved name of the type, only used for complex types
	Name string `yaml:",omitempty"`

	// The original name of the type before resolution, only used for complex types
	OriginalName string `yaml:",omitempty"`

	// The hash of the type, only used for complex types
	Hash string `yaml:",omitempty"`

	// Sometimes we were lazy and didn't write "t.OriginalName" and just wrote "t.Name"
	// This flag is used to indicate that the original name has been frozen and should not be changed
	OriginalNameFrozen bool

	// The context stack for this type only present for complex types
	ContextStack ContextStack

	// Deduplicated context stacks
	DeduplicatedContextStacks ContextStacks `yaml:"-"`

	// The location of the type in the OpenAPI document. Might be nil if the type is derived.
	Location *OpenAPILocation `yaml:"-"`

	// The actual type of this definition
	Type DataType `yaml:",omitempty"`

	// If this is a container this is the type of the items
	ItemType *TypeDef `yaml:",omitempty"`

	// If this is a container that can  hold null values then this is true
	ContainsNull bool `yaml:",omitempty"`

	// Whether this is a union that can be null
	IsNullableUnion bool `yaml:",omitempty"`

	// If this is a class, this is the list of fields
	Fields Fields `yaml:",omitempty"`

	// JSON Schema Validations
	Validations *Validations `yaml:",omitempty"`

	// If this is an any type this is a list of types it could be
	AssociatedTypes TypeDefs `yaml:",omitempty"`

	// If this is an enum, this contains the details of the enum
	Enum *Enum `yaml:",omitempty"`

	// The scope of this type, only used for complex types
	Scope Scope `yaml:",omitempty"`

	// Whether this type is an inline request
	IsInlineRequestBody bool `yaml:",omitempty"`

	// Whether this type is an inline response
	IsInlineResponseBody bool `yaml:",omitempty"`

	// Whether this type is a component
	IsComponent bool `yaml:",omitempty"`

	// Whether this is a partial class that was truncated to avoid circular references, only used for field types not templating
	Truncated bool `yaml:",omitempty"`

	// Comments for this type
	Comments *Comment `yaml:",omitempty"`

	// Whether this type is an input type, ie contains write-only fields
	Input bool `yaml:",omitempty"`

	// Whether this type is an output type, ie contains read-only fields
	Output bool `yaml:",omitempty"`

	// Extensions available for this type
	Extensions *TypeDefExtensions `yaml:",omitempty"`

	// Example values for this type
	Examples Examples `yaml:",omitempty"`

	// If data type is a string, will contain a hint as to the format of the string to help drive usage snippets (e.g. format="email"); if data type is bigint it might contain "string" to indicate it should be serialized/deserialized to a JSON string
	Format string `yaml:",omitempty"`

	// ContentMediaType is the JSON Schema content vocabulary contentMediaType value for this type (e.g. "application/json")
	ContentMediaType string `yaml:",omitempty"`

	// Discriminator for union types
	Discriminator *Discriminator `yaml:",omitempty"`

	// DiscriminatorPreApplied is the name of the discriminator property that this type is always used with
	// When a type appears in multiple unions with the same discriminator property and value, this field is set
	DiscriminatorPreApplied string `yaml:",omitempty"`

	// Whether this discriminated union should be open (tolerating unknown discriminator values)
	IsUnionOpen bool `yaml:",omitempty"`

	// Whether this is a complex any type, ie it would be a union in supported languages
	ComplexAny bool `yaml:",omitempty"`

	// The location this model will be generated to if its a top level model
	OutputLocation string `yaml:",omitempty"`

	// The name of the model the type ended up in after resolution by `getModelTypes` in the templates
	ResolvedModel string `yaml:",omitempty"`

	// Whether this is an event stream envelope type
	EventStreamEnvelope bool `yaml:",omitempty"`

	// Whether this is a response envelope type
	ResponseEnvelope bool `yaml:",omitempty"`

	// Whether this type is used in a union, used for resolving type conflicts
	UsedInUnion bool `yaml:",omitempty"`

	// Whether this type is referenced by a request (as a immediate child or reachable via a chain of children)
	UsedInRequest bool `yaml:",omitempty"`

	// Whether this type is referenced by a response (as a immediate child or reachable via a chain of children)
	UsedInResponse bool `yaml:",omitempty"`

	// Whether this type is referenced by a webhook (as a immediate child or reachable via a chain of children)
	UsedInWebhook bool `yaml:",omitempty"`

	// Whether this type is referenced by a callback (as a immediate child or reachable via a chain of children)
	UsedInCallback bool `yaml:",omitempty"`

	// Whether this type is referenced by security (as a immediate child or reachable via a chain of children)
	UsedInSecurity bool `yaml:",omitempty"`

	// The reference in the serialized JSON
	Reference string `yaml:",omitempty"`

	// Whether this type has been registered
	Registered bool `yaml:"-"`

	// Special cache for sanitized enum names used during name resolution
	CachedEnumNames []string `yaml:"-"`

	// Whether this type is a multipart file
	IsMultipartFile bool `yaml:"-"`

	// EventStreamSentinel value for event streams
	EventStreamSentinel string `yaml:",omitempty"`
}

// Ensure TypeDef implements expected interfaces.
var _ jsonpointer.KeyNavigable = (*TypeDef)(nil)
var _ jsonpointer.IndexNavigable = (*TypeDef)(nil)

// Returns a new SDK TypeDef with the given name and context stack.
func NewSDKTypeDef(name string, contextStack ContextStack) *TypeDef {
	return &TypeDef{
		ContextStack: contextStack,
		Extensions: &TypeDefExtensions{
			All: make(map[string]any),
		},
		Name:  name,
		Scope: ScopeSDK,
		Type:  DataTypeClass,
	}
}

// Clone creates a deep copy of the TypeDef
func (t *TypeDef) Clone() *TypeDef {
	if t == nil {
		return nil
	}

	return t.clone(nil, nil)
}

// clone creates a deep copy of the TypeDef, tracking visited TypeDefs to avoid
// infinite recursion on circular references.
func (t *TypeDef) clone(fieldDefVisited map[*FieldDef]*FieldDef, typeDefVisited map[*TypeDef]*TypeDef) *TypeDef {
	if t == nil {
		return nil
	}

	if typeDefVisited == nil {
		typeDefVisited = make(map[*TypeDef]*TypeDef)
	}

	if existing, ok := typeDefVisited[t]; ok {
		return existing
	}

	if fieldDefVisited == nil {
		fieldDefVisited = make(map[*FieldDef]*FieldDef)
	}

	cloned := &TypeDef{
		CachedEnumNames:           slices.Clone(t.CachedEnumNames),
		Comments:                  t.Comments.Clone(),
		ComplexAny:                t.ComplexAny,
		ContainsNull:              t.ContainsNull,
		ContentMediaType:          t.ContentMediaType,
		ContextStack:              t.ContextStack.Clone(),
		DeduplicatedContextStacks: t.DeduplicatedContextStacks.Clone(),
		Discriminator:             t.Discriminator.Clone(),
		Enum:                      t.Enum.Clone(),
		EventStreamEnvelope:       t.EventStreamEnvelope,
		EventStreamSentinel:       t.EventStreamSentinel,
		Examples:                  t.Examples.Clone(),
		Extensions:                t.Extensions.Clone(),
		Format:                    t.Format,
		Hash:                      t.Hash,
		Input:                     t.Input,
		IsComponent:               t.IsComponent,
		IsInlineRequestBody:       t.IsInlineRequestBody,
		IsInlineResponseBody:      t.IsInlineResponseBody,
		IsMultipartFile:           t.IsMultipartFile,
		IsNullableUnion:           t.IsNullableUnion,
		IsUnionOpen:               t.IsUnionOpen,
		Location:                  t.Location.Clone(),
		Name:                      t.Name,
		OriginalName:              t.OriginalName,
		OriginalNameFrozen:        t.OriginalNameFrozen,
		Output:                    t.Output,
		OutputLocation:            t.OutputLocation,
		Reference:                 t.Reference,
		Registered:                t.Registered,
		ResolvedModel:             t.ResolvedModel,
		ResponseEnvelope:          t.ResponseEnvelope,
		Scope:                     t.Scope,
		Truncated:                 t.Truncated,
		Type:                      t.Type,
		UsedInUnion:               t.UsedInUnion,
		UsedInRequest:             t.UsedInRequest,
		UsedInResponse:            t.UsedInResponse,
		UsedInWebhook:             t.UsedInWebhook,
		UsedInCallback:            t.UsedInCallback,
		UsedInSecurity:            t.UsedInSecurity,
		Validations:               t.Validations.Clone(),
	}

	// Ensure we mark the cloned TypeDef as visited before cloning children to
	// avoid infinite recursion.
	typeDefVisited[t] = cloned

	cloned.AssociatedTypes = t.AssociatedTypes.clone(fieldDefVisited, typeDefVisited)
	cloned.Fields = t.Fields.clone(fieldDefVisited, typeDefVisited)
	cloned.ItemType = t.ItemType.clone(fieldDefVisited, typeDefVisited)

	return cloned
}

// DeepClone creates a fully independent deep copy of the TypeDef tree.
// Unlike Clone(), which preserves shared *TypeDef pointers via a visited map,
// DeepClone creates independent copies for every position in the tree. The
// visited maps are used only for cycle detection (back-edges): entries are
// added before recursing into children and removed after, so shared pointers
// (cross-edges) each get their own copy.
//
// This is necessary when the cloned tree will be mutated by path-dependent
// operations (e.g. TerraformPropagateComputedParent) that propagate state
// differently depending on the path to a node. OAS $ref resolution caches
// TypeDef pointers (getCachedType in schemas.go), so the same *TypeDef can
// appear at multiple positions in the tree. Clone() preserves this sharing;
// DeepClone() breaks it.
func (t *TypeDef) DeepClone() *TypeDef {
	if t == nil {
		return nil
	}

	return t.deepClone(nil, nil)
}

// deepClone creates a fully independent copy using stack-based cycle detection.
// Unlike clone(), entries are removed from the visited maps after processing
// children, so shared pointers encountered from different paths get fresh copies.
func (t *TypeDef) deepClone(fieldDefStack map[*FieldDef]*FieldDef, typeDefStack map[*TypeDef]*TypeDef) *TypeDef {
	if t == nil {
		return nil
	}

	if typeDefStack == nil {
		typeDefStack = make(map[*TypeDef]*TypeDef)
	}

	// Only return cached clone for cycles (back-edges on the current stack).
	if existing, ok := typeDefStack[t]; ok {
		return existing
	}

	if fieldDefStack == nil {
		fieldDefStack = make(map[*FieldDef]*FieldDef)
	}

	cloned := &TypeDef{
		CachedEnumNames:           slices.Clone(t.CachedEnumNames),
		Comments:                  t.Comments.Clone(),
		ComplexAny:                t.ComplexAny,
		ContainsNull:              t.ContainsNull,
		ContentMediaType:          t.ContentMediaType,
		ContextStack:              t.ContextStack.Clone(),
		DeduplicatedContextStacks: t.DeduplicatedContextStacks.Clone(),
		Discriminator:             t.Discriminator.Clone(),
		Enum:                      t.Enum.Clone(),
		EventStreamEnvelope:       t.EventStreamEnvelope,
		EventStreamSentinel:       t.EventStreamSentinel,
		Examples:                  t.Examples.Clone(),
		Extensions:                t.Extensions.Clone(),
		Format:                    t.Format,
		Hash:                      t.Hash,
		Input:                     t.Input,
		IsComponent:               t.IsComponent,
		IsInlineRequestBody:       t.IsInlineRequestBody,
		IsInlineResponseBody:      t.IsInlineResponseBody,
		IsMultipartFile:           t.IsMultipartFile,
		IsNullableUnion:           t.IsNullableUnion,
		IsUnionOpen:               t.IsUnionOpen,
		Location:                  t.Location.Clone(),
		Name:                      t.Name,
		OriginalName:              t.OriginalName,
		OriginalNameFrozen:        t.OriginalNameFrozen,
		Output:                    t.Output,
		OutputLocation:            t.OutputLocation,
		Reference:                 t.Reference,
		Registered:                t.Registered,
		ResolvedModel:             t.ResolvedModel,
		ResponseEnvelope:          t.ResponseEnvelope,
		Scope:                     t.Scope,
		Truncated:                 t.Truncated,
		Type:                      t.Type,
		UsedInUnion:               t.UsedInUnion,
		UsedInRequest:             t.UsedInRequest,
		UsedInResponse:            t.UsedInResponse,
		UsedInWebhook:             t.UsedInWebhook,
		UsedInCallback:            t.UsedInCallback,
		UsedInSecurity:            t.UsedInSecurity,
		Validations:               t.Validations.Clone(),
	}

	// Push onto stack for cycle detection.
	typeDefStack[t] = cloned

	cloned.AssociatedTypes = t.AssociatedTypes.deepClone(fieldDefStack, typeDefStack)
	cloned.Fields = t.Fields.deepClone(fieldDefStack, typeDefStack)
	cloned.ItemType = t.ItemType.deepClone(fieldDefStack, typeDefStack)

	// Pop from stack — future encounters of t from other paths will
	// create fresh independent clones, not reuse this one.
	delete(typeDefStack, t)

	return cloned
}

// EnsureExtensions initializes Extensions if nil, including the All map.
func (t *TypeDef) EnsureExtensions() {
	if t.Extensions == nil {
		t.Extensions = &TypeDefExtensions{
			All: make(map[string]any),
		}
	}
}

// Returns the name of the given associated type using the precedence of:
// - Discriminator mapping name
// - OriginalName
// - Name
// - Associated type DataType
func (t *TypeDef) AssociatedTypeName(associatedType *TypeDef) string {
	if t == nil || associatedType == nil || len(t.AssociatedTypes) == 0 {
		return ""
	}

	if t.Discriminator != nil {
		for _, mapping := range t.Discriminator.Mapping {
			if mapping.Type.Name == associatedType.Name || (mapping.Type.OriginalName != "" && mapping.Type.OriginalName == associatedType.OriginalName) {
				return mapping.Name
			}
		}
	}

	if associatedType.OriginalName != "" {
		return associatedType.OriginalName
	}

	if associatedType.Name != "" {
		return associatedType.Name
	}

	switch associatedType.Type {
	case DataTypeArray, DataTypeMap, DataTypeSet:
		return string(associatedType.Type) + "_Of_" + associatedType.AssociatedTypeName(associatedType.ItemType)
	case DataTypeString:
		return "Str"
	default:
		return string(associatedType.Type)
	}
}

// Returns a mutated TypeDef that has been merged with the given TypeDef. The
// algorithm adds data from the given TypeDef to this TypeDef where it is
// undefined. Where there is conflicting data, this TypeDef's data is preserved
// or otherwise delegated to data-specific merge functionality. It does not
// remove any data from this TypeDef.
//
// Fields are matched using FindTerraformEquivalentField (sanitized name
// matching) and sorted by Name after merging. AssociatedTypes are matched
// using dual-context name lookup and sorted by OriginalName after merging.
//
// NOTE: This algorithm was originally designed for Terraform and did not have
// handling for every struct field. If considering for other use cases, deeply
// evaluate the behavior to ensure it is suitable.
func (t *TypeDef) TerraformMerge(other *TypeDef) error {
	if t == nil {
		return errors.New("cannot merge into nil TypeDef")
	}

	if other == nil {
		return nil
	}

	dataType, err := t.TerraformMergeDataType(other)

	if err != nil {
		return err
	}

	t.Type = dataType

	for _, otherAssociatedType := range other.AssociatedTypes {
		// Dual-context name lookup: match using both the other's and
		// receiver's discriminator context. The discriminator mapping may
		// only be present on one side due to merging order.
		otherName := other.AssociatedTypeName(otherAssociatedType)
		selfName := t.AssociatedTypeName(otherAssociatedType)

		var existingAssociatedType *TypeDef

		for _, candidate := range t.AssociatedTypes {
			candidateName := t.AssociatedTypeName(candidate)

			if candidateName == otherName || candidateName == selfName {
				existingAssociatedType = candidate
				break
			}
		}

		if existingAssociatedType == nil {
			t.AssociatedTypes = append(t.AssociatedTypes, otherAssociatedType.Clone())
			continue
		}

		if err := existingAssociatedType.TerraformMerge(otherAssociatedType); err != nil {
			name := t.AssociatedTypeName(existingAssociatedType)
			return fmt.Errorf("cannot merge associated type %q: %w", name, err)
		}
	}

	slices.SortFunc(t.AssociatedTypes, func(a, b *TypeDef) int {
		return strings.Compare(a.OriginalName, b.OriginalName)
	})

	// Source-wins comment merge: other's values take priority.
	if other.Comments != nil {
		if t.Comments == nil {
			t.Comments = other.Comments
		} else {
			mergedComments := other.Comments.Clone()
			mergedComments.Merge(t.Comments) // fills in blanks from t where other is empty
			t.Comments = mergedComments
		}
	}

	if t.Discriminator == nil {
		t.Discriminator = other.Discriminator
	}

	if t.Enum == nil {
		t.Enum = other.Enum
	}

	if t.Examples == nil {
		t.Examples = other.Examples
	} else {
		t.Examples.Merge(other.Examples)
	}

	if t.Extensions == nil {
		t.Extensions = other.Extensions
	} else {
		t.Extensions.TerraformMerge(other.Extensions)
	}

	for _, otherField := range other.Fields {
		existingField, _ := otherField.FindTerraformEquivalentField(t.Fields, false)

		if existingField == nil {
			t.Fields = append(t.Fields, otherField.Clone())
			continue
		}

		if err := existingField.Type.TerraformMerge(otherField.Type); err != nil {
			return fmt.Errorf("cannot merge field %q: %w", existingField.Name, err)
		}

		// Source-wins comment merge for fields.
		if otherField.Comments != nil {
			if existingField.Comments == nil {
				existingField.Comments = otherField.Comments
			} else {
				mergedComments := otherField.Comments.Clone()
				mergedComments.Merge(existingField.Comments)
				existingField.Comments = mergedComments
			}
		}
	}

	slices.SortFunc(t.Fields, func(a, b *FieldDef) int {
		return strings.Compare(a.Name, b.Name)
	})

	if t.ContentMediaType == "" {
		t.ContentMediaType = other.ContentMediaType
	}

	if !t.Input {
		t.Input = other.Input
	}

	if t.ItemType == nil {
		t.ItemType = other.ItemType
	} else if other.ItemType != nil {
		if err := t.ItemType.TerraformMerge(other.ItemType); err != nil {
			return fmt.Errorf("cannot merge item type: %w", err)
		}
	}

	if t.Name != other.Name && len(t.Name) > len(other.Name) {
		t.Name = other.Name
	}

	if !t.Output {
		t.Output = other.Output
	}

	if t.Validations == nil {
		t.Validations = other.Validations
	} else {
		t.Validations.Merge(other.Validations)
	}

	return nil
}

// Returns the most suitable DataType when there is a mismatch between the
// TypeDef and the given TypeDef. An error is returned if the mismatch cannot
// be reconciled.
//
// NOTE: This algorithm was originally designed for Terraform. If considering
// for other use cases, deeply evaluate the behavior to ensure it is suitable.
func (t *TypeDef) TerraformMergeDataType(other *TypeDef) (DataType, error) {
	if other == nil {
		return t.Type, nil
	}

	if t.Type == other.Type {
		return t.Type, nil
	}

	if t.Type == DataTypeClass && other.Type == DataTypeUnion {
		return t.Type, nil
	}

	if t.Type == DataTypeUnion && other.Type == DataTypeClass {
		return t.Type, nil
	}

	if t.Type == DataTypeEnum {
		if _, err := t.Enum.Type.TerraformMergeDataType(other); err != nil {
			return t.Type, fmt.Errorf("unwrapping enum: %w", err)
		}

		return t.Type, nil
	}

	if other.Type == DataTypeEnum {
		if _, err := t.TerraformMergeDataType(other.Enum.Type); err != nil {
			return t.Type, fmt.Errorf("unwrapping enum: %w", err)
		}

		return t.Type, nil
	}

	switch t.Type {
	case DataTypeBigInt, DataTypeInteger, DataTypeInt32:
		switch other.Type {
		case DataTypeBigInt, DataTypeInteger, DataTypeInt32, DataTypeFloat32:
			return t.Type, nil
		}
	case DataTypeClass:
		if other.Type == DataTypeAny {
			return t.Type, nil
		}
	case DataTypeFloat32:
		if slices.Contains([]DataType{DataTypeBigInt, DataTypeInteger, DataTypeInt32}, other.Type) {
			return other.Type, nil
		}
	}

	return t.Type, fmt.Errorf("%w: cannot reconcile between %s and %s", ErrTypeMismatch, t.Type, other.Type)
}

func (t *TypeDef) FreezeOriginalName() {
	if t.OriginalNameFrozen {
		return
	}

	if t.OriginalName == "" {
		t.OriginalName = t.Name
	}
	t.OriginalNameFrozen = true
}

func (t *TypeDef) Match(matchers Matchers) error {
	if matchers.TypeDef != nil {
		return matchers.TypeDef(t)
	}

	return nil
}

func (t TypeDef) MarshalYAML() (any, error) {
	type shadowTypeDef TypeDef
	s := shadowTypeDef(t)

	// If this isn't set we assume we are serializing the typedef in the components section
	if t.Reference == "" {
		return s, nil
	}

	// Otherwise we are serializing just a reference to the type
	type typeDefRef struct {
		Reference string
	}

	return typeDefRef{
		Reference: t.Reference,
	}, nil
}

// NavigateWithKey implements the jsonpointer.KeyNavigable interface for
// traversing JSON Pointers, such as Arazzo conditions.
func (t TypeDef) NavigateWithKey(key string) (any, error) {
	if t.Type == DataTypeUnion {
		for _, at := range t.AssociatedTypes {
			val, err := at.NavigateWithKey(key)
			if err == nil {
				return val, nil
			}
		}
		return nil, jsonpointer.ErrNotFound
	}

	if !t.IsTypeWithFields() {
		return nil, jsonpointer.ErrInvalidPath
	}

	var additionalProperties *FieldDef

	for _, f := range t.Fields {
		if f.IsAdditionalProperties {
			additionalProperties = f
		}
		if f.OriginalName == key {
			return f, nil
		}
	}

	// Fallback: match by potentially overridden Name (e.g. x-speakeasy-name-override)
	for _, f := range t.Fields {
		if f.Name == key {
			return f, nil
		}
	}

	if additionalProperties != nil {
		return additionalProperties, nil
	}

	return nil, jsonpointer.ErrNotFound
}

// NavigateWithIndex implements the jsonpointer.IndexNavigable interface for
// traversing JSON Pointers, such as Arazzo conditions.
func (t TypeDef) NavigateWithIndex(index int) (any, error) {
	if t.Type != DataTypeArray && t.Type != DataTypeSet {
		return nil, jsonpointer.ErrInvalidPath
	}

	return t.ItemType, nil
}

func NewType(typ *TypeDef, contextStack ContextStack) *TypeDef {
	t := typ

	if contextStack != nil {
		t.ContextStack = make(ContextStack, len(contextStack))
		copy(t.ContextStack, contextStack)
	}

	if t.Extensions == nil {
		t.Extensions = &TypeDefExtensions{
			All: make(map[string]any),
		}
	}

	return t
}

// Returns true if this TypeDef or any of its children are truncated (contain
// circular references).
func (t *TypeDef) ContainsTruncated() bool {
	for _, typeDef := range t.Walk() {
		if typeDef.Truncated {
			return true
		}
	}

	return false
}

// Returns the associated type that matches the given TypeDef. The algorithm
// uses the naming precedence of:
// - Discriminator mapping name
// - OriginalName
// - Name
// - Associated type DataType
func (t *TypeDef) FindAssociatedTypeByTypeDef(typeDef *TypeDef) *TypeDef {
	if t == nil || typeDef == nil || len(t.AssociatedTypes) == 0 {
		return nil
	}

	candidateName := t.AssociatedTypeName(typeDef)

	for _, associatedType := range t.AssociatedTypes {
		if candidateName == t.AssociatedTypeName(associatedType) {
			return associatedType
		}
	}

	return nil
}

// Returns this TypeDef or any of its children when the given entity name
// matches the x-speakeasy-entity configuration.
func (t *TypeDef) FindEntityTypeDef(entityName string) *TypeDef {
	for _, typeDef := range t.Walk() {
		if typeDef.HasEntityName(entityName) {
			return typeDef
		}
	}

	return nil
}

// FindEntitySDKMethodTargets returns the TerraformSDKMethodTarget for each
// TypeDef on the path from this TypeDef to the TypeDef matching the given
// entity name, ordered deepest match first (entity TypeDef) to shallowest
// (this TypeDef). Returns nil if no matching entity TypeDef is found.
//
// The optional parameter indicates whether this TypeDef was reached through an
// optional/nullable access path, which propagates into each TerraformSDKMethodTarget.
//
// Only the first matching path is returned via depth-first search through
// Fields, then ItemType, then AssociatedTypes.
func (t *TypeDef) FindEntitySDKMethodTargets(entityName string, optional bool) []TerraformSDKMethodTarget {
	return t.findEntitySDKMethodTargets(entityName, optional, make(map[*TypeDef]bool))
}

func (t *TypeDef) findEntitySDKMethodTargets(entityName string, optional bool, visited map[*TypeDef]bool) []TerraformSDKMethodTarget {
	if t == nil {
		return nil
	}

	self := TerraformSDKMethodTarget{TypeDef: t, Optional: optional}

	if t.HasEntityName(entityName) {
		return []TerraformSDKMethodTarget{self}
	}

	if visited[t] {
		return nil
	}

	visited[t] = true

	for _, field := range t.Fields {
		if field.Type == nil {
			continue
		}

		result := field.Type.findEntitySDKMethodTargets(entityName, field.Optional || field.Nullable, visited)

		if len(result) > 0 {
			return append(result, self)
		}
	}

	if t.ItemType != nil {
		result := t.ItemType.findEntitySDKMethodTargets(entityName, false, visited)

		if len(result) > 0 {
			return append(result, self)
		}
	}

	for _, associated := range t.AssociatedTypes {
		result := associated.findEntitySDKMethodTargets(entityName, true, visited)

		if len(result) > 0 {
			return append(result, self)
		}
	}

	return nil
}

// Returns true if TypeDef has x-speakeasy-entity configured with the given
// entity name.
func (t *TypeDef) HasEntityName(entityName string) bool {
	if t == nil || t.Extensions == nil || t.Extensions.Entity == nil {
		return false
	}

	return slices.Contains(t.Extensions.Entity.Names, entityName)
}

// IsCustomType returns true if this type is a named type defined by the OpenAPI Document
func (t *TypeDef) IsCustomType() bool {
	switch t.Type {
	case DataTypeClass, DataTypeEnum, DataTypeError, DataTypeUnion:
		return true
	default:
		return false
	}
}

// IsCustomClass returns true if this type is a class defined by the OpenAPI Document
func (t *TypeDef) IsCustomClass() bool {
	switch t.Type {
	case DataTypeClass, DataTypeError, DataTypeUnion:
		return true
	default:
		return false
	}
}

// HasFields returns true if this type has fields
func (t *TypeDef) IsTypeWithFields() bool {
	return t.Type == DataTypeClass || t.Type == DataTypeError
}

// IsObjectType returns true if this type is a class, map or any type and generally serializes to a JSON object
func (t *TypeDef) IsObjectType() bool {
	switch t.Type {
	case DataTypeClass, DataTypeError, DataTypeMap, DataTypeUnion:
		return true
	case DataTypeAny:
		return len(t.AssociatedTypes) > 0 || t.ComplexAny
	default:
		return false
	}
}

// IsSimpleObjectOrContainerType returns true if this type is a class, map or array that contains a simple object with no complex type fields
func (t *TypeDef) IsSimpleObjectOrContainerType() bool {
	switch t.Type {
	case DataTypeError:
		fallthrough
	case DataTypeClass:
		for _, f := range t.Fields {
			if !f.Type.IsPrimitive() {
				return false
			}
		}
		return true
	case DataTypeEventStream:
		return t.ItemType.IsPrimitive()
	case DataTypeMap:
		return t.ItemType.IsPrimitive()
	case DataTypeAny:
		for _, at := range t.AssociatedTypes {
			if !at.IsPrimitive() {
				return false
			}
		}
		return true
	case DataTypeSet:
		fallthrough
	case DataTypeArray:
		return t.ItemType.IsPrimitive()
	default:
		return false
	}
}

// IsContainer returns true if this type is an array or map
func (t *TypeDef) IsContainer() bool {
	switch t.Type {
	case DataTypeArray, DataTypeSet, DataTypeMap, DataTypeEventStream:
		return true
	default:
		return false
	}
}

// IsPrimitiveContainer returns true if this type is an array or map of primitive types
func (t *TypeDef) IsPrimitiveContainer() bool {
	switch t.Type {
	case DataTypeArray, DataTypeSet, DataTypeMap, DataTypeEventStream:
		return t.ItemType.IsPrimitive()
	default:
		return false
	}
}

// IsPrimitive returns true if this type is a primitive type
func (t *TypeDef) IsPrimitive() bool {
	return !t.IsObjectType() && !t.IsContainer() && t.Type != DataTypeBytes && t.Type != DataTypeRequestStream && t.Type != DataTypeResponseStream && t.Type != DataTypeEventStream
}

// IsInput returns true if this type is an input type, ie contains write-only fields
func (t *TypeDef) IsInput(nameResolutionFixesDec2023 bool) bool {
	for s, t := range t.Walk() {
		if _, isCircular := s.Parents.Get(t); isCircular {
			return !nameResolutionFixesDec2023
		}

		if t.Input {
			return true
		}
	}

	return false
}

// IsOutput returns true if this type is an output type, ie contains read-only fields
func (t *TypeDef) IsOutput(nameResolutionFixesDec2023 bool) bool {
	for s, t := range t.Walk() {
		if _, isCircular := s.Parents.Get(t); isCircular {
			return !nameResolutionFixesDec2023
		}

		if t.Output {
			return true
		}
	}

	return false
}

func (t *TypeDef) IsRequest() bool {
	if len(t.ContextStack) != 1 {
		return false
	}

	requestFrame := t.ContextStack.FindLastFrameOfType(ContextTypeRequestResponse)
	if requestFrame == nil {
		return false
	}

	return requestFrame.Identifier == InternalTypeRequest
}

// IsEmpty returns true if this type can be considered empty. For example, a
// class type with no child fields.
func (t *TypeDef) IsEmpty() bool {
	if t.Type == DataTypeClass {
		return len(t.Fields) == 0
	}

	return false
}

type isEqualOpts struct {
	skipAnnotations bool
	visited         map[*TypeDef]bool
}

type IsEqualOpt func(opts *isEqualOpts)

func SkipAnnotations() IsEqualOpt {
	return func(opts *isEqualOpts) {
		opts.skipAnnotations = true
	}
}

func WithVisited(visited map[*TypeDef]bool) IsEqualOpt {
	return func(opts *isEqualOpts) {
		opts.visited = visited
	}
}

func (t TypeDef) IsEqualType(other TypeDef) bool {
	return t.IsEqual(&other) == nil
}

func (t *TypeDef) IsEqual(other *TypeDef, opts ...IsEqualOpt) error {
	o := &isEqualOpts{}

	for _, opt := range opts {
		opt(o)
	}

	if o.visited == nil {
		o.visited = make(map[*TypeDef]bool)
		opts = append(opts, WithVisited(o.visited))
	}

	if o.visited[t] {
		return nil
	}

	o.visited[t] = true

	if t == other {
		return nil
	}

	if t.Type != other.Type {
		return ErrTypeMismatch.Wrap(fmt.Errorf("expected %s in %s, got %s in %s", t.Type, t.getNameOrType(), other.Type, other.getNameOrType()))
	}

	if t.Scope != other.Scope {
		return ErrScopeMismatch.Wrap(fmt.Errorf("expected %s in %s, got %s in %s", t.Scope, t.getNameOrType(), other.Scope, other.getNameOrType()))
	}

	switch t.Type {
	case DataTypeEnum:
		if t.Enum == nil && other.Enum != nil || t.Enum != nil && other.Enum == nil {
			return ErrEnumMismatch
		} else if t.Enum != nil && other.Enum != nil {
			if !slices.Equal(t.Enum.Values, other.Enum.Values) {
				return ErrEnumMismatch.Wrap(fmt.Errorf("expected values %v in %s, got %v in %s", t.Enum.Values, t.getNameOrType(), other.Enum.Values, other.getNameOrType()))
			}
			if !slices.Equal(t.Enum.Names, other.Enum.Names) {
				return ErrEnumMismatch.Wrap(fmt.Errorf("expected names %v in %s, got %v in %s", t.Enum.Names, t.getNameOrType(), other.Enum.Names, other.getNameOrType()))
			}
			if !maps.Equal(t.Enum.Descriptions, other.Enum.Descriptions) {
				return ErrEnumMismatch.Wrap(fmt.Errorf("expected descriptions %v in %s, got %v in %s", t.Enum.Descriptions, t.getNameOrType(), other.Enum.Descriptions, other.getNameOrType()))
			}
		}
	case DataTypeUnion:
		if t.Discriminator == nil && other.Discriminator != nil || t.Discriminator != nil && other.Discriminator == nil {
			return ErrUnionDiscriminatorMismatch
		}
		if t.Discriminator != nil && other.Discriminator != nil {
			if t.Discriminator.TypePropertyName != other.Discriminator.TypePropertyName {
				return ErrUnionDiscriminatorMismatch.Wrap(fmt.Errorf("expected discriminator %s in %s, got %s in %s", t.Discriminator.TypePropertyName, t.getNameOrType(), other.Discriminator.TypePropertyName, other.getNameOrType()))
			}
			for i, mapping := range t.Discriminator.Mapping {
				if mapping.Name != other.Discriminator.Mapping[i].Name {
					return ErrUnionDiscriminatorMismatch.Wrap(fmt.Errorf("expected discriminator mapping %s in %s, got %s in %s", mapping.Name, t.getNameOrType(), other.Discriminator.Mapping[i].Name, other.getNameOrType()))
				}
				if err := mapping.Type.IsEqual(other.Discriminator.Mapping[i].Type, opts...); err != nil {
					return ErrUnionDiscriminatorMismatch.Wrap(err)
				}
			}
		}
		if len(t.AssociatedTypes) != len(other.AssociatedTypes) {
			return ErrUnionTypeMismatch.Wrap(fmt.Errorf("expected %d associated types in %s, got %d in %s", len(t.AssociatedTypes), t.getNameOrType(), len(other.AssociatedTypes), other.getNameOrType()))
		}
		for i, at := range t.AssociatedTypes {
			if err := at.IsEqual(other.AssociatedTypes[i], opts...); err != nil {
				return ErrUnionTypeMismatch.Wrap(err)
			}
		}
	case DataTypeError:
		fallthrough
	case DataTypeClass:
		// If either is truncated we can only assume they are equal
		if !t.Truncated && !other.Truncated {
			var fieldErr error

			if len(t.Fields) != len(other.Fields) {
				fieldErr = fmt.Errorf("expected %d fields in %s, got %d in %s. difference=[%s]", len(t.Fields), t.getNameOrType(), len(other.Fields), other.getNameOrType(), t.Fields.Difference(other.Fields).GetFieldNames())
			} else {
				for i, f := range t.Fields {
					if f.Name != other.Fields[i].Name {
						fieldErr = fmt.Errorf("expected field %s in %s, got %s in %s", f.Name, t.getNameOrType(), other.Fields[i].Name, other.getNameOrType())
						break
					}

					if err := f.Type.IsEqual(other.Fields[i].Type, opts...); err != nil {
						fieldErr = errors.New(err.Error())
						break
					}

					if f.Optional != other.Fields[i].Optional {
						fieldErr = fmt.Errorf("expected field %s to be optional: %t in %s, got %t in %s", f.Name, f.Optional, t.getNameOrType(), other.Fields[i].Optional, other.getNameOrType())
						break
					}
				}
			}

			if fieldErr != nil {
				return ErrFieldMismatch.Wrap(fieldErr)
			}

			if !o.skipAnnotations {
				for i, f := range other.Fields {
					for _, a := range f.Annotations {
						_, existingAnnotation := t.Fields[i].Annotations.Find(a)

						if existingAnnotation != nil && !a.IsEqual(existingAnnotation) {
							return ErrAnnotationMismatch.Wrap(fmt.Errorf("expected annotation %s in %s, got %s in %s", a.Type(), t.getNameOrType(), existingAnnotation.Type(), other.getNameOrType()))
						}
					}
				}
			}
		}
	default:
		switch t.Type {
		case DataTypeMap, DataTypeSet, DataTypeArray, DataTypeEventStream:
			if t.ItemType != nil && other.ItemType == nil || t.ItemType == nil && other.ItemType != nil {
				return ErrItemTypeMismatch
			}

			if t.ItemType != nil && other.ItemType != nil {
				if err := t.ItemType.IsEqual(other.ItemType, opts...); err != nil {
					return ErrItemTypeMismatch.Wrap(err)
				}
			}

			// Check EventStreamSentinel for EventStream types
			if t.Type == DataTypeEventStream && t.EventStreamSentinel != other.EventStreamSentinel {
				return ErrTypeMismatch.Wrap(fmt.Errorf("EventStreamSentinel mismatch: expected %s in %s, got %s in %s", t.EventStreamSentinel, t.getNameOrType(), other.EventStreamSentinel, other.getNameOrType()))
			}
		}
	}

	return nil
}

// Returns true if the TypeDef is equal to given TypeDef for Terraform usage.
//
// NOTE: This logic is a subset of the full equality logic in IsEqual, focusing
// on aspects relevant to Terraform. It also differs from IsEqual in certain
// ways, such as treating matching TypeDef DataType and Enum underlying DataType
// as equivalent.
func (t *TypeDef) IsTerraformEqual(other *TypeDef) bool {
	if t == nil || other == nil {
		return t == other
	}

	if t == other {
		return true
	}

	if t.Type != other.Type && t.Enum == nil && other.Enum == nil {
		return false
	}

	if t.Enum != nil || other.Enum != nil {
		tDataType := t.Type
		otherDataType := other.Type

		if t.Enum != nil {
			tDataType = t.Enum.Type.Type
		}

		if other.Enum != nil {
			otherDataType = other.Enum.Type.Type
		}

		if tDataType != otherDataType {
			return false
		}
	}

	if len(t.AssociatedTypes) != len(other.AssociatedTypes) {
		return false
	}

	for _, associatedType := range t.AssociatedTypes {
		otherAssociatedType := other.FindAssociatedTypeByTypeDef(associatedType)

		if !associatedType.IsTerraformEqual(otherAssociatedType) {
			return false
		}
	}

	if !t.Fields.IsTerraformEqual(other.Fields) {
		return false
	}

	if !t.ItemType.IsTerraformEqual(other.ItemType) {
		return false
	}

	return true
}

// Returns true if TypeDef DataType is a primitive type for Terraform usage.
func (t *TypeDef) IsTerraformPrimitiveType() bool {
	if t == nil {
		return false
	}

	switch t.Type {
	case DataTypeAny, DataTypeBigInt, DataTypeBoolean, DataTypeBytes, DataTypeDate, DataTypeDateTime, DataTypeDecimal, DataTypeFloat32, DataTypeInt32, DataTypeInteger, DataTypeNumber, DataTypeString:
		return true
	case DataTypeEnum:
		if t.Enum == nil || t.Enum.Type == nil {
			return false
		}

		return t.Enum.Type.IsTerraformPrimitiveType()
	default:
		return false
	}
}

func (t *TypeDef) GetFullyQualifiedName() string {
	switch t.Type {
	case DataTypeUnion:
		fallthrough
	case DataTypeEnum:
		fallthrough
	case DataTypeError:
		fallthrough
	case DataTypeClass:
		return t.GetRegistrationID()
	case DataTypeSet:
		fallthrough
	case DataTypeArray:
		return fmt.Sprintf("Array<%s>", t.ItemType.GetFullyQualifiedName())
	case DataTypeMap:
		return fmt.Sprintf("Map<%s>", t.ItemType.GetFullyQualifiedName())
	case DataTypeEventStream:
		return fmt.Sprintf("EventStream<%s>", t.ItemType.GetFullyQualifiedName())
	case DataTypeJsonL:
		return fmt.Sprintf("JsonL<%s>", t.ItemType.GetFullyQualifiedName())
	default:
		return string(t.Type)
	}
}

type RegistrationIDOption func(opts *RegistrationIDOptions)

type RegistrationIDOptions struct {
	SkipDuplicateFrame   bool
	SkipInputOutputFrame bool
}

func WithSkipDuplicateFrame() RegistrationIDOption {
	return func(opts *RegistrationIDOptions) {
		opts.SkipDuplicateFrame = true
	}
}

func WithSkipInputOutputFrame() RegistrationIDOption {
	return func(opts *RegistrationIDOptions) {
		opts.SkipInputOutputFrame = true
	}
}

func (t *TypeDef) GetRegistrationID(opts ...RegistrationIDOption) string {
	// FreezeOriginalName ensures the OriginalName field is set and immutable for consistent type registration deduplication.
	// If OriginalName is uninitialized, this copies the current Name value to preserve the original naming context. Freezing
	// prevents accidental overwrites if Name is later modified, ensuring registration IDs remain stable even when types are
	// subsequently renamed or refined during processing. eg consider t.OriginalName is "" - freezing only once preserves this.
	t.FreezeOriginalName()
	o := &RegistrationIDOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if t.IsStructurallyDeduplicated() {
		return t.getStructuralRegistrationID()
	}
	return GetRegistrationID(t.Scope, t.ContextStack, t.OriginalName, opts...)
}

// This has been factored out so that it can be use to find pre-built types
// in the type register / other caches
func GetRegistrationID(scope Scope, contextStack ContextStack, originalName string, opts ...RegistrationIDOption) string {
	o := &RegistrationIDOptions{}
	for _, opt := range opts {
		opt(o)
	}
	builder := strings.Builder{}
	builder.WriteString("scope:")
	builder.WriteString(string(scope))

	if contextStack == nil {
		panic(fmt.Errorf("GetRegistrationID() originalName:\"%s\" scope:\"%s\" has no context stack", originalName, scope))
	}
	for _, frame := range contextStack {
		// Skip input/output context frames as for truncated classes they don't get added
		// or if we are skipping the duplicate frame
		if (o.SkipInputOutputFrame && frame.Type == ContextTypeInputOutput) || (o.SkipDuplicateFrame && frame.Type == ContextTypeRegisterDuplicate) {
			continue
		}
		// skip response type error from the registration ids: they don't impact if we should duplicate/deduplicate the type
		// and currently just live for debug purposes.
		if frame.Type == ContextTypeResponseError {
			continue
		}

		builder.WriteString(" ")
		builder.WriteString(string(frame.Type))
		builder.WriteString(":")
		builder.WriteString(frame.Identifier)
	}
	builder.WriteString(" originalName:")
	builder.WriteString(originalName)

	return builder.String()
}

// Unlike GetRegistrationID() this method will not panic if there's no
// ContextStack If possible GetRegistrationID will be used. Otherwise
// this is a best effort unique string for this type.
// The ID is not guaranteed to be unique.
func (t *TypeDef) GetRegistrationIDOrType() string {
	if t.IsCustomType() && t.ContextStack != nil {
		return t.GetRegistrationID()
	}

	builder := strings.Builder{}
	if t.Scope != "" {
		builder.WriteString("scope:")
		builder.WriteString(string(t.Scope))
	}
	if t.ContextStack != nil {
		builder.WriteString(t.ContextStack.String())
	}
	if builder.Len() > 0 {
		builder.WriteString(" ")
	}
	builder.WriteString("type:")
	builder.WriteString(string(t.Type))
	if t.Name != "" {
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString("name:")
		builder.WriteString(t.Name)
	}
	return builder.String()
}

func (t *TypeDef) IsStructurallyDeduplicated() bool {
	return t.Scope == ScopeErrors && t.Hash != ""
}

func (t *TypeDef) getStructuralRegistrationID() string {
	builder := strings.Builder{}
	builder.WriteString("scope:")
	builder.WriteString(string(t.Scope))

	if t.ContextStack.HasFrameOfType(ContextTypeResponseError) {
		// Errors should always be scoped to status code
		builder.WriteString(" status:")
		status := t.ContextStack.FindLastFrameOfType(ContextTypeResponseStatusCode)
		if status != nil {
			builder.WriteString(status.Identifier)
		} else {
			builder.WriteString("unknown")
		}
	}

	builder.WriteString(" hash:")
	builder.WriteString(t.Hash)
	builder.WriteString(" originalName:")
	builder.WriteString(t.OriginalName)
	return builder.String()
}

// Returns the shallow children of a type
// If it has fields, it will return the fields
// If it's a union, it will return the associated types as children (fieldDef is nil)
// If it's a container, it will return the item type as a child (fieldDef is nil)
func (t *TypeDef) Children(yield func(field *FieldDef, typ *TypeDef) bool) {
	for _, f := range t.Fields {
		if !yield(f, f.Type) {
			return
		}
	}

	if t.ItemType != nil {
		if !yield(nil, t.ItemType) {
			return
		}
	}

	for _, at := range t.AssociatedTypes {
		if !yield(nil, at) {
			return
		}
	}

	if t.Enum != nil && t.Enum.Type != nil {
		if !yield(nil, t.Enum.Type) {
			return
		}
	}
}

// Walk returns an iterator that traverses the type definition tree depth-first.
// Note: The walk includes the type itself as the first element yielded.
func (t *TypeDef) Walk() func(func(state *TypeDefWalkState, typeDef *TypeDef) bool) {
	return t.walk(nil)
}

type TypeDefWalkState struct {
	Visited map[*TypeDef]bool
	Parents *sequencedmap.Map[*TypeDef, bool]
	Aborted bool
}

func (t *TypeDef) walk(state *TypeDefWalkState) func(func(state *TypeDefWalkState, typeDef *TypeDef) bool) {
	if state == nil {
		state = &TypeDefWalkState{
			Visited: map[*TypeDef]bool{},
			Parents: sequencedmap.New[*TypeDef, bool](),
		}
	}

	return func(yield func(state *TypeDefWalkState, typeDef *TypeDef) bool) {
		if t == nil {
			return
		}

		if !yield(state, t) {
			state.Aborted = true
			return
		}

		if state.Visited[t] {
			return
		}

		state.Visited[t] = true

		state.Parents.Set(t, true)

		for _, child := range t.Children {
			if state.Aborted {
				return
			}
			child.walk(state)(yield)
		}

		state.Parents.Delete(t)
	}
}

// ShallowCopy creates a new instance of a TypeDef from an original, generally allowing for superficial changes
// to the original without affecting the original. Use carefully as it could cause conflicts if not changed sufficiently and reregistered.
// Currently only being used in tests to reference the original type but add custom examples to avoid overwriting the original.
func (t TypeDef) ShallowCopy() *TypeDef {
	return &TypeDef{
		Name:                t.Name,
		OriginalName:        t.OriginalName,
		Type:                t.Type,
		ItemType:            t.ItemType,
		ContainsNull:        t.ContainsNull,
		ContentMediaType:    t.ContentMediaType,
		Fields:              t.Fields,
		ContextStack:        t.ContextStack,
		AssociatedTypes:     t.AssociatedTypes,
		Enum:                t.Enum,
		Scope:               t.Scope,
		IsComponent:         t.IsComponent,
		Truncated:           t.Truncated,
		Comments:            t.Comments,
		Input:               t.Input,
		Output:              t.Output,
		Extensions:          t.Extensions,
		Examples:            t.Examples,
		Format:              t.Format,
		Discriminator:       t.Discriminator,
		ComplexAny:          t.ComplexAny,
		OutputLocation:      t.OutputLocation,
		ResolvedModel:       t.ResolvedModel,
		EventStreamEnvelope: t.EventStreamEnvelope,
		Location:            t.Location,
		ResponseEnvelope:    t.ResponseEnvelope,
		UsedInUnion:         t.UsedInUnion,
		UsedInRequest:       t.UsedInRequest,
		UsedInResponse:      t.UsedInResponse,
		UsedInWebhook:       t.UsedInWebhook,
		UsedInCallback:      t.UsedInCallback,
		UsedInSecurity:      t.UsedInSecurity,
		Reference:           t.Reference,
		Registered:          t.Registered,
		Validations:         t.Validations,
		CachedEnumNames:     t.CachedEnumNames,
		OriginalNameFrozen:  t.OriginalNameFrozen,
		EventStreamSentinel: t.EventStreamSentinel,
	}
}

func (t *TypeDef) FindFieldByName(fieldName string) *FieldDef {
	for _, current := range t.Walk() {
		if !current.IsCustomType() {
			continue
		}

		for _, field := range current.Fields {
			if field.Name == fieldName {
				return field
			}
		}
	}
	return nil
}

// getNameOrType returns a human-friendly representation of the TypeDef.
func (t TypeDef) getNameOrType() string {
	if t.Name != "" {
		return t.Name
	}

	return string(t.Type)
}

// Intended to be used for debugging, not for any other purpose - avoids circular references
func (t *TypeDef) JSON() map[string]any {
	id := ""
	if t.IsCustomType() && len(t.ContextStack) > 0 {
		id = t.GetRegistrationID()
	}

	ret := map[string]any{
		"ID":           id,
		"Type":         t.Type,
		"Name":         t.Name,
		"OriginalName": t.OriginalName,
		"Scope":        t.Scope,
	}

	if t.ContentMediaType != "" {
		ret["ContentMediaType"] = t.ContentMediaType
	}

	if t.ItemType != nil {
		ret["itemType"] = t.ItemType.JSON()
	}

	if t.Discriminator != nil {
		// TODO: Serialize discriminator
		ret["discriminator"] = t.Discriminator
	}

	if t.Enum != nil {
		enumRet := map[string]any{"type": t.Enum.Type.JSON(), "values": t.Enum.Values}
		if t.Enum.Open {
			enumRet["open"] = true
		}
		if len(t.Enum.Names) > 0 {
			enumRet["names"] = t.Enum.Names
		}
		if len(t.Enum.Descriptions) > 0 {
			enumRet["descriptions"] = t.Enum.Descriptions
		}
		ret["enum"] = enumRet
	}

	if len(t.Fields) > 0 {
		fields := make([]string, len(t.Fields))
		for i, f := range t.Fields {
			fields[i] = fmt.Sprintf("(%s, %s)", f.Name, f.Type.Type)
		}
		ret["Fields"] = fields
	}

	if len(t.AssociatedTypes) > 0 {
		associatedTypes := make([]string, len(t.AssociatedTypes))
		for i, at := range t.AssociatedTypes {
			associatedTypes[i] = fmt.Sprintf("(%s, %s)", at.Type, at.Name)
		}
		ret["AssociatedTypes"] = associatedTypes
	}

	return ret
}

func (t *TypeDef) GetLocationNode() *yaml.Node {
	if t.Location == nil {
		return nil
	}

	return t.Location.Node
}
