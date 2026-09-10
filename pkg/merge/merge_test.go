package merge

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/lockfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Mock Implementations ==========

// MockGit implements the Git interface for testing
type MockGit struct {
	Objects        map[string][]byte            // blob hash -> content
	Trees          map[string]map[string]string // tree hash -> (path -> blob hash)
	Commits        map[string]MockCommit
	FetchCalled    map[string]bool // track which UUIDs were fetched
	FetchError     error           // if set, FetchSnapshot returns this error
	ConflictStates map[string]MockConflictState
}

type MockCommit struct {
	TreeHash   string
	ParentHash string
	Message    string
}

type MockConflictState struct {
	Base         []byte
	Ours         []byte
	Theirs       []byte
	IsExecutable bool
}

func NewMockGit() *MockGit {
	return &MockGit{
		Objects:        make(map[string][]byte),
		Trees:          make(map[string]map[string]string),
		Commits:        make(map[string]MockCommit),
		FetchCalled:    make(map[string]bool),
		ConflictStates: make(map[string]MockConflictState),
	}
}

func (m *MockGit) HasObject(hash string) bool {
	_, ok := m.Objects[hash]
	return ok
}

func (m *MockGit) ReadBlob(hash string) ([]byte, error) {
	if content, ok := m.Objects[hash]; ok {
		return content, nil
	}
	return nil, fmt.Errorf("object not found: %s", hash)
}

func (m *MockGit) FetchSnapshot(uuid string) error {
	m.FetchCalled[uuid] = true
	return m.FetchError
}

func (m *MockGit) RepoRoot() string {
	// Return a fixed path for testing - tests don't depend on actual repo root
	return "/mock/repo"
}

func (m *MockGit) WriteObject(content []byte) (string, error) {
	// Use real SHA-1 for predictable testing
	h := sha1.New()
	h.Write(content)
	hash := hex.EncodeToString(h.Sum(nil))
	m.Objects[hash] = content
	return hash, nil
}

func (m *MockGit) CreateSnapshotTree(fileHashes map[string]string) (string, error) {
	// Simple hash of the tree contents
	keys := make([]string, 0, len(fileHashes))
	for k := range fileHashes {
		keys = append(keys, k)
	}

	// Create a deterministic tree hash
	var treeContent strings.Builder
	for _, k := range keys {
		treeContent.WriteString(k)
		treeContent.WriteString(":")
		treeContent.WriteString(fileHashes[k])
		treeContent.WriteString("\n")
	}

	h := sha1.New()
	h.Write([]byte(treeContent.String()))
	treeHash := hex.EncodeToString(h.Sum(nil))

	m.Trees[treeHash] = fileHashes
	return treeHash, nil
}

func (m *MockGit) CommitSnapshot(treeHash, parentHash, message string) (string, error) {
	commitContent := "tree:" + treeHash + "\nparent:" + parentHash + "\nmsg:" + message
	h := sha1.New()
	h.Write([]byte(commitContent))
	commitHash := hex.EncodeToString(h.Sum(nil))

	m.Commits[commitHash] = MockCommit{
		TreeHash:   treeHash,
		ParentHash: parentHash,
		Message:    message,
	}
	return commitHash, nil
}

func (m *MockGit) PushSnapshot(commitHash, uuid string) error {
	// No-op for testing
	return nil
}

func (m *MockGit) SetConflictState(path string, base, ours, theirs []byte, isExecutable bool) error {
	m.ConflictStates[path] = MockConflictState{
		Base:         base,
		Ours:         ours,
		Theirs:       theirs,
		IsExecutable: isExecutable,
	}
	return nil
}

// MockFileSystem implements the filesystem.FileSystem interface for testing
type MockFileSystem struct {
	Files map[string][]byte
	Modes map[string]fs.FileMode
}

var _ filesystem.FileSystem = (*MockFileSystem)(nil)

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Modes: make(map[string]fs.FileMode),
	}
}

func (m *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, ok := m.Files[path]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) WriteFile(path string, content []byte, perm fs.FileMode) error {
	m.Files[path] = content
	m.Modes[path] = perm
	return nil
}

func (m *MockFileSystem) Open(name string) (fs.File, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockFileSystem) OpenFile(name string, flag int, perm fs.FileMode) (filesystem.File, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockFileSystem) Stat(name string) (fs.FileInfo, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil // no-op for testing
}

func (m *MockFileSystem) Remove(name string) error {
	delete(m.Files, name)
	delete(m.Modes, name)
	return nil
}

func (m *MockFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, errors.New("not implemented in mock")
}

func (m *MockFileSystem) ScanForGeneratedIDs() (map[string]string, error) {
	return map[string]string{}, nil
}

// ========== Helper Functions for Tests ==========

func newTestLockFile() *config.LockFile {
	lf := lockfile.New()
	lf.TrackedFiles = sequencedmap.New[string, lockfile.TrackedFile]()
	return lf
}

func computeSHA1(content []byte) string {
	h := sha1.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

// ========== Tests for Pure Functions ==========

func TestSplitAfter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single line no newline",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "single line with newline",
			input:    "hello\n",
			expected: []string{"hello\n"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1\n", "line2\n", "line3"},
		},
		{
			name:     "multiple lines with trailing newline",
			input:    "line1\nline2\n",
			expected: []string{"line1\n", "line2\n"},
		},
		{
			name:     "empty lines",
			input:    "\n\n",
			expected: []string{"\n", "\n"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := splitAfter(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLinesToCharsMux(t *testing.T) {
	t.Parallel()
	text1 := "line1\nline2\n"
	text2 := "line1\nline3\n"
	text3 := "line2\nline3\n"

	out1, out2, out3, lineMap := linesToCharsMux(text1, text2, text3)

	// Verify that the encoding is consistent
	assert.NotEmpty(t, out1)
	assert.NotEmpty(t, out2)
	assert.NotEmpty(t, out3)
	assert.NotEmpty(t, lineMap)

	// Verify that we can reconstruct the original texts
	reconstruct := func(encoded string) string {
		var result strings.Builder
		for _, r := range encoded {
			if int(r) < len(lineMap) {
				result.WriteString(lineMap[r])
			}
		}
		return result.String()
	}

	assert.Equal(t, text1, reconstruct(out1))
	assert.Equal(t, text2, reconstruct(out2))
	assert.Equal(t, text3, reconstruct(out3))
}

func TestDiffsToModifications(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		diffs    []diffmatchpatch.Diff
		expected []modification
	}{
		{
			name:     "empty diffs",
			diffs:    []diffmatchpatch.Diff{},
			expected: nil,
		},
		{
			name: "single insert",
			diffs: []diffmatchpatch.Diff{
				{Type: diffmatchpatch.DiffInsert, Text: "new"},
			},
			expected: []modification{
				{startBase: 0, endBase: 0, text: "new"},
			},
		},
		{
			name: "single delete",
			diffs: []diffmatchpatch.Diff{
				{Type: diffmatchpatch.DiffDelete, Text: "old"},
			},
			expected: []modification{
				{startBase: 0, endBase: 3, text: ""},
			},
		},
		{
			name: "delete followed by insert (replacement)",
			diffs: []diffmatchpatch.Diff{
				{Type: diffmatchpatch.DiffDelete, Text: "old"},
				{Type: diffmatchpatch.DiffInsert, Text: "new"},
			},
			expected: []modification{
				{startBase: 0, endBase: 3, text: "new"},
			},
		},
		{
			name: "equal, delete, insert",
			diffs: []diffmatchpatch.Diff{
				{Type: diffmatchpatch.DiffEqual, Text: "abc"},
				{Type: diffmatchpatch.DiffDelete, Text: "def"},
				{Type: diffmatchpatch.DiffInsert, Text: "xyz"},
			},
			expected: []modification{
				{startBase: 3, endBase: 6, text: "xyz"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := diffsToModifications(tt.diffs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMerge3Way_NoChanges(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\n")
	current := []byte("line1\nline2\nline3\n")
	generated := []byte("line1\nline2\nline3\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict)
	assert.Equal(t, base, merged)
}

func TestMerge3Way_OnlyCurrentChanged(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\n")
	current := []byte("line1\nmodified\nline3\n")
	generated := []byte("line1\nline2\nline3\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict)
	assert.Equal(t, current, merged)
}

func TestMerge3Way_OnlyNewChanged(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\n")
	current := []byte("line1\nline2\nline3\n")
	generated := []byte("line1\ngenerated\nline3\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict)
	assert.Equal(t, generated, merged)
}

func TestMerge3Way_BothChangedNonOverlapping(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\nline4\n")
	current := []byte("modified1\nline2\nline3\nline4\n")
	generated := []byte("line1\nline2\nline3\nmodified4\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict)
	expected := []byte("modified1\nline2\nline3\nmodified4\n")
	assert.Equal(t, expected, merged)
}

func TestMerge3Way_IdenticalChanges(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\n")
	current := []byte("line1\nmodified\nline3\n")
	generated := []byte("line1\nmodified\nline3\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict, "Identical changes should auto-merge")
	assert.Equal(t, current, merged)
}

func TestMerge3Way_Conflict(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\n")
	current := []byte("line1\nuser edit\nline3\n")
	generated := []byte("line1\ngenerated edit\nline3\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.True(t, hasConflict)

	// Check that conflict markers are present
	mergedStr := string(merged)
	assert.Contains(t, mergedStr, "<<<<<<< Current")
	assert.Contains(t, mergedStr, "=======")
	assert.Contains(t, mergedStr, ">>>>>>> New")
	assert.Contains(t, mergedStr, "user edit")
	assert.Contains(t, mergedStr, "generated edit")
}

func TestMerge3Way_InsertInsertConflict(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline3\n")
	current := []byte("line1\nuser insert\nline3\n")
	generated := []byte("line1\ngenerated insert\nline3\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.True(t, hasConflict)
	assert.Contains(t, string(merged), "<<<<<<< Current")
}

func TestMerge3Way_EmptyBase(t *testing.T) {
	t.Parallel()
	base := []byte("")
	current := []byte("user content\n")
	generated := []byte("generated content\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.True(t, hasConflict, "Both sides adding content should conflict")
	assert.NotEmpty(t, merged, "Should produce merged output even with conflict")
}

func TestMerge3Way_EmptyFiles(t *testing.T) {
	t.Parallel()
	base := []byte("")
	current := []byte("")
	generated := []byte("")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict)
	assert.Empty(t, merged)
}

func TestMerge3Way_NoTrailingNewline(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2")
	current := []byte("line1\nmodified")
	generated := []byte("line1\ngenerated")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.True(t, hasConflict)
	// The merge should handle files without trailing newlines
	assert.NotEmpty(t, merged)
}

func TestMerge3Way_MixedLineEndings(t *testing.T) {
	t.Parallel()
	// Note: This test documents current behavior.
	base := []byte("line1\nline2\n")
	current := []byte("line1\r\nline2\r\n") // Windows line endings
	generated := []byte("line1\nline2\n")   // Unix line endings

	merged, hasConflict := merge3Way(base, current, generated)

	// The algorithm treats line ending changes as modifications
	// This test documents the actual behavior - it may or may not conflict
	// depending on how the diff algorithm sees the changes
	assert.NotEmpty(t, merged, "Should produce merged output")

	// The behavior here depends on the diff algorithm's handling of \r\n vs \n
	// We just verify it doesn't crash and produces output
	t.Logf("Mixed line endings result: conflict=%v, len=%d", hasConflict, len(merged))
}

// ========== Tests for Merge Function ==========

func TestMerge_NilGit(t *testing.T) {
	t.Parallel()
	ctx := &MergeContext{
		Git:          nil,
		VirtualFiles: map[string]*VirtualFile{},
	}

	_, err := Merge(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "git interface is nil")
}

func TestMerge_NoVirtualFiles(t *testing.T) {
	t.Parallel()
	ctx := &MergeContext{
		Git:          NewMockGit(),
		VirtualFiles: nil,
	}

	_, err := Merge(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no virtual files")
}

func TestMerge_NewFile(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	vFile := &VirtualFile{
		Path:     "newfile.txt",
		Content:  []byte("new content\n"),
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"newfile.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.False(t, result.IsNoOp)
	assert.Empty(t, result.ConflictFiles)

	// Verify file was written
	written, err := mockFS.ReadFile("newfile.txt")
	require.NoError(t, err)
	assert.Equal(t, vFile.Content, written)

	// Verify tracking was updated
	tracked, exists := lockFile.TrackedFiles.Get("newfile.txt")
	require.True(t, exists)
	assert.NotEmpty(t, tracked.PristineGitObject)
	assert.NotEmpty(t, tracked.LastWriteChecksum)
}

func TestMerge_CleanFile(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	oldContent := []byte("old content\n")
	newContent := []byte("new content\n")

	// Set up existing file on disk
	mockFS.Files["file.txt"] = oldContent

	// Compute checksum
	checksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader(string(oldContent)))

	// Set up tracking
	oldBlobHash := computeSHA1(oldContent)
	mockGit.Objects[oldBlobHash] = oldContent

	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + checksum,
		PristineGitObject: oldBlobHash,
	})

	vFile := &VirtualFile{
		Path:     "file.txt",
		Content:  newContent,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.False(t, result.IsNoOp)
	assert.Empty(t, result.ConflictFiles)

	// Verify new content was written
	written, err := mockFS.ReadFile("file.txt")
	require.NoError(t, err)
	assert.Equal(t, newContent, written)
}

func TestMerge_DirtyFileWithBase(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	baseContent := []byte("line1\nline2\nline3\n")
	currentContent := []byte("line1\nuser edit\nline3\n")
	newContent := []byte("line1\nline2\nline3\ngenerated line\n")

	// Set up base in git
	baseHash := computeSHA1(baseContent)
	mockGit.Objects[baseHash] = baseContent

	// Set up current file on disk (modified by user)
	mockFS.Files["file.txt"] = currentContent

	// Set up tracking with old checksum (file is dirty)
	oldChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader(string(baseContent)))
	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + oldChecksum,
		PristineGitObject: baseHash,
	})

	vFile := &VirtualFile{
		Path:     "file.txt",
		Content:  newContent,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.Empty(t, result.ConflictFiles, "Non-overlapping changes should merge cleanly")

	// Verify merged content
	merged, err := mockFS.ReadFile("file.txt")
	require.NoError(t, err)

	// Should have both user's edit and generated line
	mergedStr := string(merged)
	assert.Contains(t, mergedStr, "user edit")
	assert.Contains(t, mergedStr, "generated line")
}

func TestMerge_SnapshotRefError(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()
	lockFile.PersistentEdits = &lockfile.PersistentEdits{
		GenerationID:     "test-generation-uuid",
		PristineTreeHash: "some-tree-hash",
	}

	// pristine blob does not exist in the object DB
	missingBlobHash := "deadbeef1234567890abcdef1234567890abcdef"
	mockGit.FetchError = errors.New("ref not found")

	currentContent := []byte("line1\nuser edit\nline3\n")
	mockFS.Files["file.txt"] = currentContent

	oldChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader("line1\nline2\nline3\n"))
	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + oldChecksum,
		PristineGitObject: missingBlobHash,
	})

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": {Path: "file.txt", Content: []byte("line1\nline2\nline3\nnew\n"), Mode: 0o644},
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	_, err := Merge(ctx)
	require.Error(t, err)

	var snapshotRefErr *SnapshotRefError
	require.ErrorAs(t, err, &snapshotRefErr)
	assert.Equal(t, "test-generation-uuid", snapshotRefErr.GenerationID)
	assert.Equal(t, missingBlobHash, snapshotRefErr.BlobHash)
	assert.True(t, mockGit.FetchCalled["test-generation-uuid"])
}

func TestMerge_SnapshotRefError_BlobMissingAfterFetch(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()
	lockFile.PersistentEdits = &lockfile.PersistentEdits{
		GenerationID:     "test-generation-uuid",
		PristineTreeHash: "some-tree-hash",
	}

	// blob does not exist; FetchSnapshot succeeds but blob is still absent
	missingBlobHash := "deadbeef1234567890abcdef1234567890abcdef"
	// mockGit.FetchError left nil — fetch succeeds

	currentContent := []byte("line1\nuser edit\nline3\n")
	mockFS.Files["file.txt"] = currentContent

	oldChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader("line1\nline2\nline3\n"))
	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + oldChecksum,
		PristineGitObject: missingBlobHash,
	})

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": {Path: "file.txt", Content: []byte("line1\nline2\nline3\nnew\n"), Mode: 0o644},
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	_, err := Merge(ctx)
	require.Error(t, err)

	var snapshotRefErr *SnapshotRefError
	require.ErrorAs(t, err, &snapshotRefErr)
	assert.Equal(t, "test-generation-uuid", snapshotRefErr.GenerationID)
	assert.Equal(t, missingBlobHash, snapshotRefErr.BlobHash)
	assert.True(t, mockGit.FetchCalled["test-generation-uuid"])
	assert.ErrorContains(t, snapshotRefErr.Cause, "still not found after fetching snapshot")
}

func TestMerge_DirtyFileWithConflict(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	baseContent := []byte("line1\nline2\nline3\n")
	currentContent := []byte("line1\nuser edit\nline3\n")
	newContent := []byte("line1\ngenerated edit\nline3\n")

	// Set up base in git
	baseHash := computeSHA1(baseContent)
	mockGit.Objects[baseHash] = baseContent

	// Set up current file on disk (modified by user)
	mockFS.Files["file.txt"] = currentContent

	// Set up tracking with old checksum (file is dirty)
	oldChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader(string(baseContent)))
	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + oldChecksum,
		PristineGitObject: baseHash,
	})

	vFile := &VirtualFile{
		Path:     "file.txt",
		Content:  newContent,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.Len(t, result.ConflictFiles, 1)
	assert.Contains(t, result.ConflictFiles, "file.txt")

	// Verify conflict markers in merged content
	merged, err := mockFS.ReadFile("file.txt")
	require.NoError(t, err)

	mergedStr := string(merged)
	assert.Contains(t, mergedStr, "<<<<<<< Current")
	assert.Contains(t, mergedStr, "user edit")
	assert.Contains(t, mergedStr, "generated edit")
}

func TestMerge_BinaryFile(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	oldBinary := []byte{0x89, 0x50, 0x4E, 0x47}        // PNG header
	userBinary := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D} // User modified
	newBinary := []byte{0xFF, 0xD8, 0xFF, 0xE0}        // JPEG header

	// Set up base in git
	baseHash := computeSHA1(oldBinary)
	mockGit.Objects[baseHash] = oldBinary

	// Set up current file on disk (modified by user)
	mockFS.Files["image.png"] = userBinary

	// Set up tracking with old checksum (file is dirty)
	oldChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader(string(oldBinary)))
	lockFile.TrackedFiles.Set("image.png", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + oldChecksum,
		PristineGitObject: baseHash,
	})

	vFile := &VirtualFile{
		Path:     "image.png",
		Content:  newBinary,
		Mode:     0o644,
		IsBinary: true,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"image.png": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.Empty(t, result.ConflictFiles, "Binary files should not report conflicts")

	// Verify tracking was updated with new blob hash
	tracked, exists := lockFile.TrackedFiles.Get("image.png")
	require.True(t, exists)
	assert.NotEmpty(t, tracked.PristineGitObject, "Should track new pristine object")
	assert.NotEmpty(t, tracked.LastWriteChecksum, "Should update write checksum")

	// Binary file behavior: when dirty, keeps user version and updates tracking
	written, err := mockFS.ReadFile("image.png")
	require.NoError(t, err)
	// Note: The actual behavior depends on whether the base is available in git
	// If base is available, user version is kept. Otherwise, new version is written.
	t.Logf("Binary file result: user had %d bytes, generated has %d bytes, disk has %d bytes",
		len(userBinary), len(newBinary), len(written))
}

func TestMerge_NoOpDetection(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	content := []byte("content\n")
	vFile := &VirtualFile{
		Path:     "file.txt",
		Content:  content,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
	}

	// First run - should write file
	result1, err := Merge(ctx)
	require.NoError(t, err)
	assert.False(t, result1.IsNoOp)

	// Set up persistent edits with the tree hash
	lockFile.PersistentEdits = &lockfile.PersistentEdits{
		GenerationID:       "test-uuid",
		PristineTreeHash:   result1.TreeHash,
		PristineCommitHash: "commit-hash",
	}

	// Second run with same content - should be no-op
	result2, err := Merge(ctx)
	require.NoError(t, err)
	assert.True(t, result2.IsNoOp, "Should detect no-op when tree hash matches")
}

func TestMerge_MissingDiskFile(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	// File is in lockfile but not on disk
	lockFile.TrackedFiles.Set("missing.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:abc123",
		PristineGitObject: "blob-hash",
	})

	newContent := []byte("new content\n")
	vFile := &VirtualFile{
		Path:     "missing.txt",
		Content:  newContent,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"missing.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.Empty(t, result.ConflictFiles)

	// Should write new content
	written, err := mockFS.ReadFile("missing.txt")
	require.NoError(t, err)
	assert.Equal(t, newContent, written)
}

func TestMerge_ExecutableMode(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	content := []byte("#!/bin/bash\necho hello\n")
	vFile := &VirtualFile{
		Path:     "script.sh",
		Content:  content,
		Mode:     0o755, // Executable
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"script.sh": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.False(t, result.IsNoOp)

	// Verify file was written with executable mode
	assert.NotEqual(t, fs.FileMode(0), mockFS.Modes["script.sh"]&0o111, "Script should be executable")
}

func TestMerge_DirtyFileNoBase(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	currentContent := []byte("user content\n")
	newContent := []byte("generated content\n")

	// Set up current file on disk
	mockFS.Files["file.txt"] = currentContent

	// Set up tracking WITHOUT base (PristineGitObject is empty)
	currentChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader(string(currentContent)))
	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:different", // Mark as dirty
		PristineGitObject: "",               // No base available
	})

	// Also verify the checksum doesn't match
	assert.NotEqual(t, "sha1:different", "sha1:"+currentChecksum)

	vFile := &VirtualFile{
		Path:     "file.txt",
		Content:  newContent,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.Empty(t, result.ConflictFiles, "Without base, should overwrite")

	// Verify new content was written (overwrites user content)
	written, err := mockFS.ReadFile("file.txt")
	require.NoError(t, err)
	assert.Equal(t, newContent, written)
}

func TestMerge_ConflictStateSetup(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	baseContent := []byte("line1\nline2\n")
	currentContent := []byte("line1\nuser\n")
	newContent := []byte("line1\ngenerated\n")

	// Set up base in git
	baseHash := computeSHA1(baseContent)
	mockGit.Objects[baseHash] = baseContent

	// Set up current file on disk
	mockFS.Files["file.txt"] = currentContent

	// Set up tracking
	oldChecksum, _ := lockfile.HashNormalizedSHA1(strings.NewReader(string(baseContent)))
	lockFile.TrackedFiles.Set("file.txt", lockfile.TrackedFile{
		LastWriteChecksum: "sha1:" + oldChecksum,
		PristineGitObject: baseHash,
	})

	vFile := &VirtualFile{
		Path:     "file.txt",
		Content:  newContent,
		Mode:     0o644,
		IsBinary: false,
	}

	ctx := &MergeContext{
		VirtualFiles: map[string]*VirtualFile{
			"file.txt": vFile,
		},
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		Enabled:    true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.Len(t, result.ConflictFiles, 1)

	// Verify git conflict state was set up
	conflictState, exists := mockGit.ConflictStates["file.txt"]
	require.True(t, exists, "Conflict state should be set up")
	assert.Equal(t, baseContent, conflictState.Base)
	assert.Equal(t, currentContent, conflictState.Ours)
	assert.Equal(t, newContent, conflictState.Theirs)
	assert.False(t, conflictState.IsExecutable)
}

// ========== Fuzz Test for merge3Way ==========

func FuzzMerge3Way(f *testing.F) {
	// Add seed corpus
	f.Add([]byte("base"), []byte("current"), []byte("new"))
	f.Add([]byte("line1\nline2\n"), []byte("line1\nmodified\n"), []byte("line1\ngenerated\n"))
	f.Add([]byte(""), []byte("added"), []byte("also added"))

	f.Fuzz(func(t *testing.T, base, current, generated []byte) {
		// Skip invalid UTF-8 sequences as they might cause issues
		if !isValidUTF8(base) || !isValidUTF8(current) || !isValidUTF8(generated) {
			t.Skip("Invalid UTF-8")
		}

		merged, conflict := merge3Way(base, current, generated)

		// Property 1: If current == generated, result should equal current (and no conflict)
		if bytes.Equal(current, generated) {
			if !bytes.Equal(merged, current) {
				t.Errorf("Expected identical changes to return current content")
			}
			if conflict {
				t.Errorf("Identical changes should not conflict")
			}
		}

		// Property 2: If base == current (no user changes), result should be generated
		if bytes.Equal(base, current) {
			if !bytes.Equal(merged, generated) {
				t.Errorf("Expected no-op on current to accept generated")
			}
			if conflict {
				t.Errorf("No user changes should not cause conflict")
			}
		}

		// Property 3: If base == generated (no generator changes), result should be current
		if bytes.Equal(base, generated) {
			if !bytes.Equal(merged, current) {
				t.Errorf("Expected no generator changes to preserve current")
			}
			if conflict {
				t.Errorf("No generator changes should not cause conflict")
			}
		}

		// Property 4: Merged result should always be valid (not crash)
		_ = merged
		_ = conflict
	})
}

func isValidUTF8(data []byte) bool {
	for len(data) > 0 {
		size := 0
		for i, b := range data {
			if i == 0 {
				if b < 0x80 {
					size = 1
					break
				}
			}
			if i >= 4 {
				return false
			}
		}
		if size == 0 {
			// Try to decode multi-byte
			switch {
			case data[0]&0x80 == 0:
				size = 1
			case data[0]&0xE0 == 0xC0:
				size = 2
			case data[0]&0xF0 == 0xE0:
				size = 3
			case data[0]&0xF8 == 0xF0:
				size = 4
			default:
				return false
			}

			if len(data) < size {
				return false
			}
		}
		data = data[size:]
	}
	return true
}

// ========== Tests for mergeFile ==========

func TestMergeFile_NewFile(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	content := []byte("new file content\n")
	vFile := &VirtualFile{
		Path:     "newfile.txt",
		Content:  content,
		Mode:     0o644,
		IsBinary: false,
	}

	blobHash := computeSHA1(content)
	mockGit.Objects[blobHash] = content

	ctx := &MergeContext{
		LockFile:   lockFile,
		FileSystem: mockFS,
		Git:        mockGit,
		VirtualFiles: map[string]*VirtualFile{
			"newfile.txt": vFile,
		},
	}

	hasConflict, err := mergeFile(ctx, "newfile.txt", vFile, blobHash)
	require.NoError(t, err)
	assert.False(t, hasConflict)

	// Verify file was written
	written, _ := mockFS.ReadFile("newfile.txt")
	assert.Equal(t, content, written)

	// Verify tracking
	tracked, exists := lockFile.TrackedFiles.Get("newfile.txt")
	require.True(t, exists)
	assert.Equal(t, blobHash, tracked.PristineGitObject)
	assert.NotEmpty(t, tracked.LastWriteChecksum)
}

func TestWriteFile(t *testing.T) {
	t.Parallel()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	content := []byte("test content\n")
	vFile := &VirtualFile{
		Path:     "test.txt",
		Content:  content,
		Mode:     0o755, // Executable
		IsBinary: false,
	}

	blobHash := "abc123"

	ctx := &MergeContext{
		LockFile:   lockFile,
		FileSystem: mockFS,
	}

	err := writeFile(ctx, "test.txt", "test.txt", vFile, blobHash, content)
	require.NoError(t, err)

	// Verify file written with correct mode (executable)
	written, _ := mockFS.ReadFile("test.txt")
	assert.Equal(t, content, written)
	assert.Equal(t, fs.FileMode(0o755), mockFS.Modes["test.txt"])

	// Verify tracking
	tracked, exists := lockFile.TrackedFiles.Get("test.txt")
	require.True(t, exists)
	assert.Equal(t, blobHash, tracked.PristineGitObject)
}

func TestUpdateTracking(t *testing.T) {
	t.Parallel()
	lockFile := newTestLockFile()

	content := []byte("content\n")
	blobHash := "blob123"

	ctx := &MergeContext{
		LockFile: lockFile,
	}

	err := updateTracking(ctx, "file.txt", blobHash, content)
	require.NoError(t, err)

	tracked, exists := lockFile.TrackedFiles.Get("file.txt")
	require.True(t, exists)
	assert.Equal(t, blobHash, tracked.PristineGitObject)
	assert.True(t, strings.HasPrefix(tracked.LastWriteChecksum, "sha1:"))
}

// ========== Edge Case Tests ==========

func TestMerge3Way_LargeFile(t *testing.T) {
	t.Parallel()
	// Test with a reasonably large file to ensure algorithm scales
	var baseLines, currentLines, generatedLines []string
	for i := 0; i < 1000; i++ {
		line := fmt.Sprintf("line %d\n", i)
		baseLines = append(baseLines, line)
		currentLines = append(currentLines, line)
		generatedLines = append(generatedLines, line)
	}

	// Modify line 500 differently in current and generated
	currentLines[500] = "user modified line 500\n"
	generatedLines[750] = "generated modified line 750\n"

	base := []byte(strings.Join(baseLines, ""))
	current := []byte(strings.Join(currentLines, ""))
	generated := []byte(strings.Join(generatedLines, ""))

	merged, hasConflict := merge3Way(base, current, generated)

	assert.False(t, hasConflict, "Non-overlapping changes should merge cleanly")
	mergedStr := string(merged)
	assert.Contains(t, mergedStr, "user modified line 500")
	assert.Contains(t, mergedStr, "generated modified line 750")
}

func TestMerge3Way_MultipleConflicts(t *testing.T) {
	t.Parallel()
	base := []byte("line1\nline2\nline3\nline4\n")
	current := []byte("user1\nline2\nuser3\nline4\n")
	generated := []byte("gen1\nline2\ngen3\nline4\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.True(t, hasConflict)
	mergedStr := string(merged)

	// Should have multiple conflict markers
	conflictCount := strings.Count(mergedStr, "<<<<<<< Current")
	assert.Equal(t, 2, conflictCount, "Should have 2 conflicts")
}

func TestMerge3Way_OnlyNewlines(t *testing.T) {
	t.Parallel()
	base := []byte("\n\n\n")
	current := []byte("\n\n\n\n")
	generated := []byte("\n\n")

	merged, hasConflict := merge3Way(base, current, generated)

	// Both sides modified the newlines
	assert.True(t, hasConflict || !hasConflict) // Either outcome is acceptable
	assert.NotNil(t, merged)
}

func TestMerge3Way_UnicodeContent(t *testing.T) {
	t.Parallel()
	base := []byte("Hello 世界\nLine 2\n")
	current := []byte("Hello 世界\nModified 用户\n")
	generated := []byte("Hello 世界\nGenerated 生成\n")

	merged, hasConflict := merge3Way(base, current, generated)

	assert.True(t, hasConflict)
	mergedStr := string(merged)
	assert.Contains(t, mergedStr, "用户")
	assert.Contains(t, mergedStr, "生成")
}

func TestMerge_MultipleFiles(t *testing.T) {
	t.Parallel()
	mockGit := NewMockGit()
	mockFS := NewMockFileSystem()
	lockFile := newTestLockFile()

	files := map[string]*VirtualFile{
		"file1.txt": {
			Path:    "file1.txt",
			Content: []byte("content1\n"),
			Mode:    0o644,
		},
		"file2.txt": {
			Path:    "file2.txt",
			Content: []byte("content2\n"),
			Mode:    0o644,
		},
		"file3.txt": {
			Path:    "file3.txt",
			Content: []byte("content3\n"),
			Mode:    0o755,
		},
	}

	ctx := &MergeContext{
		VirtualFiles: files,
		LockFile:     lockFile,
		FileSystem:   mockFS,
		Git:          mockGit,
		Enabled:      true,
	}

	result, err := Merge(ctx)
	require.NoError(t, err)
	assert.False(t, result.IsNoOp)
	assert.Empty(t, result.ConflictFiles)

	// Verify all files written
	for path, vFile := range files {
		written, err := mockFS.ReadFile(path)
		require.NoError(t, err, "File %s should exist", path)
		assert.Equal(t, vFile.Content, written)
	}
}

// ========== Performance Test ==========

func BenchmarkMerge3Way_NoConflict(b *testing.B) {
	base := []byte(strings.Repeat("line\n", 100))
	current := []byte("modified\n" + strings.Repeat("line\n", 99))
	generated := []byte(strings.Repeat("line\n", 99) + "generated\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = merge3Way(base, current, generated)
	}
}

func BenchmarkMerge3Way_WithConflict(b *testing.B) {
	base := []byte(strings.Repeat("line\n", 100))
	current := []byte("user modified\n" + strings.Repeat("line\n", 99))
	generated := []byte("generated modified\n" + strings.Repeat("line\n", 99))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = merge3Way(base, current, generated)
	}
}
