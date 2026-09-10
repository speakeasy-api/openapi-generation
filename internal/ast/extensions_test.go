package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	openapiExtensions "github.com/speakeasy-api/openapi/extensions"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// yamlToOAExtensions is a helper function that converts a YAML string to OAExtensions
func yamlToOAExtensions(t *testing.T, yamlStr string) extensions.OAExtensions {
	t.Helper()

	if yamlStr == "" {
		return openapiExtensions.New()
	}

	var doc yaml.Node
	err := yaml.Unmarshal([]byte(yamlStr), &doc)
	require.NoError(t, err, "failed to unmarshal YAML")

	ext := openapiExtensions.New()

	// The document should be a mapping node containing the extensions
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		if doc.Content[0].Kind == yaml.MappingNode {
			// Process each key-value pair in the mapping
			for i := 0; i < len(doc.Content[0].Content); i += 2 {
				keyNode := doc.Content[0].Content[i]
				valueNode := doc.Content[0].Content[i+1]

				ext.Set(keyNode.Value, valueNode)
			}
		}
	}

	return ext
}
