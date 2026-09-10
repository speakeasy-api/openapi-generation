declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Tests\Helpers\Helpers;
use PHPUnit\Framework\TestCase;
use ReflectionMethod;

final class FlatteningAdditionalTest extends TestCase
{
    /**
     * Tests that nullable-but-required body parameters are ordered after
     * required path parameters in the generated method signature.
     * PHP 8.0+ deprecates required parameters following optional ones.
     */
    public function testNullableBodyWithRequiredParamOrdering(): void
    {
        CommonHelpers::recordTest('flattening-nullable-body-with-required-param-ordering');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $method = new ReflectionMethod($sdk->flattening, 'nullableBodyWithRequiredParam');
        $params = $method->getParameters();

        // Filter out internal params (serverURL, urlOverride, options)
        $userParams = array_filter($params, fn($p) => !in_array($p->getName(), ['serverURL', 'urlOverride', 'options']));
        $userParams = array_values($userParams);

        // requiredParam must come before the nullable body
        $this->assertCount(2, $userParams, 'Expected 2 user parameters');
        $this->assertEquals('requiredParam', $userParams[0]->getName(), 'Required path param should be first');
        $this->assertFalse($userParams[0]->isOptional(), 'requiredParam should not be optional');

        $this->assertEquals('nullableBodyObject', $userParams[1]->getName(), 'Nullable body should be second');
        $this->assertTrue($userParams[1]->isOptional(), 'Nullable body should be optional (has default = null)');
    }
}
