<?php

declare(strict_types=1);

namespace alphabetically\early\Hooks;

use GuzzleHttp\Exception\GuzzleException;
use GuzzleHttp\Promise\PromiseInterface;
use Psr\Http\Message\RequestInterface;
use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\UriInterface;


class CustomTestClient implements \GuzzleHttp\ClientInterface
{
    public \GuzzleHttp\ClientInterface $client;

    public function __construct(?\GuzzleHttp\ClientInterface $client = null)
    {
        if ($client === null) {
            $client = new \GuzzleHttp\Client();
        }
        $this->client = $client;
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
        // Inject a custom header into the outgoing request to prove the custom client is being used
        $modifiedRequest = $request->withHeader('X-Custom-Client-Active', 'true');

        // Send the modified request
        return $this->client->send($modifiedRequest, $options);
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
        // Inject a custom header into the outgoing request to prove the custom client is being used
        $modifiedRequest = $request->withHeader('X-Custom-Client-Active', 'true');

        return $this->client->sendAsync($modifiedRequest, $options);
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
        // Add custom header to request options
        if (! isset($options['headers'])) {
            $options['headers'] = [];
        }
        $options['headers']['X-Custom-Client-Active'] = 'true';

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
        // Add custom header to request options
        if (! isset($options['headers'])) {
            $options['headers'] = [];
        }
        $options['headers']['X-Custom-Client-Active'] = 'true';

        return $this->client->requestAsync($method, $uri, $options);
    }
}


class CustomClientHook implements SDKInitHook
{
    public function sdkInit(string $baseUrl, \GuzzleHttp\ClientInterface $client): SDKRequestContext
    {
        $customClient = new CustomTestClient($client);

        return new SDKRequestContext($baseUrl, $customClient);
    }
}
