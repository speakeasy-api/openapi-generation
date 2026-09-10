package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/pkg/models"
)

func HandleCustomAuth(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/auth/customsecurity/customSchemeAppId": // tests/specs/uber.yaml
		handleCustomSchemeAppId(w, r)
	case "/auth/customsecurity/customHttpOnly": // tests/overlays/custom-http/overlay.yaml
		handleCustomHttpOnly(w, r)
	default:
		utils.HandleError(w, fmt.Errorf("invalid path"))
		return
	}
}

func handleCustomSchemeAppId(w http.ResponseWriter, r *http.Request) {
	appID := r.Header.Get("X-Security-App-Id")
	if appID != "testAppID" {
		utils.HandleError(w, fmt.Errorf("invalid app id: %w", authError))
		return
	}

	secret := r.Header.Get("X-Security-Secret")
	if secret != "testSecret" {
		utils.HandleError(w, fmt.Errorf("invalid secret: %w", authError))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleCustomHttpOnly(w http.ResponseWriter, r *http.Request) {
	expectedUserID := 54321
	expectedRole := "manager"
	expectedPassphrase := "secure-passphrase-123"
	expectedAccessCode := 104

	userID, _ := strconv.Atoi(r.Header.Get("X-Security-UserID"))
	if userID != expectedUserID {
		utils.HandleError(w, fmt.Errorf("invalid 'userID': '%v'. expected '%v'.", userID, expectedUserID))
		return
	}

	role := r.Header.Get("X-Security-Role")
	if role != expectedRole {
		utils.HandleError(w, fmt.Errorf("invalid 'role': '\"%s\"'. expected '\"%s\"'.", role, expectedRole))
		return
	}

	passphrase := r.Header.Get("X-Security-Passphrase")
	if passphrase != expectedPassphrase {
		utils.HandleError(w, fmt.Errorf("invalid 'passphrase': '\"%s\"'. expected '\"%s\"'.", passphrase, expectedPassphrase))
		return
	}

	accessCode, _ := strconv.Atoi(r.Header.Get("X-Security-AccessCode"))
	if accessCode != expectedAccessCode {
		utils.HandleError(w, fmt.Errorf("invalid 'accessCode': '%v'. expected '%v'.", accessCode, expectedAccessCode))
		return
	}

	var scopes []string
	scopesStr := r.Header.Get("X-Security-Scopes")
	if scopesStr != "" {
		json.Unmarshal([]byte(scopesStr), &scopes)
	}

	response := models.CustomHttpAuthResponse{
		Grant:  "access_granted",
		Scopes: scopes,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
