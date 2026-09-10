package extensions

import (
	"context"
	stderrors "errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"gopkg.in/yaml.v3"
)

// CLIErrorRule is one exact reason-code rule declared by
// x-speakeasy-cli-errors. Type and Hints are independently optional: an
// author may declare either part of a rule, or attach hints to a reason
// whose type still falls back to HTTP status.
type CLIErrorRule struct {
	Reason   string   `json:"reason" yaml:"reason"`
	Type     string   `json:"type,omitempty" yaml:"type,omitempty"`
	Hints    []string `json:"hints,omitempty" yaml:"hints,omitempty"`
	HasHints bool     `json:"-" yaml:"-"` // generation-only presence bit; not extension shape
	line     int      // declaration line for cross-carrier diagnostics
}

// CLIErrorTypeHints replaces the generated fallback hints for one closed
// error type.
type CLIErrorTypeHints struct {
	Type  string   `json:"type" yaml:"type"`
	Hints []string `json:"hints" yaml:"hints"`
}

// CLIErrorPointerSegment is one step of a parsed carrier pointer path: a
// named member, or a [*] fan-out over every item of an array member.
type CLIErrorPointerSegment struct {
	Field  string `json:"field,omitempty" yaml:"field,omitempty"`
	IsWild bool   `json:"isWild,omitempty" yaml:"isWild,omitempty"`
}

// CLIErrorProbe is one additional declared reason carrier: a pointer into
// the response body's error object plus the reason rules matched against the
// values found there. Probes are consulted after the primary carrier, in
// declaration order.
type CLIErrorProbe struct {
	Pointer  string                   `json:"pointer" yaml:"pointer"`
	Segments []CLIErrorPointerSegment `json:"-" yaml:"-"` // parsed Pointer; generation-only
	Reasons  []CLIErrorRule           `json:"reasons,omitempty" yaml:"reasons,omitempty"`
}

// CLIErrorManifest is the decoded x-speakeasy-cli-errors document extension.
// Slices preserve authoring order so generated Go tables are deterministic.
type CLIErrorManifest struct {
	Version int `json:"version" yaml:"version"`
	// ReasonPointer is the declared primary reason carrier: the restricted
	// JSONPath, relative to the response body's error object, whose string
	// values are the structured reason codes the top-level reason rules
	// match. Empty means the default carrier $.reason. Declaring the pointer
	// also opts the carrier into promoting an unmatched reason code verbatim
	// as error_reason.
	ReasonPointer  string                   `json:"reasonPointer,omitempty" yaml:"reasonPointer,omitempty"`
	ReasonSegments []CLIErrorPointerSegment `json:"-" yaml:"-"` // parsed ReasonPointer; generation-only
	Reasons        []CLIErrorRule           `json:"reasons,omitempty" yaml:"reasons,omitempty"`
	// Probes are additional reason carriers consulted after the primary
	// carrier, in declaration order. Each probe declares its pointer
	// explicitly, so each probe promotes an unmatched reason code verbatim.
	Probes []CLIErrorProbe     `json:"probes,omitempty" yaml:"probes,omitempty"`
	Types  []CLIErrorTypeHints `json:"types,omitempty" yaml:"types,omitempty"`
	// UnwrapErrorArray opts into normalizing the single-element
	// array-wrapped error body shape ([{"error": {...}}]) to its sole
	// element before classification. Default off.
	UnwrapErrorArray bool `json:"unwrapErrorArray,omitempty" yaml:"unwrapErrorArray,omitempty"`
}

// CLIErrorTypes is the closed error_type namespace emitted by generated CLIs.
var CLIErrorTypes = []string{
	"authentication_error",
	"authorization_error",
	"service_disabled",
	"billing_disabled",
	"not_found",
	"validation_error",
	"rate_limit_error",
	"server_error",
	"connection_error",
	"protocol_error",
	"api_error",
	"runtime_error",
	"unsupported_error",
	"async_failed",
	"async_timeout",
	"async_unknown_state",
}

var cliErrorTopLevelKeys = []string{"version", "reasonPointer", "reasons", "probes", "types", "unwrapErrorArray"}
var cliErrorRuleKeys = []string{"type", "hints"}
var cliErrorTypeRuleKeys = []string{"hints"}
var cliErrorProbeKeys = []string{"pointer", "reasons"}

// HandleCLIErrorsExtension decodes the document-level
// x-speakeasy-cli-errors extension. Returns nil when the extension is absent.
func (e *Extensions) HandleCLIErrorsExtension(_ context.Context, docInfo *document.DocumentInfo) (*CLIErrorManifest, error) {
	if docInfo == nil || docInfo.Doc == nil {
		return nil, nil
	}
	exts := docInfo.Doc.GetExtensions()
	if exts.Len() == 0 {
		return nil, nil
	}
	node, ok := e.findExtension(exts, ExtCLIErrors)
	if !ok {
		return nil, nil
	}

	manifest, err := DecodeCLIErrorsManifest(node)
	if err != nil {
		return nil, errors.NewValidationError(ExtCLIErrors.Name()+": "+err.Error(), node, nil)
	}
	return manifest, nil
}

// DecodeCLIErrorsManifest performs the strict v1 decode. It is exported for
// focused decoder tests; generation should call HandleCLIErrorsExtension.
func DecodeCLIErrorsManifest(node *yaml.Node) (*CLIErrorManifest, error) {
	expanded, err := cliExpandNode(node, &cliExpandState{})
	if err != nil {
		return nil, err
	}
	entries, err := cliMapEntries(expanded, "the extension value")
	if err != nil {
		return nil, err
	}

	manifest := &CLIErrorManifest{}
	var reasonsNode, typesNode, probesNode *yaml.Node
	versionSeen := false
	unwrapSeen := false
	for _, entry := range entries {
		switch entry.Key.Value {
		case "version":
			versionSeen = true
			if entry.Value.Kind != yaml.ScalarNode || entry.Value.Tag != "!!int" {
				return nil, fmt.Errorf("line %d: version must be the integer 1", entry.Value.Line)
			}
			if entry.Value.Value != "1" {
				return nil, fmt.Errorf("line %d: unsupported version %s (only version 1 exists)", entry.Value.Line, entry.Value.Value)
			}
			manifest.Version = 1
		case "reasonPointer":
			raw, err := cliScalarString(entry.Value, "reasonPointer")
			if err != nil {
				return nil, err
			}
			segments, err := cliErrorsParseCarrierPointer(raw, "reasonPointer")
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", entry.Value.Line, err)
			}
			manifest.ReasonPointer = raw
			manifest.ReasonSegments = segments
		case "reasons":
			reasonsNode = entry.Value
		case "probes":
			probesNode = entry.Value
		case "types":
			typesNode = entry.Value
		case "unwrapErrorArray":
			if entry.Value.Kind != yaml.ScalarNode || entry.Value.Tag != "!!bool" {
				return nil, fmt.Errorf("line %d: unwrapErrorArray must be true or false", entry.Value.Line)
			}
			unwrapSeen = true
			manifest.UnwrapErrorArray = entry.Value.Value == "true"
		default:
			return nil, fmt.Errorf("line %d: unknown key %q%s (expected version, reasonPointer, reasons, probes, types, unwrapErrorArray)", entry.Key.Line, entry.Key.Value, cliDidYouMean(entry.Key.Value, cliErrorTopLevelKeys))
		}
	}
	if !versionSeen {
		return nil, stderrors.New("version is required (the integer 1)")
	}

	if reasonsNode != nil {
		manifest.Reasons, err = decodeCLIErrorReasons(reasonsNode, "reasons", true)
		if err != nil {
			return nil, err
		}
	}
	if probesNode != nil {
		manifest.Probes, err = decodeCLIErrorProbes(probesNode, manifest)
		if err != nil {
			return nil, err
		}
	}
	if typesNode != nil {
		manifest.Types, err = decodeCLIErrorTypes(typesNode)
		if err != nil {
			return nil, err
		}
	}
	if len(manifest.Reasons) == 0 && len(manifest.Types) == 0 && manifest.ReasonPointer == "" && len(manifest.Probes) == 0 && !unwrapSeen {
		return nil, stderrors.New("at least one of reasonPointer, reasons, probes, types, or unwrapErrorArray must be declared")
	}
	return manifest, nil
}

// cliErrorsParseCarrierPointer parses the reason-carrier grammar: the shared
// wildcard pointer subset from cliParseWildPath (`$.name` dot children,
// `$['name']` bracket children, and `[*]` wildcard fan-out over every array
// item) with carrier-specific error wording. The pointer is resolved against
// the response body's error object, so it must additionally begin with a
// named member and end at the named member that holds the reason string.
func cliErrorsParseCarrierPointer(path, noun string) ([]CLIErrorPointerSegment, error) {
	parsed, err := cliParseWildPath(path, cliWildPathGrammar{
		noun:              noun,
		subset:            "reason carrier subset",
		emptyExample:      "$.reason or $.details[*].reason",
		startExample:      "$.reason",
		rootHint:          "addresses the error object root; name the member that holds the reason string (e.g. $.reason)",
		scanner:           "the classifier",
		detectQuotedUnion: true,
	})
	if err != nil {
		return nil, err
	}
	segments := make([]CLIErrorPointerSegment, len(parsed))
	for i, segment := range parsed {
		segments[i] = CLIErrorPointerSegment{Field: segment.Name, IsWild: segment.IsWild}
	}
	if segments[0].IsWild {
		return nil, fmt.Errorf("%s %q begins with a [*] fan-out, but it is resolved against the error object; begin with a named member (e.g. $.details[*].reason)", noun, path)
	}
	if segments[len(segments)-1].IsWild {
		return nil, fmt.Errorf("%s %q ends with a [*] fan-out; it must end at the named member that holds the reason string (e.g. $.details[*].reason)", noun, path)
	}
	return segments, nil
}

// cliErrorCarrierKey canonicalizes parsed pointer segments so equivalent
// spellings of the same carrier ($.reason vs $['reason']) collide. Each named
// segment is quoted so the encoding stays injective: a quoted field containing
// a path delimiter ($['a.b']) must not collapse into its nested spelling
// ($.a.b), nor a field literally containing "[*]" into a wildcard fan-out.
func cliErrorCarrierKey(segments []CLIErrorPointerSegment) string {
	var b strings.Builder
	for _, segment := range segments {
		if segment.IsWild {
			b.WriteString("[*]")
			continue
		}
		b.WriteString(".")
		b.WriteString(strconv.Quote(segment.Field))
	}
	return b.String()
}

// cliErrorPrimarySegments is the primary carrier's parsed pointer: the
// declared reasonPointer, else the default $.reason member.
func cliErrorPrimarySegments(manifest *CLIErrorManifest) []CLIErrorPointerSegment {
	if len(manifest.ReasonSegments) > 0 {
		return manifest.ReasonSegments
	}
	return []CLIErrorPointerSegment{{Field: "reason"}}
}

func decodeCLIErrorProbes(node *yaml.Node, manifest *CLIErrorManifest) ([]CLIErrorProbe, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("line %d: probes must be a sequence of probe declarations", node.Line)
	}
	if len(node.Content) == 0 {
		return nil, fmt.Errorf("line %d: probes must contain at least one probe declaration", node.Line)
	}

	primaryKey := cliErrorCarrierKey(cliErrorPrimarySegments(manifest))
	seenKeys := map[string]int{}
	// Reason values already bound to a type at an earlier carrier; a reason
	// value must map to a single error type wherever it is found.
	boundTypes := map[string]string{}
	boundWhere := map[string]string{}
	for _, rule := range manifest.Reasons {
		if rule.Type != "" {
			boundTypes[rule.Reason] = rule.Type
			boundWhere[rule.Reason] = "the top-level reasons"
		}
	}

	probes := make([]CLIErrorProbe, 0, len(node.Content))
	for i, item := range node.Content {
		what := fmt.Sprintf("probes[%d]", i)
		entries, err := cliMapEntries(item, what)
		if err != nil {
			return nil, err
		}
		probe := CLIErrorProbe{}
		var probeReasonsNode *yaml.Node
		for _, entry := range entries {
			switch entry.Key.Value {
			case "pointer":
				raw, err := cliScalarString(entry.Value, what+" pointer")
				if err != nil {
					return nil, err
				}
				segments, err := cliErrorsParseCarrierPointer(raw, "pointer")
				if err != nil {
					return nil, fmt.Errorf("line %d: %s: %w", entry.Value.Line, what, err)
				}
				probe.Pointer = raw
				probe.Segments = segments
			case "reasons":
				probeReasonsNode = entry.Value
			default:
				return nil, fmt.Errorf("line %d: %s has unknown key %q%s (expected pointer, reasons)", entry.Key.Line, what, entry.Key.Value, cliDidYouMean(entry.Key.Value, cliErrorProbeKeys))
			}
		}
		if probe.Pointer == "" {
			return nil, fmt.Errorf("line %d: %s must declare pointer (the reason carrier it classifies from)", item.Line, what)
		}
		key := cliErrorCarrierKey(probe.Segments)
		if key == primaryKey {
			if manifest.ReasonPointer != "" {
				return nil, fmt.Errorf("line %d: %s pointer %q duplicates the declared reasonPointer; declare those rules under the top-level reasons", item.Line, what, probe.Pointer)
			}
			return nil, fmt.Errorf("line %d: %s pointer %q duplicates the default primary carrier $.reason; declare reasonPointer and the top-level reasons for that carrier instead", item.Line, what, probe.Pointer)
		}
		if earlier, dup := seenKeys[key]; dup {
			return nil, fmt.Errorf("line %d: %s pointer %q duplicates probes[%d]; declare one probe per carrier", item.Line, what, probe.Pointer, earlier)
		}
		seenKeys[key] = i

		if probeReasonsNode != nil {
			probe.Reasons, err = decodeCLIErrorReasons(probeReasonsNode, what+" reasons", false)
			if err != nil {
				return nil, err
			}
		}
		for _, rule := range probe.Reasons {
			if rule.Type == "" {
				continue
			}
			if bound, ok := boundTypes[rule.Reason]; ok && bound != rule.Type {
				return nil, fmt.Errorf("line %d: %s reason %q declares type %q, but the same reason is declared with type %q at %s; a reason value must map to a single error type across carriers", rule.line, what, rule.Reason, rule.Type, bound, boundWhere[rule.Reason])
			}
			if _, ok := boundTypes[rule.Reason]; !ok {
				boundTypes[rule.Reason] = rule.Type
				boundWhere[rule.Reason] = what
			}
		}
		probes = append(probes, probe)
	}
	return probes, nil
}

func decodeCLIErrorReasons(node *yaml.Node, what string, allowRuntime bool) ([]CLIErrorRule, error) {
	entries, err := cliMapEntries(node, what)
	if err != nil {
		return nil, err
	}
	rules := make([]CLIErrorRule, 0, len(entries))
	for _, entry := range entries {
		reason, err := cliErrorExactKey(entry.Key, "reason")
		if err != nil {
			return nil, err
		}
		if !allowRuntime && strings.HasPrefix(reason, "CLI_") {
			return nil, fmt.Errorf("line %d: %s reason %q is runtime-owned; declare its hints under the top-level reasons, not inside a probe", entry.Key.Line, what, reason)
		}
		ruleEntries, err := cliMapEntries(entry.Value, fmt.Sprintf("reason %q", reason))
		if err != nil {
			return nil, err
		}
		rule := CLIErrorRule{Reason: reason, line: entry.Key.Line}
		for _, ruleEntry := range ruleEntries {
			switch ruleEntry.Key.Value {
			case "type":
				if strings.HasPrefix(reason, "CLI_") {
					return nil, fmt.Errorf("line %d: reason %q is runtime-owned and may declare hints only; its type is fixed", ruleEntry.Key.Line, reason)
				}
				rule.Type, err = cliScalarString(ruleEntry.Value, "error type")
				if err != nil {
					return nil, err
				}
				if !cliContains(CLIErrorTypes, rule.Type) {
					return nil, fmt.Errorf("line %d: reason %q has unknown type %q%s (expected %s)", ruleEntry.Value.Line, reason, rule.Type, cliDidYouMean(rule.Type, CLIErrorTypes), cliCandidateList(CLIErrorTypes))
				}
			case "hints":
				rule.Hints, err = decodeCLIErrorHints(ruleEntry.Value, fmt.Sprintf("reason %q hints", reason))
				if err != nil {
					return nil, err
				}
				rule.HasHints = true
			default:
				return nil, fmt.Errorf("line %d: reason %q has unknown key %q%s", ruleEntry.Key.Line, reason, ruleEntry.Key.Value, cliDidYouMean(ruleEntry.Key.Value, cliErrorRuleKeys))
			}
		}
		if strings.HasPrefix(reason, "CLI_") {
			if !cliContains(CLIRuntimeHintReasons, reason) {
				return nil, fmt.Errorf("line %d: reason %q is not a reason the generated runtime emits (CLI_* namespace: %s)", entry.Key.Line, reason, cliCandidateList(CLIRuntimeHintReasons))
			}
			if len(rule.Hints) == 0 {
				return nil, fmt.Errorf("line %d: runtime reason %q must declare hints", entry.Value.Line, reason)
			}
		} else if rule.Type == "" && len(rule.Hints) == 0 {
			return nil, fmt.Errorf("line %d: reason %q must declare type, hints, or both", entry.Value.Line, reason)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func decodeCLIErrorTypes(node *yaml.Node) ([]CLIErrorTypeHints, error) {
	entries, err := cliMapEntries(node, "types")
	if err != nil {
		return nil, err
	}
	rules := make([]CLIErrorTypeHints, 0, len(entries))
	for _, entry := range entries {
		errorType, err := cliErrorExactKey(entry.Key, "type")
		if err != nil {
			return nil, err
		}
		if !cliContains(CLIErrorTypes, errorType) {
			return nil, fmt.Errorf("line %d: unknown error type %q%s (expected %s)", entry.Key.Line, errorType, cliDidYouMean(errorType, CLIErrorTypes), cliCandidateList(CLIErrorTypes))
		}
		typeEntries, err := cliMapEntries(entry.Value, fmt.Sprintf("type %q", errorType))
		if err != nil {
			return nil, err
		}
		rule := CLIErrorTypeHints{Type: errorType}
		for _, typeEntry := range typeEntries {
			switch typeEntry.Key.Value {
			case "hints":
				rule.Hints, err = decodeCLIErrorHints(typeEntry.Value, fmt.Sprintf("type %q hints", errorType))
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("line %d: type %q has unknown key %q%s", typeEntry.Key.Line, errorType, typeEntry.Key.Value, cliDidYouMean(typeEntry.Key.Value, cliErrorTypeRuleKeys))
			}
		}
		if len(rule.Hints) == 0 {
			return nil, fmt.Errorf("line %d: type %q must declare hints", entry.Value.Line, errorType)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func decodeCLIErrorHints(node *yaml.Node, context string) ([]string, error) {
	var hints []string
	switch node.Kind {
	case yaml.ScalarNode:
		hint, err := cliScalarString(node, context)
		if err != nil {
			return nil, err
		}
		hints = []string{hint}
	case yaml.SequenceNode:
		for _, item := range node.Content {
			hint, err := cliScalarString(item, context)
			if err != nil {
				return nil, err
			}
			hints = append(hints, hint)
		}
	default:
		return nil, fmt.Errorf("line %d: %s must be a string or a sequence of strings", node.Line, context)
	}
	if len(hints) == 0 {
		return nil, fmt.Errorf("line %d: %s must contain at least one line", node.Line, context)
	}
	for _, hint := range hints {
		if err := cliErrorValidateLine(hint); err != nil {
			return nil, fmt.Errorf("line %d: %s %w", node.Line, context, err)
		}
	}
	return hints, nil
}

func cliErrorExactKey(node *yaml.Node, kind string) (string, error) {
	value, err := cliScalarString(node, kind)
	if err != nil {
		return "", err
	}
	if err := cliErrorValidateLine(value); err != nil {
		return "", fmt.Errorf("line %d: %s %w", node.Line, kind, err)
	}
	return value, nil
}

func cliErrorValidateLine(value string) error {
	if value == "" {
		return stderrors.New("must be non-empty")
	}
	if strings.TrimSpace(value) != value {
		return stderrors.New("must not have leading or trailing whitespace")
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return stderrors.New("must not contain control characters")
		}
	}
	return nil
}
