package generate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/speakeasy-api/openapi-generation/v2/internal/patches"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

type ProgressStepID string

const (
	ProgressStepSetup             ProgressStepID = "setup"
	ProgressStepValidate          ProgressStepID = "validate"
	ProgressStepGenSDK            ProgressStepID = "genSDK"
	ProgressStepGenMockServer     ProgressStepID = "genMockServer"
	ProgressStepCleanup           ProgressStepID = "cleanup"
	ProgressStepLockFile          ProgressStepID = "lockFile"
	ProgressStepCompileSDK        ProgressStepID = "compileSDK"
	ProgressStepCompileTests      ProgressStepID = "compileTests"
	ProgressStepLintSDK           ProgressStepID = "lintSDK"
	ProgressStepCompileUsage      ProgressStepID = "compileUsage"
	ProgressStepCompileMockServer ProgressStepID = "compileMockServer"
	ProgressStepLintMockServer    ProgressStepID = "lintMockServer"
	ProgressStepDone              ProgressStepID = "done"
	ProgressStepCancel            ProgressStepID = "cancel"

	// Custom code / persistent edits merge steps
	ProgressStepMergeStart        ProgressStepID = "mergeStart"
	ProgressStepMergeSkipped      ProgressStepID = "mergeSkipped"
	ProgressStepMergeFetch        ProgressStepID = "mergeFetch"
	ProgressStepMergeSnapshot     ProgressStepID = "mergeSnapshot"
	ProgressStepMerge             ProgressStepID = "merge"
	ProgressStepMergeStepSkipped  ProgressStepID = "mergeStepSkipped"
	ProgressStepMergePush         ProgressStepID = "mergePush"
	ProgressStepMergePushSkipped  ProgressStepID = "mergePushSkipped"
	ProgressStepMergeApply        ProgressStepID = "mergeApply"
	ProgressStepMergeApplySkipped ProgressStepID = "mergeApplySkipped"
)

var ProgressMessages = map[ProgressStepID]string{
	ProgressStepSetup:             "Setup Environment",
	ProgressStepValidate:          "Load and Validate Document",
	ProgressStepGenSDK:            "Generate SDK",
	ProgressStepGenMockServer:     "Generate Mock Server",
	ProgressStepCleanup:           "Cleanup Orphaned Files",
	ProgressStepLockFile:          "Update Lock File",
	ProgressStepCompileSDK:        "Compile SDK",
	ProgressStepCompileTests:      "Compile Tests",
	ProgressStepLintSDK:           "Lint SDK",
	ProgressStepCompileUsage:      "Compile Usage Snippets",
	ProgressStepCompileMockServer: "Compile Mock Server",
	ProgressStepLintMockServer:    "Lint Mock Server",
	ProgressStepDone:              "Done",
	ProgressStepCancel:            "Cancelled",

	// Custom code / persistent edits merge steps
	ProgressStepMergeStart:        "Custom Code",
	ProgressStepMergeSkipped:      "Custom Code (skipped)",
	ProgressStepMergeFetch:        "Fetching custom code history",
	ProgressStepMergeSnapshot:     "Creating generation snapshot",
	ProgressStepMerge:             "Merging custom edits",
	ProgressStepMergeStepSkipped:  "Merging custom edits (skipped)",
	ProgressStepMergePush:         "Pushing snapshot to remote",
	ProgressStepMergePushSkipped:  "Pushing snapshot to remote (skipped)",
	ProgressStepMergeApply:        "Applying merged files",
	ProgressStepMergeApplySkipped: "Applying merged files (skipped)",
}

type ProgressStep struct {
	ID      ProgressStepID
	Message string
}

type ProgressFileStatus string

const (
	ProgressFileStatusCreated   ProgressFileStatus = "created"
	ProgressFileStatusUnchanged ProgressFileStatus = "unchanged"
	ProgressFileStatusDeleted   ProgressFileStatus = "deleted"
	ProgressFileStatusModified  ProgressFileStatus = "modified"
)

type ProgressFile struct {
	Path         string
	Status       ProgressFileStatus
	Content      *bytes.Buffer
	IsMainReadme bool
}

type ProgressUpdate struct {
	TargetID string
	Step     *ProgressStep
	File     *ProgressFile
	Merge    *patches.MergeProgress // Custom code merge events (conflicts, push warnings)
}

type Progress struct {
	TargetID         string               // identifier for the target being processed
	StepID           ProgressStepID       // ID of the current step
	OnProgressUpdate func(ProgressUpdate) // callback for progress updates
	UpdateGenSteps   bool                 // whether to send updates before each generation step
	UpdateFileStatus bool                 // whether to send file status updates
}

func (g *Generator) withStep(ctx context.Context, stepID ProgressStepID, fn func(ctx context.Context) error) error {
	if isCancelled(ctx) {
		return errors.New("context cancelled")
	}

	g.progress.StepID = stepID
	g.sendStepProgressUpdate(g.logStep(stepID))

	stepName := ProgressMessages[stepID]

	return logging.WithStepCtx(ctx, stepName, func(ctx context.Context) error {
		return fn(ctx)
	})
}

// Whether the generator has been configured to send progress updates.
func (g *Generator) canUpdateProgress() bool {
	return g.progress.OnProgressUpdate != nil
}

// Send a progress update with the given step ID and message.
func (g *Generator) sendStepProgressUpdate(message string) {
	if g.canUpdateProgress() && g.progress.UpdateGenSteps {
		g.progress.OnProgressUpdate(ProgressUpdate{
			TargetID: g.progress.TargetID,
			Step: &ProgressStep{
				ID:      g.progress.StepID,
				Message: message,
			},
		})
	}
}

// Log the start of a generation step.
func (g *Generator) logStep(stepID ProgressStepID) string {
	message, exists := ProgressMessages[stepID]
	if !exists {
		panic(fmt.Sprintf("Invalid step ID %s", stepID))
	}

	g.log.Debug(fmt.Sprintf("* * %s * *", message))

	return message
}

// Determine if a file has been created, modified, unchanged, or deleted
func (g *Generator) getFileStatus(absPath string, data []byte) ProgressFileStatus {
	if data == nil {
		return ProgressFileStatusDeleted
	}

	prevData, err := g.ReadFile(absPath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return ProgressFileStatusCreated
	case err != nil:
		panic(err)
	case bytes.Equal(prevData, data):
		return ProgressFileStatusUnchanged
	default:
		return ProgressFileStatusModified
	}
}

// Send a file status update.
func (g *Generator) sendFileProgressUpdate(path string, data []byte) {
	if !g.canUpdateProgress() {
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	fileStatus := g.getFileStatus(absPath, data)

	// ATM file content is only passed for the main README
	isMainReadme := false
	var content *bytes.Buffer
	if filepath.Base(path) == "README.md" && filepath.Dir(path) == g.outDir {
		isMainReadme = true
		content = bytes.NewBuffer(data)
	}

	if isMainReadme || (g.progress.UpdateFileStatus && fileStatus != ProgressFileStatusUnchanged) {
		g.progress.OnProgressUpdate(ProgressUpdate{
			TargetID: g.progress.TargetID,
			File: &ProgressFile{
				Status:       fileStatus,
				Path:         absPath,
				IsMainReadme: isMainReadme,
				Content:      content,
			},
		})
	}
}

// sendPatchesProgressUpdate handles progress updates from the patches subsystem.
// It logs push warnings to the generator's logger (visible in CLI-only mode) and
// emits progress updates for workflow visualization when streaming is enabled.
func (g *Generator) sendPatchesProgressUpdate(stepID string, merge *patches.MergeProgress) {
	// Always log push warnings to the generator's logger
	// (in CLI-only mode, progress updates aren't streamed but warnings should still appear)
	if merge != nil && merge.PushError != nil {
		g.log.Warn(fmt.Sprintf("Failed to push generation snapshot to remote: %v", merge.PushError))
		g.log.Warn("Your code has been generated successfully locally. The remote snapshot will be pushed on the next generation run.")
	}

	// Emit progress updates if streaming is enabled
	if g.canUpdateProgress() {
		// Emit step progress updates
		if stepID != "" {
			g.progress.OnProgressUpdate(ProgressUpdate{
				TargetID: g.progress.TargetID,
				Step: &ProgressStep{
					ID:      ProgressStepID(stepID),
					Message: ProgressMessages[ProgressStepID(stepID)],
				},
			})
		}
		// Emit merge-specific events (conflicts, push errors)
		if merge != nil {
			g.progress.OnProgressUpdate(ProgressUpdate{
				TargetID: g.progress.TargetID,
				Merge:    merge,
			})
		}
	}
}
