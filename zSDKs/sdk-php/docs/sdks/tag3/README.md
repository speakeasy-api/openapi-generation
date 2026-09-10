# TestGroup.Tag3

## Overview

### Available Operations

* [postTest](#posttest) - Post Test2

## postTest

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="php" operationID="postTest2" method="post" path="/test2" -->
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

$response = $sdk->testGroup->tag3->postTest(
    test2Request: $test2Request
);

if ($response->body !== null) {
    // handle response
}
```

### Parameters

| Parameter                                                                                                               | Type                                                                                                                    | Required                                                                                                                | Description                                                                                                             | Example                                                                                                                 |
| ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `test2Request`                                                                                                          | [Test2Request](../../Test2Request.md)                                                                                   | :heavy_check_mark:                                                                                                      | N/A                                                                                                                     |                                                                                                                         |
| `deprecatedQueryParam1`                                                                                                 | *?string*                                                                                                               | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `deprecatedQueryParam2`                                                                                                 | *?string*                                                                                                               | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `$serverURL`                                                                                                            | *string*                                                                                                                | :heavy_minus_sign:                                                                                                      | An optional server URL to use.                                                                                          | http://localhost:8080                                                                                                   |

### Response

**[?PostTest2Response](../../PostTest2Response.md)**

### Errors

| Error Type                          | Status Code                         | Content Type                        |
| ----------------------------------- | ----------------------------------- | ----------------------------------- |
| OpenAPI\BadRequestResponseException | 400                                 | application/json                    |
| OpenAPI\ErrorsError                 | 404                                 | application/json                    |
| OpenAPI\Test2ResponseException      | 500                                 | application/json                    |
| OpenAPI\SDKException                | 4XX, 5XX                            | \*/\*                               |