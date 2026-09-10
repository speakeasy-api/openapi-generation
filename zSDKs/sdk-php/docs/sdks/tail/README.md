# Group.SubGroup.Empty.Tail

## Overview

### Available Operations

* [nestedGroupOp](#nestedgroupop) - An operation at the group's deepest level

## nestedGroupOp

Notice that 'group.flattened' has no operations.


### Example Usage

<!-- UsageSnippet language="php" operationID="nestedGroupOp" method="get" path="/group/nested" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->group->subGroup->empty->tail->nestedGroupOp(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Response

**[?NestedGroupOpResponse](../../NestedGroupOpResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |