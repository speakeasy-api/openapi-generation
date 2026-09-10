package extensions

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

type ExtensionScope int

const (
	Global ExtensionScope = iota
	Operation
	Parameter
)

type NameOverride struct {
	OperationId                 string `json:"operationId" yaml:"operationId"`
	GlobalMethodNameOverride    string `json:"methodNameOverride" yaml:"methodNameOverride"`
	ParameterName               string `json:"parameterName" yaml:"parameterName"`
	GlobalParameterNameOverride string `json:"parameterNameOverride" yaml:"parameterNameOverride"`
	Name                        string
	Node                        *yaml.Node // The yaml node where this extension is defined (for error reporting)
}

func (e *Extensions) HandleGlobalNameOverrideExtensions(doc *openapi.OpenAPI) ([]*NameOverride, error) {
	return e.handleNameOverrideExtension(doc.GetExtensions(), Global)
}

func (e *Extensions) HandleOperationMethodNameExtension(operation *openapi.Operation) (*NameOverride, bool, error) {
	n, err := e.handleNameOverrideExtension(operation.GetExtensions(), Operation)
	if err != nil {
		return nil, false, err
	}

	if n != nil {
		// use first override and ignore the rest
		return &NameOverride{
			OperationId: operation.GetOperationID(),
			Name:        n[0].Name,
			Node:        n[0].Node,
		}, true, nil
	}

	return nil, false, nil
}

func (e *Extensions) handleNameOverrideExtension(extensions OAExtensions, scope ExtensionScope) ([]*NameOverride, error) {
	n, err := e.parseNameOverrideExtension(extensions, scope)
	if err != nil {
		return nil, err
	}

	if n == nil {
		return nil, nil
	}

	return n, nil
}

func validateNameOverrides(n []*NameOverride, node *yaml.Node, scope ExtensionScope) error {
	for _, override := range n {
		if scope == Global {
			if override.GlobalMethodNameOverride == "" && override.GlobalParameterNameOverride == "" {
				return errors.NewValidationError(fmt.Sprintf("Either %s.parameterNameOverride or %s.methodNameOverride is required", ExtNameOverride.Name(), ExtNameOverride.Name()), node, nil)
			}
			if override.OperationId == "" && override.ParameterName == "" {
				return errors.NewValidationError(fmt.Sprintf("Either %s.parameterName or %s.operationId is required", ExtNameOverride.Name(), ExtNameOverride.Name()), node, nil)
			}
		} else if override.Name == "" {
			return errors.NewValidationError("Value is required for "+ExtNameOverride.Name(), node, nil)
		}
	}

	return nil
}

func (e *Extensions) parseNameOverrideExtension(extensions OAExtensions, scope ExtensionScope) ([]*NameOverride, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	nameOverrideExtension, ok := e.findExtension(extensions, ExtNameOverride)
	if !ok {
		return nil, nil
	}

	if scope == Global {
		n, err := e.ParseGlobalNameOverrideExtension(nameOverrideExtension)
		if err != nil {
			return nil, err
		}

		// Set the node on all global name overrides
		for _, override := range n {
			override.Node = nameOverrideExtension
		}

		return n, nil
	}

	n := []*NameOverride{{
		Name: nameOverrideExtension.Value,
		Node: nameOverrideExtension,
	}}

	if err := validateNameOverrides(n, nameOverrideExtension, scope); err != nil {
		return nil, err
	}

	return n, nil
}

func (e *Extensions) ParseGlobalNameOverrideExtension(nameOverrideExtension *yaml.Node) ([]*NameOverride, error) {
	var n []*NameOverride
	if err := nameOverrideExtension.Decode(&n); err != nil {
		return nil, errors.NewValidationError("failed to unmarshal "+ExtNameOverride.Name(), nameOverrideExtension, err)
	}

	if err := validateNameOverrides(n, nameOverrideExtension, Global); err != nil {
		return nil, err
	}

	return n, nil
}
