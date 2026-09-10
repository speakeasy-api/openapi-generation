package extensions

// This file is the generator-side rendering layer over the schemaexec
// symbolic executor (github.com/speakeasy-api/jq): it turns verdicts and
// machine-level causes into one-line, developer-facing diagnostics that name
// the owner, the projection, what is wrong, and — when cheaply derivable
// from the response schema — how to fix it.

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	runtimejq "github.com/itchyny/gojq"
	analyzerjq "github.com/speakeasy-api/jq"
	"github.com/speakeasy-api/jq/schemaexec"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
)

func cliOperationProjectionSchema(opInfo *cliOperationInfo) *oas3.JSONSchema[oas3.Referenceable] {
	if opInfo == nil {
		return nil
	}
	if !opInfo.streamMediaSeen {
		return opInfo.responseSchemaJS
	}

	branches := make([]*oas3.JSONSchema[oas3.Referenceable], 0, len(opInfo.streamSchemaJS))
	for _, schema := range opInfo.streamSchemaJS {
		if schema == nil || schema.GetSchema() == nil {
			// A streaming media type without a schema admits any event shape.
			schema = oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{})
		}
		branches = append(branches, schema)
	}
	return oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{AnyOf: branches})
}

// lintProjection is the linker seam shared by every declared projection
// surface. Future projection kinds should lower to jq and enter here.
func (d *cliManifestDecoder) lintProjection(owner, kind, jqExpr, unavailableReason string, schemaJS *oas3.JSONSchema[oas3.Referenceable], suppressWarning bool) error {
	query, diagnostic := cliParseProjection(owner, kind, jqExpr)
	if diagnostic == nil && query != nil {
		switch schemaJS {
		case nil:
			diagnostic = newCLIProjectionDiagnostic(owner, kind, jqExpr, "warning", "cannot be statically verified — %s; it will only be checked at runtime", unavailableReason)
		default:
			response, err := cliResolveProjectionSchema(d.ctx, d.docInfo.GetResolutionOptions(d.ctx), schemaJS)
			if err != nil {
				diagnostic = newCLIProjectionDiagnostic(owner, kind, jqExpr, "warning", "cannot be statically verified — response schema references could not be resolved: %s; it will only be checked at runtime", cliFirstSentence(err.Error()))
			} else {
				diagnostic = cliAnalyzeProjection(d.ctx, owner, kind, jqExpr, query, response)
			}
		}
	}
	if diagnostic == nil {
		return nil
	}
	if diagnostic.Severity == "error" {
		return errors.New(diagnostic.Message)
	}
	if !suppressWarning {
		d.warnf("%s", diagnostic.Message)
	}
	return nil
}

func cliResolveProjectionSchema(ctx context.Context, opts oas3.ResolveOptions, root *oas3.JSONSchema[oas3.Referenceable]) (*oas3.Schema, error) {
	seenWrappers := map[*oas3.JSONSchema[oas3.Referenceable]]bool{}
	seenSchemas := map[*oas3.Schema]bool{}

	var walkSchema func(*oas3.Schema) error
	walkWrapper := func(js *oas3.JSONSchema[oas3.Referenceable]) error {
		if js == nil || seenWrappers[js] {
			return nil
		}
		seenWrappers[js] = true

		schema := js.GetSchema()
		if js.IsReference() {
			if _, err := js.Resolve(ctx, opts); err != nil {
				return fmt.Errorf("resolve %q: %w", js.GetRef(), err)
			}
			resolved := js.GetResolvedSchema()
			if resolved == nil || resolved.GetSchema() == nil {
				return fmt.Errorf("resolve %q: unresolved reference", js.GetRef())
			}
			if !schema.IsReferenceOnly() {
				if err := walkSchema(schema); err != nil {
					return err
				}
			}
			schema = resolved.GetSchema()
		}
		return walkSchema(schema)
	}
	walkSchema = func(schema *oas3.Schema) error {
		if schema == nil || seenSchemas[schema] {
			return nil
		}
		seenSchemas[schema] = true

		for _, group := range [][]*oas3.JSONSchema[oas3.Referenceable]{
			schema.GetAllOf(), schema.GetAnyOf(), schema.GetOneOf(), schema.GetPrefixItems(),
		} {
			for _, child := range group {
				if err := walkWrapper(child); err != nil {
					return err
				}
			}
		}
		for _, child := range []*oas3.JSONSchema[oas3.Referenceable]{
			schema.GetItems(), schema.GetAdditionalProperties(), schema.GetNot(),
			schema.GetIf(), schema.GetThen(), schema.GetElse(), schema.GetContains(),
			schema.GetUnevaluatedItems(), schema.GetUnevaluatedProperties(),
			schema.GetPropertyNames(), schema.GetContentSchema(),
		} {
			if err := walkWrapper(child); err != nil {
				return err
			}
		}
		if properties := schema.GetProperties(); properties != nil {
			for _, child := range properties.All() {
				if err := walkWrapper(child); err != nil {
					return err
				}
			}
		}
		if patternProperties := schema.GetPatternProperties(); patternProperties != nil {
			for _, child := range patternProperties.All() {
				if err := walkWrapper(child); err != nil {
					return err
				}
			}
		}
		if defs := schema.GetDefs(); defs != nil {
			for _, child := range defs.All() {
				if err := walkWrapper(child); err != nil {
					return err
				}
			}
		}
		if dependentSchemas := schema.GetDependentSchemas(); dependentSchemas != nil {
			for _, child := range dependentSchemas.All() {
				if err := walkWrapper(child); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := walkWrapper(root); err != nil {
		return nil, err
	}
	return cliMaterializeProjectionSchema(ctx, root), nil
}

// cliMaterializeProjectionSchema builds a fresh schema graph for the
// executor. In JSON Schema and OpenAPI 3.1, keywords beside $ref constrain
// the referenced target; the typed wrapper's GetResolvedSchema intentionally
// returns only the target. Representing the effective schema as
// allOf: [target, siblings] preserves both without mutating the document.
func cliMaterializeProjectionSchema(ctx context.Context, root *oas3.JSONSchema[oas3.Referenceable]) *oas3.Schema {
	memo := map[*oas3.Schema]*oas3.Schema{}

	var materializeSchema func(*oas3.Schema) *oas3.Schema
	materializeWrapper := func(js *oas3.JSONSchema[oas3.Referenceable]) *oas3.JSONSchema[oas3.Referenceable] {
		if js == nil {
			return nil
		}
		if value := js.GetBool(); value != nil {
			return oas3.NewJSONSchemaFromBool(*value)
		}
		source := js.GetSchema()
		if source == nil {
			return nil
		}
		if !js.IsReference() {
			return oas3.NewJSONSchemaFromSchema[oas3.Referenceable](materializeSchema(source))
		}

		resolved := js.GetResolvedSchema()
		if resolved == nil || resolved.GetSchema() == nil {
			return oas3.NewJSONSchemaFromSchema[oas3.Referenceable](materializeSchema(source))
		}
		target := materializeSchema(resolved.GetSchema())
		targetRef := oas3.NewReferencedScheme(ctx, js.GetRef(), oas3.NewJSONSchemaFromSchema[oas3.Concrete](target))
		if source.IsReferenceOnly() {
			return targetRef
		}

		siblings := materializeSchema(source).ShallowCopy()
		siblings.Ref = nil
		return oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{AllOf: []*oas3.JSONSchema[oas3.Referenceable]{
			targetRef,
			oas3.NewJSONSchemaFromSchema[oas3.Referenceable](siblings),
		}})
	}
	materializeSchema = func(source *oas3.Schema) *oas3.Schema {
		if source == nil {
			return nil
		}
		if clone, ok := memo[source]; ok {
			return clone
		}
		clone := source.ShallowCopy()
		memo[source] = clone

		materializeSlice := func(children []*oas3.JSONSchema[oas3.Referenceable]) []*oas3.JSONSchema[oas3.Referenceable] {
			if children == nil {
				return nil
			}
			result := make([]*oas3.JSONSchema[oas3.Referenceable], len(children))
			for i, child := range children {
				result[i] = materializeWrapper(child)
			}
			return result
		}
		clone.AllOf = materializeSlice(source.AllOf)
		clone.AnyOf = materializeSlice(source.AnyOf)
		clone.OneOf = materializeSlice(source.OneOf)
		clone.PrefixItems = materializeSlice(source.PrefixItems)

		clone.Items = materializeWrapper(source.Items)
		clone.AdditionalProperties = materializeWrapper(source.AdditionalProperties)
		clone.UnevaluatedProperties = materializeWrapper(source.UnevaluatedProperties)
		clone.UnevaluatedItems = materializeWrapper(source.UnevaluatedItems)
		clone.Contains = materializeWrapper(source.Contains)
		clone.Not = materializeWrapper(source.Not)
		clone.If = materializeWrapper(source.If)
		clone.Then = materializeWrapper(source.Then)
		clone.Else = materializeWrapper(source.Else)
		clone.PropertyNames = materializeWrapper(source.PropertyNames)
		clone.ContentSchema = materializeWrapper(source.ContentSchema)

		for name, child := range source.Properties.All() {
			clone.Properties.Set(name, materializeWrapper(child))
		}
		for pattern, child := range source.PatternProperties.All() {
			clone.PatternProperties.Set(pattern, materializeWrapper(child))
		}
		for name, child := range source.DependentSchemas.All() {
			clone.DependentSchemas.Set(name, materializeWrapper(child))
		}
		for name, child := range source.Defs.All() {
			clone.Defs.Set(name, materializeWrapper(child))
		}
		return clone
	}

	return cliResolveJS(materializeWrapper(root))
}

// cliProjectionDiagnostic is one rendered finding about a declared jq
// projection. Severity "error" is safe to fail generation on (the projection
// provably yields nothing useful); "warning" means the projection could not
// be verified and will only be checked at runtime.
type cliProjectionDiagnostic struct {
	Owner    string
	Kind     string
	JQ       string
	Severity string // "error" | "warning"
	Message  string
}

func newCLIProjectionDiagnostic(owner, kind, jqExpr, severity, format string, args ...any) *cliProjectionDiagnostic {
	return &cliProjectionDiagnostic{
		Owner:    owner,
		Kind:     kind,
		JQ:       jqExpr,
		Severity: severity,
		Message:  fmt.Sprintf("%s: %s projection %q %s", owner, kind, jqExpr, fmt.Sprintf(format, args...)),
	}
}

func cliParseProjection(owner, kind, jqExpr string) (*analyzerjq.Query, *cliProjectionDiagnostic) {
	if jqExpr == "" {
		return nil, nil
	}
	if _, err := runtimejq.Parse(jqExpr); err != nil {
		return nil, newCLIProjectionDiagnostic(owner, kind, jqExpr, "error", "is not valid jq: %s", cliCondenseJQParseError(err))
	}
	query, err := analyzerjq.Parse(jqExpr)
	if err != nil {
		return nil, newCLIProjectionDiagnostic(owner, kind, jqExpr, "warning", "cannot be statically verified — the analyzer does not support this syntax; it will only be checked at runtime")
	}
	return query, nil
}

func cliAnalyzeProjection(ctx context.Context, owner, kind, jqExpr string, query *analyzerjq.Query, response *oas3.Schema) *cliProjectionDiagnostic {
	if query == nil || response == nil {
		return nil
	}
	errorDiag := func(format string, args ...any) *cliProjectionDiagnostic {
		return newCLIProjectionDiagnostic(owner, kind, jqExpr, "error", format, args...)
	}
	warnDiag := func(format string, args ...any) *cliProjectionDiagnostic {
		return newCLIProjectionDiagnostic(owner, kind, jqExpr, "warning", format, args...)
	}

	// Schema-walk of the projection's leading path. When the whole expression
	// is a simple path, the walk is authoritative: a mistake it finds is a
	// hard error with the precise reason (typos with suggestions, field
	// access on arrays, iterating scalars). For longer expressions it only
	// explains verdicts the engine reaches.
	walk := cliWalkProjectionPath(jqExpr, response)
	if walk.Mistake != "" && walk.Authoritative {
		return errorDiag("always returns nothing — %s", walk.Mistake)
	}

	analysis, err := schemaexec.Analyze(ctx, query, response)
	if err != nil {
		return warnDiag("cannot be statically verified — could not be analyzed: %s; it will only be checked at runtime", cliFirstSentence(err.Error()))
	}

	switch analysis.Verdict {
	case schemaexec.VerdictProvenBroken:
		if walk.PropertyUncertain {
			return warnDiag("cannot be statically verified — the response schema may permit undeclared properties along this path; it will only be checked at runtime")
		}
		if walk.Mistake != "" {
			return errorDiag("always returns nothing — %s", walk.Mistake)
		}
		return errorDiag("always returns nothing for valid responses — %s", cliCondenseCauses(analysis.Causes))
	case schemaexec.VerdictUnverifiable:
		return warnDiag("cannot be statically verified — %s; it will only be checked at runtime", cliUnverifiableCause(analysis.Causes))
	}

	// The executor is best-effort. It currently proves three simple patterns
	// that are broken at runtime: field access directly on an array, map over a
	// missing item field, and a string builtin applied to a non-string. The
	// leading-path renderer catches their simple forms below; longer forms are
	// intentionally left silent until schemaexec can classify them itself.
	if cliSchemaIsNullOnly(analysis.Output) {
		if walk.PropertyUncertain {
			return warnDiag("cannot be statically verified — the response schema may permit undeclared properties along this path; it will only be checked at runtime")
		}
		if walk.Mistake != "" {
			return errorDiag("always returns null — %s", walk.Mistake)
		}
		return errorDiag("always returns null for valid responses — a projected field does not exist in the response schema")
	}
	if item := cliArrayItemSchema(analysis.Output); item != nil && cliSchemaIsNullOnly(item) {
		if field, arrayPath, suggestion, definitelyMissing := cliExplainMapProjection(jqExpr, response); field != "" {
			if definitelyMissing {
				return errorDiag("always returns an array of nulls — %q does not exist in the items of %q%s", "."+field, arrayPath, suggestion)
			}
			return nil
		}
		return errorDiag("always returns an array of nulls — the projected field does not exist in the array's items")
	}
	if mistake := cliStringBuiltinMismatch(jqExpr, walk); mistake != "" {
		return errorDiag("%s", mistake)
	}

	return nil
}

// cliUnverifiableCause preserves the executor's concrete location. A leading
// path may encounter a union even when that union is not what caused the
// loss of precision, so the walker's UnionAt is not used to replace this.
func cliUnverifiableCause(causes []string) string {
	if len(causes) == 0 {
		return "the output type could not be determined"
	}
	return cliFirstSentence(causes[0])
}

// cliCondenseJQParseError reduces a gojq parse error to its first line.
func cliCondenseJQParseError(err error) string {
	msg := err.Error()
	if idx := strings.IndexByte(msg, '\n'); idx != -1 {
		msg = msg[:idx]
	}
	return strings.TrimSpace(msg)
}

// cliCondenseCauses folds machine-level schemaexec causes into one
// developer-facing phrase. The cause strings themselves are substrate and
// pass through only when nothing better is known.
func cliCondenseCauses(causes []string) string {
	joined := strings.ToLower(strings.Join(causes, " | "))
	switch {
	case strings.Contains(joined, "oneof") || strings.Contains(joined, "anyof") || strings.Contains(joined, "union"):
		return "the response schema is a union along this path"
	case strings.Contains(joined, "additionalproperties") || strings.Contains(joined, "unconstrained") || strings.Contains(joined, "untyped"):
		return "the response schema is open (accepts undeclared values) along this path"
	case strings.Contains(joined, "unsupported") || strings.Contains(joined, "not implemented"):
		return "the projection uses a jq feature that cannot be statically analyzed"
	case len(causes) > 0:
		return cliFirstSentence(causes[0])
	default:
		return "the output type could not be determined"
	}
}

// --- Leading-path analysis -------------------------------------------------

type cliJQPathSegment struct {
	Field   string
	Iterate bool
	Index   bool
}

// cliProjectionWalk is the outcome of walking a projection's leading simple
// path against the response schema.
type cliProjectionWalk struct {
	// Mistake is a precise, human-phrased reason when the walk itself
	// located the problem ("" when the walked prefix resolves).
	Mistake string
	// Authoritative reports that the WHOLE expression was a simple path, so
	// a mistake needs no engine confirmation (later jq stages could not
	// rescue it, and there is no `?` or `//` to absorb it).
	Authoritative bool
	// UnionAt is the walked path at which a oneOf/anyOf stopped precise
	// analysis ("" when no union was met).
	UnionAt string
	// FinalTypes are the schema types at the end of the walked prefix, when
	// the walk completed cleanly.
	FinalTypes []oas3.SchemaType
	// PropertyUncertain records that a field was not explicitly declared but
	// an open/pattern/unevaluated property rule prevents proving it absent.
	PropertyUncertain bool
	// Rest is the unconsumed remainder of the expression.
	Rest string
}

// cliLeadingJQPath extracts the longest simple-path prefix of a jq
// expression (".a.b[].c" of ".a.b[].c | length") plus the remainder.
func cliLeadingJQPath(expr string) ([]cliJQPathSegment, string) {
	var segments []cliJQPathSegment
	rest := strings.TrimSpace(expr)
	for rest != "" {
		switch {
		case strings.HasPrefix(rest, "."):
			body := rest[1:]
			end := 0
			for end < len(body) {
				c := body[end]
				if c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (end > 0 && c >= '0' && c <= '9') {
					end++
					continue
				}
				break
			}
			if end == 0 {
				return segments, rest
			}
			segments = append(segments, cliJQPathSegment{Field: body[:end]})
			rest = body[end:]
		case strings.HasPrefix(rest, "[]"):
			segments = append(segments, cliJQPathSegment{Iterate: true})
			rest = rest[2:]
		case strings.HasPrefix(rest, "["):
			closing := strings.IndexByte(rest, ']')
			if closing == -1 {
				return segments, rest
			}
			if _, err := strconv.Atoi(rest[1:closing]); err != nil {
				return segments, rest
			}
			segments = append(segments, cliJQPathSegment{Index: true})
			rest = rest[closing+1:]
		default:
			return segments, rest
		}
	}
	return segments, ""
}

// cliWalkProjectionPath walks the leading path against the response schema.
func cliWalkProjectionPath(expr string, response *oas3.Schema) cliProjectionWalk {
	segments, rest := cliLeadingJQPath(expr)
	walk := cliProjectionWalk{Rest: strings.TrimSpace(rest), Authoritative: strings.TrimSpace(rest) == ""}
	if len(segments) == 0 {
		walk.Authoritative = false
		return walk
	}

	current := response
	walked := ""
	for _, segment := range segments {
		if current == nil {
			return walk
		}
		if len(current.OneOf) > 0 || len(current.AnyOf) > 0 {
			walk.UnionAt = cliOrResponsePlain(walked)
			return walk
		}
		types := cliProjectionSchemaTypes(current, nil)

		switch {
		case segment.Field != "":
			admitsObject := cliSchemaIsType(types, oas3.SchemaTypeObject)
			admitsArray := cliSchemaIsType(types, oas3.SchemaTypeArray)
			if len(types) == 1 && admitsArray {
				iterSuggestion := fmt.Sprintf("%s[].%s", walked, segment.Field)
				mapSuggestion := fmt.Sprintf("%s | map(.%s)", walked, segment.Field)
				if walked == "" {
					iterSuggestion = ".[]." + segment.Field
					mapSuggestion = "map(." + segment.Field + ")"
				}
				walk.Mistake = fmt.Sprintf("%s is an array; use %q or %q", cliOrResponse(walked), iterSuggestion, mapSuggestion)
				return walk
			}
			if len(types) > 0 && !admitsObject {
				if len(types) == 1 {
					walk.Mistake = fmt.Sprintf("%s is %s, not an object", cliOrResponse(walked), cliTypeWithArticle(types[0]))
				} else {
					walk.Mistake = cliOrResponse(walked) + " does not admit object values"
				}
				return walk
			}
			property, ok := cliProjectionProperty(current, segment.Field, nil)
			if !ok {
				if cliProjectionPropertyDefinitelyMissing(current, segment.Field, nil) {
					walk.Mistake = fmt.Sprintf("%q does not exist in %s%s", "."+segment.Field, cliOrResponse(walked), cliJQDidYouMean(segment.Field, cliSchemaPropertyNames(current)))
				} else {
					walk.PropertyUncertain = true
				}
				return walk
			}
			current = property
			walked += "." + segment.Field
		case segment.Iterate || segment.Index:
			admitsArray := cliSchemaIsType(types, oas3.SchemaTypeArray)
			admitsObject := cliSchemaIsType(types, oas3.SchemaTypeObject)
			if len(types) > 0 && !admitsArray && !admitsObject {
				walk.Mistake = fmt.Sprintf("%s is %s; [] iterates arrays and objects", cliOrResponse(walked), cliTypeWithArticle(types[0]))
				return walk
			}
			if len(types) == 1 && admitsArray {
				if current.Items == nil {
					return walk
				}
				current = cliResolveJS(current.Items)
				if segment.Iterate {
					walked += "[]"
				} else {
					walked += "[0]"
				}
				continue
			}
			return walk // object iteration and unknowns: engine territory
		}
	}
	walk.FinalTypes = nil
	if current != nil {
		walk.FinalTypes = cliProjectionSchemaTypes(current, nil)
	}
	return walk
}

// cliProjectionProperty looks through conjunctions because an allOf branch
// (including a materialized $ref sibling) may declare the selected field. If
// several branches constrain it, their property schemas are conjoined too.
func cliProjectionProperty(schema *oas3.Schema, name string, seen map[*oas3.Schema]bool) (*oas3.Schema, bool) {
	if schema == nil {
		return nil, false
	}
	if seen == nil {
		seen = map[*oas3.Schema]bool{}
	}
	if seen[schema] {
		return nil, false
	}
	seen[schema] = true
	defer delete(seen, schema)

	var matches []*oas3.Schema
	if property, ok := schema.Properties.Get(name); ok {
		if resolved := cliResolveJS(property); resolved != nil {
			matches = append(matches, resolved)
		}
	}
	for _, branch := range schema.AllOf {
		if property, ok := cliProjectionProperty(cliResolveJS(branch), name, seen); ok {
			matches = append(matches, property)
		}
	}
	if len(matches) == 0 {
		return nil, false
	}
	if len(matches) == 1 {
		return matches[0], true
	}
	allOf := make([]*oas3.JSONSchema[oas3.Referenceable], len(matches))
	for i, match := range matches {
		allOf[i] = oas3.NewJSONSchemaFromSchema[oas3.Referenceable](match)
	}
	return &oas3.Schema{AllOf: allOf}, true
}

// cliProjectionPropertyDefinitelyMissing applies the generator's
// closed-world object semantics conservatively. Explicitly open additional
// properties, a matching pattern property, unevaluatedProperties, or a
// composition that cannot be inspected prevents the leading walk from
// turning the miss into a build error.
func cliProjectionPropertyDefinitelyMissing(schema *oas3.Schema, name string, seen map[*oas3.Schema]bool) bool {
	if schema == nil || len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 || schema.UnevaluatedProperties != nil {
		return false
	}
	if seen == nil {
		seen = map[*oas3.Schema]bool{}
	}
	if seen[schema] {
		return false
	}
	seen[schema] = true
	defer delete(seen, schema)

	for pattern := range schema.PatternProperties.All() {
		compiled, err := regexp.Compile(pattern)
		if err != nil || compiled.MatchString(name) {
			return false
		}
	}
	if additional := schema.AdditionalProperties; additional != nil {
		value := additional.GetBool()
		if value == nil || *value {
			return false
		}
	}
	for _, branch := range schema.AllOf {
		if !cliProjectionPropertyDefinitelyMissing(cliResolveJS(branch), name, seen) {
			return false
		}
	}
	return true
}

// cliProjectionSchemaTypes treats a JSON Schema type array as the set of
// admitted runtime types. allOf branches narrow that set by intersection.
func cliProjectionSchemaTypes(schema *oas3.Schema, seen map[*oas3.Schema]bool) []oas3.SchemaType {
	if schema == nil {
		return nil
	}
	if seen == nil {
		seen = map[*oas3.Schema]bool{}
	}
	if seen[schema] {
		return nil
	}
	seen[schema] = true
	defer delete(seen, schema)

	types := append([]oas3.SchemaType(nil), schema.GetType()...)
	for _, branch := range schema.AllOf {
		branchTypes := cliProjectionSchemaTypes(cliResolveJS(branch), seen)
		if len(branchTypes) == 0 {
			continue
		}
		if len(types) == 0 {
			types = append(types, branchTypes...)
			continue
		}
		intersection := types[:0]
		for _, candidate := range types {
			if cliSchemaIsType(branchTypes, candidate) {
				intersection = append(intersection, candidate)
			}
		}
		types = intersection
	}
	return types
}

var cliMapArgPattern = regexp.MustCompile(`map\(\s*\.([A-Za-z_][A-Za-z0-9_]*)\s*\)`)

// cliExplainMapProjection explains an array-of-nulls output produced by a
// `<arraypath> | map(.field)` projection: it locates the array the leading
// path selects and suggests item properties near the map argument.
func cliExplainMapProjection(expr string, response *oas3.Schema) (field, arrayPath, suggestion string, definitelyMissing bool) {
	match := cliMapArgPattern.FindStringSubmatch(expr)
	if match == nil {
		return "", "", "", false
	}
	field = match[1]

	segments, _ := cliLeadingJQPath(expr)
	current := response
	var walked strings.Builder
	for _, segment := range segments {
		if current == nil || segment.Field == "" {
			break
		}
		property, ok := cliProjectionProperty(current, segment.Field, nil)
		if !ok {
			break
		}
		current = property
		walked.WriteString("." + segment.Field)
	}
	if current == nil || !cliSchemaIsType(cliProjectionSchemaTypes(current, nil), oas3.SchemaTypeArray) || current.Items == nil {
		return field, cliOrResponsePlain(walked.String()), "", false
	}
	item := cliResolveJS(current.Items)
	if _, exists := cliProjectionProperty(item, field, nil); exists {
		return field, cliOrResponsePlain(walked.String()), "", false
	}
	definitelyMissing = cliProjectionPropertyDefinitelyMissing(item, field, nil)
	return field, cliOrResponsePlain(walked.String()), cliJQDidYouMean(field, cliSchemaPropertyNames(item)), definitelyMissing
}

// cliStringBuiltins are jq builtins that require string input; applying one
// to a provably non-string value is a runtime error jq's engine currently
// widens through, so the renderer checks the cheap pure case itself.
var cliStringBuiltins = map[string]bool{
	"ascii_downcase": true, "ascii_upcase": true, "ltrimstr": true,
	"rtrimstr": true, "trimstr": true, "explode": true, "utf8bytelength": true,
}

var cliBuiltinRestPattern = regexp.MustCompile(`^\|\s*([a-z_0-9]+)\s*(\(.*\))?\s*$`)

// cliStringBuiltinMismatch flags `<path> | <string-builtin>` where the walked
// path provably ends on a non-string value.
func cliStringBuiltinMismatch(expr string, walk cliProjectionWalk) string {
	if walk.Mistake != "" || walk.UnionAt != "" || len(walk.FinalTypes) == 0 || walk.Rest == "" {
		return ""
	}
	match := cliBuiltinRestPattern.FindStringSubmatch(walk.Rest)
	if match == nil || !cliStringBuiltins[match[1]] {
		return ""
	}
	if cliSchemaIsType(walk.FinalTypes, oas3.SchemaTypeString) {
		return ""
	}
	segments, _ := cliLeadingJQPath(expr)
	var walked strings.Builder
	for _, segment := range segments {
		if segment.Field != "" {
			walked.WriteString("." + segment.Field)
		}
	}
	return fmt.Sprintf("applies the string operation %q to %s, which is %s", match[1], cliOrResponsePlainQuoted(walked.String()), cliTypeWithArticle(walk.FinalTypes[0]))
}

// --- helpers ---------------------------------------------------------------

func cliOrResponse(walked string) string {
	if walked == "" {
		return "the response"
	}
	return fmt.Sprintf("%q", walked)
}

func cliOrResponsePlain(walked string) string {
	if walked == "" {
		return "the response"
	}
	return walked
}

func cliOrResponsePlainQuoted(walked string) string {
	if walked == "" {
		return "the response"
	}
	return fmt.Sprintf("%q", walked)
}

func cliTypeWithArticle(t oas3.SchemaType) string {
	name := string(t)
	switch name {
	case "integer", "object", "array":
		return "an " + name
	default:
		return "a " + name
	}
}

// cliJQDidYouMean renders a jq-flavored suggestion (dotted field spelling).
func cliJQDidYouMean(got string, candidates []string) string {
	best := ""
	bestDist := 3
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
	return fmt.Sprintf(" (did you mean %q?)", "."+best)
}

func cliSchemaIsType(types []oas3.SchemaType, want oas3.SchemaType) bool {
	for _, t := range types {
		if t == want {
			return true
		}
	}
	return false
}

// cliSchemaIsNullOnly reports whether a schema admits null and nothing else.
func cliSchemaIsNullOnly(s *oas3.Schema) bool {
	if s == nil {
		return false
	}
	types := s.GetType()
	return len(types) == 1 && types[0] == oas3.SchemaTypeNull
}

// cliArrayItemSchema returns the resolved item schema when s is an array.
func cliArrayItemSchema(s *oas3.Schema) *oas3.Schema {
	if s == nil || !cliSchemaIsType(s.GetType(), oas3.SchemaTypeArray) || s.Items == nil {
		return nil
	}
	return cliResolveJS(s.Items)
}

func cliSchemaPropertyNames(s *oas3.Schema) []string {
	if s == nil {
		return nil
	}
	seenSchemas := map[*oas3.Schema]bool{}
	seenNames := map[string]bool{}
	var collect func(*oas3.Schema)
	collect = func(schema *oas3.Schema) {
		if schema == nil || seenSchemas[schema] {
			return
		}
		seenSchemas[schema] = true
		for name := range schema.Properties.All() {
			seenNames[name] = true
		}
		for _, branch := range schema.AllOf {
			collect(cliResolveJS(branch))
		}
	}
	collect(s)
	names := make([]string, 0, len(seenNames))
	for name := range seenNames {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// cliResolveJS unwraps a JSON Schema node after targeted reference resolution,
// while continuing to handle inline schemas directly.
func cliResolveJS(js *oas3.JSONSchema[oas3.Referenceable]) *oas3.Schema {
	if js == nil {
		return nil
	}
	if resolved := js.GetResolvedSchema(); resolved != nil {
		if left := resolved.GetLeft(); left != nil {
			return left
		}
	}
	return js.GetLeft()
}
