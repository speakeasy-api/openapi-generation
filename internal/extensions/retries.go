package extensions

import (
	"reflect"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"gopkg.in/yaml.v3"
)

const (
	defaultInitialInterval = 500   // 500 milliseconds
	defaultMaxInterval     = 60000 // 60 seconds
	defaultExponent        = 1.5
	defaultMaxElapsedTime  = 3600000 // 60 minutes
)

type BackoffStrategy struct {
	InitialInterval *int     `json:"initialInterval" yaml:"initialInterval,omitempty"`
	MaxInterval     *int     `json:"maxInterval" yaml:"maxInterval,omitempty"`
	Exponent        *float32 `json:"exponent" yaml:"exponent,omitempty"`
	MaxElapsedTime  *int     `json:"maxElapsedTime" yaml:"maxElapsedTime,omitempty"`
}

type Retries struct {
	Strategy              string           `json:"strategy" yaml:"strategy,omitempty"`
	DefaultApplied        bool             `json:"defaultApplied,omitempty" yaml:"defaultApplied,omitempty"`
	Disabled              *bool            `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Backoff               *BackoffStrategy `json:"backoff" yaml:"backoff,omitempty"`
	StatusCodes           []string         `json:"statusCodes" yaml:"statusCodes,omitempty"`
	RetryConnectionErrors *bool            `json:"retryConnectionErrors" yaml:"retryConnectionErrors,omitempty"`
	MaxRetries            *int             `json:"maxRetries" yaml:"maxRetries,omitempty"`
}

func (e *Extensions) HandleGlobalRetryExtension(doc *openapi.OpenAPI, defaultEnabledRetries bool) (*Retries, error) {
	r, _, err := e.handleRetryExtension(doc.GetExtensions(), defaultEnabledRetries)
	return r, err
}

func (e *Extensions) HandleOperationRetryExtension(operation *openapi.Operation) (*Retries, bool, error) {
	r, present, err := e.handleRetryExtension(operation.GetExtensions(), false)
	if err != nil {
		return nil, false, err
	}

	if !present {
		return nil, false, nil
	}

	if r == nil || reflect.DeepEqual(*r, Retries{}) {
		return nil, true, nil
	}

	return r, true, nil
}

func (e *Extensions) handleRetryExtension(extensions OAExtensions, enableDefaultRetries bool) (*Retries, bool, error) {
	r, node, err := e.parseRetryExtension(extensions)
	if err != nil {
		return nil, false, err
	}
	present := r != nil

	if r == nil {
		if enableDefaultRetries {
			r = &Retries{
				DefaultApplied: true,
				StatusCodes:    []string{"429", "500", "502", "503", "504"},
			}
		} else {
			return nil, false, nil
		}
	}

	if (r.Disabled != nil && *r.Disabled) || r.Strategy == "none" {
		return nil, present, nil
	}

	if reflect.DeepEqual(*r, Retries{}) {
		return r, present, nil
	}

	if err := validateRetries(r, node); err != nil {
		return nil, false, err
	}

	populateDefaults(r)

	return r, present, nil
}

func validateRetries(r *Retries, node *yaml.Node) error {
	if r.Strategy == "" && !r.DefaultApplied {
		return errors.NewValidationError(ExtRetries.Name()+".strategy is required", node, nil)
	}

	switch r.Strategy {
	case "", "backoff", "attempt-count-backoff":
	default:
		return errors.NewValidationError(ExtRetries.Name()+".strategy must be one of backoff, attempt-count-backoff, or none", node, nil)
	}

	if r.Strategy == "attempt-count-backoff" && r.MaxRetries == nil {
		return errors.NewValidationError(ExtRetries.Name()+".maxRetries is required for attempt-count-backoff", node, nil)
	}

	if r.MaxRetries != nil && *r.MaxRetries < 0 {
		return errors.NewValidationError(ExtRetries.Name()+".maxRetries must be greater than or equal to 0", node, nil)
	}

	// TODO validate status codes
	if len(r.StatusCodes) == 0 {
		return errors.NewValidationError(ExtRetries.Name()+".statusCodes is required", node, nil)
	}

	return nil
}

func populateDefaults(r *Retries) {
	if r.Backoff != nil {
		if r.Backoff.InitialInterval == nil {
			r.Backoff.InitialInterval = pointer.From(defaultInitialInterval)
		}

		if r.Backoff.MaxInterval == nil {
			r.Backoff.MaxInterval = pointer.From(defaultMaxInterval)
		}

		if r.Backoff.Exponent == nil {
			r.Backoff.Exponent = pointer.From[float32](defaultExponent)
		}

		if r.Backoff.MaxElapsedTime == nil {
			r.Backoff.MaxElapsedTime = pointer.From(defaultMaxElapsedTime)
		}
	} else if r.Strategy == "backoff" || r.Strategy == "attempt-count-backoff" {
		r.Backoff = &BackoffStrategy{
			InitialInterval: pointer.From(defaultInitialInterval),
			MaxInterval:     pointer.From(defaultMaxInterval),
			Exponent:        pointer.From[float32](defaultExponent),
			MaxElapsedTime:  pointer.From(defaultMaxElapsedTime),
		}
	}

	if r.RetryConnectionErrors == nil {
		r.RetryConnectionErrors = pointer.From(true)
	}
}

func (e *Extensions) parseRetryExtension(ext OAExtensions) (*Retries, *yaml.Node, error) {
	if ext.Len() == 0 {
		return nil, nil, nil
	}

	retryExtension, ok := e.findExtension(ext, ExtRetries)
	if !ok {
		return nil, nil, nil
	}

	var r Retries
	if err := retryExtension.Decode(&r); err != nil {
		return nil, nil, errors.NewValidationError("failed to unmarshal "+ExtRetries.Name(), retryExtension, err)
	}

	return &r, retryExtension, nil
}
