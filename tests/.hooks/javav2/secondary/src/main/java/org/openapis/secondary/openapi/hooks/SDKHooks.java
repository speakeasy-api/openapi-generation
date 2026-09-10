package org.openapis.secondary.openapi.hooks;

import org.openapis.secondary.openapi.utils.AsyncHooks;
import org.openapis.secondary.openapi.utils.Hooks;

public final class SDKHooks {

    private SDKHooks() {
    }

    public static final void initialize(Hooks hooks) {
        CustomSecurityHooks securityHooks = new CustomSecurityHooks();
        hooks.registerBeforeRequest(securityHooks);
    }

    public static final void initialize(AsyncHooks asyncHooks) {
        // no async hooks registered for this variant
    }

}
