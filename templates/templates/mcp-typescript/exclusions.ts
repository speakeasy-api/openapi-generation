function setupExclusions() {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride("^src/hooks/.*");
  }
  addUntrackedPattern("^src/hooks/registration\\.ts");

  addHeaderPattern("^src/hooks/registration\\.ts", true);
  addHeaderPattern(".*\\.ts$");
}
