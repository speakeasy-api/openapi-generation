package tfmockserver

import (
	"testing"
)

func TestAssertRequestBodyField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        map[string]any
		field       string
		expected    any
		expectError bool
	}{
		// Field exists and matches expected value
		{
			name:        "string value matches",
			body:        map[string]any{"name": "test"},
			field:       "name",
			expected:    "test",
			expectError: false,
		},
		{
			name:        "integer value matches",
			body:        map[string]any{"count": 42},
			field:       "count",
			expected:    42,
			expectError: false,
		},
		{
			name:        "float value matches",
			body:        map[string]any{"price": 19.99},
			field:       "price",
			expected:    19.99,
			expectError: false,
		},
		{
			name:        "boolean true matches",
			body:        map[string]any{"enabled": true},
			field:       "enabled",
			expected:    true,
			expectError: false,
		},
		{
			name:        "boolean false matches",
			body:        map[string]any{"enabled": false},
			field:       "enabled",
			expected:    false,
			expectError: false,
		},
		{
			name:        "null value matches nil expected",
			body:        map[string]any{"optional": nil},
			field:       "optional",
			expected:    nil,
			expectError: false,
		},

		// Field missing
		{
			name:        "field missing from body",
			body:        map[string]any{"other_field": "value"},
			field:       "missing_field",
			expected:    "expected",
			expectError: true,
		},
		{
			name:        "empty body",
			body:        map[string]any{},
			field:       "any_field",
			expected:    "any_value",
			expectError: true,
		},
		{
			name:        "nil body",
			body:        nil,
			field:       "any_field",
			expected:    "any_value",
			expectError: true,
		},

		// Null expected but actual is not null
		{
			name:        "expected null but got string",
			body:        map[string]any{"field": "actual_value"},
			field:       "field",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "expected null but got integer",
			body:        map[string]any{"field": 123},
			field:       "field",
			expected:    nil,
			expectError: true,
		},

		// Value mismatch
		{
			name:        "string mismatch",
			body:        map[string]any{"name": "actual"},
			field:       "name",
			expected:    "expected",
			expectError: true,
		},
		{
			name:        "integer mismatch",
			body:        map[string]any{"count": 10},
			field:       "count",
			expected:    20,
			expectError: true,
		},
		{
			name:        "boolean mismatch",
			body:        map[string]any{"enabled": true},
			field:       "enabled",
			expected:    false,
			expectError: true,
		},

		// Array matching
		{
			name:        "empty array matches",
			body:        map[string]any{"items": []interface{}{}},
			field:       "items",
			expected:    []any{},
			expectError: false,
		},
		{
			name:        "string array matches",
			body:        map[string]any{"items": []interface{}{"a", "b", "c"}},
			field:       "items",
			expected:    []any{"a", "b", "c"},
			expectError: false,
		},
		{
			name:        "integer array matches",
			body:        map[string]any{"numbers": []interface{}{1, 2, 3}},
			field:       "numbers",
			expected:    []any{1, 2, 3},
			expectError: false,
		},
		{
			name:        "array length mismatch - fewer actual",
			body:        map[string]any{"items": []interface{}{"a", "b"}},
			field:       "items",
			expected:    []any{"a", "b", "c"},
			expectError: true,
		},
		{
			name:        "array length mismatch - more actual",
			body:        map[string]any{"items": []interface{}{"a", "b", "c", "d"}},
			field:       "items",
			expected:    []any{"a", "b", "c"},
			expectError: true,
		},
		{
			name:        "array element mismatch",
			body:        map[string]any{"items": []interface{}{"a", "x", "c"}},
			field:       "items",
			expected:    []any{"a", "b", "c"},
			expectError: true,
		},
		{
			name:        "expected array but got string",
			body:        map[string]any{"items": "not an array"},
			field:       "items",
			expected:    []any{"a", "b"},
			expectError: true,
		},
		{
			name:        "expected array but got nil",
			body:        map[string]any{"items": nil},
			field:       "items",
			expected:    []any{"a", "b"},
			expectError: true,
		},

		// Mixed types - fmt.Sprintf comparison behavior
		{
			name:        "int vs float64 same value",
			body:        map[string]any{"num": float64(42)},
			field:       "num",
			expected:    42,
			expectError: false,
		},
		{
			name:        "string number vs int",
			body:        map[string]any{"num": "42"},
			field:       "num",
			expected:    42,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := assertRequestBodyField(tt.body, tt.field, tt.expected)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestAssertRequestBodyFieldOmitted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        map[string]any
		field       string
		expectError bool
	}{
		{
			name:        "field is omitted",
			body:        map[string]any{"other": "value"},
			field:       "missing",
			expectError: false,
		},
		{
			name:        "field is omitted from empty body",
			body:        map[string]any{},
			field:       "missing",
			expectError: false,
		},
		{
			name:        "field is omitted from nil body",
			body:        nil,
			field:       "missing",
			expectError: false,
		},
		{
			name:        "field is present with string value",
			body:        map[string]any{"field": "value"},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with null value",
			body:        map[string]any{"field": nil},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with zero value",
			body:        map[string]any{"field": 0},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with empty string",
			body:        map[string]any{"field": ""},
			field:       "field",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := assertRequestBodyFieldOmitted(tt.body, tt.field)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestAssertRequestBodyFieldNull(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        map[string]any
		field       string
		expectError bool
	}{
		{
			name:        "field is present and null",
			body:        map[string]any{"field": nil},
			field:       "field",
			expectError: false,
		},
		{
			name:        "field is omitted",
			body:        map[string]any{"other": "value"},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is omitted from empty body",
			body:        map[string]any{},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is omitted from nil body",
			body:        nil,
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with string value",
			body:        map[string]any{"field": "value"},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with zero int",
			body:        map[string]any{"field": 0},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with empty string",
			body:        map[string]any{"field": ""},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is present with false",
			body:        map[string]any{"field": false},
			field:       "field",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := assertRequestBodyFieldNull(tt.body, tt.field)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestAssertRequestBodyFieldEmptyArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        map[string]any
		field       string
		expectError bool
	}{
		{
			name:        "field is present with empty array",
			body:        map[string]any{"field": []interface{}{}},
			field:       "field",
			expectError: false,
		},
		{
			name:        "field is omitted",
			body:        map[string]any{"other": "value"},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is omitted from empty body",
			body:        map[string]any{},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is omitted from nil body",
			body:        nil,
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is null",
			body:        map[string]any{"field": nil},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is non-empty array",
			body:        map[string]any{"field": []interface{}{"a", "b"}},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is string",
			body:        map[string]any{"field": "not an array"},
			field:       "field",
			expectError: true,
		},
		{
			name:        "field is integer",
			body:        map[string]any{"field": 123},
			field:       "field",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := assertRequestBodyFieldEmptyArray(tt.body, tt.field)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestAssertRequestBodyFieldArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        map[string]any
		field       string
		expected    []any
		expectError bool
	}{
		{
			name:        "empty array matches",
			body:        map[string]any{"field": []interface{}{}},
			field:       "field",
			expected:    []any{},
			expectError: false,
		},
		{
			name:        "string array matches",
			body:        map[string]any{"field": []interface{}{"a", "b", "c"}},
			field:       "field",
			expected:    []any{"a", "b", "c"},
			expectError: false,
		},
		{
			name:        "integer array matches",
			body:        map[string]any{"field": []interface{}{1, 2, 3}},
			field:       "field",
			expected:    []any{1, 2, 3},
			expectError: false,
		},
		{
			name:        "mixed type array matches",
			body:        map[string]any{"field": []interface{}{"a", 1, true}},
			field:       "field",
			expected:    []any{"a", 1, true},
			expectError: false,
		},
		{
			name:        "field is omitted",
			body:        map[string]any{"other": "value"},
			field:       "field",
			expected:    []any{"a"},
			expectError: true,
		},
		{
			name:        "field is omitted from empty body",
			body:        map[string]any{},
			field:       "field",
			expected:    []any{"a"},
			expectError: true,
		},
		{
			name:        "field is omitted from nil body",
			body:        nil,
			field:       "field",
			expected:    []any{"a"},
			expectError: true,
		},
		{
			name:        "field is null",
			body:        map[string]any{"field": nil},
			field:       "field",
			expected:    []any{"a"},
			expectError: true,
		},
		{
			name:        "field is string",
			body:        map[string]any{"field": "not an array"},
			field:       "field",
			expected:    []any{"a"},
			expectError: true,
		},
		{
			name:        "array length mismatch - fewer actual",
			body:        map[string]any{"field": []interface{}{"a"}},
			field:       "field",
			expected:    []any{"a", "b"},
			expectError: true,
		},
		{
			name:        "array length mismatch - more actual",
			body:        map[string]any{"field": []interface{}{"a", "b", "c"}},
			field:       "field",
			expected:    []any{"a", "b"},
			expectError: true,
		},
		{
			name:        "array element mismatch",
			body:        map[string]any{"field": []interface{}{"a", "x", "c"}},
			field:       "field",
			expected:    []any{"a", "b", "c"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := assertRequestBodyFieldArray(tt.body, tt.field, tt.expected)

			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}
