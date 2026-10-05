package sanitization

import (
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type Sanitizer struct{}

type Result struct {
	Original string
	Results  *sequencedmap.Map[string, string]
}

func (r *Result) GetConflicting(rhs Result) (string, string) {
	for lhsKey, lhsValue := range r.Results.All() {
		rhsValue, ok := rhs.Results.Get(lhsKey)
		if !ok {
			continue
		}

		if lhsValue == rhsValue {
			return lhsValue, rhsValue
		}
	}

	return "", ""
}

func SanitizeClassName(className string) string {
	return casing.New().ToPascal(sanitization.SanitizeName(className))
}

func SanitizeFieldName(fieldName string) string {
	return casing.New().ToGoPascal(sanitization.SanitizeName(fieldName))
}

func GetSanitizedFileNameResult(filename string) Result {
	// This will try and sanitize the file names using a selection of common methods from our targets
	// We only need to add new ones here if they meaningfully differ from the existing ones
	// and can cause conflicts that aren't caught by the existing methods
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("go", func(s string) string {
		return strings.ToLower(sanitization.SanitizeFile(s, ""))
	})
	sanitizationMethods.Set("java", func(s string) string {
		return sanitization.SanitizeFile(casing.New().ToPascal(sanitization.SanitizeName(s)), "")
	})
	sanitizationMethods.Set("python", func(s string) string {
		return strings.ToLower(sanitization.SanitizeFile(s, "_"))
	})

	return getResult(filename, sanitizationMethods)
}

func GetSanitizedOperationIDResult(operationID string) Result {
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("go", strings.TrimSpace)

	return getResult(operationID, sanitizationMethods)
}

func GetSanitizedMethodNameResult(methodName string) Result {
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("typescript", func(s string) string {
		return casing.New().ToCamel(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("python", func(s string) string {
		return casing.New().ToSnake(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("csharp", func(s string) string {
		return casing.New().ToPascal(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("go", func(s string) string {
		return casing.New().ToGoPascal(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("php", func(s string) string {
		return strings.ToLower(casing.New().ToCamel(sanitization.SanitizeName(s)))
	})

	return getResult(methodName, sanitizationMethods)
}

func GetSanitizedFieldNameResult(fieldName string) Result {
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("go", func(s string) string {
		return casing.New().ToGoPascal(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("java", func(s string) string {
		return casing.New().ToCamel(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("python", func(s string) string {
		s = strings.TrimSuffix(sanitization.SanitizeName(s), "_")
		return casing.New().ToSnake(s)
	})

	return getResult(fieldName, sanitizationMethods)
}

func GetSanitizedClassNameResult(className string) Result {
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("csharp", func(s string) string {
		return casing.New().ToPascal(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("go", func(s string) string {
		return casing.New().ToGoPascal(sanitization.SanitizeName(s))
	})

	return getResult(className, sanitizationMethods)
}

func GetSanitizedNamespaceResult(namespace string) Result {
	// This will try and sanitize the namespace names using a selection of common methods from our targets
	// We only need to add new ones here if they meaningfully differ from the existing ones
	// and can cause conflicts that aren't caught by the existing methods
	//
	// Each sanitizer operates per path segment (split on "/") to mirror the template-level
	// sanitizeOutputLocation() behavior. This prevents false collisions between nested
	// namespaces like "foo/bar" and a flat "foobar".
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("go", perForwardSlashSegment(func(s string) string {
		return strings.ToLower(sanitization.SanitizeFile(s, ""))
	}))
	sanitizationMethods.Set("python", perForwardSlashSegment(func(s string) string {
		return casing.New().ToSnake(sanitization.SanitizeName(s))
	}))
	sanitizationMethods.Set("java", perForwardSlashSegment(func(s string) string {
		return strings.ToLower(sanitization.SanitizeName(s))
	}))
	sanitizationMethods.Set("csharp", perForwardSlashSegment(func(s string) string {
		return casing.New().ToPascal(sanitization.SanitizeName(s))
	}))
	sanitizationMethods.Set("typescript", perForwardSlashSegment(func(s string) string {
		return casing.New().ToCamel(sanitization.SanitizeName(s))
	}))
	sanitizationMethods.Set("php", perForwardSlashSegment(func(s string) string {
		return casing.New().ToPascal(sanitization.SanitizeName(s))
	}))

	return getResult(namespace, sanitizationMethods)
}

// perForwardSlashSegment wraps a sanitizer to apply it per "/" segment, mirroring the template-level
// sanitizeOutputLocation() which splits, sanitizes each segment, and joins.
func perForwardSlashSegment(fn func(string) string) func(string) string {
	return func(s string) string {
		segments := strings.Split(s, "/")
		for i, seg := range segments {
			segments[i] = fn(seg)
		}
		return strings.Join(segments, "/")
	}
}

func GetSanitizedEnumNamesResult(values []string) []Result {
	sanitizationMethods := getEnumSanitizationMethods()
	nameMethods := getEnumNameSanitizationMethods()

	uniqueEnumNames := sequencedmap.New[string, map[string]int]()

	for key, sanitizer := range sanitizationMethods.All() {
		unique, ok := uniqueEnumNames.Get(key)
		if !ok {
			unique = map[string]int{}
		}

		for _, value := range values {
			name := sanitizer(value)
			if _, ok := unique[name]; !ok {
				unique[name] = 0
			}

			unique[name]++
		}

		uniqueEnumNames.Set(key, unique)
	}

	resultsByValue := sequencedmap.New[string, Result]()

	for key := range sanitizationMethods.Keys() {
		unique, _ := uniqueEnumNames.Get(key)

		for _, value := range values {
			b, _ := sanitizationMethods.Get(key)
			m, _ := nameMethods.Get(key)

			name := m(value, unique[b(value)])

			res, ok := resultsByValue.Get(value)
			if !ok {
				res = Result{
					Original: value,
					Results:  sequencedmap.New[string, string](),
				}
			}

			res.Results.Set(key, name)
			resultsByValue.Set(value, res)
		}
	}

	return slices.Collect(resultsByValue.Values())
}

func GetSanitizedEnumNameResult(enum string) Result {
	return getResult(enum, getEnumSanitizationMethods())
}

func getEnumSanitizationMethods() *sequencedmap.Map[string, func(string) string] {
	sanitizationMethods := sequencedmap.New[string, func(string) string]()
	sanitizationMethods.Set("csharp", func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = "Unknown"
		}

		return casing.New().ToPascal(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("go", func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = "Unknown"
		}

		return casing.New().ToGoPascalIgnoreAcronyms(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("java", func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = "Unknown"
		}

		return casing.New().ToPascal(sanitization.SanitizeName(s))
	})
	sanitizationMethods.Set("python", func(s string) string {
		s = strings.TrimSpace(s)
		if s == "" {
			s = "unknown"
		}

		return casing.New().ToSNAKE(sanitization.SanitizeName(s))
	})

	return sanitizationMethods
}

func getEnumNameSanitizationMethods() *sequencedmap.Map[string, func(string, int) string] {
	baseMethods := getEnumSanitizationMethods()

	baseDedupeMethod := func(lang, value string, count int) string {
		c := casing.New()

		m, _ := baseMethods.Get(lang)
		s := m(value)
		if count > 1 {
			s += c.ToPascal(getCasing(value))
		}
		return s
	}

	nameMethods := sequencedmap.New[string, func(string, int) string]()
	nameMethods.Set("csharp", func(s string, count int) string {
		s = baseDedupeMethod("csharp", s, count)
		return casing.New().ToPascal(s)
	})
	nameMethods.Set("go", func(s string, count int) string {
		s = baseDedupeMethod("go", s, count)
		return casing.New().ToGoPascal(s)
	})
	nameMethods.Set("java", func(s string, count int) string {
		s = baseDedupeMethod("java", s, count)
		c := casing.New()
		return c.ToSNAKE(c.ToCamel(s))
	})
	nameMethods.Set("python", func(value string, count int) string {
		m, _ := baseMethods.Get("python")
		s := m(value)
		if count > 1 {
			s = s + "_" + strings.ToUpper(getCasing(value))
		}
		return s
	})

	return nameMethods
}

func getCasing(s string) string {
	switch {
	case strings.ToUpper(s) == s:
		return "upper"
	case strings.ToLower(s) == s:
		return "lower"
	default:
		return "mixed"
	}
}

func getResult(orig string, sanitizationMethods *sequencedmap.Map[string, func(string) string]) Result {
	results := sequencedmap.New[string, string]()

	for k, sanitizer := range sanitizationMethods.All() {
		results.Set(k, sanitizer(orig))
	}

	return Result{
		Original: orig,
		Results:  results,
	}
}
