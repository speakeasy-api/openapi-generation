import { CustomSecurityHook } from "./customsecurityhook.js";
import { Hooks } from "./types.js";

export function initHooks(hooks: Hooks) {
  const customSecurityHook = new CustomSecurityHook();
  hooks.registerBeforeRequestHook(customSecurityHook);
}
