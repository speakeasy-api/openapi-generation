declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use PHPUnit\Framework\TestCase;

final class AuthAdditionalTest extends TestCase
{
    public function testGlobalSecurityFieldsOrdering(): void
    {
        CommonHelpers::recordTest('auth-global-security-fields-ordering');

        // When `maintainOpenApiOrder: false` is set, security fields are sorted in alphabetical order.
        // Since the first non-nil field is selected ApiKeyAuth takes precedence over BasicHttp,
        // which is not valid authentication for this endpoint.
        $security = new \OpenAPI\OpenAPI\Models\Shared\Security(
            apiKeyAuth: 'testApiKey',
            basicHttp: new \OpenAPI\OpenAPI\Models\Shared\SchemeBasicHTTP(
                username: 'testUser',
                password: 'testPass',
            ),
        );
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setSecurity($security)->build();

        try {
            $sdk->auth->globalSecurityBasicHttp();
            $this->fail('Expected APIException was not thrown');
        } catch (\OpenAPI\OpenAPI\Models\Errors\APIException $e) {
            $this->assertEquals(401, $e->getCode());
        }
    }

    public function testHoistedSecurityAccessTokenOnly(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-access-token-only');

        $security = new \OpenAPI\OpenAPI\Models\Shared\Security(
            apiKeyAuth: 'testApiKey',
            basicHttp: new \OpenAPI\OpenAPI\Models\Shared\SchemeBasicHTTP(
                username: 'testUser',
                password: 'testPass',
            ),
            accessToken: 'Bearer ghp_xxxx',
        );
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setSecurity($security)->build();

        $response = $sdk->auth->hoistedSecurityAccessTokenOnly();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->token);
        $this->assertEquals('Bearer ghp_xxxx', $response->token->token);
    }

    public function testHoistedSecurityAccessTokenFirst(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-access-token-first');

        $security = new \OpenAPI\OpenAPI\Models\Shared\Security(
            apiKeyAuth: 'testApiKey',
            accessToken: 'Bearer ghp_xxxx',
        );
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setSecurity($security)->build();

        $response = $sdk->auth->hoistedSecurityAccessTokenFirst();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->token);
        $this->assertEquals('Bearer ghp_xxxx', $response->token->token);
    }

    public function testHoistedSecurityApiKeyFirst(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-api-key-first');

        $security = new \OpenAPI\OpenAPI\Models\Shared\Security(
            apiKeyAuth: 'testApiKey',
            accessToken: 'Bearer ghp_xxxx',
        );
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setSecurity($security)->build();

        $response = $sdk->auth->hoistedSecurityApiKeyFirst();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->token);
        $this->assertEquals('testApiKey', $response->token->token);
    }

    public function testHoistedSecurityBasicHttpOnly(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-basic-http-only');

        $security = new \OpenAPI\OpenAPI\Models\Shared\Security(
            basicHttp: new \OpenAPI\OpenAPI\Models\Shared\SchemeBasicHTTP(
                username: 'testUser',
                password: 'testPass',
            ),
            apiKeyAuth: 'testApiKey',
            accessToken: 'Bearer ghp_xxxx',
        );
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setSecurity($security)->build();

        $response = $sdk->auth->hoistedSecurityBasicHttpOnly();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->basicAuth);
        $this->assertTrue($response->basicAuth->authenticated);
        $this->assertEquals('testUser', $response->basicAuth->user);
    }

    public function testHoistedSecurityInvalidField(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-invalid-field');

        // Provide only BasicHttp — not valid for accessTokenFirst which expects bearer/apiKey
        $security = new \OpenAPI\OpenAPI\Models\Shared\Security(
            basicHttp: new \OpenAPI\OpenAPI\Models\Shared\SchemeBasicHTTP(
                username: 'user',
                password: 'pass',
            ),
        );
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setSecurity($security)->build();

        try {
            $sdk->auth->hoistedSecurityAccessTokenFirst();
            $this->fail('Expected APIException was not thrown');
        } catch (\OpenAPI\OpenAPI\Models\Errors\APIException $e) {
            $this->assertEquals(401, $e->getCode());
        }
    }
}
