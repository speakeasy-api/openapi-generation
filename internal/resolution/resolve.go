package resolution

import (
	"context"
	"reflect"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/references"
	"gopkg.in/yaml.v3"
)

type Resolvable[T any] interface {
	references.Resolvable[T]
	GetRootNodeLine() int
	GetRootNode() *yaml.Node
}

func Resolve[T any](ctx context.Context, resolvable Resolvable[T], docInfo *document.DocumentInfo) (*T, error) {
	if resolvable == nil || reflect.ValueOf(resolvable).IsNil() {
		return nil, nil
	}

	validationErrs, err := resolvable.Resolve(ctx, docInfo.GetResolutionOptions(ctx))
	if err != nil {
		return nil, errors.NewResolveError(err, resolvable.GetRootNode())
	}
	if len(validationErrs) > 0 {

		vWarns, vErrs := docInfo.Validator.FilterErrors(validationErrs)
		for _, warn := range vWarns {
			logging.LogWarning(ctx, "validation warning resolving reference", warn)
		}

		if len(vErrs) > 0 {
			errs := append([]error{}, vErrs...)

			return nil, errors.NewValidationError("reference validation errors "+resolvable.GetReference().String(), resolvable.GetRootNode(), errors.Join(errs...))
		}
	}

	resolved := resolvable.GetResolvedObject()
	if resolved == nil {
		return nil, errors.NewResolveError(errors.New("unresolved reference"), resolvable.GetRootNode())
	}

	return resolved, nil
}
