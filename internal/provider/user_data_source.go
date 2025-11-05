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
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

// NewUserDataSource is a helper function to simplify the provider implementation.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSource is the data source implementation.
type userDataSource struct {
	client *client.Client
}

// userDataSourceModel maps the data source schema data.
type userDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	IsOwner   types.Bool   `tfsdk:"is_owner"`
	IsPending types.Bool   `tfsdk:"is_pending"`
}

// Metadata returns the data source type name.
func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the data source.
func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an n8n user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User identifier",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "Role of the user",
				Computed:    true,
			},
			"is_owner": schema.BoolAttribute{
				Description: "Whether the user is an owner",
				Computed:    true,
			},
			"is_pending": schema.BoolAttribute{
				Description: "Whether the user account is pending activation",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the user was created",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the user was last updated",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userDataSourceModel

	// Read configuration
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from n8n
	user, err := d.client.GetUser(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading n8n User",
			"Could not read n8n user ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to state
	state.Email = types.StringValue(user.Email)
	state.Role = types.StringValue(user.GetRole())
	state.IsOwner = types.BoolValue(user.IsOwner)
	state.IsPending = types.BoolValue(user.IsPending)
	state.CreatedAt = types.StringValue(user.CreatedAt)
	state.UpdatedAt = types.StringValue(user.UpdatedAt)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
