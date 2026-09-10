# SDK

## Overview

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

Speakeasy Docs
<https://speakeasy.com/docs>

### Available Operations

* [OperationWithLeadingAndTrailingUnderscores](#operationwithleadingandtrailingunderscores)
* [PostFile](#postfile) - Post File
* [GetPolymorphism](#getpolymorphism)
* [GetUnionErrors](#getunionerrors)
* [GetRequestBodyFlattenedAway](#getrequestbodyflattenedaway)
* [GetFullyFlattenedRequest](#getfullyflattenedrequest)
* [CreateWithUnion](#createwithunion) - Create with discriminated union request body
* [TestEndpoint](#testendpoint)
* [CreateUser](#createuser) - Create User
* [GetUser](#getuser) - Get User
* [UpdateUser](#updateuser) - Update User
* [DeleteUser](#deleteuser) - Delete User
* [Login](#login) - Login
* [Validate](#validate) - Validate
* [Chat](#chat)
* [GetBinaryDefaultResponse](#getbinarydefaultresponse)
* [TestEnumFormats](#testenumformats) - Test x-speakeasy-enums in different formats
* [BinaryAndStringUpload](#binaryandstringupload)
* [GetErrorInUnion](#geterrorinunion)
* [GetDuplicateExportCollision](#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [GetNamedPrimitiveUnion](#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [GetEmptyObjectError](#getemptyobjecterror) - Get Empty Object Error
* [URLValidationStressTest](#urlvalidationstresstest)
* [ParenthesesInPathAllowed](#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [GetNestedIntegerString](#getnestedintegerstring) - Test nested struct with integer:string tag
* [RenderAsset](#renderasset) - Render Asset
* [GetAsset](#getasset) - Get Asset
* [GetErrorOnlyExample](#geterroronlyexample) - Operation with example only on error response

## OperationWithLeadingAndTrailingUnderscores

### Example Usage

<!-- UsageSnippet language="go" operationID="_operation_with_leading_and_trailing_underscores_" method="get" path="/test_operation_id_with_underscores" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.OperationWithLeadingAndTrailingUnderscores(ctx, "renamed")
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                                                                                        | Type                                                                                                                                             | Required                                                                                                                                         | Description                                                                                                                                      | Example                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `ctx`                                                                                                                                            | [context.Context](https://pkg.go.dev/context#Context)                                                                                            | :heavy_check_mark:                                                                                                                               | The context to use for the request.                                                                                                              |                                                                                                                                                  |
| `qp1`                                                                                                                                            | `string`                                                                                                                                         | :heavy_check_mark:                                                                                                                               | This parameter will not be filled in with the queryParam1 global because it uses x-speakeasy-name-override which results in a non-matching name. | renamed                                                                                                                                          |
| `opts`                                                                                                                                           | [][examplealias.Option](../../option.md)                                                                                                         | :heavy_minus_sign:                                                                                                                               | The options for this request.                                                                                                                    |                                                                                                                                                  |

### Response

**[*OperationWithLeadingAndTrailingUnderscoresResponse](../../operationwithleadingandtrailingunderscoresresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## PostFile

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="go" operationID="postFile" method="post" path="/file" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"os"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    example, fileErr := os.Open("example.file")
    if fileErr != nil {
        panic(fileErr)
    }

    res, err := s.PostFile(ctx, examplealias.PostFileRequest{
        Upload: examplealias.File{
            FileName: "example.file",
            Content: example,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.File != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `request`                                             | [PostFileRequest](../../postfilerequest.md)           | :heavy_check_mark:                                    | The request object to use for the request.            |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*PostFileResponse](../../postfileresponse.md), error**

### Errors

| Error Type               | Status Code              | Content Type             |
| ------------------------ | ------------------------ | ------------------------ |
| examplealias.ErrorsError | 415, 4XX                 | application/json         |
| examplealias.ErrorsError | 5XX                      | application/json         |

## GetPolymorphism

### Example Usage

<!-- UsageSnippet language="go" operationID="getPolymorphism" method="get" path="/polymorphism" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetPolymorphism(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        switch res.Object.OneOfWithUnionDescription.Type {
            case examplealias.OneOfWithUnionDescriptionTypeExhaustiveObject:
                // res.Object.OneOfWithUnionDescription.ExhaustiveObject is populated
            case examplealias.OneOfWithUnionDescriptionTypeSimpleObject:
                // res.Object.OneOfWithUnionDescription.SimpleObject is populated
            default:
                // Unknown type - use res.Object.OneOfWithUnionDescription.GetUnknownRaw() for raw JSON
        }

    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetPolymorphismResponse](../../getpolymorphismresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetUnionErrors

### Example Usage

<!-- UsageSnippet language="go" operationID="getUnionErrors" method="get" path="/unionErrors" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetUnionErrors(ctx, 12)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        for {
            // handle items

            res, err = res.Next()

            if err != nil {
                // handle error
            }

            if res == nil {
                break
            }
        }
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           | Example                                               |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |                                                       |
| `page`                                                | `int64`                                               | :heavy_check_mark:                                    | N/A                                                   | 12                                                    |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |                                                       |

### Response

**[*GetUnionErrorsResponse](../../getunionerrorsresponse.md), error**

### Errors

| Error Type                                     | Status Code                                    | Content Type                                   |
| ---------------------------------------------- | ---------------------------------------------- | ---------------------------------------------- |
| examplealias.ErrorsError                       | 404                                            | application/json                               |
| examplealias.GetUnionErrorsInternalServerError | 500                                            | application/json                               |
| examplealias.ClientError                       | 4XX                                            | application/json                               |
| examplealias.SDKError                          | 5XX                                            | \*/\*                                          |

## GetRequestBodyFlattenedAway

### Example Usage

<!-- UsageSnippet language="go" operationID="getRequestBodyFlattenedAway" method="get" path="/requestBodyFlattenedAway" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New(
        examplealias.WithLoneQueryParam("<value>"),
    )

    res, err := s.GetRequestBodyFlattenedAway(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetRequestBodyFlattenedAwayResponse](../../getrequestbodyflattenedawayresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetFullyFlattenedRequest

### Example Usage

<!-- UsageSnippet language="go" operationID="getFullyFlattenedRequest" method="post" path="/fullyFlattenedRequest" example="namedExampleThatIsntMatchedAcrossDifferentExamples" -->
```go
package main

import(
	"context"
	"os"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New(
        examplealias.WithSecurity(examplealias.Security{
            Option6: &examplealias.SecurityOption6{
                ClientCredentials: os.Getenv("SPEAKEASY_CLIENT_CREDENTIALS"),
            },
        }),
    )

    res, err := s.GetFullyFlattenedRequest(ctx, "en", examplealias.GetFullyFlattenedRequestRequestBody{
        Name: "<value>",
    }, nil)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                           | Type                                                                                | Required                                                                            | Description                                                                         |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `ctx`                                                                               | [context.Context](https://pkg.go.dev/context#Context)                               | :heavy_check_mark:                                                                  | The context to use for the request.                                                 |
| `lang`                                                                              | `string`                                                                            | :heavy_check_mark:                                                                  | N/A                                                                                 |
| `requestBody`                                                                       | [GetFullyFlattenedRequestRequestBody](../../getfullyflattenedrequestrequestbody.md) | :heavy_check_mark:                                                                  | N/A                                                                                 |
| `maxLength`                                                                         | `*int64`                                                                            | :heavy_minus_sign:                                                                  | N/A                                                                                 |
| `opts`                                                                              | [][examplealias.Option](../../option.md)                                            | :heavy_minus_sign:                                                                  | The options for this request.                                                       |

### Response

**[*GetFullyFlattenedRequestResponse](../../getfullyflattenedrequestresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## CreateWithUnion

Test CLI generation for discriminated unions with dot-notation flags

### Example Usage

<!-- UsageSnippet language="go" operationID="createWithUnion" method="post" path="/unionRequestBody" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.CreateWithUnion(ctx, examplealias.ShapeRequest{
        Name: "<value>",
        Shape: examplealias.NewShape(
            examplealias.Rectangle{
                Type: "rectangle",
                Width: 3125.73,
                Height: 922.51,
            },
        ),
    }, nil)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        switch res.Object.Shape.Type {
            case examplealias.ShapeTypeCircle:
                // res.Object.Shape.Circle is populated
            case examplealias.ShapeTypeRectangle:
                // res.Object.Shape.Rectangle is populated
            default:
                // Unknown type - use res.Object.Shape.GetUnknownRaw() for raw JSON
        }

    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `shapeRequest`                                        | [ShapeRequest](../../shaperequest.md)                 | :heavy_check_mark:                                    | N/A                                                   |
| `dryRun`                                              | `*bool`                                               | :heavy_minus_sign:                                    | If true, validates without creating                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*CreateWithUnionResponse](../../createwithunionresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## TestEndpoint

### Example Usage

<!-- UsageSnippet language="go" operationID="testEndpoint" method="post" path="/test/endpoint/{testName}" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.TestEndpoint(ctx, "<value>", examplealias.TestEndpointRequestBody{
        Test: "<value>",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                   | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `ctx`                                                       | [context.Context](https://pkg.go.dev/context#Context)       | :heavy_check_mark:                                          | The context to use for the request.                         |
| `testName`                                                  | `string`                                                    | :heavy_check_mark:                                          | N/A                                                         |
| `requestBody`                                               | [TestEndpointRequestBody](../../testendpointrequestbody.md) | :heavy_check_mark:                                          | N/A                                                         |
| `opts`                                                      | [][examplealias.Option](../../option.md)                    | :heavy_minus_sign:                                          | The options for this request.                               |

### Response

**[*TestEndpointResponse](../../testendpointresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## CreateUser

Creates a new user in the system. Multiple named examples demonstrate
different pairing scenarios for documentation generation.


### Example Usage: paired-example

<!-- UsageSnippet language="go" operationID="createUser" method="put" path="/user" example="paired-example" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.CreateUser(ctx, examplealias.BaseUser{
        Email: "paired@example.com",
        FirstName: examplealias.Pointer("John"),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.User != nil {
        // handle response
    }
}
```
### Example Usage: request-only

<!-- UsageSnippet language="go" operationID="createUser" method="put" path="/user" example="request-only" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.CreateUser(ctx, examplealias.BaseUser{
        Email: "request-only@example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.User != nil {
        // handle response
    }
}
```
### Example Usage: response-only

<!-- UsageSnippet language="go" operationID="createUser" method="put" path="/user" example="response-only" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.CreateUser(ctx, examplealias.BaseUser{
        ID: examplealias.Pointer("8ffac18c-7d88-4879-b057-e5f45b9ce7de"),
        Email: "Virginie47@gmail.com",
        Gender: examplealias.GenderOther.ToPointer(),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.User != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `request`                                             | [BaseUser](../../baseuser.md)                         | :heavy_check_mark:                                    | The request object to use for the request.            |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*CreateUserResponse](../../createuserresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetUser

Get User

### Example Usage

<!-- UsageSnippet language="go" operationID="getUser" method="get" path="/user/{id}" example="success" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetUser(ctx, "<id>")
    if err != nil {
        log.Fatal(err)
    }
    if res.User != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `id`                                                  | `string`                                              | :heavy_check_mark:                                    | N/A                                                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetUserResponse](../../getuserresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## UpdateUser

Update User

### Example Usage

<!-- UsageSnippet language="go" operationID="updateUser" method="post" path="/user/{id}" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.UpdateUser(ctx, "<id>", examplealias.User{
        ID: "8ffac18c-7d88-4879-b057-e5f45b9ce7de",
        Email: "Joanny.Feeney@gmail.com",
        Gender: examplealias.GenderOther.ToPointer(),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.User != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `id`                                                  | `string`                                              | :heavy_check_mark:                                    | N/A                                                   |
| `user`                                                | [User](../../user.md)                                 | :heavy_check_mark:                                    | N/A                                                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*UpdateUserResponse](../../updateuserresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## DeleteUser

Delete User

### Example Usage

<!-- UsageSnippet language="go" operationID="deleteUser" method="delete" path="/user/{id}" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.DeleteUser(ctx, "<id>")
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `id`                                                  | `string`                                              | :heavy_check_mark:                                    | N/A                                                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*DeleteUserResponse](../../deleteuserresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## Login

Login

### Example Usage

<!-- UsageSnippet language="go" operationID="login" method="get" path="/auth/login" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.Login(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*LoginResponse](../../loginresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## Validate

Validate

### Example Usage

<!-- UsageSnippet language="go" operationID="validate" method="get" path="/auth/validate" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.Validate(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*ValidateResponse](../../validateresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## Chat

### Example Usage

<!-- UsageSnippet language="go" operationID="chat" method="post" path="/chat" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.Chat(ctx, examplealias.NewChatRequest(
        examplealias.ChatModelRequest{
            Prompt: "What is the largest city in the world?",
            Stream: examplealias.Pointer(false),
        },
    ))
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        defer res.ChatStream.Close()

        for res.ChatStream.Next() {
            event := res.ChatStream.Value()
            log.Print(event)
            // Handle the event
	      }
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `request`                                             | [ChatRequest](../../chatrequest.md)                   | :heavy_check_mark:                                    | The request object to use for the request.            |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*ChatResponse](../../chatresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetBinaryDefaultResponse

### Example Usage

<!-- UsageSnippet language="go" operationID="getBinaryDefaultResponse" method="get" path="/binaryDefaultResponse" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetBinaryDefaultResponse(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Bytes != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetBinaryDefaultResponseResponse](../../getbinarydefaultresponseresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## TestEnumFormats

This endpoint tests the x-speakeasy-enums extension in both array and map formats,
including partial map coverage and both string and integer enum types.

### Example Usage

<!-- UsageSnippet language="go" operationID="testEnumFormats" method="post" path="/enumFormats" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.TestEnumFormats(ctx, examplealias.TestEnumFormatsRequest{
        StringArrayFormat: examplealias.StringArrayFormatAwaitingReviewProcess,
        StringMapFormat: examplealias.StringMapFormatModerateImportanceLevel,
        StringPartialMapFormat: examplealias.StringPartialMapFormatInitialDraftVersion,
        IntegerMapFormat: examplealias.IntegerMapFormatSuccessfulOperationComplete,
        IntegerPartialMapFormat: examplealias.IntegerPartialMapFormatPrimaryFirstOption,
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                 | Type                                                      | Required                                                  | Description                                               |
| --------------------------------------------------------- | --------------------------------------------------------- | --------------------------------------------------------- | --------------------------------------------------------- |
| `ctx`                                                     | [context.Context](https://pkg.go.dev/context#Context)     | :heavy_check_mark:                                        | The context to use for the request.                       |
| `request`                                                 | [TestEnumFormatsRequest](../../testenumformatsrequest.md) | :heavy_check_mark:                                        | The request object to use for the request.                |
| `opts`                                                    | [][examplealias.Option](../../option.md)                  | :heavy_minus_sign:                                        | The options for this request.                             |

### Response

**[*TestEnumFormatsResponse](../../testenumformatsresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## BinaryAndStringUpload

### Example Usage

<!-- UsageSnippet language="go" operationID="binaryAndStringUpload" method="post" path="/binaryAndStringUpload" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"os"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    test, fileErr := os.ReadFile("test.json")
    if fileErr != nil {
        panic(fileErr)
    }

    res, err := s.BinaryAndStringUpload(ctx, examplealias.BinaryAndStringUploadRequest{
        Binary: test,
        String: string(test),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                             | Type                                                                  | Required                                                              | Description                                                           |
| --------------------------------------------------------------------- | --------------------------------------------------------------------- | --------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `ctx`                                                                 | [context.Context](https://pkg.go.dev/context#Context)                 | :heavy_check_mark:                                                    | The context to use for the request.                                   |
| `request`                                                             | [BinaryAndStringUploadRequest](../../binaryandstringuploadrequest.md) | :heavy_check_mark:                                                    | The request object to use for the request.                            |
| `opts`                                                                | [][examplealias.Option](../../option.md)                              | :heavy_minus_sign:                                                    | The options for this request.                                         |

### Response

**[*BinaryAndStringUploadResponse](../../binaryandstringuploadresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetErrorInUnion

### Example Usage

<!-- UsageSnippet language="go" operationID="getErrorInUnion" method="get" path="/errorInUnion" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetErrorInUnion(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetErrorInUnionResponse](../../geterrorinunionresponse.md), error**

### Errors

| Error Type                                      | Status Code                                     | Content Type                                    |
| ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- |
| examplealias.GetErrorInUnionInternalServerError | 500                                             | application/json                                |
| examplealias.SDKError                           | 4XX, 5XX                                        | \*/\*                                           |

## GetDuplicateExportCollision

Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308

### Example Usage

<!-- UsageSnippet language="go" operationID="getDuplicateExportCollision" method="get" path="/duplicateExportCollision" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetDuplicateExportCollision(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetDuplicateExportCollisionResponse](../../getduplicateexportcollisionresponse.md), error**

### Errors

| Error Type                       | Status Code                      | Content Type                     |
| -------------------------------- | -------------------------------- | -------------------------------- |
| examplealias.RequestTimeoutError | 408                              | application/json                 |
| examplealias.SDKError            | 4XX, 5XX                         | \*/\*                            |

## GetNamedPrimitiveUnion

Test named primitive union options using title and x-speakeasy-name-override

### Example Usage

<!-- UsageSnippet language="go" operationID="getNamedPrimitiveUnion" method="get" path="/namedPrimitiveUnion" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetNamedPrimitiveUnion(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.SomeUnion != nil {
        switch res.SomeUnion.Type {
            case examplealias.SomeUnionTypeMyString:
                // res.SomeUnion.MyString is populated
            case examplealias.SomeUnionTypeMyObject:
                // res.SomeUnion.MyObject is populated
            case examplealias.SomeUnionTypeFoo:
                // res.SomeUnion.Foo is populated
            default:
                // Unknown type - use res.SomeUnion.GetUnknownRaw() for raw JSON
        }

    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetNamedPrimitiveUnionResponse](../../getnamedprimitiveunionresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetEmptyObjectError

This endpoint tests the behavior when an error response has an empty object schema.

### Example Usage

<!-- UsageSnippet language="go" operationID="getEmptyObjectError" method="get" path="/emptyObjectError" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetEmptyObjectError(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetEmptyObjectErrorResponse](../../getemptyobjecterrorresponse.md), error**

### Errors

| Error Type                       | Status Code                      | Content Type                     |
| -------------------------------- | -------------------------------- | -------------------------------- |
| examplealias.FailedResponseError | 500                              | application/json                 |
| examplealias.SDKError            | 4XX, 5XX                         | \*/\*                            |

## URLValidationStressTest

### Example Usage

<!-- UsageSnippet language="go" operationID="urlValidationStressTest" method="get" path="/AZaz09-._~!$&'*+,;=:@/%20%25/-._~!$&'()*+,;=:@" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.URLValidationStressTest(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*URLValidationStressTestResponse](../../urlvalidationstresstestresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## ParenthesesInPathAllowed

A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".


### Example Usage

<!-- UsageSnippet language="go" operationID="parenthesesInPathAllowed" method="post" path="/jobs/job({id})" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.ParenthesesInPathAllowed(ctx, "<id>", examplealias.TemplateBracesTest{
        FieldWithBracesInDescription: examplealias.Pointer("A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n"),
        FieldWithBracesInTitle: examplealias.Pointer("A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n"),
        FieldWithBracesInExample: examplealias.Pointer("A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n"),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.TemplateBracesTest != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `id`                                                  | `string`                                              | :heavy_check_mark:                                    | N/A                                                   |
| `templateBracesTest`                                  | [TemplateBracesTest](../../templatebracestest.md)     | :heavy_check_mark:                                    | N/A                                                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*ParenthesesInPathAllowedResponse](../../parenthesesinpathallowedresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetNestedIntegerString

This endpoint tests the behavior when a deeply nested struct contains
an integer field that should be unmarshaled from a string.

### Example Usage

<!-- UsageSnippet language="go" operationID="getNestedIntegerString" method="get" path="/nestedIntegerString" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetNestedIntegerString(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.TaskResponse != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetNestedIntegerStringResponse](../../getnestedintegerstringresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## RenderAsset

Render a media asset from a text prompt.

### Example Usage

<!-- UsageSnippet language="go" operationID="renderAsset" method="post" path="/assets/render" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.RenderAsset(ctx, examplealias.RenderAssetRequest{
        Prompt: "<value>",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.AssetResult != nil {
        defer res.AssetStream.Close()

        for res.AssetStream.Next() {
            event := res.AssetStream.Value()
            log.Print(event)
            // Handle the event
	      }
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `request`                                             | [RenderAssetRequest](../../renderassetrequest.md)     | :heavy_check_mark:                                    | The request object to use for the request.            |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*RenderAssetResponse](../../renderassetresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetAsset

Get the current result of an asset job.

### Example Usage

<!-- UsageSnippet language="go" operationID="getAsset" method="get" path="/assets/{id}" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetAsset(ctx, "<id>", examplealias.Pointer(true))
    if err != nil {
        log.Fatal(err)
    }
    if res.AssetResult != nil {
        defer res.AssetStatusStream.Close()

        for res.AssetStatusStream.Next() {
            event := res.AssetStatusStream.Value()
            log.Print(event)
            // Handle the event
	      }
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `id`                                                  | `string`                                              | :heavy_check_mark:                                    | N/A                                                   |
| `stream_`                                             | `*bool`                                               | :heavy_minus_sign:                                    | N/A                                                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetAssetResponse](../../getassetresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetErrorOnlyExample

This endpoint tests that when an operation has a named example only on
an error response (not on the success response), we still generate
a default example for the success response.

### Example Usage

<!-- UsageSnippet language="go" operationID="getErrorOnlyExample" method="get" path="/errorOnlyExample" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.GetErrorOnlyExample(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Object != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*GetErrorOnlyExampleResponse](../../geterroronlyexampleresponse.md), error**

### Errors

| Error Type               | Status Code              | Content Type             |
| ------------------------ | ------------------------ | ------------------------ |
| examplealias.ErrorsError | 404                      | application/json         |
| examplealias.SDKError    | 4XX, 5XX                 | \*/\*                    |