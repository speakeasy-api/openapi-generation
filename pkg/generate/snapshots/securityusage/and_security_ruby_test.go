package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Ruby(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `ruby:
  packageName: and-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [global_security](#global_security)
* [and_auth_hoisted](#and_auth_hoisted)
* [op_level_and_auth](#op_level_and_auth)

## global_security

### Example Usage

<!-- UsageSnippet language="ruby" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'and_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    auth1: '<YOUR_API_KEY_HERE>',
    auth2: Models::Components::SchemeAuth2.new(
      username: '',
      password: ''
    )
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

## and_auth_hoisted

### Example Usage

<!-- UsageSnippet language="ruby" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'and_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    auth1: '<YOUR_API_KEY_HERE>',
    auth2: Models::Components::SchemeAuth2.new(
      username: '',
      password: ''
    )
  )
)
res = s.and_auth_hoisted

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::AndAuthHoistedResponse)](../../models/operations/andauthhoistedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## op_level_and_auth

### Example Usage

<!-- UsageSnippet language="ruby" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'and_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.op_level_and_auth(security: Models::Operations::OpLevelAndAuthSecurity.new(
  auth2: Models::Components::SchemeAuth2.new(
    username: '',
    password: ''
  ),
  auth3: '<YOUR_BEARER_TOKEN_HERE>'
))

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                       | Type                                                                                            | Required                                                                                        | Description                                                                                     |
| ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                      | [Models::Operations::OpLevelAndAuthSecurity](../../models/operations/oplevelandauthsecurity.md) | :heavy_check_mark:                                                                              | The security requirements to use for the request.                                               |

### Response

**[T.nilable(Models::Operations::OpLevelAndAuthResponse)](../../models/operations/oplevelandauthresponse.md)**

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
