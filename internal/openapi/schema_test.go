package openapi

import (
	"testing"

	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/values"
	"github.com/stretchr/testify/assert"
)

type testExtensions struct{}

func (t *testExtensions) IsExtensionMergable(extName string) bool {
	return false
}

func (t *testExtensions) IsExtensionIdentifying(extName string) bool {
	return false
}

func TestSchema_IsEmpty(t *testing.T) {
	type args struct {
		schema                          *oas3.Schema
		onlyConsiderTypeModifyingFields bool
	}
	tests := []struct {
		name      string
		args      args
		wantEmpty bool
	}{
		{
			name: "empty schema is empty",
			args: args{
				schema: &oas3.Schema{},
			},
			wantEmpty: true,
		},
		{
			name: "schema with title is not empty",
			args: args{
				schema: &oas3.Schema{
					Title: pointer.From("some title"),
				},
			},
			wantEmpty: false,
		},
		{
			name: "schema with title but ignored is empty",
			args: args{
				schema: &oas3.Schema{
					Title: pointer.From("some title"),
				},
				onlyConsiderTypeModifyingFields: true,
			},
			wantEmpty: true,
		},
		{
			name: "schema with type is not empty",
			args: args{
				schema: &oas3.Schema{
					Type: oas3.NewTypeFromArray([]oas3.SchemaType{"string"}),
				},
			},
			wantEmpty: false,
		},
		{
			name: "schema with type but ignored non-modifying fields is not empty",
			args: args{
				schema: &oas3.Schema{
					Type: oas3.NewTypeFromArray([]oas3.SchemaType{"string"}),
				},
				onlyConsiderTypeModifyingFields: true,
			},
			wantEmpty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantEmpty, IsEmpty(tt.args.schema, tt.args.onlyConsiderTypeModifyingFields, &testExtensions{}))
		})
	}
}

func TestSchema_IsEqual(t *testing.T) {
	type args struct {
		left  *oas3.JSONSchema[oas3.Referenceable]
		right *oas3.JSONSchema[oas3.Referenceable]
	}
	tests := []struct {
		name      string
		args      args
		wantEqual bool
	}{
		{
			name: "empty schemas are equal",
			args: args{
				left:  &oas3.JSONSchema[oas3.Referenceable]{},
				right: &oas3.JSONSchema[oas3.Referenceable]{},
			},
			wantEqual: true,
		},
		{
			name: "schemas with only references are equal",
			args: args{
				left:  oas3.NewJSONSchemaFromReference("#/components/schemas/SomeSchema"),
				right: oas3.NewJSONSchemaFromReference("#/components/schemas/SomeSchema"),
			},
			wantEqual: true,
		},
		{
			name: "schemas with the same types are equal",
			args: args{
				left: oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
					Type: oas3.NewTypeFromArray([]oas3.SchemaType{"string"}),
				}),
				right: oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
					Type: oas3.NewTypeFromArray([]oas3.SchemaType{"string"}),
				}),
			},
			wantEqual: true,
		},
		{
			name: "schemas with different types are not equal",
			args: args{
				left: oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
					Type: oas3.NewTypeFromArray([]oas3.SchemaType{"string"}),
				}),
				right: oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
					Type: oas3.NewTypeFromArray([]oas3.SchemaType{"integer"}),
				}),
			},
			wantEqual: false,
		},
		{
			name: "map and slice fields are equal if they are nil or empty",
			args: args{
				left: oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
					Enum:       nil,
					Extensions: extensions.New(),
				}),
				right: oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
					Enum:       []values.Value{},
					Extensions: nil,
				}),
			},
			wantEqual: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantEqual, tt.args.left.IsEqual(tt.args.right))
		})
	}
}
