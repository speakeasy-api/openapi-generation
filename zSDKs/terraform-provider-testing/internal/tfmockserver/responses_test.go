package tfmockserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		statusCode     int
		body           any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "200 with map body",
			statusCode:     http.StatusOK,
			body:           map[string]any{"id": "123", "name": "test"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"id":"123","name":"test"}` + "\n",
		},
		{
			name:       "201 with struct body",
			statusCode: http.StatusCreated,
			body: struct {
				ID string `json:"id"`
			}{ID: "456"},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":"456"}` + "\n",
		},
		{
			name:           "204 with nil body",
			statusCode:     http.StatusNoContent,
			body:           nil,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "null\n",
		},
		{
			name:           "400 with error message",
			statusCode:     http.StatusBadRequest,
			body:           map[string]string{"error": "bad request"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"bad request"}` + "\n",
		},
		{
			name:           "500 with empty map",
			statusCode:     http.StatusInternalServerError,
			body:           map[string]any{},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "{}\n",
		},
		{
			name:           "200 with slice body",
			statusCode:     http.StatusOK,
			body:           []string{"a", "b", "c"},
			expectedStatus: http.StatusOK,
			expectedBody:   `["a","b","c"]` + "\n",
		},
		{
			name:           "200 with nested structure",
			statusCode:     http.StatusOK,
			body:           map[string]any{"data": map[string]any{"nested": true}},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"data":{"nested":true}}` + "\n",
		},
		{
			name:           "200 with string body",
			statusCode:     http.StatusOK,
			body:           "plain string",
			expectedStatus: http.StatusOK,
			expectedBody:   `"plain string"` + "\n",
		},
		{
			name:           "200 with integer body",
			statusCode:     http.StatusOK,
			body:           42,
			expectedStatus: http.StatusOK,
			expectedBody:   "42\n",
		},
		{
			name:           "200 with boolean body",
			statusCode:     http.StatusOK,
			body:           true,
			expectedStatus: http.StatusOK,
			expectedBody:   "true\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()

			JSONResponse(recorder, tc.statusCode, tc.body)

			result := recorder.Result()
			defer result.Body.Close()

			if result.StatusCode != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, result.StatusCode)
			}

			contentType := result.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", contentType)
			}

			body := recorder.Body.String()
			if body != tc.expectedBody {
				t.Errorf("expected body %q, got %q", tc.expectedBody, body)
			}
		})
	}
}
