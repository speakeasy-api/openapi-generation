package hooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/speakeasy/customhttp/models/components"
)

type CustomSecurityHook struct{}

var _ beforeRequestHook = (*CustomSecurityHook)(nil)

func (i *CustomSecurityHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	switch hookCtx.OperationID {
	case "customHttpOnly":
		if hookCtx.SecuritySource == nil {
			return nil, errors.New("security source is nil")
		}
		sec, err := hookCtx.SecuritySource(req.Context())
		if err != nil {
			return nil, err
		}

		security, ok := sec.(*components.Security)
		if !ok {
			return nil, errors.New("security source is not of type Security")
		}

		if customHTTP := security.GetCustomHTTP(); customHTTP == nil {
			return nil, errors.New("customHTTP security is not defined")
		} else {
			req.Header.Set("X-Security-UserID", fmt.Sprintf("%d", customHTTP.GetUserID()))
			req.Header.Set("X-Security-Role", string(customHTTP.GetRole()))
			req.Header.Set("X-Security-Passphrase", customHTTP.GetPassphrase())

			if accessCode := customHTTP.GetAccessCode(); accessCode != nil {
				req.Header.Set("X-Security-AccessCode", fmt.Sprintf("%d", *accessCode))
			}

			if scopes := customHTTP.GetScopes(); len(scopes) > 0 {
				scopesJSON, err := json.Marshal(scopes)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal scopes: %w", err)
				}
				req.Header.Set("X-Security-Scopes", string(scopesJSON))
			}
		}
	}

	return req, nil
}
