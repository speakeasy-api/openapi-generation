// @ts-ignore
function setupExclusions(sdkClassName: string) {
  if (!accountHasFeatureAccess("sdkHooks")) {
    // Don't allow monkey patching hooks in the free/scale-up tier
    addIgnoreOverride(`^src/.*/_hooks/.*`);
  }

  addUntrackedPattern(`^src/.*/_hooks/registration\\.py`);
  addUntrackedPattern(`^src/.*/_hooks/asyncregistration\\.py`);
  addUntrackedPattern(`.*\\.pyc$`);
  addUntrackedPattern(`.*mypy_cache/.*$`);
  addHeaderPattern("^src/.*/_hooks/registration\\.py", true);
  addHeaderPattern("^src/.*/_hooks/asyncregistration\\.py", true);
  addHeaderPattern(".*\\.py$");
  addHeaderPattern(".*\\.usage\\.py$", true);
}
