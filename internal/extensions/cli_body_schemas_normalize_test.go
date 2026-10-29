package extensions

import (
	"bytes"
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func collectBodySchemaVariantsFromYAML(t *testing.T, fullYAML string, variants ...CLIBodySchemaVariant) map[string]string {
	t.Helper()
	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(fullYAML)))
	require.NoError(t, err)
	e := &Extensions{}
	schemas, err := e.CollectCLIBodySchemas(ctx, &document.DocumentInfo{Doc: doc}, variants...)
	require.NoError(t, err)
	return schemas
}

const requestUnionSpec = `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /tasks:
    post:
      operationId: createTask
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateTask"
      responses:
        "200":
          description: ok
components:
  schemas:
    CreateTask:
      oneOf:
        - $ref: "#/components/schemas/RenderTask"
        - $ref: "#/components/schemas/WorkflowTask"
    RenderTask:
      type: object
      required: [id, engine, status]
      properties:
        id:
          type: string
          readOnly: true
        status:
          $ref: "#/components/schemas/TaskStatus"
        engine:
          $ref: "#/components/schemas/Engine"
        secret:
          type: string
          writeOnly: true
    WorkflowTask:
      type: object
      required: [workflow]
      properties:
        workflow:
          type: string
    Engine:
      type: string
      enum: [fast, slow]
    TaskStatus:
      type: string
      readOnly: true
`

func TestCollectCLIBodySchemas_PinnedVariantBundle(t *testing.T) {
	schemas := collectBodySchemaVariantsFromYAML(t, requestUnionSpec,
		CLIBodySchemaVariant{OperationID: "createTask", RequestVariant: "#/components/schemas/RenderTask"},
		CLIBodySchemaVariant{OperationID: "createTask", RequestVariant: "#/components/schemas/RenderTask"},
		CLIBodySchemaVariant{OperationID: "createTask", RequestVariant: "#/paths/~1tasks/post/requestBody"},
		CLIBodySchemaVariant{OperationID: "missingOperation", RequestVariant: "#/components/schemas/WorkflowTask"},
	)
	assert.Len(t, schemas, 2)

	union := decodeBodySchema(t, schemas["createTask"])
	assert.Equal(t, "#/$defs/CreateTask", union["$ref"])

	key := CLIBodySchemaVariantKey("createTask", "#/components/schemas/RenderTask")
	assert.Equal(t, "createTask#RenderTask", key)
	variant := decodeBodySchema(t, schemas[key])
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", variant["$schema"])
	assert.Equal(t, "#/$defs/RenderTask", variant["$ref"])
	defs := variant["$defs"].(map[string]any)
	assert.Contains(t, defs, "RenderTask")
	assert.Contains(t, defs, "Engine")
	assert.NotContains(t, defs, "WorkflowTask")
	assert.NotContains(t, defs, "CreateTask")
}

func TestCollectCLIBodySchemas_RemovesReadOnlyProperties(t *testing.T) {
	first := collectBodySchemaVariantsFromYAML(t, requestUnionSpec)
	second := collectBodySchemaVariantsFromYAML(t, requestUnionSpec)
	assert.Equal(t, first, second)

	render := decodeBodySchema(t, first["createTask"])["$defs"].(map[string]any)["RenderTask"].(map[string]any)
	props := render["properties"].(map[string]any)
	assert.NotContains(t, props, "id", "inline readOnly property")
	assert.NotContains(t, props, "status", "property referencing a readOnly definition")
	assert.Contains(t, props, "engine")
	assert.Contains(t, props, "secret", "writeOnly properties are request fields")
	assert.Equal(t, []any{"engine"}, render["required"])
}

func TestCollectCLIBodySchemas_ConvertsOpenAPI30Schemas(t *testing.T) {
	spec := `openapi: 3.0.3
info:
  title: t
  version: 1.0.0
paths:
  /things:
    post:
      operationId: createThing
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required: [created, name]
              properties:
                created:
                  type: string
                  readOnly: true
                name:
                  type: string
                  nullable: true
                mode:
                  type: string
                  enum: [a, b]
                  nullable: true
                parent:
                  description: Parent thing
                  nullable: true
                  allOf:
                    - $ref: "#/components/schemas/Parent"
                count:
                  type: integer
                  minimum: 0
                  exclusiveMinimum: true
                  maximum: 10
                  exclusiveMaximum: false
                plain:
                  type: string
                  nullable: false
      responses:
        "200":
          description: ok
components:
  schemas:
    Parent:
      type: object
      properties:
        label:
          type: string
          nullable: true
`
	schemas := collectBodySchemaVariantsFromYAML(t, spec)
	bundled := decodeBodySchema(t, schemas["createThing"])
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", bundled["$schema"])
	assert.Equal(t, []any{"name"}, bundled["required"])

	props := bundled["properties"].(map[string]any)
	assert.NotContains(t, props, "created")
	assert.Equal(t, map[string]any{"type": []any{"string", "null"}}, props["name"])
	assert.Equal(t, map[string]any{"type": []any{"string", "null"}, "enum": []any{"a", "b", nil}}, props["mode"])
	assert.Equal(t, map[string]any{
		"description": "Parent thing",
		"anyOf": []any{
			map[string]any{"allOf": []any{map[string]any{"$ref": "#/$defs/Parent"}}},
			map[string]any{"type": "null"},
		},
	}, props["parent"])
	assert.Equal(t, map[string]any{"type": "integer", "exclusiveMinimum": float64(0), "maximum": float64(10)}, props["count"])
	assert.Equal(t, map[string]any{"type": "string"}, props["plain"])

	parent := bundled["$defs"].(map[string]any)["Parent"].(map[string]any)
	assert.Equal(t, map[string]any{"type": []any{"string", "null"}}, parent["properties"].(map[string]any)["label"])
}

func TestCollectCLIBodySchemas_RemovesReadOnlyThroughCompositionAndPointers(t *testing.T) {
	spec := `openapi: 3.0.3
info:
  title: t
  version: 1.0.0
paths:
  /things:
    post:
      operationId: createThing
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Child"
      responses:
        "200":
          description: ok
components:
  schemas:
    Child:
      allOf:
        - $ref: "#/components/schemas/Base"
        - type: object
          required: [id, name, audit, stamp]
          properties:
            name:
              type: string
            audit:
              nullable: true
              allOf:
                - $ref: "#/components/schemas/Audit"
            stamp:
              $ref: "#/components/schemas/Holder/properties/stamp"
    Base:
      type: object
      properties:
        id:
          type: string
          readOnly: true
    Audit:
      type: object
      readOnly: true
    Holder:
      type: object
      properties:
        stamp:
          type: string
          readOnly: true
`
	schemas := collectBodySchemaVariantsFromYAML(t, spec)
	defs := decodeBodySchema(t, schemas["createThing"])["$defs"].(map[string]any)

	own := defs["Child"].(map[string]any)["allOf"].([]any)[1].(map[string]any)
	assert.Equal(t, []any{"name"}, own["required"], "required entries from allOf siblings and references are removed")
	props := own["properties"].(map[string]any)
	assert.Contains(t, props, "name")
	assert.NotContains(t, props, "audit", "nullable reference to a readOnly definition")
	assert.NotContains(t, props, "stamp", "deep pointer to a readOnly property")

	assert.NotContains(t, defs["Base"].(map[string]any)["properties"].(map[string]any), "id")
}

func TestCollectCLIBodySchemas_NullableRootKeepsDefinitions(t *testing.T) {
	spec := `openapi: 3.0.3
info:
  title: t
  version: 1.0.0
paths:
  /things:
    post:
      operationId: createThing
      requestBody:
        content:
          application/json:
            schema:
              nullable: true
              allOf:
                - $ref: "#/components/schemas/Parent"
      responses:
        "200":
          description: ok
components:
  schemas:
    Parent:
      type: object
      properties:
        label:
          type: string
`
	schemas := collectBodySchemaVariantsFromYAML(t, spec)
	bundled := decodeBodySchema(t, schemas["createThing"])
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", bundled["$schema"])
	assert.Contains(t, bundled["$defs"], "Parent")
	inner := bundled["anyOf"].([]any)[0].(map[string]any)
	assert.NotContains(t, inner, "$defs")
	assert.NotContains(t, inner, "$schema")
	assert.Equal(t, []any{map[string]any{"$ref": "#/$defs/Parent"}}, inner["allOf"])
}

func TestCollectCLIBodySchemas_BooleanSchemasDeclareDraft(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /accept:
    post:
      operationId: acceptAnything
      requestBody:
        content:
          application/json:
            schema: true
      responses:
        "200":
          description: ok
  /reject:
    post:
      operationId: rejectEverything
      requestBody:
        content:
          application/json:
            schema: false
      responses:
        "200":
          description: ok
`
	schemas := collectBodySchemaVariantsFromYAML(t, spec)
	assert.JSONEq(t, `{"$schema":"https://json-schema.org/draft/2020-12/schema"}`, schemas["acceptAnything"])
	assert.JSONEq(t, `{"$schema":"https://json-schema.org/draft/2020-12/schema","not":{}}`, schemas["rejectEverything"])
}

func TestCollectCLIBodySchemas_KeepsOpenAPI31Keywords(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /things:
    post:
      operationId: createThing
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                nullable:
                  type: [string, "null"]
                count:
                  type: integer
                  exclusiveMinimum: 0
      responses:
        "200":
          description: ok
`
	schemas := collectBodySchemaVariantsFromYAML(t, spec)
	props := decodeBodySchema(t, schemas["createThing"])["properties"].(map[string]any)
	assert.Contains(t, props, "nullable", "a property named like a keyword is kept")
	assert.Equal(t, map[string]any{"type": "integer", "exclusiveMinimum": float64(0)}, props["count"])
}

func TestCLICommandManifest_PinnedRequestVariants(t *testing.T) {
	manifest := &CLICommandManifest{Commands: []CLICommand{
		{Source: CLICommandSource{Type: CLICommandSourceOperation, Routes: []CLICommandRoute{
			{OperationID: "createTask", RequestVariant: "#/components/schemas/RenderTask"},
		}}},
		{Source: CLICommandSource{Type: CLICommandSourceOperation, Routes: []CLICommandRoute{
			{OperationID: "createTask", RequestVariant: "#/components/schemas/RenderTask"},
			{OperationID: "createTask", RequestVariant: "#/components/schemas/WorkflowTask"},
		}}},
		{Source: CLICommandSource{Type: CLICommandSourceOperation, Routes: []CLICommandRoute{
			{OperationID: "listTasks"},
		}}},
		{Source: CLICommandSource{Type: CLICommandSourceGroup}},
	}}
	assert.Equal(t, []CLIBodySchemaVariant{
		{OperationID: "createTask", RequestVariant: "#/components/schemas/RenderTask"},
	}, manifest.PinnedRequestVariants())
	assert.Nil(t, (*CLICommandManifest)(nil).PinnedRequestVariants())
}
