package testutils

import (
	"fmt"
	"reflect"
	"testing"
)

// DeepCompare performs a deep comparison of two values using reflection and prints differences
func DeepCompare(t *testing.T, expected, actual interface{}) bool {
	t.Helper()
	return deepCompareRecursive(t, reflect.ValueOf(expected), reflect.ValueOf(actual), "root", make(map[uintptr]bool), 0)
}

func deepCompareRecursive(t *testing.T, expected, actual reflect.Value, path string, visited map[uintptr]bool, depth int) bool {
	t.Helper()
	if depth > 100 {
		t.Logf("Max depth reached at path: %s", path)
		return true
	}

	// Handle nil cases
	if !expected.IsValid() && !actual.IsValid() {
		return true
	}
	if !expected.IsValid() || !actual.IsValid() {
		t.Logf("DIFF at %s: expected.IsValid()=%v, actual.IsValid()=%v", path, expected.IsValid(), actual.IsValid())
		return false
	}

	// Handle different types
	if expected.Type() != actual.Type() {
		t.Logf("DIFF at %s: expected type=%v, actual type=%v", path, expected.Type(), actual.Type())
		return false
	}

	// Handle pointer cycles
	if expected.Kind() == reflect.Pointer && actual.Kind() == reflect.Pointer {
		if expected.IsNil() && actual.IsNil() {
			return true
		}
		if expected.IsNil() || actual.IsNil() {
			t.Logf("DIFF at %s: expected.IsNil()=%v, actual.IsNil()=%v", path, expected.IsNil(), actual.IsNil())
			return false
		}

		expectedPtr := expected.Pointer()
		actualPtr := actual.Pointer()

		// If both pointers are the same, they're definitely equal
		if expectedPtr == actualPtr {
			return true
		}

		// Check for cycles - only skip if we've seen this exact pair before
		visitKey := expectedPtr
		if visited[visitKey] {
			return true // Already visited, assume equal to avoid infinite recursion
		}
		visited[visitKey] = true

		return deepCompareRecursive(t, expected.Elem(), actual.Elem(), path+"->", visited, depth+1)
	}

	switch expected.Kind() {
	case reflect.String:
		if expected.String() != actual.String() {
			t.Logf("DIFF at %s: expected=%q, actual=%q", path, expected.String(), actual.String())
			return false
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if expected.Int() != actual.Int() {
			t.Logf("DIFF at %s: expected=%d, actual=%d", path, expected.Int(), actual.Int())
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if expected.Uint() != actual.Uint() {
			t.Logf("DIFF at %s: expected=%d, actual=%d", path, expected.Uint(), actual.Uint())
			return false
		}
	case reflect.Bool:
		if expected.Bool() != actual.Bool() {
			t.Logf("DIFF at %s: expected=%v, actual=%v", path, expected.Bool(), actual.Bool())
			return false
		}
	case reflect.Float32, reflect.Float64:
		if expected.Float() != actual.Float() {
			t.Logf("DIFF at %s: expected=%f, actual=%f", path, expected.Float(), actual.Float())
			return false
		}
	case reflect.Slice:
		// Check for nil vs empty slice difference
		if expected.IsNil() != actual.IsNil() {
			t.Logf("DIFF at %s: expected.IsNil()=%v, actual.IsNil()=%v", path, expected.IsNil(), actual.IsNil())
			return false
		}
		if expected.Len() != actual.Len() {
			t.Logf("DIFF at %s: expected len=%d, actual len=%d", path, expected.Len(), actual.Len())
			return false
		}
		equal := true
		for i := 0; i < expected.Len(); i++ {
			if !deepCompareRecursive(t, expected.Index(i), actual.Index(i), fmt.Sprintf("%s[%d]", path, i), visited, depth+1) {
				equal = false
			}
		}
		return equal
	case reflect.Map:
		if expected.Len() != actual.Len() {
			t.Logf("DIFF at %s: expected len=%d, actual len=%d", path, expected.Len(), actual.Len())
			return false
		}
		equal := true
		for _, key := range expected.MapKeys() {
			expectedVal := expected.MapIndex(key)
			actualVal := actual.MapIndex(key)
			if !actualVal.IsValid() {
				t.Logf("DIFF at %s[%v]: key missing in actual", path, key.Interface())
				equal = false
				continue
			}
			if !deepCompareRecursive(t, expectedVal, actualVal, fmt.Sprintf("%s[%v]", path, key.Interface()), visited, depth+1) {
				equal = false
			}
		}
		return equal
	case reflect.Struct:
		equal := true
		for i := 0; i < expected.NumField(); i++ {
			field := expected.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			expectedField := expected.Field(i)
			actualField := actual.Field(i)
			fieldPath := path + "." + field.Name
			if !deepCompareRecursive(t, expectedField, actualField, fieldPath, visited, depth+1) {
				equal = false
			}
		}
		return equal
	case reflect.Interface:
		if expected.IsNil() && actual.IsNil() {
			return true
		}
		if expected.IsNil() || actual.IsNil() {
			t.Logf("DIFF at %s: expected.IsNil()=%v, actual.IsNil()=%v", path, expected.IsNil(), actual.IsNil())
			return false
		}
		return deepCompareRecursive(t, expected.Elem(), actual.Elem(), path+".(interface)", visited, depth+1)
	default:
		// For other types, use reflect.DeepEqual as fallback
		if !reflect.DeepEqual(expected.Interface(), actual.Interface()) {
			t.Logf("DIFF at %s: expected=%v, actual=%v (using reflect.DeepEqual)", path, expected.Interface(), actual.Interface())
			return false
		}
	}
	return true
}
