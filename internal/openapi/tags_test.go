package openapi

import (
	"testing"

	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/stretchr/testify/assert"
)

func TestBuildTagPathMap(t *testing.T) {
	tests := []struct {
		name     string
		tags     []*openapi.Tag
		expected map[string]string
	}{
		{
			name:     "nil tags",
			tags:     nil,
			expected: map[string]string{},
		},
		{
			name: "no parent relationships",
			tags: []*openapi.Tag{
				{Name: "users"},
				{Name: "orders"},
			},
			expected: map[string]string{},
		},
		{
			name: "single parent",
			tags: []*openapi.Tag{
				{Name: "catalog"},
				{Name: "books", Parent: pointer.From("catalog")},
			},
			expected: map[string]string{
				"books": "catalog.books",
			},
		},
		{
			name: "multi-level parent chain",
			tags: []*openapi.Tag{
				{Name: "store"},
				{Name: "catalog", Parent: pointer.From("store")},
				{Name: "books", Parent: pointer.From("catalog")},
			},
			expected: map[string]string{
				"catalog": "store.catalog",
				"books":   "store.catalog.books",
			},
		},
		{
			name: "multiple siblings with same parent",
			tags: []*openapi.Tag{
				{Name: "products"},
				{Name: "books", Parent: pointer.From("products")},
				{Name: "electronics", Parent: pointer.From("products")},
			},
			expected: map[string]string{
				"books":       "products.books",
				"electronics": "products.electronics",
			},
		},
		{
			name: "circular reference is handled gracefully",
			tags: []*openapi.Tag{
				{Name: "a", Parent: pointer.From("b")},
				{Name: "b", Parent: pointer.From("a")},
			},
			expected: map[string]string{
				"a": "b.a",
				"b": "a.b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildTagPathMap(tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildNonNavTagSet(t *testing.T) {
	tests := []struct {
		name     string
		tags     []*openapi.Tag
		expected map[string]bool
	}{
		{
			name:     "nil tags",
			tags:     nil,
			expected: map[string]bool{},
		},
		{
			name: "no kind set — all tags are nav by default",
			tags: []*openapi.Tag{
				{Name: "users"},
				{Name: "orders"},
			},
			expected: map[string]bool{},
		},
		{
			name: "nav kind is not excluded",
			tags: []*openapi.Tag{
				{Name: "users", Kind: pointer.From("nav")},
			},
			expected: map[string]bool{},
		},
		{
			name: "badge kind is excluded",
			tags: []*openapi.Tag{
				{Name: "beta", Kind: pointer.From("badge")},
				{Name: "users"},
			},
			expected: map[string]bool{
				"beta": true,
			},
		},
		{
			name: "audience kind is excluded",
			tags: []*openapi.Tag{
				{Name: "internal", Kind: pointer.From("audience")},
				{Name: "users", Kind: pointer.From("nav")},
			},
			expected: map[string]bool{
				"internal": true,
			},
		},
		{
			name: "mixed kinds",
			tags: []*openapi.Tag{
				{Name: "catalog", Kind: pointer.From("nav")},
				{Name: "beta", Kind: pointer.From("badge")},
				{Name: "internal", Kind: pointer.From("audience")},
				{Name: "misc"},
			},
			expected: map[string]bool{
				"beta":     true,
				"internal": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNonNavTagSet(tt.tags)
			assert.Equal(t, tt.expected, result)
		})
	}
}
