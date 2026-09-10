# NamespaceTests.SingleBar

## Overview

### Available Operations

* [GetSingleNamespaceBarPet](#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

## GetSingleNamespaceBarPet

This endpoint tests using a single component from the bar namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="go" operationID="getSingleNamespaceBarPet" method="get" path="/singleNamespace/bar/pet" -->
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

    res, err := s.NamespaceTests.SingleBar.GetSingleNamespaceBarPet(ctx)
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

**[*GetSingleNamespaceBarPetResponse](../../getsinglenamespacebarpetresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |