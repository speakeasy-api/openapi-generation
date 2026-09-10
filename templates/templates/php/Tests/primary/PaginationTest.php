declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Tests\Helpers\Helpers;
use OpenAPI\OpenAPI\Utils\JSON;
use OpenAPI\OpenAPI\Models\Shared;
use OpenAPI\OpenAPI\Models\Operations;
use PHPUnit\Framework\TestCase;

final class PaginationTest extends TestCase
{
    public function testPaginationLimitOffsetPageParams(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-page-params');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $serverLimit = 20;
        $responses = $sdk->pagination->paginationLimitOffsetPageParams(page: 1);
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->httpMeta);
                $this->assertNotNull($response->httpMeta->response);
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($serverLimit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->httpMeta);
                $this->assertNotNull($response->httpMeta->response);
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 2){
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationLimitOffsetUnionOutputPageParams(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-union-output-page-params');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $serverLimit = 20;
        $responses = $sdk->pagination->paginationLimitOffsetUnionOutputPageParams(page: 1);

        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->httpMeta);
                $this->assertNotNull($response->httpMeta->response);
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($serverLimit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertNotNull($response);
                $this->assertNotNull($response->httpMeta);
                $this->assertNotNull($response->httpMeta->response);
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 2){
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }


    public function testPaginationLimitOffsetPageBody(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-page-body');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationLimitOffsetPageBody(request: new Shared\LimitOffsetConfig(limit: $limit, page: 1));
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count == 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertNotNull($response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(20 - $limit, $response->res->resultArray);
            } else if ($count === 2){
                $this->assertCount(0, $response->res->resultArray);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationLimitOffsetPageBodyNullable(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-page-body-nullable');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        // first request sends a null body; later pages materialize one with
        // the advanced page (wire sequence enforced by the test service)
        $responses = $sdk->pagination->paginationLimitOffsetPageBodyNullable(request: null);

        $pages = [];
        foreach ($responses as $response) {
            $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
            $this->assertNotNull($response->res);
            $pages[] = $response->res->resultArray;
        }

        $this->assertEquals([
            [0, 1, 2, 3, 4, 5, 6],
            [7, 8, 9, 10, 11, 12, 13],
            [14, 15, 16, 17, 18, 19],
        ], $pages);
    }

    public function testPaginationLimitOffsetDeepOutputsPageBody(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-deep-outputs-page-body');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationLimitOffsetDeepOutputsPageBody(request: new Shared\LimitOffsetConfig(limit: $limit, page: 1));
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count == 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationLimitOffsetOffsetParams(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-offset-params');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationLimitOffsetOffsetParams(limit: $limit, offset: 0);
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count == 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count == 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count == 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationLimitOffsetOffsetBody(): void
    {
        CommonHelpers::recordTest('pagination-limit-offset-offset-body');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationLimitOffsetOffsetBody(request: new Shared\LimitOffsetConfig(limit: $limit, offset: 0));
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count == 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count == 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count == 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationCursorParams(): void
    {
        CommonHelpers::recordTest('pagination-cursor-params');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationCursorParams(cursor: -1);
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationCursorBody(): void
    {
        CommonHelpers::recordTest('pagination-cursor-body');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationCursorBody(request: new Operations\PaginationCursorBodyRequestBody(cursor: -1));
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationCursorNonNumeric(): void
    {
        CommonHelpers::recordTest('pagination-cursor-non-numeric');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $responses = $sdk->pagination->paginationCursorNonNumeric();
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(15, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(5, $response->res->resultArray);
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationNullableCursor(): void
    {
        CommonHelpers::recordTest('pagination-cursor-nullable-limit');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $defaultLimit = 10;
        $responses = $sdk->pagination->paginationCursorNullableLimit();

        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->paginationCursorNullableLimitNextCursor);
                $this->assertCount($defaultLimit, $response->paginationCursorNullableLimitNextCursor->results);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->paginationCursorNullableLimitNextCursor);
                $this->assertCount($defaultLimit, $response->paginationCursorNullableLimitNextCursor->results);
            }
            $count++;
        }
    }

    public function testPaginationCursorNonNumericNullable(): void
    {
        CommonHelpers::recordTest('pagination-cursor-non-numeric-nullable');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $responses = $sdk->pagination->paginationCursorNonNumericNullable("2");
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(15, $response->res->resultArray);
                $this->assertEquals("17", $response->res->cursor);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(2, $response->res->resultArray);
                $this->assertNull($response->res->cursor);
            } else if ($count === 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationCursorNonNumericEmptyString(): void
    {
        CommonHelpers::recordTest('pagination-cursor-non-numeric-empty-string');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $responses = $sdk->pagination->paginationCursorNonNumericEmptyString("", "2");
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(15, $response->res->resultArray);
                $this->assertEquals("17", $response->res->cursor);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(2, $response->res->resultArray);
                $this->assertEquals("", $response->res->cursor);
            } else if ($count === 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationUrl(): void
    {
        CommonHelpers::recordTest('pagination-url');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $responses = $sdk->pagination->paginationUrlParams(attempts: 3);
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(9, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(6, $response->res->resultArray);
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(3, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }

        $responses = $sdk->pagination->paginationUrlParams(attempts: 3, isReferencePath: 'true');
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(9, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(6, $response->res->resultArray);
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(3, $response->res->resultArray);
            } else if ($count === 3){
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationAmbiguousInput(): void
    {
        CommonHelpers::recordTest('pagination-ambiguous-input');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $rb = new Operations\PaginationAmbiguousInputRequestBody(
            cursor: -1
        );
        $responses = $sdk->pagination->paginationAmbiguousInput($rb);
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationOptionalSecurity(): void
    {
        CommonHelpers::recordTest('pagination-body-flattened-optional-security');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationBodyFlattenedOptionalSecurity(limit: 15, offset: 0, security: new Operations\PaginationBodyFlattenedOptionalSecuritySecurity('test'));
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertEquals(20-$limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationBodyFlattenedWithSecurity(): void
    {
        CommonHelpers::recordTest('pagination-body-flattened-with-security');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationBodyFlattenedWithSecurity(new Operations\PaginationBodyFlattenedWithSecuritySecurity(paginationAuth: 'test'), limit: $limit, offset: 0);
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertEquals(20-$limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationBodyWrappedRequest(): void
    {
        CommonHelpers::recordTest('pagination-body-wrapped-request');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $responses = $sdk->pagination->paginationBodyWrappedRequest(request: new Operations\PaginationBodyWrappedRequestRequest(
            limitOffsetConfig: new Shared\LimitOffsetConfig(limit: $limit, page: 1))
        );
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertEquals(20-$limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationWrappedOptionalBody(): void
    {
        CommonHelpers::recordTest('pagination-wrapped-optional-body');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $responses = $sdk->pagination->paginationWrappedOptionalBody();
    
        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(20, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 2) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }

    public function testPaginationEncapsulatedParameter(): void
    {
        CommonHelpers::recordTest('pagination-encapsulated-parameter');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $limit = 15;
        $rb = new Operations\PaginationEncapsulatedParameterRequest(
            cursor: -1
        );
        $responses = $sdk->pagination->paginationEncapsulatedParameter($rb);

        $count = 0;
        foreach ($responses as $response) {
            if ($count === 0) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount($limit, $response->res->resultArray);
            } else if ($count === 1) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertLessThan($limit, count($response->res->resultArray));
            } else if ($count === 2) {
                $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
                $this->assertNotNull($response->res);
                $this->assertCount(0, $response->res->resultArray);
            } else if ($count === 3) {
                $this->assertNull($response);
            } else {
                $this->fail('Unexpected response');
            }
            $count++;
        }
    }
}
