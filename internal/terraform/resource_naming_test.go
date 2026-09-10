package terraform

import (
	"testing"
)

func TestIsResourceNameValid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid cases - starts with letter
		{
			name:     "valid simple name",
			input:    "resource",
			expected: true,
		},
		{
			name:     "valid name with uppercase",
			input:    "MyResource",
			expected: true,
		},
		{
			name:     "valid name with numbers",
			input:    "resource123",
			expected: true,
		},
		{
			name:     "valid name with underscore",
			input:    "my_resource",
			expected: true,
		},
		{
			name:     "valid name with hyphen",
			input:    "my-resource",
			expected: true,
		},
		{
			name:     "valid name with space",
			input:    "my resource",
			expected: true,
		},
		{
			name:     "valid name with multiple spaces",
			input:    "my resource name",
			expected: true,
		},
		{
			name:     "valid name with all allowed characters",
			input:    "MyResource-123_test name",
			expected: true,
		},
		{
			name:     "valid single uppercase letter followed by number",
			input:    "A1",
			expected: true,
		},
		{
			name:     "valid single lowercase letter followed by number",
			input:    "a1",
			expected: true,
		},
		{
			name:     "valid long name",
			input:    "VeryLongResourceNameWith123NumbersAnd_Underscores-And-Hyphens And Spaces",
			expected: true,
		},

		// Invalid cases - doesn't start with letter
		{
			name:     "invalid starts with number",
			input:    "1resource",
			expected: false,
		},
		{
			name:     "invalid starts with underscore",
			input:    "_resource",
			expected: false,
		},
		{
			name:     "invalid starts with hyphen",
			input:    "-resource",
			expected: false,
		},
		{
			name:     "invalid starts with space",
			input:    " resource",
			expected: false,
		},

		// Invalid cases - contains invalid characters
		{
			name:     "invalid contains exclamation",
			input:    "resource!",
			expected: false,
		},
		{
			name:     "invalid contains at symbol",
			input:    "resource@name",
			expected: false,
		},
		{
			name:     "invalid contains hash",
			input:    "resource#1",
			expected: false,
		},
		{
			name:     "invalid contains dollar sign",
			input:    "resource$",
			expected: false,
		},
		{
			name:     "invalid contains percent",
			input:    "resource%",
			expected: false,
		},
		{
			name:     "invalid contains caret",
			input:    "resource^",
			expected: false,
		},
		{
			name:     "invalid contains ampersand",
			input:    "resource&name",
			expected: false,
		},
		{
			name:     "invalid contains asterisk",
			input:    "resource*",
			expected: false,
		},
		{
			name:     "invalid contains parentheses",
			input:    "resource(1)",
			expected: false,
		},
		{
			name:     "invalid contains brackets",
			input:    "resource[1]",
			expected: false,
		},
		{
			name:     "invalid contains braces",
			input:    "resource{1}",
			expected: false,
		},
		{
			name:     "invalid contains plus",
			input:    "resource+",
			expected: false,
		},
		{
			name:     "invalid contains equals",
			input:    "resource=",
			expected: false,
		},
		{
			name:     "invalid contains pipe",
			input:    "resource|name",
			expected: false,
		},
		{
			name:     "invalid contains backslash",
			input:    "resource\\name",
			expected: false,
		},
		{
			name:     "invalid contains forward slash",
			input:    "resource/name",
			expected: false,
		},
		{
			name:     "invalid contains period",
			input:    "resource.name",
			expected: false,
		},
		{
			name:     "invalid contains comma",
			input:    "resource,name",
			expected: false,
		},
		{
			name:     "invalid contains semicolon",
			input:    "resource;name",
			expected: false,
		},
		{
			name:     "invalid contains colon",
			input:    "resource:name",
			expected: false,
		},
		{
			name:     "invalid contains quote",
			input:    "resource'name",
			expected: false,
		},
		{
			name:     "invalid contains double quote",
			input:    `resource"name`,
			expected: false,
		},
		{
			name:     "invalid contains backtick",
			input:    "resource`name",
			expected: false,
		},
		{
			name:     "invalid contains tilde",
			input:    "resource~name",
			expected: false,
		},
		{
			name:     "invalid contains less than",
			input:    "resource<name",
			expected: false,
		},
		{
			name:     "invalid contains greater than",
			input:    "resource>name",
			expected: false,
		},
		{
			name:     "invalid contains question mark",
			input:    "resource?",
			expected: false,
		},

		// Edge cases
		{
			name:     "invalid empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid single letter",
			input:    "a",
			expected: false,
		},
		{
			name:     "invalid single uppercase letter",
			input:    "A",
			expected: false,
		},
		{
			name:     "invalid single number",
			input:    "1",
			expected: false,
		},
		{
			name:     "invalid single underscore",
			input:    "_",
			expected: false,
		},
		{
			name:     "invalid single hyphen",
			input:    "-",
			expected: false,
		},
		{
			name:     "invalid single space",
			input:    " ",
			expected: false,
		},
		{
			name:     "invalid only spaces",
			input:    "   ",
			expected: false,
		},
		{
			name:     "invalid tab character",
			input:    "resource\tname",
			expected: false,
		},
		{
			name:     "invalid newline character",
			input:    "resource\nname",
			expected: false,
		},
		{
			name:     "invalid carriage return",
			input:    "resource\rname",
			expected: false,
		},
		{
			name:     "invalid unicode character",
			input:    "resource™",
			expected: false,
		},
		{
			name:     "invalid emoji",
			input:    "resource🚀",
			expected: false,
		},
		{
			name:     "invalid non-ASCII letter",
			input:    "resourceñ",
			expected: false,
		},
		{
			name:     "valid two character minimum",
			input:    "ab",
			expected: true,
		},
		{
			name:     "valid starts with uppercase Z",
			input:    "Zresource",
			expected: true,
		},
		{
			name:     "valid starts with lowercase z",
			input:    "zresource",
			expected: true,
		},
		{
			name:     "invalid leading and trailing spaces",
			input:    " resource ",
			expected: false,
		},
		{
			name:     "valid multiple consecutive spaces",
			input:    "resource   name",
			expected: true,
		},
		{
			name:     "valid multiple consecutive underscores",
			input:    "resource___name",
			expected: true,
		},
		{
			name:     "valid multiple consecutive hyphens",
			input:    "resource---name",
			expected: true,
		},
		{
			name:     "valid mixed case throughout",
			input:    "ReSoUrCe123NaMe",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsResourceNameValid(tt.input)
			if result != tt.expected {
				t.Errorf("IsResourceNameValid(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
