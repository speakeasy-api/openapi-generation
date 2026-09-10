package readme

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Removes leading/trailing whitespace and hash characters from a header.
func sanitizeHeader(header string) string {
	re := regexp.MustCompile(`^[\s#]+|[\s#]+$`)
	return re.ReplaceAllString(header, "")
}

// Checks whether a markdown header is contained in a string.
// If a header is found, the function returns its associated level, i.e.
// the number of leading hash characters, otherwise it returns an error.
//
// Arguments:
// - line: The header content.
// - minLevel: The minimum heading depth to parsed.
// - maxLevel: The maximum heading depth to be parsed.
//
// Returns:
// - int: The header level if a header was found.
// - error: An error if no header was found.
func parseHeader(line string, minLevel, maxLevel int) (int, error) {
	level := minLevel
	prefix := strings.Repeat("#", minLevel)

	for level <= maxLevel {
		if strings.HasPrefix(line, prefix) {
			// the line starts with `N=level` hash characters
			if strings.HasPrefix(line, prefix+" ") {
				// hashes are followed by a space: it's a header.
				return level, nil
			} else {
				// must check if it is a header or hight level.
				level++
				prefix += "#"
			}
		} else {
			// the line is not a header.
			break
		}
	}

	return 0, errors.New("No header was found.")
}

// Generates a unique markdown slug from a header string.
//
// Arguments:
// - header: The header content.
// - slugs: A map of existing slugs.
//
// Returns:
// - string: The generated slug.
func getHeaderSlug(header string, slugs *map[string]struct{}) string {
	var b strings.Builder

	// Replace spaces with hyphens and remove non-alphanumeric characters.
	for _, s := range strings.ToLower(header) {
		if unicode.IsSpace(s) {
			b.WriteRune('-')
		} else if unicode.IsLetter(s) || unicode.IsNumber(s) || s == '-' {
			b.WriteRune(s)
		}
	}

	// Remove leading/trailing hyphens and clean consecutive hyphens.
	slug := strings.Trim(b.String(), "-")
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Ensure the slug is unique by appending a counter if necessary.
	if slugs != nil {
		_slug := slug
		cnt := 1
		for {
			if _, exists := (*slugs)[_slug]; !exists {
				break
			}
			_slug = fmt.Sprintf("%s-%d", slug, cnt)
			cnt++
		}
		slug = _slug
	}

	return slug
}

// Generates a Table of Contents entry for a given header.
//
// Arguments:
// - header: The heading content.
// - level: The heading level.
// - slug (optional): The heading slug.
//
// Returns:
// - string: The TOC entry formatted as "* [header](#slug)".
func GenerateTableOfContentsEntry(header string, level, minLevel int, slug *string) string {
	if slug == nil {
		s := getHeaderSlug(header, nil)
		return GenerateTableOfContentsEntry(header, level, minLevel, &s)
	}

	padding := strings.Repeat("  ", level-minLevel)
	return fmt.Sprintf("%s* [%s](#%s)\n", padding, header, *slug)
}

func stripCodeBlocks(contents string) string {
	fence := regexp.MustCompile("^([ \t]*)(`{3,}|~{3,})(.*)$")
	var out []string
	inBlock := false
	openIndent := 0
	openFence := ""
	for _, line := range strings.Split(contents, "\n") {
		m := fence.FindStringSubmatch(line)
		if m != nil && (m[2][0] != '`' || !strings.Contains(m[3], "`")) {
			indent := len(strings.ReplaceAll(m[1], "\t", "    "))
			if inBlock {
				if indent <= openIndent+3 && closesFence(m[2], m[3], openFence) {
					inBlock = false
				}
				continue
			}
			// Per CommonMark, an opening fence may be preceded by at most
			// three spaces of indentation; with four or more the line is an
			// indented code block, not a fence.
			if indent <= 3 {
				inBlock = true
				openIndent = indent
				openFence = m[2]
				continue
			}
		}
		if !inBlock {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func closesFence(fence, trailing, openFence string) bool {
	return strings.TrimSpace(trailing) == "" && fence[0] == openFence[0] && len(fence) >= len(openFence)
}

// Generates a Table of Contents from a markdown string.
//
// Arguments:
// - contents: The markdown contents.
// - minLevel: The minimum heading level to include in the TOC.
// - maxLevel: The maximum heading level to include in the TOC.
// - exclude:  The list of headers to exclude from the TOC.
//
// Returns:
// - string: The generated Table of Contents.
func GenerateTableOfContents(contents string, minLevel, maxLevel int, exclude []string) string {
	contents = stripCodeBlocks(contents)

	excludedHeaders := make(map[string]struct{})
	for _, x := range exclude {
		excludedHeaders[x] = struct{}{}
	}

	slugs := make(map[string]struct{})

	if minLevel < 1 {
		minLevel = 1
	}

	if maxLevel < minLevel {
		maxLevel = minLevel
	}

	var output strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()

		if _, exists := excludedHeaders[line]; exists {
			continue
		}

		level, err := parseHeader(line, minLevel, maxLevel)
		if err != nil {
			// no header was found on this line.
			continue
		}

		header := sanitizeHeader(line)
		slug := getHeaderSlug(header, &slugs)
		slugs[slug] = struct{}{}
		output.WriteString(GenerateTableOfContentsEntry(header, level, minLevel, &slug))
	}

	return output.String()
}
