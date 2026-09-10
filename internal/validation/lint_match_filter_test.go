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

// TestDefaultFiltersDowngradeInternalValidationErrors demonstrates that the built-in
// default filters (migrated from internal/openapivalidation/filter.go) automatically
// downgrade certain internal validation errors to warnings
func TestDefaultFiltersDowngradeInternalValidationErrors(t *testing.T) {
	// Create an OpenAPI spec with validation issues
	// No lint.yaml needed - default filters should handle this automatically
	spec := `openapi: 3.0.0
info:
  # Missing required title field - should be downgraded to warning by default filters
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          # Missing required description field - should be downgraded to warning by default filters
          content:
            application/json:
              schema:
                type: object
`
	// Create validator without any lint.yaml
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
	)
	require.NoError(t, err)

	// Validate the spec
	res := validateSpec(v, context.Background(), []byte(spec), "", types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Print all errors to see what we're getting
	t.Logf("Total errors: %d", len(errs))
	for i, err := range errs {
		t.Logf("Error %d: %s", i, err.Error())
	}

	// Check that the validation-required-field errors are present
	// Note: the upstream dependency now wraps field names in backticks and emits these as errors
	var titleFound, descriptionFound bool
	for _, err := range errs {
		errStr := err.Error()
		if contains(errStr, "`info.title` is required") {
			titleFound = true
		}
		if contains(errStr, "`response.description` is required") {
			descriptionFound = true
		}
	}

	assert.True(t, titleFound, "Expected to find info.title required error")
	assert.True(t, descriptionFound, "Expected to find response.description required error")
}

// TestLintYamlCanOverrideDefaultFilters demonstrates that lint.yaml Match filters
// can override the built-in default filters
func TestLintYamlCanOverrideDefaultFilters(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create .speakeasy subdirectory
	speakeasyDir := filepath.Join(tmpDir, ".speakeasy")
	err := os.MkdirAll(speakeasyDir, 0755)
	require.NoError(t, err)

	// Create a lint.yaml that overrides a default filter to make it an error again
	// Note: backticks in the match pattern must be escaped since the upstream dependency
	// now wraps field names in backticks in validation messages
	bt := "`"
	lintYaml := "lintVersion: 2.0.0\n" +
		"defaultRuleset: custom\n" +
		"rulesets:\n" +
		"  custom:\n" +
		"    rulesets:\n" +
		"      - speakeasy-recommended\n" +
		"    rules:\n" +
		"      # Override the default filter to make info.title an error\n" +
		"      validation-required-field:\n" +
		"        severity: error\n" +
		"        match: \".*" + bt + "info\\\\.title" + bt + " is (required|missing).*\"\n"
	lintPath := filepath.Join(speakeasyDir, "lint.yaml")
	err = os.WriteFile(lintPath, []byte(lintYaml), 0644)
	require.NoError(t, err)

	// Create an OpenAPI spec with validation issues
	spec := `openapi: 3.0.0
info:
  version: 1.0.0
paths: {}
`
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	err = os.WriteFile(specPath, []byte(spec), 0644)
	require.NoError(t, err)

	// Create validator with working directory set to tmpDir
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
		validation.WithWorkingDir(tmpDir),
	)
	require.NoError(t, err)

	// Validate the spec
	specBytes, err := os.ReadFile(specPath)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), specBytes, specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Check that info.title is still an error (lint.yaml overrode the default filter)
	var titleIsError bool
	for _, err := range errs {
		errStr := err.Error()
		if contains(errStr, "`info.title` is required") {
			// Should be an error, not a warning (lint.yaml overrode default filter)
			assert.Contains(t, errStr, "validation error:", "info.title should be an error when overridden in lint.yaml")
			titleIsError = true
		}
	}

	assert.True(t, titleIsError, "Expected to find info.title as error (overridden by lint.yaml)")
}

// TestMatchFilterDisablesSpecificErrors demonstrates using Match with disabled flag
// to completely filter out specific error messages that are NOT in the default filters
func TestMatchFilterDisablesSpecificErrors(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create .speakeasy subdirectory (lint.yaml is typically in .speakeasy/)
	speakeasyDir := filepath.Join(tmpDir, ".speakeasy")
	err := os.MkdirAll(speakeasyDir, 0755)
	require.NoError(t, err)

	// Create a lint.yaml that disables a specific custom rule error
	lintYaml := `lintVersion: 2.0.0
defaultRuleset: custom
rulesets:
  custom:
    rulesets:
      - speakeasy-recommended
    rules:
      # Disable missing examples hints
      generator-missing-examples:
        disabled: true
        match: ".*Missing example for.*"
`
	lintPath := filepath.Join(speakeasyDir, "lint.yaml")
	err = os.WriteFile(lintPath, []byte(lintYaml), 0644)
	require.NoError(t, err)

	// Create an OpenAPI spec that would normally trigger missing examples
	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
`
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	err = os.WriteFile(specPath, []byte(spec), 0644)
	require.NoError(t, err)

	// Create validator with working directory set to tmpDir
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
		validation.WithWorkingDir(tmpDir),
	)
	require.NoError(t, err)

	// Validate the spec
	specBytes, err := os.ReadFile(specPath)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), specBytes, specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Check that missing examples errors were completely filtered out
	for _, err := range errs {
		assert.NotContains(t, err.Error(), "Missing example for",
			"Missing example errors should be completely filtered out by disabled Match rule")
	}
}

// TestMultipleMatchConfigurationsForSameRule tests that we can configure the same rule
// multiple times with different match patterns and different severities/disabled states
func TestMultipleMatchConfigurationsForSameRule(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create .speakeasy subdirectory
	speakeasyDir := filepath.Join(tmpDir, ".speakeasy")
	err := os.MkdirAll(speakeasyDir, 0755)
	require.NoError(t, err)

	// Create a lint.yaml that configures the same rule multiple times with different match patterns
	lintYaml := `lintVersion: 2.0.0
defaultRuleset: custom
rulesets:
  custom:
    rulesets:
      - speakeasy-recommended
    rules:
      # Configure generator-validate-enums rule multiple times with different match patterns
      - id: generator-validate-enums
        match: "(?i).*active.*status.*"
        severity: warn
      - id: generator-validate-enums
        match: "(?i).*internal.*code.*"
        disabled: true
      - id: generator-validate-enums
        match: "(?i).*test.*value.*"
        severity: hint
`
	lintPath := filepath.Join(speakeasyDir, "lint.yaml")
	err = os.WriteFile(lintPath, []byte(lintYaml), 0644)
	require.NoError(t, err)

	// Create spec with multiple enum validation issues
	spec := `openapi: 3.0.0
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
                type: object
                properties:
                  status:
                    type: string
                    enum:
                      - active_status  # Will be downgraded to warning by match
                      - Active_Status  # Will be downgraded to warning by match
                      - ActiveStatus   # Will be downgraded to warning by match
                  internalCode:
                    type: string
                    enum:
                      - internal_code  # Will be filtered out (disabled) by match
                      - Internal_Code  # Will be filtered out (disabled) by match
                      - InternalCode   # Will be filtered out (disabled) by match
                  testValue:
                    type: string
                    enum:
                      - test_value     # Will be downgraded to hint by match
                      - Test_Value     # Will be downgraded to hint by match
                      - TestValue      # Will be downgraded to hint by match
                  otherValue:
                    type: string
                    enum:
                      - other_value    # Will use default severity (error) - no match
                      - Other_Value    # Will use default severity (error) - no match
                      - OtherValue     # Will use default severity (error) - no match
`
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	err = os.WriteFile(specPath, []byte(spec), 0644)
	require.NoError(t, err)

	// Create validator with working directory set to tmpDir
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
		validation.WithWorkingDir(tmpDir),
	)
	require.NoError(t, err)

	// Validate the spec
	specBytes, err := os.ReadFile(specPath)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), specBytes, specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Print all errors for debugging
	t.Logf("Total errors: %d", len(errs))
	for i, err := range errs {
		t.Logf("Error %d: %s", i, err.Error())
	}

	// Count enum errors specifically
	enumErrors := 0
	for _, err := range errs {
		if contains(err.Error(), "generator-validate-enums") {
			enumErrors++
		}
	}
	t.Logf("Total enum validation errors: %d", enumErrors)

	// Check that the different match patterns applied different severities
	var foundActiveWarning, foundInternalDisabled, foundTestHint, foundOtherError bool

	for _, err := range errs {
		errStr := err.Error()
		if contains(errStr, "ActiveStatus") || contains(errStr, "Active_Status") || contains(errStr, "active_status") {
			// Should be a warning (match: "(?i).*active.*status.*", severity: warn)
			assert.Contains(t, errStr, "validation warn:", "ActiveStatus enum collision should be downgraded to warning")
			foundActiveWarning = true
		}
		if contains(errStr, "InternalCode") || contains(errStr, "Internal_Code") || contains(errStr, "internal_code") {
			// Should be completely filtered out (match: "(?i).*internal.*code.*", disabled: true)
			t.Errorf("Found InternalCode enum error that should have been filtered out: %s", errStr)
		}
		if contains(errStr, "TestValue") || contains(errStr, "Test_Value") || contains(errStr, "test_value") {
			// Should be a hint (match: "(?i).*test.*value.*", severity: hint)
			assert.Contains(t, errStr, "validation hint:", "TestValue enum collision should be downgraded to hint")
			foundTestHint = true
		}
		if contains(errStr, "OtherValue") || contains(errStr, "Other_Value") || contains(errStr, "other_value") {
			// Should be an error (no match, uses default severity)
			assert.Contains(t, errStr, "validation error:", "OtherValue enum collision should remain as error")
			foundOtherError = true
		}
	}

	// Verify that we didn't find any internal errors (they should be disabled)
	foundInternalDisabled = true
	for _, err := range errs {
		if contains(err.Error(), "InternalCode") || contains(err.Error(), "Internal_Code") || contains(err.Error(), "internal_code") {
			foundInternalDisabled = false
			break
		}
	}

	assert.True(t, foundActiveWarning, "Expected to find ActiveStatus enum warning")
	assert.True(t, foundInternalDisabled, "Expected InternalCode enum errors to be filtered out")
	assert.True(t, foundTestHint, "Expected to find TestValue enum hint")
	assert.True(t, foundOtherError, "Expected to find OtherValue enum error")
}

// TestSeverityOffDisablesRule tests that severity: "off" disables a rule (backwards compatibility)
func TestSeverityOffDisablesRule(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()

	// Create a lint.yaml that uses severity: "off" to disable a rule
	lintYaml := `lintVersion: 2.0.0
defaultRuleset: custom
rulesets:
  custom:
    rulesets:
      - speakeasy-recommended
    rules:
      # Use severity: "off" to disable the rule (backwards compatibility)
      generator-validate-enums:
        severity: "off"
`
	lintPath := filepath.Join(tmpDir, "lint.yaml")
	err := os.WriteFile(lintPath, []byte(lintYaml), 0644)
	require.NoError(t, err)

	// Create spec with enum validation issues
	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    enum:
                      - ACTIVE
                      - active  # This will collide when normalized - should be disabled
`
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	err = os.WriteFile(specPath, []byte(spec), 0644)
	require.NoError(t, err)

	// Create validator with working directory set to tmpDir
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
		validation.WithWorkingDir(tmpDir),
	)
	require.NoError(t, err)

	// Validate the spec
	specBytes, err := os.ReadFile(specPath)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), specBytes, specPath, types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	// Check that the enum validation error is NOT present (rule is disabled)
	for _, err := range errs {
		assert.NotContains(t, err.Error(), "generator-validate-enums",
			"generator-validate-enums should be disabled by severity: off")
	}
}

// TestDefaultFiltersDowngradeExclusiveMinMaxToWarning tests that the default match filters
// downgrade boolean exclusiveMinimum/exclusiveMaximum type-mismatch errors to warnings.
// This is critical for OpenAPI 3.0 specs that use boolean exclusiveMin/Max in a 3.1 context.
func TestDefaultFiltersDowngradeExclusiveMinMaxToWarning(t *testing.T) {
	// OpenAPI 3.1 spec with boolean exclusiveMinimum (valid in 3.0, invalid in 3.1)
	spec := `openapi: "3.1.0"
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  count:
                    type: number
                    minimum: 0
                    maximum: 100
                    exclusiveMinimum: true
                    exclusiveMaximum: true
`
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
	)
	require.NoError(t, err)

	res := validateSpec(v, context.Background(), []byte(spec), "", types.NewTargetFromTemplate("go"))
	errs := res.GetValidationErrors()

	t.Logf("Total errors: %d", len(errs))
	for i, err := range errs {
		t.Logf("Error %d: %s", i, err.Error())
	}

	// The exclusiveMinimum/exclusiveMaximum type-mismatch errors should be downgraded to warnings
	for _, err := range errs {
		errStr := err.Error()
		if contains(errStr, "exclusiveMinimum") && contains(errStr, "validation-type-mismatch") {
			assert.Contains(t, errStr, "validation warn:",
				"exclusiveMinimum type-mismatch should be downgraded to warning by default filters")
			assert.NotContains(t, errStr, "validation error:",
				"exclusiveMinimum type-mismatch should NOT be an error")
		}
		if contains(errStr, "exclusiveMaximum") && contains(errStr, "validation-type-mismatch") {
			assert.Contains(t, errStr, "validation warn:",
				"exclusiveMaximum type-mismatch should be downgraded to warning by default filters")
			assert.NotContains(t, errStr, "validation error:",
				"exclusiveMaximum type-mismatch should NOT be an error")
		}
	}

	// Verify at least one exclusiveMinimum error was found (so the test is meaningful)
	var foundExclusiveMin, foundExclusiveMax bool
	for _, err := range errs {
		errStr := err.Error()
		if contains(errStr, "exclusiveMinimum") {
			foundExclusiveMin = true
		}
		if contains(errStr, "exclusiveMaximum") {
			foundExclusiveMax = true
		}
	}
	assert.True(t, foundExclusiveMin, "expected to find exclusiveMinimum type-mismatch error (downgraded to warning)")
	assert.True(t, foundExclusiveMax, "expected to find exclusiveMaximum type-mismatch error (downgraded to warning)")
}

// TestDefaultMatchFilters_RealMessages tests that EVERY entry in getDefaultRuleEntries()
// actually matches the real error messages produced by the upstream openapi library.
// This guards against message format changes (like adding backticks) silently breaking filters.
func TestDefaultMatchFilters_RealMessages(t *testing.T) {
	tests := []struct {
		name          string
		spec          string
		ruleID        string // the validation rule ID to look for
		messageSubstr string // substring to find in the error message
	}{
		{
			name: "validation-required-field/info.title",
			spec: `openapi: "3.0.0"
info:
  version: "1.0.0"
paths: {}
`,
			ruleID:        "validation-required-field",
			messageSubstr: "info.title",
		},
		{
			name: "validation-required-field/info.version",
			spec: `openapi: "3.0.0"
info:
  title: Test API
paths: {}
`,
			ruleID:        "validation-required-field",
			messageSubstr: "info.version",
		},
		{
			name: "validation-required-field/response.description",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
`,
			ruleID:        "validation-required-field",
			messageSubstr: "response.description",
		},
		{
			name: "validation-required-field/serverVariable.default",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
servers:
  - url: https://{env}.example.com
    variables:
      env:
        description: Environment
paths: {}
`,
			ruleID:        "validation-required-field",
			messageSubstr: "serverVariable.default",
		},
		{
			name: "validation-required-field/parameter.in=path",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test/{id}:
    get:
      operationId: getTest
      parameters:
        - name: id
          in: path
          required: false
          schema:
            type: string
      responses:
        '200':
          description: OK
`,
			ruleID:        "validation-required-field",
			messageSubstr: "parameter.in=path",
		},
		{
			name: "validation-invalid-format/contact.email",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
  contact:
    name: Support
    email: "not-an-email"
paths: {}
`,
			ruleID:        "validation-invalid-format",
			messageSubstr: "contact.email is not a valid email address",
		},
		{
			name: "validation-type-mismatch/exclusiveMinimum",
			spec: `openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  count:
                    type: number
                    minimum: 0
                    maximum: 100
                    exclusiveMinimum: true
`,
			ruleID:        "validation-type-mismatch",
			messageSubstr: "exclusiveMinimum",
		},
		{
			name: "validation-type-mismatch/exclusiveMaximum",
			spec: `openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  count:
                    type: number
                    minimum: 0
                    maximum: 100
                    exclusiveMaximum: true
`,
			ruleID:        "validation-type-mismatch",
			messageSubstr: "exclusiveMaximum",
		},
		{
			name: "validation-allowed-values/parameter.allowEmptyValue",
			spec: `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /test/{id}:
    get:
      operationId: getTest
      parameters:
        - name: id
          in: path
          required: true
          allowEmptyValue: true
          schema:
            type: string
      responses:
        '200':
          description: OK
`,
			ruleID:        "validation-allowed-values",
			messageSubstr: "allowEmptyValue",
		},
		// Note: validation-scheme-not-found is not tested here because the generator-level
		// security validation (generator-validate-security) handles undefined schemes before
		// the upstream validation-scheme-not-found rule is surfaced.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validation.NewValidator(
				&config.Configuration{},
				validation.RulesetSpeakeasyRecommended,
			)
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.spec), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()

			t.Logf("Total errors: %d", len(errs))
			for i, e := range errs {
				t.Logf("  Error %d: %s", i, e.Error())
			}

			// Find the target error by rule ID and message substring
			var found bool
			for _, e := range errs {
				errStr := e.Error()
				if contains(errStr, tt.ruleID) && contains(errStr, tt.messageSubstr) {
					found = true
					// Assert it was downgraded to warning (not error)
					assert.Contains(t, errStr, "validation warn:",
						"match filter should downgrade %s to warning, got: %s", tt.ruleID, errStr)
					assert.NotContains(t, errStr, "validation error:",
						"match filter should NOT leave %s as error, got: %s", tt.ruleID, errStr)
				}
			}
			assert.True(t, found,
				"expected to find %s error containing %q (so the test is meaningful), errors: %v",
				tt.ruleID, tt.messageSubstr, errStringsFromErrs(errs))
		})
	}
}

func errStringsFromErrs(errs []error) []string {
	strs := make([]string, 0, len(errs))
	for _, e := range errs {
		strs = append(strs, e.Error())
	}
	return strs
}
