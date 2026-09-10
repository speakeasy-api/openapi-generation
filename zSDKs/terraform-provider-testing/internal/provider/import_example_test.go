package provider_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestImportExampleFromResponseSchema verifies that the import documentation
// uses example values from the response schema when the request schema (path
// parameters) don't have examples.
//
// This is a regression test for response-schema import examples.
//
// The OASExample entity has an `id` field in the response schema with:
//
//	example: '1234567890abcdef'
//
// The import documentation should use this example value instead of the
// default placeholder "...".
func TestImportExampleFromResponseSchema(t *testing.T) {
	t.Parallel()

	// Get the path to the examples directory relative to this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	providerDir := filepath.Dir(filename)
	examplesDir := filepath.Join(providerDir, "..", "..", "examples", "resources", "testing_oas_example")

	tests := []struct {
		name            string
		file            string
		expectedValue   string
		unexpectedValue string
	}{
		{
			name:            "import.sh should contain response example",
			file:            "import.sh",
			expectedValue:   `"1234567890abcdef"`,
			unexpectedValue: `"..."`,
		},
		{
			name:            "import-by-string-id.tf should contain response example",
			file:            "import-by-string-id.tf",
			expectedValue:   `"1234567890abcdef"`,
			unexpectedValue: `"..."`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			filePath := filepath.Join(examplesDir, tc.file)
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("failed to read %s: %v", tc.file, err)
			}

			contentStr := string(content)

			// Verify the expected example value is present
			if !strings.Contains(contentStr, tc.expectedValue) {
				t.Errorf("expected %s to contain %q, got:\n%s", tc.file, tc.expectedValue, contentStr)
			}

			// Verify the placeholder is NOT present (regression check)
			if strings.Contains(contentStr, tc.unexpectedValue) {
				t.Errorf("expected %s to NOT contain placeholder %q, got:\n%s", tc.file, tc.unexpectedValue, contentStr)
			}
		})
	}
}

// TestImportExampleFallbackToPlaceholder verifies that resources without
// response examples still use the placeholder value.
func TestImportExampleFallbackToPlaceholder(t *testing.T) {
	t.Parallel()

	// Get the path to the examples directory relative to this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}
	providerDir := filepath.Dir(filename)
	examplesDir := filepath.Join(providerDir, "..", "..", "examples", "resources", "testing_import_id_string")

	filePath := filepath.Join(examplesDir, "import.sh")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read import.sh: %v", err)
	}

	contentStr := string(content)

	// This resource doesn't have an example in the spec, so it should use the placeholder
	if !strings.Contains(contentStr, `"..."`) {
		t.Errorf("expected import.sh for testing_import_id_string to contain placeholder \"...\", got:\n%s", contentStr)
	}
}
