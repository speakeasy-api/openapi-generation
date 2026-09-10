package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/pkg/models"
)

var authError = errors.New("invalid auth")

func checkAuth(req models.AuthRequest, r *http.Request) error {
	if req.BasicAuth != nil {
		if err := checkBasicAuth(*req.BasicAuth, r); err != nil {
			return err
		}
	}

	for _, headerAuth := range req.HeaderAuth {
		if err := checkHeaderAuth(headerAuth, r); err != nil {
			return err
		}
	}

	return nil
}

func checkBasicAuth(basicAuth models.BasicAuth, r *http.Request) error {
	basicAuthHeader := r.Header.Get("Authorization")
	if basicAuthHeader == "" {
		return fmt.Errorf("missing Authorization header for Basic Auth: %w", authError)
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		return fmt.Errorf("invalid Authorization header for Basic Auth: %w", authError)
	}

	if username != basicAuth.Username {
		return fmt.Errorf("invalid username for Basic Auth: %w", authError)
	}

	if password != basicAuth.Password {
		return fmt.Errorf("invalid password for Basic Auth: %w", authError)
	}

	return nil
}

func checkHeaderAuth(headerAuth models.HeaderAuth, r *http.Request) error {
	headerValue := r.Header.Get(headerAuth.HeaderName)
	if headerValue == "" {
		return fmt.Errorf("missing %s header: %w", headerAuth.HeaderName, authError)
	}

	if headerValue != headerAuth.ExpectedValue {
		return fmt.Errorf("invalid %s header: %w", headerAuth.HeaderName, authError)
	}

	return nil
}
