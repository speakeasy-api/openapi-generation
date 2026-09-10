package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestNoGlobalWithOp_Py(t *testing.T) {
	t.Parallel()

	spec := specNoGlobalWithOp

	genYaml := `python:
  packageName: no-global-auth-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [global_security_hoisted](#global_security_hoisted)
* [most_common_scheme_hoisted](#most_common_scheme_hoisted)
* [op_level_security](#op_level_security)

## global_security_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="globalSecurityHoisted" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `python
from no_global_auth_sdk import SDK


with SDK(
    auth_a="<YOUR_API_KEY_HERE>",
) as sdk:

    sdk.global_security_hoisted()

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

## most_common_scheme_hoisted

### Example Usage

<!-- UsageSnippet language="python" operationID="mostCommonSchemeHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `python
from no_global_auth_sdk import SDK


with SDK(
    auth_a="<YOUR_API_KEY_HERE>",
) as sdk:

    sdk.most_common_scheme_hoisted()

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

## op_level_security

### Example Usage

<!-- UsageSnippet language="python" operationID="opLevelSecurity" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `python
from no_global_auth_sdk import SDK, models


with SDK() as sdk:

    sdk.op_level_security(security=models.OpLevelSecuritySecurity(
        auth_a="<YOUR_API_KEY_HERE>",
    ))

    # Use the SDK ...

` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                           | Type                                                                | Required                                                            | Description                                                         |
| ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                          | [models.OpLevelSecuritySecurity](../../oplevelsecuritysecurity.md)  | :heavy_check_mark:                                                  | The security requirements to use for the request.                   |
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
