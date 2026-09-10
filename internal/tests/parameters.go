package tests

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/speakeasy-api/openapi/validation"
)

type handleOperationParameterOpts struct {
	testArazzo         *arazzo.Arazzo
	workflow           *arazzo.Workflow
	stepIdx            int
	step               *arazzo.Step
	op                 *ast.Operation
	previousSteps      []ast.ArazzoStep
	workflowInputs     *ast.FieldDef
	isTopLevelWorkflow bool
}

func handleOperationParameters(ctx context.Context, opts handleOperationParameterOpts) []string {
	op := opts.op

	requiredParams := getOpRequiredParams(op)

	incompleteMessages := []string{}

	getMissingParamMsg := func(paramType, param string) {
		stepNode := opts.workflow.GetCore().Steps.GetSliceValueNodeOrRoot(opts.stepIdx, opts.workflow.GetRootNode())

		msg := fmt.Sprintf("workflow step %s.%s referencing operation %s missing required %s parameter %s", opts.workflow.WorkflowID, opts.step.StepID, op.ID, paramType, param)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: fmt.Errorf("missing required %s parameter", paramType),
			Node:            stepNode,
		})

		incompleteMessages = append(incompleteMessages, msg)
	}

	if op.Request == nil || op.Request.Params == nil {
		if requiredParams == nil {
			return nil
		}

		for _, param := range requiredParams.Path {
			getMissingParamMsg("path", param)
		}

		for _, param := range requiredParams.Query {
			getMissingParamMsg("query", param)
		}

		for _, param := range requiredParams.Header {
			getMissingParamMsg("header", param)
		}

		return incompleteMessages
	}

	foundPathParams := map[string]bool{}
	foundQueryParams := map[string]bool{}
	foundHeaderParams := map[string]bool{}

	parameters := getStepParameters(opts.testArazzo, opts.workflow, opts.step)

	for _, rp := range parameters {
		param := rp.Get(opts.testArazzo.Components)

		if param.In == nil {
			msg := fmt.Sprintf("workflow step %s.%s with parameter %s with no `in` value not currently supported", opts.workflow.WorkflowID, opts.step.StepID, param.Name)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("unsupported step feature"),
				Node:            param.GetCore().Name.GetKeyNodeOrRoot(param.GetRootNode()),
			})

			incompleteMessages = append(incompleteMessages, msg)
			continue
		}

		val, expression, err := expression.GetValueOrExpressionValue(param.Value)
		if err != nil {
			msg := fmt.Sprintf("workflow step %s.%s with parameter %s failed to get value", opts.workflow.WorkflowID, opts.step.StepID, param.Name)

			logging.LogWarning(ctx, msg, err)

			incompleteMessages = append(incompleteMessages, msg)
			continue
		}

		handleParamFunc := func(paramType string, params []*ast.Param, found map[string]bool) bool { //nolint:unparam
			foundIdx := slices.IndexFunc(params, func(e *ast.Param) bool {
				return e.Field.OriginalName == param.Name
			})
			if foundIdx == -1 {
				msg := fmt.Sprintf("workflow step %s.%s referencing operation %s does not contain %s parameter %s", opts.workflow.WorkflowID, opts.step.StepID, op.ID, paramType, param.Name)

				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: fmt.Errorf("%s parameter not found", paramType),
					Node:            param.GetCore().Name.GetKeyNodeOrRoot(param.GetRootNode()),
				})

				incompleteMessages = append(incompleteMessages, msg)
				return false
			}

			found[param.Name] = true

			p := params[foundIdx]

			var reference *ast.ExampleReference

			if expression != nil {
				var expIncompleteMessage string
				val, reference, expIncompleteMessage = resolveExpression(ctx, resolveExpressionOpts{
					expression:         *expression,
					source:             p.Field.Type,
					msgContext:         fmt.Sprintf("workflow step %s.%s with expression in parameter", opts.workflow.WorkflowID, opts.step.StepID),
					expressionNode:     param.GetCore().Value.GetKeyNodeOrRoot(param.GetRootNode()),
					testArazzo:         opts.testArazzo,
					workflow:           opts.workflow,
					workflowSteps:      opts.previousSteps,
					workflowInputs:     opts.workflowInputs,
					isTopLevelWorkflow: opts.isTopLevelWorkflow,
				})
				if expIncompleteMessage != "" {
					incompleteMessages = append(incompleteMessages, expIncompleteMessage)
					return false
				}
			}

			exampleName := fmt.Sprintf("%s[%d]", opts.workflow.WorkflowID, opts.stepIdx)

			if reference != nil {
				p.Examples = params[foundIdx].Examples.AppendExample(ast.NewExampleReference(exampleName, "", reference))
			} else {
				p.Examples = params[foundIdx].Examples.AppendExample(ast.NewExample(exampleName, "", val))
			}

			return true
		}

		switch *param.In {
		case arazzo.InPath:
			if !handleParamFunc("path", op.Request.Params.PathParams, foundPathParams) {
				continue
			}
		case arazzo.InQuery:
			if !handleParamFunc("query", op.Request.Params.QueryParams, foundQueryParams) {
				continue
			}
		case arazzo.InHeader:
			if !handleParamFunc("header", op.Request.Params.HeaderParams, foundHeaderParams) {
				continue
			}
		}
	}

	if requiredParams != nil {
		for _, param := range requiredParams.Path {
			if !foundPathParams[param] {
				getMissingParamMsg("path", param)
			}
		}

		for _, param := range requiredParams.Query {
			if !foundQueryParams[param] {
				getMissingParamMsg("query", param)
			}
		}

		for _, param := range requiredParams.Header {
			if !foundHeaderParams[param] {
				getMissingParamMsg("header", param)
			}
		}
	}

	return incompleteMessages
}

type handleWorkflowParametersOpts struct {
	testArazzo         *arazzo.Arazzo
	workflow           *arazzo.Workflow
	stepIdx            int
	step               *arazzo.Step
	previousSteps      []ast.ArazzoStep
	workflowInputs     *ast.FieldDef
	isTopLevelWorkflow bool
}

func handleWorkflowParameters(ctx context.Context, opts handleWorkflowParametersOpts) []string {
	incompleteMessages := []string{}

	step := opts.step
	if len(step.Parameters) == 0 {
		return nil
	}

	if opts.workflowInputs == nil {
		msg := fmt.Sprintf("workflow step %s.%s targeting unresolved input schema", opts.workflow.WorkflowID, step.StepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unresolved input schema"),
			Node:            step.GetCore().Parameters.GetKeyNodeOrRoot(step.GetRootNode()),
		})
	}

	requiredParams := []string{}

	for _, field := range opts.workflowInputs.Type.Fields {
		if !field.Optional {
			requiredParams = append(requiredParams, field.Name)
		}
	}

	parameters := getStepParameters(opts.testArazzo, opts.workflow, opts.step)

	foundParams := map[string]bool{}

	for _, rp := range parameters {
		param := rp.Get(opts.testArazzo.Components)

		pIDx := slices.IndexFunc(opts.workflowInputs.Type.Fields, func(f *ast.FieldDef) bool {
			return f.OriginalName == param.Name
		})

		if pIDx == -1 {
			msg := fmt.Sprintf("workflow step %s.%s with parameter %s referencing non-existent input field %s", opts.workflow.WorkflowID, opts.step.StepID, param.Name, param.Name)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("input field not found"),
				Node:            param.GetCore().Name.GetKeyNodeOrRoot(param.GetRootNode()),
			})

			incompleteMessages = append(incompleteMessages, msg)
			continue
		}

		field := opts.workflowInputs.Type.Fields[pIDx]

		foundParams[param.Name] = true

		val, expression, err := expression.GetValueOrExpressionValue(param.Value)
		if err != nil {
			msg := fmt.Sprintf("workflow step %s.%s with parameter %s failed to get value", opts.workflow.WorkflowID, opts.step.StepID, param.Name)

			logging.LogWarning(ctx, msg, err)

			incompleteMessages = append(incompleteMessages, msg)
			continue
		}

		var reference *ast.ExampleReference

		if expression != nil {
			var expIncompleteMessage string
			val, reference, expIncompleteMessage = resolveExpression(ctx, resolveExpressionOpts{
				expression:         *expression,
				source:             field.Type,
				msgContext:         fmt.Sprintf("workflow step %s.%s with expression in parameter", opts.workflow.WorkflowID, opts.step.StepID),
				expressionNode:     param.GetCore().Value.GetKeyNodeOrRoot(param.GetRootNode()),
				testArazzo:         opts.testArazzo,
				workflow:           opts.workflow,
				workflowSteps:      opts.previousSteps,
				workflowInputs:     opts.workflowInputs,
				isTopLevelWorkflow: opts.isTopLevelWorkflow,
			})
			if expIncompleteMessage != "" {
				incompleteMessages = append(incompleteMessages, expIncompleteMessage)
				continue
			}
		}

		exampleName := fmt.Sprintf("%s[%d]", opts.workflow.WorkflowID, opts.stepIdx)

		if reference != nil {
			field.Type.Examples = field.Type.Examples.AppendExample(ast.NewExampleReference(exampleName, "", reference))
		} else {
			field.Type.Examples = field.Type.Examples.AppendExample(ast.NewExample(exampleName, "", val))
		}
	}

	for _, param := range requiredParams {
		if foundParams[param] {
			continue
		}

		stepNode := opts.workflow.GetCore().Steps.GetSliceValueNodeOrRoot(opts.stepIdx, opts.workflow.GetRootNode())

		msg := fmt.Sprintf("workflow step %s.%s referencing missing required parameter %s", opts.workflow.WorkflowID, opts.step.StepID, param)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("missing required parameter"),
			Node:            stepNode,
		})

		incompleteMessages = append(incompleteMessages, msg)
	}

	return incompleteMessages
}

type requiredOpParams struct {
	Path   []string
	Query  []string
	Header []string
}

func getOpRequiredParams(op *ast.Operation) *requiredOpParams {
	if op.Request == nil || op.Request.Params == nil {
		return nil
	}

	requiredParams := &requiredOpParams{}

	getRequiredParams := func(params []*ast.Param, paramType string) {
		for _, param := range params {
			switch paramType {
			case "path":
				if !param.Field.Optional && !slices.Contains(requiredParams.Path, param.Field.Name) {
					requiredParams.Path = append(requiredParams.Path, param.Field.Name)
				}
			case "query":
				if !param.Field.Optional && !slices.Contains(requiredParams.Query, param.Field.Name) {
					requiredParams.Query = append(requiredParams.Query, param.Field.Name)
				}
			case "header":
				if !param.Field.Optional && !slices.Contains(requiredParams.Header, param.Field.Name) {
					requiredParams.Header = append(requiredParams.Header, param.Field.Name)
				}
			}
		}
	}

	getRequiredParams(op.Request.Params.PathParams, "path")
	getRequiredParams(op.Request.Params.QueryParams, "query")
	getRequiredParams(op.Request.Params.HeaderParams, "header")

	if len(requiredParams.Path) == 0 && len(requiredParams.Query) == 0 && len(requiredParams.Header) == 0 {
		return nil
	}

	return requiredParams
}

func getStepParameters(a *arazzo.Arazzo, workflow *arazzo.Workflow, step *arazzo.Step) []*arazzo.ReusableParameter {
	// Merge the step parameters with the workflow parameters
	parameters := []*arazzo.ReusableParameter{}

	parameters = append(parameters, workflow.Parameters...)

	for _, rp := range step.Parameters {
		param := rp.Get(a.Components)

		idx := slices.IndexFunc(parameters, func(p *arazzo.ReusableParameter) bool {
			existingP := p.Get(a.Components)

			return existingP.Name == param.Name && existingP.In == param.In
		})

		if idx == -1 {
			parameters = append(parameters, rp)
		} else {
			parameters[idx] = rp
		}
	}

	return parameters
}
