package markdown

import (
	"fmt"
	"strconv"
	"strings"
)

func CreateMarkdownTable(contents [][]string) string {
	largestCellContentLength := 0
	largestNumberOfCells := 0

	for _, row := range contents {
		if len(row) > largestNumberOfCells {
			largestNumberOfCells = len(row)
		}

		for _, value := range row {
			if len(value) > largestCellContentLength {
				largestCellContentLength = len(value)
			}
		}
	}

	if largestCellContentLength < 3 {
		largestCellContentLength = 3
	}

	var table strings.Builder

	handledHeader := false
	for _, row := range contents {
		cellsAdded := 0
		for _, cell := range row {
			table.WriteString("| ")
			fmt.Fprintf(&table, "%-"+strconv.Itoa(largestCellContentLength)+"s", strings.ReplaceAll(cell, "|", "\\|"))
			table.WriteString(" ")
			cellsAdded++
		}
		if cellsAdded < largestNumberOfCells {
			for i := 0; i < largestNumberOfCells-cellsAdded; i++ {
				table.WriteString("| ")
				fmt.Fprintf(&table, "%-"+strconv.Itoa(largestCellContentLength)+"s", "")
				table.WriteString(" ")
			}
		}
		table.WriteString("|\n")

		if !handledHeader {
			for i := 0; i < largestNumberOfCells; i++ {
				table.WriteString("| ")
				table.WriteString(strings.Repeat("-", largestCellContentLength))
				table.WriteString(" ")
			}
			table.WriteString("|\n")
			handledHeader = true
		}
	}

	return strings.Trim(table.String(), "\n")
}
