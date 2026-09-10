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

const querySSEOverloadSpec = `openapi: 3.1.0
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

func TestSnapPyExplicitQuerySSEOverload(t *testing.T) {
	t.Parallel()

	genYaml := `python:
  packageName: querystream
  flattenRequests: true
  methodArguments: positional-path-params
  responseFormat: flat
`

	tempDir, errs := generateQuerySSEOverload(t, genYaml)
	require.Empty(t, errs)
	source := readGeneratedQuerySSEOverloadFile(t, tempDir, "src", "querystream", "interactions.py")

	require.Contains(t, source, "from typing import Literal, Mapping, Optional, Union, overload")
	require.NotContains(t, source, "accept_header_override")
	require.Contains(t, source, `stream: Union[Literal[False], None] = None,`)
	require.Contains(t, source, `stream: Literal[True],`)
	require.Contains(t, source, `stream: bool,`)
	require.Contains(t, source, "models.Interaction, eventstreaming.EventStream[models.InteractionSSEEvent]")
	require.Contains(t, source, `stream: Optional[bool] = False,`)
	require.Contains(t, source, `if stream is True
            else "application/json",`)
}

func TestSnapPyConfiguredSSEStreamClassNames(t *testing.T) {
	t.Parallel()

	genYaml := `python:
  packageName: querystream
  flattenRequests: true
  methodArguments: positional-path-params
  responseFormat: flat
  eventStreamClassNames:
    sync: EventStream
    async: Stream
`

	tempDir, errs := generateQuerySSEOverload(t, genYaml)
	require.Empty(t, errs)
	source := readGeneratedQuerySSEOverloadFile(t, tempDir, "src", "querystream", "interactions.py")

	require.Contains(t, source, "from querystream.utils.eventstreaming import EventStream, Stream")
	require.Contains(t, source, "models.Interaction, EventStream[models.InteractionSSEEvent]")
	require.Contains(t, source, "models.Interaction, Stream[models.InteractionSSEEvent]")
	require.Contains(t, source, "return EventStream(")
	require.Contains(t, source, "return Stream(")
	require.NotContains(t, source, "__PYTHON_EVENT_STREAM_TYPE_REF__")
	require.NotContains(t, source, "EventEventStream")
	require.NotContains(t, source, "eventstreaming.Stream[models.InteractionSSEEvent]")
	require.NotContains(t, source, "eventstreaming.EventStream[models.InteractionSSEEvent]")
	require.NotContains(t, source, "eventstreaming.EventStreamAsync[models.InteractionSSEEvent]")

	eventStreamingSource := readGeneratedQuerySSEOverloadFile(t, tempDir, "src", "querystream", "utils", "eventstreaming.py")
	require.Contains(t, eventStreamingSource, "class EventStream(Generic[T]):")
	require.Contains(t, eventStreamingSource, "class Stream(Generic[T]):")
	require.Contains(t, eventStreamingSource, "def close(self):")
	require.Contains(t, eventStreamingSource, "async def close(self):")
}

func TestSnapPyConfiguredSSEStreamClassNamesRejectPythonKeywords(t *testing.T) {
	t.Parallel()

	genYaml := `python:
  packageName: querystream
  flattenRequests: true
  methodArguments: positional-path-params
  responseFormat: flat
  eventStreamClassNames:
    sync: class
    async: Stream
`

	_, errs := generateQuerySSEOverload(t, genYaml)
	require.NotEmpty(t, errs)
	require.Contains(t, fmt.Sprint(errs), "eventStreamClassNames values must not be Python keywords")
}

func generateQuerySSEOverload(t *testing.T, genYaml string) (string, []error) {
	t.Helper()

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(genYaml), 0o644))

	generator, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	return tempDir, generator.Generate(ctx, []byte(querySSEOverloadSpec), "test-schema.yaml", "python", tempDir, false, false)
}

func readGeneratedQuerySSEOverloadFile(t *testing.T, tempDir string, pathParts ...string) string {
	t.Helper()

	sourceBytes, err := os.ReadFile(filepath.Join(append([]string{tempDir}, pathParts...)...))
	require.NoError(t, err)
	return string(sourceBytes)
}
