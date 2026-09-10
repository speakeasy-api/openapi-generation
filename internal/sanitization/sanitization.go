//go:generate go run generate/generate_lookup_table.go
//go:build !js || !wasm

package sanitization

import (
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"sync"

	"github.com/ettle/strcase"
	"github.com/gammban/numtow"
	"github.com/gammban/numtow/lang"
	"github.com/gammban/numtow/lang/en"
)

var (
	illegalChars              = `\^(|):.,@/\\$+\-<>=\[\]#&%{}'"*?;~’!£ ` + "`"
	numberPrefixRegex         = regexp.MustCompile(`(^[0-9]+)`)
	filenameIllegalCharsRegex = regexp.MustCompile(fmt.Sprintf(`([%s_])`, illegalChars))

	sanitizeNameCacheMu sync.RWMutex
	sanitizeNameCache   = map[string]string{}
)

func SanitizeName(name string) string {
	sanitizeNameCacheMu.RLock()
	if cached, ok := sanitizeNameCache[name]; ok {
		sanitizeNameCacheMu.RUnlock()
		return cached
	}
	sanitizeNameCacheMu.RUnlock()

	result := sanitizeNameUncached(name)

	sanitizeNameCacheMu.Lock()
	sanitizeNameCache[name] = result
	sanitizeNameCacheMu.Unlock()

	return result
}

func sanitizeNameUncached(name string) string {
	runes := []rune(name)

	var sb strings.Builder

	prefix := true

	var first rune
	var last rune

	for i, r := range runes {
		if i == 0 {
			first = r
		}

		// Check if the current rune is a diatritic and if so, replace it with the appropriate replacement
		if replacement, ok := normalizedDiatriticsLookup[r]; ok {
			sb.WriteRune(replacement)
			last = replacement
			prefix = false
		} else {
			switch {
			case r > unicode.MaxASCII:
				// If the current rune is not ASCII and the last rune was not an underscore and we are not at the prefix, add an underscore otherwise skip it
				if last != '_' && !prefix {
					sb.WriteRune('_')
					last = '_'
				}
			case r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9':
				// if the current rune is a letter or number, add it to the sanitized name
				if r < '0' || r > '9' {
					prefix = false // when we encounter the first letter we are no longer at the prefix
				}
				sb.WriteRune(r)
				last = r
			case prefix:
				// if we are at the prefix and weren't a letter or number, replace the current rune with the appropriate prefix replacement
				replacement := replacePrefixChar(r)
				if replacement != "" && (replacement != "_" || last != '_') {
					sb.WriteString(replacement)
					last = rune(replacement[len(replacement)-1])
				}
			default:
				// else if we are here we are likely an ascii symbol that needs to be replaced
				replacement := replaceChar(r)
				if replacement != "_" || last != '_' {
					sb.WriteString(replacement)
					last = rune(replacement[len(replacement)-1])
				}
			}
		}
	}

	// if the start of the string is a number then we need to replace it with the appropriate number replacement
	if first >= '0' && first <= '9' {
		// TODO think about how to replace this
		return numberPrefixRegex.ReplaceAllStringFunc(sb.String(), func(s string) string {
			return numtow.MustString(s, lang.EN, en.WithFmtGroupSep(""))
		})
	}

	return sb.String()
}

func SanitizeFile(name string, replacement string) string {
	name = SanitizeName(name)

	return filenameIllegalCharsRegex.ReplaceAllString(name, replacement)
}

// Convert symbols with meaning to a name otherwise if they are likely dividers an empty string
func replacePrefixChar(char rune) string {
	switch char {
	case '$':
		return "Dollar_"
	case '+':
		return "Plus_"
	case '-':
		return "Minus_"
	case '<':
		return "LessThan_"
	case '>':
		return "GreaterThan_"
	case '=':
		return "Equal_"
	case '@':
		return "At_"
	case '#':
		return "Number_"
	case '&':
		return "And_"
	case '%':
		return "Percent_"
	case '*':
		return "Wildcard_"
	case '/':
		return "Root_"
	case '!':
		return "Not_"
	case '.':
		return "Dot_"
	case '£':
		return "Pound_"
	case '_':
		return "_"
	default:
		return ""
	}
}

func replaceChar(char rune) string {
	switch char {
	case '$':
		return "Dollar_"
	case '+':
		return "Plus_"
	case '<':
		return "LessThan_"
	case '>':
		return "GreaterThan_"
	case '=':
		return "Equal_"
	case '@':
		return "At_"
	case '#':
		return "Number_"
	case '&':
		return "And_"
	case '%':
		return "Percent_"
	case '*':
		return "Wildcard_"
	case '£':
		return "Pound_"
	default:
		return "_"
	}
}

func HumanizeMediaType(mediaType string) string {
	safeMediaType, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		mediaType = safeMediaType
	}

	mediaType = strings.ReplaceAll(mediaType, "+", "_")
	mediaType = strings.ReplaceAll(mediaType, "*", "")
	split := strings.Split(mediaType, "/")
	topLevel := ""
	subLevel := ""
	if len(split) > 0 {
		topLevel = split[0]
	}
	if len(split) > 1 {
		subLevel = split[1]
	}

	if strings.ToLower(topLevel) == "application" || strings.ToLower(topLevel) == "text" {
		topLevel = ""
	}

	if slices.Contains([]string{
		"json", "ld+json", "x-ndjson",
		"json-seq", "problem_json", "vnd.api_json",
		"ld_json",
	}, strings.ToLower(subLevel)) {
		subLevel = ""
	}

	if strings.ToLower(subLevel) == "x-www-form-urlencoded" {
		subLevel = "FormEncoded"
	}

	if strings.ToLower(topLevel) == "multipart" && strings.ToLower(subLevel) == "form-data" {
		topLevel = ""
		subLevel = "FormEncoded"
	}

	if strings.HasPrefix(strings.ToLower(subLevel), "x-") {
		subLevel = subLevel[len("x-"):]
	}

	if strings.HasPrefix(strings.ToLower(subLevel), "vnd.") {
		subLevel = subLevel[len("vnd."):]
	}

	return strcase.ToGoPascal(topLevel + "_" + subLevel)
}

func HumanizeStatusCode(statusCode string) string {
	statusCode = strings.TrimSpace(statusCode)
	statusCode = strings.ToUpper(statusCode)

	if statusCode == "default" || strings.HasPrefix(statusCode, "2") {
		return ""
	}

	if statusCode == "1XX" {
		return "Informational"
	}
	if statusCode == "3XX" {
		return "Redirect"
	}
	if statusCode == "4XX" {
		return "ClientError"
	}
	if statusCode == "5XX" {
		return "ServerError"
	}
	statusCodeInt, err := strconv.Atoi(statusCode)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(http.StatusText(statusCodeInt), " ", "")
}
