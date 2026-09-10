package polling

import (
	"net/http"
	"strconv"
	"sync"
)

var (
	// Mapping of request identifiers to status codes that have sent requests
	// to this endpoint.
	successCriteriaStatusCodeRequests = make(map[string]int)

	// Mutex to protect access to successCriteriaStatusCodeRequests.
	successCriteriaStatusCodeMutex sync.Mutex
)

func SuccessCriteriaStatusCodeGetHandler(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("requestID")

	if requestID == "" {
		http.Error(w, "missing requestID parameter", http.StatusBadRequest)
		return
	}

	requestStatusCodeStr := r.URL.Query().Get("statusCode")
	requestStatusCode, err := strconv.Atoi(requestStatusCodeStr)

	if err != nil {
		http.Error(w, "invalid statusCode parameter: "+requestStatusCodeStr, http.StatusBadRequest)
		return
	}

	successCriteriaStatusCodeMutex.Lock()
	defer successCriteriaStatusCodeMutex.Unlock()
	statusCode, ok := successCriteriaStatusCodeRequests[requestID]

	if !ok {
		successCriteriaStatusCodeRequests[requestID] = requestStatusCode
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}

	WriteOKResponseAsJSON(w, PollingStatusCompleted)
}
