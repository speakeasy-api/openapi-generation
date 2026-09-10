declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use PHPUnit\Framework\TestCase;
use ReflectionMethod;

final class ParameterAdditionalTest extends TestCase
{
    /**
     * Helper: asserts no required parameter follows an optional one in
     * the given parameter list, which would trigger PHP 8.0+ deprecation.
     */
    private function assertNoRequiredAfterOptional(array $params): void
    {
        $optionalStarted = false;
        foreach ($params as $p) {
            if ($p->isOptional()) {
                $optionalStarted = true;
            } elseif ($optionalStarted) {
                $this->fail("Required parameter '{$p->getName()}' follows optional parameter");
            }
        }
    }

    /**
     * Helper: returns user-facing parameters (excludes internal 'options' param).
     */
    private function getUserParams(ReflectionMethod $method): array
    {
        $params = $method->getParameters();
        return array_values(array_filter($params, fn($p) => !in_array($p->getName(), ['options'])));
    }

    /**
     * Tests that when the request body is required, it comes before
     * path/query/header parameters in the generated method signature.
     */
    public function testParametersOrderingBodyFirst(): void
    {
        CommonHelpers::recordTest('parameters-ordering-body-first');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();
        $method = new ReflectionMethod($sdk->parameters, 'flatParametersOrdering');
        $userParams = $this->getUserParams($method);

        $this->assertEquals('requestBody', $userParams[0]->getName(), 'Required body should be first');
        $this->assertFalse($userParams[0]->isOptional(), 'requestBody should be required');
        $this->assertNoRequiredAfterOptional($userParams);
    }

    /**
     * PHP has no flatteningOrder config, so "parameters-first" with a required
     * body still produces body-first ordering (same as body-first test).
     */
    public function testParametersOrderingParametersFirst(): void
    {
        CommonHelpers::recordTest('parameters-ordering-parameters-first');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();
        $method = new ReflectionMethod($sdk->parameters, 'flatParametersOrdering');
        $userParams = $this->getUserParams($method);

        $this->assertEquals('requestBody', $userParams[0]->getName(), 'Required body should be first (PHP has no parameters-first config)');
        $this->assertFalse($userParams[0]->isOptional(), 'requestBody should be required');
        $this->assertNoRequiredAfterOptional($userParams);
    }

    /**
     * PHP has no flatteningOrder config, so "legacy" with a required body
     * produces the same body-first ordering.
     */
    public function testParametersOrderingWithLegacyFlatteningOrder(): void
    {
        CommonHelpers::recordTest('parameters-ordering-with-legacy-flattening-order');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();
        $method = new ReflectionMethod($sdk->parameters, 'flatParametersOrdering');
        $userParams = $this->getUserParams($method);

        $this->assertEquals('requestBody', $userParams[0]->getName(), 'Required body should be first (PHP has no legacy config)');
        $this->assertFalse($userParams[0]->isOptional(), 'requestBody should be required');
        $this->assertNoRequiredAfterOptional($userParams);
    }

    /**
     * Tests that when the request body is optional, required path/query/header
     * parameters come before the optional body in the generated method signature.
     */
    public function testParametersOrderingUsingOptionalRequestBodyParametersFirst(): void
    {
        CommonHelpers::recordTest('parameters-ordering-using-optional-request-body-parameters-first');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();
        $method = new ReflectionMethod($sdk->parameters, 'flatParametersOrderingUsingOptionalRequestBody');
        $userParams = $this->getUserParams($method);

        $this->assertNoRequiredAfterOptional($userParams);

        $requestBodyParam = null;
        foreach ($userParams as $p) {
            if ($p->getName() === 'requestBody') {
                $requestBodyParam = $p;
                break;
            }
        }
        $this->assertNotNull($requestBodyParam, 'requestBody parameter should exist');
        $this->assertTrue($requestBodyParam->isOptional(), 'requestBody should be optional');
    }

}
