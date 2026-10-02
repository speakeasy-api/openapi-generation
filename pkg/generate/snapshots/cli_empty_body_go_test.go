package snapshots

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/require"
)

const cliEmptyBodySpec = `openapi: 3.0.3
info:
  title: Widget API
  version: 1.0.0
servers:
  - url: https://api.example.com
x-test-operation: &mixedOperation
  description: An optional empty object body accompanies a required path parameter.
  tags: [widgets]
  security: []
  parameters:
    - name: id
      in: path
      required: true
      example: widget_123
      schema:
        type: string
  requestBody: &emptyBody
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/EmptyBody'
  responses: &responses
    '200':
      description: Accepted
paths:
  /widgets/{id}/optional:
    post:
      <<: *mixedOperation
      operationId: optional
  /widgets/{id}/optionalexample:
    post:
      <<: *mixedOperation
      operationId: optionalexample
      description: An explicit empty object example does not require an optional body.
      requestBody:
        content:
          application/json:
            example: {}
            schema:
              $ref: '#/components/schemas/EmptyBody'
  /widgets/{id}/required:
    post:
      <<: *mixedOperation
      operationId: required
      description: A required empty object body accepts an empty object without invented properties.
      requestBody:
        <<: *emptyBody
        required: true
  /widgets/standaloneoptional:
    post:
      <<: *mixedOperation
      operationId: standaloneoptional
      description: An optional empty object body is the only operation input.
      parameters: []
  /widgets/standalonerequired:
    post:
      <<: *mixedOperation
      operationId: standalonerequired
      description: A required empty object body is the only operation input.
      parameters: []
      requestBody:
        <<: *emptyBody
        required: true
  /widgets/{id}/formoptional:
    post:
      <<: *mixedOperation
      operationId: formoptional
      description: An optional empty form object accompanies a required path parameter.
      requestBody: &emptyFormBody
        content:
          application/x-www-form-urlencoded:
            schema:
              $ref: '#/components/schemas/EmptyBody'
  /widgets/{id}/formrequired:
    post:
      <<: *mixedOperation
      operationId: formrequired
      description: A required empty form object accepts an empty object without invented properties.
      requestBody:
        <<: *emptyFormBody
        required: true
  /widgets/{id}/scalaroptional:
    post:
      <<: *mixedOperation
      operationId: scalaroptional
      description: An optional scalar body without an object example does not need an invented object payload.
      requestBody:
        content:
          application/json:
            schema:
              type: string
  /widgets/{id}/nullablerequired:
    post:
      <<: *mixedOperation
      operationId: nullablerequired
      description: A nullable empty object body does not need an invented payload.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              nullable: true
              properties: {}
  /widgets/{id}/nullableoptional:
    post:
      <<: *mixedOperation
      operationId: nullableoptional
      description: An optional nullable empty object body preserves the distinction between omission and null.
      requestBody:
        content:
          application/json:
            schema:
              type: object
              nullable: true
              properties: {}
  /widgets/standalonenullable:
    post:
      <<: *mixedOperation
      operationId: standalonenullable
      description: A nullable empty object body is the only operation input.
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              nullable: true
              properties: {}
  /widgets/{id}/maprequired:
    post:
      <<: *mixedOperation
      operationId: maprequired
      description: A map body retains its whole-body input flag.
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              additionalProperties:
                type: string
  /widgets/{id}/unionrequired:
    post:
      <<: *mixedOperation
      operationId: unionrequired
      description: A union body retains its whole-body input flag and explicit example.
      requestBody:
        required: true
        content:
          application/json:
            example:
              label: sample
            schema:
              oneOf:
                - type: object
                  required: [label]
                  properties:
                    label:
                      type: string
                - type: string
  /widgets/parameter:
    get:
      operationId: parameter
      description: An empty object query parameter remains a JSON input rather than a body wrapper.
      tags: [widgets]
      security: []
      parameters:
        - name: filter
          in: query
          schema:
            $ref: '#/components/schemas/EmptyBody'
      responses: *responses
components:
  schemas:
    EmptyBody:
      description: An object with no declared properties.
      type: object
      properties: {}
`

func TestSnapCLIEmptyBodyExamplesAndFlags(t *testing.T) {
	for _, style := range []string{"compact", "full"} {
		t.Run(style, func(t *testing.T) {
			shouldCompile := false
			snaptest.DoTestSnapshot(t, snaptest.Options{
				Spec: cliEmptyBodySpec,
				GenYaml: fmt.Sprintf(`cli:
  packageName: github.com/example/widget-cli
  cliName: widget-cli
  envVarPrefix: WIDGET
  helpStyle: %s
`, style),
				ShouldCompile: &shouldCompile,
				AfterGenerate: func(t *testing.T, outputDir string) {
					t.Helper()
					readCommand := func(name string) string {
						t.Helper()
						content, err := os.ReadFile(filepath.Join(outputDir, "internal", "cli", "widgets", name+".go"))
						require.NoError(t, err)
						return string(content)
					}
					examplePattern := regexp.MustCompile(`Example:\s*"([^"\n]*)"`)
					for _, name := range []string{"optional", "optionalexample", "scalaroptional", "formoptional", "nullablerequired"} {
						command := readCommand(name)
						require.Regexp(t, examplePattern, command)
						require.Equal(t, "  widget-cli widgets "+name+" --id widget_123", examplePattern.FindStringSubmatch(command)[1])
					}
					for _, name := range []string{"optional", "optionalexample", "required", "formoptional", "formrequired"} {
						command := readCommand(name)
						require.Contains(t, command, `cmd.Flags().String("body",`)
						require.NotContains(t, command, `FlagName: "body-param"`)
					}
					usageSource, err := os.ReadFile(filepath.Join(outputDir, "internal", "usage", "schema.go"))
					require.NoError(t, err)
					usagePattern := regexp.MustCompile(`"widgets (optional|optionalexample|required|formoptional|formrequired)":\s*"(.*)"`)
					usageSchemas := usagePattern.FindAllStringSubmatch(string(usageSource), -1)
					require.Len(t, usageSchemas, 5)
					for _, schema := range usageSchemas {
						require.Contains(t, schema[2], "--body <body>")
						require.NotContains(t, schema[2], "--body-param")
					}
					for _, name := range []string{"required", "formrequired"} {
						command := readCommand(name)
						require.Regexp(t, examplePattern, command)
						require.Equal(t, "  widget-cli widgets "+name+" --id widget_123 --body '{}'", examplePattern.FindStringSubmatch(command)[1])
						require.Contains(t, command, `PromptFlagSpec{Required: true, Kind: "json", BodyFlag: true}`)
					}
					standaloneOptional := readCommand("standaloneoptional")
					require.Regexp(t, examplePattern, standaloneOptional)
					require.Equal(t, "  widget-cli widgets standaloneoptional", examplePattern.FindStringSubmatch(standaloneOptional)[1])
					standaloneNullable := readCommand("standalonenullable")
					require.Regexp(t, examplePattern, standaloneNullable)
					require.Equal(t, "  widget-cli widgets standalonenullable", examplePattern.FindStringSubmatch(standaloneNullable)[1])
					require.NotContains(t, readCommand("optional"), "alternative to individual flags")
					nullableOptional := readCommand("nullableoptional")
					require.Contains(t, nullableOptional, "Kind: flagutil.FlagKindJSON")
					require.Contains(t, nullableOptional, "alternative to individual flags")
					standaloneRequired := readCommand("standalonerequired")
					require.Contains(t, standaloneRequired, "--empty-body '{}'")
					for _, command := range []string{standaloneOptional, standaloneRequired} {
						require.Contains(t, command, `cmd.Flags().String("empty-body",`)
						require.Contains(t, command, "flagutil.BuildRequestBody[")
					}
					require.Contains(t, readCommand("parameter"), `FlagName: "filter"`)
					require.Contains(t, readCommand("parameter"), "Kind: flagutil.FlagKindJSON")
					require.Contains(t, readCommand("maprequired"), `FlagName: "body-param"`)
					require.Contains(t, readCommand("unionrequired"), "Kind: flagutil.FlagKindUnion")
					require.Contains(t, readCommand("unionrequired"), `--body '{\"label\":\"sample\"}'`)
				},
			})
		})
	}
}
