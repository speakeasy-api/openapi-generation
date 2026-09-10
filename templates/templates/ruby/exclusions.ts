// @ts-ignore
function setupExclusions(_sdkClassName: string) {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride("lib/.*/sdk_hooks/[^/]+\\.rb");
  }

  // WARNING: must match the registration filepath set in hooks/hooks.ts
  addUntrackedPattern("lib/.*/sdk_hooks/registration\\.rb");

  addHeaderPattern("lib/.*/sdk_hooks/registration\\.rb", true);
  addHeaderPattern(".*\\.rb$");
  addHeaderPattern(".*\\.usage\\.rb$", true);
}
