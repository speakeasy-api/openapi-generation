package tests

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/validation"
)

type handleSecurityOpts struct {
	ast           *ast.AST
	testArazzo    *arazzo.Arazzo
	workflow      *arazzo.Workflow
	testSecurity  *TestSecurity
	workflowSteps []ast.ArazzoStep
	msgContext    string
	exampleName   string
}

func handleSecurity(ctx context.Context, opts handleSecurityOpts) (*ast.Example, []string) {
	testSecurity := opts.testSecurity

	if testSecurity == nil {
		return nil, nil
	}

	var securityVal *ast.Example

	// TODO we need to find the source (either globally or in an operation) for the expression

	var im []string
	example, replacements, im := resolvePayloadReplacements(ctx, resolvePayloadReplacementsOpts{
		node:          &testSecurity.Value,
		testArazzo:    opts.testArazzo,
		workflow:      opts.workflow,
		msgContext:    opts.msgContext,
		nodeContext:   ExtTestSecurity,
		workflowSteps: opts.workflowSteps,
	})
	if len(im) > 0 {
		return nil, im
	}

	if example == nil {
		msg := fmt.Sprintf("%s referencing %s has invalid value", opts.msgContext, ExtTestSecurity)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: fmt.Errorf("invalid %s value", ExtTestSecurity),
			Node:            &testSecurity.Value,
		})

		return nil, []string{msg}
	} else {
		securityVal = ast.NewExample(opts.exampleName, "", example)
		securityVal.Replacements = replacements
	}

	return securityVal, nil
}
