package oas

import (
	"testing"
)

func TestDynamicReservedChecker_IsVersionSupported(t *testing.T) {
	checker := NewDynamicReservedChecker()

	supportedVersions := []string{"3.0.0", "3.0.1", "3.0.2", "3.0.3", "3.1.0", "3.1.1"}
	for _, version := range supportedVersions {
		if !checker.IsVersionSupported(version) {
			t.Errorf("Expected version %s to be supported", version)
		}
	}

	unsupportedVersions := []string{"2.0", "4.0.0", "3.2.0", ""}
	for _, version := range unsupportedVersions {
		if checker.IsVersionSupported(version) {
			t.Errorf("Expected version %s to not be supported", version)
		}
	}
}

func TestDynamicReservedChecker_UnsupportedVersionBehavior(t *testing.T) {
	checker := NewDynamicReservedChecker()

	// For unsupported versions, should allow renames (return false for reserved check)
	result := checker.IsOASReservedPropertyName("openapi", []string{}, "4.0.0")
	if result {
		t.Errorf("Expected unsupported version to allow renames, but got restricted")
	}

	result = checker.IsOASReservedPropertyName("customProperty", []string{"components", "schemas", "User"}, "4.0.0")
	if result {
		t.Errorf("Expected unsupported version to allow custom property renames, but got restricted")
	}
}

func TestDynamicReservedChecker_FallbackToStatic(t *testing.T) {
	checker := NewDynamicReservedChecker()

	// With empty version, should fallback to static checking
	result := checker.IsOASReservedPropertyName("openapi", []string{}, "")
	expected := IsOASReservedPropertyName("openapi", []string{})
	if result != expected {
		t.Errorf("Expected fallback to static checking, got %v, expected %v", result, expected)
	}
}

func TestDynamicReservedChecker_GetSupportedVersions(t *testing.T) {
	checker := NewDynamicReservedChecker()
	versions := checker.GetSupportedVersions()

	// Should return at least the versions we know about
	expectedVersions := []string{"3.0.0", "3.0.1", "3.0.2", "3.0.3", "3.1.0", "3.1.1"}
	for _, expected := range expectedVersions {
		found := false
		for _, actual := range versions {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected version %s to be in supported list", expected)
		}
	}
}

func TestIsOASReservedPropertyNameWithVersion(t *testing.T) {
	// Test the backward compatibility function

	// Should fallback to static checking when no version provided
	result1 := IsOASReservedPropertyNameWithVersion("openapi", []string{}, "")
	expected1 := IsOASReservedPropertyName("openapi", []string{})
	if result1 != expected1 {
		t.Errorf("Expected fallback to static checking, got %v, expected %v", result1, expected1)
	}

	// Should use dynamic checking when version provided (but may fallback on network errors)
	result2 := IsOASReservedPropertyNameWithVersion("openapi", []string{}, "3.1.0")
	// openapi should always be reserved
	if !result2 {
		t.Errorf("Expected 'openapi' to be reserved in any version")
	}

	// For unsupported versions, should allow renames
	result3 := IsOASReservedPropertyNameWithVersion("customProperty", []string{"components", "schemas"}, "4.0.0")
	if result3 {
		t.Errorf("Expected unsupported version to allow custom property renames")
	}
}

func TestSchemaManagerConfig(t *testing.T) {
	config := DefaultSchemaManagerConfig()

	if config.DisableRemoteSchemas {
		t.Errorf("Expected default config to enable remote schemas")
	}

	if config.CacheTTL.Hours() != 24 {
		t.Errorf("Expected default cache TTL to be 24 hours, got %v", config.CacheTTL)
	}

	if config.HTTPTimeout.Seconds() != 10 {
		t.Errorf("Expected default HTTP timeout to be 10 seconds, got %v", config.HTTPTimeout)
	}
}

func TestSchemaManager_IsSupported(t *testing.T) {
	manager := NewSchemaManager()

	if !manager.IsSupported("3.1.0") {
		t.Errorf("Expected 3.1.0 to be supported")
	}

	if manager.IsSupported("4.0.0") {
		t.Errorf("Expected 4.0.0 to not be supported")
	}

	if manager.IsSupported("") {
		t.Errorf("Expected empty version to not be supported")
	}
}
