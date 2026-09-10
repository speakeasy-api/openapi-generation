package readme

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
)

type blockBoundary struct {
	prefix string
	suffix string
	regex  string
}

type block struct {
	id       string
	title    string
	contents string
	start    blockBoundary
	end      blockBoundary
}

// Expected Block formats:
//
//	legacy:
//	    {prefix} Start {title} {suffix}
//	    {contents}
//	    {prefix} End {title} {suffix}
//
//	with ID:
//	    {prefix} Start {title} [{id}] {suffix}
//	    {contents}
//	    {prefix} End {title} [{id}] {suffix}
//
//	e.g.
//	    <!-- Start SDK Installation [installation] -->
//	    go get github.com/something/something
//	    <!-- End SDK Installation [installation] -->
//
// Replaced Block output:
//
//	{prefix} Start {newTitle} [{id}] {suffix}
//	{newContents}
//	{prefix} End {newTitle} [{id}] {suffix}
//
// e.g.
//	<!-- Start New Title [installation] -->
//	go get github.com/something/else
//	<!-- End New Title [installation] -->

// Replaces the content inside a block
//
// Arguments:
// - prefix: The block boundary prefix, e.g. `<!--`.
// - title: The block title (used in "legacy mode" only).
// - id: The block ID.
// - suffix: The block boundary suffix, e.g. `-->`.
// - contents: The README markdown contents.
// - newTitle: The title that should be placed in the block boundary.
// - newContents: The new content that should be placed inside the block.
//
// Returns:
// - string: The updated readme with the block contents replaced.
func ReplaceBlock(prefix, title, id, suffix, contents, newTitle, newContents string) (string, bool) {

	block := block{
		id:    id,
		title: newTitle,
		start: blockBoundary{
			prefix: prefix + " Start ",
			suffix: " " + suffix,
		},
		end: blockBoundary{
			prefix: prefix + " End ",
			suffix: " " + suffix,
		},
		contents: contents,
	}

	if contents == "" {
		return block.format(newContents), true
	}

	if id != "" {
		// try using ID first (title is bound to change)
		block.updateRegexWithID(id)
		if block.replaceContent(newContents) {
			return block.contents, true
		}
	}

	// falling back on using title (legacy mode)
	block.updateRegexWithTitle(title)
	if block.replaceContent(newContents) {
		return block.contents, true
	}

	return contents, false
}

func (b *block) format(replacement string) string {

	replacement = fmt.Sprintf("\n%s\n", replacement)
	if b.id != "" {
		return b.formatWithID(replacement)
	}
	return b.formatWithTitle(replacement)
}

func (b *block) formatWithID(replacement string) string {
	boundaryFormat := "%s%s [%s]%s"
	startStr := fmt.Sprintf(boundaryFormat, b.start.prefix, b.title, b.id, b.start.suffix)
	endStr := fmt.Sprintf(boundaryFormat, b.end.prefix, b.title, b.id, b.start.suffix)
	return strings.Join([]string{startStr, replacement, endStr}, "")
}

func (b *block) formatWithTitle(replacement string) string {
	boundaryFormat := "%s%s%s"
	startStr := fmt.Sprintf(boundaryFormat, b.start.prefix, b.title, b.start.suffix)
	endStr := fmt.Sprintf(boundaryFormat, b.end.prefix, b.title, b.start.suffix)
	return strings.Join([]string{startStr, replacement, endStr}, "")

}

func (b *block) updateRegexWithID(id string) {
	regexFormat := `%s[\w\s-]+\[%s\]%s`
	b.start.regex = fmt.Sprintf(
		regexFormat,
		regexp.QuoteMeta(b.start.prefix),
		regexp.QuoteMeta(id),
		regexp.QuoteMeta(b.start.suffix),
	)
	b.end.regex = fmt.Sprintf(
		regexFormat,
		regexp.QuoteMeta(b.end.prefix),
		regexp.QuoteMeta(id),
		regexp.QuoteMeta(b.end.suffix),
	)
}

func (b *block) updateRegexWithTitle(title string) {
	regexFormat := `%s%s%s`
	b.start.regex = regexp.QuoteMeta(fmt.Sprintf(regexFormat, b.start.prefix, title, b.start.suffix))
	b.end.regex = regexp.QuoteMeta(fmt.Sprintf(regexFormat, b.end.prefix, title, b.end.suffix))
}

func (b *block) replaceContent(replacement string) bool {
	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s(.*?)%s`, b.start.regex, b.end.regex))
	if re.MatchString(b.contents) {

		if replacement == "" {
			contents := ""
			for _, match := range re.FindAllStringIndex(b.contents, -1) {
				contents += b.contents[:match[0]-1] + b.contents[match[1]+1:]
			}
			b.contents = contents
			return true
		}

		b.contents = utils.ReplaceAllStringSubmatchFunc(re, b.contents, func(match []string) string {
			return b.format(replacement)
		})

		return true
	}
	return false
}

// Checks if a block is marked as "No Action", meaning the contents contain:
//
//	{prefix} No {title} [{id}] {suffix}
//
// or using the legacy format:
//
//	{prefix} No {title} {suffix}.
//
// e.g.
//
//	<!-- No SDK Installation [installation] -->
//	The installation block is marked as "No Action".
//	It is disabled and should not get updated.
//
// Arguments:
// - prefix: The block boundary prefix, e.g. `<!--`.
// - title: The block title (used in "legacy mode" only).
// - id: The block ID.
// - suffix: The block boundary suffix, e.g. `-->`.
// - contents: The README markdown contents.
//
// Returns:
// - bool: True if the block is marked as "No Action", false otherwise.
func IsBlockDisabled(prefix, title, id, suffix, contents string) bool {

	noActionPrefix := prefix + " No "
	noActionSuffix := " " + suffix

	regexWithID := fmt.Sprintf(
		`%s[\w\s-]+\[%s\]%s`,
		regexp.QuoteMeta(noActionPrefix),
		regexp.QuoteMeta(id),
		regexp.QuoteMeta(noActionSuffix),
	)

	re := regexp.MustCompile(regexWithID)

	if re.MatchString(contents) {
		return true
	}

	// falling back on using title (legacy mode)
	regexWithTitle := fmt.Sprintf(
		`%s%s%s`,
		regexp.QuoteMeta(noActionPrefix),
		regexp.QuoteMeta(title),
		regexp.QuoteMeta(noActionSuffix),
	)

	re = regexp.MustCompile(regexWithTitle)

	return re.MatchString(contents)
}

// Retrieves section IDs from all Start block boundaries found inside the README.
//
// Arguments:
// - prefix: The block boundary prefix, e.g. `<!--`.
// - suffix: The block boundary suffix, e.g. `-->`.
// - contents: The README markdown contents.
//
// Returns:
// - string: A comma-separated list of all section IDs found.
func ParseBlockIDs(prefix, suffix, contents string) string {

	idMap := make(map[string]struct{}) // tracks duplicates
	idSlice := []string{}              // maintains ordering

	startBlockRegex := fmt.Sprintf(
		`%s Start [\w\s-]+\[([a-z\-]+)\] %s`,
		regexp.QuoteMeta(prefix),
		regexp.QuoteMeta(suffix),
	)

	re := regexp.MustCompile(startBlockRegex)

	for _, match := range re.FindAllStringSubmatch(contents, -1) {
		if len(match) > 1 {
			id := match[1]
			if _, exists := idMap[id]; !exists {
				idMap[id] = struct{}{}
				idSlice = append(idSlice, id)
			}
		}
	}

	return strings.Join(idSlice, ",")
}

// Retrieves the content of the block associated with a given ID.
//
// Arguments:
// - prefix: The block boundary prefix, e.g. `<!--`.
// - id: The ID of the block to be parsed.
// - suffix: The block boundary suffix, e.g. `-->`.
// - contents: The README markdown contents.
//
// Returns:
// - string: The content of the block associated with the specified ID.
func ParseBlockWithID(prefix, id, suffix, contents string) string {

	prefix = regexp.QuoteMeta(prefix)
	suffix = regexp.QuoteMeta(suffix)
	id = regexp.QuoteMeta(id)

	startRegex := fmt.Sprintf(`%s Start [\w\s-]+\[%s\] %s`, prefix, id, suffix)
	endRegex := fmt.Sprintf(`%s End [\w\s-]+\[%s\] %s`, prefix, id, suffix)

	blockRegex := fmt.Sprintf(`(?s)%s(.*?)%s`, startRegex, endRegex)

	re := regexp.MustCompile(blockRegex)

	match := re.FindStringSubmatch(contents)

	if len(match) > 1 {
		return match[1]
	}

	return ""
}
