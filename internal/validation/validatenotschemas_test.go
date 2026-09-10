package validation_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const notSchemasSpecHeader = `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    post:
      operationId: test
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/payload'
      responses:
        '200':
          description: OK
components:
  schemas:
`

func Test_ValidateNotSchemas_Success(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name: "no not keyword",
			schema: notSchemasSpecHeader + `    payload:
      type: object
      properties:
        name:
          type: string`,
		},
		{
			name: "supported composites without not",
			schema: notSchemasSpecHeader + `    payload:
      oneOf:
        - type: string
        - type: integer`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(
				&config.Configuration{},
				validation.RulesetSpeakeasyUnsupported,
				validation.WithFilteredRules([]string{(&validation.ValidateNotSchemas{}).ID()}),
			)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))
			assert.Empty(t, res.GetValidationErrors())
		})
	}
}

func Test_ValidateNotSchemas_Errors(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		wantErrs []string
	}{
		{
			name: "component schema uses not",
			schema: notSchemasSpecHeader + `    payload:
      type: object
      not:
        type: string`,
			wantErrs: []string{
				"validation error: [line 24:7] generator-validate-not-schemas - schema uses `not`, which the SDK generator does not support — every operation reaching this schema will be silently dropped from the SDK",
			},
		},
		{
			name: "nested property uses not",
			schema: notSchemasSpecHeader + `    payload:
      type: object
      properties:
        aws:
          type: object
          not:
            required:
              - secret`,
			wantErrs: []string{
				"validation error: [line 27:11] generator-validate-not-schemas - schema uses `not`, which the SDK generator does not support — every operation reaching this schema will be silently dropped from the SDK",
			},
		},
		{
			name: "inline request body schema uses not",
			schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    post:
      operationId: test
      requestBody:
        content:
          application/json:
            schema:
              type: object
              not:
                required:
                  - forbidden
      responses:
        '200':
          description: OK
`,
			wantErrs: []string{
				"validation error: [line 14:15] generator-validate-not-schemas - schema uses `not`, which the SDK generator does not support — every operation reaching this schema will be silently dropped from the SDK",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(
				&config.Configuration{},
				validation.RulesetSpeakeasyUnsupported,
				validation.WithFilteredRules([]string{(&validation.ValidateNotSchemas{}).ID()}),
			)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := make([]string, len(errs))
			for i, e := range errs {
				errStrs[i] = e.Error()
			}
			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}

func rulesetContains(ruleset, ruleID string) bool {
	for _, r := range validation.NewRuleset(ruleset) {
		if r.ID() == ruleID {
			return true
		}
	}
	return false
}

// The rule ships only in the opt-in speakeasy-unsupported-constructs ruleset, so pipelines on
// the default rulesets don't start failing on regeneration.
func Test_ValidateNotSchemas_OnlyInUnsupportedRuleset(t *testing.T) {
	ruleID := (&validation.ValidateNotSchemas{}).ID()

	assert.True(t, rulesetContains(validation.RulesetSpeakeasyUnsupported, ruleID))
	assert.False(t, rulesetContains(validation.RulesetSpeakeasyGeneration, ruleID))
	assert.False(t, rulesetContains(validation.RulesetSpeakeasyRecommended, ruleID))
	assert.False(t, rulesetContains(validation.RulesetSpeakeasyOpenAPI, ruleID))
}

// speakeasy-unsupported-constructs must stay a strict superset of speakeasy-generation:
// pkg/generate refuses to generate unless every generation rule ran.
func Test_SpeakeasyUnsupported_IsSupersetOfGeneration(t *testing.T) {
	unsupported := map[string]bool{}
	for _, r := range validation.NewRuleset(validation.RulesetSpeakeasyUnsupported) {
		unsupported[r.ID()] = true
	}

	for _, r := range validation.NewRuleset(validation.RulesetSpeakeasyGeneration) {
		assert.True(t, unsupported[r.ID()], "generation rule %s missing from speakeasy-unsupported-constructs", r.ID())
	}
}

// End-to-end over the full ruleset (no rule filtering): the same `not` spec is
// silent by default and reports once the unsupported-constructs ruleset is selected.
func Test_ValidateNotSchemas_RulesetSelection(t *testing.T) {
	schema := notSchemasSpecHeader + `    payload:
      type: object
      not:
        type: string`

	tests := []struct {
		name        string
		ruleset     string
		wantFinding bool
	}{
		{
			name:        "generation ruleset stays silent",
			ruleset:     validation.RulesetSpeakeasyGeneration,
			wantFinding: false,
		},
		{
			name:        "unsupported-constructs ruleset reports",
			ruleset:     validation.RulesetSpeakeasyUnsupported,
			wantFinding: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			v, err := validation.NewValidator(&config.Configuration{}, tt.ruleset)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(schema), "", types.NewTargetFromTemplate("go"))

			found := false
			for _, e := range res.GetValidationErrors() {
				if strings.Contains(e.Error(), "generator-validate-not-schemas") {
					found = true
				}
			}
			assert.Equal(t, tt.wantFinding, found)
		})
	}
}

// The rule fails generation by default; a team that would rather ship an SDK
// with the affected operations dropped can downgrade it in lint.yaml.
func Test_ValidateNotSchemas_SeverityOverrideToWarn(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "true")

	schema := notSchemasSpecHeader + `    payload:
      type: object
      not:
        type: string`

	dir := t.TempDir()
	speakeasyDir := filepath.Join(dir, ".speakeasy")
	require.NoError(t, os.MkdirAll(speakeasyDir, 0o755))

	lintYaml := `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-unsupported-constructs
    rules:
      generator-validate-not-schemas:
        severity: warn
`
	require.NoError(t, os.WriteFile(filepath.Join(speakeasyDir, "lint.yaml"), []byte(lintYaml), 0o644))

	openapiPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(openapiPath, []byte(schema), 0o644))

	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyUnsupported,
		validation.WithWorkingDir(dir),
		validation.WithFilteredRules([]string{(&validation.ValidateNotSchemas{}).ID()}),
	)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), []byte(schema), openapiPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()
	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "validation warn:")
	assert.Contains(t, errs[0].Error(), "generator-validate-not-schemas")
}
