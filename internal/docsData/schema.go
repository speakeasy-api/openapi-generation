// The schema chunk represents a single piece of data. Schema chunks are
// primarily used for request/response bodies, but can also be used for other
// purposes. The format for a schema chunk is similar to a TypeDef, but there
// are some key differences, described below

package docsData

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type ValueType string

// Most of these match the types in TypeDef 1:1, but there are a few differences
const (
	StringType         ValueType = "string"
	DateStringType     ValueType = "date"
	DateTimeStringType ValueType = "date-time"
	BooleanType        ValueType = "boolean"
	NumberType         ValueType = "number"
	IntegerType        ValueType = "integer"
	Int32Type          ValueType = "int32"
	Float32Type        ValueType = "float32"
	DecimalType        ValueType = "decimal"
	BigIntType         ValueType = "bigint"
	ObjectType         ValueType = "object"
	ArrayType          ValueType = "array"
	SetType            ValueType = "set"
	MapType            ValueType = "map"
	EventStreamType    ValueType = "event-stream"
	JsonLType          ValueType = "jsonl"
	UnionType          ValueType = "union"
	EnumType           ValueType = "enum"

	// We set null as a separate type and turn any type that is nullable into a
	// union of that type and a null type. This way, null shows up like any other
	// type in the UI
	NullType ValueType = "null"

	// DataTypeBytes, DataTypeRequestStream, DataTypeResponseStream are all
	// combined into this binary type here
	BinaryType ValueType = "binary"
	AnyType    ValueType = "any"

	// This type is used to create a layer of indirection such that objects and
	// arrays are never nested directly. This way we can serialize circular
	// references easily, and also more give the UI more flexibility when deciding
	// how to display deeply nested data. This type is the primary difference with
	// TypeDefs, since TypeDefs have no such concept.
	ChunkType ValueType = "chunk"
)

type SchemaValue interface {
	GetValueType() ValueType
}

// Nullable values

// If a given type is nullable, we convert the type to a union type containing
// the given type and a null value
func NewNullableValue(valueType SchemaValue, typeDef *ast.TypeDef) *UnionValue {
	var description *string
	if typeDef.Comments != nil {
		description = &typeDef.Comments.Description
	}
	unionValue := &UnionValue{
		Type:        UnionType,
		Description: description,
		Examples:    GetFormattedExamples(typeDef.Examples),
	}
	nullValue := &PrimitiveValue{
		Type:        NullType,
		Description: nil,
		Examples:    []string{},
	}
	unionValue.Values = append(unionValue.Values, valueType)
	unionValue.Values = append(unionValue.Values, nullValue)
	return unionValue
}

// Primitive values

type PrimitiveValue struct {
	Type        ValueType `json:"type"`
	Description *string   `json:"description"`
	Examples    []string  `json:"examples"`
	Default     *string   `json:"defaultValue"`
	MinLength   *int64    `json:"minLength"`
	MaxLength   *int64    `json:"maxLength"`
	Minimum     *float64  `json:"minimum"`
	Maximum     *float64  `json:"maximum"`
	Pattern     *string   `json:"pattern"`
}

func (p *PrimitiveValue) GetValueType() ValueType {
	return p.Type
}

func NewPrimitiveValue(primitiveType ValueType, typeDef *ast.TypeDef, defaultValue *string) *PrimitiveValue {
	var description *string
	if typeDef.Comments != nil {
		description = &typeDef.Comments.Description
	}
	p := &PrimitiveValue{
		Type:        primitiveType,
		Description: description,
		Examples:    GetFormattedExamples(typeDef.Examples),
		Default:     defaultValue,
	}
	if typeDef.Validations != nil {
		p.MinLength = typeDef.Validations.MinLength
		p.MaxLength = typeDef.Validations.MaxLength
		p.Minimum = typeDef.Validations.Minimum
		p.Maximum = typeDef.Validations.Maximum
		p.Pattern = typeDef.Validations.Pattern
	}
	return p
}

// Enum values

type EnumValue struct {
	Type        ValueType `json:"type"`
	Description *string   `json:"description"`
	Examples    []string  `json:"examples"`
	Values      []string  `json:"values"`
	Default     *string   `json:"defaultValue"`
}

func (e *EnumValue) GetValueType() ValueType {
	return e.Type
}

func NewEnumValue(values []string, typeDef *ast.TypeDef, isNullable bool, defaultValue *string) *EnumValue {
	// Note: enum values can't have a description according to the JSON Schema
	// spec itself, so we only have to worry about the top-level description on
	// the enum as a whole
	var description *string
	if typeDef.Comments != nil {
		description = &typeDef.Comments.Description
	}

	// TODO: need to add real null values if it's nullable? Need to make sure the
	// UI doesn't turn it into `'null'`
	return &EnumValue{
		Type:        EnumType,
		Examples:    GetFormattedExamples(typeDef.Examples),
		Values:      values,
		Description: description,
		Default:     defaultValue,
	}
}

// Chunk indirection type

type ChunkValue struct {
	Type    ValueType `json:"type"`
	ChunkId string    `json:"chunkId"`
}

func (c *ChunkValue) GetValueType() ValueType {
	return c.Type
}

func NewChunkValue(chunkId string) *ChunkValue {
	return &ChunkValue{
		Type:    ChunkType,
		ChunkId: chunkId,
	}
}

// Object values (can only be top-level)

type ObjectValue struct {
	Type        ValueType              `json:"type"`
	Description *string                `json:"description"`
	Examples    []string               `json:"examples"`
	Properties  map[string]SchemaValue `json:"properties"`
	Required    []string               `json:"required"`
	Deprecated  bool                   `json:"deprecated"`
	Name        string                 `json:"name"`
	Default     *string                `json:"defaultValue"`
}

func (o *ObjectValue) GetValueType() ValueType {
	return o.Type
}

func NewObjectValue(name string, typeDef *ast.TypeDef, defaultValue *string) *ObjectValue {
	var description *string
	if typeDef.Comments != nil {
		description = &typeDef.Comments.Description
	}
	return &ObjectValue{
		Type:        ObjectType,
		Description: description,
		Examples:    GetFormattedExamples(typeDef.Examples),
		Properties:  make(map[string]SchemaValue),
		Required:    make([]string, 0),
		Deprecated:  typeDef.Comments != nil && typeDef.Comments.Deprecated,
		Name:        name,
	}
}

func (o *ObjectValue) ParseObjectTypeDef(docs *Docs, typeDef *ast.TypeDef) error {
	for _, field := range typeDef.Fields {
		// Sometimes we have comments defined on the field but not the type for the
		// field. Since we combine these two objects into one in docs data, we can
		// just set the type comments to the field comments to preserve them.
		if field.Type.Comments == nil {
			field.Type.Comments = field.Comments
		}
		serializedDefault := GetSerializedDefault(field)
		fieldType, err := ParseTypeDef(docs, field.Type, field.Nullable, serializedDefault)
		if err != nil {
			return err
		}
		o.Properties[field.Name] = fieldType
		if !field.Optional {
			o.Required = append(o.Required, field.Name)
		}
		o.Deprecated = field.Comments != nil && field.Comments.Deprecated
	}
	return nil
}

// Array values (can only be top-level)

type ArrayValue struct {
	Type        ValueType   `json:"type"`
	Description *string     `json:"description"`
	Examples    []string    `json:"examples"`
	Items       SchemaValue `json:"items"`
	Default     *string     `json:"defaultValue"`
	MinItems    *int64      `json:"minItems"`
	MaxItems    *int64      `json:"maxItems"`
}

func (a *ArrayValue) GetValueType() ValueType {
	return a.Type
}

func NewArrayValue(arrayType ValueType, typeDef *ast.TypeDef, defaultValue *string) *ArrayValue {
	var description *string
	if typeDef.Comments != nil {
		description = &typeDef.Comments.Description
	}
	return &ArrayValue{
		Type:        arrayType,
		Description: description,
		Examples:    GetFormattedExamples(typeDef.Examples),
	}
}

func (a *ArrayValue) ParseArrayTypeDef(docs *Docs, typeDef *ast.TypeDef) error {
	if typeDef.ItemType == nil {
		return errors.New("array type def must have an item type")
	}

	items, err := ParseTypeDef(docs, typeDef.ItemType, typeDef.ContainsNull, nil)
	if err != nil {
		return err
	}

	a.Items = items
	if typeDef.Validations != nil {
		a.MinItems = typeDef.Validations.MinItems
		a.MaxItems = typeDef.Validations.MaxItems
	}
	return nil
}

// Union values (can only be top-level)

type UnionValue struct {
	Type        ValueType     `json:"type"`
	Description *string       `json:"description"`
	Examples    []string      `json:"examples"`
	Values      []SchemaValue `json:"values"`
	Default     *string       `json:"defaultValue"`
}

func (u *UnionValue) GetValueType() ValueType {
	return u.Type
}

func NewUnionValue(typeDef *ast.TypeDef, defaultValue *string) *UnionValue {
	var description *string
	if typeDef.Comments != nil {
		description = &typeDef.Comments.Description
	}
	return &UnionValue{
		Type:        UnionType,
		Description: description,
		Examples:    GetFormattedExamples(typeDef.Examples),
	}
}

func (u *UnionValue) ParseUnionTypeDef(docs *Docs, typeDef *ast.TypeDef) error {
	for _, item := range typeDef.AssociatedTypes {
		itemType, err := ParseTypeDef(docs, item, typeDef.ContainsNull, nil)
		if err != nil {
			return err
		}
		u.Values = append(u.Values, itemType)
	}
	return nil
}

// Top level

func ParseTypeDef(docs *Docs, typeDef *ast.TypeDef, isNullable bool, defaultValue *string) (SchemaValue, error) {
	switch typeDef.Type {
	// Primitive types in docs data mostly maps directly to the TypeDef type, but
	// there are a few cases where we deviate, such as streams, to display a type
	// that more closely matches the original spec and is more user friendly
	case ast.DataTypeAny:
		primitiveValue := NewPrimitiveValue(AnyType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeDateTime:
		primitiveValue := NewPrimitiveValue(DateTimeStringType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeDate:
		primitiveValue := NewPrimitiveValue(DateStringType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeString:
		primitiveValue := NewPrimitiveValue(StringType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeBoolean:
		primitiveValue := NewPrimitiveValue(BooleanType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeNumber:
		primitiveValue := NewPrimitiveValue(NumberType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeInteger:
		primitiveValue := NewPrimitiveValue(IntegerType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeInt32:
		primitiveValue := NewPrimitiveValue(Int32Type, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeBigInt:
		primitiveValue := NewPrimitiveValue(BigIntType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeFloat32:
		primitiveValue := NewPrimitiveValue(Float32Type, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeDecimal:
		primitiveValue := NewPrimitiveValue(DecimalType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeBytes:
		primitiveValue := NewPrimitiveValue(BinaryType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeRequestStream:
		primitiveValue := NewPrimitiveValue(BinaryType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil
	case ast.DataTypeResponseStream:
		primitiveValue := NewPrimitiveValue(BinaryType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return primitiveValue, nil

	case ast.DataTypeEnum:
		values := make([]string, len(typeDef.Enum.Values))
		copy(values, typeDef.Enum.Values)
		return NewEnumValue(values, typeDef, isNullable, defaultValue), nil

	case ast.DataTypeUnion:
		// Note: handling nullable values is done in the SerializeSchema call
		chunkId, err := SerializeSchema(docs, typeDef, isNullable, defaultValue)
		if err != nil {
			return nil, err
		}
		return NewChunkValue(*chunkId), nil

	// Error is an alias for a class, and the generator codebase itself treats
	// these two types as the same thing. We do the same here
	case ast.DataTypeError:
		fallthrough
	case ast.DataTypeClass:
		// Note: handling nullable values is done in the SerializeSchema call
		chunkId, err := SerializeSchema(docs, typeDef, isNullable, defaultValue)
		if err != nil {
			return nil, err
		}
		return NewChunkValue(*chunkId), nil

	// We use array logic for sets, maps, event streams, and jsonl, but set a
	// different root type so that we can still show this in the UI
	case ast.DataTypeEventStream:
		fallthrough
	case ast.DataTypeJsonL:
		fallthrough
	case ast.DataTypeSet:
		fallthrough
	case ast.DataTypeMap:
		fallthrough
	case ast.DataTypeArray:
		// Note: handling nullable values is done in the SerializeSchema call
		chunkId, err := SerializeSchema(docs, typeDef, isNullable, defaultValue)
		if err != nil {
			return nil, err
		}
		return NewChunkValue(*chunkId), nil

	// The request and response types are used for HTTPMetadata, which is only
	// included in error cases. To handle this in docs, we convert them into an
	// `any` type to represent that we don't know what they do
	case ast.DataTypeRequest:
		fallthrough
	case ast.DataTypeResponse:
		primitiveValue := NewPrimitiveValue(AnyType, typeDef, defaultValue)
		if isNullable {
			return NewNullableValue(primitiveValue, typeDef), nil
		}
		return NewPrimitiveValue(AnyType, typeDef, defaultValue), nil

	// This shouldn't be possible, since we exhaustively handle all cases, but we
	// leave it here in case a new type is added and this code isn't updated.
	default:
		return nil, fmt.Errorf("unknown schema type: %s", typeDef.Type)
	}
}

func getName(typeDef *ast.TypeDef) string {
	if typeDef.Name != "" {
		return typeDef.Name
	}
	return ""
}

func getBaseSlug(typeDef *ast.TypeDef) string {
	// We don't generate slugs for operations, callbacks, webhooks, or responses,
	// since they will never be unrolled into a separate navigable chunk
	if typeDef.Scope == ast.ScopeOperations || typeDef.Scope == ast.ScopeCallbacks || typeDef.Scope == ast.ScopeWebhooks || typeDef.Scope == "" {
		return ""
	}

	name := typeDef.GetRegistrationID()
	pairs := strings.Fields(name)
	parsedNameParts := []string{}
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")

		// If we don't have a key:value pair, use the single part as the entire value
		if len(parts) != 2 {
			parts = []string{"", pair}
		}

		// Boolean name's aren't useful, so convert them to something more readable
		switch parts[1] {
		case "true":
			parts[1] = "is-" + parts[0]
		case "false":
			parts[1] = "not-" + parts[0]
		}

		if parts[0] == "refType" {
			// Shouldn't be possible, but we should check just in case
			if parts[1] != "Schemas" {
				panic("Unexpected refType: " + parts[1])
			}
			// If this is a schema, we don't include a segment here because we
			// prefix it with "schemas/" below
			continue
		}

		// Remove slashes that would create an unexpected path segment
		escapedName := strings.ReplaceAll(parts[1], "/", "-")
		parsedNameParts = append(parsedNameParts, escapedName)
	}

	return "schemas/" + strings.ToLower(strings.Join(parsedNameParts, "/"))
}

// Some types don't have their own scope, and thus we can't use the registration
// ID directly. We wrap these types in this case in a manner similar to
// GetFullyQualifiedName() (which solves the same problem, but doesn't produce
// the syntax we want)
func getSlug(typeDef *ast.TypeDef) string {
	switch typeDef.Type {
	case ast.DataTypeUnion:
		fallthrough
	case ast.DataTypeEnum:
		fallthrough
	case ast.DataTypeError:
		fallthrough
	case ast.DataTypeClass:
		return getBaseSlug(typeDef)
	case ast.DataTypeEventStream:
		baseSlug := getBaseSlug(typeDef.ItemType)
		if baseSlug == "" {
			return ""
		}
		return baseSlug + "/event-stream"
	case ast.DataTypeJsonL:
		baseSlug := getBaseSlug(typeDef.ItemType)
		if baseSlug == "" {
			return ""
		}
		return baseSlug + "/jsonl"
	case ast.DataTypeSet:
		baseSlug := getBaseSlug(typeDef.ItemType)
		if baseSlug == "" {
			return ""
		}
		return baseSlug + "/set"
	case ast.DataTypeArray:
		baseSlug := getBaseSlug(typeDef.ItemType)
		if baseSlug == "" {
			return ""
		}
		return baseSlug + "/array"
	case ast.DataTypeMap:
		baseSlug := getBaseSlug(typeDef.ItemType)
		if baseSlug == "" {
			return ""
		}
		return baseSlug + "/map"
	default:
		// Primitives and such don't have a registration ID, and thus we can't
		// use them as a slug. Plus it just wouldn't make sense for a primitive to
		// have a slug in the UI anyways.
		return ""
	}
}

type SchemaData struct {
	Name  string      `json:"name"`
	Value SchemaValue `json:"value"`
}

type SchemaChunk struct {
	ID        string     `json:"id"`
	Slug      string     `json:"slug"`
	ChunkData SchemaData `json:"chunkData"`
	ChunkType string     `json:"chunkType"`
}

func SerializeSchema(docs *Docs, typeDef *ast.TypeDef, isNullable bool, defaultValue *string) (*string, error) {
	var chunk SchemaChunk

	// Check if there is a scope for this typeDef, aka it's a component. If so,
	// look to see if we have already serialized it. If we have, return the
	// existing ID without further traversing the type. This prevents circular
	// references from causing infinite recursion.
	slug := getSlug(typeDef)
	if slug != "" {
		cachedId := docs.GetCachedIdFromSlug(slug)
		if cachedId != "" {
			return &cachedId, nil
		}
	}
	chunk.ID = docs.GenerateID()
	if slug != "" {
		chunk.Slug = slug
		docs.RegisterSlug(chunk.ID, slug)
	}
	chunk.ChunkType = "schema"

	switch typeDef.Type {
	case ast.DataTypeError:
		fallthrough
	case ast.DataTypeClass:
		objectValue := NewObjectValue(getName(typeDef), typeDef, defaultValue)
		err := objectValue.ParseObjectTypeDef(docs, typeDef)
		if err != nil {
			return nil, err
		}
		if isNullable {
			nullableValue := NewNullableValue(objectValue, typeDef)
			chunk.ChunkData = SchemaData{
				Name:  getName(typeDef),
				Value: nullableValue,
			}
		} else {
			chunk.ChunkData = SchemaData{
				Name:  getName(typeDef),
				Value: objectValue,
			}
		}
	case ast.DataTypeSet:
		fallthrough
	case ast.DataTypeMap:
		fallthrough
	case ast.DataTypeEventStream:
		fallthrough
	case ast.DataTypeJsonL:
		fallthrough
	case ast.DataTypeArray:
		var arrayType ValueType
		switch typeDef.Type {
		case ast.DataTypeSet:
			arrayType = SetType
		case ast.DataTypeMap:
			arrayType = MapType
		case ast.DataTypeEventStream:
			arrayType = EventStreamType
		case ast.DataTypeJsonL:
			arrayType = JsonLType
		default:
			arrayType = ArrayType
		}
		arrayValue := NewArrayValue(arrayType, typeDef, defaultValue)
		err := arrayValue.ParseArrayTypeDef(docs, typeDef)
		if err != nil {
			return nil, err
		}
		if isNullable {
			nullableValue := NewNullableValue(arrayValue, typeDef)
			chunk.ChunkData = SchemaData{
				Name:  getName(typeDef),
				Value: nullableValue,
			}
		} else {
			chunk.ChunkData = SchemaData{
				Name:  getName(typeDef),
				Value: arrayValue,
			}
		}
	case ast.DataTypeUnion:
		unionValue := NewUnionValue(typeDef, defaultValue)
		err := unionValue.ParseUnionTypeDef(docs, typeDef)
		if err != nil {
			return nil, err
		}
		if isNullable {
			unionValue.Values = append(unionValue.Values, NewPrimitiveValue(NullType, typeDef, nil))
		}
		chunk.ChunkData = SchemaData{
			Name:  getName(typeDef),
			Value: unionValue,
		}
	default:
		chunkData, err := ParseTypeDef(docs, typeDef, isNullable, nil)
		if err != nil {
			return nil, err
		}
		chunk.ChunkData = SchemaData{
			Name:  getName(typeDef),
			Value: chunkData,
		}
	}

	b, err := json.Marshal(chunk)
	if err != nil {
		return nil, err
	}
	docs.SaveSerializedChunk(chunk.ID, string(b))
	return &chunk.ID, nil
}
