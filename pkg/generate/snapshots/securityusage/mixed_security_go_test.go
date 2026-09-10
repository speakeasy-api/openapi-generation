package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Go(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `go:
  packageName: mixedsecuritysdk
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

<!-- UsageSnippet language="go" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"mixedsecuritysdk/models/components"
	"mixedsecuritysdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := mixedsecuritysdk.New(
        mixedsecuritysdk.WithSecurity(components.Security{
            Option1: &components.SecurityOption1{
                AuthA1: "<YOUR_API_KEY_HERE>",
                AuthA2: components.SchemeAuthA2{
                    Username: "",
                    Password: "",
                },
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

## Option1Hoisted

### Example Usage

<!-- UsageSnippet language="go" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"mixedsecuritysdk/models/components"
	"mixedsecuritysdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := mixedsecuritysdk.New(
        mixedsecuritysdk.WithSecurity(components.Security{
            Option1: &components.SecurityOption1{
                AuthA1: "<YOUR_API_KEY_HERE>",
                AuthA2: components.SchemeAuthA2{
                    Username: "",
                    Password: "",
                },
            },
        }),
    )

    res, err := s.Option1Hoisted(ctx)
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

**[*operations.Option1HoistedResponse](../../models/operations/option1hoistedresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## Option2Hoisted

### Example Usage

<!-- UsageSnippet language="go" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"mixedsecuritysdk/models/components"
	"mixedsecuritysdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := mixedsecuritysdk.New(
        mixedsecuritysdk.WithSecurity(components.Security{
            Option2: &components.SecurityOption2{
                AuthB: "<YOUR_JWT>",
            },
        }),
    )

    res, err := s.Option2Hoisted(ctx)
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

**[*operations.Option2HoistedResponse](../../models/operations/option2hoistedresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## Option1NotAllowed

### Example Usage

<!-- UsageSnippet language="go" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"mixedsecuritysdk/models/components"
	"mixedsecuritysdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := mixedsecuritysdk.New(
        mixedsecuritysdk.WithSecurity(components.Security{
            Option3: &components.SecurityOption3{
                AuthC: "<YOUR_AUTH_C_HERE>",
            },
        }),
    )

    res, err := s.Option1NotAllowed(ctx)
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

**[*operations.Option1NotAllowedResponse](../../models/operations/option1notallowedresponse.md), error**

### Errors

| Error Type         | Status Code        | Content Type       |
| ------------------ | ------------------ | ------------------ |
| apierrors.APIError | 4XX, 5XX           | \*/\*              |

## OpLevelMixedAuth

### Example Usage

<!-- UsageSnippet language="go" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `go
package main

import(
	"context"
	"mixedsecuritysdk"
	"mixedsecuritysdk/models/operations"
	"log"
)

func main() {
    ctx := context.Background()

    s := mixedsecuritysdk.New()

    res, err := s.OpLevelMixedAuth(ctx, operations.OpLevelMixedAuthSecurity{
        Option1: &operations.OpLevelMixedAuthSecurityOption1{
            AuthA1: "<YOUR_API_KEY_HERE>",
            AuthB: "<YOUR_JWT>",
        },
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

| Parameter                                                                                  | Type                                                                                       | Required                                                                                   | Description                                                                                |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ |
| ` + "`" + `ctx` + "`" + `                                                                                      | [context.Context](https://pkg.go.dev/context#Context)                                      | :heavy_check_mark:                                                                         | The context to use for the request.                                                        |
| ` + "`" + `security` + "`" + `                                                                                 | [operations.OpLevelMixedAuthSecurity](../../models/operations/oplevelmixedauthsecurity.md) | :heavy_check_mark:                                                                         | The security requirements to use for the request.                                          |
| ` + "`" + `opts` + "`" + `                                                                                     | [][operations.Option](../../models/operations/option.md)                                   | :heavy_minus_sign:                                                                         | The options for this request.                                                              |

### Response

**[*operations.OpLevelMixedAuthResponse](../../models/operations/oplevelmixedauthresponse.md), error**

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
