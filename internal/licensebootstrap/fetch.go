package licensebootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var ErrUnauthorized = errors.New("the Speakeasy API key was rejected; run `speakeasy auth login` again")

var ErrNotEntitled = errors.New("workspace has no commercial entitlement; no license token was issued")

var ErrAccessTokenMissing = errors.New("the platform allowed generation but issued no token")

var ErrInsecureServerURL = errors.New("the Speakeasy server URL must use https (plain http is allowed for loopback addresses only)")

func requireSecureURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "https":
		return nil
	case "http":
		host := parsed.Hostname()
		if host == "localhost" {
			return nil
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrInsecureServerURL, raw)
}

type ValidateResponse struct {
	WorkspaceID   string  `json:"workspace_id"`
	WorkspaceSlug string  `json:"workspace_slug"`
	OrgSlug       string  `json:"org_slug"`
	AccountType   string  `json:"account_type_v2"`
	LicenseJWT    *string `json:"license_jwt"`
}

type AccessResponse struct {
	GenerationAllowed bool    `json:"generation_allowed"`
	Message           string  `json:"message"`
	Level             string  `json:"level"`
	LicenseJWT        *string `json:"license_jwt"`
}

func FetchLicenseToken(ctx context.Context, client *http.Client, serverURL, apiKey string) (string, ValidateResponse, error) {
	var details ValidateResponse

	body, err := fetchResponse(ctx, client, serverURL, apiKey, "/v1/auth/validate", "validate API key")
	if err != nil {
		return "", details, err
	}

	if err := json.Unmarshal(body, &details); err != nil {
		return "", details, fmt.Errorf("decode validate response: %w", err)
	}
	if details.LicenseJWT == nil || strings.TrimSpace(*details.LicenseJWT) == "" {
		return "", details, ErrNotEntitled
	}
	return strings.TrimSpace(*details.LicenseJWT), details, nil
}

func FetchAccessLicenseToken(ctx context.Context, client *http.Client, serverURL, apiKey, target string) (string, AccessResponse, error) {
	var details AccessResponse
	query := url.Values{"targetType": []string{target}}
	body, err := fetchResponse(ctx, client, serverURL, apiKey, "/v1/workspace/access?"+query.Encode(), "fetch workspace access")
	if err != nil {
		return "", details, err
	}
	if err := json.Unmarshal(body, &details); err != nil {
		return "", details, fmt.Errorf("decode workspace access response: %w", err)
	}
	if !details.GenerationAllowed && details.Level == "blocked" {
		return "", details, ErrNotEntitled
	}
	if details.LicenseJWT == nil || strings.TrimSpace(*details.LicenseJWT) == "" {
		return "", details, ErrAccessTokenMissing
	}
	return strings.TrimSpace(*details.LicenseJWT), details, nil
}

func fetchResponse(ctx context.Context, client *http.Client, serverURL, apiKey, endpoint, operation string) ([]byte, error) {
	if err := requireSecureURL(serverURL); err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(serverURL, "/")+endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("x-api-key", apiKey)
	request.Header.Set("Accept", "application/json")

	noRedirect := *client
	noRedirect.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	response, err := noRedirect.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", request.URL.Redacted(), err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	switch {
	case response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden:
		return nil, ErrUnauthorized
	case response.StatusCode >= 300 && response.StatusCode <= 399:
		return nil, fmt.Errorf("%s: %s answered with a redirect (status %d); redirects are not followed so the API key is never forwarded", operation, serverURL, response.StatusCode)
	case response.StatusCode < 200 || response.StatusCode > 299:
		return nil, fmt.Errorf("%s: unexpected status %d from %s", operation, response.StatusCode, serverURL)
	}
	return body, nil
}
