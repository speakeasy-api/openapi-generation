<?php

declare(strict_types=1);

namespace OpenAPI\OpenAPI\Hooks;

class HookRegistration
{
    /**
     * @param  Hooks  $hooks
     */
    public static function initHooks(Hooks $hooks): void
    {
        $customSecurityHook = new CustomSecurityHook();
        $hooks->registerBeforeRequestHook($customSecurityHook);
    }
}
