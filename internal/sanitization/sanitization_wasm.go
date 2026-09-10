//go:build js && wasm

package sanitization

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/sync/singleflight"
)

var (
	illegalChars              = `\^(|):.,@/\\$+\-<>=\[\]#&%{}'"*?;~'!£ ` + "`"
	numberPrefixRegex         = regexp.MustCompile(`(^[0-9]+)`)
	filenameIllegalCharsRegex = regexp.MustCompile(fmt.Sprintf(`([%s_])`, illegalChars))
	sanitizeNameGroup         = new(singleflight.Group)
)

// Simple number to word conversion for WASM (replaces numtow dependency)
func numberToWords(num string) string {
	// Simple mapping for common numbers - much lighter than full numtow
	switch num {
	case "0":
		return "zero"
	case "1":
		return "one"
	case "2":
		return "two"
	case "3":
		return "three"
	case "4":
		return "four"
	case "5":
		return "five"
	case "6":
		return "six"
	case "7":
		return "seven"
	case "8":
		return "eight"
	case "9":
		return "nine"
	case "10":
		return "ten"
	default:
		// For longer numbers, use a simple prefix approach
		if n, err := strconv.Atoi(num); err == nil && n < 100 {
			if n < 20 {
				teens := []string{"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
				if n >= 10 {
					return teens[n-10]
				}
			} else {
				tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
				units := []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
				return tens[n/10] + units[n%10]
			}
		}
		// For very large numbers or complex cases, just use "num" + the number
		return "num" + num
	}
}

func SanitizeName(name string) string {
	rawResult, _, _ := sanitizeNameGroup.Do(name, func() (any, error) {
		runes := []rune(name)

		var sb strings.Builder

		prefix := true

		var first rune
		var last rune

		for i, r := range runes {
			if i == 0 {
				first = r
			}

			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				prefix = false
				sb.WriteRune(r)
				last = r
			} else if prefix {
				// if we are at the prefix and weren't a letter or number, replace the current rune with the appropriate prefix replacement
				replacement := replacePrefixChar(r)
				if replacement != "" && (replacement != "_" || last != '_') {
					sb.WriteString(replacement)
					last = rune(replacement[len(replacement)-1])
				}
			} else {
				// else if we are here we are likely an ascii symbol that needs to be replaced
				replacement := replaceChar(r)
				if replacement != "_" || last != '_' {
					sb.WriteString(replacement)
					last = rune(replacement[len(replacement)-1])
				}
			}
		}

		// if the start of the string is a number then we need to replace it with the appropriate number replacement
		if first >= '0' && first <= '9' {
			// Use our simple number to words conversion instead of numtow
			return numberPrefixRegex.ReplaceAllStringFunc(sb.String(), func(s string) string {
				return numberToWords(s)
			}), nil
		}

		return sb.String(), nil
	})

	return rawResult.(string)
}

func SanitizeFile(name string, replacement string) string {
	name = SanitizeName(name)

	return filenameIllegalCharsRegex.ReplaceAllString(name, replacement)
}

func HumanizeMediaType(mediaType string) string {
	// Simple implementation for WASM
	return strings.ReplaceAll(mediaType, "/", "_")
}

func HumanizeStatusCode(statusCode string) string {
	// Simple implementation for WASM
	switch statusCode {
	case "200":
		return "OK"
	case "201":
		return "Created"
	case "400":
		return "BadRequest"
	case "401":
		return "Unauthorized"
	case "403":
		return "Forbidden"
	case "404":
		return "NotFound"
	case "500":
		return "InternalServerError"
	default:
		return "Status" + statusCode
	}
}

// Convert symbols with meaning to a name otherwise if they are likely dividers an empty string
func replacePrefixChar(char rune) string {
	switch char {
	case '$':
		return "dollar"
	case '@':
		return "at"
	case '&':
		return "and"
	case '%':
		return "percent"
	case '#':
		return "hash"
	case '+':
		return "plus"
	case '-':
		return "minus"
	case '=':
		return "equals"
	case '<':
		return "less"
	case '>':
		return "greater"
	case '!':
		return "exclamation"
	case '?':
		return "question"
	case '*':
		return "star"
	case '/':
		return "slash"
	case '\\':
		return "backslash"
	case '|':
		return "pipe"
	case '^':
		return "caret"
	case '~':
		return "tilde"
	case '`':
		return "backtick"
	case '\'':
		return "quote"
	case '"':
		return "doublequote"
	default:
		return ""
	}
}

func replaceChar(char rune) string {
	switch char {
	case '$':
		return "_dollar_"
	case '@':
		return "_at_"
	case '&':
		return "_and_"
	case '%':
		return "_percent_"
	case '#':
		return "_hash_"
	case '+':
		return "_plus_"
	case '-':
		return "_minus_"
	case '=':
		return "_equals_"
	case '<':
		return "_less_"
	case '>':
		return "_greater_"
	case '!':
		return "_exclamation_"
	case '?':
		return "_question_"
	case '*':
		return "_star_"
	case '/':
		return "_slash_"
	case '\\':
		return "_backslash_"
	case '|':
		return "_pipe_"
	case '^':
		return "_caret_"
	case '~':
		return "_tilde_"
	case '`':
		return "_backtick_"
	case '\'':
		return "_quote_"
	case '"':
		return "_doublequote_"
	case ':':
		return "_colon_"
	case ';':
		return "_semicolon_"
	case ',':
		return "_comma_"
	case '.':
		return "_dot_"
	case '(':
		return "_open_paren_"
	case ')':
		return "_close_paren_"
	case '[':
		return "_open_bracket_"
	case ']':
		return "_close_bracket_"
	case '{':
		return "_open_brace_"
	case '}':
		return "_close_brace_"
	default:
		return "_"
	}
}
