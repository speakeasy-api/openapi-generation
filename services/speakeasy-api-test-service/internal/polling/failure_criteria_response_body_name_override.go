package polling

import (
	"net/http"
)

// Verifies polling failureCriteria functionality based on response body when
// the polling condition references a property with x-speakeasy-name-override.
// This endpoint always returns a 200 OK with a status of "failed", which
// should always cause a polling error due to the failureCriteria.
func FailureCriteriaResponseBodyNameOverrideGetHandler(w http.ResponseWriter, r *http.Request) {
	WriteOKResponseAsJSON(w, PollingStatusFailed)
}
