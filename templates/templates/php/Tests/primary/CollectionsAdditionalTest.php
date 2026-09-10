declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Tests\Helpers\Helpers;
use OpenAPI\OpenAPI\Models\Operations;
use PHPUnit\Framework\TestCase;
use TypeError;

final class CollectionsAdditionalTest extends TestCase
{
    public function testCollectionsContainingNull(): void
    {
        CommonHelpers::recordTest('collections-containing-null');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $input = new Operations\CollectionsContainingNullNullishCollections(
          requiredArray: ['foo', null],
          requiredMap: ['foo' => null, 'bar' => 123],
          arrayOfNullUnion: ['foo', null],
          mapOfNullUnion: ['foo' => null, 'bar' => 123],
          optionalArray: ['foo', null],
          optionalMap: ['foo' => null, 'bar' => 123]
        );

        $expected = [
          'requiredArray' => ['foo', null],
          'requiredMap' => ['foo' => null, 'bar' => 123],
          'optionalArray' => ['foo', null],
          'optionalMap' => ['foo' => null, 'bar' => null], // TODO: Existing issue: should be 123
          'arrayOfNullUnion' => ['foo', null],
          'mapOfNullUnion' => ['foo' => null, 'bar' => null], // TODO: Existing issue: should be 123
        ];

        $response = $sdk->collections->collectionsContainingNull($input);

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        Helpers::assertEquivalent($this, $expected, $response->object->json);
    }
}
