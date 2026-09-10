package delay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

type DelayResponse struct {
	Message      string `json:"message"`
	DelaySeconds int    `json:"delay_seconds"`
}

func HandleDelay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	delaySecondsStr := vars["seconds"]

	delaySeconds, err := strconv.Atoi(delaySecondsStr)
	if err != nil || delaySeconds < 0 {
		utils.HandleError(w, fmt.Errorf("invalid delay seconds: %s", delaySecondsStr))
		return
	}

	// Wait for specified duration
	time.Sleep(time.Duration(delaySeconds) * time.Second)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := DelayResponse{
		Message:      fmt.Sprintf("Delayed response after %d seconds", delaySeconds),
		DelaySeconds: delaySeconds,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HandleError(w, err)
		return
	}
}
