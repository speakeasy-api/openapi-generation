package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Php(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `php:
  packageName: openapi/and-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [andAuthHoisted](#andauthhoisted)
* [opLevelAndAuth](#oplevelandauth)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="php" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            auth1: '<YOUR_API_KEY_HERE>',
            auth2: new Components\SchemeAuth2(
                username: '',
                password: '',
            ),
        )
    )
    ->build();



$response = $sdk->globalSecurity(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\GlobalSecurityResponse](../../Models/Operations/GlobalSecurityResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## andAuthHoisted

### Example Usage

<!-- UsageSnippet language="php" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            auth1: '<YOUR_API_KEY_HERE>',
            auth2: new Components\SchemeAuth2(
                username: '',
                password: '',
            ),
        )
    )
    ->build();



$response = $sdk->andAuthHoisted(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\AndAuthHoistedResponse](../../Models/Operations/AndAuthHoistedResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## opLevelAndAuth

### Example Usage

<!-- UsageSnippet language="php" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;
use OpenAPI\OpenAPI\Models\Operations;

$sdk = OpenAPI\SDK::builder()->build();


$requestSecurity = new Operations\OpLevelAndAuthSecurity(
    auth2: new Components\SchemeAuth2(
        username: '',
        password: '',
    ),
    auth3: '<YOUR_BEARER_TOKEN_HERE>',
);

$response = $sdk->opLevelAndAuth(
    security: $requestSecurity
);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                              | Type                                                                                   | Required                                                                               | Description                                                                            |
| -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                             | [Operations\OpLevelAndAuthSecurity](../../Models/Operations/OpLevelAndAuthSecurity.md) | :heavy_check_mark:                                                                     | The security requirements to use for the request.                                      |

### Response

**[?Operations\OpLevelAndAuthResponse](../../Models/Operations/OpLevelAndAuthResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
