package logging

import "context"

func WithStepCtx(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	step := From(ctx).StartStep(name)

	tracker := &stepTracker{allSkipped: true}
	trackedCtx := withStepTracker(ctx, tracker)

	scopedLogger := From(ctx).Scope(name)
	trackedCtx = With(trackedCtx, scopedLogger)

	err := fn(trackedCtx)

	if err == nil && tracker.hasSubsteps && tracker.allSkipped {
		step.Skip()
		return nil
	}

	if err != nil {
		step.Fail()
	} else {
		step.Succeed()
	}
	return err
}

func StartStepCtx(ctx context.Context, name string) Step {
	step := From(ctx).StartStep(name)

	if tracker := stepTrackerFrom(ctx); tracker != nil {
		return &trackedStep{
			Step:    step,
			tracker: tracker,
		}
	}

	return step
}
