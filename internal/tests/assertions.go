package tests

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/speakeasy-api/openapi/jsonpointer"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type getOpAssertionsOpts struct {
	workflow *arazzo.Workflow
	step     *arazzo.Step
	stepIdx  int
	op       *ast.Operation
}

func getOpAssertions(ctx context.Context, opts getOpAssertionsOpts) ([]ast.Assertion, string, []string) {
	type responseBodyAssertion struct {
		condition *criterion.Condition
		node      *yaml.Node
	}

	type requiredAssertions struct {
		statusCode               *criterion.Condition
		statusCodeNode           *yaml.Node
		contentType              *criterion.Condition
		contentTypeNode          *yaml.Node
		responseBodiesAssertions []responseBodyAssertion
	}

	workflow := opts.workflow
	step := opts.step

	requiredAsserts := requiredAssertions{}
	incompleteMessages := []string{}

	if len(opts.step.SuccessCriteria) == 0 {
		msg := fmt.Sprintf("workflow step %s.%s does not contain success criteria", workflow.WorkflowID, step.StepID)

		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("success criteria not found"),
			Node:            step.GetCore().SuccessCriteria.GetKeyNodeOrRoot(step.GetRootNode()),
		})

		incompleteMessages = append(incompleteMessages, msg)
		return nil, "", incompleteMessages
	}

	for _, c := range opts.step.SuccessCriteria {
		switch c.Type.GetType() {
		case criterion.CriterionTypeSimple:
			cond, _ := c.GetCondition()

			if cond == nil && c.Context != nil {
				switch c.Context.GetType() {
				case expression.ExpressionTypeStatusCode:
					fallthrough
				case expression.ExpressionTypeResponse:
					cond = &criterion.Condition{
						Expression: *c.Context,
						Operator:   criterion.OperatorEQ,
						Value:      c.Condition,
					}
				default:
					msg := fmt.Sprintf("workflow step %s.%s contains unsupported context runtime expression type %s", workflow.WorkflowID, step.StepID, c.Context.GetType())

					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("context runtime expression type not currently supported"),
						Node:            c.GetCore().Context.GetKeyNodeOrRoot(c.GetRootNode()),
					})

					incompleteMessages = append(incompleteMessages, msg)
					continue
				}
			}

			if cond == nil {
				msg := fmt.Sprintf("workflow step %s.%s contains criterion %s with invalid condition", workflow.WorkflowID, step.StepID, c.Type.GetType())

				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("condition not found"),
					Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
				})

				incompleteMessages = append(incompleteMessages, msg)
				continue
			}

			switch cond.Expression.GetType() {
			case expression.ExpressionTypeStatusCode:
				if cond.Operator != criterion.OperatorEQ {
					msg := fmt.Sprintf("workflow step %s.%s contains unsupported operator %s for status code expression", workflow.WorkflowID, step.StepID, cond.Operator)

					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("unsupported operator for status code expression"),
						Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
					})

					incompleteMessages = append(incompleteMessages, msg)
					continue
				}

				if requiredAsserts.statusCode != nil {
					logging.LogWarning(ctx, "multiple status code expressions found, skipping condition "+c.Condition, &validation.Error{
						UnderlyingError: errors.New("multiple status code expressions found"),
						Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
					})
				} else {
					requiredAsserts.statusCode = cond
					requiredAsserts.statusCodeNode = c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode())
				}
			case expression.ExpressionTypeResponse:
				_, responseTarget, expressionParts, _ := cond.Expression.GetParts()
				switch responseTarget {
				case "header":
					if len(expressionParts) != 1 {
						msg := fmt.Sprintf("workflow step %s.%s contains invalid token %s for header", workflow.WorkflowID, step.StepID, strings.Join(expressionParts, "."))

						logging.LogWarning(ctx, msg, &validation.Error{
							UnderlyingError: errors.New("invalid token for header"),
							Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
						})

						incompleteMessages = append(incompleteMessages, msg)
						continue
					}

					token := strings.ToLower(expressionParts[0])
					switch token {
					case "content-type":
						if requiredAsserts.contentType != nil {
							logging.LogWarning(ctx, "multiple content type expressions found, skipping condition "+c.Condition, &validation.Error{
								UnderlyingError: errors.New("multiple content type expressions found"),
								Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
							})
						} else {
							requiredAsserts.contentType = cond
							requiredAsserts.contentTypeNode = c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode())
						}
					default:
						msg := fmt.Sprintf("workflow step %s.%s contains unsupported header token %s", workflow.WorkflowID, step.StepID, token)

						logging.LogWarning(ctx, msg, &validation.Error{
							UnderlyingError: errors.New("header token not currently supported"),
							Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
						})

						incompleteMessages = append(incompleteMessages, msg)
						continue
					}
				case "body":
					requiredAsserts.responseBodiesAssertions = append(requiredAsserts.responseBodiesAssertions, responseBodyAssertion{
						condition: cond,
						node:      c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
					})
				default:
					msg := fmt.Sprintf("workflow step %s.%s contains unsupported response target %s", workflow.WorkflowID, step.StepID, responseTarget)

					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("response target not currently supported"),
						Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
					})

					incompleteMessages = append(incompleteMessages, msg)
					continue
				}
			default:
				msg := fmt.Sprintf("workflow step %s.%s contains unsupported expression type %s", workflow.WorkflowID, step.StepID, cond.Expression.GetType())

				logging.LogWarning(ctx, msg, &validation.Error{
					UnderlyingError: errors.New("expression type not currently supported"),
					Node:            c.GetCore().Condition.GetKeyNodeOrRoot(c.GetRootNode()),
				})

				incompleteMessages = append(incompleteMessages, msg)
				continue
			}
		default:
			msg := fmt.Sprintf("workflow step %s.%s contains unsupported criterion type %s", workflow.WorkflowID, step.StepID, c.Type.GetType())

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("criterion type not currently supported"),
				Node:            c.GetCore().Type.GetKeyNodeOrRoot(c.GetRootNode()),
			})

			incompleteMessages = append(incompleteMessages, msg)
			continue
		}
	}

	assertions := []ast.Assertion{}

	if requiredAsserts.statusCode != nil {
		assertionType, errorMsg := getAssertionType(ctx, requiredAsserts.statusCode, step.GetCore().SuccessCriteria.GetKeyNodeOrRoot(step.GetRootNode()), opts)
		if errorMsg != "" {
			incompleteMessages = append(incompleteMessages, errorMsg)
		} else {
			assertions = append(assertions, ast.Assertion{
				Type:       assertionType,
				TargetType: ast.AssertionTargetStatusCode,
				Target:     opts.op.Response,
				Value:      requiredAsserts.statusCode.Value,
			})
		}
	}

	responseContentType := ""
	if requiredAsserts.contentType != nil {
		responseContentType = strings.Trim(fmt.Sprint(requiredAsserts.contentType.Value), `"`)
	}

	if len(requiredAsserts.responseBodiesAssertions) > 0 {
		if requiredAsserts.statusCode == nil {
			msg := fmt.Sprintf("workflow step %s.%s does not contain $statusCode successCriteria and required for response body assertion", workflow.WorkflowID, step.StepID)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("response body assertion without status code"),
				Node:            step.GetCore().SuccessCriteria.GetKeyNodeOrRoot(step.GetRootNode()),
			})

			incompleteMessages = append(incompleteMessages, msg)
		}

		if requiredAsserts.contentType == nil {
			msg := fmt.Sprintf("workflow step %s.%s does not contain $contentType successCriteria and required for response body assertion", workflow.WorkflowID, step.StepID)

			logging.LogWarning(ctx, msg, &validation.Error{
				UnderlyingError: errors.New("response body assertion without content type"),
				Node:            step.GetCore().SuccessCriteria.GetKeyNodeOrRoot(step.GetRootNode()),
			})

			incompleteMessages = append(incompleteMessages, msg)
		}

		if requiredAsserts.statusCode != nil && requiredAsserts.contentType != nil {
			// TODO handle other types of response body assertions ie using json-pointer
			for _, responseBodyAssert := range requiredAsserts.responseBodiesAssertions {
				foundSubResponseIdx := slices.IndexFunc(opts.op.Response.Responses, func(e *ast.SubResponse) bool {
					return slices.Contains(e.Code, fmt.Sprintf("%s", requiredAsserts.statusCode.Value))
				})

				assertionType, errorMsg := getAssertionType(ctx, responseBodyAssert.condition, responseBodyAssert.node, opts)
				if errorMsg != "" {
					incompleteMessages = append(incompleteMessages, errorMsg)
					continue
				}

				if foundSubResponseIdx == -1 {
					msg := fmt.Sprintf("workflow step %s.%s referencing operation %s does not contain response with status code %s", workflow.WorkflowID, step.StepID, opts.op.ID, requiredAsserts.statusCode.Value)

					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("response not found with status code"),
						Node:            requiredAsserts.statusCodeNode,
					})

					incompleteMessages = append(incompleteMessages, msg)
					continue
				}

				contentTypeValue := strings.Trim(fmt.Sprintf("%s", requiredAsserts.contentType.Value), `"`)
				foundContentIdx := slices.IndexFunc(opts.op.Response.Responses[foundSubResponseIdx].Content, func(e *ast.ResponseBodyContent) bool {
					return e.ContentType == contentTypeValue
				})

				if foundContentIdx == -1 {
					msg := fmt.Sprintf("workflow step %s.%s referencing operation %s does not contain response body with content type %s", workflow.WorkflowID, step.StepID, opts.op.ID, requiredAsserts.contentType.Value)

					logging.LogWarning(ctx, msg, &validation.Error{
						UnderlyingError: errors.New("content type not found for response"),
						Node:            requiredAsserts.contentTypeNode,
					})

					incompleteMessages = append(incompleteMessages, msg)
					continue
				}

				// TODO this is not likely to work for all possible schema types and content types
				rawJSON := fmt.Sprint(responseBodyAssert.condition.Value)

				isSetAssertion := false
				// This is a special case where the end user is trying to just check anything was received.
				// The null or empty string might not actually represent the correct type but this will
				// just short circuit the checks to NotNil or NotEmpty assertions
				if assertionType == ast.AssertionTypeNotEqual && (rawJSON == `""` || rawJSON == "null") {
					isSetAssertion = true
				}

				content := opts.op.Response.Responses[foundSubResponseIdx].Content[foundContentIdx]
				var target any = content.Content

				jp := responseBodyAssert.condition.Expression.GetJSONPointer()
				if jp != "" {
					jpTarget, err := jsonpointer.GetTarget(target, jp)
					if err != nil {
						msg := fmt.Sprintf("workflow step %s.%s contains unresolvable json-pointer %s for response body: %s", workflow.WorkflowID, step.StepID, jp, err.Error())
						logging.LogWarning(ctx, msg, &validation.Error{
							UnderlyingError: errors.New("json-pointer not currently supported on response body"),
							Node:            responseBodyAssert.node,
						})
						incompleteMessages = append(incompleteMessages, msg)
						continue
					}

					if jpTargetTyp, ok := jpTarget.(*ast.FieldDef); ok {
						if jpTargetTyp.IsAdditionalProperties {
							jpTarget = jpTargetTyp.Type.ItemType
						}
					}

					target = jpTarget
				}

				var example *ast.Example

				if !isSetAssertion {
					var node yaml.Node
					if err := yaml.Unmarshal([]byte(rawJSON), &node); err != nil {
						msg := fmt.Sprintf("workflow step %s.%s referencing operation %s failed to unmarshal json %s with error [%s] for response body assertion", workflow.WorkflowID, step.StepID, opts.op.ID, rawJSON, err.Error())

						logging.LogWarning(ctx, msg, &validation.Error{
							UnderlyingError: errors.New("failed to unmarshal json response body condition"),
							Node:            responseBodyAssert.node,
						})

						incompleteMessages = append(incompleteMessages, msg)
						continue
					}

					example = ast.NewExample(fmt.Sprintf("%s[%d]", workflow.WorkflowID, opts.stepIdx), "", &node)
				}

				assert := ast.ResponseBodyAssertion{
					Path:    string(jp),
					Value:   example,
					Content: content,
				}

				assertions = append(assertions, ast.Assertion{
					Type:       assertionType,
					TargetType: ast.AssertionTargetResponseBody,
					Target:     target,
					Value:      assert,
				})
			}
		}
	}

	return assertions, responseContentType, incompleteMessages
}

func getAssertionType(ctx context.Context, condition *criterion.Condition, node *yaml.Node, opts getOpAssertionsOpts) (ast.AssertionType, string) {
	var assertionType ast.AssertionType

	switch condition.Operator {
	case criterion.OperatorEQ:
		assertionType = ast.AssertionTypeEqual
	case criterion.OperatorNE:
		assertionType = ast.AssertionTypeNotEqual
	default:
		msg := fmt.Sprintf("workflow step %s.%s contains unsupported operator %s for response body assertion", opts.workflow.WorkflowID, opts.step.StepID, condition.Operator)
		logging.LogWarning(ctx, msg, &validation.Error{
			UnderlyingError: errors.New("unsupported operator for response body assertion"),
			Node:            node,
		})
		return "", msg
	}

	return assertionType, ""
}
