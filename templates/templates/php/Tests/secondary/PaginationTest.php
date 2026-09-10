declare(strict_types=1);

namespace alphabetically\early\Tests;


use alphabetically\early\Tests\CommonHelpers;
use PHPUnit\Framework\TestCase;

final class PaginationTest extends TestCase
{

    public function testPaginationLimitOffsetPageParamsFlat(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-page-params-flat');

        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        $serverLimit = 20;
        $responses = $sdk->pagination->paginationLimitOffsetPageParams(page: 1);

        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->result);
                $this->assertCount($serverLimit, $response->result->resultArray);
            } else if ($count === 1) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->result);
                $this->assertCount(0, $response->result->resultArray);
            } else if ($count === 2){
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationLimitOffsetUnionOutputPageParamsFlat(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-union-output-page-params-flat');

        $sdk = \alphabetically\early\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\alphabetically\early\SDK::class, $sdk);

        $serverLimit = 20;
        $responses = $sdk->pagination->paginationLimitOffsetUnionOutputPageParams(page: 1);

        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->result);
                $this->assertCount($serverLimit, $response->result->resultArray);
            } else if ($count === 1) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->result);
                $this->assertCount(0, $response->result->resultArray);
            } else if ($count === 2){
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }
}
