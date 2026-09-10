package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/runner"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap/zapcore"
)

var logger = logging.NewLogger(zapcore.DebugLevel)

func main() {
	schema := flag.String("s", "", "path to the schema file")
	outDir := flag.String("o", ".", "path to the output directory")
	lang := flag.String("l", "", "language to generate")
	installationURL := flag.String("i", "", "installation url for the SDK")
	operationIDs := flag.String("op", "", "operationIDs for usage snippet (comma separated)")
	groupName := flag.String("g", "", "Group Name for usage snippet")
	all := flag.Bool("a", false, "generate usage snippet for all operations")
	flag.Parse()

	outputBuffer := &bytes.Buffer{}

	operationIDsArray := strings.Split(*operationIDs, ",")

	if *operationIDs == "" {
		*all = true
	}

	if err := genUsageSnippet(*lang, *schema, *outDir, *installationURL, operationIDsArray, *groupName, *all, outputBuffer); len(err) > 0 {
		for _, e := range err {
			logger.Error(e.Error())
		}

		logger.Fatal("Failed to generate SDK")
	}

	fmt.Print(outputBuffer.String())
}

func genUsageSnippet(lang, schemaPath, outDir, installationURL string, operationIDs []string, groupName string, all bool, outputBuffer *bytes.Buffer) []error {
	fmt.Printf("Generating Usage Snippet for %s...\n", lang)

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return []error{fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)}
	}

	opts := []generate.GeneratorOptions{
		generate.WithDebuggingEnabled(),
		generate.WithFileSystem(&FileSystem{buf: outputBuffer}),
		generate.WithInstallationURL(installationURL),
		generate.WithLogger(logger),
		generate.WithRunLocation("cli"),
	}

	switch {
	case all:
		opts = append(opts, generate.WithUsageSnippetArgsGenerateAll())
	case len(operationIDs) == 0 && groupName == "":
		opts = append(opts, generate.WithUsageSnippetArgsByRootExample())
	case len(operationIDs) > 0:
		opts = append(opts, generate.WithUsageSnippetArgsByOperationID(strings.Join(operationIDs, ",")))
	default:
		opts = append(opts, generate.WithUsageSnippetArgsByNamespace(groupName))
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

	fmt.Printf("Generated Usage Snippet for %s\n", lang)

	return nil
}

type FileSystem struct {
	buf *bytes.Buffer
}

var _ filesystem.FileSystem = &FileSystem{}

func (fs *FileSystem) ReadFile(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}

func (fs *FileSystem) Remove(name string) error {
	return nil
}

func (fs *FileSystem) WriteFile(outFileName string, data []byte, mode os.FileMode) error {
	// Make this resilient to additional files being inadvertently written
	if strings.Contains(outFileName, ".usage.") {
		snippet := string(data)
		snippet += "\n\n"
		_, err := fs.buf.Write([]byte(snippet))
		return err
	}

	return nil
}

func (fs *FileSystem) MkdirAll(path string, mode os.FileMode) error {
	return nil
}

func (fs *FileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (fs *FileSystem) OpenFile(name string, flag int, perm os.FileMode) (filesystem.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (fs *FileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (f *FileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (fs *FileSystem) ScanForGeneratedIDs() (map[string]string, error) {
	return nil, nil
}
