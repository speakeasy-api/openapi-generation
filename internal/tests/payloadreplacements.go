package tests

import (
	"context"
	"fmt"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type resolvePayloadReplacementsOpts struct {
	node          *yaml.Node
	testArazzo    *arazzo.Arazzo
	workflow      *arazzo.Workflow
	msgContext    string
	nodeContext   string
	workflowSteps []ast.ArazzoStep
}

func resolvePayloadReplacements(ctx context.Context, opts resolvePayloadReplacementsOpts) (*yaml.Node, []*ast.ExampleReplacement, []string) {
	return handleNode(ctx, "", opts.node, opts)
}

func handleNode(ctx context.Context, path string, node *yaml.Node, opts resolvePayloadReplacementsOpts) (*yaml.Node, []*ast.ExampleReplacement, []string) {
	switch node.Kind {
	case yaml.MappingNode:
		return handleMappingNode(ctx, path, node, opts)
	case yaml.SequenceNode:
		return handleSequenceNode(ctx, path, node, opts)
	case yaml.AliasNode:
		// TODO support alias nodes we will need to dereference before passing into `resolveRequestPayload` and any other nodes encountered if they alias a string scalar node we need to handle as such
		msg := fmt.Sprintf("%s has unsupported %s alias node", opts.msgContext, opts.nodeContext)
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: fmt.Errorf("unsupported %s alias node", opts.nodeContext),
			Node:            node,
		})
		return nil, nil, []string{msg}
	default:
		panic(fmt.Sprintf("unexpected node kind %v", node.Kind))
	}
}

func handleMappingNode(ctx context.Context, path string, node *yaml.Node, opts resolvePayloadReplacementsOpts) (*yaml.Node, []*ast.ExampleReplacement, []string) {
	replacements := []*ast.ExampleReplacement{}
	incompleteMessages := []string{}

	toDelete := []int{}

	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i] // Currently not supporting replacements in keys
		v := node.Content[i+1]

		currentPath := path + "/" + k.Value

		switch v.Kind {
		case yaml.ScalarNode:
			replacement, incompleteMessage := handleScalarNode(ctx, currentPath, v, opts)
			if incompleteMessage != "" {
				incompleteMessages = append(incompleteMessages, incompleteMessage)
				continue
			}
			if replacement != nil {
				replacements = append(replacements, replacement)
				toDelete = append(toDelete, i)
			}
		default:
			n, r, im := handleNode(ctx, currentPath, v, opts)
			node.Content[i+1] = n
			replacements = append(replacements, r...)
			incompleteMessages = append(incompleteMessages, im...)
		}
	}

	contentToRetain := []*yaml.Node{}
	for i := 0; i < len(node.Content); i += 2 {
		if !slices.Contains(toDelete, i) {
			contentToRetain = append(contentToRetain, node.Content[i], node.Content[i+1])
		}
	}

	node.Content = contentToRetain

	return node, replacements, incompleteMessages
}

func handleSequenceNode(ctx context.Context, path string, node *yaml.Node, opts resolvePayloadReplacementsOpts) (*yaml.Node, []*ast.ExampleReplacement, []string) {
	replacements := []*ast.ExampleReplacement{}
	incompleteMessages := []string{}

	toDelete := []int{}

	for i := 0; i < len(node.Content); i++ {
		v := node.Content[i]

		currentPath := fmt.Sprintf("%s/%d", path, i)

		switch v.Kind {
		case yaml.ScalarNode:
			replacement, incompleteMessage := handleScalarNode(ctx, currentPath, v, opts)
			if incompleteMessage != "" {
				incompleteMessages = append(incompleteMessages, incompleteMessage)
				continue
			}
			if replacement != nil {
				replacements = append(replacements, replacement)
				toDelete = append(toDelete, i)
			}
		default:
			n, r, im := handleNode(ctx, currentPath, v, opts)
			node.Content[i] = n
			replacements = append(replacements, r...)
			incompleteMessages = append(incompleteMessages, im...)
		}
	}

	contentToRetain := []*yaml.Node{}
	for i := 0; i < len(node.Content); i++ {
		if !slices.Contains(toDelete, i) {
			contentToRetain = append(contentToRetain, node.Content[i])
		}
	}

	node.Content = contentToRetain

	return node, replacements, incompleteMessages
}

func handleScalarNode(ctx context.Context, path string, node *yaml.Node, opts resolvePayloadReplacementsOpts) (*ast.ExampleReplacement, string) {
	if node.Tag != "!!str" {
		return nil, ""
	}

	exp := expression.Expression(node.Value)

	if exp.IsExpression() {
		val, reference, incompleteMessage := resolveExpression(ctx, resolveExpressionOpts{
			expression:     exp,
			msgContext:     opts.msgContext + " with expression in requestBody",
			expressionNode: node,
			testArazzo:     opts.testArazzo,
			workflow:       opts.workflow,
			workflowSteps:  opts.workflowSteps,
			// TODO not sure what to set these to yet need a use case
			source:             nil,
			workflowInputs:     nil,
			isTopLevelWorkflow: false,
		})
		if incompleteMessage != "" {
			return nil, incompleteMessage
		}

		var example *ast.Example

		if val != nil {
			example = ast.NewExample("", "", val)
		} else {
			example = ast.NewExampleReference("", "", reference)
		}

		return &ast.ExampleReplacement{
			Path:  path,
			Value: example,
		}, ""
	}

	return nil, ""
}
