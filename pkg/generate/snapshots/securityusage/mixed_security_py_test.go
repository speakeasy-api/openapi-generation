package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Py(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `python:
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

<!-- UsageSnippet language="python" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `python
from mixed_security_sdk import SDK, models


with SDK(
    security=models.Security(
        option1=models.SecurityOption1(
            auth_a1="<YOUR_API_KEY_HERE>",
            auth_a2=models.SchemeAuthA2(
                username="",
                password="",
            ),
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

## option1_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `python
from mixed_security_sdk import SDK, models


with SDK(
    security=models.Security(
        option1=models.SecurityOption1(
            auth_a1="<YOUR_API_KEY_HERE>",
            auth_a2=models.SchemeAuthA2(
                username="",
                password="",
            ),
        ),
    ),
) as sdk:

    sdk.option1_hoisted()

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

## option2_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `python
from mixed_security_sdk import SDK, models


with SDK(
    security=models.Security(
        option2=models.SecurityOption2(
            auth_b="<YOUR_JWT>",
        ),
    ),
) as sdk:

    sdk.option2_hoisted()

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

## option1_not_allowed

### Example Usage

<!-- UsageSnippet language="python" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `python
from mixed_security_sdk import SDK, models


with SDK(
    security=models.Security(
        option3=models.SecurityOption3(
            auth_c="<YOUR_AUTH_C_HERE>",
        ),
    ),
) as sdk:

    sdk.option1_not_allowed()

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

## op_level_mixed_auth

### Example Usage

<!-- UsageSnippet language="python" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `python
from mixed_security_sdk import SDK, models


with SDK() as sdk:

    sdk.op_level_mixed_auth(security=models.OpLevelMixedAuthSecurity(
        option1=models.OpLevelMixedAuthSecurityOption1(
            auth_a1="<YOUR_API_KEY_HERE>",
            auth_b="<YOUR_JWT>",
        ),
    ))

    # Use the SDK ...

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                            | Type                                                                 | Required                                                             | Description                                                          |
| -------------------------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------- | -------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                           | [models.OpLevelMixedAuthSecurity](../../oplevelmixedauthsecurity.md) | :heavy_check_mark:                                                   | The security requirements to use for the request.                    |
| ` + "`" + `retries` + "`" + `                                                            | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)     | :heavy_minus_sign:                                                   | Configuration to override the default retry behavior of the client.  |

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
