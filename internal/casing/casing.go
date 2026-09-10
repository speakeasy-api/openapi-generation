package casing

import (
	"slices"
	"strings"
	"sync"
	"unicode"

	"github.com/ettle/strcase"
)

var (
	defaultCaser   = strcase.NewCaser(false, map[string]bool{"SDK": true}, nil)
	defaultGoCaser = strcase.NewCaser(true, map[string]bool{"SDK": true}, nil)

	caserCacheMu sync.RWMutex
	caserCache   = map[string]*strcase.Caser{}

	acronymCacheMu sync.RWMutex
	acronymCache   = map[string]map[string]bool{}
)

// Casing provides a set of methods for converting strings to different casing styles.
// It generally tries to retain the original casing of acronyms in the input string, unless otherwise specified.
type Casing struct {
	symbolCasingOverrides []string
}

type symbolCasingOverride struct {
	normalized string
	exact      string
}

func New() *Casing {
	return &Casing{}
}

func NewWithSymbolCasingOverrides(overrides []string) *Casing {
	return &Casing{
		symbolCasingOverrides: slices.Clone(overrides),
	}
}

func CustomCasingOverrides(raw any) []string {
	switch customCasings := raw.(type) {
	case map[string]any:
		return customCasingMapOverrides(customCasings)
	default:
		return nil
	}
}

func customCasingMapOverrides(customCasings map[string]any) []string {
	keys := make([]string, 0, len(customCasings))
	for key := range customCasings {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	overrides := make([]string, 0, len(customCasings))
	for _, key := range keys {
		if exact, ok := customCasingOverride(key, customCasings[key]); ok {
			overrides = append(overrides, exact)
		}
	}

	return overrides
}

func customCasingOverride(key string, raw any) (string, bool) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false
	}

	switch value := raw.(type) {
	case map[string]any:
		if truthy(value["initialism"]) {
			return strings.ToUpper(key), true
		}
		if pascal, ok := stringValue(value["pascal"]); ok {
			return pascal, true
		}
		if capital, ok := stringValue(value["capital"]); ok {
			return capital, true
		}
	case map[string]string:
		if pascal, ok := stringValue(value["pascal"]); ok {
			return pascal, true
		}
		if capital, ok := stringValue(value["capital"]); ok {
			return capital, true
		}
	}

	return "", false
}

func truthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}

func stringValue(value any) (string, bool) {
	str, ok := value.(string)
	if !ok {
		return "", false
	}

	str = strings.TrimSpace(str)
	return str, str != ""
}

// ToGoPascal converts a string to PascalCase using Go initialisms
func (c *Casing) ToGoPascal(s string) string {
	return c.applySymbolCasing(getCaser(true, s).ToPascal(s))
}

// ToGoPascalIgnoreAcronyms converts a string to PascalCase using Go initialisms but ignoring acronyms in the original string
func (c *Casing) ToGoPascalIgnoreAcronyms(s string) string {
	return strcase.ToPascal(s)
}

// ToPascal converts a string to PascalCase
func (c *Casing) ToPascal(s string) string {
	return c.applySymbolCasing(getCaser(false, s).ToPascal(s))
}

// ToSnake converts a string to lower snake_case
func (c *Casing) ToSnake(s string) string {
	return strcase.ToSnake(s)
}

// ToSNAKE converts a string to upper SNAKE_CASE
func (c *Casing) ToSNAKE(s string) string {
	return strcase.ToSNAKE(s)
}

// ToGoCamel converts a string to camelCase using Go initialisms
func (c *Casing) ToGoCamel(s string) string {
	return getCaser(true, s).ToCamel(s)
}

// ToCamel converts a string to camelCase
func (c *Casing) ToCamel(s string) string {
	return getCaser(false, s).ToCamel(s)
}

// ToKebab converts a string to lower kebab-case
func (c *Casing) ToKebab(s string) string {
	return getCaser(false, s).ToKebab(s)
}

// ToKEBAB converts a string to upper KEBAB-CASE
func (c *Casing) ToKEBAB(s string) string {
	return getCaser(false, s).ToKEBAB(s)
}

func (c *Casing) applySymbolCasing(s string) string {
	if c == nil || len(c.symbolCasingOverrides) == 0 {
		return s
	}

	return ApplySymbolCasing(s, c.symbolCasingOverrides)
}

// ApplySymbolCasing rewrites PascalCase word segments to the exact casing
// configured by the caller. It intentionally only replaces complete Pascal
// segments, so "ApiKey" becomes "APIKey" while "Apiary" is left unchanged.
func ApplySymbolCasing(s string, overrides []string) string {
	replacements := make([]symbolCasingOverride, 0, len(overrides))
	seen := map[string]bool{}

	for _, override := range overrides {
		exact := strings.TrimSpace(override)
		if exact == "" {
			continue
		}

		normalized := strcase.ToPascal(strings.ToLower(exact))
		if normalized == "" || seen[normalized] {
			continue
		}

		seen[normalized] = true
		replacements = append(replacements, symbolCasingOverride{
			normalized: normalized,
			exact:      exact,
		})
	}

	slices.SortFunc(replacements, func(a, b symbolCasingOverride) int {
		if len(a.normalized) != len(b.normalized) {
			return len(b.normalized) - len(a.normalized)
		}

		return strings.Compare(a.normalized, b.normalized)
	})

	result := s
	for _, replacement := range replacements {
		result = replaceSymbolCasingSegment(result, replacement.normalized, replacement.exact)
	}

	return result
}

func replaceSymbolCasingSegment(input, normalized, exact string) string {
	var b strings.Builder
	start := 0
	replaced := false

	for {
		relativeIndex := strings.Index(input[start:], normalized)
		if relativeIndex == -1 {
			if !replaced {
				return input
			}

			b.WriteString(input[start:])
			return b.String()
		}

		index := start + relativeIndex
		end := index + len(normalized)

		if isSymbolCasingBoundary(input, end) {
			b.WriteString(input[start:index])
			b.WriteString(exact)
			start = end
			replaced = true
			continue
		}

		b.WriteString(input[start : index+1])
		start = index + 1
	}
}

func isSymbolCasingBoundary(input string, index int) bool {
	if index >= len(input) {
		return true
	}

	r := rune(input[index])
	return unicode.IsUpper(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

func getCaser(golang bool, input string) *strcase.Caser {
	acronyms := cachedFindAcronyms(input)

	if len(acronyms) == 1 {
		if golang {
			return defaultGoCaser
		}

		return defaultCaser
	}

	// Build a stable cache key from the acronym set and golang flag.
	keys := make([]string, 0, len(acronyms))
	for k := range acronyms {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	prefix := "std:"
	if golang {
		prefix = "go:"
	}
	cacheKey := prefix + strings.Join(keys, ",")

	caserCacheMu.RLock()
	if cached, ok := caserCache[cacheKey]; ok {
		caserCacheMu.RUnlock()
		return cached
	}
	caserCacheMu.RUnlock()

	c := strcase.NewCaser(golang, acronyms, nil)

	caserCacheMu.Lock()
	caserCache[cacheKey] = c
	caserCacheMu.Unlock()

	return c
}

func cachedFindAcronyms(input string) map[string]bool {
	acronymCacheMu.RLock()
	if cached, ok := acronymCache[input]; ok {
		acronymCacheMu.RUnlock()
		return cached
	}
	acronymCacheMu.RUnlock()

	acronyms := findAcronyms(input)
	acronyms["SDK"] = true

	acronymCacheMu.Lock()
	acronymCache[input] = acronyms
	acronymCacheMu.Unlock()

	return acronyms
}

func findAcronyms(input string) map[string]bool {
	var current, prev, next rune

	acronyms := map[string]bool{}

	if (strings.ToUpper(input) == input) || (strings.ToLower(input) == input) {
		return acronyms
	}

	var currentAcronym string

	for i, r := range input {
		prev = current
		current = r
		if i < len(input)-1 {
			next = rune(input[i+1])
		} else {
			next = 0
		}

		if unicode.IsUpper(current) {
			if (!unicode.IsLetter(prev) || unicode.IsLower(prev)) && unicode.IsUpper(next) {
				currentAcronym = string(current)
			} else if unicode.IsUpper(prev) && !unicode.IsLower(next) {
				currentAcronym += string(current)
			}
		} else {
			if currentAcronym != "" {
				acronyms[currentAcronym] = true
			}
			currentAcronym = ""
		}
	}

	if currentAcronym != "" {
		acronyms[currentAcronym] = true
	}

	return acronyms
}
