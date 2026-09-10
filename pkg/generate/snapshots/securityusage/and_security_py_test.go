package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Py(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `python:
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

<!-- UsageSnippet language="python" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `python
from and_security_sdk import SDK, models


with SDK(
    security=models.Security(
        auth1="<YOUR_API_KEY_HERE>",
        auth2=models.SchemeAuth2(
            username="",
            password="",
        ),
    ),
) as sdk:

    sdk.global_security()

    # Use the SDK ...

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| ` + "`" + `retries` + "`" + `                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

## and_auth_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `python
from and_security_sdk import SDK, models


with SDK(
    security=models.Security(
        auth1="<YOUR_API_KEY_HERE>",
        auth2=models.SchemeAuth2(
            username="",
            password="",
        ),
    ),
) as sdk:

    sdk.and_auth_hoisted()

    # Use the SDK ...

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| ` + "`" + `retries` + "`" + `                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

## op_level_and_auth

### Example Usage

<!-- UsageSnippet language="python" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `python
from and_security_sdk import SDK, models


with SDK() as sdk:

    sdk.op_level_and_auth(security=models.OpLevelAndAuthSecurity(
        auth2=models.SchemeAuth2(
            username="",
            password="",
        ),
        auth3="<YOUR_BEARER_TOKEN_HERE>",
    ))

    # Use the SDK ...

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                          | [models.OpLevelAndAuthSecurity](../../oplevelandauthsecurity.md)    | :heavy_check_mark:                                                  | The security requirements to use for the request.                   |
| ` + "`" + `retries` + "`" + `                                                           | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)    | :heavy_minus_sign:                                                  | Configuration to override the default retry behavior of the client. |

### Errors

| Error Type             | Status Code            | Content Type           |
| ---------------------- | ---------------------- | ---------------------- |
| errors.SDKDefaultError | 4XX, 5XX               | \*/\*                  |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
