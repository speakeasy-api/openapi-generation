package org.openapis.openapi.hooks;

import org.openapis.openapi.utils.AsyncHooks;
import org.openapis.openapi.utils.Hooks;

public final class SDKHooks {

    private SDKHooks() {
        // prevent instantiation
    }

    public static void initialize(Hooks hooks) {
        TestHooks testHooks = new TestHooks();
        hooks.registerBeforeRequest(testHooks);
        hooks.registerAfterSuccess(testHooks);
        hooks.registerAfterError(testHooks);
        hooks.registerSdkInit(testHooks);

        CustomSecurityHooks securityHooks = new CustomSecurityHooks();
        hooks.registerBeforeRequest(securityHooks);
    }

    public static void initialize(AsyncHooks asyncHooks) {
        AsyncTestHooks testHooks = new AsyncTestHooks();
        asyncHooks.registerBeforeRequest(testHooks);
        asyncHooks.registerAfterSuccess(testHooks);
        asyncHooks.registerAfterError(testHooks);
    }

}
