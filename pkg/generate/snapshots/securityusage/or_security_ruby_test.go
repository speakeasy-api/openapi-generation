package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Ruby(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `ruby:
  packageName: multi-auth-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [global_security](#global_security)
* [auth1_hoisted](#auth1_hoisted)
* [auth2_hoisted](#auth2_hoisted)
* [auth2_preferred](#auth2_preferred)
* [op_level_client_credentials](#op_level_client_credentials)

## global_security

### Example Usage

<!-- UsageSnippet language="ruby" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'multi_auth_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    auth1: '<YOUR_API_KEY_HERE>'
  )
)
res = s.global_security

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::GlobalSecurityResponse)](../../models/operations/globalsecurityresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## auth1_hoisted

### Example Usage

<!-- UsageSnippet language="ruby" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'multi_auth_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    auth1: '<YOUR_API_KEY_HERE>'
  )
)
res = s.auth1_hoisted

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::Auth1HoistedResponse)](../../models/operations/auth1hoistedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## auth2_hoisted

### Example Usage

<!-- UsageSnippet language="ruby" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'multi_auth_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    auth2: Models::Components::SchemeAuth2.new(
      username: '',
      password: ''
    )
  )
)
res = s.auth2_hoisted

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::Auth2HoistedResponse)](../../models/operations/auth2hoistedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## auth2_preferred

### Example Usage

<!-- UsageSnippet language="ruby" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'multi_auth_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    auth2: Models::Components::SchemeAuth2.new(
      username: '',
      password: ''
    )
  )
)
res = s.auth2_preferred

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::Auth2PreferredResponse)](../../models/operations/auth2preferredresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## op_level_client_credentials

### Example Usage

<!-- UsageSnippet language="ruby" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'multi_auth_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.op_level_client_credentials(security: Models::Operations::OpLevelClientCredentialsSecurity.new(
  client_id: '<YOUR_CLIENT_ID_HERE>',
  client_secret: '<YOUR_CLIENT_SECRET_HERE>'
))

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                           | Type                                                                                                                | Required                                                                                                            | Description                                                                                                         |
| ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                                          | [Models::Operations::OpLevelClientCredentialsSecurity](../../models/operations/oplevelclientcredentialssecurity.md) | :heavy_check_mark:                                                                                                  | The security requirements to use for the request.                                                                   |

### Response

**[T.nilable(Models::Operations::OpLevelClientCredentialsResponse)](../../models/operations/oplevelclientcredentialsresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
