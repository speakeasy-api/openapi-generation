package imports

import "strings"

// PythonRelativeImport returns a Python relative dotted path from fromPkg to toPkg.
// Both args are absolute dotted package paths.
//
// Output:
//   - Leading dots encode the climb from fromPkg up to the common
//     ancestor of fromPkg and toPkg, plus one. The "plus one" matches
//     Python's convention where the first dot denotes the caller's
//     own package (e.g. "from . import X" means "from caller-pkg import X").
//   - Trailing segments are the path from the common ancestor down to toPkg,
//     joined by dots.
//
// Examples:
//
//	PythonRelativeImport("a", "a.b") == ".b"
//	PythonRelativeImport("a.b", "a.b") == "."
//	PythonRelativeImport("a.b", "a.b.c") == ".c"
//	PythonRelativeImport("a.b.c","a.b") == ".."
//	PythonRelativeImport("a.b.c", "a.b.d") == "..d"
//	PythonRelativeImport("a.b.c","a.d") == "...d"
//
// Empty fromPkg or toPkg are treated as zero-segment paths.
func PythonRelativeImport(fromPkg, toPkg string) string {
	separator := "."

	a := splitOrEmpty(fromPkg, separator)
	b := splitOrEmpty(toPkg, separator)

	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}

	upLevels := len(a) - i
	dots := strings.Repeat(separator, upLevels+1)
	rest := strings.Join(b[i:], separator)

	if rest == "" {
		return dots
	}
	return dots + rest
}

func splitOrEmpty(s, separator string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, separator)
}
