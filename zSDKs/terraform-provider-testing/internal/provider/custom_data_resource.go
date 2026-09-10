package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = (*CustomDataResource)(nil)

func NewCustomDataResource() datasource.DataSource {
	return &CustomDataResource{}
}

// CustomDataResource defines a stub custom managed resource type implementation.
// This implementation is intentionally empty as this is used for verifying
// the inclusion of a custom managed resource in the generated provider.
type CustomDataResource struct {
	// Provider configured SDK client.
	client *sdk.SDK
}

type CustomDataResourceModel struct {
	ID types.String `tfsdk:"id"`
}

type CustomDataResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *CustomDataResource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom"
}

func (r *CustomDataResource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *CustomDataResource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*TestingProviderConfigureData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *TestingProviderConfigureData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = providerData.SDKClient
}

func (r *CustomDataResource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Intentionally Missing Read Implementation",
		"This custom data resource type intentionally has no Read implementation.",
	)
}
