package ast

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi/jsonpointer"
)

// FieldAddOptions contains options for adding fields to a Fields collection
type FieldAddOptions struct {
	MaintainOriginalOrder bool
	Sanitize              bool
}

// DefaultFieldAddOptions returns the default options for adding fields
func DefaultFieldAddOptions() FieldAddOptions {
	return FieldAddOptions{
		MaintainOriginalOrder: false,
		Sanitize:              true,
	}
}

// FieldDef represents a field in a TypeDef if it is a class
type FieldDef struct {
	Name                   string               `yaml:",omitempty"` // The name of the field within the class
	OriginalName           string               `yaml:",omitempty"` // The name of a field as it appears in the source document if it was a property in an object schema
	Type                   *TypeDef             `yaml:",omitempty"` // The type of the field
	Comments               *Comment             `yaml:",omitempty"` // The comments associated with the field
	Annotations            Annotations          `yaml:",omitempty"` // Any annotations applied to the field
	Nullable               bool                 `yaml:",omitempty"` // Whether the field is nullable
	Optional               bool                 `yaml:",omitempty"` // In the case of an object property, indicates whether the field is non-required. Otherwise, matches Nullable
	SerializationMethod    *SerializationMethod `yaml:",omitempty"` // The serialization method to use for the field if any
	ErrorMessage           bool                 `yaml:",omitempty"` // Whether or not the field is an error message when used in an error type
	Const                  *AnyValue            `yaml:",omitempty"` // The constant value of the field if any
	Default                *AnyValue            `yaml:",omitempty"` // The default value of the field if any
	IsAdditionalProperties bool                 `yaml:",omitempty"` // Whether or not the field is an additional properties field
	IsResponseHeaders      bool                 `yaml:",omitempty"` // Whether or not the field is the response headers field
	IsResponseMetadata     bool                 `yaml:",omitempty"` // Whether or not the field is HTTP response metadata (ContentType/StatusCode/RawResponse)
	ParameterIndex         *int                 `yaml:",omitempty"` // The index of the parameter in the operation if the field is a request parameter
}

var _ jsonpointer.NavigableNoder = (*FieldDef)(nil)

// Returns this FieldDef or any of its children when the given entity name
// matches the x-speakeasy-entity configuration.
func (f *FieldDef) FindEntityFieldDef(entityName string) *FieldDef {
	if f == nil || f.Type == nil {
		return nil
	}

	if f.Type.HasEntityName(entityName) {
		return f
	}

	for _, fieldDef := range f.Type.Fields {
		foundFieldDef := fieldDef.Type.FindEntityTypeDef(entityName)

		if foundFieldDef != nil {
			return fieldDef
		}
	}

	if f.Type.ItemType != nil {
		foundTypeDef := f.Type.ItemType.FindEntityTypeDef(entityName)

		if foundTypeDef != nil {
			// TODO: This logic should return any nested FieldDef that has the
			// entity, but for now avoiding excessive complexity with FieldDef
			// versus TypeDef given logic calling this method does not need full
			// correctness at the moment.
			return f
		}
	}

	for _, typeDef := range f.Type.AssociatedTypes {
		foundTypeDef := typeDef.FindEntityTypeDef(entityName)

		if foundTypeDef != nil {
			// TODO: This logic should return any nested FieldDef that has the
			// entity, but for now avoiding excessive complexity with FieldDef
			// versus TypeDef given logic calling this method does not need full
			// correctness at the moment.
			return f
		}
	}

	return nil
}

// Initial use case of `GetID` was to avoid rendering
// the same field multiple times in a code sample
// Note: that this field may be duplicated in the AST in the sense it
// has the same ID
// eg { a: A, b: { a: A } } (field a is duplicated)
func (f *FieldDef) GetID() string {
	builder := strings.Builder{}
	builder.WriteString("fieldName:")
	builder.WriteString(f.Name)

	for _, annotation := range f.Annotations {
		builder.WriteString(" ")
		builder.WriteString("annotation:")
		builder.WriteString(string(annotation.Type()))
	}

	builder.WriteString(" ")

	t := f.Type
	if t.ItemType != nil {
		builder.WriteString("type:")
		builder.WriteString(string(t.Type))
		t = t.ItemType
		builder.WriteString(" ")
	}
	builder.WriteString(t.GetRegistrationIDOrType())
	return builder.String()
}

// Clone creates a deep copy of the FieldDef
func (f *FieldDef) Clone() *FieldDef {
	if f == nil {
		return nil
	}

	return f.clone(nil, nil)
}

// clone creates a deep copy of the FieldDef, tracking visited FieldDefs to
// avoid infinite recursion on circular references.
func (f *FieldDef) clone(fieldDefVisited map[*FieldDef]*FieldDef, typeDefVisited map[*TypeDef]*TypeDef) *FieldDef {
	if f == nil {
		return nil
	}

	if fieldDefVisited == nil {
		fieldDefVisited = make(map[*FieldDef]*FieldDef)
	}

	if existing, ok := fieldDefVisited[f]; ok {
		return existing
	}

	if typeDefVisited == nil {
		typeDefVisited = make(map[*TypeDef]*TypeDef)
	}

	cloned := &FieldDef{
		Annotations:            f.Annotations.Clone(),
		Comments:               f.Comments.Clone(),
		Const:                  f.Const.Clone(),
		Default:                f.Default.Clone(),
		ErrorMessage:           f.ErrorMessage,
		IsAdditionalProperties: f.IsAdditionalProperties,
		IsResponseHeaders:      f.IsResponseHeaders,
		IsResponseMetadata:     f.IsResponseMetadata,
		Name:                   f.Name,
		Nullable:               f.Nullable,
		Optional:               f.Optional,
		OriginalName:           f.OriginalName,
		ParameterIndex:         clonePtr(f.ParameterIndex),
		SerializationMethod:    clonePtr(f.SerializationMethod),
	}

	// Ensure we mark the cloned FieldDef as visited before cloning children to
	// avoid infinite recursion.
	fieldDefVisited[f] = cloned

	cloned.Type = f.Type.clone(fieldDefVisited, typeDefVisited)

	return cloned
}

// deepClone creates a fully independent copy using stack-based cycle detection.
// See TypeDef.DeepClone for details on why this exists.
func (f *FieldDef) deepClone(fieldDefStack map[*FieldDef]*FieldDef, typeDefStack map[*TypeDef]*TypeDef) *FieldDef {
	if f == nil {
		return nil
	}

	if fieldDefStack == nil {
		fieldDefStack = make(map[*FieldDef]*FieldDef)
	}

	// Only return cached clone for cycles (back-edges on the current stack).
	if existing, ok := fieldDefStack[f]; ok {
		return existing
	}

	if typeDefStack == nil {
		typeDefStack = make(map[*TypeDef]*TypeDef)
	}

	cloned := &FieldDef{
		Annotations:            f.Annotations.Clone(),
		Comments:               f.Comments.Clone(),
		Const:                  f.Const.Clone(),
		Default:                f.Default.Clone(),
		ErrorMessage:           f.ErrorMessage,
		IsAdditionalProperties: f.IsAdditionalProperties,
		IsResponseHeaders:      f.IsResponseHeaders,
		IsResponseMetadata:     f.IsResponseMetadata,
		Name:                   f.Name,
		Nullable:               f.Nullable,
		Optional:               f.Optional,
		OriginalName:           f.OriginalName,
		ParameterIndex:         clonePtr(f.ParameterIndex),
		SerializationMethod:    clonePtr(f.SerializationMethod),
	}

	// Push onto stack for cycle detection.
	fieldDefStack[f] = cloned

	cloned.Type = f.Type.deepClone(fieldDefStack, typeDefStack)

	// Pop from stack.
	delete(fieldDefStack, f)

	return cloned
}

func (f *FieldDef) Match(matchers Matchers) error {
	if matchers.FieldDef != nil {
		return matchers.FieldDef(f)
	}

	return nil
}

func (def *FieldDef) IsEqual(other *FieldDef) bool {
	if def == nil && other == nil || def == other {
		return true
	}
	if def == nil && other != nil || other == nil && def != nil {
		return false
	}
	if def.Name != other.Name {
		return false
	}
	if def.Type.IsEqual(other.Type) != nil {
		return false
	}
	if def.Optional != other.Optional {
		return false
	}
	if def.SerializationMethod == nil && other.SerializationMethod != nil || def.SerializationMethod != nil && other.SerializationMethod == nil {
		return false
	}
	if def.SerializationMethod != nil && other.SerializationMethod != nil && *def.SerializationMethod != *other.SerializationMethod {
		return false
	}

	if len(def.Annotations) != len(other.Annotations) {
		return false
	}
	for i := range def.Annotations {
		if !def.Annotations[i].IsEqual(other.Annotations[i]) {
			return false
		}
	}
	return true
}

// Returns true if the FieldDef is equal to given FieldDef for Terraform usage.
func (f *FieldDef) IsTerraformEqual(other *FieldDef) bool {
	if f == nil || other == nil {
		return f == other
	}

	if f == other {
		return true
	}

	if f.Nullable != other.Nullable {
		return false
	}

	if f.Optional != other.Optional {
		return false
	}

	if !f.Type.IsTerraformEqual(other.Type) {
		return false
	}

	return true
}

func (def FieldDef) GetNavigableNode() (any, error) {
	return def.Type, nil
}

// Returns the ParamAnnotation from the FieldDef if present.
func (f *FieldDef) getParamAnnotation() *ParamAnnotation {
	if f == nil || f.Annotations == nil {
		return nil
	}

	return f.Annotations.GetParam()
}

type Fields []*FieldDef

// Clone creates a deep copy of the Fields
func (f Fields) Clone() Fields {
	if f == nil {
		return nil
	}

	return f.clone(nil, nil)
}

// clone creates a deep copy of the Fields, tracking visited FieldDefs and
// TypeDefs to avoid infinite recursion on circular references.
func (f Fields) clone(fieldDefVisited map[*FieldDef]*FieldDef, typeDefVisited map[*TypeDef]*TypeDef) Fields {
	if f == nil {
		return nil
	}

	if fieldDefVisited == nil {
		fieldDefVisited = make(map[*FieldDef]*FieldDef)
	}

	if typeDefVisited == nil {
		typeDefVisited = make(map[*TypeDef]*TypeDef)
	}

	allVisited := true

	for _, field := range f {
		if _, ok := fieldDefVisited[field]; !ok {
			allVisited = false
			break
		}
	}

	if allVisited {
		return f
	}

	cloned := make(Fields, len(f))

	for i, field := range f {
		cloned[i] = field.clone(fieldDefVisited, typeDefVisited)
	}

	return cloned
}

// deepClone creates fully independent copies using stack-based cycle detection.
// See TypeDef.DeepClone for details on why this exists.
func (f Fields) deepClone(fieldDefStack map[*FieldDef]*FieldDef, typeDefStack map[*TypeDef]*TypeDef) Fields {
	if f == nil {
		return nil
	}

	if fieldDefStack == nil {
		fieldDefStack = make(map[*FieldDef]*FieldDef)
	}

	if typeDefStack == nil {
		typeDefStack = make(map[*TypeDef]*TypeDef)
	}

	cloned := make(Fields, len(f))

	for i, field := range f {
		cloned[i] = field.deepClone(fieldDefStack, typeDefStack)
	}

	return cloned
}

// TODO: add tests for panics
// TODO: add tests for field renaming
func (f Fields) MustAddField(field *FieldDef, maintainOriginalOrder bool) Fields {
	opts := DefaultFieldAddOptions()
	opts.MaintainOriginalOrder = maintainOriginalOrder
	return f.MustAddFieldWithOptions(field, opts)
}

func (f Fields) MustAddFieldWithOptions(field *FieldDef, opts FieldAddOptions) Fields {
	if field == nil {
		panic("field is nil")
	}
	if field.Name == "" {
		panic("field name is empty")
	}

	for {
		if exists, _ := FieldExists(f, field, opts.Sanitize); exists {
			field.Name = utils.IncrementName(field.Name)
		} else {
			break
		}
	}

	ff, err := f.AddFieldWithOptions(field, opts)
	if err != nil {
		panic(err)
	}

	return ff
}

func (f Fields) AddField(field *FieldDef, maintainOriginalOrder bool) (Fields, error) {
	opts := DefaultFieldAddOptions()
	opts.MaintainOriginalOrder = maintainOriginalOrder
	return f.AddFieldWithOptions(field, opts)
}

func (f Fields) AddFieldWithOptions(field *FieldDef, opts FieldAddOptions) (Fields, error) {
	if field == nil {
		return nil, errors.New("field is nil")
	}

	exists, _ := FieldExists(f, field, opts.Sanitize)
	if exists {
		return nil, errors.New("field already exists")
	}

	if opts.MaintainOriginalOrder {
		return append(f, field), nil
	} else {
		return utils.AppendSorted(f, field, func(i, j *FieldDef) bool {
			return i.Name >= j.Name
		}), nil
	}
}

// Returns the FieldDef with the given name, if it exists.
func (f Fields) GetField(name string) *FieldDef {
	for _, field := range f {
		if field.Name == name {
			return field
		}
	}

	return nil
}

// Returns true if the Fields is equal to given Fields for Terraform usage.
func (f Fields) IsTerraformEqual(other Fields) bool {
	if len(f) != len(other) {
		return false
	}

	for _, field := range f {
		otherField := other.GetField(field.Name)

		if !field.IsTerraformEqual(otherField) {
			return false
		}
	}

	return true
}

// IsTerraformSymbolEqual reports whether two Fields slices are structurally
// equal for symbol deduplication, using sanitized field name matching. Unlike
// IsTerraformEqual which uses exact name matching via GetField, this method
// accounts for casing/formatting differences by comparing SanitizeFieldName
// values. Only unidirectional matching is needed because lengths are equal and
// sanitized field names are unique within a single Fields slice.
func (f Fields) IsTerraformSymbolEqual(other Fields) bool {
	if len(f) != len(other) {
		return false
	}

	for _, field := range f {
		sanitizedName := SanitizeFieldName(field.Name)
		var otherField *FieldDef

		for _, of := range other {
			if SanitizeFieldName(of.Name) == sanitizedName {
				otherField = of
				break
			}
		}

		if otherField == nil {
			return false
		}

		if field.Optional != otherField.Optional {
			return false
		}

		if field.Nullable != otherField.Nullable {
			return false
		}

		if !field.Type.IsTerraformSymbolEqual(otherField.Type) {
			return false
		}
	}

	return true
}

func (f Fields) ContainsAdditionalProperties() bool {
	for _, field := range f {
		if field.IsAdditionalProperties {
			return true
		}
	}

	return false
}

func (f Fields) Difference(fields Fields) Fields {
	diff := make(map[string]*FieldDef)

	// Populate map with elements from f and mark them as seen
	for _, item := range f {
		if _, exists := diff[item.Name]; !exists {
			diff[item.Name] = item
		}
	}

	// Toggle the presence in the map for fields
	// If an item is already in diff (from f), it's a duplicate and removed.
	// If not, it is added.
	for _, item := range fields {
		if _, exists := diff[item.Name]; exists {
			delete(diff, item.Name)
		} else {
			diff[item.Name] = item
		}
	}

	// Build the result in the original order as much as possible
	var result Fields
	seen := make(map[string]bool)
	// Check from f first
	for _, item := range f {
		if _, exists := diff[item.Name]; exists && !seen[item.Name] {
			result = append(result, item)
			seen[item.Name] = true
		}
	}
	// Then check from fields
	for _, item := range fields {
		if _, exists := diff[item.Name]; exists && !seen[item.Name] {
			result = append(result, item)
			seen[item.Name] = true
		}
	}

	return result
}

func (f Fields) GetFieldNames() string {
	names := make([]string, 0, len(f))

	for _, field := range f {
		names = append(names, field.Name)
	}

	return strings.Join(names, ", ")
}

func (f Fields) EnsureField(field *FieldDef, maintainOriginalOrder bool) Fields {
	opts := DefaultFieldAddOptions()
	opts.MaintainOriginalOrder = maintainOriginalOrder
	return f.EnsureFieldWithOptions(field, opts)
}

func (f Fields) EnsureFieldWithOptions(field *FieldDef, opts FieldAddOptions) Fields {
	if field == nil {
		panic("field is nil")
	}
	if field.Name == "" {
		panic("field name is empty")
	}

	if exists, i := FieldExists(f, field, opts.Sanitize); exists && f[i].Type.IsEqual(field.Type) == nil {
		return f
	}

	return f.MustAddFieldWithOptions(field, opts)
}

func (f Fields) PrintFields() {
	fieldNames := make([]string, 0, len(f))

	for _, field := range f {
		fieldNames = append(fieldNames, field.Name)
	}

	if len(fieldNames) == 0 {
		return
	}

	fmt.Printf("%p: %v\n", &f, fieldNames)
}

func FieldExists(ff Fields, field *FieldDef, sanitize bool) (bool, int) {
	for i, f := range ff {
		var existingName, newName string

		// If either field is a parameter, always use sanitized comparison
		// to ensure consistent naming and avoid conflicts
		useSanitized := sanitize || f.Annotations.Has(AnnotationTypeParam) || field.Annotations.Has(AnnotationTypeParam)

		if useSanitized {
			existingName = SanitizeFieldName(f.Name)
			newName = SanitizeFieldName(field.Name)
		} else {
			existingName = f.Name
			newName = field.Name
		}

		if existingName == newName {
			return true, i
		}
	}

	return false, -1
}

func sanitizeFieldNameWithCache() func(name string) string {
	cache := sync.Map{}

	return func(original string) string {
		cacheHit, ok := cache.Load(original)
		if ok {
			return cacheHit.(string)
		}

		name := sanitization.SanitizeName(original)
		name = strcase.ToGoPascal(name)
		cache.Store(original, name)
		return name
	}
}

var SanitizeFieldName = sanitizeFieldNameWithCache()
