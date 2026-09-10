package extensions

import (
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (e *Extensions) HandleClassNameExtension(extensions OAExtensions) (*NameOverride, error) {
	c, node := e.parseClassNameExtension(extensions)

	if c == nil {
		return nil, nil
	}

	if err := validateClassName(c, node); err != nil {
		return nil, err
	}

	return c, nil
}

func validateClassName(c *NameOverride, node *yaml.Node) error {
	if c.Name == "" {
		return errors.NewValidationError("Value is required for "+ExtNameOverride.Name(), node, nil)
	}

	return nil
}

func (e *Extensions) parseClassNameExtension(extensions OAExtensions) (*NameOverride, *yaml.Node) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	classNameExtension, ok := e.findExtension(extensions, ExtNameOverride)
	if !ok {
		return nil, nil
	}

	return &NameOverride{Name: classNameExtension.Value}, classNameExtension
}
