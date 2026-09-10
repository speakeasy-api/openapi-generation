package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestGlobalOptional_Py(t *testing.T) {
	t.Parallel()

	spec := specGlobalOptional

	genYaml := `python:
  packageName: global-optional-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [global_security_optional](#global_security_optional)
* [hoisted_optional](#hoisted_optional)
* [op_level_required](#op_level_required)

## global_security_optional

### Example Usage

<!-- UsageSnippet language="python" operationID="globalSecurityOptional" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `python
from global_optional_sdk import SDK


with SDK() as sdk:

    sdk.global_security_optional()

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

## hoisted_optional

### Example Usage

<!-- UsageSnippet language="python" operationID="hoistedOptional" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `python
from global_optional_sdk import SDK, models


with SDK(
    security=models.Security(
        auth_b="<YOUR_JWT>",
    ),
) as sdk:

    sdk.hoisted_optional()

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

## op_level_required

### Example Usage

<!-- UsageSnippet language="python" operationID="opLevelRequired" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `python
from global_optional_sdk import SDK, models


with SDK(
    security=models.Security(
        auth_b="<YOUR_JWT>",
    ),
) as sdk:

    sdk.op_level_required()

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
