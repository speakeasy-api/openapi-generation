package openapi

import (
	"strings"

	"github.com/speakeasy-api/openapi/openapi"
)

// BuildTagPathMap resolves OpenAPI 3.2 tag parent chains into dot-notation paths.
// For example, tag "books" with parent "products" resolves to "products.books".
// Multi-level: "books" → parent "products" → parent "catalog" → "catalog.products.books".
// Only tags that have a parent are included in the returned map.
func BuildTagPathMap(tags []*openapi.Tag) map[string]string {
	parentOf := map[string]string{}
	for _, t := range tags {
		if t.GetParent() != "" {
			parentOf[t.Name] = t.GetParent()
		}
	}

	result := map[string]string{}
	for name := range parentOf {
		chain := []string{name}
		current := name
		visited := map[string]bool{name: true}
		for {
			parent, ok := parentOf[current]
			if !ok || visited[parent] {
				break
			}
			visited[parent] = true
			chain = append([]string{parent}, chain...)
			current = parent
		}
		result[name] = strings.Join(chain, ".")
	}
	return result
}

// BuildNonNavTagSet returns a set of tag names that have a non-nav kind (e.g., "badge", "audience").
// Tags with these kinds should not create SubSDKs — they represent cross-cutting concerns rather than
// navigation groups. Tags without a kind are treated as nav by default.
func BuildNonNavTagSet(tags []*openapi.Tag) map[string]bool {
	result := map[string]bool{}
	for _, t := range tags {
		kind := t.GetKind()
		if kind != "" && kind != "nav" {
			result[t.Name] = true
		}
	}
	return result
}
