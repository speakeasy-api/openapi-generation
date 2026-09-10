package polling

import (
	"net/http"
	"sync"
	"time"
)

var (
	// Mapping of request identifiers to last request time.
	intervalSecondsRequests = make(map[string]time.Time)

	// Mutex to protect access to intervalSecondsRequests.
	intervalSecondsMutex sync.Mutex
)

// Verifies polling intervalSeconds functionality. Based on requestID, the
// server will ensure requests are delayed by the expected interval.
//
// The server will respond with a 204 No Content:
//   - When interval is <3 seconds after the last request
//   - When interval is >7 seconds after the last request
//
// This ensures that the default intervalSeconds of 1 second is not used
// and that the 5 second polling interval is used instead.
func IntervalSecondsGetHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("requestID")

	if requestID == "" {
		http.Error(w, "missing requestID parameter", http.StatusBadRequest)
		return
	}

	intervalSecondsMutex.Lock()
	lastRequestTime, ok := intervalSecondsRequests[requestID]

	intervalSecondsRequests[requestID] = time.Now()
	intervalSecondsMutex.Unlock()

	if !ok {
		w.WriteHeader(http.StatusCreated)
		return
	}

	switch {
	case time.Since(lastRequestTime) < 3*time.Second:
		w.WriteHeader(http.StatusNoContent)
	case time.Since(lastRequestTime) > 7*time.Second:
		w.WriteHeader(http.StatusNoContent)
	default:
		WriteOKResponseAsJSON(w, PollingStatusCompleted)
	}
}
