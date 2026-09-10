// @ts-ignore
function setupExclusions() {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride("^internal/sdk/internal/hooks/.*");
  }

  addUntrackedPattern("^internal/sdk/internal/hooks/registration\\.go");

  addHeaderPattern("^internal/sdk/internal/hooks/registration\\.go", true);
  addHeaderPattern(".*\\.go$");
  addHeaderPattern(".*\\.usage\\.go$", true);
}
