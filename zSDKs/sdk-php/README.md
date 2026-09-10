# openapi/openapi

Developer-friendly & type-safe Php SDK specifically catered to leverage *openapi/openapi* API.

[![Built by Speakeasy](https://img.shields.io/badge/Built_by-SPEAKEASY-374151?style=for-the-badge&labelColor=f3f4f6)](https://www.speakeasy.com/?utm_source=openapi/openapi&utm_campaign=php)
[![License: MIT](https://img.shields.io/badge/LICENSE_//_MIT-3b5bdb?style=for-the-badge&labelColor=eff6ff)](https://opensource.org/licenses/MIT)


<br /><br />
> [!IMPORTANT]
> This SDK is not yet ready for production use. Delete this section before > publishing to a package manager.

<!-- Start Summary [summary] -->
## Summary

SDK Review: A test document for reviewing the SDK.

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

For more information about the API: [Speakeasy Docs](https://speakeasy.com/docs)
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [openapi/openapi](#openapiopenapi)
  * [SDK Installation](#sdk-installation)
  * [SDK Example Usage](#sdk-example-usage)
  * [Authentication](#authentication)
  * [Available Resources and Operations](#available-resources-and-operations)
  * [Global Parameters](#global-parameters)
  * [Pagination](#pagination)
  * [Retries](#retries)
  * [Error Handling](#error-handling)
  * [Server Selection](#server-selection)
* [Development](#development)
  * [Maturity](#maturity)
  * [Contributions](#contributions)

<!-- End Table of Contents [toc] -->

<!-- Start SDK Installation [installation] -->
## SDK Installation

> [!TIP]
> To finish publishing your SDK you must [run your first generation action](https://www.speakeasy.com/docs/github-setup#step-by-step-guide).


The SDK relies on [Composer](https://getcomposer.org/) to manage its dependencies.

To install the SDK first add the below to your `composer.json` file:

```json
{
    "repositories": [
        {
            "type": "github",
            "url": "<UNSET>.git"
        }
    ],
    "require": {
        "openapi/openapi": "*"
    }
}
```

Then run the following command:

```bash
composer update
```
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## SDK Example Usage

### Example 1

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

### Example 2

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

$request = new OpenAPI\PostFileWithEncodingRequest(
    file: new OpenAPI\PostFileWithEncodingFile(
        fileName: 'example.file',
        content: file_get_contents('example.file');,
    ),
);

$response = $sdk->tag1->postFileWithEncoding(
    request: $request
);

if ($response->object !== null) {
    // handle response
}
```

### Example 3

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use Brick\DateTime\LocalDate;
use Brick\Math\BigDecimal;
use Brick\Math\BigInteger;
use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Utils;

$sdk = OpenAPI\SDK::builder()
    ->setDeprecatedQueryParam1('some example query param')
    ->setDeprecatedQueryParam2('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();

$test2Request = new OpenAPI\Test2Request(
    obj: new OpenAPI\ExhaustiveObject(
        str: 'example',
        bool: true,
        integer: 999999,
        int32: 1,
        num: 1.1,
        float32: 8499.3,
        date: LocalDate::parse('2020-01-01'),
        dateTime: Utils\Utils::parseDateTime('2020-01-01T00:00:00Z'),
        anything: '<value>',
        boolOpt: true,
        intOptNull: 999999,
        numOptNull: 1.1,
        intEnum: OpenAPI\IntEnum::Third,
        int32Enum: OpenAPI\Int32Enum::SixtyNine,
        bigint: BigInteger::of('702830'),
        decimalStr: BigDecimal::of('3858.6'),
        obj: new OpenAPI\SimpleObject(
            str: 'example',
        ),
        map: [
            'key' => new OpenAPI\SimpleObject(
                str: 'example',
            ),
        ],
        arr: [
            new OpenAPI\SimpleObject(
                str: 'example',
            ),
            new OpenAPI\SimpleObject(
                str: 'example',
            ),
        ],
        any: new OpenAPI\SimpleObject(
            str: 'example',
        ),
        nullableIntEnum: OpenAPI\NullableIntEnum::Third,
        nullableStringEnum: OpenAPI\NullableStringEnum::Second,
        color: OpenAPI\Color::Green,
        icon: OpenAPI\Icon::Tick,
        heroWidth: OpenAPI\HeroWidth::FourHundredAndEighty,
    ),
    type: OpenAPI\Type::SuperType1,
);

$response = $sdk->testGroup->tag2->postTest(
    test2Request: $test2Request
);

if ($response->body !== null) {
    // handle response
}
```

### A custom readme heading

A custom usage description

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setQueryParam1('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();



$responses = $sdk->tag1->listTest1(
    page: 100,
    queryParam2: OpenAPI\QueryParam2::One,
    headerParam1: 'some example header param'

);


foreach ($responses as $response) {
    if ($response->statusCode === 200) {
        // handle response
    }
}
```
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
## Authentication

### Per-Client Security Schemes

This SDK supports multiple security scheme combinations globally. You can choose from one of the alternatives through the `setSecurity` function on the `SDKBuilder` when initializing the SDK client instance. The selected option will be used by default to authenticate with the API for all operations that support it.

#### UserPassAuth

The `UserPassAuth` alternative relies on the following scheme:

| Name                      | Type | Scheme     | Environment Variable                      |
| ------------------------- | ---- | ---------- | ----------------------------------------- |
| `username`<br/>`password` | http | HTTP Basic | `OPENAPI_USERNAME`<br/>`OPENAPI_PASSWORD` |

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            userPassAuth: new OpenAPI\UserPassAuth(
                username: '<USERNAME>',
                password: '<PASSWORD>',
            ),
        )
    )
    ->build();

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

#### Option2

All of the following schemes must be satisfied to use the `Option2` alternative:

| Name         | Type   | Scheme      | Environment Variable  |
| ------------ | ------ | ----------- | --------------------- |
| `bearerAuth` | http   | HTTP Bearer | `OPENAPI_BEARER_AUTH` |
| `myApiKey`   | apiKey | API key     | `OPENAPI_MY_API_KEY`  |

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            option2: new OpenAPI\SecurityOption2(
                bearerAuth: '<YOUR_JWT>',
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();

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

#### Option3

The `Option3` alternative relies on the following scheme:

| Name     | Type   | Scheme       | Environment Variable |
| -------- | ------ | ------------ | -------------------- |
| `oauth2` | oauth2 | OAuth2 token | `OPENAPI_OAUTH2`     |

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            option3: new OpenAPI\SecurityOption3(
                oauth2: 'Bearer <YOUR_OAUTH2_TOKEN>',
            ),
        )
    )
    ->build();

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

#### Option4

The `Option4` alternative relies on the following scheme:

| Name                 | Type | Scheme      | Environment Variable                  |
| -------------------- | ---- | ----------- | ------------------------------------- |
| `appId`<br/>`secret` | http | Custom HTTP | `OPENAPI_APP_ID`<br/>`OPENAPI_SECRET` |

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            option4: new OpenAPI\SecurityOption4(
                appId: 'app-speakeasy-123',
                secret: 'MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI',
            ),
        )
    )
    ->build();

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

#### Option5

The `Option5` alternative relies on the following scheme:

| Name         | Type   | Scheme       | Environment Variable  |
| ------------ | ------ | ------------ | --------------------- |
| `mobileAuth` | oauth2 | OAuth2 token | `OPENAPI_MOBILE_AUTH` |

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            option5: new OpenAPI\SecurityOption5(
                mobileAuth: 'Bearer <YOUR_OAUTH2_TOKEN>',
            ),
        )
    )
    ->build();

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

#### Option6

The `Option6` alternative relies on the following scheme:

| Name                                         | Type   | Scheme                         | Environment Variable                                                    |
| -------------------------------------------- | ------ | ------------------------------ | ----------------------------------------------------------------------- |
| `clientID`<br/>`clientSecret`<br/>`tokenURL` | oauth2 | OAuth2 Client Credentials Flow | `OPENAPI_CLIENT_ID`<br/>`OPENAPI_CLIENT_SECRET`<br/>`OPENAPI_TOKEN_URL` |

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

#### MyApiKey

The `MyApiKey` alternative relies on the following scheme:

| Name       | Type   | Scheme  | Environment Variable |
| ---------- | ------ | ------- | -------------------- |
| `myApiKey` | apiKey | API key | `OPENAPI_MY_API_KEY` |

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();

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

### Per-Operation Security Schemes

Some operations in this SDK require the security scheme to be specified at the request level. For example:
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();


$requestSecurity = new OpenAPI\AuthSecurity(
    accessToken: '<YOUR_ACCESS_TOKEN_HERE>',
);

$response = $sdk->tag1->auth(
    security: $requestSecurity
);

if ($response->statusCode === 200) {
    // handle response
}
```
<!-- End Authentication [security] -->

<!-- Start Available Resources and Operations [operations] -->
## Available Resources and Operations

<details open>
<summary>Available methods</summary>

### [SDK](docs/sdks/sdk/README.md)

* [operationWithLeadingAndTrailingUnderscores](docs/sdks/sdk/README.md#operationwithleadingandtrailingunderscores)
* [postFile](docs/sdks/sdk/README.md#postfile) - Post File
* [getPolymorphism](docs/sdks/sdk/README.md#getpolymorphism)
* [getUnionErrors](docs/sdks/sdk/README.md#getunionerrors)
* [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway)
* [getFullyFlattenedRequest](docs/sdks/sdk/README.md#getfullyflattenedrequest)
* [createWithUnion](docs/sdks/sdk/README.md#createwithunion) - Create with discriminated union request body
* [testEndpoint](docs/sdks/sdk/README.md#testendpoint)
* [createUser](docs/sdks/sdk/README.md#createuser) - Create User
* [getUser](docs/sdks/sdk/README.md#getuser) - Get User
* [updateUser](docs/sdks/sdk/README.md#updateuser) - Update User
* [deleteUser](docs/sdks/sdk/README.md#deleteuser) - Delete User
* [login](docs/sdks/sdk/README.md#login) - Login
* [validate](docs/sdks/sdk/README.md#validate) - Validate
* [getBinaryDefaultResponse](docs/sdks/sdk/README.md#getbinarydefaultresponse)
* [testEnumFormats](docs/sdks/sdk/README.md#testenumformats) - Test x-speakeasy-enums in different formats
* [binaryAndStringUpload](docs/sdks/sdk/README.md#binaryandstringupload)
* [getErrorInUnion](docs/sdks/sdk/README.md#geterrorinunion)
* [getDuplicateExportCollision](docs/sdks/sdk/README.md#getduplicateexportcollision) - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [getNamedPrimitiveUnion](docs/sdks/sdk/README.md#getnamedprimitiveunion) - Test named primitive union options using title and x-speakeasy-name-override
* [getEmptyObjectError](docs/sdks/sdk/README.md#getemptyobjecterror) - Get Empty Object Error
* [urlValidationStressTest](docs/sdks/sdk/README.md#urlvalidationstresstest)
* [parenthesesInPathAllowed](docs/sdks/sdk/README.md#parenthesesinpathallowed) - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".

* [getNestedIntegerString](docs/sdks/sdk/README.md#getnestedintegerstring) - Test nested struct with integer:string tag
* [getErrorOnlyExample](docs/sdks/sdk/README.md#geterroronlyexample) - Operation with example only on error response

### [Group](docs/sdks/group/README.md)

* [rootGroupOp](docs/sdks/group/README.md#rootgroupop) - An operation at the group's root level

#### [Group.SubGroup](docs/sdks/subgroup/README.md)

* [subGroupOp](docs/sdks/subgroup/README.md#subgroupop) - An operation at the group's top level

##### [Group.SubGroup.Empty.Tail](docs/sdks/tail/README.md)

* [nestedGroupOp](docs/sdks/tail/README.md#nestedgroupop) - An operation at the group's deepest level

### [NamespaceTests.Conflicts](docs/sdks/conflicts/README.md)

* [getNamespaceConflict](docs/sdks/conflicts/README.md#getnamespaceconflict) - Get Namespace Conflict Test
* [putNamespaceConflict](docs/sdks/conflicts/README.md#putnamespaceconflict) - Put Property Name Conflicts Behind
* [createNamespaceConflict](docs/sdks/conflicts/README.md#createnamespaceconflict) - Create Namespace Conflict Test
* [getTripleNamespaceConflict](docs/sdks/conflicts/README.md#gettriplenamespaceconflict) - Get Triple Namespace Conflict Test
* [getPetOwners](docs/sdks/conflicts/README.md#getpetowners) - Get Pet Owners

### [NamespaceTests.SingleBar](docs/sdks/singlebar/README.md)

* [getSingleNamespaceBarPet](docs/sdks/singlebar/README.md#getsinglenamespacebarpet) - Get Single Namespace Bar Pet

### [NamespaceTests.SingleFoo](docs/sdks/singlefoo/README.md)

* [getSingleNamespaceFooPet](docs/sdks/singlefoo/README.md#getsinglenamespacefoopet) - Get Single Namespace Foo Pet
* [createSingleNamespaceFooPet](docs/sdks/singlefoo/README.md#createsinglenamespacefoopet) - Create Single Namespace Foo Pet

### [NamespaceTests.Types](docs/sdks/types/README.md)

* [getNamespaceTypes](docs/sdks/types/README.md#getnamespacetypes) - Get Namespace Types Test
* [getNamespaceAnimal](docs/sdks/types/README.md#getnamespaceanimal) - Get Namespace Animal (Discriminated Union)
* [getNamespaceVehicle](docs/sdks/types/README.md#getnamespacevehicle) - Get Namespace Vehicle (Non-Discriminated Union)
* [getNamespaceOrganization](docs/sdks/types/README.md#getnamespaceorganization) - Get Namespace Organization (Nested Inline Schemas)

### [~~Obsolete~~](docs/sdks/obsolete/README.md)

* [~~deprecated1~~](docs/sdks/obsolete/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.

### [Tag1](docs/sdks/tag1/README.md)

* [~~deprecated1~~](docs/sdks/tag1/README.md#deprecated1) - Deprecated Operation :warning: **Deprecated** Use [getRequestBodyFlattenedAway](docs/sdks/sdk/README.md#getrequestbodyflattenedaway) instead.
* [auth](docs/sdks/tag1/README.md#auth) - This operation aims at testing available OAuth2 scopes collection:
 - only operation with oauth2 authorizationCode security flow
 - belongs to a subSDK

* [listTest1](docs/sdks/tag1/README.md#listtest1) - Get Test1
* [postFileWithEncoding](docs/sdks/tag1/README.md#postfilewithencoding) - Post File With Encoding

### [TestGroup.Tag2](docs/sdks/tag2/README.md)

* [postTest](docs/sdks/tag2/README.md#posttest) - Post Test2

### [TestGroup.Tag3](docs/sdks/tag3/README.md)

* [postTest](docs/sdks/tag3/README.md#posttest) - Post Test2

</details>
<!-- End Available Resources and Operations [operations] -->

<!-- Start Global Parameters [global-parameters] -->
## Global Parameters

Certain parameters are configured globally. These parameters may be set on the SDK client instance itself during initialization. When configured as an option during SDK initialization, These global values will be used as defaults on the operations that use them. When such operations are called, there is a place in each to override the global value, if needed.

For example, you can set `queryParam1` to `'some example query param'` at SDK initialization and then you do not have to pass the same value on calls to operations like `getRequestBodyFlattenedAway`. But if you want to do so you may, which will locally override the global setting. See the example code below for a demonstration.


### Available Globals

The following global parameters are available.
Global parameters can also be set via environment variable.

| Name                  | Type   | Description                                                                        | Environment                     |
| --------------------- | ------ | ---------------------------------------------------------------------------------- | ------------------------------- |
| queryParam1           | string | A long winded, multi-line description<br/>for the query parameter number one.<br/> | OPENAPI_QUERY_PARAM1            |
| deprecatedQueryParam1 | string | A deprecated description                                                           | OPENAPI_DEPRECATED_QUERY_PARAM1 |
| deprecatedQueryParam2 | string | The deprecatedQueryParam2 parameter.                                               | OPENAPI_DEPRECATED_QUERY_PARAM2 |
| loneQueryParam        | string | The loneQueryParam parameter.                                                      | OPENAPI_LONE_QUERY_PARAM        |

### Example

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setLoneQueryParam('<value>')
    ->setQueryParam1('some example query param')
    ->setDeprecatedQueryParam1('some example query param')
    ->setDeprecatedQueryParam2('some example query param')
    ->build();



$response = $sdk->getRequestBodyFlattenedAway(

);

if ($response->statusCode === 200) {
    // handle response
}
```
<!-- End Global Parameters [global-parameters] -->

<!-- Start Pagination [pagination] -->
## Pagination

Some of the endpoints in this SDK support pagination. To use pagination, you make your SDK calls as usual, but the
returned object will be a `Generator` instead of an individual response.

Working with generators is as simple as iterating over the responses in a `foreach` loop, and you can see an example below:
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setQueryParam1('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();



$responses = $sdk->tag1->listTest1(
    page: 100,
    queryParam2: OpenAPI\QueryParam2::One,
    headerParam1: 'some example header param'

);


foreach ($responses as $response) {
    if ($response->statusCode === 200) {
        // handle response
    }
}
```
<!-- End Pagination [pagination] -->

<!-- Start Retries [retries] -->
## Retries

Some of the endpoints in this SDK support retries. If you use the SDK without any configuration, it will fall back to the default retry strategy provided by the API. However, the default retry strategy can be overridden on a per-operation basis, or across the entire SDK.

To change the default retry strategy for a single API call, simply provide an `Options` object built with a `RetryConfig` object to the call:
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use Brick\DateTime\LocalDate;
use Brick\Math\BigDecimal;
use Brick\Math\BigInteger;
use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Utils;
use OpenAPI\OpenAPI\Utils\Retry;

$sdk = OpenAPI\SDK::builder()
    ->setDeprecatedQueryParam1('some example query param')
    ->setDeprecatedQueryParam2('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();

$test2Request = new OpenAPI\Test2Request(
    obj: new OpenAPI\ExhaustiveObject(
        str: 'example',
        bool: true,
        integer: 999999,
        int32: 1,
        num: 1.1,
        float32: 8499.3,
        date: LocalDate::parse('2020-01-01'),
        dateTime: Utils\Utils::parseDateTime('2020-01-01T00:00:00Z'),
        anything: '<value>',
        boolOpt: true,
        intOptNull: 999999,
        numOptNull: 1.1,
        intEnum: OpenAPI\IntEnum::Third,
        int32Enum: OpenAPI\Int32Enum::SixtyNine,
        bigint: BigInteger::of('702830'),
        decimalStr: BigDecimal::of('3858.6'),
        obj: new OpenAPI\SimpleObject(
            str: 'example',
        ),
        map: [
            'key' => new OpenAPI\SimpleObject(
                str: 'example',
            ),
        ],
        arr: [
            new OpenAPI\SimpleObject(
                str: 'example',
            ),
            new OpenAPI\SimpleObject(
                str: 'example',
            ),
        ],
        any: new OpenAPI\SimpleObject(
            str: 'example',
        ),
        nullableIntEnum: OpenAPI\NullableIntEnum::Third,
        nullableStringEnum: OpenAPI\NullableStringEnum::Second,
        color: OpenAPI\Color::Green,
        icon: OpenAPI\Icon::Tick,
        heroWidth: OpenAPI\HeroWidth::FourHundredAndEighty,
    ),
    type: OpenAPI\Type::SuperType1,
);

$response = $sdk->testGroup->tag2->postTest(
    test2Request: $test2Request,
    options: Utils\Options->builder()->setRetryConfig(
        new Retry\RetryConfigBackoff(
            initialInterval: 1,
            maxInterval:     50,
            exponent:        1.1,
            maxElapsedTime:  100,
            retryConnectionErrors: false,
        ))->build()

);

if ($response->body !== null) {
    // handle response
}
```

If you'd like to override the default retry strategy for all operations that support retries, you can pass a `RetryConfig` object to the `SDKBuilder->setRetryConfig` function when initializing the SDK:
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use Brick\DateTime\LocalDate;
use Brick\Math\BigDecimal;
use Brick\Math\BigInteger;
use OpenAPI\OpenAPI;
use OpenAPI\OpenAPI\Utils;
use OpenAPI\OpenAPI\Utils\Retry;

$sdk = OpenAPI\SDK::builder()
    ->setRetryConfig(
        new Retry\RetryConfigBackoff(
            initialInterval: 1,
            maxInterval:     50,
            exponent:        1.1,
            maxElapsedTime:  100,
            retryConnectionErrors: false,
        )
  )
    ->setDeprecatedQueryParam1('some example query param')
    ->setDeprecatedQueryParam2('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();

$test2Request = new OpenAPI\Test2Request(
    obj: new OpenAPI\ExhaustiveObject(
        str: 'example',
        bool: true,
        integer: 999999,
        int32: 1,
        num: 1.1,
        float32: 8499.3,
        date: LocalDate::parse('2020-01-01'),
        dateTime: Utils\Utils::parseDateTime('2020-01-01T00:00:00Z'),
        anything: '<value>',
        boolOpt: true,
        intOptNull: 999999,
        numOptNull: 1.1,
        intEnum: OpenAPI\IntEnum::Third,
        int32Enum: OpenAPI\Int32Enum::SixtyNine,
        bigint: BigInteger::of('702830'),
        decimalStr: BigDecimal::of('3858.6'),
        obj: new OpenAPI\SimpleObject(
            str: 'example',
        ),
        map: [
            'key' => new OpenAPI\SimpleObject(
                str: 'example',
            ),
        ],
        arr: [
            new OpenAPI\SimpleObject(
                str: 'example',
            ),
            new OpenAPI\SimpleObject(
                str: 'example',
            ),
        ],
        any: new OpenAPI\SimpleObject(
            str: 'example',
        ),
        nullableIntEnum: OpenAPI\NullableIntEnum::Third,
        nullableStringEnum: OpenAPI\NullableStringEnum::Second,
        color: OpenAPI\Color::Green,
        icon: OpenAPI\Icon::Tick,
        heroWidth: OpenAPI\HeroWidth::FourHundredAndEighty,
    ),
    type: OpenAPI\Type::SuperType1,
);

$response = $sdk->testGroup->tag2->postTest(
    test2Request: $test2Request
);

if ($response->body !== null) {
    // handle response
}
```
<!-- End Retries [retries] -->

<!-- Start Error Handling [errors] -->
## Error Handling

Handling errors in this SDK should largely match your expectations. All operations return a response object or throw an exception.

By default an API error will raise a `OpenAPI\SDKException` exception, which has the following properties:

| Property       | Type                                    | Description           |
|----------------|-----------------------------------------|-----------------------|
| `$message`     | *string*                                | The error message     |
| `$statusCode`  | *int*                                   | The HTTP status code  |
| `$rawResponse` | *?\Psr\Http\Message\ResponseInterface*  | The raw HTTP response |
| `$body`        | *string*                                | The response content  |

When custom error responses are specified for an operation, the SDK may also throw their associated exception. You can refer to respective *Errors* tables in SDK docs for more details on possible exception types for each operation. For example, the `getUnionErrors` method throws the following exceptions:

| Error Type           | Status Code | Content Type     |
| -------------------- | ----------- | ---------------- |
| OpenAPI\ErrorsError  | 404         | application/json |
| OpenAPI\ErrorType1   | 500         | application/json |
| OpenAPI\ErrorType2   | 500         | application/json |
| OpenAPI\TaggedError1 | 4XX         | application/json |
| OpenAPI\TaggedError2 | 4XX         | application/json |
| OpenAPI\SDKException | 5XX         | \*/\*            |

### Example

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()->build();

try {
    $responses = $sdk->getUnionErrors(
        page: 12
    );

    foreach ($responses as $response) {
        if ($response->statusCode === 200) {
            // handle response
        }
    }
} catch (OpenAPI\ErrorsErrorThrowable $e) {
    // handle $e->$container data
    throw $e;
} catch (OpenAPI\ErrorType1|OpenAPI\ErrorType2Throwable $e) {
    // handle $e->$container data
    throw $e;
} catch (OpenAPI\TaggedError1|OpenAPI\TaggedError2Throwable $e) {
    // handle $e->$container data
    throw $e;
} catch (OpenAPI\SDKException $e) {
    // handle default exception
    throw $e;
}
```
<!-- End Error Handling [errors] -->

<!-- Start Server Selection [server] -->
## Server Selection

### Select Server by Index

You can override the default server globally using the `setServerIndex(int $serverIdx)` builder method when initializing the SDK client instance. The selected server will then be used as the default on the operations that use it. This table lists the indexes associated with the available servers:

| #   | Server                                     | Variables                 | Description                     |
| --- | ------------------------------------------ | ------------------------- | ------------------------------- |
| 0   | `http://localhost:35123`                   |                           | The default server.             |
| 1   | `http://{subdomain}.domain.com/v{version}` | `subdomain`<br/>`version` |                                 |
| 2   | `http://{HostName}:{PORT}`                 | `HostName`<br/>`PORT`     | A server with an enum variable. |

If the selected server has variables, you may override its default values using the associated builder method(s):

| Variable    | BuilderMethod                      | Supported Values                      | Default       | Description                              |
| ----------- | ---------------------------------- | ------------------------------------- | ------------- | ---------------------------------------- |
| `subdomain` | `setSubdomain(string subdomain)`   | string                                | `"api"`       |                                          |
| `version`   | `setVersion(string version)`       | string                                | `"1"`         |                                          |
| `HostName`  | `setHostName(string hostName)`     | string                                | `"localhost"` | The hostname of the server.              |
| `PORT`      | `setPORT(OpenAPI\ServerPORT port)` | - `"80"`<br/>- `"8080"`<br/>- `"443"` | `"8080"`      | The port on which the server is running. |

#### Example

```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setServerIndex(2)
    ->setHostName('localhost')
    ->setPORT('443')
    ->build();

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

### Override Server URL Per-Client

The default server can also be overridden globally using the `setServerUrl(string $serverUrl)` builder method when initializing the SDK client instance. For example:
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setServerURL('http://localhost:8080')
    ->build();

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

### Override Server URL Per-Operation

The server URL can also be overridden on a per-operation basis, provided a server list was specified for the operation. For example:
```php
declare(strict_types=1);

require 'vendor/autoload.php';

use OpenAPI\OpenAPI;

$sdk = OpenAPI\SDK::builder()
    ->setQueryParam1('some example query param')
    ->setSecurity(
        new OpenAPI\Security(
            myApiKey: new OpenAPI\MyApiKey(
                myApiKey: '<YOUR_API_KEY_HERE>',
            ),
        )
    )
    ->build();



$responses = $sdk->tag1->listTest1(
    'http://localhost:35123',
    page: 100,
    queryParam2: OpenAPI\QueryParam2::One,
    headerParam1: 'some example header param'

);


foreach ($responses as $response) {
    if ($response->statusCode === 200) {
        // handle response
    }
}
```
<!-- End Server Selection [server] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

# Development

## Maturity

This SDK is in beta, and there may be breaking changes between versions without a major version update. Therefore, we recommend pinning usage
to a specific package version. This way, you can install the same version each time without breaking changes unless you are intentionally
looking for the latest version.

## Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=openapi/openapi&utm_campaign=php)
