package validation_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLintConfig_BackwardsCompatibility tests that old lint.yaml configurations
// continue to work with the new validation system
func TestLintConfig_BackwardsCompatibility(t *testing.T) {
	// Common test OpenAPI spec with an enum collision issue
	testSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
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
                $ref: '#/components/schemas/TestEnum'
components:
  schemas:
    TestEnum:
      type: string
      enum:
        - EAN_13
        - ean13
        - EAN13
`

	tests := []struct {
		name             string
		lintYaml         string
		expectErrors     bool
		expectedRuleID   string
		expectedSeverity string
		description      string
	}{
		{
			name: "old rule name with generator- prefix",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      generator-validate-enums:
        severity: hint
`,
			expectErrors:     true,
			expectedRuleID:   "generator-validate-enums",
			expectedSeverity: "hint",
			description:      "Old rule names with generator- prefix should still work",
		},
		{
			name: "new rule name without generator- prefix",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      validate-enums:
        severity: warn
`,
			expectErrors:     true,
			expectedRuleID:   "validate-enums",
			expectedSeverity: "warn",
			description:      "New rule names without generator- prefix should work",
		},
		{
			name: "disable specific rule",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      generator-validate-enums:
        enabled: false
`,
			expectErrors: false,
			description:  "Disabling a rule should prevent errors from that rule",
		},
		{
			name: "unknown rule IDs should be ignored",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      unknown-rule-that-does-not-exist:
        severity: error
      generator-validate-enums:
        severity: warn
`,
			expectErrors:     true,
			expectedRuleID:   "generator-validate-enums",
			expectedSeverity: "warn",
			description:      "Unknown rule IDs should be ignored, not cause validation to fail",
		},
		{
			name: "old lintVersion 0.1.0",
			lintYaml: `lintVersion: 0.1.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      generator-validate-enums:
        severity: error
`,
			expectErrors:     true,
			expectedRuleID:   "unsupported lint version",
			expectedSeverity: "", // Not checking severity for config errors
			description:      "Older unsupported lintVersion should produce an error",
		},
		{
			name: "future lintVersion 99.0.0",
			lintYaml: `lintVersion: 99.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      generator-validate-enums:
        severity: warn
`,
			expectErrors:     true,
			expectedRuleID:   "unsupported lint version",
			expectedSeverity: "", // Not checking severity for config errors
			description:      "Future unsupported lintVersion should produce an error",
		},
		{
			name: "multiple extends",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
      - speakeasy-recommended
    rules:
      generator-validate-enums:
        severity: hint
`,
			expectErrors:     true,
			expectedRuleID:   "generator-validate-enums",
			expectedSeverity: "hint",
			description:      "Multiple ruleset extends should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")

			// Create temp directory with .speakeasy/lint.yaml
			dir := t.TempDir()
			speakeasyDir := filepath.Join(dir, ".speakeasy")
			err := os.MkdirAll(speakeasyDir, 0o755)
			require.NoError(t, err)

			// Write lint.yaml
			lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
			err = os.WriteFile(lintYamlPath, []byte(tt.lintYaml), 0o644)
			require.NoError(t, err)

			// Write OpenAPI spec
			openapiYamlPath := filepath.Join(dir, "openapi.yaml")
			err = os.WriteFile(openapiYamlPath, []byte(testSpec), 0o644)
			require.NoError(t, err)

			// Create validator and validate (with working directory so lint.yaml is loaded)
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithWorkingDir(dir))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(testSpec), openapiYamlPath, types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()

			if tt.expectErrors {
				assert.NotEmpty(t, errs, tt.description)

				if len(errs) > 0 && tt.expectedRuleID != "" {
					// Check that we got an error containing the expected text
					foundExpectedText := false
					for _, e := range errs {
						if contains(e.Error(), tt.expectedRuleID) {
							foundExpectedText = true
							// Only check severity if it's specified and we're looking for a validation error
							if tt.expectedSeverity != "" && contains(tt.expectedRuleID, "-") {
								assert.Contains(t, e.Error(), "validation "+tt.expectedSeverity, "Expected severity: "+tt.expectedSeverity)
							}
							break
						}
					}
					assert.True(t, foundExpectedText, "Expected error containing: "+tt.expectedRuleID)
				}
			} else if tt.expectedRuleID != "" {
				// Should have no errors from the disabled rule
				for _, e := range errs {
					assert.NotContains(t, e.Error(), tt.expectedRuleID, tt.description)
				}
			}
		})
	}
}

// TestLintConfig_EdgeCases tests edge cases and error handling
func TestLintConfig_EdgeCases(t *testing.T) {
	testSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: test
      responses:
        '200':
          description: OK
`

	tests := []struct {
		name        string
		lintYaml    string
		shouldWork  bool
		description string
	}{
		{
			name: "minimal valid lint.yaml",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: speakeasy-generation
`,
			shouldWork:  true,
			description: "Minimal valid configuration should work",
		},
		{
			name: "empty rulesets section",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets: []
    rules: {}
`,
			shouldWork:  true,
			description: "Empty rulesets and rules should work",
		},
		{
			name: "extra unknown fields",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
futureFeature: someValue
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    unknownField: ignored
`,
			shouldWork:  true,
			description: "Unknown fields should be ignored for forward compatibility",
		},
		{
			name: "rule with only severity",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      generator-validate-enums: warn
`,
			shouldWork:  true,
			description: "Shorthand rule syntax (just severity) should work",
		},
		{
			name: "rule with severity and enabled",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      generator-validate-enums:
        severity: error
        enabled: true
`,
			shouldWork:  true,
			description: "Rule with both severity and enabled should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")

			dir := t.TempDir()
			speakeasyDir := filepath.Join(dir, ".speakeasy")
			err := os.MkdirAll(speakeasyDir, 0o755)
			require.NoError(t, err)

			lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
			err = os.WriteFile(lintYamlPath, []byte(tt.lintYaml), 0o644)
			require.NoError(t, err)

			openapiYamlPath := filepath.Join(dir, "openapi.yaml")
			err = os.WriteFile(openapiYamlPath, []byte(testSpec), 0o644)
			require.NoError(t, err)

			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration)
			require.NoError(t, err)

			// Should not panic or error during validation
			assert.NotPanics(t, func() {
				res := validateSpec(v, context.Background(), []byte(testSpec), openapiYamlPath, types.NewTargetFromTemplate("go"))
				if tt.shouldWork {
					// Result should be valid (even if it has validation errors for the spec itself)
					assert.NotNil(t, res, tt.description)
				}
			}, tt.description)
		})
	}
}

// TestLintConfig_RulesetInheritance tests ruleset extension and inheritance
func TestLintConfig_RulesetInheritance(t *testing.T) {
	testSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: https://example.com/api
paths:
  /test:
    get:
      operationId: test
      responses:
        '200':
          description: OK
`

	tests := []struct {
		name          string
		lintYaml      string
		expectWarning bool
		warningRuleID string
		description   string
	}{
		{
			name: "inherit from speakeasy-recommended",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-recommended
`,
			expectWarning: true,
			warningRuleID: "style-oas3-host-not-example",
			description:   "Should inherit rules from speakeasy-recommended (including example.com check)",
		},
		{
			name: "override inherited rule severity",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-recommended
    rules:
      style-oas3-host-not-example:
        disabled: true
`,
			expectWarning: false,
			warningRuleID: "style-oas3-host-not-example",
			description:   "Should be able to disable inherited rules",
		},
		{
			name: "custom ruleset extending generation",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: strict
rulesets:
  strict:
    rulesets:
      - speakeasy-generation
      - speakeasy-recommended
`,
			expectWarning: true,
			warningRuleID: "style-oas3-host-not-example",
			description:   "Custom ruleset can extend multiple rulesets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")

			dir := t.TempDir()
			speakeasyDir := filepath.Join(dir, ".speakeasy")
			err := os.MkdirAll(speakeasyDir, 0o755)
			require.NoError(t, err)

			lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
			err = os.WriteFile(lintYamlPath, []byte(tt.lintYaml), 0o644)
			require.NoError(t, err)

			openapiYamlPath := filepath.Join(dir, "openapi.yaml")
			err = os.WriteFile(openapiYamlPath, []byte(testSpec), 0o644)
			require.NoError(t, err)

			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithWorkingDir(dir))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(testSpec), openapiYamlPath, types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()

			foundRule := false
			for _, e := range errs {
				if contains(e.Error(), tt.warningRuleID) {
					foundRule = true
					break
				}
			}

			if tt.expectWarning {
				assert.True(t, foundRule, "Expected to find rule: "+tt.warningRuleID)
			} else {
				assert.False(t, foundRule, "Expected NOT to find rule: "+tt.warningRuleID)
			}
		})
	}
}

// TestLintConfig_SeverityOverrides tests various severity override scenarios
func TestLintConfig_SeverityOverrides(t *testing.T) {
	testSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
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
                type: string
                enum:
                  - value1
                  - value1
`

	tests := []struct {
		name             string
		lintYaml         string
		expectedSeverity string
		ruleID           string
		description      string
	}{
		{
			name: "override to error",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-recommended
    rules:
      semantic-duplicated-enum:
        severity: error
`,
			expectedSeverity: "error",
			ruleID:           "semantic-duplicated-enum",
			description:      "Can upgrade warning to error",
		},
		{
			name: "override to hint",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-recommended
    rules:
      semantic-duplicated-enum:
        severity: hint
`,
			expectedSeverity: "hint",
			ruleID:           "semantic-duplicated-enum",
			description:      "Can downgrade warning to hint",
		},
		{
			name: "override to warn",
			lintYaml: `lintVersion: 1.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-recommended
    rules:
      semantic-duplicated-enum:
        severity: warn
`,
			expectedSeverity: "warn",
			ruleID:           "semantic-duplicated-enum",
			description:      "Can keep severity as warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")

			dir := t.TempDir()
			speakeasyDir := filepath.Join(dir, ".speakeasy")
			err := os.MkdirAll(speakeasyDir, 0o755)
			require.NoError(t, err)

			lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
			err = os.WriteFile(lintYamlPath, []byte(tt.lintYaml), 0o644)
			require.NoError(t, err)

			openapiYamlPath := filepath.Join(dir, "openapi.yaml")
			err = os.WriteFile(openapiYamlPath, []byte(testSpec), 0o644)
			require.NoError(t, err)

			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithWorkingDir(dir))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(testSpec), openapiYamlPath, types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()

			foundWithCorrectSeverity := false
			for _, e := range errs {
				errStr := e.Error()
				if contains(errStr, tt.ruleID) {
					if contains(errStr, "validation "+tt.expectedSeverity) {
						foundWithCorrectSeverity = true
						break
					}
				}
			}

			assert.True(t, foundWithCorrectSeverity,
				"Expected to find rule '%s' with severity '%s', got errors: %v",
				tt.ruleID, tt.expectedSeverity, errs)
		})
	}
}

// TestLintConfig_CustomRules tests that custom JS/TS rules configured in lint.yaml
// are loaded and executed during validation
func TestLintConfig_CustomRules(t *testing.T) {
	// Use the require-description.ts custom rule from testdata
	// (copied from github.com/speakeasy-api/openapi/openapi/linter/customrules/testdata)
	customRulePath, err := filepath.Abs("testdata/require-description.ts")
	require.NoError(t, err)

	// OpenAPI spec with an operation that has NO description
	testSpec := `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
`

	tests := []struct {
		name         string
		lintYaml     string
		expectCustom bool
		customRuleID string
		description  string
	}{
		{
			name: "custom rule fires on spec without description",
			lintYaml: `lintVersion: 2.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
customRules:
  paths:
    - "%s"
`,
			expectCustom: true,
			customRuleID: "custom-require-operation-description",
			description:  "Custom TS rule should fire and report missing description",
		},
		{
			name: "custom rule can be disabled via lint.yaml rules",
			lintYaml: `lintVersion: 2.0.0
defaultRuleset: myRuleset
rulesets:
  myRuleset:
    rulesets:
      - speakeasy-generation
    rules:
      - id: custom-require-operation-description
        disabled: true
customRules:
  paths:
    - "%s"
`,
			expectCustom: false,
			customRuleID: "custom-require-operation-description",
			description:  "Custom rule should be disabled when configured in rules section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			// Write lint.yaml with the testdata rule path substituted
			speakeasyDir := filepath.Join(dir, ".speakeasy")
			err := os.MkdirAll(speakeasyDir, 0o755)
			require.NoError(t, err)

			lintYaml := fmt.Sprintf(tt.lintYaml, customRulePath)
			lintYamlPath := filepath.Join(speakeasyDir, "lint.yaml")
			err = os.WriteFile(lintYamlPath, []byte(lintYaml), 0o644)
			require.NoError(t, err)

			// Write OpenAPI spec
			openapiYamlPath := filepath.Join(dir, "openapi.yaml")
			err = os.WriteFile(openapiYamlPath, []byte(testSpec), 0o644)
			require.NoError(t, err)

			// Create validator with working directory so lint.yaml is found
			v, err := validation.NewValidator(&config.Configuration{}, validation.RulesetSpeakeasyGeneration, validation.WithWorkingDir(dir))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(testSpec), openapiYamlPath, types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()

			foundCustomRule := false
			for _, e := range errs {
				if contains(e.Error(), tt.customRuleID) {
					foundCustomRule = true
					break
				}
			}

			if tt.expectCustom {
				assert.True(t, foundCustomRule, "%s: expected custom rule error '%s' but didn't find it in: %v", tt.description, tt.customRuleID, errs)
			} else {
				assert.False(t, foundCustomRule, "%s: expected custom rule '%s' NOT to fire but it did", tt.description, tt.customRuleID)
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
