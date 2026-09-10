package extensions

import (
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
)

func (e *Extensions) GetDiscriminatorNameOverrides(schema *oas3.Schema) (map[string]string, error) {
	if schema.GetExtensions().Len() == 0 {
		return nil, nil
	}

	ext, ok := e.findExtension(schema.GetExtensions(), ExtDiscriminator)
	if !ok {
		return nil, nil
	}

	var overrides map[string]string
	if err := ext.Decode(&overrides); err != nil {
		return nil, err
	}

	return overrides, nil
}
