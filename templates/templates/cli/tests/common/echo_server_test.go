package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// newEchoServer creates an httptest.Server that reflects requests as JSON.
//
// Routing:
//   - /status/{code} → returns that HTTP status code with {"status_code": <code>, "error": "<text>"}
//   - /errors/{code} → returns that status code with {"message": "error from echo server", "code": "ERROR_CODE", "status_code": <code>}
//   - /followRedirect/oldPage → 302 redirect to /followRedirect/newPage (which returns {"name": "John Doe"})
//   - /responseObjectWith* → returns optional/nullable mock responses based on ?scenario= query param
//   - All other paths → returns 200 with a JSON reflection of the request:
//     {"url": "/path", "method": "GET", "headers": {...}, "args": {...}, "data": "...", "json": {...}}
//
//nolint:unused // used by generated test code
func newEchoServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Handle /status/{code} paths (for error tests like statusGetError)
		if strings.HasPrefix(r.URL.Path, "/status/") {
			codeStr := strings.TrimPrefix(r.URL.Path, "/status/")
			code, err := strconv.Atoi(codeStr)
			if err != nil {
				code = 200
			}

			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status_code": code,
				"error":       http.StatusText(code),
			})
			return
		}

		// Handle /errors/{code} paths (for x-speakeasy-errors tests like statusGetXSpeakeasyErrors)
		if strings.HasPrefix(r.URL.Path, "/errors/") {
			codeStr := strings.TrimPrefix(r.URL.Path, "/errors/")
			code, err := strconv.Atoi(codeStr)
			if err != nil {
				code = 500
			}

			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"message":     "error from echo server",
				"code":        "ERROR_CODE",
				"status_code": code,
			})
			return
		}

		// Handle /followRedirect/oldPage → 302 redirect to /followRedirect/newPage
		if r.URL.Path == "/followRedirect/oldPage" {
			http.Redirect(w, r, "/followRedirect/newPage", http.StatusFound)
			return
		}
		if r.URL.Path == "/followRedirect/newPage" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "John Doe",
			})
			return
		}

		// Handle /responseObjectWith* paths (for optional/nullable tests)
		if strings.HasPrefix(r.URL.Path, "/responseObjectWith") {
			scenario := r.URL.Query().Get("scenario")
			handleOptionalNullableResponse(w, scenario)
			return
		}

		// Read request body
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr := string(bodyBytes)

		// Parse query parameters
		args := map[string]interface{}{}
		for k, v := range r.URL.Query() {
			if len(v) == 1 {
				args[k] = v[0]
			} else {
				args[k] = v
			}
		}

		// Parse headers
		headers := map[string]interface{}{}
		for k, v := range r.Header {
			if len(v) == 1 {
				headers[k] = v[0]
			} else {
				headers[k] = v
			}
		}

		// Try to parse body as JSON
		var jsonBody interface{}
		if len(bodyBytes) > 0 {
			_ = json.Unmarshal(bodyBytes, &jsonBody)
		}

		// Build reflection response
		resp := map[string]interface{}{
			"url":     r.URL.Path,
			"method":  r.Method,
			"headers": headers,
			"args":    args,
			"data":    bodyStr,
		}
		if jsonBody != nil {
			resp["json"] = jsonBody
		}

		_ = json.NewEncoder(w).Encode(resp)
	}))
}

// handleOptionalNullableResponse returns mock pet data based on the scenario.
//
//nolint:unused // used by generated test code
func handleOptionalNullableResponse(w http.ResponseWriter, scenario string) {
	base := map[string]interface{}{
		"id":   123,
		"name": "fluffy",
		"type": "cat",
	}
	switch scenario {
	case "full":
		base["medicalRecord"] = map[string]interface{}{
			"description": "fluffy is sick",
		}
	case "medicalRecordNull":
		base["medicalRecord"] = nil
	case "medicalRecordAbsent":
		// do not add medicalRecord — it should be absent from the JSON
	default:
		base["medicalRecord"] = map[string]interface{}{
			"description": "fluffy is sick",
		}
	}
	_ = json.NewEncoder(w).Encode(base)
}
