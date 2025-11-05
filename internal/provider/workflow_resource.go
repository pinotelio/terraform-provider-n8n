package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-n8n/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &workflowResource{}
	_ resource.ResourceWithConfigure   = &workflowResource{}
	_ resource.ResourceWithImportState = &workflowResource{}
)

// NewWorkflowResource is a helper function to simplify the provider implementation.
func NewWorkflowResource() resource.Resource {
	return &workflowResource{}
}

// workflowResource is the resource implementation.
type workflowResource struct {
	client *client.Client
}

// workflowResourceModel maps the resource schema data.
type workflowResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	WorkflowJSON types.String `tfsdk:"workflow_json"`
	Nodes        types.String `tfsdk:"nodes"`
	Connections  types.String `tfsdk:"connections"`
	Settings     types.String `tfsdk:"settings"`
	Tags         types.String `tfsdk:"tags"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *workflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

// Schema defines the schema for the resource.
func (r *workflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an n8n workflow. You can either specify individual attributes (name, nodes, connections, etc.) or provide a complete workflow JSON using the workflow_json attribute.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Workflow identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the workflow. Optional if workflow_json is provided.",
				Optional:    true,
				Computed:    true,
			},
			"nodes": schema.StringAttribute{
				Description: "JSON string representing the workflow nodes. Optional if workflow_json is provided.",
				Optional:    true,
				Computed:    true,
			},
			"connections": schema.StringAttribute{
				Description: "JSON string representing the workflow connections. Optional if workflow_json is provided.",
				Optional:    true,
				Computed:    true,
			},
			"settings": schema.StringAttribute{
				Description: "JSON string representing the workflow settings",
				Optional:    true,
				Computed:    true,
			},
			"tags": schema.StringAttribute{
				Description: "JSON string representing the workflow tags",
				Optional:    true,
				Computed:    true,
			},
			"workflow_json": schema.StringAttribute{
				Description: "Complete workflow JSON. When provided, individual attributes (name, nodes, connections, etc.) are extracted from this JSON. This allows you to paste an entire n8n workflow export directly.",
				Optional:    true,
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

// Configure adds the provider configured client to the resource.
func (r *workflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan workflowResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name string
	var active bool
	var nodes []interface{}
	var connections map[string]interface{}
	var settings map[string]interface{}
	var tags []map[string]string

	// Check if workflow_json is provided
	if !plan.WorkflowJSON.IsNull() && plan.WorkflowJSON.ValueString() != "" {
		// Parse the complete workflow JSON
		var workflowData map[string]interface{}
		if err := json.Unmarshal([]byte(plan.WorkflowJSON.ValueString()), &workflowData); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing workflow_json",
				"Could not parse workflow_json: "+err.Error(),
			)
			return
		}

		// Extract name
		if nameVal, ok := workflowData["name"].(string); ok {
			name = nameVal
		} else {
			resp.Diagnostics.AddError(
				"Missing required field",
				"workflow_json must contain a 'name' field",
			)
			return
		}

		// Extract active (default to false if not present)
		if activeVal, ok := workflowData["active"].(bool); ok {
			active = activeVal
		} else {
			active = false
		}

		// Extract nodes
		if nodesVal, ok := workflowData["nodes"].([]interface{}); ok {
			nodes = nodesVal
		} else {
			resp.Diagnostics.AddError(
				"Missing required field",
				"workflow_json must contain a 'nodes' array",
			)
			return
		}

		// Extract connections
		if connectionsVal, ok := workflowData["connections"].(map[string]interface{}); ok {
			connections = connectionsVal
		} else {
			resp.Diagnostics.AddError(
				"Missing required field",
				"workflow_json must contain a 'connections' object",
			)
			return
		}

		// Extract settings (optional)
		if settingsVal, ok := workflowData["settings"].(map[string]interface{}); ok {
			settings = settingsVal
		}

		// Extract tags (optional)
		if tagsVal, ok := workflowData["tags"].([]interface{}); ok {
			tags = make([]map[string]string, 0, len(tagsVal))
			for _, tag := range tagsVal {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					tagStr := make(map[string]string)
					for k, v := range tagMap {
						if vStr, ok := v.(string); ok {
							tagStr[k] = vStr
						}
					}
					tags = append(tags, tagStr)
				}
			}
		}

		// Update plan with extracted values for state management
		plan.Name = types.StringValue(name)
		// plan.Active = types.BoolValue(active)

		nodesJSON, err := json.Marshal(nodes)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling nodes",
				"Could not marshal nodes to JSON: "+err.Error(),
			)
			return
		}
		plan.Nodes = types.StringValue(string(nodesJSON))

		connectionsJSON, err := json.Marshal(connections)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling connections",
				"Could not marshal connections to JSON: "+err.Error(),
			)
			return
		}
		plan.Connections = types.StringValue(string(connectionsJSON))

		if settings != nil {
			settingsJSON, err := json.Marshal(settings)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error marshaling settings",
					"Could not marshal settings to JSON: "+err.Error(),
				)
				return
			}
			plan.Settings = types.StringValue(string(settingsJSON))
		}

		if tags != nil {
			tagsJSON, err := json.Marshal(tags)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error marshaling tags",
					"Could not marshal tags to JSON: "+err.Error(),
				)
				return
			}
			plan.Tags = types.StringValue(string(tagsJSON))
		}
	} else {
		// Use individual attributes
		if plan.Name.IsNull() || plan.Nodes.IsNull() || plan.Connections.IsNull() {
			resp.Diagnostics.AddError(
				"Missing required attributes",
				"Either workflow_json or all of (name, nodes, connections) must be provided",
			)
			return
		}

		name = plan.Name.ValueString()
		// 		active = plan.Active.ValueBool()

		// Parse JSON strings
		if err := json.Unmarshal([]byte(plan.Nodes.ValueString()), &nodes); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing nodes JSON",
				"Could not parse nodes JSON: "+err.Error(),
			)
			return
		}

		if err := json.Unmarshal([]byte(plan.Connections.ValueString()), &connections); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing connections JSON",
				"Could not parse connections JSON: "+err.Error(),
			)
			return
		}

		if !plan.Settings.IsNull() && plan.Settings.ValueString() != "" {
			if err := json.Unmarshal([]byte(plan.Settings.ValueString()), &settings); err != nil {
				resp.Diagnostics.AddError(
					"Error parsing settings JSON",
					"Could not parse settings JSON: "+err.Error(),
				)
				return
			}
		}

		if !plan.Tags.IsNull() && plan.Tags.ValueString() != "" {
			if err := json.Unmarshal([]byte(plan.Tags.ValueString()), &tags); err != nil {
				resp.Diagnostics.AddError(
					"Error parsing tags JSON",
					"Could not parse tags JSON: "+err.Error(),
				)
				return
			}
		}
	}

	// Create new workflow
	workflow := &client.Workflow{
		Name:        name,
		Active:      active,
		Nodes:       nodes,
		Connections: connections,
		Settings:    settings,
		Tags:        tags,
	}

	createdWorkflow, err := r.client.CreateWorkflow(workflow)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating workflow",
			"Could not create workflow, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(createdWorkflow.ID)
	plan.CreatedAt = types.StringValue(createdWorkflow.CreatedAt)
	plan.UpdatedAt = types.StringValue(createdWorkflow.UpdatedAt)

	// Ensure tags is set (even if empty)
	if plan.Tags.IsNull() || plan.Tags.IsUnknown() {
		if len(createdWorkflow.Tags) > 0 {
			tagsJSON, err := json.Marshal(createdWorkflow.Tags)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error marshaling tags",
					"Could not marshal tags to JSON: "+err.Error(),
				)
				return
			}
			plan.Tags = types.StringValue(string(tagsJSON))
		} else {
			plan.Tags = types.StringValue("[]")
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state workflowResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed workflow value from n8n
	workflow, err := r.client.GetWorkflow(state.ID.ValueString())
	if err != nil {
		// Check if the workflow was deleted outside of Terraform (404 error)
		if strings.Contains(err.Error(), "404") {
			// Remove from state - Terraform will recreate it on next apply
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading n8n Workflow",
			"Could not read n8n workflow ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Name = types.StringValue(workflow.Name)
	// 	state.Active = types.BoolValue(workflow.Active)
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
	if len(workflow.Tags) > 0 {
		tagsJSON, err := json.Marshal(workflow.Tags)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling tags",
				"Could not marshal tags to JSON: "+err.Error(),
			)
			return
		}
		state.Tags = types.StringValue(string(tagsJSON))
	} else {
		// Set empty array for tags if none exist
		state.Tags = types.StringValue("[]")
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan workflowResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var name string
	var active bool
	var nodes []interface{}
	var connections map[string]interface{}
	var settings map[string]interface{}
	var tags []map[string]string

	// Check if workflow_json is provided
	if !plan.WorkflowJSON.IsNull() && plan.WorkflowJSON.ValueString() != "" {
		// Parse the complete workflow JSON
		var workflowData map[string]interface{}
		if err := json.Unmarshal([]byte(plan.WorkflowJSON.ValueString()), &workflowData); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing workflow_json",
				"Could not parse workflow_json: "+err.Error(),
			)
			return
		}

		// Extract name
		if nameVal, ok := workflowData["name"].(string); ok {
			name = nameVal
		} else {
			resp.Diagnostics.AddError(
				"Missing required field",
				"workflow_json must contain a 'name' field",
			)
			return
		}

		// Extract active (default to false if not present)
		if activeVal, ok := workflowData["active"].(bool); ok {
			active = activeVal
		} else {
			active = false
		}

		// Extract nodes
		if nodesVal, ok := workflowData["nodes"].([]interface{}); ok {
			nodes = nodesVal
		} else {
			resp.Diagnostics.AddError(
				"Missing required field",
				"workflow_json must contain a 'nodes' array",
			)
			return
		}

		// Extract connections
		if connectionsVal, ok := workflowData["connections"].(map[string]interface{}); ok {
			connections = connectionsVal
		} else {
			resp.Diagnostics.AddError(
				"Missing required field",
				"workflow_json must contain a 'connections' object",
			)
			return
		}

		// Extract settings (optional)
		if settingsVal, ok := workflowData["settings"].(map[string]interface{}); ok {
			settings = settingsVal
		}

		// Extract tags (optional)
		if tagsVal, ok := workflowData["tags"].([]interface{}); ok {
			tags = make([]map[string]string, 0, len(tagsVal))
			for _, tag := range tagsVal {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					tagStr := make(map[string]string)
					for k, v := range tagMap {
						if vStr, ok := v.(string); ok {
							tagStr[k] = vStr
						}
					}
					tags = append(tags, tagStr)
				}
			}
		}

		// Update plan with extracted values for state management
		plan.Name = types.StringValue(name)
		// plan.Active = types.BoolValue(active)

		nodesJSON, err := json.Marshal(nodes)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling nodes",
				"Could not marshal nodes to JSON: "+err.Error(),
			)
			return
		}
		plan.Nodes = types.StringValue(string(nodesJSON))

		connectionsJSON, err := json.Marshal(connections)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling connections",
				"Could not marshal connections to JSON: "+err.Error(),
			)
			return
		}
		plan.Connections = types.StringValue(string(connectionsJSON))

		if settings != nil {
			settingsJSON, err := json.Marshal(settings)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error marshaling settings",
					"Could not marshal settings to JSON: "+err.Error(),
				)
				return
			}
			plan.Settings = types.StringValue(string(settingsJSON))
		}

		if tags != nil {
			tagsJSON, err := json.Marshal(tags)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error marshaling tags",
					"Could not marshal tags to JSON: "+err.Error(),
				)
				return
			}
			plan.Tags = types.StringValue(string(tagsJSON))
		}
	} else {
		// Use individual attributes
		name = plan.Name.ValueString()
		// active = plan.Active.ValueBool()

		// Parse JSON strings
		if err := json.Unmarshal([]byte(plan.Nodes.ValueString()), &nodes); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing nodes JSON",
				"Could not parse nodes JSON: "+err.Error(),
			)
			return
		}

		if err := json.Unmarshal([]byte(plan.Connections.ValueString()), &connections); err != nil {
			resp.Diagnostics.AddError(
				"Error parsing connections JSON",
				"Could not parse connections JSON: "+err.Error(),
			)
			return
		}

		if !plan.Settings.IsNull() && plan.Settings.ValueString() != "" {
			if err := json.Unmarshal([]byte(plan.Settings.ValueString()), &settings); err != nil {
				resp.Diagnostics.AddError(
					"Error parsing settings JSON",
					"Could not parse settings JSON: "+err.Error(),
				)
				return
			}
		}

		if !plan.Tags.IsNull() && plan.Tags.ValueString() != "" {
			if err := json.Unmarshal([]byte(plan.Tags.ValueString()), &tags); err != nil {
				resp.Diagnostics.AddError(
					"Error parsing tags JSON",
					"Could not parse tags JSON: "+err.Error(),
				)
				return
			}
		}
	}

	// Update existing workflow
	workflow := &client.Workflow{
		Name:        name,
		Active:      active,
		Nodes:       nodes,
		Connections: connections,
		Settings:    settings,
		Tags:        tags,
	}

	updatedWorkflow, err := r.client.UpdateWorkflow(plan.ID.ValueString(), workflow)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating n8n Workflow",
			"Could not update workflow, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with updated items and timestamps
	plan.CreatedAt = types.StringValue(updatedWorkflow.CreatedAt)
	plan.UpdatedAt = types.StringValue(updatedWorkflow.UpdatedAt)

	// Ensure tags is set (even if empty)
	if len(updatedWorkflow.Tags) > 0 {
		tagsJSON, err := json.Marshal(updatedWorkflow.Tags)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error marshaling tags",
				"Could not marshal tags to JSON: "+err.Error(),
			)
			return
		}
		plan.Tags = types.StringValue(string(tagsJSON))
	} else {
		plan.Tags = types.StringValue("[]")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state workflowResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing workflow
	err := r.client.DeleteWorkflow(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting n8n Workflow",
			"Could not delete workflow, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
