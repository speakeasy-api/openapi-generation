package arazzo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/tests"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/workspace"
)

const (
	TestArazzoDocFileName = "tests.arazzo.yaml"
	ExtTestRebuild        = "x-speakeasy-test-rebuild"
)

type ManageArazzoDocOptions struct {
	OutDir             string
	OpenAPIDocPath     string
	AST                *ast.AST
	FileSystem         workspace.FS
	LockFile           *config.LockFile
	IsInternalTestSpec bool
	LoadOnly           bool
	Config             *configuration.Config
}

func ManageArazzoDoc(ctx context.Context, opts ManageArazzoDocOptions) (*tests.ArazzoDocumentInfo, error) {
	generateTests := opts.Config.Generation.Tests.GenerateTests

	if !generateTests {
		return nil, nil
	}

	arazzoDoc, arazzoPath, validationErrors, err := loadArazzoDoc(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(validationErrors) > 0 {
		// TODO a document exists it just isn't passing validation so make sure not to overwrite it or modify it
		for _, err := range validationErrors {
			logging.LogWarning(ctx, "arazzo validation error", err)
		}
	}
	docInfo := &tests.ArazzoDocumentInfo{
		Location: arazzoPath,
		Doc:      arazzoDoc,
	}

	if opts.LoadOnly {
		return docInfo, nil
	}

	if opts.LockFile.GeneratedTests == nil {
		opts.LockFile.GeneratedTests = sequencedmap.New[string, string]()
	}

	regenerateTests := make(map[string]bool)

	if arazzoDoc != nil {
		// Record any new tests that might have been added to the arazzo document so we don't auto generate a test for that operation
		for item := range arazzo.Walk(ctx, arazzoDoc) {
			_ = item.Match(arazzo.Matcher{
				Workflow: func(workflow *arazzo.Workflow) error {
					if workflow.Extensions == nil {
						return nil
					}

					val, ok := workflow.Extensions.Get(ExtTestRebuild)
					if ok {
						regenerateTests[workflow.WorkflowID] = val.Value == "true"
					} else {
						regenerateTests[workflow.WorkflowID] = false
					}
					return nil
				},
				Step: func(step *arazzo.Step) error {
					if step.OperationID == nil {
						return nil
					}

					// TODO will need to deal with expressions at some point
					if !step.OperationID.IsExpression() {
						opID := string(*step.OperationID)
						if _, ok := opts.LockFile.GeneratedTests.Get(opID); !ok {
							opts.LockFile.GeneratedTests.Set(opID, time.Now().Format(time.RFC3339))
						}
					}
					return nil
				},
			})
		}
	}

	arazzoDoc, err = generateTestsFromSpec(ctx, generateTestsFromSpecOptions{
		AST:                opts.AST,
		ArazzoDoc:          arazzoDoc,
		RegenerateTests:    regenerateTests,
		OpenAPIDocPath:     opts.OpenAPIDocPath,
		LockFile:           opts.LockFile,
		IsInternalTestSpec: opts.IsInternalTestSpec,
		Config:             opts.Config,
	})
	if err != nil {
		return nil, err
	}

	if arazzoDoc == nil || len(arazzoDoc.Workflows) == 0 {
		return nil, nil
	}

	if err := arazzoDoc.Sync(ctx); err != nil {
		return nil, err
	}

	if err := writeArazzoDoc(ctx, arazzoDoc, arazzoPath, opts); err != nil {
		return nil, err
	}

	arazzoDoc, _, validationErrors, err = loadArazzoDoc(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(validationErrors) > 0 {
		for _, err := range validationErrors {
			logging.LogWarning(ctx, "arazzo validation error", err)
		}
		return nil, errors.New("failed to validate arazzo doc")
	}
	docInfo.Doc = arazzoDoc

	return docInfo, nil
}

func loadArazzoDoc(ctx context.Context, opts ManageArazzoDocOptions) (*arazzo.Arazzo, string, []error, error) {
	arazzoData, arazzoPath, err := findTestArazzoDoc(ctx, opts.OutDir, opts.FileSystem)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, "", nil, err
	}

	var arazzoDoc *arazzo.Arazzo
	var validationErrors []error

	if arazzoData != nil && arazzoPath != "" {
		var err error
		arazzoDoc, validationErrors, err = arazzo.Unmarshal(ctx, bytes.NewReader(arazzoData))
		if err != nil {
			return nil, "", nil, err
		}
	}

	return arazzoDoc, arazzoPath, validationErrors, nil
}

func writeArazzoDoc(ctx context.Context, arazzoDoc *arazzo.Arazzo, arazzoPath string, opts ManageArazzoDocOptions) error {
	if arazzoPath == "" {
		res, err := workspace.FindWorkspace(opts.OutDir, workspace.FindWorkspaceOptions{
			FS:        opts.FileSystem,
			Recursive: true,
		})
		if err == nil && res != nil {
			arazzoPath = filepath.Join(res.Path, TestArazzoDocFileName)
		}
	}
	if arazzoPath == "" {
		arazzoPath = filepath.Join(opts.OutDir, workspace.SpeakeasyFolder, TestArazzoDocFileName)
	}

	arazzoPath, err := filepath.Abs(arazzoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", arazzoPath, err)
	}

	f, err := os.Create(arazzoPath)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", arazzoPath, err)
	}
	defer f.Close()
	if err := arazzo.Marshal(ctx, arazzoDoc, f); err != nil {
		return err
	}
	logging.From(ctx).Info(arazzoPath + " updated")

	return nil
}

func findTestArazzoDoc(_ context.Context, outDir string, fileSystem workspace.FS) ([]byte, string, error) {
	// TODO we may want this path set and retrieved from the workflow.yaml instead of a hardcoded file that needs to exist
	testArazzoRes, err := workspace.FindWorkspace(outDir, workspace.FindWorkspaceOptions{
		FindFile:  TestArazzoDocFileName,
		FS:        fileSystem,
		Recursive: true,
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("failed to find test arazzo doc: %w", err)
	}

	if testArazzoRes == nil {
		return nil, "", nil
	}

	return testArazzoRes.Data, testArazzoRes.Path, nil
}

func createArazzoDoc(summary, openAPIDocPath string) *arazzo.Arazzo {
	return &arazzo.Arazzo{
		Arazzo: arazzo.Version,
		Info: arazzo.Info{
			Title:   "Test Suite",
			Version: "0.0.1",
			Summary: pointer.From(summary),
		},
		SourceDescriptions: []*arazzo.SourceDescription{
			{
				Name: openAPIDocPath,
				Type: arazzo.SourceDescriptionTypeOpenAPI,
				URL:  "https://TBD.com", // TODO how can we get a usable URL here?
			},
		},
		Workflows: []*arazzo.Workflow{},
	}
}
