package configuration

import (
	"testing"

	config "github.com/speakeasy-api/sdk-gen-config"
)

func TestConfigDocumentation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		value        any
		wantMode     string
		wantMintlify bool
		wantDisabled bool
	}{
		{
			name:         "defaults to standard when unset",
			value:        nil,
			wantMode:     DocumentationStandard,
			wantMintlify: false,
			wantDisabled: false,
		},
		{
			name:         "defaults to standard for empty string",
			value:        "",
			wantMode:     DocumentationStandard,
			wantMintlify: false,
			wantDisabled: false,
		},
		{
			name:         "defaults to standard for non-string value",
			value:        true,
			wantMode:     DocumentationStandard,
			wantMintlify: false,
			wantDisabled: false,
		},
		{
			name:         "mintlify",
			value:        DocumentationMintlify,
			wantMode:     DocumentationMintlify,
			wantMintlify: true,
			wantDisabled: false,
		},
		{
			name:         "none disables docs",
			value:        DocumentationNone,
			wantMode:     DocumentationNone,
			wantMintlify: false,
			wantDisabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			additional := map[string]any{}
			if tt.value != nil {
				additional["documentation"] = tt.value
			}

			c := &Config{
				Configuration: config.Configuration{
					Generation: config.Generation{
						AdditionalProperties: additional,
					},
				},
			}

			if got := c.Documentation(); got != tt.wantMode {
				t.Errorf("Documentation() = %q, want %q", got, tt.wantMode)
			}
			if got := c.IsMintlify(); got != tt.wantMintlify {
				t.Errorf("IsMintlify() = %v, want %v", got, tt.wantMintlify)
			}
			if got := c.DocsDisabled(); got != tt.wantDisabled {
				t.Errorf("DocsDisabled() = %v, want %v", got, tt.wantDisabled)
			}
		})
	}
}
