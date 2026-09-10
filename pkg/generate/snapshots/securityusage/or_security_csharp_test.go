package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Csharp(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `csharp:
  packageName: MultiAuthSDK
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [GlobalSecurity](#globalsecurity)
* [Auth1Hoisted](#auth1hoisted)
* [Auth2Hoisted](#auth2hoisted)
* [Auth2Preferred](#auth2preferred)
* [OpLevelClientCredentials](#oplevelclientcredentials)

## GlobalSecurity

### Example Usage

<!-- UsageSnippet language="csharp" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MultiAuthSDK;
using MultiAuthSDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Auth1 = "<YOUR_API_KEY_HERE>",
});

var res = await sdk.GlobalSecurityAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[GlobalSecurityResponse](../../Models/Requests/GlobalSecurityResponse.md)**

### Errors

| Error Type                              | Status Code                             | Content Type                            |
| --------------------------------------- | --------------------------------------- | --------------------------------------- |
| MultiAuthSDK.Models.Errors.APIException | 4XX, 5XX                                | \*/\*                                   |

## Auth1Hoisted

### Example Usage

<!-- UsageSnippet language="csharp" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MultiAuthSDK;
using MultiAuthSDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Auth1 = "<YOUR_API_KEY_HERE>",
});

var res = await sdk.Auth1HoistedAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Auth1HoistedResponse](../../Models/Requests/Auth1HoistedResponse.md)**

### Errors

| Error Type                              | Status Code                             | Content Type                            |
| --------------------------------------- | --------------------------------------- | --------------------------------------- |
| MultiAuthSDK.Models.Errors.APIException | 4XX, 5XX                                | \*/\*                                   |

## Auth2Hoisted

### Example Usage

<!-- UsageSnippet language="csharp" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MultiAuthSDK;
using MultiAuthSDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Auth2 = new SchemeAuth2() {
        Username = "",
        Password = "",
    },
});

var res = await sdk.Auth2HoistedAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Auth2HoistedResponse](../../Models/Requests/Auth2HoistedResponse.md)**

### Errors

| Error Type                              | Status Code                             | Content Type                            |
| --------------------------------------- | --------------------------------------- | --------------------------------------- |
| MultiAuthSDK.Models.Errors.APIException | 4XX, 5XX                                | \*/\*                                   |

## Auth2Preferred

### Example Usage

<!-- UsageSnippet language="csharp" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MultiAuthSDK;
using MultiAuthSDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Auth2 = new SchemeAuth2() {
        Username = "",
        Password = "",
    },
});

var res = await sdk.Auth2PreferredAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Auth2PreferredResponse](../../Models/Requests/Auth2PreferredResponse.md)**

### Errors

| Error Type                              | Status Code                             | Content Type                            |
| --------------------------------------- | --------------------------------------- | --------------------------------------- |
| MultiAuthSDK.Models.Errors.APIException | 4XX, 5XX                                | \*/\*                                   |

## OpLevelClientCredentials

### Example Usage

<!-- UsageSnippet language="csharp" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MultiAuthSDK;
using MultiAuthSDK.Models.Requests;

var sdk = new SDK();

var res = await sdk.OpLevelClientCredentialsAsync(security: new OpLevelClientCredentialsSecurity() {
    ClientID = "<YOUR_CLIENT_ID_HERE>",
    ClientSecret = "<YOUR_CLIENT_SECRET_HERE>",
});

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                     | Type                                                                                          | Required                                                                                      | Description                                                                                   |
| --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                    | [OpLevelClientCredentialsSecurity](../../Models/Requests/OpLevelClientCredentialsSecurity.md) | :heavy_check_mark:                                                                            | The security requirements to use for the request.                                             |

### Response

**[OpLevelClientCredentialsResponse](../../Models/Requests/OpLevelClientCredentialsResponse.md)**

### Errors

| Error Type                              | Status Code                             | Content Type                            |
| --------------------------------------- | --------------------------------------- | --------------------------------------- |
| MultiAuthSDK.Models.Errors.APIException | 4XX, 5XX                                | \*/\*                                   |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
