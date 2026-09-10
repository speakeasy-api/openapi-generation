package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi/arazzo/criterion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// testCondition creates a criterion.Condition using the same method as UnmarshalYAML
// This ensures the unexported rawCondition field is properly set
func testCondition(t *testing.T, conditionStr string) *criterion.Condition {
	t.Helper()

	c := &criterion.Criterion{
		Condition: conditionStr,
	}

	condition, err := c.GetCondition()

	if err != nil {
		t.Fatalf("failed to create test condition from '%s': %v", conditionStr, err)
	}

	return condition
}

func TestParsePolling(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected *Polling
		wantErr  bool
	}{
		{
			name: "valid sequence with single polling option",
			yaml: `
- name: "default"
  successCriteria:
    - condition: "$statusCode == 200"`,
			expected: &Polling{
				Options: PollingOptions{
					{
						Name:            "default",
						DelaySeconds:    ptr(int64(1)),
						IntervalSeconds: ptr(int64(1)),
						LimitCount:      ptr(int64(60)),
						SuccessCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$statusCode == 200"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid sequence with multiple polling options",
			yaml: `
- name: "fast"
  delaySeconds: 0
  intervalSeconds: 1
  limitCount: 10
  successCriteria:
    - condition: "$statusCode == 200"
- name: "slow"
  delaySeconds: 5
  intervalSeconds: 10
  limitCount: 30
  successCriteria:
    - condition: "$response.body#/status == 'complete'"`,
			expected: &Polling{
				Options: PollingOptions{
					{
						Name:            "fast",
						DelaySeconds:    ptr(int64(0)),
						IntervalSeconds: ptr(int64(1)),
						LimitCount:      ptr(int64(10)),
						SuccessCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$statusCode == 200"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
					},
					{
						Name:            "slow",
						DelaySeconds:    ptr(int64(5)),
						IntervalSeconds: ptr(int64(10)),
						LimitCount:      ptr(int64(30)),
						SuccessCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$response.body#/status == 'complete'"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid polling option with multiple success criteria",
			yaml: `
- name: "multi-criteria"
  successCriteria:
    - condition: "$statusCode == 200"
    - condition: "$response.body#/ready == true"`,
			expected: &Polling{
				Options: PollingOptions{
					{
						Name:            "multi-criteria",
						DelaySeconds:    ptr(int64(1)),
						IntervalSeconds: ptr(int64(1)),
						LimitCount:      ptr(int64(60)),
						SuccessCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$statusCode == 200"),
								Type:      criterion.CriterionTypeSimple,
							},
							{
								Condition: testCondition(t, "$response.body#/ready == true"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid polling option with failure criteria",
			yaml: `
- name: "with-failure"
  successCriteria:
    - condition: "$response.body#/status == 'success'"
  failureCriteria:
    - condition: "$response.body#/status == 'failed'"`,
			expected: &Polling{
				Options: PollingOptions{
					{
						Name:            "with-failure",
						DelaySeconds:    ptr(int64(1)),
						IntervalSeconds: ptr(int64(1)),
						LimitCount:      ptr(int64(60)),
						SuccessCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$response.body#/status == 'success'"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
						FailureCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$response.body#/status == 'failed'"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid polling option with custom timing values",
			yaml: `
- name: "custom-timing"
  delaySeconds: 10
  intervalSeconds: 5
  limitCount: 100
  successCriteria:
    - condition: "$response.body#/done == true"`,
			expected: &Polling{
				Options: PollingOptions{
					{
						Name:            "custom-timing",
						DelaySeconds:    ptr(int64(10)),
						IntervalSeconds: ptr(int64(5)),
						LimitCount:      ptr(int64(100)),
						SuccessCriteria: PollingCriteria{
							{
								Condition: testCondition(t, "$response.body#/done == true"),
								Type:      criterion.CriterionTypeSimple,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty sequence",
			yaml: `[]`,
			expected: &Polling{
				Options: PollingOptions{},
			},
			wantErr: false,
		},
		{
			name: "invalid - missing name",
			yaml: `
- successCriteria:
    - condition: "$statusCode == 200"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid - empty name",
			yaml: `
- name: ""
  successCriteria:
    - condition: "$statusCode == 200"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid - missing success criteria",
			yaml: `
- name: "no-criteria"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid - empty success criteria",
			yaml: `
- name: "empty-criteria"
  successCriteria: []`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid - success criterion without condition",
			yaml: `
- name: "no-condition"
  successCriteria:
    - other: "value"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid scalar node - string",
			yaml:     `"invalid_scalar_value"`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid null node",
			yaml:     `null`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid boolean node",
			yaml:     `false`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid mapping node",
			yaml:     `{}`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid numeric node",
			yaml:     `42`,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid sequence with non-mapping items",
			yaml: `
- "string item"
- 123`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yaml), &node)
			require.NoError(t, err)

			// The root node is a document node, we need the content node
			require.Len(t, node.Content, 1, "Expected exactly one content node")
			contentNode := node.Content[0]

			e := &Extensions{}
			result, err := e.parsePolling(contentNode)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
