package schemas

import (
	"context"
	"os"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestHandleComplexSchemas tests the handling of complex schemas with oneOf discriminators
// and various property combinations like the OASOneOfRequest and OASOneOfResponse schemas
func TestHandleComplexSchemas(t *testing.T) {
	tests := []struct {
		name         string
		schemaName   string
		fixes        *config.Fixes                        // nil means use default
		mockFeatures func() *testutils.MockFeaturesConfig // nil means use default (all features supported)
		isRequest    *bool                                // nil means use default (false), otherwise use specified value
		inDoc        string
		expectedDoc  string
	}{
		{
			name:        "OASOneOfRequest",
			schemaName:  "TestSchema",
			isRequest:   pointer.From(true),
			inDoc:       "testdata/OASOneOfRequest.yaml",
			expectedDoc: "testdata/OASOneOfRequest_ast_expected.yaml",
		},
		{
			name:        "OASOneOfResponse",
			schemaName:  "TestSchema",
			isRequest:   pointer.From(false),
			inDoc:       "testdata/OASOneOfResponse.yaml",
			expectedDoc: "testdata/OASOneOfResponse_ast_expected.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment using common testutils
			in, err := os.ReadFile(tt.inDoc)
			require.NoError(t, err)

			var mockFeatures *testutils.MockFeaturesConfig
			if tt.mockFeatures != nil {
				mockFeatures = tt.mockFeatures()
			} else {
				mockFeatures = testutils.NewMockFeaturesConfig().WithSupportAllFeatures()
			}

			opts := testutils.TestEnvironmentOptions{
				OpenAPIYAML:  string(in),
				Fixes:        tt.fixes,
				MockFeatures: mockFeatures,
			}

			common, err := testutils.SetupTestEnvironment(opts)
			require.NoError(t, err)

			// Create the Schemas instance
			schemas := &Schemas{
				Config:    common.Config,
				Target:    common.Target,
				Subsystem: common.Subsystem,
				Namer:     common.Namer,
			}

			// Get the test schema
			testSchema, exists := common.DocInfo.Doc.GetComponents().GetSchemas().Get(tt.schemaName)
			require.True(t, exists)
			require.NotNil(t, testSchema)

			// Create test parameters
			isRequest := false
			if tt.isRequest != nil {
				isRequest = *tt.isRequest
			}

			params := Params{
				ContextStack:        ast.ContextStack{},
				SerializationMethod: ast.SerializationMethodJSON,
				Schema:              testSchema,
				Scope:               ast.ScopeShared,
				IsRequest:           isRequest,
				Depth:               0,
				MaxDepth:            10,
				Parents:             []string{},
				TypeDefCache:        make(map[string]*ast.TypeDef),
				LoopContext:         []LoopFrame{},
				Nullable:            false,
				CircularReference:   false,
				DocInfo:             common.DocInfo,
			}

			// Call the top-level HandleSchema method instead of handleAnyOfOneOf directly
			ctx := context.Background()
			result, err := schemas.HandleSchema(ctx, params)

			// Verify the result
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify the result
			require.NoError(t, err)
			require.NotNil(t, result)

			out, err := yaml.Marshal(result)
			require.NoError(t, err)

			expected, err := os.ReadFile(tt.expectedDoc)
			require.NoError(t, err)

			assert.Equal(t, string(expected), string(out))
		})
	}
}
