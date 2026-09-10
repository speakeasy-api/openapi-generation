// @ts-ignore
function setupExclusions() {
  if (!accountHasFeatureAccess("sdkHooks")) {
    addIgnoreOverride("^internal/sdk/sdkinternal/hooks/.*");
  }

  addUntrackedPattern("^internal/sdk/sdkinternal/hooks/registration\\.go");
  addHeaderPattern("^internal/sdk/sdkinternal/hooks/registration\\.go", true);

  // Hand-written command layer: user-owned, survives regeneration untouched.
  addUntrackedPattern("^internal/cli/custom/.*");
  addHeaderPattern("^internal/cli/custom/.*", true);
  addHeaderPattern(".*\\.go$");
  addHeaderPattern(".*\\.usage\\.go$", true);
}
