package generate

import (
	"context"
	"fmt"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filetracking"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/js/template"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/merge"
)

const defaultFileMode = 0o644 // readable by all, writable by owner

var _ filesystem.FileSystem = &Generator{}

func (g *Generator) Open(name string) (fs.File, error) {
	if g.fs == nil {
		return os.Open(name)
	}

	return g.fs.Open(name)
}

func (g *Generator) OpenFile(name string, flag int, perm fs.FileMode) (filesystem.File, error) {
	if g.fs == nil {
		return os.OpenFile(name, flag, perm)
	}

	return g.fs.OpenFile(name, flag, perm)
}

func (g *Generator) Stat(name string) (fs.FileInfo, error) {
	if g.fs == nil {
		return os.Stat(name)
	}

	return g.fs.Stat(name)
}

func (g *Generator) Remove(name string) error {
	if g.fs == nil {
		return os.Remove(name)
	}

	return g.fs.Remove(name)
}

func (g *Generator) WriteFile(name string, data []byte, perm fs.FileMode) error {
	if g.dontWrite {
		return nil
	}

	if err := g.createDirectory(name); err != nil {
		return err
	}

	if g.fs == nil {
		return os.WriteFile(name, data, perm)
	}

	return g.fs.WriteFile(name, data, perm)
}

func (g *Generator) ReadFile(name string) ([]byte, error) {
	if g.fs == nil {
		return os.ReadFile(name)
	}

	return g.fs.ReadFile(name)
}

func (g *Generator) MkdirAll(path string, perm os.FileMode) error {
	if g.fs == nil {
		return os.MkdirAll(path, perm)
	}

	return g.fs.MkdirAll(path, perm)
}

func (g *Generator) ReadDir(name string) ([]fs.DirEntry, error) {
	if g.fs == nil {
		return os.ReadDir(name)
	}

	return g.fs.ReadDir(name)
}

func (g *Generator) ScanForGeneratedIDs() (map[string]string, error) {
	if g.fs == nil {
		// No filesystem injected, return empty map
		return nil, nil
	}

	return g.fs.ScanForGeneratedIDs()
}

func (g *Generator) createDirectory(filename string) error {
	dir := filepath.Dir(filename)

	_, err := g.Stat(dir)
	if err != nil {
		err := g.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
	}
	return nil
}

// docsDisabled reports whether per-operation documentation generation is
// turned off via `generation.documentation: none` in gen.yaml.
func (g *Generator) docsDisabled() bool {
	return g.subsystem.Config != nil && g.subsystem.Config.DocsDisabled()
}

// mintlifyEnabled reports whether documentation should be emitted as
// Mintlify MDX (`generation.documentation: mintlify` in gen.yaml).
func (g *Generator) mintlifyEnabled() bool {
	return g.subsystem.Config != nil && g.subsystem.Config.IsMintlify()
}

func (g *Generator) onWriteFile(filePathPrefix string) template.WriteFileFunc {
	return func(ctx context.Context, filename string, data []byte, perm fs.FileMode, checkExisting bool) error {
		filename = filetracking.NormalizePath(filename)

		// Documentation disabled (`generation.documentation: none`):
		// skip writing per-operation docs entirely instead of emitting and then
		// deleting them. Checked against the template-relative path — before
		// the filePathPrefix join — so docs of nested SDKs (e.g. terraform's
		// internal/sdk) are suppressed too. Skipped files are never tracked, so
		// docs left over from a previous run become untracked and are cleaned
		// up as orphans on the next generation.
		if g.docsDisabled() && isDocumentationFile(filepath.ToSlash(filename)) {
			return nil
		}

		if filePathPrefix != "" {
			filename = filepath.Join(filePathPrefix, filename)
		}

		filename = filepath.ToSlash(filename)

		// If persistent edits is enabled, check if user moved this file
		// and remap to the actual location to respect their change.
		// The patches DB is keyed by the template-output filename, so this
		// lookup must happen against the original `.md` name — before the
		// mintlify transform renames it to `.mdx`.
		originalFilename := filename
		filename = g.subsystem.Patches.GetActualPath(filename)

		// Mintlify: after persistent-edit resolution, transform data and
		// rename `.md` → `.mdx` for the disk path. Also rewrite
		// originalFilename so FileTracker records the actual disk name —
		// cleanup compares the tracker against files on disk, so any later
		// run that disables the flag can collect the orphaned `.mdx` files.
		if g.mintlifyEnabled() {
			filename, data = applyMintlifyTransform(filename, data)
			originalFilename = mintlifyOutputPath(originalFilename)
		}

		if !g.shouldWrite(ctx, filename) || g.validationOnly {
			return nil
		}

		if perm == 0 {
			perm = defaultFileMode
		}

		// Don't write empty files
		if strings.TrimSpace(string(data)) == "" {
			return nil
		}

		// Add header if needed
		if !g.subsystem.FileTracker.IsUntrackedPatternFile(filename) {
			if g.subsystem.FileTracker.AddHeaderToFile(filename) {
				// Get the generated ID if persistent edits is enabled
				var generatedID string
				if g.subsystem.Patches.IsEnabled() {
					generatedID = g.subsystem.Patches.GetOrCreateID(g.outDir + "/" + filename)
				}

				header := g.subsystem.Features.GetAutoGeneratedHeader(context.Background(), generatedID)

				if header != "" {
					data = append([]byte(header), data...)
				}
			}

			g.subsystem.FileTracker.TrackFile(originalFilename, data)
		}

		// Check if binary by looking for null bytes in first 8KB
		isBinary := IsBinaryContent(data)

		g.virtualFiles.Store(filename, &merge.VirtualFile{
			Path:     filename,
			Content:  data,
			Mode:     perm,
			IsBinary: isBinary,
		})

		// Write directly if custom code not enabled
		outFileName := filepath.Join(g.outDir, filename)

		if (env.IsDebug() && g.verboseOutput) || env.DebugWriteFiles() {
			absFile, err := filepath.Abs(outFileName)
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("Writing File %s - len=%v", absFile, len(data))
			writeFileMutex.Lock()
			writeFileCount[outFileName]++
			if writeFileCount[outFileName] > 1 {
				msg += fmt.Sprintf(" (duplicate_write=%v)", writeFileCount[outFileName])
			}
			writeFileMutex.Unlock()
			logging.From(ctx).Debug(msg)
			defer logging.From(ctx).Debug("Done " + filename)
		}

		g.sendFileProgressUpdate(outFileName, data)

		if g.subsystem.Patches.IsEnabled() {
			// we write out virtual files at end after merging.
			return nil
		}

		return g.WriteFile(outFileName, data, perm)
	}
}

var (
	writeFileMutex = sync.Mutex{}
	writeFileCount = map[string]int{}
)

// IsBinaryContent detects if content is binary by checking for null bytes (same as git)
func IsBinaryContent(data []byte) bool {
	checkLen := 8192
	if len(data) < checkLen {
		checkLen = len(data)
	}
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}

func (g *Generator) onReadFile(filePathPrefix string) template.ReadFileFunc {
	return func(filename string) ([]byte, error) {
		filename = filetracking.NormalizePath(filename)

		if filePathPrefix != "" {
			filename = filepath.Join(filePathPrefix, filename)
		}

		filename = filepath.ToSlash(filename)

		// If persistent edits is enabled, check if user moved this file
		// and remap to the actual location to respect their change
		filename = g.subsystem.Patches.GetActualPath(filename)

		// Mirror the write-side mintlify rename so a template that wrote
		// `docs/foo.md` and reads it back by name finds the `.mdx`-keyed
		// virtualFiles entry (and the `.mdx` file on disk).
		if g.mintlifyEnabled() {
			filename = mintlifyOutputPath(filename)
		}

		if virtualFile, ok := g.virtualFiles.Load(filename); ok {
			return virtualFile.Content, nil
		}

		return g.ReadFile(filepath.Join(g.outDir, filename))
	}
}
