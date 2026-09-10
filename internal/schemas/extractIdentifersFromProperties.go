package schemas

import (
	"context"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

// Certain properties are fixed in the schema and therefore can be used as helpful identifiers
// e.g.
// ```yaml
//
//	type: object
//	properties:
//	  type:
//	    type: string
//	    enum:
//	    	- dog
//	required:
//	  - type
//
// ```
// In the above schema, `dog` quite clearly identifies this schema
func (s *Schemas) extractIdentifiersFromProperties(ctx context.Context, ps *oas3.JSONSchema[oas3.Referenceable], js *oas3.JSONSchema[oas3.Referenceable], docInfo *document.DocumentInfo) *sequencedmap.Map[string, string] {
	identifiers := sequencedmap.New[string, string]()

	// map[constPropKey] = constPropValue
	requiredConstProps := sequencedmap.New[string, string]()
	constProps := sequencedmap.New[string, string]()

	requiredPropertiesOnParentOneOf := make(map[string]bool)

	if parentOneOfSchema := ps.MustGetResolvedSchema().GetSchema(); parentOneOfSchema != nil {
		for _, key := range parentOneOfSchema.GetRequired() {
			requiredPropertiesOnParentOneOf[key] = true
		}
	}

	schema := js.MustGetResolvedSchema().GetSchema()
	if schema == nil {
		return identifiers
	}

	for key, propJS := range schema.GetProperties().All() {
		p, err := resolution.Resolve(ctx, propJS, docInfo)
		if err != nil {
			continue
		}
		if p.IsBool() {
			continue
		}
		prop := p.GetSchema()

		value := s.extractSingleStringValue(prop)
		if value == "" {
			continue
		}

		if slices.Contains(schema.GetRequired(), key) || requiredPropertiesOnParentOneOf[key] {
			requiredConstProps.Set(key, value)
		} else {
			constProps.Set(key, value)
		}
	}

	if requiredConstProps.Len() > 0 {
		for key, value := range requiredConstProps.All() {
			identifiers.Set(key, value)
		}
		return identifiers
	}

	if constProps.Len() == 1 {
		for key, value := range constProps.All() {
			identifiers.Set(key, value)
		}
	}
	return identifiers
}

func (s *Schemas) extractSingleStringValue(prop *oas3.Schema) string {
	if len(prop.GetType()) != 1 || prop.GetType()[0] != "string" {
		return ""
	}

	if len(prop.GetEnum()) == 1 {
		var enumVal any
		if err := prop.GetEnum()[0].Decode(&enumVal); err != nil {
			return ""
		}

		strValue, ok := enumVal.(string)
		if ok {
			isOpenEnum, err := s.Subsystem.Extensions.IsOpenEnum(prop)
			if err != nil {
				return ""
			}
			if isOpenEnum {
				return ""
			}
			return strValue
		}
	}

	if prop.Const != nil {
		return prop.GetConst().Value
	}

	return ""
}
