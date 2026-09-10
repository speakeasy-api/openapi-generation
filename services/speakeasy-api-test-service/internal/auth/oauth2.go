package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
)

const (
	oauth2JWTSigningSecret = "fancy-jwt-signing-secret"
)

type OAuth2ErrorCode string

const (
	ErrCodeInvalidRequest       OAuth2ErrorCode = "invalid_request"
	ErrCodeInvalidClient        OAuth2ErrorCode = "invalid_client"
	ErrCodeInvalidGrant         OAuth2ErrorCode = "invalid_grant"
	ErrCodeUnauthorizedClient   OAuth2ErrorCode = "unauthorized_client"
	ErrCodeUnsupportedGrantType OAuth2ErrorCode = "unsupported_grant_type"
	ErrCodeInvalidScope         OAuth2ErrorCode = "invalid_scope"
)

func (e OAuth2ErrorCode) Error() string {
	return string(e)
}

type oauth2Error struct {
	Code    OAuth2ErrorCode `json:"error"`
	Message string          `json:"error_description,omitempty"`
}

func SendOAuth2Error(w http.ResponseWriter, code OAuth2ErrorCode, description string) {
	payload := oauth2Error{
		Code:    code,
		Message: description,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := enc.Encode(payload); err != nil {
		log.Println(err)

		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

type TokenForm struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Code         string `json:"code"`
	RedirectURI  string `json:"redirect_uri"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func HandleOAuth2InspectToken(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	w.Header().Set("Content-Type", "application/json")

	authz := r.Header.Get("Authorization")
	if authz == "" {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}
	if !strings.HasPrefix(authz, "Bearer ") {
		http.Error(w, `{"error": "invalid authorization"}`, http.StatusBadRequest)
		return
	}

	token := authz[len("Bearer "):]
	claims, err := ParseToken(token)
	if err != nil {
		http.Error(w, `{"error": "invalid token"}`, http.StatusBadRequest)
		return
	}

	updatedExpiry := GetTokenExpiry(claims)

	claims["exp"] = float64(updatedExpiry.Unix())

	if err := enc.Encode(claims); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

func HandleOAuth2(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		SendOAuth2Error(w, ErrCodeInvalidRequest, "cannot parse url-encoded request body")
		return
	}

	form := TokenForm{
		GrantType:    r.PostForm.Get("grant_type"),
		ClientID:     r.PostForm.Get("client_id"),
		ClientSecret: r.PostForm.Get("client_secret"),
		Username:     r.PostForm.Get("username"),
		Password:     r.PostForm.Get("password"),
		Code:         r.PostForm.Get("code"),
		RedirectURI:  r.PostForm.Get("redirect_uri"),
		RefreshToken: r.PostForm.Get("refresh_token"),
		Scope:        r.PostForm.Get("scope"),
	}

	switch form.GrantType {
	case "client_credentials":
		if form.ClientID == "" || form.ClientSecret == "" {
			SendOAuth2Error(w, ErrCodeInvalidRequest, "missing client credentials")
			return
		}

		if !validateClientCredentials(r, form) {
			SendOAuth2Error(w, ErrCodeInvalidClient, "invalid client id or secret")
			return
		}
	case "password":
		if form.ClientID == "" || form.ClientSecret == "" || form.Username == "" || form.Password == "" {
			SendOAuth2Error(w, ErrCodeInvalidRequest, "missing resource owner password credentials")
			return
		}
		if !validateClientCredentials(r, form) {
			SendOAuth2Error(w, ErrCodeInvalidClient, "invalid client id or secret")
			return
		}
		if form.Username != "testuser" || form.Password != "testpassword" {
			SendOAuth2Error(w, ErrCodeInvalidGrant, "invalid username or password")
			return
		}
	case "authorization_code":
		if form.ClientID == "" || form.Code == "" {
			SendOAuth2Error(w, ErrCodeInvalidRequest, "missing authorization code credentials")
			return
		}
		if form.ClientID != "beezy" {
			SendOAuth2Error(w, ErrCodeInvalidClient, "invalid client id")
			return
		}
		if form.Code != "secret-auth-code" {
			SendOAuth2Error(w, ErrCodeInvalidGrant, "")
			return
		}
	case "refresh_token":
		if form.RefreshToken == "" {
			SendOAuth2Error(w, ErrCodeInvalidRequest, "missing refresh token")
			return
		}

		if !validateClientCredentials(r, form) {
			SendOAuth2Error(w, ErrCodeInvalidGrant, "invalid client id or secret")
			return
		}

		rt, err := ParseToken(form.RefreshToken)
		if err != nil {
			SendOAuth2Error(w, ErrCodeInvalidRequest, "invalid refresh token")
			return
		}
		if IsTokenExpired(rt) {
			SendOAuth2Error(w, ErrCodeInvalidGrant, "refresh token has expired")
			return
		}
	default:
		SendOAuth2Error(w, ErrCodeUnsupportedGrantType, "unsupported grant type")
		return
	}

	now := time.Now()
	expires := now.Add(time.Hour)
	forcedExpiry := r.Header.Get("x-oauth2-expire-at")
	if exp, err := time.Parse(time.RFC3339, forcedExpiry); err == nil {
		expires = exp
	}

	accessTokenID := gofakeit.UUID()
	accessTokenClaims := jwt.MapClaims{
		"exp":       float64(expires.Unix()),
		"id":        accessTokenID,
		"grantType": form.GrantType,
		"clientID":  form.ClientID,
		"username":  form.Username,
		"scope":     form.Scope,
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	signedAccessToken, err := accessToken.SignedString([]byte(oauth2JWTSigningSecret))
	if err != nil {
		log.Println(err)
		SendOAuth2Error(w, ErrCodeInvalidRequest, err.Error())
		return
	}

	refreshTokenClaims := jwt.MapClaims{
		"exp":       float64(now.Add(60 * 24 * time.Hour).Unix()),
		"id":        gofakeit.UUID(),
		"sub":       accessTokenID,
		"grantType": "refresh_token",
		"clientID":  form.ClientID,
		"username":  form.Username,
		"scope":     form.Scope,
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(oauth2JWTSigningSecret))
	if err != nil {
		log.Println(err)
		http.Error(w, `{"error": "failed to sign refresh token"}`, http.StatusInternalServerError)
		return
	}

	res := OAuth2TokenResponse{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    max(int(expires.Sub(now).Seconds()), 0),
	}

	RegisterToken(accessTokenClaims)
	RegisterToken(refreshTokenClaims)

	if err := enc.Encode(res); err != nil {
		http.Error(w, `{"error": "failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

func validateClientCredentials(r *http.Request, form TokenForm) bool {
	clientID := form.ClientID
	clientSecret := form.ClientSecret
	if clientID == "" && clientSecret == "" {
		clientID, clientSecret, _ = r.BasicAuth()
	}

	return clientID == "beezy" && clientSecret == "super-secret"
}

var tokenDB sync.Map
var tokenDBLastAccess atomic.Value

func RegisterToken(tokenClaims jwt.MapClaims) {
	tokenDBLastAccess.Store(time.Now())

	tokenID := tokenClaims["id"].(string)
	expiry, err := tokenClaims.GetExpirationTime()
	if err != nil {
		panic(err)
	}
	tokenDB.Store(tokenID, expiry.Time)
}

func RefreshToken(refreshClaims jwt.MapClaims) {
	tokenDBLastAccess.Store(time.Now())

	tokenID := refreshClaims["sub"].(string)
	expiry := time.Now().Add(time.Hour)
	tokenDB.Store(tokenID, expiry)
}

func GetTokenExpiry(tokenClaims jwt.MapClaims) time.Time {
	tokenDBLastAccess.Store(time.Now())

	tokenID := tokenClaims["id"].(string)

	exp, found := tokenDB.Load(tokenID)
	if found {
		return exp.(time.Time)
	}

	expiryClaim, err := tokenClaims.GetExpirationTime()
	if err != nil {
		panic(err)
	}

	return expiryClaim.Time
}

func IsTokenExpired(tokenClaims jwt.MapClaims) bool {
	tokenDBLastAccess.Store(time.Now())

	tokenID := tokenClaims["id"].(string)
	expiryClaim, err := tokenClaims.GetExpirationTime()
	if err != nil {
		panic(err)
	}

	exp, found := tokenDB.Load(tokenID)
	if !found {
		RegisterToken(tokenClaims)
		exp = expiryClaim.Time
	}

	expiry := exp.(time.Time)

	return expiry.Before(time.Now())
}

func StartTokenDBCompaction(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lastAccess, ok := tokenDBLastAccess.Load().(time.Time)
			now := time.Now()
			if !ok {
				lastAccess = now
			}

			delta := now.Sub(lastAccess)
			if delta > 5*time.Minute {
				tokenDB.Clear()
			}
		}
	}
}

func ParseToken(encodedToken string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{"HS256"}))
	token, err := parser.Parse(encodedToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(oauth2JWTSigningSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid access token string")
	}

	return claims, nil
}
