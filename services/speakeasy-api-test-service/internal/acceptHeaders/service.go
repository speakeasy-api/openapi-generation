package acceptHeaders

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

func headersContains(headers []string, toCheck string) bool {
	for _, a := range headers {
		if strings.Contains(a, toCheck) {
			return true
		}
	}
	return false
}

func HandleAcceptHeaderMultiplexing(w http.ResponseWriter, r *http.Request) {
	var obj interface{}
	if headersContains(r.Header["Accept"], "application/json") {
		err := json.Unmarshal([]byte("{\"type\":\"obj1\", \"value\": \"JSON\"}"), &obj)
		if err != nil {
			utils.HandleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if err := json.NewEncoder(w).Encode(obj); err != nil {
			utils.HandleError(w, err)
		}
	} else if headersContains(r.Header["Accept"], "text/plain") {

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Success"))
	}

}
