package snapshots

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/stretchr/testify/require"
)

func TestSnapPyRawResponseHelpers(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Raw Response Helper Test
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
  /agents:
    post:
      operationId: createAgent
      x-speakeasy-group: agents
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Agent"
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
          type: string
    Agent:
      type: object
      required: [id]
      properties:
        id:
          type: string`

	genYaml := `python:
  packageName: RawHelpers
  flattenRequests: true
  methodArguments: positional-path-with-extras
  responseFormat: flat
  rawResponseHelpers: true
`

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(genYaml), 0o644))

	generator, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	errs := generator.Generate(ctx, []byte(spec), "test-schema.yaml", "python", tempDir, false, false)
	require.Empty(t, errs)

	interactionsBytes, err := os.ReadFile(filepath.Join(tempDir, "src", "rawhelpers", "interactions.py"))
	require.NoError(t, err)
	interactions := string(interactionsBytes)

	require.Contains(t, interactions, "response_helpers")
	require.Contains(t, interactions, "def with_raw_response(self):")
	require.Contains(t, interactions, "def with_streaming_response(self):")
	// Response mode is derived from a stripped request header, not a public method argument.
	require.NotContains(t, interactions, `_speakeasy_response_mode: Literal`)
	require.Contains(t, interactions, `response_helpers.consume_response_mode(`)
	require.NotContains(t, interactions, `is_error_status_code=(lambda c: False)`)
	require.Contains(t, interactions, `if _speakeasy_response_mode != "parsed"`)
	require.Contains(t, interactions, `stream=stream is True or _speakeasy_response_mode == "streaming",`)
	require.Contains(t, interactions, `mode=("event_stream" if stream is True else "buffered"),`)
	require.Contains(t, interactions, `return cast(`)
	require.Contains(t, interactions, `_speakeasy_response_cls = (`)
	require.Contains(t, interactions, `response_helpers.StreamedAPIResponse`)
	require.Contains(t, interactions, `else response_helpers.APIResponse`)
	require.NotContains(t, interactions, `stream=True,`)
	require.NotContains(t, interactions, `stream=True or _speakeasy_response_mode == "streaming"`)
	require.Contains(t, interactions, `self.get = response_helpers.to_raw_response_wrapper(sdk.get, "extra_headers")`)
	require.Contains(t, interactions, "self.get = response_helpers.to_streamed_response_wrapper(\n            sdk.get, \"extra_headers\"\n        )")
	require.NotContains(t, interactions, `response_helpers.APIResponse[Any], self._sdk.get(*args, **kwargs)`)
	require.NotContains(t, interactions, "x-speakeasy-response-mode")

	agentsBytes, err := os.ReadFile(filepath.Join(tempDir, "src", "rawhelpers", "agents.py"))
	require.NoError(t, err)
	agents := string(agentsBytes)
	require.Contains(t, agents, `) -> models.Agent:`)
	require.NotContains(t, agents, `) -> Any:`)
	require.Contains(t, agents, `stream=_speakeasy_response_mode == "streaming"`)
	require.Contains(t, agents, `mode="buffered",`)
	require.NotContains(t, agents, `stream=True or _speakeasy_response_mode == "streaming"`)

	responseHelpersBytes, err := os.ReadFile(filepath.Join(tempDir, "src", "rawhelpers", "utils", "response_helpers.py"))
	require.NoError(t, err)
	responseHelpers := string(responseHelpersBytes)
	versionBytes, err := os.ReadFile(filepath.Join(tempDir, "src", "rawhelpers", "_version.py"))
	require.NoError(t, err)
	version := string(versionBytes)
	require.Contains(t, responseHelpers, "class APIResponse")
	require.Contains(t, responseHelpers, "class StreamedAPIResponse")
	require.Contains(t, responseHelpers, "class AsyncStreamedAPIResponse")
	require.Contains(t, responseHelpers, "class ResponseContextManager")
	require.Contains(t, responseHelpers, "def parse(self, *, to: Type[U]) -> U")
	require.Contains(t, responseHelpers, "def parse(self) -> T")
	require.Contains(t, responseHelpers, "self._parsed_by_type: dict[Any, Any] = {}")
	require.Contains(t, responseHelpers, "parse_key = to or _DEFAULT_PARSE_KEY")
	require.Contains(t, responseHelpers, "unmarshal_json_response(")
	require.Contains(t, responseHelpers, "def request_id(self)")
	require.Contains(t, responseHelpers, "def elapsed(self) -> timedelta")
	require.Contains(t, responseHelpers, "raw_stream / event_stream wrappers leave the body unread")
	require.Contains(t, responseHelpers, "def __enter__(self: ResponseT) -> ResponseT")
	require.Contains(t, responseHelpers, "async def __aenter__(self: AsyncResponseT) -> AsyncResponseT")
	require.Contains(t, responseHelpers, "def to_raw_response_wrapper")
	require.Contains(t, responseHelpers, "def to_streamed_response_wrapper")
	require.Contains(t, responseHelpers, "Callable[P, APIResponse[T]]")
	require.Contains(t, responseHelpers, "Callable[P, ResponseContextManager[StreamedAPIResponse[T]]]")
	require.Contains(t, responseHelpers, "def consume_response_mode(")
	require.Contains(t, responseHelpers, "cast(Dict[str, Any], kwargs), _RAW_RESPONSE_MODE, header_kwarg")
	require.Contains(t, responseHelpers, `cast(Dict[str, Any], kwargs), _STREAMING_RESPONSE_MODE, header_kwarg`)
	require.Contains(t, responseHelpers, "functools.partial(func, *args, **kwargs)")
	// Marker header is whitelabeled with the SDK package name and normalized at
	// generation time, never "speakeasy".
	require.NotContains(t, responseHelpers, "from .._version import __title__")
	require.Contains(t, responseHelpers, "from .._version import __response_mode_header__")
	require.Contains(t, responseHelpers, `_RESPONSE_MODE_HEADER = __response_mode_header__.lower()`)
	require.Contains(t, version, `__title__: str = "RawHelpers"`)
	require.Contains(t, version, `__response_mode_header__: str = "x-rawhelpers-response-mode"`)
	require.NotContains(t, responseHelpers, "x-speakeasy")
}

func TestSnapPyRawResponseHelpersDisabledDoesNotEmitResponseModeHeader(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Raw Response Helper Disabled Test
  version: 1.0.0
servers:
  - url: https://api.example.com
paths:
  /agents:
    post:
      operationId: createAgent
      x-speakeasy-group: agents
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Agent"
components:
  schemas:
    Agent:
      type: object
      required: [id]
      properties:
        id:
          type: string`

	genYaml := `python:
  packageName: RawHelpersDisabled
  flattenRequests: true
  responseFormat: flat
  rawResponseHelpers: false
`

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(genYaml), 0o644))

	generator, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	errs := generator.Generate(ctx, []byte(spec), "test-schema.yaml", "python", tempDir, false, false)
	require.Empty(t, errs)

	versionBytes, err := os.ReadFile(filepath.Join(tempDir, "src", "rawhelpersdisabled", "_version.py"))
	require.NoError(t, err)
	version := string(versionBytes)
	require.Contains(t, version, `__title__: str = "RawHelpersDisabled"`)
	require.NotContains(t, version, "__response_mode_header__")
	require.NotContains(t, version, "response-mode")

	_, err = os.Stat(filepath.Join(tempDir, "src", "rawhelpersdisabled", "utils", "response_helpers.py"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
