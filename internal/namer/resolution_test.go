package namer

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	config "github.com/speakeasy-api/sdk-gen-config"
	"go.uber.org/zap/zapcore"
)

func TestResolveErrorConflict(t *testing.T) {
	tests := []struct {
		name           string
		originalError  string
		baseOrDefault  string
		expectedResult string
		description    string
	}{
		{
			name:           "insert_base_before_error_suffix",
			originalError:  "ValidationError",
			baseOrDefault:  "Base",
			expectedResult: "ValidationBaseError",
			description:    "Should insert 'Base' before 'Error' suffix",
		},
		{
			name:           "insert_default_before_exception_suffix",
			originalError:  "APIException",
			baseOrDefault:  "Default",
			expectedResult: "APIDefaultException",
			description:    "Should insert 'Default' before 'Exception' suffix",
		},
		{
			name:           "prepend_sdk_when_base_present_but_sdk_missing",
			originalError:  "BaseException",
			baseOrDefault:  "Base",
			expectedResult: "TestSDKBaseException",
			description:    "Should prepend SDK name when 'Base' is present but SDK name is missing",
		},
		{
			name:           "prepend_sdk_when_default_present_but_sdk_missing",
			originalError:  "CustomDefaultError",
			baseOrDefault:  "Default",
			expectedResult: "TestSDKCustomDefaultError",
			description:    "Should prepend SDK name when 'Default' is present but SDK name is missing",
		},
		{
			name:           "append_base_as_last_resort",
			originalError:  "TestSDKBaseError",
			baseOrDefault:  "Base",
			expectedResult: "TestSDKBaseErrorBase",
			description:    "Should append 'Base' as last resort when both SDK and 'Base' are present",
		},
		{
			name:           "append_default_as_last_resort",
			originalError:  "DefaultErrorForTestSDK",
			baseOrDefault:  "Default",
			expectedResult: "DefaultErrorForTestSDKDefault",
			description:    "Should append 'Default' as last resort when both SDK and 'Default' are present",
		},
	}

	// testing targets with "Error" and "Exception" preferred suffixes respectively
	for _, target := range []string{"go", "csharp"} {
		resolver := createTestResolver("TestSDK", target)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				result := resolver.resolveErrorConflict(tt.originalError, tt.baseOrDefault)

				if result != tt.expectedResult {
					t.Errorf("resolveErrorConflict() = %q, expected %q", result, tt.expectedResult)
				}
			})
		}
	}
}

func TestInsertBeforePreferredErrorSuffix(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		expectedResult string
		description    string
	}{
		{
			name:           "insert_base_before_error_suffix",
			target:         "go",
			expectedResult: "OriginalBaseErrorException",
			description:    "Should insert 'Base' before 'Error' suffix in Go.",
		},
		{
			name:           "insert_base_before_exception_suffix",
			target:         "java",
			expectedResult: "OriginalErrorBaseException",
			description:    "Should insert 'Base' before 'Exception' suffix in Java.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := createTestResolver("TestSDK", tt.target)
			result := resolver.resolveErrorConflict("OriginalErrorException", "Base")

			if result != tt.expectedResult {
				t.Errorf("resolveErrorConflict() = %q, expected %q", result, tt.expectedResult)
			}
		})
	}

}

func TestGetBucketNameForType(t *testing.T) {
	cfg := configuration.ImportConfig{
		Option: configuration.ImportOptionOpenAPI,
		Paths: map[string]string{
			"shared":     "models/shared",
			"errors":     "models/errors",
			"operations": "models/operations",
		},
	}

	singleSegmentCfg := configuration.ImportConfig{
		Option: configuration.ImportOptionOpenAPI,
		Paths: map[string]string{
			"shared":     "models",
			"errors":     "models/errors",
			"operations": "models/operations",
		},
	}

	ns := func(s string) *string { return &s }

	tests := []struct {
		name     string
		cfg      configuration.ImportConfig
		typeDef  *ast.TypeDef
		expected string
	}{
		{
			name: "shared_type_no_namespace",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
			},
			expected: "models/shared",
		},
		{
			name: "shared_type_with_namespace_extension",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
				Extensions: &ast.TypeDefExtensions{
					ModelNamespace: ns("identity"),
				},
			},
			expected: "models/identity",
		},
		{
			name: "error_type_with_namespace_extension",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeErrors,
				Extensions: &ast.TypeDefExtensions{
					ModelNamespace: ns("identity"),
				},
			},
			expected: "models/identity",
		},
		{
			name: "shared_and_error_land_in_same_bucket",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
				Extensions: &ast.TypeDefExtensions{
					ModelNamespace: ns("identity"),
				},
			},
			expected: "models/identity",
		},
		{
			name: "single_segment_shared_path_with_namespace",
			cfg:  singleSegmentCfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
				Extensions: &ast.TypeDefExtensions{
					ModelNamespace: ns("identity"),
				},
			},
			expected: "models/identity",
		},
		{
			name: "inherited_namespace_from_context_stack",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
				ContextStack: ast.ContextStack{
					{Type: ast.ContextTypeModelNamespace, Identifier: "billing"},
				},
			},
			expected: "models/billing",
		},
		{
			name: "extension_takes_precedence_over_context_stack",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
				Extensions: &ast.TypeDefExtensions{
					ModelNamespace: ns("identity"),
				},
				ContextStack: ast.ContextStack{
					{Type: ast.ContextTypeModelNamespace, Identifier: "billing"},
				},
			},
			expected: "models/identity",
		},
		{
			name: "empty_namespace_extension_ignored",
			cfg:  cfg,
			typeDef: &ast.TypeDef{
				Scope: ast.ScopeShared,
				Extensions: &ast.TypeDefExtensions{
					ModelNamespace: ns(""),
				},
			},
			expected: "models/shared",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBucketPathForType(tt.cfg, tt.typeDef)
			if result != tt.expected {
				t.Errorf("getBucketPathForType() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func createTestResolver(sdkClassName, target string) *Resolver {
	baseConfig := &config.Configuration{
		Generation: config.Generation{
			SDKClassName: sdkClassName,
		},
	}

	cfg := configuration.New(baseConfig, target)

	subsys := &subsystem.Subsystem{
		Config: cfg,
		Target: types.Target{Target: target},
	}

	logger := logging.NewLogger(zapcore.InfoLevel)

	resolver, _ := NewResolver(subsys, logger)
	return resolver
}
