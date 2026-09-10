package polling

import (
	"net/http"
)

// Verifies polling failureCriteria functionality based on status code.
// This endpoint always returns a 202 Accepted, which should always cause
// an error due to the failureCriteria.
func FailureCriteriaStatusCodeGetHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
