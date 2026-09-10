package generate

import (
	"context"
	"errors"
	"fmt"
)

// Thin wrapper around Generate that checks if the context was cancelled.
func (g *Generator) GenerateWithCancel(cancelCtx context.Context, schema []byte, schemaPath, target, outDir string, isRemote, compileOutput bool) (bool, []error) {
	// check if context has already been cancelled
	if err := cancelCtx.Err(); err != nil {
		return true, []error{g.cancelError()}
	}

	errs := g.Generate(cancelCtx, schema, schemaPath, target, outDir, isRemote, compileOutput)

	if cancelCtx.Err() != nil {
		// if cancellation occurred, we ignore generation errors
		return true, []error{g.cancelError()}
	}

	return false, errs
}

// Check if the context is cancelled or expired.
func isCancelled(ctx context.Context) bool {
	if ctx.Done() == nil {
		// context is not cancellable
		return false
	}

	return ctx.Err() != nil
}

// Format a context cancellation error based on current StepID, if applicable.
func (g *Generator) cancelError() error {
	g.logStep(ProgressStepCancel)
	msg := "Generation was cancelled"
	if g.progress.StepID != "" {
		msg += fmt.Sprintf(" during step '%s'", g.progress.StepID)
	}
	g.log.Warn(msg)
	return errors.New(msg)
}
