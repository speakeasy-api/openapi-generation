// @ts-ignore
function setupExclusions(sdkClassName: string) {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride(".*/Hooks/[^/]+\\.php");
  }

  // WARNING: must match the registration filepath set in hooks/hooks.ts
  addUntrackedPattern(".*/Hooks/HookRegistration\\.php");

  addHeaderPattern(".*/Hooks/HookRegistration\\.php", true);
  addHeaderPattern(".*\\.php$");
  addHeaderPattern(".*\\.usage\\.php$", true);
}
