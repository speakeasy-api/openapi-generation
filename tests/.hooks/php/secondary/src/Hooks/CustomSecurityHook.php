<?php

declare(strict_types=1);

namespace alphabetically\early\Hooks;

use alphabetically\early\Models\Operations\CustomSchemeAppIdSecurity;
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

                $customSecurity = $context->securitySource->call($context);

                if (! $customSecurity instanceof CustomSchemeAppIdSecurity) {
                    throw new \InvalidArgumentException('Security source is not of type CustomSchemeAppIdSecurity');
                }

                $request = $request->withAddedHeader('X-Security-App-Id', $customSecurity->appId);
                $request = $request->withAddedHeader('X-Security-Secret', $customSecurity->secret);
                break;
        }

        return $request;
    }
}
