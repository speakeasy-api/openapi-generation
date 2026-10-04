declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Tests\Helpers\Helpers;
use OpenAPI\OpenAPI\Models\Shared;
use PHPUnit\Framework\TestCase;

final class RequestBodiesAdditionalTest extends TestCase
{
    public function testRequestBodiesBase64FileInputIdempotent(): void
    {
        CommonHelpers::recordTest('request-bodies-base64-file-input-idempotent');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $preEncoded = base64_encode("\xFF\xFE\x00\x01binary\xC3\x28");

        $request = new Shared\Base64InputFileModeRequest(
            dataByte: $preEncoded,
            dataContentEncoding: $preEncoded,
            dataPlain: 'plain-idempotent',
        );

        $first = $sdk->requestBodies->postBase64InputMode($request);

        $this->assertNotNull($first);
        $this->assertEquals(200, $first->httpMeta->response->getStatusCode());
        $this->assertNotNull($first->res);
        $this->assertEquals($preEncoded, $first->res->json->dataByte);
        $this->assertEquals($preEncoded, $first->res->json->dataContentEncoding);
        $this->assertEquals('plain-idempotent', $first->res->json->dataPlain);

        $second = $sdk->requestBodies->postBase64InputMode($request);

        $this->assertNotNull($second);
        $this->assertEquals(200, $second->httpMeta->response->getStatusCode());
        $this->assertNotNull($second->res);
        $this->assertEquals($first->res->json, $second->res->json);
    }
}
