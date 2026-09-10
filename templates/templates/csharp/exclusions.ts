// @ts-ignore
function setupExclusions(sdkClassName: string) {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride(".*/Hooks/[^/]+\\.cs");
  }

  // WARNING: must match the registration filepath set in hooks/hooks.ts
  addUntrackedPattern(".*/Hooks/HookRegistration\\.cs");
  addHeaderPattern(".*/Hooks/HookRegistration\\.cs", true);
  addHeaderPattern(".*\\.cs$");
  addHeaderPattern(".*\\.usage\\.cs$", true);
}
