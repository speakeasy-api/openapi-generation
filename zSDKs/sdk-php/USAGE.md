<!-- Start SDK Example Usage [usage] -->
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