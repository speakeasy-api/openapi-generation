package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ action.Action = (*CustomAction)(nil)
var _ action.ActionWithConfigure = (*CustomAction)(nil)

func NewCustomAction() action.Action {
	return &CustomAction{}
}

// CustomAction defines a stub custom action type implementation.
// This implementation is intentionally empty as this is used for verifying
// the inclusion of a custom action in the generated provider.
type CustomAction struct {
	// Provider configured SDK client.
	client *sdk.SDK
}

type CustomActionModel struct{}

func (a *CustomAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom"
}

func (a *CustomAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (a *CustomAction) Configure(ctx context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*TestingProviderConfigureData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *TestingProviderConfigureData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	a.client = providerData.SDKClient
}

func (a *CustomAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	resp.Diagnostics.AddError("Intentionally missing action implementation", "This custom action implementation intentionally has no Invoke implementation.")
}
