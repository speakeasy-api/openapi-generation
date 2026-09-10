package polling

import (
	"net/http"
	"sync"
)

var (
	// Mapping of request identifiers to request count.
	limitCountRequests = make(map[string]int)

	// Mutex to protect access to limitCountRequests.
	limitCountMutex sync.Mutex
)

// Verifies polling limitCount functionality. Based on requestID, the
// server will ensure requests are limited by the expected count.
//
// The server will respond with a "running" status for the first 5 requests
// then respond with "failed" status beyond.
//
// This ensures that the default limitCount of 60 is not used and that the
// the 5 polling request limit is used instead.
func LimitCountGetHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("requestID")

	if requestID == "" {
		http.Error(w, "missing requestID parameter", http.StatusBadRequest)
		return
	}

	limitCountMutex.Lock()
	count := limitCountRequests[requestID]
	count += 1
	limitCountRequests[requestID] = count
	limitCountMutex.Unlock()

	switch count {
	case 1, 2, 3, 4, 5:
		WriteOKResponseAsJSON(w, PollingStatusRunning)
	default:
		WriteOKResponseAsJSON(w, PollingStatusFailed)
	}
}
