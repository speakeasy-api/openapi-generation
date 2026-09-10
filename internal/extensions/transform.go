package extensions

import (
	"strings"

	"github.com/itchyny/gojq"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"gopkg.in/yaml.v3"
)

// TransformerType is a string enum that can right now only be "jq"
type TransformerType string

const (
	Jq TransformerType = "jq"
)

type TransformerConfig struct {
	Type   TransformerType
	Config string
}

func (e *Extensions) HandleTransformExtension(ext OAExtensions, transformExt Extension) (*TransformerConfig, error) {
	if ext.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(ext, transformExt)
	if !ok {
		return nil, nil
	}

	// Only support structured YAML format: { jq: "expression" }
	if yamlNode.Kind != yaml.MappingNode {
		return nil, errors.NewValidationError(transformExt.Name()+" must be a mapping with a 'jq' field", yamlNode, nil)
	}

	var transformMap map[string]string
	if err := yamlNode.Decode(&transformMap); err != nil {
		return nil, errors.NewValidationError(transformExt.Name()+" must be a valid mapping", yamlNode, nil)
	}

	jqExpr, ok := transformMap["jq"]
	if !ok {
		return nil, errors.NewValidationError(transformExt.Name()+" requires a 'jq' field", yamlNode, nil)
	}

	configValue := strings.TrimSpace(jqExpr)
	if configValue == "" {
		return nil, errors.NewValidationError(transformExt.Name()+" jq expression cannot be empty", yamlNode, nil)
	}

	q, err := gojq.Parse(configValue)
	if err != nil {
		return nil, errors.NewValidationError(transformExt.Name()+": '"+configValue+"' is an invalid jq expression: "+err.Error(), yamlNode, nil)
	}

	value := q.String()

	return &TransformerConfig{
		Type:   Jq,
		Config: value,
	}, nil
}
