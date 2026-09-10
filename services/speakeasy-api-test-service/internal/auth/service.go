package auth

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/pkg/models"
)

func HandleAuthInspectToken(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"authenticated": true,
		"token":         authHeader,
	})
}

func HandleAuth(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	var req models.AuthRequest
	if err := json.Unmarshal(body, &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	if err := checkAuth(req, r); err != nil {
		utils.HandleError(w, err)
		return
	}
}
