// @ts-ignore
function setupExclusions() {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride("^src/hooks/.*");
  }
  addUntrackedPattern("^src/hooks/registration\\.ts");
  addUntrackedPattern("^src/hooks/webhook-security-custom\\.ts");

  addHeaderPattern("^src/hooks/registration\\.ts", true);
  addHeaderPattern("^src/hooks/webhook-security-custom\\.ts", true);
  addHeaderPattern(".*\\.ts$");
  addHeaderPattern(".*\\.usage\\.ts$", true);
}
