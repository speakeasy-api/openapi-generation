package hooks

import (
	"errors"
	"net/http"

	"example.com/openapi-go-sdk/pkg/models/shared"
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

		security, ok := sec.(shared.Security)
		if !ok {
			return nil, errors.New("security source is not of type Security")
		}

		if customSecurity := security.GetCustomSchemeAppID(); customSecurity == nil {
			return nil, errors.New("customSchemeAppId security is not defined")
		} else {
			req.Header.Set("X-Security-App-Id", customSecurity.AppID)
			req.Header.Set("X-Security-Secret", customSecurity.Secret)
		}
	}

	return req, nil
}
