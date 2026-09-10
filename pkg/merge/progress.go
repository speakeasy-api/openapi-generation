package merge

import "fmt"

// ConflictsError is returned when merge conflicts are detected.
// The Files slice contains relative paths of all conflicting files.
// The CLI should render this as a git-status style message and exit non-zero.
type ConflictsError struct {
	Files []string
}

func (e *ConflictsError) Error() string {
	return fmt.Sprintf("merge conflicts detected in %d file(s)", len(e.Files))
}

// SnapshotRefError is returned when persistentEdits is enabled and the
// merge-base blob cannot be recovered. The three-way merge cannot proceed
// without the base, so generation is aborted rather than writing conflict
// markers or silently losing customer edits.
type SnapshotRefError struct {
	GenerationID string
	BlobHash     string
	Cause        error
}

func (e *SnapshotRefError) Error() string {
	return fmt.Sprintf("could not fetch snapshot ref for generation %s (blob %s): %v", e.GenerationID, e.BlobHash, e.Cause)
}

func (e *SnapshotRefError) Unwrap() error {
	return e.Cause
}
