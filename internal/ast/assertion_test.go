package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
)

// testCondition creates a criterion.Condition using the same method as UnmarshalYAML
// This ensures the unexported rawCondition field is properly set
func testCondition(t *testing.T, conditionStr string) *criterion.Condition {
	t.Helper()

	c := &criterion.Criterion{
		Condition: conditionStr,
	}

	condition, err := c.GetCondition()

	if err != nil {
		t.Fatalf("failed to create test condition from '%s': %v", conditionStr, err)
	}

	return condition
}
