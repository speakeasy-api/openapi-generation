package extensions

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/openapi"
)

// defaultSSEOverloadSelectorName is the property/parameter name the SSE-overload validators look for.
const defaultSSEOverloadSelectorName = "stream"

// SSEOverloadConfig identifies the boolean field that toggles SSE overload for an operation.
// Name is not user-configurable today: always falls back to defaultSSEOverloadSelectorName.
type SSEOverloadConfig struct {
	In   string // "body" | "query"
	Name string // resolved field name
}

// OperationCanHaveSSEOverload validates that an operation is compatible with the x-speakeasy-sse-overload extension.
// The operation must have a boolean `stream` discriminator either in the request body or as a query parameter,
// and exactly two successful response content types: 1 text/event-stream and 1 application/json.
// Returns the resolved stream-field ref on success.
func (e *Extensions) OperationCanHaveSSEOverload(ctx context.Context, op *openapi.Operation, docInfo *document.DocumentInfo) (*SSEOverloadConfig, error) {
	ref, err := operationHasBodyStreamField(ctx, op, docInfo)
	if err != nil {
		return nil, err
	}
	if ref != nil {
		if err := operationHasSSEResponses(ctx, op, docInfo); err != nil {
			return nil, err
		}
		return ref, nil
	}

	ref, err = operationHasQueryStreamParameter(ctx, op, docInfo)
	if err != nil {
		return nil, err
	}
	if ref != nil {
		if err := operationHasSSEResponses(ctx, op, docInfo); err != nil {
			return nil, err
		}
		return ref, nil
	}

	return nil, errors.NewValidationError("x-speakeasy-sse-overload requires a 'stream' boolean field in the request body or query parameters", nil, nil)
}

// OperationCanInferSSEOverload validates body-stream eligibility for generation.inferSSEOverload.
// Inference intentionally remains body-only; query-stream operations must opt in explicitly.
// Returns (nil, nil) when the operation is simply not a candidate (no body, no stream field, etc.).
// Fatal failures (e.g. broken $ref) are surfaced as errors.
func (e *Extensions) OperationCanInferSSEOverload(ctx context.Context, op *openapi.Operation, docInfo *document.DocumentInfo) (*SSEOverloadConfig, error) {
	ref, err := operationHasBodyStreamField(ctx, op, docInfo)
	if err != nil {
		return nil, err
	}
	if ref == nil {
		return nil, nil
	}

	if err := operationHasSSEResponses(ctx, op, docInfo); err != nil {
		return nil, nil //nolint:nilerr // response shape mismatch is a soft skip for inference
	}

	return ref, nil
}

// operationHasBodyStreamField probes the request body for a boolean stream discriminator.
// Non-nil err is non-recoverable (e.g. resolver failure) and should be surfaced.
// A (nil, nil) return means the strategy doesn't apply (non-fatal): callers should try the next probe.
func operationHasBodyStreamField(ctx context.Context, op *openapi.Operation, docInfo *document.DocumentInfo) (*SSEOverloadConfig, error) {
	if op.RequestBody == nil {
		return nil, nil
	}

	requestBody, err := resolution.Resolve(ctx, op.GetRequestBody(), docInfo)
	if err != nil {
		return nil, err
	}

	if !requestBody.GetRequired() || requestBody.GetContent().Len() == 0 {
		return nil, nil
	}

	hasStreamField := false
	for _, mediaType := range requestBody.GetContent().All() {
		if mediaType.Schema != nil {
			hasStreamField, err = checkSchemaForStreamField(ctx, mediaType.GetSchema(), docInfo, defaultSSEOverloadSelectorName)
			if err != nil {
				return nil, err
			}
		}
		if hasStreamField {
			break
		}
	}

	if !hasStreamField {
		return nil, nil
	}

	return &SSEOverloadConfig{In: "body", Name: defaultSSEOverloadSelectorName}, nil
}

// operationHasQueryStreamParameter probes the operation's query parameters for a boolean stream discriminator.
// See operationHasBodyStreamField for the fatal/no-match convention.
func operationHasQueryStreamParameter(ctx context.Context, op *openapi.Operation, docInfo *document.DocumentInfo) (*SSEOverloadConfig, error) {
	for _, refParam := range op.GetParameters() {
		param, err := resolution.Resolve(ctx, refParam, docInfo)
		if err != nil {
			return nil, err
		}
		if param == nil || param.GetName() != defaultSSEOverloadSelectorName || param.GetIn() != openapi.ParameterInQuery {
			continue
		}
		ok, err := schemaAcceptsBoolean(ctx, param.GetSchema(), docInfo)
		if err != nil {
			return nil, err
		}
		if ok {
			return &SSEOverloadConfig{In: "query", Name: defaultSSEOverloadSelectorName}, nil
		}
	}

	return nil, nil
}

func operationHasSSEResponses(ctx context.Context, op *openapi.Operation, docInfo *document.DocumentInfo) error {
	if op.GetResponses() == nil {
		return errors.NewValidationError("x-speakeasy-sse-overload requires responses", nil, nil)
	}

	hasEventStream := false
	hasJSON := false
	responseCount := 0

	for statusCode, r := range op.GetResponses().All() {
		response, err := resolution.Resolve(ctx, r, docInfo)
		if err != nil {
			return err
		}

		// Only count 2XX responses (successful responses)
		if len(statusCode) == 3 && statusCode[0] == '2' {
			if response.Content != nil {
				for contentType := range response.GetContent().All() {
					responseCount++
					if contentType == "text/event-stream" {
						hasEventStream = true
					}
					if contentType == "application/json" {
						hasJSON = true
					}
				}
			}
		}
	}

	if !hasEventStream || !hasJSON {
		return errors.NewValidationError("x-speakeasy-sse-overload requires exactly one text/event-stream and one application/json response", nil, nil)
	}

	if responseCount != 2 {
		return errors.NewValidationError(fmt.Sprintf("x-speakeasy-sse-overload requires exactly 2 response content types (found %d)", responseCount), nil, nil)
	}

	return nil
}

// schemaAcceptsBoolean reports whether a schema accepts a boolean value,
// walking composition keywords with the right satisfiability semantics:
//   - allOf: intersection — every member must accept boolean.
//   - anyOf / oneOf: union — at least one member must accept boolean.
func schemaAcceptsBoolean(ctx context.Context, s *oas3.JSONSchema[oas3.Referenceable], docInfo *document.DocumentInfo) (bool, error) {
	if s == nil {
		return false, nil
	}

	js, err := resolution.Resolve(ctx, s, docInfo)
	if err != nil {
		return false, err
	}

	if js.IsBool() {
		return false, nil
	}

	schema := js.GetSchema()

	for _, t := range schema.GetType() {
		if t == "boolean" {
			return true, nil
		}
	}

	if len(schema.GetAllOf()) > 0 {
		for _, member := range schema.GetAllOf() {
			ok, err := schemaAcceptsBoolean(ctx, member, docInfo)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil
	}

	for _, composition := range [][]*oas3.JSONSchema[oas3.Referenceable]{
		schema.GetAnyOf(),
		schema.GetOneOf(),
	} {
		for _, member := range composition {
			ok, err := schemaAcceptsBoolean(ctx, member, docInfo)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}

	return false, nil
}

// checkSchemaForStreamField reports whether the schema declares a property named fieldName whose schema
// accepts a boolean value, walking composition keywords with the appropriate semantics:
//   - direct properties: lookup by name.
//   - allOf: any member declaring the field is sufficient (all members apply).
//   - oneOf/anyOf: every member must declare the field, otherwise some valid
//     instance could legally omit it and reporting "found" would be unsound.
func checkSchemaForStreamField(ctx context.Context, s *oas3.JSONSchema[oas3.Referenceable], docInfo *document.DocumentInfo, fieldName string) (bool, error) {
	if s == nil {
		return false, nil
	}

	js, err := resolution.Resolve(ctx, s, docInfo)
	if err != nil {
		return false, err
	}

	if js.IsBool() {
		return false, nil
	}

	schema := js.GetSchema()

	// Check direct properties
	if schema.Properties != nil {
		for propName, p := range schema.GetProperties().All() {
			if propName == fieldName {
				return schemaAcceptsBoolean(ctx, p, docInfo)
			}
		}
	}

	// Check allOf schemas
	if len(schema.AllOf) > 0 {
		for _, allOfSchema := range schema.AllOf {
			hasStreamField, err := checkSchemaForStreamField(ctx, allOfSchema, docInfo, fieldName)
			if err != nil {
				return false, err
			}
			if hasStreamField {
				return true, nil
			}
		}
	}

	// Check oneOf / anyOf — field must be present in every member.
	for _, composition := range [][]*oas3.JSONSchema[oas3.Referenceable]{
		schema.GetOneOf(),
		schema.GetAnyOf(),
	} {
		if len(composition) == 0 {
			continue
		}
		allHave := true
		for _, member := range composition {
			hasStreamField, err := checkSchemaForStreamField(ctx, member, docInfo, fieldName)
			if err != nil {
				return false, err
			}
			if !hasStreamField {
				allHave = false
				break
			}
		}
		if allHave {
			return true, nil
		}
	}

	return false, nil
}
