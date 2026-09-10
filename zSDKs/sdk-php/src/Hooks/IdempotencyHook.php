<?php

declare(strict_types=1);

namespace OpenAPI\OpenAPI\Hooks;

use Psr\Http\Message\RequestInterface;

class IdempotencyHook implements BeforeRequestHook
{
    public function beforeRequest(BeforeRequestContext $context, RequestInterface $request): RequestInterface
    {
        $request->withHeader('Idempotency-Key', uniqid());

        return $request;
    }
}
