package generate

import (
	"bufio"
	"bytes"
	"fmt"
	"path"
	"regexp"
	"strings"
)

// htmlCommentLineRegex matches whole-line HTML comments such as the
// `<!-- UsageSnippet ... -->` markers the generator emits around usage
// snippets, plus any other single-line `<!-- ... -->` directives Mintlify
// would render as raw text.
var htmlCommentLineRegex = regexp.MustCompile(`(?m)^[ \t]*<!--.*?-->[ \t]*\n?`)

// overviewHeadingRegex matches the `## Overview` heading we anchor the
// description on. Case-insensitive to mirror the upstream Python
// post-processor.
var overviewHeadingRegex = regexp.MustCompile(`(?i)^[ \t]*##[ \t]+Overview[ \t]*$`)

// existingFrontmatterRegex matches a YAML frontmatter block at the very start
// of a document. Used to detect (and replace) frontmatter on re-runs.
var existingFrontmatterRegex = regexp.MustCompile(`\A---\n[\s\S]*?\n---\n`)

// markdownLinkRegex matches markdown link targets ending in `.md`, optionally
// followed by a `#fragment`. Captures: 1=path, 2=fragment (with leading #) or
// empty.
var markdownLinkRegex = regexp.MustCompile(`\]\(([^)\s]+?\.md)(#[^)\s]*)?\)`)

// schemeRegex matches absolute URLs / external schemes we should NOT rewrite.
var schemeRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)

// applyMintlifyTransform converts a generated `docs/**/*.md` file into a
// Mintlify-flavored MDX file. Files outside `docs/` and non-`.md` files pass
// through unchanged. The returned filename has `.mdx` extension when
// transformed.
func applyMintlifyTransform(filename string, data []byte) (string, []byte) {
	if !isDocumentationFile(filename) {
		return filename, data
	}

	body := string(data)
	body = stripHTMLCommentLines(body)
	body = existingFrontmatterRegex.ReplaceAllString(body, "")

	title, description, body := extractTitleAndDescription(body)
	body = rewriteIntraDocLinks(body)
	body = escapeMDXHazards(body)

	out := buildFrontmatter(title, description) + strings.TrimLeft(body, "\n")

	return mintlifyOutputPath(filename), []byte(out)
}

// mintlifyOutputPath returns the disk path mintlify mode emits for the given
// template-output filename. Files outside `docs/` or non-`.md` files are
// returned unchanged. This is the filename-only counterpart of
// applyMintlifyTransform — kept in sync so write-side and read-side code can
// agree on the post-transform name without re-running content transforms.
func mintlifyOutputPath(filename string) string {
	if !isDocumentationFile(filename) {
		return filename
	}
	return strings.TrimSuffix(filename, ".md") + ".mdx"
}

// isDocumentationFile reports whether filename is a generated per-operation
// documentation file — a `.md` under the `docs/` directory. This is the set
// the mintlify transform rewrites and the set `generation.documentation: none`
// skips writing entirely.
func isDocumentationFile(filename string) bool {
	if !strings.HasSuffix(filename, ".md") {
		return false
	}
	clean := path.Clean(filename)
	return clean == "docs" || strings.HasPrefix(clean, "docs/")
}

func stripHTMLCommentLines(body string) string {
	return htmlCommentLineRegex.ReplaceAllString(body, "")
}

// extractTitleAndDescription pulls the H1 as `title` and the paragraph
// directly under the first `## Overview` heading as `description`. The H1
// line (and any immediately-following blank lines) is removed from the body
// so Mintlify's frontmatter title doesn't render twice. The `## Overview`
// heading and its paragraph are KEPT in the body — duplicating the lead
// paragraph there mirrors the upstream Python post-processor and lets the
// page read naturally when viewed standalone. If `## Overview` is absent or
// has no paragraph under it, the description is empty.
func extractTitleAndDescription(body string) (string, string, string) {
	lines := splitLines(body)

	titleIdx := -1
	for i, line := range lines {
		t := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(t, "# ") {
			titleIdx = i
			break
		}
		if t != "" {
			break
		}
	}

	title := ""
	if titleIdx != -1 {
		title = strings.TrimSpace(strings.TrimPrefix(strings.TrimLeft(lines[titleIdx], " \t"), "#"))
		end := titleIdx + 1
		for end < len(lines) && strings.TrimSpace(lines[end]) == "" {
			end++
		}
		lines = append(lines[:titleIdx], lines[end:]...)
	}

	description := findOverviewDescription(lines)
	return title, description, joinLines(lines)
}

// findOverviewDescription returns the paragraph immediately under the first
// `## Overview` heading. Termination on blank line or any heading mirrors the
// upstream Python script.
func findOverviewDescription(lines []string) string {
	for i, line := range lines {
		if !overviewHeadingRegex.MatchString(line) {
			continue
		}
		j := i + 1
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		var paragraph []string
		for j < len(lines) {
			if strings.TrimSpace(lines[j]) == "" {
				break
			}
			if strings.HasPrefix(strings.TrimLeft(lines[j], " \t"), "#") {
				break
			}
			paragraph = append(paragraph, strings.TrimSpace(lines[j]))
			j++
		}
		if len(paragraph) == 0 {
			return ""
		}
		// Match the upstream Python script: keep inline markdown (links,
		// code, bold, italic) verbatim. The script does
		// `" ".join(p.strip() for p in paragraph).strip()`.
		return strings.Join(paragraph, " ")
	}
	return ""
}

// escapeMDXHazards escapes characters in body prose that would otherwise
// crash the MDX parser:
//   - `<` followed by an uppercase letter or digit (would be parsed as a JSX
//     component opening or invalid tag).
//   - `{` (would open a JSX expression). Body prose like `Types: { kind: ... }`
//     would otherwise fail to parse.
//
// Lowercase tags like `<br/>` are valid HTML in MDX and are left alone.
// Content inside fenced code blocks and inline-code backticks is left
// untouched. Already-escaped sequences (`\<`, `\{`) are not double-escaped.
func escapeMDXHazards(body string) string {
	lines := splitLines(body)
	inFence := false
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		lines[i] = escapeHazardsOutsideCode(line)
	}
	return joinLines(lines)
}

func escapeHazardsOutsideCode(line string) string {
	var out strings.Builder
	out.Grow(len(line))
	i := 0
	for i < len(line) {
		c := line[i]
		// Inline code spans open and close with backtick runs of EQUAL length.
		// A run with no matching closer on the line is a literal backtick, not
		// a delimiter, so it must not suppress escaping of the rest of the line
		// (matches CommonMark and MDX). Toggling on every backtick — as this
		// once did — let a stray backtick shield an unescaped `<Foo, ...>` and
		// crash the MDX parser. A backslash-escaped backtick (`\``) is likewise
		// a literal and never opens a span.
		if c == '`' {
			if isCharEscaped(line, i) {
				out.WriteByte(c)
				i++
				continue
			}
			runLen := 1
			for i+runLen < len(line) && line[i+runLen] == '`' {
				runLen++
			}
			if closeIdx := findClosingBacktickRun(line, i+runLen, runLen); closeIdx != -1 {
				end := closeIdx + runLen
				out.WriteString(line[i:end])
				i = end
				continue
			}
			out.WriteString(line[i : i+runLen])
			i += runLen
			continue
		}
		// A character is escaped only when preceded by an ODD number of
		// backslashes — `\<` is escaped, but `\\<` is a literal backslash
		// followed by an unescaped `<` (which still needs escaping).
		alreadyEscaped := isCharEscaped(line, i)
		if c == '<' && i+1 < len(line) && !alreadyEscaped {
			next := line[i+1]
			isUpper := next >= 'A' && next <= 'Z'
			isDigit := next >= '0' && next <= '9'
			isLower := next >= 'a' && next <= 'z'
			switch {
			case isUpper || isDigit:
				// Uppercase looks like a JSX component, digit isn't a valid
				// tag name start -- safe to always escape.
				out.WriteString(`\<`)
				i++
				continue
			case isLower && !looksLikeHTMLTag(line, i):
				// Lowercase could be a real HTML tag like `<br/>` (leave alone
				// so MDX renders it), or a hazard like `Record<string, X>`,
				// `array<string, X>`, or `<https://example.com>` whose
				// "tag name" contains characters MDX won't accept. Escape
				// only the hazard cases.
				out.WriteString(`\<`)
				i++
				continue
			}
		}
		if c == '{' && !alreadyEscaped {
			out.WriteString(`\{`)
			i++
			continue
		}
		out.WriteByte(c)
		i++
	}
	return out.String()
}

// findClosingBacktickRun returns the start index of the first backtick run of
// exactly n backticks at or after `from`, or -1 if none exists. A run longer
// than n does not close an n-length code span (CommonMark rule), so it is
// skipped and the search continues. Backslash-escaped backticks are literals,
// not delimiters, so they are ignored.
func findClosingBacktickRun(s string, from, n int) int {
	for j := from; j < len(s); {
		if s[j] != '`' || isCharEscaped(s, j) {
			j++
			continue
		}
		runLen := 1
		for j+runLen < len(s) && s[j+runLen] == '`' {
			runLen++
		}
		if runLen == n {
			return j
		}
		j += runLen
	}
	return -1
}

// allowedHTMLInlineTags is the set of HTML tag names treated as legitimate
// inline markup in mintlify body prose. Tags outside this set get their `<`
// escaped so MDX won't try to parse them as an unknown JSX component. The
// list mirrors `cSharpAllowedXMLInlineTags` in spirit -- conservative, only
// tags that may appear legitimately in generated descriptions from OpenAPI
// specs. Adding a tag is safe; removing one means it renders as escaped text.
var allowedHTMLInlineTags = map[string]bool{
	"a": true, "b": true, "br": true, "code": true, "em": true,
	"i": true, "img": true, "li": true, "ol": true, "p": true,
	"pre": true, "span": true, "strong": true, "sub": true, "sup": true,
	"table": true, "tbody": true, "td": true, "th": true, "thead": true,
	"tr": true, "u": true, "ul": true,
}

// looksLikeHTMLTag reports whether `<` at position i begins what reads like a
// valid HTML tag whose name is in [allowedHTMLInlineTags]. It scans forward
// from i+1 and accepts the run only when every character up to the first
// space, `/`, `>`, or end-of-line is a valid tag-name character (letter,
// digit, or hyphen) AND the assembled name is in the whitelist. Anything
// else -- `:`, `,`, `=`, `[`, unknown tag names like `myCustomElement`, etc.
// -- means MDX would fail to parse it as a known tag, so the `<` needs
// escaping.
func looksLikeHTMLTag(s string, i int) bool {
	if i+1 >= len(s) {
		return false
	}
	for j := i + 1; j < len(s); j++ {
		c := s[j]
		switch {
		case c == '>' || c == ' ' || c == '/' || c == '\t':
			if j == i+1 {
				return false
			}
			return allowedHTMLInlineTags[strings.ToLower(s[i+1:j])]
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-':
			// valid tag-name character; keep scanning
		default:
			return false
		}
	}
	// Reached end of line without a closing `>` -- not a tag.
	return false
}

// isCharEscaped reports whether the byte at position i in s is preceded by an
// odd number of backslashes — i.e., escaped under standard backslash-pairing
// rules. `\<` and `\\\<` are escaped; `\\<` (escaped backslash literal + raw
// `<`) is not.
func isCharEscaped(s string, i int) bool {
	count := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		count++
	}
	return count%2 == 1
}

// rewriteIntraDocLinks rewrites markdown link targets like `(../models/foo.md)`
// to `(../models/foo.mdx)`. Absolute URLs (anything with a `scheme:` prefix)
// are left untouched.
func rewriteIntraDocLinks(body string) string {
	return markdownLinkRegex.ReplaceAllStringFunc(body, func(match string) string {
		groups := markdownLinkRegex.FindStringSubmatch(match)
		linkPath := groups[1]
		fragment := groups[2]
		if schemeRegex.MatchString(linkPath) {
			return match
		}
		return fmt.Sprintf("](%sx%s)", linkPath, fragment)
	})
}

// buildFrontmatter renders a YAML frontmatter block. Title and description are
// always quoted to avoid YAML escaping pitfalls (colons, leading symbols).
// Description is omitted when empty.
func buildFrontmatter(title, description string) string {
	var b bytes.Buffer
	b.WriteString("---\n")
	if title != "" {
		fmt.Fprintf(&b, "title: %s\n", yamlQuote(title))
	}
	if description != "" {
		fmt.Fprintf(&b, "description: %s\n", yamlQuote(description))
	}
	b.WriteString("---\n\n")
	return b.String()
}

func yamlQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func splitLines(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if strings.HasSuffix(s, "\n") {
		lines = append(lines, "")
	}
	return lines
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
