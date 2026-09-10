package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestFrameworkTypeResourceModel_ToSharedFrameworkTypeRequest_NullableMaps tests that
// nullable map fields with default: null are properly populated in the SDK request,
// rather than being left as nil due to variable shadowing.
//
// This test verifies the fix for the variable shadowing bug where using `:=` inside
// an if block would create a new local variable instead of assigning to the outer scope.
func TestFrameworkTypeResourceModel_ToSharedFrameworkTypeRequest_NullableMaps(t *testing.T) {
	ctx := context.Background()

	// Create a model with nullable map fields populated
	model := &FrameworkTypeResourceModel{
		Bool: types.BoolValue(true),

		// Test nullable map with string values
		MapNullString: map[string]types.String{
			"key1": types.StringValue("value1"),
			"key2": types.StringValue("value2"),
		},

		// Test nullable map with bool values
		MapNullBool: map[string]types.Bool{
			"enabled":  types.BoolValue(true),
			"disabled": types.BoolValue(false),
		},

		// Test nullable map with int64 values
		MapNullInt64: map[string]types.Int64{
			"count": types.Int64Value(42),
			"limit": types.Int64Value(100),
		},

		// Test nullable map with JSON/any values
		MapNullAny: map[string]jsontypes.Normalized{
			"key1": jsontypes.NewNormalizedValue(`{"nested": "value"}`),
			"key2": jsontypes.NewNormalizedValue(`"simple string"`),
		},

		// Test nullable map with list values
		MapNullListString: map[string][]types.String{
			"list1": {
				types.StringValue("a"),
				types.StringValue("b"),
			},
		},
	}

	// Convert to SDK request
	request, diags := model.ToSharedFrameworkTypeRequest(ctx)
	if diags.HasError() {
		t.Fatalf("ToSharedFrameworkTypeRequest returned errors: %v", diags)
	}

	// Verify nullable maps are NOT nil (the bug would leave them as nil)
	if request.MapNullString == nil {
		t.Error("MapNullString should not be nil after conversion")
	} else {
		if len(request.MapNullString) != 2 {
			t.Errorf("MapNullString length = %d, want 2", len(request.MapNullString))
		}
		if request.MapNullString["key1"] != "value1" {
			t.Errorf("MapNullString[key1] = %q, want %q", request.MapNullString["key1"], "value1")
		}
		if request.MapNullString["key2"] != "value2" {
			t.Errorf("MapNullString[key2] = %q, want %q", request.MapNullString["key2"], "value2")
		}
	}

	if request.MapNullBool == nil {
		t.Error("MapNullBool should not be nil after conversion")
	} else {
		if len(request.MapNullBool) != 2 {
			t.Errorf("MapNullBool length = %d, want 2", len(request.MapNullBool))
		}
		if request.MapNullBool["enabled"] != true {
			t.Errorf("MapNullBool[enabled] = %v, want true", request.MapNullBool["enabled"])
		}
		if request.MapNullBool["disabled"] != false {
			t.Errorf("MapNullBool[disabled] = %v, want false", request.MapNullBool["disabled"])
		}
	}

	if request.MapNullInt64 == nil {
		t.Error("MapNullInt64 should not be nil after conversion")
	} else {
		if len(request.MapNullInt64) != 2 {
			t.Errorf("MapNullInt64 length = %d, want 2", len(request.MapNullInt64))
		}
		if request.MapNullInt64["count"] != 42 {
			t.Errorf("MapNullInt64[count] = %d, want 42", request.MapNullInt64["count"])
		}
		if request.MapNullInt64["limit"] != 100 {
			t.Errorf("MapNullInt64[limit] = %d, want 100", request.MapNullInt64["limit"])
		}
	}

	if request.MapNullAny == nil {
		t.Error("MapNullAny should not be nil after conversion")
	} else {
		if len(request.MapNullAny) != 2 {
			t.Errorf("MapNullAny length = %d, want 2", len(request.MapNullAny))
		}
		// Verify the JSON was properly unmarshaled
		if _, ok := request.MapNullAny["key1"]; !ok {
			t.Error("MapNullAny[key1] should exist")
		}
		if _, ok := request.MapNullAny["key2"]; !ok {
			t.Error("MapNullAny[key2] should exist")
		}
	}

	if request.MapNullListString == nil {
		t.Error("MapNullListString should not be nil after conversion")
	} else {
		if len(request.MapNullListString) != 1 {
			t.Errorf("MapNullListString length = %d, want 1", len(request.MapNullListString))
		}
		if list, ok := request.MapNullListString["list1"]; !ok {
			t.Error("MapNullListString[list1] should exist")
		} else if len(list) != 2 {
			t.Errorf("MapNullListString[list1] length = %d, want 2", len(list))
		}
	}
}

// TestFrameworkTypeResourceModel_ToSharedFrameworkTypeRequest_NullableMapsEmpty tests that
// when nullable maps are provided but empty, they are still properly initialized (not nil).
func TestFrameworkTypeResourceModel_ToSharedFrameworkTypeRequest_NullableMapsEmpty(t *testing.T) {
	ctx := context.Background()

	// Create a model with empty nullable maps (not nil, but empty)
	model := &FrameworkTypeResourceModel{
		Bool:          types.BoolValue(true),
		MapNullString: map[string]types.String{}, // Empty but not nil
		MapNullBool:   map[string]types.Bool{},
	}

	// Convert to SDK request
	request, diags := model.ToSharedFrameworkTypeRequest(ctx)
	if diags.HasError() {
		t.Fatalf("ToSharedFrameworkTypeRequest returned errors: %v", diags)
	}

	// Empty maps should still be initialized, not nil
	if request.MapNullString == nil {
		t.Error("MapNullString should not be nil even when empty")
	} else if len(request.MapNullString) != 0 {
		t.Errorf("MapNullString length = %d, want 0", len(request.MapNullString))
	}

	if request.MapNullBool == nil {
		t.Error("MapNullBool should not be nil even when empty")
	} else if len(request.MapNullBool) != 0 {
		t.Errorf("MapNullBool length = %d, want 0", len(request.MapNullBool))
	}
}

// TestFrameworkTypeResourceModel_ToSharedFrameworkTypeRequest_NullableMapsNil tests that
// when nullable maps are nil, they remain nil in the SDK request.
func TestFrameworkTypeResourceModel_ToSharedFrameworkTypeRequest_NullableMapsNil(t *testing.T) {
	ctx := context.Background()

	// Create a model with nil nullable maps
	model := &FrameworkTypeResourceModel{
		Bool:          types.BoolValue(true),
		MapNullString: nil, // Explicitly nil
		MapNullBool:   nil,
	}

	// Convert to SDK request
	request, diags := model.ToSharedFrameworkTypeRequest(ctx)
	if diags.HasError() {
		t.Fatalf("ToSharedFrameworkTypeRequest returned errors: %v", diags)
	}

	// Nil maps should remain nil (the if condition shouldn't be entered)
	if request.MapNullString != nil {
		t.Error("MapNullString should be nil when input is nil")
	}

	if request.MapNullBool != nil {
		t.Error("MapNullBool should be nil when input is nil")
	}
}

