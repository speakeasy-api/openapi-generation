// @ts-ignore
function setupExclusions() {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride("^internal/hooks/.*");
    addIgnoreOverride("^sdkinternal/hooks/.*");
  }

  addUntrackedPattern("^internal/hooks/registration\\.go");
  addUntrackedPattern("^sdkinternal/hooks/registration\\.go");
  addHeaderPattern("^internal/hooks/registration\\.go", true);
  addHeaderPattern("^sdkinternal/hooks/registration\\.go", true);
  addHeaderPattern(".*\\.go$");
  addHeaderPattern(".*\\.usage\\.go$", true);
}
