# Group.SubGroup

## Overview

### Available Operations

* [subGroupOp](#subgroupop) - An operation at the group's top level

## subGroupOp

An operation at the group's top level

### Example Usage

<!-- UsageSnippet language="java" operationID="subGroupOp" method="get" path="/group/subgroup" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.SubGroupOpResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        SubGroupOpResponse res = sdk.group().subGroup().subGroupOp()
                .call();

        // handle response
    }
}
```

### Response

**[SubGroupOpResponse](../../models/operations/SubGroupOpResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |