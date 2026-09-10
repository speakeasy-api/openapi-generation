package ast

// RequestParams represent the different parameter types that can be passed to an operation
type RequestParams struct {
	QueryParams  []*Param `yaml:",omitempty"`
	PathParams   []*Param `yaml:",omitempty"`
	HeaderParams []*Param `yaml:",omitempty"`
}

func (r *RequestParams) Match(matchers Matchers) error {
	if matchers.RequestParams != nil {
		return matchers.RequestParams(r)
	}

	return nil
}

func (r RequestParams) HasQueryParams() bool {
	return len(r.QueryParams) > 0
}

func (r RequestParams) HasPathParams() bool {
	return len(r.PathParams) > 0
}

func (r RequestParams) HasHeaderParams() bool {
	return len(r.HeaderParams) > 0
}

func (r *RequestParams) IsEmpty() bool {
	if r == nil {
		return true
	}

	return !r.HasQueryParams() && !r.HasPathParams() && !r.HasHeaderParams()
}

func (r *RequestParams) HasVisibleParams() bool {
	if r == nil {
		return false
	}

	for _, param := range r.QueryParams {
		if !param.Hidden {
			return true
		}
	}

	for _, param := range r.PathParams {
		if !param.Hidden {
			return true
		}
	}

	for _, param := range r.HeaderParams {
		if !param.Hidden {
			return true
		}
	}

	return false
}

// Request represents the input to an operation
type Request struct {
	Field                 *FieldDef      `yaml:",omitempty"` // The field representing the entire request object
	RequestBody           *FieldDef      `yaml:",omitempty"` // The request body field
	MatchedContentTypes   []string       `yaml:",omitempty"` // The content types that were matched and flattened into this request body
	IsRequestBody         bool           `yaml:",omitempty"` // Whether the request is the same as the request body after flattening
	IsRequestBodyRequired bool           `yaml:",omitempty"` // Whether the request is required
	Params                *RequestParams `yaml:",omitempty"` // The parameters that can be passed to the operation
	Examples              Examples       `yaml:",omitempty"` // The examples of the request
}

// Returns true if the Request contains a truncated (circular reference) type.
func (r *Request) ContainsTruncated() bool {
	if r == nil || r.Field == nil || r.Field.Type == nil {
		return false
	}

	return r.Field.Type.ContainsTruncated()
}

// Returns the TypeDef or underlying TypeDef where the given entity name matches
// the x-speakeasy-entity extension configuration.
func (r *Request) FindEntityTypeDef(entityName string) *TypeDef {
	if r == nil || r.Field == nil || r.Field.Type == nil {
		return nil
	}

	return r.Field.Type.FindEntityTypeDef(entityName)
}

func (r *Request) Match(matchers Matchers) error {
	if matchers.Request != nil {
		return matchers.Request(r)
	}

	return nil
}
