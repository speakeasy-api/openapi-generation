# NamespaceTests.SingleBar

## Overview

### Available Operations

* [getSingleNamespaceBarPet](#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

## getSingleNamespaceBarPet

This endpoint tests using a single component from the bar namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="php" operationID="getSingleNamespaceBarPet" method="get" path="/singleNamespace/bar/pet" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->singleBar->getSingleNamespaceBarPet(

);

if ($response->pet !== null) {
    // handle response
}
```

### Response

**[?GetSingleNamespaceBarPetResponse](../../GetSingleNamespaceBarPetResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |