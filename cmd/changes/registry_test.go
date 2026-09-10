package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type registryRoundTripFunc func(*http.Request) (*http.Response, error)

func (f registryRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSetupClientValidatesAPIKeyAndResolvesValidatedWorkspace(t *testing.T) {
	for _, tt := range []struct {
		name               string
		serverURL          string
		wantValidationHost string
		wantSDKHost        string
	}{
		{
			name:               "defaults diverge",
			wantValidationHost: "app.speakeasy.com",
			wantSDKHost:        "api.prod.speakeasy.com",
		},
		{
			name:               "allowlisted override is shared",
			serverURL:          "https://staging.speakeasy.com",
			wantValidationHost: "staging.speakeasy.com",
			wantSDKHost:        "staging.speakeasy.com",
		},
		{
			name:               "custom override only applies to registry SDK",
			serverURL:          "https://custom.example.invalid",
			wantValidationHost: "app.speakeasy.com",
			wantSDKHost:        "custom.example.invalid",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			resetRegistryTestState(t)
			t.Setenv("SPEAKEASY_SERVER_URL", tt.serverURL)
			apiKey = "test-api-key"

			var paths []string
			httpClient := &http.Client{Transport: registryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				paths = append(paths, req.URL.Path)
				switch req.URL.Path {
				case "/v1/auth/validate":
					assert.Equal(t, tt.wantValidationHost, req.URL.Host)
					assert.Equal(t, "test-api-key", req.Header.Get("x-api-key"))
					return registryJSONResponse(req, http.StatusOK, `{"workspace_id":"validated-workspace-id"}`), nil
				case "/v1/workspace/validated-workspace-id":
					assert.Equal(t, tt.wantSDKHost, req.URL.Host)
					return registryJSONResponse(req, http.StatusOK, `{"id":"validated-workspace-id","name":"Workspace Name","organization_id":"organization-id","slug":"workspace-slug"}`), nil
				case "/v1/organization/organization-id":
					assert.Equal(t, tt.wantSDKHost, req.URL.Host)
					return registryJSONResponse(req, http.StatusOK, `{"id":"organization-id","name":"Organization Name","slug":"organization-slug"}`), nil
				default:
					t.Fatalf("unexpected request: %s", req.URL)
					return nil, nil
				}
			})}

			ctx, client, err := setupClientWithHTTPClient(httpClient)
			require.NoError(t, err)
			state, err := registryStateFromContext(ctx)
			require.NoError(t, err)
			assert.Equal(t, "validated-workspace-id", state.workspaceID)
			accessState, ok := generationaccess.StateFromContext(ctx)
			require.True(t, ok)
			assert.Equal(t, generationaccess.ModeDirect, accessState.Mode())

			require.NoError(t, resolveWorkspace(ctx, client, logging.NewLoggerFromZap(zap.NewNop())))
			assert.Equal(t, "workspace-slug", workspace)
			assert.Equal(t, "organization-slug", org)
			assert.Equal(t, []string{
				"/v1/auth/validate",
				"/v1/workspace/validated-workspace-id",
				"/v1/organization/organization-id",
			}, paths)
		})
	}
}

func TestSetupClientRejectsMissingWorkspaceIdentity(t *testing.T) {
	resetRegistryTestState(t)
	apiKey = "test-api-key"

	httpClient := &http.Client{Transport: registryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return registryJSONResponse(req, http.StatusOK, `{"workspace_id":"  "}`), nil
	})}

	ctx, client, err := setupClientWithHTTPClient(httpClient)
	require.ErrorContains(t, err, "API key validation did not return a workspace identity")
	assert.Nil(t, ctx)
	assert.Nil(t, client)
}

func TestSetupClientRejectsInvalidAPIKey(t *testing.T) {
	resetRegistryTestState(t)
	apiKey = "invalid-api-key"

	requests := 0
	httpClient := &http.Client{Transport: registryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		assert.Equal(t, "/v1/auth/validate", req.URL.Path)
		return registryJSONResponse(req, http.StatusUnauthorized, `{"message":"invalid API key"}`), nil
	})}

	ctx, client, err := setupClientWithHTTPClient(httpClient)
	require.ErrorContains(t, err, "failed to validate API key")
	assert.Nil(t, ctx)
	assert.Nil(t, client)
	assert.Equal(t, 1, requests)
}

func TestCallPreflightUsesRegistryStateAndInjectedClient(t *testing.T) {
	requests := 0
	httpClient := &http.Client{Transport: registryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "app.speakeasy.com", req.URL.Host)
		assert.Equal(t, "/v1/artifacts/preflight", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", req.Header.Get("x-api-key"))
		assert.Equal(t, "workspace-id", req.Header.Get("x-workspace-id"))

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		var payload map[string]string
		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, map[string]string{"namespace_name": "pets"}, payload)
		return registryJSONResponse(req, http.StatusOK, `{"auth_token":"local-token","namespace_id":"namespace-id"}`), nil
	})}
	ctx := context.WithValue(context.Background(), registryContextKey{}, registryContext{
		apiKey:       "test-api-key",
		workspaceID:  "workspace-id",
		httpClient:   httpClient,
		preflightURL: preflightServerURL(""),
	})

	response, err := callPreflight(ctx, "pets")
	require.NoError(t, err)
	assert.Equal(t, "local-token", response.AuthToken)
	assert.Equal(t, "namespace-id", response.NamespaceID)
	assert.Equal(t, 1, requests)
}

func TestAuthenticatedArtifactClientUsesInjectedClientAndSDKDefaultServer(t *testing.T) {
	requests := 0
	httpClient := &http.Client{Transport: registryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		switch req.URL.Path {
		case "/v1/artifacts/preflight":
			assert.Equal(t, "staging.speakeasy.com", req.URL.Host)
			return registryJSONResponse(req, http.StatusOK, `{"auth_token":"local-token","namespace_id":"namespace-id"}`), nil
		case "/v1/oci/v2/organization/workspace/pets/manifests/revision":
			assert.Equal(t, "api.prod.speakeasy.com", req.URL.Host)
			assert.Equal(t, "Bearer local-token", req.Header.Get("Authorization"))
			response := registryJSONResponse(req, http.StatusOK, `{"schemaVersion":2,"layers":[{"digest":"sha256:abc123"}]}`)
			response.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
			return response, nil
		default:
			t.Fatalf("unexpected request: %s", req.URL)
			return nil, nil
		}
	})}
	ctx := context.WithValue(context.Background(), registryContextKey{}, registryContext{
		apiKey:       "test-api-key",
		workspaceID:  "workspace-id",
		httpClient:   httpClient,
		preflightURL: "https://staging.speakeasy.com",
	})
	logger := logging.NewLoggerFromZap(zap.NewNop())

	client, err := getAuthenticatedClient(ctx, "pets", logger)
	require.NoError(t, err)
	digest, err := getLayerDigest(context.Background(), client, "organization", "workspace", "pets", "revision", logger)
	require.NoError(t, err)
	assert.Equal(t, "sha256:abc123", digest)
	assert.Equal(t, 2, requests)
}

func TestPreflightServerURLPreservesKnownOverrideBehavior(t *testing.T) {
	assert.Equal(t, "https://app.speakeasy.com", preflightServerURL(""))
	assert.Equal(t, "https://staging.speakeasy.com", preflightServerURL("https://staging.speakeasy.com"))
	assert.Equal(t, "https://app.speakeasy.com", preflightServerURL("https://custom.example.invalid"))
}

func resetRegistryTestState(t *testing.T) {
	t.Helper()

	originalAPIKey := apiKey
	originalOrg := org
	originalWorkspace := workspace
	apiKey = ""
	org = ""
	workspace = ""
	t.Setenv("SPEAKEASY_API_KEY", "")
	t.Setenv("SPEAKEASY_SERVER_URL", "")
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() {
		apiKey = originalAPIKey
		org = originalOrg
		workspace = originalWorkspace
	})
}

func registryJSONResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(strings.TrimSpace(body))),
		Request:    req,
	}
}
