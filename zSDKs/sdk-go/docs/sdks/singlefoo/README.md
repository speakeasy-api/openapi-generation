# NamespaceTests.SingleFoo

## Overview

### Available Operations

* [GetSingleNamespaceFooPet](#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [CreateSingleNamespaceFooPet](#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

## GetSingleNamespaceFooPet

This endpoint tests using a single component from the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="go" operationID="getSingleNamespaceFooPet" method="get" path="/singleNamespace/foo/pet" -->
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

    res, err := s.NamespaceTests.SingleFoo.GetSingleNamespaceFooPet(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Pet != nil {
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

**[*GetSingleNamespaceFooPetResponse](../../getsinglenamespacefoopetresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## CreateSingleNamespaceFooPet

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="go" operationID="createSingleNamespaceFooPet" method="post" path="/singleNamespace/foo/pet" -->
```go
package main

import(
	"context"
	examplealias "example.com/openapi-go-sdk"
	"example.com/openapi-go-sdk/foo"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New()

    res, err := s.NamespaceTests.SingleFoo.CreateSingleNamespaceFooPet(ctx, foo.Pet{
        ID: "pet-foo-123",
        Name: "Fluffy",
        Species: "cat",
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Pet != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                             | Type                                                  | Required                                              | Description                                           |
| ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| `ctx`                                                 | [context.Context](https://pkg.go.dev/context#Context) | :heavy_check_mark:                                    | The context to use for the request.                   |
| `request`                                             | [foo.Pet](../../foo/pet.md)                           | :heavy_check_mark:                                    | The request object to use for the request.            |
| `opts`                                                | [][examplealias.Option](../../option.md)              | :heavy_minus_sign:                                    | The options for this request.                         |

### Response

**[*CreateSingleNamespaceFooPetResponse](../../createsinglenamespacefoopetresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |