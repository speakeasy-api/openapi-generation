# Group

## Overview

### Available Operations

* [rootGroupOp](#rootgroupop) - An operation at the group's root level

## rootGroupOp

'group' differs from 'TestGroup' in that it not only contains subgroups,
but also an operation.


### Example Usage

<!-- UsageSnippet language="java" operationID="rootGroupOp" method="get" path="/group/root" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.RootGroupOpResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        RootGroupOpResponse res = sdk.group().rootGroupOp()
                .call();

        // handle response
    }
}
```

### Response

**[RootGroupOpResponse](../../models/operations/RootGroupOpResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |