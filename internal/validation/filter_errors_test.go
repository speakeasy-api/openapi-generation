package validation_test

import (
	"errors"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	openapiValidation "github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestFilterErrors tests the FilterErrors method without any lint.yaml configuration
func TestFilterErrors_WithDefaultFilters(t *testing.T) {
	// Create validator with default filters
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
	)
	require.NoError(t, err)

	// Create some validation errors with different severities
	errors := []error{
		&openapiValidation.Error{
			Rule:            "validation-required-field",
			Severity:        openapiValidation.SeverityError,
			UnderlyingError: errors.New("info.title is required"),
			Node:            &yaml.Node{Line: 1},
		},
		&openapiValidation.Error{
			Rule:            "validation-required-field",
			Severity:        openapiValidation.SeverityError,
			UnderlyingError: errors.New("response.description is required"),
			Node:            &yaml.Node{Line: 10},
		},
		&openapiValidation.Error{
			Rule:            "generator-validate-enums",
			Severity:        openapiValidation.SeverityError,
			UnderlyingError: errors.New("enum values collide"),
			Node:            &yaml.Node{Line: 20},
		},
	}

	// Apply filters
	warns, errs := v.FilterErrors(errors)

	// The first two should be downgraded to warnings by default filters
	assert.Len(t, warns, 2, "Expected 2 warnings (info.title and response.description)")
	assert.Len(t, errs, 1, "Expected 1 error (enum validation)")

	// Verify warning messages
	assert.Contains(t, warns[0].Error(), "info.title is required")
	assert.Contains(t, warns[1].Error(), "response.description is required")

	// Verify error message
	assert.Contains(t, errs[0].Error(), "enum values collide")
}

// TestFilterErrors_WithSeverityDowngrade tests that FilterErrors respects severity changes
func TestFilterErrors_WithSeverityDowngrade(t *testing.T) {
	// Create validator
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
	)
	require.NoError(t, err)

	// Create errors with different severities
	errors := []error{
		&openapiValidation.Error{
			Rule:            "test-rule",
			Severity:        openapiValidation.SeverityError,
			UnderlyingError: errors.New("this is an error"),
			Node:            &yaml.Node{Line: 1},
		},
		&openapiValidation.Error{
			Rule:            "test-rule",
			Severity:        openapiValidation.SeverityWarning,
			UnderlyingError: errors.New("this is a warning"),
			Node:            &yaml.Node{Line: 2},
		},
		&openapiValidation.Error{
			Rule:            "test-rule",
			Severity:        openapiValidation.SeverityHint,
			UnderlyingError: errors.New("this is a hint"),
			Node:            &yaml.Node{Line: 3},
		},
	}

	// Apply filters
	warns, errs := v.FilterErrors(errors)

	// Warning and hint should be in warns, error should be in errs
	assert.Len(t, warns, 2, "Expected 2 warnings (warning + hint)")
	assert.Len(t, errs, 1, "Expected 1 error")

	assert.Contains(t, errs[0].Error(), "this is an error")
	assert.Contains(t, warns[0].Error(), "this is a warning")
	assert.Contains(t, warns[1].Error(), "this is a hint")
}

// TestFilterErrors_WithNonOpenAPIErrors tests that non-openapi validation errors are passed through
func TestFilterErrors_WithNonOpenAPIErrors(t *testing.T) {
	// Create validator
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
	)
	require.NoError(t, err)

	// Mix of openapi validation errors and regular errors
	errors := []error{
		&openapiValidation.Error{
			Rule:            "validation-required-field",
			Severity:        openapiValidation.SeverityError,
			UnderlyingError: errors.New("info.title is required"),
			Node:            &yaml.Node{Line: 1},
		},
		errors.New("regular error that is not from openapi validation"),
	}

	// Apply filters
	warns, errs := v.FilterErrors(errors)

	// info.title should be downgraded to warning
	// regular error should remain as error
	assert.Len(t, warns, 1, "Expected 1 warning")
	assert.Len(t, errs, 1, "Expected 1 error")

	assert.Contains(t, warns[0].Error(), "info.title is required")
	assert.Contains(t, errs[0].Error(), "regular error that is not from openapi validation")
}

// TestFilterErrors_EmptyList tests FilterErrors with an empty error list
func TestFilterErrors_EmptyList(t *testing.T) {
	// Create validator
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyRecommended,
	)
	require.NoError(t, err)

	// Apply filters to empty list
	warns, errs := v.FilterErrors([]error{})

	assert.Empty(t, warns, "Expected 0 warnings")
	assert.Empty(t, errs, "Expected 0 errors")
}
