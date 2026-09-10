package polling

import (
	"net/http"
	"sync"
)

var (
	// Mapping of request identifiers to request count.
	successCriteriaResponseBodyRequests = make(map[string]int)

	// Mutex to protect access to successCriteriaResponseBodyRequests.
	successCriteriaResponseBodyMutex sync.Mutex
)

// Verifies polling successCriteria functionality based on response body.
// Based on requestID, this endpoint returns 200 OK with initial status of
// "pending", then "running", and then "completed".
func SuccessCriteriaResponseBodyGetHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("requestID")

	if requestID == "" {
		http.Error(w, "missing requestID parameter", http.StatusBadRequest)
		return
	}

	successCriteriaResponseBodyMutex.Lock()
	defer successCriteriaResponseBodyMutex.Unlock()
	count, ok := successCriteriaResponseBodyRequests[requestID]

	if !ok {
		successCriteriaResponseBodyRequests[requestID] = 1
		WriteOKResponseAsJSON(w, PollingStatusPending)
		return
	}

	successCriteriaResponseBodyRequests[requestID] = count + 1

	switch count {
	case 1:
		WriteOKResponseAsJSON(w, PollingStatusRunning)
	default:
		WriteOKResponseAsJSON(w, PollingStatusCompleted)
	}
}
