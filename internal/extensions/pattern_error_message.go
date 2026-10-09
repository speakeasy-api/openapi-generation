package extensions

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

func (e *Extensions) HandlePatternErrorMessageExtension(exts OAExtensions) (string, error) {
	node, ok := e.findExtension(exts, ExtPatternErrorMessage)
	if !ok {
		return "", nil
	}

	var value any
	if err := node.Decode(&value); err != nil {
		return "", errors.NewValidationError("failed to unmarshal "+e.GetResolvedName(ExtPatternErrorMessage), node, ErrUnmarshal.Wrap(err))
	}
	message, ok := value.(string)
	if !ok || strings.TrimSpace(message) == "" {
		return "", errors.NewValidationError(e.GetResolvedName(ExtPatternErrorMessage)+" must be a non-empty string", node, nil)
	}
	return message, nil
}
