package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ ephemeral.EphemeralResource = (*CustomEphemeralResource)(nil)
var _ ephemeral.EphemeralResourceWithConfigure = (*CustomEphemeralResource)(nil)

func NewCustomEphemeralResource() ephemeral.EphemeralResource {
	return &CustomEphemeralResource{}
}

// CustomEphemeralResource defines a stub custom ephemeral resource type implementation.
// This implementation is intentionally empty as this is used for verifying
// the inclusion of a custom ephemeral resource in the generated provider.
type CustomEphemeralResource struct {
	// Provider configured SDK client.
	client *sdk.SDK
}

type CustomEphemeralResourceModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *CustomEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom"
}

func (r *CustomEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *CustomEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*TestingProviderConfigureData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *TestingProviderConfigureData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = providerData.SDKClient
}

func (r *CustomEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	resp.Diagnostics.AddError(
		"Intentionally Missing Open Implementation",
		"This custom Ephemeral resource type intentionally has no Open implementation.",
	)
}
