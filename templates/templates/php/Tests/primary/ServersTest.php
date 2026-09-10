declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\ServerSomething;
use OpenAPI\OpenAPI\Utils;
use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Tests\Helpers\Helpers;
use PHPUnit\Framework\TestCase;

final class ServersTest extends TestCase
{
    public function testSelectGlobalServerValid(): void
    {
        CommonHelpers::recordTest('servers-select-global-server-valid');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->selectGlobalServer();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testSelectGlobalServerBroken(): void
    {
        CommonHelpers::recordTest('servers-select-global-server-broken');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerIndex(1)->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(\GuzzleHttp\Exception\ConnectException::class);

        $sdk->servers->selectGlobalServer();
    }

    public function testSelectServerWithIDDefault(): void
    {
        CommonHelpers::recordTest('servers-select-server-with-id-default');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->selectServerWithID(Helpers::getHttpBinUrl());

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testSelectServerWithIDValid(): void
    {
        CommonHelpers::recordTest('servers-select-server-with-id-valid');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->selectServerWithID(Helpers::getHttpBinUrl());

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testSelectServerWithIDBroken(): void
    {
        CommonHelpers::recordTest('servers-select-server-with-id-broken');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(\OpenAPI\OpenAPI\SDK::SERVERS[1])->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(\GuzzleHttp\Exception\ConnectException::class);

        $sdk->servers->selectServerWithID(\OpenAPI\OpenAPI\Servers::SELECT_SERVER_WITH_ID_SERVERS[\OpenAPI\OpenAPI\Servers::SELECT_SERVER_WITH_ID_SERVER_BROKEN]);
    }

    public function testServerWithTemplatesGlobal(): void
    {
        CommonHelpers::recordTest('servers-server-with-templates-global');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerIndex(2)->setHostname('localhost')->setPort(Helpers::getHttpBinPort())->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->serverWithTemplatesGlobal();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testServerWithTemplatesGlobalDefaults(): void
    {
        CommonHelpers::recordTest('servers-server-with-templates-global-defaults');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerIndex(2)->setHostname('localhost')->setPort(Helpers::getHttpBinPort())->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->serverWithTemplatesGlobal();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testServerWithTemplatesGlobalEnum(): void
    {
        CommonHelpers::recordTest('servers-server-with-templates-global-enum');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl() . '/anything/' . ServerSomething::SomethingElseAgain->value)->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->serverWithTemplatesGlobal();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testServerWithTemplates(): void
    {
        CommonHelpers::recordTest('servers-server-with-templates');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->serverWithTemplates(\OpenAPI\OpenAPI\Utils\Utils::templateUrl(\OpenAPI\OpenAPI\Servers::SERVER_WITH_TEMPLATES_SERVERS[0], [
            'hostname' => 'localhost',
            'port' => Helpers::getHttpBinPort(),
        ]));

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testServerWithTemplatesDefaults(): void
    {
        CommonHelpers::recordTest('servers-server-with-templates-defaults');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->serverWithTemplates(Utils\Utils::templateUrl(\OpenAPI\OpenAPI\Servers::SERVER_WITH_TEMPLATES_SERVERS[0], [
            'hostname' => 'localhost',
            'port' => Helpers::getHttpBinPort(),
        ]));

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testServerByIDWithTemplates(): void
    {
        CommonHelpers::recordTest('servers-server-by-id-with-templates');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->serversByIDWithTemplates(Utils\Utils::templateUrl('http://{hostname}:{port}', [
            'hostname' => 'localhost',
            'port' => Helpers::getHttpBinPort(),
        ]));

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testGlobalServerWithTemplatedProtocol(): void
    {
        CommonHelpers::recordTest('servers-global-server-with-templated-protocol');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerIndex(4)->setProtocol('http')->setHostname('localhost')->setPort(Helpers::getHttpBinPort())->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->servers->selectGlobalServer();

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }


    public function testServerWithInvalidTemplatedProtocol(): void
    {
        CommonHelpers::recordTest('servers-global-server-with-invalid-templated-protocol');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerIndex(4)->setProtocol("invalid")->setHostname("localhost")->setPort(Helpers::getHttpBinPort())->build();
        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(\GuzzleHttp\Exception\RequestException::class);

        $sdk->servers->selectGlobalServer();
    }

    public function testServerWithProtocolTemplate(): void
    {
        CommonHelpers::recordTest('servers-server-with-protocol-template');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();

        $response = $sdk->servers->serverWithProtocolTemplate(Utils\Utils::templateUrl(
            \OpenAPI\OpenAPI\Servers::SERVER_WITH_PROTOCOL_TEMPLATE_SERVERS[\OpenAPI\OpenAPI\Servers::SERVER_WITH_PROTOCOL_TEMPLATE_SERVER_MAIN],
            [
                'protocol' => 'http',
                'hostname' => 'localhost',
                'port' => Helpers::getHttpBinPort(),
            ]
        ));

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
    }

    public function testServerWithInvalidProtocolTemplate(): void
    {
        CommonHelpers::recordTest('servers-server-with-invalid-protocol-template');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->build();

        $this->expectException(\GuzzleHttp\Exception\RequestException::class);

        $sdk->servers->serverWithProtocolTemplate(Utils\Utils::templateUrl(
            \OpenAPI\OpenAPI\Servers::SERVER_WITH_PROTOCOL_TEMPLATE_SERVERS[\OpenAPI\OpenAPI\Servers::SERVER_WITH_PROTOCOL_TEMPLATE_SERVER_MAIN],
            [
                'protocol' => 'invalid',
                'hostname' => 'localhost',
                'port' => Helpers::getHttpBinPort(),
            ]
        ));
    }

}
