package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = (*CustomResource)(nil)
var _ resource.ResourceWithIdentity = (*CustomResource)(nil)

func NewCustomResource() resource.Resource {
	return &CustomResource{}
}

// CustomResource defines a stub custom managed resource type implementation.
// This implementation is intentionally empty as this is used for verifying
// the inclusion of a custom managed resource in the generated provider.
type CustomResource struct {
	// Provider configured SDK client.
	client *sdk.SDK
}

type CustomResourceModel struct {
	ID types.String `tfsdk:"id"`
}

type CustomResourceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

func (r *CustomResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom"
}

func (r *CustomResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *CustomResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *CustomResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CustomResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Intentionally Missing Create Implementation",
		"This custom managed resource type intentionally has no Create implementation.",
	)
}

func (r *CustomResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.AddError(
		"Intentionally Missing Read Implementation",
		"This custom managed resource type intentionally has no Read implementation.",
	)
}

func (r *CustomResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Intentionally Missing Update Implementation",
		"This custom managed resource type intentionally has no Update implementation.",
	)
}

func (r *CustomResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Intentionally Missing Delete Implementation",
		"This custom managed resource type intentionally has no Delete implementation.",
	)
}
