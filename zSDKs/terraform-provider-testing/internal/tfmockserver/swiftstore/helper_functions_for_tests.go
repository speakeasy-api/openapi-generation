package swiftstore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RouteConfig defines the configuration for a single route
type RouteConfig struct {
	Method   string // HTTP method (e.g., "POST", "GET", "PUT", "DELETE")
	Endpoint string // Endpoint path (e.g., "/data", "/users", "/orders")
	AutoID   bool   // Whether to use auto-generated ID or custom ID
}

// StoreHandler wraps the store and provides HTTP endpoints
type StoreHandler struct {
	store *Store
	mux   *http.ServeMux
}

// NewStoreHandler creates a new handler with a store instance and configurable routes
func NewStoreHandler(routes []RouteConfig) *StoreHandler {
	handler := &StoreHandler{
		store: New(),
	}

	// Create mux with dynamic route configuration
	mux := http.NewServeMux()

	// Register routes based on configuration
	for _, route := range routes {
		switch route.Method {
		case "POST":
			if route.AutoID {
				// Auto-generated ID route
				mux.HandleFunc("POST "+route.Endpoint, handler.handlePostAutoID)
			} else {
				// Custom ID route - add {id} parameter
				customEndpoint := route.Endpoint + "/{id}"
				mux.HandleFunc("POST "+customEndpoint, handler.handlePostCustomID)
			}
		case "DELETE":
			// DELETE always requires an ID, so add {id} parameter
			deleteEndpoint := route.Endpoint + "/{id}"
			mux.HandleFunc("DELETE "+deleteEndpoint, handler.handleDelete)
		case "PUT":
			// PUT always requires an ID, so add {id} parameter
			putEndpoint := route.Endpoint + "/{id}"
			mux.HandleFunc("PUT "+putEndpoint, handler.handlePut)
		case "GET":
			// GET always requires an ID, so add {id} parameter
			getEndpoint := route.Endpoint + "/{id}"
			mux.HandleFunc("GET "+getEndpoint, handler.handleGet)
			// Future HTTP methods can be added here
		}
	}

	handler.mux = mux
	return handler
}

// ServeHTTP delegates to the mux for routing
func (h *StoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// handlePostAutoID handles POST requests to /data (auto-generated ID)
func (h *StoreHandler) handlePostAutoID(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Add data with auto-generated ID
	id, err := h.store.Add(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error adding data: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the stored data to return (preserves types)
	storedData, err := h.store.Get(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving stored data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Create flattened response by merging ID with data
	response := map[string]interface{}{
		"id": id,
	}
	// Add all data fields at the top level
	for key, value := range storedData {
		response[key] = value
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handlePostCustomID handles POST requests to /data/custom/{id}
func (h *StoreHandler) handlePostCustomID(w http.ResponseWriter, r *http.Request) {
	// Extract custom ID from path parameters
	customID := r.PathValue("id")
	if customID == "" {
		http.Error(w, "Custom ID cannot be empty", http.StatusBadRequest)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Add data with custom ID
	if err := h.store.AddWithID(customID, data); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, fmt.Sprintf("ID %s already exists", customID), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Error adding data: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Get the stored data to return (preserves types)
	storedData, err := h.store.Get(customID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving stored data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Create flattened response by merging ID with data
	response := map[string]interface{}{
		"id": customID,
	}
	// Add all data fields at the top level
	for key, value := range storedData {
		response[key] = value
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handleDelete handles DELETE requests to /{endpoint}/{id}
func (h *StoreHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path parameters
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID cannot be empty", http.StatusBadRequest)
		return
	}

	// Delete data from store
	if err := h.store.Delete(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Data with ID %s not found", id), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error deleting data: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return 204 No Content for successful deletion
	w.WriteHeader(http.StatusNoContent)
}

// handlePut handles PUT requests to /{endpoint}/{id}
func (h *StoreHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path parameters
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID cannot be empty", http.StatusBadRequest)
		return
	}

	// Parse the update data from request body
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Update data in store
	mergedData, err := h.store.Update(id, updateData)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Data with ID %s not found", id), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error updating data: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return the merged data flattened (ID + data fields at same level)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create flattened response by merging ID with data
	response := map[string]interface{}{
		"id": id,
	}
	// Add all data fields at the top level
	for key, value := range mergedData {
		response[key] = value
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// handleGet handles GET requests to /{endpoint}/{id}
func (h *StoreHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path parameters
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID cannot be empty", http.StatusBadRequest)
		return
	}

	// Fetch data from store
	data, err := h.store.Get(id)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, fmt.Sprintf("Data with ID %s not found", id), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Error retrieving data: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return the data as flattened JSON (ID + data fields at same level)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Create flattened response by merging ID with data
	response := map[string]interface{}{
		"id": id,
	}
	// Add all data fields at the top level
	for key, value := range data {
		response[key] = value
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
