# NamespaceTests.Types

## Overview

### Available Operations

* [GetNamespaceTypes](#getnamespacetypes) - Get Namespace Types Test
* [GetNamespaceAnimal](#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [GetNamespaceVehicle](#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [GetNamespaceOrganization](#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

## GetNamespaceTypes

This endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,
non-discriminated unions, and models with nested inline schemas.

### Example Usage

<!-- UsageSnippet language="go" operationID="getNamespaceTypes" method="get" path="/namespaceTypes" -->
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

    res, err := s.NamespaceTests.Types.GetNamespaceTypes(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.NamespaceTypesTest != nil {
        switch res.NamespaceTypesTest.FooAnimal.Type {
            case examplealias.AnimalTypeDog:
                // res.NamespaceTypesTest.FooAnimal.Dog is populated
            case examplealias.AnimalTypeCat:
                // res.NamespaceTypesTest.FooAnimal.Cat is populated
            default:
                // Unknown type - use res.NamespaceTypesTest.FooAnimal.GetUnknownRaw() for raw JSON
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

**[*GetNamespaceTypesResponse](../../getnamespacetypesresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetNamespaceAnimal

This endpoint tests a discriminated union type in a custom namespace.
The response can be either a foo.Dog or foo.Cat.

### Example Usage

<!-- UsageSnippet language="go" operationID="getNamespaceAnimal" method="get" path="/namespaceAnimal" -->
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

    res, err := s.NamespaceTests.Types.GetNamespaceAnimal(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Animal != nil {
        switch res.Animal.Type {
            case examplealias.AnimalTypeDog:
                // res.Animal.Dog is populated
            case examplealias.AnimalTypeCat:
                // res.Animal.Cat is populated
            default:
                // Unknown type - use res.Animal.GetUnknownRaw() for raw JSON
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

**[*GetNamespaceAnimalResponse](../../getnamespaceanimalresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetNamespaceVehicle

This endpoint tests a non-discriminated union type in a custom namespace.
The response can be either a bar.Car or bar.Bike.

### Example Usage

<!-- UsageSnippet language="go" operationID="getNamespaceVehicle" method="get" path="/namespaceVehicle" -->
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

    res, err := s.NamespaceTests.Types.GetNamespaceVehicle(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Vehicle != nil {
        switch res.Vehicle.Type {
            case examplealias.VehicleTypeCar:
                // res.Vehicle.Car is populated
            case examplealias.VehicleTypeBike:
                // res.Vehicle.Bike is populated
            default:
                // Unknown type - use res.Vehicle.GetUnknownRaw() for raw JSON
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

**[*GetNamespaceVehicleResponse](../../getnamespacevehicleresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |

## GetNamespaceOrganization

This endpoint tests nested inline object schemas in a custom namespace.
The organization model contains nested address and department types that
should inherit the foo namespace.

### Example Usage

<!-- UsageSnippet language="go" operationID="getNamespaceOrganization" method="get" path="/namespaceOrganization" -->
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

    res, err := s.NamespaceTests.Types.GetNamespaceOrganization(ctx)
    if err != nil {
        log.Fatal(err)
    }
    if res.Organization != nil {
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

**[*GetNamespaceOrganizationResponse](../../getnamespaceorganizationresponse.md), error**

### Errors

| Error Type            | Status Code           | Content Type          |
| --------------------- | --------------------- | --------------------- |
| examplealias.SDKError | 4XX, 5XX              | \*/\*                 |