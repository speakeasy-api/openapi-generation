package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleAllOf tests the handleAllOf method with various allOf scenarios
func TestHandleAllOf(t *testing.T) {
	tests := []allOfTestCase{
		{
			name: "SingleAllOfString",
			schemasYAML: `TestSchema:
  allOf:
    - type: string`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					Build()).
				Build(),
		},
		{
			name: "SingleAllOfObject",
			schemasYAML: `TestSchema:
  allOf:
    - type: object
      properties:
        name:
          type: string
        age:
          type: integer`,
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
				Build(),
		},
		{
			name: "SingleAllOfReference",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("BaseSchema",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("BaseSchema").
					WithOriginalName("BaseSchema").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("BaseSchema").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithDescriptionOverride",
			schemasYAML: `TestSchema:
  allOf:
    - type: string
    - description: "A string with custom description"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 10, Column: 11}).
					Build()).
				Build(),
		},
		{
			name: "AllOfMergeObjects",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
    name:
      type: string
ExtendedSchema:
  type: object
  properties:
    age:
      type: integer
    email:
      type: string
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - $ref: "#/components/schemas/ExtendedSchema"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 23, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/BaseSchema,#/components/schemas/ExtendedSchema]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 19, Column: 11}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("email", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 11}).Build()).
							WithOriginalName("email").
							WithOptional(true).
							WithJSONAnnotation("email").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfMergeWithInlineProperties",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
TestSchema:
  type: object
  properties:
    name:
      type: string
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        age:
          type: integer`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 14, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "44e38c29eb5e8489-allOf[#/components/schemas/BaseSchema,0032b2b30ac5ee82]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 23, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 17, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithRequiredFields",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
  required:
    - id
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        name:
          type: string
      required:
        - name`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 16, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/BaseSchema,d8072bb8c5d8a2ef]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(false).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(false).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNullable",
			schemasYAML: `TestSchema:
  allOf:
    - type: string
  nullable: true`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "23f0fa9ca7e4fef6-allOf[9cfeba6fdf2261c3]",
					}).
					Build()).
				WithNullable(true).
				WithOptional(true).
				Build(),
		},
		{
			name: "AllOfMergePrimitiveTypes",
			schemasYAML: `TestSchema:
  allOf:
    - type: string
      minLength: 1
    - type: string
      maxLength: 10
      pattern: "^[a-z]+$"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[03797e2df1c4e4bc,36286511f6a09863]",
					}).
					WithValidations(&ast.Validations{
						MinLength: func() *int64 { v := int64(1); return &v }(),
						MaxLength: func() *int64 { v := int64(10); return &v }(),
						Pattern:   func() *string { v := "^[a-z]+$"; return &v }(),
					}).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithEnum",
			schemasYAML: `TestSchema:
  allOf:
    - type: string
    - enum:
        - "active"
        - "inactive"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("enum",
				testutils.NewEnumTypeDef(ast.DataTypeString, []string{"active", "inactive"}, "enum").
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithEnumUnderlyingLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[65c0b3f1276ee5ee,9cfeba6fdf2261c3]",
					}).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithAdditionalProperties",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
  additionalProperties: true`,
			schemaName: "TestSchema",
			expected: func() *ast.FieldDef {
				// Create the map type with any item type
				mapType := testutils.NewTypeDef(ast.DataTypeMap).
					WithLocation(&yaml.Node{Line: 14, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "1a2e46e344784173-allOf[#/components/schemas/BaseSchema]",
					}).
					Build()
				mapType.ItemType = testutils.NewTypeDef(ast.DataTypeAny).Build()

				// Create the additional properties field
				additionalPropsField := testutils.NewFieldDef("AdditionalProperties", mapType).
					WithOptional(true).
					WithAnnotations(&ast.JSONAnnotation{Ignore: true, FieldName: "-"}, &ast.NeedsCasingAnnotation{}).
					Build()
				additionalPropsField.IsAdditionalProperties = true

				return testutils.NewFieldDef("object",
					testutils.NewTypeDef(ast.DataTypeClass).
						WithLocation(&yaml.Node{Line: 14, Column: 7}).
						WithScope(ast.ScopeShared).
						WithRegistered().
						WithOriginalNameFrozen().
						WithContextStack(ast.ContextStack{}).
						WithExtensions(map[string]interface{}{
							"x-speakeasy-reference-override": "1a2e46e344784173-allOf[#/components/schemas/BaseSchema]",
						}).
						WithFields(
							additionalPropsField,
							testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
								WithOriginalName("id").
								WithOptional(true).
								WithJSONAnnotation("id").
								Build(),
						).
						Build()).
					Build()
			}(),
		},
		{
			name: "AllOfEmptyArray",
			schemasYAML: `TestSchema:
  allOf: []`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("any",
				testutils.NewTypeDef(ast.DataTypeAny).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithComplexAny(false).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithCircularReference",
			schemasYAML: `Node:
  type: object
  properties:
    value:
      type: string
    next:
      $ref: "#/components/schemas/Node"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/Node"`,
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

				return testutils.NewFieldDef("Node", nodeType).Build()
			}(),
		},
		{
			name: "AllOfWithInputOutputTypes",
			schemasYAML: `BaseInput:
  type: object
  properties:
    name:
      type: string
      writeOnly: true
    id:
      type: string
  required:
    - name
BaseOutput:
  type: object
  properties:
    name:
      type: string
      readOnly: true
    id:
      type: string
  required:
    - name
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseInput"
    - $ref: "#/components/schemas/BaseOutput"`,
			schemaName: "TestSchema",
			isRequest:  pointer.From(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 29, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithInputOutput("Input").Build(),
					).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/BaseInput,#/components/schemas/BaseOutput]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 25, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
					).
					WithInput(true).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithConflictingTypes",
			schemasYAML: `TestSchema:
  allOf:
    - type: string
    - type: integer`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("oneOf",
				testutils.NewTypeDef(ast.DataTypeUnion).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithAssociatedTypes(
						testutils.NewTypeDef(ast.DataTypeString).Build(),
						testutils.NewTypeDef(ast.DataTypeInteger).Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNestedAllOf",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
NestedSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        name:
          type: string
TestSchema:
  allOf:
    - $ref: "#/components/schemas/NestedSchema"
    - type: object
      properties:
        age:
          type: integer`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 21, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "42eeecad5ca81fe2-allOf[#/components/schemas/BaseSchema,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 26, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 19, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithValidations",
			schemasYAML: `TestSchema:
  allOf:
    - type: integer
      minimum: 0
    - type: integer
      maximum: 100`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("integer",
				testutils.NewTypeDef(ast.DataTypeInteger).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[42d1454f7adecc80,7c32ac9492a85979]",
					}).
					WithValidations(&ast.Validations{
						Minimum: func() *float64 { v := float64(0); return &v }(),
						Maximum: func() *float64 { v := float64(100); return &v }(),
					}).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithTitle",
			schemasYAML: `TestSchema:
  title: "Test Schema Title"
  allOf:
    - type: string`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("Test Schema Title",
				testutils.NewTypeDef(ast.DataTypeString).
					WithName("Test Schema Title").
					WithOriginalName("").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "1ac43ecd59ad3b14-allOf[9cfeba6fdf2261c3]",
					}).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithExample",
			schemasYAML: `TestSchema:
  allOf:
    - type: string
  example: "test value"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("string",
				testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "63f93381bfa9315a-allOf[9cfeba6fdf2261c3]",
					}).
					WithExamples(&ast.Example{
						Value: &yaml.Node{
							Kind:   yaml.ScalarNode,
							Style:  yaml.DoubleQuotedStyle,
							Tag:    "!!str",
							Value:  "test value",
							Line:   11,
							Column: 16,
						},
					}).
					Build()).
				Build(),
		},
		{
			name: "TopLevelComponentReferenceWithAllOf",
			schemasYAML: `BaseUser:
  type: object
  properties:
    name:
      type: string
User:
  allOf:
    - $ref: "#/components/schemas/BaseUser"
    - required: ["name"]
TestSchema:
  $ref: "#/components/schemas/User"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("User",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("User").
					WithOriginalName("User").
					WithLocation(&yaml.Node{Line: 14, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("User").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "#/components/schemas/User-allOf[#/components/schemas/BaseUser,87e5519238c53c12]",
					}).
					WithFields(
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(false).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithParentNameOverride",
			schemasYAML: `TestSchema:
  x-speakeasy-name-override: "ParentOverride"
  allOf:
    - type: object
      properties:
        id:
          type: string
    - type: object
      properties:
        name:
          type: string
      x-speakeasy-group: "child-group"`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("ParentOverride",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("ParentOverride").
					WithOriginalName("ParentOverride").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "ParentOverride",
						"x-speakeasy-group":              "child-group",
						"x-speakeasy-reference-override": "1b5bec8e7a4cfdf3-allOf[340d40e0d55e80f3,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 15}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 18, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithExtensionHoistingTwoLevelsDeep_NameOverrideBleeds",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
  x-speakeasy-name-override: "BaseOverride"
MiddleSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        name:
          type: string
      x-speakeasy-group: "middle-group"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/MiddleSchema"
    - type: object
      properties:
        age:
          type: integer
      x-speakeasy-example: "test-example"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("BaseOverride",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("BaseOverride").
					WithOriginalName("BaseOverride").
					WithLocation(&yaml.Node{Line: 23, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "BaseOverride",
						"x-speakeasy-group":              "middle-group",
						"x-speakeasy-example":            "test-example",
						"x-speakeasy-reference-override": "42eeecad5ca81fe2-allOf[#/components/schemas/BaseSchema,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 28, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithExtensionHoistingTwoLevelsDeep_NameOverrideIgnored",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
  x-speakeasy-name-override: "BaseOverride"
MiddleSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        name:
          type: string
      x-speakeasy-group: "middle-group"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/MiddleSchema"
    - type: object
      properties:
        age:
          type: integer
      x-speakeasy-example: "test-example"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 23, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-group":              "middle-group",
						"x-speakeasy-example":            "test-example",
						"x-speakeasy-reference-override": "42eeecad5ca81fe2-allOf[#/components/schemas/BaseSchema,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 28, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithExtensionHoistingCircularReference_NameOverrideBleeds",
			schemasYAML: `NodeWithExtensions:
  type: object
  properties:
    value:
      type: string
    next:
      $ref: "#/components/schemas/NodeWithExtensions"
  x-speakeasy-name-override: "CustomNode"
  x-speakeasy-group: "node-group"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/NodeWithExtensions"
    - type: object
      properties:
        metadata:
          type: string
      x-speakeasy-example: "circular-example"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: func() *ast.FieldDef {
				// Create the NodeWithExtensions type definition first
				nodeType := testutils.NewTypeDef(ast.DataTypeClass).
					WithName("CustomNode").
					WithOriginalName("CustomNode").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("CustomNode").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithDeduplicatedContextStacks(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("CustomNode").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override": "CustomNode",
						"x-speakeasy-group":         "node-group",
					}).
					Build()

				// Add fields that reference the same type (circular reference)
				nodeType.Fields = append(nodeType.Fields,
					testutils.NewFieldDef("CustomNode", nodeType).
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

				// Create the merged schema with hoisted extensions and input context
				mergedType := testutils.NewTypeDef(ast.DataTypeClass).
					WithName("CustomNode").
					WithOriginalName("CustomNode").
					WithLocation(&yaml.Node{Line: 18, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "CustomNode",
						"x-speakeasy-group":              "node-group",
						"x-speakeasy-example":            "circular-example",
						"x-speakeasy-reference-override": "3227167712ca05a9-allOf[#/components/schemas/NodeWithExtensions,51e94c7eb203a44b]",
					}).
					WithFields(
						testutils.NewFieldDef("CustomNode", nodeType).
							WithOriginalName("next").
							WithOptional(true).
							WithJSONAnnotation("next").
							Build(),
						testutils.NewFieldDef("metadata", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 23, Column: 15}).Build()).
							WithOriginalName("metadata").
							WithOptional(true).
							WithJSONAnnotation("metadata").
							Build(),
						testutils.NewFieldDef("value", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("value").
							WithOptional(true).
							WithJSONAnnotation("value").
							Build(),
					).
					Build()

				return testutils.NewFieldDef("CustomNode", mergedType).Build()
			}(),
		},
		{
			name: "AllOfWithExtensionHoistingCircularReference_NameOverrideIgnored",
			schemasYAML: `NodeWithExtensions:
  type: object
  properties:
    value:
      type: string
    next:
      $ref: "#/components/schemas/NodeWithExtensions"
  x-speakeasy-name-override: "CustomNode"
  x-speakeasy-group: "node-group"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/NodeWithExtensions"
    - type: object
      properties:
        metadata:
          type: string
      x-speakeasy-example: "circular-example"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: func() *ast.FieldDef {
				nodeType := testutils.NewTypeDef(ast.DataTypeClass).
					WithName("CustomNode").
					WithOriginalName("CustomNode").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("CustomNode").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithDeduplicatedContextStacks(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("CustomNode").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override": "CustomNode",
						"x-speakeasy-group":         "node-group",
					}).
					Build()

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

				mergedType := testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 18, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-group":              "node-group",
						"x-speakeasy-example":            "circular-example",
						"x-speakeasy-reference-override": "3227167712ca05a9-allOf[#/components/schemas/NodeWithExtensions,51e94c7eb203a44b]",
					}).
					WithFields(
						testutils.NewFieldDef("metadata", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 23, Column: 15}).Build()).
							WithOriginalName("metadata").
							WithOptional(true).
							WithJSONAnnotation("metadata").
							Build(),
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
					).
					Build()

				return testutils.NewFieldDef("object", mergedType).Build()
			}(),
		},
		{
			name: "AllOfWithExtensionHoistingThreeLevelsDeep_NameOverrideBleeds",
			schemasYAML: `Level1Schema:
  type: object
  properties:
    level1:
      type: string
  x-speakeasy-name-override: "Level1Override"
Level2Schema:
  allOf:
    - $ref: "#/components/schemas/Level1Schema"
    - type: object
      properties:
        level2:
          type: string
      x-speakeasy-group: "level2-group"
Level3Schema:
  allOf:
    - $ref: "#/components/schemas/Level2Schema"
    - type: object
      properties:
        level3:
          type: string
      x-speakeasy-example: "level3-example"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/Level3Schema"
    - type: object
      properties:
        final:
          type: string
      x-speakeasy-docs: "final-docs"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("Level1Override",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("Level1Override").
					WithOriginalName("Level1Override").
					WithLocation(&yaml.Node{Line: 31, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "Level1Override",
						"x-speakeasy-group":              "level2-group",
						"x-speakeasy-example":            "level3-example",
						"x-speakeasy-docs":               "final-docs",
						"x-speakeasy-reference-override": "13cbbd645c330b0e-allOf[#/components/schemas/Level1Schema,70dc8779a6142df0]",
					}).
					WithFields(
						testutils.NewFieldDef("final", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 36, Column: 15}).Build()).
							WithOriginalName("final").
							WithOptional(true).
							WithJSONAnnotation("final").
							Build(),
						testutils.NewFieldDef("level1", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("level1").
							WithOptional(true).
							WithJSONAnnotation("level1").
							Build(),
						testutils.NewFieldDef("level2", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
							WithOriginalName("level2").
							WithOptional(true).
							WithJSONAnnotation("level2").
							Build(),
						testutils.NewFieldDef("level3", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 28, Column: 15}).Build()).
							WithOriginalName("level3").
							WithOptional(true).
							WithJSONAnnotation("level3").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithExtensionHoistingThreeLevelsDeep_NameOverrideIgnored",
			schemasYAML: `Level1Schema:
  type: object
  properties:
    level1:
      type: string
  x-speakeasy-name-override: "Level1Override"
Level2Schema:
  allOf:
    - $ref: "#/components/schemas/Level1Schema"
    - type: object
      properties:
        level2:
          type: string
      x-speakeasy-group: "level2-group"
Level3Schema:
  allOf:
    - $ref: "#/components/schemas/Level2Schema"
    - type: object
      properties:
        level3:
          type: string
      x-speakeasy-example: "level3-example"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/Level3Schema"
    - type: object
      properties:
        final:
          type: string
      x-speakeasy-docs: "final-docs"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 31, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-group":              "level2-group",
						"x-speakeasy-example":            "level3-example",
						"x-speakeasy-docs":               "final-docs",
						"x-speakeasy-reference-override": "13cbbd645c330b0e-allOf[#/components/schemas/Level1Schema,70dc8779a6142df0]",
					}).
					WithFields(
						testutils.NewFieldDef("final", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 36, Column: 15}).Build()).
							WithOriginalName("final").
							WithOptional(true).
							WithJSONAnnotation("final").
							Build(),
						testutils.NewFieldDef("level1", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("level1").
							WithOptional(true).
							WithJSONAnnotation("level1").
							Build(),
						testutils.NewFieldDef("level2", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
							WithOriginalName("level2").
							WithOptional(true).
							WithJSONAnnotation("level2").
							Build(),
						testutils.NewFieldDef("level3", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 28, Column: 15}).Build()).
							WithOriginalName("level3").
							WithOptional(true).
							WithJSONAnnotation("level3").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNonMergableExtensions_NameOverrideBleeds",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
  x-speakeasy-include: true
  x-speakeasy-name-override: "BaseOverride"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        name:
          type: string
      x-speakeasy-group: "test-group"
      x-custom-extension: "should-not-hoist"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("BaseOverride",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("BaseOverride").
					WithOriginalName("BaseOverride").
					WithLocation(&yaml.Node{Line: 16, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						// x-speakeasy-include should not be hoisted (blacklisted)
						// x-custom-extension should not be hoisted (not x-speakeasy-*)
						"x-speakeasy-name-override":      "BaseOverride",
						"x-speakeasy-group":              "test-group",
						"x-speakeasy-reference-override": "c0a212b0f6a820f7-allOf[#/components/schemas/BaseSchema,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNonMergableExtensions_NameOverrideIgnored",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    id:
      type: string
  x-speakeasy-include: true
  x-speakeasy-name-override: "BaseOverride"
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        name:
          type: string
      x-speakeasy-group: "test-group"
      x-custom-extension: "should-not-hoist"`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 16, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						// x-speakeasy-include should not be hoisted (blacklisted)
						// x-custom-extension should not be hoisted (not x-speakeasy-*)
						// x-speakeasy-name-override should not be hoisted (fix enabled)
						"x-speakeasy-group":              "test-group",
						"x-speakeasy-reference-override": "c0a212b0f6a820f7-allOf[#/components/schemas/BaseSchema,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithFactoredOutProperties",
			schemasYAML: `TestSchema:
  $ref: "#/components/schemas/AllOfWithFactoredOutProperties"
AllOfWithFactoredOutProperties:
  description: "An object with allOf and factored out properties."
  properties:
    anExtraProperty:
      type: string
      description: "An extra property."
      example: "example"
    anOverridingProperty:
      type: string
      description: "An overriding property."
      example: "example"
  allOf:
    - type: object
      properties:
        anOverridingProperty:
          type: integer
          description: "An overriding property."
          example: 1
    - type: object
      properties:
        anOverridingProperty:
          type: number
          description: "An overriding property."
          example: 1.1
        anotherProperty:
          type: string
          description: "Another property."
          example: "example"`,
			schemaName:           "TestSchema",
			maintainOpenAPIOrder: pointer.From(true),
			expected: testutils.NewFieldDef("AllOfWithFactoredOutProperties",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("AllOfWithFactoredOutProperties").
					WithOriginalName("AllOfWithFactoredOutProperties").
					WithLocation(&yaml.Node{Line: 11, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(true).
					WithContextStack(
						testutils.NewContextStack().WithRefType("Schemas").WithRefName("AllOfWithFactoredOutProperties").WithIdentifierForNaming(nil).WithComponent().Build(),
					).
					WithComments(&ast.Comment{
						Description:      "An object with allOf and factored out properties.",
						ExtendedComments: make(map[string]*ast.ExtendedComment),
					}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "#/components/schemas/AllOfWithFactoredOutProperties-allOf[41e86b6181ae811d,b76f1247f32cf1a2]",
					}).
					WithFields(
						testutils.NewFieldDef("anOverridingProperty", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 18, Column: 11}).WithValidations(&ast.Validations{}).WithExamples(&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.ScalarNode,
								Style:  yaml.DoubleQuotedStyle,
								Tag:    "!!str",
								Value:  "example",
								Line:   20,
								Column: 20,
							},
						}).Build()).
							WithOriginalName("anOverridingProperty").
							WithOptional(true).
							WithJSONAnnotation("anOverridingProperty").
							WithComments(&ast.Comment{
								Description:      "An overriding property.",
								ExtendedComments: make(map[string]*ast.ExtendedComment),
							}).
							Build(),
						testutils.NewFieldDef("anotherProperty", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 35, Column: 15}).WithExamples(&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.ScalarNode,
								Style:  yaml.DoubleQuotedStyle,
								Tag:    "!!str",
								Value:  "example",
								Line:   37,
								Column: 24,
							},
						}).Build()).
							WithOriginalName("anotherProperty").
							WithOptional(true).
							WithJSONAnnotation("anotherProperty").
							WithComments(&ast.Comment{
								Description:      "Another property.",
								ExtendedComments: make(map[string]*ast.ExtendedComment),
							}).
							Build(),
						testutils.NewFieldDef("anExtraProperty", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 11}).WithExamples(&ast.Example{
							Value: &yaml.Node{
								Kind:   yaml.ScalarNode,
								Style:  yaml.DoubleQuotedStyle,
								Tag:    "!!str",
								Value:  "example",
								Line:   16,
								Column: 20,
							},
						}).Build()).
							WithOriginalName("anExtraProperty").
							WithOptional(true).
							WithJSONAnnotation("anExtraProperty").
							WithComments(&ast.Comment{
								Description:      "An extra property.",
								ExtendedComments: make(map[string]*ast.ExtendedComment),
							}).
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNameOverride_NameOverrideBleeds",
			schemasYAML: `TestSchema:
  allOf:
    - type: object
      properties:
        id:
          type: string
    - x-speakeasy-name-override: ChildOverride`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("ChildOverride",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("ChildOverride").
					WithOriginalName("ChildOverride").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "ChildOverride",
						"x-speakeasy-reference-override": "082ee8fa7c9864e3-allOf[340d40e0d55e80f3,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 13, Column: 15}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNameOverride_NameOverrideIgnored",
			schemasYAML: `TestSchema:
  allOf:
    - type: object
      properties:
        id:
          type: string
    - x-speakeasy-name-override: NoEffect`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[340d40e0d55e80f3,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 13, Column: 15}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNestedNameOverride_NameOverrideBleeds",
			schemasYAML: `TestSchema:
  allOf:
    - type: object
      properties:
        prop:
          type: number
    - allOf:
        - type: object
          properties:
            id:
              type: string
        - x-speakeasy-name-override: deepOverriddenAllOf`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("deepOverriddenAllOf",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("deepOverriddenAllOf").
					WithOriginalName("deepOverriddenAllOf").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "deepOverriddenAllOf",
						"x-speakeasy-reference-override": "04d9ba2d088cd9e3-allOf[340d40e0d55e80f3,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 18, Column: 19}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("prop", testutils.NewTypeDef(ast.DataTypeNumber).WithLocation(&yaml.Node{Line: 13, Column: 15}).Build()).
							WithOriginalName("prop").
							WithOptional(true).
							WithJSONAnnotation("prop").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNestedNameOverride_NameOverrideIgnored",
			schemasYAML: `TestSchema:
  allOf:
    - type: object
      properties:
        prop:
          type: number
    - allOf:
        - type: object
          properties:
            id:
              type: string
        - x-speakeasy-name-override: deeplyLostExtension`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "04d9ba2d088cd9e3-allOf[340d40e0d55e80f3,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 18, Column: 19}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
						testutils.NewFieldDef("prop", testutils.NewTypeDef(ast.DataTypeNumber).WithLocation(&yaml.Node{Line: 13, Column: 15}).Build()).
							WithOriginalName("prop").
							WithOptional(true).
							WithJSONAnnotation("prop").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithMultipleNameOverrides_NameOverrideBleeds",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    name:
      type: string
  x-speakeasy-name-override: BaseOverride
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        age:
          type: integer
    - x-speakeasy-name-override: FinalOverride`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("FinalOverride",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("FinalOverride").
					WithOriginalName("FinalOverride").
					WithLocation(&yaml.Node{Line: 15, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "FinalOverride",
						"x-speakeasy-reference-override": "9b04e51a3d8537ac-allOf[#/components/schemas/BaseSchema,0032b2b30ac5ee82,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithMultipleNameOverrides_NameOverrideIgnored",
			schemasYAML: `BaseSchema:
  type: object
  properties:
    name:
      type: string
  x-speakeasy-name-override: BaseOverride
TestSchema:
  allOf:
    - $ref: "#/components/schemas/BaseSchema"
    - type: object
      properties:
        age:
          type: integer
    - x-speakeasy-name-override: FinalOverride`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 15, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/BaseSchema,0032b2b30ac5ee82,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 20, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNameOverrideAndProperties_NameOverrideBleeds",
			schemasYAML: `TestSchema:
  type: object
  properties:
    existingProp:
      type: string
  allOf:
    - type: object
      properties:
        id:
          type: string
    - x-speakeasy-name-override: MergedOverride`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("MergedOverride",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("MergedOverride").
					WithOriginalName("MergedOverride").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "MergedOverride",
						"x-speakeasy-reference-override": "5df4b1ac636be865-allOf[340d40e0d55e80f3,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("existingProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("existingProp").
							WithOptional(true).
							WithJSONAnnotation("existingProp").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 17, Column: 15}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithNameOverrideAndProperties_NameOverrideIgnored",
			schemasYAML: `TestSchema:
  type: object
  properties:
    existingProp:
      type: string
  allOf:
    - type: object
      properties:
        id:
          type: string
    - x-speakeasy-name-override: UnmergedOverride`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "5df4b1ac636be865-allOf[340d40e0d55e80f3,cbf29ce484222325]",
					}).
					WithFields(
						testutils.NewFieldDef("existingProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 12, Column: 11}).Build()).
							WithOriginalName("existingProp").
							WithOptional(true).
							WithJSONAnnotation("existingProp").
							Build(),
						testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 17, Column: 15}).Build()).
							WithOriginalName("id").
							WithOptional(true).
							WithJSONAnnotation("id").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithParentNameOverrideInlineChildren",
			schemasYAML: `TestSchema:
  x-speakeasy-name-override: Profile
  allOf:
    - type: object
      properties:
        name:
          type: string
    - type: object
      properties:
        age:
          type: integer`,
			schemaName: "TestSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("Profile",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("Profile").
					WithOriginalName("Profile").
					WithLocation(&yaml.Node{Line: 9, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "Profile",
						"x-speakeasy-reference-override": "a4727bbfc5e65cc6-allOf[0032b2b30ac5ee82,97ccf04bb202fda9]",
					}).
					WithFields(
						testutils.NewFieldDef("age", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 18, Column: 15}).Build()).
							WithOriginalName("age").
							WithOptional(true).
							WithJSONAnnotation("age").
							Build(),
						testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 14, Column: 15}).Build()).
							WithOriginalName("name").
							WithOptional(true).
							WithJSONAnnotation("name").
							Build(),
					).
					Build()).
				Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runAllOfTest(t, tt, false)
		})
	}
}

func TestHandleAllOf_DeepMerge(t *testing.T) {
	tests := []allOfTestCase{
		{
			name: "DeepMerge: Simple nested object merge",
			schemasYAML: `
Base:
    type: object
    properties:
      config:
        type: object
        properties:
          timeout:
            type: integer
TestSchema:
    allOf:
      - $ref: "#/components/schemas/Base"
      - type: object
        properties:
          config:
            type: object
            properties:
              retries:
                type: integer
`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 18, Column: 9}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/Base,f33fc51c4e336608]",
					}).
					WithFields(
						testutils.NewFieldDef("config",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("config").
								WithOriginalName("config").
								WithLocation(&yaml.Node{Line: 13, Column: 13}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(ast.ContextStack{}).
								WithFields(
									testutils.NewFieldDef("retries", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 26, Column: 21}).Build()).
										WithOriginalName("retries").
										WithOptional(true).
										WithJSONAnnotation("retries").
										Build(),
									testutils.NewFieldDef("timeout", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 16, Column: 17}).Build()).
										WithOriginalName("timeout").
										WithOptional(true).
										WithJSONAnnotation("timeout").
										Build(),
								).
								Build()).
							WithOriginalName("config").
							WithOptional(true).
							WithJSONAnnotation("config").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "DeepMerge: Nested object with required fields",
			schemasYAML: `
Base:
    type: object
    properties:
      metadata:
        type: object
        properties:
          name:
            type: string
TestSchema:
    allOf:
      - $ref: "#/components/schemas/Base"
      - type: object
        required:
          - metadata
        properties:
          metadata:
            type: object
            required:
              - name
`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 18, Column: 9}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/Base,6137f97fc588561f]",
					}).
					WithFields(
						testutils.NewFieldDef("metadata",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("metadata").
								WithOriginalName("metadata").
								WithLocation(&yaml.Node{Line: 13, Column: 13}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(ast.ContextStack{}).
								WithFields(
									testutils.NewFieldDef("name", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 16, Column: 17}).Build()).
										WithOriginalName("name").
										WithOptional(false).
										WithJSONAnnotation("name").
										Build(),
								).
								Build()).
							WithOriginalName("metadata").
							WithOptional(false).
							WithJSONAnnotation("metadata").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "DeepMerge: Multiple levels of nesting",
			schemasYAML: `
Base:
    type: object
    properties:
      level1:
        type: object
        properties:
          level2:
            type: object
            properties:
              value:
                type: string
TestSchema:
    allOf:
      - $ref: "#/components/schemas/Base"
      - type: object
        properties:
          level1:
            type: object
            properties:
              level2:
                type: object
                properties:
                  extra:
                    type: integer
`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 21, Column: 9}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/Base,3aaa97dbda1fab0a]",
					}).
					WithFields(
						testutils.NewFieldDef("level1",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("level1").
								WithOriginalName("level1").
								WithLocation(&yaml.Node{Line: 13, Column: 13}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(ast.ContextStack{}).
								WithFields(
									testutils.NewFieldDef("level2",
										testutils.NewTypeDef(ast.DataTypeClass).
											WithName("level2").
											WithOriginalName("level2").
											WithLocation(&yaml.Node{Line: 16, Column: 17}).
											WithScope(ast.ScopeShared).
											WithRegistered().
											WithOriginalNameFrozen().
											WithContextStack(
												ast.ContextStack{
													{Type: ast.ContextTypeProperty, Identifier: "level1"},
												},
											).
											WithFields(
												testutils.NewFieldDef("extra", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 32, Column: 25}).Build()).
													WithOriginalName("extra").
													WithOptional(true).
													WithJSONAnnotation("extra").
													Build(),
												testutils.NewFieldDef("value", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 19, Column: 21}).Build()).
													WithOriginalName("value").
													WithOptional(true).
													WithJSONAnnotation("value").
													Build(),
											).
											Build()).
										WithOriginalName("level2").
										WithOptional(true).
										WithJSONAnnotation("level2").
										Build(),
								).
								Build()).
							WithOriginalName("level1").
							WithOptional(true).
							WithJSONAnnotation("level1").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "DeepMerge: Nested object with examples",
			schemasYAML: `Base:
    type: object
    properties:
      settings:
        type: object
        properties:
          enabled:
            type: boolean
TestSchema:
    allOf:
      - $ref: "#/components/schemas/Base"
      - type: object
        properties:
          settings:
            example: {"enabled": true, "mode": "auto"}
            type: object
            properties:
              mode:
                type: string`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 17, Column: 9}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/Base,988dc253f9617883]",
					}).
					WithFields(
						testutils.NewFieldDef("settings",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("settings").
								WithOriginalName("settings").
								WithLocation(&yaml.Node{Line: 12, Column: 13}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(ast.ContextStack{}).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind:   yaml.MappingNode,
										Style:  yaml.FlowStyle,
										Tag:    "!!map",
										Line:   22,
										Column: 26,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "enabled", Line: 22, Column: 27},
											{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true", Line: 22, Column: 38},
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "mode", Line: 22, Column: 44},
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "auto", Line: 22, Column: 52},
										},
									},
								}).
								WithFields(
									testutils.NewFieldDef("enabled", testutils.NewTypeDef(ast.DataTypeBoolean).Build()).
										WithOriginalName("enabled").
										WithOptional(true).
										WithJSONAnnotation("enabled").
										Build(),
									testutils.NewFieldDef("mode", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 26, Column: 21}).Build()).
										WithOriginalName("mode").
										WithOptional(true).
										WithJSONAnnotation("mode").
										Build(),
								).
								Build()).
							WithOriginalName("settings").
							WithOptional(true).
							WithJSONAnnotation("settings").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "DeepMerge: Multiple allOf entries with nested objects",
			schemasYAML: `Base1:
    type: object
    properties:
      data:
        type: object
        properties:
          field1:
            type: string
Base2:
    type: object
    properties:
      data:
        type: object
        properties:
          field2:
            type: integer
TestSchema:
    allOf:
      - $ref: "#/components/schemas/Base1"
      - $ref: "#/components/schemas/Base2"
      - type: object
        properties:
          data:
            type: object
            properties:
              field3:
                type: boolean`,
			schemaName: "TestSchema",
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 25, Column: 9}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/Base1,#/components/schemas/Base2,b48781f37becbce1]",
					}).
					WithFields(
						testutils.NewFieldDef("data",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("data").
								WithOriginalName("data").
								WithLocation(&yaml.Node{Line: 12, Column: 13}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(ast.ContextStack{}).
								WithFields(
									testutils.NewFieldDef("field1", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 15, Column: 17}).Build()).
										WithOriginalName("field1").
										WithOptional(true).
										WithJSONAnnotation("field1").
										Build(),
									testutils.NewFieldDef("field2", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 23, Column: 17}).Build()).
										WithOriginalName("field2").
										WithOptional(true).
										WithJSONAnnotation("field2").
										Build(),
									testutils.NewFieldDef("field3", testutils.NewTypeDef(ast.DataTypeBoolean).Build()).
										WithOriginalName("field3").
										WithOptional(true).
										WithJSONAnnotation("field3").
										Build(),
								).
								Build()).
							WithOriginalName("data").
							WithOptional(true).
							WithJSONAnnotation("data").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "DeepMerge: required nested objects with examples",
			schemasYAML: `
WidgetSpec:
  type: object
  properties:
    folder:
      type: object
      properties:
        id:
          type: string
    storage:
      type: object
      properties:
        id:
          type: string
    author:
      type: object
      properties:
        id:
          type: string
Widget:
  type: object
  properties:
    api_version:
      type: string
      enum:
      - example/v1
      readOnly: true
    kind:
      type: string
      readOnly: true
      enum:
      - Widget
    id:
      type: string
      maxLength: 64
      readOnly: true
      example: widget-001
    spec:
      "$ref": "#/components/schemas/WidgetSpec"
TestSchema:
  allOf:
    - {"$ref": "#/components/schemas/Widget"}
    - type: "object"
      required:
        - "spec"
      properties:
        "spec":
          type: "object"
          required:
            - "folder"
            - "storage"
            - "author"
    - type: "object"
      properties:
        "spec":
          type: "object"
          properties:
            "folder":
              example: {"id": "folder-001"}
            "storage":
              example: {"id": "storage-001"}
            "author":
              example: {"id": "author-001"}
`,
			schemaName: "TestSchema",
			expected: func() *ast.FieldDef {
				widgetSpecType := testutils.NewTypeDef(ast.DataTypeClass).
					WithName("spec").
					WithOriginalName("spec").
					WithLocation(&yaml.Node{Line: 10, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithIsComponent(false).
					WithContextStack(ast.ContextStack{}).
					WithFields(
						testutils.NewFieldDef("author",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("author").
								WithOriginalName("author").
								WithLocation(&yaml.Node{Line: 23, Column: 11}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(
									ast.ContextStack{
										{Type: ast.ContextTypeProperty, Identifier: "spec"},
									},
								).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind:   yaml.MappingNode,
										Style:  yaml.FlowStyle,
										Tag:    "!!map",
										Line:   70,
										Column: 28,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "id", Line: 70, Column: 29},
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "author-001", Line: 70, Column: 35},
										},
									},
								}).
								WithFields(
									testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 26, Column: 15}).Build()).
										WithOriginalName("id").
										WithOptional(true).
										WithJSONAnnotation("id").
										Build(),
								).
								Build()).
							WithOriginalName("author").
							WithOptional(false).
							WithJSONAnnotation("author").
							Build(),
						testutils.NewFieldDef("folder",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("folder").
								WithOriginalName("folder").
								WithLocation(&yaml.Node{Line: 13, Column: 11}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(
									ast.ContextStack{
										{Type: ast.ContextTypeProperty, Identifier: "spec"},
									},
								).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind:   yaml.MappingNode,
										Style:  yaml.FlowStyle,
										Tag:    "!!map",
										Line:   66,
										Column: 28,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "id", Line: 66, Column: 29},
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "folder-001", Line: 66, Column: 35},
										},
									},
								}).
								WithFields(
									testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 16, Column: 15}).Build()).
										WithOriginalName("id").
										WithOptional(true).
										WithJSONAnnotation("id").
										Build(),
								).
								Build()).
							WithOriginalName("folder").
							WithOptional(false).
							WithJSONAnnotation("folder").
							Build(),
						testutils.NewFieldDef("storage",
							testutils.NewTypeDef(ast.DataTypeClass).
								WithName("storage").
								WithOriginalName("storage").
								WithLocation(&yaml.Node{Line: 18, Column: 11}).
								WithScope(ast.ScopeShared).
								WithRegistered().
								WithOriginalNameFrozen().
								WithContextStack(
									ast.ContextStack{
										{Type: ast.ContextTypeProperty, Identifier: "spec"},
									},
								).
								WithExamples(&ast.Example{
									Value: &yaml.Node{
										Kind:   yaml.MappingNode,
										Style:  yaml.FlowStyle,
										Tag:    "!!map",
										Line:   68,
										Column: 28,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "id", Line: 68, Column: 29},
											{Kind: yaml.ScalarNode, Style: yaml.DoubleQuotedStyle, Tag: "!!str", Value: "storage-001", Line: 68, Column: 35},
										},
									},
								}).
								WithFields(
									testutils.NewFieldDef("id", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 21, Column: 15}).Build()).
										WithOriginalName("id").
										WithOptional(true).
										WithJSONAnnotation("id").
										Build(),
								).
								Build()).
							WithOriginalName("storage").
							WithOptional(false).
							WithJSONAnnotation("storage").
							Build(),
					).
					Build()
				apiVersionEnumType := testutils.NewEnumTypeDef(ast.DataTypeString, []string{"example/v1"}, "enum").
					WithName("api_version").
					WithOriginalName("api_version").
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithConstProperty("example/v1").WithIdentifierForNaming(pointer.From("Example/v1")).Build(),
					).
					Build()
				apiVersionEnumType.Enum.Type.Location = &ast.OpenAPILocation{Node: &yaml.Node{Line: 31, Column: 11}}
				kindEnumType := testutils.NewEnumTypeDef(ast.DataTypeString, []string{"Widget"}, "enum").
					WithName("kind").
					WithOriginalName("kind").
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(
						testutils.NewContextStack().WithConstProperty("Widget").WithIdentifierForNaming(pointer.From("Widget")).Build(),
					).
					Build()
				kindEnumType.Enum.Type.Location = &ast.OpenAPILocation{Node: &yaml.Node{Line: 36, Column: 11}}
				idType := testutils.NewTypeDef(ast.DataTypeString).
					WithLocation(&yaml.Node{Line: 41, Column: 11}).
					WithValidations(&ast.Validations{
						MaxLength: func() *int64 { v := int64(64); return &v }(),
					}).
					WithExamples(&ast.Example{
						Value: &yaml.Node{
							Kind:   yaml.ScalarNode,
							Tag:    "!!str",
							Value:  "widget-001",
							Line:   44,
							Column: 20,
						},
					}).
					Build()

				return testutils.NewFieldDef("object",
					testutils.NewTypeDef(ast.DataTypeClass).
						WithLocation(&yaml.Node{Line: 48, Column: 7}).
						WithScope(ast.ScopeShared).
						WithRegistered().
						WithOriginalNameFrozen().
						WithContextStack(ast.ContextStack{}).
						WithExtensions(map[string]interface{}{
							"x-speakeasy-reference-override": "allOf[#/components/schemas/Widget,461f91e91df1e702,aae7ea3eae1636c6]",
						}).
						WithFields(
							testutils.NewFieldDef("api_version", apiVersionEnumType).
								WithOriginalName("api_version").
								WithOptional(true).
								WithJSONAnnotation("api_version").
								Build(),
							testutils.NewFieldDef("id", idType).
								WithOriginalName("id").
								WithOptional(true).
								WithJSONAnnotation("id").
								Build(),
							testutils.NewFieldDef("kind", kindEnumType).
								WithOriginalName("kind").
								WithOptional(true).
								WithJSONAnnotation("kind").
								Build(),
							testutils.NewFieldDef("spec", widgetSpecType).
								WithOriginalName("spec").
								WithOptional(false).
								WithJSONAnnotation("spec").
								Build(),
						).
						Build()).
					Build()
			}(),
		},
		{
			name: "AllOfWithRefToRenamedComponent_NameOverrideBleeds",
			schemasYAML: `RenamedSchema:
  x-speakeasy-name-override: "renamed"
  type: object
  properties:
    strProp:
      type: string
CustomSchema:
  allOf:
    - $ref: '#/components/schemas/RenamedSchema'
    - type: object
      properties:
        intProp:
          type: integer`,
			schemaName: "CustomSchema",
			fixes:      testutils.WithNameOverrideFeb2026(false),
			expected: testutils.NewFieldDef("renamed",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithName("renamed").
					WithOriginalName("renamed").
					WithLocation(&yaml.Node{Line: 15, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-name-override":      "renamed",
						"x-speakeasy-reference-override": "06e406650f965ab6-allOf[#/components/schemas/RenamedSchema,bf4ef3bc1b80577d]",
					}).
					WithFields(
						testutils.NewFieldDef("intProp", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 20, Column: 15}).WithValidations(&ast.Validations{}).WithExtensions(map[string]interface{}{}).Build()).
							WithOriginalName("intProp").
							WithOptional(true).
							WithJSONAnnotation("intProp").
							Build(),
						testutils.NewFieldDef("strProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 13, Column: 11}).WithValidations(&ast.Validations{}).WithExtensions(map[string]interface{}{}).Build()).
							WithOriginalName("strProp").
							WithOptional(true).
							WithJSONAnnotation("strProp").
							Build(),
					).
					Build()).
				Build(),
		},
		{
			name: "AllOfWithRefToRenamedComponent_NameOverrideIgnored",
			schemasYAML: `RenamedSchema:
  x-speakeasy-name-override: "renamed"
  type: object
  properties:
    strProp:
      type: string
CustomSchema:
  allOf:
    - $ref: '#/components/schemas/RenamedSchema'
    - type: object
      properties:
        intProp:
          type: integer`,
			schemaName: "CustomSchema",
			fixes:      testutils.WithNameOverrideFeb2026(true),
			expected: testutils.NewFieldDef("object",
				testutils.NewTypeDef(ast.DataTypeClass).
					WithLocation(&yaml.Node{Line: 15, Column: 7}).
					WithScope(ast.ScopeShared).
					WithRegistered().
					WithOriginalNameFrozen().
					WithContextStack(ast.ContextStack{}).
					WithExtensions(map[string]interface{}{
						"x-speakeasy-reference-override": "allOf[#/components/schemas/RenamedSchema,bf4ef3bc1b80577d]",
					}).
					WithFields(
						testutils.NewFieldDef("intProp", testutils.NewTypeDef(ast.DataTypeInteger).WithLocation(&yaml.Node{Line: 20, Column: 15}).WithValidations(&ast.Validations{}).WithExtensions(map[string]interface{}{}).Build()).
							WithOriginalName("intProp").
							WithOptional(true).
							WithJSONAnnotation("intProp").
							Build(),
						testutils.NewFieldDef("strProp", testutils.NewTypeDef(ast.DataTypeString).WithLocation(&yaml.Node{Line: 13, Column: 11}).WithValidations(&ast.Validations{}).WithExtensions(map[string]interface{}{}).Build()).
							WithOriginalName("strProp").
							WithOptional(true).
							WithJSONAnnotation("strProp").
							Build(),
					).
					Build()).
				Build(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runAllOfTest(t, tt, true)
		})
	}
}

type allOfTestCase struct {
	name                 string
	schemasYAML          string
	schemaName           string
	fixes                *config.Fixes                        // nil means use default
	mockFeatures         func() *testutils.MockFeaturesConfig // nil means use default (all features supported)
	isRequest            *bool                                // nil means use default (false), otherwise use specified value
	expected             *ast.FieldDef
	maintainOpenAPIOrder *bool
}

func runAllOfTest(t *testing.T, tt allOfTestCase, deepMerge bool) {
	t.Helper()

	// Set up test environment using common testutils
	fullYAML := testutils.CreateOpenAPIDoc(tt.schemasYAML)

	var mockFeatures *testutils.MockFeaturesConfig
	if tt.mockFeatures != nil {
		mockFeatures = tt.mockFeatures()
	} else {
		mockFeatures = testutils.NewMockFeaturesConfig().WithSupportAllFeatures()
	}

	opts := testutils.TestEnvironmentOptions{
		OpenAPIYAML:          fullYAML,
		Fixes:                tt.fixes,
		MockFeatures:         mockFeatures,
		MaintainOpenAPIOrder: tt.maintainOpenAPIOrder,
	}

	if deepMerge {
		opts.Schemas = &config.Schemas{
			AllOfMergeStrategy: config.AllOfMergeStrategyDeepMerge,
		}
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

	// Call the top-level HandleSchema method instead of handleAllOf directly
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
}
