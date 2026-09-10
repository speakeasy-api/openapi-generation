declare(strict_types=1);

namespace OpenAPI\OpenAPI\Tests;

use GuzzleHttp\Exception\ConnectException;
use OpenAPI\OpenAPI\Models\Errors\APIException;
use OpenAPI\OpenAPI\Models\Errors\ErrorThrowable;
use OpenAPI\OpenAPI\Models\Errors\ErrorType1Throwable;
use OpenAPI\OpenAPI\Models\Errors\ErrorType2Throwable;
use OpenAPI\OpenAPI\Models\Errors\StatusGetXSpeakeasyErrorsResponseBodyThrowable;
use OpenAPI\OpenAPI\Models\Errors\TaggedError1Throwable;
use OpenAPI\OpenAPI\Models\Errors\TaggedError2Throwable;
use OpenAPI\OpenAPI\Models\Operations;

use OpenAPI\OpenAPI\Tests\CommonHelpers;
use PHPUnit\Framework\TestCase;

final class ErrorsTest extends TestCase
{
    public function testStatusGetError_DefaultErrorCodes(): void
    {
        CommonHelpers::recordTest('errors-status-get-error-default-error-codes');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        // When `clientServerStatusCodesAsErrors: true` is set (default), 4XX and 5XX ranges
        // are automatically treated as errors.

        // 400 and 500 responses are explicitly defined
        try {
            $sdk->errors->statusGetError(400);
            $this->fail('Expected APIException for 400');
        } catch (APIException $ex400) {
            $this->assertEquals(400, $ex400->getCode());
        }

        try {
            $sdk->errors->statusGetError(500);
            $this->fail('Expected APIException for 500');
        } catch (APIException $ex500) {
            $this->assertEquals(500, $ex500->getCode());
        }

        // 404 and 503 responses are undefined but still treated as errors by default
        try {
            $sdk->errors->statusGetError(404);
            $this->fail('Expected APIException for 404');
        } catch (APIException $ex404) {
            $this->assertEquals(404, $ex404->getCode());
        }

        try {
            $sdk->errors->statusGetError(503);
            $this->fail('Expected APIException for 503');
        } catch (APIException $ex503) {
            $this->assertEquals(503, $ex503->getCode());
        }
    }

    public function testStatusGetError_300_NonError(): void
    {
        CommonHelpers::recordTest('errors-status-get-error300-non-error');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        $response = $sdk->errors->statusGetError(300);
        $this->assertNotNull($response);
        $this->assertEquals(300, $response->httpMeta->response->getStatusCode());
    }

    public function testStatusgetErrorXSpeakeasyErrors(): void
    {
        CommonHelpers::recordTest('errors-status-get-error-x-speakeasy-errors');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        // 400 response is explicitly defined and is marked as an error in `x-speakeasy-errors`
        $this->expectException(APIException::class);
        $this->expectExceptionCode(400);
        $this->expectExceptionMessage('API error occurred');
        $response = $sdk->errors->statusGetXSpeakeasyErrors(400);
        $this->assertNull($response);
    }

    public function testStatusgetErrorXSpeakeasyErrors401(): void
    {
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        // 401 response is undefined but it is marked as an error in `x-speakeasy-errors`
        $this->expectException(APIException::class);
        $this->expectExceptionCode(401);
        $this->expectExceptionMessage('API error occurred');
        $response = $sdk->errors->statusGetXSpeakeasyErrors(401);
        $this->assertNull($response);
    }

    public function testStatusgetErrorXSpeakeasyErrors402(): void
    {
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        // 402 response is undefined and is not treated as an API error since it's not listed in `x-speakeasy-errors`.
        // Instead we raise a "Unknown status code received" exception.
        $this->expectException(APIException::class);
        $this->expectExceptionCode(402);
        $this->expectExceptionMessage('Unknown status code received');
        $response = $sdk->errors->statusGetXSpeakeasyErrors(402);
        $this->assertNull($response);
    }

    public function testStatusgetErrorXSpeakeasyErrors500(): void
    {
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        // Both 500 and 501 responses are marked as errors since `5XX` is listed `x-speakeasy-errors`.
        $this->expectException(ErrorThrowable::class);
        $this->expectExceptionCode(500);
        $this->expectExceptionMessage('an error occurred');
        $response = $sdk->errors->statusGetXSpeakeasyErrors(500);
        $this->assertNull($response);
    }

    public function testStatusgetErrorXSpeakeasyErrors501(): void
    {
        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        $this->expectException(StatusGetXSpeakeasyErrorsResponseBodyThrowable::class);
        $this->expectExceptionCode(501);
        $this->expectExceptionMessage('an error occurred');
        $response = $sdk->errors->statusGetXSpeakeasyErrors(501);
        $this->assertNull($response);
    }

    public function testStatusGetSuccessXSpeakeasyErrorsEmptyStatusCodeList(): void
    {
        CommonHelpers::recordTest('errors-status-get-error-x-speakeasy-errors-none');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()
            ->build();

        // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
        // with a dummy list, meaning all responses are treated as non-errors.

        $response = $sdk->errors->statusGetNonError(200);
        $this->assertNotNull($response);
        $this->assertEquals(200, $response->httpMeta->response->getStatusCode());

        $response = $sdk->errors->statusGetNonError(400);
        $this->assertNotNull($response);
        $this->assertEquals(400, $response->httpMeta->response->getStatusCode());

        $response = $sdk->errors->statusGetNonError(500);
        $this->assertNotNull($response);
        $this->assertEquals(500, $response->httpMeta->response->getStatusCode());
    }

    public function testStatusGetSuccessXSpeakeasyErrorsUnspecifiedResponses(): void
    {
        CommonHelpers::recordTest('errors-status-get-error-x-speakeasy-errors-default');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()
            ->build();

        // `x-speakeasy-errors` overrides the `clientServerStatusCodesAsErrors` setting
        // by marking all *unspecified* responses as errors.

        // 200 and 400 responses are explicitly defined, so they are treated as non-errors.
        $res200 = $sdk->errors->statusGetDefaultError(200);
        $this->assertNotNull($res200);
        $this->assertEquals(200, $res200->httpMeta->response->getStatusCode());

        $res400 = $sdk->errors->statusGetDefaultError(400);
        $this->assertNotNull($res400);
        $this->assertEquals(400, $res400->httpMeta->response->getStatusCode());

        // 404 and 500 responses are undefined, so they are treated as errors.
        try {
            $sdk->errors->statusGetDefaultError(404);
            $this->fail('Expected APIException for 404');
        } catch (APIException $ex404) {
            $this->assertEquals(404, $ex404->getCode());
        }

        try {
            $sdk->errors->statusGetDefaultError(500);
            $this->fail('Expected APIException for 500');
        } catch (APIException $ex500) {
            $this->assertEquals(500, $ex500->getCode());
        }

        // To make sure the catch-all "default" code gets properly applied in `templateErrorStatusCodesCheck`,
        // an AfterError hook was added to Hooks/TestHook.php to recover from 418 error.
        $res418 = $sdk->errors->statusGetDefaultError(418);
        $this->assertNotNull($res418);
        $this->assertEquals(200, $res418->httpMeta->response->getStatusCode());
    }

    public function testConnectionErrorGet(): void
    {
        CommonHelpers::recordTest('errors-connection-error');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        $this->expectException(ConnectException::class);

        try {
            $response = $sdk->errors->connectionErrorGet();
            $this->assertNull($response);
        } catch (ConnectException $exception) {
            $this->assertStringContainsString('cURL error 6: Could not resolve host: somebrokenapi.broken', $exception->getMessage());
            $this->assertMatchesRegularExpression('#https://curl\.(se|haxx\.se)/libcurl/c/libcurl-errors\.html#', $exception->getMessage());
            $this->assertStringContainsString('for http://somebrokenapi.broken/anything/connectionError', $exception->getMessage());
            throw $exception;
        }
    }

    public function testUnionOfErrors(): void
    {
        CommonHelpers::recordTest('errors-union-of-errors');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        $request1 = new Operations\ErrorType1RequestBody(
          error: 'Error1'
        );

        $this->expectException(ErrorType1Throwable::class);
        $this->expectExceptionCode(-1);
        $this->expectExceptionMessage('{"error":"Error1"}');
        $response1 = $sdk->errors->errorUnionPost($request1);
        $this->assertNull($response1);

        $request2 = new Operations\ErrorType2RequestBody(
          errorType2Message: new Operations\ErrorType2Message(
            message: 'Error2'
          )
        );

        $this->expectException(ErrorType2Throwable::class);
        $this->expectExceptionCode(-1);
        $this->expectExceptionMessage('{"error":{"message":"Error2"}}');
        $response2 = $sdk->errors->errorUnionPost($request2);
        $this->assertNull($response2);
    }

    public function testUnionOfErrorsDiscriminated(): void
    {
        CommonHelpers::recordTest('errors-union-of-errors-discriminated');

        $sdk = \OpenAPI\OpenAPI\SDK::builder()->setServerUrl(CommonHelpers::getHttpBinUrl())
            ->build();

        $request1 = new Operations\TaggedError1RequestBody(
          error: 'Error1',
          tag: 'tag1'
        );

        $this->expectException(TaggedError1Throwable::class);
        $this->expectExceptionCode(-1);
        $this->expectExceptionMessage('{"error":"Error1","tag":"tag1"}');
        $response1 = $sdk->errors->errorUnionDiscriminatedPost($request1);
        $this->assertNull($response1);

        $request2 = new Operations\TaggedError2RequestBody(
          tag: 'tag2',
          taggedError2Message: new Operations\TaggedError2Message(
            message: 'Error2'
          )
        );

        $this->expectException(TaggedError2Throwable::class);
        $this->expectExceptionCode(-1);
        $this->expectExceptionMessage('{"error":{"message":"Error2"},"tag":"tag2"}');
        $response2 = $sdk->errors->errorUnionDiscriminatedPost($request2);
        $this->assertNull($response2);
    }
}
