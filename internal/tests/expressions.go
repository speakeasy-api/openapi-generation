package tests

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math/rand"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/speakeasy-api/openapi/jsonpointer"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type resolveExpressionOpts struct {
	expression         expression.Expression
	source             *ast.TypeDef
	msgContext         string
	expressionNode     *yaml.Node
	testArazzo         *arazzo.Arazzo
	workflow           *arazzo.Workflow
	workflowSteps      []ast.ArazzoStep
	workflowInputs     *ast.FieldDef
	isTopLevelWorkflow bool
}

func resolveExpression(ctx context.Context, opts resolveExpressionOpts) (*yaml.Node, *ast.ExampleReference, string) {
	typ, source, expressionParts, jp := opts.expression.GetParts()

	// TODO renable if we add other expression types that use json-pointers that we don't support
	// if jp != "" {
	// 	msg := fmt.Sprintf("%s with type %s has unsupported json-pointer component", opts.msgContext, opts.expression.GetType())
	// 	logging.LogWarning(ctx, msg, &validation.Error{
	// 		Message: "json-pointer not currently supported on expression",
	// 		Line:    opts.expressionLine,
	// 		Column:  opts.expressionColumn,
	// 	})

	// 	return nil, nil, msg
	// }

	switch typ {
	case expression.ExpressionTypeInputs:
		return resolveInputExpression(ctx, source, opts)
	case expression.ExpressionTypeSteps:
		return resolveStepExpression(ctx, source, expressionParts, jp, opts)
	default:
		msg := fmt.Sprintf("%s with type %s not currently supported", opts.msgContext, opts.expression.GetType())

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("expression type not currently supported"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}
}

func resolveInputExpression(ctx context.Context, expressionSource string, opts resolveExpressionOpts) (*yaml.Node, *ast.ExampleReference, string) {
	workflow := opts.workflow

	if opts.workflowInputs == nil {
		msg := opts.msgContext + " targeting unresolved input schema"

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("input expression targeting unresolved input schema"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}

	if !workflow.Inputs.IsSchema() {
		msg := opts.msgContext + " targeting non-object input schema"

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("input expression targeting non-object input schema"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}

	inputsSchema := workflow.Inputs.GetSchema()

	input, ok := inputsSchema.Properties.Get(expressionSource)
	if !ok {
		msg := fmt.Sprintf("%s referencing non-existent input %s", opts.msgContext, expressionSource)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("input expression referencing non-existent input"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}

	if opts.isTopLevelWorkflow {
		if !input.IsSchema() {
			msg := fmt.Sprintf("%s referencing boolean json schema input %s with no examples", opts.msgContext, expressionSource)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("input expression referencing boolean json schema input"),
				Node:            opts.expressionNode,
			})

			return nil, nil, msg
		}

		examples := input.GetSchema().Examples

		if len(examples) == 0 {
			msg := fmt.Sprintf("%s referencing input %s with no examples", opts.msgContext, expressionSource)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("input expression referencing input with no examples"),
				Node:            opts.expressionNode,
			})

			return nil, nil, msg
		}

		r := rand.New(rand.NewSource(hashString(fmt.Sprintf("%s.%s", opts.msgContext, expressionSource))))
		return examples[r.Intn(len(examples))], nil, ""
	} else {
		fields := opts.workflowInputs.Type.Fields

		inputField := fields.GetField(expressionSource)
		if inputField == nil {
			msg := fmt.Sprintf("%s referencing non-existent input %s", opts.msgContext, expressionSource)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("input expression referencing non-existent input"),
				Node:            opts.expressionNode,
			})

			return nil, nil, msg
		}

		if inputField.Type != nil {
			// TODO we need to check if the type is the same or different
		} else {
			// Set the type to the source type using the input
			// This is assuming the source is the same type as the input
			// This is a workaround to having to calculate the type from the arazzo json schema
			inputField.Type = opts.source.ShallowCopy()
		}

		target := ast.InputTarget{
			Target: inputField,
			Path:   "/" + expressionSource,
			Inputs: opts.workflowInputs,
		}

		return nil, &ast.ExampleReference{
			Target: target,
			Type:   "inputs",
		}, ""
	}
}

func resolveStepExpression(ctx context.Context, stepID string, expressionParts []string, jp jsonpointer.JSONPointer, opts resolveExpressionOpts) (*yaml.Node, *ast.ExampleReference, string) {
	var step *arazzo.Step
	var stepIdx int

	for i, s := range opts.workflow.Steps {
		if s.StepID == stepID {
			step = s
			stepIdx = i
			break
		}
	}

	if step == nil {
		msg := fmt.Sprintf("%s referencing non-existent step %s", opts.msgContext, stepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("expression referencing non-existent step"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}

	if len(expressionParts) == 0 {
		msg := fmt.Sprintf("%s referencing step %s with no expression", opts.msgContext, stepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("expression referencing step with no expression"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}

	switch expressionParts[0] {
	case "outputs":
		if len(expressionParts) != 2 {
			msg := opts.msgContext + " missing outputs id"

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("expression missing outputs id"),
				Node:            opts.expressionNode,
			})

			return nil, nil, msg
		}

		outputID := expressionParts[1]

		return handleOutputExpression(ctx, handleOutputExpressionOpts{
			testArazzo:     opts.testArazzo,
			workflow:       opts.workflow,
			msgContext:     opts.msgContext,
			expressionNode: opts.expressionNode,
			stepIdx:        stepIdx,
			stepID:         stepID,
			step:           step,
			outputID:       outputID,
			workflowSteps:  opts.workflowSteps,
			jp:             jp,
		})
	case "description", "stepId", "operationId", "operationPath", "workflowId", "parameters", "requestBody", "successCriteria", "onSuccess", "onFailure":
		msg := fmt.Sprintf("%s contains unsupported step expression type %s", opts.msgContext, expressionParts[0])

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("expression referencing output with unsupported expression type"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	default:
		return handleOutputExpression(ctx, handleOutputExpressionOpts{
			testArazzo:     opts.testArazzo,
			workflow:       opts.workflow,
			msgContext:     opts.msgContext,
			expressionNode: opts.expressionNode,
			stepIdx:        stepIdx,
			stepID:         stepID,
			step:           step,
			outputID:       expressionParts[0],
			workflowSteps:  opts.workflowSteps,
		})
	}
}

type handleOutputExpressionOpts struct {
	testArazzo     *arazzo.Arazzo
	workflow       *arazzo.Workflow
	msgContext     string
	expressionNode *yaml.Node
	stepIdx        int
	stepID         string
	step           *arazzo.Step
	outputID       string
	jp             jsonpointer.JSONPointer
	workflowSteps  []ast.ArazzoStep
}

func handleOutputExpression(ctx context.Context, opts handleOutputExpressionOpts) (*yaml.Node, *ast.ExampleReference, string) {
	var outputExpression expression.Expression

	for id, o := range opts.step.Outputs.All() {
		if id == opts.outputID {
			outputExpression = o
			break
		}
	}

	if outputExpression == "" {
		msg := fmt.Sprintf("%s references empty output %s", opts.msgContext, opts.outputID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("expression references empty output"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}

	typ, source, expressionParts, jp := outputExpression.GetParts()

	switch typ {
	case expression.ExpressionTypeResponse:
		switch source {
		case "body":
			if opts.stepIdx >= len(opts.workflowSteps) {
				msg := fmt.Sprintf("%s referencing unresolved step %s", opts.msgContext, opts.stepID)
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("expression referencing unresolved step"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			step := opts.workflowSteps[opts.stepIdx]

			if step.GetType() != ast.ArazzoStepTypeOperation {
				msg := opts.msgContext + " referencing workflow step that doesn't support body expressions"
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("expression referencing step that doesn't support body expressions"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			op := step.(*ast.ArazzoOperationStep)

			var responseBodyAssertion *ast.ResponseBodyAssertion

			for _, assertion := range op.UsageContext.Assertions {
				if assertion.TargetType == ast.AssertionTargetResponseBody {
					rba, ok := assertion.Value.(ast.ResponseBodyAssertion)
					if !ok {
						msg := opts.msgContext + " referencing step with invalid response body assertion"
						logging.LogWarning(ctx, msg, &validation.Error{
							UnderlyingError: errors.New("response.body expression referencing step currently needs a response body assertion"),
							Node:            opts.expressionNode,
						})
						return nil, nil, msg
					}
					responseBodyAssertion = &rba
					break
				}
			}

			if responseBodyAssertion == nil {
				msg := opts.msgContext + " referencing step with no response body assertion"
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("response.body expression referencing step currently needs a response body assertion"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			if jp == "" && opts.jp == "" {
				jp = "/"
			}

			if opts.jp != "" {
				jp += opts.jp
			}

			t, err := jsonpointer.GetTarget(responseBodyAssertion.Content.Content, jp)
			if err != nil {
				msg := fmt.Sprintf("%s contains unresolvable json-pointer %s for response body: %s", opts.msgContext, jp, err.Error())
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("json-pointer not currently supported on response body"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			target := ast.ResponseBodyTarget{
				Target:  t,
				Path:    string(jp),
				Body:    responseBodyAssertion.Content,
				StepIdx: opts.stepIdx,
			}

			return nil, &ast.ExampleReference{
				Target: target,
				Type:   "response.body",
			}, ""
		default:
			msg := fmt.Sprintf("%s referencing output %s with unsupported response expression source %s", opts.msgContext, opts.outputID, source)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("expression referencing output with unsupported response expression source"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}
	case expression.ExpressionTypeWorkflows:
		if opts.step.WorkflowID == nil {
			msg := fmt.Sprintf("%s referencing output %s has workflow expression in non-workflow step", opts.msgContext, opts.outputID)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("workflow expression in non-workflow step"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		if source != string(*opts.step.WorkflowID) {
			msg := fmt.Sprintf("%s referencing output %s references workflow %s but is not the current workflow", opts.msgContext, opts.outputID, source)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("output expression not referencing current workflow"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		targetArazzoWorkflow := opts.testArazzo.Workflows.Find(source)
		if targetArazzoWorkflow == nil {
			msg := fmt.Sprintf("%s referencing output %s references non-existent workflow %s", opts.msgContext, opts.outputID, source)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("output expression referencing non-existent workflow"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		if opts.stepIdx >= len(opts.workflowSteps) {
			msg := fmt.Sprintf("%s referencing output %s is referencing unresolved step %s", opts.msgContext, opts.outputID, source)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("output expression referencing unresolved step"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		targetWorkflowStep, ok := opts.workflowSteps[opts.stepIdx].(*ast.ArazzoWorkflowStep)
		if !ok {
			msg := fmt.Sprintf("%s referencing output %s is not a workflow step", opts.msgContext, opts.outputID)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("output expression is not a workflow step"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		targetWorkflow := targetWorkflowStep.Workflow
		if targetWorkflow == nil {
			msg := fmt.Sprintf("%s referencing output %s has unresolved workflow step %s", opts.msgContext, opts.outputID, source)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("output expression referencing unresolved workflow step"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		if len(expressionParts) == 0 {
			msg := fmt.Sprintf("%s referencing output %s does not contain expected expression parts", opts.msgContext, opts.outputID)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("invalid output expression"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}

		switch expressionParts[0] {
		case "outputs":
			if len(expressionParts) != 2 {
				msg := fmt.Sprintf("%s referencing output %s has output expression missing target", opts.msgContext, opts.outputID)
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("invalid output expression"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			outputTarget := expressionParts[1]

			if !targetArazzoWorkflow.Outputs.Has(outputTarget) {
				msg := fmt.Sprintf("%s referencing output %s has output expression referencing non-existent output %s", opts.msgContext, opts.outputID, outputTarget)
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("invalid output expression"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			jp = jsonpointer.JSONPointer("/"+outputTarget) + jp

			if opts.jp != "" {
				jp += opts.jp
			}

			t, err := jsonpointer.GetTarget(targetWorkflow.Outputs, jp)
			if err != nil {
				msg := fmt.Sprintf("%s contains unresolvable json-pointer %s for response body: %s", opts.msgContext, jp, err.Error())
				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("json-pointer not currently supported on response body"),
					Node:            opts.expressionNode,
				})
				return nil, nil, msg
			}

			target := ast.OutputTarget{
				Target:  t,
				Path:    string(jp),
				Outputs: targetWorkflow.Outputs,
				StepIdx: opts.stepIdx,
			}

			return nil, &ast.ExampleReference{
				Target: target,
				Type:   "outputs",
			}, ""
		default:
			msg := fmt.Sprintf("%s referencing output %s has unsupported expression type %s", opts.msgContext, opts.outputID, expressionParts[0])
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("unsupported workflow expression type"),
				Node:            opts.expressionNode,
			})
			return nil, nil, msg
		}
	default:
		msg := fmt.Sprintf("%s referencing output %s with unsupported expression type %s", opts.msgContext, opts.outputID, typ)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("expression referencing output with unsupported expression type"),
			Node:            opts.expressionNode,
		})

		return nil, nil, msg
	}
}

func hashString(s string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return int64(h.Sum64())
}
