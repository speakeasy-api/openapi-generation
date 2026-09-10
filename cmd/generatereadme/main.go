package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/runner"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap/zapcore"
)

var logger = logging.NewLogger(zapcore.DebugLevel)

func main() {
	lang := flag.String("l", "", "language to generate standalone README for")
	schemaPath := flag.String("s", "", "path to the schema file")
	outDir := flag.String("o", ".", "path to the output directory")
	fileName := flag.String("f", "README.md", "name of the README file")
	headersOnly := flag.Bool("headers-only", false, "template section headers only")

	flag.Parse()

	if err := genReadme(*lang, *schemaPath, *outDir, *fileName, *headersOnly); len(err) > 0 {
		for _, e := range err {
			logger.Error(e.Error())
		}
		logger.Fatal("Failed to generate Standalone README")
	}
}

func genReadme(lang, schemaPath, outDir, fileName string, headersOnly bool) []error {
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return []error{fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)}
	}

	outFile := fmt.Sprintf("%s/%s", outDir, fileName)
	whitelist := map[string]interface{}{
		outFile: nil,
	}

	opts := []generate.GeneratorOptions{
		generate.WithStandaloneReadme(fileName, headersOnly),
		generate.WithDebuggingEnabled(),
		generate.WithFileSystem(&fileSystem{whiteList: whitelist}),
		generate.WithLogger(logger),
		generate.WithRunLocation("cli"),
	}

	g, err := generate.New(opts...)
	if err != nil {
		return []error{err}
	}

	ctx := generationaccess.WithDirect(context.Background())
	if factory := runner.ResolveGenerationContextFactory("", "", os.Getenv); factory != nil {
		if ctx, err = factory(); err != nil {
			return []error{err}
		}
	}
	if errs := g.Generate(ctx, schema, schemaPath, lang, outDir, false, false); len(errs) > 0 {
		return errs
	}

	logger.Debug(fmt.Sprintf("Generated Standalone README %s for %s with %s", outFile, lang, filepath.Base(schemaPath)))

	return nil
}

type fileSystem struct {
	whiteList map[string]interface{}
}

var _ filesystem.FileSystem = &fileSystem{}

func (fs *fileSystem) ReadFile(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}

func (fs *fileSystem) Remove(name string) error {
	return nil
}

func (fs *fileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	if _, ok := fs.whiteList[path]; ok {
		dir := filepath.Dir(path)
		_, err := fs.Stat(dir)
		if err != nil {
			err := os.MkdirAll(path, 0o755)
			if err != nil {
				return err
			}
		}

		return os.WriteFile(path, data, perm)
	}

	return nil
}

func (fs *fileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (fs *fileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (fs *fileSystem) OpenFile(name string, flag int, perm os.FileMode) (filesystem.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (fs *fileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (f *fileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (fs *fileSystem) ScanForGeneratedIDs() (map[string]string, error) {
	return nil, nil
}
