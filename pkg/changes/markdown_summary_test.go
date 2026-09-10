package changes

import (
	"fmt"
	"strings"
	"testing"
)

// TestToMarkdownCompactSummarizesWhenTooLarge verifies the compact changelog
// degrades to a count summary once it would exceed the inline size limit, so it
// stays under GitHub's 125000-char release body cap. Every summarized change type
// (added / removed / deprecated / changed) is represented with a distinct count.
func TestToMarkdownCompactSummarizesWhenTooLarge(t *testing.T) {
	// Distinct counts per type so the assertion pins each line unambiguously.
	// ArgumentsChanged summarizes as "changed" (the default case).
	counts := []struct {
		typ   ChangeType
		count int
	}{
		{MethodAdded, 800},
		{MethodDeleted, 500},
		{MethodDeprecated, 200},
		{ArgumentsChanged, 300},
	}

	var changes []MethodDiff
	total := 0
	for _, c := range counts {
		for i := range c.count {
			changes = append(changes, MethodDiff{
				Type:        c.typ,
				MethodParts: []string{"sdk", "group", fmt.Sprintf("%s_operation_%d", c.typ, i)},
				MethodKey:   fmt.Sprintf("sdk.group.%s_operation_%d", c.typ, i),
			})
		}
		total += c.count
	}

	diff := SDKDiff{Changes: changes}

	compact := ToMarkdown(diff, DetailLevelCompact)

	// Output is fully deterministic: both subsystems are nil (no lang header), so
	// only the summary lines appear, in the order added / removed / deprecated / changed.
	want := fmt.Sprintf("%d method changes summary (please refer to the full report for details):\n", total) +
		"* 800 added\n" +
		"* 500 removed\n" +
		"* 200 deprecated\n" +
		"* 300 changed\n"
	if compact != want {
		t.Fatalf("summarized output mismatch:\ngot:\n%s\nwant:\n%s", compact, want)
	}
	if len(compact) > maxCompactMarkdownBytes {
		t.Fatalf("summarized output exceeds limit: got %d bytes, want <= %d", len(compact), maxCompactMarkdownBytes)
	}
}

// TestToMarkdownCompactKeepsDetailWhenSmall verifies small changelogs keep full
// per-method detail (no summarization).
func TestToMarkdownCompactKeepsDetailWhenSmall(t *testing.T) {
	diff := SDKDiff{Changes: []MethodDiff{
		{Type: MethodAdded, MethodParts: []string{"sdk", "foo"}, MethodKey: "sdk.foo"},
	}}

	compact := ToMarkdown(diff, DetailLevelCompact)
	if strings.Contains(compact, "method changes (detail omitted") {
		t.Fatalf("small changelog should not be summarized, got:\n%s", compact)
	}
	if !strings.Contains(compact, "`sdk.foo()`") {
		t.Fatalf("expected per-method detail, got:\n%s", compact)
	}
}
