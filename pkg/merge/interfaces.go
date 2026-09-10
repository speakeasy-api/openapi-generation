package merge

// Git abstracts the object database and network operations.
// The CLI implements this interface and injects it into the Generator.
type Git interface {
	// --- Read Operations ---

	// HasObject checks if a blob or commit exists in the local object DB.
	HasObject(hash string) bool

	// ReadBlob returns the content of a specific blob hash.
	ReadBlob(hash string) ([]byte, error)

	// FetchSnapshot ensures the history for a specific UUID exists locally.
	// This triggers the network call to origin.
	FetchSnapshot(uuid string) error

	// RepoRoot returns the root directory of the git repository.
	// This is the directory containing .git (the worktree root).
	// Returns empty string if not in a git repository.
	RepoRoot() string

	// --- Write Operations ---

	// WriteObject hashes content into the DB (blob) and returns the SHA1.
	WriteObject(content []byte) (string, error)

	// CreateSnapshotTree builds a Git Tree object from a map of "path" -> "blobHash".
	// Returns the tree hash.
	CreateSnapshotTree(fileHashes map[string]string) (string, error)

	// CommitSnapshot creates a commit object.
	// parentHash: The Previous Commit Hash (links the history).
	// Returns the commit hash.
	CommitSnapshot(treeHash, parentHash, message string) (string, error)

	// PushSnapshot syncs the ref to the server synchronously.
	// The push must complete before returning to ensure the commit is available
	// for future generations that may reference it as a parent.
	PushSnapshot(commitHash, uuid string) error

	// --- Conflict Management ---

	// SetConflictState sets up git's index to show a file as conflicted.
	// This writes the base, ours, and theirs versions as blobs and creates
	// stage 1, 2, 3 index entries, enabling standard git conflict resolution:
	//   - git status shows "both modified"
	//   - git mergetool can resolve conflicts
	//   - git checkout --ours/--theirs works
	//   - git add marks as resolved
	//
	// Parameters:
	//   - path: relative file path
	//   - base: content from common ancestor (stage 1)
	//   - ours: content from current/HEAD version (stage 2)
	//   - theirs: content from incoming/generated version (stage 3)
	//   - isExecutable: whether the file should be marked executable
	//
	// If base is nil, it indicates a new file conflict (no common ancestor).
	SetConflictState(path string, base, ours, theirs []byte, isExecutable bool) error
}
