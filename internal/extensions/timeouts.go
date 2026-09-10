package extensions

import (
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

func (e *Extensions) handleTimeoutExtension(extensions OAExtensions) (*int64, error) {
	r, node, err := e.parseTimeoutExtension(extensions)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, nil
	}

	if *r <= 0 {
		return nil, errors.NewValidationError(ExtTimeout.Name()+" must be greater than 0 milliseconds", node, nil)
	}

	return r, nil
}

func (e *Extensions) HandleGlobalTimeoutExtension(doc *openapi.OpenAPI) (*int64, error) {
	return e.handleTimeoutExtension(doc.GetExtensions())
}

func (e *Extensions) HandleOperationTimeoutExtension(operation *openapi.Operation) (*int64, bool, error) {
	r, err := e.handleTimeoutExtension(operation.GetExtensions())
	if err != nil {
		return nil, false, err
	}

	if r != nil {
		return r, true, nil
	}

	return nil, false, nil
}

func (e *Extensions) parseTimeoutExtension(ext OAExtensions) (*int64, *yaml.Node, error) {
	if ext.Len() == 0 {
		return nil, nil, nil
	}

	timeoutExtension, ok := e.findExtension(ext, ExtTimeout)
	if !ok {
		return nil, nil, nil
	}

	var r int64
	if err := timeoutExtension.Decode(&r); err != nil {
		return nil, nil, errors.NewValidationError("failed to unmarshal "+ExtTimeout.Name(), timeoutExtension, err)
	}

	return &r, timeoutExtension, nil
}
