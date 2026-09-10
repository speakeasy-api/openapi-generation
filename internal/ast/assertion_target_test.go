package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertionTargetFromExpression(t *testing.T) {
	tests := []struct {
		name           string
		conditionStr   string
		expectedTarget AssertionTarget
		expectError    bool
		errorContains  string
	}{
		{
			name:           "status code expression",
			conditionStr:   "$statusCode == 200",
			expectedTarget: AssertionTargetStatusCode,
			expectError:    false,
		},
		{
			name:           "status code expression with different operator",
			conditionStr:   "$statusCode != 404",
			expectedTarget: AssertionTargetStatusCode,
			expectError:    false,
		},
		{
			name:           "response body expression without JSON pointer",
			conditionStr:   "$response.body == 'OK'",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response body expression with root JSON pointer",
			conditionStr:   "$response.body#/ == 'value'",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response body expression with JSON pointer",
			conditionStr:   "$response.body#/status == 'complete'",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response body expression with deep JSON pointer",
			conditionStr:   "$response.body#/data/user/status == 'active'",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response body expression with array index",
			conditionStr:   "$response.body#/items/0/id == 123",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response body expression with boolean",
			conditionStr:   "$response.body#/ready == true",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response body expression with not equal",
			conditionStr:   "$response.body#/error != null",
			expectedTarget: AssertionTargetResponseBody,
			expectError:    false,
		},
		{
			name:           "response header expression - unsupported",
			conditionStr:   "$response.header.content-type == 'application/json'",
			expectedTarget: "",
			expectError:    true,
			errorContains:  "unsupported response target for assertion: header",
		},
		{
			name:           "request expression - unsupported",
			conditionStr:   "$request.body#/id == 123",
			expectedTarget: "",
			expectError:    true,
			errorContains:  "unsupported target for assertion:",
		},
		{
			name:           "url expression - unsupported",
			conditionStr:   "$url.path == '/api/v1/users'",
			expectedTarget: "",
			expectError:    true,
			errorContains:  "unsupported target for assertion:",
		},
		{
			name:           "method expression - unsupported",
			conditionStr:   "$method == 'GET'",
			expectedTarget: "",
			expectError:    true,
			errorContains:  "unsupported target for assertion:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a condition from the string
			condition := testCondition(t, tt.conditionStr)
			require.NotNil(t, condition, "failed to create test condition")
			require.NotNil(t, condition.Expression, "condition should have an expression")

			// Extract the expression from the condition
			result, err := AssertionTargetFromExpression(condition.Expression)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Equal(t, tt.expectedTarget, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTarget, result)
			}
		})
	}
}
