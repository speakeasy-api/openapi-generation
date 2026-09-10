// @ts-ignore
function setupExclusions() {
  // TODO might be more specific but don't know how to access templatePackageName() here!
  const hooksPackagePathRegex = `^src/main/java/.*/hooks`;
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride(hooksPackagePathRegex + "/.*");
  }
  addUntrackedPattern(hooksPackagePathRegex + "/SDKHooks\\.java");
  addUntrackedPattern("^build-extras\\.gradle");

  addHeaderPattern(hooksPackagePathRegex + "/SDKHooks\\.java", true);
  addHeaderPattern(".*\\.java$");
  addHeaderPattern(".*\\.usage\\.java$", true);
}
