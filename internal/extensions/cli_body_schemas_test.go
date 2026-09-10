package extensions

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func collectBodySchemasFromYAML(t *testing.T, fullYAML string) (map[string]string, error) {
	t.Helper()
	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(fullYAML)))
	require.NoError(t, err)
	e := &Extensions{}
	return e.CollectCLIBodySchemas(ctx, &document.DocumentInfo{Doc: doc})
}

func collectBodySchemasFromFiles(t *testing.T, files map[string]string) (map[string]string, error) {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(files["openapi.yaml"])))
	require.NoError(t, err)
	e := &Extensions{}
	return e.CollectCLIBodySchemas(ctx, &document.DocumentInfo{Doc: doc, SchemaPath: filepath.Join(dir, "openapi.yaml")})
}

func decodeBodySchema(t *testing.T, schema string) map[string]any {
	t.Helper()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(schema), &decoded))
	return decoded
}

func schemaRefAt(t *testing.T, schema map[string]any, path ...string) string {
	t.Helper()
	current := schema
	for _, key := range path {
		next, ok := current[key].(map[string]any)
		require.True(t, ok, "expected object at %q", key)
		current = next
	}
	ref, _ := current["$ref"].(string)
	return ref
}

const externalComponentsYAML = `components:
  schemas:
    External:
      type: object
      properties:
        nested:
          $ref: "#/components/schemas/Nested"
        again:
          $ref: "other.yaml#/components/schemas/Nested"
    Nested:
      type: string
      enum: [a, b]
`

func TestCollectCLIBodySchemas_BundlesTransitiveComponents(t *testing.T) {
	schemas, err := collectBodySchemasFromYAML(t, cliCommandsTestSpec)
	require.NoError(t, err)
	require.Contains(t, schemas, "CreateTask")

	var bundled map[string]any
	require.NoError(t, json.Unmarshal([]byte(schemas["CreateTask"]), &bundled))
	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", bundled["$schema"])

	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok, "expected $defs in bundled schema")
	// Transitive closure: the union members plus everything they reference.
	for _, name := range []string{"CreateRenderTaskParams", "CreateWorkflowTaskParams", "TaskBase", "ContentPart", "EngineOption", "Modality"} {
		assert.Contains(t, defs, name)
	}
	// Every reference is rewritten into the bundle; none dangle.
	assert.NotContains(t, schemas["CreateTask"], "#/components/schemas/")

	// Operations without an application/json request body are absent.
	assert.NotContains(t, schemas, "GetTask")
	assert.NotContains(t, schemas, "ListTasks")
}

func TestCollectCLIBodySchemas_ExternalRefsAreBundled(t *testing.T) {
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
              description: 'See other.yaml#/components/schemas/External for details.'
              properties:
                external:
                  $ref: "other.yaml#/components/schemas/External"
                name:
                  type: string
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromFiles(t, map[string]string{"openapi.yaml": spec, "other.yaml": externalComponentsYAML})
	require.NoError(t, err)
	require.Contains(t, schemas, "createThing")

	bundled := decodeBodySchema(t, schemas["createThing"])
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok, "expected $defs in bundled schema")
	assert.Contains(t, defs, "External")
	assert.Contains(t, defs, "Nested")
	assert.Equal(t, "#/$defs/External", schemaRefAt(t, bundled, "properties", "external"))
	assert.Equal(t, "#/$defs/Nested", schemaRefAt(t, bundled, "$defs", "External", "properties", "nested"))
	assert.Equal(t, "#/$defs/Nested", schemaRefAt(t, bundled, "$defs", "External", "properties", "again"))
	assert.Equal(t, []any{"a", "b"}, defs["Nested"].(map[string]any)["enum"])
	assert.Equal(t, "See other.yaml#/components/schemas/External for details.", bundled["description"])
	walkSchemaRefs(bundled, func(ref string) string {
		assert.NotContains(t, ref, "other.yaml")
		return ref
	})
}

func TestCollectCLIBodySchemas_ExternalRootRefIsBundled(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /things:
    put:
      operationId: putThing
      requestBody:
        content:
          application/json:
            schema:
              $ref: "other.yaml#/components/schemas/External"
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromFiles(t, map[string]string{"openapi.yaml": spec, "other.yaml": externalComponentsYAML})
	require.NoError(t, err)
	require.Contains(t, schemas, "putThing")

	bundled := decodeBodySchema(t, schemas["putThing"])
	assert.Equal(t, "#/$defs/External", bundled["$ref"])
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok, "expected $defs in bundled schema")
	assert.Contains(t, defs, "External")
	assert.Contains(t, defs, "Nested")
	assert.NotContains(t, schemas["putThing"], "other.yaml")
}

func TestCollectCLIBodySchemas_ExternalAliasAvoidsComponentCollision(t *testing.T) {
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
                local:
                  $ref: "#/components/schemas/External"
                external:
                  $ref: "other.yaml#/components/schemas/External"
      responses:
        "200":
          description: ok
components:
  schemas:
    External:
      type: integer
`
	schemas, err := collectBodySchemasFromFiles(t, map[string]string{"openapi.yaml": spec, "other.yaml": externalComponentsYAML})
	require.NoError(t, err)

	bundled := decodeBodySchema(t, schemas["createThing"])
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok, "expected $defs in bundled schema")
	assert.Equal(t, "integer", defs["External"].(map[string]any)["type"])
	assert.Equal(t, "object", defs["External_2"].(map[string]any)["type"])
	assert.Contains(t, defs, "Nested")
	assert.Equal(t, "#/$defs/External", schemaRefAt(t, bundled, "properties", "local"))
	assert.Equal(t, "#/$defs/External_2", schemaRefAt(t, bundled, "properties", "external"))
	assert.Equal(t, "#/$defs/Nested", schemaRefAt(t, bundled, "$defs", "External_2", "properties", "nested"))
}

func TestCollectCLIBodySchemas_UnresolvableExternalRefIsAnError(t *testing.T) {
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
                external:
                  $ref: "missing.yaml#/components/schemas/External"
      responses:
        "200":
          description: ok
`
	_, err := collectBodySchemasFromFiles(t, map[string]string{"openapi.yaml": spec})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `request body schema for createThing references "missing.yaml#/components/schemas/External"`)
}

func TestCollectCLIBodySchemas_IgnoredOperationsAreSkipped(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /ignored:
    post:
      operationId: ignoredOp
      x-speakeasy-ignore: true
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Missing"
      responses:
        "200":
          description: ok
  /ignored-path:
    x-speakeasy-ignore: true
    post:
      operationId: ignoredPathOp
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Missing"
      responses:
        "200":
          description: ok
  /kept:
    post:
      operationId: keptOp
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	require.Len(t, schemas, 1)
	assert.Contains(t, schemas, "keptOp")
	assert.NotContains(t, schemas, "ignoredOp")
	assert.NotContains(t, schemas, "ignoredPathOp")
}

func TestCollectCLIBodySchemas_NonObjectRootsAreEmitted(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /tags:
    put:
      operationId: replaceTags
      requestBody:
        content:
          application/json:
            schema:
              type: array
              items:
                $ref: "#/components/schemas/Tag"
      responses:
        "200":
          description: ok
  /name:
    put:
      operationId: setName
      requestBody:
        content:
          application/json:
            schema:
              type: string
              minLength: 1
      responses:
        "200":
          description: ok
components:
  schemas:
    Tag:
      type: string
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)

	tags := decodeBodySchema(t, schemas["replaceTags"])
	assert.Equal(t, "array", tags["type"])
	assert.Equal(t, "#/$defs/Tag", schemaRefAt(t, tags, "items"))
	assert.Contains(t, tags["$defs"], "Tag")

	name := decodeBodySchema(t, schemas["setName"])
	assert.Equal(t, "string", name["type"])
	assert.InDelta(t, float64(1), name["minLength"], 0)
}

func TestCollectCLIBodySchemas_MissingComponentIsAnError(t *testing.T) {
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
                broken:
                  $ref: "#/components/schemas/DoesNotExist"
      responses:
        "200":
          description: ok
`
	_, err := collectBodySchemasFromYAML(t, spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DoesNotExist")
}

func TestCollectCLIBodySchemas_DeepLocalRefsBundleTheComponent(t *testing.T) {
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
                nested:
                  $ref: "#/components/schemas/Wrapper/properties/inner"
      responses:
        "200":
          description: ok
components:
  schemas:
    Wrapper:
      type: object
      properties:
        inner:
          type: string
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	require.Contains(t, schemas, "createThing")
	// The referenced component is bundled and the deep suffix preserved, so
	// the emitted reference resolves inside the printed document.
	assert.Contains(t, schemas["createThing"], `"$ref":"#/$defs/Wrapper/properties/inner"`)
	var bundled map[string]any
	require.NoError(t, json.Unmarshal([]byte(schemas["createThing"]), &bundled))
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, defs, "Wrapper")
}

func TestCollectCLIBodySchemas_NonStringMapKeysAreNormalized(t *testing.T) {
	// YAML permits non-string mapping keys (booleans, numbers); the bundle
	// must stringify them so the schema stays JSON-marshalable.
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
                mode:
                  type: string
                  x-lookup:
                    true: enabled
                    1: numbered
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	require.Contains(t, schemas, "createThing")
	assert.Contains(t, schemas["createThing"], `"true":"enabled"`)
	assert.Contains(t, schemas["createThing"], `"1":"numbered"`)
}

func TestCollectCLIBodySchemas_PreservesRootLocalDefs(t *testing.T) {
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
              $defs:
                Local:
                  type: string
              properties:
                local:
                  $ref: "#/$defs/Local"
                shared:
                  $ref: "#/components/schemas/Shared"
      responses:
        "200":
          description: ok
components:
  schemas:
    Shared:
      type: integer
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)

	var bundled map[string]any
	require.NoError(t, json.Unmarshal([]byte(schemas["createThing"]), &bundled))
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, defs, "Local")
	assert.Contains(t, defs, "Shared")
	assert.Contains(t, schemas["createThing"], `"$ref":"#/$defs/Local"`)
	assert.Contains(t, schemas["createThing"], `"$ref":"#/$defs/Shared"`)
}

func TestCollectCLIBodySchemas_LocalDefComponentCollisionBundlesAlias(t *testing.T) {
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
              $defs:
                Shared:
                  type: string
              properties:
                local:
                  $ref: "#/$defs/Shared"
                shared:
                  $ref: "#/components/schemas/Shared"
                deep:
                  $ref: "#/components/schemas/Shared/properties/inner"
      responses:
        "200":
          description: ok
components:
  schemas:
    Shared:
      type: object
      properties:
        inner:
          type: integer
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)

	var bundled map[string]any
	require.NoError(t, json.Unmarshal([]byte(schemas["createThing"]), &bundled))
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, defs, "Shared")
	require.Contains(t, defs, "Shared_2")
	assert.Equal(t, map[string]any{"type": "string"}, defs["Shared"])

	// Local references keep their target; component references follow the
	// alias, deep suffixes included.
	assert.Contains(t, schemas["createThing"], `"local":{"$ref":"#/$defs/Shared"}`)
	assert.Contains(t, schemas["createThing"], `"shared":{"$ref":"#/$defs/Shared_2"}`)
	assert.Contains(t, schemas["createThing"], `"deep":{"$ref":"#/$defs/Shared_2/properties/inner"}`)
}

func TestCollectCLIBodySchemas_CollisionAliasSkipsTakenNames(t *testing.T) {
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
              $defs:
                Shared:
                  type: string
                Shared_2:
                  type: boolean
              properties:
                shared:
                  $ref: "#/components/schemas/Shared"
      responses:
        "200":
          description: ok
components:
  schemas:
    Shared:
      type: integer
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)

	var bundled map[string]any
	require.NoError(t, json.Unmarshal([]byte(schemas["createThing"]), &bundled))
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, map[string]any{"type": "string"}, defs["Shared"])
	assert.Equal(t, map[string]any{"type": "boolean"}, defs["Shared_2"])
	assert.Equal(t, map[string]any{"type": "integer"}, defs["Shared_3"])
	assert.Contains(t, schemas["createThing"], `"shared":{"$ref":"#/$defs/Shared_3"}`)
}

func TestCollectCLIBodySchemas_CollisionAliasesArePerOperation(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: t
  version: 1.0.0
paths:
  /colliding:
    post:
      operationId: createColliding
      requestBody:
        content:
          application/json:
            schema:
              type: object
              $defs:
                Shared:
                  type: string
              properties:
                shared:
                  $ref: "#/components/schemas/Shared"
      responses:
        "200":
          description: ok
  /plain:
    post:
      operationId: createPlain
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                shared:
                  $ref: "#/components/schemas/Shared"
      responses:
        "200":
          description: ok
components:
  schemas:
    Shared:
      type: integer
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	assert.Contains(t, schemas["createColliding"], `"shared":{"$ref":"#/$defs/Shared_2"}`)
	assert.Contains(t, schemas["createPlain"], `"shared":{"$ref":"#/$defs/Shared"}`)
}

// contentSchema holds a subschema (3.1): its component references must be
// bundled so --schema stays self-contained.
func TestCollectCLIBodySchemas_ContentSchemaRefsAreBundled(t *testing.T) {
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
                blob:
                  type: string
                  contentMediaType: application/json
                  contentSchema:
                    $ref: "#/components/schemas/Inner"
      responses:
        "200":
          description: ok
components:
  schemas:
    Inner:
      type: integer
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	assert.Contains(t, schemas["createThing"], `"$ref":"#/$defs/Inner"`)
	assert.Contains(t, schemas["createThing"], `"Inner":{"type":"integer"}`)
}

// Integers beyond float64's exact range must survive the bundling copy
// verbatim instead of rounding.
func TestCollectCLIBodySchemas_LargeIntegersSurviveBundling(t *testing.T) {
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
                big:
                  type: integer
                  maximum: 9007199254740993
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	assert.Contains(t, schemas["createThing"], "9007199254740993")
}

// A "$ref" appearing inside example payload data is content, not a schema
// reference: it must be neither collected (a missing-component error) nor
// rewritten into the bundle.
func TestCollectCLIBodySchemas_ExampleDataRefsAreContent(t *testing.T) {
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
                payload:
                  type: object
                  example:
                    $ref: "#/components/schemas/DoesNotExist"
                    note: literal payload content
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	assert.Contains(t, schemas["createThing"], `"$ref":"#/components/schemas/DoesNotExist"`)
	assert.NotContains(t, schemas["createThing"], `#/$defs/DoesNotExist`)
}

func TestCollectCLIBodySchemas_MissingComponentCannotShadowLocalDef(t *testing.T) {
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
              $defs:
                Shared:
                  type: string
              properties:
                shared:
                  $ref: "#/components/schemas/Shared"
      responses:
        "200":
          description: ok
`
	_, err := collectBodySchemasFromYAML(t, spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "createThing")
	assert.Contains(t, err.Error(), `component "Shared"`)
	assert.Contains(t, err.Error(), "does not exist in components.schemas")
}

func TestCollectCLIBodySchemas_BooleanRequestSchema(t *testing.T) {
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
            schema: true
      responses:
        "200":
          description: ok
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)
	assert.Equal(t, "true", schemas["createThing"])
}

func TestCollectCLIBodySchemas_BundlesBooleanComponent(t *testing.T) {
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
                anything:
                  $ref: "#/components/schemas/Anything"
      responses:
        "200":
          description: ok
components:
  schemas:
    Anything: true
`
	schemas, err := collectBodySchemasFromYAML(t, spec)
	require.NoError(t, err)

	var bundled map[string]any
	require.NoError(t, json.Unmarshal([]byte(schemas["createThing"]), &bundled))
	defs, ok := bundled["$defs"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, true, defs["Anything"])
	assert.Contains(t, schemas["createThing"], `"$ref":"#/$defs/Anything"`)
}
