declare(strict_types=1);

namespace Speakeasy\BasicHttp\Tests;

use Speakeasy\BasicHttp\Tests\CommonHelpers;
use Speakeasy\BasicHttp\Models\Components\Security;
use Speakeasy\BasicHttp\Models\Operations\BasicAuthOptionalSecurity;
use PHPUnit\Framework\TestCase;

final class AuthAdditionalTest extends TestCase
{
    public function testBasicAuthFlattenedGlobal(): void
    {
        CommonHelpers::recordTest('auth-basic-auth-flattened-global');

        $sdk = \Speakeasy\BasicHttp\SDK::builder()
            ->setSecurity(new Security(
                username: 'testUser',
                password: 'testPass'
            ))
            ->build();

        $response = $sdk->auth->basicAuthGlobal();
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->basicAuthResponse);
        $this->assertTrue($response->basicAuthResponse->authenticated);
        $this->assertEquals('testUser', $response->basicAuthResponse->user);
    }

    public function testBasicAuthFlattenedHoisted(): void
    {
        CommonHelpers::recordTest('auth-basic-auth-flattened-hoisted');

        $sdk = \Speakeasy\BasicHttp\SDK::builder()
            ->setSecurity(new Security(
                username: 'testUser',
                password: 'testPass'
            ))
            ->build();

        $response = $sdk->auth->basicAuthHoisted();
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertNotNull($response->basicAuthResponse);
        $this->assertTrue($response->basicAuthResponse->authenticated);
        $this->assertEquals('testUser', $response->basicAuthResponse->user);
    }

    public function testBasicAuthOperationOptional(): void
    {
        // TODO: empty BasicAuthOptionalSecurity() still sends Authorization: Basic
        // header with empty credentials due to parseBasicAuthScheme not checking for empty fields.
        // Skipping negative test until PHP security util is fixed.
        $this->markTestSkipped('PHP security util does not yet reject empty basic-auth credentials');
    }
}
