package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleAnyOfOneOf tests the handleAnyOfOneOf method with various oneOf/anyOf scenarios
func TestHandleAnyOfOneOf(t *testing.T) {
	tests := []struct {
		name         string
		schemasYAML  string
		schemaName   string
		fixes        *config.Fixes                        // nil means use default
		mockFeatures func() *testutils.MockFeaturesConfig // nil means use default (all features supported)
		isRequest    *bool                                // nil means use default (false), otherwise use specified value
		expected     *ast.FieldDef
	}{
		{
			name: "StringOrNull",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: "null"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "StringOrNullReference",
			schemasYAML: `StringOrNullType:
  oneOf:
    - type: string
    - type: "null"
TestSchema:
  "$ref": "#/components/schemas/StringOrNullType"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "ObjectOrNull",
			schemasYAML: `TestSchema:
  oneOf:
    - type: object
      properties:
        name:
          type: string
        age:
          type: integer
    - type: "null"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 15, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 13, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "OneOfMultiplePrimitives",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: integer
    - type: boolean`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 10, Column: 11}).Build(),
						testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 11, Column: 11}).Build(),
						testutils.NewTypeDef(ast.DataTypeBoolean).Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "AnyOfMultiplePrimitives",
			schemasYAML: `TestSchema:
  anyOf:
    - type: string
    - type: integer
    - type: boolean`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 10, Column: 11}).Build(),
						testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 11, Column: 11}).Build(),
						testutils.NewTypeDef(ast.DataTypeBoolean).Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "MultipleTypesArray",
			schemasYAML: `TestSchema:
  type:
    - string
    - integer
    - boolean`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).Build(),
						testutils.NewTypeDef(ast.DataTypeInteger).Build(),
						testutils.NewTypeDef(ast.DataTypeBoolean).Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "MultipleTypesWithNull",
			schemasYAML: `TestSchema:
  type:
    - string
    - "null"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "MultipleTypesWithStringEnums",
			schemasYAML: `TestSchema:
  type:
    - string
    - integer
  enum:
    - "active"
    - "inactive"
    - "pending"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithStandardOneOfContext("1").Build(),
					).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeEnum).
							WithOriginalNameFrozen().
							WithContextStack(
								testutils.NewContextStack().WithStandardOneOfContext("1").Build(),
							).
							WithEnum(
								testutils.NewTypeDef(ast.DataTypeString).WithValidations(nil).Build(),
								[]string{"active", "inactive", "pending"},
								"enum",
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							Build(),
						testutils.NewTypeDef(ast.DataTypeInteger).Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "MultipleTypesWithIntegerEnums",
			schemasYAML: `TestSchema:
  type:
    - boolean
    - integer
  enum:
    - 1
    - 2
    - 3`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeBoolean).Build(),
						testutils.NewTypeDef(ast.DataTypeEnum).
							WithOriginalNameFrozen().
							WithContextStack(
								testutils.NewContextStack().WithStandardOneOfContext("2").Build(),
							).
							WithEnum(
								testutils.NewTypeDef(ast.DataTypeInteger).WithValidations(nil).Build(),
								[]string{"1", "2", "3"},
								"enum",
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "MultipleTypesStringIntegerWithIntegerEnums",
			schemasYAML: `TestSchema:
  type:
    - string
    - integer
  enum:
    - 1
    - 2
    - 3`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).Build(),
						testutils.NewTypeDef(ast.DataTypeEnum).
							WithOriginalNameFrozen().
							WithContextStack(
								testutils.NewContextStack().WithStandardOneOfContext("2").Build(),
							).
							WithEnum(
								testutils.NewTypeDef(ast.DataTypeInteger).WithValidations(nil).Build(),
								[]string{"1", "2", "3"},
								"enum",
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfNumbersWithDifferentValidations",
			schemasYAML: `TestSchema:
  oneOf:
    - type: number
    - type: number
      minimum: 1
    - type: number
      maximum: 20`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithValidations(&ast.Validations{
						Maximum: nil, // no maximum across all options
						Minimum: nil, // no minimum across all options
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfNumbersWithRepeatValidations",
			schemasYAML: `TestSchema:
  oneOf:
    - type: number
      maximum: 4
      minimum: 1
    - type: number
      maximum: 5
      minimum: 2
    - type: number
      maximum: 6
      minimum: 3`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("number",
				testutils.NewTypeDef(ast.DataTypeNumber).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithValidations(&ast.Validations{
						Maximum: pointer.From(float64(6)),
						Minimum: pointer.From(float64(1)),
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfStringsWithDifferentValidations",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
      pattern: "^[a-z]+$"
      minLength: 3
    - type: string
      maxLength: 10
      pattern: "^[A-Z]+$"
    - type: string
      minLength: 1
      maxLength: 20`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithValidations(&ast.Validations{
						MaxLength: nil, // no maxLength across all options
						MinLength: nil, // no minLength across all options
						Pattern:   nil, // no pattern across all options
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfStringsWithRepeatValidations",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
      pattern: "^[a-z]+$"
      maxLength: 4
      minLength: 1
    - type: string
      maxLength: 5
      minLength: 2
      pattern: "^[A-Z]+$"
    - type: string
      maxLength: 6
      minLength: 3
      pattern: "^[0-9]+$"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithValidations(&ast.Validations{
						MaxLength: pointer.From(int64(6)),
						MinLength: pointer.From(int64(1)),
						Pattern:   pointer.From("(^[a-z]+$|^[A-Z]+$|^[0-9]+$)"),
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithParentDescription",
			schemasYAML: `TestSchema:
  description: "Parent description for the oneOf"
  oneOf:
    - type: string
      description: "First sub-schema description"
    - type: string
      description: "Second sub-schema description"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 11, Column: 11}).
					WithComments(&ast.Comment{
						Description: "Parent description for the oneOf",
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithSubSchemaDescription",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
      description: "First sub-schema description"
    - type: string
      description: "Second sub-schema description"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithComments(&ast.Comment{
						Description: "First sub-schema description",
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithMixedDescriptions",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: string
      description: "Second sub-schema description"
    - type: string
      description: "Third sub-schema description"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithComments(&ast.Comment{
						Description: "Second sub-schema description",
					}).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithNoDescriptions",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: string`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithParentExample",
			schemasYAML: `TestSchema:
  example: "parent example"
  oneOf:
    - type: string
      example: "first sub-schema example"
    - type: string
      example: "second sub-schema example"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 11, Column: 11}).
					WithExamples(ast.NewExample("", "", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Style:  yaml.DoubleQuotedStyle,
						Tag:    "!!str",
						Value:  "parent example",
						Line:   9,
						Column: 16,
					})).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithSubSchemaExample",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
      example: "first sub-schema example"
    - type: string
      example: "second sub-schema example"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithExamples(ast.NewExample("", "", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Style:  yaml.DoubleQuotedStyle,
						Tag:    "!!str",
						Value:  "first sub-schema example",
						Line:   11,
						Column: 20,
					})).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithMixedExamples",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: string
      example: "second sub-schema example"
    - type: string
      example: "third sub-schema example"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					WithExamples(ast.NewExample("", "", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Style:  yaml.DoubleQuotedStyle,
						Tag:    "!!str",
						Value:  "second sub-schema example",
						Line:   12,
						Column: 20,
					})).
					Build()).
				Build(),
		},
		{
			name: "OneOfPrimitivesWithNoExamples",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: string`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					Build()).
				Build(),
		},
		{
			name: "OneOfMixedTypesWithStringExampleOnlyPropagatedToStringMember",
			schemasYAML: `TestSchema:
  example: "parent string example"
  oneOf:
    - type: string
    - type: number`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 10, Column: 7}).
					WithExamples(ast.NewExample("", "", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Style:  yaml.DoubleQuotedStyle,
						Tag:    "!!str",
						Value:  "parent string example",
						Line:   9,
						Column: 16,
					})).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 11, Column: 11}).
							WithExamples(ast.NewExample("", "", &yaml.Node{
								Kind:   yaml.ScalarNode,
								Style:  yaml.DoubleQuotedStyle,
								Tag:    "!!str",
								Value:  "parent string example",
								Line:   9,
								Column: 16,
							})).
							Build(),
						testutils.NewTypeDef(ast.DataTypeNumber).
							WithLocation(&yaml.Node{Line: 12, Column: 11}).
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfMixedTypesWithIntegerExampleOnlyPropagatedToNumberMember",
			schemasYAML: `TestSchema:
  example: 42
  oneOf:
    - type: string
    - type: number`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 10, Column: 7}).
					WithExamples(ast.NewExample("", "", &yaml.Node{
						Kind:   yaml.ScalarNode,
						Tag:    "!!int",
						Value:  "42",
						Line:   9,
						Column: 16,
					})).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).
							WithLocation(&yaml.Node{Line: 11, Column: 11}).
							Build(),
						testutils.NewTypeDef(ast.DataTypeNumber).
							WithLocation(&yaml.Node{Line: 12, Column: 11}).
							WithExamples(ast.NewExample("", "", &yaml.Node{
								Kind:   yaml.ScalarNode,
								Tag:    "!!int",
								Value:  "42",
								Line:   9,
								Column: 16,
							})).
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfWithDescriptionMerging",
			schemasYAML: `PersonObject:
  type: object
  properties:
    name:
      type: string
    age:
      type: integer
  required:
    - name
TestSchema:
  description: "A test schema with description"
  title: "Test Schema"
  properties:
    extraField:
      type: string
  oneOf:
    - $ref: "#/components/schemas/PersonObject"
    - type: "null"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("Schemas").
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("extraField", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
							WithOriginalName("extraField").
							WithOptional(true).
							WithJSONAnnotation("extraField").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(false).
							WithJSONAnnotation("name").
							Build(),
					).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "#/components/schemas/PersonObject",
					}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(false).
					WithContextStack(
						testutils.NewContextStack().WithRefType("PersonObject").WithIdentifierForNaming(pointer.From("PersonObject")).Build(),
					).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "OneOfWithDescriptionMergingDec2023",
			schemasYAML: `PersonObject:
  type: object
  properties:
    name:
      type: string
    age:
      type: integer
  required:
    - name
TestSchema:
  description: "A test schema with description"
  title: "Test Schema"
  properties:
    extraField:
      type: string
  oneOf:
    - $ref: "#/components/schemas/PersonObject"
    - type: "null"`,
			schemaName: "TestSchema",
			fixes:      withNameResolutionDec2023(),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("PersonObject").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("extraField", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
							WithOriginalName("extraField").
							WithOptional(true).
							WithJSONAnnotation("extraField").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(false).
							WithJSONAnnotation("name").
							Build(),
					).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "#/components/schemas/PersonObject",
					}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(false).
					WithContextStack(
						testutils.NewContextStack().WithOneOf("Test Schema").Build(),
					).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "MultipleTypesWithObjectProperties",
			schemasYAML: `TestSchema:
  type:
    - object
    - string
  properties:
    name:
      type: string
    age:
      type: integer
  required:
    - name`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithOriginalNameFrozen().
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithContextStack(
								testutils.NewContextStack().WithStandardOneOfContext("1").Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 16, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeString).
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "UnionsNotSupported",
			schemasYAML: `TestSchema:
  oneOf:
    - type: string
    - type: integer`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig().WithTarget(types.Target{Target: "go"})
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureUnions: false, // Disable unions
				}
				return config
			},
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithComplexAny(true).
					Build()).
				Build(),
		},
		{
			name: "EmptyOneOf",
			schemasYAML: `TestSchema:
  oneOf: []`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithComplexAny(false).
					Build()).
				Build(),
		},
		{
			name: "OneOfWithMergedReferencesNameResolutionDec2023",
			schemasYAML: `PersonObject:
  type: object
  properties:
    name:
      type: string
    age:
      type: integer
  required:
    - name
CompanyObject:
  type: object
  properties:
    companyName:
      type: string
    employees:
      type: integer
  required:
    - companyName
TestSchema:
  description: "A test schema with merged references"
  title: "Test Schema"
  properties:
    extraField:
      type: string
  oneOf:
    - $ref: "#/components/schemas/PersonObject"
    - $ref: "#/components/schemas/CompanyObject"`,
			schemaName: "TestSchema",
			fixes:      withNameResolutionDec2023(),
			expected: testutils.NewFieldDef("Test Schema",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithName("Test Schema").
					WithOriginalName("Test Schema").
					WithLocation(&yaml.Node{Line: 32, Column: 7}).
					WithContextStack(
						testutils.NewContextStack().WithOneOf("Test Schema").Build(),
					).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("PersonObject").
							WithOriginalName("PersonObject").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithOneOf("Test Schema").Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("extraField", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 31, Column: 11}).Build()).
									WithOriginalName("extraField").
									WithOptional(true).
									WithJSONAnnotation("extraField").
									Build(),
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
							).
							WithExtensions(map[string]interface{}{
								"x-speakeasy-reference-override": "#/components/schemas/PersonObject",
							}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(false).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("CompanyObject").
							WithOriginalName("CompanyObject").
							WithLocation(&yaml.Node{Line: 18, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithOneOf("Test Schema").Build(),
							).
							WithFields(
								testutils.NewFieldDef("companyName", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 11}).Build()).
									WithOriginalName("companyName").
									WithOptional(false).
									WithJSONAnnotation("companyName").
									Build(),
								testutils.NewFieldDef("employees", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 23, Column: 11}).Build()).
									WithOriginalName("employees").
									WithOptional(true).
									WithJSONAnnotation("employees").
									Build(),
								testutils.NewFieldDef("extraField", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 31, Column: 11}).Build()).
									WithOriginalName("extraField").
									WithOptional(true).
									WithJSONAnnotation("extraField").
									Build(),
							).
							WithExtensions(map[string]interface{}{
								"x-speakeasy-reference-override": "#/components/schemas/CompanyObject",
							}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(false).
							WithUsedInUnion().
							Build(),
					).
					WithComments(&ast.Comment{
						Description:      "A test schema with merged references",
						ExtendedComments: make(map[string]*ast.ExtendedComment),
					}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfWithMergedReferencesNameResolution",
			schemasYAML: `PersonObject:
  type: object
  properties:
    name:
      type: string
    age:
      type: integer
  required:
    - name
CompanyObject:
  type: object
  properties:
    companyName:
      type: string
    employees:
      type: integer
  required:
    - companyName
TestSchema:
  description: "A test schema with merged references"
  title: "Test Schema"
  properties:
    extraField:
      type: string
  oneOf:
    - $ref: "#/components/schemas/PersonObject"
    - $ref: "#/components/schemas/CompanyObject"`,
			schemaName: "TestSchema",
			// No fixes specified - uses default behavior (NameResolutionDec2023: false)
			expected: testutils.NewFieldDef("Test Schema",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithName("Test Schema").
					WithOriginalName("Test Schema").
					WithLocation(&yaml.Node{Line: 32, Column: 7}).
					WithContextStack(
						testutils.NewContextStack().WithOneOf("Test Schema").Build(),
					).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Schemas").
							WithOriginalName("Schemas").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("PersonObject").WithIdentifierForNaming(pointer.From("PersonObject")).Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("extraField", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 31, Column: 11}).Build()).
									WithOriginalName("extraField").
									WithOptional(true).
									WithJSONAnnotation("extraField").
									Build(),
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
							).
							WithExtensions(map[string]interface{}{
								"x-speakeasy-reference-override": "#/components/schemas/PersonObject",
							}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(false).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Schemas").
							WithOriginalName("Schemas").
							WithLocation(&yaml.Node{Line: 18, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("CompanyObject").WithIdentifierForNaming(pointer.From("CompanyObject")).Build(),
							).
							WithFields(
								testutils.NewFieldDef("companyName", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 11}).Build()).
									WithOriginalName("companyName").
									WithOptional(false).
									WithJSONAnnotation("companyName").
									Build(),
								testutils.NewFieldDef("employees", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 23, Column: 11}).Build()).
									WithOriginalName("employees").
									WithOptional(true).
									WithJSONAnnotation("employees").
									Build(),
								testutils.NewFieldDef("extraField", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 31, Column: 11}).Build()).
									WithOriginalName("extraField").
									WithOptional(true).
									WithJSONAnnotation("extraField").
									Build(),
							).
							WithExtensions(map[string]interface{}{
								"x-speakeasy-reference-override": "#/components/schemas/CompanyObject",
							}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(false).
							WithUsedInUnion().
							Build(),
					).
					WithComments(&ast.Comment{
						Description:      "A test schema with merged references",
						ExtendedComments: make(map[string]*ast.ExtendedComment),
					}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfWithDiscriminatorExplicitMapping",
			schemasYAML: `Dog:
  type: object
  properties:
    petType:
      type: string
    name:
      type: string
  required:
    - petType
    - name
Cat:
  type: object
  properties:
    petType:
      type: string
    age:
      type: integer
  required:
    - petType
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/Dog"
    - $ref: "#/components/schemas/Cat"
  discriminator:
    propertyName: petType
    mapping:
      dog: "#/components/schemas/Dog"
      cat: "#/components/schemas/Cat"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 28, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Dog").
							WithOriginalName("Dog").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
								testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("petType").
									WithOptional(false).
									WithJSONAnnotation("petType").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Cat").
							WithOriginalName("Cat").
							WithLocation(&yaml.Node{Line: 19, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 24, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
									WithOriginalName("petType").
									WithOptional(false).
									WithJSONAnnotation("petType").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
					).
					WithDiscriminator(testutils.NewDiscriminator("petType",
						testutils.NewDiscriminatorMapping("cat",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("Cat").
								WithOriginalName("Cat").
								WithLocation(&yaml.Node{Line: 19, Column: 7}).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithFields(
									testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 24, Column: 11}).Build()).
										WithOriginalName("age").
										WithOptional(true).
										WithJSONAnnotation("age").
										Build(),
									testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
										WithOriginalName("petType").
										WithOptional(false).
										WithJSONAnnotation("petType").
										Build(),
								).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								Build()),
						testutils.NewDiscriminatorMapping("dog",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("Dog").
								WithOriginalName("Dog").
								WithLocation(&yaml.Node{Line: 9, Column: 7}).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithFields(
									testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
										WithOriginalName("name").
										WithOptional(false).
										WithJSONAnnotation("name").
										Build(),
									testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
										WithOriginalName("petType").
										WithOptional(false).
										WithJSONAnnotation("petType").
										Build(),
								).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								Build()),
					)).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfWithDiscriminatorNoMapping",
			schemasYAML: `Dog:
  type: object
  properties:
    petType:
      type: string
    name:
      type: string
  required:
    - petType
    - name
Cat:
  type: object
  properties:
    petType:
      type: string
    age:
      type: integer
  required:
    - petType
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/Dog"
    - $ref: "#/components/schemas/Cat"
  discriminator:
    propertyName: petType`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 28, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Dog").
							WithOriginalName("Dog").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
								testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("petType").
									WithOptional(false).
									WithJSONAnnotation("petType").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Cat").
							WithOriginalName("Cat").
							WithLocation(&yaml.Node{Line: 19, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 24, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
									WithOriginalName("petType").
									WithOptional(false).
									WithJSONAnnotation("petType").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
					).
					WithDiscriminator(testutils.NewDiscriminator("petType",
						testutils.NewDiscriminatorMapping("Dog",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("Dog").
								WithOriginalName("Dog").
								WithLocation(&yaml.Node{Line: 9, Column: 7}).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithFields(
									testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
										WithOriginalName("name").
										WithOptional(false).
										WithJSONAnnotation("name").
										Build(),
									testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
										WithOriginalName("petType").
										WithOptional(false).
										WithJSONAnnotation("petType").
										Build(),
								).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								Build()),
						testutils.NewDiscriminatorMapping("Cat",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("Cat").
								WithOriginalName("Cat").
								WithLocation(&yaml.Node{Line: 19, Column: 7}).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithFields(
									testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 24, Column: 11}).Build()).
										WithOriginalName("age").
										WithOptional(true).
										WithJSONAnnotation("age").
										Build(),
									testutils.NewFieldDef("petType", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
										WithOriginalName("petType").
										WithOptional(false).
										WithJSONAnnotation("petType").
										Build(),
								).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								Build()),
					)).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfWithDiscriminatorEnumProperty",
			schemasYAML: `Dog:
  type: object
  properties:
    petType:
      type: string
      enum: ["Dog"]
    name:
      type: string
  required:
    - petType
    - name
Cat:
  type: object
  properties:
    petType:
      type: string
      enum: ["Cat"]
    age:
      type: integer
  required:
    - petType
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/Dog"
    - $ref: "#/components/schemas/Cat"
  discriminator:
    propertyName: petType`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 30, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Dog").
							WithOriginalName("Dog").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithComponent().
									WithConstProperty("Dog").WithIdentifierForNaming(nil).Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 15, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
								testutils.NewFieldDef("petType",
									testutils.NewEnumTypeDef(ast.DataTypeString, []string{"Dog"}, "enum").
										WithName("petType").
										WithOriginalName("petType").
										WithOriginalNameFrozen().
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithConstProperty("Dog").Build(),
										).
										WithEnumUnderlyingLocation(&yaml.Node{Line: 12, Column: 11}).
										WithScope(ast.ScopeShared).
										WithRegistered().
										Build()).
									WithOriginalName("petType").
									WithOptional(false).
									WithJSONAnnotation("petType").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("Cat").
							WithOriginalName("Cat").
							WithLocation(&yaml.Node{Line: 20, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithComponent().
									WithConstProperty("Cat").WithIdentifierForNaming(nil).Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 26, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("petType",
									testutils.NewEnumTypeDef(ast.DataTypeString, []string{"Cat"}, "enum").
										WithName("petType").
										WithOriginalName("petType").
										WithOriginalNameFrozen().
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithConstProperty("Cat").Build(),
										).
										WithEnumUnderlyingLocation(&yaml.Node{Line: 23, Column: 11}).
										WithScope(ast.ScopeShared).
										WithRegistered().
										Build()).
									WithOriginalName("petType").
									WithOptional(false).
									WithJSONAnnotation("petType").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
					).
					WithDiscriminator(testutils.NewDiscriminator("petType",
						testutils.NewDiscriminatorMapping("Dog",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("Dog").
								WithOriginalName("Dog").
								WithLocation(&yaml.Node{Line: 9, Column: 7}).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithComponent().
										WithConstProperty("Dog").WithIdentifierForNaming(nil).Build(),
								).
								WithFields(
									testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 15, Column: 11}).Build()).
										WithOriginalName("name").
										WithOptional(false).
										WithJSONAnnotation("name").
										Build(),
									testutils.NewFieldDef("petType",
										testutils.NewEnumTypeDef(ast.DataTypeString, []string{"Dog"}, "enum").
											WithName("petType").
											WithOriginalName("petType").
											WithOriginalNameFrozen().
											WithContextStack(
												testutils.NewContextStack().WithRefType("Schemas").WithRefName("Dog").WithIdentifierForNaming(nil).WithConstProperty("Dog").Build(),
											).
											WithEnumUnderlyingLocation(&yaml.Node{Line: 12, Column: 11}).
											WithScope(ast.ScopeShared).
											WithRegistered().
											Build()).
										WithOriginalName("petType").
										WithOptional(false).
										WithJSONAnnotation("petType").
										Build(),
								).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								Build()),
						testutils.NewDiscriminatorMapping("Cat",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("Cat").
								WithOriginalName("Cat").
								WithLocation(&yaml.Node{Line: 20, Column: 7}).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithComponent().
										WithConstProperty("Cat").WithIdentifierForNaming(nil).Build(),
								).
								WithFields(
									testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 26, Column: 11}).Build()).
										WithOriginalName("age").
										WithOptional(true).
										WithJSONAnnotation("age").
										Build(),
									testutils.NewFieldDef("petType",
										testutils.NewEnumTypeDef(ast.DataTypeString, []string{"Cat"}, "enum").
											WithName("petType").
											WithOriginalName("petType").
											WithOriginalNameFrozen().
											WithContextStack(
												testutils.NewContextStack().WithRefType("Schemas").WithRefName("Cat").WithIdentifierForNaming(nil).WithConstProperty("Cat").Build(),
											).
											WithEnumUnderlyingLocation(&yaml.Node{Line: 23, Column: 11}).
											WithScope(ast.ScopeShared).
											WithRegistered().
											Build()).
										WithOriginalName("petType").
										WithOptional(false).
										WithJSONAnnotation("petType").
										Build(),
								).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								Build()),
					)).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "InputOutputTypesRequest",
			schemasYAML: `PersonInput:
  type: object
  properties:
    name:
      type: string
      writeOnly: true
    age:
      type: integer
  required:
    - name
PersonOutput:
  type: object
  properties:
    name:
      type: string
      readOnly: true
    age:
      type: integer
  required:
    - name
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/PersonInput"
    - $ref: "#/components/schemas/PersonOutput"
`,
			schemaName: "TestSchema",
			isRequest:  pointer.From(true),
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 29, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("PersonInput").
							WithOriginalName("PersonInput").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonInput").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 15, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("PersonOutput").
							WithOriginalName("PersonOutput").
							WithLocation(&yaml.Node{Line: 19, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonOutput").WithIdentifierForNaming(nil).WithComponent().WithInputOutput("Input").Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 25, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							WithInput(true).
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithInputOutput("Input").Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "InputOutputTypesResponse",
			schemasYAML: `PersonInput:
  type: object
  properties:
    name:
      type: string
      writeOnly: true
    age:
      type: integer
  required:
    - name
PersonOutput:
  type: object
  properties:
    name:
      type: string
      readOnly: true
    age:
      type: integer
  required:
    - name
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/PersonInput"
    - $ref: "#/components/schemas/PersonOutput"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 29, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("PersonInput").
							WithOriginalName("PersonInput").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonInput").WithIdentifierForNaming(nil).WithComponent().WithInputOutput("Output").Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 15, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							WithOutput(true).
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("PersonOutput").
							WithOriginalName("PersonOutput").
							WithLocation(&yaml.Node{Line: 19, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonOutput").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 25, Column: 11}).Build()).
									WithOriginalName("age").
									WithOptional(true).
									WithJSONAnnotation("age").
									Build(),
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(false).
									WithJSONAnnotation("name").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithInputOutput("Output").Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "CanBeUnionTypeFalse",
			schemasYAML: `DuplicateType:
  type: object
  properties:
    name:
      type: string
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/DuplicateType"
    - $ref: "#/components/schemas/DuplicateType"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithLocation(&yaml.Node{Line: 14, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("DuplicateType").
							WithOriginalName("DuplicateType").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("DuplicateType").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(true).
									WithJSONAnnotation("name").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("DuplicateType").
							WithOriginalName("DuplicateType").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("DuplicateType").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
									WithOriginalName("name").
									WithOptional(true).
									WithJSONAnnotation("name").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "CircularReference",
			schemasYAML: `Node:
  type: object
  properties:
    value:
      type: string
    next:
      $ref: "#/components/schemas/Node"
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/Node"
    - type: "null"`,
			schemaName: "TestSchema",
			expected: func() *ast.FieldDef {
				// Create the Node type definition first
				nodeType := testutils.NewTypeDef(ast.DataTypeClass).
					WithName("Node").
					WithOriginalName("Node").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("Node").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithDeduplicatedContextStacks(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("Node").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					Build()

				// Add fields that reference the same type (circular reference)
				nodeType.Fields = append(nodeType.Fields,
					testutils.NewFieldDef("next", nodeType).
						WithOriginalName("next").
						WithOptional(true).
						WithJSONAnnotation("next").
						Build(),
					testutils.NewFieldDef("value", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
						WithOriginalName("value").
						WithOptional(true).
						WithJSONAnnotation("value").
						Build(),
				)

				return testutils.NewFieldDef("Node", nodeType).
					WithNullable(true).
					WithOptional(true).
					Build()
			}(),
		},
		{
			name: "TopLevelComponentReferenceWithOneOf",
			schemasYAML: `PersonType:
  oneOf:
    - type: object
      properties:
        name:
          type: string
        role:
          type: string
          enum: ["admin"]
    - type: object
      properties:
        name:
          type: string
        role:
          type: string
          enum: ["user"]
TestSchema:
  $ref: "#/components/schemas/PersonType"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("PersonType",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithName("PersonType").
					WithOriginalName("PersonType").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonType").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithLocation(&yaml.Node{Line: 10, Column: 11}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonType").WithIdentifierForNaming(nil).WithComponent().WithOneOfPosition("1", pointer.From("")).WithConstProperty("Admin").WithIdentifierForNaming(nil).Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 13, Column: 15}).Build()).
									WithOriginalName("name").
									WithOptional(true).
									WithJSONAnnotation("name").
									Build(),
								testutils.NewFieldDef("role",
									testutils.NewEnumTypeDef(ast.DataTypeString, []string{"admin"}, "enum").
										WithName("role").
										WithOriginalName("role").
										WithOriginalNameFrozen().
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonType").WithIdentifierForNaming(nil).WithComponent().WithOneOfPosition("1", pointer.From("")).WithConstProperty("admin").WithIdentifierForNaming(pointer.From("Admin")).Build(),
										).
										WithEnumUnderlyingLocation(&yaml.Node{Line: 15, Column: 15}).
										WithScope(ast.ScopeShared).
										WithRegistered().
										Build()).
									WithOriginalName("role").
									WithOptional(true).
									WithJSONAnnotation("role").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithUsedInUnion().
							Build(),
						testutils.NewTypeDef(ast.DataTypeClass).
							WithLocation(&yaml.Node{Line: 17, Column: 11}).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonType").WithIdentifierForNaming(nil).WithComponent().WithOneOfPosition("2", pointer.From("")).WithConstProperty("User").WithIdentifierForNaming(nil).Build(),
							).
							WithFields(
								testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
									WithOriginalName("name").
									WithOptional(true).
									WithJSONAnnotation("name").
									Build(),
								testutils.NewFieldDef("role",
									testutils.NewEnumTypeDef(ast.DataTypeString, []string{"user"}, "enum").
										WithName("role").
										WithOriginalName("role").
										WithOriginalNameFrozen().
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("PersonType").WithIdentifierForNaming(nil).WithComponent().WithOneOfPosition("2", pointer.From("")).WithConstProperty("user").WithIdentifierForNaming(pointer.From("User")).Build(),
										).
										WithEnumUnderlyingLocation(&yaml.Node{Line: 22, Column: 15}).
										WithScope(ast.ScopeShared).
										WithRegistered().
										Build()).
									WithOriginalName("role").
									WithOptional(true).
									WithJSONAnnotation("role").
									Build(),
							).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithUsedInUnion().
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "DiscriminatedUnionWithOneMember",
			schemasYAML: `ExhaustiveObject:
  type: object
  properties:
    str:
      type: string
      description: "A string property."
      example: "example"
    type:
      type: string
      default: "0"
  required:
    - str
    - type
DisriminatedUnionWithOneMember:
  oneOf:
    - $ref: "#/components/schemas/ExhaustiveObject"
  discriminator:
    propertyName: type
    mapping:
      type1: "#/components/schemas/ExhaustiveObject"
TestSchema:
  $ref: "#/components/schemas/DisriminatedUnionWithOneMember"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("DisriminatedUnionWithOneMember",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithName("DisriminatedUnionWithOneMember").
					WithOriginalName("DisriminatedUnionWithOneMember").
					WithLocation(&yaml.Node{Line: 22, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("DisriminatedUnionWithOneMember").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeClass).
							WithName("ExhaustiveObject").
							WithOriginalName("ExhaustiveObject").
							WithLocation(&yaml.Node{Line: 9, Column: 7}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithUsedInUnion().
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("ExhaustiveObject").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithFields(
								testutils.NewFieldDef("str", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).WithExamples(ast.NewExample("", "", &yaml.Node{
									Kind:   yaml.ScalarNode,
									Style:  yaml.DoubleQuotedStyle,
									Tag:    "!!str",
									Value:  "example",
									Line:   14,
									Column: 20,
								},
								)).Build()).
									WithOriginalName("str").
									WithOptional(false).
									WithJSONAnnotation("str").
									WithComments(&ast.Comment{
										Description:      "A string property.",
										ExtendedComments: make(map[string]*ast.ExtendedComment),
									}).
									Build(),
								testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 16, Column: 11}).Build()).
									WithOriginalName("type").
									WithOptional(true).
									WithJSONAnnotation("type").
									WithDefault(&ast.AnyValue{Value: "0"}).
									Build(),
							).
							Build(),
					).
					WithDiscriminator(testutils.NewDiscriminator("type",
						testutils.NewDiscriminatorMapping("type1",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("ExhaustiveObject").
								WithOriginalName("ExhaustiveObject").
								WithLocation(&yaml.Node{Line: 9, Column: 7}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithUsedInUnion().
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("ExhaustiveObject").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithFields(
									testutils.NewFieldDef("str", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).WithExamples(ast.NewExample("", "", &yaml.Node{
										Kind:   yaml.ScalarNode,
										Style:  yaml.DoubleQuotedStyle,
										Tag:    "!!str",
										Value:  "example",
										Line:   14,
										Column: 20,
									})).Build()).
										WithOriginalName("str").
										WithOptional(false).
										WithJSONAnnotation("str").
										WithComments(&ast.Comment{
											Description:      "A string property.",
											ExtendedComments: make(map[string]*ast.ExtendedComment),
										}).
										Build(),
									testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 16, Column: 11}).Build()).
										WithOriginalName("type").
										WithOptional(true).
										WithJSONAnnotation("type").
										WithDefault(&ast.AnyValue{Value: "0"}).
										Build(),
								).
								Build()),
					)).
					Build()).
				Build(),
		},
		{
			name: "UnionOfUnions",
			schemasYAML: `Foo:
    type: object
    properties:
      fooProp:
        type: string
    required:
      - fooProp
Bar:
  type: object
  properties:
    barProp:
      type: number
  required:
    - barProp
Bat:
  type: object
  properties:
    batProp:
      type: boolean
  required:
    - batProp
Baz:
  type: object
  properties:
    bazProp:
      type: array
      items:
        type: string
  required:
    - bazProp
FooOrBar:
  oneOf:
    - $ref: "#/components/schemas/Foo"
    - $ref: "#/components/schemas/Bar"
BatOrBaz:
  oneOf:
    - $ref: "#/components/schemas/Bat"
    - $ref: "#/components/schemas/Baz"
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/FooOrBar"
    - $ref: "#/components/schemas/BatOrBaz"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 47, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeUnion).
							WithName("FooOrBar").
							WithOriginalName("FooOrBar").
							WithLocation(&yaml.Node{Line: 39, Column: 7}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("FooOrBar").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithAssociatedTypes(
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Foo").
									WithOriginalName("Foo").
									WithLocation(&yaml.Node{Line: 9, Column: 9}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Foo").WithIdentifierForNaming(nil).WithComponent().Build(),
									).
									WithFields(
										testutils.NewFieldDef("fooProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 13}).Build()).
											WithOriginalName("fooProp").
											WithOptional(false).
											WithJSONAnnotation("fooProp").
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Bar").
									WithOriginalName("Bar").
									WithLocation(&yaml.Node{Line: 16, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Bar").WithIdentifierForNaming(nil).WithComponent().Build(),
									).
									WithFields(
										testutils.NewFieldDef("barProp", testutils.NewTypeDef(ast.DataTypeNumber).WithLocation(&yaml.Node{Line: 19, Column: 11}).Build()).
											WithOriginalName("barProp").
											WithOptional(false).
											WithJSONAnnotation("barProp").
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
							).
							Build(),
						testutils.NewTypeDef(ast.DataTypeUnion).
							WithName("BatOrBaz").
							WithOriginalName("BatOrBaz").
							WithLocation(&yaml.Node{Line: 43, Column: 7}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("BatOrBaz").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithAssociatedTypes(
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Bat").
									WithOriginalName("Bat").
									WithLocation(&yaml.Node{Line: 23, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Bat").WithIdentifierForNaming(nil).WithComponent().Build(),
									).
									WithFields(
										testutils.NewFieldDef("batProp", testutils.NewTypeDef(ast.DataTypeBoolean).Build()).
											WithOriginalName("batProp").
											WithOptional(false).
											WithJSONAnnotation("batProp").
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Baz").
									WithOriginalName("Baz").
									WithLocation(&yaml.Node{Line: 30, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Baz").WithIdentifierForNaming(nil).WithComponent().Build(),
									).
									WithFields(
										testutils.NewFieldDef("bazProp", testutils.NewTypeDef(ast.DataTypeArray).WithLocation(&yaml.Node{Line: 33, Column: 11}).WithItemType(testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 35, Column: 13}).Build()).WithValidations(&ast.Validations{}).Build()).
											WithOriginalName("bazProp").
											WithOptional(false).
											WithJSONAnnotation("bazProp").
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
							).
							Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "UnionOfUnionsDiscriminated",
			schemasYAML: `Foo:
  type: object
  properties:
    type:
      type: string
      const: "foo-or-bar"
    fooProp:
      type: string
  required:
    - foo
Bar:
  type: object
  properties:
    type:
      type: string
      const: "foo-or-bar"
    barProp:
      type: number
  required:
    - foo
Bat:
  type: object
  properties:
    type:
      type: string
      const: "bat-or-baz"
    batProp:
      type: boolean
  required:
    - batProp
Baz:
  type: object
  properties:
    type:
      type: string
      const: "bat-or-baz"
    bazProp:
      type: array
      items:
        type: string
  required:
    - bazProp
FooOrBar:
  oneOf:
    - $ref: "#/components/schemas/Foo"
    - $ref: "#/components/schemas/Bar"
BatOrBaz:
  oneOf:
    - $ref: "#/components/schemas/Bat"
    - $ref: "#/components/schemas/Baz"
TestSchema:
  oneOf:
    - $ref: "#/components/schemas/FooOrBar"
    - $ref: "#/components/schemas/BatOrBaz"
  discriminator:
    propertyName: type
    mapping:
      foo-or-bar: "#/components/schemas/FooOrBar"
      bat-or-baz: "#/components/schemas/BatOrBaz"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithLocation(&yaml.Node{Line: 59, Column: 7}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeUnion).
							WithName("FooOrBar").
							WithOriginalName("FooOrBar").
							WithLocation(&yaml.Node{Line: 51, Column: 7}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("FooOrBar").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithAssociatedTypes(
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Foo").
									WithOriginalName("Foo").
									WithLocation(&yaml.Node{Line: 9, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Foo").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("FooOrBar").WithIdentifierForNaming(nil).Build(),
									).
									WithFields(
										testutils.NewFieldDef("fooProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 15, Column: 11}).Build()).
											WithOriginalName("fooProp").
											WithOptional(true).
											WithJSONAnnotation("fooProp").
											Build(),
										testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
											WithOriginalName("type").
											WithOptional(true).
											WithJSONAnnotation("type").
											WithConst(&ast.AnyValue{Value: "foo-or-bar"}).
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Bar").
									WithOriginalName("Bar").
									WithLocation(&yaml.Node{Line: 19, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Bar").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("FooOrBar").WithIdentifierForNaming(nil).Build(),
									).
									WithFields(
										testutils.NewFieldDef("barProp", testutils.NewTypeDef(ast.DataTypeNumber).WithLocation(&yaml.Node{Line: 25, Column: 11}).Build()).
											WithOriginalName("barProp").
											WithOptional(true).
											WithJSONAnnotation("barProp").
											Build(),
										testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
											WithOriginalName("type").
											WithOptional(true).
											WithJSONAnnotation("type").
											WithConst(&ast.AnyValue{Value: "foo-or-bar"}).
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
							).
							Build(),
						testutils.NewTypeDef(ast.DataTypeUnion).
							WithName("BatOrBaz").
							WithOriginalName("BatOrBaz").
							WithLocation(&yaml.Node{Line: 55, Column: 7}).
							WithScope(ast.ScopeShared).
							WithRegistered().
							WithOriginalNameFrozen().
							WithIsComponent(true).
							WithContextStack(
								testutils.NewContextStack().WithRefType("Schemas").WithRefName("BatOrBaz").WithIdentifierForNaming(nil).WithComponent().Build(),
							).
							WithAssociatedTypes(
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Bat").
									WithOriginalName("Bat").
									WithLocation(&yaml.Node{Line: 29, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Bat").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("BatOrBaz").WithIdentifierForNaming(nil).Build(),
									).
									WithFields(
										testutils.NewFieldDef("batProp", testutils.NewTypeDef(ast.DataTypeBoolean).Build()).
											WithOriginalName("batProp").
											WithOptional(false).
											WithJSONAnnotation("batProp").
											Build(),
										testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 32, Column: 11}).Build()).
											WithOriginalName("type").
											WithOptional(true).
											WithJSONAnnotation("type").
											WithConst(&ast.AnyValue{Value: "bat-or-baz"}).
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
								testutils.NewTypeDef(ast.DataTypeClass).
									WithName("Baz").
									WithOriginalName("Baz").
									WithLocation(&yaml.Node{Line: 39, Column: 7}).
									WithContextStack(
										testutils.NewContextStack().WithRefType("Schemas").WithRefName("Baz").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("BatOrBaz").WithIdentifierForNaming(nil).Build(),
									).
									WithFields(
										testutils.NewFieldDef("bazProp", testutils.NewTypeDef(ast.DataTypeArray).WithLocation(&yaml.Node{Line: 45, Column: 11}).WithItemType(testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 47, Column: 13}).Build()).WithValidations(&ast.Validations{}).Build()).
											WithOriginalName("bazProp").
											WithOptional(false).
											WithJSONAnnotation("bazProp").
											Build(),
										testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 42, Column: 11}).Build()).
											WithOriginalName("type").
											WithOptional(true).
											WithJSONAnnotation("type").
											WithConst(&ast.AnyValue{Value: "bat-or-baz"}).
											Build(),
									).
									WithScope(ast.ScopeShared).
									WithRegistered().
									WithOriginalNameFrozen().
									WithIsComponent(true).
									WithUsedInUnion().
									Build(),
							).
							Build(),
					).
					WithDiscriminator(testutils.NewDiscriminator("type",
						testutils.NewDiscriminatorMapping("bat-or-baz",
							testutils.NewTypeDef(ast.DataTypeUnion).
								WithName("BatOrBaz").
								WithOriginalName("BatOrBaz").
								WithLocation(&yaml.Node{Line: 55, Column: 7}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("BatOrBaz").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithAssociatedTypes(
									testutils.NewTypeDef(ast.DataTypeClass).
										WithName("Bat").
										WithOriginalName("Bat").
										WithLocation(&yaml.Node{Line: 29, Column: 7}).
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("Bat").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("BatOrBaz").WithIdentifierForNaming(nil).Build(),
										).
										WithFields(
											testutils.NewFieldDef("batProp", testutils.NewTypeDef(ast.DataTypeBoolean).Build()).
												WithOriginalName("batProp").
												WithOptional(false).
												WithJSONAnnotation("batProp").
												Build(),
											testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 32, Column: 11}).Build()).
												WithOriginalName("type").
												WithOptional(true).
												WithJSONAnnotation("type").
												WithConst(&ast.AnyValue{Value: "bat-or-baz"}).
												Build(),
										).
										WithScope(ast.ScopeShared).
										WithRegistered().
										WithOriginalNameFrozen().
										WithIsComponent(true).
										WithUsedInUnion().
										Build(),
									testutils.NewTypeDef(ast.DataTypeClass).
										WithName("Baz").
										WithOriginalName("Baz").
										WithLocation(&yaml.Node{Line: 39, Column: 7}).
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("Baz").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("BatOrBaz").WithIdentifierForNaming(nil).Build(),
										).
										WithFields(
											testutils.NewFieldDef("bazProp", testutils.NewTypeDef(ast.DataTypeArray).WithLocation(&yaml.Node{Line: 45, Column: 11}).WithItemType(testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 47, Column: 13}).Build()).WithValidations(&ast.Validations{}).Build()).
												WithOriginalName("bazProp").
												WithOptional(false).
												WithJSONAnnotation("bazProp").
												Build(),
											testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 42, Column: 11}).Build()).
												WithOriginalName("type").
												WithOptional(true).
												WithJSONAnnotation("type").
												WithConst(&ast.AnyValue{Value: "bat-or-baz"}).
												Build(),
										).
										WithScope(ast.ScopeShared).
										WithRegistered().
										WithOriginalNameFrozen().
										WithIsComponent(true).
										WithUsedInUnion().
										Build(),
								).
								Build()),
						testutils.NewDiscriminatorMapping("foo-or-bar",
							testutils.NewTypeDef(ast.DataTypeUnion).
								WithName("FooOrBar").
								WithOriginalName("FooOrBar").
								WithLocation(&yaml.Node{Line: 51, Column: 7}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithIsComponent(true).
								WithContextStack(
									testutils.NewContextStack().WithRefType("Schemas").WithRefName("FooOrBar").WithIdentifierForNaming(nil).WithComponent().Build(),
								).
								WithAssociatedTypes(
									testutils.NewTypeDef(ast.DataTypeClass).
										WithName("Foo").
										WithOriginalName("Foo").
										WithLocation(&yaml.Node{Line: 9, Column: 7}).
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("Foo").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("FooOrBar").WithIdentifierForNaming(nil).Build(),
										).
										WithFields(
											testutils.NewFieldDef("fooProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 15, Column: 11}).Build()).
												WithOriginalName("fooProp").
												WithOptional(true).
												WithJSONAnnotation("fooProp").
												Build(),
											testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
												WithOriginalName("type").
												WithOptional(true).
												WithJSONAnnotation("type").
												WithConst(&ast.AnyValue{Value: "foo-or-bar"}).
												Build(),
										).
										WithScope(ast.ScopeShared).
										WithRegistered().
										WithOriginalNameFrozen().
										WithIsComponent(true).
										WithUsedInUnion().
										Build(),
									testutils.NewTypeDef(ast.DataTypeClass).
										WithName("Bar").
										WithOriginalName("Bar").
										WithLocation(&yaml.Node{Line: 19, Column: 7}).
										WithContextStack(
											testutils.NewContextStack().WithRefType("Schemas").WithRefName("Bar").WithIdentifierForNaming(nil).WithComponent().WithConstProperty("FooOrBar").WithIdentifierForNaming(nil).Build(),
										).
										WithFields(
											testutils.NewFieldDef("barProp", testutils.NewTypeDef(ast.DataTypeNumber).WithLocation(&yaml.Node{Line: 25, Column: 11}).Build()).
												WithOriginalName("barProp").
												WithOptional(true).
												WithJSONAnnotation("barProp").
												Build(),
											testutils.NewFieldDef("type", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 22, Column: 11}).Build()).
												WithOriginalName("type").
												WithOptional(true).
												WithJSONAnnotation("type").
												WithConst(&ast.AnyValue{Value: "foo-or-bar"}).
												Build(),
										).
										WithScope(ast.ScopeShared).
										WithRegistered().
										WithOriginalNameFrozen().
										WithIsComponent(true).
										WithUsedInUnion().
										Build(),
								).
								Build()),
					)).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfAllConstInt32FlattensToNamedEnum",
			schemasYAML: `TestSchema:
  title: "CommandType"
  type: integer
  format: int32
  oneOf:
    - const: 1
      title: "CHAT"
    - const: 2
      title: "USER"
    - const: 3
      title: "MESSAGE"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("CommandType",
				testutils.NewEnumTypeDef(ast.DataTypeInt32, []string{"1", "2", "3"}, "enum").
					WithName("CommandType").
					WithOriginalName("CommandType").
					WithEnumUnderlyingLocation(&yaml.Node{Line: 13, Column: 11}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					Build()).
				Build(),
		},
		{
			name: "OneOfSingleConstInt32RemainsTypedConst",
			schemasYAML: `TestSchema:
  type: integer
  format: int32
  oneOf:
    - const: 1
      title: "CHAT"`,
			schemaName: "TestSchema",
			mockFeatures: func() *testutils.MockFeaturesConfig {
				config := testutils.NewMockFeaturesConfig()
				config.SupportedFeatures = map[features.Feature]bool{
					features.FeatureConstsAndDefaults: true,
				}
				return config
			},
			expected: testutils.NewFieldDef("int32",
				testutils.NewTypeDef(ast.DataTypeInt32).
					WithLocation(&yaml.Node{Line: 12, Column: 11}).
					WithValidations(&ast.Validations{}).
					Build()).
				WithConst(&ast.AnyValue{Value: 1}).
				Build(),
		},
		// TODO: Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment using common testutils
			fullYAML := testutils.CreateOpenAPIDoc(tt.schemasYAML)

			var mockFeatures *testutils.MockFeaturesConfig
			if tt.mockFeatures != nil {
				mockFeatures = tt.mockFeatures()
			} else {
				mockFeatures = testutils.NewMockFeaturesConfig().WithSupportAllFeatures()
			}

			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  fullYAML,
				Fixes:        tt.fixes,
				MockFeatures: mockFeatures,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Create the Schemas instance
			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			// Get the test schema
			testSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get(tt.schemaName)
			require.True(t, exists)
			require.NotNil(t, testSchema)

			// Create test parameters
			isRequest := false
			if tt.isRequest != nil {
				isRequest = *tt.isRequest
			}

			params := Params{
				ContextStack:        ast.ContextStack{},
				SerializationMethod: ast.SerializationMethodJSON,
				Schema:              testSchema,
				Scope:               ast.ScopeShared,
				IsRequest:           isRequest,
				Depth:               0,
				MaxDepth:            10,
				Parents:             []string{},
				TypeDefCache:        make(map[string]*ast.TypeDef),
				LoopContext:         []LoopFrame{},
				Nullable:            false,
				CircularReference:   false,
				DocInfo:             common.DocInfo,
			}

			// Call the top-level HandleSchema method instead of handleAnyOfOneOf directly
			ctx := context.Background()
			result, err := schemas.HandleSchema(ctx, params)

			// Verify the result
			require.NoError(t, err)
			require.NotNil(t, result)

			// Use our deep comparison function to find exact differences
			isEqual := testutils.DeepCompare(t, tt.expected, result)
			if !isEqual {
				t.Errorf("Deep comparison failed for test case: %s", tt.name)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHandleAnyOfOneOfPathParamPrimitiveUnion verifies that a path parameter
// whose members all resolve to the same primitive collapses to that primitive,
// while the same union elsewhere is still generated as a union.
func TestHandleAnyOfOneOfPathParamPrimitiveUnion(t *testing.T) {
	// The enum member is deliberately covered in both positions: the flattened
	// schema is copied from the first member, so only the enum-first ordering
	// exercises dropping the inherited enum constraint.
	const enumLastYAML = `MeLiteral:
  type: string
  enum:
    - me
TestSchema:
  oneOf:
    - type: string
      pattern: "^pfl_.+$"
    - "$ref": "#/components/schemas/MeLiteral"`

	const enumFirstYAML = `MeLiteral:
  type: string
  enum:
    - me
TestSchema:
  oneOf:
    - "$ref": "#/components/schemas/MeLiteral"
    - type: string
      pattern: "^pfl_.+$"`

	const inlineEnumFirstYAML = `TestSchema:
  oneOf:
    - type: string
      enum:
        - me
    - type: string
      pattern: "^pfl_.+$"`

	const anyOfEnumFirstYAML = `TestSchema:
  anyOf:
    - type: string
      enum:
        - me
    - type: string
      pattern: "^pfl_.+$"`

	const allEnumYAML = `TestSchema:
  oneOf:
    - type: string
      enum:
        - me
    - type: string
      enum:
        - me
        - all`

	const singleEnumYAML = `MeLiteral:
  type: string
  enum:
    - me
    - all
TestSchema:
  oneOf:
    - "$ref": "#/components/schemas/MeLiteral"`

	const reorderedEnumYAML = `TestSchema:
  oneOf:
    - type: string
      enum:
        - first
        - second
    - type: string
      enum:
        - second
        - first`

	const mixedPrimitiveEnumYAML = `TestSchema:
  oneOf:
    - type: string
      enum:
        - me
    - type: integer
      enum:
        - 1
        - 2`

	const nullableOpenEnumYAML = `TestSchema:
  oneOf:
    - type: string
      enum:
        - me
    - type: string
      pattern: "^pfl_.+$"
    - type: "null"`

	const nullableClosedEnumYAML = `TestSchema:
  oneOf:
    - type: string
      enum:
        - me
        - all
    - type: "null"`

	const constAndEnumYAML = `TestSchema:
  oneOf:
    - const: one
    - type: string
      enum:
        - two
        - three`

	const constAndOpenStringYAML = `TestSchema:
  oneOf:
    - const: one
    - type: string
      pattern: "^pfl_.+$"`

	const refConstAndEnumYAML = `FreeLiteral:
  const: free
TestSchema:
  oneOf:
    - "$ref": "#/components/schemas/FreeLiteral"
    - type: string
      enum:
        - pro
        - enterprise`

	const refConstAndOpenStringYAML = `FreeLiteral:
  const: free
TestSchema:
  oneOf:
    - "$ref": "#/components/schemas/FreeLiteral"
    - type: string
      pattern: "^pfl_.+$"`

	handle := func(t *testing.T, schemasYAML string, paramType ParamType) *ast.FieldDef {
		t.Helper()

		common, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
			OpenAPIYAML:  testutils.CreateOpenAPIDoc(schemasYAML),
			MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
		})
		require.NoError(t, err)

		schemas := &Schemas{
			Config:    common.Config,
			Target:    common.Target,
			Subsystem: common.Subsystem,
			Namer:     common.Namer,
		}

		testSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get("TestSchema")
		require.True(t, exists)
		require.NotNil(t, testSchema)

		result, err := schemas.HandleSchema(context.Background(), Params{
			ContextStack:        ast.ContextStack{},
			SerializationMethod: ast.SerializationMethodJSON,
			Schema:              testSchema,
			Scope:               ast.ScopeShared,
			IsRequest:           true,
			Depth:               0,
			MaxDepth:            10,
			Parents:             []string{},
			TypeDefCache:        make(map[string]*ast.TypeDef),
			LoopContext:         []LoopFrame{},
			DocInfo:             common.DocInfo,
			ParamType:           paramType,
		})
		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotNil(t, result.Type)
		return result
	}

	t.Run("PathParamCollapsesToPrimitiveEnumLast", func(t *testing.T) {
		result := handle(t, enumLastYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Empty(t, result.Type.AssociatedTypes)
		assert.Nil(t, result.Type.Enum)
	})

	t.Run("PathParamCollapsesToPrimitiveEnumFirst", func(t *testing.T) {
		result := handle(t, enumFirstYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Empty(t, result.Type.AssociatedTypes)
		assert.Nil(t, result.Type.Enum)
	})

	t.Run("PathParamCollapsesToPrimitiveInlineEnumFirst", func(t *testing.T) {
		result := handle(t, inlineEnumFirstYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Empty(t, result.Type.AssociatedTypes)
		assert.Nil(t, result.Type.Enum)
	})

	t.Run("PathParamCollapsesToPrimitiveAnyOf", func(t *testing.T) {
		result := handle(t, anyOfEnumFirstYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Empty(t, result.Type.AssociatedTypes)
		assert.Nil(t, result.Type.Enum)
	})

	t.Run("PathParamMergesEnumMembers", func(t *testing.T) {
		result := handle(t, allEnumYAML, ParamTypePath)
		require.NotNil(t, result.Type.Enum)
		assert.Equal(t, []string{"me", "all"}, result.Type.Enum.Values)
	})

	t.Run("PathParamKeepsSingleMemberEnum", func(t *testing.T) {
		result := handle(t, singleEnumYAML, ParamTypePath)
		require.NotNil(t, result.Type.Enum)
		assert.Equal(t, []string{"me", "all"}, result.Type.Enum.Values)
	})

	t.Run("PathParamMergesEnumMembersIgnoringOrder", func(t *testing.T) {
		result := handle(t, reorderedEnumYAML, ParamTypePath)
		require.NotNil(t, result.Type.Enum)
		assert.Equal(t, []string{"first", "second"}, result.Type.Enum.Values)
	})

	// Members of different primitive types can't collapse to a single
	// parameter, so the union survives and the operation is still skipped.
	t.Run("PathParamMixedPrimitiveEnumsRemainUnion", func(t *testing.T) {
		result := handle(t, mixedPrimitiveEnumYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeUnion, result.Type.Type)
	})

	t.Run("PathParamCollapsesNullableOpenUnion", func(t *testing.T) {
		result := handle(t, nullableOpenEnumYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Nil(t, result.Type.Enum)
		assert.True(t, result.Nullable)
	})

	// A single enum member alongside null keeps its constraint: it still
	// describes the whole union.
	t.Run("PathParamKeepsNullableClosedEnum", func(t *testing.T) {
		result := handle(t, nullableClosedEnumYAML, ParamTypePath)
		require.NotNil(t, result.Type.Enum)
		assert.Equal(t, []string{"me", "all"}, result.Type.Enum.Values)
		assert.True(t, result.Nullable)
	})

	// A const describes a closed set of one, so a const alongside an enum
	// merges into the combined set rather than leaking the first member's
	// const onto the collapsed parameter.
	t.Run("PathParamMergesConstWithEnum", func(t *testing.T) {
		result := handle(t, constAndEnumYAML, ParamTypePath)
		assert.Nil(t, result.Const)
		require.NotNil(t, result.Type.Enum)
		assert.Equal(t, []string{"one", "two", "three"}, result.Type.Enum.Values)
	})

	// A const paired with an unconstrained string is an open union, so both
	// the const and the enum have to be dropped.
	t.Run("PathParamDropsConstForOpenUnion", func(t *testing.T) {
		result := handle(t, constAndOpenStringYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Nil(t, result.Const)
		assert.Nil(t, result.Type.Enum)
	})

	// A referenced const carries its value on the component it points at, so
	// the resolved member has to be inlined before the constraint is rewritten.
	t.Run("PathParamMergesReferencedConstWithEnum", func(t *testing.T) {
		result := handle(t, refConstAndEnumYAML, ParamTypePath)
		assert.Nil(t, result.Const)
		require.NotNil(t, result.Type.Enum)
		assert.Equal(t, []string{"free", "pro", "enterprise"}, result.Type.Enum.Values)
	})

	t.Run("PathParamDropsReferencedConstForOpenUnion", func(t *testing.T) {
		result := handle(t, refConstAndOpenStringYAML, ParamTypePath)
		assert.Equal(t, ast.DataTypeString, result.Type.Type)
		assert.Nil(t, result.Const)
		assert.Nil(t, result.Type.Enum)
	})

	t.Run("NonPathParamRemainsUnion", func(t *testing.T) {
		for _, paramType := range []ParamType{"", ParamTypeQuery, ParamTypeHeader} {
			result := handle(t, enumLastYAML, paramType)
			assert.Equal(t, ast.DataTypeUnion, result.Type.Type, "paramType %q", paramType)
		}
	})
}

// Helper function to create custom fixes configuration
func withNameResolutionDec2023() *config.Fixes {
	return &config.Fixes{
		NameResolutionFeb2025:                true,
		NameResolutionDec2023:                true, // Override to true
		RequestResponseComponentNamesFeb2024: false,
	}
}
