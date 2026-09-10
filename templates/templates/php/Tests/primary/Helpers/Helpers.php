declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests\Helpers;

use Brick\DateTime\LocalDate;
use Brick\Math\BigInteger;
use Brick\Math\BigDecimal;
use OpenAPI\OpenAPI\Models\Shared;
use OpenAPI\OpenAPI\Utils;
use Speakeasy\Serializer\DeserializationContext;

//fwrite(STDERR, print_r($metadata, TRUE));

class Helpers
{
    // Test service URLs - read from environment with defaults for local development
    public const HTTPBIN_PORT = 'HTTPBIN_PORT';
    public const API_TEST_SERVICE_PORT = 'API_TEST_SERVICE_PORT';

    public static function getHttpBinPort(): string
    {
        return getenv(self::HTTPBIN_PORT) ?: '35123';
    }

    public static function getApiTestServicePort(): string
    {
        return getenv(self::API_TEST_SERVICE_PORT) ?: '35456';
    }

    public static function getHttpBinUrl(): string
    {
        return 'http://localhost:' . self::getHttpBinPort();
    }

    public static function getApiTestServiceUrl(): string
    {
        return 'http://localhost:' . self::getApiTestServicePort();
    }

    public static function createSimpleObject(): Shared\SimpleObject
    {
        $object = new Shared\SimpleObject(
            str: 'test',
            bool: true,
            bigint: BigInteger::of(8821239038968084),
            bigintStr: BigInteger::of('9223372036854775808'),
            decimal: BigDecimal::of('3.1415926535898'),
            decimalStr: BigDecimal::of('3.14159265358979344719667586'),
            int: 1,
            int32: 1,
            int64Str: "100",
            int32Enum: Shared\Int32Enum::FiftyFive,
            intEnum: Shared\IntEnum::Second,
            num: 1.1,
            float32: 1.1,
            float64Str: "1.1",
            enum: Shared\Enum::One,
            any: 'any',
            date: LocalDate::parse('2020-01-01'),
            dateTime: \DateTime::createFromFormat('Y-m-d\TH:i:s.up', '2020-01-01T00:00:00.000001Z'),
            boolOpt: true,
            strOpt: 'testOptional'
        );

        return $object;
    }

    public static function createSimpleObjectCamelCase(): Shared\SimpleObjectCamelCase
    {
        $object = new Shared\SimpleObjectCamelCase(
            anyVal: 'any',
            boolVal: true,
            boolOptVal: true,
            dateVal: LocalDate::parse('2020-01-01'),
            dateTimeVal: \DateTime::createFromFormat('Y-m-d\TH:i:s.up', '2020-01-01T00:00:00.000001Z'),
            enumVal: Shared\Enum::One,
            float32Val: 1.1,
            int32Val: 1,
            int32EnumVal: Shared\Int32EnumVal::FiftyFive,
            intEnumVal: Shared\IntEnumVal::Second,
            intOptNullVal: null,
            intVal: 1,
            numVal: 1.1,
            numOptNullVal: null,
            strVal: 'test',
            strOptVal: 'test_optional'
        );

        return $object;
    }

    public static function createSimpleObjectWithType(): Shared\SimpleObjectWithType
    {
        $object = new Shared\SimpleObjectWithType(
            str: 'test',
            bool: true,
            int: 1,
            int32: 1,
            int32Enum: Shared\SimpleObjectWithTypeInt32Enum::FiftyFive,
            intEnum: Shared\SimpleObjectWithTypeIntEnum::Second,
            num: 1.1,
            float32: 1.1,
            enum: Shared\Enum::One,
            any: 'any',
            date: LocalDate::parse('2020-01-01'),
            dateTime: \DateTime::createFromFormat('Y-m-d\TH:i:s.up', '2020-01-01T00:00:00.000001Z'),
            boolOpt: true,
            strOpt: 'testOptional',
            type: 'simpleObjectWithType',
        );

        return $object;
    }

    public static function createDeepObject(): Shared\DeepObject
    {
        $simpleObj = Helpers::createSimpleObject();

        $deep = new Shared\DeepObject(
            any: $simpleObj,
            str: 'test',
            bool: true,
            int: 1,
            num: 1.1,
            obj: $simpleObj,
            arr: [$simpleObj, $simpleObj],
            map: ['key' => $simpleObj]
        );

        return $deep;
    }

    public static function createDeepObjectCamelCase(): Shared\DeepObjectCamelCase
    {
        $simpleObj = Helpers::createSimpleObjectCamelCase();

        $deep = new Shared\DeepObjectCamelCase(
            anyVal: $simpleObj,
            strVal: 'test',
            boolVal: true,
            intVal: 1,
            numVal: 1.1,
            objVal: $simpleObj,
            arrVal: [$simpleObj, $simpleObj],
            mapVal: ['key' => $simpleObj]
        );

        return $deep;
    }

    public static function createDeepObjectWithType(): Shared\DeepObjectWithType
    {
        $simpleObj = Helpers::createSimpleObject();

        $deep = new Shared\DeepObjectWithType(
            any: $simpleObj,
            str: 'test',
            bool: true,
            int: 1,
            num: 1.1,
            obj: $simpleObj,
            arr: [$simpleObj, $simpleObj],
            map: ['key' => $simpleObj],
            type: 'deepObjectWithType',
        );

        return $deep;
    }

    public static function createSimpleObjectWithNonStandardTypeName(): Shared\SimpleObjectWithNonStandardTypeName
    {
        $object = new Shared\SimpleObjectWithNonStandardTypeName(
            any: 'any',
            bool: true,
            boolOpt: true,
            date: LocalDate::parse('2020-01-01'),
            dateTime: \DateTime::createFromFormat('Y-m-d\TH:i:s.up', '2020-01-01T00:00:00.000001Z'),
            enum: Shared\Enum::One,
            float32: 1.1,
            int: 1,
            int32: 1,
            int32Enum: Shared\SimpleObjectWithNonStandardTypeNameInt32Enum::FiftyFive,
            intEnum: Shared\SimpleObjectWithNonStandardTypeNameIntEnum::Second,
            intOptNull: null,
            num: 1.1,
            numOptNull: null,
            str: 'test',
            strOpt: 'testOptional',
            objType: 'simpleObjectWithNonStandardTypeName'
        );
        return $object;
    }

    public static function assertEquivalent($test, $obj, $other, string $path = 'root'): void
    {
        if (is_array($obj) && is_array($other)) {
            $test->assertEquals(
                count($obj),
                count($other),
                sprintf('Array count mismatch at %s: expected %d elements, got %d', $path, count($obj), count($other))
            );
            foreach ($obj as $key => $value) {
                $test->assertTrue(
                    array_key_exists($key, $other),
                    sprintf('Missing key "%s" at %s', $key, $path)
                );
                Helpers::assertEquivalent($test, $value, $other[$key], $path . '.' . $key);
            }
            return;
        }

        if (is_array($obj) !== is_array($other)) {
            $test->fail(sprintf('Type mismatch at %s: expected %s, got %s', $path, gettype($obj), gettype($other)));
        }

        $test->assertEquals($obj, $other);
    }

    public static function endToEndObject($obj)
    {
        $t = get_class($obj);
        $serializer = Utils\JSON::createSerializer();
        $serialized = $serializer->serialize($obj, 'json');
        $newObj = $serializer->deserialize($serialized, $t, 'json', DeserializationContext::create()->setRequireAllRequiredProperties(true));
        return $newObj;
    }
}
