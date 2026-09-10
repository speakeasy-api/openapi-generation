package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Csharp(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `csharp:
  packageName: MixedSecuritySDK
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [GlobalSecurity](#globalsecurity)
* [Option1Hoisted](#option1hoisted)
* [Option2Hoisted](#option2hoisted)
* [Option1NotAllowed](#option1notallowed)
* [OpLevelMixedAuth](#oplevelmixedauth)

## GlobalSecurity

### Example Usage

<!-- UsageSnippet language="csharp" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MixedSecuritySDK;
using MixedSecuritySDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Option1 = new SecurityOption1() {
        AuthA1 = "<YOUR_API_KEY_HERE>",
        AuthA2 = new SchemeAuthA2() {
            Username = "",
            Password = "",
        },
    },
});

var res = await sdk.GlobalSecurityAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[GlobalSecurityResponse](../../Models/Requests/GlobalSecurityResponse.md)**

### Errors

| Error Type                                  | Status Code                                 | Content Type                                |
| ------------------------------------------- | ------------------------------------------- | ------------------------------------------- |
| MixedSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                    | \*/\*                                       |

## Option1Hoisted

### Example Usage

<!-- UsageSnippet language="csharp" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MixedSecuritySDK;
using MixedSecuritySDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Option1 = new SecurityOption1() {
        AuthA1 = "<YOUR_API_KEY_HERE>",
        AuthA2 = new SchemeAuthA2() {
            Username = "",
            Password = "",
        },
    },
});

var res = await sdk.Option1HoistedAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Option1HoistedResponse](../../Models/Requests/Option1HoistedResponse.md)**

### Errors

| Error Type                                  | Status Code                                 | Content Type                                |
| ------------------------------------------- | ------------------------------------------- | ------------------------------------------- |
| MixedSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                    | \*/\*                                       |

## Option2Hoisted

### Example Usage

<!-- UsageSnippet language="csharp" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MixedSecuritySDK;
using MixedSecuritySDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Option2 = new SecurityOption2() {
        AuthB = "<YOUR_JWT>",
    },
});

var res = await sdk.Option2HoistedAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Option2HoistedResponse](../../Models/Requests/Option2HoistedResponse.md)**

### Errors

| Error Type                                  | Status Code                                 | Content Type                                |
| ------------------------------------------- | ------------------------------------------- | ------------------------------------------- |
| MixedSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                    | \*/\*                                       |

## Option1NotAllowed

### Example Usage

<!-- UsageSnippet language="csharp" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MixedSecuritySDK;
using MixedSecuritySDK.Models.Components;

var sdk = new SDK(security: new Security() {
    Option3 = new SecurityOption3() {
        AuthC = "<YOUR_AUTH_C_HERE>",
    },
});

var res = await sdk.Option1NotAllowedAsync();

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Option1NotAllowedResponse](../../Models/Requests/Option1NotAllowedResponse.md)**

### Errors

| Error Type                                  | Status Code                                 | Content Type                                |
| ------------------------------------------- | ------------------------------------------- | ------------------------------------------- |
| MixedSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                    | \*/\*                                       |

## OpLevelMixedAuth

### Example Usage

<!-- UsageSnippet language="csharp" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `csharp
using MixedSecuritySDK;
using MixedSecuritySDK.Models.Requests;

var sdk = new SDK();

var res = await sdk.OpLevelMixedAuthAsync(security: new OpLevelMixedAuthSecurity() {
    Option1 = new OpLevelMixedAuthSecurityOption1() {
        AuthA1 = "<YOUR_API_KEY_HERE>",
        AuthB = "<YOUR_JWT>",
    },
});

// handle response
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                     | Type                                                                          | Required                                                                      | Description                                                                   |
| ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                    | [OpLevelMixedAuthSecurity](../../Models/Requests/OpLevelMixedAuthSecurity.md) | :heavy_check_mark:                                                            | The security requirements to use for the request.                             |

### Response

**[OpLevelMixedAuthResponse](../../Models/Requests/OpLevelMixedAuthResponse.md)**

### Errors

| Error Type                                  | Status Code                                 | Content Type                                |
| ------------------------------------------- | ------------------------------------------- | ------------------------------------------- |
| MixedSecuritySDK.Models.Errors.APIException | 4XX, 5XX                                    | \*/\*                                       |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
