package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeDefExtensions_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *TypeDefExtensions
		test func(t *testing.T, original, cloned *TypeDefExtensions)
	}{
		{
			name: "nil TypeDefExtensions",
			orig: nil,
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "empty TypeDefExtensions",
			orig: &TypeDefExtensions{},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Nil(t, cloned.All)
				assert.Nil(t, cloned.Entity)
				assert.Nil(t, cloned.EntityDescription)
				assert.Nil(t, cloned.EntityVersion)
				assert.Nil(t, cloned.ExampleUnset)
				assert.Nil(t, cloned.Ignore)
				assert.Nil(t, cloned.Pagination)
				assert.Nil(t, cloned.TerraformCustomDefault)
				assert.Nil(t, cloned.TerraformIgnore)
			},
		},
		{
			name: "TypeDefExtensions with All map",
			orig: &TypeDefExtensions{
				All: map[string]any{
					"x-custom-1": "value1",
					"x-custom-2": 123,
					"x-custom-3": true,
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.All)
				// Maps are not comparable with NotSame, but we can verify they're equal
				assert.Len(t, cloned.All, len(original.All))
				for key, val := range original.All {
					assert.Equal(t, val, cloned.All[key])
				}
			},
		},
		{
			name: "TypeDefExtensions with Entity",
			orig: &TypeDefExtensions{
				Entity: &extensions.Entity{
					Names: []string{"user", "account"},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.Entity)
				assert.NotSame(t, original.Entity, cloned.Entity)
				for i, name := range original.Entity.Names {
					assert.Equal(t, name, cloned.Entity.Names[i])
				}
				assert.Equal(t, original.Entity.Names, cloned.Entity.Names)
			},
		},
		{
			name: "TypeDefExtensions with EntityDescription",
			orig: &TypeDefExtensions{
				EntityDescription: &extensions.EntityDescription{
					TerraformDataResource:    "data_user",
					TerraformManagedResource: "user",
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.EntityDescription)
				assert.NotSame(t, original.EntityDescription, cloned.EntityDescription)
				assert.Equal(t, original.EntityDescription.TerraformDataResource, cloned.EntityDescription.TerraformDataResource)
				assert.Equal(t, original.EntityDescription.TerraformManagedResource, cloned.EntityDescription.TerraformManagedResource)
			},
		},
		{
			name: "TypeDefExtensions with EntityVersion",
			orig: &TypeDefExtensions{
				EntityVersion: &extensions.EntityVersion{
					TerraformManagedResource: 42,
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.EntityVersion)
				assert.NotSame(t, original.EntityVersion, cloned.EntityVersion)
				assert.Equal(t, original.EntityVersion.TerraformManagedResource, cloned.EntityVersion.TerraformManagedResource)
			},
		},
		{
			name: "TypeDefExtensions with MatchConfig",
			orig: &TypeDefExtensions{
				MatchConfig: &extensions.MatchConfig{Path: ptr("some-match-config")},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.MatchConfig)
				assert.NotSame(t, original.MatchConfig, cloned.MatchConfig)
				assert.Equal(t, *original.MatchConfig, *cloned.MatchConfig)
			},
		},
		{
			name: "TypeDefExtensions with ExampleUnset",
			orig: &TypeDefExtensions{
				ExampleUnset: ptr(true),
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.ExampleUnset)
				assert.NotSame(t, original.ExampleUnset, cloned.ExampleUnset)
				assert.Equal(t, *original.ExampleUnset, *cloned.ExampleUnset)
			},
		},
		{
			name: "TypeDefExtensions with Ignore true",
			orig: &TypeDefExtensions{
				Ignore: ptr(true),
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.Ignore)
				assert.NotSame(t, original.Ignore, cloned.Ignore)
				assert.Equal(t, *original.Ignore, *cloned.Ignore)
			},
		},
		{
			name: "TypeDefExtensions with Ignore false",
			orig: &TypeDefExtensions{
				Ignore: ptr(false),
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.Ignore)
				assert.NotSame(t, original.Ignore, cloned.Ignore)
				assert.Equal(t, *original.Ignore, *cloned.Ignore)
			},
		},
		{
			name: "TypeDefExtensions with Pagination",
			orig: &TypeDefExtensions{
				Pagination: &extensions.Pagination{
					Type: extensions.PaginationTypeOffsetLimit,
					Inputs: []extensions.PaginationInputs{
						{
							Name: "offset",
							In:   extensions.PaginationInputInTypeParameters,
							Type: extensions.PaginationInputTypeOffset,
						},
						{
							Name:     "limit",
							In:       extensions.PaginationInputInTypeParameters,
							Type:     extensions.PaginationInputTypeLimit,
							Optional: true,
						},
					},
					Outputs: extensions.PaginationOutputs{
						NextURL:    "$.next",
						NextCursor: "$.cursor",
						Results:    "$.items",
						NumPages:   "$.totalPages",
					},
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.Pagination)
				assert.NotSame(t, original.Pagination, cloned.Pagination)
				assert.Equal(t, original.Pagination.Type, cloned.Pagination.Type)

				if original.Pagination.Inputs != nil {
					require.NotNil(t, cloned.Pagination.Inputs)
					assert.NotSame(t, &original.Pagination.Inputs, &cloned.Pagination.Inputs)
					assert.Equal(t, original.Pagination.Inputs, cloned.Pagination.Inputs)
				}

				assert.Equal(t, original.Pagination.Outputs.NextURL, cloned.Pagination.Outputs.NextURL)
				assert.Equal(t, original.Pagination.Outputs.NextCursor, cloned.Pagination.Outputs.NextCursor)
				assert.Equal(t, original.Pagination.Outputs.Results, cloned.Pagination.Outputs.Results)
				assert.Equal(t, original.Pagination.Outputs.NumPages, cloned.Pagination.Outputs.NumPages)
			},
		},
		{
			name: "TypeDefExtensions with TerraformCustomDefault",
			orig: &TypeDefExtensions{
				TerraformCustomDefault: &extensions.TerraformCustomDefault{
					Imports:          []string{"fmt", "strings"},
					SchemaDefinition: "stringdefault.StaticString(\"default value\")",
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.TerraformCustomDefault)
				assert.NotSame(t, original.TerraformCustomDefault, cloned.TerraformCustomDefault)
				for i, imp := range original.TerraformCustomDefault.Imports {
					assert.Equal(t, imp, cloned.TerraformCustomDefault.Imports[i])
				}
				assert.Equal(t, original.TerraformCustomDefault.Imports, cloned.TerraformCustomDefault.Imports)
				assert.Equal(t, original.TerraformCustomDefault.SchemaDefinition, cloned.TerraformCustomDefault.SchemaDefinition)
			},
		},
		{
			name: "TypeDefExtensions with TerraformIgnore",
			orig: &TypeDefExtensions{
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: true,
					Schema:    true,
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.TerraformIgnore)
				assert.NotSame(t, original.TerraformIgnore, cloned.TerraformIgnore)
				assert.Equal(t, original.TerraformIgnore.DataModel, cloned.TerraformIgnore.DataModel)
				assert.Equal(t, original.TerraformIgnore.Schema, cloned.TerraformIgnore.Schema)

				// Verify deep copy by modifying original
				original.TerraformIgnore.DataModel = false
				assert.True(t, cloned.TerraformIgnore.DataModel)
			},
		},
		{
			name: "TypeDefExtensions with TerraformIgnore schema only",
			orig: &TypeDefExtensions{
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: false,
					Schema:    true,
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				require.NotNil(t, cloned.TerraformIgnore)
				assert.NotSame(t, original.TerraformIgnore, cloned.TerraformIgnore)
				assert.False(t, cloned.TerraformIgnore.DataModel)
				assert.True(t, cloned.TerraformIgnore.Schema)
			},
		},
		{
			name: "TypeDefExtensions with all fields",
			orig: &TypeDefExtensions{
				All: map[string]any{
					"x-custom": "value",
				},
				Entity: &extensions.Entity{
					Names: []string{"item"},
				},
				EntityDescription: &extensions.EntityDescription{
					TerraformDataResource:    "data_item",
					TerraformManagedResource: "item",
				},
				EntityVersion: &extensions.EntityVersion{
					TerraformManagedResource: 1,
				},
				ExampleUnset: ptr(false),
				Ignore:       ptr(true),
				Pagination: &extensions.Pagination{
					Type: extensions.PaginationTypeCursor,
				},
				TerraformCustomDefault: &extensions.TerraformCustomDefault{
					SchemaDefinition: "default value",
				},
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: false,
					Schema:    true,
				},
			},
			test: func(t *testing.T, original, cloned *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)

				require.NotNil(t, cloned.All)
				// Maps are not comparable with NotSame, but we can verify they're equal
				assert.Equal(t, original.All, cloned.All)

				require.NotNil(t, cloned.Entity)
				assert.NotSame(t, original.Entity, cloned.Entity)

				require.NotNil(t, cloned.EntityDescription)
				assert.NotSame(t, original.EntityDescription, cloned.EntityDescription)

				require.NotNil(t, cloned.EntityVersion)
				assert.NotSame(t, original.EntityVersion, cloned.EntityVersion)

				require.NotNil(t, cloned.ExampleUnset)
				assert.NotSame(t, original.ExampleUnset, cloned.ExampleUnset)
				assert.Equal(t, *original.ExampleUnset, *cloned.ExampleUnset)

				require.NotNil(t, cloned.Ignore)
				assert.NotSame(t, original.Ignore, cloned.Ignore)
				assert.Equal(t, *original.Ignore, *cloned.Ignore)

				require.NotNil(t, cloned.Pagination)
				assert.NotSame(t, original.Pagination, cloned.Pagination)

				require.NotNil(t, cloned.TerraformCustomDefault)
				assert.NotSame(t, original.TerraformCustomDefault, cloned.TerraformCustomDefault)

				require.NotNil(t, cloned.TerraformIgnore)
				assert.NotSame(t, original.TerraformIgnore, cloned.TerraformIgnore)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.orig.Clone()
			tt.test(t, tt.orig, cloned)
		})
	}
}

func TestTypeDefExtensions_Get(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extensions    *TypeDefExtensions
		name          string
		expectedValue any
		expectedOK    bool
	}{
		"nil All": {
			extensions:    &TypeDefExtensions{},
			name:          "x-test",
			expectedValue: nil,
			expectedOK:    false,
		},
		"missing key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-other": "value"},
			},
			name:          "x-test",
			expectedValue: nil,
			expectedOK:    false,
		},
		"present key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-test": "found"},
			},
			name:          "x-test",
			expectedValue: "found",
			expectedOK:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, ok := tt.extensions.Get(tt.name)
			assert.Equal(t, tt.expectedValue, value)
			assert.Equal(t, tt.expectedOK, ok)
		})
	}
}

func TestTypeDefExtensions_Has(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extensions *TypeDefExtensions
		name       string
		expected   bool
	}{
		"nil All": {
			extensions: &TypeDefExtensions{},
			name:       "x-test",
			expected:   false,
		},
		"missing key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-other": "value"},
			},
			name:     "x-test",
			expected: false,
		},
		"present key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-test": true},
			},
			name:     "x-test",
			expected: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.extensions.Has(tt.name))
		})
	}
}

func TestTypeDefExtensions_Remove(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extensions *TypeDefExtensions
		name       string
		validate   func(t *testing.T, e *TypeDefExtensions)
	}{
		"nil All": {
			extensions: &TypeDefExtensions{},
			name:       "x-test",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Nil(t, e.All)
			},
		},
		"missing key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-other": "value"},
			},
			name: "x-test",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, map[string]any{"x-other": "value"}, e.All)
			},
		},
		"present key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-test": true, "x-other": "value"},
			},
			name: "x-test",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, map[string]any{"x-other": "value"}, e.All)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tt.extensions.Remove(tt.name)
			tt.validate(t, tt.extensions)
		})
	}
}

func TestTypeDefExtensions_Rename(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extensions *TypeDefExtensions
		from       string
		to         string
		validate   func(t *testing.T, e *TypeDefExtensions)
	}{
		"nil All": {
			extensions: &TypeDefExtensions{},
			from:       "x-old",
			to:         "x-new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Nil(t, e.All)
			},
		},
		"missing key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-other": "value"},
			},
			from: "x-old",
			to:   "x-new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, map[string]any{"x-other": "value"}, e.All)
			},
		},
		"false value stays": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-old": false},
			},
			from: "x-old",
			to:   "x-new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, false, e.All["x-old"])
				assert.Nil(t, e.All["x-new"])
			},
		},
		"nil value stays": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-old": nil},
			},
			from: "x-old",
			to:   "x-new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Contains(t, e.All, "x-old")
				assert.NotContains(t, e.All, "x-new")
			},
		},
		"truthy value moves": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-old": "value"},
			},
			from: "x-old",
			to:   "x-new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.NotContains(t, e.All, "x-old")
				assert.Equal(t, "value", e.All["x-new"])
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tt.extensions.Rename(tt.from, tt.to)
			tt.validate(t, tt.extensions)
		})
	}
}

func TestTypeDefExtensions_Set(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extensions *TypeDefExtensions
		name       string
		value      any
		validate   func(t *testing.T, e *TypeDefExtensions)
	}{
		"nil All auto-initializes": {
			extensions: &TypeDefExtensions{},
			name:       "x-test",
			value:      true,
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, e.All)
				assert.Equal(t, true, e.All["x-test"])
			},
		},
		"new key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-existing": "value"},
			},
			name:  "x-new",
			value: 42,
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, "value", e.All["x-existing"])
				assert.Equal(t, 42, e.All["x-new"])
			},
		},
		"overwrite existing": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-test": "old"},
			},
			name:  "x-test",
			value: "new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, "new", e.All["x-test"])
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tt.extensions.Set(tt.name, tt.value)
			tt.validate(t, tt.extensions)
		})
	}
}

func TestTypeDefExtensions_SetIfAbsent(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		extensions *TypeDefExtensions
		name       string
		value      any
		validate   func(t *testing.T, e *TypeDefExtensions)
	}{
		"nil All auto-initializes": {
			extensions: &TypeDefExtensions{},
			name:       "x-test",
			value:      true,
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, e.All)
				assert.Equal(t, true, e.All["x-test"])
			},
		},
		"new key": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-existing": "value"},
			},
			name:  "x-new",
			value: 42,
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, "value", e.All["x-existing"])
				assert.Equal(t, 42, e.All["x-new"])
			},
		},
		"existing key preserved": {
			extensions: &TypeDefExtensions{
				All: map[string]any{"x-test": "original"},
			},
			name:  "x-test",
			value: "new",
			validate: func(t *testing.T, e *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, "original", e.All["x-test"])
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tt.extensions.SetIfAbsent(tt.name, tt.value)
			tt.validate(t, tt.extensions)
		})
	}
}

func TestTypeDefExtensions_MergeWithoutOverwrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		existing         *TypeDefExtensions
		oaExtensionsYAML string
		expected         *TypeDefExtensions
		expectError      bool
	}{
		{
			name:             "nil existing extensions",
			existing:         nil,
			oaExtensionsYAML: ``,
			expected:         nil,
			expectError:      false,
		},
		{
			name:             "empty OAExtensions",
			existing:         &TypeDefExtensions{All: map[string]any{}},
			oaExtensionsYAML: ``,
			expected:         &TypeDefExtensions{All: map[string]any{}},
			expectError:      false,
		},
		{
			name: "merge Entity when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-entity:
  - user
  - account
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity": []any{"user", "account"},
				},
				Entity: &extensions.Entity{
					Names: []string{"account", "user"},
				},
			},
			expectError: false,
		},
		{
			name: "do not overwrite existing Entity",
			existing: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity": []any{"original"},
				},
				Entity: &extensions.Entity{
					Names: []string{"original"},
				},
			},
			oaExtensionsYAML: `
x-speakeasy-entity:
  - new
  - entity
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity": []any{"original"},
				},
				Entity: &extensions.Entity{
					Names: []string{"original"},
				},
			},
			expectError: false,
		},
		{
			name: "merge multiple fields when existing are nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-entity:
  - item
x-speakeasy-example-unset: true
x-speakeasy-match: "pattern"
x-speakeasy-terraform-alias-to: "old_field"
x-custom-extension: "custom-value"
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity":             []any{"item"},
					"x-speakeasy-example-unset":      true,
					"x-speakeasy-match":              "pattern",
					"x-speakeasy-terraform-alias-to": "old_field",
					"x-custom-extension":             "custom-value",
				},
				Entity: &extensions.Entity{
					Names: []string{"item"},
				},
				ExampleUnset:     ptr(true),
				MatchConfig:      &extensions.MatchConfig{Path: ptr("pattern")},
				TerraformAliasTo: ptr("old_field"),
			},
			expectError: false,
		},
		{
			name: "merge PublicExports when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-exports:
  - group: chat.completions
    name: chat_completion_audio
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-exports": []any{
						map[string]any{
							"group": "chat.completions",
							"name":  "chat_completion_audio",
						},
					},
				},
				PublicExports: []extensions.PublicExport{
					{Group: "chat.completions", Name: "chat_completion_audio"},
				},
			},
			expectError: false,
		},
		{
			name: "partially merge - only add missing fields",
			existing: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity":        []any{"existing"},
					"x-speakeasy-example-unset": false,
				},
				Entity:       &extensions.Entity{Names: []string{"existing"}},
				ExampleUnset: ptr(false),
			},
			oaExtensionsYAML: `
x-speakeasy-entity:
  - item
x-speakeasy-example-unset: true
x-speakeasy-match: "pattern"
x-speakeasy-terraform-alias-to: "old_field"
x-custom-extension: "custom-value"
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity":             []any{"existing"},
					"x-speakeasy-example-unset":      false,
					"x-speakeasy-match":              "pattern",
					"x-speakeasy-terraform-alias-to": "old_field",
					"x-custom-extension":             "custom-value",
				},
				Entity:           &extensions.Entity{Names: []string{"existing"}},
				ExampleUnset:     ptr(false),
				MatchConfig:      &extensions.MatchConfig{Path: ptr("pattern")},
				TerraformAliasTo: ptr("old_field"),
			},
			expectError: false,
		},
		{
			name: "merge TerraformCustomDefault when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-terraform-custom-default:
  imports:
    - fmt
    - strings
  schemaDefinition: 'stringdefault.StaticString("default")'
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-custom-default": map[string]any{
						"imports":          []any{"fmt", "strings"},
						"schemaDefinition": "stringdefault.StaticString(\"default\")",
					},
				},
				TerraformCustomDefault: &extensions.TerraformCustomDefault{
					Imports:          []string{"fmt", "strings"},
					SchemaDefinition: "stringdefault.StaticString(\"default\")",
				},
			},
			expectError: false,
		},
		{
			name: "do not overwrite existing TerraformCustomDefault",
			existing: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-custom-default": map[string]any{
						"schemaDefinition": "existing",
					},
				},
				TerraformCustomDefault: &extensions.TerraformCustomDefault{
					SchemaDefinition: "existing",
				},
			},
			oaExtensionsYAML: `
x-speakeasy-terraform-custom-default:
  imports:
    - fmt
    - strings
  schemaDefinition: 'stringdefault.StaticString("default")'
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-custom-default": map[string]any{
						"schemaDefinition": "existing",
					},
				},
				TerraformCustomDefault: &extensions.TerraformCustomDefault{
					SchemaDefinition: "existing",
				},
			},
			expectError: false,
		},
		{
			name: "merge TerraformIgnore when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-terraform-ignore: true
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-ignore": true,
				},
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: true,
					Schema:    true,
				},
			},
			expectError: false,
		},
		{
			name: "merge TerraformIgnore schema only when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-terraform-ignore: schema
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-ignore": "schema",
				},
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: false,
					Schema:    true,
				},
			},
			expectError: false,
		},
		{
			name: "do not overwrite existing TerraformIgnore",
			existing: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-ignore": "schema",
				},
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: false,
					Schema:    true,
				},
			},
			oaExtensionsYAML: `
x-speakeasy-terraform-ignore: true
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-terraform-ignore": "schema",
				},
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: false,
					Schema:    true,
				},
			},
			expectError: false,
		},
		{
			name: "merge EntityDescription and EntityVersion",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-entity-description: "user"
x-speakeasy-entity-version: 42
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-entity-description": "user",
					"x-speakeasy-entity-version":     42,
				},
				EntityDescription: &extensions.EntityDescription{
					TerraformAction:            "user",
					TerraformDataResource:      "user",
					TerraformEphemeralResource: "user",
					TerraformManagedResource:   "user",
				},
				EntityVersion: &extensions.EntityVersion{
					TerraformManagedResource: 42,
				},
			},
			expectError: false,
		},
		{
			name: "do not overwrite All map existing entries",
			existing: &TypeDefExtensions{
				All: map[string]any{
					"x-custom-1": "existing-value",
					"x-custom-2": 123,
				},
			},
			oaExtensionsYAML: `
x-custom-1: "new-value"
x-custom-2: 456
x-custom-3: "added-value"
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-custom-1": "existing-value",
					"x-custom-2": 123,
					"x-custom-3": "added-value",
				},
			},
			expectError: false,
		},
		{
			name: "merge Ignore true when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-ignore: true
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-ignore": true,
				},
				Ignore: ptr(true),
			},
			expectError: false,
		},
		{
			name: "merge Ignore false when existing is nil",
			existing: &TypeDefExtensions{
				All: map[string]any{},
			},
			oaExtensionsYAML: `
x-speakeasy-ignore: false
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-ignore": false,
				},
				Ignore: ptr(false),
			},
			expectError: false,
		},
		{
			name: "do not overwrite existing Ignore",
			existing: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-ignore": false,
				},
				Ignore: ptr(false),
			},
			oaExtensionsYAML: `
x-speakeasy-ignore: true
`,
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-speakeasy-ignore": false,
				},
				Ignore: ptr(false),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Convert YAML to OAExtensions
			oaExtensions := yamlToOAExtensions(t, tt.oaExtensionsYAML)

			// Create a minimal TypeDef for testing
			typeDef := &TypeDef{
				Extensions: tt.existing,
			}

			// Create an Extensions instance for handling
			ext := &extensions.Extensions{}

			err := tt.existing.MergeWithoutOverwrite(ext, typeDef, oaExtensions)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected != nil {
					assert.Equal(t, tt.expected, tt.existing)
				}
			}
		})
	}
}

func TestTypeDefExtensions_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		e1       *TypeDefExtensions
		e2       *TypeDefExtensions
		expected *TypeDefExtensions
	}{
		{
			name:     "merges nil extensions",
			e1:       nil,
			e2:       &TypeDefExtensions{AllowEmptyValue: true},
			expected: nil,
		},
		{
			name:     "merges with nil other",
			e1:       &TypeDefExtensions{AllowEmptyValue: true},
			e2:       nil,
			expected: &TypeDefExtensions{AllowEmptyValue: true},
		},
		{
			name: "merges All map",
			e1: &TypeDefExtensions{
				All: map[string]any{"x-custom": "value1"},
			},
			e2: &TypeDefExtensions{
				All: map[string]any{"x-other": "value2"},
			},
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-custom": "value1",
					"x-other":  "value2",
				},
			},
		},
		{
			name: "preserves existing All map entries",
			e1: &TypeDefExtensions{
				All: map[string]any{"x-custom": "original"},
			},
			e2: &TypeDefExtensions{
				All: map[string]any{"x-custom": "new"},
			},
			expected: &TypeDefExtensions{
				All: map[string]any{"x-custom": "original"},
			},
		},
		{
			name: "merges AllowEmptyValue",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{AllowEmptyValue: true},
			expected: &TypeDefExtensions{
				AllowEmptyValue: true,
			},
		},
		{
			name: "preserves existing AllowEmptyValue",
			e1:   &TypeDefExtensions{AllowEmptyValue: true},
			e2:   &TypeDefExtensions{AllowEmptyValue: false},
			expected: &TypeDefExtensions{
				AllowEmptyValue: true,
			},
		},
		{
			name: "merges Entity",
			e1:   &TypeDefExtensions{},
			e2: &TypeDefExtensions{
				Entity: &extensions.Entity{Names: []string{"entity1"}},
			},
			expected: &TypeDefExtensions{
				Entity: &extensions.Entity{Names: []string{"entity1"}},
			},
		},
		{
			name: "merges EntityDescription",
			e1:   &TypeDefExtensions{},
			e2: &TypeDefExtensions{
				EntityDescription: &extensions.EntityDescription{
					TerraformDataResource: "data_resource",
				},
			},
			expected: &TypeDefExtensions{
				EntityDescription: &extensions.EntityDescription{
					TerraformDataResource: "data_resource",
				},
			},
		},
		{
			name: "merges EntityVersion",
			e1:   &TypeDefExtensions{},
			e2: &TypeDefExtensions{
				EntityVersion: &extensions.EntityVersion{
					TerraformManagedResource: 42,
				},
			},
			expected: &TypeDefExtensions{
				EntityVersion: &extensions.EntityVersion{
					TerraformManagedResource: 42,
				},
			},
		},
		{
			name: "merges ExampleUnset",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{ExampleUnset: ptr(true)},
			expected: &TypeDefExtensions{
				ExampleUnset: ptr(true),
			},
		},
		{
			name: "merges Ignore",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{Ignore: ptr(true)},
			expected: &TypeDefExtensions{
				Ignore: ptr(true),
			},
		},
		{
			name: "merges MatchConfig",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{MatchConfig: &extensions.MatchConfig{}},
			expected: &TypeDefExtensions{
				MatchConfig: &extensions.MatchConfig{},
			},
		},
		{
			name: "merges ModelNamespace",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{ModelNamespace: ptr("namespace")},
			expected: &TypeDefExtensions{
				ModelNamespace: ptr("namespace"),
			},
		},
		{
			name: "merges OverridableOAuth2Scopes",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{OverridableOAuth2Scopes: true},
			expected: &TypeDefExtensions{
				OverridableOAuth2Scopes: true,
			},
		},
		{
			name: "preserves existing OverridableOAuth2Scopes",
			e1:   &TypeDefExtensions{OverridableOAuth2Scopes: true},
			e2:   &TypeDefExtensions{OverridableOAuth2Scopes: false},
			expected: &TypeDefExtensions{
				OverridableOAuth2Scopes: true,
			},
		},
		{
			name: "merges TerraformAliasTo",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{TerraformAliasTo: ptr("alias")},
			expected: &TypeDefExtensions{
				TerraformAliasTo: ptr("alias"),
			},
		},
		{
			name: "merges TerraformWriteOnly",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{TerraformWriteOnly: ptr(true)},
			expected: &TypeDefExtensions{
				TerraformWriteOnly: ptr(true),
			},
		},
		{
			name: "merges WrappedAttribute",
			e1:   &TypeDefExtensions{},
			e2:   &TypeDefExtensions{WrappedAttribute: ptr("wrapped")},
			expected: &TypeDefExtensions{
				WrappedAttribute: ptr("wrapped"),
			},
		},
		{
			name: "complex merge preserves and adds data",
			e1: &TypeDefExtensions{
				All:             map[string]any{"x-existing": "value"},
				AllowEmptyValue: true,
				Entity:          &extensions.Entity{Names: []string{"entity1"}},
			},
			e2: &TypeDefExtensions{
				All:                map[string]any{"x-new": "value2"},
				EntityDescription:  &extensions.EntityDescription{TerraformDataResource: "data"},
				TerraformWriteOnly: ptr(true),
			},
			expected: &TypeDefExtensions{
				All: map[string]any{
					"x-existing": "value",
					"x-new":      "value2",
				},
				AllowEmptyValue:    true,
				Entity:             &extensions.Entity{Names: []string{"entity1"}},
				EntityDescription:  &extensions.EntityDescription{TerraformDataResource: "data"},
				TerraformWriteOnly: ptr(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone e1 to avoid mutation affecting the test table
			var testE1 *TypeDefExtensions
			if tt.e1 != nil {
				testE1 = tt.e1.Clone()
			}

			testE1.Merge(tt.e2)

			if tt.expected == nil {
				assert.Nil(t, testE1)
			} else {
				require.NotNil(t, testE1)

				// Compare fields
				if tt.expected.All != nil {
					assert.Equal(t, tt.expected.All, testE1.All)
				}
				assert.Equal(t, tt.expected.AllowEmptyValue, testE1.AllowEmptyValue)
				assert.Equal(t, tt.expected.Entity, testE1.Entity)
				assert.Equal(t, tt.expected.EntityDescription, testE1.EntityDescription)
				assert.Equal(t, tt.expected.EntityVersion, testE1.EntityVersion)
				assert.Equal(t, tt.expected.ExampleUnset, testE1.ExampleUnset)
				assert.Equal(t, tt.expected.Ignore, testE1.Ignore)
				assert.Equal(t, tt.expected.MatchConfig, testE1.MatchConfig)
				assert.Equal(t, tt.expected.ModelNamespace, testE1.ModelNamespace)
				assert.Equal(t, tt.expected.OverridableOAuth2Scopes, testE1.OverridableOAuth2Scopes)
				assert.Equal(t, tt.expected.Pagination, testE1.Pagination)
				assert.Equal(t, tt.expected.TerraformAliasTo, testE1.TerraformAliasTo)
				assert.Equal(t, tt.expected.TerraformCustomDefault, testE1.TerraformCustomDefault)
				assert.Equal(t, tt.expected.TerraformIgnore, testE1.TerraformIgnore)
				assert.Equal(t, tt.expected.TerraformWriteOnly, testE1.TerraformWriteOnly)
				assert.Equal(t, tt.expected.WrappedAttribute, testE1.WrappedAttribute)
			}
		})
	}
}

func TestTypeDefExtensions_TerraformMerge(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		e1       *TypeDefExtensions
		e2       *TypeDefExtensions
		validate func(t *testing.T, result *TypeDefExtensions)
	}{
		"nil receiver": {
			e1: nil,
			e2: &TypeDefExtensions{AllowEmptyValue: true},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		"nil other": {
			e1: &TypeDefExtensions{AllowEmptyValue: true},
			e2: nil,
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				assert.True(t, result.AllowEmptyValue)
			},
		},
		"base merge behavior preserved": {
			e1: &TypeDefExtensions{
				All: map[string]any{"x-custom": "value1"},
			},
			e2: &TypeDefExtensions{
				All: map[string]any{"x-other": "value2"},
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				assert.Equal(t, "value1", result.All["x-custom"])
				assert.Equal(t, "value2", result.All["x-other"])
			},
		},
		"Ignore other wins": {
			e1: &TypeDefExtensions{
				Ignore: ptr(false),
			},
			e2: &TypeDefExtensions{
				Ignore: ptr(true),
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.Ignore)
				assert.True(t, *result.Ignore)
			},
		},
		"Ignore preserved when other nil": {
			e1: &TypeDefExtensions{
				Ignore: ptr(true),
			},
			e2: &TypeDefExtensions{},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.Ignore)
				assert.True(t, *result.Ignore)
			},
		},
		"TerraformIgnore OR merges DataModel and Schema": {
			e1: &TypeDefExtensions{
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: true,
					Schema:    false,
				},
			},
			e2: &TypeDefExtensions{
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: false,
					Schema:    true,
				},
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.TerraformIgnore)
				assert.True(t, result.TerraformIgnore.DataModel)
				assert.True(t, result.TerraformIgnore.Schema)
			},
		},
		"TerraformIgnore sets from other when receiver nil": {
			e1: &TypeDefExtensions{},
			e2: &TypeDefExtensions{
				TerraformIgnore: &extensions.TerraformIgnore{
					DataModel: true,
				},
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.TerraformIgnore)
				assert.True(t, result.TerraformIgnore.DataModel)
				assert.False(t, result.TerraformIgnore.Schema)
			},
		},
		"TerraformWriteOnly other wins": {
			e1: &TypeDefExtensions{
				TerraformWriteOnly: ptr(false),
			},
			e2: &TypeDefExtensions{
				TerraformWriteOnly: ptr(true),
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.TerraformWriteOnly)
				assert.True(t, *result.TerraformWriteOnly)
			},
		},
		"TerraformWriteOnly preserved when other nil": {
			e1: &TypeDefExtensions{
				TerraformWriteOnly: ptr(true),
			},
			e2: &TypeDefExtensions{},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.TerraformWriteOnly)
				assert.True(t, *result.TerraformWriteOnly)
			},
		},
		"TerraformHoistedFrom other wins": {
			e1: &TypeDefExtensions{
				TerraformHoistedFrom: []TerraformHoistedSource{{FieldName: "original"}},
			},
			e2: &TypeDefExtensions{
				TerraformHoistedFrom: []TerraformHoistedSource{{FieldName: "updated"}},
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.Len(t, result.TerraformHoistedFrom, 1)
				assert.Equal(t, "updated", result.TerraformHoistedFrom[0].FieldName)
			},
		},
		"TerraformHoistedFrom preserved when other nil": {
			e1: &TypeDefExtensions{
				TerraformHoistedFrom: []TerraformHoistedSource{{FieldName: "keep"}},
			},
			e2: &TypeDefExtensions{},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.Len(t, result.TerraformHoistedFrom, 1)
				assert.Equal(t, "keep", result.TerraformHoistedFrom[0].FieldName)
			},
		},
		"TerraformAliasTo preserves existing": {
			e1: &TypeDefExtensions{
				TerraformAliasTo: ptr("original"),
			},
			e2: &TypeDefExtensions{
				TerraformAliasTo: ptr("other"),
			},
			validate: func(t *testing.T, result *TypeDefExtensions) {
				t.Helper()
				require.NotNil(t, result.TerraformAliasTo)
				assert.Equal(t, "original", *result.TerraformAliasTo)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var testE1 *TypeDefExtensions
			if tt.e1 != nil {
				testE1 = tt.e1.Clone()
			}

			testE1.TerraformMerge(tt.e2)
			tt.validate(t, testE1)
		})
	}
}
