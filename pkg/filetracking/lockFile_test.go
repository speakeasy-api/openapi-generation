package filetracking

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
)

// Implements enough of the filesystem interface to test the lockFile package
type MockFileSystem struct {
	filesystem.FileSystem
}

func (m *MockFileSystem) OpenFile(name string, flag int, perm fs.FileMode) (filesystem.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (m *MockFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (m *MockFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func TestNewLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)

	lf, previousFiles, err := newLockFile(Options{
		OutDir:    tmpDir,
		GenLockId: "test",
		FS:        &MockFileSystem{},
	})
	if err != nil {
		t.Fatalf("newLockFile() error = %v", err)
	}
	defer lf.Destroy() //nolint:errcheck

	// Expect the file does not exist
	if _, err := os.Stat(lf.path); !os.IsNotExist(err) {
		t.Errorf("file exists after newLockFile: %v", err)
	}

	if len(previousFiles) != 0 {
		t.Errorf("expected empty previousFiles, got %v", previousFiles)
	}
}

func TestExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)
	fs := &MockFileSystem{}

	dir := filepath.Join(tmpDir, ".speakeasy")
	_ = fs.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "generated-files-test.lock")
	content := "path1\npath2\npath3\n"
	_ = fs.WriteFile(path, []byte(content), 0644)

	lf, previousFiles, err := newLockFile(Options{
		OutDir:    tmpDir,
		GenLockId: "test",
		FS:        fs,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer lf.Destroy() //nolint:errcheck

	expected := []string{"path1", "path2", "path3"}
	if !reflect.DeepEqual(previousFiles, expected) {
		t.Errorf("expected previousFiles %v, got %v", expected, previousFiles)
	}
}

func TestWriteLine(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)

	lf, previousFiles, err := newLockFile(Options{
		OutDir:    tmpDir,
		GenLockId: "test",
		FS:        &MockFileSystem{},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer lf.Destroy() //nolint:errcheck

	// Expect previousFiles to be empty
	if len(previousFiles) != 0 {
		t.Errorf("expected empty previousFiles, got %v", previousFiles)
	}

	line := "test/path"
	err = lf.WriteLine(line)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(lf.path)
	if err != nil {
		t.Fatal(err)
	}
	expectedContent := line + "\n"
	if string(data) != expectedContent {
		t.Errorf("file content = %q, want %q", data, expectedContent)
	}

	lf2, previousFiles2, err := newLockFile(Options{
		OutDir:    tmpDir,
		GenLockId: "test",
		FS:        &MockFileSystem{},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer lf2.Destroy() //nolint:errcheck

	// Expect previousFiles2 to not be empty and contain the line
	if len(previousFiles2) != 1 || previousFiles2[0] != line {
		t.Errorf("expected previousFiles2 to contain %q, got %v", line, previousFiles2)
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("TMPDIR", tmpDir)

	lf, _, err := newLockFile(Options{
		OutDir:    tmpDir,
		GenLockId: "test",
		FS:        &MockFileSystem{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := lf.WriteLine("test/path"); err != nil {
		t.Fatal(err)
	}

	err = lf.Destroy()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(lf.path); !os.IsNotExist(err) {
		t.Errorf("file still exists after Clear: %v", err)
	}
}
