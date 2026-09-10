<?php

declare(strict_types=1);

namespace OpenAPI\OpenAPI\Hooks;

use GuzzleHttp\Exception\GuzzleException;
use GuzzleHttp\Promise\PromiseInterface;
use OpenAPI\OpenAPI\SDKConfiguration;
use OpenAPI\OpenAPI\Utils;
use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\UriInterface;


class TestClient implements \GuzzleHttp\ClientInterface
{
    public \GuzzleHttp\ClientInterface $client;

    public function __construct(?\GuzzleHttp\ClientInterface $client = null)
    {
        if ($client === null) {
            $client = new \GuzzleHttp\Client();
        } else {
            $this->client = $client;
        }
    }

    /**
     * Send an HTTP request.
     *
     * @param  RequestInterface  $request  Request to send
     * @param  array<string, mixed>  $options  Request options to apply to the given
     *                                         request and to the transfer.
     *
     * @throws GuzzleException
     */
    public function send(RequestInterface $request, array $options = []): ResponseInterface
    {
        $options['headers']['Client-Level-Header'] = 'added by client';

        return $this->client->send($request, $options);
    }


    /**
     * Asynchronously send an HTTP request.
     *
     * @param  RequestInterface  $request  Request to send
     * @param  array<string, mixed>  $options  Request options to apply to the given
     *                                         request and to the transfer.
     */
    public function sendAsync(RequestInterface $request, array $options = []): PromiseInterface
    {
        $options['headers']['Client-Level-Header'] = 'added by client';

        return $this->client->sendAsync($request, $options);
    }

    /**
     * Get a client configuration option.
     *
     * These options include default request options of the client, a "handler"
     * (if utilized by the concrete client), and a "base_uri" if utilized by
     * the concrete client.
     *
     * @param  string|null  $option  The config option to retrieve.
     *
     * @return mixed
     *
     * @deprecated ClientInterface::getConfig will be removed in guzzlehttp/guzzle:8.0.
     */
    public function getConfig(?string $option = null): mixed
    {
        return $this->client->getConfig($option);
    }

    /**
     * Create and send an HTTP request.
     *
     * Use an absolute path to override the base path of the client, or a
     * relative path to append to the base path of the client. The URL can
     * contain the query string as well.
     *
     * @param  string  $method  HTTP method.
     * @param  string|UriInterface  $uri  URI object or string.
     * @param  array<string, mixed>  $options  Request options to apply.
     *
     * @throws GuzzleException
     */
    public function request(string $method, $uri, array $options = []): ResponseInterface
    {
        $options['headers']['Client-Level-Header'] = 'added by client';

        return $this->client->request($method, $uri, $options);
    }

    /**
     * Create and send an asynchronous HTTP request.
     *
     * Use an absolute path to override the base path of the client, or a
     * relative path to append to the base path of the client. The URL can
     * contain the query string as well. Use an array to provide a URL
     * template and additional variables to use in the URL template expansion.
     *
     * @param  string  $method  HTTP method
     * @param  string|UriInterface  $uri  URI object or string.
     * @param  array<string, mixed>  $options  Request options to apply.
     */
    public function requestAsync(string $method, $uri, array $options = []): PromiseInterface
    {
        $options['headers']['Client-Level-Header'] = 'added by client';

        return $this->client->requestAsync($method, $uri, $options);
    }
}


class TestHook implements AfterErrorHook, AfterSuccessHook, BeforeRequestHook, SDKInitHook
{
    private string $initSdkVersion = '';

    public function sdkInit(SDKConfiguration $config): SDKConfiguration
    {
        $this->initSdkVersion = $config->sdkVersion;
        $config->client = new TestClient($config->client);

        return $config;
    }

    public function beforeRequest(BeforeRequestContext $context, RequestInterface $request): RequestInterface
    {
        $request = $request->withHeader('Idempotency-Key', 'some-key');

        switch ($context->operationID) {
            case 'testHooks':
                $requestParams = Utils\Utils::proper_parse_str($request->getUri()->getQuery());
                $requestParams['someParam'] = 'overriddenParam';
                $uri = $request->getUri()->withQuery(http_build_query($requestParams));
                $request = $request->withUri($uri);
                break;
            case 'authorizationHeaderModification':
                $token = $context->securitySource->call($context)->apiKeyAuth;
                $request = $request->withoutHeader('Authorization')->withAddedHeader('Authorization', "$token modified");
                break;
            case 'testHooksBeforeCreateRequestPaths':
                $request = $request->withAddedHeader('old-pathname', $request->getUri()->getPath());
                break;
            case 'hooksCustomUserAgent':
                $request = $request
                    ->withHeader('user-agent', 'acme-corp/'.$context->config->sdkVersion.' acme-corp/'.PHP_VERSION)
                    ->withHeader('X-Test-Gen-Version', $context->config->genVersion)
                    ->withHeader('X-Test-Doc-Version', $context->config->openapiDocVersion)
                    ->withHeader('X-Test-Init-Sdk-Version', $this->initSdkVersion);
                break;
        }

        return $request;
    }

    public function afterSuccess(AfterSuccessContext $context, ResponseInterface $response): ResponseInterface
    {
        if ($context->operationID === 'testHooksAfterResponse') {
            throw new \Exception('validation failed');
        }

        return $response;
    }

    public function afterError(AfterErrorContext $context, ?ResponseInterface $response, ?\Throwable $exception): ErrorResponseContext
    {
        if ($context->operationID == 'testHooksError') {
            if ($response != null && $response->getStatusCode() !== 400) {
                return new ErrorResponseContext(null, new \Exception('expected status code 400'));
            }

            return new ErrorResponseContext(null, new \Exception('special test error case'));
        }

        if ($context->operationID == 'statusGetDefaultError') {
            if ($response != null && $response->getStatusCode() === 418) {
                $okResponse = new \GuzzleHttp\Psr7\Response(200);

                return new ErrorResponseContext($okResponse, null);
            }
        }

        return new ErrorResponseContext($response, $exception);
    }
}
