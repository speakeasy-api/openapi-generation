// Package licensetoken validates registry-signed license tokens against the embedded production JWK set.
package licensetoken

import (
	"bytes"
	"context"
	"crypto/ed25519"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
)

const (
	issuer                 = "speakeasy-registry"
	audience               = "speakeasy-generation"
	licenseTypeCommercial  = "commercial"
	supportedSchemaVersion = 1

	clockSkew = time.Minute
)

var ErrInvalidLicenseToken = errors.New("invalid license token")

var ErrConflictingLicenseElection = errors.New("a license token conflicts with an AGPL election: supply one or the other")

var targetPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Production public keys only; never add dev, staging, or test keys.
//
//go:embed embedded_jwks.json
var embeddedProductionJWKS []byte

type validator interface {
	validate(token []byte) (*claims, error)
}

type jwksValidator struct {
	jwks []byte
	now  func() time.Time
}

func (v jwksValidator) validate(token []byte) (*claims, error) {
	return validateToken(token, v.jwks, v.now())
}

var productionValidator validator = jwksValidator{jwks: embeddedProductionJWKS, now: time.Now}

type tokenContextKey struct{}

type TokenInfo struct {
	WorkspaceID   string
	WorkspaceSlug string
	OrgSlug       string
	Tier          string
	Targets       []string
	IssuedAt      time.Time
	ExpiresAt     time.Time
	Message       string
}

func (i TokenInfo) Covers(target string) bool {
	return slices.Contains(i.Targets, target) || slices.Contains(i.Targets, "*")
}

// Inspect validates token exactly as generation does and returns its summary without touching a context.
func Inspect(token []byte) (TokenInfo, error) {
	verifiedClaims, _, err := verifiedInfo(token)
	if err != nil {
		return TokenInfo{}, err
	}
	return TokenInfo{
		WorkspaceID:   verifiedClaims.WorkspaceID,
		WorkspaceSlug: verifiedClaims.WorkspaceSlug,
		OrgSlug:       verifiedClaims.OrgSlug,
		Tier:          verifiedClaims.Tier,
		Targets:       append([]string(nil), verifiedClaims.Targets...),
		IssuedAt:      verifiedClaims.IssuedAt,
		ExpiresAt:     verifiedClaims.ExpiresAt,
		Message:       verifiedClaims.Message,
	}, nil
}

// WithToken attaches a raw license token; the generator validates it during initialization.
func WithToken(ctx context.Context, token []byte) context.Context {
	trimmed := bytes.TrimSpace(token)
	if len(trimmed) == 0 {
		return ctx
	}
	return context.WithValue(ctx, tokenContextKey{}, append([]byte(nil), trimmed...))
}

type Verification struct {
	Targets []string
}

func (v Verification) Covers(target string) bool {
	return slices.Contains(v.Targets, target) || slices.Contains(v.Targets, "*")
}

// ResolveGenerationAccess validates an attached token into commercial access state and returns its target coverage.
func ResolveGenerationAccess(ctx context.Context) (context.Context, *Verification, error) {
	token, ok := ctx.Value(tokenContextKey{}).([]byte)
	if !ok {
		return ctx, nil, nil
	}
	if state, ok := generationaccess.StateFromContext(ctx); ok && state.GeneratedLicense() == generationaccess.GeneratedLicenseAGPL {
		return ctx, nil, ErrConflictingLicenseElection
	}

	verifiedClaims, info, err := verifiedInfo(token)
	if err != nil {
		return ctx, nil, err
	}

	ctx, err = generationaccess.WithAuthenticated(ctx, info)
	if err != nil {
		return ctx, nil, fmt.Errorf("%w: %w", ErrInvalidLicenseToken, err)
	}
	return ctx, &Verification{Targets: append([]string(nil), verifiedClaims.Targets...)}, nil
}

func verifiedInfo(token []byte) (*claims, generationaccess.AuthenticatedInfo, error) {
	verifiedClaims, err := productionValidator.validate(token)
	if err != nil {
		return nil, generationaccess.AuthenticatedInfo{}, err
	}

	info, err := verifiedClaims.authenticatedInfo()
	if err != nil {
		return nil, generationaccess.AuthenticatedInfo{}, err
	}

	if _, err := generationaccess.WithAuthenticated(context.Background(), info); err != nil {
		return nil, generationaccess.AuthenticatedInfo{}, fmt.Errorf("%w: %w", ErrInvalidLicenseToken, err)
	}
	return verifiedClaims, info, nil
}

type claims struct {
	Type               string    `json:"type"`
	Version            int       `json:"version"`
	WorkspaceID        string    `json:"workspace_id"`
	WorkspaceSlug      string    `json:"workspace_slug"`
	OrgID              string    `json:"org_id"`
	OrgSlug            string    `json:"org_slug"`
	Tier               string    `json:"tier"`
	Targets            []string  `json:"targets"`
	Features           []string  `json:"features"`
	AddOns             []string  `json:"add_ons"`
	WorkspaceCreatedAt time.Time `json:"workspace_created_at"`
	TelemetryDisabled  bool      `json:"telemetry_disabled"`
	Message            string    `json:"message,omitempty"`

	ExpiresAt time.Time `json:"-"`
	IssuedAt  time.Time `json:"-"`
}

func (c *claims) authenticatedInfo() (generationaccess.AuthenticatedInfo, error) {
	var accountType shared.AccountType
	switch c.Tier {
	case string(shared.AccountTypeScaleUp):
		accountType = shared.AccountTypeScaleUp
	case string(shared.AccountTypeBusiness):
		accountType = shared.AccountTypeBusiness
	case string(shared.AccountTypeEnterprise):
		accountType = shared.AccountTypeEnterprise
	case string(shared.AccountTypeFree):
		accountType = shared.AccountTypeFree
	case "oss":
		accountType = shared.AccountTypeEnterprise
	default:
		return generationaccess.AuthenticatedInfo{}, fmt.Errorf("%w: unsupported tier %q", ErrInvalidLicenseToken, c.Tier)
	}

	if c.WorkspaceCreatedAt.IsZero() {
		return generationaccess.AuthenticatedInfo{}, fmt.Errorf("%w: missing workspace_created_at", ErrInvalidLicenseToken)
	}

	var addOns []shared.BillingAddOn
	for _, addOn := range c.AddOns {
		switch candidate := shared.BillingAddOn(addOn); candidate {
		case shared.BillingAddOnWebhooks,
			shared.BillingAddOnSDKTesting,
			shared.BillingAddOnCustomCodeRegions,
			shared.BillingAddOnSnippetAi:
			addOns = append(addOns, candidate)
		default:
		}
	}

	createdAt := c.WorkspaceCreatedAt

	return generationaccess.AuthenticatedInfo{
		AccountType:        accountType,
		BillingAddOns:      addOns,
		WorkspaceCreatedAt: &createdAt,
		WorkspaceID:        c.WorkspaceID,
		TelemetryDisabled:  c.TelemetryDisabled,
		GeneratedLicense:   generationaccess.GeneratedLicenseCommercial,
	}, nil
}

type joseHeader struct {
	Alg  string   `json:"alg"`
	Kid  string   `json:"kid"`
	Crit []string `json:"crit"`
}

type audienceClaim []string

func (a *audienceClaim) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = audienceClaim{single}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*a = audienceClaim(many)
	return nil
}

type tokenPayload struct {
	Issuer    string        `json:"iss"`
	Audience  audienceClaim `json:"aud"`
	Subject   string        `json:"sub"`
	ExpiresAt *int64        `json:"exp"`
	IssuedAt  *int64        `json:"iat"`
	NotBefore *int64        `json:"nbf"`
	License   *claims       `json:"license"`
}

func validateToken(token []byte, jwksJSON []byte, now time.Time) (*claims, error) {
	token = bytes.TrimSpace(token)
	headerPart, payloadPart, signaturePart, err := splitCompact(token)
	if err != nil {
		return nil, err
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(string(headerPart))
	if err != nil {
		return nil, fmt.Errorf("%w: malformed header encoding", ErrInvalidLicenseToken)
	}
	var header joseHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("%w: malformed header", ErrInvalidLicenseToken)
	}
	if header.Alg != "EdDSA" {
		return nil, fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidLicenseToken, header.Alg)
	}
	if header.Kid == "" {
		return nil, fmt.Errorf("%w: missing kid", ErrInvalidLicenseToken)
	}
	if len(header.Crit) > 0 {
		return nil, fmt.Errorf("%w: unsupported crit header", ErrInvalidLicenseToken)
	}

	publicKey, err := resolveKey(jwksJSON, header.Kid)
	if err != nil {
		return nil, err
	}

	signature, err := base64.RawURLEncoding.DecodeString(string(signaturePart))
	if err != nil || len(signature) != ed25519.SignatureSize {
		return nil, fmt.Errorf("%w: malformed signature", ErrInvalidLicenseToken)
	}
	signingInput := token[:len(headerPart)+1+len(payloadPart)]
	if !ed25519.Verify(publicKey, signingInput, signature) {
		return nil, fmt.Errorf("%w: signature verification failed", ErrInvalidLicenseToken)
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(string(payloadPart))
	if err != nil {
		return nil, fmt.Errorf("%w: malformed payload encoding", ErrInvalidLicenseToken)
	}
	var payload tokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrInvalidLicenseToken)
	}

	if err := validatePayload(payload, now); err != nil {
		return nil, err
	}

	verifiedClaims := *payload.License
	verifiedClaims.ExpiresAt = time.Unix(*payload.ExpiresAt, 0)
	verifiedClaims.IssuedAt = time.Unix(*payload.IssuedAt, 0)

	return &verifiedClaims, nil
}

func splitCompact(token []byte) (header, payload, signature []byte, err error) {
	parts := bytes.Split(token, []byte("."))
	if len(parts) != 3 {
		return nil, nil, nil, fmt.Errorf("%w: not a compact JWS", ErrInvalidLicenseToken)
	}
	for _, part := range parts {
		if len(part) == 0 {
			return nil, nil, nil, fmt.Errorf("%w: not a compact JWS", ErrInvalidLicenseToken)
		}
	}
	return parts[0], parts[1], parts[2], nil
}

type jsonWebKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	X   string `json:"x"`
}

type jsonWebKeySet struct {
	Keys []jsonWebKey `json:"keys"`
}

func resolveKey(jwksJSON []byte, kid string) (ed25519.PublicKey, error) {
	var set jsonWebKeySet
	if err := json.Unmarshal(jwksJSON, &set); err != nil {
		return nil, fmt.Errorf("%w: malformed JWK set", ErrInvalidLicenseToken)
	}

	for _, key := range set.Keys {
		if key.Kid != kid {
			continue
		}
		if key.Kty != "OKP" || key.Crv != "Ed25519" {
			return nil, fmt.Errorf("%w: key %q is not an Ed25519 key", ErrInvalidLicenseToken, kid)
		}
		raw, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil || len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("%w: key %q has malformed key material", ErrInvalidLicenseToken, kid)
		}
		return ed25519.PublicKey(raw), nil
	}

	return nil, fmt.Errorf("%w: unknown kid %q", ErrInvalidLicenseToken, kid)
}

func validatePayload(payload tokenPayload, now time.Time) error {
	if payload.Issuer != issuer {
		return fmt.Errorf("%w: unexpected issuer %q", ErrInvalidLicenseToken, payload.Issuer)
	}

	audienceMatched := false
	for _, aud := range payload.Audience {
		if aud == audience {
			audienceMatched = true
			break
		}
	}
	if !audienceMatched {
		return fmt.Errorf("%w: audience does not include %q", ErrInvalidLicenseToken, audience)
	}

	if payload.ExpiresAt == nil || payload.IssuedAt == nil || payload.NotBefore == nil {
		return fmt.Errorf("%w: exp, iat, and nbf are required", ErrInvalidLicenseToken)
	}
	if now.After(time.Unix(*payload.ExpiresAt, 0).Add(clockSkew)) {
		return fmt.Errorf("%w: token expired", ErrInvalidLicenseToken)
	}
	if now.Before(time.Unix(*payload.NotBefore, 0).Add(-clockSkew)) {
		return fmt.Errorf("%w: token not yet valid", ErrInvalidLicenseToken)
	}
	if now.Before(time.Unix(*payload.IssuedAt, 0).Add(-clockSkew)) {
		return fmt.Errorf("%w: token issued in the future", ErrInvalidLicenseToken)
	}

	if payload.License == nil {
		return fmt.Errorf("%w: missing license claim", ErrInvalidLicenseToken)
	}
	if payload.License.Type != licenseTypeCommercial {
		return fmt.Errorf("%w: license type %q is not commercial proof", ErrInvalidLicenseToken, payload.License.Type)
	}
	if payload.License.Version != supportedSchemaVersion {
		return fmt.Errorf("%w: unsupported license schema version %d", ErrInvalidLicenseToken, payload.License.Version)
	}
	if payload.License.Targets == nil {
		return fmt.Errorf("%w: missing targets", ErrInvalidLicenseToken)
	}
	if len(payload.License.Targets) == 0 {
		return fmt.Errorf("%w: targets must not be empty", ErrInvalidLicenseToken)
	}
	wildcard := false
	for _, target := range payload.License.Targets {
		if target == "*" {
			wildcard = true
			continue
		}
		if !targetPattern.MatchString(target) {
			return fmt.Errorf("%w: invalid target %q", ErrInvalidLicenseToken, target)
		}
	}
	if wildcard && len(payload.License.Targets) != 1 {
		return fmt.Errorf("%w: wildcard target cannot be combined with named targets", ErrInvalidLicenseToken)
	}
	return nil
}
