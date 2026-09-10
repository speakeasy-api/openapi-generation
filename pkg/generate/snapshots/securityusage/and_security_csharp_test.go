package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Csharp(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `csharp:
  packageName: AndSecuritySDK
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [GlobalSecurity](#globalsecurity)
* [AndAuthHoisted](#andauthhoisted)
* [OpLevelAndAuth](#oplevelandauth)

## GlobalSecurity

### Example Usage

<!-- UsageSnippet language="csharp" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using AndSecuritySDK;
using AndSecuritySDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Auth1 = "<YOUR_API_KEY_HERE>",
    Auth2 = new SchemeAuth2() {
        Username = "",
        Password = "",
    },
});

var res = await sdk.GlobalSecurityAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[GlobalSecurityResponse](../../Models/Requests/GlobalSecurityResponse.md)**

### Errors

| Error Type                                | Status Code                               | Content Type                              |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| AndSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                  | \*/\*                                     |

## AndAuthHoisted

### Example Usage

<!-- UsageSnippet language="csharp" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using AndSecuritySDK;
using AndSecuritySDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Auth1 = "<YOUR_API_KEY_HERE>",
    Auth2 = new SchemeAuth2() {
        Username = "",
        Password = "",
    },
});

var res = await sdk.AndAuthHoistedAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[AndAuthHoistedResponse](../../Models/Requests/AndAuthHoistedResponse.md)**

### Errors

| Error Type                                | Status Code                               | Content Type                              |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| AndSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                  | \*/\*                                     |

## OpLevelAndAuth

### Example Usage

<!-- UsageSnippet language="csharp" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using AndSecuritySDK;
using AndSecuritySDK.Models.Components;
using AndSecuritySDK.Models.Requests;

var sdk = new SDK();

var res = await sdk.OpLevelAndAuthAsync(security: new OpLevelAndAuthSecurity() {
    Auth2 = new SchemeAuth2() {
        Username = "",
        Password = "",
    },
    Auth3 = "<YOUR_BEARER_TOKEN_HERE>",
});

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                 | Type                                                                      | Required                                                                  | Description                                                               |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                | [OpLevelAndAuthSecurity](../../Models/Requests/OpLevelAndAuthSecurity.md) | :heavy_check_mark:                                                        | The security requirements to use for the request.                         |

### Response

**[OpLevelAndAuthResponse](../../Models/Requests/OpLevelAndAuthResponse.md)**

### Errors

| Error Type                                | Status Code                               | Content Type                              |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| AndSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                  | \*/\*                                     |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
