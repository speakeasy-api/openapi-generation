package generate

import (
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/testutils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/require"
)

func TestHandleParametersAllowEmptyValue(t *testing.T) {
	tests := []struct {
		name        string
		openAPIYAML string
		validate    func(t *testing.T, result *ast.RequestParams, err error)
	}{
		{
			name: "QueryParameterWithAllowEmptyValueTrue",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: test
      parameters:
        - name: filter
          in: query
          schema:
            type: string
          x-speakeasy-allow-empty-value: true
      responses:
        '200':
          description: Success`,
			validate: func(t *testing.T, result *ast.RequestParams, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.QueryParams, 1)
				require.True(t, result.QueryParams[0].AllowEmptyValue)
			},
		},
		{
			name: "QueryParameterWithAllowEmptyValueFalse",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: test
      parameters:
        - name: filter
          in: query
          schema:
            type: string
          x-speakeasy-allow-empty-value: false
      responses:
        '200':
          description: Success`,
			validate: func(t *testing.T, result *ast.RequestParams, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.QueryParams, 1)
				require.False(t, result.QueryParams[0].AllowEmptyValue)
			},
		},
		{
			name: "QueryParameterWithoutExtension",
			openAPIYAML: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: test
      parameters:
        - name: filter
          in: query
          schema:
            type: string
      responses:
        '200':
          description: Success`,
			validate: func(t *testing.T, result *ast.RequestParams, err error) {
				t.Helper()
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Len(t, result.QueryParams, 1)
				require.False(t, result.QueryParams[0].AllowEmptyValue)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
				OpenAPIYAML: tt.openAPIYAML,
			})
			require.NoError(t, err)

			schemasInstance := &schemas.Schemas{
				Config:    setup.Config,
				Target:    setup.Target,
				Subsystem: setup.Subsystem,
				Namer:     setup.Namer,
			}

			generator := &Generator{
				subsystem: setup.Subsystem,
				schemas:   schemasInstance,
			}

			pi, exists := setup.DocInfo.Doc.GetPaths().Get("/test")
			require.True(t, exists)
			require.NotNil(t, pi)

			pathItem, err := resolution.Resolve(t.Context(), pi, setup.DocInfo)
			require.NoError(t, err)
			require.NotNil(t, pathItem)

			operation := pathItem.Get()
			require.NotNil(t, operation)

			var itemParams []*openapi.ReferencedParameter
			if pathItem.Parameters != nil {
				itemParams = append(itemParams, pathItem.GetParameters()...)
			}

			var paramsOAS []*openapi.ReferencedParameter
			if operation.Parameters != nil {
				paramsOAS = append(paramsOAS, operation.GetParameters()...)
			}

			opts := handleParametersOptions{
				contextStack: ast.ContextStack{},
				itemParams:   itemParams,
				paramsOAS:    paramsOAS,
				scope:        ast.ScopeShared,
				opID:         operation.GetOperationID(),
				docInfo:      setup.DocInfo,
			}

			result, _, _, err := generator.handleParameters(context.Background(), opts)
			tt.validate(t, result, err)
		})
	}
}

// Targets that do not implement allowReserved previously failed generation with
// an unsupported error, dropping the whole operation. They now generate the
// parameter with standard percent encoding and warn instead.
func TestHandleParametersWarnsWhenAllowReservedIsUnsupported(t *testing.T) {
	const pathEncodingOverrideYAML = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test/{id}:
    get:
      operationId: test
      parameters:
        - name: id
          in: path
          required: true
          x-speakeasy-param-encoding-override: allowReserved
          schema:
            type: string
      responses:
        '200':
          description: Success`

	const queryAllowReservedYAML = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test/{id}:
    get:
      operationId: test
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: filter
          in: query
          allowReserved: true
          schema:
            type: string
      responses:
        '200':
          description: Success`

	const queryEncodingOverrideYAML = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test/{id}:
    get:
      operationId: test
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: filter
          in: query
          x-speakeasy-param-encoding-override: allowReserved
          schema:
            type: string
      responses:
        '200':
          description: Success`

	const pathAllowReservedKeywordYAML = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test/{id}:
    get:
      operationId: test
      parameters:
        - name: id
          in: path
          required: true
          allowReserved: true
          schema:
            type: string
      responses:
        '200':
          description: Success`

	tests := []struct {
		name            string
		path            string
		openAPIYAML     string
		allowReserved   bool
		expectedWarning string
		validate        func(t *testing.T, result *ast.RequestParams)
	}{
		{
			name:            "PathParameterEncodingOverrideUnsupported",
			path:            "/test/{id}",
			openAPIYAML:     pathEncodingOverrideYAML,
			expectedWarning: "path parameter \"id\" has `x-speakeasy-param-encoding-override: allowReserved`, which this language does not support; the parameter will use standard percent encoding",
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.PathParams, 1)
				ann := result.PathParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				// The annotation must not reach templates on a target that
				// cannot honour it.
				require.False(t, paramAnn.AllowReserved)
			},
		},
		{
			name:            "QueryParameterAllowReservedUnsupported",
			path:            "/test/{id}",
			openAPIYAML:     queryAllowReservedYAML,
			expectedWarning: "query parameter \"filter\" has `allowReserved: true`, which this language does not support; the parameter value will use standard percent encoding",
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.QueryParams, 1)
				ann := result.QueryParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				// The annotation must not reach templates on a target that
				// cannot honour it.
				require.False(t, paramAnn.AllowReserved)
			},
		},
		{
			name:          "QueryParameterAllowReservedSupported",
			path:          "/test/{id}",
			openAPIYAML:   queryAllowReservedYAML,
			allowReserved: true,
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.QueryParams, 1)
				ann := result.QueryParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				require.True(t, paramAnn.AllowReserved)
			},
		},
		{
			// The encoding-override extension is documented for path
			// parameters only; query parameters opt in via the spec's own
			// allowReserved keyword. The extension on a query parameter is
			// never honoured — the generator warns and points at the
			// keyword instead, regardless of target support.
			name:            "QueryParameterEncodingOverrideMisplacedWhenSupported",
			path:            "/test/{id}",
			openAPIYAML:     queryEncodingOverrideYAML,
			allowReserved:   true,
			expectedWarning: "query parameter \"filter\" has `x-speakeasy-param-encoding-override: allowReserved`, which only applies to path parameters; use `allowReserved: true` on the parameter instead",
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.QueryParams, 1)
				ann := result.QueryParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				require.False(t, paramAnn.AllowReserved)
			},
		},
		{
			name:            "QueryParameterEncodingOverrideMisplacedWhenUnsupported",
			path:            "/test/{id}",
			openAPIYAML:     queryEncodingOverrideYAML,
			expectedWarning: "query parameter \"filter\" has `x-speakeasy-param-encoding-override: allowReserved`, which only applies to path parameters; use `allowReserved: true` on the parameter instead",
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.QueryParams, 1)
				ann := result.QueryParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				require.False(t, paramAnn.AllowReserved)
			},
		},
		{
			// The mirror of the misplaced-extension cases: the spec's
			// allowReserved keyword only applies to query parameters, so a
			// path parameter carrying it is ignored with a warning pointing
			// at the extension, regardless of target support.
			name:            "PathParameterKeywordMisplacedWhenSupported",
			path:            "/test/{id}",
			openAPIYAML:     pathAllowReservedKeywordYAML,
			allowReserved:   true,
			expectedWarning: "path parameter \"id\" has `allowReserved: true`, which only applies to query parameters; use `x-speakeasy-param-encoding-override: allowReserved` on the parameter instead",
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.PathParams, 1)
				ann := result.PathParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				require.False(t, paramAnn.AllowReserved)
			},
		},
		{
			name:            "PathParameterKeywordMisplacedWhenUnsupported",
			path:            "/test/{id}",
			openAPIYAML:     pathAllowReservedKeywordYAML,
			expectedWarning: "path parameter \"id\" has `allowReserved: true`, which only applies to query parameters; use `x-speakeasy-param-encoding-override: allowReserved` on the parameter instead",
			validate: func(t *testing.T, result *ast.RequestParams) {
				t.Helper()
				require.Len(t, result.PathParams, 1)
				ann := result.PathParams[0].Field.Annotations.Get("param")
				require.NotNil(t, ann)
				paramAnn, ok := ann.(*ast.ParamAnnotation)
				require.True(t, ok)
				require.False(t, paramAnn.AllowReserved)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A non-nil SupportedFeatures map reports every absent feature as
			// unsupported, so support everything explicitly and vary only
			// allowReserved. Otherwise the "unsupported" cases would really be
			// testing a target that supports nothing at all.
			supported := map[features.Feature]bool{}
			for _, feature := range features.GetAllFeatures() {
				supported[feature] = true
			}
			supported[features.FeatureAllowReserved] = tt.allowReserved

			mockFeatures := testutils.NewMockFeaturesConfig()
			mockFeatures.SupportedFeatures = supported

			setup, err := testutils.SetupTestEnvironment(testutils.TestEnvironmentOptions{
				OpenAPIYAML:  tt.openAPIYAML,
				MockFeatures: mockFeatures,
			})
			require.NoError(t, err)

			generator := &Generator{
				subsystem: setup.Subsystem,
				schemas: &schemas.Schemas{
					Config:    setup.Config,
					Target:    setup.Target,
					Subsystem: setup.Subsystem,
					Namer:     setup.Namer,
				},
			}

			pi, exists := setup.DocInfo.Doc.GetPaths().Get(tt.path)
			require.True(t, exists)

			pathItem, err := resolution.Resolve(t.Context(), pi, setup.DocInfo)
			require.NoError(t, err)
			require.NotNil(t, pathItem)

			operation := pathItem.Get()
			require.NotNil(t, operation)

			opts := handleParametersOptions{
				contextStack: ast.ContextStack{},
				itemParams:   pathItem.GetParameters(),
				paramsOAS:    operation.GetParameters(),
				scope:        ast.ScopeShared,
				opID:         operation.GetOperationID(),
				docInfo:      setup.DocInfo,
			}

			warningLogger := logging.NewWarningLogger(false, nil)
			ctx := logging.WithWarningLogger(t.Context(), warningLogger)

			result, _, _, err := generator.handleParameters(ctx, opts)
			require.NoError(t, err)
			require.NotNil(t, result)
			tt.validate(t, result)

			warnings := warningLogger.GetWarnings()
			if tt.expectedWarning == "" {
				require.Empty(t, warnings)
				return
			}
			// Match on the message rather than the warning count, so an
			// unrelated warning from another feature cannot fail this test.
			var matched []error
			for _, warning := range warnings {
				if strings.Contains(warning.Error(), tt.expectedWarning) {
					matched = append(matched, warning)
				}
			}
			require.Len(t, matched, 1, "warnings: %v", warnings)
		})
	}
}
