package generate

// compilePlan determines which compilation stages to run.
//
// Returns (runPristine, runFinal) where:
//   - runPristine: compile to normalize files so checksums match post-format state
//   - runFinal: compile after patches are applied (or if pristine was skipped)
//
// Invariants when compileEnabled=true:
//   - At least one compile pass always runs
//   - If pristine ran and no patches exist, final is skipped (optimization)
//   - If pristine is skipped, final must run (fallback)
func compilePlan(compileEnabled, pristineMode, patchesEnabled bool) (runPristine, runFinal bool) {
	if !compileEnabled {
		return false, false
	}

	runPristine = pristineMode
	runFinal = patchesEnabled || !runPristine

	return runPristine, runFinal
}
