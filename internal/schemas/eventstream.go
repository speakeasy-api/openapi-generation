package schemas

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
)

func postProcessEventStreamEnvelope(objType *ast.TypeDef) error {
	for _, f := range objType.Fields {
		if f == nil {
			continue
		}

		// Use the original property name from the OpenAPI spec for SSE field matching.
		// f.Name may have been overridden by x-speakeasy-name-override on a $ref target,
		// but SSE envelope fields must use their original names (data, event, id, retry).
		if f.OriginalName != "" && f.Name != f.OriginalName {
			f.Name = f.OriginalName
		}

		switch {
		case f.Name == "data":
			typ := f.Type.Type
			switch {
			case f.Type.IsObjectType(), f.Type.IsContainer():
				// We assume objects are encoded as JSON strings
				f.Annotations = append(f.Annotations, &ast.EncodingAnnotation{MediaType: "application/json"})
			case f.Type.Enum != nil && f.Type.Enum.Type != nil && f.Type.Enum.Type.Type == ast.DataTypeString:
				// nothing to do
			case typ == ast.DataTypeString:
				// nothing to do
			default:
				return errors.NewValidationError("server-sent event: 'data' field may be a string, object, or array", nil, nil)
			}
		case f.Name == "event":
			if f.Type.Type != ast.DataTypeString && (f.Type.Type != ast.DataTypeEnum || f.Type.Enum.Type.Type != ast.DataTypeString) {
				return errors.NewValidationError("server-sent event: 'event' field must have a string type", nil, nil)
			}
		case f.Name == "retry":
			if f.Type.Type != ast.DataTypeInteger && f.Type.Type != ast.DataTypeInt32 {
				return errors.NewValidationError("server-sent event: 'retry' field must have an integer type", nil, nil)
			}
		case f.Name == "id":
			if f.Type.Type != ast.DataTypeString {
				return errors.NewValidationError("server-sent event: 'id' field must have a string type", nil, nil)
			}
		case f.Const != nil:
			// Additional fields that are consts are fine to have
		default:
			return errors.NewValidationError("server-sent event: unknown field '"+f.Name+"'", nil, nil)
		}

		f.Annotations = append(f.Annotations, &ast.JSONAnnotation{FieldName: f.Name})
	}

	return nil
}

func selectEventStreamFieldSerialization(
	ctx context.Context,
	prop string,
	propSchema *oas3.JSONSchema[oas3.Concrete],
	docInfo *document.DocumentInfo,
) (ast.SerializationMethod, error) {
	switch prop {
	case "data":
		isComplex, err := openapi.IsComplex(ctx, propSchema, docInfo)
		if err != nil {
			return "", err
		}

		if isComplex {
			return ast.SerializationMethodJSON, nil
		}
	case "retry":
		return ast.SerializationMethodJSON, nil
	}

	return ast.SerializationMethodString, nil
}
