package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestResponseBodyContent_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		original *ResponseBodyContent
		check    func(t *testing.T, original, cloned *ResponseBodyContent)
	}{
		{
			name:     "nil ResponseBodyContent",
			original: nil,
			check: func(t *testing.T, original, cloned *ResponseBodyContent) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "ResponseBodyContent with all fields",
			original: &ResponseBodyContent{
				SerializationMethod: "json",
				ContentType:         "application/json",
				Content: &FieldDef{
					Name: "responseField",
					Type: &TypeDef{
						Type: DataTypeString,
					},
				},
				UsageExample: true,
				Examples: Examples{
					NewExample("example1", "description", &yaml.Node{Value: "value"}),
				},
				SSESentinel: "sentinel",
			},
			check: func(t *testing.T, original, cloned *ResponseBodyContent) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.SerializationMethod, cloned.SerializationMethod)
				assert.Equal(t, original.ContentType, cloned.ContentType)
				assert.NotSame(t, original.Content, cloned.Content)
				assert.Equal(t, original.Content.Name, cloned.Content.Name)
				assert.Equal(t, original.UsageExample, cloned.UsageExample)
				assert.Len(t, cloned.Examples, len(original.Examples))
				for i := range original.Examples {
					assert.NotSame(t, original.Examples[i], cloned.Examples[i])
					assert.Equal(t, original.Examples[i].Name(), cloned.Examples[i].Name())
				}
				assert.Equal(t, original.SSESentinel, cloned.SSESentinel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.original.Clone()
			tt.check(t, tt.original, cloned)
		})
	}
}
