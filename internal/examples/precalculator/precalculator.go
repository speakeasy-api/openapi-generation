package precalculator

import (
	"context"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/js/template"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type ExamplePreCalculator struct {
	engine *template.Engine
}

func New(ctx context.Context, engine *template.Engine) (*ExamplePreCalculator, error) {
	if err := engine.RunScript(ctx, "examples.ts"); err != nil {
		return nil, err
	}

	return &ExamplePreCalculator{
		engine: engine,
	}, nil
}

func (e *ExamplePreCalculator) PreCalculateExamples(ctx context.Context, a *ast.AST, precalculatedExamples config.Examples) (config.Examples, error) {
	calculatedExamples, err := e.engine.RunFunction(ctx, "getPrecalculatedExamples", a.MainSDK, precalculatedExamples)
	if err != nil {
		return nil, err
	}

	ex, ok := calculatedExamples.Export().(config.Examples)
	if !ok {
		return nil, fmt.Errorf("invalid type returned from getPrecalculatedExamples: %T", calculatedExamples.Export())
	}
	return ex, nil
}
