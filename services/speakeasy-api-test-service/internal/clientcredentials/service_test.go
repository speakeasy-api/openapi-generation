package clientcredentials

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandleTokenRequest(t *testing.T) {
	tests := []struct {
		name            string
		setupRequest    func() *http.Request
		wantStatus      int
		wantAccessToken string
	}{
		{
			name: "valid form credentials",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				form.Set("client_id", "speakeasy-sdks")
				form.Set("client_secret", "supersecret-123")
				form.Set("scope", "read write")

				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			wantStatus:      http.StatusOK,
			wantAccessToken: firstAccessToken,
		},
		{
			name: "valid basic auth",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				form.Set("scope", "read write")

				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				// Create basic auth header
				creds := base64.StdEncoding.EncodeToString([]byte("speakeasy-sdks:supersecret-123"))
				req.Header.Set("Authorization", fmt.Sprintf("Basic %s", creds))

				return req
			},
			wantStatus:      http.StatusOK,
			wantAccessToken: firstAccessToken,
		},
		{
			name: "invalid basic auth format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", nil)
				req.Header.Set("Authorization", "Basic invalid-base64")
				return req
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing credentials in basic auth",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", nil)
				// Encode just username without password
				creds := base64.StdEncoding.EncodeToString([]byte("speakeasy-sdks"))
				req.Header.Set("Authorization", fmt.Sprintf("Basic %s", creds))
				return req
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid credentials in basic auth",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				creds := base64.StdEncoding.EncodeToString([]byte("wrong:wrong"))
				req.Header.Set("Authorization", fmt.Sprintf("Basic %s", creds))

				return req
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "unexpected scope in basic auth",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				form.Set("scope", "unknown") // missing write scope

				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				creds := base64.StdEncoding.EncodeToString([]byte("speakeasy-sdks:supersecret-123"))
				req.Header.Set("Authorization", fmt.Sprintf("Basic %s", creds))

				return req
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "case insensitive basic prefix",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				form.Set("scope", "read write")

				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				creds := base64.StdEncoding.EncodeToString([]byte("speakeasy-sdks:supersecret-123"))
				req.Header.Set("Authorization", fmt.Sprintf("BASIC %s", creds))

				return req
			},
			wantStatus:      http.StatusOK,
			wantAccessToken: firstAccessToken,
		},
		{
			name: "Non-expired token",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				form.Set("client_id", "speakeasy-sdks")
				form.Set("client_secret", "supersecret-123")
				form.Set("scope", "read write")

				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token?expires_in=60", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return req
			},
			wantStatus:      http.StatusOK,
			wantAccessToken: firstAccessToken,
		},
		{
			name: "custom token type",
			setupRequest: func() *http.Request {
				form := url.Values{}
				form.Set("grant_type", "client_credentials")
				form.Set("client_id", "speakeasy-sdks")
				form.Set("client_secret", "supersecret-123")
				form.Set("scope", "read write")

				req := httptest.NewRequest(http.MethodPost, "/clientcredentials/token?token_type=custom", strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return req
			},
			wantStatus:      http.StatusOK,
			wantAccessToken: firstAccessToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			HandleTokenRequest(w, tt.setupRequest())

			if got := w.Code; got != tt.wantStatus {
				t.Errorf("HandleTokenRequest() status = %v, want %v", got, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				if !strings.Contains(w.Body.String(), tt.wantAccessToken) {
					t.Errorf("HandleTokenRequest() response doesn't contain expected access token")
				}

				// Check expires_in query parameter
				if tt.name == "Non-expired token" {
					if !strings.Contains(w.Body.String(), `"expires_in":60`) {
						t.Errorf("HandleTokenRequest() response doesn't contain expected expires_in value")
					}
				}

				// Check token_type query parameter
				if tt.name == "custom token type" {
					if !strings.Contains(w.Body.String(), `"token_type":"custom"`) {
						t.Errorf("HandleTokenRequest() response doesn't contain expected token_type value")
					}
				}
			}
		})
	}
}
