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

<!-- UsageSnippet language="php" operationID="getNamespaceConflict" method="get" path="/namespaceConflict" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->conflicts->getNamespaceConflict(

);

if ($response->namespaceConflictTest !== null) {
    // handle response
}
```

### Response

**[?GetNamespaceConflictResponse](../../GetNamespaceConflictResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## putNamespaceConflict

This endpoint tests property name conflict resolution through
x-speakeasy-name-override and x-speakeasy-model-namespace extensions.

### Example Usage

<!-- UsageSnippet language="php" operationID="putNamespaceConflict" method="put" path="/namespaceConflict" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->conflicts->putNamespaceConflict(
    request: $request
);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                                                             | Type                                                                  | Required                                                              | Description                                                           |
| --------------------------------------------------------------------- | --------------------------------------------------------------------- | --------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `$request`                                                            | [OpenAPI\ObjWithRenamedProperties](../../ObjWithRenamedProperties.md) | :heavy_check_mark:                                                    | The request object to use for the request.                            |

### Response

**[?PutNamespaceConflictResponse](../../PutNamespaceConflictResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## createNamespaceConflict

This endpoint tests creating with models from different namespaces.
Uses foo.Pet in the request and bar.Pet in the response.

### Example Usage

<!-- UsageSnippet language="php" operationID="createNamespaceConflict" method="post" path="/namespaceConflict" -->
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

$response = $sdk->namespaceTests->conflicts->createNamespaceConflict(
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

**[?CreateNamespaceConflictResponse](../../CreateNamespaceConflictResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getTripleNamespaceConflict

This endpoint tests the x-speakeasy-model-namespace extension by returning
a model that references three different Pet types from different namespaces.
The SDK should properly import and alias all three Pet types.

### Example Usage

<!-- UsageSnippet language="php" operationID="getTripleNamespaceConflict" method="get" path="/tripleNamespaceConflict" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->conflicts->getTripleNamespaceConflict(

);

if ($response->tripleNamespaceConflictTest !== null) {
    // handle response
}
```

### Response

**[?GetTripleNamespaceConflictResponse](../../GetTripleNamespaceConflictResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getPetOwners

This endpoint tests using PetOwner models from different namespaces.
Returns both foo.PetOwner and bar.PetOwner in the response.

### Example Usage

<!-- UsageSnippet language="php" operationID="getPetOwners" method="get" path="/petOwners" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->conflicts->getPetOwners(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?GetPetOwnersResponse](../../GetPetOwnersResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |