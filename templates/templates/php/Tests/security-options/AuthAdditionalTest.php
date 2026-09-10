declare(strict_types=1);

namespace Speakeasy\SecurityOptions\Tests;

use Speakeasy\SecurityOptions\Tests\CommonHelpers;
use Speakeasy\SecurityOptions\Models\Shared;
use PHPUnit\Framework\TestCase;

final class AuthAdditionalTest extends TestCase
{
    public function testGlobalSecurityBasicHttpSuccess(): void
    {
        CommonHelpers::recordTest('auth-basic-http-global-option');

        // Expected to succeed since BasicHTTP takes priority over AccessToken in global security definition
        $sdk = \Speakeasy\SecurityOptions\SDK::builder()
            ->setSecurity(new Shared\Security(
                basicHttp: new Shared\BasicHttp(
                    username: 'testUser',
                    password: 'testPass',
                ),
                accessToken: new Shared\AccessToken(
                    accessToken: 'Bearer ignored',
                ),
            ))
            ->build();

        $response = $sdk->auth->globalSecurityOptionBasicHttp();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->basicAuth);
        $this->assertTrue($response->basicAuth->authenticated);
        $this->assertEquals('testUser', $response->basicAuth->user);
    }

    public function testGlobalSecurityFieldsOrdering(): void
    {
        CommonHelpers::recordTest('auth-global-security-option-fields-ordering');

        // When `maintainOpenApiOrder: false` is set, security fields are sorted in alphabetical order.
        // Since the first non-nil field is selected ApiKeyAuth takes precedence over BasicHttp,
        // which is not valid authentication for this endpoint.
        $sdk = \Speakeasy\SecurityOptions\SDK::builder()
            ->setSecurity(new Shared\Security(
                apiKeyAuth: new Shared\ApiKeyAuth(
                    apiKeyAuth: 'Bearer test_api_key',
                ),
                basicHttp: new Shared\BasicHttp(
                    username: 'testUser',
                    password: 'testPass',
                ),
            ))
            ->build();

        try {
            $sdk->auth->globalSecurityOptionBasicHttp();
            $this->fail('Expected APIException was not thrown');
        } catch (\Speakeasy\SecurityOptions\Models\Errors\APIException $e) {
            $this->assertEquals(401, $e->getCode());
        }
    }

    public function testHoistedSecurityAccessTokenFirst(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-option-access-token-first');

        $sdk = \Speakeasy\SecurityOptions\SDK::builder()
            ->setSecurity(new Shared\Security(
                apiKeyAuth: new Shared\ApiKeyAuth(
                    apiKeyAuth: 'testApiKey',
                ),
                basicHttp: new Shared\BasicHttp(
                    username: 'username',
                    password: 'password',
                ),
                accessToken: new Shared\AccessToken(
                    accessToken: 'Bearer ghp_xxxx',
                ),
            ))
            ->build();

        $response = $sdk->auth->hoistedSecurityOptionAccessTokenFirst();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->tokenAuthResponse);
        $this->assertEquals('Bearer ghp_xxxx', $response->tokenAuthResponse->token);
    }

    public function testHoistedSecurityApiKeyFirst(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-option-api-key-first');

        $sdk = \Speakeasy\SecurityOptions\SDK::builder()
            ->setSecurity(new Shared\Security(
                apiKeyAuth: new Shared\ApiKeyAuth(
                    apiKeyAuth: 'testApiKey',
                ),
                basicHttp: new Shared\BasicHttp(
                    username: 'username',
                    password: 'password',
                ),
                accessToken: new Shared\AccessToken(
                    accessToken: 'Bearer ghp_xxxx',
                ),
            ))
            ->build();

        $response = $sdk->auth->hoistedSecurityOptionApiKeyFirst();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->tokenAuthResponse);
        $this->assertEquals('testApiKey', $response->tokenAuthResponse->token);
    }

    public function testHoistedSecurityBasicHttpOnly(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-option-basic-http-only');

        $sdk = \Speakeasy\SecurityOptions\SDK::builder()
            ->setSecurity(new Shared\Security(
                apiKeyAuth: new Shared\ApiKeyAuth(
                    apiKeyAuth: 'testApiKey',
                ),
                basicHttp: new Shared\BasicHttp(
                    username: 'testUser',
                    password: 'testPass',
                ),
                accessToken: new Shared\AccessToken(
                    accessToken: 'Bearer ghp_xxxx',
                ),
            ))
            ->build();

        $response = $sdk->auth->hoistedSecurityOptionBasicHttpOnly();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->basicAuth);
        $this->assertTrue($response->basicAuth->authenticated);
        $this->assertEquals('testUser', $response->basicAuth->user);
    }

    public function testHoistedSecurityInvalidOption(): void
    {
        CommonHelpers::recordTest('auth-hoisted-security-invalid-option');

        $sdk = \Speakeasy\SecurityOptions\SDK::builder()
            ->setSecurity(new Shared\Security(
                basicHttp: new Shared\BasicHttp(
                    username: 'username',
                    password: 'password',
                ),
            ))
            ->build();

        try {
            $sdk->auth->hoistedSecurityOptionAccessTokenFirst();
            $this->fail('Expected APIException was not thrown');
        } catch (\Speakeasy\SecurityOptions\Models\Errors\APIException $e) {
            $this->assertEquals(401, $e->getCode());
        }
    }
}
