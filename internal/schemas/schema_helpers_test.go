package schemas

import (
	"testing"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/values"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// testValue creates a values.Value for testing.
func testValue(v string) values.Value {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: v}
}

// testValueWithTag creates a values.Value with a specific YAML tag for testing.
func testValueWithTag(v string, tag string) values.Value {
	return &yaml.Node{Kind: yaml.ScalarNode, Value: v, Tag: tag}
}

// testArrayValue creates a values.Value representing a YAML array for testing.
func testArrayValue() values.Value {
	return &yaml.Node{Kind: yaml.SequenceNode}
}

// testObjectValue creates a values.Value representing a YAML object for testing.
func testObjectValue() values.Value {
	return &yaml.Node{Kind: yaml.MappingNode}
}

func TestLeastRestrictiveMax(t *testing.T) {
	t.Parallel()

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			values   []*float64
			expected *float64
		}{
			{
				name:     "all nil",
				values:   []*float64{nil, nil, nil},
				expected: nil,
			},
			{
				name:     "empty slice",
				values:   []*float64{},
				expected: nil,
			},
			{
				name:     "single value",
				values:   []*float64{pointer.From(5.0)},
				expected: pointer.From(5.0),
			},
			{
				name:     "first is nil",
				values:   []*float64{nil, pointer.From(5.0), pointer.From(10.0)},
				expected: nil,
			},
			{
				name:     "middle is nil",
				values:   []*float64{pointer.From(5.0), nil, pointer.From(10.0)},
				expected: nil,
			},
			{
				name:     "last is nil",
				values:   []*float64{pointer.From(5.0), pointer.From(10.0), nil},
				expected: nil,
			},
			{
				name:     "returns largest",
				values:   []*float64{pointer.From(5.0), pointer.From(10.0), pointer.From(3.0)},
				expected: pointer.From(10.0),
			},
			{
				name:     "all same value",
				values:   []*float64{pointer.From(5.0), pointer.From(5.0), pointer.From(5.0)},
				expected: pointer.From(5.0),
			},
			{
				name:     "negative values",
				values:   []*float64{pointer.From(-5.0), pointer.From(-10.0), pointer.From(-3.0)},
				expected: pointer.From(-3.0),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result := leastRestrictiveMax(tt.values)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.InDelta(t, *tt.expected, *result, 0.0001)
				}
			})
		}
	})

	t.Run("int64", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			values   []*int64
			expected *int64
		}{
			{
				name:     "all nil",
				values:   []*int64{nil, nil, nil},
				expected: nil,
			},
			{
				name:     "empty slice",
				values:   []*int64{},
				expected: nil,
			},
			{
				name:     "single value",
				values:   []*int64{pointer.From(int64(5))},
				expected: pointer.From(int64(5)),
			},
			{
				name:     "first is nil",
				values:   []*int64{nil, pointer.From(int64(5)), pointer.From(int64(10))},
				expected: nil,
			},
			{
				name:     "returns largest",
				values:   []*int64{pointer.From(int64(5)), pointer.From(int64(10)), pointer.From(int64(3))},
				expected: pointer.From(int64(10)),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result := leastRestrictiveMax(tt.values)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, *tt.expected, *result)
				}
			})
		}
	})
}

func TestLeastRestrictiveMin(t *testing.T) {
	t.Parallel()

	t.Run("float64", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			values   []*float64
			expected *float64
		}{
			{
				name:     "all nil",
				values:   []*float64{nil, nil, nil},
				expected: nil,
			},
			{
				name:     "empty slice",
				values:   []*float64{},
				expected: nil,
			},
			{
				name:     "single value",
				values:   []*float64{pointer.From(5.0)},
				expected: pointer.From(5.0),
			},
			{
				name:     "first is nil",
				values:   []*float64{nil, pointer.From(5.0), pointer.From(10.0)},
				expected: nil,
			},
			{
				name:     "middle is nil",
				values:   []*float64{pointer.From(5.0), nil, pointer.From(10.0)},
				expected: nil,
			},
			{
				name:     "last is nil",
				values:   []*float64{pointer.From(5.0), pointer.From(10.0), nil},
				expected: nil,
			},
			{
				name:     "returns smallest",
				values:   []*float64{pointer.From(5.0), pointer.From(10.0), pointer.From(3.0)},
				expected: pointer.From(3.0),
			},
			{
				name:     "all same value",
				values:   []*float64{pointer.From(5.0), pointer.From(5.0), pointer.From(5.0)},
				expected: pointer.From(5.0),
			},
			{
				name:     "negative values",
				values:   []*float64{pointer.From(-5.0), pointer.From(-10.0), pointer.From(-3.0)},
				expected: pointer.From(-10.0),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result := leastRestrictiveMin(tt.values)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.InDelta(t, *tt.expected, *result, 0.0001)
				}
			})
		}
	})

	t.Run("int64", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			values   []*int64
			expected *int64
		}{
			{
				name:     "all nil",
				values:   []*int64{nil, nil, nil},
				expected: nil,
			},
			{
				name:     "empty slice",
				values:   []*int64{},
				expected: nil,
			},
			{
				name:     "single value",
				values:   []*int64{pointer.From(int64(5))},
				expected: pointer.From(int64(5)),
			},
			{
				name:     "first is nil",
				values:   []*int64{nil, pointer.From(int64(5)), pointer.From(int64(10))},
				expected: nil,
			},
			{
				name:     "returns smallest",
				values:   []*int64{pointer.From(int64(5)), pointer.From(int64(10)), pointer.From(int64(3))},
				expected: pointer.From(int64(3)),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				result := leastRestrictiveMin(tt.values)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, *tt.expected, *result)
				}
			})
		}
	})
}

func TestMergeDescriptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		parentDescription *string
		descriptions      []*string
		expected          *string
	}{
		{
			name:              "parent nil and empty slice",
			parentDescription: nil,
			descriptions:      []*string{},
			expected:          nil,
		},
		{
			name:              "parent nil and all nil descriptions",
			parentDescription: nil,
			descriptions:      []*string{nil, nil, nil},
			expected:          nil,
		},
		{
			name:              "parent description takes precedence",
			parentDescription: pointer.From("parent description"),
			descriptions:      []*string{pointer.From("first"), pointer.From("second")},
			expected:          pointer.From("parent description"),
		},
		{
			name:              "parent description takes precedence over nil descriptions",
			parentDescription: pointer.From("parent description"),
			descriptions:      []*string{nil, nil},
			expected:          pointer.From("parent description"),
		},
		{
			name:              "first non-nil description used when parent nil",
			parentDescription: nil,
			descriptions:      []*string{pointer.From("first"), pointer.From("second")},
			expected:          pointer.From("first"),
		},
		{
			name:              "skips nil to find first non-nil description",
			parentDescription: nil,
			descriptions:      []*string{nil, pointer.From("second"), pointer.From("third")},
			expected:          pointer.From("second"),
		},
		{
			name:              "finds description at end",
			parentDescription: nil,
			descriptions:      []*string{nil, nil, pointer.From("third")},
			expected:          pointer.From("third"),
		},
		{
			name:              "single description",
			parentDescription: nil,
			descriptions:      []*string{pointer.From("only")},
			expected:          pointer.From("only"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mergeDescriptions(tt.parentDescription, tt.descriptions)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestMergePatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		patterns []*string
		expected *string
	}{
		{
			name:     "empty slice",
			patterns: []*string{},
			expected: nil,
		},
		{
			name:     "all nil",
			patterns: []*string{nil, nil, nil},
			expected: nil,
		},
		{
			name:     "single pattern",
			patterns: []*string{pointer.From("^[a-z]+$")},
			expected: pointer.From("^[a-z]+$"),
		},
		{
			name:     "first is nil",
			patterns: []*string{nil, pointer.From("^[a-z]+$"), pointer.From("^[A-Z]+$")},
			expected: nil,
		},
		{
			name:     "middle is nil",
			patterns: []*string{pointer.From("^[a-z]+$"), nil, pointer.From("^[A-Z]+$")},
			expected: nil,
		},
		{
			name:     "last is nil",
			patterns: []*string{pointer.From("^[a-z]+$"), pointer.From("^[A-Z]+$"), nil},
			expected: nil,
		},
		{
			name:     "two patterns",
			patterns: []*string{pointer.From("^[a-z]+$"), pointer.From("^[A-Z]+$")},
			expected: pointer.From("(^[a-z]+$|^[A-Z]+$)"),
		},
		{
			name:     "three patterns",
			patterns: []*string{pointer.From("^[a-z]+$"), pointer.From("^[A-Z]+$"), pointer.From("^[0-9]+$")},
			expected: pointer.From("(^[a-z]+$|^[A-Z]+$|^[0-9]+$)"),
		},
		{
			name:     "duplicate patterns at start deduplicated",
			patterns: []*string{pointer.From("^test$"), pointer.From("^test$"), pointer.From("^other$")},
			expected: pointer.From("(^test$|^other$)"),
		},
		{
			name:     "duplicate patterns at end deduplicated",
			patterns: []*string{pointer.From("^first$"), pointer.From("^test$"), pointer.From("^test$")},
			expected: pointer.From("(^first$|^test$)"),
		},
		{
			name:     "duplicate patterns in middle deduplicated",
			patterns: []*string{pointer.From("^first$"), pointer.From("^test$"), pointer.From("^test$"), pointer.From("^last$")},
			expected: pointer.From("(^first$|^test$|^last$)"),
		},
		{
			name:     "all duplicate patterns returns single",
			patterns: []*string{pointer.From("^test$"), pointer.From("^test$")},
			expected: pointer.From("^test$"),
		},
		{
			name:     "multiple duplicates across input",
			patterns: []*string{pointer.From("^A$"), pointer.From("^B$"), pointer.From("^A$"), pointer.From("^C$"), pointer.From("^B$")},
			expected: pointer.From("(^A$|^B$|^C$)"),
		},
		{
			name:     "preserves order of first occurrence",
			patterns: []*string{pointer.From("^C$"), pointer.From("^A$"), pointer.From("^B$"), pointer.From("^A$"), pointer.From("^C$")},
			expected: pointer.From("(^C$|^A$|^B$)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mergePatterns(tt.patterns)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestMergeExamples(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		parentExample    values.Value
		parentExamples   []values.Value
		subExamples      []subSchemaExamples
		expectedExample  values.Value
		expectedExamples []values.Value
	}{
		{
			name:             "all empty",
			parentExample:    nil,
			parentExamples:   nil,
			subExamples:      []subSchemaExamples{},
			expectedExample:  nil,
			expectedExamples: nil,
		},
		{
			name:             "parent example takes precedence over everything",
			parentExample:    testValue("parent"),
			parentExamples:   []values.Value{testValue("parent1"), testValue("parent2")},
			subExamples:      []subSchemaExamples{{Example: testValue("sub")}},
			expectedExample:  testValue("parent"),
			expectedExamples: nil,
		},
		{
			name:             "parent examples takes precedence over sub-schemas",
			parentExample:    nil,
			parentExamples:   []values.Value{testValue("parent1"), testValue("parent2")},
			subExamples:      []subSchemaExamples{{Example: testValue("sub")}},
			expectedExample:  nil,
			expectedExamples: []values.Value{testValue("parent1"), testValue("parent2")},
		},
		{
			name:           "first sub-schema example used when parent empty",
			parentExample:  nil,
			parentExamples: nil,
			subExamples: []subSchemaExamples{
				{Example: testValue("first")},
				{Example: testValue("second")},
			},
			expectedExample:  testValue("first"),
			expectedExamples: nil,
		},
		{
			name:           "first sub-schema examples used when parent empty and no example",
			parentExample:  nil,
			parentExamples: nil,
			subExamples: []subSchemaExamples{
				{Examples: []values.Value{testValue("first1"), testValue("first2")}},
				{Example: testValue("second")},
			},
			expectedExample:  nil,
			expectedExamples: []values.Value{testValue("first1"), testValue("first2")},
		},
		{
			name:           "skips empty sub-schemas to find first example",
			parentExample:  nil,
			parentExamples: nil,
			subExamples: []subSchemaExamples{
				{},
				{Example: testValue("second")},
				{Example: testValue("third")},
			},
			expectedExample:  testValue("second"),
			expectedExamples: nil,
		},
		{
			name:           "skips empty sub-schemas to find first examples",
			parentExample:  nil,
			parentExamples: nil,
			subExamples: []subSchemaExamples{
				{},
				{},
				{Examples: []values.Value{testValue("third1"), testValue("third2")}},
			},
			expectedExample:  nil,
			expectedExamples: []values.Value{testValue("third1"), testValue("third2")},
		},
		{
			name:           "example preferred over examples in same sub-schema",
			parentExample:  nil,
			parentExamples: nil,
			subExamples: []subSchemaExamples{
				{
					Example:  testValue("single"),
					Examples: []values.Value{testValue("multi1"), testValue("multi2")},
				},
			},
			expectedExample:  testValue("single"),
			expectedExamples: nil,
		},
		{
			name:             "all sub-schemas empty returns nil",
			parentExample:    nil,
			parentExamples:   nil,
			subExamples:      []subSchemaExamples{{}, {}, {}},
			expectedExample:  nil,
			expectedExamples: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resultExample, resultExamples := mergeExamples(tt.parentExample, tt.parentExamples, tt.subExamples)

			if tt.expectedExample == nil {
				assert.Nil(t, resultExample)
			} else {
				assert.NotNil(t, resultExample)
				assert.Equal(t, tt.expectedExample.Value, resultExample.Value)
			}

			if tt.expectedExamples == nil {
				assert.Nil(t, resultExamples)
			} else {
				assert.Len(t, resultExamples, len(tt.expectedExamples))
				for i, expected := range tt.expectedExamples {
					assert.Equal(t, expected.Value, resultExamples[i].Value)
				}
			}
		})
	}
}

func TestExampleMatchesSchemaType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		example     values.Value
		schemaTypes []oas3.SchemaType
		expected    bool
	}{
		{
			name:        "nil example always matches",
			example:     nil,
			schemaTypes: []oas3.SchemaType{"string"},
			expected:    true,
		},
		{
			name:        "empty schema types always matches",
			example:     testValueWithTag("test", "!!str"),
			schemaTypes: []oas3.SchemaType{},
			expected:    true,
		},
		{
			name:        "nil example with empty schema types",
			example:     nil,
			schemaTypes: []oas3.SchemaType{},
			expected:    true,
		},
		{
			name:        "string example matches string schema",
			example:     testValueWithTag("hello", "!!str"),
			schemaTypes: []oas3.SchemaType{"string"},
			expected:    true,
		},
		{
			name:        "string example does not match number schema",
			example:     testValueWithTag("hello", "!!str"),
			schemaTypes: []oas3.SchemaType{"number"},
			expected:    false,
		},
		{
			name:        "string example does not match integer schema",
			example:     testValueWithTag("hello", "!!str"),
			schemaTypes: []oas3.SchemaType{"integer"},
			expected:    false,
		},
		{
			name:        "integer example matches integer schema",
			example:     testValueWithTag("42", "!!int"),
			schemaTypes: []oas3.SchemaType{"integer"},
			expected:    true,
		},
		{
			name:        "integer example matches number schema",
			example:     testValueWithTag("42", "!!int"),
			schemaTypes: []oas3.SchemaType{"number"},
			expected:    true,
		},
		{
			name:        "integer example does not match string schema",
			example:     testValueWithTag("42", "!!int"),
			schemaTypes: []oas3.SchemaType{"string"},
			expected:    false,
		},
		{
			name:        "float example matches number schema",
			example:     testValueWithTag("3.14", "!!float"),
			schemaTypes: []oas3.SchemaType{"number"},
			expected:    true,
		},
		{
			name:        "float example does not match integer schema",
			example:     testValueWithTag("3.14", "!!float"),
			schemaTypes: []oas3.SchemaType{"integer"},
			expected:    false,
		},
		{
			name:        "float example does not match string schema",
			example:     testValueWithTag("3.14", "!!float"),
			schemaTypes: []oas3.SchemaType{"string"},
			expected:    false,
		},
		{
			name:        "boolean example matches boolean schema",
			example:     testValueWithTag("true", "!!bool"),
			schemaTypes: []oas3.SchemaType{"boolean"},
			expected:    true,
		},
		{
			name:        "boolean example does not match string schema",
			example:     testValueWithTag("true", "!!bool"),
			schemaTypes: []oas3.SchemaType{"string"},
			expected:    false,
		},
		{
			name:        "array example matches array schema",
			example:     testArrayValue(),
			schemaTypes: []oas3.SchemaType{"array"},
			expected:    true,
		},
		{
			name:        "array example does not match object schema",
			example:     testArrayValue(),
			schemaTypes: []oas3.SchemaType{"object"},
			expected:    false,
		},
		{
			name:        "object example matches object schema",
			example:     testObjectValue(),
			schemaTypes: []oas3.SchemaType{"object"},
			expected:    true,
		},
		{
			name:        "object example does not match array schema",
			example:     testObjectValue(),
			schemaTypes: []oas3.SchemaType{"array"},
			expected:    false,
		},
		{
			name:        "null example matches null schema",
			example:     testValueWithTag("null", "!!null"),
			schemaTypes: []oas3.SchemaType{"null"},
			expected:    true,
		},
		{
			name:        "null example does not match string schema",
			example:     testValueWithTag("null", "!!null"),
			schemaTypes: []oas3.SchemaType{"string"},
			expected:    false,
		},
		{
			name:        "matches any of multiple schema types",
			example:     testValueWithTag("hello", "!!str"),
			schemaTypes: []oas3.SchemaType{"number", "string"},
			expected:    true,
		},
		{
			name:        "does not match if none of multiple schema types match",
			example:     testValueWithTag("true", "!!bool"),
			schemaTypes: []oas3.SchemaType{"number", "string"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := exampleMatchesSchemaType(tt.example, tt.schemaTypes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterExamplesBySchemaType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		examples       []values.Value
		schemaTypes    []oas3.SchemaType
		expectedValues []string
	}{
		{
			name:           "nil examples returns nil",
			examples:       nil,
			schemaTypes:    []oas3.SchemaType{"string"},
			expectedValues: nil,
		},
		{
			name:           "empty examples returns empty",
			examples:       []values.Value{},
			schemaTypes:    []oas3.SchemaType{"string"},
			expectedValues: []string{},
		},
		{
			name:           "empty schema types returns all examples",
			examples:       []values.Value{testValueWithTag("a", "!!str"), testValueWithTag("1", "!!int")},
			schemaTypes:    []oas3.SchemaType{},
			expectedValues: []string{"a", "1"},
		},
		{
			name: "filters to only matching types",
			examples: []values.Value{
				testValueWithTag("hello", "!!str"),
				testValueWithTag("42", "!!int"),
				testValueWithTag("world", "!!str"),
			},
			schemaTypes:    []oas3.SchemaType{"string"},
			expectedValues: []string{"hello", "world"},
		},
		{
			name: "returns nil when no examples match",
			examples: []values.Value{
				testValueWithTag("hello", "!!str"),
				testValueWithTag("world", "!!str"),
			},
			schemaTypes:    []oas3.SchemaType{"integer"},
			expectedValues: nil,
		},
		{
			name: "filters with multiple allowed schema types",
			examples: []values.Value{
				testValueWithTag("hello", "!!str"),
				testValueWithTag("42", "!!int"),
				testValueWithTag("true", "!!bool"),
				testValueWithTag("3.14", "!!float"),
			},
			schemaTypes:    []oas3.SchemaType{"number"},
			expectedValues: []string{"42", "3.14"},
		},
		{
			name: "all examples match",
			examples: []values.Value{
				testValueWithTag("a", "!!str"),
				testValueWithTag("b", "!!str"),
			},
			schemaTypes:    []oas3.SchemaType{"string"},
			expectedValues: []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := filterExamplesBySchemaType(tt.examples, tt.schemaTypes)

			if tt.expectedValues == nil {
				assert.Nil(t, result)
			} else {
				assert.Len(t, result, len(tt.expectedValues))
				for i, expected := range tt.expectedValues {
					assert.Equal(t, expected, result[i].Value)
				}
			}
		})
	}
}
