import { CustomHttpHook } from "./customhttphook.js";
import { Hooks } from "./types.js";

export function initHooks(hooks: Hooks) {
  const customHttpHook = new CustomHttpHook();
  hooks.registerBeforeRequestHook(customHttpHook);
}
