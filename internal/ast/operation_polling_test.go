package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/speakeasy-api/openapi/expression"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseOperation_pollingAssertionsFromExtensionsPollingCriteria(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		operation          *BaseOperation
		extensionCriteria  extensions.PollingCriteria
		expectedAssertions Assertions
		expectError        bool
		errorContains      string
	}{
		{
			name:               "nil operation",
			operation:          nil,
			extensionCriteria:  extensions.PollingCriteria{},
			expectedAssertions: nil,
			expectError:        true,
			errorContains:      "operation is nil",
		},
		{
			name:               "nil criteria",
			operation:          &BaseOperation{},
			extensionCriteria:  nil,
			expectedAssertions: nil,
			expectError:        false,
		},
		{
			name:               "empty criteria",
			operation:          &BaseOperation{},
			extensionCriteria:  extensions.PollingCriteria{},
			expectedAssertions: Assertions{},
			expectError:        false,
		},
		{
			name:      "nil criterion in criteria",
			operation: &BaseOperation{},
			extensionCriteria: extensions.PollingCriteria{
				nil,
			},
			expectedAssertions: Assertions{},
			expectError:        false,
		},
		{
			name:      "criterion with nil condition",
			operation: &BaseOperation{},
			extensionCriteria: extensions.PollingCriteria{
				{Condition: nil},
			},
			expectedAssertions: Assertions{},
			expectError:        false,
		},
		{
			name: "status code equals condition",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content:             &FieldDef{Name: "body"},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, "$statusCode == 200"),
			},
			expectedAssertions: Assertions{
				{
					TargetType: AssertionTargetStatusCode,
					Type:       AssertionTypeEqual,
					Value:      "200",
				},
			},
			expectError: false,
		},
		{
			name: "status code not equals condition",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"500"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content:             &FieldDef{Name: "body"},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, "$statusCode != 500"),
			},
			expectedAssertions: Assertions{
				{
					TargetType: AssertionTargetStatusCode,
					Type:       AssertionTypeNotEqual,
					Value:      "500",
				},
			},
			expectError: false,
		},
		{
			name: "response body equals condition requires prior status code",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content:             &FieldDef{Name: "body"},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, `$response.body#/status == "completed"`),
			},
			expectedAssertions: nil,
			expectError:        true,
			errorContains:      "polling response body assertions require a prior status code assertion",
		},
		{
			name: "response body equals condition with prior status code",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content: &FieldDef{
										Name: "body",
										Type: &TypeDef{
											Type: DataTypeClass,
											Fields: []*FieldDef{
												{Name: "status", OriginalName: "status", Type: &TypeDef{Type: DataTypeString}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, "$statusCode == 200"),
				testSimplePollingCriterion(t, `$response.body#/status == "completed"`),
			},
			expectedAssertions: Assertions{
				{
					TargetType: AssertionTargetStatusCode,
					Type:       AssertionTypeEqual,
					Value:      "200",
				},
				{
					TargetType: AssertionTargetResponseBody,
					Type:       AssertionTypeEqual,
				},
			},
			expectError: false,
		},
		{
			name: "response body not equals condition with prior status code",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content: &FieldDef{
										Name: "body",
										Type: &TypeDef{
											Type: DataTypeClass,
											Fields: []*FieldDef{
												{Name: "status", OriginalName: "status", Type: &TypeDef{Type: DataTypeString}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, "$statusCode == 200"),
				testSimplePollingCriterion(t, `$response.body#/status != "failed"`),
			},
			expectedAssertions: Assertions{
				{
					TargetType: AssertionTargetStatusCode,
					Type:       AssertionTypeEqual,
					Value:      "200",
				},
				{
					TargetType: AssertionTargetResponseBody,
					Type:       AssertionTypeNotEqual,
				},
			},
			expectError: false,
		},
		{
			name: "unsupported assertion type - less than",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content:             &FieldDef{Name: "body"},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, "$statusCode < 300"),
			},
			expectedAssertions: nil,
			expectError:        true,
			errorContains:      "unsupported operator for assertion type",
		},
		{
			name: "unsupported assertion target - request",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content:             &FieldDef{Name: "body"},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, `$request.body#/id == "123"`),
			},
			expectedAssertions: nil,
			expectError:        true,
			errorContains:      "unsupported target for assertion",
		},
		{
			name: "regex response body condition with prior status code",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content: &FieldDef{
										Name: "body",
										Type: &TypeDef{
											Type: DataTypeClass,
											Fields: []*FieldDef{
												{Name: "status", OriginalName: "status", Type: &TypeDef{Type: DataTypeString}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testSimplePollingCriterion(t, "$statusCode == 200"),
				testRegexPollingCriterion(t, "$response.body#/status", "^(completed|ready-for-next-step)$"),
			},
			expectedAssertions: Assertions{
				{
					TargetType: AssertionTargetStatusCode,
					Type:       AssertionTypeEqual,
					Value:      "200",
				},
				{
					TargetType: AssertionTargetResponseBody,
					Type:       AssertionTypeRegex,
				},
			},
			expectError: false,
		},
		{
			name: "regex response body condition requires prior status code",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content: &FieldDef{
										Name: "body",
										Type: &TypeDef{
											Type: DataTypeClass,
											Fields: []*FieldDef{
												{Name: "status", OriginalName: "status", Type: &TypeDef{Type: DataTypeString}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testRegexPollingCriterion(t, "$response.body#/status", "^(completed|ready-for-next-step)$"),
			},
			expectedAssertions: nil,
			expectError:        true,
			errorContains:      "polling response body assertions require a prior status code assertion",
		},
		{
			name: "regex status code condition",
			operation: &BaseOperation{
				Response: &Response{
					Responses: SubResponses{
						{
							Code: []string{"200"},
							Content: []*ResponseBodyContent{
								{
									SerializationMethod: "json",
									ContentType:         "application/json",
									Content:             &FieldDef{Name: "body"},
								},
							},
						},
					},
				},
			},
			extensionCriteria: extensions.PollingCriteria{
				testRegexPollingCriterion(t, "$statusCode", "^2[0-9]{2}$"),
			},
			expectedAssertions: Assertions{
				{
					TargetType: AssertionTargetStatusCode,
					Type:       AssertionTypeRegex,
					Value:      "^2[0-9]{2}$",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.operation.pollingAssertionsFromExtensionsPollingCriteria(tt.extensionCriteria)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)

				if tt.expectedAssertions == nil {
					assert.Nil(t, result)
				} else {
					require.Len(t, result, len(tt.expectedAssertions))

					for i, expected := range tt.expectedAssertions {
						assert.Equal(t, expected.TargetType, result[i].TargetType, "assertion %d target type mismatch", i)
						assert.Equal(t, expected.Type, result[i].Type, "assertion %d type mismatch", i)

						// Only check value for status code assertions since response body assertions have complex values
						if expected.TargetType == AssertionTargetStatusCode {
							assert.Equal(t, expected.Value, result[i].Value, "assertion %d value mismatch", i)
						}
					}
				}
			}
		})
	}
}

// Creates a simple type PollingCriterion from a condition string for testing.
func testSimplePollingCriterion(t *testing.T, conditionStr string) *extensions.PollingCriterion {
	t.Helper()

	return &extensions.PollingCriterion{
		Condition: testCondition(t, conditionStr),
		Type:      criterion.CriterionTypeSimple,
	}
}

// Creates a regex type PollingCriterion for testing.
func testRegexPollingCriterion(t *testing.T, context string, pattern string) *extensions.PollingCriterion {
	t.Helper()

	ctx := expression.Expression(context)

	return &extensions.PollingCriterion{
		Condition: &criterion.Condition{
			Value: pattern,
		},
		Context: &ctx,
		Type:    criterion.CriterionTypeRegex,
	}
}
