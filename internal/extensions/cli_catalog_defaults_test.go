package extensions

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const cliCatalogPresetSpec = `openapi: 3.1.0
info: {title: Catalog preset test, version: 1.0.0}
paths:
  /widgets:
    post:
      operationId: createWidget
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                choice: {$ref: '#/components/schemas/ChoiceAlias'}
                choices:
                  type: array
                  items: {$ref: '#/components/schemas/Choices'}
                unrelated: {$ref: '#/components/schemas/UnrelatedChoices'}
      responses:
        '200': {description: OK}
components:
  schemas:
    ChoiceAlias:
      allOf:
        - $ref: '#/components/schemas/Choices'
    Choices:
      type: string
      enum: [alpha, beta]
      x-speakeasy-unknown-values: allow
      x-speakeasy-cli-catalog: {command: choices}
    UnrelatedChoices:
      type: string
      enum: [alpha, beta]
      x-speakeasy-cli-catalog: {command: unrelated-choices}
`

func TestCLICatalogDefaults_LinkedPresets(t *testing.T) {
	for _, tt := range []struct {
		name   string
		preset string
		want   []CLICatalogDefault
	}{
		{name: "reference and composition", preset: "{$.choice: alpha}", want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "alpha", Label: "create"}}},
		{name: "array", preset: "{$.choices: [beta, alpha, beta]}", want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "beta", Label: "create"}, {CatalogCommand: "choices", Value: "alpha", Label: "create"}}},
		{name: "promoted array", preset: "{$.choices: beta}", want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "beta", Label: "create"}}},
		{name: "unknown open enum value", preset: "{$.choice: future}"},
		{name: "separate same-valued enum", preset: "{$.unrelated: alpha}", want: []CLICatalogDefault{{CatalogCommand: "unrelated-choices", Value: "alpha", Label: "create"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manifest, _, err := decodeCLIWithSpec(t, cliCatalogPresetSpec, "version: 1\ncommands:\n  create:\n    op: createWidget\n    preset: "+tt.preset+"\n")
			require.NoError(t, err)
			require.Len(t, manifest.Commands, 1)
			assert.Equal(t, tt.want, manifest.Commands[0].CatalogDefaults)
		})
	}
}

func TestCLICatalogDefaults_EffectiveRoutePresets(t *testing.T) {
	spec := strings.ReplaceAll(cliRouteDispatchSpec, "        background:\n", "        choice: {$ref: '#/components/schemas/Choices'}\n        background:\n")
	spec += `    Choices:
      type: string
      enum: [alpha, beta]
      x-speakeasy-cli-catalog: {command: choices}
`
	for _, tt := range []struct {
		name        string
		routePreset string
		want        []CLICatalogDefault
	}{
		{name: "shared", want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "alpha", Label: "jobs run"}}},
		{name: "route override", routePreset: "        preset: {$.choice: beta}\n", want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "alpha", Label: "jobs run --engine"}, {CatalogCommand: "choices", Value: "beta", Label: "jobs run --pipeline"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manifestYAML := strings.Replace(cliRouteDispatchManifest, "        op: createJob#PipelineJobParams\n", "        op: createJob#PipelineJobParams\n"+tt.routePreset, 1)
			manifestYAML += "    preset: {$.choice: alpha}\n"
			manifest, _, err := decodeCLIWithSpec(t, spec, manifestYAML)
			require.NoError(t, err)
			assert.Equal(t, tt.want, manifest.Commands[0].CatalogDefaults)
		})
	}
}

func TestCLICatalogDefaults_PartialRouteCoverage(t *testing.T) {
	spec := strings.Replace(cliRouteDispatchSpec, "          description: Engine selection.\n", "          description: Engine selection.\n          x-speakeasy-cli-catalog: {command: engines}\n", 1)
	manifestYAML := strings.Replace(cliRouteDispatchManifest, "        selector: engine\n", "        selector: engine\n        preset: {$.engine: text-2}\n", 1)
	manifest, _, err := decodeCLIWithSpec(t, spec, manifestYAML)
	require.NoError(t, err)
	assert.Equal(t, []CLICatalogDefault{{CatalogCommand: "engines", Value: "text-2", Label: "jobs run --engine"}}, manifest.Commands[0].CatalogDefaults)
}

func TestCLICatalogDefaults_UnionAssociation(t *testing.T) {
	d := &cliManifestDecoder{}
	catalogArm := map[string]any{"type": "string", "enum": []any{"alpha"}, "x-speakeasy-cli-catalog": map[string]any{"command": "choices"}}
	for _, tt := range []struct {
		name  string
		other map[string]any
		want  []CLICatalogDefault
	}{
		{name: "unique matching arm", other: map[string]any{"type": "string", "enum": []any{"beta"}}, want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "alpha"}}},
		{name: "plain string arm", other: map[string]any{"type": "string"}, want: []CLICatalogDefault{{CatalogCommand: "choices", Value: "alpha"}}},
		{name: "ambiguous matching arms", other: map[string]any{"type": "string", "enum": []any{"alpha"}, "x-speakeasy-cli-catalog": map[string]any{"command": "other-choices"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			facts, err := d.propertyFacts(map[string]any{"oneOf": []any{catalogArm, tt.other}}, nil)
			require.NoError(t, err)
			defaults, err := d.catalogDefaultsForValue("alpha", facts)
			require.NoError(t, err)
			assert.Equal(t, tt.want, defaults)
		})
	}
}
