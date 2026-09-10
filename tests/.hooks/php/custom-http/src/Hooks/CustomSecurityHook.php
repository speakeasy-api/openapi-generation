<?php

declare(strict_types=1);

namespace OpenAPI\OpenAPI\Hooks;

use OpenAPI\OpenAPI\Models\Components\Security;
use Psr\Http\Message\RequestInterface;

class CustomSecurityHook implements BeforeRequestHook
{
    public function beforeRequest(BeforeRequestContext $context, RequestInterface $request): RequestInterface
    {
        switch ($context->operationID) {
            case 'customHttpOnly':
                if ($context->securitySource === null) {
                    throw new \InvalidArgumentException('Security source is null');
                }

                $security = $context->securitySource->call($context);

                if (! $security instanceof Security) {
                    throw new \InvalidArgumentException('Security source is not of type Security');
                }

                if ($security->customHttp === null) {
                    throw new \RuntimeException('CustomHttp security is not defined');
                }

                $customHttp = $security->customHttp;

                $request = $request->withAddedHeader('X-Security-UserID', (string) $customHttp->userID);
                $request = $request->withAddedHeader('X-Security-Role', $customHttp->role->value);
                $request = $request->withAddedHeader('X-Security-Passphrase', $customHttp->passphrase);

                if ($customHttp->accessCode !== null) {
                    $request = $request->withAddedHeader('X-Security-AccessCode', (string) $customHttp->accessCode);
                }

                if ($customHttp->scopes !== null && count($customHttp->scopes) > 0) {
                    $scopesJson = json_encode($customHttp->scopes);
                    if ($scopesJson === false) {
                        throw new \RuntimeException('Failed to encode scopes to JSON');
                    }
                    $request = $request->withAddedHeader('X-Security-Scopes', $scopesJson);
                }

                break;
        }

        return $request;
    }
}
