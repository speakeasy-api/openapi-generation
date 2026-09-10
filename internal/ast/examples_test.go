package ast

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestExample_Clone(t *testing.T) {
	t.Parallel()

	yamlNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "example value",
	}

	original := NewExample("exampleName", "example description", yamlNode)
	cloned := original.Clone()

	require.NotNil(t, cloned)
	assert.NotSame(t, original, cloned)
	assert.Equal(t, original.Name(), cloned.Name())
	assert.Equal(t, original.Description, cloned.Description)
	if original.Value != nil {
		assert.NotSame(t, original.Value, cloned.Value)
		assert.Equal(t, original.Value.Value, cloned.Value.Value)
	}
}

func TestExamples_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		examples Examples
		test     func(t *testing.T, original, cloned Examples)
	}{
		{
			name:     "nil Examples",
			examples: nil,
			test: func(t *testing.T, original, cloned Examples) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name:     "empty Examples",
			examples: Examples{},
			test: func(t *testing.T, original, cloned Examples) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				assert.Empty(t, cloned)
			},
		},
		{
			name: "Examples with single example",
			examples: Examples{
				NewExample("example1", "description 1", &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value1",
				}),
			},
			test: func(t *testing.T, original, cloned Examples) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 1)

				assert.NotSame(t, original[0], cloned[0])
				assert.Equal(t, original[0].Name(), cloned[0].Name())
				assert.Equal(t, original[0].Description, cloned[0].Description)
				if original[0].Value != nil {
					assert.NotSame(t, original[0].Value, cloned[0].Value)
					assert.Equal(t, original[0].Value.Value, cloned[0].Value.Value)
				}
			},
		},
		{
			name: "Examples with multiple examples",
			examples: Examples{
				NewExample("example1", "description 1", &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "value1",
				}),
				NewExample("example2", "description 2", &yaml.Node{
					Kind:  yaml.MappingNode,
					Value: "map value",
				}),
				NewExample("example3", "description 3", nil),
			},
			test: func(t *testing.T, original, cloned Examples) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, &original, &cloned)
				require.Len(t, cloned, 3)

				for i := range original {
					assert.NotSame(t, original[i], cloned[i])
					assert.Equal(t, original[i].Name(), cloned[i].Name())
					assert.Equal(t, original[i].Description, cloned[i].Description)

					if original[i].Value != nil {
						require.NotNil(t, cloned[i].Value)
						assert.NotSame(t, original[i].Value, cloned[i].Value)
						assert.Equal(t, original[i].Value.Kind, cloned[i].Value.Kind)
						assert.Equal(t, original[i].Value.Value, cloned[i].Value.Value)
					} else {
						assert.Nil(t, cloned[i].Value)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.examples.Clone()
			tt.test(t, tt.examples, cloned)
		})
	}
}

func TestExample_ToJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		example     *Example
		expected    string
		description string
	}{
		{
			name: "nil value returns null",
			example: &Example{
				name:  "test",
				Value: nil,
			},
			expected:    "null",
			description: "When Value is nil, should return 'null'",
		},
		{
			name: "scalar string value",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: "hello",
				},
			},
			expected:    `"hello"`,
			description: "String scalars should be JSON-encoded strings",
		},
		{
			name: "scalar number value",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!int",
					Value: "42",
				},
			},
			expected:    "42",
			description: "Number scalars should be JSON numbers",
		},
		{
			name: "scalar boolean true",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!bool",
					Value: "true",
				},
			},
			expected:    "true",
			description: "Boolean scalars should be JSON booleans",
		},
		{
			name: "scalar boolean false",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!bool",
					Value: "false",
				},
			},
			expected:    "false",
			description: "Boolean scalars should be JSON booleans",
		},
		{
			name: "simple map/object - should be formatted as JSON not Go map",
			example: func() *Example {
				yamlStr := `custom_field_1: value1
custom_field_2: value2`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{"custom_field_1":"value1","custom_field_2":"value2"}`,
			description: "Maps should be formatted as JSON objects, not Go map representations",
		},
		{
			name: "nested object",
			example: func() *Example {
				yamlStr := `name: John
address:
  street: Main St
  city: Springfield`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{"name":"John","address":{"street":"Main St","city":"Springfield"}}`,
			description: "Nested objects should be properly formatted as JSON",
		},
		{
			name: "array of strings",
			example: func() *Example {
				yamlStr := `- apple
- banana
- cherry`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `["apple","banana","cherry"]`,
			description: "Arrays should be formatted as JSON arrays",
		},
		{
			name: "array of objects",
			example: func() *Example {
				yamlStr := `- id: 1
  name: Alice
- id: 2
  name: Bob`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`,
			description: "Arrays of objects should be formatted as JSON",
		},
		{
			name: "object with array property",
			example: func() *Example {
				yamlStr := `users:
  - Alice
  - Bob
count: 2`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{"users":["Alice","Bob"],"count":2}`,
			description: "Objects with array properties should be formatted as JSON",
		},
		{
			name: "complex nested structure",
			example: func() *Example {
				yamlStr := `company:
  name: ACME Corp
  employees:
    - name: John
      age: 30
      skills:
        - Go
        - Python
    - name: Jane
      age: 25
      skills:
        - JavaScript
active: true`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{"company":{"name":"ACME Corp","employees":[{"name":"John","age":30,"skills":["Go","Python"]},{"name":"Jane","age":25,"skills":["JavaScript"]}]},"active":true}`,
			description: "Complex nested structures should be formatted as JSON",
		},
		{
			name: "object with null value",
			example: func() *Example {
				yamlStr := `field1: value1
field2: null
field3: value3`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{"field1":"value1","field2":null,"field3":"value3"}`,
			description: "Null values in objects should be preserved as JSON null",
		},
		{
			name: "object with numeric values",
			example: func() *Example {
				yamlStr := `integer: 42
float: 3.14
negative: -10`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{"integer":42,"float":3.14,"negative":-10}`,
			description: "Numeric values should be formatted as JSON numbers",
		},
		{
			name: "empty object",
			example: func() *Example {
				yamlStr := `{}`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `{}`,
			description: "Empty objects should be formatted as empty JSON objects",
		},
		{
			name: "empty array",
			example: func() *Example {
				yamlStr := `[]`
				var node yaml.Node
				err := yaml.Unmarshal([]byte(yamlStr), &node)
				require.NoError(t, err)
				return &Example{
					name:  "test",
					Value: &node,
				}
			}(),
			expected:    `[]`,
			description: "Empty arrays should be formatted as empty JSON arrays",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.example.ToJSON()

			// Normalize both expected and actual JSON for comparison
			// This handles any whitespace differences
			var expectedNormalized, actualNormalized any
			err := json.Unmarshal([]byte(tt.expected), &expectedNormalized)
			require.NoError(t, err, "Expected value should be valid JSON")

			err = json.Unmarshal([]byte(result), &actualNormalized)
			require.NoError(t, err, "Result should be valid JSON: %s", result)

			assert.Equal(t, expectedNormalized, actualNormalized,
				"Description: %s\nExpected JSON: %s\nActual JSON: %s",
				tt.description, tt.expected, result)

			// Also verify that the result does NOT look like a Go map representation
			if strings.Contains(result, "map[") {
				t.Errorf("Result should be JSON formatted, not Go map representation. Got: %s", result)
			}
		})
	}
}

func TestExample_ToJSON_DocumentNode(t *testing.T) {
	t.Parallel()

	// Test with a document node (which contains content)
	yamlStr := `key: value`
	var doc yaml.Node
	err := yaml.Unmarshal([]byte(yamlStr), &doc)
	require.NoError(t, err)

	example := &Example{
		name:  "test",
		Value: &doc,
	}

	result := example.ToJSON()

	var actual any
	err = json.Unmarshal([]byte(result), &actual)
	require.NoError(t, err, "Result should be valid JSON")

	expected := map[string]any{"key": "value"}
	assert.Equal(t, expected, actual)
}

func TestExample_ToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		example     *Example
		expected    string
		description string
	}{
		{
			name: "nil value returns empty string",
			example: &Example{
				name:  "test",
				Value: nil,
			},
			expected:    "",
			description: "When Value is nil, should return empty string",
		},
		{
			name: "yaml null scalar returns 'null'",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!null",
					Value: "null",
				},
			},
			expected:    "null",
			description: "yaml !!null nodes must render as 'null' (preserves user intent, mirrors ToJSON; previously printed '<nil>' which broke MDX and leaked Go formatting)",
		},
		{
			name: "scalar string value",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!str",
					Value: "hello",
				},
			},
			expected:    "hello",
			description: "String scalars should be returned as plain strings",
		},
		{
			name: "scalar number value",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!int",
					Value: "42",
				},
			},
			expected:    "42",
			description: "Number scalars should be returned as plain numbers",
		},
		{
			name: "scalar boolean true",
			example: &Example{
				name: "test",
				Value: &yaml.Node{
					Kind:  yaml.ScalarNode,
					Tag:   "!!bool",
					Value: "true",
				},
			},
			expected:    "true",
			description: "Boolean scalars should be returned as plain booleans",
		},
		{
			name: "mapping node should return JSON not Go map representation",
			example: func() *Example {
				return &Example{
					name: "test",
					Value: &yaml.Node{
						Kind: yaml.MappingNode,
						Tag:  "!!map",
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "custom_field_1"},
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "value1"},
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "custom_field_2"},
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "value2"},
						},
					},
				}
			}(),
			expected:    `{"custom_field_1":"value1","custom_field_2":"value2"}`,
			description: "Mapping nodes should be formatted as JSON, not as Go map[key:value]",
		},
		{
			name: "document node containing map - this is the bug case",
			example: func() *Example {
				// When YAML is unmarshaled, it creates a DocumentNode wrapping the actual content
				yamlStr := `custom_field_1: value1
custom_field_2: value2`
				var doc yaml.Node
				_ = yaml.Unmarshal([]byte(yamlStr), &doc)
				return &Example{
					name:  "test",
					Value: &doc, // This is a DocumentNode containing a MappingNode
				}
			}(),
			expected:    `{"custom_field_1":"value1","custom_field_2":"value2"}`,
			description: "DocumentNode containing a map should be formatted as JSON, not as map[custom_field_1:value1 custom_field_2:value2]",
		},
		{
			name: "sequence node should return JSON array",
			example: func() *Example {
				return &Example{
					name: "test",
					Value: &yaml.Node{
						Kind: yaml.SequenceNode,
						Tag:  "!!seq",
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "a"},
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "b"},
							{Kind: yaml.ScalarNode, Tag: "!!str", Value: "c"},
						},
					},
				}
			}(),
			expected:    `["a","b","c"]`,
			description: "Sequence nodes should be formatted as JSON arrays",
		},
		{
			name: "document node containing sequence",
			example: func() *Example {
				yamlStr := `- item1
- item2`
				var doc yaml.Node
				_ = yaml.Unmarshal([]byte(yamlStr), &doc)
				return &Example{
					name:  "test",
					Value: &doc,
				}
			}(),
			expected:    `["item1","item2"]`,
			description: "DocumentNode containing a sequence should be formatted as JSON array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.example.ToString()

			// For non-JSON expected values (scalars), just compare strings
			if !strings.HasPrefix(tt.expected, "{") && !strings.HasPrefix(tt.expected, "[") {
				assert.Equal(t, tt.expected, result,
					"Description: %s\nExpected: %s\nActual: %s",
					tt.description, tt.expected, result)
				return
			}

			// For JSON expected values, normalize and compare
			var expectedNormalized, actualNormalized any
			err := json.Unmarshal([]byte(tt.expected), &expectedNormalized)
			require.NoError(t, err, "Expected value should be valid JSON")

			err = json.Unmarshal([]byte(result), &actualNormalized)
			require.NoError(t, err, "Result should be valid JSON, not Go representation: %s", result)

			assert.Equal(t, expectedNormalized, actualNormalized,
				"Description: %s\nExpected JSON: %s\nActual: %s",
				tt.description, tt.expected, result)

			// Verify result does not look like a Go map representation
			if strings.Contains(result, "map[") {
				t.Errorf("Result should be JSON formatted, not Go map representation. Got: %s", result)
			}
		})
	}
}
