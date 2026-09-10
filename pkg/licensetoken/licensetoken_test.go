package licensetoken

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/speakeasy-client-sdk-go/v3/pkg/models/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var goldenClock = time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

func readGolden(t *testing.T, name string) (token []byte, jwks []byte) {
	t.Helper()
	token, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	jwks, err = os.ReadFile("testdata/golden-jwks.json")
	require.NoError(t, err)
	return token, jwks
}

func withValidator(t *testing.T, jwks []byte, now time.Time) {
	t.Helper()
	previous := productionValidator
	productionValidator = jwksValidator{jwks: jwks, now: func() time.Time { return now }}
	t.Cleanup(func() { productionValidator = previous })
}

func TestValidateToken_GoldenToken(t *testing.T) {
	token, jwks := readGolden(t, "golden-token.jwt")

	claims, err := validateToken(token, jwks, goldenClock)
	require.NoError(t, err)

	assert.Equal(t, licenseTypeCommercial, claims.Type)
	assert.Equal(t, supportedSchemaVersion, claims.Version)
	assert.Equal(t, "golden-workspace", claims.WorkspaceID)
	assert.Equal(t, "golden-ws", claims.WorkspaceSlug)
	assert.Equal(t, "golden-org", claims.OrgID)
	assert.Equal(t, "golden", claims.OrgSlug)
	assert.Equal(t, "business", claims.Tier)
	assert.Equal(t, []string{"go"}, claims.Targets)
	assert.Equal(t, []string{"sdk_testing"}, claims.AddOns)
	assert.Equal(t, time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC), claims.WorkspaceCreatedAt.UTC())
	assert.True(t, claims.TelemetryDisabled)
	assert.Equal(t, goldenClock.Add(30*24*time.Hour).Unix(), claims.ExpiresAt.Unix())
}

func TestValidateToken_GoldenFreeToken(t *testing.T) {
	token, jwks := readGolden(t, "golden-token-free.jwt")
	withValidator(t, jwks, goldenClock)

	verifiedClaims, err := validateToken(token, jwks, goldenClock)
	require.NoError(t, err)
	assert.Equal(t, "free", verifiedClaims.Tier)
	assert.Equal(t, []string{"typescript"}, verifiedClaims.Targets)

	authenticated, err := verifiedClaims.authenticatedInfo()
	require.NoError(t, err)
	assert.Equal(t, shared.AccountTypeFree, authenticated.AccountType)

	info, err := Inspect(token)
	require.NoError(t, err)
	assert.Equal(t, []string{"typescript"}, info.Targets)
	assert.True(t, info.Covers("typescript"))
	assert.False(t, info.Covers("go"))
}

type testMinter struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
	kid  string
}

func newTestMinter(t *testing.T) *testMinter {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sum := sha256.Sum256(pub)
	return &testMinter{priv: priv, pub: pub, kid: base64.RawURLEncoding.EncodeToString(sum[:])}
}

func (m *testMinter) jwks(t *testing.T) []byte {
	t.Helper()
	set := map[string]any{"keys": []map[string]any{{
		"kty": "OKP",
		"crv": "Ed25519",
		"alg": "EdDSA",
		"use": "sig",
		"kid": m.kid,
		"x":   base64.RawURLEncoding.EncodeToString(m.pub),
	}}}
	out, err := json.Marshal(set)
	require.NoError(t, err)
	return out
}

func defaultPayload(now time.Time) map[string]any {
	return map[string]any{
		"iss": issuer,
		"aud": []string{audience},
		"sub": "ws_1",
		"jti": "test-jti",
		"iat": now.Unix(),
		"nbf": now.Add(-time.Minute).Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
		"license": map[string]any{
			"type":                 licenseTypeCommercial,
			"version":              supportedSchemaVersion,
			"workspace_id":         "ws_1",
			"workspace_slug":       "ws-slug",
			"org_id":               "org_1",
			"org_slug":             "org-slug",
			"tier":                 "business",
			"targets":              []string{"*"},
			"features":             []string{},
			"add_ons":              []string{"webhooks"},
			"workspace_created_at": now.Add(-30 * 24 * time.Hour).Format(time.RFC3339),
		},
	}
}

func (m *testMinter) mint(t *testing.T, now time.Time, mutateHeader func(map[string]any), mutatePayload func(map[string]any)) []byte {
	t.Helper()
	header := map[string]any{"alg": "EdDSA", "kid": m.kid, "typ": "JWT"}
	if mutateHeader != nil {
		mutateHeader(header)
	}
	payload := defaultPayload(now)
	if mutatePayload != nil {
		mutatePayload(payload)
	}

	headerJSON, err := json.Marshal(header)
	require.NoError(t, err)
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := ed25519.Sign(m.priv, []byte(signingInput))
	return []byte(signingInput + "." + base64.RawURLEncoding.EncodeToString(sig))
}

func TestValidateToken_AcceptsOwnMintedToken(t *testing.T) {
	m := newTestMinter(t)
	now := time.Now().UTC().Truncate(time.Second)

	claims, err := validateToken(m.mint(t, now, nil, nil), m.jwks(t), now)
	require.NoError(t, err)
	assert.Equal(t, "ws_1", claims.WorkspaceID)
	assert.Equal(t, []string{"webhooks"}, claims.AddOns)
	assert.Equal(t, []string{"*"}, claims.Targets)
	assert.Empty(t, claims.Message)
}

func TestValidateToken_IgnoresSurroundingWhitespace(t *testing.T) {
	m := newTestMinter(t)
	now := time.Now().UTC().Truncate(time.Second)
	token := m.mint(t, now, nil, nil)

	for _, padded := range [][]byte{
		append(append([]byte{}, token...), '\n'),
		append(append([]byte{}, token...), "\r\n"...),
		append([]byte("  \n"), token...),
	} {
		claims, err := validateToken(padded, m.jwks(t), now)
		require.NoError(t, err)
		assert.Equal(t, "ws_1", claims.WorkspaceID)
	}

	_, err := validateToken([]byte("  \n"), m.jwks(t), now)
	require.ErrorIs(t, err, ErrInvalidLicenseToken)
}

func TestValidateToken_CarriesOptionalMessage(t *testing.T) {
	m := newTestMinter(t)
	now := time.Now().UTC().Truncate(time.Second)
	const message = "Licensed to Example Corp under agreement EX-2026-01."

	claims, err := validateToken(m.mint(t, now, nil, func(p map[string]any) {
		p["license"].(map[string]any)["message"] = message
	}), m.jwks(t), now)
	require.NoError(t, err)
	assert.Equal(t, message, claims.Message)

	info, err := claims.authenticatedInfo()
	require.NoError(t, err)
	assert.Equal(t, generationaccess.GeneratedLicenseCommercial, info.GeneratedLicense)
}

func TestValidateToken_Rejections(t *testing.T) {
	m := newTestMinter(t)
	other := newTestMinter(t)
	now := time.Now().UTC().Truncate(time.Second)

	testCases := []struct {
		name        string
		token       []byte
		jwks        []byte
		wantMessage string
	}{
		{
			name:  "empty token",
			token: []byte(""),
			jwks:  m.jwks(t),
		},
		{
			name:  "not compact - two segments",
			token: []byte("abc.def"),
			jwks:  m.jwks(t),
		},
		{
			name:  "not compact - four segments",
			token: append(m.mint(t, now, nil, nil), []byte(".extra")...),
			jwks:  m.jwks(t),
		},
		{
			name:  "json jws serialization",
			token: []byte(`{"payload":"e30","signatures":[{"protected":"e30","signature":""}]}`),
			jwks:  m.jwks(t),
		},
		{
			name:  "tampered signature",
			token: tamperPayload(m.mint(t, now, nil, nil)),
			jwks:  m.jwks(t),
		},
		{
			name:  "signed by unknown key",
			token: other.mint(t, now, nil, nil),
			jwks:  m.jwks(t),
		},
		{
			name:  "kid not in set",
			token: m.mint(t, now, func(h map[string]any) { h["kid"] = "unknown-kid" }, nil),
			jwks:  m.jwks(t),
		},
		{
			name:  "missing kid",
			token: m.mint(t, now, func(h map[string]any) { delete(h, "kid") }, nil),
			jwks:  m.jwks(t),
		},
		{
			name:  "wrong alg",
			token: m.mint(t, now, func(h map[string]any) { h["alg"] = "RS256" }, nil),
			jwks:  m.jwks(t),
		},
		{
			name:  "wrong issuer",
			token: m.mint(t, now, nil, func(p map[string]any) { p["iss"] = "someone-else" }),
			jwks:  m.jwks(t),
		},
		{
			name:  "wrong audience",
			token: m.mint(t, now, nil, func(p map[string]any) { p["aud"] = []string{"other-audience"} }),
			jwks:  m.jwks(t),
		},
		{
			name:  "expired beyond skew",
			token: m.mint(t, now, nil, func(p map[string]any) { p["exp"] = now.Add(-2 * time.Minute).Unix() }),
			jwks:  m.jwks(t),
		},
		{
			name:  "not yet valid beyond skew",
			token: m.mint(t, now, nil, func(p map[string]any) { p["nbf"] = now.Add(2 * time.Minute).Unix() }),
			jwks:  m.jwks(t),
		},
		{
			name:  "issued in the future beyond skew",
			token: m.mint(t, now, nil, func(p map[string]any) { p["iat"] = now.Add(2 * time.Minute).Unix() }),
			jwks:  m.jwks(t),
		},
		{
			name:  "missing exp",
			token: m.mint(t, now, nil, func(p map[string]any) { delete(p, "exp") }),
			jwks:  m.jwks(t),
		},
		{
			name:  "missing iat",
			token: m.mint(t, now, nil, func(p map[string]any) { delete(p, "iat") }),
			jwks:  m.jwks(t),
		},
		{
			name:  "missing nbf",
			token: m.mint(t, now, nil, func(p map[string]any) { delete(p, "nbf") }),
			jwks:  m.jwks(t),
		},
		{
			name:  "missing license claim",
			token: m.mint(t, now, nil, func(p map[string]any) { delete(p, "license") }),
			jwks:  m.jwks(t),
		},
		{
			name: "license type not commercial",
			token: m.mint(t, now, nil, func(p map[string]any) {
				p["license"].(map[string]any)["type"] = "agpl"
			}),
			jwks: m.jwks(t),
		},
		{
			name: "unsupported license schema version",
			token: m.mint(t, now, nil, func(p map[string]any) {
				p["license"].(map[string]any)["version"] = 2
			}),
			jwks: m.jwks(t),
		},
		{
			name: "missing targets",
			token: m.mint(t, now, nil, func(p map[string]any) {
				delete(p["license"].(map[string]any), "targets")
			}),
			jwks:        m.jwks(t),
			wantMessage: "missing targets",
		},
		{
			name: "empty targets",
			token: m.mint(t, now, nil, func(p map[string]any) {
				p["license"].(map[string]any)["targets"] = []string{}
			}),
			jwks:        m.jwks(t),
			wantMessage: "targets must not be empty",
		},
		{
			name: "uppercase target",
			token: m.mint(t, now, nil, func(p map[string]any) {
				p["license"].(map[string]any)["targets"] = []string{"Go"}
			}),
			jwks:        m.jwks(t),
			wantMessage: `invalid target "Go"`,
		},
		{
			name: "target containing a space",
			token: m.mint(t, now, nil, func(p map[string]any) {
				p["license"].(map[string]any)["targets"] = []string{"type script"}
			}),
			jwks:        m.jwks(t),
			wantMessage: `invalid target "type script"`,
		},
		{
			name: "wildcard combined with a named target",
			token: m.mint(t, now, nil, func(p map[string]any) {
				p["license"].(map[string]any)["targets"] = []string{"*", "go"}
			}),
			jwks:        m.jwks(t),
			wantMessage: "wildcard target cannot be combined with named targets",
		},
		{
			name:  "empty jwk set",
			token: m.mint(t, now, nil, nil),
			jwks:  []byte(`{"keys":[]}`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := validateToken(testCase.token, testCase.jwks, now)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrInvalidLicenseToken)
			if testCase.wantMessage != "" {
				assert.Contains(t, err.Error(), testCase.wantMessage)
			}
		})
	}
}

func TestValidateToken_ToleratesSkew(t *testing.T) {
	m := newTestMinter(t)
	now := time.Now().UTC().Truncate(time.Second)

	token := m.mint(t, now, nil, func(p map[string]any) {
		p["exp"] = now.Add(-30 * time.Second).Unix()
		p["iat"] = now.Add(30 * time.Second).Unix()
	})
	_, err := validateToken(token, m.jwks(t), now)
	require.NoError(t, err)
}

func TestValidateToken_AcceptsFlattenedAudienceString(t *testing.T) {
	m := newTestMinter(t)
	now := time.Now().UTC().Truncate(time.Second)

	token := m.mint(t, now, nil, func(p map[string]any) { p["aud"] = audience })
	_, err := validateToken(token, m.jwks(t), now)
	require.NoError(t, err)
}

func TestClaims_AuthenticatedInfo(t *testing.T) {
	createdAt := time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name            string
		claims          claims
		wantErr         bool
		wantAccountType shared.AccountType
		wantAddOns      []shared.BillingAddOn
	}{
		{
			name: "business tier maps directly",
			claims: claims{
				Type: licenseTypeCommercial, Version: supportedSchemaVersion,
				WorkspaceID: "ws_1", Tier: "business",
				AddOns:             []string{"webhooks", "sdk_testing"},
				WorkspaceCreatedAt: createdAt,
				TelemetryDisabled:  true,
			},
			wantAccountType: shared.AccountTypeBusiness,
			wantAddOns:      []shared.BillingAddOn{shared.BillingAddOnWebhooks, shared.BillingAddOnSDKTesting},
		},
		{
			name: "free tier maps directly",
			claims: claims{
				Type: licenseTypeCommercial, Version: supportedSchemaVersion,
				WorkspaceID: "ws_1", Tier: "free",
				WorkspaceCreatedAt: createdAt,
			},
			wantAccountType: shared.AccountTypeFree,
		},
		{
			name: "oss tier maps to enterprise",
			claims: claims{
				Type: licenseTypeCommercial, Version: supportedSchemaVersion,
				WorkspaceID: "ws_1", Tier: "oss",
				WorkspaceCreatedAt: createdAt,
			},
			wantAccountType: shared.AccountTypeEnterprise,
		},
		{
			name: "unknown add-ons are dropped not fatal",
			claims: claims{
				Type: licenseTypeCommercial, Version: supportedSchemaVersion,
				WorkspaceID: "ws_1", Tier: "enterprise",
				AddOns:             []string{"webhooks", "some_future_add_on"},
				WorkspaceCreatedAt: createdAt,
			},
			wantAccountType: shared.AccountTypeEnterprise,
			wantAddOns:      []shared.BillingAddOn{shared.BillingAddOnWebhooks},
		},
		{
			name: "unknown tier is rejected",
			claims: claims{
				Type: licenseTypeCommercial, Version: supportedSchemaVersion,
				WorkspaceID: "ws_1", Tier: "some-future-tier",
				WorkspaceCreatedAt: createdAt,
			},
			wantErr: true,
		},
		{
			name: "missing workspace creation time is rejected",
			claims: claims{
				Type: licenseTypeCommercial, Version: supportedSchemaVersion,
				WorkspaceID: "ws_1", Tier: "business",
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			info, err := testCase.claims.authenticatedInfo()
			if testCase.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.wantAccountType, info.AccountType)
			assert.Equal(t, testCase.wantAddOns, info.BillingAddOns)
			assert.Equal(t, testCase.claims.TelemetryDisabled, info.TelemetryDisabled)
			assert.Equal(t, generationaccess.GeneratedLicenseCommercial, info.GeneratedLicense)
			require.NotNil(t, info.WorkspaceCreatedAt)
			assert.Equal(t, createdAt, info.WorkspaceCreatedAt.UTC())

			ctx, err := generationaccess.WithAuthenticated(context.Background(), info)
			require.NoError(t, err)
			state, ok := generationaccess.StateFromContext(ctx)
			require.True(t, ok)
			assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())
		})
	}
}

func TestResolveGenerationAccess_NoTokenIsANoOp(t *testing.T) {
	ctx := context.Background()

	resolved, verification, err := ResolveGenerationAccess(ctx)
	require.NoError(t, err)
	assert.Nil(t, verification)
	assert.Equal(t, ctx, resolved)

	_, ok := generationaccess.StateFromContext(resolved)
	assert.False(t, ok)
}

func TestResolveGenerationAccess_ValidTokenEstablishesCommercialState(t *testing.T) {
	token, jwks := readGolden(t, "golden-token.jwt")
	withValidator(t, jwks, goldenClock)

	ctx := WithToken(context.Background(), token)
	resolved, verification, err := ResolveGenerationAccess(ctx)
	require.NoError(t, err)
	require.NotNil(t, verification)
	assert.Equal(t, []string{"go"}, verification.Targets)
	assert.True(t, verification.Covers("go"))
	assert.False(t, verification.Covers("typescript"))

	state, ok := generationaccess.StateFromContext(resolved)
	require.True(t, ok)
	assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())
	workspaceID, authenticated := state.WorkspaceID()
	assert.True(t, authenticated)
	assert.Equal(t, "golden-workspace", workspaceID)
	assert.True(t, state.TelemetryDisabled())
}

func TestResolveGenerationAccess_TokenOverridesPreexistingDirectState(t *testing.T) {
	token, jwks := readGolden(t, "golden-token.jwt")
	withValidator(t, jwks, goldenClock)

	ctx := WithToken(generationaccess.WithDirect(context.Background()), token)
	resolved, _, err := ResolveGenerationAccess(ctx)
	require.NoError(t, err)

	state, ok := generationaccess.StateFromContext(resolved)
	require.True(t, ok)
	assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())
}

func TestResolveGenerationAccess_InvalidTokenIsAHardError(t *testing.T) {
	token, jwks := readGolden(t, "golden-token.jwt")
	withValidator(t, jwks, goldenClock)

	ctx := WithToken(generationaccess.WithDirect(context.Background()), tamperPayload(token))
	_, _, err := ResolveGenerationAccess(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidLicenseToken)
}

func TestEmbeddedProductionJWKS_TrustsOnlyProductionKeys(t *testing.T) {
	var set jsonWebKeySet
	require.NoError(t, json.Unmarshal(embeddedProductionJWKS, &set))
	require.NotEmpty(t, set.Keys, "embedded production JWKS must not be empty")
	for _, k := range set.Keys {
		assert.Equal(t, "OKP", k.Kty)
		assert.Equal(t, "Ed25519", k.Crv)
		assert.NotEmpty(t, k.Kid)
		goldenJWKS, err := os.ReadFile("testdata/golden-jwks.json")
		require.NoError(t, err)
		var golden jsonWebKeySet
		require.NoError(t, json.Unmarshal(goldenJWKS, &golden))
		assert.NotEqual(t, golden.Keys[0].Kid, k.Kid, "test key present in production JWKS")
	}

	token, _ := readGolden(t, "golden-token.jwt")
	_, _, err := ResolveGenerationAccess(WithToken(context.Background(), token))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidLicenseToken)
}

func TestInspect(t *testing.T) {
	token, jwks := readGolden(t, "golden-token.jwt")
	withValidator(t, jwks, goldenClock)

	info, err := Inspect(token)
	require.NoError(t, err)
	assert.Equal(t, "golden-workspace", info.WorkspaceID)
	assert.Equal(t, "golden-ws", info.WorkspaceSlug)
	assert.Equal(t, "golden", info.OrgSlug)
	assert.Equal(t, "business", info.Tier)
	assert.Equal(t, []string{"go"}, info.Targets)
	assert.Equal(t, goldenClock.Unix(), info.IssuedAt.Unix())
	assert.Equal(t, goldenClock.Add(30*24*time.Hour).Unix(), info.ExpiresAt.Unix())

	_, err = Inspect([]byte("not-a-token"))
	require.ErrorIs(t, err, ErrInvalidLicenseToken)
}

func TestResolveGenerationAccess_ReportsVerification(t *testing.T) {
	token, jwks := readGolden(t, "golden-token.jwt")
	withValidator(t, jwks, goldenClock)

	_, verified, err := ResolveGenerationAccess(context.Background())
	require.NoError(t, err)
	assert.Nil(t, verified)

	resolved, verified, err := ResolveGenerationAccess(WithToken(context.Background(), token))
	require.NoError(t, err)
	require.NotNil(t, verified)
	assert.Equal(t, []string{"go"}, verified.Targets)
	state, ok := generationaccess.StateFromContext(resolved)
	require.True(t, ok)
	assert.Equal(t, generationaccess.GeneratedLicenseCommercial, state.GeneratedLicense())

	_, verified, err = ResolveGenerationAccess(WithToken(generationaccess.WithDirect(context.Background()), token))
	require.NoError(t, err)
	require.NotNil(t, verified)

	_, verified, err = ResolveGenerationAccess(WithToken(context.Background(), []byte("forged.token.value")))
	require.ErrorIs(t, err, ErrInvalidLicenseToken)
	assert.Nil(t, verified)

	_, verified, err = ResolveGenerationAccess(WithToken(generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()), token))
	require.ErrorIs(t, err, ErrConflictingLicenseElection)
	assert.Nil(t, verified)
}

func TestTokenCoverage(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		targets []string
		target  string
		want    bool
	}{
		{name: "named target", targets: []string{"go"}, target: "go", want: true},
		{name: "different named target", targets: []string{"go"}, target: "typescript"},
		{name: "wildcard", targets: []string{"*"}, target: "typescript", want: true},
		{name: "empty coverage", target: "go"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want, TokenInfo{Targets: testCase.targets}.Covers(testCase.target))
			assert.Equal(t, testCase.want, Verification{Targets: testCase.targets}.Covers(testCase.target))
		})
	}
}

func TestValidateToken_RejectsCritHeader(t *testing.T) {
	m := newTestMinter(t)
	token := m.mint(t, goldenClock, func(h map[string]any) { h["crit"] = []string{"b64"} }, nil)

	_, err := validateToken(token, m.jwks(t), goldenClock)
	require.ErrorIs(t, err, ErrInvalidLicenseToken)
	assert.Contains(t, err.Error(), "crit")
}

func TestInspect_RejectsTokensGenerationCannotUse(t *testing.T) {
	m := newTestMinter(t)
	withValidator(t, m.jwks(t), goldenClock)

	licenseOf := func(payload map[string]any) map[string]any { return payload["license"].(map[string]any) }
	for name, mutate := range map[string]func(map[string]any){
		"unsupported tier":                func(p map[string]any) { licenseOf(p)["tier"] = "future-tier" },
		"missing workspace id":            func(p map[string]any) { licenseOf(p)["workspace_id"] = "" },
		"missing workspace creation time": func(p map[string]any) { delete(licenseOf(p), "workspace_created_at") },
	} {
		t.Run(name, func(t *testing.T) {
			token := m.mint(t, goldenClock, nil, mutate)

			_, err := Inspect(token)
			require.ErrorIs(t, err, ErrInvalidLicenseToken)

			_, _, err = ResolveGenerationAccess(WithToken(context.Background(), token))
			require.ErrorIs(t, err, ErrInvalidLicenseToken)
		})
	}

	info, err := Inspect(m.mint(t, goldenClock, nil, nil))
	require.NoError(t, err)
	assert.Equal(t, "ws_1", info.WorkspaceID)
	assert.Equal(t, "business", info.Tier)
}

// Flips the first payload character: its top bits always change the decoded byte, unlike the padding-carrying last character.
func tamperPayload(token []byte) []byte {
	out := append([]byte(nil), token...)
	i := bytes.IndexByte(out, '.') + 1
	if out[i] == 'A' {
		out[i] = 'B'
	} else {
		out[i] = 'A'
	}
	return out
}
