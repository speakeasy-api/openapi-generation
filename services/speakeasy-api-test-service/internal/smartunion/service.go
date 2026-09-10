package smartunion

import (
	"encoding/json"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

// HandleNestedUnion returns a response with unknown enum values to test
// that nested unions only count the winning inner option's unrecognized values.
//
// Test structure:
// OuterUnion:
//   - Option A: { data: InnerUnion } where InnerUnion has 3 options with open enums
//   - Option B: { data: { kind: OpenEnum, name: OpenEnum } } has 2 open enum fields
//
// InnerUnion (inside Option A):
//   - { kind: OpenEnum["cat"], name: string } - 2 fields, wins for payload with name
//   - { kind: OpenEnum["dog"] } - 1 field
//   - { kind: OpenEnum["bird"] } - 1 field
//
// Response payload: { data: { kind: "unknown", name: "also_unknown" } }
//
// Expected: Option A should win because:
//   - Inner union picks first variant (has name field), counts 1 unrecognized (kind="unknown")
//   - Option B would count 2 unrecognized (both kind and name are unknown enum values)
//   - With correct counting: A=1 < B=2, so A wins
//   - With buggy counting: A=3 (accumulated from all inner options) > B=2, so B would wrongly win
func HandleNestedUnion(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"json": map[string]interface{}{
			"data": map[string]interface{}{
				"kind": "unknown",
				"name": "also_unknown",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HandleError(w, err)
	}
}
