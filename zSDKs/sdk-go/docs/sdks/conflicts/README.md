# NamespaceTests.Conflicts

## Overview

### Available Operations

* [GetNamespaceConflict](#getnamespaceconflict) - Get Namespace Conflict Test
* [PutNamespaceConflict](#putnamespaceconflict) - Put Property Name Conflicts Behind
* [CreateNamespaceConflict](#createnamespaceconflict) - Create Namespace Conflict Test
* [GetTripleNamespaceConflict](#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
* [GetPetOwners](#getpetowners) - Get Pet Owners

## GetNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references two different Pet types from different namespaces.
The SDK should properly import and alias both Pet types.

### Example Usage

<!-- UsageSnippet language="go" operationID="getNamespaceConflict" method="get" path="/namespaceConflict" -->
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

    res, err := s.NamespaceTests.Conflicts.GetNamespaceConflict(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.NamespaceConflictTest != nil {
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

**[*GetNamespaceConflictResponse](../../getnamespaceconflictresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## PutNamespaceConflict

This endpoint tests property name conflict resolution through
x-speakeasy-name-override and x-speakeasy-model-namespace extensions.

### Example Usage

<!-- UsageSnippet language="go" operationID="putNamespaceConflict" method="put" path="/namespaceConflict" -->
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

    res, err := s.NamespaceTests.Conflicts.PutNamespaceConflict(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    if res != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                     | Type                                                          | Required                                                      | Description                                                   |
| ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- |
| `ctx`                                                         | [context.Context](https://pkg.go.dev/context#Context)         | :heavy_check_mark:                                            | The context to use for the request.                           |
| `request`                                                     | [ObjWithRenamedProperties](../../objwithrenamedproperties.md) | :heavy_check_mark:                                            | The request object to use for the request.                    |
| `opts`                                                        | [][examplealias.Option](../../option.md)                      | :heavy_minus_sign:                                            | The options for this request.                                 |

### Response

**[*PutNamespaceConflictResponse](../../putnamespaceconflictresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## CreateNamespaceConflict

This endpoint tests creating with models from different namespaces.
Uses foo.Pet in the request and bar.Pet in the response.

### Example Usage

<!-- UsageSnippet language="go" operationID="createNamespaceConflict" method="post" path="/namespaceConflict" -->
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

    res, err := s.NamespaceTests.Conflicts.CreateNamespaceConflict(ctx, foo.Pet{
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

**[*CreateNamespaceConflictResponse](../../createnamespaceconflictresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetTripleNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references three different Pet types from different namespaces.
The SDK should properly import and alias all three Pet types.

### Example Usage

<!-- UsageSnippet language="go" operationID="getTripleNamespaceConflict" method="get" path="/tripleNamespaceConflict" -->
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

    res, err := s.NamespaceTests.Conflicts.GetTripleNamespaceConflict(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.TripleNamespaceConflictTest != nil {
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

**[*GetTripleNamespaceConflictResponse](../../gettriplenamespaceconflictresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetPetOwners

This endpoint tests using PetOwner models from different namespaces.
Returns both foo.PetOwner and bar.PetOwner in the response.

### Example Usage

<!-- UsageSnippet language="go" operationID="getPetOwners" method="get" path="/petOwners" -->
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

    res, err := s.NamespaceTests.Conflicts.GetPetOwners(ctx)
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

**[*GetPetOwnersResponse](../../getpetownersresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |