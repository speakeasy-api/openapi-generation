package validation

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation/sanitization"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

type nameReference struct {
	Name   string
	Line   int
	Result sanitization.Result
	In     openapi.ParameterIn
	Schema *oas3.Schema
}

type operationNameInfo struct {
	OperationNode           *yaml.Node
	OperationID             string
	HasOperationID          bool
	OperationIDNode         *yaml.Node
	MethodNameOverride      string
	MethodNameOverrideNode  *yaml.Node
	MethodGroupOverride     string
	MethodGroupOverrideNode *yaml.Node
	Tags                    []string
	TagsNode                *yaml.Node
	HTTPPath                string
	HTTPMethod              string
}

type sanitizedOperationNameResult struct {
	Line       int
	HTTPPath   string
	HTTPMethod string
	Result     sanitization.Result
}

func (o *sanitizedOperationNameResult) GetPath() string {
	if o.HTTPPath == "" {
		return o.HTTPMethod + " (callback)"
	}
	return o.HTTPMethod + " " + o.HTTPPath
}
