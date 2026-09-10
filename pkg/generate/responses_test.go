package generate_test

import (
	"slices"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/stretchr/testify/require"
)

func TestCompareCodePredicates(t *testing.T) {
	testcases := []struct {
		left, right []string
		expected    int
	}{
		{[]string{"200"}, []string{"201"}, -1},
		{[]string{"201"}, []string{"201"}, 0},
		{[]string{"201"}, []string{"200"}, 1},

		{[]string{"201"}, []string{"404"}, -1},
		{[]string{"401"}, []string{"204"}, 1},

		{[]string{"200"}, []string{"default"}, -1},
		{[]string{"default"}, []string{"200"}, 1},
		{[]string{"default"}, []string{"default"}, 0},
		{[]string{"5XX"}, []string{"default"}, -1},

		{[]string{"200"}, []string{"2XX"}, -1},
		{[]string{"2XX"}, []string{"200"}, 1},
		{[]string{"2xx"}, []string{"2XX"}, 0}, // case insensitive
		{[]string{"2XX"}, []string{"2XX"}, 0},
		{[]string{"2XX"}, []string{"default"}, -1},
		{[]string{"default"}, []string{"2XX"}, 1},

		{[]string{"2XX"}, []string{"3XX"}, -1},
		{[]string{"3XX"}, []string{"2XX"}, 1},
		{[]string{"2XX"}, []string{"2XX", "3XX"}, -1},
		{[]string{"2XX", "3XX"}, []string{"2XX"}, 1},

		{[]string{"404"}, []string{"4XX", "5XX"}, -1},
		{[]string{"502"}, []string{"4XX", "5XX"}, -1},
	}

	terms := []string{"less than", "equal to", "greater than"}

	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			result := generate.CompareCodePredicates(tc.left, tc.right)
			term := terms[tc.expected+1]
			require.Equal(t, tc.expected, result, "expected %v to be %s %v", tc.left, term, tc.right)
		})
	}
}

func TestCompareMediaRanges(t *testing.T) {
	testcases := []struct {
		title string
		input []string
		// leaving this nil means we expect no change in the order
		expected []string
	}{
		{
			title: "empty",
			input: []string{},
		},
		{
			title: "single entry",
			input: []string{"application/json"},
		},
		{
			title: "preserves original order",
			input: []string{"text/csv", "application/zlib", "application/json"},
		},
		{
			title:    "moves wildcard to end",
			input:    []string{"text/csv", "*/*", "application/zlib", "application/json"},
			expected: []string{"text/csv", "application/zlib", "application/json", "*/*"},
		},
		{
			title:    "ranges precede wildcards",
			input:    []string{"*/*", "application/*"},
			expected: []string{"application/*", "*/*"},
		},
		{
			title:    "demotes ranges",
			input:    []string{"text/csv", "text/*", "application/*", "application/zlib", "application/json"},
			expected: []string{"text/csv", "application/zlib", "application/json", "text/*", "application/*"},
		},
		{
			title:    "ranges precede wildcards with mixed set",
			input:    []string{"text/csv", "*/*", "text/*", "*/*", "application/*", "application/zlib", "application/json"},
			expected: []string{"text/csv", "application/zlib", "application/json", "text/*", "application/*", "*/*", "*/*"},
		},
	}

	for _, tc := range testcases {
		input := append([]string{}, tc.input...)
		expected := tc.expected
		if expected == nil {
			expected = append([]string{}, tc.input...)
		}

		t.Run(tc.title, func(t *testing.T) {
			slices.SortStableFunc(input, generate.CompareMediaRanges)
			require.Equal(t, expected, input)
		})
	}
}
