package document

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/mode"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/references"
)

type Validator interface {
	FilterErrors(errs []error) (warns []error, errors []error)
}

type DocumentInfo struct {
	Doc        *openapi.OpenAPI
	Schema     []byte
	SchemaPath string
	IsRemote   bool
	Validator  Validator
}

func (d DocumentInfo) GetResolutionOptions(ctx context.Context) references.ResolveOptions {
	return references.ResolveOptions{
		TargetLocation:      d.SchemaPath,
		RootDocument:        d.Doc,
		DisableExternalRefs: mode.IsSpeakeasyExecutionContextEmbedded(ctx),
	}
}
