package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOpOptional_Py(t *testing.T) {
	t.Parallel()

	spec := specOpOptional

	genYaml := `python:
  packageName: op-optional-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [global_security_required](#global_security_required)
* [op_level_optional](#op_level_optional)
* [disable_auth_with_empty_object](#disable_auth_with_empty_object)
* [disable_auth_with_empty_array](#disable_auth_with_empty_array)

## global_security_required

### Example Usage

<!-- UsageSnippet language="python" operationID="globalSecurityRequired" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `python
from op_optional_sdk import SDK


with SDK(
    auth1="<YOUR_API_KEY_HERE>",
) as sdk:

    sdk.global_security_required()

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

## op_level_optional

### Example Usage

<!-- UsageSnippet language="python" operationID="opLevelOptional" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `python
from op_optional_sdk import SDK


with SDK(
    auth1="<YOUR_API_KEY_HERE>",
) as sdk:

    sdk.op_level_optional()

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

## disable_auth_with_empty_object

### Example Usage

<!-- UsageSnippet language="python" operationID="disableAuthWithEmptyObject" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `python
from op_optional_sdk import SDK


with SDK() as sdk:

    sdk.disable_auth_with_empty_object()

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

## disable_auth_with_empty_array

### Example Usage

<!-- UsageSnippet language="python" operationID="disableAuthWithEmptyArray" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `python
from op_optional_sdk import SDK


with SDK() as sdk:

    sdk.disable_auth_with_empty_array()

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

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
