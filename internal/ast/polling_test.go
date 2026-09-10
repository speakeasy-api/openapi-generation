package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPolling_Clone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		polling  *Polling
		testFunc func(t *testing.T, original, cloned *Polling)
	}{
		{
			name:    "nil polling",
			polling: nil,
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.Nil(t, cloned)
			},
		},
		{
			name: "empty polling",
			polling: &Polling{
				Options: PollingOptions{},
			},
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Empty(t, cloned.Options)
				assert.NotSame(t, &original.Options, &cloned.Options)
			},
		},
		{
			name: "polling with single option",
			polling: &Polling{
				Options: PollingOptions{
					{
						Name:            "default",
						DelaySeconds:    ptr(int64(1)),
						IntervalSeconds: ptr(int64(5)),
						LimitCount:      ptr(int64(60)),
						SuccessCriteria: Assertions{
							{
								TargetType: AssertionTargetStatusCode,
								Type:       AssertionTypeEqual,
								Value:      "200",
							},
						},
					},
				},
			},
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.Options, 1)
				assert.NotSame(t, original.Options[0], cloned.Options[0])

				// Check that the option fields are cloned properly
				assert.Equal(t, original.Options[0].Name, cloned.Options[0].Name)
				assert.Equal(t, *original.Options[0].DelaySeconds, *cloned.Options[0].DelaySeconds)
				assert.NotSame(t, original.Options[0].DelaySeconds, cloned.Options[0].DelaySeconds)
				assert.Equal(t, *original.Options[0].IntervalSeconds, *cloned.Options[0].IntervalSeconds)
				assert.NotSame(t, original.Options[0].IntervalSeconds, cloned.Options[0].IntervalSeconds)
				assert.Equal(t, *original.Options[0].LimitCount, *cloned.Options[0].LimitCount)
				assert.NotSame(t, original.Options[0].LimitCount, cloned.Options[0].LimitCount)

				// Check success criteria
				assert.Len(t, cloned.Options[0].SuccessCriteria, 1)
				assert.NotSame(t, original.Options[0].SuccessCriteria[0], cloned.Options[0].SuccessCriteria[0])
				assert.Equal(t, original.Options[0].SuccessCriteria[0].TargetType, cloned.Options[0].SuccessCriteria[0].TargetType)
				assert.Equal(t, original.Options[0].SuccessCriteria[0].Type, cloned.Options[0].SuccessCriteria[0].Type)
				assert.Equal(t, original.Options[0].SuccessCriteria[0].Value, cloned.Options[0].SuccessCriteria[0].Value)

				// Verify deep copy by modifying original
				*original.Options[0].DelaySeconds = 10
				assert.Equal(t, int64(1), *cloned.Options[0].DelaySeconds)
				original.Options[0].Name = "modified"
				assert.Equal(t, "default", cloned.Options[0].Name)
			},
		},
		{
			name: "polling with multiple options",
			polling: &Polling{
				Options: PollingOptions{
					{
						Name:            "fast",
						DelaySeconds:    ptr(int64(0)),
						IntervalSeconds: ptr(int64(1)),
						LimitCount:      ptr(int64(10)),
						SuccessCriteria: Assertions{
							{
								TargetType: AssertionTargetStatusCode,
								Type:       AssertionTypeEqual,
								Value:      "200",
							},
						},
					},
					{
						Name:            "slow",
						DelaySeconds:    ptr(int64(5)),
						IntervalSeconds: ptr(int64(10)),
						LimitCount:      ptr(int64(30)),
						SuccessCriteria: Assertions{
							{
								TargetType: AssertionTargetResponseBody,
								Type:       AssertionTypeEqual,
								Value:      "true",
							},
						},
						FailureCriteria: Assertions{
							{
								TargetType: AssertionTargetResponseBody,
								Type:       AssertionTypeEqual,
								Value:      "true",
							},
						},
					},
				},
			},
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.Options, 2)

				// Check that each option is cloned
				for i := range original.Options {
					assert.NotSame(t, original.Options[i], cloned.Options[i])
					assert.Equal(t, original.Options[i].Name, cloned.Options[i].Name)
				}

				// Verify deep copy by modifying original
				original.Options = append(original.Options, &PollingOption{Name: "new"})
				assert.Len(t, cloned.Options, 2)
			},
		},
		{
			name: "polling with nil pointers in options",
			polling: &Polling{
				Options: PollingOptions{
					{
						Name:            "minimal",
						DelaySeconds:    nil,
						IntervalSeconds: nil,
						LimitCount:      nil,
						SuccessCriteria: Assertions{
							{
								TargetType: AssertionTargetResponseBody,
								Type:       AssertionTypeEqual,
								Value:      "true",
							},
						},
					},
				},
			},
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.Options, 1)
				assert.Nil(t, cloned.Options[0].DelaySeconds)
				assert.Nil(t, cloned.Options[0].IntervalSeconds)
				assert.Nil(t, cloned.Options[0].LimitCount)
			},
		},
		{
			name: "polling with multiple criteria",
			polling: &Polling{
				Options: PollingOptions{
					{
						Name:            "multi-criteria",
						DelaySeconds:    ptr(int64(2)),
						IntervalSeconds: ptr(int64(3)),
						LimitCount:      ptr(int64(20)),
						SuccessCriteria: Assertions{
							{
								TargetType: AssertionTargetStatusCode,
								Type:       AssertionTypeEqual,
								Value:      "200",
							},
							{
								TargetType: AssertionTargetResponseBody,
								Type:       AssertionTypeEqual,
								Value:      "true",
							},
						},
						FailureCriteria: Assertions{
							{
								TargetType: AssertionTargetStatusCode,
								Type:       AssertionTypeEqual,
								Value:      "500",
							},
						},
					},
				},
			},
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Len(t, cloned.Options, 1)

				// Check success criteria are cloned
				assert.Len(t, cloned.Options[0].SuccessCriteria, 2)
				assert.NotNil(t, cloned.Options[0].SuccessCriteria[0])
				assert.NotNil(t, cloned.Options[0].SuccessCriteria[1])

				// Check failure criteria are cloned
				assert.Len(t, cloned.Options[0].FailureCriteria, 1)
				assert.NotNil(t, cloned.Options[0].FailureCriteria[0])

				// Verify deep copy of criteria
				for i := range original.Options[0].SuccessCriteria {
					assert.NotSame(t, original.Options[0].SuccessCriteria[i], cloned.Options[0].SuccessCriteria[i])
				}
				for i := range original.Options[0].FailureCriteria {
					assert.NotSame(t, original.Options[0].FailureCriteria[i], cloned.Options[0].FailureCriteria[i])
				}
			},
		},
		{
			name: "polling with nil options",
			polling: &Polling{
				Options: nil,
			},
			testFunc: func(t *testing.T, original, cloned *Polling) {
				t.Helper()
				assert.NotNil(t, cloned)
				assert.NotSame(t, original, cloned)
				assert.Nil(t, cloned.Options)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cloned := tt.polling.Clone()
			tt.testFunc(t, tt.polling, cloned)
		})
	}
}
