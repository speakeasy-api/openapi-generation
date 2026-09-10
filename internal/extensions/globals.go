package extensions

import (
	"context"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/marshaller"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/openapi/core"
	"github.com/speakeasy-api/openapi/validation"
)

func init() {
	marshaller.RegisterType(func() *Globals {
		return &Globals{}
	})
	marshaller.RegisterType(func() *CoreGlobals {
		return &CoreGlobals{}
	})
}

type Globals struct {
	marshaller.Model[CoreGlobals]
	Parameters []*openapi.ReferencedParameter
}

type CoreGlobals struct {
	marshaller.CoreModel `model:"globals"`

	Parameters marshaller.Node[[]marshaller.Node[*core.Reference[*core.Parameter]]] `key:"parameters"`
}

func (e *Extensions) HandleGlobalsExtension(ctx context.Context, doc *openapi.OpenAPI) (*Globals, error) {
	if doc.GetExtensions().Len() == 0 {
		return nil, nil
	}

	var globals Globals

	vErrs, err := extensions.UnmarshalExtensionModel[Globals, CoreGlobals](ctx, doc.GetExtensions(), e.GetResolvedName(ExtGlobals), &globals)
	if err != nil {
		if errors.Is(err, extensions.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if len(vErrs) > 0 {
		var vErr validation.Error
		if !errors.As(vErrs[0], &vErr) {
			vErr = validation.Error{
				UnderlyingError: vErrs[0],
			}
		}

		return nil, errors.NewValidationError("failed to validate "+ExtGlobals.Name(), vErr.GetNode(), errors.Join(vErrs...))
	}

	return &globals, nil
}

func (e *Extensions) IsGlobalHidden(ctx context.Context, parameter *openapi.Parameter) bool {
	if parameter == nil {
		return false
	}

	hidden, err := getExtensionValue(e.GetResolvedName(ExtGlobalsHidden), parameter.GetExtensions(), false)
	if err != nil {
		return false
	}
	return hidden
}
