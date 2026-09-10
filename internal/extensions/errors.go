package extensions

// TODO: validate errors extension
// TODO: validate all errors use json deserialization
// TODO: validate error message only live in top level error type
// TODO: validate only one error message per error type

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

type Errors struct {
	StatusCodes []string `yaml:"statusCodes"`
	Override    bool     `yaml:"override"`
}

func (e *Errors) IsErrorStatusCode(statusCode string) bool {
	for _, s := range e.StatusCodes {
		if s == statusCode || (strings.HasSuffix(s, "XX") && strings.HasPrefix(statusCode, s[:len(s)-2])) {
			return true
		}
	}

	return false
}

func (e *Extensions) IsErrorMessage(extensions OAExtensions) (bool, error) {
	if extensions.Len() == 0 {
		return false, nil
	}

	return getExtensionValue(e.GetResolvedName(ExtErrorMessage), extensions, false)
}

func (e *Extensions) HandleErrors(extensions OAExtensions) (*Errors, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	errorsExt, ok := e.findExtension(extensions, ExtErrors)
	if !ok {
		return nil, nil
	}

	var errs Errors
	if err := errorsExt.Decode(&errs); err != nil {
		return nil, errors.NewValidationError("failed to unmarshal "+ExtErrors.Name(), errorsExt, err)
	}

	if len(errs.StatusCodes) == 0 {
		return nil, errors.NewValidationError(ExtErrors.Name()+".statusCodes is required", errorsExt, nil)
	}

	return &errs, nil
}

func MergeErrors(a, b *Errors) *Errors {
	if a != nil {
		if b == nil {
			aCopy := *a
			b = &aCopy
		} else if !b.Override {
			b.StatusCodes = append(b.StatusCodes, a.StatusCodes...)
		}
	}

	return b
}
