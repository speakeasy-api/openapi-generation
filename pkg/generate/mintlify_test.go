package generate

import (
	"strings"
	"testing"
)

func TestApplyMintlifyTransform_FilenameGating(t *testing.T) {
	cases := []struct {
		name      string
		filename  string
		want      string
		untouched bool
	}{
		{"docs sdks readme", "docs/sdks/foo/README.md", "docs/sdks/foo/README.mdx", false},
		{"docs models component", "docs/models/components/foo.md", "docs/models/components/foo.mdx", false},
		{"docs types one off", "docs/types/date.md", "docs/types/date.mdx", false},
		{"non-docs markdown", "README.md", "README.md", true},
		{"go source", "pkg/foo.go", "pkg/foo.go", true},
		{"docs but not md", "docs/README.html", "docs/README.html", true},
		{"docs prefix collision", "documentation/foo.md", "documentation/foo.md", true},
	}

	body := []byte("# Hello\n\nDescription.\n")
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotName, gotData := applyMintlifyTransform(c.filename, body)
			if gotName != c.want {
				t.Errorf("filename: got %q want %q", gotName, c.want)
			}
			if c.untouched && string(gotData) != string(body) {
				t.Errorf("expected unchanged data, got: %q", gotData)
			}
		})
	}
}

// TestMintlifyOutputPath_MatchesApplyMintlifyTransform asserts that the
// filename-only helper produces the same output filename as the full
// content-transforming pipeline, for the same set of inputs. Write-side
// (`onWriteFile`) and read-side (`onReadFile`) code rely on this equivalence
// to agree on the post-transform name without running content transforms.
func TestMintlifyOutputPath_MatchesApplyMintlifyTransform(t *testing.T) {
	cases := []string{
		"docs/sdks/foo/README.md",
		"docs/models/components/foo.md",
		"docs/types/date.md",
		"README.md",
		"pkg/foo.go",
		"docs/README.html",
		"documentation/foo.md",
	}
	for _, filename := range cases {
		t.Run(filename, func(t *testing.T) {
			wantName, _ := applyMintlifyTransform(filename, []byte("# Hello\n"))
			gotName := mintlifyOutputPath(filename)
			if gotName != wantName {
				t.Errorf("mintlifyOutputPath(%q) = %q; applyMintlifyTransform returned %q", filename, gotName, wantName)
			}
		})
	}
}

func TestApplyMintlifyTransform_ExtractsTitleAndDescription(t *testing.T) {
	in := []byte("# MyClass\n\n## Overview\n\nDoes a thing for the API.\n\n## Methods\n\n* foo\n")
	_, data := applyMintlifyTransform("docs/sdks/myclass/README.md", in)
	got := string(data)

	wantSubs := []string{
		"---\n",
		`title: "MyClass"`,
		`description: "Does a thing for the API."`,
		"\n## Overview\n",
		"\n## Methods\n",
	}
	for _, sub := range wantSubs {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in:\n%s", sub, got)
		}
	}
	if strings.Contains(got, "# MyClass") {
		t.Errorf("H1 not stripped:\n%s", got)
	}
}

func TestApplyMintlifyTransform_NoOverviewMeansNoDescription(t *testing.T) {
	// First-paragraph-after-H1 is no longer used as a description source — only
	// the paragraph under `## Overview` counts. Mirrors the upstream Python
	// post-processor.
	in := []byte("# Foo\n\nA plain paragraph that is NOT under Overview.\n\n## Methods\n")
	_, data := applyMintlifyTransform("docs/sdks/foo/README.md", in)
	got := string(data)
	if !strings.Contains(got, `title: "Foo"`) {
		t.Errorf("title missing: %s", got)
	}
	if strings.Contains(got, "description:") {
		t.Errorf("description should be omitted when no `## Overview` heading is present:\n%s", got)
	}
	if !strings.Contains(got, "A plain paragraph that is NOT under Overview.") {
		t.Errorf("body paragraph should remain since it isn't claimed as a description:\n%s", got)
	}
}

func TestApplyMintlifyTransform_NoH1(t *testing.T) {
	in := []byte("Just a paragraph with no heading.\n")
	_, data := applyMintlifyTransform("docs/types/date.md", in)
	got := string(data)
	if !strings.HasPrefix(got, "---\n---\n") {
		t.Errorf("expected empty frontmatter block, got:\n%s", got)
	}
	if !strings.Contains(got, "Just a paragraph") {
		t.Errorf("body lost:\n%s", got)
	}
}

func TestApplyMintlifyTransform_TitleOnlyNoDescription(t *testing.T) {
	in := []byte("# OnlyTitle\n\n| Col | Col |\n| --- | --- |\n")
	_, data := applyMintlifyTransform("docs/models/components/onlytitle.md", in)
	got := string(data)
	if !strings.Contains(got, `title: "OnlyTitle"`) {
		t.Errorf("title missing: %s", got)
	}
	if strings.Contains(got, "description:") {
		t.Errorf("description should be omitted when only a table follows H1: %s", got)
	}
}

func TestApplyMintlifyTransform_DescriptionUnderOverview(t *testing.T) {
	// Sub-SDK READMEs follow the `# Title \n ## Overview \n description` shape.
	// The Overview heading and its paragraph are retained in the body so the
	// page reads naturally when viewed standalone.
	in := []byte("# Cancellation\n\n## Overview\n\nEndpoints for testing request cancellation.\n\n### Available Operations\n")
	_, data := applyMintlifyTransform("docs/sdks/cancellation/README.md", in)
	got := string(data)
	if !strings.Contains(got, `title: "Cancellation"`) {
		t.Errorf("title missing: %s", got)
	}
	if !strings.Contains(got, `description: "Endpoints for testing request cancellation."`) {
		t.Errorf("description should be pulled from under '## Overview':\n%s", got)
	}
	if !strings.Contains(got, "## Overview") || !strings.Contains(got, "Endpoints for testing request cancellation.") {
		t.Errorf("`## Overview` heading and its paragraph should remain in body:\n%s", got)
	}
}

func TestApplyMintlifyTransform_OverviewParagraphRetainedInBody(t *testing.T) {
	// When the description IS sourced from `## Overview`, the paragraph stays
	// in the body too (mirrors the upstream Python script). Mintlify renders
	// the frontmatter title separately, so the duplication is intentional.
	in := []byte("# RFCDate\n\n## Overview\n\nA wrapper around Date.\n\n## Usage\n\nMore text.\n")
	_, data := applyMintlifyTransform("docs/types/rfcdate.md", in)
	got := string(data)
	if !strings.Contains(got, `description: "A wrapper around Date."`) {
		t.Errorf("description missing:\n%s", got)
	}
	if strings.Count(got, "A wrapper around Date.") != 2 {
		t.Errorf("description paragraph should appear twice (frontmatter + body):\n%s", got)
	}
	if !strings.Contains(got, "## Usage") {
		t.Errorf("subsequent body content lost:\n%s", got)
	}
}

func TestApplyMintlifyTransform_OverviewHeadingPreservedAlongsideDescription(t *testing.T) {
	// When `## Overview` has a paragraph under it, the heading + paragraph
	// remain in the body even though the description was extracted into
	// frontmatter — they are not orphaned.
	in := []byte("# Cancellation\n\n## Overview\n\nEndpoints for testing.\n\n### Available Operations\n\n* foo\n")
	_, data := applyMintlifyTransform("docs/sdks/cancellation/README.md", in)
	got := string(data)
	if !strings.Contains(got, `description: "Endpoints for testing."`) {
		t.Errorf("description missing:\n%s", got)
	}
	if !strings.Contains(got, "## Overview") {
		t.Errorf("`## Overview` heading should remain in body:\n%s", got)
	}
	if !strings.Contains(got, "Endpoints for testing.") {
		t.Errorf("Overview paragraph should remain in body:\n%s", got)
	}
	if !strings.Contains(got, "### Available Operations") {
		t.Errorf("subsequent heading should remain:\n%s", got)
	}
}

func TestApplyMintlifyTransform_OverviewMultipleParagraphsOnlyFirstExtracted(t *testing.T) {
	// When `## Overview` has multiple paragraphs, only the FIRST is lifted
	// into the frontmatter description. Both paragraphs stay in the body
	// (mirrors the upstream Python script — second paragraph is never
	// promoted to frontmatter).
	in := []byte("# Cancellation\n\n## Overview\n\nFirst paragraph should end up in the description and stay in document.\n\nSecond paragraph should not be collected but stay in document.\n\n## Methods\n\n* foo\n")
	_, data := applyMintlifyTransform("docs/sdks/cancellation/README.md", in)
	got := string(data)

	if !strings.Contains(got, `description: "First paragraph should end up in the description and stay in document."`) {
		t.Errorf("first paragraph not lifted into description:\n%s", got)
	}
	if strings.Contains(got, `description: "Second paragraph`) {
		t.Errorf("second paragraph should NOT be promoted into description:\n%s", got)
	}
	// First paragraph appears twice (frontmatter + body).
	if strings.Count(got, "First paragraph should end up in the description and stay in document.") != 2 {
		t.Errorf("first paragraph should appear twice (frontmatter + body):\n%s", got)
	}
	// Second paragraph appears exactly once — body only.
	if strings.Count(got, "Second paragraph should not be collected but stay in document.") != 1 {
		t.Errorf("second paragraph should appear once in body:\n%s", got)
	}
	if !strings.Contains(got, "## Methods") {
		t.Errorf("subsequent heading should remain:\n%s", got)
	}
}

func TestApplyMintlifyTransform_PreservesEmptyOverviewHeading(t *testing.T) {
	// Mintlify's upstream Python script keeps `## Overview` in the body even
	// when there's no paragraph under it — preserves the heading hierarchy
	// (h2 Overview → h3 Available Operations) so files don't end up with an
	// h3 directly preceding a sibling h2. Match that behavior.
	in := []byte("# Collections\n\n## Overview\n\n### Available Operations\n\n* [foo](#foo)\n")
	_, data := applyMintlifyTransform("docs/sdks/collections/README.md", in)
	got := string(data)
	if !strings.Contains(got, "## Overview") {
		t.Errorf("`## Overview` heading should be preserved (matches upstream Python script):\n%s", got)
	}
	if !strings.Contains(got, "### Available Operations") {
		t.Errorf("subsequent heading should remain:\n%s", got)
	}
}

func TestApplyMintlifyTransform_PreservesAllHeadingsEvenWithoutContent(t *testing.T) {
	// We do NOT drop empty headings — the upstream script keeps them, so we
	// keep them too. Stacked empty headings stay as-is.
	in := []byte("# X\n\n## A\n\n## B\n\n### C\n\n* foo\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)
	for _, h := range []string{"## A", "## B", "### C"} {
		if !strings.Contains(got, h) {
			t.Errorf("heading %q should be preserved:\n%s", h, got)
		}
	}
}

func TestApplyMintlifyTransform_HeadingWithContentPreserved(t *testing.T) {
	// When the parent heading has content under it (additional paragraph
	// or list) after stripping the description, the heading is preserved
	// and so is its remaining content.
	in := []byte("# X\n\n## Overview\n\nDesc paragraph.\n\nMore content under Overview.\n\n## Other\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)
	if !strings.Contains(got, `description: "Desc paragraph."`) {
		t.Errorf("description missing:\n%s", got)
	}
	if !strings.Contains(got, "## Overview") {
		t.Errorf("'## Overview' should remain because content still follows:\n%s", got)
	}
	if !strings.Contains(got, "More content under Overview.") {
		t.Errorf("remaining content under heading should be preserved:\n%s", got)
	}
}

func TestApplyMintlifyTransform_NoDescriptionWhenBodyStartsWithCode(t *testing.T) {
	// Model docs typically go `# TypeName \n ## Example Usage \n ```typescript`,
	// which has no usable description. We should NOT grab anything from after
	// the code block — better to omit than to misuse code as a description.
	in := []byte("# NoAuthResponse\n\n## Example Usage\n\n```typescript\nimport x from \"y\";\n```\n\n## Fields\n\n| Col | Col |\n")
	_, data := applyMintlifyTransform("docs/sdk/models/operations/noauthresponse.md", in)
	got := string(data)
	if strings.Contains(got, "description:") {
		t.Errorf("description should be omitted when only headings + code/tables follow H1:\n%s", got)
	}
}

func TestApplyMintlifyTransform_DescriptionPreservesInlineMarkdown(t *testing.T) {
	// Mirror the upstream Python script: inline markdown in the Overview
	// paragraph is kept verbatim in the frontmatter description. Mintlify
	// renders frontmatter description as markdown in some surfaces, and
	// preserving the source matches the script's contract.
	in := []byte("# Foo\n\n## Overview\n\nUses the [link](./other.md) and `code` and **bold**.\n")
	_, data := applyMintlifyTransform("docs/sdks/foo/README.md", in)
	got := string(data)
	if !strings.Contains(got, `description: "Uses the [link](./other.md) and `+"`code`"+` and **bold**."`) {
		t.Errorf("inline markdown should be preserved in description (matches upstream Python script):\n%s", got)
	}
}

func TestApplyMintlifyTransform_DescriptionJoinsMultilineParagraph(t *testing.T) {
	// A single Overview paragraph that spans multiple lines (no blank line
	// breaking it) must be collected as one description and joined with
	// spaces — matches the script's `" ".join(p.strip() for p in paragraph)`.
	in := []byte("# Foo\n\n## Overview\n\nFirst line of the paragraph.\nSecond line continues.\nThird line wraps.\n\n## Methods\n")
	_, data := applyMintlifyTransform("docs/sdks/foo/README.md", in)
	got := string(data)
	if !strings.Contains(got, `description: "First line of the paragraph. Second line continues. Third line wraps."`) {
		t.Errorf("multi-line single paragraph should be joined with spaces:\n%s", got)
	}
	// All three lines should still appear in the body too.
	for _, line := range []string{"First line of the paragraph.", "Second line continues.", "Third line wraps."} {
		if !strings.Contains(got, line) {
			t.Errorf("body should retain line %q:\n%s", line, got)
		}
	}
}

func TestApplyMintlifyTransform_RewritesIntraDocLinks(t *testing.T) {
	in := []byte("# X\n\nDesc.\n\nSee [A](../models/foo.md) and [B](../models/bar.md#frag).\n" +
		"External: [docs](https://example.com/page.md).\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)
	if !strings.Contains(got, "(../models/foo.mdx)") {
		t.Errorf("intra-doc link not rewritten:\n%s", got)
	}
	if !strings.Contains(got, "(../models/bar.mdx#frag)") {
		t.Errorf("link with fragment not rewritten:\n%s", got)
	}
	if !strings.Contains(got, "(https://example.com/page.md)") {
		t.Errorf("external link should not be rewritten:\n%s", got)
	}
}

func TestApplyMintlifyTransform_StripsUsageSnippetMarkers(t *testing.T) {
	in := []byte("# Op\n\n## Overview\n\nDoes op.\n\n<!-- UsageSnippet language=\"typescript\" operationID=\"foo\" -->\n```typescript\nawait sdk.foo();\n```\n")
	_, data := applyMintlifyTransform("docs/sdks/op/README.md", in)
	got := string(data)
	if strings.Contains(got, "UsageSnippet") {
		t.Errorf("usage snippet marker not stripped:\n%s", got)
	}
	if !strings.Contains(got, "await sdk.foo();") {
		t.Errorf("usage snippet body should be retained:\n%s", got)
	}
}

func TestApplyMintlifyTransform_StripsAllWholeLineHTMLComments(t *testing.T) {
	// Any single-line HTML comment occupying its own line should be removed,
	// not just `<!-- UsageSnippet ... -->`.
	in := []byte("# Op\n\n## Overview\n\nDoes op.\n\n<!-- speakeasy-no-tracking -->\n<!--   GeneratorMeta key=value -->\nBody continues.\n")
	_, data := applyMintlifyTransform("docs/sdks/op/README.md", in)
	got := string(data)
	if strings.Contains(got, "<!--") {
		t.Errorf("whole-line HTML comments should be stripped:\n%s", got)
	}
	if !strings.Contains(got, "Body continues.") {
		t.Errorf("non-comment content should be retained:\n%s", got)
	}
}

func TestApplyMintlifyTransform_QuotesYAMLSpecialChars(t *testing.T) {
	in := []byte("# Title: With Colon\n\n## Overview\n\nA \"quoted\" value.\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)
	if !strings.Contains(got, `title: "Title: With Colon"`) {
		t.Errorf("colon in title not preserved by quoting:\n%s", got)
	}
	if !strings.Contains(got, `description: "A \"quoted\" value."`) {
		t.Errorf("quotes in description not escaped:\n%s", got)
	}
}

func TestApplyMintlifyTransform_EscapesMDXBraceHazards(t *testing.T) {
	in := []byte("# X\n\nBody.\n\nTypes: { kind: \"dog\" } | { kind: \"cat\" }\nInline `{a}` is fine.\n\n```typescript\nconst x = { a: 1 };\n```\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)

	if !strings.Contains(got, `Types: \{ kind: "dog" } | \{ kind: "cat" }`) {
		t.Errorf("bare `{` in body prose not escaped:\n%s", got)
	}
	if !strings.Contains(got, "`{a}`") {
		t.Errorf("inline-code `{a}` should not be escaped:\n%s", got)
	}
	if !strings.Contains(got, "const x = { a: 1 };") {
		t.Errorf("fenced code block content should not be escaped:\n%s", got)
	}
}

func TestApplyMintlifyTransform_EscapesMDXAngleHazards(t *testing.T) {
	in := []byte("# X\n\nBody.\n\n* Item — produces List<Map<String, Object>> in Java.\n* Wait for <3 seconds.\n* HTML break <br/> is fine.\n* Inline `<Map>` in code is fine.\n\n```typescript\nconst x: List<Map<String>> = [];\n```\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)

	if !strings.Contains(got, `\<Map`) {
		t.Errorf("uppercase JSX-like tag not escaped:\n%s", got)
	}
	if !strings.Contains(got, `\<3 seconds`) {
		t.Errorf("digit-following-`<` not escaped:\n%s", got)
	}
	if !strings.Contains(got, "<br/> is fine") {
		t.Errorf("lowercase HTML tag should not be escaped:\n%s", got)
	}
	if !strings.Contains(got, "`<Map>`") {
		t.Errorf("inline code containing `<Map>` should not be escaped:\n%s", got)
	}
	// Code fences are preserved as-is.
	if !strings.Contains(got, "List<Map<String>> = []") {
		t.Errorf("fenced code content should not be escaped:\n%s", got)
	}
}

// TestApplyMintlifyTransform_EscapesLowercaseAngleHazards covers cases where
// `<` is followed by a lowercase letter but the run isn't a valid HTML tag --
// the "name" contains characters MDX rejects (`,`, `:`, etc.). These broke
// MDX parsing pre-fix because the predicate only escaped uppercase/digits.
func TestApplyMintlifyTransform_EscapesLowercaseAngleHazards(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"TS generic map cell (comma in apparent tag)",
			"# X\n\nBody.\n\nType: Record<string, *any*>.\n",
			`Record\<string, *any*>`,
		},
		{
			"PHP generic map cell (comma in apparent tag)",
			"# X\n\nBody.\n\nType: array<string, *mixed*>.\n",
			`array\<string, *mixed*>`,
		},
		{
			"bare https autolink (colon in apparent tag)",
			"# X\n\nBody.\n\nSee <https://example.com>.\n",
			`\<https://example.com>`,
		},
		{
			"bare http autolink (colon in apparent tag)",
			"# X\n\nBody.\n\nSee <http://example.com>.\n",
			`\<http://example.com>`,
		},
		{
			"legitimate self-closing HTML tag stays",
			"# X\n\nBody.\n\nLine <br/> break.\n",
			"<br/> break",
		},
		{
			"legitimate HTML tag with attrs stays",
			"# X\n\nBody.\n\nLink <a href=\"x\">y</a>.\n",
			`<a href="x"`,
		},
		{
			"angle hazard inside inline code stays",
			"# X\n\nBody.\n\nUse `Record<string, X>` here.\n",
			"`Record<string, X>`",
		},
		{
			// Tag names outside [allowedHTMLInlineTags] are escaped even when
			// they look syntactically valid. Matches the C# XML doc
			// sanitization approach: whitelist over heuristic.
			"unknown lowercase tag name escapes",
			"# X\n\nBody.\n\nA <myCustomElement> here.\n",
			`\<myCustomElement>`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, data := applyMintlifyTransform("docs/x.md", []byte(tc.in))
			got := string(data)
			if !strings.Contains(got, tc.want) {
				t.Errorf("expected %q in output:\n%s", tc.want, got)
			}
		})
	}
}

// TestApplyMintlifyTransform_StrayBacktickDoesNotShieldAngleHazard is a
// regression test for the MDX crash `Unexpected character , (U+002C) in name`.
// An unmatched backtick on a line must NOT be treated as an inline-code
// delimiter — otherwise the `<Foo, ...>` after it reaches MDX unescaped and the
// JSX parser chokes on the comma. Only backtick runs with a matching closer of
// equal length form a code span.
func TestApplyMintlifyTransform_StrayBacktickDoesNotShieldAngleHazard(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantEscape string // substring that must appear (hazard escaped)
		wantSafe   string // substring that must NOT appear (hazard raw)
	}{
		{
			"single stray backtick before uppercase generic",
			"# X\n\nBody.\n\nUse the ` marker then Either<Foo, Bar> works.\n",
			`Either\<Foo, Bar>`,
			`Either<Foo,`,
		},
		{
			"single stray backtick before lowercase generic",
			"# X\n\nBody.\n\nA lone ` and then Record<string, X> here.\n",
			`Record\<string, X>`,
			`Record<string,`,
		},
		{
			"matched single-backtick span still shields its own hazard",
			"# X\n\nBody.\n\nInline `Record<string, X>` and then Either<Foo, Bar>.\n",
			`Either\<Foo, Bar>`,
			`Either<Foo,`,
		},
		{
			"double-backtick span with matching closer stays code",
			"# X\n\nBody.\n\nInline ``Map<K, V>`` stays code.\n",
			"``Map<K, V>``",
			`Map\<K, V>`,
		},
		{
			"escaped backticks are literals and do not form code span",
			"# X\n\nBody.\n\nA literal \\` marker then Either<Foo, Bar> works before \\` this.\n",
			`Either\<Foo, Bar>`,
			`Either<Foo,`,
		},
		{
			"typo backtick makes later generic-looking code prose",
			"# X\n\nBody.\n\nA typo ` before `T<A,B>`.\n",
			`T\<A,B>`,
			`T<A,B>`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, data := applyMintlifyTransform("docs/models/x.md", []byte(c.in))
			got := string(data)
			body := got
			if idx := strings.Index(got, "\n---\n"); idx != -1 {
				body = got[idx+len("\n---\n"):] // ignore YAML frontmatter (not MDX-parsed)
			}
			if c.wantEscape != "" && !strings.Contains(body, c.wantEscape) {
				t.Errorf("expected %q in body:\n%s", c.wantEscape, body)
			}
			if c.wantSafe != "" && strings.Contains(body, c.wantSafe) {
				t.Errorf("unescaped hazard %q reached MDX body:\n%s", c.wantSafe, body)
			}
		})
	}
}

func TestApplyMintlifyTransform_DoubleEscapeIdempotent(t *testing.T) {
	// Running the transform on already-escaped content should not double-escape.
	in := []byte("# X\n\nBody.\n\nProduces \\<Map\\<String>>.\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)
	if strings.Contains(got, `\\<`) {
		t.Errorf("double-escape produced — should be idempotent:\n%s", got)
	}
}

// TestApplyMintlifyTransform_EscapesAfterEvenBackslashes covers the case where
// an MDX hazard is preceded by an even number of backslashes — i.e., the
// backslashes cancel out and the hazard is unescaped. The transform must
// still escape the hazard. Regression test for the single-char `alreadyEscaped`
// check that mistook `\\<` for an escaped `<`.
func TestApplyMintlifyTransform_EscapesAfterEvenBackslashes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"two backslashes before angle hazard",
			"# X\n\nBody.\n\nLiteral backslash then tag: \\\\<Map>.\n",
			`Literal backslash then tag: \\\<Map>.`,
		},
		{
			"two backslashes before brace hazard",
			"# X\n\nBody.\n\nLiteral backslash then brace: \\\\{Map}.\n",
			`Literal backslash then brace: \\\{Map}.`,
		},
		{
			"three backslashes (odd, already escaped)",
			"# X\n\nBody.\n\nEscaped: \\\\\\<Map>.\n",
			`Escaped: \\\<Map>.`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, data := applyMintlifyTransform("docs/sdks/x/README.md", []byte(c.in))
			got := string(data)
			if !strings.Contains(got, c.want) {
				t.Errorf("expected %q in output, got:\n%s", c.want, got)
			}
		})
	}
}

func TestIsCharEscaped(t *testing.T) {
	cases := []struct {
		name string
		s    string
		i    int
		want bool
	}{
		{"no preceding backslash", "<Map>", 0, false},
		{"one preceding backslash", `\<Map>`, 1, true},
		{"two preceding backslashes", `\\<Map>`, 2, false},
		{"three preceding backslashes", `\\\<Map>`, 3, true},
		{"four preceding backslashes", `\\\\<Map>`, 4, false},
		{"backslash earlier in string but not adjacent", `\ a<Map>`, 3, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isCharEscaped(c.s, c.i); got != c.want {
				t.Errorf("isCharEscaped(%q, %d) = %v, want %v", c.s, c.i, got, c.want)
			}
		})
	}
}

func TestApplyMintlifyTransform_ReplacesExistingFrontmatter(t *testing.T) {
	in := []byte("---\ntitle: Old\n---\n# New\n\nNew description.\n")
	_, data := applyMintlifyTransform("docs/sdks/x/README.md", in)
	got := string(data)
	if strings.Contains(got, "Old") {
		t.Errorf("old frontmatter not replaced:\n%s", got)
	}
	if !strings.Contains(got, `title: "New"`) {
		t.Errorf("new title missing:\n%s", got)
	}
	if strings.Count(got, "---\n") != 2 {
		t.Errorf("expected exactly one frontmatter block:\n%s", got)
	}
}
