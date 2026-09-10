# NamespaceTests.SingleFoo

## Overview

### Available Operations

* [getSingleNamespaceFooPet](#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [createSingleNamespaceFooPet](#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

## getSingleNamespaceFooPet

This endpoint tests using a single component from the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="java" operationID="getSingleNamespaceFooPet" method="get" path="/singleNamespace/foo/pet" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetSingleNamespaceFooPetResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetSingleNamespaceFooPetResponse res = sdk.namespaceTests().singleFoo().getSingleNamespaceFooPet()
                .call();

        if (res.pet().isPresent()) {
            System.out.println(res.pet().get());
        }
    }
}
```

### Response

**[GetSingleNamespaceFooPetResponse](../../models/operations/GetSingleNamespaceFooPetResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## createSingleNamespaceFooPet

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="java" operationID="createSingleNamespaceFooPet" method="post" path="/singleNamespace/foo/pet" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.foo.Pet;
import org.openapis.openapi.models.operations.CreateSingleNamespaceFooPetResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        Pet req = Pet.builder()
                .id("pet-foo-123")
                .name("Fluffy")
                .species("cat")
                .build();

        CreateSingleNamespaceFooPetResponse res = sdk.namespaceTests().singleFoo().createSingleNamespaceFooPet()
                .request(req)
                .call();

        if (res.pet().isPresent()) {
            System.out.println(res.pet().get());
        }
    }
}
```

### Parameters

| Parameter                                  | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `request`                                  | [Pet](../../models/shared/Pet.md)          | :heavy_check_mark:                         | The request object to use for the request. |

### Response

**[CreateSingleNamespaceFooPetResponse](../../models/operations/CreateSingleNamespaceFooPetResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |