package swiftstore

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreWithHTTPServer(t *testing.T) {
	// Create handler and server with route configuration
	routes := []RouteConfig{
		{Method: "POST", Endpoint: "/data", AutoID: true},
		{Method: "POST", Endpoint: "/data/custom", AutoID: false},
		{Method: "GET", Endpoint: "/data", AutoID: false},    // GET always requires ID
		{Method: "PUT", Endpoint: "/data", AutoID: false},    // PUT always requires ID
		{Method: "DELETE", Endpoint: "/data", AutoID: false}, // DELETE always requires ID
	}
	handler := NewStoreHandler(routes)
	server := httptest.NewServer(handler)
	defer server.Close()

	// Test adding data with auto-generated ID
	t.Run("Add with auto-generated ID", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John Doe",
			"age":  30.0, // Use float64 to match JSON unmarshaling behavior
		}

		jsonData, _ := json.Marshal(data)
		resp, err := http.Post(server.URL+"/data", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		// Extract the generated ID
		id := result["id"].(string)
		require.NotEmpty(t, id)

		// Test retrieving the data
		resp, err = http.Get(server.URL + "/data/" + id)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var retrieved map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&retrieved)
		resp.Body.Close()

		require.Equal(t, id, retrieved["id"])
		require.Equal(t, "John Doe", retrieved["name"])
		require.Equal(t, 30.0, retrieved["age"])

		// Test updating the data
		updateData := map[string]interface{}{
			"age":  31.0,
			"city": "New York",
		}
		updateJSON, _ := json.Marshal(updateData)

		req, err := http.NewRequest("PUT", server.URL+"/data/"+id, bytes.NewBuffer(updateJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&updatedResult)
		resp.Body.Close()

		// Verify the update response
		require.Equal(t, id, updatedResult["id"])
		require.Equal(t, "John Doe", updatedResult["name"]) // Original field preserved
		require.Equal(t, 31.0, updatedResult["age"])        // Updated field
		require.Equal(t, "New York", updatedResult["city"]) // New field added

		// Test deleting the data
		req, err = http.NewRequest("DELETE", server.URL+"/data/"+id, nil)
		require.NoError(t, err)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

	})

	// Test adding data with custom ID
	t.Run("Add with custom ID", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "Jane Smith",
			"age":  25.0, // Use float64 to match JSON unmarshaling behavior
		}

		jsonData, _ := json.Marshal(data)
		resp, err := http.Post(server.URL+"/data/custom/user-123", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Test retrieving the data
		resp, err = http.Get(server.URL + "/data/user-123")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var retrieved map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&retrieved)
		resp.Body.Close()

		require.Equal(t, "user-123", retrieved["id"])
		require.Equal(t, "Jane Smith", retrieved["name"])
		require.Equal(t, 25.0, retrieved["age"])

		// Test updating the data
		updateData := map[string]interface{}{
			"age":        26.0,
			"department": "Engineering",
		}
		updateJSON, _ := json.Marshal(updateData)

		req, err := http.NewRequest("PUT", server.URL+"/data/user-123", bytes.NewBuffer(updateJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedResult map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&updatedResult)
		resp.Body.Close()

		// Verify the update response
		require.Equal(t, "user-123", updatedResult["id"])
		require.Equal(t, "Jane Smith", updatedResult["name"])        // Original field preserved
		require.Equal(t, 26.0, updatedResult["age"])                 // Updated field
		require.Equal(t, "Engineering", updatedResult["department"]) // New field added

		// Test deleting the data
		req, err = http.NewRequest("DELETE", server.URL+"/data/user-123", nil)
		require.NoError(t, err)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		resp.Body.Close()

	})

	// Test GET with non-existent ID
	t.Run("Get non-existent ID", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/data/non-existent")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})

	// Test PUT with non-existent ID
	t.Run("Update non-existent ID", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name": "Test User",
		}
		updateJSON, _ := json.Marshal(updateData)

		req, err := http.NewRequest("PUT", server.URL+"/data/non-existent", bytes.NewBuffer(updateJSON))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})

	// Test DELETE with non-existent ID
	t.Run("Delete non-existent ID", func(t *testing.T) {
		req, err := http.NewRequest("DELETE", server.URL+"/data/non-existent", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})
}
