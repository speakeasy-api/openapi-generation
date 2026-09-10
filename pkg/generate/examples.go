package generate

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/examples/precalculator"
	"github.com/speakeasy-api/openapi/openapi"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

// examplesVersion is the version of the examples that are precalculated. This should be updated whenever a change to the generated examples code is made to ensure customers get the fixes
// NOTE: this will cause README and DOCs churn when changed though, the amount of churn depends on the changes between generations in their spec and how reliant they are on auto generated examples
const examplesVersion = "1.0.2"

func (g *Generator) preCalculateExamples(ctx context.Context, a *ast.AST) error {
	ctx, span := g.tracer.Start(ctx, "Generator.preCalculateExamples")
	defer span.End()

	var preCalculatedExamples config.Examples

	// If the examplesVersion matches use the stored examples otherwise regenerate
	if g.lockFile.ExamplesVersion == examplesVersion {
		preCalculatedExamples = g.lockFile.Examples
	}

	if a == nil {
		return nil
	}

	g.log.Debug("starting preCalculateExamples")

	start := time.Now()

	// TODO don't re-calculate examples for the mock server
	if g.target.Target != "mockserver" {
		e := g.getExecutor(ctx, g.target, g.onWriteFile(""), g.onReadFile(""))
		if err := e.Init(ctx, GlobalContext{
			Config: g.getBaseTemplateConfigs(g.target, g.subsystem.Config, g.lockFile),
			AST:    a,
		}); err != nil {
			return err
		}

		preCalculator, err := precalculator.New(ctx, e)
		if err != nil {
			return err
		}
		preCalculatedExamples, err = preCalculator.PreCalculateExamples(ctx, a, preCalculatedExamples)
		if err != nil {
			return err
		}
	}

	if err := g.populateSDKExamples(ctx, a.MainSDK, preCalculatedExamples); err != nil {
		return err
	}

	elapsed := time.Since(start)
	g.log.Debug(fmt.Sprintf("preCalculateExamples completed in '%s'", elapsed))

	g.lockFile.Examples = preCalculatedExamples
	g.lockFile.ExamplesVersion = examplesVersion

	return nil
}

func (g *Generator) populateSDKExamples(ctx context.Context, sdk *ast.SDK, preCalculatedExamples config.Examples) error {
	if sdk == nil {
		return nil
	}

	for _, op := range sdk.Operations {
		if err := g.populateOperationExamples(ctx, op, preCalculatedExamples); err != nil {
			return err
		}
	}

	for _, sub := range sdk.SubSDKs {
		if err := g.populateSDKExamples(ctx, sub, preCalculatedExamples); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) populateOperationExamples(_ context.Context, op *ast.Operation, preCalculateExamples config.Examples) error {
	if op == nil || preCalculateExamples == nil {
		return nil
	}

	opExamples, ok := preCalculateExamples.Get(op.OriginalID)
	if !ok {
		return nil
	}

	for exampleName, opExample := range opExamples.All() {
		if opExample.RequestBody != nil && op.Request != nil && op.Request.RequestBody != nil {
			examplePair := opExample.RequestBody.First() // TODO at some point we are likely going to need to match examples with content types instead of just using the first one
			if examplePair != nil {
				preCalculatedExample := ast.NewExample(exampleName, "", &examplePair.Value)
				op.Request.Examples = updateExamples(op.Request.Examples, preCalculatedExample)
			}
		}

		if opExample.Parameters != nil && op.Request != nil && op.Request.Params != nil {
			if opExample.Parameters.Query.Len() > 0 && len(op.Request.Params.QueryParams) > 0 {
				for paramName, paramExample := range opExample.Parameters.Query.All() {
					for i, param := range op.Request.Params.QueryParams {
						if paramName != param.Field.OriginalName {
							continue
						}

						preCalculatedExample := ast.NewExample(exampleName, "", &paramExample)

						op.Request.Params.QueryParams[i].Examples = updateExamples(param.Examples, preCalculatedExample)
					}
				}
			}

			if opExample.Parameters.Path.Len() > 0 && len(op.Request.Params.PathParams) > 0 {
				for paramName, paramExample := range opExample.Parameters.Path.All() {
					for i, param := range op.Request.Params.PathParams {
						if paramName != param.Field.OriginalName {
							continue
						}

						preCalculatedExample := ast.NewExample(exampleName, "", &paramExample)

						op.Request.Params.PathParams[i].Examples = updateExamples(param.Examples, preCalculatedExample)
					}
				}
			}

			if opExample.Parameters.Header.Len() > 0 && len(op.Request.Params.HeaderParams) > 0 {
				for paramName, paramExample := range opExample.Parameters.Header.All() {
					for i, param := range op.Request.Params.HeaderParams {
						if paramName != param.Field.OriginalName {
							continue
						}

						preCalculatedExample := ast.NewExample(exampleName, "", &paramExample)

						op.Request.Params.HeaderParams[i].Examples = updateExamples(param.Examples, preCalculatedExample)
					}
				}
			}
		}

		if opExample.Responses.Len() > 0 {
			for code, contentExamples := range opExample.Responses.All() {
				for contentType, contentExample := range contentExamples.All() {
					for i, resp := range op.Response.Responses {
						if !slices.Contains(resp.Code, code) {
							continue
						}

						for j, content := range resp.Content {
							if content.ContentType != contentType {
								continue
							}

							preCalculatedExample := ast.NewExample(exampleName, "", &contentExample)

							resp.Content[j].Examples = updateExamples(content.Examples, preCalculatedExample)
							op.Response.Responses[i] = resp
						}
					}
				}
			}
		}
	}

	return nil
}

// resolveExampleValue returns the best available value from an OpenAPI Example object.
// OpenAPI 3.2 priority: dataValue > value > serializedValue (best-effort parse).
func resolveExampleValue(example *openapi.Example) *yaml.Node {
	if v := example.GetDataValue(); v != nil {
		return v
	}
	if v := example.GetValue(); v != nil {
		return v
	}
	if s := example.GetSerializedValue(); s != "" {
		// serializedValue is a pre-serialized string (could be JSON, form-encoded, etc.)
		// Try to parse it as YAML/JSON — if it's valid, use the structured form.
		// If not parseable, ignore it (non-JSON serialized forms aren't useful for SDK codegen).
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(s), &node); err == nil && len(node.Content) > 0 {
			return node.Content[0]
		}
	}
	return nil
}

func updateExamples(examples ast.Examples, example *ast.Example) ast.Examples {
	if len(examples) == 0 {
		examples = ast.Examples{}
	}

	for i, ex := range examples {
		if ex.Name() == example.Name() {
			examples[i] = example
			return examples
		}
	}
	examples = append(examples, example)
	return examples
}
