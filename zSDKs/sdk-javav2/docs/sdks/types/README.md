# NamespaceTests.Types

## Overview

### Available Operations

* [getNamespaceTypes](#getnamespacetypes) - Get Namespace Types Test
* [getNamespaceAnimal](#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [getNamespaceVehicle](#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [getNamespaceOrganization](#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

## getNamespaceTypes

This endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,
non-discriminated unions, and models with nested inline schemas.

### Example Usage

<!-- UsageSnippet language="java" operationID="getNamespaceTypes" method="get" path="/namespaceTypes" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetNamespaceTypesResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNamespaceTypesResponse res = sdk.namespaceTests().types().getNamespaceTypes()
                .call();

        if (res.namespaceTypesTest().isPresent()) {
            System.out.println(res.namespaceTypesTest().get());
        }
    }
}
```

### Response

**[GetNamespaceTypesResponse](../../models/operations/GetNamespaceTypesResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getNamespaceAnimal

This endpoint tests a discriminated union type in a custom namespace.
The response can be either a foo.Dog or foo.Cat.

### Example Usage

<!-- UsageSnippet language="java" operationID="getNamespaceAnimal" method="get" path="/namespaceAnimal" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.foo.Animal;
import org.openapis.openapi.models.operations.GetNamespaceAnimalResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNamespaceAnimalResponse res = sdk.namespaceTests().types().getNamespaceAnimal()
                .call();

        if (res.animal().isPresent()) {
            Animal unionValue = res.animal().get();
            switch (unionValue.animalType()) {
                case "dog":
                    // Handle dog discriminator variant
                    break;
                case "cat":
                    // Handle cat discriminator variant
                    break;
                default:
                    // Handle unknown discriminator variant
            }
        }
    }
}
```

### Response

**[GetNamespaceAnimalResponse](../../models/operations/GetNamespaceAnimalResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getNamespaceVehicle

This endpoint tests a non-discriminated union type in a custom namespace.
The response can be either a bar.Car or bar.Bike.

### Example Usage

<!-- UsageSnippet language="java" operationID="getNamespaceVehicle" method="get" path="/namespaceVehicle" -->
```java
package hello.world;

import com.fasterxml.jackson.databind.JsonNode;
import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.bar.Bike;
import org.openapis.openapi.models.bar.Car;
import org.openapis.openapi.models.bar.Vehicle;
import org.openapis.openapi.models.operations.GetNamespaceVehicleResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNamespaceVehicleResponse res = sdk.namespaceTests().types().getNamespaceVehicle()
                .call();

        if (res.vehicle().isPresent()) {
            Vehicle unionValue = res.vehicle().get();
            if (unionValue.car().isPresent()) {
                Car carValue = unionValue.car().get();
                // Handle car variant
            } else if (unionValue.bike().isPresent()) {
                Bike bikeValue = unionValue.bike().get();
                // Handle bike variant
            } else if (unionValue.asJson().isPresent()) {
                JsonNode raw = unionValue.asJson().get();
                // Handle unknown variant fallback
            }
        }
    }
}
```

### Response

**[GetNamespaceVehicleResponse](../../models/operations/GetNamespaceVehicleResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getNamespaceOrganization

This endpoint tests nested inline object schemas in a custom namespace.
The organization model contains nested address and department types that
should inherit the foo namespace.

### Example Usage

<!-- UsageSnippet language="java" operationID="getNamespaceOrganization" method="get" path="/namespaceOrganization" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetNamespaceOrganizationResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNamespaceOrganizationResponse res = sdk.namespaceTests().types().getNamespaceOrganization()
                .call();

        if (res.organization().isPresent()) {
            System.out.println(res.organization().get());
        }
    }
}
```

### Response

**[GetNamespaceOrganizationResponse](../../models/operations/GetNamespaceOrganizationResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |