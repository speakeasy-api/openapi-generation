package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Py(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `python:
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

<!-- UsageSnippet language="python" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `python
from multi_auth_sdk import SDK, models


with SDK(
    security=models.Security(
        auth1="<YOUR_API_KEY_HERE>",
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

## auth1_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `python
from multi_auth_sdk import SDK, models


with SDK(
    security=models.Security(
        auth1="<YOUR_API_KEY_HERE>",
    ),
) as sdk:

    sdk.auth1_hoisted()

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

## auth2_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `python
from multi_auth_sdk import SDK, models


with SDK(
    security=models.Security(
        auth2=models.SchemeAuth2(
            username="",
            password="",
        ),
    ),
) as sdk:

    sdk.auth2_hoisted()

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

## auth2_preferred

### Example Usage

<!-- UsageSnippet language="python" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `python
from multi_auth_sdk import SDK, models


with SDK(
    security=models.Security(
        auth2=models.SchemeAuth2(
            username="",
            password="",
        ),
    ),
) as sdk:

    sdk.auth2_preferred()

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

## op_level_client_credentials

### Example Usage

<!-- UsageSnippet language="python" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `python
from multi_auth_sdk import SDK, models


with SDK() as sdk:

    sdk.op_level_client_credentials(security=models.OpLevelClientCredentialsSecurity(
        client_id="<YOUR_CLIENT_ID_HERE>",
        client_secret="<YOUR_CLIENT_SECRET_HERE>",
    ))

    # Use the SDK ...

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                            | Type                                                                                 | Required                                                                             | Description                                                                          |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| ` + "`" + `security` + "`" + `                                                                           | [models.OpLevelClientCredentialsSecurity](../../oplevelclientcredentialssecurity.md) | :heavy_check_mark:                                                                   | The security requirements to use for the request.                                    |
| ` + "`" + `retries` + "`" + `                                                                            | [Optional[utils.RetryConfig]](../../models/utils/retryconfig.md)                     | :heavy_minus_sign:                                                                   | Configuration to override the default retry behavior of the client.                  |

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
