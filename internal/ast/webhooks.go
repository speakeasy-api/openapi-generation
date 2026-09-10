package ast

import "github.com/speakeasy-api/openapi-generation/v2/internal/extensions"

// A webhook configuration indicates that the operation is a webhook operation
type Webhook struct {
	// A webhook key is used to identify the PathItemObject in the OpenAPI document, similar to the `Operation.Path` but for webhooks
	Key string `yaml:",omitempty"`

	// A webhook security configuration
	Security *extensions.WebhookSecurity `yaml:",omitempty"`
}
