package templates

import "testing"

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "jar file should be binary",
			path:     "templates/javav2/auxiliary/gradle/wrapper/gradle-wrapper.jar",
			expected: true,
		},
		{
			name:     "another jar file should be binary",
			path:     "templates/javav2/testusage/gradle/wrapper/gradle-wrapper.jar",
			expected: true,
		},
		{
			name:     "go file should not be binary",
			path:     "templates/go/common.go",
			expected: false,
		},
		{
			name:     "typescript file should not be binary",
			path:     "templates/typescriptv2/index.ts",
			expected: false,
		},
		{
			name:     "unknown file should not be binary",
			path:     "templates/unknown/file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBinary(tt.path)
			if result != tt.expected {
				t.Errorf("IsBinary(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsBinaryExhaustive(t *testing.T) {
	// Test that all known binary files return true
	binaryFiles := []string{
		"templates/javav2/auxiliary/gradle/wrapper/gradle-wrapper.jar",
		"templates/javav2/testusage/gradle/wrapper/gradle-wrapper.jar",
	}

	for _, path := range binaryFiles {
		t.Run(path, func(t *testing.T) {
			if !IsBinary(path) {
				t.Errorf("IsBinary(%q) should return true for known binary file", path)
			}
		})
	}
}
