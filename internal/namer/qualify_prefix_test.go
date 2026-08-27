package namer

import (
	"testing"
)

// shouldStripParentPrefix mirrors the redundant-prefix decision in
// qualifyInlineName so the word-boundary behaviour can be exercised directly.
func shouldStripParentPrefix(parent string, props ...string) bool {
	_, consumedRefName := stripRedundantParent(append([]string{parent}, props...))
	return !consumedRefName
}

func TestRedundantPrefixIsTokenWise(t *testing.T) {
	cases := []struct {
		name   string
		parent string
		props  []string
		strip  bool
	}{
		{"redundant token prefix keeps embedded parent", "Redundant", []string{"redundantStatus"}, true},
		{"redundant nested prefix", "Redundant", []string{"redundantSettings", "inner"}, true},
		{"mid-token letters are not a prefix", "Card", []string{"cardinality"}, false},
		{"pluralized parent is not a prefix", "Order", []string{"orders"}, false},
		{"suffixed parent is not a prefix", "Account", []string{"accountable"}, false},
		{"equal name is not a strict prefix", "Order", []string{"order"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldStripParentPrefix(tc.parent, tc.props...); got != tc.strip {
				t.Fatalf("shouldStripParentPrefix(%q, %v) = %v, want %v", tc.parent, tc.props, got, tc.strip)
			}
		})
	}
}
