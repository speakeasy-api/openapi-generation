package schemas

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	internalOpenAPI "github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLibopenAPINestedRefBug documents a bug in libopenapi where multiple schemas
// that reference the same shared schema will incorrectly return the first schema's
// reference name.
//
// For example, given:
//
//	Schema1:
//	  $ref: "#/components/schemas/SchemaShared"
//	Schema2:
//	  $ref: "#/components/schemas/SchemaShared"
//	SchemaShared:
//	  type: object
//	  properties:
//	    name:
//	      type: string
//
// When libopenapi resolves Schema2, internally it may return Schema1 as the reference
// because libopenapi indexes nested schemas by the first schema's name that contains them.
//
// This test documents the CURRENT BUGGY BEHAVIOR so that when we refactor, we can:
// 1. Maintain backward compatibility with a feature flag
// 2. Eventually fix the behavior
func TestLibopenAPINestedRefBug(t *testing.T) {
	// Create an OpenAPI spec with two schemas referencing the same shared schema
	// This is used in request/response contexts where the ref resolution matters
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /endpoint1:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Schema1"
  /endpoint2:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Schema2"
  /endpoint3:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/SchemaShared"
components:
  schemas:
    Schema1:
      $ref: "#/components/schemas/SchemaShared"
    Schema2:
      $ref: "#/components/schemas/SchemaShared"
    SchemaShared:
      type: object
      properties:
        name:
          type: string
        value:
          type: integer
`

	tests := []struct {
		name        string
		fixes       *config.Fixes
		wantResult1 string
		wantResult2 string
		wantResult3 string // Direct reference to SchemaShared
	}{
		{
			name: "Buggy behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: false,
			},
			wantResult1: "Schema1",
			wantResult2: "Schema1",
			// In buggy behavior, ALL references to SchemaShared (including direct ones)
			// resolve to the first referencer's name because libopenapi's indexing
			// "claims" the shared schema under the first referencer's name.
			wantResult3: "Schema1",
		},
		{
			name: "Buggy behavior - fixed",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: true,
			},
			wantResult1: "Schema1",
			wantResult2: "Schema2",
			wantResult3: "SchemaShared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  openAPIYAML,
				MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
				Fixes:        tt.fixes,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Debug: log the registry contents
			registry := internalOpenAPI.GlobalNestedRefRegistry()
			t.Logf("Registry contents after population:")
			for k, v := range registry.All() {
				t.Logf("  %s -> %s", k, v)
			}
			t.Logf("Registry length: %d", registry.Len())

			// Create the Schemas instance
			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			// Get the schemas referencing Schema1, Schema2, and SchemaShared directly from the document
			schema1 := common.DocInfo.Doc.GetPaths().GetOrZero("/endpoint1").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, schema1)

			schema2 := common.DocInfo.Doc.GetPaths().GetOrZero("/endpoint2").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, schema2)

			// Direct reference to SchemaShared (not through an alias)
			schema3 := common.DocInfo.Doc.GetPaths().GetOrZero("/endpoint3").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, schema3)

			// Document the current (buggy) behavior:
			// Both schemas reference SchemaShared, but due to libopenapi's internal indexing,
			// the Ref() method may return the same reference for both.

			ref1 := schema1.GetRef()
			ref2 := schema2.GetRef()
			ref3 := schema3.GetRef()

			t.Logf("Schema1 ref: %s", ref1)
			t.Logf("Schema2 ref: %s", ref2)
			t.Logf("Schema3 ref (direct): %s", ref3)

			// This test documents the CURRENT behavior, which may be buggy.
			// When we resolve Schema1's ref, we expect it to be Schema1 (which references SchemaShared)
			// When we resolve Schema2's ref, we expect it to be Schema2 (which references SchemaShared)
			//
			// However, due to libopenapi's indexing behavior, both might resolve to Schema1
			// because libopenapi internally stores nested schemas under the first parent's name.

			// Process Schema1 through HandleSchema
			params1 := CreateTestParams(schema1, common.DocInfo)
			result1, err := schemas.HandleSchema(context.Background(), params1)
			require.NoError(t, err)
			require.NotNil(t, result1)

			// Process Schema2 through HandleSchema
			params2 := CreateTestParams(schema2, common.DocInfo)
			result2, err := schemas.HandleSchema(context.Background(), params2)
			require.NoError(t, err)
			require.NotNil(t, result2)

			// Process Schema3 (direct reference to SchemaShared) through HandleSchema
			params3 := CreateTestParams(schema3, common.DocInfo)
			result3, err := schemas.HandleSchema(context.Background(), params3)
			require.NoError(t, err)
			require.NotNil(t, result3)

			t.Logf("Result1 Name: %s", result1.Name)
			t.Logf("Result1 Type Name: %s", result1.Type.Name)
			t.Logf("Result2 Name: %s", result2.Name)
			t.Logf("Result2 Type Name: %s", result2.Type.Name)
			t.Logf("Result3 Name: %s", result3.Name)
			t.Logf("Result3 Type Name: %s", result3.Type.Name)

			// Document the CURRENT (potentially buggy) behavior:
			// Both results currently resolve to the same type name because libopenapi
			// returns the same reference for both Schema1 and Schema2 when they both
			// point to SchemaShared.
			//
			// This assertion documents the current behavior - if this test fails after
			// a libopenapi update or fix, it means the behavior has changed.

			// The expected CORRECT behavior would be:
			// - result1.Name == "Schema1" or "SchemaShared" (depending on how we want to resolve)
			// - result2.Name == "Schema2" or "SchemaShared" (depending on how we want to resolve)
			//
			// The CURRENT (buggy) behavior is that both resolve to the same name based on
			// whichever schema libopenapi encountered first in its indexing.

			// The type name should be Schema1 since showcasing the actual bug where ideally they should resolve to Schema1 and Schema2 respectively
			assert.Equal(t, tt.wantResult1, result1.Type.Name,
				"Schema1 should resolve to "+tt.wantResult1)
			assert.Equal(t, tt.wantResult2, result2.Type.Name,
				"Schema2 should resolve to "+tt.wantResult2)
			// In buggy behavior: direct refs to SchemaShared also resolve to Schema1
			// In fixed behavior: direct refs to SchemaShared stay as SchemaShared
			assert.Equal(t, tt.wantResult3, result3.Type.Name,
				"Direct reference to SchemaShared should resolve to "+tt.wantResult3)
		})
	}
}

func TestAliasedReferencePositions(t *testing.T) {
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API - Expanded Reference Types
  version: 1.0.0
paths:
  /widgets/{id}:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/GetWidgetResponse"
  /widgets:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ListWidgetsResponse"
  /container:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ContainerWithDirectRef"
  /map:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/MapOfWidgets"
  /oneof:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/OneOfContainer"
  /nested:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/DeepNestedContainer"
  /direct:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Widget"
  /second-alias:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/SecondAliasForWidget"
components:
  schemas:

    GetWidgetResponse:
      $ref: "#/components/schemas/Widget"

    ListWidgetsResponse:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: "#/components/schemas/Widget"

    ContainerWithDirectRef:
      type: object
      properties:
        widget:
          $ref: "#/components/schemas/Widget"

    MapOfWidgets:
      type: object
      additionalProperties:
        $ref: "#/components/schemas/Widget"

    OneOfContainer:
      oneOf:
        - $ref: "#/components/schemas/Widget"
        - type: object
          properties:
            fallback:
              type: string

    DeepNestedContainer:
      type: object
      properties:
        level1:
          type: object
          properties:
            level2:
              type: array
              items:
                type: object
                properties:
                  widget:
                    $ref: "#/components/schemas/Widget"
    
    # The actual schema (comes AFTER all refs in doc)
    Widget:
      type: object
      properties:
        id:
          type: string
        name:
          type: string

    SecondAliasForWidget:
      $ref: "#/components/schemas/Widget"
`

	tests := []struct {
		name                       string
		fixes                      *config.Fixes
		wantAliasType              string // GetWidgetResponse -> ?
		wantArrayItemsType         string // ListWidgetsResponse.data.items -> ?
		wantDirectPropertyType     string // ContainerWithDirectRef.widget -> ?
		wantAdditionalPropsType    string // MapOfWidgets additionalProperties -> ?
		wantOneOfComponentType     string // OneOfContainer oneOf[0] -> ?
		wantDeepNestedPropertyType string // DeepNestedContainer...widget -> ?
		wantDirectRefType          string // /direct response -> ?
		wantSecondAliasType        string // SecondAliasForWidget -> ?
	}{
		{
			name: "Buggy behavior - libopenapi nested ref bug",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: false,
			},
			// First alias keeps its name
			wantAliasType: "GetWidgetResponse",
			// Array items are NOT replaced (libopenapi bug doesn't affect them)
			wantArrayItemsType: "Widget",
			// Direct property refs ARE replaced with first alias
			wantDirectPropertyType: "GetWidgetResponse",
			// AdditionalProperties are NOT replaced (libopenapi bug doesn't affect them)
			wantAdditionalPropsType: "Widget",
			// OneOf components ARE replaced with first alias
			wantOneOfComponentType: "GetWidgetResponse",
			// Deep nested property refs ARE replaced with first alias
			wantDeepNestedPropertyType: "GetWidgetResponse",
			// Direct path refs to Widget ARE replaced with first alias
			wantDirectRefType: "GetWidgetResponse",
			// Second alias (after Widget) gets first alias's name
			wantSecondAliasType: "GetWidgetResponse",
		},
		{
			name: "Fixed behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: true,
			},
			// Fixed: everything uses its proper name
			wantAliasType:              "GetWidgetResponse",
			wantArrayItemsType:         "Widget",
			wantDirectPropertyType:     "Widget",
			wantAdditionalPropsType:    "Widget",
			wantOneOfComponentType:     "Widget",
			wantDeepNestedPropertyType: "Widget",
			wantDirectRefType:          "Widget",
			wantSecondAliasType:        "SecondAliasForWidget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  openAPIYAML,
				MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
				Fixes:        tt.fixes,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			t.Log("\n=== Testing Multiple Reference Types ===")

			// 1. First Alias (GetWidgetResponse)
			t.Log("\n--- 1. First Alias (GetWidgetResponse) ---")
			getWidgetSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/widgets/{id}").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, getWidgetSchema)
			paramsGet := CreateTestParams(getWidgetSchema, common.DocInfo)
			resultGet, err := schemas.HandleSchema(context.Background(), paramsGet)
			require.NoError(t, err)
			t.Logf("GetWidgetResponse Type Name: %s", resultGet.Type.Name)
			assert.Equal(t, tt.wantAliasType, resultGet.Type.Name, "First alias type")

			// 2. Array items
			t.Log("\n--- 2. Array Items (ListWidgetsResponse.data.items) ---")
			listWidgetsSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/widgets").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, listWidgetsSchema)
			paramsListWidgets := CreateTestParams(listWidgetsSchema, common.DocInfo)
			resultListWidgets, err := schemas.HandleSchema(context.Background(), paramsListWidgets)
			require.NoError(t, err)
			var dataPropType *ast.TypeDef
			for _, field := range resultListWidgets.Type.Fields {
				if field.Name == "data" || field.OriginalName == "data" {
					dataPropType = field.Type
					break
				}
			}
			require.NotNil(t, dataPropType)
			require.NotNil(t, dataPropType.ItemType)
			t.Logf("Array items Type Name: %s", dataPropType.ItemType.Name)
			assert.Equal(t, tt.wantArrayItemsType, dataPropType.ItemType.Name, "Array items type")

			// 3. Direct property reference
			t.Log("\n--- 3. Direct Property (ContainerWithDirectRef.widget) ---")
			containerSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/container").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, containerSchema)
			paramsContainer := CreateTestParams(containerSchema, common.DocInfo)
			resultContainer, err := schemas.HandleSchema(context.Background(), paramsContainer)
			require.NoError(t, err)
			var widgetPropType *ast.TypeDef
			for _, field := range resultContainer.Type.Fields {
				if field.Name == "widget" || field.OriginalName == "widget" {
					widgetPropType = field.Type
					break
				}
			}
			require.NotNil(t, widgetPropType)
			t.Logf("Direct property Type Name: %s", widgetPropType.Name)
			assert.Equal(t, tt.wantDirectPropertyType, widgetPropType.Name, "Direct property type")

			// 4. AdditionalProperties
			t.Log("\n--- 4. AdditionalProperties (MapOfWidgets) ---")
			mapSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/map").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, mapSchema)
			paramsMap := CreateTestParams(mapSchema, common.DocInfo)
			resultMap, err := schemas.HandleSchema(context.Background(), paramsMap)
			require.NoError(t, err)
			// Map's value type is stored in ItemType (for additionalProperties)
			if resultMap.Type.ItemType != nil {
				t.Logf("AdditionalProperties Type Name: %s", resultMap.Type.ItemType.Name)
				assert.Equal(t, tt.wantAdditionalPropsType, resultMap.Type.ItemType.Name, "AdditionalProperties type")
			} else {
				t.Log("AdditionalProperties ItemType is nil - check implementation")
			}

			// 5. OneOf component
			t.Log("\n--- 5. OneOf Component (OneOfContainer) ---")
			oneOfSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/oneof").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, oneOfSchema)
			paramsOneOf := CreateTestParams(oneOfSchema, common.DocInfo)
			resultOneOf, err := schemas.HandleSchema(context.Background(), paramsOneOf)
			require.NoError(t, err)
			t.Logf("OneOf result type: %s", resultOneOf.Type.Name)
			if len(resultOneOf.Type.AssociatedTypes) > 0 {
				t.Logf("OneOf has %d union types", len(resultOneOf.Type.AssociatedTypes))
				for i, ut := range resultOneOf.Type.AssociatedTypes {
					t.Logf("  Union[%d]: %s", i, ut.Name)
				}
				// First union type should be Widget (or its alias)
				if len(resultOneOf.Type.AssociatedTypes) > 0 {
					assert.Equal(t, tt.wantOneOfComponentType, resultOneOf.Type.AssociatedTypes[0].Name, "OneOf component type")
				}
			} else {
				t.Log("OneOf UnionTypes is empty - check implementation")
			}

			// 6. Deeply nested reference
			t.Log("\n--- 6. Deeply Nested Property ---")
			deepNestedSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/nested").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, deepNestedSchema)
			paramsDeepNested := CreateTestParams(deepNestedSchema, common.DocInfo)
			resultDeepNested, err := schemas.HandleSchema(context.Background(), paramsDeepNested)
			require.NoError(t, err)
			// Navigate: level1 -> level2 (array) -> items -> widget
			var deepNestedWidgetType string
			for _, field := range resultDeepNested.Type.Fields {
				if field.Name == "level1" || field.OriginalName == "level1" {
					for _, level1Field := range field.Type.Fields {
						if level1Field.Name == "level2" || level1Field.OriginalName == "level2" {
							if level1Field.Type.ItemType != nil {
								for _, itemField := range level1Field.Type.ItemType.Fields {
									if itemField.Name == "widget" || itemField.OriginalName == "widget" {
										deepNestedWidgetType = itemField.Type.Name
										t.Logf("Deep nested 'widget' Type Name: %s", deepNestedWidgetType)
									}
								}
							}
						}
					}
				}
			}
			if deepNestedWidgetType != "" {
				assert.Equal(t, tt.wantDeepNestedPropertyType, deepNestedWidgetType, "Deep nested property type")
			} else {
				t.Log("Could not navigate to deep nested widget property")
			}

			// 7. Direct reference from path
			t.Log("\n--- 7. Direct Path Reference to Widget ---")
			directSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/direct").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, directSchema)
			paramsDirect := CreateTestParams(directSchema, common.DocInfo)
			resultDirect, err := schemas.HandleSchema(context.Background(), paramsDirect)
			require.NoError(t, err)
			t.Logf("Direct path ref Type Name: %s", resultDirect.Type.Name)
			assert.Equal(t, tt.wantDirectRefType, resultDirect.Type.Name, "Direct path reference type")

			// 8. Second alias (comes AFTER Widget in document)
			t.Log("\n--- 8. Second Alias (after Widget in doc) ---")
			secondAliasSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/second-alias").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, secondAliasSchema)
			paramsSecondAlias := CreateTestParams(secondAliasSchema, common.DocInfo)
			resultSecondAlias, err := schemas.HandleSchema(context.Background(), paramsSecondAlias)
			require.NoError(t, err)
			t.Logf("Second alias Type Name: %s", resultSecondAlias.Type.Name)
			assert.Equal(t, tt.wantSecondAliasType, resultSecondAlias.Type.Name, "Second alias type")

			t.Log("\n=== Summary ===")
			t.Logf("1. First alias (GetWidgetResponse): %s", resultGet.Type.Name)
			t.Logf("2. Array items: %s", dataPropType.ItemType.Name)
			t.Logf("3. Direct property: %s", widgetPropType.Name)
			if resultMap.Type.ItemType != nil {
				t.Logf("4. AdditionalProperties: %s", resultMap.Type.ItemType.Name)
			}
			if len(resultOneOf.Type.AssociatedTypes) > 0 {
				t.Logf("5. OneOf[0]: %s", resultOneOf.Type.AssociatedTypes[0].Name)
			}
			if deepNestedWidgetType != "" {
				t.Logf("6. Deep nested: %s", deepNestedWidgetType)
			}
			t.Logf("7. Direct path ref: %s", resultDirect.Type.Name)
			t.Logf("8. Second alias: %s", resultSecondAlias.Type.Name)
		})
	}
}

func TestAliasedResponseReferences(t *testing.T) {
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /widgets/{id}:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/GetWidgetResponse"
  /widgets:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ListWidgetsResponse"

  /get-profile:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/GetWidgetProfileResponse"
  /update-profile:
    put:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/UpdateWidgetProfileResponse"


  /widget-settings-aliases:
    get:
      responses:
        '200':
          description: A list of settings aliases
          content:
            application/json:
              schema:
                type: object
                title: ListWidgetSettingsAliasesResponse
                properties:
                  data:
                    type: array
                    items:
                      $ref: "#/components/schemas/WidgetSettingsAlias"

  /settings:
    get:
      responses:
        '200':
          description: A list of settings
          content:
            application/json:
              schema:
                type: object
                title: ListSettingsResponse
                properties:
                  data:
                    type: array
                    items:
                      $ref: "#/components/schemas/WidgetSettings"

  /settings/{id}:
    get:
      responses:
        '200':
          description: Settings retrieved by ID
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/WidgetSettings"

  /create-settings:
    post:
      responses:
        '200':
          description: Created settings
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/WidgetSettings"
components:
  schemas:
    GetWidgetResponse:
      $ref: "#/components/schemas/Widget"
    ListWidgetsResponse:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: "#/components/schemas/Widget"
    Widget:
      type: object
      properties:
        id:
          type: string
        name:
          type: string

    GetWidgetProfileResponse:
      $ref: "#/components/schemas/WidgetProfile"
    UpdateWidgetProfileResponse:
      $ref: "#/components/schemas/WidgetProfile"
    WidgetProfile:
      type: object
      properties:
        id:
          type: string
        name:
          type: string

    WidgetSettingsAlias:
      $ref: "#/components/schemas/WidgetSettings"
    WidgetSettings:
      type: object
      properties:
        version:
          type: string
        data:
          type: string
`

	tests := []struct {
		name                       string
		fixes                      *config.Fixes
		wantGetWidgetResponse      string
		wantListWidgetItemsType    string
		wantGetProfileResponse     string // GetWidgetProfileResponse -> ?
		wantUpdateProfileResponse  string // UpdateWidgetProfileResponse -> ?
		wantSettingsAliasItemsType string // /widget-settings-aliases array items (uses WidgetSettingsAlias) -> ?
		wantListSettingsItemType   string // /settings array items (uses WidgetSettings) -> ?
		wantRetrieveSettings       string // /settings/{id} (direct ref to WidgetSettings) -> ?
		wantCreateSettings         string // /create-settings (direct ref to WidgetSettings) -> ?
	}{
		{
			name: "Buggy behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: false,
			},
			wantGetWidgetResponse: "GetWidgetResponse",

			wantListWidgetItemsType: "Widget",

			wantGetProfileResponse:    "GetWidgetProfileResponse",
			wantUpdateProfileResponse: "GetWidgetProfileResponse", // Bug: second alias also becomes first alias name
			// Array items using WidgetSettingsAlias stay as WidgetSettingsAlias
			wantSettingsAliasItemsType: "WidgetSettingsAlias",
			// Array items using WidgetSettings stay as WidgetSettings (array items not affected)
			wantListSettingsItemType: "WidgetSettings",
			wantRetrieveSettings:     "WidgetSettings",
			wantCreateSettings:       "WidgetSettings",
		},
		{
			name: "Fixed behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: true,
			},
			wantGetWidgetResponse:   "GetWidgetResponse",
			wantListWidgetItemsType: "Widget",
			// Same as above - no renaming should happen
			wantGetProfileResponse:     "GetWidgetProfileResponse",
			wantUpdateProfileResponse:  "UpdateWidgetProfileResponse",
			wantSettingsAliasItemsType: "WidgetSettingsAlias",
			// FIXED: Array items using WidgetSettings stay as WidgetSettings
			wantListSettingsItemType: "WidgetSettings",
			// FIXED: Direct refs to WidgetSettings stay as WidgetSettings
			wantRetrieveSettings: "WidgetSettings",
			wantCreateSettings:   "WidgetSettings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  openAPIYAML,
				MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
				Fixes:        tt.fixes,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Debug: log the registry contents
			registry := internalOpenAPI.GlobalNestedRefRegistry()
			t.Logf("Registry contents after population:")
			for k, v := range registry.All() {
				t.Logf("  %s -> %s", k, v)
			}

			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			// Get GetWidgetResponse
			getWidgetSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/widgets/{id}").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, getWidgetSchema)

			// Get ListWidgetsResponse
			listWidgetsSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/widgets").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, listWidgetsSchema)

			t.Logf("GetWidgetResponse ref: %s", getWidgetSchema.GetRef())
			t.Logf("ListWidgetsResponse ref: %s", listWidgetsSchema.GetRef())

			// Process GetWidgetResponse
			paramsGet := CreateTestParams(getWidgetSchema, common.DocInfo)
			resultGet, err := schemas.HandleSchema(context.Background(), paramsGet)
			require.NoError(t, err)
			require.NotNil(t, resultGet)

			t.Logf("GetWidgetResponse Name: %s", resultGet.Name)
			t.Logf("GetWidgetResponse Type Name: %s", resultGet.Type.Name)

			// Process ListWidgetsResponse - we need to check the items type
			paramsListWidgets := CreateTestParams(listWidgetsSchema, common.DocInfo)
			resultListWidgets, err := schemas.HandleSchema(context.Background(), paramsListWidgets)
			require.NoError(t, err)
			require.NotNil(t, resultListWidgets)

			t.Logf("ListWidgetsResponse Name: %s", resultListWidgets.Name)
			t.Logf("ListWidgetsResponse Type Name: %s", resultListWidgets.Type.Name)

			// Get the data property and check its items type
			var dataPropType *ast.TypeDef
			for _, field := range resultListWidgets.Type.Fields {
				if field.Name == "data" || field.OriginalName == "data" {
					dataPropType = field.Type
					break
				}
			}
			require.NotNil(t, dataPropType, "ListWidgetsResponse should have 'data' property")
			require.NotNil(t, dataPropType.ItemType, "data property should be an array")

			t.Logf("data.items Type Name: %s", dataPropType.ItemType.Name)

			assert.Equal(t, tt.wantGetWidgetResponse, resultGet.Type.Name,
				"GetWidgetResponse should resolve to "+tt.wantGetWidgetResponse)
			assert.Equal(t, tt.wantListWidgetItemsType, dataPropType.ItemType.Name,
				"ListWidgetsResponse.data.items should resolve to "+tt.wantListWidgetItemsType)

			t.Log("\n=== Additional AliasBeforeTarget scenarios ===")

			// Test GetWidgetProfileResponse
			t.Log("\n--- GetWidgetProfileResponse ---")
			getProfileSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/get-profile").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, getProfileSchema)
			paramsGetProfile := CreateTestParams(getProfileSchema, common.DocInfo)
			resultGetProfile, err := schemas.HandleSchema(context.Background(), paramsGetProfile)
			require.NoError(t, err)
			t.Logf("GetWidgetProfileResponse Type Name: %s", resultGetProfile.Type.Name)
			assert.Equal(t, tt.wantGetProfileResponse, resultGetProfile.Type.Name, "GetWidgetProfileResponse type")

			// Test UpdateWidgetProfileResponse
			t.Log("\n--- UpdateWidgetProfileResponse ---")
			updateProfileSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/update-profile").GetResolvedObject().Put().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, updateProfileSchema)
			paramsUpdateProfile := CreateTestParams(updateProfileSchema, common.DocInfo)
			resultUpdateProfile, err := schemas.HandleSchema(context.Background(), paramsUpdateProfile)
			require.NoError(t, err)
			t.Logf("UpdateWidgetProfileResponse Type Name: %s", resultUpdateProfile.Type.Name)
			assert.Equal(t, tt.wantUpdateProfileResponse, resultUpdateProfile.Type.Name, "UpdateWidgetProfileResponse type")

			// Test /widget-settings-aliases - array items using WidgetSettingsAlias
			t.Log("\n--- /widget-settings-aliases array items (WidgetSettingsAlias) ---")
			settingsAliasListSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/widget-settings-aliases").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, settingsAliasListSchema)
			paramsSettingsAliasList := CreateTestParams(settingsAliasListSchema, common.DocInfo)
			resultSettingsAliasList, err := schemas.HandleSchema(context.Background(), paramsSettingsAliasList)
			require.NoError(t, err)
			var settingsAliasDataType *ast.TypeDef
			for _, field := range resultSettingsAliasList.Type.Fields {
				if field.Name == "data" || field.OriginalName == "data" {
					settingsAliasDataType = field.Type
					break
				}
			}
			require.NotNil(t, settingsAliasDataType, "widget-settings-aliases should have 'data' property")
			require.NotNil(t, settingsAliasDataType.ItemType, "data property should be an array")
			t.Logf("/widget-settings-aliases data.items Type Name: %s", settingsAliasDataType.ItemType.Name)
			assert.Equal(t, tt.wantSettingsAliasItemsType, settingsAliasDataType.ItemType.Name, "/widget-settings-aliases array items type")

			// Test /settings - array items using WidgetSettings
			t.Log("\n--- /settings array items (WidgetSettings) ---")
			settingsListSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/settings").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, settingsListSchema)
			paramsSettingsList := CreateTestParams(settingsListSchema, common.DocInfo)
			resultSettingsList, err := schemas.HandleSchema(context.Background(), paramsSettingsList)
			require.NoError(t, err)
			var settingsDataType *ast.TypeDef
			for _, field := range resultSettingsList.Type.Fields {
				if field.Name == "data" || field.OriginalName == "data" {
					settingsDataType = field.Type
					break
				}
			}
			require.NotNil(t, settingsDataType, "settings should have 'data' property")
			require.NotNil(t, settingsDataType.ItemType, "data property should be an array")
			t.Logf("/settings data.items Type Name: %s", settingsDataType.ItemType.Name)
			assert.Equal(t, tt.wantListSettingsItemType, settingsDataType.ItemType.Name, "/settings array items type")

			// Test /settings/{id} - direct ref to WidgetSettings (RetrieveSettings)
			t.Log("\n--- /settings/{id} (direct ref to WidgetSettings) ---")
			retrieveConfigSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/settings/{id}").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, retrieveConfigSchema)
			paramsRetrieveConfig := CreateTestParams(retrieveConfigSchema, common.DocInfo)
			resultRetrieveConfig, err := schemas.HandleSchema(context.Background(), paramsRetrieveConfig)
			require.NoError(t, err)
			t.Logf("RetrieveSettings Type Name: %s", resultRetrieveConfig.Type.Name)
			assert.Equal(t, tt.wantRetrieveSettings, resultRetrieveConfig.Type.Name, "RetrieveSettings type")

			// Test /create-settings - direct ref to WidgetSettings (CreateSettings)
			t.Log("\n--- /create-settings (direct ref to WidgetSettings) ---")
			createConfigSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/create-settings").GetResolvedObject().Post().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, createConfigSchema)
			paramsCreateConfig := CreateTestParams(createConfigSchema, common.DocInfo)
			resultCreateConfig, err := schemas.HandleSchema(context.Background(), paramsCreateConfig)
			require.NoError(t, err)
			t.Logf("CreateSettings Type Name: %s", resultCreateConfig.Type.Name)
			assert.Equal(t, tt.wantCreateSettings, resultCreateConfig.Type.Name, "CreateSettings type")
		})
	}
}

// TestTargetFirstLikeScenario tests a scenario similar to the TargetFirst spec where:
// - The alias (AliasSchema) comes AFTER the shared schema (SharedSchema)
// - ContainerSchema has a property with direct ref to SharedSchema
//
// In the buggy libopenapi behavior, the direct ref becomes "AliasSchema".
func TestTargetFirstLikeScenario(t *testing.T) {
	// Document order: SharedSchema (actual) -> AliasSchema (refs SharedSchema) -> ContainerSchema (has property with direct ref)
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /alias:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/AliasSchema"
  /container:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ContainerSchema"
components:
  schemas:
    SharedSchema:
      type: object
      properties:
        value:
          type: string
    AliasSchema:
      $ref: "#/components/schemas/SharedSchema"
    ContainerSchema:
      type: object
      properties:
        nested:
          $ref: "#/components/schemas/SharedSchema"
`

	tests := []struct {
		name                    string
		fixes                   *config.Fixes
		wantAliasType           string
		wantContainerNestedType string
	}{
		{
			name: "Buggy behavior - alias before direct ref",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: false,
			},
			wantAliasType: "AliasSchema",
			// In TargetFirst, OLD libopenapi would return the alias name for direct refs
			// because the alias was encountered first and "claimed" the shared schema
			wantContainerNestedType: "AliasSchema",
		},
		{
			name: "Fixed behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: true,
			},
			wantAliasType:           "AliasSchema",
			wantContainerNestedType: "SharedSchema",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  openAPIYAML,
				MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
				Fixes:        tt.fixes,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Debug: log the registry contents
			registry := internalOpenAPI.GlobalNestedRefRegistry()
			t.Logf("Registry contents after population:")
			for k, v := range registry.All() {
				t.Logf("  %s -> %s", k, v)
			}

			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			// Get AliasSchema (the alias)
			aliasSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/alias").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, aliasSchema)

			// Get ContainerSchema
			containerSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/container").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, containerSchema)

			t.Logf("AliasSchema ref: %s", aliasSchema.GetRef())
			t.Logf("ContainerSchema ref: %s", containerSchema.GetRef())

			// Process AliasSchema
			paramsAlias := CreateTestParams(aliasSchema, common.DocInfo)
			resultAlias, err := schemas.HandleSchema(context.Background(), paramsAlias)
			require.NoError(t, err)
			require.NotNil(t, resultAlias)

			t.Logf("AliasSchema Name: %s", resultAlias.Name)
			t.Logf("AliasSchema Type Name: %s", resultAlias.Type.Name)

			// Process ContainerSchema - we need to check the nested property type
			paramsContainer := CreateTestParams(containerSchema, common.DocInfo)
			resultContainer, err := schemas.HandleSchema(context.Background(), paramsContainer)
			require.NoError(t, err)
			require.NotNil(t, resultContainer)

			t.Logf("ContainerSchema Name: %s", resultContainer.Name)
			t.Logf("ContainerSchema Type Name: %s", resultContainer.Type.Name)

			// Get the nested property and check its type
			var nestedPropType *ast.TypeDef
			for _, field := range resultContainer.Type.Fields {
				if field.Name == "nested" || field.OriginalName == "nested" {
					nestedPropType = field.Type
					break
				}
			}
			require.NotNil(t, nestedPropType, "ContainerSchema should have 'nested' property")

			t.Logf("ContainerSchema.nested Type Name: %s", nestedPropType.Name)

			// Log the reference chain for debugging
			t.Logf("\n=== Reference Chain Analysis ===")
			resolvedAlias := aliasSchema.GetResolvedSchema()
			if resolvedAlias != nil {
				aliasChain := resolvedAlias.GetReferenceChain()
				t.Logf("AliasSchema resolved chain length: %d", len(aliasChain))
				for i, entry := range aliasChain {
					t.Logf("  Chain[%d]: %s", i, entry.Reference)
				}
			}

			assert.Equal(t, tt.wantAliasType, resultAlias.Type.Name,
				"AliasSchema should resolve to "+tt.wantAliasType)
			assert.Equal(t, tt.wantContainerNestedType, nestedPropType.Name,
				"ContainerSchema.nested should resolve to "+tt.wantContainerNestedType)
		})
	}
}

// TestMixedOrderingScenario tests a document with BOTH orderings in the same file:

// - TargetFirst-like: Target (TargetFirstTarget) defined BEFORE Alias (TargetFirstAlias)
//
// In buggy libopenapi behavior, ALL direct refs are replaced with the alias name.
func TestMixedOrderingScenario(t *testing.T) {
	// Document order:

	// 4. TargetFirstTarget (actual schema) - target BEFORE alias
	// 5. TargetFirstAlias (alias to TargetFirstTarget) - alias AFTER target
	// 6. TargetFirstContainer (has property with direct ref to TargetFirstTarget)
	openAPIYAML := `openapi: 3.0.0
info:
  title: Mixed Ordering Test API
  version: 1.0.0
paths:
  /aliasBeforeTarget-alias:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/AliasBeforeTargetAlias"
  /aliasBeforeTarget-container:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/AliasBeforeTargetContainer"
  /targetFirst-alias:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/TargetFirstAlias"
  /targetFirst-container:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/TargetFirstContainer"
components:
  schemas:

    AliasBeforeTargetAlias:
      $ref: "#/components/schemas/AliasBeforeTargetTarget"
    AliasBeforeTargetContainer:
      type: object
      properties:
        directRef:
          $ref: "#/components/schemas/AliasBeforeTargetTarget"
    AliasBeforeTargetTarget:
      type: object
      properties:
        id:
          type: string
    # TargetFirst-like ordering: Target BEFORE Alias
    TargetFirstTarget:
      type: object
      properties:
        value:
          type: string
    TargetFirstAlias:
      $ref: "#/components/schemas/TargetFirstTarget"
    TargetFirstContainer:
      type: object
      properties:
        directRef:
          $ref: "#/components/schemas/TargetFirstTarget"
`

	tests := []struct {
		name                               string
		fixes                              *config.Fixes
		wantAliasBeforeTargetAliasType     string
		wantAliasBeforeTargetDirectRefType string
		wantTargetFirstAliasType           string
		wantTargetFirstDirectRefType       string
	}{
		{
			name: "Buggy behavior - mixed orderings",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: false,
			},
			// In buggy behavior, ALL direct refs are replaced with the alias name
			wantAliasBeforeTargetAliasType:     "AliasBeforeTargetAlias",
			wantAliasBeforeTargetDirectRefType: "AliasBeforeTargetAlias",
			wantTargetFirstAliasType:           "TargetFirstAlias",
			wantTargetFirstDirectRefType:       "TargetFirstAlias",
		},
		{
			name: "Fixed behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: true,
			},
			// Fixed: no replacements for any direct refs
			wantAliasBeforeTargetAliasType:     "AliasBeforeTargetAlias",
			wantAliasBeforeTargetDirectRefType: "AliasBeforeTargetTarget",
			wantTargetFirstAliasType:           "TargetFirstAlias",
			wantTargetFirstDirectRefType:       "TargetFirstTarget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  openAPIYAML,
				MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
				Fixes:        tt.fixes,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Debug: log the registry contents
			registry := internalOpenAPI.GlobalNestedRefRegistry()
			t.Logf("Registry contents after population:")
			for k, v := range registry.All() {
				t.Logf("  %s -> %s", k, v)
			}

			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			t.Log("\n=== AliasBeforeTarget-like (alias before target) ===")

			aliasBeforeTargetAliasSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/aliasBeforeTarget-alias").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, aliasBeforeTargetAliasSchema)

			aliasBeforeTargetContainerSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/aliasBeforeTarget-container").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, aliasBeforeTargetContainerSchema)

			paramsAliasBeforeTargetAlias := CreateTestParams(aliasBeforeTargetAliasSchema, common.DocInfo)
			resultAliasBeforeTargetAlias, err := schemas.HandleSchema(context.Background(), paramsAliasBeforeTargetAlias)
			require.NoError(t, err)
			require.NotNil(t, resultAliasBeforeTargetAlias)
			t.Logf("AliasBeforeTargetAlias Type Name: %s", resultAliasBeforeTargetAlias.Type.Name)

			paramsAliasBeforeTargetContainer := CreateTestParams(aliasBeforeTargetContainerSchema, common.DocInfo)
			resultAliasBeforeTargetContainer, err := schemas.HandleSchema(context.Background(), paramsAliasBeforeTargetContainer)
			require.NoError(t, err)
			require.NotNil(t, resultAliasBeforeTargetContainer)

			var aliasBeforeTargetDirectRefType *ast.TypeDef
			for _, field := range resultAliasBeforeTargetContainer.Type.Fields {
				if field.Name == "directRef" || field.OriginalName == "directRef" {
					aliasBeforeTargetDirectRefType = field.Type
					break
				}
			}
			require.NotNil(t, aliasBeforeTargetDirectRefType, "AliasBeforeTargetContainer should have 'directRef' property")
			t.Logf("AliasBeforeTargetContainer.directRef Type Name: %s", aliasBeforeTargetDirectRefType.Name)

			// === TargetFirst-like tests ===
			t.Log("\n=== TargetFirst-like (target before alias) ===")

			targetFirstAliasSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/targetFirst-alias").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, targetFirstAliasSchema)

			targetFirstContainerSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/targetFirst-container").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, targetFirstContainerSchema)

			// Process TargetFirstAlias
			paramsTargetFirstAlias := CreateTestParams(targetFirstAliasSchema, common.DocInfo)
			resultTargetFirstAlias, err := schemas.HandleSchema(context.Background(), paramsTargetFirstAlias)
			require.NoError(t, err)
			require.NotNil(t, resultTargetFirstAlias)
			t.Logf("TargetFirstAlias Type Name: %s", resultTargetFirstAlias.Type.Name)

			// Process TargetFirstContainer - check the directRef property
			paramsTargetFirstContainer := CreateTestParams(targetFirstContainerSchema, common.DocInfo)
			resultTargetFirstContainer, err := schemas.HandleSchema(context.Background(), paramsTargetFirstContainer)
			require.NoError(t, err)
			require.NotNil(t, resultTargetFirstContainer)

			var targetFirstDirectRefType *ast.TypeDef
			for _, field := range resultTargetFirstContainer.Type.Fields {
				if field.Name == "directRef" || field.OriginalName == "directRef" {
					targetFirstDirectRefType = field.Type
					break
				}
			}
			require.NotNil(t, targetFirstDirectRefType, "TargetFirstContainer should have 'directRef' property")
			t.Logf("TargetFirstContainer.directRef Type Name: %s", targetFirstDirectRefType.Name)

			// === Assertions ===
			t.Log("\n=== Assertions ===")

			assert.Equal(t, tt.wantAliasBeforeTargetAliasType, resultAliasBeforeTargetAlias.Type.Name,
				"AliasBeforeTargetAlias should resolve to "+tt.wantAliasBeforeTargetAliasType)
			assert.Equal(t, tt.wantAliasBeforeTargetDirectRefType, aliasBeforeTargetDirectRefType.Name,
				"AliasBeforeTargetContainer.directRef should resolve to "+tt.wantAliasBeforeTargetDirectRefType)

			// TargetFirst-like assertions
			assert.Equal(t, tt.wantTargetFirstAliasType, resultTargetFirstAlias.Type.Name,
				"TargetFirstAlias should resolve to "+tt.wantTargetFirstAliasType)
			assert.Equal(t, tt.wantTargetFirstDirectRefType, targetFirstDirectRefType.Name,
				"TargetFirstContainer.directRef should resolve to "+tt.wantTargetFirstDirectRefType)
		})
	}
}

// TestLibopenAPIReferenceAccessBehaviour tests the behavior when accessing a schema
// directly from components (not via $ref resolution in paths).
//
// When you access Schema1 directly (which is a $ref to SchemaShared), libopenapi
// returns "SchemaShared" as the name because there's no intermediate reference chain.
func TestLibopenAPIReferenceAccessBehaviour(t *testing.T) {
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /endpoint1:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/SchemaShared"
components:
  schemas:
    Schema1:
      $ref: "#/components/schemas/SchemaShared"
    SchemaShared:
      type: object
      properties:
        name:
          type: string
        value:
          type: integer
`

	tests := []struct {
		name                   string
		fixes                  *config.Fixes
		wantPathSchemaResult   string
		wantDirectAccessResult string
	}{
		{
			name: "Buggy behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: false,
			},
			// When accessed via path $ref, we get SchemaShared
			wantPathSchemaResult: "SchemaShared",
			// When accessed directly from components, Schema1 resolves to SchemaShared
			wantDirectAccessResult: "SchemaShared",
		},
		{
			name: "Fixed behavior",
			fixes: &config.Fixes{
				SharedNestedComponentsJan2026: true,
			},
			wantPathSchemaResult:   "SchemaShared",
			wantDirectAccessResult: "SchemaShared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  openAPIYAML,
				MockFeatures: testutils.NewMockFeaturesConfig().WithSupportAllFeatures(),
				Fixes:        tt.fixes,
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

			// Access schema via path $ref
			pathSchema := common.DocInfo.Doc.GetPaths().GetOrZero("/endpoint1").GetResolvedObject().Get().GetResponses().GetOrZero("200").GetResolvedObject().GetContent().GetOrZero("application/json").GetSchema()
			require.NotNil(t, pathSchema)

			// Access Schema1 directly from components
			componentsSchemas := common.DocInfo.Doc.GetComponents().GetSchemas()
			directSchema, _ := componentsSchemas.Get("Schema1")
			require.NotNil(t, directSchema)

			t.Logf("Path schema ref: %s", pathSchema.GetRef())
			t.Logf("Direct schema ref: %s", directSchema.GetRef())

			// Process path schema through HandleSchema
			paramsPath := CreateTestParams(pathSchema, common.DocInfo)
			resultPath, err := schemas.HandleSchema(context.Background(), paramsPath)
			require.NoError(t, err)
			require.NotNil(t, resultPath)

			// Process directly accessed schema through HandleSchema
			paramsDirect := CreateTestParams(directSchema, common.DocInfo)
			resultDirect, err := schemas.HandleSchema(context.Background(), paramsDirect)
			require.NoError(t, err)
			require.NotNil(t, resultDirect)

			t.Logf("Path schema result Name: %s", resultPath.Name)
			t.Logf("Path schema result Type Name: %s", resultPath.Type.Name)
			t.Logf("Direct schema result Name: %s", resultDirect.Name)
			t.Logf("Direct schema result Type Name: %s", resultDirect.Type.Name)

			assert.Equal(t, tt.wantPathSchemaResult, resultPath.Type.Name,
				"Path schema should resolve to "+tt.wantPathSchemaResult)
			assert.Equal(t, tt.wantDirectAccessResult, resultDirect.Type.Name,
				"Directly accessed schema should resolve to "+tt.wantDirectAccessResult)
		})
	}
}
