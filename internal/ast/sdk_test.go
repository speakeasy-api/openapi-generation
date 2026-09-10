package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi/openapi"
)

func TestSDKHasOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sdk      *SDK
		expected bool
	}{
		{
			name: "SDK with no operations and no subSDKs",
			sdk: &SDK{
				Operations: nil,
				SubSDKs:    nil,
			},
			expected: false,
		},
		{
			name: "SDK with empty operations and empty subSDKs",
			sdk: &SDK{
				Operations: []*Operation{},
				SubSDKs:    []*SDK{},
			},
			expected: false,
		},
		{
			name: "SDK with operations",
			sdk: &SDK{
				Operations: []*Operation{
					{},
					{},
				},
				SubSDKs: nil,
			},
			expected: true,
		},
		{
			name: "SDK with no operations but subSDK has operations",
			sdk: &SDK{
				Operations: nil,
				SubSDKs: []*SDK{
					{
						Operations: []*Operation{
							{},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "SDK with no operations and subSDK with no operations",
			sdk: &SDK{
				Operations: nil,
				SubSDKs: []*SDK{
					{
						Operations: nil,
						SubSDKs:    nil,
					},
				},
			},
			expected: false,
		},
		{
			name: "SDK with operations and subSDKs with operations",
			sdk: &SDK{
				Operations: []*Operation{
					{},
				},
				SubSDKs: []*SDK{
					{
						Operations: []*Operation{
							{},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "SDK with nested subSDKs with operations",
			sdk: &SDK{
				Operations: nil,
				SubSDKs: []*SDK{
					{
						Operations: nil,
						SubSDKs: []*SDK{
							{
								Operations: []*Operation{
									{},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "SDK with multiple subSDKs, one with operations",
			sdk: &SDK{
				Operations: nil,
				SubSDKs: []*SDK{
					{
						Operations: nil,
						SubSDKs:    nil,
					},
					{
						Operations: []*Operation{
							{},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.sdk.HasOperations()

			if result != tt.expected {
				t.Errorf("expected %t, got: %t", tt.expected, result)
			}
		})
	}
}

func TestSortSubSDKsByTagOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		sdk           *SDK
		tags          []*openapi.Tag
		expectedOrder []string
	}{
		{
			name: "sorts SubSDKs according to tag order",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "pets"},
					{FieldName: "users"},
					{FieldName: "orders"},
				},
			},
			tags: []*openapi.Tag{
				{Name: "users"},
				{Name: "orders"},
				{Name: "pets"},
			},
			expectedOrder: []string{"users", "orders", "pets"},
		},
		{
			name: "handles SubSDKs not in tag list",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "pets"},
					{FieldName: "users"},
					{FieldName: "orders"},
					{FieldName: "inventory"},
				},
			},
			tags: []*openapi.Tag{
				{Name: "users"},
				{Name: "orders"},
			},
			expectedOrder: []string{"users", "orders", "pets", "inventory"},
		},
		{
			name: "handles tags not matching any SubSDKs",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "pets"},
					{FieldName: "users"},
				},
			},
			tags: []*openapi.Tag{
				{Name: "users"},
				{Name: "nonexistent"},
				{Name: "pets"},
			},
			expectedOrder: []string{"users", "pets"},
		},
		{
			name: "empty SubSDKs",
			sdk: &SDK{
				SubSDKs: []*SDK{},
			},
			tags: []*openapi.Tag{
				{Name: "users"},
				{Name: "pets"},
			},
			expectedOrder: []string{},
		},
		{
			name: "empty tags",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "pets"},
					{FieldName: "users"},
				},
			},
			tags:          []*openapi.Tag{},
			expectedOrder: []string{"pets", "users"},
		},
		{
			name: "nil tags",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "pets"},
					{FieldName: "users"},
				},
			},
			tags:          nil,
			expectedOrder: []string{"pets", "users"},
		},
		{
			name: "single SubSDK with matching tag",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "users"},
				},
			},
			tags: []*openapi.Tag{
				{Name: "users"},
			},
			expectedOrder: []string{"users"},
		},
		{
			name: "preserves SubSDKs order when no tags match",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "alpha"},
					{FieldName: "beta"},
					{FieldName: "gamma"},
				},
			},
			tags: []*openapi.Tag{
				{Name: "delta"},
				{Name: "epsilon"},
			},
			expectedOrder: []string{"alpha", "beta", "gamma"},
		},
		{
			name: "partial tag match with mixed ordering",
			sdk: &SDK{
				SubSDKs: []*SDK{
					{FieldName: "delta"},
					{FieldName: "alpha"},
					{FieldName: "gamma"},
					{FieldName: "beta"},
				},
			},
			tags: []*openapi.Tag{
				{Name: "beta"},
				{Name: "gamma"},
			},
			expectedOrder: []string{"beta", "gamma", "delta", "alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.sdk.SortSubSDKsByTags(tt.tags)

			if len(tt.sdk.SubSDKs) != len(tt.expectedOrder) {
				t.Errorf("expected %d SubSDKs, got %d", len(tt.expectedOrder), len(tt.sdk.SubSDKs))
				return
			}

			for i, sdk := range tt.sdk.SubSDKs {
				if sdk.FieldName != tt.expectedOrder[i] {
					t.Errorf("expected SubSDK at index %d to be %q, got %q", i, tt.expectedOrder[i], sdk.FieldName)
				}
			}
		})
	}
}
