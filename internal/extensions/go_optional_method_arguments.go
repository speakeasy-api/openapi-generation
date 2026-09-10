package extensions

import "fmt"

type GoOptionalMethodArguments string

const (
	GoOptionalMethodArgumentsPointers      GoOptionalMethodArguments = "pointers"
	GoOptionalMethodArgumentsSharedOptions GoOptionalMethodArguments = "shared-options"
	GoOptionalMethodArgumentsMethodOptions GoOptionalMethodArguments = "method-options"
)

func (e *Extensions) GetGoOptionalMethodArguments(extensions OAExtensions) (*GoOptionalMethodArguments, error) {
	if extensions == nil || extensions.Len() == 0 {
		return nil, nil
	}

	return getExtensionValueWithValidation(e.GetResolvedName(ExtGoOptionalMethodArguments), extensions, (*GoOptionalMethodArguments)(nil), func(value *GoOptionalMethodArguments) error {
		if value == nil {
			return nil
		}

		switch *value {
		case GoOptionalMethodArgumentsPointers,
			GoOptionalMethodArgumentsSharedOptions,
			GoOptionalMethodArgumentsMethodOptions:
			return nil
		default:
			return fmt.Errorf("%s: value is not allowed (supported: %s, %s, %s)", *value, GoOptionalMethodArgumentsPointers, GoOptionalMethodArgumentsSharedOptions, GoOptionalMethodArgumentsMethodOptions)
		}
	})
}
