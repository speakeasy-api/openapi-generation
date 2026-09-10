package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/auth"
)

type oauth2CtxKey string

var oauth2ClaimsKey oauth2CtxKey = "oauth2Claims"

func OAuth2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authhdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(authhdr, "Bearer ") {
			auth.SendOAuth2Error(w, auth.ErrCodeInvalidRequest, "missing bearer token")
			return
		}

		claims, err := auth.ParseToken(authhdr[7:])
		if err != nil {
			auth.SendOAuth2Error(w, auth.ErrCodeInvalidRequest, err.Error())
			return
		}

		if auth.IsTokenExpired(claims) {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "token has expired"}`, http.StatusUnauthorized)
			return
		}

		if claims["grantType"].(string) == "refresh_token" {
			auth.SendOAuth2Error(w, auth.ErrCodeInvalidRequest, "cannot use refresh token as access token")
			return
		}

		w.Header().Set("x-oauth2", "pass")
		ctx = context.WithValue(ctx, oauth2ClaimsKey, claims)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func OAuth2Scopes(r *http.Request) (Scopes, bool) {
	v := r.Context().Value(oauth2ClaimsKey)
	if v == nil {
		return nil, false
	}

	claims, ok := v.(jwt.MapClaims)
	if !ok {
		return nil, false
	}

	rawscopes, ok := claims["scope"].(string)
	if !ok {
		return nil, false
	}

	elements := strings.Split(rawscopes, " ")
	var scopes Scopes
	for _, scope := range elements {
		scopes = append(scopes, scope)
	}

	return scopes, true
}

type Scopes []string

func (s Scopes) Has(requiredScopes []string) bool {
	if len(s) < len(requiredScopes) {
		return false
	}

	for _, required := range requiredScopes {
		if !slices.Contains(s, required) {
			return false
		}
	}

	return true
}
