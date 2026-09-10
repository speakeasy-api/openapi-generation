package responseHeaders

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

func HandleVendorJsonResponseHeaders(w http.ResponseWriter, r *http.Request) {
	var obj interface{}

	err := json.Unmarshal([]byte("{\"name\":\"Panda\"}"), &obj)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.api+json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(obj); err != nil {
		utils.HandleError(w, err)
	}
}

func HandleResponseHeaders(w http.ResponseWriter, r *http.Request) {
	// Extract path parameter
	vars := mux.Vars(r)
	code := vars["statusCode"]

	switch code {
	case "200":
		// Return 200 OK with required header and token
		if r.URL.Query().Get("includeRequiredHeader") != "false" {
			w.Header().Set("X-Required-Header", "required")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{"token": "test-token"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			utils.HandleError(w, err)
		}
	case "400":
		// Return 400 Bad Request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
	default:
		// Return 500 Error with response body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		errorStr := "Internal server error."
		if err := json.NewEncoder(w).Encode(errorStr); err != nil {
			utils.HandleError(w, err)
		}
	}
}

func HandleErrorResponseHeaders(w http.ResponseWriter, r *http.Request) {
	// Extract path parameter
	vars := mux.Vars(r)
	code := vars["statusCode"]

	// Check query parameter for optional headers
	includeHeader := r.URL.Query().Get("includeHeader") != "false"

	switch code {
	case "200":
		// Return 200 OK with no custom headers
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]string{"token": "test-token"}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			utils.HandleError(w, err)
		}
	case "400":
		// Return 400 Bad Request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
	case "429":
		// If requested, include "Retry-After" header
		if includeHeader {
			w.Header().Set("Retry-After", "60")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)

		errorStr := "Too many attempts. Please try again later."
		if err := json.NewEncoder(w).Encode(errorStr); err != nil {
			utils.HandleError(w, err)
		}

	default:
		// Return the requested status code in X-Status-Header
		if includeHeader {
			w.Header().Set("X-500-Header", "ok")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		errorStr := "Internal server error."
		if err := json.NewEncoder(w).Encode(errorStr); err != nil {
			utils.HandleError(w, err)
		}
	}
}
