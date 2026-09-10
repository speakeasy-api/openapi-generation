package reflect

import (
	"io"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"
)

func HandleReflect(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.HandleError(w, err)
		return
	}

	contentType := r.Header.Get("Content-Type")

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}

	if _, err := w.Write(body); err != nil {
		utils.HandleError(w, err)
	}
}
