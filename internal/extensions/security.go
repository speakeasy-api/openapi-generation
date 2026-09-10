package extensions

import (
	"context"

	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/jsonschema/oas3/core"
	"github.com/speakeasy-api/openapi/marshaller"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

type CustomSecurityConfig struct {
	marshaller.Model[CoreCustomSecurityConfig]

	UsesScopes *bool
	Schema     *oas3.JSONSchema[oas3.Referenceable]
}

type CoreCustomSecurityConfig struct {
	marshaller.CoreModel `model:"customSecurityConfig"`

	UsesScopes marshaller.Node[*bool]           `key:"usesScopes"`
	Schema     marshaller.Node[core.JSONSchema] `key:"schema" required:"true"`
}

func (e *Extensions) HandleCustomSecurityConfig(ctx context.Context, exts OAExtensions) (*CustomSecurityConfig, error) {
	var css CustomSecurityConfig
	vErrs, err := extensions.UnmarshalExtensionModel[CustomSecurityConfig, CoreCustomSecurityConfig](ctx, exts, e.GetResolvedName(ExtCustomSecurityScheme), &css)
	if err != nil {
		if errors.Is(err, extensions.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if len(vErrs) > 0 {
		var vErr validation.Error
		if !errors.As(vErrs[0], &vErr) {
			vErr = validation.Error{
				UnderlyingError: vErrs[0],
			}
		}

		return nil, errors.NewValidationError("failed to validate "+ExtGlobals.Name(), vErr.GetNode(), errors.Join(vErrs...))
	}
	return &css, nil
}

func (e *Extensions) GetTokenServerAuthentication(securityScheme *openapi.OAuthFlow) (string, error) {
	val, err := getExtensionValueWithValidation(e.GetResolvedName(ExtTokenEndpointAuth), securityScheme.GetExtensions(), "client_secret_post", func(t string) error {
		if t == "client_secret_post" || t == "client_secret_basic" {
			return nil
		}
		return errors.Error("x-speakeasy-token-endpoint-authentication only supports type client_secret_post and client_secret_basic")
	})
	if err != nil {
		return "", err
	}
	if val == "" {
		val = "client_secret_post"
	}
	return val, nil
}

func (e *Extensions) GetOverridableOAuth2Scopes(flow *openapi.OAuthFlow) (bool, error) {
	// The extension is expected at the flow level (not inside the scopes node)
	// E.g.
	//     clientCredentials:
	//       type: oauth2
	//       flows:
	//         clientCredentials:
	//           tokenUrl: /clientcredentials/token
	//           x-speakeasy-overridable-scopes: true
	//           scopes:
	//             read: Read access
	//             write: Write access

	if flow == nil {
		return false, nil
	}

	extName := e.GetResolvedName(ExtOverridableOAuth2Scopes)

	// Check for the extension at the flow level
	ext, ok := flow.GetExtensions().Get(extName)
	if ok && ext != nil {
		var overridable bool
		if err := ext.Decode(&overridable); err != nil {
			return false, errors.NewValidationError("failed to unmarshal "+extName, nil, ErrUnmarshal.Wrap(err))
		}
		return overridable, nil
	}

	return false, nil
}
