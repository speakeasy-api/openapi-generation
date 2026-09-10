package tests

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/speakeasy-api/openapi/jsonpointer"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type handleStepOpts struct {
	ast                   *ast.AST
	testArazzo            *arazzo.Arazzo
	target                types.Target
	internalTestGroupName string
	subsystem             *subsystem.Subsystem
	workflow              *arazzo.Workflow
	stepIdx               int
	step                  *arazzo.Step
	operationsByID        map[string][]*ast.Operation
	previousSteps         []ast.ArazzoStep
	workflowInputs        *ast.FieldDef
	isTopLevelWorkflow    bool
	docInfo               *ArazzoDocumentInfo
}

func handleStep(ctx context.Context, opts handleStepOpts) (ast.ArazzoStep, []string) {
	step := opts.step

	if step.WorkflowID != nil {
		return handleWorkflowStep(ctx, opts)
	}

	return handleOperationStep(ctx, opts)
}

func handleOperationStep(ctx context.Context, opts handleStepOpts) (ast.ArazzoStep, []string) {
	incompleteMessages := []string{}

	step := opts.step
	workflow := opts.workflow

	if step.OperationPath != nil {
		msg := fmt.Sprintf("workflow step %s.%s referencing an operationPath not currently supported", workflow.WorkflowID, step.StepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("workflow step referencing an operationPath not currently supported"),
			Node:            step.GetCore().OperationPath.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	if step.OperationID.IsExpression() {
		msg := fmt.Sprintf("workflow step %s.%s with operationId using runtime expression not currently supported", workflow.WorkflowID, step.StepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported step feature"),
			Node:            step.GetCore().OperationPath.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	operationID := string(*step.OperationID)

	ops, ok := opts.operationsByID[operationID]
	if !ok || len(ops) == 0 {
		msg := fmt.Sprintf("workflow step %s.%s referencing operation %s not found in document", workflow.WorkflowID, step.StepID, operationID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("operation not found"),
			Node:            step.GetCore().OperationPath.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)

		return nil, incompleteMessages
	}

	// TODO in the future we may want to allow selecting which SDK to use for the test via an extensions
	var op *ast.Operation
	for _, o := range ops {
		if step.RequestBody == nil || step.RequestBody.ContentType == nil {
			op = o
			break
		}
		if o.Request != nil && slices.Contains(o.Request.MatchedContentTypes, *step.RequestBody.ContentType) {
			op = o
			break
		}
	}
	if op == nil {
		msg := fmt.Sprintf("workflow step %s.%s unable to find matching operation", workflow.WorkflowID, step.StepID)
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unable to find matching operation"),
			Node:            step.GetCore().OperationPath.GetKeyNodeOrRoot(step.GetRootNode()),
		})
		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	if !IsOperationSupported(op, opts.subsystem.Config) {
		msg := fmt.Sprintf("workflow step %s.%s referencing operation %s is not currently supported", workflow.WorkflowID, step.StepID, operationID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("operation not currently supported"),
			Node:            workflow.GetCore().Steps.GetSliceValueNodeOrRoot(opts.stepIdx, workflow.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	isRequestBodyRequired := isOpRequestBodyRequired(op)

	if isRequestBodyRequired && step.RequestBody == nil {
		msg := fmt.Sprintf("workflow step %s.%s referencing operation %s is missing required request body", workflow.WorkflowID, step.StepID, operationID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("missing required request body"),
			Node:            workflow.GetCore().Steps.GetSliceValueNodeOrRoot(opts.stepIdx, workflow.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
	} else if step.RequestBody != nil {
		incompleteMessages = append(incompleteMessages, handleStepRequestBody(ctx, handleStepRequestBodyOpts{
			testArazzo:         opts.testArazzo,
			workflow:           workflow,
			step:               step,
			stepIdx:            opts.stepIdx,
			op:                 op,
			previousSteps:      opts.previousSteps,
			incompleteMessages: incompleteMessages,
			workflowInputs:     opts.workflowInputs,
			isTopLevelWorkflow: opts.isTopLevelWorkflow,
		})...)
	}

	incompleteMessages = append(incompleteMessages, handleOperationParameters(ctx, handleOperationParameterOpts{
		testArazzo:         opts.testArazzo,
		workflow:           workflow,
		step:               step,
		stepIdx:            opts.stepIdx,
		op:                 op,
		previousSteps:      opts.previousSteps,
		workflowInputs:     opts.workflowInputs,
		isTopLevelWorkflow: opts.isTopLevelWorkflow,
	})...)

	assertions, responseContentType, aIncompleteMessages := getOpAssertions(ctx, getOpAssertionsOpts{
		workflow: workflow,
		step:     step,
		stepIdx:  opts.stepIdx,
		op:       op,
	})
	incompleteMessages = append(incompleteMessages, aIncompleteMessages...)

	if len(incompleteMessages) > 0 {
		return nil, incompleteMessages
	}

	securityVal, im := handleSecurity(ctx, handleSecurityOpts{
		testArazzo:    opts.testArazzo,
		workflow:      opts.workflow,
		testSecurity:  getStepTestSecurity(ctx, step),
		workflowSteps: opts.previousSteps,
		msgContext:    fmt.Sprintf("workflow step %s.%s referencing operation %s", opts.workflow.WorkflowID, opts.step.StepID, *opts.step.OperationID),
		exampleName:   fmt.Sprintf("%s[%d]", opts.workflow.WorkflowID, opts.stepIdx),
	})
	if len(im) > 0 {
		incompleteMessages = append(incompleteMessages, im...)
		return nil, incompleteMessages
	}

	usageContext := ast.CreateUsageContext(op.OwningSDK, op, op.Extensions.UsageExample, opts.subsystem.Config.Generation.Tests.SkipResponseBodyAssertions)
	usageContext.Assertions = assertions

	return &ast.ArazzoOperationStep{
		ArazzoStepBase: ast.ArazzoStepBase{
			Type:   string(ast.ArazzoStepTypeOperation),
			StepID: opts.step.StepID,
		},
		Invocation: &ast.ArazzoInvocationContext{
			SDK:       op.OwningSDK,
			Operation: op,
			StepIdx:   opts.stepIdx,
			StepID:    opts.step.StepID,
		},
		UsageContext:        usageContext,
		Operation:           op,
		StepIdx:             opts.stepIdx,
		ResponseContentType: responseContentType,
		Security:            securityVal,
	}, nil
}

func handleWorkflowStep(ctx context.Context, opts handleStepOpts) (ast.ArazzoStep, []string) {
	incompleteMessages := []string{}

	step := opts.step
	workflow := opts.workflow

	if step.WorkflowID.IsExpression() {
		msg := fmt.Sprintf("workflow step %s.%s with workflowId using runtime expression not currently supported", workflow.WorkflowID, step.StepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported step workflowId expression"),
			Node:            step.GetCore().WorkflowID.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	referencedWorkflow := opts.testArazzo.Workflows.Find(string(*step.WorkflowID))
	if referencedWorkflow == nil {
		msg := fmt.Sprintf("workflow step %s.%s referencing non-existent workflow %s", workflow.WorkflowID, step.StepID, *step.WorkflowID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("workflow not found"),
			Node:            step.GetCore().WorkflowID.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	_, w, incompleteMessages := handleWorkflow(ctx, handleWorkflowOpts{
		ast:                   opts.ast,
		testArazzo:            opts.testArazzo,
		workflow:              referencedWorkflow,
		topLevel:              false,
		operationsByID:        opts.operationsByID,
		target:                opts.target,
		internalTestGroupName: opts.internalTestGroupName,
		subsystem:             opts.subsystem,
		stepID:                step.StepID,
		isTopLevelWorkflow:    false,
		docInfo:               opts.docInfo,
	})
	if len(incompleteMessages) > 0 {
		return nil, incompleteMessages
	}
	if w == nil {
		msg := fmt.Sprintf("unabled to handle step %s.%s referencing workflow %s", workflow.WorkflowID, step.StepID, *step.WorkflowID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("workflow invalid"),
			Node:            step.GetCore().WorkflowID.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, incompleteMessages
	}

	incompleteMessages = append(incompleteMessages, handleWorkflowParameters(ctx, handleWorkflowParametersOpts{
		testArazzo:         opts.testArazzo,
		workflow:           opts.workflow,
		stepIdx:            opts.stepIdx,
		step:               step,
		previousSteps:      opts.previousSteps,
		workflowInputs:     w.Inputs,
		isTopLevelWorkflow: false,
	})...)

	if len(incompleteMessages) > 0 {
		return nil, incompleteMessages
	}

	return &ast.ArazzoWorkflowStep{
		ArazzoStepBase: ast.ArazzoStepBase{
			Type:   string(ast.ArazzoStepTypeWorkflow),
			StepID: step.StepID,
		},
		StepIdx:    opts.stepIdx,
		WorkflowID: string(*step.WorkflowID),
		Workflow:   w,
	}, nil
}

type handleStepRequestBodyOpts struct {
	testArazzo         *arazzo.Arazzo
	workflow           *arazzo.Workflow
	step               *arazzo.Step
	stepIdx            int
	op                 *ast.Operation
	previousSteps      []ast.ArazzoStep
	incompleteMessages []string
	workflowInputs     *ast.FieldDef
	isTopLevelWorkflow bool
}

func handleStepRequestBody(ctx context.Context, opts handleStepRequestBodyOpts) []string {
	workflow := opts.workflow
	step := opts.step
	op := opts.op

	stepContentType := ""
	if len(op.Request.MatchedContentTypes) > 0 {
		stepContentType = op.Request.MatchedContentTypes[0]
	}
	if step.RequestBody.ContentType != nil {
		stepContentType = *step.RequestBody.ContentType
	}

	var example *yaml.Node
	var reference *ast.ExampleReference
	replacements := []*ast.ExampleReplacement{}

	value, exp, err := expression.GetValueOrExpressionValue(step.RequestBody.Payload)
	if err != nil {
		msg := fmt.Sprintf("workflow step %s.%s referencing operation %s has invalid request body payload: %s", workflow.WorkflowID, step.StepID, *step.OperationID, err.Error())
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("invalid request body payload"),
			Node:            step.GetCore().RequestBody.GetKeyNodeOrRoot(step.GetRootNode()),
		})
		opts.incompleteMessages = append(opts.incompleteMessages, msg)
		return opts.incompleteMessages
	}

	if exp != nil {
		var incompleteMessage string
		example, reference, incompleteMessage = resolveExpression(ctx, resolveExpressionOpts{
			expression:         *exp,
			source:             opts.op.Request.RequestBody.Type,
			msgContext:         fmt.Sprintf("workflow step %s.%s with expression in requestBody", opts.workflow.WorkflowID, opts.step.StepID),
			expressionNode:     step.RequestBody.GetCore().Payload.GetKeyNodeOrRoot(step.RequestBody.GetRootNode()),
			testArazzo:         opts.testArazzo,
			workflow:           opts.workflow,
			workflowSteps:      opts.previousSteps,
			workflowInputs:     opts.workflowInputs,
			isTopLevelWorkflow: opts.isTopLevelWorkflow,
		})
		if incompleteMessage != "" {
			opts.incompleteMessages = append(opts.incompleteMessages, incompleteMessage)
			return opts.incompleteMessages
		}

		foundErr := false

		for i, replacement := range step.RequestBody.Replacements {
			val, e, err := expression.GetValueOrExpressionValue(replacement.Value)
			if err != nil {
				msg := fmt.Sprintf("workflow step %s.%s referencing operation %s has invalid request body replacement value: %s", workflow.WorkflowID, step.StepID, *step.OperationID, err.Error())
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("invalid request body replacement value"),
					Node:            step.RequestBody.GetCore().Replacements.GetSliceValueNodeOrRoot(i, step.RequestBody.GetRootNode()),
				})
				opts.incompleteMessages = append(opts.incompleteMessages, msg)
				foundErr = true
				continue
			}

			source, err := jsonpointer.GetTarget(opts.op.Request.RequestBody.Type, replacement.Target)
			if err != nil {
				msg := fmt.Sprintf("workflow step %s.%s referencing operation %s has invalid request body replacement target: %s", workflow.WorkflowID, step.StepID, *step.OperationID, err.Error())
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("invalid request body replacement target"),
					Node:            step.RequestBody.GetCore().Replacements.GetSliceValueNodeOrRoot(i, step.RequestBody.GetRootNode()),
				})
				opts.incompleteMessages = append(opts.incompleteMessages, msg)
				foundErr = true
				continue
			}

			var sourceType *ast.TypeDef

			sourceField, ok := source.(*ast.FieldDef)
			if !ok {
				sourceType, ok = source.(*ast.TypeDef)
				if !ok {
					msg := fmt.Sprintf("workflow step %s.%s referencing operation %s has invalid request body replacement target: %s", workflow.WorkflowID, step.StepID, *step.OperationID, err.Error())
					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("invalid request body replacement target"),
						Node:            step.RequestBody.GetCore().Replacements.GetSliceValueNodeOrRoot(i, step.RequestBody.GetRootNode()),
					})
					opts.incompleteMessages = append(opts.incompleteMessages, msg)
					foundErr = true
					continue
				}
			} else {
				sourceType = sourceField.Type
			}

			var reference *ast.ExampleReference

			if e != nil {
				var im string
				val, reference, im = resolveExpression(ctx, resolveExpressionOpts{
					expression:         *e,
					source:             sourceType,
					msgContext:         fmt.Sprintf("workflow step %s.%s with expression in requestBody", opts.workflow.WorkflowID, opts.step.StepID),
					expressionNode:     step.RequestBody.GetCore().Replacements.GetSliceValueNodeOrRoot(i, step.RequestBody.GetRootNode()),
					testArazzo:         opts.testArazzo,
					workflow:           opts.workflow,
					workflowSteps:      opts.previousSteps,
					workflowInputs:     opts.workflowInputs,
					isTopLevelWorkflow: opts.isTopLevelWorkflow,
				})
				if im != "" {
					opts.incompleteMessages = append(opts.incompleteMessages, im)
					foundErr = true
					continue
				}
			}

			var example *ast.Example
			if val != nil {
				example = ast.NewExample("", "", val)
			} else {
				example = ast.NewExampleReference("", "", reference)
			}

			replacements = append(replacements, &ast.ExampleReplacement{
				Path:  string(replacement.Target),
				Value: example,
			})
		}

		if foundErr {
			return opts.incompleteMessages
		}
	} else {
		switch value.Kind {
		case yaml.ScalarNode:
			// Check if the node contains json in string form and if so parse it and get as a yaml node
			if contenttypes.IsJSON(stepContentType) && value.Tag == "!!str" {
				isJson := false

				switch {
				case strings.HasPrefix(value.Value, "{"):
					fallthrough
				case strings.HasPrefix(value.Value, "["):
					fallthrough
				case strings.HasPrefix(value.Value, "\""):
					fallthrough
				case value.Value == "true", value.Value == "false":
					isJson = true
				}

				// Ignoring numbers now as they should be handled correctly anyway if they are in string format
				if isJson {
					var node yaml.Node
					if err := yaml.Unmarshal([]byte(value.Value), &node); err == nil {
						example = &node
					}
				} else {
					example = value
				}
			} else {
				example = value
			}
		default:
			var im []string
			example, replacements, im = resolvePayloadReplacements(ctx, resolvePayloadReplacementsOpts{
				node:          value,
				testArazzo:    opts.testArazzo,
				workflow:      opts.workflow,
				msgContext:    fmt.Sprintf("workflow step %s.%s referencing operation %s", opts.workflow.WorkflowID, opts.step.StepID, *opts.step.OperationID),
				workflowSteps: opts.previousSteps,
				nodeContext:   "requestBody",
			})
			if len(im) > 0 {
				return append(opts.incompleteMessages, im...)
			}
		}
	}

	switch {
	case example == nil && reference == nil:
		msg := fmt.Sprintf("workflow step %s.%s referencing operation %s has invalid request body", workflow.WorkflowID, step.StepID, *step.OperationID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("invalid request body"),
			Node:            step.GetCore().RequestBody.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		opts.incompleteMessages = append(opts.incompleteMessages, msg)
	case example != nil:
		ex := ast.NewExample(fmt.Sprintf("%s[%d]", workflow.WorkflowID, opts.stepIdx), "", example)
		ex.Replacements = replacements

		op.Request.Examples = op.Request.Examples.AppendExample(ex)
	case reference != nil:
		ex := ast.NewExampleReference(fmt.Sprintf("%s[%d]", workflow.WorkflowID, opts.stepIdx), "", reference)
		ex.Replacements = replacements

		op.Request.Examples = op.Request.Examples.AppendExample(ex)
	}

	return opts.incompleteMessages
}
