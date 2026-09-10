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
        $testHook = new TestHook();
        $hooks->registerSDKInitHook($testHook);
        $hooks->registerBeforeRequestHook($testHook);
        $hooks->registerAfterSuccessHook($testHook);
        $hooks->registerAfterErrorHook($testHook);

        $customSecurityHook = new CustomSecurityHook();
        $hooks->registerBeforeRequestHook($customSecurityHook);
    }
}
