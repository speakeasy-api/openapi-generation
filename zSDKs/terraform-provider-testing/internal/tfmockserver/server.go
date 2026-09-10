package tfmockserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-provider-testing/internal/tfmockserver/swiftstore"
)

const StoreKey = "default"

type Error struct {
	// The error code
	HTTPStatus *int `json:"httpStatus,omitempty"`
	// The error message
	Message *string `json:"message,omitempty"`
	// A technical code to identify the error
	TechnicalCode *string `json:"technicalCode,omitempty"`
	// A map of parameters to be used in the error message
	Parameters map[string]string `json:"parameters,omitempty"`
}

// CapturedRequest represents a single request captured by the mock server.
type CapturedRequest struct {
	Method     string
	Path       string
	PathParams map[string]string
	Body       map[string]any
}

// RequestLog is a concurrency-safe log of captured requests.
type RequestLog struct {
	mu   sync.Mutex
	reqs []CapturedRequest
}

func (l *RequestLog) Append(req CapturedRequest) {
	l.mu.Lock()
	l.reqs = append(l.reqs, req)
	l.mu.Unlock()
}

func (l *RequestLog) All() []CapturedRequest {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]CapturedRequest, len(l.reqs))
	copy(out, l.reqs)
	return out
}

func (l *RequestLog) Bodies() []map[string]any {
	all := l.All()
	bodies := make([]map[string]any, len(all))
	for i, r := range all {
		bodies[i] = r.Body
	}
	return bodies
}

func (l *RequestLog) Reset() {
	l.mu.Lock()
	l.reqs = nil
	l.mu.Unlock()
}

// ResourceEndpoints defines the API endpoints for a server.
type ResourceEndpoints struct {
	// Close endpoints
	Close Endpoints

	// Create endpoints
	Create Endpoints

	// Get endpoints
	Get Endpoints

	// Update endpoints
	Update Endpoints

	// Delete endpoints
	Delete Endpoints

	// Invoke endpoints
	Invoke Endpoints

	// Open endpoints
	Open Endpoints
}

// Sets path parameters for each endpoint.
func (e ResourceEndpoints) setParams() {
	e.Close.setParams()
	e.Create.setParams()
	e.Get.setParams()
	e.Update.setParams()
	e.Delete.setParams()
	e.Invoke.setParams()
	e.Open.setParams()
}

// Sets status codes for each endpoint.
func (e ResourceEndpoints) setStatusCodes() {
	e.Close.setStatusCodes()
	e.Create.setStatusCodes()
	e.Get.setStatusCodes()
	e.Update.setStatusCodes()
	e.Delete.setStatusCodes()
	e.Invoke.setStatusCodes()
	e.Open.setStatusCodes()
}

// Validates the ResourceEndpoints configuration.
func (e ResourceEndpoints) validate() error {
	if len(e.Close) == 0 && len(e.Create) == 0 && len(e.Delete) == 0 && len(e.Get) == 0 && len(e.Invoke) == 0 && len(e.Open) == 0 && len(e.Update) == 0 {
		return errors.New("must set at least one of Close, Create, Delete, Get, Invoke, Open, or Update")
	}

	if err := e.Close.validate(); err != nil {
		return fmt.Errorf("invalid Close endpoints: %w", err)
	}

	if len(e.Close) > 0 && len(e.Open) == 0 {
		return errors.New("close endpoints require at least one Open endpoint")
	}

	if err := e.Create.validate(); err != nil {
		return fmt.Errorf("invalid Create endpoints: %w", err)
	}

	if err := e.Get.validate(); err != nil {
		return fmt.Errorf("invalid Get endpoints: %w", err)
	}

	if err := e.Update.validate(); err != nil {
		return fmt.Errorf("invalid Update endpoints: %w", err)
	}

	if err := e.Delete.validate(); err != nil {
		return fmt.Errorf("invalid Delete endpoints: %w", err)
	}

	if err := e.Invoke.validate(); err != nil {
		return fmt.Errorf("invalid Invoke endpoints: %w", err)
	}

	if err := e.Open.validate(); err != nil {
		return fmt.Errorf("invalid Open endpoints: %w", err)
	}

	return nil
}

// Collection of ResourceEndpoints.
type Endpoints []*ResourceEndpoint

// Sets path parameters for each endpoint.
func (e Endpoints) setParams() {
	for _, endpoint := range e {
		endpoint.setParams()
	}
}

// Sets status codes for each endpoint.
func (e Endpoints) setStatusCodes() {
	for _, endpoint := range e {
		endpoint.setStatusCodes()
	}
}

// Validates the Endpoints configuration.
func (e Endpoints) validate() error {
	for index, endpoint := range e {
		if endpoint.Endpoint == "" {
			return fmt.Errorf("index %d: must set Endpoint, such as 'GET /v0/resource/{id}'", index)
		}

		// Count how many response options are set
		responseOptionsSet := 0
		if endpoint.Response != nil {
			responseOptionsSet++
		}
		if endpoint.ResponseOverlay != nil {
			responseOptionsSet++
		}
		if endpoint.ResponseFunc != nil {
			responseOptionsSet++
		}
		if endpoint.ResponseFuncWithStore != nil {
			responseOptionsSet++
		}
		if endpoint.ResponseFuncWithBody != nil {
			responseOptionsSet++
		}

		if responseOptionsSet > 1 {
			return fmt.Errorf("index %d: cannot set more than one of Response, ResponseOverlay, ResponseFunc, ResponseFuncWithStore, or ResponseFuncWithBody", index)
		}
	}

	return nil
}

type ResourceEndpoint struct {
	// The HTTP method and path. For example: "GET /v0/resource/{id}"
	Endpoint string

	// Override the 404 Not Found HTTP status code when store data is missing.
	// Applicable for Delete, Get, and Update endpoints.
	NotFoundStatusCode int

	// Full response. If set, this will be returned as the full JSON response,
	// bypassing underlying store data. Conflicts with ResponseOverlay and ResponseFunc.
	Response map[string]any

	// Optional overlay to apply to the store data in the response. Conflicts
	// with Response and ResponseFunc.
	ResponseOverlay map[string]any

	// Optional function to dynamically generate responses based on the request.
	// This is useful for testing pagination or other scenarios where responses
	// need to vary based on request parameters. Conflicts with Response, ResponseOverlay,
	// and ResponseFuncWithStore.
	// The function receives the *http.Request and http.ResponseWriter for full control
	// over the HTTP response. Use JSONResponse helper for common JSON responses.
	ResponseFunc func(r *http.Request, w http.ResponseWriter)

	// Optional function to dynamically generate responses with access to the stored data.
	// Similar to ResponseFunc but receives the stored data as a third parameter.
	// This is useful when you need to modify specific fields while preserving the rest
	// of the stored data. Conflicts with Response, ResponseOverlay, and ResponseFunc.
	// For GET requests only.
	ResponseFuncWithStore func(r *http.Request, w http.ResponseWriter, storedData map[string]any)

	// Optional function to dynamically generate responses with access to the request body.
	// Similar to ResponseFunc but receives the parsed request body as a third parameter.
	// The data is automatically stored after this function returns.
	// This is useful for Create/Update when you need to modify the response while still
	// storing the data. Conflicts with Response, ResponseOverlay, and ResponseFunc.
	// For POST/PUT/PATCH requests only.
	ResponseFuncWithBody func(r *http.Request, w http.ResponseWriter, requestBody map[string]any)

	// Requests holds all requests captured for this configured endpoint.
	// Tests can access this after resource.Test() completes to validate request bodies.
	Requests RequestLog

	// Extracted path parameters (e.g. {...} sequences) from the Endpoint.
	params []string
}

// Captured returns all captured requests for this endpoint.
func (e *ResourceEndpoint) Captured() []CapturedRequest {
	return e.Requests.All()
}

// CapturedBodies returns only the request bodies for all captured requests.
func (e *ResourceEndpoint) CapturedBodies() []map[string]any {
	return e.Requests.Bodies()
}

// Extracts path parameters from the Endpoint and sets the params field.
func (e *ResourceEndpoint) setParams() {
	e.params = extractPathParams(e.Endpoint)
}

// Sets status codes with default values as necessary, such as
// NotFoundStatusCode to 404 Not Found.
func (e *ResourceEndpoint) setStatusCodes() {
	if e.NotFoundStatusCode == 0 {
		e.NotFoundStatusCode = http.StatusNotFound
	}
}

// ResourceHandler handles resource-specific API endpoints
type ResourceHandler struct {
	store          *swiftstore.Store
	mux            *http.ServeMux
	endpoints      ResourceEndpoints
	testingContext *testing.T
}

// Method to extract path params from a url
func extractPathParams(path string) []string {
	// This regex finds all occurrences of {any_text_without_braces}
	// The parentheses ( ... ) create a "capturing group" for the parameter name.
	re := regexp.MustCompile(`\{([^{}]+)\}`)

	// FindAllStringSubmatch returns a slice for each match.
	// Each inner slice contains the full match (e.g., "{param1}")
	// and the captured group (e.g., "param1").
	matches := re.FindAllStringSubmatch(path, -1)

	if matches == nil {
		return []string{}
	}

	// We only want the captured group (the parameter name itself).
	params := make([]string, len(matches))
	for i, match := range matches {
		params[i] = match[1] // Index 1 is the first capturing group
	}

	return params
}

// Validates the endpoints configuration, then starts and returns a new mock
// server based on that configuration.
func StartServer(endpoints ResourceEndpoints, t *testing.T) *httptest.Server {
	t.Helper()

	if err := endpoints.validate(); err != nil {
		t.Fatalf("invalid ResourceEndpoints: %v", err)
	}

	endpoints.setParams()
	endpoints.setStatusCodes()

	t.Logf("MockServer in use")
	handler := GenerateResouceHandler(endpoints, t)
	server := httptest.NewServer(handler)
	t.Logf("🚀 [TEST] Mock server started at: %s", server.URL)
	return server
}

// GenerateResouceHandler creates a new handler for resource API endpoints
func GenerateResouceHandler(endpoints ResourceEndpoints, t *testing.T) *ResourceHandler {
	handler := &ResourceHandler{
		store:          swiftstore.New(),
		endpoints:      endpoints,
		testingContext: t,
	}

	mux := http.NewServeMux()

	// Register the endpoints dynamically
	for _, resourceEndpoint := range endpoints.Close {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleCloseWithConfig(w, r, ep)
		})
	}
	for _, resourceEndpoint := range endpoints.Create {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleCreateOrUpdateWithConfig(w, r, ep)
		})
	}
	for _, resourceEndpoint := range endpoints.Get {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleGetWithConfig(w, r, ep)
		})
	}
	for _, resourceEndpoint := range endpoints.Update {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleUpdateWithConfig(w, r, ep)
		})
	}
	for _, resourceEndpoint := range endpoints.Delete {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleDeleteWithConfig(w, r, ep)
		})
	}
	for _, resourceEndpoint := range endpoints.Invoke {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleInvokeWithConfig(w, r, ep)
		})
	}
	for _, resourceEndpoint := range endpoints.Open {
		// Capture loop variable to avoid closure bug
		ep := resourceEndpoint
		mux.HandleFunc(ep.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			handler.handleOpenWithConfig(w, r, ep)
		})
	}

	handler.mux = mux
	return handler
}

// ServeHTTP delegates to the mux for routing
func (h *ResourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handleCreateOrUpdate handles PUT requests to create or update shared policy groups
func (h *ResourceHandler) handleCreateOrUpdateWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {

	h.testingContext.Logf("🔵 [MOCK SERVER] %s Request: %s\n", r.Method, r.URL.Path)
	h.testingContext.Logf("🔵 [MOCK SERVER] Headers: %+v\n", r.Header)

	// Parse request body
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.testingContext.Logf("🔴 [MOCK SERVER] Failed to decode JSON: %v\n", err)

		// Capture the request for test validation even if decode fails
		pathParams := map[string]string{}
		for _, p := range config.params {
			pathParams[p] = r.PathValue(p)
		}
		config.Requests.Append(CapturedRequest{
			Method:     r.Method,
			Path:       r.URL.Path,
			PathParams: pathParams,
			Body:       nil, // Body decode failed
		})

		h.writeError(w, 400, "Invalid JSON data", "VALIDATION_ERROR")
		return
	}

	h.testingContext.Logf("🔵 [MOCK SERVER] Request body: %+v\n", requestData)

	// Capture the request for test validation
	pathParams := map[string]string{}
	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}
	config.Requests.Append(CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       requestData,
	})

	// Check if ResponseFunc is configured - if so, use it instead of default behavior
	if config.ResponseFunc != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] %s - Using custom ResponseFunc\n\n", r.Method)
		config.ResponseFunc(r, w)
		return
	}

	// Check if ResponseFuncWithBody is configured
	if config.ResponseFuncWithBody != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] %s - Using custom ResponseFuncWithBody\n\n", r.Method)
		// Add ID to request data
		requestData["id"] = StoreKey
		// Call user function first - it writes the response
		config.ResponseFuncWithBody(r, w, requestData)
		// Store whatever the user function left in requestData (they may have modified it)
		if err := h.store.AddWithID(StoreKey, requestData); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				h.store.Update(StoreKey, requestData)
			}
		}
		return
	}

	responseData := make(map[string]interface{})
	for k, v := range requestData {
		responseData[k] = v
	}
	if len(config.params) > 0 {
		for _, param := range config.params {
			responseData[param] = r.PathValue(param)
		}
	}
	h.testingContext.Logf("stored data in store: %v\n", h.store.Data)
	responseData["id"] = StoreKey

	// Store the data
	if err := h.store.AddWithID(StoreKey, responseData); err != nil {
		// If already exists, update it
		if strings.Contains(err.Error(), "already exists") {
			if _, updateErr := h.store.Update(StoreKey, responseData); updateErr != nil {
				h.writeError(w, 500, "Failed to update resource", "INTERNAL_ERROR")
				return
			}
		} else {
			h.writeError(w, 500, "Failed to create resource", "INTERNAL_ERROR")
			return
		}
	}

	if config.ResponseOverlay != nil {
		for k, v := range config.ResponseOverlay {
			responseData[k] = v
		}
	}

	// Return success response
	h.testingContext.Logf("🟢 [MOCK SERVER] Returning 200 response: %+v\n\n", responseData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(responseData)
}

// handleGet handles GET requests to retrieve resources
func (h *ResourceHandler) handleGetWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {
	h.testingContext.Logf("🔵 [MOCK SERVER] GET Request: %s\n", r.URL.Path)

	// Capture the request for test validation
	pathParams := map[string]string{}
	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}
	config.Requests.Append(CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       nil, // GET requests typically don't have a body
	})

	if config.Response != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] GET - Returning configured response: %+v\n\n", config.Response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(config.Response)
		return
	}

	if config.ResponseFunc != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] GET - Using custom ResponseFunc\n\n")
		config.ResponseFunc(r, w)
		return
	}

	h.testingContext.Logf("🔵 [MOCK SERVER] GET - Id: %s,\n", StoreKey)
	h.testingContext.Logf("🔵 [MOCK SERVER] GET - Looking up key: %s\n", StoreKey)
	responseData := make(map[string]interface{})
	// If no ID in path, get the first (and only) item from store

	data, err := h.store.Get(StoreKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.testingContext.Logf("🔴 [MOCK SERVER] GET - Resource not found: %s\n", StoreKey)
			w.WriteHeader(config.NotFoundStatusCode)
		} else {
			h.testingContext.Logf("🔴 [MOCK SERVER] GET - Internal error: %v\n", err)
			h.writeError(w, 500, "Internal server error", "INTERNAL_ERROR")
		}
		return
	}

	// If ResponseFuncWithStore is set, call it with the stored data
	if config.ResponseFuncWithStore != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] GET - Using custom ResponseFuncWithStore\n\n")
		// Add id to the data before passing to the function
		data["id"] = StoreKey
		config.ResponseFuncWithStore(r, w, data)
		return
	}

	// Create response with the ID included
	for k, v := range data {
		responseData[k] = v
	}

	responseData["id"] = StoreKey

	if config.ResponseOverlay != nil {
		for k, v := range config.ResponseOverlay {
			responseData[k] = v
		}
	}

	// Return the stored data
	h.testingContext.Logf("🟢 [MOCK SERVER] GET - Returning data: %+v\n\n", responseData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(responseData)
}

// handleDeleteWithConfig handles DELETE requests to remove resources
func (h *ResourceHandler) handleDeleteWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {
	h.testingContext.Logf("🔵 [MOCK SERVER] DELETE Request: %s\n", r.URL.Path)

	// Capture the request for test validation
	pathParams := map[string]string{}
	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}
	config.Requests.Append(CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       nil, // DELETE requests typically don't have a body
	})

	h.testingContext.Logf("🔵 [MOCK SERVER] DELETE - Id: %s,\n", StoreKey)

	// If no ID in path, delete the first (and only) item from store
	h.testingContext.Logf("🔵 [MOCK SERVER] DELETE - Deleting key: %s\n", StoreKey)

	if err := h.store.Delete(StoreKey); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.testingContext.Logf("🔴 [MOCK SERVER] DELETE - Resource not found: %s\n", StoreKey)
			w.WriteHeader(config.NotFoundStatusCode)
		} else {
			h.testingContext.Logf("🔴 [MOCK SERVER] DELETE - Internal error: %v\n", err)
			h.writeError(w, 500, "Internal server error", "INTERNAL_ERROR")
		}
		return
	}

	responseData := make(map[string]interface{})
	if config.ResponseOverlay != nil {
		for k, v := range config.ResponseOverlay {
			responseData[k] = v
		}
	}

	// Return 200 OK for successful deletion
	h.testingContext.Logf("🟢 [MOCK SERVER] DELETE - Successfully deleted: %s\n", StoreKey)
	if len(responseData) == 0 {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"message": "Resource deleted successfully"})
	} else {
		h.testingContext.Logf("🟢 [MOCK SERVER] DELETE - Returning data after successfully deleting data: %+v\n\n", responseData)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(responseData)
	}
}

// handleUpdate handles PATCH requests to update resources
func (h *ResourceHandler) handleUpdateWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {
	h.testingContext.Logf("🔵 [MOCK SERVER] %s Request: %s\n", r.Method, r.URL.Path)
	h.testingContext.Logf("🔵 [MOCK SERVER] Headers: %+v\n", r.Header)

	// Parse request body
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.testingContext.Logf("🔴 [MOCK SERVER] Failed to decode JSON: %v\n", err)
		h.writeError(w, 400, "Invalid JSON data", "VALIDATION_ERROR")
		return
	}

	h.testingContext.Logf("🔵 [MOCK SERVER] Request body: %+v\n", requestData)

	// Capture the request for test validation
	pathParams := map[string]string{}
	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}
	config.Requests.Append(CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       requestData,
	})

	h.testingContext.Logf("🔵 [MOCK SERVER] UPDATE - Id: %s,\n", StoreKey)

	// Check if resource exists
	existingData, err := h.store.Get(StoreKey)
	if err != nil {
		h.testingContext.Logf("🔴 [MOCK SERVER] UPDATE - Resource not found: %s\n", StoreKey)
		w.WriteHeader(config.NotFoundStatusCode)
		return
	}

	// Merge existing data with update data
	responseData := make(map[string]interface{})
	for k, v := range existingData {
		responseData[k] = v
	}
	for k, v := range requestData {
		responseData[k] = v
	}
	responseData["id"] = StoreKey

	// Update the data in store
	if _, updateErr := h.store.Update(StoreKey, responseData); updateErr != nil {
		h.testingContext.Logf("🔴 [MOCK SERVER] UPDATE - Failed to update: %v\n", updateErr)
		h.writeError(w, 500, "Failed to update resource", "INTERNAL_ERROR")
		return
	}

	// Apply update response overlay if configured
	if config.ResponseOverlay != nil {
		for k, v := range config.ResponseOverlay {
			responseData[k] = v
		}
	}

	// Return success response
	h.testingContext.Logf("🟢 [MOCK SERVER] UPDATE - Returning 200 response: %+v\n\n", responseData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(responseData)
}

func (h *ResourceHandler) handleInvokeWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {
	h.testingContext.Logf("🔵 [MOCK SERVER] %s Request: %s\n", r.Method, r.URL.Path)
	h.testingContext.Logf("🔵 [MOCK SERVER] Headers: %+v\n", r.Header)

	pathParams := make(map[string]string, len(config.params))

	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}

	capturedRequest := CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       make(map[string]any),
	}

	if err := json.NewDecoder(r.Body).Decode(&capturedRequest.Body); err != nil {
		h.testingContext.Logf("🔴 [MOCK SERVER] Failed to decode JSON: %v\n", err)
		config.Requests.Append(capturedRequest)
		h.writeError(w, 400, "Invalid JSON data", "VALIDATION_ERROR")

		return
	}

	h.testingContext.Logf("🔵 [MOCK SERVER] Request body: %+v\n", capturedRequest.Body)
	config.Requests.Append(capturedRequest)

	if config.ResponseFunc != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] %s - Using custom ResponseFunc\n\n", r.Method)
		config.ResponseFunc(r, w)
		return
	}

	if config.ResponseFuncWithBody != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] %s - Using custom ResponseFuncWithBody\n\n", r.Method)
		config.ResponseFuncWithBody(r, w, capturedRequest.Body)
		return
	}

	responseData := make(map[string]any, len(capturedRequest.Body)+len(config.params))

	for k, v := range capturedRequest.Body {
		responseData[k] = v
	}

	if len(config.params) > 0 {
		for _, param := range config.params {
			responseData[param] = r.PathValue(param)
		}
	}

	if config.ResponseOverlay != nil {
		for k, v := range config.ResponseOverlay {
			responseData[k] = v
		}
	}

	h.testingContext.Logf("🟢 [MOCK SERVER] INVOKE - Returning 200 response: %+v\n\n", responseData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(responseData)
}

func (h *ResourceHandler) handleCloseWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {
	h.testingContext.Logf("🔵 [MOCK SERVER] CLOSE Request: %s\n", r.URL.Path)

	pathParams := make(map[string]string, len(config.params))

	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}

	capturedRequest := CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       make(map[string]any),
	}

	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest.Body); err != nil {
			h.testingContext.Logf("🔴 [MOCK SERVER] CLOSE - Failed to decode JSON: %v\n", err)
			config.Requests.Append(capturedRequest)
			h.writeError(w, 400, "Invalid JSON data", "VALIDATION_ERROR")

			return
		}
	}

	h.testingContext.Logf("🔵 [MOCK SERVER] CLOSE - Request body: %+v\n", capturedRequest.Body)
	config.Requests.Append(capturedRequest)

	if config.ResponseFunc != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] CLOSE - Using custom ResponseFunc\n\n")
		config.ResponseFunc(r, w)
		return
	}

	// Look up stored data from Open operation. Unlike delete, close
	// operations should succeed even if the store is empty (e.g., after a
	// previous close operation in a chain already cleaned it up).
	storedData, err := h.store.Get(StoreKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.testingContext.Logf("🔵 [MOCK SERVER] CLOSE - No stored data found in store: %s (this is OK for chained close operations)\n", StoreKey)
		} else {
			h.testingContext.Logf("🔴 [MOCK SERVER] CLOSE - Internal error: %v\n", err)
			h.writeError(w, 500, "Internal server error", "INTERNAL_ERROR")

			return
		}
	} else {
		h.testingContext.Logf("🔵 [MOCK SERVER] CLOSE - Found stored data: %+v\n", storedData)

		// Clean up stored data
		if deleteErr := h.store.Delete(StoreKey); deleteErr != nil {
			h.testingContext.Logf("🔴 [MOCK SERVER] CLOSE - Failed to clean up stored data: %v\n", deleteErr)
		}
	}

	h.testingContext.Logf("🟢 [MOCK SERVER] CLOSE - Returning 200 response\n\n")
	w.WriteHeader(200)
}

func (h *ResourceHandler) handleOpenWithConfig(w http.ResponseWriter, r *http.Request, config *ResourceEndpoint) {
	h.testingContext.Logf("🔵 [MOCK SERVER] %s Request: %s\n", r.Method, r.URL.Path)
	h.testingContext.Logf("🔵 [MOCK SERVER] Headers: %+v\n", r.Header)

	pathParams := make(map[string]string, len(config.params))

	for _, p := range config.params {
		pathParams[p] = r.PathValue(p)
	}

	capturedRequest := CapturedRequest{
		Method:     r.Method,
		Path:       r.URL.Path,
		PathParams: pathParams,
		Body:       make(map[string]any),
	}

	if err := json.NewDecoder(r.Body).Decode(&capturedRequest.Body); err != nil {
		h.testingContext.Logf("🔴 [MOCK SERVER] Failed to decode JSON: %v\n", err)
		config.Requests.Append(capturedRequest)
		h.writeError(w, 400, "Invalid JSON data", "VALIDATION_ERROR")

		return
	}

	h.testingContext.Logf("🔵 [MOCK SERVER] Request body: %+v\n", capturedRequest.Body)
	config.Requests.Append(capturedRequest)

	if config.ResponseFunc != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] %s - Using custom ResponseFunc\n\n", r.Method)
		config.ResponseFunc(r, w)
		return
	}

	if config.ResponseFuncWithBody != nil {
		h.testingContext.Logf("🟢 [MOCK SERVER] %s - Using custom ResponseFuncWithBody\n\n", r.Method)

		// Call user function first - it writes the response
		config.ResponseFuncWithBody(r, w, capturedRequest.Body)

		// Store whatever the user function left in capturedRequest.Body (they may have modified it)
		if err := h.store.AddWithID(StoreKey, capturedRequest.Body); err != nil {
			if strings.Contains(err.Error(), "already exists") {
				h.store.Update(StoreKey, capturedRequest.Body)
			}
		}

		return
	}

	responseData := make(map[string]any, len(capturedRequest.Body)+len(config.params))

	for k, v := range capturedRequest.Body {
		responseData[k] = v
	}

	if len(config.params) > 0 {
		for _, param := range config.params {
			responseData[param] = r.PathValue(param)
		}
	}

	if config.ResponseOverlay != nil {
		for k, v := range config.ResponseOverlay {
			responseData[k] = v
		}
	}

	if err := h.store.AddWithID(StoreKey, responseData); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			if _, updateErr := h.store.Update(StoreKey, responseData); updateErr != nil {
				h.writeError(w, 500, "Failed to update resource", "INTERNAL_ERROR")

				return
			}
		} else {
			h.writeError(w, 500, "Failed to create resource", "INTERNAL_ERROR")

			return
		}
	}

	h.testingContext.Logf("🟢 [MOCK SERVER] stored data in store: %v\n", h.store.Data)
	h.testingContext.Logf("🟢 [MOCK SERVER] Returning 200 response: %+v\n\n", responseData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(responseData)
}

// writeError writes a FrameworkTypeResource-formatted error response
func (h *ResourceHandler) writeError(w http.ResponseWriter, statusCode int, message, technicalCode string) {
	errorResponse := Error{
		HTTPStatus:    &statusCode,
		Message:       &message,
		TechnicalCode: &technicalCode,
		Parameters:    make(map[string]string),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// DecodeJSONBody decodes the JSON request body into the provided map.
func DecodeJSONBody(r *http.Request, body *map[string]any) error {
	return json.NewDecoder(r.Body).Decode(body)
}
