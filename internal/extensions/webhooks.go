package extensions

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

// A Webhooks extension is used to configure webhooks for an API.
type Webhooks struct {
	// A WebhookSecurity object is used to configure the security for a webhook.
	Security *WebhookSecurity `json:"security" yaml:"security,omitempty"`
}

type WebhookSecurity struct {
	// Valid values for this field are:
	//
	// * "signature" a configurable signature which respects header name, text encoding and algorithm
	//
	// * "custom" a custom signature down to the API producer to complete the implementation
	//
	// * "signatureStandardWebhooks" a preset which conforms to "Standard Webhooks" naming / guidance
	//
	// * "apiKey" an API key - no signing is performed
	Type string `json:"type" yaml:"type"`
	// Applicable when Type is "signature"
	HeaderName string `json:"headerName" yaml:"headerName,omitempty"`
	// Applicable when Type is "signature"
	SignatureTextEncoding string `json:"signatureTextEncoding" yaml:"signatureTextEncoding,omitempty"`
	// Applicable when Type is "signature"
	SignatureAlgorithm string `json:"algorithm" yaml:"algorithm,omitempty"`
	// ConsumerShouldProvideSecret is true if the webhook consumer should provide the secret - allows for "custom" type to override the default behavior
	ConsumerShouldProvideSecret *bool `json:"consumerShouldProvideSecret" yaml:"consumerShouldProvideSecret,omitempty"`
}

func (w *WebhookSecurity) isValidType() error {
	validSecurityTypes := []string{"signature", "custom"}
	if !slices.Contains(validSecurityTypes, w.Type) {
		return fmt.Errorf("invalid x-speakeasy-webhooks.security.type supported: %q, got: %q", strings.Join(validSecurityTypes, ", "), w.Type)
	}
	return nil
}

func (w *WebhookSecurity) isValidHeaderName() error {
	if strings.TrimSpace(w.HeaderName) == "" {
		return fmt.Errorf("invalid x-speakeasy-webhooks.security.headerName got: %q", w.HeaderName)
	}
	return nil
}

func (w *WebhookSecurity) isValidSignatureTextEncoding() error {
	if w.Type != "signature" {
		return nil
	}

	validFormats := []string{"base64", "base64url", "hex"}
	if !slices.Contains(validFormats, w.SignatureTextEncoding) {
		return fmt.Errorf("invalid x-speakeasy-webhooks.security.signatureFormat supported: %q, got: %q", strings.Join(validFormats, ","), w.SignatureTextEncoding)
	}

	return nil
}

func (w *WebhookSecurity) isValidSignatureAlgorithm() error {
	if w.Type != "signature" {
		return nil
	}

	validSignatureAlgorithms := []string{"hmac-sha256"}
	if !slices.Contains(validSignatureAlgorithms, w.SignatureAlgorithm) {
		return fmt.Errorf(
			"invalid x-speakeasy-webhooks.security.algorithm supported: %q, got: %q",
			strings.Join(validSignatureAlgorithms, ","),
			w.SignatureAlgorithm,
		)
	}
	return nil
}

// applyDefaults fills in default values (if missing).
func (w *WebhookSecurity) applyDefaults() {
	if strings.TrimSpace(w.Type) == "" {
		w.Type = "signature"
	}

	if strings.TrimSpace(w.HeaderName) == "" {
		w.HeaderName = "x-webhook-signature"
	}

	if w.ConsumerShouldProvideSecret == nil {
		consumerShouldProvideSecret := true
		w.ConsumerShouldProvideSecret = &consumerShouldProvideSecret
	}

	if w.Type == "signature" {
		if strings.TrimSpace(w.SignatureAlgorithm) == "" {
			w.SignatureAlgorithm = "hmac-sha256"
		}
		if strings.TrimSpace(w.SignatureTextEncoding) == "" {
			w.SignatureTextEncoding = "base64"
		}
		consumerShouldProvideSecret := true
		w.ConsumerShouldProvideSecret = &consumerShouldProvideSecret
	}
}

func (e *Extensions) HandleWebhooksExtension(extensions OAExtensions) (*Webhooks, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	extension, ok := e.findExtension(extensions, ExtWebhooks)
	if !ok {
		return nil, nil
	}

	lineNumber := extension

	var cfg Webhooks
	if err := extension.Decode(&cfg); err != nil {
		return nil, errors.NewValidationError(
			"failed to unmarshal "+e.GetResolvedName(ExtWebhooks),
			lineNumber,
			ErrUnmarshal.Wrap(err),
		)
	}

	if reflect.DeepEqual(cfg, Webhooks{}) {
		return nil, errors.NewValidationError("x-speakeasy-webhooks: value is empty or unrecognized", lineNumber, nil)
	}

	// Fill in defaults
	cfg.Security.applyDefaults()

	if err := cfg.Security.isValidType(); err != nil {
		return nil, errors.NewValidationError(err.Error(), lineNumber, nil)
	}

	if err := cfg.Security.isValidHeaderName(); err != nil {
		return nil, errors.NewValidationError(err.Error(), lineNumber, nil)
	}

	if err := cfg.Security.isValidSignatureTextEncoding(); err != nil {
		return nil, errors.NewValidationError(err.Error(), lineNumber, nil)
	}

	if err := cfg.Security.isValidSignatureAlgorithm(); err != nil {
		return nil, errors.NewValidationError(err.Error(), lineNumber, nil)
	}

	return &cfg, nil
}
