<?php

declare(strict_types=1);

namespace OpenAPI\OpenAPI\Hooks;

use OpenAPI\OpenAPI\Models\Shared\Security;
use Psr\Http\Message\RequestInterface;

class CustomSecurityHook implements BeforeRequestHook
{
    public function beforeRequest(BeforeRequestContext $context, RequestInterface $request): RequestInterface
    {
        $request = $request->withHeader('Idempotency-Key', 'some-key');

        switch ($context->operationID) {
            case 'customSchemeAppId':
                if ($context->securitySource === null) {
                    throw new \InvalidArgumentException('Security source is null');
                }

                $security = $context->securitySource->call($context);

                if (! $security instanceof Security) {
                    throw new \InvalidArgumentException('Security source is not of type Security');
                }

                if ($security->customSchemeAppId === null) {
                    throw new \RuntimeException('CustomSchemeAppID security is not defined');
                }

                $request = $request->withAddedHeader('X-Security-App-Id', $security->customSchemeAppId->appId);
                $request = $request->withAddedHeader('X-Security-Secret', $security->customSchemeAppId->secret);
                break;
        }

        return $request;
    }
}
