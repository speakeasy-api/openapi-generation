package tfmockserver

import (
	"encoding/json"
	"net/http"
)

// JSONResponse is a helper function for ResponseFunc implementations to write JSON responses.
// It sets the appropriate headers, status code, and encodes the body as JSON.
//
// Example usage:
//
//	ResponseFunc: func(r *http.Request, w http.ResponseWriter) {
//	    tfmockserver.JSONResponse(w, 200, map[string]any{"id": "123", "name": "test"})
//	}
func JSONResponse(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(body)
}
