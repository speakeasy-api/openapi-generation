package validation_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_OverrideRuleSeverity(t *testing.T) {
	t.Setenv("SPEAKEASY_DEBUG", "true")

	dir := t.TempDir()
	speakeasyDir := filepath.Join(dir, ".speakeasy")

	err := os.MkdirAll(speakeasyDir, 0o755)
	require.NoError(t, err)

	lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
	lintYaml := `lintVersion: 1.0.0
defaultRuleset:  barRuleset
rulesets:
  barRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      validate-enums:
        severity: warn
`
	openapiYamlPath := filepath.Join(dir, "openapi.yaml")
	openapiYaml := `openapi: 3.1.0
info:
  title: Validation testing
  version: "0.0.1"
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/enum'
        '4XX':
          description: Not OK
        '5XX':
          description: Not OK
components:
  schemas:
    enum:
      type: string
      enum:
        - EAN_13
        - ean13
        - EAN13
`

	err = os.WriteFile(lintYamlPath, []byte(lintYaml), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(openapiYamlPath, []byte(openapiYaml), 0o644)
	require.NoError(t, err)

	v, err := validation.NewValidator(&config.Configuration{
		Generation: config.Generation{},
	}, validation.RulesetSpeakeasyGeneration)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), []byte(openapiYaml), openapiYamlPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	errStrings := make([]string, 0, len(errs))
	for _, e := range errs {
		errStrings = append(errStrings, e.Error())
	}

	assert.Equal(t, []string{
		"validation hint: [line 1:1] generator-retries - retries should be configured - consider adding `x-speakeasy-retries` extension globally or per-operation",
		"validation warn: [line 29:11] generator-validate-enums - enum value `EAN13` (`Ean13Upper`) will collide with `EAN_13` (`Ean13Upper`) [line `27`] when normalized, try using `x-speakeasy-enums`",
	}, errStrings)
}
