package tfmockserver

import (
	"fmt"
	"testing"
)

// AssertRequestBodyField checks that a specific field in the request body has the expected value.
// It handles nil, arrays, and other types appropriately.
//
// Example usage:
//
//	captured := endpoints.Create[0].Captured()
//	tfmockserver.AssertRequestBodyField(t, captured[0].Body, "my_field", nil)  // expect null
//	tfmockserver.AssertRequestBodyField(t, captured[0].Body, "my_list", []any{})  // expect empty array
//	tfmockserver.AssertRequestBodyField(t, captured[0].Body, "my_list", []any{"a", "b"})  // expect values
func AssertRequestBodyField(t *testing.T, body map[string]any, field string, expected any) {
	t.Helper()

	if err := assertRequestBodyField(body, field, expected); err != nil {
		t.Errorf("%s", err)
	}
}

func assertRequestBodyField(body map[string]any, field string, expected any) error {
	actual, exists := body[field]

	if !exists {
		return fmt.Errorf("expected field %q to be present in request body, but it was not", field)
	}

	if expected == nil {
		if actual != nil {
			return fmt.Errorf("expected field %q to be null, got %v (%T)", field, actual, actual)
		}
		return nil
	}

	if expectedArr, ok := expected.([]any); ok {
		actualArr, ok := actual.([]interface{})
		if !ok {
			return fmt.Errorf("expected field %q to be an array, got %v (%T)", field, actual, actual)
		}
		if len(expectedArr) != len(actualArr) {
			return fmt.Errorf("expected field %q to have %d elements, got %d: %v", field, len(expectedArr), len(actualArr), actualArr)
		}
		for i, exp := range expectedArr {
			if fmt.Sprintf("%v", actualArr[i]) != fmt.Sprintf("%v", exp) {
				return fmt.Errorf("expected field %q[%d] to be %v, got %v", field, i, exp, actualArr[i])
			}
		}
		return nil
	}

	if fmt.Sprintf("%v", actual) != fmt.Sprintf("%v", expected) {
		return fmt.Errorf("expected field %q to be %v (%T), got %v (%T)", field, expected, expected, actual, actual)
	}

	return nil
}

// AssertRequestBodyFieldOmitted checks that a specific field is NOT present in the request body.
//
// Example usage:
//
//	captured := endpoints.Create[0].Captured()
//	tfmockserver.AssertRequestBodyFieldOmitted(t, captured[0].Body, "optional_field")
func AssertRequestBodyFieldOmitted(t *testing.T, body map[string]any, field string) {
	t.Helper()

	if err := assertRequestBodyFieldOmitted(body, field); err != nil {
		t.Errorf("%s", err)
	}
}

func assertRequestBodyFieldOmitted(body map[string]any, field string) error {
	if actual, exists := body[field]; exists {
		return fmt.Errorf("expected field %q to be omitted from request body, but it was present with value: %v (%T)", field, actual, actual)
	}
	return nil
}

// AssertRequestBodyFieldNull checks that a specific field is present and has a null value.
//
// Example usage:
//
//	captured := endpoints.Create[0].Captured()
//	tfmockserver.AssertRequestBodyFieldNull(t, captured[0].Body, "nullable_field")
func AssertRequestBodyFieldNull(t *testing.T, body map[string]any, field string) {
	t.Helper()

	if err := assertRequestBodyFieldNull(body, field); err != nil {
		t.Errorf("%s", err)
	}
}

func assertRequestBodyFieldNull(body map[string]any, field string) error {
	actual, exists := body[field]

	if !exists {
		return fmt.Errorf("expected field %q to be present with null value, but it was omitted", field)
	}

	if actual != nil {
		return fmt.Errorf("expected field %q to be null, got %v (%T)", field, actual, actual)
	}

	return nil
}

// AssertRequestBodyFieldEmptyArray checks that a specific field is present and has an empty array value.
//
// Example usage:
//
//	captured := endpoints.Create[0].Captured()
//	tfmockserver.AssertRequestBodyFieldEmptyArray(t, captured[0].Body, "list_field")
func AssertRequestBodyFieldEmptyArray(t *testing.T, body map[string]any, field string) {
	t.Helper()

	if err := assertRequestBodyFieldEmptyArray(body, field); err != nil {
		t.Errorf("%s", err)
	}
}

func assertRequestBodyFieldEmptyArray(body map[string]any, field string) error {
	actual, exists := body[field]

	if !exists {
		return fmt.Errorf("expected field %q to be present with empty array, but it was omitted", field)
	}

	if actual == nil {
		return fmt.Errorf("expected field %q to be an empty array, got null", field)
	}

	arr, ok := actual.([]interface{})

	if !ok {
		return fmt.Errorf("expected field %q to be an array, got %v (%T)", field, actual, actual)
	}

	if len(arr) != 0 {
		return fmt.Errorf("expected field %q to be an empty array, got %v", field, arr)
	}

	return nil
}

// AssertRequestBodyFieldArray checks that a specific field is present and has the expected array values.
//
// Example usage:
//
//	captured := endpoints.Create[0].Captured()
//	tfmockserver.AssertRequestBodyFieldArray(t, captured[0].Body, "list_field", []any{"foo", "bar"})
func AssertRequestBodyFieldArray(t *testing.T, body map[string]any, field string, expected []any) {
	t.Helper()

	if err := assertRequestBodyFieldArray(body, field, expected); err != nil {
		t.Errorf("%s", err)
	}
}

func assertRequestBodyFieldArray(body map[string]any, field string, expected []any) error {
	actual, exists := body[field]

	if !exists {
		return fmt.Errorf("expected field %q to be present with array %v, but it was omitted", field, expected)
	}

	if actual == nil {
		return fmt.Errorf("expected field %q to be an array %v, got null", field, expected)
	}

	arr, ok := actual.([]interface{})

	if !ok {
		return fmt.Errorf("expected field %q to be an array, got %v (%T)", field, actual, actual)
	}

	if len(arr) != len(expected) {
		return fmt.Errorf("expected field %q to have %d elements, got %d: %v", field, len(expected), len(arr), arr)
	}

	for i, exp := range expected {
		if fmt.Sprintf("%v", arr[i]) != fmt.Sprintf("%v", exp) {
			return fmt.Errorf("expected field %q[%d] to be %v, got %v", field, i, exp, arr[i])
		}
	}

	return nil
}
