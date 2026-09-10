package hooks

import (
	"net/http"

	"github.com/google/uuid"
)

type IdempotencyHook struct{}

var _ beforeRequestHook = (*IdempotencyHook)(nil)

func (i *IdempotencyHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	u := uuid.New()

	req.Header.Set("Idempotency-Key", u.String())
	return req, nil
}
