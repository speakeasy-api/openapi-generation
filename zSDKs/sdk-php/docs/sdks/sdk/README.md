# SDK

## Overview

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

Speakeasy Docs
<https://speakeasy.com/docs>

### Available Operations

* [operationWithLeadingAndTrailingUnderscores](#operationwithleadingandtrailingunderscores)
* [postFile](#postfile) - Post File
* [getPolymorphism](#getpolymorphism)
* [getUnionErrors](#getunionerrors)
* [getRequestBodyFlattenedAway](#getrequestbodyflattenedaway)
* [getFullyFlattenedRequest](#getfullyflattenedrequest)
* [createWithUnion](#createwithunion) - Create with discriminated union request body
* [testEndpoint](#testendpoint)
* [createUser](#createuser) - Create User
* [getUser](#getuser) - Get User
* [updateUser](#updateuser) - Update User
* [deleteUser](#deleteuser) - Delete User
* [login](#login) - Login
* [validate](#validate) - Validate
* [getBinaryDefaultResponse](#getbinarydefaultresponse)
* [testEnumFormats](#testenumformats) - Test x-speakeasy-enums in different formats
* [binaryAndStringUpload](#binaryandstringupload)
* [getErrorInUnion](#geterrorinunion)
* [getDuplicateExportCollision](#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [getNamedPrimitiveUnion](#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [getEmptyObjectError](#getemptyobjecterror) - Get Empty Object Error
* [urlValidationStressTest](#urlvalidationstresstest)
* [parenthesesInPathAllowed](#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [getNestedIntegerString](#getnestedintegerstring) - Test nested struct with integer:string tag
* [getErrorOnlyExample](#geterroronlyexample) - Operation with example only on error response

## operationWithLeadingAndTrailingUnderscores

### Example Usage

<!-- UsageSnippet language="php" operationID="_operation_with_leading_and_trailing_underscores_" method="get" path="/test_operation_id_with_underscores" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->operationWithLeadingAndTrailingUnderscores(
    qp1: 'renamed'
);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                                                                                                                                        | Type                                                                                                                                             | Required                                                                                                                                         | Description                                                                                                                                      | Example                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `qp1`                                                                                                                                            | *string*                                                                                                                                         | :heavy_check_mark:                                                                                                                               | This parameter will not be filled in with the queryParam1 global because it uses x-speakeasy-name-override which results in a non-matching name. | renamed                                                                                                                                          |

### Response

**[?OperationWithLeadingAndTrailingUnderscoresResponse](../../OperationWithLeadingAndTrailingUnderscoresResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## postFile

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="php" operationID="postFile" method="post" path="/file" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\PostFileRequest(
    upload: new OpenAPI\File(
        fileName: 'example.file',
        content: file_get_contents('example.file');,
    ),
);

$response = $sdk->postFile(
    request: $request
);

if ($response->file !== null) {
    // handle response
}
```

### Parameters

| Parameter                                           | Type                                                | Required                                            | Description                                         |
| --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- |
| `$request`                                          | [OpenAPI\PostFileRequest](../../PostFileRequest.md) | :heavy_check_mark:                                  | The request object to use for the request.          |

### Response

**[?PostFileResponse](../../PostFileResponse.md)**

### Errors

| Error Type          | Status Code         | Content Type        |
| ------------------- | ------------------- | ------------------- |
| OpenAPI\ErrorsError | 415, 4XX            | application/json    |
| OpenAPI\ErrorsError | 5XX                 | application/json    |

## getPolymorphism

### Example Usage

<!-- UsageSnippet language="php" operationID="getPolymorphism" method="get" path="/polymorphism" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getPolymorphism(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?GetPolymorphismResponse](../../GetPolymorphismResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getUnionErrors

### Example Usage

<!-- UsageSnippet language="php" operationID="getUnionErrors" method="get" path="/unionErrors" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$responses = $sdk->getUnionErrors(
    page: 12
);


foreach ($responses as $response) {
    if ($response->statusCode === 200) {
        // handle response
    }
}
```

### Parameters

| Parameter          | Type               | Required           | Description        | Example            |
| ------------------ | ------------------ | ------------------ | ------------------ | ------------------ |
| `page`             | *int*              | :heavy_check_mark: | N/A                | 12                 |

### Response

**[?GetUnionErrorsResponse](../../GetUnionErrorsResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\ErrorsError  | 404                  | application/json     |
| OpenAPI\ErrorType1   | 500                  | application/json     |
| OpenAPI\ErrorType2   | 500                  | application/json     |
| OpenAPI\TaggedError1 | 4XX                  | application/json     |
| OpenAPI\TaggedError2 | 4XX                  | application/json     |
| OpenAPI\SDKException | 5XX                  | \*/\*                |

## getRequestBodyFlattenedAway

### Example Usage

<!-- UsageSnippet language="php" operationID="getRequestBodyFlattenedAway" method="get" path="/requestBodyFlattenedAway" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setLoneQueryParam('<value>')
    ->build();



$response = $sdk->getRequestBodyFlattenedAway(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `loneQueryParam`   | *?string*          | :heavy_minus_sign: | N/A                |

### Response

**[?GetRequestBodyFlattenedAwayResponse](../../GetRequestBodyFlattenedAwayResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getFullyFlattenedRequest

### Example Usage

<!-- UsageSnippet language="php" operationID="getFullyFlattenedRequest" method="post" path="/fullyFlattenedRequest" example="namedExampleThatIsntMatchedAcrossDifferentExamples" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            option6: new OpenAPI\SecurityOption6(
                clientID: '<YOUR_CLIENT_ID_HERE>',
                clientSecret: '<YOUR_CLIENT_SECRET_HERE>',
                tokenURL: '/clientcredentials/token',
            ),
        )
    )
    ->build();

$requestBody = new OpenAPI\GetFullyFlattenedRequestRequestBody(
    name: '<value>',
);

$response = $sdk->getFullyFlattenedRequest(
    lang: 'en',
    requestBody: $requestBody

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                                                                           | Type                                                                                | Required                                                                            | Description                                                                         |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `lang`                                                                              | *string*                                                                            | :heavy_check_mark:                                                                  | N/A                                                                                 |
| `requestBody`                                                                       | [GetFullyFlattenedRequestRequestBody](../../GetFullyFlattenedRequestRequestBody.md) | :heavy_check_mark:                                                                  | N/A                                                                                 |
| `maxLength`                                                                         | *?int*                                                                              | :heavy_minus_sign:                                                                  | N/A                                                                                 |

### Response

**[?GetFullyFlattenedRequestResponse](../../GetFullyFlattenedRequestResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## createWithUnion

Test CLI generation for discriminated unions with dot-notation flags

### Example Usage

<!-- UsageSnippet language="php" operationID="createWithUnion" method="post" path="/unionRequestBody" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$shapeRequest = new OpenAPI\ShapeRequest(
    name: '<value>',
    shape: new OpenAPI\Rectangle(
        type: 'rectangle',
        width: 3125.73,
        height: 922.51,
    ),
);

$response = $sdk->createWithUnion(
    shapeRequest: $shapeRequest
);

if ($response->object !== null) {
    // handle response
}
```

### Parameters

| Parameter                             | Type                                  | Required                              | Description                           |
| ------------------------------------- | ------------------------------------- | ------------------------------------- | ------------------------------------- |
| `shapeRequest`                        | [ShapeRequest](../../ShapeRequest.md) | :heavy_check_mark:                    | N/A                                   |
| `dryRun`                              | *?bool*                               | :heavy_minus_sign:                    | If true, validates without creating   |

### Response

**[?CreateWithUnionResponse](../../CreateWithUnionResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## testEndpoint

### Example Usage

<!-- UsageSnippet language="php" operationID="testEndpoint" method="post" path="/test/endpoint/{testName}" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$requestBody = new OpenAPI\TestEndpointRequestBody(
    test: '<value>',
);

$response = $sdk->testEndpoint(
    testName: '<value>',
    requestBody: $requestBody

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                                                   | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `testName`                                                  | *string*                                                    | :heavy_check_mark:                                          | N/A                                                         |
| `requestBody`                                               | [TestEndpointRequestBody](../../TestEndpointRequestBody.md) | :heavy_check_mark:                                          | N/A                                                         |

### Response

**[?TestEndpointResponse](../../TestEndpointResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## createUser

Creates a new user in the system. Multiple named examples demonstrate
different pairing scenarios for documentation generation.


### Example Usage: paired-example

<!-- UsageSnippet language="php" operationID="createUser" method="put" path="/user" example="paired-example" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\BaseUser(
    email: 'paired@example.com',
    firstName: 'John',
);

$response = $sdk->createUser(
    request: $request
);

if ($response->user !== null) {
    // handle response
}
```
### Example Usage: request-only

<!-- UsageSnippet language="php" operationID="createUser" method="put" path="/user" example="request-only" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\BaseUser(
    email: 'request-only@example.com',
);

$response = $sdk->createUser(
    request: $request
);

if ($response->user !== null) {
    // handle response
}
```
### Example Usage: response-only

<!-- UsageSnippet language="php" operationID="createUser" method="put" path="/user" example="response-only" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\BaseUser(
    id: '8ffac18c-7d88-4879-b057-e5f45b9ce7de',
    email: 'Virginie47@gmail.com',
    gender: OpenAPI\Gender::Other,
);

$response = $sdk->createUser(
    request: $request
);

if ($response->user !== null) {
    // handle response
}
```

### Parameters

| Parameter                                  | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `$request`                                 | [OpenAPI\BaseUser](../../BaseUser.md)      | :heavy_check_mark:                         | The request object to use for the request. |

### Response

**[?CreateUserResponse](../../CreateUserResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getUser

Get User

### Example Usage

<!-- UsageSnippet language="php" operationID="getUser" method="get" path="/user/{id}" example="success" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getUser(
    id: '<id>'
);

if ($response->user !== null) {
    // handle response
}
```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `id`               | *string*           | :heavy_check_mark: | N/A                |

### Response

**[?GetUserResponse](../../GetUserResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## updateUser

Update User

### Example Usage

<!-- UsageSnippet language="php" operationID="updateUser" method="post" path="/user/{id}" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$user = new OpenAPI\User(
    id: '8ffac18c-7d88-4879-b057-e5f45b9ce7de',
    email: 'Joanny.Feeney@gmail.com',
    gender: OpenAPI\Gender::Other,
);

$response = $sdk->updateUser(
    id: '<id>',
    user: $user

);

if ($response->user !== null) {
    // handle response
}
```

### Parameters

| Parameter             | Type                  | Required              | Description           |
| --------------------- | --------------------- | --------------------- | --------------------- |
| `id`                  | *string*              | :heavy_check_mark:    | N/A                   |
| `user`                | [User](../../User.md) | :heavy_check_mark:    | N/A                   |

### Response

**[?UpdateUserResponse](../../UpdateUserResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## deleteUser

Delete User

### Example Usage

<!-- UsageSnippet language="php" operationID="deleteUser" method="delete" path="/user/{id}" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->deleteUser(
    id: '<id>'
);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter          | Type               | Required           | Description        |
| ------------------ | ------------------ | ------------------ | ------------------ |
| `id`               | *string*           | :heavy_check_mark: | N/A                |

### Response

**[?DeleteUserResponse](../../DeleteUserResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## login

Login

### Example Usage

<!-- UsageSnippet language="php" operationID="login" method="get" path="/auth/login" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->login(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?LoginResponse](../../LoginResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## validate

Validate

### Example Usage

<!-- UsageSnippet language="php" operationID="validate" method="get" path="/auth/validate" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->validate(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?ValidateResponse](../../ValidateResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getBinaryDefaultResponse

### Example Usage

<!-- UsageSnippet language="php" operationID="getBinaryDefaultResponse" method="get" path="/binaryDefaultResponse" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getBinaryDefaultResponse(

);

if ($response->bytes !== null) {
    // handle response
}
```

### Response

**[?GetBinaryDefaultResponseResponse](../../GetBinaryDefaultResponseResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## testEnumFormats

This endpoint tests the x-speakeasy-enums extension in both array and map formats,
including partial map coverage and both string and integer enum types.

### Example Usage

<!-- UsageSnippet language="php" operationID="testEnumFormats" method="post" path="/enumFormats" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\TestEnumFormatsRequest(
    stringArrayFormat: OpenAPI\StringArrayFormat::AwaitingReviewProcess,
    stringMapFormat: OpenAPI\StringMapFormat::ModerateImportanceLevel,
    stringPartialMapFormat: OpenAPI\StringPartialMapFormat::InitialDraftVersion,
    integerMapFormat: OpenAPI\IntegerMapFormat::SuccessfulOperationComplete,
    integerPartialMapFormat: OpenAPI\IntegerPartialMapFormat::PrimaryFirstOption,
);

$response = $sdk->testEnumFormats(
    request: $request
);

if ($response->object !== null) {
    // handle response
}
```

### Parameters

| Parameter                                                         | Type                                                              | Required                                                          | Description                                                       |
| ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------- |
| `$request`                                                        | [OpenAPI\TestEnumFormatsRequest](../../TestEnumFormatsRequest.md) | :heavy_check_mark:                                                | The request object to use for the request.                        |

### Response

**[?TestEnumFormatsResponse](../../TestEnumFormatsResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## binaryAndStringUpload

### Example Usage

<!-- UsageSnippet language="php" operationID="binaryAndStringUpload" method="post" path="/binaryAndStringUpload" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\BinaryAndStringUploadRequest(
    binary: file_get_contents('test.json');,
    string: file_get_contents('test.json');,
);

$response = $sdk->binaryAndStringUpload(
    request: $request
);

if ($response->statusCode === 200) {
    // handle response
}
```

### Parameters

| Parameter                                                                     | Type                                                                          | Required                                                                      | Description                                                                   |
| ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `$request`                                                                    | [OpenAPI\BinaryAndStringUploadRequest](../../BinaryAndStringUploadRequest.md) | :heavy_check_mark:                                                            | The request object to use for the request.                                    |

### Response

**[?BinaryAndStringUploadResponse](../../BinaryAndStringUploadResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getErrorInUnion

### Example Usage

<!-- UsageSnippet language="php" operationID="getErrorInUnion" method="get" path="/errorInUnion" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getErrorInUnion(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Response

**[?GetErrorInUnionResponse](../../GetErrorInUnionResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\ErrorsError  | 500                  | application/json     |
| OpenAPI\TaggedError1 | 500                  | application/json     |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getDuplicateExportCollision

Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308

### Example Usage

<!-- UsageSnippet language="php" operationID="getDuplicateExportCollision" method="get" path="/duplicateExportCollision" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getDuplicateExportCollision(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?GetDuplicateExportCollisionResponse](../../GetDuplicateExportCollisionResponse.md)**

### Errors

| Error Type                  | Status Code                 | Content Type                |
| --------------------------- | --------------------------- | --------------------------- |
| OpenAPI\RequestTimeoutError | 408                         | application/json            |
| OpenAPI\SDKException        | 4XX, 5XX                    | \*/\*                       |

## getNamedPrimitiveUnion

Test named primitive union options using title and x-speakeasy-name-override

### Example Usage

<!-- UsageSnippet language="php" operationID="getNamedPrimitiveUnion" method="get" path="/namedPrimitiveUnion" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getNamedPrimitiveUnion(

);

if ($response->someUnion !== null) {
    // handle response
}
```

### Response

**[?GetNamedPrimitiveUnionResponse](../../GetNamedPrimitiveUnionResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getEmptyObjectError

This endpoint tests the behavior when an error response has an empty object schema.

### Example Usage

<!-- UsageSnippet language="php" operationID="getEmptyObjectError" method="get" path="/emptyObjectError" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getEmptyObjectError(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?GetEmptyObjectErrorResponse](../../GetEmptyObjectErrorResponse.md)**

### Errors

| Error Type                      | Status Code                     | Content Type                    |
| ------------------------------- | ------------------------------- | ------------------------------- |
| OpenAPI\FailedResponseException | 500                             | application/json                |
| OpenAPI\SDKException            | 4XX, 5XX                        | \*/\*                           |

## urlValidationStressTest

### Example Usage

<!-- UsageSnippet language="php" operationID="urlValidationStressTest" method="get" path="/AZaz09-._~!$&'*+,;=:@/%20%25/-._~!$&'()*+,;=:@" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->urlValidationStressTest(

);

if ($response->statusCode === 200) {
    // handle response
}
```

### Response

**[?UrlValidationStressTestResponse](../../UrlValidationStressTestResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## parenthesesInPathAllowed

A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".


### Example Usage

<!-- UsageSnippet language="php" operationID="parenthesesInPathAllowed" method="post" path="/jobs/job({id})" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$templateBracesTest = new OpenAPI\TemplateBracesTest(
    fieldWithBracesInDescription: 'A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand \'single quotes\' and "double quotes".\nand  \\\'escaped single quotes\\\' and \\"escaped double quotes\\".\n',
    fieldWithBracesInTitle: 'A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand \'single quotes\' and "double quotes".\nand  \\\'escaped single quotes\\\' and \\"escaped double quotes\\".\n',
    fieldWithBracesInExample: 'A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand \'single quotes\' and "double quotes".\nand  \\\'escaped single quotes\\\' and \\"escaped double quotes\\".\n',
);

$response = $sdk->parenthesesInPathAllowed(
    id: '<id>',
    templateBracesTest: $templateBracesTest

);

if ($response->templateBracesTest !== null) {
    // handle response
}
```

### Parameters

| Parameter                                         | Type                                              | Required                                          | Description                                       |
| ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------- |
| `id`                                              | *string*                                          | :heavy_check_mark:                                | N/A                                               |
| `templateBracesTest`                              | [TemplateBracesTest](../../TemplateBracesTest.md) | :heavy_check_mark:                                | N/A                                               |

### Response

**[?ParenthesesInPathAllowedResponse](../../ParenthesesInPathAllowedResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getNestedIntegerString

This endpoint tests the behavior when a deeply nested struct contains
an integer field that should be unmarshaled from a string.

### Example Usage

<!-- UsageSnippet language="php" operationID="getNestedIntegerString" method="get" path="/nestedIntegerString" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getNestedIntegerString(

);

if ($response->taskResponse !== null) {
    // handle response
}
```

### Response

**[?GetNestedIntegerStringResponse](../../GetNestedIntegerStringResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |

## getErrorOnlyExample

This endpoint tests that when an operation has a named example only on
an error response (not on the success response), we still generate
a default example for the success response.

### Example Usage

<!-- UsageSnippet language="php" operationID="getErrorOnlyExample" method="get" path="/errorOnlyExample" -->
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();



$response = $sdk->getErrorOnlyExample(

);

if ($response->object !== null) {
    // handle response
}
```

### Response

**[?GetErrorOnlyExampleResponse](../../GetErrorOnlyExampleResponse.md)**

### Errors

| Error Type           | Status Code          | Content Type         |
| -------------------- | -------------------- | -------------------- |
| OpenAPI\ErrorsError  | 404                  | application/json     |
| OpenAPI\SDKException | 4XX, 5XX             | \*/\*                |