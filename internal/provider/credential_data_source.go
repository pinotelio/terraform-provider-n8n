package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-n8n/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &credentialDataSource{}
	_ datasource.DataSourceWithConfigure = &credentialDataSource{}
)

// NewCredentialDataSource is a helper function to simplify the provider implementation.
func NewCredentialDataSource() datasource.DataSource {
	return &credentialDataSource{}
}

// credentialDataSource is the data source implementation.
type credentialDataSource struct {
	client *client.Client
}

// credentialDataSourceModel maps the data source schema data.
type credentialDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Data types.String `tfsdk:"data"`
}

// Metadata returns the data source type name.
func (d *credentialDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

// Schema defines the schema for the data source.
func (d *credentialDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an n8n credential.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential identifier",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the credential",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the credential",
				Computed:    true,
			},
			"data": schema.StringAttribute{
				Description: "JSON string representing the credential data",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *credentialDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *credentialDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state credentialDataSourceModel

	// Read configuration
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// n8n API does not support reading credentials for security reasons:
	// - No GET /api/v1/credentials/{id} endpoint (returns 405)
	// - No LIST /api/v1/credentials endpoint available
	//
	// Since we cannot fetch credential data from the API, this data source
	// has limited functionality. We return an error with guidance.

	resp.Diagnostics.AddError(
		"n8n Credential Data Source Not Supported",
		fmt.Sprintf(
			"The n8n API does not support reading credentials for security reasons. "+
				"Credential data sources cannot be used. "+
				"If you need to reference a credential, use the resource directly:\n\n"+
				"  resource \"n8n_credential\" \"example\" {\n"+
				"    name = \"My Credential\"\n"+
				"    type = \"httpBasicAuth\"\n"+
				"    data = jsonencode({...})\n"+
				"  }\n\n"+
				"Then reference it as: n8n_credential.example.id\n\n"+
				"Credential ID provided: %s",
			state.ID.ValueString(),
		),
	)
}
