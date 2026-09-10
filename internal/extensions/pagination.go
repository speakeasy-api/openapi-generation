package extensions

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/token"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

// simpleTokens is the set of tokens which are allowed in a simple JSONPath
// expression. Such expressions can be evaluated with a simple, performant
// object drilling algorithm instead of a full-blow JSONPath evaluator.
var simpleTokens = map[token.Token]struct{}{
	token.ROOT:   {},
	token.CHILD:  {},
	token.STRING: {},
}

type PaginationType string

const (
	PaginationTypeOffsetLimit PaginationType = "offsetLimit"
	PaginationTypeCursor      PaginationType = "cursor"
	PaginationTypeURL         PaginationType = "url"
)

type PaginationInputInType string

const (
	PaginationInputInTypeParameters  PaginationInputInType = "parameters"
	PaginationInputInTypeRequestBody PaginationInputInType = "requestBody"
)

type PaginationInputType string

const (
	PaginationInputTypeLimit  PaginationInputType = "limit"
	PaginationInputTypeOffset PaginationInputType = "offset"
	PaginationInputTypePage   PaginationInputType = "page"
	PaginationInputTypeCursor PaginationInputType = "cursor"
)

type PaginationInputs struct {
	Name     string                `json:"name" yaml:"name"`
	In       PaginationInputInType `json:"in" yaml:"in"`
	Type     PaginationInputType   `json:"type" yaml:"type"`
	Optional bool
}

// Clone creates a deep copy of the PaginationInputs
func (i PaginationInputs) Clone() PaginationInputs {
	return PaginationInputs{
		In:       i.In,
		Name:     i.Name,
		Optional: i.Optional,
		Type:     i.Type,
	}
}

type PaginationOutputs struct {
	// CanUseDotNotation indicates that the JSONPath expressions in this
	// pagination config are simple enough and it is possible to use a
	// lightweight and more performant object-drilling library instead of a
	// JSONPath library.
	CanUseDotNotation bool `json:"-" yaml:"-"`

	Results    string `json:"results" yaml:"results"`
	ResultsDot string `json:"-" yaml:"-"`

	NumPages    string `json:"numPages" yaml:"numPages"`
	NumPagesDot string `json:"-" yaml:"-"`

	NextCursor    string `json:"nextCursor" yaml:"nextCursor"`
	NextCursorDot string `json:"-" yaml:"-"`

	NextURL    string `json:"nextUrl" yaml:"nextUrl"`
	NextURLDot string `json:"-" yaml:"-"`
}

// Clone creates a deep copy of the PaginationOutputs
func (o PaginationOutputs) Clone() PaginationOutputs {
	return PaginationOutputs{
		CanUseDotNotation: o.CanUseDotNotation,
		NextCursor:        o.NextCursor,
		NextCursorDot:     o.NextCursorDot,
		NextURL:           o.NextURL,
		NextURLDot:        o.NextURLDot,
		NumPages:          o.NumPages,
		NumPagesDot:       o.NumPagesDot,
		Results:           o.Results,
		ResultsDot:        o.ResultsDot,
	}
}

type Pagination struct {
	Type    PaginationType     `json:"type" yaml:"type"`
	Inputs  []PaginationInputs `json:"inputs" yaml:"inputs"`
	Outputs PaginationOutputs  `json:"outputs" yaml:"outputs"`
}

// Clone creates a deep copy of the Pagination
func (p *Pagination) Clone() *Pagination {
	if p == nil {
		return nil
	}

	clonedInputs := make([]PaginationInputs, 0, len(p.Inputs))

	for _, input := range p.Inputs {
		clonedInputs = append(clonedInputs, input.Clone())
	}

	cloned := &Pagination{
		Type:    p.Type,
		Inputs:  clonedInputs,
		Outputs: p.Outputs.Clone(),
	}

	return cloned
}

var jsonPathRegex = regexp.MustCompile(`^\$.*`)

func (e *Extensions) HandleOperationPaginationExtension(ctx context.Context, operation *openapi.Operation, docInfo *document.DocumentInfo) (*Pagination, error) {
	config, node, err := e.parsePaginationExtension(operation.GetExtensions())
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	if reflect.DeepEqual(*config, Pagination{}) {
		return config, nil
	}

	err = validatePaginationConfig(ctx, config, operation, node, docInfo)
	if err != nil {
		return nil, err
	}

	_ = markOptionalInputs(ctx, config, operation, docInfo)

	ComputePaginationDotNotation(config)

	return config, nil
}

func (e *Extensions) parsePaginationExtension(ext OAExtensions) (*Pagination, *yaml.Node, error) {
	if ext.Len() == 0 {
		return nil, nil, nil
	}

	paginationExtension, ok := e.findExtension(ext, ExtPagination)
	if !ok {
		return nil, nil, nil
	}

	var p Pagination
	if err := paginationExtension.Decode(&p); err != nil {
		return nil, nil, errors.NewValidationError("failed to unmarshal "+ExtPagination.Name(), paginationExtension, err)
	}

	return &p, paginationExtension, nil
}

func validatePaginationConfig(ctx context.Context, config *Pagination, operation *openapi.Operation, node *yaml.Node, docInfo *document.DocumentInfo) error {
	err := validatePaginationInput(ctx, config.Inputs, PaginationInputTypeLimit, operation, node, false, docInfo)
	if err != nil {
		return errors.NewValidationError(ExtPagination.Name()+" configuration error: invalid inputs.limit", node, err)
	}
	switch config.Type {
	case PaginationTypeCursor:
		err := validatePaginationInput(ctx, config.Inputs, PaginationInputTypeCursor, operation, node, true, docInfo)
		if err != nil {
			return err
		}

		if !jsonPathRegex.MatchString(config.Outputs.NextCursor) {
			return errors.NewValidationError(fmt.Sprintf("%s type %s requires output.nextCursor", ExtPagination.Name(), PaginationTypeCursor), node, nil)
		}
	case PaginationTypeURL:
		if !jsonPathRegex.MatchString(config.Outputs.NextURL) {
			return errors.NewValidationError(fmt.Sprintf("%s type %s requires output.nextUrl", ExtPagination.Name(), PaginationTypeURL), node, nil)
		}
	case PaginationTypeOffsetLimit:
		pageErr := validatePaginationInput(ctx, config.Inputs, PaginationInputTypePage, operation, node, true, docInfo)
		offsetErr := validatePaginationInput(ctx, config.Inputs, PaginationInputTypeOffset, operation, node, true, docInfo)

		if pageErr != nil && offsetErr != nil {
			return errors.NewValidationError(fmt.Sprintf("%s type %s requires either input.page input.offset", ExtPagination.Name(), PaginationTypeOffsetLimit), node, offsetErr)
		}

		numPagesOutputValid := jsonPathRegex.MatchString(config.Outputs.NumPages)
		resultsOutputValid := jsonPathRegex.MatchString(config.Outputs.Results)

		if pageErr == nil {
			// page requires either output.numPages _or_ output.results
			if !numPagesOutputValid && !resultsOutputValid {
				return errors.NewValidationError(fmt.Sprintf("%s type %s requires output.numPages OR output.Results", ExtPagination.Name(), PaginationTypeCursor), node, nil)
			}
		}

		if offsetErr == nil {
			// offset requires output.results
			if !resultsOutputValid {
				return errors.NewValidationError(fmt.Sprintf("%s type %s requires output.Results", ExtPagination.Name(), PaginationTypeCursor), node, nil)
			}
		}
	default:
		return errors.NewValidationError(fmt.Sprintf("invalid %s.type '%s'", ExtPagination.Name(), config.Type), node, nil)
	}
	return nil
}

func validatePaginationInput(ctx context.Context, inputs []PaginationInputs, t PaginationInputType, operation *openapi.Operation, node *yaml.Node, isRequired bool, docInfo *document.DocumentInfo) error {
	for _, in := range inputs {
		if in.Type == t {
			switch in.In {
			case PaginationInputInTypeParameters:
				exists, err := paramExists(ctx, operation, in.Name, docInfo)
				if err != nil {
					return err
				}
				if !exists {
					return errors.NewValidationError(fmt.Sprintf("parameter %s not found for operation %s, but is required for %s", in.Name, operation.GetOperationID(), ExtPagination.Name()), node, nil)
				}
				return nil
			case PaginationInputInTypeRequestBody:
				exists, err := requestBodyFieldExists(ctx, operation, in.Name, docInfo)
				if err != nil {
					return err
				}
				if !exists {
					return errors.NewValidationError(fmt.Sprintf("request body field %s not found for operation %s, but is required for %s", in.Name, operation.GetOperationID(), ExtPagination.Name()), node, nil)
				}
				return nil
			default:
				return errors.NewValidationError(fmt.Sprintf("invalid field 'in' for %s input %s '%s'", ExtPagination.Name(), t, in.In), node, nil)
			}
		}
	}
	if isRequired {
		return errors.NewValidationError(fmt.Sprintf("%s type %s requires %s input", ExtPagination.Name(), "", t), node, nil)
	}
	return nil
}

func markOptionalInputs(ctx context.Context, config *Pagination, operation *openapi.Operation, docInfo *document.DocumentInfo) error {
	for i, input := range config.Inputs {
		config.Inputs[i].Optional = true
		switch input.In {
		case PaginationInputInTypeParameters:
			for _, p := range operation.Parameters {
				param, err := resolution.Resolve(ctx, p, docInfo)
				if err != nil {
					return err
				}

				if param.GetName() == input.Name {
					config.Inputs[i].Optional = !param.GetRequired()
					break
				}
			}
		case PaginationInputInTypeRequestBody:
			rb, err := resolution.Resolve(ctx, operation.RequestBody, docInfo)
			if err != nil {
				return err
			}

			for mt := range rb.GetContent().Values() {
				schema, err := resolution.Resolve(ctx, mt.Schema, docInfo)
				if err != nil {
					return err
				}
				if schema.IsBool() {
					continue
				}

				for _, s := range schema.GetSchema().Required {
					if s == input.Name {
						config.Inputs[i].Optional = false
					}
				}
			}
		}
	}

	return nil
}

func paramExists(ctx context.Context, operation *openapi.Operation, name string, docInfo *document.DocumentInfo) (bool, error) {
	for _, p := range operation.Parameters {
		param, err := resolution.Resolve(ctx, p, docInfo)
		if err != nil {
			return false, err
		}
		if param.Name != name {
			continue
		}
		return true, nil
	}
	return false, nil
}

func requestBodyFieldExists(ctx context.Context, operation *openapi.Operation, name string, docInfo *document.DocumentInfo) (bool, error) {
	if operation.RequestBody == nil {
		return false, nil
	}
	requestBody, err := resolution.Resolve(ctx, operation.RequestBody, docInfo)
	if err != nil {
		return false, err
	}
	mt, ok := requestBody.Content.Get("application/json")
	if !ok {
		return false, nil
	}
	schema, err := resolution.Resolve(ctx, mt.Schema, docInfo)
	if err != nil {
		return false, err
	}
	if schema.IsBool() {
		return false, nil
	}

	_, found := schema.GetSchema().Properties.Get(name)

	return found, nil
}

func ComputePaginationDotNotation(config *Pagination) {
	config.Outputs.CanUseDotNotation = canUseDotNotation(config)
	if !config.Outputs.CanUseDotNotation {
		return
	}

	config.Outputs.ResultsDot = jsonpathToDot(config.Outputs.Results)
	config.Outputs.NumPagesDot = jsonpathToDot(config.Outputs.NumPages)
	config.Outputs.NextCursorDot = jsonpathToDot(config.Outputs.NextCursor)
	config.Outputs.NextURLDot = jsonpathToDot(config.Outputs.NextURL)
}

func canUseDotNotation(config *Pagination) bool {
	outputs := []string{
		config.Outputs.Results,
		config.Outputs.NumPages,
		config.Outputs.NextCursor,
		config.Outputs.NextURL,
	}

	simple := true
	for _, jp := range outputs {
		if jp == "" {
			continue
		}

		stream := token.NewTokenizer(jp).Tokenize()
		for _, tokenInfo := range stream {
			_, allowed := simpleTokens[tokenInfo.Token]
			if !allowed {
				return false
			}
		}
	}

	return simple
}

func jsonpathToDot(expression string) string {
	if expression == "" {
		return ""
	}

	result := ""
	for i, tokenInfo := range token.NewTokenizer(expression).Tokenize() {
		if i < 2 { // Drop the leading `$.`
			continue
		}

		switch tokenInfo.Token {
		case token.STRING:
			result += tokenInfo.Literal
		case token.CHILD:
			result += "."
		default:
			panic(fmt.Errorf("unexpected JSONPath token '%v' in '%s' at %d:%d", tokenInfo.Token, expression, tokenInfo.Line, tokenInfo.Column))
		}
	}

	return result
}
