package hooks

import (
	"errors"
	"net/http"

	"openapi/internal/sdk/models/operations"
)

type CustomSecurityHook struct{}

var _ beforeRequestHook = (*CustomSecurityHook)(nil)

func (i *CustomSecurityHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	switch hookCtx.OperationID {
	case "customSchemeAppId":
		if hookCtx.SecuritySource == nil {
			return nil, errors.New("security source is nil")
		}
		sec, err := hookCtx.SecuritySource(req.Context())
		if err != nil {
			return nil, err
		}
		customSchemeAppIDSecurity, ok := sec.(operations.CustomSchemeAppIDSecurity)
		if !ok {
			return nil, errors.New("security source is not of type CustomSchemeAppIDSecurity")
		}
		req.Header.Set("X-Security-App-Id", customSchemeAppIDSecurity.AppID)
		req.Header.Set("X-Security-Secret", customSchemeAppIDSecurity.Secret)
	}

	return req, nil
}
