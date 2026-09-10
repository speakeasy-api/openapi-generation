package generate

import (
	"bytes"
	"context"

	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"

	"github.com/speakeasy-api/openapi-generation/v2/internal/analytics"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

type validationResult struct {
	RunGenerator bool
	Result       *validation.Result
}

func (g *Generator) LoadAndValidateDoc(ctx context.Context, docInfo *document.DocumentInfo, ad *analytics.Data, workingDir string, validationOnly bool) (*validationResult, error) {
	doc, validationErrs, err := g.loadOpenAPIDoc(ctx, docInfo)
	if err != nil {
		return nil, err
	}

	// Store parsed document in docInfo for reuse during validation (eliminates duplicate parsing)
	docInfo.Doc = doc

	if validationOnly {
		g.validationOnly = true
	}

	res, runGenerator, err := g.validateDoc(ctx, docInfo, workingDir, validationErrs)
	if err != nil {
		return nil, err
	}

	version := doc.OpenAPI
	g.openAPIVersion = version
	ad.OpenAPIVersion = version

	ad.DocTitle = doc.Info.Title
	if doc.Info.Contact != nil {
		ad.SupportEmail = pointer.Value(doc.Info.Contact.Email)
		ad.SupportURL = pointer.Value(doc.Info.Contact.URL)
	}

	return &validationResult{
		RunGenerator: runGenerator,
		Result:       res,
	}, nil
}

func (g *Generator) loadOpenAPIDoc(ctx context.Context, docInfo *document.DocumentInfo) (*openapi.OpenAPI, []error, error) {
	ctx, span := g.tracer.Start(ctx, "Generator.loadOpenAPIDoc")
	defer span.End()

	doc, validationErrs, err := openapi.Unmarshal(ctx, bytes.NewReader(docInfo.Schema))
	if err != nil {
		return nil, nil, errors.NewValidationError("failed to load document", nil, err)
	}

	return doc, validationErrs, nil
}

func (g *Generator) validateDoc(ctx context.Context, docInfo *document.DocumentInfo, workingDir string, parseErrors []error) (*validation.Result, bool, error) {
	ctx, span := g.tracer.Start(ctx, "Generator.validateDoc")
	defer span.End()

	runGenerator := true

	opts := []validation.Option{
		validation.WithCliVersion(g.cliVersion),
	}
	if workingDir != "" {
		opts = append(opts, validation.WithWorkingDir(workingDir))
	}
	if g.validationRuleset != "" {
		opts = append(opts, validation.WithRuleset(g.validationRuleset))
	}
	defaultRuleset := validation.RulesetSpeakeasyGeneration
	if g.validationOnly {
		defaultRuleset = validation.RulesetSpeakeasyRecommended
		opts = append(opts, validation.WithUniformSeverity())
	}

	if g.parseValidOperations {
		opts = append(opts, validation.WithParseValidOperations())
	}

	v, err := validation.NewValidator(&g.subsystem.Config.Configuration, defaultRuleset, opts...)
	if err != nil {
		return nil, false, err
	}
	docInfo.Validator = v

	validationTarget := g.target
	// Use ValidateDocument to reuse the already-parsed document from docInfo
	// This eliminates duplicate parsing and improves performance by 40-50%
	// Pass parse errors to be filtered through match filters
	res := v.ValidateDocument(ctx, docInfo, validationTarget, parseErrors...)
	if res == nil {
		return nil, false, errors.New("validation result is nil")
	}
	if !v.AreRulesUsed(v.GetRulesets()[validation.RulesetSpeakeasyGeneration]) {
		if !g.validationOnly {
			return nil, false, errors.New("generation must validate against the generation ruleset")
		}
		runGenerator = false
	}

	// If debug infer discriminators is enabled, always run the generator
	if env.DebugInferDiscriminators() {
		runGenerator = true
	}

	return res, runGenerator, nil
}
