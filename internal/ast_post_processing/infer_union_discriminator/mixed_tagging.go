package infer_union_discriminator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

// SummarizeMixedTagging inspects a union whose discriminator inference failed
// and reports how its members are tagged, so the silent loss of the
// forward-compatible Unknown* fallback can be surfaced to the user.
//
// A member is considered "tagged" when it has a required property constrained
// to known values (const or closed enum) — i.e. it looks like it was designed
// to participate in a discriminated union.
//
// Returns ok=false (no warning warranted) when:
//   - fewer than two members are tagged: a single incidentally-constrained
//     member (e.g. a `status` enum) next to a primitive does not indicate
//     discriminated-union intent, or
//   - any member is a map/any type: such a member already absorbs unknown
//     object shapes, so the union is effectively forward-compatible.
func SummarizeMixedTagging(typeDef *ast.TypeDef) (summary string, ok bool) {
	taggedBy := map[string][]string{}
	taggedMembers := 0
	var untagged []string

	for _, member := range typeDef.AssociatedTypes {
		if member.Type == ast.DataTypeMap || member.Type == ast.DataTypeAny {
			// This member absorbs arbitrary objects: unknown variants from
			// newer servers will still parse into it.
			return "", false
		}
		props := taggedProperties(member)
		if len(props) == 0 {
			untagged = append(untagged, memberName(member))
			continue
		}
		taggedMembers++
		for _, prop := range props {
			taggedBy[prop] = append(taggedBy[prop], memberName(member))
		}
	}

	if taggedMembers < 2 {
		return "", false
	}

	props := make([]string, 0, len(taggedBy))
	for prop := range taggedBy {
		props = append(props, prop)
	}
	sort.Strings(props)

	parts := make([]string, 0, len(props))
	for _, prop := range props {
		parts = append(parts, fmt.Sprintf("%q (%s)", prop, strings.Join(taggedBy[prop], ", ")))
	}

	summary = "members tag on: " + strings.Join(parts, "; ")
	if len(untagged) > 0 {
		summary += "; untagged members: " + strings.Join(untagged, ", ")
	}
	return summary, true
}

// taggedProperties returns the names of a member's required, value-constrained
// properties — its candidate discriminator tags.
func taggedProperties(member *ast.TypeDef) []string {
	if !member.IsTypeWithFields() {
		return nil
	}

	var props []string
	for _, field := range member.Fields {
		if field.Optional {
			continue
		}
		// Mirror the inference rules: the field must be a valid discriminator
		// type (parsePrimitiveDataType rejects open enums, which accept
		// arbitrary values) and constrained to known values.
		if parsePrimitiveDataType(field) == "" {
			continue
		}
		if getPossibleConsts(field) == nil {
			continue
		}
		props = append(props, getOriginalFieldName(field))
	}
	return props
}

func memberName(member *ast.TypeDef) string {
	if member.Name != "" {
		return member.Name
	}
	return string(member.Type)
}
