declare(strict_types=1);

namespace alphabetically\early\Tests;

use alphabetically\early\Tests\CommonHelpers;
use alphabetically\early\Tests\RequestRecorderClient;
use alphabetically\early\Tests\RequestLogEntry;
use function alphabetically\early\Tests\createRequestRecorderClient;
use PHPUnit\Framework\TestCase;

final class AuthTest extends TestCase
{
    public function testNoAuth(): void
    {
        CommonHelpers::recordTest('auth-hoisted-no-auth-retained');

        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        $response = $sdk->auth->noAuth();
    }

    public function testBasicAuth(): void
    {
        CommonHelpers::recordTest('auth-hoisted-basic-auth');

        $security = new \alphabetically\early\Models\Shared\Security();
        $security->username = 'testUser';
        $security->password = 'testPass';

        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->setSecurity($security)->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        $response = $sdk->auth->basicAuth('testPass', 'testUser');

        $this->assertNotNull($response);
        $this->assertTrue($response->authenticated);
    }

    public function testMultipleMixedOptionsAuth(): void
    {
        CommonHelpers::recordTest('auth-hoisted-operation-auth-retained');
        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        $request = new \alphabetically\early\Models\Shared\AuthServiceRequestBody();
        $request->basicAuth = new \alphabetically\early\Models\Shared\AuthServiceRequestBodyBasicAuth(
            username: 'testUser',
            password: 'testPass',
        );

        $security = new \alphabetically\early\Models\Operations\MultipleMixedOptionsAuthSecurity();
        $security->basicAuth = new \alphabetically\early\Models\Shared\SchemeBasicAuth(
            username: 'testUser',
            password: 'testPass',
        );

        $sdk->authNew->multipleMixedOptionsAuth($security, $request);
    }

    public function testOperationLevelOauth2(): void
    {
        CommonHelpers::recordTest('auth-operation-level-oauth2');

        $log = [];
        $client = createRequestRecorderClient($log);
        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->setClient($client)
            ->build();
        $clientSecret = 'supersecret-'.CommonHelpers::randSequence(10);

        // A token should be requested with 'read', 'write' and 'erase' scopes.
        $sdk->hooks->authenticatedRequest(
            new \alphabetically\early\Models\Operations\AuthenticatedRequestSecurity(
                audience: '',
                clientID: 'speakeasy-sdks',
                clientSecret: $clientSecret,
            )
        );

        // This operation requires 'read' and 'write' scopes.
        // The same token should be reused since [read, write, erase] is a superset of [read, write].
        $sdk->hooks->authenticatedRequestUnflattened(
            new \alphabetically\early\Models\Operations\AuthenticatedRequestUnflattenedSecurity(
                clientCredentials: new \alphabetically\early\Models\Shared\SchemeClientCredentials(
                    audience: '',
                    clientID: 'speakeasy-sdks',
                    clientSecret: $clientSecret,
                )
            )
        );

        // Verify that only a single token was requested
        $tokenRequests = array_filter($log, function($entry) {
            return str_contains($entry->requestBody ?? '', 'grant_type=client_credentials');
        });
        $this->assertCount(1, $tokenRequests, 'Expected only a single token request');
        $firstRequest = reset($tokenRequests);
        $this->assertStringContainsString('scope=read+write+erase', $firstRequest->requestBody ?? '');
    }

    public function testCustomSecuritySchemeAddId(): void
    {
        CommonHelpers::recordTest('auth-custom-security-scheme-app-id');

        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        $response = $sdk->authNew->customSchemeAppId(
            security: new \alphabetically\early\Models\Operations\CustomSchemeAppIdSecurity(
                appId: 'testAppID',
                secret: 'testSecret'
            )
        );
    }
}
