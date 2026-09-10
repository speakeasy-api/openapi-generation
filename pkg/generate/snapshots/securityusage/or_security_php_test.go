package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Php(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `php:
  packageName: openapi/openapi
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [auth1Hoisted](#auth1hoisted)
* [auth2Hoisted](#auth2hoisted)
* [auth2Preferred](#auth2preferred)
* [opLevelClientCredentials](#oplevelclientcredentials)

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

## auth1Hoisted

### Example Usage

<!-- UsageSnippet language="php" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            auth1: '<YOUR_API_KEY_HERE>',
        )
    )
    ->build();



$response = $sdk->auth1Hoisted(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\Auth1HoistedResponse](../../Models/Operations/Auth1HoistedResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## auth2Hoisted

### Example Usage

<!-- UsageSnippet language="php" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            auth2: new Components\SchemeAuth2(
                username: '',
                password: '',
            ),
        )
    )
    ->build();



$response = $sdk->auth2Hoisted(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\Auth2HoistedResponse](../../Models/Operations/Auth2HoistedResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## auth2Preferred

### Example Usage

<!-- UsageSnippet language="php" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Components;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new Components\Security(
            auth2: new Components\SchemeAuth2(
                username: '',
                password: '',
            ),
        )
    )
    ->build();



$response = $sdk->auth2Preferred(

);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[?Operations\Auth2PreferredResponse](../../Models/Operations/Auth2PreferredResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| Errors\APIException | 4XX, 5XX            | \*/\*               |

## opLevelClientCredentials

### Example Usage

<!-- UsageSnippet language="php" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Models\Operations;

$sdk = OpenAPI\SDK::builder()->build();


$requestSecurity = new Operations\OpLevelClientCredentialsSecurity(
    clientID: '<YOUR_CLIENT_ID_HERE>',
    clientSecret: '<YOUR_CLIENT_SECRET_HERE>',
);

$response = $sdk->opLevelClientCredentials(
    security: $requestSecurity
);

if ($response->statusCode === 200) {
    // handle response
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                  | Type                                                                                                       | Required                                                                                                   | Description                                                                                                |
| ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                                 | [Operations\OpLevelClientCredentialsSecurity](../../Models/Operations/OpLevelClientCredentialsSecurity.md) | :heavy_check_mark:                                                                                         | The security requirements to use for the request.                                                          |

### Response

**[?Operations\OpLevelClientCredentialsResponse](../../Models/Operations/OpLevelClientCredentialsResponse.md)**

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
