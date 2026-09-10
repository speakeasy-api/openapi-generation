package extensions

import (
	"reflect"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

type WindowStrategy struct {
	Rate   int    `json:"rate" yaml:"rate"`
	Period string `json:"period" yaml:"period"`
}

type RateLimit struct {
	Strategy      string          `json:"strategy" yaml:"strategy"`
	SlidingWindow *WindowStrategy `json:"sliding_window" yaml:"sliding_window"`
	Identifier    string          `json:"identifier" yaml:"identifier"`
	Description   string          `json:"description" yaml:"description"`
}

func (e *Extensions) HandleDocsRateLimitExtension(operation *openapi.Operation) ([]RateLimit, error) {
	r, err := e.handleRateLimitExtension(operation.GetExtensions())
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (e *Extensions) handleRateLimitExtension(extensions OAExtensions) ([]RateLimit, error) {
	r, _, err := e.parseRateLimit(extensions)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, nil
	}

	if reflect.DeepEqual(r, []RateLimit{}) {
		return r, nil
	}

	return r, nil
}

func (e *Extensions) parseRateLimit(ext OAExtensions) ([]RateLimit, *yaml.Node, error) {
	if ext.Len() == 0 {
		return nil, nil, nil
	}

	rateLimitExtension, ok := e.findExtension(ext, ExtDocsRateLimits)
	if !ok {
		return nil, nil, nil
	}

	var r []RateLimit
	if err := rateLimitExtension.Decode(&r); err != nil {
		return nil, nil, errors.NewValidationError("failed to unmarshal "+ExtDocsRateLimits.Name(), rateLimitExtension, err)
	}

	return r, rateLimitExtension, nil
}
