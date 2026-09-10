package generate_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

var logger = logging.NewLogger(zapcore.DebugLevel)

const (
	cancelTestLang       = "typescript"
	cancelTestOutDir     = "../../testSDKs/sdk-ts-cancel"
	cancelTestSchemaPath = "../../tests/specs/cancel.yaml"
)

func generateWithCancel(ctx context.Context, opts ...generate.GeneratorOptions) (bool, []error) {
	ctx = generationaccess.WithDirect(ctx, generationaccess.ElectAGPL())

	schema, err := os.ReadFile(cancelTestSchemaPath)
	if err != nil {
		panic(err)
	}

	g, err := generate.New(opts...)
	if err != nil {
		panic(err)
	}

	return g.GenerateWithCancel(ctx, schema, cancelTestSchemaPath, cancelTestLang, cancelTestOutDir, false, true)
}

// Retrieve the temporary lock file (gets cleaned up and is only present when generation is aborted).
func getGeneratedFilesTempLockFile() (string, error) {
	pattern := filepath.Join(cancelTestOutDir, ".speakeasy", "generated-files-*.lock")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	if len(files) == 1 {
		return files[0], nil
	}
	return "", fmt.Errorf("expected exactly one generated-files lock file, found %d", len(files))
}

// Parse number of generated files from the temporary generated-files-<id>.lock file.
func getTempGeneratedFilesCount() (int, error) {
	filename, err := getGeneratedFilesTempLockFile()
	if err != nil {
		return 0, err
	}

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return count, nil
}

// Parse the number of generated files from the generatedFiles section in the final `gen.lock` file.
func getFinalGeneratedFilesCount() (int, error) {
	genlock := filepath.Join(cancelTestOutDir, ".speakeasy", "gen.lock")
	data, err := os.ReadFile(genlock)
	if err != nil {
		return 0, err
	}

	var lock struct {
		TrackedFiles map[string]any `yaml:"trackedFiles"`
	}

	if err := yaml.Unmarshal(data, &lock); err != nil {
		panic(err)
	}

	return len(lock.TrackedFiles), nil
}

func TestCancelBeforeCall(t *testing.T) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	aborted, errs := generateWithCancel(cancelCtx)

	assert.True(t, aborted)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Generation was cancelled", errs[0].Error())
}

func cancelDuringCleanup(cancelFunc context.CancelFunc) func(generate.ProgressUpdate) {
	return func(progressUpdate generate.ProgressUpdate) {
		if progressUpdate.File != nil && progressUpdate.File.Status == generate.ProgressFileStatusDeleted {
			cancelFunc()
		}
	}
}

func TestCancelDuringCleanup(t *testing.T) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
		generate.WithProgressUpdates(
			"testTargetID",
			cancelDuringCleanup(cancelFunc),
			true,
			true,
		),
	}

	_ = os.RemoveAll(cancelTestOutDir)

	// Create a temporary lock file with a couple orphaned files
	testID := "1234"
	speakeasy := filepath.Join(cancelTestOutDir, ".speakeasy")
	if err := os.MkdirAll(speakeasy, 0o755); err != nil {
		t.Fatal(err)
	}
	tmplock := filepath.Join(speakeasy, fmt.Sprintf("generated-files-%s.lock", testID))
	var tmpdata strings.Builder
	numOrphans := 3
	for i := 1; i <= numOrphans; i++ {
		orphan := fmt.Sprintf("orphan%d.md", i)
		tmpdata.WriteString(orphan)
		tmpdata.WriteString("\n")
		if _, err := os.Create(filepath.Join(cancelTestOutDir, orphan)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(tmplock, []byte(tmpdata.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a minimal gen.lock file with the same ID
	genlock := filepath.Join(speakeasy, "gen.lock")
	gendata := fmt.Sprintf("lockVersion: 2.0.0\nid: %s\nmanagement: {}\n", testID)
	if err := os.WriteFile(genlock, []byte(gendata), 0o644); err != nil {
		t.Fatal(err)
	}

	aborted, errs := generateWithCancel(cancelCtx, opts...)
	assert.True(t, aborted)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Generation was cancelled during step 'cleanup'", errs[0].Error())

	orphans, err := filepath.Glob(filepath.Join(cancelTestOutDir, "orphan*.md"))
	require.NoError(t, err)
	assert.Len(t, orphans, numOrphans-1, "expected %s orphaned files to remain", numOrphans-1)

	// The following test will check that orphaned files get cleaned up after a full run
}

// Expected number of generated files after a full successful run.
// Gets populated by TestNoCancel or runBaselineGeneration.
var expectedGeneratedFilesCount int

// ensureBaselineGenerated returns the expected file count, running a full
// generation first if TestNoCancel hasn't already populated the baseline.
func ensureBaselineGenerated(t *testing.T) int {
	t.Helper()
	if expectedGeneratedFilesCount > 0 {
		return expectedGeneratedFilesCount
	}

	_ = os.RemoveAll(cancelTestOutDir)
	ctx := context.Background()
	aborted, errs := generateWithCancel(ctx,
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
	)
	require.False(t, aborted, "baseline generation should not abort")
	require.Empty(t, errs, "baseline generation should not error")

	count, err := getFinalGeneratedFilesCount()
	require.NoError(t, err)
	require.Positive(t, count, "baseline should generate files")

	expectedGeneratedFilesCount = count
	return count
}

func TestNoCancel(t *testing.T) {
	// Keeping cancelTestOutDir from previous run to test orphaned files removal

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc() // Ensure context is cleaned up

	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
		generate.WithProgressUpdates(
			"testTargetID",
			func(generate.ProgressUpdate) {},
			false,
			false,
		),
	}

	aborted, errs := generateWithCancel(cancelCtx, opts...)
	assert.False(t, aborted)
	assert.Empty(t, errs)

	_, err := getTempGeneratedFilesCount()
	require.Error(t, err, "expected temp lock file to have been cleaned up")

	expectedGeneratedFilesCount, err = getFinalGeneratedFilesCount()
	require.NoError(t, err)
	assert.Positive(t, expectedGeneratedFilesCount, "expected some files to have been generated")

	orphans, err := filepath.Glob(filepath.Join(cancelTestOutDir, "orphan*.md"))
	require.NoError(t, err)
	assert.Empty(t, orphans, "expected all orphaned files to have been cleaned up")
}

func cancelBeforeGenerate(cancelFunc context.CancelFunc) func(generate.ProgressUpdate) {
	return func(progressUpdate generate.ProgressUpdate) {
		if progressUpdate.Step != nil && progressUpdate.Step.ID == generate.ProgressStepGenSDK {
			cancelFunc()
		}
	}
}

func TestCancelBeforeGenerate(t *testing.T) {
	_ = os.RemoveAll(cancelTestOutDir)

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
		generate.WithProgressUpdates(
			"testTargetID",
			cancelBeforeGenerate(cancelFunc),
			true,  // send update before each generation step
			false, // do not send file status updates
		),
	}
	aborted, errs := generateWithCancel(cancelCtx, opts...)
	assert.True(t, aborted)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Generation was cancelled during step 'genSDK'", errs[0].Error())

	entries, err := os.ReadDir(cancelTestOutDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, ".speakeasy", entries[0].Name())

	_, err = getTempGeneratedFilesCount()
	assert.Error(t, err, "expected no temp lock file to have been created yet")
}

func cancelOnFirstFileCreated(cancelFunc context.CancelFunc) func(generate.ProgressUpdate) {
	return func(progressUpdate generate.ProgressUpdate) {
		if progressUpdate.File != nil && progressUpdate.File.Status == generate.ProgressFileStatusCreated {
			cancelFunc()
		}
	}
}

func TestCancelDuringGenerate(t *testing.T) {
	_ = os.RemoveAll(cancelTestOutDir)

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
		generate.WithProgressUpdates(
			"testTargetID",
			cancelOnFirstFileCreated(cancelFunc),
			false, // do not send updates for each generation step
			true,  // send file status updates
		),
	}

	t.Setenv("SPEAKEASY_DEBUG_CONCURRENCY", "9")
	maxConcurrency := env.DebugConcurrency()

	aborted, errs := generateWithCancel(cancelCtx, opts...)
	assert.True(t, aborted)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Generation was cancelled during step 'genSDK'", errs[0].Error())

	count, err := getTempGeneratedFilesCount()
	require.NoError(t, err)
	assert.LessOrEqual(t, count, maxConcurrency, "expected no more than %v files to be generated, found %v", maxConcurrency, count)

	none, err := getFinalGeneratedFilesCount()
	require.NoError(t, err)
	assert.Equal(t, 0, none, "expected no generatedFiles section in gen.lock file")
}

var cancelTimeStamp time.Time

func cancelDuringCompile(cancelFunc context.CancelFunc, delay time.Duration) func(generate.ProgressUpdate) {
	return func(progressUpdate generate.ProgressUpdate) {
		if progressUpdate.Step != nil && progressUpdate.Step.ID == generate.ProgressStepCompileSDK {
			go func() {
				<-time.After(delay)
				cancelTimeStamp = time.Now()
				cancelFunc()
			}()
		}
	}
}

func TestCancelDuringCompile(t *testing.T) {
	baseline := ensureBaselineGenerated(t)

	_ = os.RemoveAll(cancelTestOutDir)
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithLogger(logger),
		generate.WithProgressUpdates(
			"testTargetID",
			cancelDuringCompile(cancelFunc, 1*time.Second),
			true,
			false,
		),
	}

	aborted, errs := generateWithCancel(cancelCtx, opts...)
	timeToExit := time.Since(cancelTimeStamp)

	assert.True(t, aborted)
	assert.Len(t, errs, 1)
	assert.Equal(t, "Generation was cancelled during step 'compileSDK'", errs[0].Error())

	// Use a generous slack (400ms) to account for slow CI runners where
	// process cleanup can take much longer than on local machines.
	maxTimeToExit := processrunner.CanceledCommandWaitDelay + (400 * time.Millisecond)
	assert.Less(t, timeToExit, maxTimeToExit, "Exiting took %v", timeToExit)

	count, err := getFinalGeneratedFilesCount()
	require.NoError(t, err)
	assert.Equal(t, baseline, count, "expected %v files to be generated, found %v", baseline, count)
}
