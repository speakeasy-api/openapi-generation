namespace Speakeasy.CustomHttp.Hooks
{
    public static class HookRegistration
    {
        public static void InitHooks(IHooks hooks)
        {
            var customSecurityHook = new CustomSecurityHook();
            hooks.RegisterBeforeRequestHook(customSecurityHook);
        }
    }
}
