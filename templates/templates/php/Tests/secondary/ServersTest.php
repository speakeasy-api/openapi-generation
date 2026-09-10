declare(strict_types=1);

namespace alphabetically\early\Tests;

use alphabetically\early\ServerSomething;
use alphabetically\early\Tests\CommonHelpers;
use PHPUnit\Framework\TestCase;

final class ServersTest extends TestCase
{

    public function testSelectGlobalServerByNameDefault(): void
    {
        CommonHelpers::recordTest('servers-select-global-server-by-name-default');

        $sdk = \alphabetically\early\SDK::builder()->build();

        $response = $sdk->servers->selectGlobalServer();
        $this->assertNotNull($response);
    }

    public function testSelectGlobalServerByNameValid(): void
    {
        CommonHelpers::recordTest('servers-select-global-server-by-name-valid');

        $sdk = \alphabetically\early\SDK::builder()->setServer(\alphabetically\early\SDK::SERVER_DEFAULT_SERVER)->build();

        $response = $sdk->servers->selectGlobalServer();
        $this->assertNotNull($response);
    }

    public function testSelectGlobalServerByNameBroken(): void
    {
        CommonHelpers::recordTest('servers-select-global-server-by-name-broken');

        $sdk = \alphabetically\early\SDK::builder()->setServer(\alphabetically\early\SDK::SERVER_BROKEN_SERVER)->build();

        $this->expectException(\GuzzleHttp\Exception\ConnectException::class);
        $sdk->servers->selectGlobalServer();
    }

}
