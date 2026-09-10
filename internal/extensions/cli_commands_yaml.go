package extensions

import (
	stderrors "errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// This file contains the raw yaml.Node machinery for the strict
// x-speakeasy-cli-commands decode: alias/merge expansion with duplicate-key
// re-checking, strict mapping access with did-you-mean diagnostics, and the
// singular-JSONPath-to-RFC-6901 path grammar shared by input binds and preset
// keys.

const (
	// cliExpandNodeBudget bounds alias/merge expansion so a hostile document
	// cannot amplify a small YAML body into a huge tree (billion-laughs).
	cliExpandNodeBudget = 50000
	// cliExpandDepthBudget bounds tree depth during expansion, independent of
	// the node budget, so deeply nested alias chains fail fast.
	cliExpandDepthBudget = 200
)

// cliExpandState tracks the node and depth budgets across a whole expansion.
type cliExpandState struct {
	visited int
	depth   int
}

// cliExpandNode returns an alias-free, merge-free deep copy of node. Every
// mapping in the result is re-checked for duplicate keys after expansion, so
// aliases and << merge keys cannot smuggle in duplicates that a purely
// syntactic check would miss.
func cliExpandNode(node *yaml.Node, state *cliExpandState) (*yaml.Node, error) {
	if node == nil {
		return nil, nil
	}
	state.visited++
	if state.visited > cliExpandNodeBudget {
		return nil, fmt.Errorf("document exceeds the %d-node expansion budget (aliases and merge keys are expanded before validation)", cliExpandNodeBudget)
	}
	state.depth++
	defer func() { state.depth-- }()
	if state.depth > cliExpandDepthBudget {
		return nil, fmt.Errorf("document exceeds the %d-level depth budget during alias and merge-key expansion", cliExpandDepthBudget)
	}

	switch node.Kind {
	case yaml.AliasNode:
		if node.Alias == nil {
			return nil, fmt.Errorf("line %d: alias %q does not resolve", node.Line, node.Value)
		}
		// yaml.v3 guarantees anchors precede aliases, so alias expansion
		// cannot recurse into itself; the node budget bounds pathological
		// fan-out.
		return cliExpandNode(node.Alias, state)
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return nil, fmt.Errorf("line %d: expected a single document node", node.Line)
		}
		return cliExpandNode(node.Content[0], state)
	case yaml.ScalarNode:
		copied := *node
		copied.Content = nil
		return &copied, nil
	case yaml.SequenceNode:
		copied := *node
		copied.Content = make([]*yaml.Node, 0, len(node.Content))
		for _, child := range node.Content {
			expanded, err := cliExpandNode(child, state)
			if err != nil {
				return nil, err
			}
			copied.Content = append(copied.Content, expanded)
		}
		return &copied, nil
	case yaml.MappingNode:
		copied := *node
		copied.Content = nil
		seen := map[string]int{}
		appendPair := func(key, value *yaml.Node) error {
			if prev, ok := seen[key.Value]; ok {
				return fmt.Errorf("line %d: duplicate key %q (first defined on line %d; duplicates are rejected after alias and merge-key expansion)", key.Line, key.Value, prev)
			}
			seen[key.Value] = key.Line
			copied.Content = append(copied.Content, key, value)
			return nil
		}
		var mergePairs [][2]*yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			if keyNode.Tag == "!!merge" {
				expandedVal, err := cliExpandNode(valNode, state)
				if err != nil {
					return nil, err
				}
				sources := []*yaml.Node{expandedVal}
				if expandedVal.Kind == yaml.SequenceNode {
					sources = expandedVal.Content
				}
				for _, src := range sources {
					if src.Kind != yaml.MappingNode {
						return nil, fmt.Errorf("line %d: merge key value must be a mapping or sequence of mappings", keyNode.Line)
					}
					for j := 0; j+1 < len(src.Content); j += 2 {
						mergePairs = append(mergePairs, [2]*yaml.Node{src.Content[j], src.Content[j+1]})
					}
				}
				continue
			}
			expandedKey, err := cliExpandNode(keyNode, state)
			if err != nil {
				return nil, err
			}
			expandedVal, err := cliExpandNode(valNode, state)
			if err != nil {
				return nil, err
			}
			if err := appendPair(expandedKey, expandedVal); err != nil {
				return nil, err
			}
		}
		for _, pair := range mergePairs {
			if err := appendPair(pair[0], pair[1]); err != nil {
				return nil, err
			}
		}
		return &copied, nil
	default:
		return nil, fmt.Errorf("line %d: unsupported YAML node kind", node.Line)
	}
}

// cliMapEntry is one key/value pair of a strict mapping, in document order.
type cliMapEntry struct {
	Key   *yaml.Node
	Value *yaml.Node
}

// cliMapEntries requires node to be a mapping and returns its entries in
// document order. The caller receives precise context for the error message.
func cliMapEntries(node *yaml.Node, what string) ([]cliMapEntry, error) {
	if node == nil || node.Kind != yaml.MappingNode {
		line := 0
		if node != nil {
			line = node.Line
		}
		return nil, fmt.Errorf("line %d: %s must be a mapping", line, what)
	}
	entries := make([]cliMapEntry, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("line %d: %s keys must be scalars", node.Content[i].Line, what)
		}
		entries = append(entries, cliMapEntry{Key: node.Content[i], Value: node.Content[i+1]})
	}
	return entries, nil
}

func cliScalarString(node *yaml.Node, what string) (string, error) {
	if node == nil || node.Kind != yaml.ScalarNode || (node.Tag != "!!str" && node.Tag != "!!null") {
		line := 0
		if node != nil {
			line = node.Line
		}
		return "", fmt.Errorf("line %d: %s must be a string", line, what)
	}
	if node.Tag == "!!null" {
		return "", fmt.Errorf("line %d: %s must be a non-empty string", node.Line, what)
	}
	return node.Value, nil
}

func cliScalarBool(node *yaml.Node, what string) (bool, error) {
	if node == nil || node.Kind != yaml.ScalarNode || node.Tag != "!!bool" {
		line := 0
		if node != nil {
			line = node.Line
		}
		return false, fmt.Errorf("line %d: %s must be a boolean", line, what)
	}
	// yaml.v3 tags True/TRUE as !!bool while preserving the authored
	// spelling, so comparing node.Value against "true" would silently invert
	// those. Decode resolves every accepted spelling.
	var b bool
	if err := node.Decode(&b); err != nil {
		return false, fmt.Errorf("line %d: %s must be a boolean: %w", node.Line, what, err)
	}
	return b, nil
}

// cliDecodeValue decodes a YAML node into a plain Go value (maps, slices,
// scalars) while rejecting non-JSON-representable tags such as !!timestamp
// and !!binary — the manifest lowers to JSON-shaped IR.
func cliDecodeValue(node *yaml.Node) (any, error) {
	if node == nil {
		return nil, nil
	}
	switch node.Kind {
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!str":
			return node.Value, nil
		case "!!int":
			n, err := strconv.ParseInt(node.Value, 0, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid integer %q", node.Line, node.Value)
			}
			return n, nil
		case "!!float":
			f, err := strconv.ParseFloat(node.Value, 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid number %q", node.Line, node.Value)
			}
			return f, nil
		case "!!bool":
			// True/TRUE keep their authored spelling under the !!bool tag;
			// Decode resolves them instead of silently yielding false.
			var b bool
			if err := node.Decode(&b); err != nil {
				return nil, fmt.Errorf("line %d: invalid boolean %q", node.Line, node.Value)
			}
			return b, nil
		case "!!null":
			return nil, nil
		default:
			return nil, fmt.Errorf("line %d: unsupported YAML tag %s (values must be JSON-representable)", node.Line, node.Tag)
		}
	case yaml.SequenceNode:
		out := make([]any, 0, len(node.Content))
		for _, child := range node.Content {
			v, err := cliDecodeValue(child)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.MappingNode:
		out := map[string]any{}
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			if keyNode.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("line %d: object keys must be scalar strings", keyNode.Line)
			}
			if keyNode.Tag != "!!str" {
				return nil, fmt.Errorf("line %d: object keys must be strings; quote key %q", keyNode.Line, keyNode.Value)
			}
			v, err := cliDecodeValue(node.Content[i+1])
			if err != nil {
				return nil, err
			}
			out[keyNode.Value] = v
		}
		return out, nil
	default:
		return nil, fmt.Errorf("line %d: unsupported YAML node kind in value", node.Line)
	}
}

// cliDidYouMean returns a ` (did you mean "x"?)` suffix when a close match for
// got exists in the candidate set, and "" otherwise.
func cliDidYouMean(got string, candidates []string) string {
	best := ""
	bestDist := 3 // suggest only within edit distance 2
	for _, candidate := range candidates {
		d := cliEditDistance(strings.ToLower(got), strings.ToLower(candidate))
		if d < bestDist {
			bestDist = d
			best = candidate
		}
	}
	if best == "" {
		return ""
	}
	return fmt.Sprintf(" (did you mean %q?)", best)
}

// cliCandidateList renders a deterministic, comma-separated candidate list.
func cliCandidateList(candidates []string) string {
	sorted := append([]string(nil), candidates...)
	sort.Strings(sorted)
	return strings.Join(sorted, ", ")
}

func cliEditDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, min(curr[j-1]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}

// cliPathSegment is one segment of a parsed singular JSONPath. IsWild is only
// produced by the artifact content-pointer grammar ([*] fan-out); the singular
// bind grammar rejects wildcards.
type cliPathSegment struct {
	Name    string
	Index   int
	IsIndex bool
	IsWild  bool
}

// cliParseSingularPath parses the overlay-style singular JSONPath subset used
// by input binds and preset keys: `$.name` dot children, `$['name']` /
// `$["name"]` bracket children (for names that are not identifiers), and
// `$[0]` numeric indexes. Wildcards, recursive descent, filters, slices, and
// unions are rejected with targeted errors — the grammar is deliberately the
// singular subset that identifies exactly one location.
func cliParseSingularPath(path string) ([]cliPathSegment, error) {
	if path == "" {
		return nil, stderrors.New("path is empty (expected a singular JSONPath such as $.field)")
	}
	if strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path %q uses JSON Pointer syntax; use the singular JSONPath spelling instead (e.g. $.%s)", path, strings.TrimPrefix(path, "/"))
	}
	if !strings.HasPrefix(path, "$") {
		return nil, fmt.Errorf("path %q must start with $ (expected a singular JSONPath such as $.field)", path)
	}
	rest := path[1:]
	if rest == "" {
		return nil, fmt.Errorf("path %q addresses the document root; bind a specific field (e.g. $.field)", path)
	}

	var segments []cliPathSegment
	for len(rest) > 0 {
		switch {
		case strings.HasPrefix(rest, ".."):
			return nil, fmt.Errorf("path %q uses recursive descent (..), which is not part of the singular path subset", path)
		case strings.HasPrefix(rest, "."):
			rest = rest[1:]
			if rest == "" {
				return nil, fmt.Errorf("path %q ends with a dangling dot", path)
			}
			if strings.HasPrefix(rest, "*") {
				return nil, fmt.Errorf("path %q uses a wildcard (*), which is not part of the singular path subset", path)
			}
			end := strings.IndexAny(rest, ".[")
			var name string
			if end == -1 {
				name, rest = rest, ""
			} else {
				name, rest = rest[:end], rest[end:]
			}
			if !cliIsPathIdentifier(name) {
				return nil, fmt.Errorf("path segment %q is not a plain identifier; use the bracket form $['%s']", name, name)
			}
			segments = append(segments, cliPathSegment{Name: name})
		case strings.HasPrefix(rest, "["):
			closing := cliFindBracketEnd(rest)
			if closing == -1 {
				return nil, fmt.Errorf("path %q has an unterminated bracket segment", path)
			}
			inner := rest[1:closing]
			rest = rest[closing+1:]
			trimmed := strings.TrimSpace(inner)
			switch {
			case trimmed == "*":
				return nil, fmt.Errorf("path %q uses a wildcard (*), which is not part of the singular path subset", path)
			case strings.HasPrefix(trimmed, "?"):
				return nil, fmt.Errorf("path %q uses a filter expression, which is not part of the singular path subset", path)
			case strings.Contains(trimmed, ":"):
				return nil, fmt.Errorf("path %q uses a slice, which is not part of the singular path subset", path)
			case strings.HasPrefix(trimmed, "'") || strings.HasPrefix(trimmed, "\""):
				quote := trimmed[0]
				closing := cliFindQuoteEnd(trimmed)
				if closing == -1 {
					return nil, fmt.Errorf("path %q has an unterminated quoted segment", path)
				}
				if closing != len(trimmed)-1 {
					if strings.Contains(trimmed[closing+1:], ",") {
						return nil, fmt.Errorf("path %q uses a union selector, which is not part of the singular path subset", path)
					}
					return nil, fmt.Errorf("path %q has trailing characters after the quoted segment", path)
				}
				name, err := cliUnescapeQuoted(trimmed[1:closing])
				if err != nil {
					return nil, fmt.Errorf("path %q: %w", path, err)
				}
				_ = quote
				segments = append(segments, cliPathSegment{Name: name})
			default:
				if idx, err := strconv.Atoi(trimmed); err == nil {
					if idx < 0 {
						return nil, fmt.Errorf("path %q uses a negative index, which is not part of the singular path subset", path)
					}
					segments = append(segments, cliPathSegment{Index: idx, IsIndex: true})
				} else if strings.Contains(trimmed, ",") {
					return nil, fmt.Errorf("path %q uses a union selector, which is not part of the singular path subset", path)
				} else {
					return nil, fmt.Errorf("path %q has an invalid bracket segment %q (use $['name'] or a numeric index)", path, trimmed)
				}
			}
		default:
			return nil, fmt.Errorf("path %q has unexpected trailing characters %q", path, rest)
		}
	}
	return segments, nil
}

// cliParseArtifactPath parses the artifact content-pointer grammar: `$.name`
// dot children, `$['name']` bracket children, and `[*]` wildcard fan-out over
// every array item. Numeric indexes, filters, slices, unions, and recursive
// descent are rejected with targeted errors — the pointer names where content
// items live, not a single item.
func cliParseArtifactPath(path string) ([]cliPathSegment, error) {
	return cliParseWildPath(path, cliWildPathGrammar{
		noun:         "pointer",
		subset:       "artifact pointer subset",
		emptyExample: "$.steps[*].content[*]",
		startExample: "$.steps[*].content[*]",
		rootHint:     "addresses the document root; name where the content items live (e.g. $.steps[*].content[*])",
		scanner:      "the artifact runtime",
	})
}

// cliWildPathGrammar parameterizes cliParseWildPath: the error-message
// wording that differs between the wildcard pointer dialects, plus the one
// grammar nicety (union detection after a quoted segment) only some dialects
// report. The scan itself is shared so grammar and escaping fixes land in one
// place.
type cliWildPathGrammar struct {
	noun              string // subject of every error message ("pointer", "reasonPointer")
	subset            string // name of the restricted subset in rejection messages
	emptyExample      string // example path(s) shown when the path is empty
	startExample      string // example path shown when the path does not start with $
	rootHint          string // full clause shown when the path addresses only the root
	scanner           string // subject of "... scans every item" in the numeric-index error
	detectQuotedUnion bool   // report `$['a','b']` as a union selector instead of trailing characters
}

// cliParseWildPath parses the shared wildcard pointer grammar: `$.name` dot
// children, `$['name']` bracket children, and `[*]` wildcard fan-out over
// every array item. Numeric indexes, filters, slices, unions, and recursive
// descent are rejected with targeted errors worded per the grammar config.
func cliParseWildPath(path string, g cliWildPathGrammar) ([]cliPathSegment, error) {
	if path == "" {
		return nil, fmt.Errorf("%s is empty (expected a restricted JSONPath such as %s)", g.noun, g.emptyExample)
	}
	if strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("%s %q uses JSON Pointer syntax; use the JSONPath spelling instead (e.g. $.%s)", g.noun, path, strings.TrimPrefix(path, "/"))
	}
	if !strings.HasPrefix(path, "$") {
		return nil, fmt.Errorf("%s %q must start with $ (expected a restricted JSONPath such as %s)", g.noun, path, g.startExample)
	}
	rest := path[1:]
	if rest == "" {
		return nil, fmt.Errorf("%s %q %s", g.noun, path, g.rootHint)
	}

	var segments []cliPathSegment
	for len(rest) > 0 {
		switch {
		case strings.HasPrefix(rest, ".."):
			return nil, fmt.Errorf("%s %q uses recursive descent (..), which is not part of the %s", g.noun, path, g.subset)
		case strings.HasPrefix(rest, "."):
			rest = rest[1:]
			if rest == "" {
				return nil, fmt.Errorf("%s %q ends with a dangling dot", g.noun, path)
			}
			if strings.HasPrefix(rest, "*") {
				return nil, fmt.Errorf("%s %q uses a dot wildcard (.*); arrays fan out with the bracket form [*]", g.noun, path)
			}
			end := strings.IndexAny(rest, ".[")
			var name string
			if end == -1 {
				name, rest = rest, ""
			} else {
				name, rest = rest[:end], rest[end:]
			}
			if !cliIsPathIdentifier(name) {
				return nil, fmt.Errorf("%s segment %q is not a plain identifier; use the bracket form $['%s']", g.noun, name, name)
			}
			segments = append(segments, cliPathSegment{Name: name})
		case strings.HasPrefix(rest, "["):
			closing := cliFindBracketEnd(rest)
			if closing == -1 {
				return nil, fmt.Errorf("%s %q has an unterminated bracket segment", g.noun, path)
			}
			inner := rest[1:closing]
			rest = rest[closing+1:]
			trimmed := strings.TrimSpace(inner)
			switch {
			case trimmed == "*":
				segments = append(segments, cliPathSegment{IsWild: true})
			case strings.HasPrefix(trimmed, "?"):
				return nil, fmt.Errorf("%s %q uses a filter expression, which is not part of the %s", g.noun, path, g.subset)
			case strings.Contains(trimmed, ":"):
				return nil, fmt.Errorf("%s %q uses a slice, which is not part of the %s", g.noun, path, g.subset)
			case strings.HasPrefix(trimmed, "'") || strings.HasPrefix(trimmed, "\""):
				closingQuote := cliFindQuoteEnd(trimmed)
				if closingQuote == -1 {
					return nil, fmt.Errorf("%s %q has an unterminated quoted segment", g.noun, path)
				}
				if closingQuote != len(trimmed)-1 {
					if g.detectQuotedUnion && strings.Contains(trimmed[closingQuote+1:], ",") {
						return nil, fmt.Errorf("%s %q uses a union selector, which is not part of the %s", g.noun, path, g.subset)
					}
					return nil, fmt.Errorf("%s %q has trailing characters after the quoted segment", g.noun, path)
				}
				name, err := cliUnescapeQuoted(trimmed[1:closingQuote])
				if err != nil {
					return nil, fmt.Errorf("%s %q: %w", g.noun, path, err)
				}
				segments = append(segments, cliPathSegment{Name: name})
			default:
				if _, err := strconv.Atoi(trimmed); err == nil {
					return nil, fmt.Errorf("%s %q selects a numeric index; %s scans every item, so spell the segment as [*]", g.noun, path, g.scanner)
				}
				return nil, fmt.Errorf("%s %q has an invalid bracket segment %q (use [*], $['name'], or a dot child)", g.noun, path, trimmed)
			}
		default:
			return nil, fmt.Errorf("%s %q has unexpected trailing characters %q", g.noun, path, rest)
		}
	}
	return segments, nil
}

// cliFindQuoteEnd locates the index of the closing quote of a quoted string
// starting at s[0], honoring backslash escapes. Returns -1 when unterminated.
func cliFindQuoteEnd(s string) int {
	quote := s[0]
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case quote:
			return i
		}
	}
	return -1
}

// cliFindBracketEnd locates the closing bracket of the segment starting at
// rest[0] == '[', honoring quoted strings with backslash escapes.
func cliFindBracketEnd(rest string) int {
	inQuote := byte(0)
	for i := 1; i < len(rest); i++ {
		c := rest[i]
		if inQuote != 0 {
			if c == '\\' {
				i++
				continue
			}
			if c == inQuote {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '\'', '"':
			inQuote = c
		case ']':
			return i
		}
	}
	return -1
}

func cliUnescapeQuoted(s string) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			if i >= len(s) {
				return "", stderrors.New("dangling escape in quoted segment")
			}
			switch s[i] {
			case '\'', '"', '\\':
				b.WriteByte(s[i])
			default:
				return "", fmt.Errorf("unsupported escape \\%c in quoted segment", s[i])
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String(), nil
}

func cliIsPathIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r == '_', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// cliSegmentsToPointer converts parsed path segments to an RFC 6901 JSON
// Pointer, the canonical form stored in the IR.
func cliSegmentsToPointer(segments []cliPathSegment) string {
	var b strings.Builder
	for _, seg := range segments {
		b.WriteByte('/')
		if seg.IsIndex {
			b.WriteString(strconv.Itoa(seg.Index))
			continue
		}
		escaped := strings.ReplaceAll(seg.Name, "~", "~0")
		escaped = strings.ReplaceAll(escaped, "/", "~1")
		b.WriteString(escaped)
	}
	return b.String()
}
