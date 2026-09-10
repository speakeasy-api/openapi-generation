package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Go(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `go:
  packageName: andsecuritysdk
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

<!-- UsageSnippet language="go" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"andsecuritysdk/models/components"
	"andsecuritysdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := andsecuritysdk.New(
        andsecuritysdk.WithSecurity(components.Security{
            Auth1: andsecuritysdk.Pointer("<YOUR_API_KEY_HERE>"),
            Auth2: &components.SchemeAuth2{
                Username: "",
                Password: "",
            },
        }),
    )

    res, err := s.GlobalSecurity(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                | Type                                                     | Required                                                 | Description                                              |
| -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| ` + "`" + `ctx` + "`" + `                                                    | [context.Context](https://pkg.go.dev/context#Context)    | :heavy_check_mark:                                       | The context to use for the request.                      |
| ` + "`" + `opts` + "`" + `                                                   | [][operations.Option](../../models/operations/option.md) | :heavy_minus_sign:                                       | The options for this request.                            |

### Response

**[*operations.GlobalSecurityResponse](../../models/operations/globalsecurityresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## AndAuthHoisted

### Example Usage

<!-- UsageSnippet language="go" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"andsecuritysdk/models/components"
	"andsecuritysdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := andsecuritysdk.New(
        andsecuritysdk.WithSecurity(components.Security{
            Auth1: andsecuritysdk.Pointer("<YOUR_API_KEY_HERE>"),
            Auth2: &components.SchemeAuth2{
                Username: "",
                Password: "",
            },
        }),
    )

    res, err := s.AndAuthHoisted(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                | Type                                                     | Required                                                 | Description                                              |
| -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- | -------------------------------------------------------- |
| ` + "`" + `ctx` + "`" + `                                                    | [context.Context](https://pkg.go.dev/context#Context)    | :heavy_check_mark:                                       | The context to use for the request.                      |
| ` + "`" + `opts` + "`" + `                                                   | [][operations.Option](../../models/operations/option.md) | :heavy_minus_sign:                                       | The options for this request.                            |

### Response

**[*operations.AndAuthHoistedResponse](../../models/operations/andauthhoistedresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## OpLevelAndAuth

### Example Usage

<!-- UsageSnippet language="go" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"andsecuritysdk"
	"andsecuritysdk/models/components"
	"andsecuritysdk/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := andsecuritysdk.New()

    res, err := s.OpLevelAndAuth(ctx, operations.OpLevelAndAuthSecurity{
        Auth2: components.SchemeAuth2{
            Username: "",
            Password: "",
        },
        Auth3: "<YOUR_BEARER_TOKEN_HERE>",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                              | Type                                                                                   | Required                                                                               | Description                                                                            |
| -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| ` + "`" + `ctx` + "`" + `                                                                                  | [context.Context](https://pkg.go.dev/context#Context)                                  | :heavy_check_mark:                                                                     | The context to use for the request.                                                    |
| ` + "`" + `security` + "`" + `                                                                             | [operations.OpLevelAndAuthSecurity](../../models/operations/oplevelandauthsecurity.md) | :heavy_check_mark:                                                                     | The security requirements to use for the request.                                      |
| ` + "`" + `opts` + "`" + `                                                                                 | [][operations.Option](../../models/operations/option.md)                               | :heavy_minus_sign:                                                                     | The options for this request.                                                          |

### Response

**[*operations.OpLevelAndAuthResponse](../../models/operations/oplevelandauthresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
