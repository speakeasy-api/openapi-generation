package extensions

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	upstreamjq "github.com/itchyny/gojq"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cliProjectionEdgeSpecPrefix = `openapi: 3.1.0
info:
  title: Projection Service
  version: 1.0.0
paths:
  /value:
    get:
      operationId: GetValue
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
`

const cliProjectionTestSpec = `openapi: 3.1.0
info:
  title: Projection Service
  version: 1.0.0
paths:
  /result:
    get:
      operationId: GetResult
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TaskResult'
  /events:
    get:
      operationId: StreamEvents
      responses:
        "200":
          description: events
          content:
            text/event-stream:
              schema:
                oneOf:
                  - $ref: '#/components/schemas/DeltaEvent'
                  - $ref: '#/components/schemas/DoneEvent'
  /mixed-events:
    get:
      operationId: StreamMixedEvents
      responses:
        "200":
          description: events with one unconstrained media type
          content:
            text/event-stream:
              schema:
                type: object
                properties:
                  text:
                    type: string
            application/x-ndjson: {}
components:
  schemas:
    TaskResult:
      type: object
      properties:
        id:
          type: string
        status:
          type: string
          enum: [queued, running, done]
        output_text:
          type: string
        count:
          type: integer
        usage:
          $ref: '#/components/schemas/UsageStats'
        items:
          type: array
          items:
            $ref: '#/components/schemas/ResultItem'
        embedding:
          $ref: '#/components/schemas/Embedding'
        response_format:
          oneOf:
            - $ref: '#/components/schemas/TextFormat'
            - $ref: '#/components/schemas/SchemaFormat'
            - {}
        tree:
          $ref: '#/components/schemas/TreeNode'
      required: [id, status]
    UsageStats:
      type: object
      properties:
        total_units:
          type: integer
        breakdown:
          $ref: '#/components/schemas/UsageBreakdown'
    UsageBreakdown:
      type: object
      properties:
        engine_units:
          type: integer
        queue_units:
          type: integer
    ResultItem:
      type: object
      properties:
        name:
          type: string
        score:
          type: number
    Embedding:
      type: object
      properties:
        values:
          type: array
          items:
            type: number
    TextFormat:
      type: object
      properties:
        type:
          type: string
    SchemaFormat:
      type: object
      properties:
        json_schema:
          type: object
    TreeNode:
      type: object
      properties:
        label:
          type: string
        child:
          $ref: '#/components/schemas/TreeNode'
    DeltaEvent:
      type: object
      properties:
        data:
          type: object
          properties:
            delta:
              type: object
              properties:
                text:
                  type: string
            empty:
              type: "null"
    DoneEvent:
      type: object
      properties:
        data:
          type: object
          properties:
            result:
              type: string
            empty:
              type: "null"
`

type cliProjectionExpectation struct {
	name     string
	jq       string
	severity string
	contains []string
}

func TestCLISingularPathToJQLowering(t *testing.T) {
	for path, want := range map[string]string{
		"$.data.delta.text": ".data.delta.text",
		"$.items[0].x":      ".items[0].x",
		"$['body/text']":    `.["body/text"]`,
	} {
		segments, err := cliParseSingularPath(path)
		require.NoError(t, err)
		assert.Equal(t, want, cliSingularPathToJQ(segments), path)
	}
}

func TestCLIProjectionLintRuntimeParserAuthority(t *testing.T) {
	t.Run("runtime accepts syntax unsupported by analyzer", func(t *testing.T) {
		const expression = `.known + "!" as $value | $value`
		query, err := upstreamjq.Parse(expression)
		require.NoError(t, err)
		value, ok := query.Run(map[string]any{"known": "ok"}).Next()
		require.True(t, ok)
		assert.Equal(t, "ok!", value)

		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  known:
                    type: string
`
		warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: '.known + "!" as $value | $value'
`)
		require.NoError(t, err)
		require.Equal(t, []string{`command "show-value": jq projection ".known + \"!\" as $value | $value" cannot be statically verified — the analyzer does not support this syntax; it will only be checked at runtime`}, warnings)
	})

	t.Run("runtime parse failure is an error", func(t *testing.T) {
		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  known:
                    type: string
`
		warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: '.known |'
`)
		require.EqualError(t, err, `command "show-value": jq projection ".known |" is not valid jq: unexpected EOF`)
		assert.Empty(t, warnings)
	})
}

func TestCLIProjectionLintOpenObjectProperties(t *testing.T) {
	t.Run("additional properties", func(t *testing.T) {
		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  known:
                    type: string
                additionalProperties: true
`
		warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .dynamic
`)
		require.NoError(t, err)
		require.Equal(t, []string{`command "show-value": jq projection ".dynamic" cannot be statically verified — unknown (unconstrained) schema at $; it will only be checked at runtime`}, warnings)
	})

	t.Run("pattern property", func(t *testing.T) {
		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  known:
                    type: string
                patternProperties:
                  '^x_[a-z]+$':
                    type: string
                additionalProperties: false
`
		warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .x_feature
`)
		require.NoError(t, err)
		assert.Empty(t, warnings)
	})

	t.Run("unevaluated properties", func(t *testing.T) {
		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  known:
                    type: string
                unevaluatedProperties: true
`
		warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .dynamic
`)
		require.NoError(t, err)
		require.Equal(t, []string{`command "show-value": jq projection ".dynamic" cannot be statically verified — the response schema may permit undeclared properties along this path; it will only be checked at runtime`}, warnings)
	})

	t.Run("open map item", func(t *testing.T) {
		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  items:
                    type: array
                    items:
                      type: object
                      properties:
                        known:
                          type: string
                      additionalProperties: true
`
		_, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: '.items | map(.dynamic)'
`)
		require.NoError(t, err)
	})

	t.Run("unevaluated map item", func(t *testing.T) {
		spec := cliProjectionEdgeSpecPrefix + `                type: object
                properties:
                  items:
                    type: array
                    items:
                      type: object
                      properties:
                        known:
                          type: string
                      unevaluatedProperties: true
`
		_, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: '.items | map(.dynamic)'
`)
		require.NoError(t, err)
	})
}

func TestCLIProjectionLintObjectArrayTypeSet(t *testing.T) {
	spec := cliProjectionEdgeSpecPrefix + `                type: [object, array]
                properties:
                  name:
                    type: string
                items:
                  type: string
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .name
`)
	require.NoError(t, err)
	require.Equal(t, []string{`command "show-value": jq projection ".name" cannot be statically verified — property access on non-object type at $; it will only be checked at runtime`}, warnings)
}

func TestCLIProjectionLintIterableTypeSet(t *testing.T) {
	spec := cliProjectionEdgeSpecPrefix + `                type: [array, string]
                items:
                  type: string
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: '.[]'
`)
	require.NoError(t, err)
	require.Equal(t, []string{`command "show-value": jq projection ".[]" cannot be statically verified — iteration over unknown type at $; it will only be checked at runtime`}, warnings)
}

func TestCLIProjectionLintReferenceSiblings(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Projection Service
  version: 1.0.0
paths:
  /value:
    get:
      operationId: GetValue
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BaseValue'
                type: object
                properties:
                  added:
                    type: string
                required: [added]
components:
  schemas:
    BaseValue:
      type: object
      properties:
        base:
          type: string
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .added
`)
	require.NoError(t, err)
	assert.Empty(t, warnings)
}

func TestCLIProjectionLintNestedStreamReferenceIsUnverifiable(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Streaming Projection Service
  version: 1.0.0
paths:
  /events:
    get:
      operationId: StreamValues
      responses:
        "200":
          description: ok
          content:
            text/event-stream:
              schema:
                $ref: '#/components/schemas/Envelope/properties/event'
components:
  schemas:
    Envelope:
      type: object
      properties:
        event:
          type: object
          properties:
            text:
              type: string
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  stream-values:
    op: StreamValues
    output:
      stream:
        select: $.text
`)
	require.NoError(t, err)
	require.Equal(t, []string{`command "stream-values" output.stream.select $.text could not be verified against the streaming response schema (reference "#/components/schemas/Envelope/properties/event" is not a top-level component schema reference); events without a string at that path are skipped at runtime`}, warnings)
}

func TestCLIProjectionLintOpenStreamSelect(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Streaming Projection Service
  version: 1.0.0
paths:
  /events:
    get:
      operationId: StreamValues
      responses:
        "200":
          description: ok
          content:
            text/event-stream:
              schema:
                type: object
                properties:
                  known:
                    type: string
                additionalProperties: true
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  stream-values:
    op: StreamValues
    output:
      stream:
        select: $.dynamic
`)
	require.NoError(t, err)
	require.Equal(t, []string{`command "stream-values" output.stream.select $.dynamic could not be verified against the streaming response schema ("dynamic" is not a declared property and the schema accepts unknown keys); events without a string at that path are skipped at runtime`}, warnings)
}

func TestCLIProjectionLintMissingJSONResponseSchema(t *testing.T) {
	tests := map[string]string{
		"non-JSON media": `      responses:
        "200":
          description: ok
          content:
            text/plain:
              schema:
                type: string
`,
		"missing responses": `      responses: {}
`,
		"default response only": `      responses:
        default:
          description: fallback
          content:
            application/json:
              schema:
                type: object
`,
		"boolean schema": `      responses:
        "200":
          description: ok
          content:
            application/json:
              schema: true
`,
	}
	for name, responses := range tests {
		t.Run(name, func(t *testing.T) {
			spec := `openapi: 3.1.0
info:
  title: Projection Service
  version: 1.0.0
paths:
  /value:
    get:
      operationId: GetValue
` + responses
			warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .dynamic
`)
			require.NoError(t, err)
			require.Equal(t, []string{`command "show-value": jq projection ".dynamic" cannot be statically verified — operation "GetValue" declares no 2xx application/json response schema; it will only be checked at runtime`}, warnings)
		})
	}
}

func TestCLIProjectionLintUnresolvedStreamSelectHasOneDiagnostic(t *testing.T) {
	spec := `openapi: 3.1.0
info:
  title: Streaming Projection Service
  version: 1.0.0
paths:
  /events:
    get:
      operationId: StreamValues
      responses:
        "200":
          description: ok
          content:
            text/event-stream:
              schema:
                $defs:
                  Event:
                    type: object
                    properties:
                      known:
                        type: string
                    additionalProperties: false
                $ref: '#/paths/~1events/get/responses/200/content/text~1event-stream/schema/$defs/Event'
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  stream-values:
    op: StreamValues
    output:
      stream:
        select: $.missing
`)
	require.NoError(t, err)
	require.Equal(t, []string{`command "stream-values" output.stream.select $.missing could not be verified against the streaming response schema (reference "#/paths/~1events/get/responses/200/content/text~1event-stream/schema/$defs/Event" is not a top-level component schema reference); events without a string at that path are skipped at runtime`}, warnings)
}

func TestCLIProjectionLintRootArraySuggestion(t *testing.T) {
	spec := cliProjectionEdgeSpecPrefix + `                type: array
                items:
                  type: object
                  properties:
                    name:
                      type: string
`
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  show-value:
    op: GetValue
    jq: .name
`)
	require.EqualError(t, err, `command "show-value": jq projection ".name" always returns nothing — the response is an array; use ".[].name" or "map(.name)"`)
	assert.Empty(t, warnings)
}

func TestCLIProjectionLintMatrix(t *testing.T) {
	cases := []cliProjectionExpectation{
		{name: "good-simple", jq: ".usage.total_units"},
		{name: "good-nested", jq: ".usage.breakdown.engine_units"},
		{name: "good-map", jq: ".items | map(.name)"},
		{name: "good-iter", jq: ".items[].name"},
		{name: "good-embed", jq: ".embedding.values"},
		{name: "typo-leaf", jq: ".embedding.value", severity: "error", contains: []string{
			`command "typo-leaf": jq projection ".embedding.value" always returns nothing`, `".value" does not exist`, `did you mean ".values"?`,
		}},
		{name: "typo-branch", jq: ".embeding.values", severity: "error", contains: []string{
			`command "typo-branch": jq projection ".embeding.values"`, `".embeding" does not exist`, `did you mean ".embedding"?`,
		}},
		{name: "wrong-container", jq: ".items.name", severity: "error", contains: []string{
			`command "wrong-container"`, `".items" is an array`, `".items[].name"`, `".items | map(.name)"`,
		}},
		{name: "iterate-scalar", jq: ".count[]", severity: "error", contains: []string{
			`command "iterate-scalar"`, `".count" is an integer`, "[] iterates arrays",
		}},
		{name: "oneof-arm", jq: ".response_format.json_schema", severity: "warning", contains: []string{
			`command "oneof-arm": jq projection ".response_format.json_schema"`, "cannot be statically verified", "property access on non-object type at $", "runtime",
		}},
		{name: "drift-renamed", jq: ".token_usage.total", severity: "error", contains: []string{
			`command "drift-renamed"`, `".token_usage" does not exist`,
		}},
		{name: "map-missing", jq: ".items | map(.scor)", severity: "error", contains: []string{
			`command "map-missing"`, "always returns an array of nulls", `".scor" does not exist`, `did you mean ".score"?`,
		}},
		{name: "builtin-mismatch", jq: ".usage | ascii_downcase", severity: "error", contains: []string{
			`command "builtin-mismatch"`, `applies the string operation "ascii_downcase"`, `".usage"`, "an object",
		}},
		{name: "syntax-error", jq: ".items | map(.name", severity: "error", contains: []string{
			`command "syntax-error": jq projection ".items | map(.name" is not valid jq`,
		}},
	}

	restore := cliCaptureProjectionStdio(t)
	results := make(map[string]struct {
		warnings []string
		err      error
	}, len(cases))
	for _, test := range cases {
		warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, fmt.Sprintf("version: 1\ncommands:\n  %s:\n    op: GetResult\n    jq: %s\n", test.name, strconv.Quote(test.jq)))
		results[test.name] = struct {
			warnings []string
			err      error
		}{warnings: warnings, err: err}
	}
	stdout, stderr := restore()
	assert.Empty(t, stdout, "projection analysis must be silent on stdout")
	assert.Empty(t, stderr, "projection analysis must be silent on stderr")

	for _, test := range cases {
		result := results[test.name]
		switch test.severity {
		case "":
			require.NoError(t, result.err, test.name)
			assert.Empty(t, result.warnings, test.name)
		case "error":
			if assert.Error(t, result.err, test.name) {
				for _, substring := range test.contains {
					assert.Contains(t, result.err.Error(), substring, test.name)
				}
			}
			assert.Empty(t, result.warnings, test.name)
		case "warning":
			require.NoError(t, result.err, test.name)
			if assert.Len(t, result.warnings, 1, test.name) {
				for _, substring := range test.contains {
					assert.Contains(t, result.warnings[0], substring, test.name)
				}
			}
		}
	}
}

func TestCLIProjectionLintBrokenIntent(t *testing.T) {
	warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-result:
    op: GetResult
    jq: .embedding.value
`)
	require.EqualError(t, err, `command "inspect-result": jq projection ".embedding.value" always returns nothing — ".value" does not exist in ".embedding" (did you mean ".values"?)`)
	assert.Empty(t, warnings)
}

func TestCLIProjectionLintFormerFixturePathIsBroken(t *testing.T) {
	_, warnings, err := decodeCLITest(t, `
version: 1
commands:
  image:
    op: CreateTask
    jq: .result.content
`)
	require.EqualError(t, err, `command "image": jq projection ".result.content" always returns nothing — ".result" does not exist in the response`)
	assert.Empty(t, warnings)
}

func TestCLIProjectionLintSoundIntent(t *testing.T) {
	warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-result:
    op: GetResult
    jq: .usage.breakdown.engine_units
`)
	require.NoError(t, err)
	assert.Empty(t, warnings)
}

func TestCLIProjectionLintUnverifiableIntent(t *testing.T) {
	warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-format:
    op: GetResult
    jq: .response_format.json_schema
`)
	require.NoError(t, err)
	require.Equal(t, []string{`command "inspect-format": jq projection ".response_format.json_schema" cannot be statically verified — property access on non-object type at $; it will only be checked at runtime`}, warnings)
}

func TestCLIProjectionLintAnalyzerFailureWarns(t *testing.T) {
	warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-result:
    op: GetResult
    jq: projection_function_not_defined
`)
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], `command "inspect-result": jq projection "projection_function_not_defined" cannot be statically verified — could not be analyzed:`)
	assert.Contains(t, warnings[0], "it will only be checked at runtime")
}

func TestCLIProjectionLintResolutionFailureWarns(t *testing.T) {
	spec := strings.Replace(cliProjectionTestSpec, "#/components/schemas/TaskResult", "#/components/schemas/MissingResult", 1)
	warnings, err := decodeCLIProjectionTest(t, spec, `version: 1
commands:
  inspect-result:
    op: GetResult
    jq: .id
`)
	require.NoError(t, err)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], `command "inspect-result": jq projection ".id" cannot be statically verified — response schema references could not be resolved:`)
	assert.Contains(t, warnings[0], "it will only be checked at runtime")
}

func TestCLIProjectionLintStreaming(t *testing.T) {
	t.Run("intent uses event union", func(t *testing.T) {
		warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  show-delta:
    op: StreamEvents
    jq: .data.delta.text
`)
		require.NoError(t, err)
		assert.Empty(t, warnings)
	})

	t.Run("intent typo is broken", func(t *testing.T) {
		_, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  show-delta:
    op: StreamEvents
    jq: .data.dleta.text
`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `command "show-delta": jq projection ".data.dleta.text" always returns nothing`)
	})

	t.Run("select typo keeps walker error", func(t *testing.T) {
		_, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  show-delta:
    op: StreamEvents
    output:
      stream: {select: $.data.delta.txt}
`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `output.stream.select $.data.delta.txt does not resolve in any event shape`)
		assert.Contains(t, err.Error(), `did you mean "text"?`)
	})

	t.Run("resolved select is escalated when proven broken", func(t *testing.T) {
		warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  show-empty:
    op: StreamEvents
    output:
      stream: {select: $.data.empty}
`)
		require.Error(t, err)
		assert.Empty(t, warnings)
		assert.Contains(t, err.Error(), `command "show-empty": output.stream.select projection ".data.empty" always returns nothing`)
	})

	t.Run("walker warning remains single", func(t *testing.T) {
		warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
operations:
  StreamMixedEvents:
    output:
      stream: {select: $.text}
`)
		require.NoError(t, err)
		require.Len(t, warnings, 1)
		assert.Contains(t, warnings[0], `operation "StreamMixedEvents" output.stream.select $.text resolves in 1 event shape(s) but could not be verified in every arm`)
	})
}

func TestCLIProjectionLintReferencesAndRecursion(t *testing.T) {
	t.Run("nested ref typo", func(t *testing.T) {
		_, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-usage:
    op: GetResult
    jq: .usage.breakdown.engine_unots
`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `".engine_unots" does not exist in ".usage.breakdown"`)
		assert.Contains(t, err.Error(), `did you mean ".engine_units"?`)
	})

	t.Run("recursive ref terminates and resolves", func(t *testing.T) {
		warnings, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-tree:
    op: GetResult
    jq: .tree.child.label
`)
		require.NoError(t, err)
		assert.Empty(t, warnings)
	})

	t.Run("typo below recursive ref", func(t *testing.T) {
		_, err := decodeCLIProjectionTest(t, cliProjectionTestSpec, `version: 1
commands:
  inspect-tree:
    op: GetResult
    jq: .tree.child.lable
`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `".lable" does not exist in ".tree.child"`)
		assert.Contains(t, err.Error(), `did you mean ".label"?`)
	})
}

func decodeCLIProjectionTest(t *testing.T, spec, manifestYAML string) ([]string, error) {
	t.Helper()
	var indented strings.Builder
	indented.WriteString("x-speakeasy-cli-commands:\n")
	for _, line := range strings.Split(strings.TrimRight(manifestYAML, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			indented.WriteString("\n")
			continue
		}
		indented.WriteString("  " + line + "\n")
	}
	fullYAML := spec + indented.String()
	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(fullYAML)))
	require.NoError(t, err)
	docInfo := &document.DocumentInfo{Doc: doc, Schema: []byte(fullYAML), SchemaPath: "projection-test.yaml"}
	for name, node := range doc.GetExtensions().All() {
		if name == "x-speakeasy-cli-commands" {
			_, warnings, err := DecodeCLICommandsManifest(ctx, docInfo, node)
			return warnings, err
		}
	}
	t.Fatalf("extension node not found in projection test document")
	return nil, nil
}

func cliCaptureProjectionStdio(t *testing.T) func() (string, string) {
	t.Helper()
	originalOut, originalErr := os.Stdout, os.Stderr
	outReader, outWriter, err := os.Pipe()
	require.NoError(t, err)
	errReader, errWriter, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout, os.Stderr = outWriter, errWriter
	restored := false
	t.Cleanup(func() {
		if !restored {
			os.Stdout, os.Stderr = originalOut, originalErr
			_ = outWriter.Close()
			_ = errWriter.Close()
		}
		_ = outReader.Close()
		_ = errReader.Close()
	})
	return func() (string, string) {
		_ = outWriter.Close()
		_ = errWriter.Close()
		os.Stdout, os.Stderr = originalOut, originalErr
		restored = true
		var stdout, stderr bytes.Buffer
		_, _ = stdout.ReadFrom(outReader)
		_, _ = stderr.ReadFrom(errReader)
		return stdout.String(), stderr.String()
	}
}

func TestCLIProjectionLintAsyncUsesPollResponse(t *testing.T) {
	spec := strings.Replace(cliCommandsTestSpec, "$ref: '#/components/schemas/TaskResult'\n  /notes:", `type: object
                properties:
                  status:
                    type: string
                    enum: [in_progress, completed, failed, requires_action]
                  outcome:
                    type: string
            text/event-stream:
              schema:
                type: object
                properties:
                  data:
                    type: object
  /notes:`, 1)
	require.NotEqual(t, cliCommandsTestSpec, spec)
	manifest := `
version: 1
commands:
  produce:
    op: CreateTask#CreateRenderTaskParams
    async:
      op: GetTask
      id:
        from: $.id
        to: {in: path, name: taskId}
      statePointer: $.status
      states:
        in_progress: pending
        completed: success
        failed: failure
        requires_action: handoff
    jq: %s
`
	_, warnings, err := decodeCLIWithSpec(t, spec, fmt.Sprintf(manifest, ".outcome"))
	require.NoError(t, err)
	assert.Empty(t, warnings)

	_, warnings, err = decodeCLIWithSpec(t, spec, fmt.Sprintf(manifest, ".steps"))
	require.EqualError(t, err, `command "produce": jq projection ".steps" always returns nothing — ".steps" does not exist in the response`)
	assert.Empty(t, warnings)
}
