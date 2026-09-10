package polling

import (
	"encoding/json"
	"net/http"
)

// Describes the polling endpoint response object.
type OKResponse struct {
	// Current status of the polling operation.
	Status PollingStatus `json:"status"`
}

// Describes the nested polling endpoint response object.
type OKNestedResponse struct {
	// Nested polling information.
	Nested OKResponse `json:"nested"`
}

// Writes an OKResponse with given status as JSON to the HTTP response.
func WriteOKResponseAsJSON(w http.ResponseWriter, status PollingStatus) {
	body := OKResponse{
		Status: status,
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		http.Error(w, "Failed to marshal OKResponse into JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyBytes)
}

// Writes an OKNestedResponse with given status as JSON to the HTTP response.
func WriteOKNestedResponseAsJSON(w http.ResponseWriter, status PollingStatus) {
	body := OKNestedResponse{
		Nested: OKResponse{
			Status: status,
		},
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		http.Error(w, "Failed to marshal OKNestedResponse into JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyBytes)
}
