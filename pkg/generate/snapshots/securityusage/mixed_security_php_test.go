package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Php(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `php:
  packageName: openapi/mixed-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [option1Hoisted](#option1hoisted)
* [option2Hoisted](#option2hoisted)
* [option1NotAllowed](#option1notallowed)
* [opLevelMixedAuth](#oplevelmixedauth)

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
            option1: new Components\SecurityOption1(
                authA1: '<YOUR_API_KEY_HERE>',
                authA2: new Components\SchemeAuthA2(
                    username: '',
                    password: '',
                ),
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

## option1Hoisted

### Example Usage

<!-- UsageSnippet language="php" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            option1: new Components\SecurityOption1(
                authA1: '<YOUR_API_KEY_HERE>',
                authA2: new Components\SchemeAuthA2(
                    username: '',
                    password: '',
                ),
            ),
        )
    )
    ->build();



$response = $sdk->option1Hoisted(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\Option1HoistedResponse](../../Models/Operations/Option1HoistedResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## option2Hoisted

### Example Usage

<!-- UsageSnippet language="php" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            option2: new Components\SecurityOption2(
                authB: '<YOUR_JWT>',
            ),
        )
    )
    ->build();



$response = $sdk->option2Hoisted(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\Option2HoistedResponse](../../Models/Operations/Option2HoistedResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## option1NotAllowed

### Example Usage

<!-- UsageSnippet language="php" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            option3: new Components\SecurityOption3(
                authC: '<YOUR_AUTH_C_HERE>',
            ),
        )
    )
    ->build();



$response = $sdk->option1NotAllowed(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\Option1NotAllowedResponse](../../Models/Operations/Option1NotAllowedResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## opLevelMixedAuth

### Example Usage

<!-- UsageSnippet language="php" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Operations;

$sdk = OpenAPI\SDK::builder()->build();


$requestSecurity = new Operations\OpLevelMixedAuthSecurity(
    option1: new Operations\OpLevelMixedAuthSecurityOption1(
        authA1: '<YOUR_API_KEY_HERE>',
        authB: '<YOUR_JWT>',
    ),
);

$response = $sdk->opLevelMixedAuth(
    security: $requestSecurity
);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                  | Type                                                                                       | Required                                                                                   | Description                                                                                |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ |
| ` + "`" + `security` + "`" + `                                                                                 | [Operations\OpLevelMixedAuthSecurity](../../Models/Operations/OpLevelMixedAuthSecurity.md) | :heavy_check_mark:                                                                         | The security requirements to use for the request.                                          |

### Response

**[?Operations\OpLevelMixedAuthResponse](../../Models/Operations/OpLevelMixedAuthResponse.md)**

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
