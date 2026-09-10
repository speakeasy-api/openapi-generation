package arazzo

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/examples"
	genTests "github.com/speakeasy-api/openapi-generation/v2/internal/tests"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/json"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

type workflowsFromTestsOptions struct {
	AST                *ast.AST
	Tests              *sequencedmap.Map[string, []test]
	RegenerateTests    map[string]bool
	ArazzoDoc          *arazzo.Arazzo
	OpenAPIDocPath     string
	LockFile           *config.LockFile
	Config             *configuration.Config
	IsInternalTestSpec bool
}

func workflowsFromTests(ctx context.Context, opts workflowsFromTestsOptions) error {
	operationsById := genTests.GetOperationsByID(opts.AST, &genTests.GetOperationsByIDOptions{
		IncludeWebhooks: false,
	})

	for operationID, tts := range opts.Tests.All() {
		for _, test := range tts {
			// Find the request body type and then the matching operation from the AST
			var requestBody *arazzo.RequestBody

			requestBodyContentType := ""

			if test.RequestBody.Len() > 0 {
				for contentType, content := range test.RequestBody.All() {
					sanitizedContent, err := roundTripYamlNode(&content)
					if err != nil {
						return err
					}

					requestBodyContentType = contentType

					requestBody = &arazzo.RequestBody{
						ContentType: pointer.From(requestBodyContentType),
						Payload:     sanitizedContent,
					}
				}
			}

			ops, ok := operationsById[operationID]
			if !ok || len(ops) == 0 {
				continue
			}

			var op *ast.Operation
			for _, o := range ops {
				if requestBodyContentType == "" {
					op = o
					break
				}
				if o.Request != nil && slices.Contains(o.Request.MatchedContentTypes, requestBodyContentType) {
					op = o
					break
				}
			}
			if op == nil {
				logging.From(ctx).Debug(fmt.Sprintf("skipping test %s as no operation matches request body content type %s", test.Name, requestBodyContentType))
				continue
			}

			workflowID, existingWorkflowIdx := generateWorkflowID(op, test, opts.ArazzoDoc)

			// Now determine if we are even going to bootstrap the test
			skipBootstrappingTest := false

			_, alreadyGenerated := opts.LockFile.GeneratedTests.Get(workflowID)
			// If we haven't already generated the test before
			if !alreadyGenerated {
				regenerateTest, ok := opts.RegenerateTests[workflowID]

				generateNewTests := opts.Config.Generation.Tests.GenerateNewTests

				if ok && !regenerateTest {
					// If the test does exist but we are not regenerating the test then skip bootstrapping the test
					skipBootstrappingTest = true
				} else if !ok && !generateNewTests {
					// If the test doesn't already exist and we are not generating new tests then skip bootstrapping the test
					skipBootstrappingTest = true
				}
			} else {
				// If we have already generated the test and it hasn't been marked for regeneration then skip bootstrapping the test
				if regenerateTest := opts.RegenerateTests[workflowID]; !regenerateTest {
					skipBootstrappingTest = true
				}
			}

			if skipBootstrappingTest {
				continue
			}

			var description *string
			if test.Description != "" {
				description = pointer.From(test.Description)
			}

			params := []*arazzo.ReusableParameter{}

			if test.Parameters != nil {
				if test.Parameters.Path.Len() > 0 {
					for param, value := range test.Parameters.Path.All() {
						sanitizedValue, err := roundTripYamlNode(&value)
						if err != nil {
							return err
						}

						params = append(params, &arazzo.ReusableParameter{
							Object: &arazzo.Parameter{
								Name:  param,
								In:    pointer.From(arazzo.InPath),
								Value: sanitizedValue,
							},
						})
					}
				}
				if test.Parameters.Query != nil {
					for param, value := range test.Parameters.Query.All() {
						sanitizedValue, err := roundTripYamlNode(&value)
						if err != nil {
							return err
						}

						params = append(params, &arazzo.ReusableParameter{
							Object: &arazzo.Parameter{
								Name:  param,
								In:    pointer.From(arazzo.InQuery),
								Value: sanitizedValue,
							},
						})
					}
				}
				if test.Parameters.Header != nil {
					for param, value := range test.Parameters.Header.All() {
						sanitizedValue, err := roundTripYamlNode(&value)
						if err != nil {
							return err
						}

						params = append(params, &arazzo.ReusableParameter{
							Object: &arazzo.Parameter{
								Name:  param,
								In:    pointer.From(arazzo.InHeader),
								Value: sanitizedValue,
							},
						})
					}
				}
			}

			assertions := []*criterion.Criterion{}

			for statusCode := range test.Responses.All() {
				contentTypes, assertStatusCode, err := test.GetResponse(statusCode)
				if err != nil {
					continue
				}
				if contentTypes.Len() > 0 {
					for contentType, content := range contentTypes.All() {
						assertions = append(assertions, &criterion.Criterion{
							Condition: "$statusCode == " + statusCode,
						})
						assertions = append(assertions, &criterion.Criterion{
							Condition: "$response.header.Content-Type == " + contentType,
						})

						out := bytes.NewBuffer([]byte{})
						if err := json.YAMLToJSON(&content, 2, out); err != nil {
							return fmt.Errorf("failed to convert response body to JSON: %w", err)
						}

						assertions = append(assertions, &criterion.Criterion{
							Type: criterion.CriterionTypeUnion{
								Type: pointer.From(criterion.CriterionTypeSimple),
							},
							Context:   pointer.From(expression.Expression("$response.body")),
							Condition: out.String(),
						})
					}
				} else if assertStatusCode || contentTypes.Len() == 0 {
					assertions = append(assertions, &criterion.Criterion{
						Condition: "$statusCode == " + statusCode,
					})
				}
			}

			if len(params) == 0 {
				params = nil
			}

			step := &arazzo.Step{
				StepID:          "test",
				RequestBody:     requestBody,
				Parameters:      params,
				OperationID:     pointer.From(expression.Expression(operationID)),
				SuccessCriteria: assertions,
			}

			workflowExtElems := []*extensions.Element{}
			stepExtElems := []*extensions.Element{}

			testGroupExt := genTests.NewTestGroupExtension(op.OwningSDK.Type.Name)
			if testGroupExt != nil {
				workflowExtElems = append(workflowExtElems, testGroupExt)
			}

			if test.Server != "" {
				serverExt := genTests.NewTestServerExtension(&genTests.TestServer{
					BaseURL: test.Server,
				})
				if serverExt != nil {
					workflowExtElems = append(workflowExtElems, serverExt)
				}
			}

			if len(test.Security.Content) > 0 || test.Security.Value != "" {
				securityExt := genTests.NewTestSecurityExtension(&genTests.TestSecurity{
					Value: test.Security,
				})
				if securityExt != nil {
					if op.Security != nil {
						stepExtElems = append(stepExtElems, securityExt)
					} else {
						workflowExtElems = append(workflowExtElems, securityExt)
					}
				}
			}

			if len(test.Targets) > 0 {
				targetsExt := genTests.NewTestTargetsExtension(test.Targets)
				if targetsExt != nil {
					workflowExtElems = append(workflowExtElems, targetsExt)
				}
			}

			if test.InternalID != "" {
				internalIDExt := genTests.NewInternalIDExtension(test.InternalID)
				if internalIDExt != nil {
					workflowExtElems = append(workflowExtElems, internalIDExt)
				}
			}

			if len(test.TestGroups) > 0 {
				internalTestGroupsExt := genTests.NewInternalTestGroupsExtension(test.TestGroups)
				if internalTestGroupsExt != nil {
					workflowExtElems = append(workflowExtElems, internalTestGroupsExt)
				}
			}

			if test.InternalEnvVars.Len() > 0 {
				internalEnvVars := []ast.TestEnvVar{}

				for key, val := range test.InternalEnvVars.All() {
					internalEnvVars = append(internalEnvVars, ast.TestEnvVar{
						Name:  key,
						Value: val,
					})
				}

				internalEnvVarsExt := genTests.NewInternalEnvVarsExtension(internalEnvVars)
				if internalEnvVarsExt != nil {
					workflowExtElems = append(workflowExtElems, internalEnvVarsExt)
				}
			}

			step.Extensions = extensions.New(stepExtElems...)

			// All new tests that are bootstrapped by us will be marked for regeneration
			var regenerateExtensionNode yaml.Node
			if err := regenerateExtensionNode.Encode(true); err != nil {
				return nil
			}
			workflowExtElems = append(workflowExtElems, extensions.NewElem(ExtTestRebuild, pointer.From(regenerateExtensionNode)))

			newWorkflow := &arazzo.Workflow{
				WorkflowID:  workflowID,
				Description: description,
				Extensions:  extensions.New(workflowExtElems...),
				Steps:       []*arazzo.Step{step},
			}

			if existingWorkflowIdx >= 0 {
				opts.ArazzoDoc.Workflows[existingWorkflowIdx] = newWorkflow
			} else {
				opts.ArazzoDoc.Workflows = append(opts.ArazzoDoc.Workflows, newWorkflow)
			}

			// Record the test as generated, we only record the first time we generate it
			if _, ok := opts.LockFile.GeneratedTests.Get(workflowID); !ok {
				opts.LockFile.GeneratedTests.Set(workflowID, time.Now().Format(time.RFC3339))
			}
		}
	}

	return nil
}

func generateWorkflowID(op *ast.Operation, test test, a *arazzo.Arazzo) (string, int) {
	workflowID := op.ID
	if test.Name != "" && test.Name != "test" && !examples.IsDefaultExample(test.Name) {
		workflowID += "-" + test.Name
	}

	for {
		// First check if the workflow with the exact name match already exists
		existingWorkflowIdx := slices.IndexFunc(a.Workflows, func(w *arazzo.Workflow) bool {
			return w.WorkflowID == workflowID
		})

		if existingWorkflowIdx >= 0 {
			return workflowID, existingWorkflowIdx
		}

		// If not then check if there is a workflow with the same name just different casing and if so increment the number suffix until we don't have a conflict
		conflictIdx := slices.IndexFunc(a.Workflows, func(w *arazzo.Workflow) bool {
			return strings.EqualFold(w.WorkflowID, workflowID)
		})

		if conflictIdx == -1 {
			break
		}

		// parse number suffix and increment
		parts := strings.Split(workflowID, "-")

		suffixStr := "0"
		if len(parts) > 1 {
			suffixStr = parts[len(parts)-1]
		}
		if len(suffixStr) == 0 {
			suffixStr = "0"
		}

		prefix := parts[0]

		suffixNum, err := strconv.Atoi(suffixStr)
		if err != nil {
			suffixNum = 0
		}
		workflowID = fmt.Sprintf("%s-%d", prefix, suffixNum+1)
	}

	return workflowID, -1
}

func roundTripYamlNode(node *yaml.Node) (*yaml.Node, error) {
	var out any
	if err := node.Decode(&out); err != nil {
		return nil, err
	}

	var outNode yaml.Node
	if err := outNode.Encode(out); err != nil {
		return nil, err
	}

	return &outNode, nil
}
