import { CustomSecurityHook } from "./customsecurityhook.js";
import { ForceRetryHook } from "./forceretryhook.js";
import { Hooks } from "./types.js";
import { TestHook } from "./testhook.js";

export function initHooks(hooks: Hooks) {
  const testHook = new TestHook();
  const customSecurityHook = new CustomSecurityHook();
  const forceRetryHook = new ForceRetryHook();

  hooks.registerSDKInitHook(testHook);

  hooks.registerBeforeCreateRequestHook(testHook);
  hooks.registerBeforeCreateRequestHook(forceRetryHook);

  hooks.registerBeforeRequestHook(testHook);
  hooks.registerBeforeRequestHook(customSecurityHook);

  hooks.registerAfterSuccessHook(testHook);
  hooks.registerAfterSuccessHook(forceRetryHook);

  hooks.registerAfterErrorHook(testHook);
  hooks.registerAfterErrorHook(forceRetryHook);
}
