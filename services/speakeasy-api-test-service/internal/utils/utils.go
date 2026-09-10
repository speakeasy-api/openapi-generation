package utils

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/pkg/models"
)

var authError = errors.New("invalid auth")

func HandleError(w http.ResponseWriter, err error) {
	log.Println(err)

	data, marshalErr := json.Marshal(models.ErrorResponse{
		Error: models.Error{
			Message: err.Error(),
		},
	})
	if marshalErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if errors.Is(err, authError) {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, _ = w.Write(data)
}
