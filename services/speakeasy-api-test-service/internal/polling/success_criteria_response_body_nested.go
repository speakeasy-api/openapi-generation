package polling

import (
	"net/http"
	"sync"
)

var (
	// Mapping of request identifiers to request count.
	successCriteriaResponseBodyNestedRequests = make(map[string]int)

	// Mutex to protect access to successCriteriaResponseBodyNestedRequests.
	successCriteriaResponseBodyNestedMutex sync.Mutex
)

// Verifies polling successCriteria functionality based on response body.
// Based on requestID, this endpoint returns 200 OK with initial status of
// "pending", then "running", and then "completed".
func SuccessCriteriaResponseBodyNestedGetHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("requestID")

	if requestID == "" {
		http.Error(w, "missing requestID parameter", http.StatusBadRequest)
		return
	}

	successCriteriaResponseBodyNestedMutex.Lock()
	defer successCriteriaResponseBodyNestedMutex.Unlock()
	count, ok := successCriteriaResponseBodyNestedRequests[requestID]

	if !ok {
		successCriteriaResponseBodyNestedRequests[requestID] = 1
		WriteOKNestedResponseAsJSON(w, PollingStatusPending)
		return
	}

	successCriteriaResponseBodyNestedRequests[requestID] = count + 1

	switch count {
	case 1:
		WriteOKNestedResponseAsJSON(w, PollingStatusRunning)
	default:
		WriteOKNestedResponseAsJSON(w, PollingStatusCompleted)
	}
}
