package formattemplateerror

import (
	"regexp"
	"strings"
)

// formatTemplateErrorString takes a flat error string from Go’s text/template execution
// and splits it into a nicely indented, multi-line form.
// WARNING: This was vibe-coded
func FormatTemplateErrorString(errStr string) string {
	if errStr == "" {
		return ""
	}

	// Split the error string into logical parts based on known patterns
	result := processError(errStr, 0)
	return strings.TrimSuffix(result, "\n")
}

func processError(errStr string, baseDepth int) string {
	var sb strings.Builder
	indent := strings.Repeat(" ", baseDepth*2)

	// Check for GoError pattern
	if strings.HasPrefix(errStr, "GoError: failed to execute template:") {
		sb.WriteString(indent + "GoError: failed to execute template:\n")
		rest := strings.TrimPrefix(errStr, "GoError: failed to execute template:")
		rest = strings.TrimSpace(rest)
		sb.WriteString(processError(rest, baseDepth+1))
		return sb.String()
	}

	// Check for plain "failed to execute template:" pattern
	if strings.HasPrefix(errStr, "failed to execute template:") {
		sb.WriteString(indent + "failed to execute template:\n")
		rest := strings.TrimPrefix(errStr, "failed to execute template:")
		rest = strings.TrimSpace(rest)
		sb.WriteString(processError(rest, baseDepth+1))
		return sb.String()
	}

	// Check for template pattern
	templateRe := regexp.MustCompile(`^template: ([^:]+:\d+:\d+)(.*)`)
	if match := templateRe.FindStringSubmatch(errStr); match != nil {
		sb.WriteString(indent + "template: " + match[1] + "\n")
		rest := strings.TrimSpace(match[2])
		if rest != "" && rest != ":" {
			if strings.HasPrefix(rest, ":") {
				rest = strings.TrimSpace(rest[1:])
			}
			if rest != "" {
				sb.WriteString(processError(rest, baseDepth+1))
			}
		}
		return sb.String()
	}

	// Check for executing pattern
	executingRe := regexp.MustCompile(`^executing "([^"]*)" at <([^>]*)>(.*)`)
	if match := executingRe.FindStringSubmatch(errStr); match != nil {
		rest := strings.TrimSpace(match[3])
		switch {
		case strings.HasPrefix(rest, ":"):
			rest = strings.TrimSpace(rest[1:])
			// Rule: if the next content starts with "error calling" followed by a NON-template error,
			// don't add a colon (same logical level). If it's followed by a template error, add colon.
			if strings.HasPrefix(rest, "error calling") && !hasNestedTemplateError(rest) {
				sb.WriteString(indent + "executing \"" + match[1] + "\" at <" + match[2] + ">\n")
				sb.WriteString(processError(rest, baseDepth))
			} else {
				sb.WriteString(indent + "executing \"" + match[1] + "\" at <" + match[2] + ">:\n")
				if rest != "" {
					sb.WriteString(processError(rest, baseDepth))
				}
			}
		case rest != "":
			// No colon in original, so don't add one and keep same indentation
			sb.WriteString(indent + "executing \"" + match[1] + "\" at <" + match[2] + ">\n")
			sb.WriteString(processError(rest, baseDepth))
		default:
			// No rest content
			sb.WriteString(indent + "executing \"" + match[1] + "\" at <" + match[2] + ">\n")
		}
		return sb.String()
	}

	// Check for error calling pattern
	errorCallingRe := regexp.MustCompile(`^error calling ([^:]+):(.*)`)
	if match := errorCallingRe.FindStringSubmatch(errStr); match != nil {
		sb.WriteString(indent + "error calling " + match[1] + ":\n")
		rest := strings.TrimSpace(match[2])
		if rest != "" {
			sb.WriteString(processError(rest, baseDepth+1))
		}
		return sb.String()
	}

	// Check for "at" pattern (stack trace lines) - handles both start of string and space-prefixed
	// This needs to handle nested parentheses like "at templateImports (includes/imports.ts:361:17(47))"
	atRe := regexp.MustCompile(`^(\s*)at ([^(]*\([^)]*\([^)]*\)[^)]*\))(.*)`)
	if match := atRe.FindStringSubmatch(errStr); match != nil {
		sb.WriteString(indent + "at " + match[2] + "\n")
		rest := strings.TrimSpace(match[3])
		if rest != "" {
			sb.WriteString(processError(rest, baseDepth))
		}
		return sb.String()
	}

	// Also handle simple "at" patterns like "at github.com/... (native)"
	// Use non-greedy matching to stop at the first complete parenthetical expression before " at" or end
	simpleAtRe := regexp.MustCompile(`^(\s*)at (.+?\([^)]*\))(\s+at.*|$)`)
	if match := simpleAtRe.FindStringSubmatch(errStr); match != nil {
		sb.WriteString(indent + "at " + match[2] + "\n")
		rest := strings.TrimSpace(match[3])
		if rest != "" {
			sb.WriteString(processError(rest, baseDepth))
		}
		return sb.String()
	}

	// Check for "Error:" at the start (like "Error: Tried to import duplicate: models")
	if strings.HasPrefix(errStr, "Error:") {
		// Find the end of this error message (before the next "at" clause)
		atIndex := strings.Index(errStr, " at ")
		if atIndex > 0 {
			errorMsg := errStr[:atIndex]
			rest := errStr[atIndex+1:] // +1 to include the space before "at"
			sb.WriteString(indent + errorMsg + "\n")
			sb.WriteString(processError(rest, baseDepth))
		} else {
			sb.WriteString(indent + errStr + "\n")
		}
		return sb.String()
	}

	// If we have any remaining text that doesn't match patterns, just add it
	if strings.TrimSpace(errStr) != "" && strings.TrimSpace(errStr) != ":" {
		sb.WriteString(indent + strings.TrimSpace(errStr) + "\n")
	}

	return sb.String()
}

// hasNestedTemplateError checks if the error string will eventually lead to another template error
// (as opposed to a direct error message). This helps determine nesting levels.
func hasNestedTemplateError(errStr string) bool {
	// Check if it starts with "error calling" followed by a template error
	if strings.HasPrefix(errStr, "error calling") {
		// Find the end of the "error calling xxx:" part
		colonIndex := strings.Index(errStr, ":")
		if colonIndex > 0 && colonIndex < len(errStr)-1 {
			afterColon := strings.TrimSpace(errStr[colonIndex+1:])
			// Check if what follows is another template error pattern
			return strings.HasPrefix(afterColon, "GoError: failed to execute template") ||
				strings.HasPrefix(afterColon, "failed to execute template")
		}
	}
	return false
}
