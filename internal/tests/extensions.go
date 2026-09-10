package tests

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/marshaller"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

const (
	ExtTestGroup          = "x-speakeasy-test-group"
	ExtTestServer         = "x-speakeasy-test-server"
	ExtTestTargets        = "x-speakeasy-test-targets"
	ExtTestTargetsExclude = "x-speakeasy-test-targets-exclude"
	ExtTestSecurity       = "x-speakeasy-test-security"
	ExtInternalID         = "x-speakeasy-test-internal-id"
	ExtInternalEnv        = "x-speakeasy-test-internal-env-vars"
	ExtInternalTestGroups = "x-speakeasy-test-internal-test-groups"
)

func NewTestGroupExtension(group string) *extensions.Element {
	var node yaml.Node
	if err := node.Encode(group); err != nil {
		return nil
	}

	return extensions.NewElem(ExtTestGroup, pointer.From(node))
}

func getTestGroup(ctx context.Context, workflow *arazzo.Workflow) string {
	group, err := extensions.GetExtensionValue[string](workflow.Extensions, ExtTestGroup)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get test group from workflow %s extension %s: %s", workflow.WorkflowID, ExtTestGroup, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get test group from workflow extension"),
			})
		}

		return "sdk" // Default to sdk if not found or unable to parse
	}

	return strings.ToLower(*group)
}

type TestServer struct {
	marshaller.Model[TestServerCore]

	BaseURL string `yaml:"baseUrl"`
}

type TestServerCore struct {
	marshaller.CoreModel `model:"testServer"`

	BaseURL marshaller.Node[string] `key:"baseUrl"`
}

func NewTestServerExtension(server *TestServer) *extensions.Element {
	var node yaml.Node
	if err := node.Encode(server); err != nil {
		return nil
	}

	return extensions.NewElem(ExtTestServer, pointer.From(node))
}

func getTestServer(ctx context.Context, doc *arazzo.Arazzo, workflow *arazzo.Workflow) *TestServer {
	testServer := getTestServerExtension(ctx, "workflow "+workflow.WorkflowID, workflow.Extensions)
	if testServer != nil {
		return testServer
	}

	return getTestServerExtension(ctx, "document "+doc.Info.Title, doc.Extensions)
}

func getTestServerExtension(ctx context.Context, target string, exts *extensions.Extensions) *TestServer {
	var server TestServer
	validationErrs, err := extensions.UnmarshalExtensionModel[TestServer, TestServerCore](ctx, exts, ExtTestServer, &server)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get test server from %s extension %s: %s", target, ExtTestServer, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get test server from extension"),
			})
		}
		return nil
	}
	for _, validationErr := range validationErrs {
		logging.LogWarning(ctx, fmt.Sprintf("validation failed for test server from %s extension %s", target, ExtTestServer), validationErr)
	}
	if len(validationErrs) > 0 {
		return nil
	}

	return &server
}

func NewTestTargetsExtension(targets []string) *extensions.Element {
	var node yaml.Node
	if err := node.Encode(targets); err != nil {
		return nil
	}

	return extensions.NewElem(ExtTestTargets, pointer.From(node))
}

func getTestTargets(ctx context.Context, workflow *arazzo.Workflow) []string {
	targets, err := extensions.GetExtensionValue[[]string](workflow.Extensions, ExtTestTargets)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get test targets from workflow %s extension %s: %s", workflow.WorkflowID, ExtTestTargets, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get test targets from workflow extension"),
			})
		}
		return nil
	}

	if targets == nil {
		return []string{}
	}
	return *targets
}

func getTestTargetsExclude(ctx context.Context, workflow *arazzo.Workflow) []string {
	targets, err := extensions.GetExtensionValue[[]string](workflow.Extensions, ExtTestTargetsExclude)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get excluded test targets from workflow %s extension %s: %s", workflow.WorkflowID, ExtTestTargets, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get excluded test targets from workflow extension"),
			})
		}
		return nil
	}

	if targets == nil {
		return []string{}
	}
	return *targets
}

type TestSecurity struct {
	marshaller.Model[TestSecurityCore]

	Value yaml.Node `yaml:"value"`
}

type TestSecurityCore struct {
	marshaller.CoreModel `model:"testSecurity"`

	Value marshaller.Node[yaml.Node] `key:"value"`
}

func NewTestSecurityExtension(security *TestSecurity) *extensions.Element {
	var node yaml.Node
	if err := node.Encode(security); err != nil {
		return nil
	}

	return extensions.NewElem(ExtTestSecurity, pointer.From(node))
}

func getWorkflowTestSecurity(ctx context.Context, doc *arazzo.Arazzo, workflow *arazzo.Workflow) *TestSecurity {
	testSecurity := getTestSecurityExtension(ctx, "workflow "+workflow.WorkflowID, workflow.Extensions)
	if testSecurity != nil {
		return testSecurity
	}

	return getTestSecurityExtension(ctx, "document "+doc.Info.Title, doc.Extensions)
}

func getStepTestSecurity(ctx context.Context, step *arazzo.Step) *TestSecurity {
	return getTestSecurityExtension(ctx, "step "+step.StepID, step.Extensions)
}

func getTestSecurityExtension(ctx context.Context, target string, exts *extensions.Extensions) *TestSecurity {
	var security TestSecurity
	validationErrs, err := extensions.UnmarshalExtensionModel[TestSecurity, TestSecurityCore](ctx, exts, ExtTestSecurity, &security)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get test security from %s extension %s: %s", target, ExtTestSecurity, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get test security from extension"),
			})
		}
		return nil
	}
	for _, validationErr := range validationErrs {
		logging.LogWarning(ctx, fmt.Sprintf("validation failed for test security from %s extension %s", target, ExtTestSecurity), validationErr)
	}
	if len(validationErrs) > 0 {
		return nil
	}

	return &security
}

func NewInternalIDExtension(id string) *extensions.Element {
	var node yaml.Node
	if err := node.Encode(id); err != nil {
		return nil
	}

	return extensions.NewElem(ExtInternalID, pointer.From(node))
}

func getInternalID(ctx context.Context, workflow *arazzo.Workflow) string {
	id, err := extensions.GetExtensionValue[string](workflow.Extensions, ExtInternalID)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get internal id from workflow %s extension %s: %s", workflow.WorkflowID, ExtInternalID, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get internal id from workflow extension"),
			})
		}
		return ""
	}

	if id == nil {
		return ""
	}

	return *id
}

func NewInternalEnvVarsExtension(envVars []ast.TestEnvVar) *extensions.Element {
	vars := map[string]string{}

	for _, envVar := range envVars {
		vars[envVar.Name] = envVar.Value
	}

	var node yaml.Node
	if err := node.Encode(vars); err != nil {
		return nil
	}

	return extensions.NewElem(ExtInternalEnv, pointer.From(node))
}

func getInternalEnvVars(ctx context.Context, workflow *arazzo.Workflow) []ast.TestEnvVar {
	envVars, err := extensions.GetExtensionValue[map[string]string](workflow.Extensions, ExtInternalEnv)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get internal env vars from workflow %s extension %s: %s", workflow.WorkflowID, ExtInternalEnv, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get internal env vars from workflow extension"),
			})
		}
		return nil
	}

	vars := []ast.TestEnvVar{}

	if envVars != nil {
		for k, v := range *envVars {
			vars = append(vars, ast.TestEnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	return vars
}

func NewInternalTestGroupsExtension(groups []string) *extensions.Element {
	var node yaml.Node
	if err := node.Encode(groups); err != nil {
		return nil
	}

	return extensions.NewElem(ExtInternalTestGroups, pointer.From(node))
}

func getInternalTestGroups(ctx context.Context, workflow *arazzo.Workflow) []string {
	groups, err := extensions.GetExtensionValue[[]string](workflow.Extensions, ExtInternalTestGroups)
	if err != nil {
		if !errors.Is(err, extensions.ErrNotFound) {
			logging.LogWarning(ctx, fmt.Sprintf("failed to get internal test groups from workflow %s extension %s: %s", workflow.WorkflowID, ExtInternalTestGroups, err.Error()), validation.Error{
				UnderlyingError: errors.New("failed to get internal test groups from workflow extension"),
			})
		}
		return nil
	}

	return *groups
}
