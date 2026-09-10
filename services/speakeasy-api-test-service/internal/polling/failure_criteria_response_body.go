package polling

import (
	"net/http"
)

// Verifies polling failureCriteria functionality based on response body.
// This endpoint always returns a 200 OK with a status of "failed", which
// should always cause an polling error due to the failureCriteria.
func FailureCriteriaResponseBodyGetHandler(w http.ResponseWriter, r *http.Request) {
	WriteOKResponseAsJSON(w, PollingStatusFailed)
}
