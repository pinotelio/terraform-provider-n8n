package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-n8n/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &workflowDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowDataSource{}
)

// NewWorkflowDataSource is a helper function to simplify the provider implementation.
func NewWorkflowDataSource() datasource.DataSource {
	return &workflowDataSource{}
}

// workflowDataSource is the data source implementation.
type workflowDataSource struct {
	client *client.Client
}

// workflowDataSourceModel maps the data source schema data.
type workflowDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Nodes       types.String `tfsdk:"nodes"`
	Connections types.String `tfsdk:"connections"`
	Settings    types.String `tfsdk:"settings"`
	Tags        types.String `tfsdk:"tags"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	Active      types.Bool   `tfsdk:"active"`
}

// Metadata returns the data source type name.
func (d *workflowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

// Schema defines the schema for the data source.
func (d *workflowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an n8n workflow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Workflow identifier",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the workflow",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the workflow is active",
				Computed:    true,
			},
			"nodes": schema.StringAttribute{
				Description: "JSON string representing the workflow nodes",
				Computed:    true,
			},
			"connections": schema.StringAttribute{
				Description: "JSON string representing the workflow connections",
				Computed:    true,
			},
			"settings": schema.StringAttribute{
				Description: "JSON string representing the workflow settings",
				Computed:    true,
			},
			"tags": schema.StringAttribute{
				Description: "JSON string representing the workflow tags",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the workflow was created",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the workflow was last updated",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *workflowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *workflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workflowDataSourceModel

	// Read configuration
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get workflow from n8n
	workflow, err := d.client.GetWorkflow(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading n8n Workflow",
			"Could not read n8n workflow ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response to state
	state.Name = types.StringValue(workflow.Name)
	state.Active = types.BoolValue(workflow.Active)
	state.CreatedAt = types.StringValue(workflow.CreatedAt)
	state.UpdatedAt = types.StringValue(workflow.UpdatedAt)

	// Convert nodes to JSON string
	nodesJSON, err := json.Marshal(workflow.Nodes)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling nodes",
			"Could not marshal nodes to JSON: "+err.Error(),
		)
		return
	}
	state.Nodes = types.StringValue(string(nodesJSON))

	// Convert connections to JSON string
	connectionsJSON, err := json.Marshal(workflow.Connections)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshaling connections",
			"Could not marshal connections to JSON: "+err.Error(),
		)
		return
	}
	state.Connections = types.StringValue(string(connectionsJSON))

	// Convert settings to JSON string
	if workflow.Settings != nil {
		settingsJSON, err := json.Marshal(workflow.Settings)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling settings",
				"Could not marshal settings to JSON: "+err.Error(),
			)
			return
		}
		state.Settings = types.StringValue(string(settingsJSON))
	}

	// Convert tags to JSON string
	if workflow.Tags != nil {
		tagsJSON, err := json.Marshal(workflow.Tags)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling tags",
				"Could not marshal tags to JSON: "+err.Error(),
			)
			return
		}
		state.Tags = types.StringValue(string(tagsJSON))
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
