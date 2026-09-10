# Group

## Overview

### Available Operations

* [rootGroupOp](#rootgroupop) - An operation at the group's root level

## rootGroupOp

'group' differs from 'TestGroup' in that it not only contains subgroups,
but also an operation.


### Example Usage

<!-- UsageSnippet language="php" operationID="rootGroupOp" method="get" path="/group/root" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->group->rootGroupOp(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Response

**[?RootGroupOpResponse](../../RootGroupOpResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |