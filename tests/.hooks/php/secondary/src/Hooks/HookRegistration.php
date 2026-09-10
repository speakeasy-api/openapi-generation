<?php

declare(strict_types=1);

namespace alphabetically\early\Hooks;

class HookRegistration
{
    /**
     * @param  Hooks  $hooks
     */
    public static function initHooks(Hooks $hooks): void
    {
        $customClientHook = new CustomClientHook();
        $hooks->registerSDKInitHook($customClientHook);

        $customSecurityHook = new CustomSecurityHook();
        $hooks->registerBeforeRequestHook($customSecurityHook);
    }
}
