# NamespaceTests.SingleBar

## Overview

### Available Operations

* [getSingleNamespaceBarPet](#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

## getSingleNamespaceBarPet

This endpoint tests using a single component from the bar namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="java" operationID="getSingleNamespaceBarPet" method="get" path="/singleNamespace/bar/pet" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.GetSingleNamespaceBarPetResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        GetSingleNamespaceBarPetResponse res = sdk.namespaceTests().singleBar().getSingleNamespaceBarPet()
                .call();

        if (res.pet().isPresent()) {
            System.out.println(res.pet().get());
        }
    }
}
```

### Response

**[GetSingleNamespaceBarPetResponse](../../models/operations/GetSingleNamespaceBarPetResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |