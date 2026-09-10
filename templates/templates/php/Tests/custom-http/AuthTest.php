declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Models\Components\SchemeCustomHTTPSecurity;
use OpenAPI\OpenAPI\Models\Components\Role;
use PHPUnit\Framework\TestCase;

final class AuthTest extends TestCase
{
    public function testCustomHttpSchemeOnly(): void
    {
        CommonHelpers::recordTest('auth-custom-security-scheme-only');

        $testScopes = ['read:products', 'write:products'];

        $sdk = \OpenAPI\OpenAPI\SDK::builder()
            ->setServerUrl(CommonHelpers::getApiTestServiceUrl())
            ->setSecurity(new SchemeCustomHTTPSecurity(
                userID: 54321,
                role: Role::Manager,
                passphrase: 'secure-passphrase-123',
                accessCode: 104,
                scopes: $testScopes
            ))
            ->build();

        $response = $sdk->auth->customHttpOnly();
        $this->assertNotNull($response);
        $this->assertNotNull($response->object);
        $this->assertEquals('access_granted', $response->object->grant);
        $this->assertEquals($testScopes, $response->object->scopes);
    }
}
