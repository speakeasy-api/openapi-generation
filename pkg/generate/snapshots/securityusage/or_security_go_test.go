package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Go(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `go:
  packageName: github.com/example/multi-auth-sdk
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

<!-- UsageSnippet language="go" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"github.com/example/multi-auth-sdk/models/components"
	multiauthsdk "github.com/example/multi-auth-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := multiauthsdk.New(
        multiauthsdk.WithSecurity(components.Security{
            Auth1: multiauthsdk.Pointer("<YOUR_API_KEY_HERE>"),
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

## Auth1Hoisted

### Example Usage

<!-- UsageSnippet language="go" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"github.com/example/multi-auth-sdk/models/components"
	multiauthsdk "github.com/example/multi-auth-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := multiauthsdk.New(
        multiauthsdk.WithSecurity(components.Security{
            Auth1: multiauthsdk.Pointer("<YOUR_API_KEY_HERE>"),
        }),
    )

    res, err := s.Auth1Hoisted(ctx)
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

**[*operations.Auth1HoistedResponse](../../models/operations/auth1hoistedresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## Auth2Hoisted

### Example Usage

<!-- UsageSnippet language="go" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"github.com/example/multi-auth-sdk/models/components"
	multiauthsdk "github.com/example/multi-auth-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := multiauthsdk.New(
        multiauthsdk.WithSecurity(components.Security{
            Auth2: &components.SchemeAuth2{
                Username: "",
                Password: "",
            },
        }),
    )

    res, err := s.Auth2Hoisted(ctx)
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

**[*operations.Auth2HoistedResponse](../../models/operations/auth2hoistedresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## Auth2Preferred

### Example Usage

<!-- UsageSnippet language="go" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"github.com/example/multi-auth-sdk/models/components"
	multiauthsdk "github.com/example/multi-auth-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := multiauthsdk.New(
        multiauthsdk.WithSecurity(components.Security{
            Auth2: &components.SchemeAuth2{
                Username: "",
                Password: "",
            },
        }),
    )

    res, err := s.Auth2Preferred(ctx)
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

**[*operations.Auth2PreferredResponse](../../models/operations/auth2preferredresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## OpLevelClientCredentials

### Example Usage

<!-- UsageSnippet language="go" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	multiauthsdk "github.com/example/multi-auth-sdk"
	"github.com/example/multi-auth-sdk/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := multiauthsdk.New()

    res, err := s.OpLevelClientCredentials(ctx, operations.OpLevelClientCredentialsSecurity{
        ClientID: "<YOUR_CLIENT_ID_HERE>",
        ClientSecret: "<YOUR_CLIENT_SECRET_HERE>",
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

| Parameter                                                                                                  | Type                                                                                                       | Required                                                                                                   | Description                                                                                                |
| ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| ` + "`" + `ctx` + "`" + `                                                                                                      | [context.Context](https://pkg.go.dev/context#Context)                                                      | :heavy_check_mark:                                                                                         | The context to use for the request.                                                                        |
| ` + "`" + `security` + "`" + `                                                                                                 | [operations.OpLevelClientCredentialsSecurity](../../models/operations/oplevelclientcredentialssecurity.md) | :heavy_check_mark:                                                                                         | The security requirements to use for the request.                                                          |
| ` + "`" + `opts` + "`" + `                                                                                                     | [][operations.Option](../../models/operations/option.md)                                                   | :heavy_minus_sign:                                                                                         | The options for this request.                                                                              |

### Response

**[*operations.OpLevelClientCredentialsResponse](../../models/operations/oplevelclientcredentialsresponse.md), error**

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
