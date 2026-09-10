package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Ruby(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `ruby:
  packageName: mixed-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [global_security](#global_security)
* [option1_hoisted](#option1_hoisted)
* [option2_hoisted](#option2_hoisted)
* [option1_not_allowed](#option1_not_allowed)
* [op_level_mixed_auth](#op_level_mixed_auth)

## global_security

### Example Usage

<!-- UsageSnippet language="ruby" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'mixed_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    option1: Models::Components::SecurityOption1.new(
      auth_a1: '<YOUR_API_KEY_HERE>',
      auth_a2: Models::Components::SchemeAuthA2.new(
        username: '',
        password: ''
      )
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

## option1_hoisted

### Example Usage

<!-- UsageSnippet language="ruby" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'mixed_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    option1: Models::Components::SecurityOption1.new(
      auth_a1: '<YOUR_API_KEY_HERE>',
      auth_a2: Models::Components::SchemeAuthA2.new(
        username: '',
        password: ''
      )
    )
  )
)
res = s.option1_hoisted

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::Option1HoistedResponse)](../../models/operations/option1hoistedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## option2_hoisted

### Example Usage

<!-- UsageSnippet language="ruby" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'mixed_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    option2: Models::Components::SecurityOption2.new(
      auth_b: '<YOUR_JWT>'
    )
  )
)
res = s.option2_hoisted

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::Option2HoistedResponse)](../../models/operations/option2hoistedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## option1_not_allowed

### Example Usage

<!-- UsageSnippet language="ruby" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'mixed_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new(
  security: Models::Components::Security.new(
    option3: Models::Components::SecurityOption3.new(
      auth_c: '<YOUR_AUTH_C_HERE>'
    )
  )
)
res = s.option1_not_allowed

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Response

**[T.nilable(Models::Operations::Option1NotAllowedResponse)](../../models/operations/option1notallowedresponse.md)**

### Errors

| Error Type       | Status Code      | Content Type     |
| ---------------- | ---------------- | ---------------- |
| Errors::APIError | 4XX, 5XX         | \*/\*            |

## op_level_mixed_auth

### Example Usage

<!-- UsageSnippet language="ruby" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `ruby
require 'mixed_security_sdk'

Models = ::OpenApiSDK::Models
s = ::OpenApiSDK::SDK.new
res = s.op_level_mixed_auth(security: Models::Operations::OpLevelMixedAuthSecurity.new(
  option1: Models::Operations::OpLevelMixedAuthSecurityOption1.new(
    auth_a1: '<YOUR_API_KEY_HERE>',
    auth_b: '<YOUR_JWT>'
  )
))

if res.status_code == 200
  # handle response
end

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                           | Type                                                                                                | Required                                                                                            | Description                                                                                         |
| --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                          | [Models::Operations::OpLevelMixedAuthSecurity](../../models/operations/oplevelmixedauthsecurity.md) | :heavy_check_mark:                                                                                  | The security requirements to use for the request.                                                   |

### Response

**[T.nilable(Models::Operations::OpLevelMixedAuthResponse)](../../models/operations/oplevelmixedauthresponse.md)**

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
