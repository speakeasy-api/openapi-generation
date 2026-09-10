# Group

## Overview

### Available Operations

* [RootGroupOp](#rootgroupop) - An operation at the group's root level

## RootGroupOp

'group' differs from 'TestGroup' in that it not only contains subgroups,
but also an operation.


### Example Usage

<!-- UsageSnippet language="go" operationID="rootGroupOp" method="get" path="/group/root" -->
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

    res, err := s.Group.RootGroupOp(ctx)
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

**[*RootGroupOpResponse](../../rootgroupopresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |