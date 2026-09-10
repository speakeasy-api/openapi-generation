package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalDocs_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *ExternalDocs
		test func(t *testing.T, original, cloned *ExternalDocs)
	}{
		{
			name: "nil ExternalDocs",
			orig: nil,
			test: func(t *testing.T, original, cloned *ExternalDocs) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "ExternalDocs with all fields",
			orig: &ExternalDocs{
				Description: "API documentation",
				URL:         "https://example.com/docs",
			},
			test: func(t *testing.T, original, cloned *ExternalDocs) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Description, cloned.Description)
				assert.Equal(t, original.URL, cloned.URL)
			},
		},
		{
			name: "ExternalDocs with empty fields",
			orig: &ExternalDocs{},
			test: func(t *testing.T, original, cloned *ExternalDocs) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Description, cloned.Description)
				assert.Equal(t, original.URL, cloned.URL)
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

func TestExtendedComment_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *ExtendedComment
		test func(t *testing.T, original, cloned *ExtendedComment)
	}{
		{
			name: "nil ExtendedComment",
			orig: nil,
			test: func(t *testing.T, original, cloned *ExtendedComment) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "ExtendedComment with all fields",
			orig: &ExtendedComment{
				Summary:     "Summary text",
				Description: "Description text",
			},
			test: func(t *testing.T, original, cloned *ExtendedComment) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Summary, cloned.Summary)
				assert.Equal(t, original.Description, cloned.Description)
			},
		},
		{
			name: "ExtendedComment with empty fields",
			orig: &ExtendedComment{},
			test: func(t *testing.T, original, cloned *ExtendedComment) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Summary, cloned.Summary)
				assert.Equal(t, original.Description, cloned.Description)
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

func TestComment_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		orig *Comment
		test func(t *testing.T, original, cloned *Comment)
	}{
		{
			name: "nil Comment",
			orig: nil,
			test: func(t *testing.T, original, cloned *Comment) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "Comment with all fields",
			orig: &Comment{
				Summary:                "API endpoint",
				Description:            "This endpoint does something",
				Deprecated:             true,
				DeprecationMessage:     "Use v2 instead",
				DeprecationReplacement: "/v2/endpoint",
				ExternalDocs: &ExternalDocs{
					Description: "Documentation",
					URL:         "https://docs.example.com",
				},
				ExtendedComments: map[string]*ExtendedComment{
					"lang1": {
						Summary:     "Summary 1",
						Description: "Description 1",
					},
					"lang2": {
						Summary:     "Summary 2",
						Description: "Description 2",
					},
				},
			},
			test: func(t *testing.T, original, cloned *Comment) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Summary, cloned.Summary)
				assert.Equal(t, original.Description, cloned.Description)
				assert.Equal(t, original.Deprecated, cloned.Deprecated)
				assert.Equal(t, original.DeprecationMessage, cloned.DeprecationMessage)
				assert.Equal(t, original.DeprecationReplacement, cloned.DeprecationReplacement)

				if original.ExternalDocs != nil {
					require.NotNil(t, cloned.ExternalDocs)
					assert.NotSame(t, original.ExternalDocs, cloned.ExternalDocs)
					assert.Equal(t, original.ExternalDocs.Description, cloned.ExternalDocs.Description)
					assert.Equal(t, original.ExternalDocs.URL, cloned.ExternalDocs.URL)
				} else {
					assert.Nil(t, cloned.ExternalDocs)
				}

				if original.ExtendedComments != nil {
					require.NotNil(t, cloned.ExtendedComments)
					// Maps are not comparable with NotSame, but we can verify they're equal
					assert.Len(t, cloned.ExtendedComments, len(original.ExtendedComments))

					for key, origEC := range original.ExtendedComments {
						clonedEC, exists := cloned.ExtendedComments[key]
						require.True(t, exists)
						assert.NotSame(t, origEC, clonedEC)
						assert.Equal(t, origEC.Summary, clonedEC.Summary)
						assert.Equal(t, origEC.Description, clonedEC.Description)
					}
				} else {
					assert.Nil(t, cloned.ExtendedComments)
				}
			},
		},
		{
			name: "Comment with nil ExternalDocs and ExtendedComments",
			orig: &Comment{
				Summary:     "Simple comment",
				Description: "Simple description",
			},
			test: func(t *testing.T, original, cloned *Comment) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Equal(t, original.Summary, cloned.Summary)
				assert.Equal(t, original.Description, cloned.Description)
				assert.Nil(t, cloned.ExternalDocs)
				assert.Nil(t, cloned.ExtendedComments)
			},
		},
		{
			name: "Comment with empty ExtendedComments map",
			orig: &Comment{
				Summary:          "Comment",
				ExtendedComments: map[string]*ExtendedComment{},
			},
			test: func(t *testing.T, original, cloned *Comment) {
				t.Helper()
				require.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.NotNil(t, cloned.ExtendedComments)
				// Maps are not comparable with NotSame, but we can verify they're equal
				assert.Empty(t, cloned.ExtendedComments)
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

func TestExternalDocs_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		e1       *ExternalDocs
		e2       *ExternalDocs
		expected *ExternalDocs
	}{
		{
			name:     "merges nil ExternalDocs",
			e1:       nil,
			e2:       &ExternalDocs{Description: "Test", URL: "https://example.com"},
			expected: nil,
		},
		{
			name: "merges into nil fields",
			e1:   &ExternalDocs{},
			e2: &ExternalDocs{
				Description: "API docs",
				URL:         "https://example.com/docs",
			},
			expected: &ExternalDocs{
				Description: "API docs",
				URL:         "https://example.com/docs",
			},
		},
		{
			name: "preserves existing values",
			e1: &ExternalDocs{
				Description: "Original description",
				URL:         "https://original.com",
			},
			e2: &ExternalDocs{
				Description: "New description",
				URL:         "https://new.com",
			},
			expected: &ExternalDocs{
				Description: "Original description",
				URL:         "https://original.com",
			},
		},
		{
			name: "merges partial data",
			e1:   &ExternalDocs{Description: "Has description"},
			e2:   &ExternalDocs{URL: "https://example.com"},
			expected: &ExternalDocs{
				Description: "Has description",
				URL:         "https://example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone e1 to avoid mutation affecting the test table
			var testE1 *ExternalDocs
			if tt.e1 != nil {
				testE1 = tt.e1.Clone()
			}

			testE1.Merge(tt.e2)

			if tt.expected == nil {
				assert.Nil(t, testE1)
			} else {
				require.NotNil(t, testE1)
				assert.Equal(t, tt.expected.Description, testE1.Description)
				assert.Equal(t, tt.expected.URL, testE1.URL)
			}
		})
	}
}

func TestExtendedComment_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		e1       *ExtendedComment
		e2       *ExtendedComment
		expected *ExtendedComment
	}{
		{
			name:     "merges nil ExtendedComment",
			e1:       nil,
			e2:       &ExtendedComment{Summary: "Test"},
			expected: nil,
		},
		{
			name: "merges into nil fields",
			e1:   &ExtendedComment{},
			e2: &ExtendedComment{
				Summary:     "Test summary",
				Description: "Test description",
			},
			expected: &ExtendedComment{
				Summary:     "Test summary",
				Description: "Test description",
			},
		},
		{
			name: "preserves existing values",
			e1: &ExtendedComment{
				Summary:     "Original summary",
				Description: "Original description",
			},
			e2: &ExtendedComment{
				Summary:     "New summary",
				Description: "New description",
			},
			expected: &ExtendedComment{
				Summary:     "Original summary",
				Description: "Original description",
			},
		},
		{
			name: "merges partial data",
			e1:   &ExtendedComment{Summary: "Has summary"},
			e2:   &ExtendedComment{Description: "Has description"},
			expected: &ExtendedComment{
				Summary:     "Has summary",
				Description: "Has description",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone e1 to avoid mutation affecting the test table
			var testE1 *ExtendedComment
			if tt.e1 != nil {
				testE1 = tt.e1.Clone()
			}

			testE1.Merge(tt.e2)

			if tt.expected == nil {
				assert.Nil(t, testE1)
			} else {
				require.NotNil(t, testE1)
				assert.Equal(t, tt.expected.Summary, testE1.Summary)
				assert.Equal(t, tt.expected.Description, testE1.Description)
			}
		})
	}
}

func TestComment_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		c1       *Comment
		c2       *Comment
		validate func(t *testing.T, result *Comment)
	}{
		{
			name: "merges nil Comment",
			c1:   nil,
			c2:   &Comment{Summary: "Test"},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				assert.Nil(t, result)
			},
		},
		{
			name: "merges into nil fields",
			c1:   &Comment{},
			c2: &Comment{
				Summary:                "API endpoint",
				Description:            "Endpoint description",
				Deprecated:             true,
				DeprecationMessage:     "Use v2",
				DeprecationReplacement: "/v2/endpoint",
			},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				assert.Equal(t, "API endpoint", result.Summary)
				assert.Equal(t, "Endpoint description", result.Description)
				assert.True(t, result.Deprecated)
				assert.Equal(t, "Use v2", result.DeprecationMessage)
				assert.Equal(t, "/v2/endpoint", result.DeprecationReplacement)
			},
		},
		{
			name: "preserves existing values",
			c1: &Comment{
				Summary:                "Original summary",
				Description:            "Original description",
				Deprecated:             true,
				DeprecationMessage:     "Original message",
				DeprecationReplacement: "Original replacement",
			},
			c2: &Comment{
				Summary:                "New summary",
				Description:            "New description",
				Deprecated:             false,
				DeprecationMessage:     "New message",
				DeprecationReplacement: "New replacement",
			},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				assert.Equal(t, "Original summary", result.Summary)
				assert.Equal(t, "Original description", result.Description)
				assert.True(t, result.Deprecated)
				assert.Equal(t, "Original message", result.DeprecationMessage)
				assert.Equal(t, "Original replacement", result.DeprecationReplacement)
			},
		},
		{
			name: "merges ExternalDocs when nil",
			c1:   &Comment{Summary: "Test"},
			c2: &Comment{
				ExternalDocs: &ExternalDocs{
					Description: "External docs",
					URL:         "https://example.com",
				},
			},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				require.NotNil(t, result.ExternalDocs)
				assert.Equal(t, "External docs", result.ExternalDocs.Description)
				assert.Equal(t, "https://example.com", result.ExternalDocs.URL)
			},
		},
		{
			name: "merges ExternalDocs when both present",
			c1: &Comment{
				ExternalDocs: &ExternalDocs{Description: "Original docs"},
			},
			c2: &Comment{
				ExternalDocs: &ExternalDocs{
					Description: "New docs",
					URL:         "https://example.com",
				},
			},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				require.NotNil(t, result.ExternalDocs)
				assert.Equal(t, "Original docs", result.ExternalDocs.Description)
				assert.Equal(t, "https://example.com", result.ExternalDocs.URL)
			},
		},
		{
			name: "merges ExtendedComments when nil",
			c1:   &Comment{Summary: "Test"},
			c2: &Comment{
				ExtendedComments: map[string]*ExtendedComment{
					"lang1": {Summary: "Summary 1", Description: "Description 1"},
				},
			},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				require.NotNil(t, result.ExtendedComments)
				require.Len(t, result.ExtendedComments, 1)
				assert.Equal(t, "Summary 1", result.ExtendedComments["lang1"].Summary)
				assert.Equal(t, "Description 1", result.ExtendedComments["lang1"].Description)
			},
		},
		{
			name: "merges ExtendedComments when both present",
			c1: &Comment{
				ExtendedComments: map[string]*ExtendedComment{
					"lang1": {Summary: "Original summary"},
				},
			},
			c2: &Comment{
				ExtendedComments: map[string]*ExtendedComment{
					"lang1": {Summary: "New summary", Description: "New description"},
					"lang2": {Summary: "Lang2 summary"},
				},
			},
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				require.NotNil(t, result.ExtendedComments)
				require.Len(t, result.ExtendedComments, 2)
				assert.Equal(t, "Original summary", result.ExtendedComments["lang1"].Summary)
				assert.Equal(t, "New description", result.ExtendedComments["lang1"].Description)
				assert.Equal(t, "Lang2 summary", result.ExtendedComments["lang2"].Summary)
			},
		},
		{
			name: "handles nil other Comment",
			c1:   &Comment{Summary: "Test"},
			c2:   nil,
			validate: func(t *testing.T, result *Comment) {
				t.Helper()
				assert.Equal(t, "Test", result.Summary)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clone c1 to avoid mutation affecting the test table
			var testC1 *Comment
			if tt.c1 != nil {
				testC1 = tt.c1.Clone()
			}

			testC1.Merge(tt.c2)
			tt.validate(t, testC1)
		})
	}
}
