package utils

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/references"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"gopkg.in/yaml.v3"
)

var (
	numberAtEndRegex = regexp.MustCompile(`(\d+)$`)

	paramTypeToSuffix = map[string]string{
		"pathParam":  "PathParameter",
		"queryParam": "QueryParameter",
	}
)

func MapToSequencedMap[K cmp.Ordered, V any](m map[K]V) *sequencedmap.Map[K, V] {
	om := sequencedmap.New[K, V]()

	if m == nil {
		return nil
	}

	keys := GetSortedKeys(m)

	for _, k := range keys {
		om.Set(k, m[k])
	}

	return om
}

func GetSortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	slices.SortStableFunc(keys, func(i, j K) int {
		switch {
		case i == j:
			return 0
		case i < j:
			return -1
		default:
			return 1
		}
	})

	return keys
}

func AppendSorted[A any](a []A, i A, comp func(A, A) bool) []A {
	aa := a

	idx := sort.Search(len(aa), func(idx int) bool {
		return comp(a[idx], i)
	})
	aa = append(aa, i)
	copy(aa[idx+1:], aa[idx:])
	aa[idx] = i

	return aa
}

func IncrementName(name string) string {
	if !numberAtEndRegex.MatchString(name) {
		return name + "1"
	}

	return numberAtEndRegex.ReplaceAllStringFunc(name, func(s string) string {
		i, err := strconv.Atoi(s)
		if err != nil {
			i = 0
		}

		return strconv.Itoa(i + 1)
	})
}

func RemoveEndNumber(name string) string {
	if numberAtEndRegex.MatchString(name) {
		return numberAtEndRegex.ReplaceAllString(name, "")
	}
	return name
}

func MapArray[T any, O any](vs []T, f func(T) O) []O {
	vsm := make([]O, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func MapToArray[T comparable, O any](m map[T]O) []O {
	a := make([]O, 0, len(m))
	for _, v := range m {
		a = append(a, v)
	}
	return a
}

func ReplaceAllStringSubmatchFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	var result strings.Builder
	lastIndex := 0

	for _, v := range re.FindAllStringSubmatchIndex(str, -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			if v[i] == -1 || v[i+1] == -1 {
				groups = append(groups, "")
			} else {
				groups = append(groups, str[v[i]:v[i+1]])
			}
		}

		result.WriteString(str[lastIndex:v[0]])
		result.WriteString(repl(groups))
		lastIndex = v[1]
	}

	result.WriteString(str[lastIndex:])
	return result.String()
}

func SuffixParamType(name, paramType string) string {
	return fmt.Sprintf("%s%s", name, paramTypeToSuffix[paramType])
}

func ToBool(value interface{}) bool {
	return value != nil && value.(bool)
}

func GetEndLine(node *yaml.Node) int {
	if len(node.Content) == 0 {
		return node.Line
	}

	lastChild := node.Content[len(node.Content)-1]
	return GetEndLine(lastChild)
}

// GetSimplifiedRef will return the reference without the preceding file path
// Though if we get a customer that has the same ref in two different files they include this may identify them incorrectly
// But that will be a problem anyway as libopenapi when returning references from an external file won't include the file path
// for a local reference with that file and so we might fail to distinguish between them that way.
// The fix we need is for libopenapi to also track which file the reference is in so we can always prefix them with the file path
func GetSimplifiedRef(ref string) references.Reference {
	if ref == "" {
		return ""
	}

	refParts := strings.Split(ref, "#/")
	return references.Reference("#/" + refParts[len(refParts)-1])
}

func ExtensionsToGoMap(extensions *extensions.Extensions) map[string]any {
	if extensions.Len() == 0 {
		return map[string]any{}
	}

	m := map[string]any{}
	for k, v := range extensions.All() {
		var val any
		if err := v.Decode(&val); err != nil {
			continue
		}

		m[k] = val
	}

	return m
}

func Dedent(s string) string {
	lines := strings.Split(s, "\n")
	// Remove leading and trailing empty lines
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	lines = lines[start:end]

	// Find the minimum indentation
	minIndent := -1
	for _, line := range lines {
		trimmedLine := strings.TrimLeft(line, " \t")
		if trimmedLine == "" {
			continue
		}
		indent := len(line) - len(trimmedLine)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// Dedent lines
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// A OneManQueue is much like golang's SingleFlight but not every call has to execute.
// Like SingleFlight, if there's an ongoing call it will enqueue the call. At most
// At most one call live in the queue. Hence the name. The last call always replaces
// the currently enqueued call.
func OneManQueue(fn func()) func() {
	mutex := sync.Mutex{}
	queueSize := 0

	return func() {
		mutex.Lock()
		if queueSize > 0 {
			if queueSize < 2 {
				// Enqueue this call
				queueSize++
			}
			mutex.Unlock()
			return
		}

		// No function is running; proceed to run
		queueSize++
		mutex.Unlock()

		for {
			fn()

			mutex.Lock()
			queueSize--
			if queueSize == 0 {
				mutex.Unlock()
				return
			}
			mutex.Unlock()
		}
	}
}

func RecursivelyListDirectories(paths ...string) []string {
	var dirs []string
	for _, path := range paths {
		_ = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				dirs = append(dirs, path)
			}
			return nil
		})
	}
	return dirs
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func InsertBefore(s, match, insert string) (string, error) {
	idx := strings.Index(s, match)
	if idx == -1 {
		return s, errors.New("no match found")
	}

	return s[:idx] + insert + s[idx:], nil
}

func InsertBeforeOrAppend(s, match, insert string) string {
	idx := strings.Index(s, match)
	if idx == -1 {
		idx = len(s)
	}
	return s[:idx] + insert + s[idx:]
}
