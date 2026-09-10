package snapshots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/stretchr/testify/require"
)

const querySSEOverloadTSSpec = `openapi: 3.1.0
info:
  title: Query SSE Test
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /interactions/{id}:
    get:
      operationId: get
      x-speakeasy-group: interactions
      x-speakeasy-sse-overload: true
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: stream
          in: query
          schema:
            type: boolean
            default: false
        - name: last_event_id
          in: query
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Interaction"
            text/event-stream:
              schema:
                $ref: "#/components/schemas/InteractionSSEEvent"
components:
  schemas:
    Interaction:
      type: object
      required: [id]
      properties:
        id:
          type: string
    InteractionSSEEvent:
      type: object
      required: [data]
      properties:
        event:
          type: string
        data:
          type: string`

func TestSnapTsExplicitQuerySSEOverload(t *testing.T) {
	t.Parallel()

	genYaml := `configVersion: 2.0.0
typescript:
  packageName: querystream
  responseFormat: flat
  templateVersion: v2
`

	tempDir, errs := generateQuerySSEOverloadTS(t, genYaml)
	require.Empty(t, errs)

	funcSource := readGeneratedQuerySSEOverloadTSFile(t, tempDir, "src", "funcs", "interactions-get.ts")

	require.Contains(t, funcSource, `request: operations.GetRequest & { stream?: false }`)
	require.Contains(t, funcSource, `request: operations.GetRequest & { stream: true }`)
	require.Contains(t, funcSource, `Accept: request?.stream ? "text/event-stream" : "application/json",`)

	methodSource := readGeneratedQuerySSEOverloadTSFile(t, tempDir, "src", "sdk", "interactions.ts")

	require.Contains(t, methodSource, `request: operations.GetRequest & { stream?: false | undefined }`)
	require.Contains(t, methodSource, `request: operations.GetRequest & { stream: true }`)
}

func TestSnapTsConfiguredSSEStreamClassName(t *testing.T) {
	t.Parallel()

	genYaml := `configVersion: 2.0.0
typescript:
  packageName: querystream
  responseFormat: flat
  templateVersion: v2
  eventStreamClassName: Stream
`

	tempDir, errs := generateQuerySSEOverloadTS(t, genYaml)
	require.Empty(t, errs)

	funcSource := readGeneratedQuerySSEOverloadTSFile(t, tempDir, "src", "funcs", "interactions-get.ts")
	require.Contains(t, funcSource, `import { Stream } from "../lib/event-streams.js";`)
	require.Contains(t, funcSource, `Stream<models.InteractionSSEEvent>`)
	require.NotContains(t, funcSource, `EventStream<models.InteractionSSEEvent>`)

	methodSource := readGeneratedQuerySSEOverloadTSFile(t, tempDir, "src", "sdk", "interactions.ts")
	require.Contains(t, methodSource, `Stream<models.InteractionSSEEvent>`)
	require.NotContains(t, methodSource, `EventStream<models.InteractionSSEEvent>`)

	operationSource := readGeneratedQuerySSEOverloadTSFile(t, tempDir, "src", "models", "operations", "get.ts")
	require.Contains(t, operationSource, `import { Stream } from "../../lib/event-streams.js";`)
	require.Contains(t, operationSource, `return new Stream(stream, rawEvent => {`)
	require.NotContains(t, operationSource, `EventStream<models.InteractionSSEEvent>`)

	eventStreamsSource := readGeneratedQuerySSEOverloadTSFile(t, tempDir, "src", "lib", "event-streams.ts")
	require.Contains(t, eventStreamsSource, `export class Stream<T extends SseMessage<unknown>>`)
	require.NotContains(t, eventStreamsSource, `export class EventStream`)
}

func TestSnapTsConfiguredSSEStreamClassNameRejectsReservedName(t *testing.T) {
	t.Parallel()

	genYaml := `configVersion: 2.0.0
typescript:
  packageName: querystream
  responseFormat: flat
  templateVersion: v2
  eventStreamClassName: ReadableStream
`

	_, errs := generateQuerySSEOverloadTS(t, genYaml)
	require.NotEmpty(t, errs)
	require.Contains(t, fmt.Sprint(errs), "collides with the reserved name")
}

func generateQuerySSEOverloadTS(t *testing.T, genYaml string) (string, []error) {
	t.Helper()

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(genYaml), 0o644))

	generator, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	return tempDir, generator.Generate(ctx, []byte(querySSEOverloadTSSpec), "test-schema.yaml", "typescript", tempDir, false, false)
}

func readGeneratedQuerySSEOverloadTSFile(t *testing.T, tempDir string, pathParts ...string) string {
	t.Helper()

	sourceBytes, err := os.ReadFile(filepath.Join(append([]string{tempDir}, pathParts...)...))
	require.NoError(t, err)
	return string(sourceBytes)
}
