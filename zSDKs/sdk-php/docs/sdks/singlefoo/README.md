# NamespaceTests.SingleFoo

## Overview

### Available Operations

* [getSingleNamespaceFooPet](#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [createSingleNamespaceFooPet](#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

## getSingleNamespaceFooPet

This endpoint tests using a single component from the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="php" operationID="getSingleNamespaceFooPet" method="get" path="/singleNamespace/foo/pet" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->singleFoo->getSingleNamespaceFooPet(

);

if ($response->pet !== null) {
    // handle response
}
```

### Response

**[?GetSingleNamespaceFooPetResponse](../../GetSingleNamespaceFooPetResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## createSingleNamespaceFooPet

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

### Example Usage

<!-- UsageSnippet language="php" operationID="createSingleNamespaceFooPet" method="post" path="/singleNamespace/foo/pet" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Foo;

$sdk = OpenAPI\SDK::builder()->build();

$request = new \OpenAPI\OpenAPI\Foo\Pet(
    id: 'pet-foo-123',
    name: 'Fluffy',
    species: 'cat',
);

$response = $sdk->namespaceTests->singleFoo->createSingleNamespaceFooPet(
    request: $request
);

if ($response->pet !== null) {
    // handle response
}
```

### Parameters

| Parameter                                    | Type                                         | Required                                     | Description                                  |
| -------------------------------------------- | -------------------------------------------- | -------------------------------------------- | -------------------------------------------- |
| `$request`                                   | [\OpenAPI\OpenAPI\Foo\Pet](../../foo/Pet.md) | :heavy_check_mark:                           | The request object to use for the request.   |

### Response

**[?CreateSingleNamespaceFooPetResponse](../../CreateSingleNamespaceFooPetResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |