package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertionTypeFromArazzoOperator(t *testing.T) {
	tests := []struct {
		name        string
		operator    criterion.Operator
		expected    AssertionType
		expectError bool
		errorMsg    string
	}{
		{
			name:        "equal operator",
			operator:    criterion.OperatorEQ,
			expected:    AssertionTypeEqual,
			expectError: false,
		},
		{
			name:        "not equal operator",
			operator:    criterion.OperatorNE,
			expected:    AssertionTypeNotEqual,
			expectError: false,
		},
		{
			name:        "less than operator - unsupported",
			operator:    criterion.OperatorLT,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: <",
		},
		{
			name:        "less than or equal operator - unsupported",
			operator:    criterion.OperatorLTE,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: <=",
		},
		{
			name:        "greater than operator - unsupported",
			operator:    criterion.OperatorGT,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: >",
		},
		{
			name:        "greater than or equal operator - unsupported",
			operator:    criterion.OperatorGTE,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: >=",
		},
		{
			name:        "not operator - unsupported",
			operator:    criterion.OperatorNot,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: !",
		},
		{
			name:        "and operator - unsupported",
			operator:    criterion.OperatorAnd,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: &&",
		},
		{
			name:        "or operator - unsupported",
			operator:    criterion.OperatorOr,
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: ||",
		},
		{
			name:        "empty operator - unsupported",
			operator:    criterion.Operator(""),
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: ",
		},
		{
			name:        "custom operator - unsupported",
			operator:    criterion.Operator("custom"),
			expected:    "",
			expectError: true,
			errorMsg:    "unsupported operator for assertion type: custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AssertionTypeFromArazzoOperator(tt.operator)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Equal(t, tt.expected, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
