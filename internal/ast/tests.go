package ast

// Arazzo contains test-generation data derived from an input Arazzo document.
//
// This keeps Arazzo-specific graph data namespaced under AST.Arazzo rather than
// flattening it onto the top-level AST alongside general SDK/OpenAPI fields.
type Arazzo struct {
	// Workflows are reusable workflow definitions keyed by workflow name and used
	// by templates to resolve workflow-step references.
	Workflows []*ArazzoWorkflow
}

type Tests struct {
	// TestGroups represent different groups of tests that will be grouped with the x-speakeasy-test-group extension in an Arazzo document.
	// Tests live at the top level of the AST as they aren't associated with any particular SDK or operation and could contain calls to multiple.
	TestGroups          []*TestGroup
	GenerateExampleFile bool
}

type TestEnvVar struct {
	Name  string
	Value string
}

type TestGroup struct {
	Name  string
	Tests []*Test
}

type Test struct {
	Name            string
	Workflow        *ArazzoWorkflow
	UsingMockServer bool
	Incomplete      []string
	InternalID      string
	InternalEnvVars []TestEnvVar
}

type ArazzoStepType string

const (
	ArazzoStepTypeOperation ArazzoStepType = "operation"
	ArazzoStepTypeWorkflow  ArazzoStepType = "workflow"
)

type ArazzoStepBase struct {
	Type string
	// StepID is the ID of the step in an Arazzo workflow that this step is in
	StepID string
}

func (t *ArazzoStepBase) GetType() ArazzoStepType {
	return ArazzoStepType(t.Type)
}

func (t *ArazzoStepBase) GetStepID() string {
	return t.StepID
}

type ArazzoStep interface {
	GetType() ArazzoStepType
	GetStepID() string
}

// ArazzoWorkflow represents a reusable workflow definition.
type ArazzoWorkflow struct {
	Name        string
	Description string
	Server      string
	Security    *Example
	Steps       []ArazzoStep
	Inputs      *FieldDef
	Outputs     *FieldDef
}

// ArazzoWorkflowStep represents a step that invokes another workflow definition.
type ArazzoWorkflowStep struct {
	ArazzoStepBase

	// StepIdx is the original index of this step in the source Arazzo workflow.
	//
	// This must be used for example-name reconstruction because generated AST
	// step slices can omit unsupported steps (for example webhooks), which would
	// otherwise skew index-based lookups.
	StepIdx int
	// WorkflowID references the invoked workflow definition by name.
	WorkflowID string
	// Workflow is the resolved workflow definition when available.
	Workflow *ArazzoWorkflow
}

// ArazzoInvocationContext represents generic invocation metadata for an operation step.
type ArazzoInvocationContext struct {
	SDK       *SDK
	Operation *Operation
	StepIdx   int
	StepID    string
}

// ArazzoOperationStep represents an operation step node and its required context/metadata for test generation.
type ArazzoOperationStep struct {
	ArazzoStepBase

	Invocation *ArazzoInvocationContext
	// The context of the test operation, this is used to provide details for things such as security, server_url etc
	UsageContext *UsageContext
	// Operation and StepIdx are transitional compatibility fields while consumers migrate to Invocation.
	// Prefer Invocation.Operation and Invocation.StepIdx in new code.
	Operation *Operation
	// StepIdx is the original index of this step in the source Arazzo workflow.
	//
	// Two different operation step nodes can reference the same operation but
	// have different StepIdx values based on which step in the workflow they are in.
	StepIdx int
	// Security is an example of the security that should be used for the operation
	Security *Example
	// ResponseContentType is the content type of the response that should be used for the operation
	ResponseContentType string
}

func GetArazzoOperationStep(step ArazzoStep) *ArazzoOperationStep {
	switch step := step.(type) {
	case *ArazzoOperationStep:
		return step
	default:
		return nil
	}
}

func GetArazzoWorkflow(step ArazzoStep) *ArazzoWorkflowStep {
	switch step := step.(type) {
	case *ArazzoWorkflowStep:
		return step
	default:
		return nil
	}
}
