declare(strict_types=1);

namespace Speakeasy\Client\Credentials\Tests;

use Speakeasy\Client\Credentials\Tests\CommonHelpers;
use Speakeasy\Client\Credentials\Models\Components\Security;
use Speakeasy\Client\Credentials\Hooks\ClientCredentialsOAuth2Scope;
use Speakeasy\Client\Credentials\Utils\Utils;
use PHPUnit\Framework\TestCase;

final class ClientCredentialsHookTest extends TestCase
{
    public function testUrljoinResolvesTokenUrlAgainstBaseUrl(): void
    {
        // absolute-path token URLs replace the base URL's path
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com/api', '/auth/token'));
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com/api/', '/auth/token'));
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com/api/v1', '/auth/token'));
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com', '/auth/token'));
        $this->assertEquals('https://example.com:8443/auth/token', Utils::urljoin('https://example.com:8443/api', '/auth/token'));
        $this->assertEquals('https://example.com/auth/token?a=b', Utils::urljoin('https://example.com/api', '/auth/token?a=b'));

        // relative token URLs resolve against the base URL's path
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com', 'auth/token'));
        $this->assertEquals('https://example.com/api/auth/token', Utils::urljoin('https://example.com/api/', 'auth/token'));
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com/api', 'auth/token'));

        // fully-qualified token URLs win outright, host-only URLs keep their empty path
        $this->assertEquals('https://auth.example.com/token', Utils::urljoin('https://example.com/api', 'https://auth.example.com/token'));
        $this->assertEquals('https://auth.example.com', Utils::urljoin('https://example.com/api', 'https://auth.example.com'));

        // an empty token URL leaves the base URL untouched, query and fragment included
        $this->assertEquals('https://example.com/api', Utils::urljoin('https://example.com/api', ''));
        $this->assertEquals('https://example.com/api?a=b#f', Utils::urljoin('https://example.com/api?a=b#f', ''));

        // parent segments above the root are dropped
        $this->assertEquals('https://example.com/auth/token', Utils::urljoin('https://example.com/api', '../auth/token'));
        $this->assertEquals('https://example.com/x', Utils::urljoin('https://example.com/', '../../x'));
    }

    public function testClientCredentialsHookSuccessfullyAuthenticates(): void
    {
        CommonHelpers::recordTest('hooks-client-credentials-success');

        $sdk = \Speakeasy\Client\Credentials\SDK::builder()
            ->setServerUrl(CommonHelpers::getApiTestServiceUrl())
            ->setSecurity(
            new Security(
                clientID: 'speakeasy-sdks',
                clientSecret: 'supersecret-'.CommonHelpers::randSequence(10)
            )
        )->build();

        $response = $sdk->hooks->authenticatedRequest(null);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertFalse($response->httpMeta->response->hasHeader("clientId"));
        $this->assertFalse($response->httpMeta->response->hasHeader("clientSecret"));
    }

    public function testClientCredentialsHookSuccessfullyAuthenticatesGlobalServer(): void
    {
        CommonHelpers::recordTest('hooks-client-credentials-success-global-server');

        $sdk = \Speakeasy\Client\Credentials\SDK::builder()
            ->setSecurity(
            new Security(
                clientID: 'speakeasy-sdks',
                clientSecret: 'supersecret-'.CommonHelpers::randSequence(10)
            )
        )->build();

        $response = $sdk->hooks->authenticatedRequestGlobalServer(null);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertFalse($response->httpMeta->response->hasHeader("clientId"));
        $this->assertFalse($response->httpMeta->response->hasHeader("clientSecret"));
    }

    public function testClientCredentialsHookSuccessfullyAuthenticatesWithAltTokenUrl(): void
    {
        CommonHelpers::recordTest('hooks-client-credentials-success-alt-token-url');

        $log = [];
        $client = createRequestRecorderClient($log);
        $tokenUrl = '/clientcredentials/alt/token';
        $sdk = \Speakeasy\Client\Credentials\SDK::builder()
            ->setServerUrl(CommonHelpers::getApiTestServiceUrl())
            ->setClient($client)
            ->setSecurity(new Security(
                clientID: 'speakeasy-sdks',
                clientSecret: 'supersecret-'.CommonHelpers::randSequence(10),
                tokenURL: $tokenUrl,
                scopes: ['alt:one', ClientCredentialsOAuth2Scope::AltTwo->value]
            ))->build();

        $response = $sdk->hooks->authenticatedRequest(null);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());

        // since the token is already expired, a new one should be requested
        $response = $sdk->hooks->authenticatedRequest(null);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());

        print_r($log);
        $tokenRequests = array_filter($log, function($entry) {
                $body = $entry->requestBody ?? '';
                return str_contains($body, 'grant_type=client_credentials') &&
                        str_contains($body, 'scope=alt%3Aone+alt%3Atwo');
            }
        );
        $this->assertCount(2, $tokenRequests);
    }

    public function testClientCredentialsHookNoScopes(): void
    {
        CommonHelpers::recordTest('hooks-client-credentials-no-scopes');

        $clientID = 'speakeasy-sdks';
        $clientSecret = 'supersecret-'.CommonHelpers::randSequence(10);

        $sdk = \Speakeasy\Client\Credentials\SDK::builder()
            ->setServerUrl(CommonHelpers::getApiTestServiceUrl())
            ->setSecurity(
            new Security(
                clientID: $clientID,
                clientSecret: $clientSecret
            )
        )->build();

        // expected to fail since the token endpoint requires 'read' and 'write' scopes
        // but the authenticatedRequestNoScopes operation does not specify any.
        try {
            $sdk->hooks->authenticatedRequestNoScopes(null);
            $this->fail('Expected exception to be thrown');
        } catch (\Exception $e) {
            $this->assertInstanceOf(\Exception::class, $e);
            $this->assertEquals(400, $e->getCode());
            $this->assertEquals('An error occurred while calling BeforeRequest hook.', $e->getMessage());
            $this->assertNotNull($e->getPrevious());
            $this->assertStringContainsString('empty_scopes', $e->getPrevious()->getMessage());
        }

        // same check but this time we override the default scopes with an empty list
        $sdk = \Speakeasy\Client\Credentials\SDK::builder()
            ->setServerUrl(CommonHelpers::getApiTestServiceUrl())
            ->setSecurity(
            new Security(
                clientID: $clientID,
                clientSecret: $clientSecret,
                scopes: [] // overrides global scopes
            )
        )->build();

        try {
            $sdk->hooks->authenticatedRequest(null); // would normally require [read, write]
            $this->fail('Expected exception to be thrown');
        } catch (\Exception $e) {
            $this->assertInstanceOf(\Exception::class, $e);
            $this->assertEquals(400, $e->getCode());
            $this->assertEquals('An error occurred while calling BeforeRequest hook.', $e->getMessage());
            $this->assertNotNull($e->getPrevious());
            $this->assertStringContainsString('empty_scopes', $e->getPrevious()->getMessage());
        }

        // now use a different tokenUrl that will allow no scopes to be requested
        $tokenUrl = '/clientcredentials/token?expires_in=90&skip_scopes=true';
        $log = [];
        $sdk = \Speakeasy\Client\Credentials\SDK::builder()
            ->setServerUrl(CommonHelpers::getApiTestServiceUrl())
            ->setClient(createRequestRecorderClient($log))
            ->setSecurity(new Security(
                clientID: $clientID,
                clientSecret: $clientSecret,
                tokenURL: $tokenUrl
            ))->build();

        $response = $sdk->hooks->authenticatedRequestNoScopes(null);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());

        // since the token is not expired, it should be reused on subsequent call
        $response = $sdk->hooks->authenticatedRequestNoScopes(null);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());

        $tokenRequests = array_filter($log, fn($entry) => str_contains($entry->requestUri, $tokenUrl));
        $this->assertCount(1, $tokenRequests);
        $tokenRequest = array_values($tokenRequests)[0];
        $this->assertStringNotContainsString('scope=', $tokenRequest->requestBody);
    }
}
