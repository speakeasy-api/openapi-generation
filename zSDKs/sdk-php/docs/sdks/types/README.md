# NamespaceTests.Types

## Overview

### Available Operations

* [getNamespaceTypes](#getnamespacetypes) - Get Namespace Types Test
* [getNamespaceAnimal](#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [getNamespaceVehicle](#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [getNamespaceOrganization](#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

## getNamespaceTypes

This endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,
non-discriminated unions, and models with nested inline schemas.

### Example Usage

<!-- UsageSnippet language="php" operationID="getNamespaceTypes" method="get" path="/namespaceTypes" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->types->getNamespaceTypes(

);

if ($response->namespaceTypesTest !== null) {
    // handle response
}
```

### Response

**[?GetNamespaceTypesResponse](../../GetNamespaceTypesResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getNamespaceAnimal

This endpoint tests a discriminated union type in a custom namespace.
The response can be either a foo.Dog or foo.Cat.

### Example Usage

<!-- UsageSnippet language="php" operationID="getNamespaceAnimal" method="get" path="/namespaceAnimal" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->types->getNamespaceAnimal(

);

if ($response->animal !== null) {
    // handle response
}
```

### Response

**[?GetNamespaceAnimalResponse](../../GetNamespaceAnimalResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getNamespaceVehicle

This endpoint tests a non-discriminated union type in a custom namespace.
The response can be either a bar.Car or bar.Bike.

### Example Usage

<!-- UsageSnippet language="php" operationID="getNamespaceVehicle" method="get" path="/namespaceVehicle" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->types->getNamespaceVehicle(

);

if ($response->vehicle !== null) {
    // handle response
}
```

### Response

**[?GetNamespaceVehicleResponse](../../GetNamespaceVehicleResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getNamespaceOrganization

This endpoint tests nested inline object schemas in a custom namespace.
The organization model contains nested address and department types that
should inherit the foo namespace.

### Example Usage

<!-- UsageSnippet language="php" operationID="getNamespaceOrganization" method="get" path="/namespaceOrganization" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->namespaceTests->types->getNamespaceOrganization(

);

if ($response->organization !== null) {
    // handle response
}
```

### Response

**[?GetNamespaceOrganizationResponse](../../GetNamespaceOrganizationResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |