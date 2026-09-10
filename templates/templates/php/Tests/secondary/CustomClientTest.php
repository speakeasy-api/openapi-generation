declare(strict_types=1);

namespace alphabetically\early\Tests;

use alphabetically\early\Tests\CommonHelpers;
use PHPUnit\Framework\TestCase;

final class CustomClientTest extends TestCase
{
    public function testCustomClientInjection(): void
    {
        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        // Use telemetry endpoint which returns headers in the response object
        $response = $sdk->telemetry->telemetryUserAgentGet();

        $this->assertNotNull($response);
        $this->assertIsArray($response->headers);
        
        // Verify that our custom client request header was added and echoed back
        $this->assertArrayHasKey('X-Custom-Client-Active', $response->headers);
        $this->assertEquals('true', $response->headers['X-Custom-Client-Active']);
    }

    public function testCustomClientWithAuthentication(): void
    {
        $security = new \alphabetically\early\Models\Shared\Security();
        $security->username = 'testUser';
        $security->password = 'testPass';

        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->setSecurity($security)->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        // Test with authenticated endpoint and then verify with telemetry
        $response = $sdk->auth->basicAuth('testPass', 'testUser');

        $this->assertNotNull($response);
        $this->assertTrue($response->authenticated);
        
        // Verify that our custom client works by checking telemetry headers
        $telemetryResponse = $sdk->telemetry->telemetryUserAgentGet();
        $this->assertNotNull($telemetryResponse);
        $this->assertIsArray($telemetryResponse->headers);
        $this->assertArrayHasKey('X-Custom-Client-Active', $telemetryResponse->headers);
        $this->assertEquals('true', $telemetryResponse->headers['X-Custom-Client-Active']);
    }
}
