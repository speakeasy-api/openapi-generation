package redirects

import (
	"encoding/json"
	"net/http"
)

// HandleRedirectOldPage handles the /followRedirect endpoint that returns a 302 redirect
func HandleRedirectOldPage(w http.ResponseWriter, r *http.Request) {
	// Set the Location header to the redirect target
	w.Header().Set("Location", "/followRedirect/newPage")
	// Return 302 Found status
	w.WriteHeader(http.StatusFound)
}

// HandleRedirectNewPage handles the /followRedirect/newPage endpoint that returns a 200 with JSON
func HandleRedirectNewPage(w http.ResponseWriter, r *http.Request) {
	// Set content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Create the response object according to the OpenAPI spec
	response := map[string]string{
		"name": "John Doe",
	}

	// Encode and send the JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
