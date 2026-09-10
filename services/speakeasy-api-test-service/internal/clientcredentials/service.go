package clientcredentials

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var state = sync.Map{}

const (
	firstAccessToken  = "super-duper-access-token"
	secondAccessToken = "second-super-duper-access-token"
)

func handleBasicAuth(authHeader string) (clientID, clientSecret string, ok bool) {
	// Remove "Basic " prefix (case-insensitive)
	if !strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
		return "", "", false
	}
	encodedCreds := strings.TrimSpace(authHeader[6:])

	// Decode base64
	decodedCreds, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encodedCreds))
	if err != nil {
		return "", "", false
	}

	// Split into username:password
	creds := strings.SplitN(string(decodedCreds), ":", 2)
	if len(creds) != 2 {
		return "", "", false
	}

	return creds[0], creds[1], true
}

// Generate a fake access token for valid client credentials.
// Supports both Basic Auth and form-encoded client_id/client_secret
// Valid Credentials:
// - client_id: "speakeasy-sdks"
// - client_secret: "supersecret-<anything>"
//
// Optional Query Parameters:
// - token_type: defaults to "Bearer"
// - expires_in: defaults to 0 (already expired)
// - skip_scopes: defaults to false
func handleTokenRequest(w http.ResponseWriter, r *http.Request, requiredScopes ...string) {
	var clientID, clientSecret string
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	// Check for Basic Auth header
	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
		var ok bool
		clientID, clientSecret, ok = handleBasicAuth(authHeader)
		if !ok {
			http.Error(w, "invalid_client", http.StatusUnauthorized)
			return
		}
	} else {
		clientID = r.Form.Get("client_id")
		clientSecret = r.Form.Get("client_secret")
	}
	grant_type := r.Form.Get("grant_type")
	if grant_type != "client_credentials" {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}

	if clientID == "" || clientSecret == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	if clientID != "speakeasy-sdks" || !strings.HasPrefix(clientSecret, "supersecret-") {
		http.Error(w, "invalid_client", http.StatusUnauthorized)
		return
	}

	if r.URL.Query().Get("skip_scopes") != "true" {
		// Check all required scopes were provided
		scopes := strings.Fields(r.Form.Get("scope")) // splits and remove empty strings
		if len(scopes) == 0 {
			http.Error(w, "empty_scopes", http.StatusBadRequest)
			return
		}
		for _, s := range requiredScopes {
			if !slices.Contains(scopes, s) {
				http.Error(w, fmt.Sprintf("missing_scope: %s. requested %s", s, scopes), http.StatusBadRequest)
				return
			}
		}
	}

	tokenType := r.URL.Query().Get("token_type")
	if tokenType == "" {
		tokenType = "Bearer" // default
	}

	expiresIn := 0 // default
	expiresInStr := r.URL.Query().Get("expires_in")
	if expiresInStr != "" {
		var err error
		if expiresIn, err = strconv.Atoi(expiresInStr); err != nil {
			http.Error(w, "could not convert expires_in query parameter to integer", http.StatusBadRequest)
			return
		}
	}

	accessToken := firstAccessToken

	_, ok := state.Load(clientSecret)
	if !ok {
		state.Store(clientSecret, true)
	} else {
		accessToken = secondAccessToken
	}

	w.Header().Set("Content-Type", "application/json")

	type tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	response := tokenResponse{
		AccessToken: accessToken,
		TokenType:   tokenType,
		ExpiresIn:   expiresIn,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
}
func HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	handleTokenRequest(w, r, "read", "write")
}

func HandleAltTokenRequest(w http.ResponseWriter, r *http.Request) {
	handleTokenRequest(w, r, "alt:one", "alt:two")
}

func HandleAuthenticatedRequest(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	if accessToken != firstAccessToken && accessToken != secondAccessToken {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
