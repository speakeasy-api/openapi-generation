package polling

import (
	"net/http"
	"sync"
	"time"
)

var (
	// Mapping of request identifiers to initial request time.
	delaySecondsRequests = make(map[string]time.Time)

	// Mutex to protect access to delaySecondsRequests.
	delaySecondsMutex sync.Mutex
)

// Verifies polling delaySeconds functionality. An initial (non-polling)
// request should be sent to the endpoint before starting polling requests,
// which will set timestamp on the server based on requestID.
//
// The server will respond with a 204 No Content:
//   - For <3 seconds after the initial request
//   - For >7 seconds after the initial request
//
// This ensures that the default delaySeconds of 1 second is not used and
// that the 5 second polling delay is used instead.
func DelaySecondsGetHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("requestID")

	if requestID == "" {
		http.Error(w, "missing requestID parameter", http.StatusBadRequest)
		return
	}

	delaySecondsMutex.Lock()
	defer delaySecondsMutex.Unlock()
	initialRequestTime, ok := delaySecondsRequests[requestID]

	if !ok {
		delaySecondsRequests[requestID] = time.Now()
		w.WriteHeader(http.StatusCreated)
		return
	}

	switch {
	case time.Since(initialRequestTime) < 3*time.Second:
		w.WriteHeader(http.StatusNoContent)
	case time.Since(initialRequestTime) > 7*time.Second:
		w.WriteHeader(http.StatusNoContent)
	default:
		WriteOKResponseAsJSON(w, PollingStatusCompleted)
	}
}
