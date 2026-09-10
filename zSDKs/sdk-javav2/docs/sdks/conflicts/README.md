# NamespaceTests.Conflicts

## Overview

### Available Operations

* [getNamespaceConflict](#getnamespaceconflict) - Get Namespace Conflict Test
* [putNamespaceConflict](#putnamespaceconflict) - Put Property Name Conflicts Behind
* [createNamespaceConflict](#createnamespaceconflict) - Create Namespace Conflict Test
* [getTripleNamespaceConflict](#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
* [getPetOwners](#getpetowners) - Get Pet Owners

## getNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references two different Pet types from different namespaces.
The SDK should properly import and alias both Pet types.

### Example Usage

<!-- UsageSnippet language="java" operationID="getNamespaceConflict" method="get" path="/namespaceConflict" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetNamespaceConflictResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetNamespaceConflictResponse res = sdk.namespaceTests().conflicts().getNamespaceConflict()
                .call();

        if (res.namespaceConflictTest().isPresent()) {
            System.out.println(res.namespaceConflictTest().get());
        }
    }
}
```

### Response

**[GetNamespaceConflictResponse](../../models/operations/GetNamespaceConflictResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## putNamespaceConflict

This endpoint tests property name conflict resolution through
x-speakeasy-name-override and x-speakeasy-model-namespace extensions.

### Example Usage

<!-- UsageSnippet language="java" operationID="putNamespaceConflict" method="put" path="/namespaceConflict" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.PutNamespaceConflictResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        PutNamespaceConflictResponse res = sdk.namespaceTests().conflicts().putNamespaceConflict()
                .call();

        // handle response
    }
}
```

### Parameters

| Parameter                                                                   | Type                                                                        | Required                                                                    | Description                                                                 |
| --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `request`                                                                   | [ObjWithRenamedProperties](../../models/shared/ObjWithRenamedProperties.md) | :heavy_check_mark:                                                          | The request object to use for the request.                                  |

### Response

**[PutNamespaceConflictResponse](../../models/operations/PutNamespaceConflictResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## createNamespaceConflict

This endpoint tests creating with models from different namespaces.
Uses foo.Pet in the request and bar.Pet in the response.

### Example Usage

<!-- UsageSnippet language="java" operationID="createNamespaceConflict" method="post" path="/namespaceConflict" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.foo.Pet;
import org.openapis.openapi.models.operations.CreateNamespaceConflictResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        Pet req = Pet.builder()
                .id("pet-foo-123")
                .name("Fluffy")
                .species("cat")
                .build();

        CreateNamespaceConflictResponse res = sdk.namespaceTests().conflicts().createNamespaceConflict()
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

**[CreateNamespaceConflictResponse](../../models/operations/CreateNamespaceConflictResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getTripleNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references three different Pet types from different namespaces.
The SDK should properly import and alias all three Pet types.

### Example Usage

<!-- UsageSnippet language="java" operationID="getTripleNamespaceConflict" method="get" path="/tripleNamespaceConflict" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetTripleNamespaceConflictResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetTripleNamespaceConflictResponse res = sdk.namespaceTests().conflicts().getTripleNamespaceConflict()
                .call();

        if (res.tripleNamespaceConflictTest().isPresent()) {
            System.out.println(res.tripleNamespaceConflictTest().get());
        }
    }
}
```

### Response

**[GetTripleNamespaceConflictResponse](../../models/operations/GetTripleNamespaceConflictResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |

## getPetOwners

This endpoint tests using PetOwner models from different namespaces.
Returns both foo.PetOwner and bar.PetOwner in the response.

### Example Usage

<!-- UsageSnippet language="java" operationID="getPetOwners" method="get" path="/petOwners" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetPetOwnersResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetPetOwnersResponse res = sdk.namespaceTests().conflicts().getPetOwners()
                .call();

        if (res.object().isPresent()) {
            System.out.println(res.object().get());
        }
    }
}
```

### Response

**[GetPetOwnersResponse](../../models/operations/GetPetOwnersResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |