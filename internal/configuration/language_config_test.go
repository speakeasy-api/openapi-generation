package configuration

import (
	"testing"

	config "github.com/speakeasy-api/sdk-gen-config"
)

func TestLanguageConfigHasMaintainTagBasedOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		langConfig config.LanguageConfig
		expected   bool
	}{
		{
			name: "returns true when maintainTagBasedOrdering is set to true",
			langConfig: config.LanguageConfig{
				Cfg: map[string]interface{}{
					"maintainTagBasedOrdering": true,
				},
			},
			expected: true,
		},
		{
			name: "returns false when maintainTagBasedOrdering is set to false",
			langConfig: config.LanguageConfig{
				Cfg: map[string]interface{}{
					"maintainTagBasedOrdering": false,
				},
			},
			expected: false,
		},
		{
			name: "returns false when maintainTagBasedOrdering is not present",
			langConfig: config.LanguageConfig{
				Cfg: map[string]interface{}{
					"otherConfig": "value",
				},
			},
			expected: false,
		},
		{
			name: "returns false when Cfg map is empty",
			langConfig: config.LanguageConfig{
				Cfg: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "returns false when maintainTagBasedOrdering is not a boolean",
			langConfig: config.LanguageConfig{
				Cfg: map[string]interface{}{
					"maintainTagBasedOrdering": "true",
				},
			},
			expected: false,
		},
		{
			name: "returns false when maintainTagBasedOrdering is nil",
			langConfig: config.LanguageConfig{
				Cfg: map[string]interface{}{
					"maintainTagBasedOrdering": nil,
				},
			},
			expected: false,
		},
		{
			name: "returns false when Cfg map is nil",
			langConfig: config.LanguageConfig{
				Cfg: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := LanguageConfigHasMaintainTagBasedOrdering(tt.langConfig)

			if result != tt.expected {
				t.Errorf("expected: %t, got %t", tt.expected, result)
			}
		})
	}
}
