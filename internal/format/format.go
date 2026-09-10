//go:build !js || !wasm

package format

import (
	"bytes"
	"context"
	gofmt "go/format"
	"regexp"
	"strings"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

var (
	initOnce      sync.Once
	tfFormatter   = tfFormat{}
	tsFormatter   dprintFormatter
	pyFormatter   dprintFormatter
	javaFormatter dprintFormatter
	phpFormatter  dprintFormatter
	mtx           sync.Mutex

	formattingEnabled   = map[string]bool{}
	formattingEnabledMu sync.RWMutex
)

// SetFormattingEnabled enables or disables formatting for a specific target.
// Used to gate formatting behind a gen.yaml config flag.
func SetFormattingEnabled(target string, enabled bool) {
	formattingEnabledMu.Lock()
	defer formattingEnabledMu.Unlock()
	formattingEnabled[target] = enabled
}

func isFormattingEnabled(target string) bool {
	formattingEnabledMu.RLock()
	defer formattingEnabledMu.RUnlock()
	enabled, ok := formattingEnabled[target]
	return ok && enabled
}

func Format(ctx context.Context, target types.Target, fileName string, data []byte) ([]byte, error) {
	if skip, cleaned := ShouldSkipFormat(data); skip {
		return cleaned, nil
	}
	ensureDPrint()
	switch target.Target {
	case "php":
		if strings.HasSuffix(fileName, ".php") {
			return formatPHP(fileName, data)
		}
	case "java":
		if strings.HasSuffix(fileName, ".java") && isFormattingEnabled("java") {
			if formatted, err := formatJava(fileName, data); err == nil {
				return formatted, nil
			}
			// Formatting failed; return unformatted data so generation is not blocked.
			return data, nil
		}
	case "mcp-typescript", "typescript":
		if strings.HasSuffix(fileName, ".ts") || strings.HasSuffix(fileName, ".js") {
			return formatTypeScript(fileName, data)
		}
	case "python":
		if strings.HasSuffix(fileName, ".py") && fileName != "example.py" {
			return formatPython(fileName, data)
		}
	case "ruby":
		if (strings.HasSuffix(fileName, ".rb") || strings.HasSuffix(fileName, ".rbi")) && isFormattingEnabled("ruby") {
			return formatRuby(ctx, fileName, data)
		}
	case "csharp", "unity":
		if strings.HasSuffix(fileName, ".cs") && isFormattingEnabled(target.Target) {
			return formatCSharp(ctx, fileName, data)
		}
	case "terraform":
		if tfFormatter.MatchesSuffix(fileName) {
			return tfFormatter.formatSourceCode(data, fileName), nil
		}
		fallthrough
	case "go", "cli", "mockserver":
		if strings.HasSuffix(fileName, ".go") {
			return gofmt.Source(data)
		}
	}

	return data, nil
}

func formatPHP(fileName string, data []byte) ([]byte, error) {
	mtx.Lock()
	defer mtx.Unlock()

	formatted, err := phpFormatter.Format(context.Background(), fileName, data)
	if err != nil {
		// Formatting failed; return unformatted data so generation is not blocked.
		return data, nil
	}

	// Preserve prior Prettier formatting behavior: Prettier did not enforce a
	// trailing newline at EOF, but mago always adds one. To avoid unnecessary
	// code churn across all PHP SDKs on the formatter migration, match the
	// input's EOF style. This post-processing can be removed anytime the team
	// is comfortable with a one-time trailing-newline diff.
	if len(data) > 0 && data[len(data)-1] != '\n' {
		formatted = bytes.TrimRight(formatted, "\n")
	}

	return formatted, nil
}

func formatPython(fileName string, data []byte) ([]byte, error) {
	mtx.Lock()
	defer mtx.Unlock()

	return pyFormatter.Format(context.Background(), fileName, data)
}

func formatTypeScript(fileName string, data []byte) ([]byte, error) {
	mtx.Lock()
	defer mtx.Unlock()

	return tsFormatter.Format(context.Background(), fileName, data)
}

func formatJava(fileName string, data []byte) ([]byte, error) {
	mtx.Lock()
	defer mtx.Unlock()

	// Pre-process: strip trailing empty "//" comments used as line-continuation
	// hints. The dprint Java formatter handles line-wrapping on its own, and
	// these bare "//" comments cause the formatter to produce broken output
	// (commas placed on separate lines after the comment).
	cleaned := stripTrailingEmptyComments(string(data))

	// Pre-process: convert trailing line comments on annotations to block
	// comments. The formatter incorrectly relocates line comments from
	// annotation lines (e.g. @Disabled // reason) into method signatures
	// (e.g. public // reason void method()), breaking compilation.
	cleaned = annotationLineCommentToBlock(cleaned)

	formatted, err := javaFormatter.Format(context.Background(), fileName, []byte(cleaned))
	if err != nil {
		return nil, err
	}

	result := string(formatted)

	// Post-process: fix a dprint bug where inline comments inside argument
	// lists cause the comma to be placed on a new line after the comment.
	// Pattern: "// comment\n<whitespace>," should become ",\n// comment"
	result = fixCommaAfterComment(result)

	return []byte(result), nil
}

// annotationLineCommentRe matches annotation lines with a trailing // comment.
// e.g. "    @Disabled // reason here"
var annotationLineCommentRe = regexp.MustCompile(`(?m)^(\s*@\w+(?:\([^)]*\))?)\s*//\s*(.+?)\s*$`)

// annotationLineCommentToBlock converts trailing line comments on annotation
// lines to block comments to prevent the formatter from relocating them.
func annotationLineCommentToBlock(input string) string {
	return annotationLineCommentRe.ReplaceAllString(input, "$1 /* $2 */")
}

// trailingEmptyCommentRe matches a trailing "//" with optional whitespace but
// no actual comment text. These are used in Java templates as line-continuation
// hints for method chains.
var trailingEmptyCommentRe = regexp.MustCompile(`\s*//\s*$`)

// stripTrailingEmptyComments removes bare trailing "//" comments that serve
// only as line-continuation hints. Real comments (with text after //) are
// preserved.
func stripTrailingEmptyComments(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		lines[i] = trailingEmptyCommentRe.ReplaceAllString(line, "")
	}
	return strings.Join(lines, "\n")
}

// commaAfterCommentRe matches a line that is only a comment followed by a line
// that contains only a comma (with optional surrounding whitespace). This is a
// dprint bug where it introduces a spurious comma on a standalone line after an
// inline comment inside an argument list.
var commaAfterCommentRe = regexp.MustCompile(`(?m)(^[ \t]*//.+\n)[ \t]*,[ \t]*\n`)

// fixCommaAfterComment removes spurious standalone comma lines that dprint
// inserts after comment lines inside argument lists.
func fixCommaAfterComment(input string) string {
	return commaAfterCommentRe.ReplaceAllString(input, "$1")
}

func ensureDPrint() {
	initOnce.Do(func() {
		var err error
		tsFormatter, err = newDprint("typescript")
		if err != nil {
			panic(err)
		}

		pyFormatter, err = newDprint("python")
		if err != nil {
			panic(err)
		}

		javaFormatter, err = newDprint("java")
		if err != nil {
			panic(err)
		}

		phpFormatter, err = newDprint("php")
		if err != nil {
			panic(err)
		}
	})
}

func init() {
	go ensureDPrint()
}
