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

// testSpecWithEnumCollision triggers generator-validate-enums
var testSpecWithEnumCollision = `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: getTest
      tags:
        - test
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TestEnum'
        '4XX':
          description: Error
        '5XX':
          description: Error
components:
  schemas:
    TestEnum:
      type: string
      enum:
        - EAN_13
        - ean13
        - EAN13
`

func setupLintTest(t *testing.T, lintYaml, spec string) (*validation.Validator, string) {
	t.Helper()
	t.Setenv("SPEAKEASY_DEBUG", "true")

	dir := t.TempDir()
	speakeasyDir := filepath.Join(dir, ".speakeasy")
	err := os.MkdirAll(speakeasyDir, 0o755)
	require.NoError(t, err)

	lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
	err = os.WriteFile(lintYamlPath, []byte(lintYaml), 0o644)
	require.NoError(t, err)

	openapiYamlPath := filepath.Join(dir, "openapi.yaml")
	err = os.WriteFile(openapiYamlPath, []byte(spec), 0o644)
	require.NoError(t, err)

	v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithWorkingDir(dir))
	require.NoError(t, err)

	return v, openapiYamlPath
}

func errStrings(errs []error) []string {
	strs := make([]string, 0, len(errs))
	for _, e := range errs {
		strs = append(strs, e.Error())
	}
	return strs
}

func findError(errs []error, substr string) (string, bool) {
	for _, e := range errs {
		if strings.Contains(e.Error(), substr) {
			return e.Error(), true
		}
	}
	return "", false
}

// TestRuleAliases_MissingAliases tests that recently added aliases translate correctly
func TestRuleAliases_MissingAliases(t *testing.T) {
	tests := []struct {
		name    string
		oldName string
		newName string
	}{
		{
			name:    "duplicate-schemas maps to generator-duplicate-inline-schemas",
			oldName: "duplicate-schemas",
			newName: "generator-duplicate-inline-schemas",
		},
		{
			name:    "validate-document maps to speakeasy-validate-document",
			oldName: "validate-document",
			newName: "speakeasy-validate-document",
		},
		{
			name:    "oas3-host-trailing-slash maps to style-oas3-host-trailing-slash",
			oldName: "oas3-host-trailing-slash",
			newName: "style-oas3-host-trailing-slash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.TranslateRuleName(tt.oldName)
			assert.Equal(t, tt.newName, result)
		})
	}
}

// TestRuleAliases_AllDocumentedRules verifies that all old rule names from the
// public docs (https://www.speakeasy.com/docs/sdks/prep-openapi/linting#available-rules)
// either translate to a new name or are already valid new rule names
func TestRuleAliases_AllDocumentedRules(t *testing.T) {
	// These are the old rule names from the docs that need aliases because
	// their new names are different
	aliasedRules := map[string]string{
		// Semantic rules
		"typed-enum":                         "semantic-typed-enum",
		"path-params":                        "semantic-path-params",
		"path-declarations-must-exist":       "semantic-path-declarations",
		"path-not-include-query":             "semantic-path-query",
		"duplicated-entry-in-enum":           "semantic-duplicated-enum",
		"no-eval-in-markdown":                "semantic-no-eval-in-markdown",
		"no-script-tags-in-markdown":         "semantic-no-script-tags-in-markdown",
		"operation-operationId":              "semantic-operation-operation-id",
		"operation-operationId-valid-in-url": "semantic-operation-id-valid-in-url",
		"no-ambiguous-paths":                 "semantic-no-ambiguous-paths",
		"oas3-unused-component":              "semantic-unused-component",
		"oas2-unused-definition":             "semantic-unused-component",
		"oas3-missing-example":               "oas3-example-missing",

		// Style rules
		"contact-properties":          "style-contact-properties",
		"info-contact":                "style-info-contact",
		"info-description":            "style-info-description",
		"info-license":                "style-info-license",
		"license-url":                 "style-license-url",
		"openapi-tags":                "style-openapi-tags",
		"openapi-tags-alphabetical":   "style-tags-alphabetical",
		"operation-tags":              "style-operation-tags",
		"operation-description":       "style-operation-description",
		"component-description":       "style-component-description",
		"tag-description":             "style-tag-description",
		"no-$ref-siblings":            "style-no-ref-siblings",
		"oas3-host-not-example":       "style-oas3-host-not-example",
		"oas3-server-trailing-slash":  "style-oas3-host-trailing-slash",
		"oas3-host-trailing-slash":    "style-oas3-host-trailing-slash",
		"oas3-parameter-description":  "style-oas3-parameter-description",
		"description-duplication":     "style-description-duplication",
		"oas3-api-servers":            "style-oas3-api-servers",
		"no-http-verbs-in-path":       "style-no-verbs-in-path",
		"paths-kebab-case":            "style-paths-kebab-case",
		"operation-4xx-response":      "style-operation-error-response",
		"operation-success-response":  "style-operation-success-response",
		"operation-singular-tag":      "style-operation-singular-tag",
		"operation-tag-defined":       "style-operation-tag-defined",
		"path-keys-no-trailing-slash": "style-path-trailing-slash",

		// OWASP rules
		"owasp-no-additionalProperties":          "owasp-no-additional-properties",
		"owasp-constrained-additionalProperties": "owasp-additional-properties-constrained",

		// Generator rules
		"duplicate-schema-name":      "generator-duplicate-schema-name",
		"duplicate-schemas":          "generator-duplicate-inline-schemas",
		"duplicate-operation-name":   "generator-duplicate-operation-name",
		"duplicate-errors":           "generator-duplicate-errors",
		"duplicate-path-params":      "generator-duplicate-path-params",
		"duplicate-properties":       "generator-duplicate-properties",
		"duplicate-tag":              "generator-duplicate-tag",
		"missing-error-response":     "generator-missing-error-response",
		"missing-examples":           "generator-missing-examples",
		"pagination":                 "generator-pagination",
		"retries":                    "generator-retries",
		"validate-composite-schemas": "generator-validate-composite-schemas",
		"validate-consts-defaults":   "generator-validate-consts-defaults",
		"validate-content-type":      "generator-validate-content-type",
		"validate-deprecation":       "generator-validate-deprecation",
		"validate-document":          "speakeasy-validate-document",
		"validate-enums":             "generator-validate-enums",
		"validate-extensions":        "generator-validate-extensions",
		"validate-parameters":        "generator-validate-parameters",
		"validate-paths":             "generator-validate-paths",
		"validate-requests":          "generator-validate-requests",
		"validate-responses":         "generator-validate-responses",
		"validate-security":          "generator-validate-security",
		"validate-servers":           "generator-validate-servers",
		"validate-types":             "generator-validate-types",
	}

	for oldName, expectedNewName := range aliasedRules {
		t.Run(oldName, func(t *testing.T) {
			result := validation.TranslateRuleName(oldName)
			assert.Equal(t, expectedNewName, result, "alias for %q should be %q", oldName, expectedNewName)
		})
	}
}

// Tests a config that extends a built-in ruleset and downgrades two rules
// referenced by their pre-alias names: path-params, validate-composite-schemas.
func TestRealWorld_OldRuleNamesWarnSeverity(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: restAPIRuleset
rulesets:
  restAPIRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      path-params: { severity: warn }
      validate-composite-schemas: { severity: warn }
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Should not panic or fail to load config
	assert.NotNil(t, res)

	// Should still get enum collision errors from speakeasy-generation
	_, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors from speakeasy-generation base ruleset, got: %v", errStrings(errs))
}

// Tests a config using pre-alias rule names typed-enum and validate-enums
// with severity overrides.
func TestRealWorld_AliasedRuleSeverityOverride(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: enumSeverityRuleset
rulesets:
  enumSeverityRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      typed-enum: { severity: warn }
      validate-enums: { severity: warn }
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// validate-enums should be downgraded to warn
	errStr, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors, got: %v", errStrings(errs))
	if found {
		assert.Contains(t, errStr, "validation warn", "validate-enums severity should be overridden to warn")
	}
}

// Tests a config that extends both speakeasy-recommended and
// speakeasy-generation and disables oas3-missing-example.
func TestRealWorld_MultipleExtendsDisabledRule(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
  defaultRuleset: ai-api
  rulesets:
    ai-api:
    rulesets:
      - speakeasy-recommended
      - speakeasy-generation
    rules:
      oas3-missing-example:
        severity: "off"
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// oas3-missing-example should be disabled (translated to oas3-example-missing, then disabled)
	_, found := findError(errs, "oas3-example-missing")
	assert.False(t, found, "oas3-example-missing should be disabled, but found in errors: %v", errStrings(errs))
}

// Tests a config that uses severity: off to disable validate-content-type.
func TestRealWorld_SeverityOffDisablesRule(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: contentTypeRuleset
rulesets:
  contentTypeRuleset:
    rulesets:
      - speakeasy-recommended
    rules:
      validate-content-type:
        severity: off
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// validate-content-type should be disabled
	_, found := findError(errs, "generator-validate-content-type")
	assert.False(t, found, "generator-validate-content-type should be disabled, but found in errors: %v", errStrings(errs))
}

// Tests a config that overrides many rules by their pre-alias names and
// references a custom ruleset via the rulesets extension.
func TestRealWorld_ManyOldRuleNamesErrorSeverity(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: strictSDKRuleset
rulesets:
  strictSDKRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      operation-tag-defined:
        severity: error
      oas3-unused-component:
        severity: error
      validate-parameters:
        severity: error
      validate-security:
        severity: error
      missing-error-response:
        severity: error
      validate-consts-defaults:
        severity: error
      duplicate-schemas:
        severity: error
      operation-success-response:
        severity: error
      operation-operationId:
        severity: error
      duplicated-entry-in-enum:
        severity: error
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// Should get enum collision errors from speakeasy-generation
	_, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors, got: %v", errStrings(errs))
}

// TestRealWorld_DefaultRulesetPointsToBuiltIn tests behavior when defaultRuleset
// in lint.yaml matches a built-in ruleset name, while the custom rulesets have different names.
func TestRealWorld_DefaultRulesetPointsToBuiltIn(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: speakeasy-generation
rulesets:
  myCustomRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      validate-enums:
        severity: warn
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// When defaultRuleset points to a built-in, the custom ruleset overrides are NOT applied.
	// The severity override to "warn" is ignored because "myCustomRuleset" is never loaded.
	// This documents current behavior - the built-in speakeasy-generation ruleset is used as-is.
	errStr, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors, got: %v", errStrings(errs))
	if found {
		// The error severity should be the built-in default (error), NOT warn
		assert.Contains(t, errStr, "validation error",
			"when defaultRuleset points to built-in, custom ruleset overrides should not apply; severity should be default 'error' not 'warn'")
	}
}

// TestRealWorld_SpectralCustomRulesIgnored tests that custom spectral-style rules
// (with given/then/function fields) are silently ignored and don't cause errors.
// These fields are not part of the lint.Rule struct and are dropped during YAML parsing.
func TestRealWorld_SpectralCustomRulesIgnored(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: customSpectralRuleset
rulesets:
  customSpectralRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      require-endpoint-renamings:
        description: All operations must have x-speakeasy-group and x-speakeasy-name-override
        severity: error
        given: "$.paths.*[?(@['x-speakeasy-ignore'] != true)]"
        then:
          - field: "x-speakeasy-group"
            function: truthy
          - field: "x-speakeasy-name-override"
            function: truthy
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// The spectral custom rule should be silently ignored (not cause a crash or config error)
	// Validation should still work with the base speakeasy-generation rules
	_, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "base speakeasy-generation rules should still work when spectral rules are present, got: %v", errStrings(errs))

	// The custom spectral rule should NOT appear in errors (it's not a registered rule)
	_, found = findError(errs, "require-endpoint-renamings")
	assert.False(t, found, "spectral custom rule should be silently ignored, not appear in errors")
}

// TestRealWorld_SpectralCustomRulesWithCustomFunction tests that custom spectral rules
// referencing a user-defined function are silently ignored
func TestRealWorld_SpectralCustomRulesWithCustomFunction(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: customFunctionRuleset
rulesets:
  customFunctionRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      duplicate-examples-rule:
        description: Detects duplicate example fields
        message: Duplicate example field detected
        severity: error
        given: $.paths
        then:
          function: customRuleFunction
      validate-enums:
        severity: warn
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// The custom function rule should be silently ignored
	_, found := findError(errs, "duplicate-examples-rule")
	assert.False(t, found, "custom spectral rule should be silently ignored")

	// validate-enums override should still work
	errStr, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors, got: %v", errStrings(errs))
	if found {
		assert.Contains(t, errStr, "validation warn", "validate-enums severity should be warn")
	}
}

// TestRealWorld_MultipleExtendsWithOverrides tests extending multiple rulesets
// with severity overrides on rules from different rulesets
func TestRealWorld_MultipleExtendsWithOverrides(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-recommended
      - speakeasy-generation
    rules:
      validate-enums:
        severity: warn
      oas3-missing-example:
        severity: "off"
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// validate-enums from speakeasy-generation should be overridden to warn
	errStr, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors, got: %v", errStrings(errs))
	if found {
		assert.Contains(t, errStr, "validation warn", "validate-enums severity should be warn")
	}

	// oas3-missing-example (translated from oas3-missing-example) should be disabled
	_, found = findError(errs, "oas3-example-missing")
	assert.False(t, found, "oas3-example-missing should be disabled, but found in errors")
}

// TestRealWorld_DuplicateSchemasAlias tests the duplicate-schemas -> generator-duplicate-inline-schemas alias,
// where duplicate-schemas is the pre-alias name.
func TestRealWorld_DuplicateSchemasAlias(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      duplicate-schemas:
        severity: error
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))

	// Should not error during config loading - the alias should translate correctly
	assert.NotNil(t, res)
}

// TestRealWorld_SimpleExtendOnly tests a lint.yaml that just extends a built-in
// ruleset with no custom rules.
func TestRealWorld_SimpleExtendOnly(t *testing.T) {
	lintYaml := `lintVersion: 1.0.0
defaultRuleset: PathParamRuleset
rulesets:
  PathParamRuleset:
    rulesets:
      - speakeasy-generation
`
	v, specPath := setupLintTest(t, lintYaml, testSpecWithEnumCollision)
	res := validateSpec(v, context.Background(), []byte(testSpecWithEnumCollision), specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	assert.NotNil(t, res)

	// Should get the default speakeasy-generation errors
	_, found := findError(errs, "generator-validate-enums")
	assert.True(t, found, "expected generator-validate-enums errors from speakeasy-generation, got: %v", errStrings(errs))
}
