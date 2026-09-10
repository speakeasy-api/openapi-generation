package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestValidateModelNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		namespace   *string
		wantErr     bool
		errContains string
	}{
		{
			name:      "nil namespace",
			namespace: nil,
			wantErr:   false,
		},
		{
			name:      "valid simple namespace",
			namespace: ptr("foo"),
			wantErr:   false,
		},
		{
			name:      "valid namespace with underscore",
			namespace: ptr("foo_bar"),
			wantErr:   false,
		},
		{
			name:      "valid namespace with hyphen",
			namespace: ptr("foo-bar"),
			wantErr:   false,
		},
		{
			name:      "valid namespace with dot",
			namespace: ptr("foo.bar"),
			wantErr:   false,
		},
		{
			name:      "valid namespace with numbers",
			namespace: ptr("foo123"),
			wantErr:   false,
		},
		{
			name:      "valid namespace with mixed valid characters",
			namespace: ptr("Foo_Bar-123.baz"),
			wantErr:   false,
		},
		{
			name:        "empty namespace",
			namespace:   ptr(""),
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "namespace with forward slash",
			namespace:   ptr("foo/bar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with nested path",
			namespace:   ptr("foo/bar/baz"),
			wantErr:     true,
			errContains: "Nested namespaces using forward slashes",
		},
		{
			name:        "namespace with backslash",
			namespace:   ptr("foo\\bar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with newline",
			namespace:   ptr("foo\nbar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with tab",
			namespace:   ptr("foo\tbar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with space",
			namespace:   ptr("foo bar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with colon",
			namespace:   ptr("foo:bar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with at sign",
			namespace:   ptr("foo@bar"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name:        "namespace with special characters",
			namespace:   ptr("foo#bar$baz"),
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateModelNamespace(tt.namespace)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetModelNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		yaml        string
		expected    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "no extension",
			yaml:     `type: object`,
			expected: "",
			wantErr:  false,
		},
		{
			name: "valid namespace",
			yaml: `
type: object
x-speakeasy-model-namespace: foo`,
			expected: "foo",
			wantErr:  false,
		},
		{
			name: "valid namespace with underscore",
			yaml: `
type: object
x-speakeasy-model-namespace: foo_bar`,
			expected: "foo_bar",
			wantErr:  false,
		},
		{
			name: "invalid namespace with slash",
			yaml: `
type: object
x-speakeasy-model-namespace: foo/bar`,
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
		{
			name: "invalid namespace with nested path",
			yaml: `
type: object
x-speakeasy-model-namespace: foo/bar/baz`,
			wantErr:     true,
			errContains: "Nested namespaces using forward slashes",
		},
		{
			name: "invalid empty namespace",
			yaml: `
type: object
x-speakeasy-model-namespace: ""`,
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name: "invalid namespace with space",
			yaml: `
type: object
x-speakeasy-model-namespace: "foo bar"`,
			wantErr:     true,
			errContains: "must contain only alphanumeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse the YAML
			var schemaNode yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &schemaNode)
			require.NoError(t, err)

			// Build the extensions map from the YAML
			extensions := extensions.New()
			if schemaNode.Kind == yaml.DocumentNode && len(schemaNode.Content) > 0 {
				mappingNode := schemaNode.Content[0]
				if mappingNode.Kind == yaml.MappingNode {
					for i := 0; i < len(mappingNode.Content); i += 2 {
						key := mappingNode.Content[i].Value
						val := mappingNode.Content[i+1]
						extensions.Set(key, val)
					}
				}
			}

			// Create Extensions instance and call GetModelNamespace
			e := New(types.Target{Target: "go"})
			result, err := e.GetModelNamespace(extensions)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestModelNamespaceValidCharsRegex(t *testing.T) {
	t.Parallel()

	// Test the regex directly to document expected behavior
	tests := []struct {
		input string
		valid bool
	}{
		// Valid cases
		{"foo", true},
		{"Foo", true},
		{"FOO", true},
		{"foo123", true},
		{"123foo", true},
		{"foo_bar", true},
		{"foo-bar", true},
		{"foo.bar", true},
		{"Foo_Bar-123.baz", true},
		{"a", true},
		{"A", true},
		{"1", true},
		{"_", true},
		{"-", true},
		{".", true},

		// Invalid cases
		{"", false},              // empty
		{"foo/bar", false},       // forward slash
		{"foo\\bar", false},      // backslash
		{"foo bar", false},       // space
		{"foo\nbar", false},      // newline
		{"foo\tbar", false},      // tab
		{"foo\rbar", false},      // carriage return
		{"foo:bar", false},       // colon
		{"foo@bar", false},       // at sign
		{"foo#bar", false},       // hash
		{"foo$bar", false},       // dollar
		{"foo%bar", false},       // percent
		{"foo^bar", false},       // caret
		{"foo&bar", false},       // ampersand
		{"foo*bar", false},       // asterisk
		{"foo(bar", false},       // open paren
		{"foo)bar", false},       // close paren
		{"foo+bar", false},       // plus
		{"foo=bar", false},       // equals
		{"foo[bar", false},       // open bracket
		{"foo]bar", false},       // close bracket
		{"foo{bar", false},       // open brace
		{"foo}bar", false},       // close brace
		{"foo|bar", false},       // pipe
		{"foo;bar", false},       // semicolon
		{"foo'bar", false},       // single quote
		{"foo\"bar", false},      // double quote
		{"foo<bar", false},       // less than
		{"foo>bar", false},       // greater than
		{"foo,bar", false},       // comma
		{"foo?bar", false},       // question mark
		{"foo`bar", false},       // backtick
		{"foo~bar", false},       // tilde
		{"foo!bar", false},       // exclamation
		{"foo/bar/baz", false},   // nested path
		{"foo\\bar\\baz", false}, // Windows-style path
		{" foo", false},          // leading space
		{"foo ", false},          // trailing space
		{"\x00foo", false},       // null byte
		{"foo\x1fbar", false},    // control character
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			result := modelNamespaceValidChars.MatchString(tt.input)
			assert.Equal(t, tt.valid, result, "input: %q", tt.input)
		})
	}
}
