# Group.SubGroup.Empty.Tail

## Overview

### Available Operations

* [nestedGroupOp](#nestedgroupop) - An operation at the group's deepest level

## nestedGroupOp

Notice that 'group.flattened' has no operations.


### Example Usage

<!-- UsageSnippet language="java" operationID="nestedGroupOp" method="get" path="/group/nested" -->
```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.NestedGroupOpResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        NestedGroupOpResponse res = sdk.group().subGroup().empty().tail().nestedGroupOp()
                .call();

        // handle response
    }
}
```

### Response

**[NestedGroupOpResponse](../../models/operations/NestedGroupOpResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/SDKException | 4XX, 5XX                   | \*/\*                      |