package imports

import "testing"

func TestPythonRelativeImport(t *testing.T) {
	cases := []struct {
		name string
		from string
		to   string
		want string
	}{
		// Worked examples from the doc comment.
		{"sibling submodule", "openapi.utils", "openapi.utils.logger", ".logger"},
		{"deep up to sibling subtree", "openapi.types.interactions", "openapi.utils", "...utils"},
		{"root down one", "openapi", "openapi.utils", ".utils"},
		{"sibling at same depth", "openapi.types.foo", "openapi.types", ".."},
		{"identical", "openapi.utils", "openapi.utils", "."},
		{"siblings under common parent", "openapi.types.foo", "openapi.types.bar", "..bar"},

		// Edge: empty inputs are zero-segment paths.
		{"both empty", "", "", "."},
		{"empty from, non-empty to", "", "openapi.utils", ".openapi.utils"},
		{"non-empty from, empty to", "openapi.utils", "", "..."},

		// Edge: from and to share no prefix at all.
		{"no common prefix", "a.b", "c.d", "...c.d"},

		// Edge: single-segment paths.
		{"single-segment identical", "a", "a", "."},
		{"single-segment to child", "a", "a.b", ".b"},
		{"single-segment to unrelated", "a", "b", "..b"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PythonRelativeImport(tc.from, tc.to)
			if got != tc.want {
				t.Errorf("PythonRelativeImport(%q, %q) = %q, want %q",
					tc.from, tc.to, got, tc.want)
			}
		})
	}
}
