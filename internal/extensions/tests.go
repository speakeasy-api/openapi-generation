package extensions

import (
	"os"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
)

func (e *Extensions) IsTestingEnabled(extensions OAExtensions) (*bool, error) {
	t := e.target.Target
	if os.Getenv("SPEAKEASY_DEBUG") == "true" {
		t = e.target.Template
	}

	if extensions.Len() == 0 {
		return nil, nil
	}

	isTest, err := getExtensionValue[*bool](e.GetResolvedName(ExtTest), extensions, nil)
	if err != nil {
		if !errors.Is(err, ErrUnmarshal) {
			return nil, err
		}

		targets, err := getExtensionValue[[]string](e.GetResolvedName(ExtTest), extensions, nil)
		if err != nil {
			return nil, err
		}

		return pointer.From(slices.Contains(targets, t)), nil
	}

	return isTest, nil
}

func (e *Extensions) GetTestID(operation *openapi.Operation) (string, error) {
	if operation.GetExtensions().Len() == 0 {
		return "", nil
	}

	return getExtensionValue(e.GetResolvedName(ExtTestID), operation.GetExtensions(), "")
}

func (e *Extensions) GetTestDirectives(extensions OAExtensions) ([]string, error) {
	return getExtensionValue(e.GetResolvedName(ExtTestInternalDirectives), extensions, []string{})
}

func (e *Extensions) IsTestIgnored(extensions OAExtensions) (bool, error) {
	if extensions.Len() == 0 {
		return false, nil
	}
	return getExtensionValue(e.GetResolvedName(ExtTestIgnore), extensions, false)
}
