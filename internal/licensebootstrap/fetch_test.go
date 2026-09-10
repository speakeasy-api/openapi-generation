package licensebootstrap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchLicenseToken(t *testing.T) {
	newServer := func(t *testing.T, status int, body string) *httptest.Server {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/auth/validate", r.URL.Path)
			assert.Equal(t, "secret-key", r.Header.Get("x-api-key"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(server.Close)
		return server
	}

	t.Run("entitled workspace returns the token and details", func(t *testing.T) {
		server := newServer(t, http.StatusOK, `{"workspace_id":"ws","workspace_slug":"acme","org_slug":"acme-org","account_type_v2":"enterprise","license_jwt":"h.p.s"}`)
		token, details, err := FetchLicenseToken(context.Background(), server.Client(), server.URL, "secret-key")
		require.NoError(t, err)
		assert.Equal(t, "h.p.s", token)
		assert.Equal(t, "acme", details.WorkspaceSlug)
		assert.Equal(t, "enterprise", details.AccountType)
	})

	t.Run("workspace without entitlement", func(t *testing.T) {
		server := newServer(t, http.StatusOK, `{"workspace_id":"ws","workspace_slug":"hobby","account_type_v2":"free","license_jwt":null}`)
		_, details, err := FetchLicenseToken(context.Background(), server.Client(), server.URL, "secret-key")
		require.ErrorIs(t, err, ErrNotEntitled)
		assert.Equal(t, "hobby", details.WorkspaceSlug)
	})

	t.Run("rejected API key", func(t *testing.T) {
		server := newServer(t, http.StatusUnauthorized, `{"message":"unauthorized"}`)
		_, _, err := FetchLicenseToken(context.Background(), server.Client(), server.URL, "secret-key")
		require.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("redirects are not followed and never receive the API key", func(t *testing.T) {
		var sinkCalled, keyLeaked atomic.Bool
		sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sinkCalled.Store(true)
			if r.Header.Get("x-api-key") != "" {
				keyLeaked.Store(true)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"license_jwt":"h.p.s"}`))
		}))
		t.Cleanup(sink.Close)
		redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, sink.URL+"/collect", http.StatusFound)
		}))
		t.Cleanup(redirecting.Close)

		_, _, err := FetchLicenseToken(context.Background(), redirecting.Client(), redirecting.URL, "secret-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redirect")
		assert.False(t, sinkCalled.Load(), "redirect target must never be contacted")
		assert.False(t, keyLeaked.Load(), "API key must never reach a redirect target")
	})

	t.Run("plain http to a non-loopback host is refused before any request", func(t *testing.T) {
		_, _, err := FetchLicenseToken(context.Background(), http.DefaultClient, "http://example.com", "secret-key")
		require.ErrorIs(t, err, ErrInsecureServerURL)
		_, _, err = FetchLicenseToken(context.Background(), http.DefaultClient, "ftp://example.com", "secret-key")
		require.ErrorIs(t, err, ErrInsecureServerURL)
	})

	t.Run("server error", func(t *testing.T) {
		server := newServer(t, http.StatusBadGateway, `oops`)
		_, _, err := FetchLicenseToken(context.Background(), server.Client(), server.URL, "secret-key")
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrUnauthorized)
		require.NotErrorIs(t, err, ErrNotEntitled)
	})
}

func TestFetchAccessLicenseToken(t *testing.T) {
	newServer := func(t *testing.T, status int, body string) *httptest.Server {
		t.Helper()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/workspace/access", r.URL.Path)
			assert.Equal(t, "typescript", r.URL.Query().Get("targetType"))
			assert.Equal(t, "secret-key", r.Header.Get("x-api-key"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(server.Close)
		return server
	}

	t.Run("allowed target returns the token and details", func(t *testing.T) {
		server := newServer(t, http.StatusOK, `{"generation_allowed":true,"level":"allowed","message":"","license_jwt":"h.p.s"}`)
		token, details, err := FetchAccessLicenseToken(context.Background(), server.Client(), server.URL, "secret-key", "typescript")
		require.NoError(t, err)
		assert.Equal(t, "h.p.s", token)
		assert.True(t, details.GenerationAllowed)
		assert.Equal(t, "allowed", details.Level)
	})

	t.Run("blocked target returns the response message", func(t *testing.T) {
		server := newServer(t, http.StatusOK, `{"generation_allowed":false,"level":"blocked","message":"typescript is not available","license_jwt":null}`)
		_, details, err := FetchAccessLicenseToken(context.Background(), server.Client(), server.URL, "secret-key", "typescript")
		require.ErrorIs(t, err, ErrNotEntitled)
		assert.Equal(t, "typescript is not available", details.Message)
	})

	t.Run("non-blocked response without a token", func(t *testing.T) {
		server := newServer(t, http.StatusOK, `{"generation_allowed":true,"level":"allowed","message":"","license_jwt":null}`)
		_, _, err := FetchAccessLicenseToken(context.Background(), server.Client(), server.URL, "secret-key", "typescript")
		require.ErrorIs(t, err, ErrAccessTokenMissing)
		assert.Equal(t, "the platform allowed generation but issued no token", err.Error())
	})

	t.Run("rejected API key", func(t *testing.T) {
		server := newServer(t, http.StatusUnauthorized, `{"message":"unauthorized"}`)
		_, _, err := FetchAccessLicenseToken(context.Background(), server.Client(), server.URL, "secret-key", "typescript")
		require.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("redirects are not followed and never receive the API key", func(t *testing.T) {
		var sinkCalled, keyLeaked atomic.Bool
		sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sinkCalled.Store(true)
			if r.Header.Get("x-api-key") != "" {
				keyLeaked.Store(true)
			}
			_, _ = w.Write([]byte(`{"license_jwt":"h.p.s"}`))
		}))
		t.Cleanup(sink.Close)
		redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, sink.URL+"/collect", http.StatusFound)
		}))
		t.Cleanup(redirecting.Close)

		_, _, err := FetchAccessLicenseToken(context.Background(), redirecting.Client(), redirecting.URL, "secret-key", "typescript")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "redirect")
		assert.False(t, sinkCalled.Load(), "redirect target must never be contacted")
		assert.False(t, keyLeaked.Load(), "API key must never reach a redirect target")
	})

	t.Run("plain http to a non-loopback host is refused before any request", func(t *testing.T) {
		_, _, err := FetchAccessLicenseToken(context.Background(), http.DefaultClient, "http://example.com", "secret-key", "typescript")
		require.ErrorIs(t, err, ErrInsecureServerURL)
		_, _, err = FetchAccessLicenseToken(context.Background(), http.DefaultClient, "ftp://example.com", "secret-key", "typescript")
		require.ErrorIs(t, err, ErrInsecureServerURL)
	})

	t.Run("server error", func(t *testing.T) {
		server := newServer(t, http.StatusBadGateway, `oops`)
		_, _, err := FetchAccessLicenseToken(context.Background(), server.Client(), server.URL, "secret-key", "typescript")
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrUnauthorized)
		require.NotErrorIs(t, err, ErrNotEntitled)
	})
}
