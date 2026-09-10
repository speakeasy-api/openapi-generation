package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDiscriminatorMappings_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mappings DiscriminatorMappings
		test     func(t *testing.T, original, cloned DiscriminatorMappings)
	}{
		{
			name:     "nil DiscriminatorMappings",
			mappings: nil,
			test: func(t *testing.T, original, cloned DiscriminatorMappings) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:     "empty DiscriminatorMappings",
			mappings: DiscriminatorMappings{},
			test: func(t *testing.T, original, cloned DiscriminatorMappings) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				assert.Empty(t, cloned)
			},
		},
		{
			name: "DiscriminatorMappings with single mapping",
			mappings: DiscriminatorMappings{
				{
					Name: "mapping1",
					Type: &TypeDef{
						Name: "MappedType1",
						Type: DataTypeClass,
					},
				},
			},
			test: func(t *testing.T, original, cloned DiscriminatorMappings) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 1)

				for i := range original {
					assert.NotSame(t, original[i], cloned[i])
					assert.Equal(t, original[i].Name, cloned[i].Name)
					assert.NotSame(t, original[i].Type, cloned[i].Type)
					assert.Equal(t, original[i].Type.Name, cloned[i].Type.Name)
					assert.Equal(t, original[i].Type.Type, cloned[i].Type.Type)
				}
			},
		},
		{
			name: "DiscriminatorMappings with multiple mappings",
			mappings: DiscriminatorMappings{
				{
					Name: "mapping1",
					Type: &TypeDef{
						Name: "MappedType1",
						Type: DataTypeClass,
					},
				},
				{
					Name: "mapping2",
					Type: &TypeDef{
						Name: "MappedType2",
						Type: DataTypeEnum,
						Enum: &Enum{
							Type: &TypeDef{
								Type: DataTypeString,
							},
							Values: []string{"value1", "value2"},
						},
					},
				},
				{
					Name: "mapping3",
					Type: &TypeDef{
						Name: "MappedType3",
						Type: DataTypeUnion,
						AssociatedTypes: []*TypeDef{
							{
								Type: DataTypeString,
							},
							{
								Type: DataTypeInteger,
							},
						},
					},
				},
			},
			test: func(t *testing.T, original, cloned DiscriminatorMappings) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 3)

				for i := range original {
					assert.NotSame(t, original[i], cloned[i])
					assert.Equal(t, original[i].Name, cloned[i].Name)
					assert.NotSame(t, original[i].Type, cloned[i].Type)
					assert.Equal(t, original[i].Type.Name, cloned[i].Type.Name)
					assert.Equal(t, original[i].Type.Type, cloned[i].Type.Type)

					// Verify deep cloning of nested structures
					if original[i].Type.Enum != nil {
						require.NotNil(t, cloned[i].Type.Enum)
						assert.NotSame(t, original[i].Type.Enum, cloned[i].Type.Enum)
						assert.Equal(t, original[i].Type.Enum.Values, cloned[i].Type.Enum.Values)
					}

					if len(original[i].Type.AssociatedTypes) > 0 {
						require.Len(t, cloned[i].Type.AssociatedTypes, len(original[i].Type.AssociatedTypes))
						for j := range original[i].Type.AssociatedTypes {
							assert.NotSame(t, original[i].Type.AssociatedTypes[j], cloned[i].Type.AssociatedTypes[j])
							assert.Equal(t, original[i].Type.AssociatedTypes[j].Type, cloned[i].Type.AssociatedTypes[j].Type)
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.mappings.Clone()
			tt.test(t, tt.mappings, cloned)
		})
	}
}

func TestTypeDef_Clone(t *testing.T) {
	t.Parallel()

	recursiveTypeDef := &TypeDef{
		Name: "RecursiveType",
		Type: DataTypeClass,
		Fields: Fields{
			{
				Name: "self",
			},
		},
	}
	recursiveTypeDef.Fields[0].Type = recursiveTypeDef

	tests := []struct {
		name string
		td   *TypeDef
		test func(t *testing.T, original, cloned *TypeDef)
	}{
		{
			name: "nil TypeDef",
			td:   nil,
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "simple TypeDef with basic fields",
			td: &TypeDef{
				Name:                "TestType",
				OriginalName:        "OriginalTestType",
				Hash:                "hash123",
				Type:                DataTypeClass,
				Scope:               ScopeSDK,
				Format:              "email",
				ContentMediaType:    "application/json",
				ComplexAny:          true,
				Truncated:           true,
				Input:               true,
				Output:              false,
				ContainsNull:        true,
				IsNullableUnion:     false,
				IsComponent:         true,
				OutputLocation:      "/path/to/output",
				ResolvedModel:       "ResolvedModel",
				EventStreamEnvelope: true,
				ResponseEnvelope:    false,
				UsedInUnion:         true,
				Reference:           "#/components/schemas/Test",
				EventStreamSentinel: "sentinel",
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Name, cloned.Name)
				assert.Equal(t, original.OriginalName, cloned.OriginalName)
				assert.Equal(t, original.Hash, cloned.Hash)
				assert.Equal(t, original.Type, cloned.Type)
				assert.Equal(t, original.Scope, cloned.Scope)
				assert.Equal(t, original.Format, cloned.Format)
				assert.Equal(t, original.ContentMediaType, cloned.ContentMediaType)
				assert.Equal(t, original.ComplexAny, cloned.ComplexAny)
				assert.Equal(t, original.Truncated, cloned.Truncated)
				assert.Equal(t, original.Input, cloned.Input)
				assert.Equal(t, original.Output, cloned.Output)
				assert.Equal(t, original.ContainsNull, cloned.ContainsNull)
				assert.Equal(t, original.IsNullableUnion, cloned.IsNullableUnion)
				assert.Equal(t, original.IsComponent, cloned.IsComponent)
				assert.Equal(t, original.OutputLocation, cloned.OutputLocation)
				assert.Equal(t, original.ResolvedModel, cloned.ResolvedModel)
				assert.Equal(t, original.EventStreamEnvelope, cloned.EventStreamEnvelope)
				assert.Equal(t, original.ResponseEnvelope, cloned.ResponseEnvelope)
				assert.Equal(t, original.UsedInUnion, cloned.UsedInUnion)
				assert.Equal(t, original.Reference, cloned.Reference)
				assert.Equal(t, original.EventStreamSentinel, cloned.EventStreamSentinel)
			},
		},
		{
			name: "TypeDef with ContextStack",
			td: &TypeDef{
				Name: "TypeWithContext",
				ContextStack: ContextStack{
					{Type: ContextTypeOperation, Identifier: "createUser", Used: true},
					{Type: ContextTypeResponseBody, Identifier: "200", Used: false},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.ContextStack, len(original.ContextStack))
				// ContextStack should be deeply cloned
				assert.NotSame(t, &original.ContextStack, &cloned.ContextStack)
				for i := range original.ContextStack {
					assert.Equal(t, original.ContextStack[i].Type, cloned.ContextStack[i].Type)
					assert.Equal(t, original.ContextStack[i].Identifier, cloned.ContextStack[i].Identifier)
					assert.Equal(t, original.ContextStack[i].Used, cloned.ContextStack[i].Used)
				}
			},
		},
		{
			name: "TypeDef with ItemType",
			td: &TypeDef{
				Name: "ArrayType",
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Name: "StringItem",
					Type: DataTypeString,
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.ItemType)
				assert.NotSame(t, original.ItemType, cloned.ItemType)
				assert.Equal(t, original.ItemType.Name, cloned.ItemType.Name)
				assert.Equal(t, original.ItemType.Type, cloned.ItemType.Type)
			},
		},
		{
			name: "TypeDef with Fields",
			td: &TypeDef{
				Name: "ClassType",
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name:         "field1",
						OriginalName: "originalField1",
						Optional:     true,
						Nullable:     false,
						Type: &TypeDef{
							Type: DataTypeString,
						},
					},
					{
						Name:         "field2",
						OriginalName: "originalField2",
						Optional:     false,
						Nullable:     true,
						Type: &TypeDef{
							Type: DataTypeInteger,
						},
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.Fields, len(original.Fields))
				for i := range original.Fields {
					assert.NotSame(t, original.Fields[i], cloned.Fields[i])
					assert.Equal(t, original.Fields[i].Name, cloned.Fields[i].Name)
					assert.Equal(t, original.Fields[i].OriginalName, cloned.Fields[i].OriginalName)
					assert.Equal(t, original.Fields[i].Optional, cloned.Fields[i].Optional)
					assert.Equal(t, original.Fields[i].Nullable, cloned.Fields[i].Nullable)
					assert.NotSame(t, original.Fields[i].Type, cloned.Fields[i].Type)
					assert.Equal(t, original.Fields[i].Type.Type, cloned.Fields[i].Type.Type)
				}
			},
		},
		{
			name: "TypeDef with Enum",
			td: &TypeDef{
				Name: "EnumType",
				Type: DataTypeEnum,
				Enum: &Enum{
					Type: &TypeDef{
						Type: DataTypeString,
					},
					Values: []string{"value1", "value2", "value3"},
					Names:  []string{"NAME1", "NAME2", "NAME3"},
					Open:   false,
					Format: "enum",
					Descriptions: map[string]string{
						"value1": "First value",
						"value2": "Second value",
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Enum)
				assert.NotSame(t, original.Enum, cloned.Enum)
				assert.NotSame(t, original.Enum.Type, cloned.Enum.Type)
				assert.Equal(t, original.Enum.Type.Type, cloned.Enum.Type.Type)
				assert.Equal(t, original.Enum.Values, cloned.Enum.Values)
				assert.Equal(t, original.Enum.Names, cloned.Enum.Names)
				assert.Equal(t, original.Enum.Open, cloned.Enum.Open)
				assert.Equal(t, original.Enum.Format, cloned.Enum.Format)
				assert.Equal(t, original.Enum.Descriptions, cloned.Enum.Descriptions)
				// Ensure slices are different instances
				if len(original.Enum.Values) > 0 {
					assert.NotSame(t, &original.Enum.Values[0], &cloned.Enum.Values[0])
				}
			},
		},
		{
			name: "TypeDef with Validations",
			td: &TypeDef{
				Name: "ValidatedType",
				Type: DataTypeString,
				Validations: &Validations{
					MinLength: ptr(int64(5)),
					MaxLength: ptr(int64(100)),
					Pattern:   ptr("^[a-zA-Z]+$"),
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Validations)
				assert.NotSame(t, original.Validations, cloned.Validations)
				assert.NotSame(t, original.Validations.MinLength, cloned.Validations.MinLength)
				assert.Equal(t, *original.Validations.MinLength, *cloned.Validations.MinLength)
				assert.NotSame(t, original.Validations.MaxLength, cloned.Validations.MaxLength)
				assert.Equal(t, *original.Validations.MaxLength, *cloned.Validations.MaxLength)
				assert.NotSame(t, original.Validations.Pattern, cloned.Validations.Pattern)
				assert.Equal(t, *original.Validations.Pattern, *cloned.Validations.Pattern)
			},
		},
		{
			name: "TypeDef with Comments",
			td: &TypeDef{
				Name: "CommentedType",
				Type: DataTypeClass,
				Comments: &Comment{
					Summary:            "Summary text",
					Description:        "Description text",
					Deprecated:         true,
					DeprecationMessage: "Use NewType instead",
					ExternalDocs: &ExternalDocs{
						Description: "External docs",
						URL:         "https://example.com",
					},
					ExtendedComments: map[string]*ExtendedComment{
						"key1": {
							Summary:     "Extended summary",
							Description: "Extended description",
						},
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Comments)
				assert.NotSame(t, original.Comments, cloned.Comments)
				assert.Equal(t, original.Comments.Summary, cloned.Comments.Summary)
				assert.Equal(t, original.Comments.Description, cloned.Comments.Description)
				assert.Equal(t, original.Comments.Deprecated, cloned.Comments.Deprecated)
				assert.Equal(t, original.Comments.DeprecationMessage, cloned.Comments.DeprecationMessage)
				assert.NotSame(t, original.Comments.ExternalDocs, cloned.Comments.ExternalDocs)
				assert.Equal(t, original.Comments.ExternalDocs.Description, cloned.Comments.ExternalDocs.Description)
				assert.Equal(t, original.Comments.ExternalDocs.URL, cloned.Comments.ExternalDocs.URL)
				// Maps can't be compared with NotSame, check they have equal content
				assert.Len(t, cloned.Comments.ExtendedComments, len(original.Comments.ExtendedComments))
				// Verify deep copy by checking if values are different instances
				for k, v := range original.Comments.ExtendedComments {
					clonedVal := cloned.Comments.ExtendedComments[k]
					assert.NotNil(t, clonedVal)
					assert.NotSame(t, v, clonedVal)
					assert.Equal(t, v.Summary, clonedVal.Summary)
					assert.Equal(t, v.Description, clonedVal.Description)
				}
			},
		},
		{
			name: "TypeDef with Extensions",
			td: &TypeDef{
				Name: "ExtendedType",
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					All: map[string]any{
						"x-custom": "value",
						"x-other":  123,
					},
					Entity: &extensions.Entity{
						Names: []string{"entity1", "entity2"},
					},
					EntityDescription: &extensions.EntityDescription{
						TerraformDataResource:    "data resource",
						TerraformManagedResource: "managed resource",
					},
					EntityVersion: &extensions.EntityVersion{
						TerraformManagedResource: 42,
					},
					ExampleUnset: ptr(true),
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Extensions)
				assert.NotSame(t, original.Extensions, cloned.Extensions)
				// Maps can't be compared with NotSame, just check equal content
				assert.Equal(t, original.Extensions.All, cloned.Extensions.All)
				assert.NotSame(t, original.Extensions.Entity, cloned.Extensions.Entity)
				assert.Equal(t, original.Extensions.Entity.Names, cloned.Extensions.Entity.Names)
				assert.NotSame(t, original.Extensions.EntityDescription, cloned.Extensions.EntityDescription)
				assert.Equal(t, original.Extensions.EntityDescription.TerraformDataResource, cloned.Extensions.EntityDescription.TerraformDataResource)
				assert.NotSame(t, original.Extensions.ExampleUnset, cloned.Extensions.ExampleUnset)
				assert.Equal(t, *original.Extensions.ExampleUnset, *cloned.Extensions.ExampleUnset)
			},
		},
		{
			name: "TypeDef with AssociatedTypes",
			td: &TypeDef{
				Name: "UnionType",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "AssocType1",
						Type: DataTypeString,
					},
					{
						Name: "AssocType2",
						Type: DataTypeInteger,
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.AssociatedTypes, len(original.AssociatedTypes))
				for i := range original.AssociatedTypes {
					assert.NotSame(t, original.AssociatedTypes[i], cloned.AssociatedTypes[i])
					assert.Equal(t, original.AssociatedTypes[i].Name, cloned.AssociatedTypes[i].Name)
					assert.Equal(t, original.AssociatedTypes[i].Type, cloned.AssociatedTypes[i].Type)
				}
			},
		},
		{
			name: "TypeDef with Discriminator",
			td: &TypeDef{
				Name: "DiscriminatedUnion",
				Type: DataTypeUnion,
				Discriminator: &Discriminator{
					TypePropertyName: "type",
					Mapping: []*DiscriminatorMapping{
						{
							Name: "mapping1",
							Type: &TypeDef{
								Name: "MappedType1",
								Type: DataTypeClass,
							},
						},
						{
							Name: "mapping2",
							Type: &TypeDef{
								Name: "MappedType2",
								Type: DataTypeClass,
							},
						},
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Discriminator)
				assert.NotSame(t, original.Discriminator, cloned.Discriminator)
				assert.Equal(t, original.Discriminator.TypePropertyName, cloned.Discriminator.TypePropertyName)
				assert.Len(t, cloned.Discriminator.Mapping, len(original.Discriminator.Mapping))
				for i := range original.Discriminator.Mapping {
					assert.NotSame(t, original.Discriminator.Mapping[i], cloned.Discriminator.Mapping[i])
					assert.Equal(t, original.Discriminator.Mapping[i].Name, cloned.Discriminator.Mapping[i].Name)
					assert.NotSame(t, original.Discriminator.Mapping[i].Type, cloned.Discriminator.Mapping[i].Type)
					assert.Equal(t, original.Discriminator.Mapping[i].Type.Name, cloned.Discriminator.Mapping[i].Type.Name)
				}
			},
		},
		{
			name: "TypeDef with Location",
			td: &TypeDef{
				Name: "LocatedType",
				Type: DataTypeClass,
				Location: &OpenAPILocation{
					Node: &yaml.Node{
						Line:   42,
						Column: 10,
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Location)
				assert.NotSame(t, original.Location, cloned.Location)
				assert.Equal(t, original.Location.Node, cloned.Location.Node)

			},
		},
		{
			name: "TypeDef with nested recursive structure",
			td:   recursiveTypeDef,
			test: func(t *testing.T, original, cloned *TypeDef) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.Fields)
				assert.NotSame(t, original.Fields[0], cloned.Fields[0])
				assert.NotSame(t, original.Fields[0].Type, cloned.Fields[0].Type)
				assert.Equal(t, original.Fields[0].Type.Name, cloned.Fields[0].Type.Name)
				assert.NotNil(t, cloned.Fields[0].Type.Fields)
				assert.Len(t, cloned.Fields[0].Type.Fields, len(original.Fields[0].Type.Fields))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.td.Clone()
			tt.test(t, tt.td, cloned)
		})
	}
}

func TestTypeDef_DeepClone(t *testing.T) {
	t.Parallel()

	t.Run("nil TypeDef", func(t *testing.T) {
		t.Parallel()

		var td *TypeDef
		cloned := td.DeepClone()
		assert.Nil(t, cloned)
	})

	t.Run("basic fields are copied", func(t *testing.T) {
		t.Parallel()

		td := &TypeDef{
			Name:         "TestType",
			OriginalName: "OriginalTestType",
			Type:         DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{"x-test": true},
			},
			Fields: Fields{
				{
					Name: "field1",
					Type: &TypeDef{
						Name: "FieldType",
						Type: DataTypeString,
					},
				},
			},
		}

		cloned := td.DeepClone()
		require.NotNil(t, cloned)
		assert.NotSame(t, td, cloned)
		assert.Equal(t, td.Name, cloned.Name)
		assert.Equal(t, td.OriginalName, cloned.OriginalName)
		assert.Equal(t, td.Type, cloned.Type)
		assert.NotSame(t, td.Fields[0], cloned.Fields[0])
		assert.NotSame(t, td.Fields[0].Type, cloned.Fields[0].Type)
		assert.Equal(t, "FieldType", cloned.Fields[0].Type.Name)
	})

	t.Run("shared pointers are broken", func(t *testing.T) {
		t.Parallel()

		// Create a shared TypeDef referenced from two fields (simulating
		// OAS $ref caching where the same schema is used in multiple places).
		sharedType := &TypeDef{
			Name: "SharedType",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{"x-original": true},
			},
		}

		td := &TypeDef{
			Name: "Root",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "direct_ref", Type: sharedType},
				{Name: "nested_ref", Type: &TypeDef{
					Name:     "Container",
					Type:     DataTypeArray,
					ItemType: sharedType,
				}},
			},
		}

		// Verify sharing exists before clone.
		assert.Same(t, td.Fields[0].Type, td.Fields[1].Type.ItemType)

		// Clone() preserves sharing.
		cloned := td.Clone()
		assert.Same(t, cloned.Fields[0].Type, cloned.Fields[1].Type.ItemType,
			"Clone() should preserve shared pointers")

		// DeepClone() breaks sharing.
		deepCloned := td.DeepClone()
		assert.NotSame(t, deepCloned.Fields[0].Type, deepCloned.Fields[1].Type.ItemType,
			"DeepClone() should break shared pointers")

		// Both copies should have the same data.
		assert.Equal(t, "SharedType", deepCloned.Fields[0].Type.Name)
		assert.Equal(t, "SharedType", deepCloned.Fields[1].Type.ItemType.Name)
	})

	t.Run("mutation isolation after deep clone", func(t *testing.T) {
		t.Parallel()

		sharedType := &TypeDef{
			Name: "SharedType",
			Type: DataTypeClass,
			Extensions: &TypeDefExtensions{
				All: map[string]any{},
			},
		}

		td := &TypeDef{
			Name: "Root",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "ref1", Type: sharedType},
				{Name: "ref2", Type: &TypeDef{
					Name:     "Wrapper",
					Type:     DataTypeArray,
					ItemType: sharedType,
				}},
			},
		}

		deepCloned := td.DeepClone()

		// Mutate one copy's extensions.
		deepCloned.Fields[0].Type.Extensions.All["x-speakeasy-param-computed"] = true

		// The other copy should not be affected.
		_, exists := deepCloned.Fields[1].Type.ItemType.Extensions.All["x-speakeasy-param-computed"]
		assert.False(t, exists,
			"mutating one copy should not affect the other after DeepClone()")
	})

	t.Run("cycle handling", func(t *testing.T) {
		t.Parallel()

		// Create a cycle: A -> field -> B -> field -> A
		typeA := &TypeDef{
			Name: "TypeA",
			Type: DataTypeClass,
		}
		typeB := &TypeDef{
			Name: "TypeB",
			Type: DataTypeClass,
			Fields: Fields{
				{Name: "back_ref", Type: typeA},
			},
		}
		typeA.Fields = Fields{
			{Name: "forward_ref", Type: typeB},
		}

		// DeepClone should not infinite loop.
		cloned := typeA.DeepClone()
		require.NotNil(t, cloned)
		assert.Equal(t, "TypeA", cloned.Name)
		assert.Equal(t, "TypeB", cloned.Fields[0].Type.Name)
		// The back-reference should point to the cloned TypeA (cycle preserved).
		assert.Same(t, cloned, cloned.Fields[0].Type.Fields[0].Type)
	})
}

func TestValidations_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		v    *Validations
		test func(t *testing.T, original, cloned *Validations)
	}{
		{
			name: "nil Validations",
			v:    nil,
			test: func(t *testing.T, original, cloned *Validations) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "Validations with all fields",
			v: &Validations{
				MinItems:    ptr(int64(1)),
				MinLength:   ptr(int64(5)),
				Minimum:     ptr(10.5),
				MaxItems:    ptr(int64(100)),
				MaxLength:   ptr(int64(255)),
				Maximum:     ptr(999.9),
				Pattern:     ptr("^[a-z]+$"),
				UniqueItems: ptr(true),
			},
			test: func(t *testing.T, original, cloned *Validations) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)

				// Check all pointer fields are deep copied
				if original.MinItems != nil {
					require.NotNil(t, cloned.MinItems)
					assert.NotSame(t, original.MinItems, cloned.MinItems)
					assert.Equal(t, *original.MinItems, *cloned.MinItems)
				}
				if original.MinLength != nil {
					require.NotNil(t, cloned.MinLength)
					assert.NotSame(t, original.MinLength, cloned.MinLength)
					assert.Equal(t, *original.MinLength, *cloned.MinLength)
				}
				if original.Minimum != nil {
					require.NotNil(t, cloned.Minimum)
					assert.NotSame(t, original.Minimum, cloned.Minimum)
					assert.InEpsilon(t, *original.Minimum, *cloned.Minimum, 1e-9)
				}
				if original.MaxItems != nil {
					require.NotNil(t, cloned.MaxItems)
					assert.NotSame(t, original.MaxItems, cloned.MaxItems)
					assert.Equal(t, *original.MaxItems, *cloned.MaxItems)
				}
				if original.MaxLength != nil {
					require.NotNil(t, cloned.MaxLength)
					assert.NotSame(t, original.MaxLength, cloned.MaxLength)
					assert.Equal(t, *original.MaxLength, *cloned.MaxLength)
				}
				if original.Maximum != nil {
					require.NotNil(t, cloned.Maximum)
					assert.NotSame(t, original.Maximum, cloned.Maximum)
					assert.InEpsilon(t, *original.Maximum, *cloned.Maximum, 1e-9)
				}
				if original.Pattern != nil {
					require.NotNil(t, cloned.Pattern)
					assert.NotSame(t, original.Pattern, cloned.Pattern)
					assert.Equal(t, *original.Pattern, *cloned.Pattern)
				}
				if original.UniqueItems != nil {
					require.NotNil(t, cloned.UniqueItems)
					assert.NotSame(t, original.UniqueItems, cloned.UniqueItems)
					assert.Equal(t, *original.UniqueItems, *cloned.UniqueItems)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.v.Clone()
			tt.test(t, tt.v, cloned)
		})
	}
}

func TestValidations_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v1       *Validations
		v2       *Validations
		validate func(t *testing.T, result *Validations)
	}{
		{
			name: "merges nil validations",
			v1:   nil,
			v2:   &Validations{MaxLength: ptr(int64(10))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "merges into nil fields",
			v1:   &Validations{},
			v2: &Validations{
				MaxLength:   ptr(int64(10)),
				Maximum:     ptr(100.0),
				Minimum:     ptr(10.0),
				MaxItems:    ptr(int64(50)),
				Pattern:     ptr("^[a-z]+$"),
				UniqueItems: ptr(true),
			},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(10), *result.MaxLength)
				assert.InDelta(t, 100.0, *result.Maximum, 0.001)
				assert.InDelta(t, 10.0, *result.Minimum, 0.001)
				assert.Equal(t, int64(50), *result.MaxItems)
				assert.Equal(t, "^[a-z]+$", *result.Pattern)
				assert.True(t, *result.UniqueItems)
			},
		},
		{
			name: "Maximum uses stricter (smaller) value",
			v1:   &Validations{Maximum: ptr(100.0)},
			v2:   &Validations{Maximum: ptr(50.0)},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.InDelta(t, 50.0, *result.Maximum, 0.001)
			},
		},
		{
			name: "Maximum preserves stricter when other is larger",
			v1:   &Validations{Maximum: ptr(50.0)},
			v2:   &Validations{Maximum: ptr(100.0)},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.InDelta(t, 50.0, *result.Maximum, 0.001)
			},
		},
		{
			name: "Minimum uses stricter (larger) value",
			v1:   &Validations{Minimum: ptr(10.0)},
			v2:   &Validations{Minimum: ptr(50.0)},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.InDelta(t, 50.0, *result.Minimum, 0.001)
			},
		},
		{
			name: "Minimum preserves stricter when other is smaller",
			v1:   &Validations{Minimum: ptr(50.0)},
			v2:   &Validations{Minimum: ptr(10.0)},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.InDelta(t, 50.0, *result.Minimum, 0.001)
			},
		},
		{
			name: "MaxLength smaller value",
			v1:   &Validations{MaxLength: ptr(int64(100))},
			v2:   &Validations{MaxLength: ptr(int64(50))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(50), *result.MaxLength)
			},
		},
		{
			name: "MaxLength preserves stricter when other is larger",
			v1:   &Validations{MaxLength: ptr(int64(50))},
			v2:   &Validations{MaxLength: ptr(int64(100))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(50), *result.MaxLength)
			},
		},
		{
			name: "MaxItems smaller value",
			v1:   &Validations{MaxItems: ptr(int64(100))},
			v2:   &Validations{MaxItems: ptr(int64(50))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(50), *result.MaxItems)
			},
		},
		{
			name: "MaxItems preserves stricter when other is larger",
			v1:   &Validations{MaxItems: ptr(int64(50))},
			v2:   &Validations{MaxItems: ptr(int64(100))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(50), *result.MaxItems)
			},
		},
		{
			name: "MinItems uses stricter (larger) value",
			v1:   &Validations{MinItems: ptr(int64(10))},
			v2:   &Validations{MinItems: ptr(int64(50))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(50), *result.MinItems)
			},
		},
		{
			name: "MinItems preserves stricter when other is smaller",
			v1:   &Validations{MinItems: ptr(int64(50))},
			v2:   &Validations{MinItems: ptr(int64(10))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(50), *result.MinItems)
			},
		},
		{
			name: "MinLength uses stricter (larger) value",
			v1:   &Validations{MinLength: ptr(int64(5))},
			v2:   &Validations{MinLength: ptr(int64(20))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(20), *result.MinLength)
			},
		},
		{
			name: "MinLength preserves stricter when other is smaller",
			v1:   &Validations{MinLength: ptr(int64(20))},
			v2:   &Validations{MinLength: ptr(int64(5))},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, int64(20), *result.MinLength)
			},
		},
		{
			name: "combines different patterns with alternation",
			v1:   &Validations{Pattern: ptr("^[a-z]+$")},
			v2:   &Validations{Pattern: ptr("^[0-9]+$")},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, "(^[a-z]+$|^[0-9]+$)", *result.Pattern)
			},
		},
		{
			name: "keeps original pattern when patterns match",
			v1:   &Validations{Pattern: ptr("^[a-z]+$")},
			v2:   &Validations{Pattern: ptr("^[a-z]+$")},
			validate: func(t *testing.T, result *Validations) {
				t.Helper()
				assert.Equal(t, "^[a-z]+$", *result.Pattern)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone v1 to avoid mutation
			var testV1 *Validations
			if tt.v1 != nil {
				testV1 = tt.v1.Clone()
			}

			testV1.Merge(tt.v2)
			tt.validate(t, testV1)
		})
	}
}

func TestTypeDef_AssociatedTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		td        *TypeDef
		assocType *TypeDef
		expected  string
	}{
		{
			name:      "returns empty for nil TypeDef",
			td:        nil,
			assocType: &TypeDef{Name: "Test"},
			expected:  "",
		},
		{
			name:      "returns empty for nil associated type",
			td:        &TypeDef{Name: "Test"},
			assocType: nil,
			expected:  "",
		},
		{
			name: "returns discriminator mapping name when available",
			td: func() *TypeDef {
				assocType := &TypeDef{
					Name:         "ActualName",
					OriginalName: "OriginalName",
				}
				return &TypeDef{
					Name:            "UnionType",
					AssociatedTypes: []*TypeDef{assocType},
					Discriminator: &Discriminator{
						TypePropertyName: "type",
						Mapping: []*DiscriminatorMapping{
							{
								Name: "MappingName",
								Type: assocType,
							},
						},
					},
				}
			}(),
			assocType: &TypeDef{
				Name:         "ActualName",
				OriginalName: "OriginalName",
			},
			expected: "MappingName",
		},
		{
			name: "returns OriginalName when no discriminator",
			td: func() *TypeDef {
				assocType := &TypeDef{
					Name:         "ActualName",
					OriginalName: "OriginalName",
				}
				return &TypeDef{
					Name:            "UnionType",
					AssociatedTypes: []*TypeDef{assocType},
				}
			}(),
			assocType: &TypeDef{
				Name:         "ActualName",
				OriginalName: "OriginalName",
			},
			expected: "OriginalName",
		},
		{
			name: "returns Name when no OriginalName",
			td: func() *TypeDef {
				assocType := &TypeDef{Name: "ActualName"}
				return &TypeDef{
					Name:            "UnionType",
					AssociatedTypes: []*TypeDef{assocType},
				}
			}(),
			assocType: &TypeDef{Name: "ActualName"},
			expected:  "ActualName",
		},
		{
			name: "does not match discriminator mapping on empty OriginalName",
			td: func() *TypeDef {
				assocType1 := &TypeDef{
					Name:         "FirstType",
					OriginalName: "",
				}
				assocType2 := &TypeDef{
					Name:         "SecondType",
					OriginalName: "",
				}
				return &TypeDef{
					Name:            "UnionType",
					AssociatedTypes: []*TypeDef{assocType1, assocType2},
					Discriminator: &Discriminator{
						TypePropertyName: "type",
						Mapping: []*DiscriminatorMapping{
							{
								Name: "first_mapping",
								Type: assocType1,
							},
							{
								Name: "second_mapping",
								Type: assocType2,
							},
						},
					},
				}
			}(),
			assocType: &TypeDef{
				Name:         "SecondType",
				OriginalName: "",
			},
			expected: "second_mapping",
		},
		{
			name: "returns data type for string when no name",
			td: func() *TypeDef {
				assocType := &TypeDef{Type: DataTypeString}
				return &TypeDef{
					Name:            "UnionType",
					AssociatedTypes: []*TypeDef{assocType},
				}
			}(),
			assocType: &TypeDef{Type: DataTypeString},
			expected:  "Str",
		},
		{
			name: "returns data type for primitive when no name",
			td: func() *TypeDef {
				assocType := &TypeDef{Type: DataTypeInteger}
				return &TypeDef{
					Name:            "UnionType",
					AssociatedTypes: []*TypeDef{assocType},
				}
			}(),
			assocType: &TypeDef{Type: DataTypeInteger},
			expected:  "integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.td.AssociatedTypeName(tt.assocType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeDef_FindAssociatedTypeByTypeDef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		td         *TypeDef
		searchType *TypeDef
		validate   func(t *testing.T, result *TypeDef)
	}{
		{
			name:       "returns nil for nil TypeDef",
			td:         nil,
			searchType: &TypeDef{Name: "Test"},
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "returns nil for nil search TypeDef",
			td: &TypeDef{
				Name:            "Union",
				AssociatedTypes: []*TypeDef{{Name: "Test"}},
			},
			searchType: nil,
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name:       "returns nil for empty associated types",
			td:         &TypeDef{Name: "Union"},
			searchType: &TypeDef{Name: "Test"},
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "finds associated type by name",
			td: func() *TypeDef {
				assoc1 := &TypeDef{Name: "Type1"}
				assoc2 := &TypeDef{Name: "Type2"}
				return &TypeDef{
					Name:            "Union",
					AssociatedTypes: []*TypeDef{assoc1, assoc2},
				}
			}(),
			searchType: &TypeDef{Name: "Type2"},
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result)
				assert.Equal(t, "Type2", result.Name)
			},
		},
		{
			name: "finds associated type by original name",
			td: func() *TypeDef {
				assoc1 := &TypeDef{Name: "Type1", OriginalName: "OrigType1"}
				assoc2 := &TypeDef{Name: "Type2", OriginalName: "OrigType2"}
				return &TypeDef{
					Name:            "Union",
					AssociatedTypes: []*TypeDef{assoc1, assoc2},
				}
			}(),
			searchType: &TypeDef{OriginalName: "OrigType1"},
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result)
				assert.Equal(t, "Type1", result.Name)
			},
		},
		{
			name: "finds associated type by discriminator mapping",
			td: func() *TypeDef {
				assoc1 := &TypeDef{Name: "Type1"}
				assoc2 := &TypeDef{Name: "Type2"}
				return &TypeDef{
					Name:            "Union",
					AssociatedTypes: []*TypeDef{assoc1, assoc2},
					Discriminator: &Discriminator{
						TypePropertyName: "type",
						Mapping: []*DiscriminatorMapping{
							{Name: "mapping1", Type: assoc1},
							{Name: "mapping2", Type: assoc2},
						},
					},
				}
			}(),
			searchType: &TypeDef{Name: "Type1"},
			validate: func(t *testing.T, result *TypeDef) {
				t.Helper()
				require.NotNil(t, result)
				assert.Equal(t, "Type1", result.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.td.FindAssociatedTypeByTypeDef(tt.searchType)
			tt.validate(t, result)
		})
	}
}

func TestTypeDef_FindEntitySDKMethodTargets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		td         *TypeDef
		entityName string
		optional   bool
		expected   []TerraformSDKMethodTarget
	}{
		{
			name:       "nil TypeDef returns nil",
			td:         nil,
			entityName: "Thing",
			optional:   true,
			expected:   nil,
		},
		{
			name: "no entity match returns nil",
			td: &TypeDef{
				Name: "Request",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{Name: "field1", Type: &TypeDef{Name: "String", Type: DataTypeString}},
				},
			},
			entityName: "Thing",
			optional:   true,
			expected:   nil,
		},
		{
			name: "self is entity returns single element",
			td: &TypeDef{
				Name: "Thing",
				Type: DataTypeClass,
				Extensions: &TypeDefExtensions{
					Entity: &extensions.Entity{Names: []string{"Thing"}},
				},
			},
			entityName: "Thing",
			optional:   true,
			expected: []TerraformSDKMethodTarget{
				{Optional: true}, // TypeDef set below
			},
		},
		{
			name: "entity in field returns path with field optional tracking",
			td: &TypeDef{
				Name: "Request",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{
						Name:     "data",
						Optional: true,
						Type: &TypeDef{
							Name: "Thing",
							Type: DataTypeClass,
							Extensions: &TypeDefExtensions{
								Entity: &extensions.Entity{Names: []string{"Thing"}},
							},
						},
					},
				},
			},
			entityName: "Thing",
			optional:   false,
			expected: []TerraformSDKMethodTarget{
				{Optional: true},  // entity TypeDef (from field.Optional)
				{Optional: false}, // root TypeDef (from initial optional)
			},
		},
		{
			name: "entity nested two levels deep",
			td: &TypeDef{
				Name: "Request",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{
						Name: "wrapper",
						Type: &TypeDef{
							Name: "Wrapper",
							Type: DataTypeClass,
							Fields: []*FieldDef{
								{
									Name:     "thing",
									Nullable: true,
									Type: &TypeDef{
										Name: "Thing",
										Type: DataTypeClass,
										Extensions: &TypeDefExtensions{
											Entity: &extensions.Entity{Names: []string{"Thing"}},
										},
									},
								},
							},
						},
					},
				},
			},
			entityName: "Thing",
			optional:   true,
			expected: []TerraformSDKMethodTarget{
				{Optional: true},  // entity TypeDef (from field.Nullable)
				{Optional: false}, // Wrapper (from field not optional/nullable)
				{Optional: true},  // Request (from initial optional)
			},
		},
		{
			name: "entity in ItemType",
			td: &TypeDef{
				Name: "List",
				Type: DataTypeArray,
				ItemType: &TypeDef{
					Name: "Thing",
					Type: DataTypeClass,
					Extensions: &TypeDefExtensions{
						Entity: &extensions.Entity{Names: []string{"Thing"}},
					},
				},
			},
			entityName: "Thing",
			optional:   true,
			expected: []TerraformSDKMethodTarget{
				{Optional: false}, // entity TypeDef (ItemType always passes false)
				{Optional: true},  // root array TypeDef
			},
		},
		{
			name: "entity in AssociatedType",
			td: &TypeDef{
				Name: "Union",
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{
						Name: "Thing",
						Type: DataTypeClass,
						Extensions: &TypeDefExtensions{
							Entity: &extensions.Entity{Names: []string{"Thing"}},
						},
					},
				},
			},
			entityName: "Thing",
			optional:   false,
			expected: []TerraformSDKMethodTarget{
				{Optional: true},  // entity TypeDef (AssociatedType always passes true)
				{Optional: false}, // root union TypeDef
			},
		},
		{
			name: "returns first matching path only",
			td: &TypeDef{
				Name: "Request",
				Type: DataTypeClass,
				Fields: []*FieldDef{
					{
						Name: "first",
						Type: &TypeDef{
							Name: "First",
							Type: DataTypeClass,
							Extensions: &TypeDefExtensions{
								Entity: &extensions.Entity{Names: []string{"Thing"}},
							},
						},
					},
					{
						Name: "second",
						Type: &TypeDef{
							Name: "Second",
							Type: DataTypeClass,
							Extensions: &TypeDefExtensions{
								Entity: &extensions.Entity{Names: []string{"Thing"}},
							},
						},
					},
				},
			},
			entityName: "Thing",
			optional:   true,
			expected: []TerraformSDKMethodTarget{
				{Optional: false}, // First (first field, not optional)
				{Optional: true},  // Request
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.td.FindEntitySDKMethodTargets(tt.entityName, tt.optional)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			require.Len(t, result, len(tt.expected))

			for i, expected := range tt.expected {
				assert.Equal(t, expected.Optional, result[i].Optional, "index %d Optional", i)
				assert.NotNil(t, result[i].TypeDef, "index %d TypeDef", i)
			}
		})
	}
}

func TestTypeDef_IsTerraformPrimitiveType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeDef  *TypeDef
		expected bool
	}{
		{
			name:     "nil TypeDef",
			typeDef:  nil,
			expected: false,
		},
		{
			name:     "boolean is primitive",
			typeDef:  &TypeDef{Type: DataTypeBoolean},
			expected: true,
		},
		{
			name:     "bytes is primitive",
			typeDef:  &TypeDef{Type: DataTypeBytes},
			expected: true,
		},
		{
			name:     "float32 is primitive",
			typeDef:  &TypeDef{Type: DataTypeFloat32},
			expected: true,
		},
		{
			name:     "int32 is primitive",
			typeDef:  &TypeDef{Type: DataTypeInt32},
			expected: true,
		},
		{
			name:     "integer is primitive",
			typeDef:  &TypeDef{Type: DataTypeInteger},
			expected: true,
		},
		{
			name:     "number is primitive",
			typeDef:  &TypeDef{Type: DataTypeNumber},
			expected: true,
		},
		{
			name:     "string is primitive",
			typeDef:  &TypeDef{Type: DataTypeString},
			expected: true,
		},
		{
			name:     "class is not primitive",
			typeDef:  &TypeDef{Type: DataTypeClass},
			expected: false,
		},
		{
			name:     "array is not primitive",
			typeDef:  &TypeDef{Type: DataTypeArray},
			expected: false,
		},
		{
			name:     "union is not primitive",
			typeDef:  &TypeDef{Type: DataTypeUnion},
			expected: false,
		},
		{
			name: "enum with string underlying type is primitive",
			typeDef: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{
					Type: &TypeDef{Type: DataTypeString},
				},
			},
			expected: true,
		},
		{
			name: "enum with integer underlying type is primitive",
			typeDef: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{
					Type: &TypeDef{Type: DataTypeInteger},
				},
			},
			expected: true,
		},
		{
			name: "enum without type is not primitive",
			typeDef: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.typeDef.IsTerraformPrimitiveType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeDef_IsTerraformEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		td1      *TypeDef
		td2      *TypeDef
		expected bool
	}{
		{
			name:     "nil TypeDefs are equal",
			td1:      nil,
			td2:      nil,
			expected: true,
		},
		{
			name:     "different data types are not equal",
			td1:      &TypeDef{Type: DataTypeString},
			td2:      &TypeDef{Type: DataTypeInteger},
			expected: false,
		},
		{
			name:     "same data types are equal",
			td1:      &TypeDef{Type: DataTypeString},
			td2:      &TypeDef{Type: DataTypeString},
			expected: true,
		},
		{
			name: "enum with same underlying type is equal",
			td1: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
			},
			td2:      &TypeDef{Type: DataTypeString},
			expected: true,
		},
		{
			name: "equal associated types",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			expected: true,
		},
		{
			name: "different number of associated types not equal",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			expected: false,
		},
		{
			name: "equal fields",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			expected: true,
		},
		{
			name: "equal item types",
			td1: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			td2: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			expected: true,
		},
		{
			name: "different item types not equal",
			td1: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeString},
			},
			td2: &TypeDef{
				Type:     DataTypeArray,
				ItemType: &TypeDef{Type: DataTypeInteger},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.td1.IsTerraformEqual(tt.td2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypeDef_TerraformMergeDataType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		td1           *TypeDef
		td2           *TypeDef
		expectedType  DataType
		expectedError error
	}{
		{
			name:          "returns error for type mismatch",
			td1:           &TypeDef{Type: DataTypeString},
			td2:           &TypeDef{Type: DataTypeClass},
			expectedError: ErrTypeMismatch,
		},
		{
			name:         "returns same type when types match",
			td1:          &TypeDef{Type: DataTypeString},
			td2:          &TypeDef{Type: DataTypeString},
			expectedType: DataTypeString,
		},
		{
			name:         "class takes precedence over union",
			td1:          &TypeDef{Type: DataTypeClass},
			td2:          &TypeDef{Type: DataTypeUnion},
			expectedType: DataTypeClass,
		},
		{
			name:         "union takes precedence over class",
			td1:          &TypeDef{Type: DataTypeUnion},
			td2:          &TypeDef{Type: DataTypeClass},
			expectedType: DataTypeUnion,
		},
		{
			name:         "integer takes precedence over float32",
			td1:          &TypeDef{Type: DataTypeFloat32},
			td2:          &TypeDef{Type: DataTypeInteger},
			expectedType: DataTypeInteger,
		},
		{
			name:         "int32 takes precedence over float32",
			td1:          &TypeDef{Type: DataTypeFloat32},
			td2:          &TypeDef{Type: DataTypeInt32},
			expectedType: DataTypeInt32,
		},
		{
			name:         "class takes precedence over any",
			td1:          &TypeDef{Type: DataTypeClass},
			td2:          &TypeDef{Type: DataTypeAny},
			expectedType: DataTypeClass,
		},
		{
			name: "preserves enum type when inner type matches",
			td1: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
			},
			td2:          &TypeDef{Type: DataTypeString},
			expectedType: DataTypeEnum,
		},
		{
			name: "preserves receiver type when other is enum with matching inner type",
			td1:  &TypeDef{Type: DataTypeString},
			td2: &TypeDef{
				Type: DataTypeEnum,
				Enum: &Enum{Type: &TypeDef{Type: DataTypeString}},
			},
			expectedType: DataTypeString,
		},
		{
			name:         "int32 reconciles with integer",
			td1:          &TypeDef{Type: DataTypeInt32},
			td2:          &TypeDef{Type: DataTypeInteger},
			expectedType: DataTypeInt32,
		},
		{
			name:         "integer reconciles with int32",
			td1:          &TypeDef{Type: DataTypeInteger},
			td2:          &TypeDef{Type: DataTypeInt32},
			expectedType: DataTypeInteger,
		},
		{
			name:         "bigint reconciles with integer",
			td1:          &TypeDef{Type: DataTypeBigInt},
			td2:          &TypeDef{Type: DataTypeInteger},
			expectedType: DataTypeBigInt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.td1.TerraformMergeDataType(tt.td2)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedType, result)
			}
		})
	}
}

func TestTypeDef_TerraformMerge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		td1      *TypeDef
		td2      *TypeDef
		validate func(t *testing.T, td *TypeDef, err error)
	}{
		{
			name: "returns error for nil TypeDef",
			td1:  nil,
			td2:  &TypeDef{Type: DataTypeString},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.Error(t, err)
			},
		},
		{
			name: "handles nil other TypeDef",
			td1:  &TypeDef{Type: DataTypeString},
			td2:  nil,
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
			},
		},
		{
			name: "merges associated types",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type1", Type: DataTypeString},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "Type2", Type: DataTypeInteger},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, td.AssociatedTypes, 2)
			},
		},
		{
			name: "merges fields",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field1", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "field2", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, td.Fields, 2)
			},
		},
		{
			name: "merges comments",
			td1: &TypeDef{
				Type:     DataTypeString,
				Comments: &Comment{Summary: "Original"},
			},
			td2: &TypeDef{
				Type:     DataTypeString,
				Comments: &Comment{Description: "Additional"},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "Original", td.Comments.Summary)
				assert.Equal(t, "Additional", td.Comments.Description)
			},
		},
		{
			name: "merges content media type from other when empty",
			td1:  &TypeDef{Type: DataTypeString},
			td2:  &TypeDef{Type: DataTypeString, ContentMediaType: "application/json"},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "application/json", td.ContentMediaType)
			},
		},
		{
			name: "preserves content media type when already set",
			td1:  &TypeDef{Type: DataTypeString, ContentMediaType: "application/json"},
			td2:  &TypeDef{Type: DataTypeString, ContentMediaType: "application/xml"},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "application/json", td.ContentMediaType)
			},
		},
		{
			name: "merges extensions",
			td1: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-custom": "value1"},
				},
			},
			td2: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{"x-other": "value2"},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, td.Extensions.All, 2)
				assert.Equal(t, "value1", td.Extensions.All["x-custom"])
				assert.Equal(t, "value2", td.Extensions.All["x-other"])
			},
		},
		{
			name: "merges validations",
			td1: &TypeDef{
				Type:        DataTypeString,
				Validations: &Validations{MaxLength: ptr(int64(100))},
			},
			td2: &TypeDef{
				Type:        DataTypeString,
				Validations: &Validations{Pattern: ptr("^test$")},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.NotNil(t, td.Validations.MaxLength)
				assert.NotNil(t, td.Validations.Pattern)
				assert.Equal(t, int64(100), *td.Validations.MaxLength)
				assert.Equal(t, "^test$", *td.Validations.Pattern)
			},
		},
		{
			name: "uses shorter name",
			td1:  &TypeDef{Type: DataTypeString, Name: "LongTypeName"},
			td2:  &TypeDef{Type: DataTypeString, Name: "Short"},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				assert.Equal(t, "Short", td.Name)
			},
		},
		{
			name: "matches fields by sanitized name",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "MyField", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "my_field", Type: &TypeDef{Type: DataTypeString, Name: "updated"}},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				// Should match by sanitized name, not create a duplicate
				assert.Len(t, td.Fields, 1)
				assert.Equal(t, "MyField", td.Fields[0].Name)
			},
		},
		{
			name: "merges field comments",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name:     "field1",
						Type:     &TypeDef{Type: DataTypeString},
						Comments: &Comment{Summary: "Original summary"},
					},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{
						Name:     "field1",
						Type:     &TypeDef{Type: DataTypeString},
						Comments: &Comment{Description: "Added description"},
					},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.Len(t, td.Fields, 1)
				require.NotNil(t, td.Fields[0].Comments)
				assert.Equal(t, "Original summary", td.Fields[0].Comments.Summary)
				assert.Equal(t, "Added description", td.Fields[0].Comments.Description)
			},
		},
		{
			name: "sorts fields by name after merge",
			td1: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "zebra", Type: &TypeDef{Type: DataTypeString}},
					{Name: "apple", Type: &TypeDef{Type: DataTypeString}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeClass,
				Fields: Fields{
					{Name: "mango", Type: &TypeDef{Type: DataTypeInteger}},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.Len(t, td.Fields, 3)
				assert.Equal(t, "apple", td.Fields[0].Name)
				assert.Equal(t, "mango", td.Fields[1].Name)
				assert.Equal(t, "zebra", td.Fields[2].Name)
			},
		},
		{
			name: "sorts associated types by original name after merge",
			td1: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "TypeZ", OriginalName: "TypeZ", Type: DataTypeString},
					{Name: "TypeA", OriginalName: "TypeA", Type: DataTypeInteger},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "TypeM", OriginalName: "TypeM", Type: DataTypeBoolean},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.Len(t, td.AssociatedTypes, 3)
				assert.Equal(t, "TypeA", td.AssociatedTypes[0].OriginalName)
				assert.Equal(t, "TypeM", td.AssociatedTypes[1].OriginalName)
				assert.Equal(t, "TypeZ", td.AssociatedTypes[2].OriginalName)
			},
		},
		{
			name: "clones associated types when adding new",
			td1: &TypeDef{
				Type:            DataTypeUnion,
				AssociatedTypes: []*TypeDef{},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "NewType", Type: DataTypeString},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.Len(t, td.AssociatedTypes, 1)
				assert.Equal(t, "NewType", td.AssociatedTypes[0].Name)
			},
		},
		{
			name: "extension ignore other wins",
			td1: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All:    map[string]any{},
					Ignore: ptr(false),
				},
			},
			td2: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All:    map[string]any{},
					Ignore: ptr(true),
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, td.Extensions.Ignore)
				assert.True(t, *td.Extensions.Ignore)
			},
		},
		{
			name: "extension terraform ignore OR merges",
			td1: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{},
					TerraformIgnore: &extensions.TerraformIgnore{
						DataModel: true,
						Schema:    false,
					},
				},
			},
			td2: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All: map[string]any{},
					TerraformIgnore: &extensions.TerraformIgnore{
						DataModel: false,
						Schema:    true,
					},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, td.Extensions.TerraformIgnore)
				assert.True(t, td.Extensions.TerraformIgnore.DataModel)
				assert.True(t, td.Extensions.TerraformIgnore.Schema)
			},
		},
		{
			name: "extension terraform write only other wins",
			td1: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All:                map[string]any{},
					TerraformWriteOnly: ptr(false),
				},
			},
			td2: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All:                map[string]any{},
					TerraformWriteOnly: ptr(true),
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, td.Extensions.TerraformWriteOnly)
				assert.True(t, *td.Extensions.TerraformWriteOnly)
			},
		},
		{
			name: "extension terraform hoisted from other wins",
			td1: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All:                  map[string]any{},
					TerraformHoistedFrom: []TerraformHoistedSource{{FieldName: "original"}},
				},
			},
			td2: &TypeDef{
				Type: DataTypeString,
				Extensions: &TypeDefExtensions{
					All:                  map[string]any{},
					TerraformHoistedFrom: []TerraformHoistedSource{{FieldName: "updated"}},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				require.Len(t, td.Extensions.TerraformHoistedFrom, 1)
				assert.Equal(t, "updated", td.Extensions.TerraformHoistedFrom[0].FieldName)
			},
		},
		{
			name: "associated type dual context matching",
			td1: &TypeDef{
				Type: DataTypeUnion,
				Discriminator: &Discriminator{
					Mapping: DiscriminatorMappings{
						{Name: "mapped_name", Type: &TypeDef{Name: "InternalType"}},
					},
				},
				AssociatedTypes: []*TypeDef{
					{Name: "InternalType", OriginalName: "InternalType", Type: DataTypeClass},
				},
			},
			td2: &TypeDef{
				Type: DataTypeUnion,
				AssociatedTypes: []*TypeDef{
					{Name: "InternalType", OriginalName: "InternalType", Type: DataTypeClass,
						Fields: Fields{{Name: "new_field", Type: &TypeDef{Type: DataTypeString}}}},
				},
			},
			validate: func(t *testing.T, td *TypeDef, err error) {
				t.Helper()
				require.NoError(t, err)
				// Should merge into existing, not create duplicate
				require.Len(t, td.AssociatedTypes, 1)
				// The merged associated type should have the new field
				assert.Len(t, td.AssociatedTypes[0].Fields, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone td1 to avoid mutation affecting the test table
			var testTd1 *TypeDef
			if tt.td1 != nil {
				testTd1 = tt.td1.Clone()
			}

			err := testTd1.TerraformMerge(tt.td2)
			tt.validate(t, testTd1, err)
		})
	}
}
