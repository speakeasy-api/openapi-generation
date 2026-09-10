package extensions

import (
	"fmt"

	"github.com/speakeasy-api/openapi/jsonschema/oas3"
)

// Base64InputModeFile is the only currently-supported value
// for the x-speakeasy-base64-input-mode extension.
const Base64InputModeFile = "file"

// Base64InputMode returns the value of the x-speakeasy-base64-input-mode extension
// on a schema. Any value other than "file" is treated as unset (returns "").
func (e *Extensions) Base64InputMode(schema *oas3.Schema) (string, error) {
	if schema.GetExtensions().Len() == 0 {
		return "", nil
	}

	allowed := map[string]bool{"": true, Base64InputModeFile: true}

	return getExtensionValueWithValidation(e.GetResolvedName(ExtBase64InputMode), schema.GetExtensions(), "", func(value string) error {
		if _, ok := allowed[value]; !ok {
			return fmt.Errorf("%s: value is not allowed (supported: %s)", value, Base64InputModeFile)
		}
		return nil
	})
}
