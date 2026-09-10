package generate

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeUrl(t *testing.T) {
	type testCase struct {
		name       string
		url        string
		baseUrl    string
		expected   string
		isAbsolute bool
	}

	testCases := []testCase{
		{
			name:       "absolute http url",
			url:        "http://example.com",
			baseUrl:    "",
			expected:   "http://example.com",
			isAbsolute: true,
		},
		{
			name:       "relative path without base url",
			url:        "/api/v1",
			baseUrl:    "",
			expected:   "/api/v1",
			isAbsolute: false,
		},
		{
			name:       "relative path with base url",
			url:        "/api/v1",
			baseUrl:    "http://example.com",
			expected:   "http://example.com/api/v1",
			isAbsolute: true,
		},
		{
			name:       "domain without scheme",
			url:        "example.com/api/v1",
			baseUrl:    "",
			expected:   "https://example.com/api/v1",
			isAbsolute: true,
		},
		{
			name:       "domain without scheme with http base url",
			url:        "example.com/api/v1",
			baseUrl:    "http://example.com",
			expected:   "https://example.com/api/v1", // prefer the url in this case
			isAbsolute: true,
		},
		{
			name:       "domain without scheme with https base url",
			url:        "example.com/api/v1",
			baseUrl:    "https://example.com",
			expected:   "https://example.com/api/v1",
			isAbsolute: true,
		},
		{
			name:       "url with variables",
			url:        "{protocol}://{hostname}:{port}",
			expected:   "{protocol}://{hostname}:{port}",
			isAbsolute: true,
		},
		{
			name:       "empty url",
			url:        "",
			expected:   "",
			isAbsolute: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, isAbsolute, err := normalizeUrl(tc.url, tc.baseUrl)
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.isAbsolute, isAbsolute)
			require.NoError(t, err)
		})
	}
}

func TestNormalizeUrlError(t *testing.T) {
	_, _, err := normalizeUrl("ws://example.com", "")
	require.Error(t, err)

	_, _, err = normalizeUrl("wss://example.com", "")
	require.Error(t, err)
}

func TestHandleGlobalServers(t *testing.T) {
	t.Parallel()

	// Mock Generator for testing
	mockGenerator := func(baseURL string) *Generator {
		return &Generator{
			subsystem: &subsystem.Subsystem{
				Config: &configuration.Config{
					Configuration: config.Configuration{
						Generation: config.Generation{
							BaseServerURL: baseURL,
						},
					},
				},
				Features: features.NewMock(&features.MockConfig{}),
			},
		}
	}

	t.Run("with provided servers and empty base URL", func(t *testing.T) {
		t.Parallel()

		g := mockGenerator("")
		servers := []*openapi.Server{
			{
				URL:         "https://api.example.com",
				Description: pointer.From("Production API"),
			},
		}
		a := ast.NewAST()
		a.MainSDK = ast.NewMainSDK(nil)

		result, err := g.handleGlobalServers(context.Background(), servers, a)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Servers, 1)
		assert.Equal(t, "https://api.example.com", result.Servers[0].URL)
		assert.False(t, result.Servers[0].IsRelative)
		assert.Equal(t, "Production API", result.Servers[0].Comments.Description)
		assert.NotNil(t, a.MainSDK.Servers)
	})

	t.Run("with provided servers and non-empty base URL", func(t *testing.T) {
		t.Parallel()

		g := mockGenerator("https://IGNORE-THIS-BASE-URL.com")
		servers := []*openapi.Server{
			{
				URL:         "https://USE-THIS-SERVER.com",
				Description: pointer.From("Production API"),
			},
		}
		a := ast.NewAST()
		a.MainSDK = ast.NewMainSDK(nil)

		result, err := g.handleGlobalServers(context.Background(), servers, a)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Servers, 1)
		assert.Equal(t, "https://USE-THIS-SERVER.com", result.Servers[0].URL)
		assert.False(t, result.Servers[0].IsRelative)
		assert.NotNil(t, a.MainSDK.Servers)
	})

	t.Run("with empty servers and empty base URL", func(t *testing.T) {
		t.Parallel()

		g := mockGenerator("")
		a := ast.NewAST()
		a.MainSDK = ast.NewMainSDK(nil)

		result, err := g.handleGlobalServers(context.Background(), nil, a)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Servers, 1)
		assert.Equal(t, "/", result.Servers[0].URL)
		assert.True(t, result.Servers[0].IsRelative)
		assert.NotNil(t, a.MainSDK.Servers)
	})

	t.Run("with empty servers and non-empty base URL", func(t *testing.T) {
		t.Parallel()

		g := mockGenerator("https://custom.example.com")
		a := ast.NewAST()
		a.MainSDK = ast.NewMainSDK(nil)

		result, err := g.handleGlobalServers(context.Background(), nil, a)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Servers, 1)
		assert.Equal(t, "https://custom.example.com", result.Servers[0].URL)
		assert.False(t, result.Servers[0].IsRelative)
		assert.NotNil(t, a.MainSDK.Servers)
	})

	t.Run("with empty servers and relative base URL", func(t *testing.T) {
		t.Parallel()

		g := mockGenerator("/api/v2")
		a := ast.NewAST()
		a.MainSDK = ast.NewMainSDK(nil)

		result, err := g.handleGlobalServers(context.Background(), nil, a)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Servers, 1)
		assert.Equal(t, "/api/v2", result.Servers[0].URL)
		assert.True(t, result.Servers[0].IsRelative)
		assert.NotNil(t, a.MainSDK.Servers)
	})
}
