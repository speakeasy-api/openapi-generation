package validation_test

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/require"
)

// TestRulesetTracking_GenerationRuleset tests that the generation ruleset rules are tracked
func TestRulesetTracking_GenerationRuleset(t *testing.T) {
	// Create validator with generation ruleset
	v, err := validation.NewValidator(
		&config.Configuration{},
		validation.RulesetSpeakeasyGeneration,
	)
	require.NoError(t, err)

	// Create a simple valid OpenAPI spec
	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: testOp
      responses:
        '200':
          description: OK
`

	// Validate
	res := validateSpec(v, t.Context(), []byte(spec), "openapi.yaml", types.NewTargetFromTemplate("go"))
	require.NotNil(t, res)

	// Check that generation ruleset rules were used
	genRules := v.GetRulesets()[validation.RulesetSpeakeasyGeneration]
	require.NotEmpty(t, genRules, "Generation ruleset should have rules")

	// This should return true if any generation rule fired
	rulesUsed := v.AreRulesUsed(genRules)

	// Log for debugging
	t.Logf("Generation rules used: %v", rulesUsed)
	t.Logf("Total errors: %d", len(res.GetValidationErrors()))
	for i, err := range res.GetValidationErrors() {
		t.Logf("Error %d: %s", i, err.Error())
	}

	// Debug: check what rules were actually used
	t.Logf("Used rules count: %d", len(v.GetUsedRules()))
	for _, rule := range v.GetUsedRules() {
		t.Logf("  Used rule: %s", rule.ID())
	}

	// Debug: check what rules are in the generation ruleset
	t.Logf("Generation ruleset count: %d", len(genRules))
	for i, rule := range genRules {
		if i >= 5 {
			break // Just show first 5
		}
		t.Logf("  Gen rule: %s", rule.ID())
	}

	// This test checks if ANY generation rules fired
	// With a minimal valid spec, we might not trigger all rules, but we should trigger some
	// If this fails, it means no generation rules fired at all, which is the bug
	require.True(t, rulesUsed, "At least one generation ruleset rule should have been used during validation")
}
