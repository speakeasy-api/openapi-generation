import { Hooks } from "./types.js";
import { TestHook } from "./testhook.js";

export function initHooks(hooks: Hooks) {
  const testHook = new TestHook();

  hooks.registerSDKInitHook(testHook);
  hooks.registerBeforeCreateRequestHook(testHook);
  hooks.registerBeforeRequestHook(testHook);
  hooks.registerAfterSuccessHook(testHook);
  hooks.registerAfterErrorHook(testHook);
}
