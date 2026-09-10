package filetracking

import (
	"context"
	stderrors "errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkipInaccessibleDir(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "dir"), 0o755))
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	dir := entries[0]

	permissionErr := &fs.PathError{Op: "open", Path: "dir", Err: fs.ErrPermission}
	require.ErrorIs(t, skipInaccessibleDir(dir, permissionErr), fs.SkipDir)

	require.NoError(t, skipInaccessibleDir(dir, fs.ErrNotExist))

	readErr := stderrors.New("read failed")
	require.ErrorIs(t, skipInaccessibleDir(dir, readErr), readErr)
	require.ErrorIs(t, skipInaccessibleDir(nil, permissionErr), fs.ErrPermission)
}

func TestTrackerGetResultImmediatelyAfterNewTracker(t *testing.T) {
	previousMaxProcs := runtime.GOMAXPROCS(1)
	defer runtime.GOMAXPROCS(previousMaxProcs)

	for i := 0; i < 100; i++ {
		tracker := NewTracker(context.Background(), Options{
			OutDir:           t.TempDir(),
			GenLockId:        fmt.Sprintf("test-%d", i),
			FS:               &MockFileSystem{},
			SkipTempLockFile: true,
		})

		done := make(chan struct{})
		go func() {
			_ = tracker.GetResult()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("GetResult blocked when called immediately after NewTracker")
		}
	}
}

func TestTrackerSkipChecksums(t *testing.T) {
	for _, tt := range []struct {
		name          string
		skipChecksums bool
	}{
		{name: "checksums recorded by default", skipChecksums: false},
		{name: "checksums omitted when skipped", skipChecksums: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(context.Background(), Options{
				OutDir:           t.TempDir(),
				GenLockId:        "test",
				FS:               &MockFileSystem{},
				SkipTempLockFile: true,
				SkipChecksums:    tt.skipChecksums,
			})

			tracker.TrackFile("channel.go", []byte("package channel"))
			require.NoError(t, tracker.TrackFileImmediate("immediate.go", []byte("package immediate")))

			result := tracker.GetResult()
			for _, path := range []string{"channel.go", "immediate.go"} {
				tracked, ok := result.NewFiles.Get(path)
				require.True(t, ok, path)
				if tt.skipChecksums {
					assert.Empty(t, tracked.LastWriteChecksum, path)
				} else {
					assert.NotEmpty(t, tracked.LastWriteChecksum, path)
				}
			}

			tracker.UpdateTrackedFileChecksum("channel.go", []byte("package channel2"))
			tracked, ok := tracker.newFiles.Get("channel.go")
			require.True(t, ok)
			if tt.skipChecksums {
				assert.Empty(t, tracked.LastWriteChecksum)
			} else {
				assert.NotEmpty(t, tracked.LastWriteChecksum)
			}
		})
	}
}
