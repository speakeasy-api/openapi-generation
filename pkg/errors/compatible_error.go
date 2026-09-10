package errors

import (
	"reflect"
	"strings"
)

const errorSeparator = " -- "

// Error is a comparable string error suitable for const sentinels.
type Error string

func (s Error) Error() string {
	return string(s)
}

func (s Error) Is(target error) bool {
	return s.Error() == target.Error() || strings.HasPrefix(target.Error(), s.Error()+errorSeparator)
}

func (s Error) As(target interface{}) bool {
	value := reflect.ValueOf(target).Elem()
	if value.Type().Name() == "Error" && value.CanSet() {
		value.SetString(string(s))
		return true
	}
	return false
}

// Wrap adds err as a cause while retaining Error's compatible message matching.
func (s Error) Wrap(err error) error {
	return wrappedError{cause: err, message: string(s)}
}

type wrappedError struct {
	cause   error
	message string
}

func (e wrappedError) Error() string {
	if e.cause == nil {
		return e.message
	}
	return e.message + errorSeparator + e.cause.Error()
}

func (e wrappedError) Unwrap() error {
	return e.cause
}

func (e wrappedError) Is(target error) bool {
	return Error(e.message).Is(target)
}

func (e wrappedError) As(target interface{}) bool {
	return Error(e.message).As(target)
}
