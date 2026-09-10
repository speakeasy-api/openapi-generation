package hooks

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"example.com/openapi-go-sdk/internal/config"
)

type TestHook struct {
	initSDKVersion string
}

var (
	_ sdkInitHook       = (*TestHook)(nil)
	_ beforeRequestHook = (*TestHook)(nil)
	_ afterSuccessHook  = (*TestHook)(nil)
	_ afterErrorHook    = (*TestHook)(nil)
)

func (i *TestHook) SDKInit(config config.SDKConfiguration) config.SDKConfiguration {
	i.initSDKVersion = config.SDKVersion

	client := config.Client

	if client == nil {
		client = &http.Client{}
	}

	config.Client = &TestClient{
		underlyingClient: client,
	}

	return config
}

func (i *TestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	req.Header.Set("Idempotency-Key", "some-key")

	switch hookCtx.OperationID {
	case "testHooks":
		values := req.URL.Query()
		values.Set("someParam", "overriddenParam")
		req.URL.RawQuery = values.Encode()
	case "authorizationHeaderModification":
		req.Header.Set("Authorization", req.Header.Get("Authorization")+" modified")
	case "hooksCustomUserAgent":
		req.Header.Set("User-Agent", fmt.Sprintf("acme-corp/%s acme-corp/%s", hookCtx.SDKConfiguration.SDKVersion, runtime.Version()))
		req.Header.Set("X-Test-Gen-Version", hookCtx.SDKConfiguration.GenVersion)
		req.Header.Set("X-Test-Doc-Version", hookCtx.SDKConfiguration.OpenAPIDocVersion)
		req.Header.Set("X-Test-Init-Sdk-Version", i.initSDKVersion)
	}

	return req, nil
}

func (i *TestHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "testHooksAfterResponse":
		return nil, errors.New("validation failed")
	}

	return res, nil
}

func (i *TestHook) AfterError(hookCtx AfterErrorContext, res *http.Response, err error) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "testHooksError":
		if res != nil && res.StatusCode != 400 {
			return nil, errors.New("expected status code 400")
		}
		return nil, errors.New("special test error case")
	case "statusGetDefaultError":
		if res != nil && res.StatusCode == 418 {
			return &http.Response{
				StatusCode: 200,
				Status:     "200 OK",
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		}
	}

	return res, err
}

type TestClient struct {
	underlyingClient HTTPClient
}

var _ HTTPClient = (*TestClient)(nil)

func (c *TestClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("Client-Level-Header", "added by client")

	return c.underlyingClient.Do(req)
}
