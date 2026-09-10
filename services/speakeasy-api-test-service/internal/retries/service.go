package retries

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
)

var (
	callCounts      = map[string]int{}
	callCountsMutex sync.Mutex
)

type retriesResponse struct {
	Retries int    `json:"retries"`
	Body    string `json:"body"`
}

// HandleRetries serves GET and POST /retries: attempts are counted per
// request-id; each fails with status-code (default 503) until the
// num-retries'th (default 3), which succeeds with {"retries": <attempt
// count>, "body": <request body>}. The echo lets callers assert a retried
// body was replayed intact (empty for GET).
func HandleRetries(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request-id")
	numRetriesStr := r.URL.Query().Get("num-retries")
	retryAfterVal := r.URL.Query().Get("retry-after-val")
	retryAfterMillisecondsVal := r.URL.Query().Get("retry-after-ms-val")

	retryAfter := 0
	if retryAfterVal != "" {
		var err error
		retryAfter, err = strconv.Atoi(retryAfterVal)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("retry-after-val must be an integer"))
			return
		}
	}

	retryAfterMilliseconds := 0
	if retryAfterMillisecondsVal != "" {
		var err error
		retryAfterMilliseconds, err = strconv.Atoi(retryAfterMillisecondsVal)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("retry-after-ms-val must be an integer"))
			return
		}
	}

	numRetries := 3
	if numRetriesStr != "" {
		var err error
		numRetries, err = strconv.Atoi(numRetriesStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("num-retries must be an integer"))
			return
		}
	}

	failStatus := http.StatusServiceUnavailable
	if v := r.URL.Query().Get("status-code"); v != "" {
		var err error
		failStatus, err = strconv.Atoi(v)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("status-code must be an integer"))
			return
		}
		if failStatus < 100 || failStatus > 999 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("status-code must be a valid HTTP status code"))
			return
		}
	}

	if requestID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("request-id is required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("failed to read request body"))
		return
	}

	callCountsMutex.Lock()
	_, ok := callCounts[requestID]
	if !ok {
		callCounts[requestID] = 0
	}
	callCounts[requestID]++
	count := callCounts[requestID]
	callCountsMutex.Unlock()

	if count < numRetries {
		if retryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		}
		if retryAfterMilliseconds > 0 {
			w.Header().Set("retry-after-ms", strconv.Itoa(retryAfterMilliseconds))
		}
		w.WriteHeader(failStatus)
		_, _ = w.Write([]byte("request failed please retry"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(retriesResponse{
		Retries: count,
		Body:    string(body),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to marshal response"))
		return
	}
	_, _ = w.Write(data)

	callCountsMutex.Lock()
	delete(callCounts, requestID)
	callCountsMutex.Unlock()
}
