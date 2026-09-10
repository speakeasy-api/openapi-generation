package template

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/js/template"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	template.Config
	TemplateConfig  map[string]any
	AST             *ast.AST
	EnabledFeatures map[string]bool
}

type GlobalContext struct {
	Config                  map[string]any
	AST                     *ast.AST
	EnabledTemplateFeatures map[string]bool
}

type Template struct {
	cfg    Config
	target types.Target

	vmPool *sync.Pool
}

func New(ctx context.Context, target types.Target, cfg Config) *Template {
	vmPool := sync.Pool{
		New: createEngine(ctx, target, cfg),
	}

	return &Template{
		cfg:    cfg,
		target: target,
		vmPool: &vmPool,
	}
}

func (t *Template) Execute(ctx context.Context) error {
	jobs, err := t.getJobs(ctx)
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil
	}

	jobs = randomizeJobs(jobs)
	jobs = prioritizeReadmeJob(jobs)

	start := time.Now()

	maxConcurrency := env.DebugConcurrency()
	if t.cfg.DebugGojaPort > 0 {
		// When the goja debugger is active, always force a single engine so
		// breakpoints work across getJobs/runJobs/postJobs and we don't try
		// to bind the DAP port more than once.
		maxConcurrency = 1
	} else if maxConcurrency == 0 {
		// assuming all CPUs we are targeting are hyperthreaded use 3/4 the number of threads
		maxConcurrency = int(float64(runtime.NumCPU()) * 0.75)
	}

	// split jobs into chunks of maxConcurrency
	chunkSize := len(jobs) / maxConcurrency

	// Handle job count less than maxConcurrency and prevent slices.Chunk panic
	if chunkSize == 0 {
		chunkSize = len(jobs)
	}

	chunks := slices.Collect(slices.Chunk(jobs, chunkSize))

	logging.From(ctx).Debug(fmt.Sprintf("Running %d jobs with %d workers", len(jobs), len(chunks)))

	errs, ctx2 := errgroup.WithContext(ctx)

	for _, chunk := range chunks {
		errs.Go(func(chunk []job) func() error {
			return func() error {
				start := time.Now()
				if err := t.runJobs(ctx2, chunk); err != nil {
					return err
				}
				logging.From(ctx).Debug(fmt.Sprintf("Run worker completed in '%s' jobs %d", time.Since(start), len(chunk)))
				return nil
			}
		}(chunk))
	}

	err = errs.Wait()
	logging.From(ctx).Debug(fmt.Sprintf("Run all jobs completed in '%s' total jobs %d total workers %d", time.Since(start), len(jobs), len(chunks)))

	if err != nil {
		return err
	}

	if err := t.postJobs(ctx); err != nil {
		if strings.Contains(err.Error(), "failed to find function") {
			logging.From(ctx).Debug("postJobs function not found in current context, skipping")
		} else {
			return err
		}
	}
	return nil
}

type job struct {
	ID       string
	FileName string
	Context  any
}

func (t *Template) getJobs(ctx context.Context) ([]job, error) {
	start := time.Now()

	res := t.vmPool.Get()
	if err, ok := res.(error); ok {
		return nil, err
	}

	e := res.(*template.Engine)
	defer t.vmPool.Put(e) // TODO cleanup?

	jobsVal, err := e.RunFunction(ctx, "getJobs", nil)
	if err != nil {
		return nil, err
	}

	jobsInt := jobsVal.Export().([]any)

	jobs := make([]job, len(jobsInt))

	for i, jobInt := range jobsInt {
		j := jobInt.(map[string]any)

		jobs[i] = job{
			ID:       j["ID"].(string),
			FileName: j["FileName"].(string),
			Context:  j["Context"],
		}
	}

	logging.From(ctx).Debug(fmt.Sprintf("getJobs completed in '%s'", time.Since(start)))

	return jobs, nil
}

func (t *Template) runJobs(ctx context.Context, jobs []job) error {
	var e *template.Engine

	if t.cfg.DebugGojaPort > 0 {
		// When the goja debugger is active, reuse the pool engine so
		// breakpoints in the single debugger session cover runJobs too.
		res := t.vmPool.Get()
		if err, ok := res.(error); ok {
			return err
		}
		e = res.(*template.Engine)
		defer t.vmPool.Put(e)
	} else {
		res := createEngine(ctx, t.target, t.cfg)()
		if err, ok := res.(error); ok {
			return err
		}
		e = res.(*template.Engine)
	}

	if _, err := e.RunFunction(ctx, "runJobs", jobs); err != nil {
		return err
	}

	return nil
}

func (t *Template) postJobs(ctx context.Context) error {
	start := time.Now()

	res := t.vmPool.Get()
	if err, ok := res.(error); ok {
		return err
	}

	e := res.(*template.Engine)
	// put the engine back in the pool when this function returns
	defer t.vmPool.Put(e)

	if _, err := e.RunFunction(ctx, "postJobs", ctx); err != nil {
		return err
	}

	logging.From(ctx).Debug(fmt.Sprintf("postJobs action completed in '%s'", time.Since(start)))
	return nil
}

// Shuffle jobs
func randomizeJobs(jobs []job) []job {
	// Create a new random source seeded with current time
	src := rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()>>32))
	rng := rand.New(src)

	// Then use this rng instead of the global rand
	rng.Shuffle(len(jobs), func(i, j int) {
		jobs[i], jobs[j] = jobs[j], jobs[i]
	})

	return jobs
}

// Move readme job to the front of the jobs slice, if present
func prioritizeReadmeJob(jobs []job) []job {
	var ret []job
	for _, j := range jobs {
		if j.ID == "readme" {
			ret = append([]job{j}, ret...)
		} else {
			ret = append(ret, j)
		}
	}

	return ret
}

func createEngine(ctx context.Context, target types.Target, cfg Config) func() any {
	return func() any {
		e := template.New(ctx, target, cfg.Config)

		if err := e.Init(ctx, GlobalContext{
			Config:                  cfg.TemplateConfig,
			AST:                     cfg.AST,
			EnabledTemplateFeatures: cfg.EnabledFeatures,
		}); err != nil {
			return err
		}

		if err := e.RunScript(ctx, "main.ts"); err != nil {
			return parseScriptError(err)
		}

		return e
	}
}

var validationErrorRegex = regexp.MustCompile(`(?s).*?ValidationError: (.*?)\n\tat .*`)

func parseScriptError(err error) error {
	// A bit crap we need to parse the error message but its what we have for now
	if strings.Contains(err.Error(), "ValidationError: ") {
		matches := validationErrorRegex.FindAllStringSubmatch(err.Error(), -1)
		if len(matches) > 0 {
			return errors.NewValidationError(matches[0][1], nil, nil)
		}
	}

	return errors.ErrGeneration.Wrap(err)
}
