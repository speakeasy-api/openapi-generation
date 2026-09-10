package ast

import (
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

type UsageExampleScope struct {
	// OpFilter determines which operations should be used for usage the snippet
	// See SelectExampleOperations() function for more details.
	// If the string starts with an exclamation mark (`!`), then the filter is
	// used to _exclude_ matching operations.
	OpFilter string

	// Feature is a tag that can be used to determine which aspect or config needs
	// to be covered in the usage snippet.
	Feature string

	// If IsGlobal is true the usage snippet SDK should be configured globally,
	// otherwise the feature should be enabled at the operation level.
	IsGlobal bool

	// Value is the value of the feature that should be used in the snippet.
	Value any
}

// UsageContext represents the information required to render a usage snippet.
type UsageContext struct {
	// Operation is the operation to include in the snippet. It is technically okay
	// not to set Operation to nil, in which case the snippet will describe
	// initialization, but not show an example of calling any operation.
	Operation *Operation

	// StepIdx is the index of the step and therefore the operation in a multi-step arazzo workflow
	StepIdx int
	// StepID is the ID of the step in a multi-step arazzo workflow
	StepID string

	// SDK is the SDK that contains the operation.
	SDK *SDK

	// Whether to instantiate the SDK for this snippet (in a test for example multiple operations may use the same SDK or their own instance)
	SkipSDKInstantiation bool
	// ContextIndex is the index of this context in a multi-context snippet. If its the first then for example in go the SDK variable will be declared instead of overridden
	ContextIndex int

	// Config is the content of the x-usage-example extension associated with this
	// snippet (if any). It should always be set to a struct containing zero values.
	Config *extensions.UsageExampleConfig

	// Scopes hold information on which feature(s) the usage snippet is rendered for
	Scopes []UsageExampleScope

	// SkipResponseBodyAssertions indicates whether the usage snippet should skip templating response body assertions
	SkipResponseBodyAssertions bool

	// The name of the example this usage snippet is associated with if being used with a specific named example from the OpenAPI doc
	ExampleName string

	// IsMainExample indicates whether this usage snippet is a main example for the SDK.
	IsMainExample bool

	// Test is the parent test workflow associated with this usage snippet
	Test *ArazzoWorkflow

	// Assertions are test-only assertions associated with this step's usage context.
	Assertions []Assertion

	// UsingMockServer indicates whether the enclosing test is targeting the generated mock server.
	UsingMockServer bool

	// Controls if this is rendered as a usage snippet in the async paradigm(used in 1 target: java)
	AsyncMode bool
}

func (usageContext UsageContext) RenderFeature(feature string, global bool) bool {
	for _, scope := range usageContext.Scopes {
		if scope.Feature == feature && scope.IsGlobal == global {
			return true
		}
	}
	return false
}

func CreateUsageContext(sdk *SDK, operation *Operation, usageExample *extensions.UsageExampleConfig, skipResponseBodyAssertions bool) *UsageContext {
	return &UsageContext{
		SDK:                        sdk,
		Operation:                  operation,
		Config:                     usageExample,
		Scopes:                     []UsageExampleScope{},
		SkipResponseBodyAssertions: skipResponseBodyAssertions,
	}
}

func (u *UsageContext) PopulateGlobalParameterScopes(exampleName string, shouldIncludeServerSelection bool) {
	// Populate global parameter scopes
	u.populateGlobalParameterScopes(exampleName)

	// Populate server selection scope if multiple servers are defined and requested
	if shouldIncludeServerSelection {
		u.populateServerSelectionScope()
	}
}

func (u *UsageContext) populateGlobalParameterScopes(exampleName string) {
	if u.Operation == nil || u.Operation.Globals == nil {
		return
	}

	opParams := []*Param{}

	if u.Operation.Request != nil && u.Operation.Request.Params != nil {
		opParams = append(opParams, u.Operation.Request.Params.QueryParams...)
		opParams = append(opParams, u.Operation.Request.Params.PathParams...)
		opParams = append(opParams, u.Operation.Request.Params.HeaderParams...)
	}

	for _, global := range u.Operation.Globals.Fields {
		idx := slices.IndexFunc(opParams, func(p *Param) bool {
			paramName := p.Field.OriginalName
			if paramName == "" {
				paramName = p.Field.Name
			}

			globalName := global.OriginalName
			if globalName == "" {
				globalName = global.Name
			}

			return paramName == globalName
		})
		if idx == -1 {
			continue
		}

		opParam := opParams[idx]
		if opParam.Field.Type.Type != global.Type.Type {
			continue
		}

		anno := opParam.Field.Annotations.Get(AnnotationTypeParam)
		if anno == nil {
			continue
		}
		paramAnno, ok := anno.(*ParamAnnotation)
		if !ok {
			continue
		}

		var paramExample *Example
		for _, e := range opParam.Examples {
			if e.Name() == exampleName {
				paramExample = e
				break
			}
		}
		if paramExample == nil && len(opParam.Examples) > 0 {
			paramExample = opParam.Examples[0]
		}

		u.Scopes = append(u.Scopes, UsageExampleScope{
			Feature:  "parameter",
			IsGlobal: true,
			Value: map[string]any{
				"example":              paramExample,
				"field":                global,
				"hidden":               paramAnno.Hidden,
				"requiredForOperation": paramAnno.RequiredForOperation,
			},
		})
	}
}

func (u *UsageContext) populateServerSelectionScope() {
	if u.SDK == nil || u.SDK.Servers == nil {
		return
	}

	// Check if we already have a server_selection scope to avoid duplicates
	for _, scope := range u.Scopes {
		if scope.Feature == "server_selection" && scope.IsGlobal {
			return
		}
	}

	// Only add server selection scope if multiple servers are defined
	if len(u.SDK.Servers.Servers) < 2 {
		return
	}

	// Add server_selection scope to include server parameter in code samples
	u.Scopes = append(u.Scopes, UsageExampleScope{
		Feature:  "server_selection",
		IsGlobal: true,
		OpFilter: "",
		Value:    false, // getUsageGlobalServer() will return the first server unless ServerToShowInSnippets is defined in gen.yaml
	})
}
