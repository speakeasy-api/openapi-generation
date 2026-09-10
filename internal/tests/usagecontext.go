package tests

import (
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

func applyTestContextToUsageContext(usageContext *ast.UsageContext, testWorkflow *ast.ArazzoWorkflow, testStep *ast.ArazzoOperationStep, usingMockServer bool) *ast.UsageContext {
	// Skip templating globals in SDK configuration and set at the operation level so values are isolated
	// TODO: we may want to allow this to be configured or remove this restriction to allow the testing of globals
	usageContext.ExampleName = fmt.Sprintf("%s[%d]", testWorkflow.Name, testStep.StepIdx)
	usageContext.Test = testWorkflow
	usageContext.UsingMockServer = usingMockServer
	usageContext.StepIdx = testStep.StepIdx
	usageContext.StepID = testStep.StepID

	// Set the content_type for the operation if multiple content types are possible and the response we want is not the default
	numContentTypes := 0
	for _, subResponse := range testStep.Operation.Response.Responses {
		for range subResponse.Content {
			numContentTypes++
		}
	}

	if numContentTypes > 1 && testStep.ResponseContentType != "" {
		acceptTypes := testStep.Operation.GetAcceptTypes()
		if len(acceptTypes) > 0 && !strings.HasPrefix(acceptTypes[0], testStep.ResponseContentType) {
			usageContext.Scopes = append(usageContext.Scopes, ast.UsageExampleScope{
				Feature:  "content_type",
				IsGlobal: false,
				OpFilter: testStep.ResponseContentType,
			})
		}
	}

	// Set server url if provided
	if testWorkflow.Server != "" {
		isGlobalServerUrl := true

		if testStep.Operation.Servers != nil {
			isGlobalServerUrl = false
		}

		usageContext.Scopes = append(usageContext.Scopes, ast.UsageExampleScope{
			Feature:  "server_url",
			IsGlobal: isGlobalServerUrl,
			OpFilter: "",
			Value:    testWorkflow.Server,
		})
	}

	if testStep.Security != nil {
		isGlobalSecurity := true

		if testStep.Operation.Security != nil {
			isGlobalSecurity = false
		}

		usageContext.Scopes = append(usageContext.Scopes, ast.UsageExampleScope{
			Feature:  "security",
			IsGlobal: isGlobalSecurity,
			OpFilter: "security",
			Value:    testStep.Security,
		})
	} else if testWorkflow.Security != nil {
		usageContext.Scopes = append(usageContext.Scopes, ast.UsageExampleScope{
			Feature:  "security",
			IsGlobal: true,
			OpFilter: "security",
			Value:    testWorkflow.Security,
		})
	}

	usageContext.PopulateGlobalParameterScopes(usageContext.ExampleName, false) // Tests should not auto-include server selection
	filterUsageContextParameterScopes(usageContext)

	usageContext.Scopes = append(usageContext.Scopes, ast.UsageExampleScope{
		Feature:  "http_client",
		IsGlobal: true,
		Value: map[string]any{
			"type":     "test",
			"testName": testWorkflow.Name,
		},
	})

	return usageContext
}

// filterUsageContextParameterScopes removes any global parameters that aren't hidden to allow the parameters to be set at the operation level for tests
func filterUsageContextParameterScopes(usageContext *ast.UsageContext) {
	scopes := []ast.UsageExampleScope{}

	for _, scope := range usageContext.Scopes {
		if scope.Feature != "parameter" {
			scopes = append(scopes, scope)
			continue
		}

		if scope.Value.(map[string]any)["hidden"].(bool) {
			scopes = append(scopes, scope)
			continue
		}
	}

	usageContext.Scopes = scopes
}
