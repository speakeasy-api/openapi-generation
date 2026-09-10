namespace Speakeasy.OpenAPI
{
    /// <summary>
    /// This file was modified and should not get overridden.
    /// </summary>
    public static class HookRegistration
    {
        public static void InitHooks(IHooks hooks)
        {
            hooks.RegisterBeforeRequestHook(new IdempotencyHook());
        }
    }
}
