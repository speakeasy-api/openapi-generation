package snapshots

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/stretchr/testify/require"
)

func TestSnapPyPositionalOperationArguments(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Positional Args Test
  version: 1.0.0
servers:
  - url: https://api.example.com
x-speakeasy-globals:
  parameters:
    - $ref: '#/components/parameters/api_version'
paths:
  /widgets:
    post:
      operationId: createWidget
      x-speakeasy-group: widgets
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
                color:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Widget'
  /{api_version}/widgets:
    parameters:
      - $ref: '#/components/parameters/api_version'
    get:
      operationId: listWidgets
      x-speakeasy-group: widgets
      parameters:
        - name: cursor
          in: query
          required: false
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListWidgetsResponse'
      x-speakeasy-pagination:
        type: cursor
        inputs:
          - name: cursor
            in: parameters
            type: cursor
        outputs:
          results: $.items
          nextCursor: $.next_cursor
  /{api_version}/widgets/{id}:
    parameters:
      - $ref: '#/components/parameters/api_version'
      - name: id
        in: path
        required: true
        schema:
          type: string
    get:
      operationId: getWidget
      x-speakeasy-group: widgets
      parameters:
        - name: expand
          in: query
          required: false
          schema:
            type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Widget'
    patch:
      operationId: updateWidget
      x-speakeasy-group: widgets
      parameters:
        - name: update_mask
          in: query
          required: false
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Widget'
components:
  parameters:
    api_version:
      name: api_version
      in: path
      required: true
      x-speakeasy-globals-hidden: true
      schema:
        type: string
  schemas:
    ListWidgetsResponse:
      type: object
      required: [items]
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/Widget'
        next_cursor:
          type: string
    Widget:
      type: object
      required: [id, name]
      properties:
        id:
          type: string
        name:
          type: string`

	genYaml := `python:
  packageName: posargs
  flattenRequests: true
  methodArguments: positional-path-params
  imports:
    option: openapi
    paths:
      callbacks: ""
      errors: ""
      operations: ""
      shared: ""
      webhooks: ""
`

	tempDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, ".speakeasy", "gen.yaml"), []byte(genYaml), 0o644))

	generator, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	errs := generator.Generate(ctx, []byte(spec), "test-schema.yaml", "python", tempDir, false, false)
	require.Empty(t, errs)

	widgets, err := os.ReadFile(filepath.Join(tempDir, "src", "posargs", "widgets.py"))
	require.NoError(t, err)
	text := string(widgets)
	baseSDK, err := os.ReadFile(filepath.Join(tempDir, "src", "posargs", "basesdk.py"))
	require.NoError(t, err)
	baseSDKText := string(baseSDK)
	requestBodies, err := os.ReadFile(filepath.Join(tempDir, "src", "posargs", "utils", "requestbodies.py"))
	require.NoError(t, err)
	requestBodiesText := string(requestBodies)

	require.Contains(t, text, `    def create_widget(
        self,
        *,
        name: str,`)
	require.Contains(t, text, `    def get_widget(
        self,
        id: str,
        *,
        expand: Optional[str] = None,`)
	require.Contains(t, text, `    def update_widget(
        self,
        id: str,
        *,
        update_mask: Optional[str] = None,
        name: Optional[str] = None,`)
	require.NotContains(t, text, "\n        api_version: str,\n")
	require.NotContains(t, text, "extra_body")
	require.NotContains(t, text, "extra_query_params")
	require.NotContains(t, text, "_coerce_timeout_ms")
	require.NotContains(t, baseSDKText, "extra_query_params")
	require.NotContains(t, baseSDKText, "_coerce_timeout_ms")
	require.NotContains(t, requestBodiesText, "extra_body")

	genYamlWithExtras := `python:
  packageName: posargs
  flattenRequests: true
  methodArguments: positional-path-with-extras
  imports:
    option: openapi
    paths:
      callbacks: ""
      errors: ""
      operations: ""
      shared: ""
      webhooks: ""
`

	tempDirWithExtras := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDirWithExtras, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDirWithExtras, ".speakeasy", "gen.yaml"), []byte(genYamlWithExtras), 0o644))

	generatorWithExtras, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	errs = generatorWithExtras.Generate(ctx, []byte(spec), "test-schema.yaml", "python", tempDirWithExtras, false, false)
	require.Empty(t, errs)

	widgetsWithExtras, err := os.ReadFile(filepath.Join(tempDirWithExtras, "src", "posargs", "widgets.py"))
	require.NoError(t, err)
	textWithExtras := string(widgetsWithExtras)

	getWidgetWithExtras := pythonMethodBlock(t, textWithExtras, "def get_widget(")
	createWidgetWithExtras := pythonMethodBlock(t, textWithExtras, "def create_widget(")
	updateWidgetWithExtras := pythonMethodBlock(t, textWithExtras, "def update_widget(")
	listWidgetsWithExtras := pythonMethodBlock(t, textWithExtras, "def list_widgets(")
	require.Contains(t, getWidgetWithExtras, `extra_headers: Optional[Mapping[str, str]] = None,`)
	require.Contains(t, getWidgetWithExtras, `extra_query: Optional[Mapping[str, Any]] = None,`)
	require.Contains(t, getWidgetWithExtras, `timeout_ms: Optional[int] = None,`)
	require.NotContains(t, getWidgetWithExtras, "extra_body")
	require.Contains(t, createWidgetWithExtras, `extra_body: Optional[Mapping[str, Any]] = None,`)
	require.Contains(t, createWidgetWithExtras, `extra_body=extra_body,`)
	require.Contains(t, updateWidgetWithExtras, `extra_body: Optional[Mapping[str, Any]] = None,`)
	require.Contains(t, updateWidgetWithExtras, `extra_body=extra_body,`)
	require.Contains(t, listWidgetsWithExtras, `extra_headers=extra_headers,`)
	require.Contains(t, listWidgetsWithExtras, `extra_query=extra_query,`)
	require.NotContains(t, listWidgetsWithExtras, "extra_body")
	require.Contains(t, textWithExtras, `        timeout_ms: Optional[int] = None,`)
	require.Contains(t, textWithExtras, `:param timeout_ms: Override the default request timeout configuration for this method in milliseconds`)
	require.Contains(t, textWithExtras, `        http_headers = extra_headers`)
	require.Contains(t, textWithExtras, `            extra_query_params=extra_query,`)
	require.Contains(t, textWithExtras, `                extra_body=extra_body,`)
	require.Contains(t, textWithExtras, `                timeout_ms=timeout_ms,`)
	require.Contains(t, textWithExtras, `                extra_headers=extra_headers,`)
	require.Contains(t, textWithExtras, `                extra_query=extra_query,`)
	require.NotContains(t, textWithExtras, "timeout: Optional[Union[float, httpx.Timeout]] = None")
	require.NotContains(t, textWithExtras, "\n        http_headers: Optional[Mapping[str, str]] = None,\n")
	require.NotContains(t, textWithExtras, "\n                http_headers=http_headers,\n")
	require.NotContains(t, textWithExtras, "\n                server_url=server_url,\n")
	require.NotContains(t, textWithExtras, "\n        api_version: str,\n")

	genYamlWithExtrasAndTimeout := `python:
  packageName: posargs
  flattenRequests: true
  methodArguments: positional-path-with-extras
  methodTimeoutArgument: timeout
  methodTimeoutUnits: seconds
  imports:
    option: openapi
    paths:
      callbacks: ""
      errors: ""
      operations: ""
      shared: ""
      webhooks: ""
`

	tempDirWithExtrasAndTimeout := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tempDirWithExtrasAndTimeout, ".speakeasy"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDirWithExtrasAndTimeout, ".speakeasy", "gen.yaml"), []byte(genYamlWithExtrasAndTimeout), 0o644))

	generatorWithExtrasAndTimeout, err := generate.New(generate.WithDebuggingEnabled())
	require.NoError(t, err)

	errs = generatorWithExtrasAndTimeout.Generate(ctx, []byte(spec), "test-schema.yaml", "python", tempDirWithExtrasAndTimeout, false, false)
	require.Empty(t, errs)

	widgetsWithExtrasAndTimeout, err := os.ReadFile(filepath.Join(tempDirWithExtrasAndTimeout, "src", "posargs", "widgets.py"))
	require.NoError(t, err)
	textWithExtrasAndTimeout := string(widgetsWithExtrasAndTimeout)

	getWidgetWithExtrasAndTimeout := pythonMethodBlock(t, textWithExtrasAndTimeout, "def get_widget(")
	require.Contains(t, getWidgetWithExtrasAndTimeout, `extra_headers: Optional[Mapping[str, str]] = None,`)
	require.Contains(t, getWidgetWithExtrasAndTimeout, `extra_query: Optional[Mapping[str, Any]] = None,`)
	require.Contains(t, getWidgetWithExtrasAndTimeout, `timeout: Optional[Union[float, httpx.Timeout]] = None,`)
	require.NotContains(t, getWidgetWithExtrasAndTimeout, "extra_body")
	require.Contains(t, textWithExtrasAndTimeout, `        timeout_ms = self._coerce_timeout_ms(timeout)`)
	require.Contains(t, textWithExtrasAndTimeout, `:param timeout: Override the default request timeout configuration for this method in seconds`)
	require.Contains(t, textWithExtrasAndTimeout, `                timeout=timeout,`)
	require.NotContains(t, textWithExtrasAndTimeout, "\n        timeout_ms: Optional[int] = None,\n")
	require.NotContains(t, textWithExtrasAndTimeout, "\n                timeout_ms=timeout_ms,\n")
	require.NotContains(t, textWithExtrasAndTimeout, "pyright: ignore[reportOverlappingOverload]")
}

func pythonMethodBlock(t *testing.T, text, signature string) string {
	t.Helper()

	start := strings.Index(text, "    "+signature)
	require.NotEqual(t, -1, start)

	end := len(text)
	for _, marker := range []string{"\n    def ", "\n    async def "} {
		if idx := strings.Index(text[start+1:], marker); idx != -1 && start+1+idx < end {
			end = start + 1 + idx
		}
	}

	return text[start:end]
}
