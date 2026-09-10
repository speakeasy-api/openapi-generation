package markdown

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func CreateMarkdownTable(contents [][]string) string {
	largestNumberOfCells := 0

	for _, row := range contents {
		if len(row) > largestNumberOfCells {
			largestNumberOfCells = len(row)
		}
	}

	// Calculate per-column max widths (using escaped length since pipes get escaped)
	colWidths := make([]int, largestNumberOfCells)
	for _, row := range contents {
		for i, value := range row {
			escapedLen := len(strings.ReplaceAll(value, "|", "\\|"))
			if escapedLen > colWidths[i] {
				colWidths[i] = escapedLen
			}
		}
	}
	for i := range colWidths {
		if colWidths[i] < 3 {
			colWidths[i] = 3
		}
	}

	var table strings.Builder

	handledHeader := false
	for _, row := range contents {
		cellsAdded := 0
		for i, cell := range row {
			table.WriteString("| ")
			fmt.Fprintf(&table, "%-"+strconv.Itoa(colWidths[i])+"s", strings.ReplaceAll(cell, "|", "\\|"))
			table.WriteString(" ")
			cellsAdded++
		}
		if cellsAdded < largestNumberOfCells {
			for i := cellsAdded; i < largestNumberOfCells; i++ {
				table.WriteString("| ")
				fmt.Fprintf(&table, "%-"+strconv.Itoa(colWidths[i])+"s", "")
				table.WriteString(" ")
			}
		}
		table.WriteString("|\n")

		if !handledHeader {
			for i := 0; i < largestNumberOfCells; i++ {
				table.WriteString("| ")
				table.WriteString(strings.Repeat("-", colWidths[i]))
				table.WriteString(" ")
			}
			table.WriteString("|\n")
			handledHeader = true
		}
	}

	return strings.Trim(table.String(), "\n")
}

// headingMatch matches a line starting with one or more # characters followed by
// zero or more spaces followed by at least one non-space.
var headingMatch = regexp.MustCompile(`^(#+)(\s*\S.*)$`)

func FindHeadingLevel(content string) int {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		match := headingMatch.FindStringSubmatch(line)
		if match != nil {
			return len(match[1])
		}
	}
	return 0
}

func AdjustHeadings(parentLevel int, content string) string {
	baseLevel := parentLevel + 1
	presentLevel := FindHeadingLevel(content)
	if presentLevel == 0 {
		return content
	}

	adjustBy := baseLevel - presentLevel
	if adjustBy == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if parts := headingMatch.FindStringSubmatch(line); parts != nil {
			headingLevel := len(parts[1])
			lines[i] = strings.Repeat("#", headingLevel+adjustBy) + parts[2]
		}
	}

	return strings.Join(lines, "\n")
}
