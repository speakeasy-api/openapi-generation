package ast

import (
	"testing"
)

func TestSDKsDeleteByFieldName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sdks      SDKs
		fieldName string
		expected  SDKs
	}{
		{
			name:      "delete from empty collection",
			sdks:      SDKs{},
			fieldName: "nonexistent",
			expected:  SDKs{},
		},
		{
			name: "delete non-existent SDK",
			sdks: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
			},
			fieldName: "nonexistent",
			expected: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
			},
		},
		{
			name: "delete first SDK",
			sdks: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
				{FieldName: "sdk3"},
			},
			fieldName: "sdk1",
			expected: SDKs{
				{FieldName: "sdk2"},
				{FieldName: "sdk3"},
			},
		},
		{
			name: "delete middle SDK",
			sdks: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
				{FieldName: "sdk3"},
			},
			fieldName: "sdk2",
			expected: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk3"},
			},
		},
		{
			name: "delete last SDK",
			sdks: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
				{FieldName: "sdk3"},
			},
			fieldName: "sdk3",
			expected: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
			},
		},
		{
			name: "delete only SDK",
			sdks: SDKs{
				{FieldName: "sdk1"},
			},
			fieldName: "sdk1",
			expected:  SDKs{},
		},
		{
			name: "delete first matching SDK when duplicates exist",
			sdks: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "duplicate"},
				{FieldName: "sdk2"},
				{FieldName: "duplicate"},
			},
			fieldName: "duplicate",
			expected: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
				{FieldName: "duplicate"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sdks := make(SDKs, len(tt.sdks))
			copy(sdks, tt.sdks)

			sdks.DeleteByFieldName(tt.fieldName)

			if len(sdks) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(sdks))
			}

			for i, sdk := range sdks {
				if i >= len(tt.expected) || sdk.FieldName != tt.expected[i].FieldName {
					t.Errorf("expected SDK field name %s at index %d, got %s", tt.expected[i].FieldName, i, sdk.FieldName)
				}
			}
		})
	}
}

func TestSDKsGetByFieldName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sdks      SDKs
		fieldName string
		expected  *SDK
	}{
		{
			name:      "get from empty collection",
			sdks:      SDKs{},
			fieldName: "nonexistent",
			expected:  nil,
		},
		{
			name:      "get from nil collection",
			sdks:      nil,
			fieldName: "nonexistent",
			expected:  nil,
		},
		{
			name: "get non-existent SDK",
			sdks: SDKs{
				{FieldName: "sdk1"},
				{FieldName: "sdk2"},
			},
			fieldName: "nonexistent",
			expected:  nil,
		},
		{
			name: "get first SDK",
			sdks: SDKs{
				{FieldName: "sdk1", Group: "group1"},
				{FieldName: "sdk2", Group: "group2"},
				{FieldName: "sdk3", Group: "group3"},
			},
			fieldName: "sdk1",
			expected:  &SDK{FieldName: "sdk1", Group: "group1"},
		},
		{
			name: "get middle SDK",
			sdks: SDKs{
				{FieldName: "sdk1", Group: "group1"},
				{FieldName: "sdk2", Group: "group2"},
				{FieldName: "sdk3", Group: "group3"},
			},
			fieldName: "sdk2",
			expected:  &SDK{FieldName: "sdk2", Group: "group2"},
		},
		{
			name: "get last SDK",
			sdks: SDKs{
				{FieldName: "sdk1", Group: "group1"},
				{FieldName: "sdk2", Group: "group2"},
				{FieldName: "sdk3", Group: "group3"},
			},
			fieldName: "sdk3",
			expected:  &SDK{FieldName: "sdk3", Group: "group3"},
		},
		{
			name: "get only SDK",
			sdks: SDKs{
				{FieldName: "sdk1", Group: "group1"},
			},
			fieldName: "sdk1",
			expected:  &SDK{FieldName: "sdk1", Group: "group1"},
		},
		{
			name: "get first matching SDK when duplicates exist",
			sdks: SDKs{
				{FieldName: "duplicate", Group: "group1"},
				{FieldName: "sdk1"},
				{FieldName: "duplicate", Group: "group2"},
			},
			fieldName: "duplicate",
			expected:  &SDK{FieldName: "duplicate", Group: "group1"},
		},
		{
			name: "case-sensitive field name matching",
			sdks: SDKs{
				{FieldName: "SDK1"},
				{FieldName: "sdk1"},
			},
			fieldName: "sdk1",
			expected:  &SDK{FieldName: "sdk1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.sdks.GetByFieldName(tt.fieldName)

			if tt.expected != nil {
				if result == nil {
					t.Errorf("expected SDK with field name %s, got none", tt.fieldName)
				} else if result.FieldName != tt.expected.FieldName {
					t.Errorf("expected SDK with field name %s, got %s", tt.expected.FieldName, result.FieldName)
				}
			} else {
				if result != nil {
					t.Errorf("expected no SDK with field name %s, got %s", tt.fieldName, result.FieldName)
				}
			}
		})
	}
}

func TestSDKsHasOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sdks     SDKs
		expected bool
	}{
		{
			name:     "empty collection",
			sdks:     SDKs{},
			expected: false,
		},
		{
			name:     "nil collection",
			sdks:     nil,
			expected: false,
		},
		{
			name: "single SDK with no operations",
			sdks: SDKs{
				{Operations: nil},
			},
			expected: false,
		},
		{
			name: "single SDK with empty operations",
			sdks: SDKs{
				{Operations: []*Operation{}},
			},
			expected: false,
		},
		{
			name: "single SDK with operations",
			sdks: SDKs{
				{Operations: []*Operation{{}, {}}},
			},
			expected: true,
		},
		{
			name: "multiple SDKs all without operations",
			sdks: SDKs{
				{Operations: nil},
				{Operations: []*Operation{}},
				{Operations: nil},
			},
			expected: false,
		},
		{
			name: "multiple SDKs with first having operations",
			sdks: SDKs{
				{Operations: []*Operation{{}}},
				{Operations: nil},
				{Operations: []*Operation{}},
			},
			expected: true,
		},
		{
			name: "multiple SDKs with middle having operations",
			sdks: SDKs{
				{Operations: nil},
				{Operations: []*Operation{{}}},
				{Operations: []*Operation{}},
			},
			expected: true,
		},
		{
			name: "multiple SDKs with last having operations",
			sdks: SDKs{
				{Operations: nil},
				{Operations: []*Operation{}},
				{Operations: []*Operation{{}}},
			},
			expected: true,
		},
		{
			name: "SDK with no direct operations but subSDK has operations",
			sdks: SDKs{
				{
					Operations: nil,
					SubSDKs: SDKs{
						{Operations: []*Operation{{}}},
					},
				},
			},
			expected: true,
		},
		{
			name: "SDK with operations and subSDK with operations",
			sdks: SDKs{
				{
					Operations: []*Operation{{}},
					SubSDKs: SDKs{
						{Operations: []*Operation{{}}},
					},
				},
			},
			expected: true,
		},
		{
			name: "nested subSDKs with operations at deep level",
			sdks: SDKs{
				{
					Operations: nil,
					SubSDKs: SDKs{
						{
							Operations: nil,
							SubSDKs: SDKs{
								{
									Operations: nil,
									SubSDKs: SDKs{
										{Operations: []*Operation{{}}},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "complex hierarchy with operations in multiple places",
			sdks: SDKs{
				{
					Operations: nil,
					SubSDKs: SDKs{
						{Operations: []*Operation{{}}},
						{Operations: nil},
					},
				},
				{
					Operations: []*Operation{{}, {}},
				},
				{
					Operations: nil,
					SubSDKs: SDKs{
						{
							Operations: nil,
							SubSDKs: SDKs{
								{Operations: []*Operation{{}}},
							},
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

			result := tt.sdks.HasOperations()

			if result != tt.expected {
				t.Errorf("expected %t, got: %t", tt.expected, result)
			}
		})
	}
}
