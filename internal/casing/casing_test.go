package casing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindAcronyms_Success(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want map[string]bool
	}{
		{
			name: "find URL at end of string",
			args: args{
				input: "nativeOrgURL",
			},
			want: map[string]bool{
				"URL": true,
			},
		},
		{
			name: "find URL in middle of string",
			args: args{
				input: "nativeURLOrg",
			},
			want: map[string]bool{
				"URL": true,
			},
		},
		{
			name: "find URL at start of string",
			args: args{
				input: "URLNativeOrg",
			},
			want: map[string]bool{
				"URL": true,
			},
		},
		{
			name: "find URL at start of string",
			args: args{
				input: "URLNativeOrg",
			},
			want: map[string]bool{
				"URL": true,
			},
		},
		{
			name: "find multiple acronyms",
			args: args{
				input: "URLHeavyAMTOfAcronymsYO",
			},
			want: map[string]bool{
				"URL": true,
				"AMT": true,
				"YO":  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acronyms := findAcronyms(tt.args.input)
			assert.Equal(t, tt.want, acronyms)
		})
	}
}

func TestCasingPreservesAcronymsFromInput(t *testing.T) {
	caser := New()

	assert.Equal(t, "NativeOrgURL", caser.ToPascal("nativeOrgURL"))
	assert.Equal(t, "AdminAPIKey", caser.ToPascal("AdminAPIKey"))
	assert.Equal(t, "AdminApiKey", caser.ToPascal("AdminApiKey"))
}

func TestCasingAppliesSymbolCasingOverrides(t *testing.T) {
	caser := NewWithSymbolCasingOverrides([]string{
		"API",
		"DataKit",
		"ID",
		"JSON",
		"MCP",
		"OAuth",
		"ExampleAI",
		"URL",
	})

	assert.Equal(t, "AdminAPIKeyCreateResponse", caser.ToPascal("AdminApiKeyCreateResponse"))
	assert.Equal(t, "ResponseFormatJSONObject", caser.ToPascal("ResponseFormatJsonObject"))
	assert.Equal(t, "ProtocolMCP", caser.ToPascal("ProtocolMcp"))
	assert.Equal(t, "ImageURL", caser.ToPascal("ImageUrl"))
	assert.Equal(t, "OAuth2ClientID", caser.ToPascal("Oauth2ClientId"))
	assert.Equal(t, "DataKitSession", caser.ToPascal("DatakitSession"))
	assert.Equal(t, "ExampleAITraceID", caser.ToGoPascal("ExampleaiTraceId"))
	assert.Equal(t, "ApiaryIdentifier", caser.ToPascal("ApiaryIdentifier"))
}

func TestCustomCasingOverrides(t *testing.T) {
	overrides := CustomCasingOverrides(map[string]any{
		"api": map[string]any{
			"initialism": true,
		},
		"datakit": map[string]any{
			"pascal":  "DataKit",
			"camel":   "datakit",
			"snake":   "datakit",
			"capital": "DataKit",
		},
		"exampleai": map[string]any{
			"pascal": "ExampleAI",
		},
		"webp": map[string]string{
			"capital": "WebP",
		},
		"unused": map[string]any{
			"camel": "unused",
		},
	})

	assert.Equal(t, []string{"API", "DataKit", "ExampleAI", "WebP"}, overrides)
}

func TestApplySymbolCasing(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		overrides []string
		want      string
	}{
		{
			name:      "preserves configured acronyms in pascal symbols",
			input:     "AdminApiKeyCreateResponse",
			overrides: []string{"API"},
			want:      "AdminAPIKeyCreateResponse",
		},
		{
			name:      "preserves adjacent configured acronyms",
			input:     "ResponseFormatJsonObject",
			overrides: []string{"JSON"},
			want:      "ResponseFormatJSONObject",
		},
		{
			name:      "preserves mixed case initialisms",
			input:     "ChatSessionDatakitConfiguration",
			overrides: []string{"DataKit"},
			want:      "ChatSessionDataKitConfiguration",
		},
		{
			name:      "preserves mixed case compound name",
			input:     "ExampleaiFileObject",
			overrides: []string{"ExampleAI"},
			want:      "ExampleAIFileObject",
		},
		{
			name:      "does not replace inside larger words",
			input:     "ApiaryIdentifier",
			overrides: []string{"API", "ID"},
			want:      "ApiaryIdentifier",
		},
		{
			name:      "does not replace id inside identity words",
			input:     "IdentityValidationID",
			overrides: []string{"ID"},
			want:      "IdentityValidationID",
		},
		{
			name:      "does not replace url inside larger words",
			input:     "UrlencodedURLValidator",
			overrides: []string{"URL"},
			want:      "UrlencodedURLValidator",
		},
		{
			name:      "allows digit boundary",
			input:     "Oauth2ClientId",
			overrides: []string{"OAuth", "ID"},
			want:      "OAuth2ClientID",
		},
		{
			name:      "uses longer overrides first",
			input:     "ExampleaiAssistantAiModel",
			overrides: []string{"AI", "ExampleAI"},
			want:      "ExampleAIAssistantAIModel",
		},
		{
			name:      "empty overrides leave input unchanged",
			input:     "ProtocolMcp",
			overrides: nil,
			want:      "ProtocolMcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ApplySymbolCasing(tt.input, tt.overrides))
		})
	}
}
