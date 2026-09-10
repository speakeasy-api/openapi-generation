package tests

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/validation"
)

type ArazzoDocumentInfo struct {
	Location string
	Doc      *arazzo.Arazzo
}

type PopulateTestsOptions struct {
	AST                   *ast.AST
	DocInfo               *ArazzoDocumentInfo
	Subsystem             *subsystem.Subsystem
	Target                types.Target
	InternalTestGroupName string
	OutDir                string
}

func PopulateTests(ctx context.Context, opts PopulateTestsOptions) []*ast.TestGroup {
	accountHasAccess := licensing.AccountHasFeatureAccess(ctx, features.FeatureTests)

	if !opts.Subsystem.Features.IsFeatureSupported(ctx, features.FeatureTests) || (!accountHasAccess) || opts.DocInfo == nil || opts.DocInfo.Doc == nil {
		return nil
	}

	workflows, projected := populateArazzoWorkflows(ctx, opts)
	if len(workflows) > 0 {
		opts.AST.Arazzo = &ast.Arazzo{Workflows: workflows}
	}

	tests := projectTests(projected)

	opts.Subsystem.Features.RecordFeatureUsage(ctx, features.FeatureTests)

	return tests
}

const (
	testServerURLSentinel = "TEST_SERVER_URL"
	testServerBaseURL     = "http://localhost:18080"
)

type projectedTestWorkflow struct {
	TestGroup       string
	Workflow        *ast.ArazzoWorkflow
	UsingMockServer bool
	Incomplete      []string
	InternalID      string
	InternalEnvVars []ast.TestEnvVar
}

func populateArazzoWorkflows(ctx context.Context, opts PopulateTestsOptions) ([]*ast.ArazzoWorkflow, []*projectedTestWorkflow) {
	workflows := []*ast.ArazzoWorkflow{}
	projected := []*projectedTestWorkflow{}

	numOpenAPIFiles := 0

	for item := range arazzo.Walk(ctx, opts.DocInfo.Doc) {
		err := item.Match(arazzo.Matcher{
			SourceDescription: func(sd *arazzo.SourceDescription) error {
				if sd.Type == arazzo.SourceDescriptionTypeOpenAPI {
					numOpenAPIFiles++
				}

				if numOpenAPIFiles > 1 {
					return &validation.Error{
						UnderlyingError: errors.New("multiple openapi files not currently supported"),
						Node:            opts.DocInfo.Doc.GetCore().SourceDescriptions.GetKeyNodeOrRoot(opts.DocInfo.Doc.GetRootNode()),
					}
				}

				return nil
			},
		})
		if err != nil {
			logging.LogWarning(ctx, "arazzo document contains unsupported features, skipping test generation", err)
			return nil, nil
		}
	}

	operationsByID := GetOperationsByID(opts.AST, &GetOperationsByIDOptions{
		IncludeWebhooks: false,
	})

	for _, workflow := range opts.DocInfo.Doc.Workflows {
		testGroup, w, incompleteMessages := handleWorkflow(ctx, handleWorkflowOpts{
			ast:                   opts.AST,
			testArazzo:            opts.DocInfo.Doc,
			workflow:              workflow,
			topLevel:              true,
			operationsByID:        operationsByID,
			target:                opts.Target,
			internalTestGroupName: opts.InternalTestGroupName,
			subsystem:             opts.Subsystem,
			isTopLevelWorkflow:    true,
			docInfo:               opts.DocInfo,
		})

		if w == nil {
			continue
		}

		workflows = append(workflows, w)
		projected = append(projected, &projectedTestWorkflow{
			TestGroup:       testGroup,
			Workflow:        w,
			UsingMockServer: usingMockServer(w),
			Incomplete:      incompleteMessages,
			InternalID:      getInternalID(ctx, workflow),
			InternalEnvVars: getInternalEnvVars(ctx, workflow),
		})
	}

	return workflows, projected
}

func usingMockServer(workflow *ast.ArazzoWorkflow) bool {
	if workflow == nil {
		return false
	}
	if strings.Contains(workflow.Server, testServerURLSentinel) || strings.Contains(workflow.Server, testServerBaseURL) {
		return true
	}
	for _, step := range workflow.Steps {
		if nested := ast.GetArazzoWorkflow(step); nested != nil && usingMockServer(nested.Workflow) {
			return true
		}
	}
	return false
}

func projectTests(projected []*projectedTestWorkflow) []*ast.TestGroup {
	testsByTestGroup := map[string][]*ast.Test{}

	for _, testWorkflow := range projected {
		t := ast.Test{
			Name:            testWorkflow.Workflow.Name,
			Workflow:        testWorkflow.Workflow,
			UsingMockServer: testWorkflow.UsingMockServer,
			Incomplete:      testWorkflow.Incomplete,
			InternalID:      testWorkflow.InternalID,
			InternalEnvVars: testWorkflow.InternalEnvVars,
		}

		tests := testsByTestGroup[testWorkflow.TestGroup]
		testsByTestGroup[testWorkflow.TestGroup] = append(tests, &t)
	}

	testGroups := make([]*ast.TestGroup, 0, len(testsByTestGroup))
	for testGroup, tests := range testsByTestGroup {
		testGroups = append(testGroups, &ast.TestGroup{Name: testGroup, Tests: tests})
	}

	return testGroups
}

type handleWorkflowOpts struct {
	ast                   *ast.AST
	testArazzo            *arazzo.Arazzo
	workflow              *arazzo.Workflow
	topLevel              bool
	operationsByID        map[string][]*ast.Operation
	target                types.Target
	internalTestGroupName string
	subsystem             *subsystem.Subsystem
	stepID                string
	isTopLevelWorkflow    bool
	docInfo               *ArazzoDocumentInfo
}

func handleWorkflow(ctx context.Context, opts handleWorkflowOpts) (string, *ast.ArazzoWorkflow, []string) {
	workflow := opts.workflow

	// TODO: handle `dependsOn` like a step referencing another workflow but these steps run first

	workflowIdx := slices.Index(opts.testArazzo.Workflows, workflow)

	// The `x-speakeasy-test` annotation is only supported for disabling a workflow from getting its own test
	// if its a reference workflow then we still handle it
	if isTestingDisabledForWorkflow(workflow) && opts.topLevel {
		return "", nil, nil
	}

	targets := getTestTargets(ctx, workflow)
	excludedTargets := getTestTargetsExclude(ctx, workflow)
	internalTestGroups := getInternalTestGroups(ctx, workflow)

	target := opts.target.Target
	if env.IsDebug() {
		target = opts.target.Template
	}

	if len(excludedTargets) > 0 && slices.Contains(excludedTargets, target) {
		return "", nil, nil
	}
	if len(targets) > 0 && !slices.Contains(targets, target) && target != "mockserver" {
		return "", nil, nil
	}
	if len(internalTestGroups) > 0 && !slices.Contains(internalTestGroups, opts.internalTestGroupName) {
		return "", nil, nil
	}

	testGroup := getTestGroup(ctx, workflow)

	incompleteMessages := []string{}

	if !workflow.Valid {
		msg := fmt.Sprintf("workflow %s is not valid", workflow.WorkflowID)

		logging.LogWarning(ctx, msg+", see previous warnings", &validation.Error{
			UnderlyingError: errors.New("invalid workflow"),
			Node:            opts.testArazzo.GetCore().Workflows.GetSliceValueNodeOrRoot(workflowIdx, opts.testArazzo.GetRootNode()),
		})
		incompleteMessages = append(incompleteMessages, msg)
	}

	server := getTestServer(ctx, opts.testArazzo, workflow)

	var serverVal string
	if server != nil {
		serverVal = server.BaseURL
	}

	mockServerEnabled := true
	if opts.subsystem.Config != nil && opts.subsystem.Config.Generation.MockServer != nil {
		mockServerEnabled = !opts.subsystem.Config.Generation.MockServer.Disabled
	}

	generateNewTests := opts.subsystem.Config != nil && opts.subsystem.Config.Generation.Tests.GenerateNewTests

	if serverVal == "" &&
		mockServerEnabled &&
		(opts.internalTestGroupName == "" || opts.internalTestGroupName == "review" || generateNewTests) {
		serverVal = "x-env: " + testServerURLSentinel + "; " + testServerBaseURL // TODO we may want to make this configurable
	}

	inputs, im := handleWorkflowInputs(ctx, handleWorkflowInputsOpts{
		workflow:  workflow,
		subsystem: opts.subsystem,
		docInfo:   opts.docInfo,
	})
	if len(im) > 0 {
		incompleteMessages = append(incompleteMessages, im...)
	}

	w := &ast.ArazzoWorkflow{
		Name:        workflow.WorkflowID,
		Description: pointer.Value(workflow.Description),
		Server:      serverVal,
		Inputs:      inputs,
	}

	steps := []ast.ArazzoStep{}

	for idx, step := range workflow.Steps {
		s, sIncompleteMessages := handleStep(ctx, handleStepOpts{
			ast:                   opts.ast,
			testArazzo:            opts.testArazzo,
			target:                opts.target,
			internalTestGroupName: opts.internalTestGroupName,
			subsystem:             opts.subsystem,
			workflow:              workflow,
			stepIdx:               idx,
			step:                  step,
			previousSteps:         steps,
			operationsByID:        opts.operationsByID,
			workflowInputs:        inputs,
			isTopLevelWorkflow:    opts.isTopLevelWorkflow,
			docInfo:               opts.docInfo,
		})
		incompleteMessages = append(incompleteMessages, sIncompleteMessages...)

		if s != nil {
			if s.GetType() == ast.ArazzoStepTypeOperation {
				testOp := s.(*ast.ArazzoOperationStep)

				// TODO: webhooks are not supported yet so we skip this operation for now
				if testOp.Operation.Webhook != nil {
					msg := "webhooks are not supported yet, skipping webhook operation " + testOp.Operation.OriginalID

					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("webhooks not supported"),
						Node:            workflow.GetCore().Steps.GetSliceValueNodeOrRoot(idx, workflow.GetRootNode()),
					})

					incompleteMessages = append(incompleteMessages, msg)
					continue
				}
			}

			steps = append(steps, s)
		}
	}

	if len(steps) == 0 {
		usageContext := ast.CreateUsageContext(nil, nil, nil, opts.subsystem.Config.Generation.Tests.SkipResponseBodyAssertions)
		usageContext.ExampleName = fmt.Sprintf("%s[%d]", w.Name, 0)
		usageContext.Test = w

		steps = append(steps, &ast.ArazzoOperationStep{
			ArazzoStepBase: ast.ArazzoStepBase{Type: string(ast.ArazzoStepTypeOperation)},
			UsageContext:   usageContext,
		})
	}
	w.Steps = steps

	securityVal, im := handleSecurity(ctx, handleSecurityOpts{
		ast:           opts.ast,
		testArazzo:    opts.testArazzo,
		workflow:      opts.workflow,
		testSecurity:  getWorkflowTestSecurity(ctx, opts.testArazzo, workflow),
		workflowSteps: w.Steps,
		msgContext:    "workflow " + opts.workflow.WorkflowID,
		exampleName:   opts.workflow.WorkflowID,
	})
	if len(im) > 0 {
		incompleteMessages = append(incompleteMessages, im...)
	}
	w.Security = securityVal

	contextIdx := 0
	for _, step := range w.Steps {
		if step.GetType() == ast.ArazzoStepTypeOperation {
			testOp := step.(*ast.ArazzoOperationStep)
			if testOp.UsageContext.Test != nil {
				continue
			}

			testOp.UsageContext = applyTestContextToUsageContext(testOp.UsageContext, w, testOp, usingMockServer(w))
			testOp.UsageContext.ContextIndex = contextIdx
			// TODO we prob need server_url at the operation level as well
			if contextIdx > 0 && testOp.Security == nil {
				testOp.UsageContext.SkipSDKInstantiation = true
			}

			contextIdx++
		}
	}

	outputs, im := handleWorkflowOutputs(ctx, handleWorkflowOutputsOpts{
		testArazzo:         opts.testArazzo,
		arazzoWorkflow:     workflow,
		workflow:           w,
		subsystem:          opts.subsystem,
		isTopLevelWorkflow: opts.isTopLevelWorkflow,
	})
	if len(im) > 0 {
		incompleteMessages = append(incompleteMessages, im...)
	}
	w.Outputs = outputs

	if w.Inputs != nil {
		// add "any" type to any unresolved inputs
		for _, field := range w.Inputs.Type.Fields {
			if field.Type == nil {
				field.Type = ast.NewType(&ast.TypeDef{
					Type: ast.DataTypeAny,
				}, ast.ContextStack{})
			}
		}
	}

	return testGroup, w, incompleteMessages
}

type handleWorkflowOutputsOpts struct {
	testArazzo         *arazzo.Arazzo
	arazzoWorkflow     *arazzo.Workflow
	workflow           *ast.ArazzoWorkflow
	subsystem          *subsystem.Subsystem
	isTopLevelWorkflow bool
}

func handleWorkflowOutputs(ctx context.Context, opts handleWorkflowOutputsOpts) (*ast.FieldDef, []string) {
	arazzoWorkflow := opts.arazzoWorkflow

	if arazzoWorkflow.Outputs == nil {
		return nil, nil
	}

	incompleteMessages := []string{}

	outputFields := ast.Fields{}

	for name, output := range arazzoWorkflow.Outputs.All() {
		example, ref, im := resolveExpression(ctx, resolveExpressionOpts{
			expression:         output,
			source:             nil, // Shouldn't be needed for outputs
			msgContext:         fmt.Sprintf("workflow %s with expression in output %s", arazzoWorkflow.WorkflowID, name),
			expressionNode:     arazzoWorkflow.GetCore().Outputs.GetMapKeyNodeOrRoot(name, arazzoWorkflow.GetRootNode()),
			testArazzo:         opts.testArazzo,
			workflow:           opts.arazzoWorkflow,
			workflowSteps:      opts.workflow.Steps,
			workflowInputs:     opts.workflow.Inputs,
			isTopLevelWorkflow: opts.isTopLevelWorkflow,
		})
		if im != "" {
			incompleteMessages = append(incompleteMessages, im)
			continue
		}
		if example != nil {
			// TODO is this the correct messaging and way to handle this?
			msg := fmt.Sprintf("workflow %s with expression in output %s didn't contain expression", arazzoWorkflow.WorkflowID, name)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: fmt.Errorf("workflow %s with expression in output %s didn't contain expression", arazzoWorkflow.WorkflowID, name),
				Node:            arazzoWorkflow.GetCore().Outputs.GetMapKeyNodeOrRoot(name, arazzoWorkflow.GetRootNode()),
			})
			incompleteMessages = append(incompleteMessages, msg)
			continue
		}
		target, ok := ref.Target.(ast.ResponseBodyTarget)
		if !ok {
			msg := fmt.Sprintf("workflow %s with expression in output %s didn't resolve to a response body target", arazzoWorkflow.WorkflowID, name)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: fmt.Errorf("workflow %s with expression in output %s didn't resolve to a response body target", arazzoWorkflow.WorkflowID, name),
				Node:            arazzoWorkflow.GetCore().Outputs.GetMapKeyNodeOrRoot(name, arazzoWorkflow.GetRootNode()),
			})
			incompleteMessages = append(incompleteMessages, msg)
			continue
		}

		field, ok := target.Target.(*ast.FieldDef)
		if !ok {
			msg := fmt.Sprintf("workflow %s with expression in output %s didn't resolve to a field", arazzoWorkflow.WorkflowID, name)
			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: fmt.Errorf("workflow %s with expression in output %s didn't resolve to a field", arazzoWorkflow.WorkflowID, name),
				Node:            arazzoWorkflow.GetCore().Outputs.GetMapKeyNodeOrRoot(name, arazzoWorkflow.GetRootNode()),
			})
			incompleteMessages = append(incompleteMessages, msg)
			continue
		}

		typ := field.Type.ShallowCopy()
		typ.Examples = []*ast.Example{
			ast.NewExampleReference(opts.workflow.Name, "", ref),
		}
		typ.ResolvedModel = typ.Name // A little hack to rewire some required context for pythonv2, this doesn't impact anything meaningfully in tests but there was a check to make sure this was set

		outputFields = outputFields.MustAddField(&ast.FieldDef{
			Name:         name,
			OriginalName: name,
			Type:         typ,
		}, opts.subsystem.Config.MaintainOpenAPIOrder())
	}

	if len(incompleteMessages) > 0 {
		return nil, incompleteMessages
	}

	if len(outputFields) == 0 {
		return nil, nil
	}

	name := opts.workflow.Name + "_outputs"
	outputs := &ast.FieldDef{
		Name:         "Outputs",
		OriginalName: "Outputs",
		Type: ast.NewType(&ast.TypeDef{
			Name:           name,
			Type:           ast.DataTypeClass,
			OutputLocation: "tests",
			Fields:         outputFields,
			ResolvedModel:  name,
		}, ast.ContextStack{}),
	}

	return outputs, nil
}

type handleWorkflowInputsOpts struct {
	workflow  *arazzo.Workflow
	subsystem *subsystem.Subsystem
	docInfo   *ArazzoDocumentInfo
}

func handleWorkflowInputs(ctx context.Context, opts handleWorkflowInputsOpts) (*ast.FieldDef, []string) {
	workflow := opts.workflow

	if workflow == nil || workflow.Inputs == nil {
		return nil, nil
	}

	if workflow.Inputs.IsReference() {
		msg := fmt.Sprintf("workflow %s has input reference which is currently unsupported", workflow.WorkflowID)
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported workflow feature"),
			Node:            workflow.GetCore().Inputs.GetValueNodeOrRoot(workflow.GetRootNode()),
		})
		return nil, []string{
			msg,
		}
	}

	if workflow.Inputs.IsBool() {
		msg := fmt.Sprintf("workflow %s has input type which is currently unsupported", workflow.WorkflowID)
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported workflow feature"),
			Node:            workflow.GetCore().Inputs.GetValueNodeOrRoot(workflow.GetRootNode()),
		})
		return nil, []string{
			msg,
		}
	}

	types, err := getSchemaTypes(ctx, workflow.Inputs, "object", opts.docInfo)
	if err != nil {
		return nil, []string{err.Error()}
	}

	if len(types) > 1 {
		msg := fmt.Sprintf("workflow %s has multiple input types which is currently unsupported [%s]", workflow.WorkflowID, strings.Join(types, ", "))
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported workflow feature"),
			Node:            workflow.GetCore().Inputs.GetValueNodeOrRoot(workflow.GetRootNode()),
		})
		return nil, []string{
			msg,
		}
	}

	if types[0] != "object" {
		msg := fmt.Sprintf("workflow %s has non-object input type %s", workflow.WorkflowID, types[0])
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported workflow input type"),
			Node:            workflow.GetCore().Inputs.GetValueNodeOrRoot(workflow.GetRootNode()),
		})
		return nil, []string{
			msg,
		}
	}

	inputsSchema := workflow.Inputs.GetSchema()

	if inputsSchema.Properties.Len() == 0 {
		msg := fmt.Sprintf("workflow %s has no inputs", workflow.WorkflowID)
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("workflow declares input object with no properties"),
			Node:            workflow.GetCore().Inputs.GetValueNodeOrRoot(workflow.GetRootNode()),
		})
		return nil, []string{
			msg,
		}
	}

	name := workflow.WorkflowID + "_inputs"
	inputType := ast.NewType(&ast.TypeDef{
		Type:           ast.DataTypeClass,
		Name:           name,
		OutputLocation: "tests",
		ResolvedModel:  name,
	}, ast.ContextStack{})

	// We are going to create placeholder fields and then add the type later when we have resolved its type via usage
	for name, schema := range inputsSchema.Properties.All() {
		required := slices.Contains(inputsSchema.Required, name)

		nullable := false
		if schema.IsBool() {
			nullable = true
		} else {
			s := schema.GetSchema()
			if s.Nullable != nil && *s.Nullable {
				nullable = true
			}
			if slices.Contains(s.GetType(), "null") {
				nullable = true
			}
		}

		inputType.Fields = inputType.Fields.MustAddField(&ast.FieldDef{
			Name:         name,
			OriginalName: name,
			Nullable:     nullable,
			Optional:     !required,
		}, opts.subsystem.Config.MaintainOpenAPIOrder())
	}

	return &ast.FieldDef{
		Name:         "Inputs",
		OriginalName: "Inputs",
		Type:         inputType,
	}, nil
}

func getSchemaTypes(ctx context.Context, s *oas3.JSONSchema[oas3.Referenceable], defaultType string, docInfo *ArazzoDocumentInfo) ([]string, error) {
	types := []string{}

	_, err := s.Resolve(ctx, oas3.ResolveOptions{
		TargetLocation: docInfo.Location,
		RootDocument:   docInfo.Doc,
	})
	if err != nil {
		return nil, err
	}
	schema := s.GetResolvedSchema()

	if schema.IsSchema() {
		for _, t := range schema.GetSchema().GetType() {
			types = append(types, string(t))
		}
	} else {
		types = append(types, "any")
	}

	if len(types) == 0 {
		// assume to be the default type
		types = append(types, defaultType)
	}

	return types, nil
}
