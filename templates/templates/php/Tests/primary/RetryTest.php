declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use OpenAPI\OpenAPI\Tests\Helpers\Helpers;
use OpenAPI\OpenAPI\Models\Operations;
use OpenAPI\OpenAPI\Models\Errors;
use OpenAPI\OpenAPI\Utils;
use OpenAPI\OpenAPI\Utils\Retry;
use PHPUnit\Framework\TestCase;
use GuzzleHttp\Client;
use GuzzleHttp\HandlerStack;
use GuzzleHttp\Middleware;


final class RetryTest extends TestCase
{
    function testRetriesSucceeds()
    {
        CommonHelpers::recordTest('retries-succeeds');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->retries->retriesGet(requestId: uniqid());

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertEquals(3, $response->retries->retries);
    }

    function testRetriesSuceedWithBody()
    {
        CommonHelpers::recordTest('retries-succeeds-with-body');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->retries->retriesPost(requestId: uniqid(), requestBody: new Operations\RetriesPostRequestBody(fieldOne: 'one'));

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertEquals(3, $response->retries->retries);
    }

    function testGlobalRetryConfigDisable()
    {
        CommonHelpers::recordTest('retries-global-config-disable');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->setRetryConfig(new Retry\RetryConfigNone())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(Errors\APIException::class);
        $this->expectExceptionMessage('API error occurred');

        $response = $sdk->retries->retriesGet(requestId: uniqid(), numRetries: 2);
    }

    function testGlobalRetryConfigSuccess()
    {
        CommonHelpers::recordTest('retries-global-config-success');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->setRetryConfig(new Retry\RetryConfigBackoff(1, 50, 1.1, 1000, false))->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->retries->retriesGet(requestId: uniqid(), numRetries: 10);

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertEquals(10, $response->retries->retries);
    }

    function testGlobalRetryConfigTimeout()
    {
        CommonHelpers::recordTest('retries-global-config-timeout');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->setRetryConfig(new Retry\RetryConfigBackoff(1, 50, 1.1, 100, false))->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(Errors\APIException::class);
        $this->expectExceptionMessage('API error occurred');

        $response = $sdk->retries->retriesGet(requestId: uniqid(), numRetries: 100);
    }

    function testRetriesHeader()
    {
        CommonHelpers::recordTest('retries-header');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $options = Utils\Options::builder()->setRetryCodes(["4xx", "5xx"])->setRetryConfig(new Retry\RetryConfigBackoff(5000, 10000, 1.1, 10000, false))->build();

        $response = $sdk->retries->retriesAfter(requestId: uniqid(), numRetries: 3, options: $options);

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertEquals(3, $response->retries->retries);
    }

    function testRetriesAttemptCountBackoff()
    {
        CommonHelpers::recordTest('retries-attempt-count-backoff');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $response = $sdk->retries->retriesAttemptCount(
            requestId: uniqid(),
            numRetries: 3,
            retryAfterMsVal: 1
        );

        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());
        $this->assertEquals(3, $response->retries->retries);
    }

    function testRetriesAttemptCountBackoffZero()
    {
        CommonHelpers::recordTest('retries-attempt-count-zero');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(Errors\APIException::class);
        $this->expectExceptionMessage('API error occurred');

        $response = $sdk->retries->retriesAttemptCountZero(
            requestId: uniqid(),
            numRetries: 2,
            retryAfterMsVal: 1
        );
    }

    function testRetriesTimeout()
    {
        CommonHelpers::recordTest('retries-timeout');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $options = Utils\Options::builder()->setRetryCodes(["4xx", "5xx"])->setRetryConfig(new Retry\RetryConfigBackoff(1, 50, 1.1, 100, false))->build();
        $this->expectException(Errors\APIException::class);
        $this->expectExceptionMessage('API error occurred');

        $response = $sdk->retries->retriesGet(requestId: uniqid(), numRetries: 1000000000, options: $options);
    }

    function testRetriesConnectError()
    {
        CommonHelpers::recordTest('retries-connect-error');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->setRetryConfig(new Retry\RetryConfigBackoff(1, 50, 1.1, 1000, false))->build();

        $this->assertInstanceOf(\OpenAPI\OpenAPI\SDK::class, $sdk);

        $this->expectException(\GuzzleHttp\Exception\ConnectException::class);
        $this->expectExceptionMessage('Failed to connect');

        $response = $sdk->retries->retriesConnectErrorGet();
    }

    function testPaginationWithRetries()
    {
        CommonHelpers::recordTest('pagination-with-retries');

        $container = [];
        $history = Middleware::history($container);

        $handlerStack = HandlerStack::create();
        $handlerStack->push($history);

        $client = new \GuzzleHttp\Client([
            'handler' => $handlerStack
        ]);

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(Helpers::getHttpBinUrl())->setClient($client)->build();

        $result = $sdk->pagination->PaginationWithRetries();
        $count = 0;
        foreach ($result as $page) {
            $count += $page->res->resultArray ? count($page->res->resultArray) : 0;
        }

        $this->assertEquals(20, $count);
        $i = 0;
        foreach ($container as $transaction) {
            $this->assertEquals("/pagination/cursor_non_numeric", $transaction['request']->getUri()->getPath());
            $this->assertEquals("GET", $transaction['request']->getMethod());
            if ($i > 3) {
                $this->assertEquals(503, $transaction['response']->getStatusCode());    
            } else {
                $this->assertEquals(200, $transaction['response']->getStatusCode());
            }
        }
    }

}
