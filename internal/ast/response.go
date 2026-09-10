package ast

import (
	"errors"
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
)

// SubResponse represents one of the possible responses that an operation can return
type SubResponse struct {
	Code    []string               // The status code associated with this particular response
	Headers bool                   // Whether or not the response has headers
	Content []*ResponseBodyContent // The content that is returned in the response
	Error   bool                   // Whether or not the response is an error
}

func (r *SubResponse) Match(matchers Matchers) error {
	if matchers.SubResponse != nil {
		return matchers.SubResponse(r)
	}

	return nil
}

func (r *SubResponse) IsEqualExceptExactCode(other *SubResponse) bool {

	if r == nil && other == nil {
		return true
	}
	if r == nil && other != nil || other == nil && r != nil {
		return false
	}

	if r.Headers != other.Headers {
		return false
	}
	if len(r.Content) != len(other.Content) {
		return false
	}
	if r.Error != other.Error {
		return false
	}
	if r.Code[0][0] != other.Code[0][0] {
		return false
	}

	for i := range r.Content {
		if !r.Content[i].IsEqual(other.Content[i]) {
			return false
		}
	}
	return true
}

// Collection of SubResponse.
type SubResponses []*SubResponse

// If found, returns the SubResponse with given code.
func (r SubResponses) FindByCode(code string) *SubResponse {
	for _, subResponse := range r {
		for _, c := range subResponse.Code {
			if c == code {
				return subResponse
			}
		}
	}

	return nil
}

// Response represents the response output from an operation
type Response struct {
	Type      *TypeDef     // The type of the response object
	Responses SubResponses // The list of possible responses that the operation can return
}

// Returns true if the Response contains a truncated (circular reference) type.
func (r *Response) ContainsTruncated() bool {
	if r == nil || r.Type == nil {
		return false
	}

	return r.Type.ContainsTruncated()
}

// Returns the TypeDef or underlying TypeDef where the given entity name matches
// the x-speakeasy-entity extension configuration.
func (r *Response) FindEntityTypeDef(entityName string) *TypeDef {
	if r == nil || r.Type == nil {
		return nil
	}

	return r.Type.FindEntityTypeDef(entityName)
}

// Returns the first SubResponse with a non-error status code.
func (r *Response) FirstSuccessCodeSubResponse() *SubResponse {
	for _, subResponse := range r.Responses {
		for _, code := range subResponse.Code {
			if code[0] == '2' {
				return subResponse
			}
		}
	}

	return nil
}

func (r *Response) Match(matchers Matchers) error {
	if matchers.Response != nil {
		return matchers.Response(r)
	}

	return nil
}

// Returns the root FieldDef representing the body of the response for
// Terraform.
func (r *Response) TerraformBodyFieldDef(entityName string) *FieldDef {
	if r == nil {
		return nil
	}

	if r.Type != nil {
		for _, fieldDef := range r.Type.Fields {
			if fieldDef.FindEntityFieldDef(entityName) == nil {
				continue
			}

			// Always return the original FieldDef, not the found one, so that
			// any parent data to x-speakeasy-entity is preserved for later
			// hoisting logic.
			return fieldDef
		}
	}

	subResponse := r.FirstSuccessCodeSubResponse()

	if subResponse == nil {
		return nil
	}

	for _, responseBodyContent := range subResponse.Content {
		if responseBodyContent.SerializationMethod == "json" {
			return responseBodyContent.Content
		}
	}

	// NOTE: This logic was copied as-is from its prior implementation. The
	// media type might not be supported by the templating. Other validation
	// logic should raise an error before reaching here, if necessary.
	if len(subResponse.Content) > 0 {
		return subResponse.Content[0].Content
	}

	return nil
}

func (r Response) GetErrorStatusCodes() []string {
	codes := []string{}

	for _, resp := range r.Responses {
		if resp.Error {
			codes = append(codes, resp.Code...)
		}
	}

	slices.Sort(codes)
	return codes
}

func (r *Response) pollingResponseBodyAssertion(statusCode string, contentType string, assertionType AssertionType, condition *criterion.Condition) (*Assertion, error) {
	if r == nil {
		return nil, errors.New("response is nil for polling response body assertion")
	}

	statusCodeSubResponse := r.Responses.FindByCode(statusCode)

	if statusCodeSubResponse == nil {
		return nil, fmt.Errorf("no response found for status code %s in polling assertion", statusCode)
	}

	if len(statusCodeSubResponse.Content) == 0 {
		return nil, errors.New("polling response body assertions are not supported for responses with no content")
	}

	if len(statusCodeSubResponse.Content) > 1 && contentType == "" {
		// TODO: Support multiple content types by allowing the user to specify
		// $response.header.Content-Type == value assertion, then use another
		// priorAssertions.FindAssertionByTarget() above to only raise this
		// error when that assertion or subresponse content is not found.
		return nil, errors.New("polling response body assertions are not currently supported for responses with multiple content types")
	}

	responseBodyContent := statusCodeSubResponse.Content[0]

	return responseBodyContent.pollingResponseBodyAssertion(assertionType, condition)
}
