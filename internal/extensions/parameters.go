package extensions

import (
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
)

func (e *Extensions) HandleOperationParameterNameExtension(param *openapi.Parameter) (*NameOverride, bool, error) {
	n, err := e.handleNameOverrideExtension(param.GetExtensions(), Parameter)
	if err != nil {
		return nil, false, err
	}

	if n != nil {
		// use first override and ignore the rest
		return &NameOverride{
			ParameterName: param.Name,
			Name:          n[0].Name,
		}, true, nil
	}

	return nil, false, nil
}

func (e *Extensions) HandleAllowEmptyQueryParameterValueExtension(param *openapi.Parameter) (bool, error) {
	if param.Extensions.Len() == 0 {
		return false, nil
	}

	ext, ok := e.findExtension(param.Extensions, ExtAllowEmptyValue)
	if !ok {
		return false, nil
	}

	var allowEmptyValue bool
	if err := ext.Decode(&allowEmptyValue); err != nil {
		return false, errors.NewValidationError("failed to unmarshal "+ExtAllowEmptyValue.Name(), ext, err)
	}

	return allowEmptyValue, nil
}

// DoesParamAllowReserved reports whether the parameter requests reserved characters to pass through unencoded.
// Query parameters opt in via the spec's allowReserved keyword; Path parameters only support the extension.
func (e *Extensions) DoesParamAllowReserved(param *openapi.Parameter) (allowed bool, unsupported bool) {
	val, ok := e.findExtension(param.GetExtensions(), ExtParamEncodingOverride)
	hasOverride := ok && val != nil

	if param.GetIn() != openapi.ParameterInPath {
		return param.GetAllowReserved(), hasOverride && val.Value == "allowReserved"
	}

	if !hasOverride {
		return false, param.GetAllowReserved()
	}

	return val.Value == "allowReserved", false
}
