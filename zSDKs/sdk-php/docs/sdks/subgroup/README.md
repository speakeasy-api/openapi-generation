# Group.SubGroup

## Overview

### Available Operations

* [subGroupOp](#subgroupop) - An operation at the group's top level

## subGroupOp

An operation at the group's top level

### Example Usage

<!-- UsageSnippet language="php" operationID="subGroupOp" method="get" path="/group/subgroup" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->group->subGroup->subGroupOp(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Response

**[?SubGroupOpResponse](../../SubGroupOpResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |