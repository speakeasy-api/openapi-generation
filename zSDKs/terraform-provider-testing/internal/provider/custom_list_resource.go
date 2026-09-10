package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-testing/internal/sdk"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ list.ListResource = (*CustomListResource)(nil)
var _ list.ListResourceWithConfigure = (*CustomListResource)(nil)

func NewCustomListResource() list.ListResource {
	return &CustomListResource{}
}

// CustomListResource defines a stub custom list resource type implementation.
// This implementation is intentionally empty as this is used for verifying
// the inclusion of a custom list resource in the generated provider.
type CustomListResource struct {
	// Provider configured SDK client.
	client *sdk.SDK
}

type CustomListResourceModel struct{}

func (r *CustomListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom"
}

func (r *CustomListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (r *CustomListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*TestingProviderConfigureData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected List Configure Type",
			fmt.Sprintf("Expected *TestingProviderConfigureData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = providerData.SDKClient
}

func (r *CustomListResource) List(_ context.Context, _ list.ListRequest, stream *list.ListResultsStream) {
	diags := diag.Diagnostics{}
	diags.AddError(
		"Intentionally Missing List Implementation",
		"This custom list resource type intentionally has no List implementation.",
	)

	stream.Results = list.ListResultsStreamDiagnostics(diags)
}
