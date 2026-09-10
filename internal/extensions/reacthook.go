package extensions

import (
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
)

type ReactHookQueryKey struct {
	IncludeRequestBody bool `json:"includeRequestBody" yaml:"includeRequestBody"`
}

type ReactHook struct {
	Disabled bool               `json:"disabled" yaml:"disabled"`
	Name     string             `json:"name" yaml:"name"`
	Type     string             `json:"type" yaml:"type"`
	QueryKey *ReactHookQueryKey `json:"queryKey" yaml:"queryKey"`
}

func (e *Extensions) HandleReactHookExtension(operation *openapi.Operation) (*ReactHook, error) {
	if operation.GetExtensions().Len() == 0 {
		return nil, nil
	}

	extension, ok := e.findExtension(operation.GetExtensions(), ExtReactHook)
	if !ok {
		return nil, nil
	}

	var reactHook ReactHook
	if err := extension.Decode(&reactHook); err != nil {
		return nil, errors.NewValidationError("failed to unmarshal "+ExtReactHook.Name(), extension, err)
	}

	if reactHook.Type == "" {
		reactHook.Type = "infer"
	}

	if reactHook.Type != "infer" && reactHook.Type != "query" && reactHook.Type != "mutation" {
		return nil, errors.NewValidationError(ExtReactHook.Name()+": react hook type must be either 'infer' or 'query' or 'mutation'", extension, nil)
	}

	if reactHook.QueryKey != nil && reactHook.QueryKey.IncludeRequestBody && reactHook.Type == "mutation" {
		return nil, errors.NewValidationError(ExtReactHook.Name()+": queryKey.includeRequestBody only applies to query hooks", extension, nil)
	}

	return &reactHook, nil
}
