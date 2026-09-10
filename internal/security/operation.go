package security

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type HandlePerOpSecurityOpts struct {
	Opts
	ContextStack          ast.ContextStack
	SecurityReqs          []*openapi.SecurityRequirement
	SecuritySchemes       *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]
	Optional              ast.SecurityOptionalityReason
	GlobalSecurityMatcher GlobalSecurityMatcher
	Scope                 ast.Scope
	DocInfo               *document.DocumentInfo
}

func HandlePerOpSecurity(ctx context.Context, opts HandlePerOpSecurityOpts) (*ast.Security, error) {
	return determineSecurity(ctx, determineSecurityOpts{
		Opts:                  opts.Opts,
		contextStack:          opts.ContextStack,
		securityReqs:          opts.SecurityReqs,
		securitySchemes:       opts.SecuritySchemes,
		optional:              opts.Optional,
		globalSecurityMatcher: opts.GlobalSecurityMatcher,
		scope:                 opts.Scope,
		global:                false,
		docInfo:               opts.DocInfo,
	})
}
