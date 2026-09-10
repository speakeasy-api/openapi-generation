namespace Openapi.Hooks
{
    public static class HookRegistration
    {
        public static void InitHooks(IHooks hooks)
        {
            var testHook = new TestHook();
            hooks.RegisterSDKInitHook(testHook);
            hooks.RegisterBeforeRequestHook(testHook);
            hooks.RegisterAfterSuccessHook(testHook);
            hooks.RegisterAfterErrorHook(testHook);

            var customSecurityHook = new CustomSecurityHook();
            hooks.RegisterBeforeRequestHook(customSecurityHook);

            var testCancellationHook = new TestCancellationHook();
            hooks.RegisterBeforeRequestHook(testCancellationHook);
        }
    }
}
