package provider

import (
	"context"
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
	_ resource.Resource                = &workflowActivationResource{}
	_ resource.ResourceWithConfigure   = &workflowActivationResource{}
	_ resource.ResourceWithImportState = &workflowActivationResource{}
)

// NewWorkflowActivationResource is a helper function to simplify the provider implementation.
func NewWorkflowActivationResource() resource.Resource {
	return &workflowActivationResource{}
}

// workflowActivationResource is the resource implementation.
type workflowActivationResource struct {
	client *client.Client
}

// workflowActivationResourceModel maps the resource schema data.
type workflowActivationResourceModel struct {
	ID         types.String `tfsdk:"id"`
	WorkflowID types.String `tfsdk:"workflow_id"`
	Active     types.Bool   `tfsdk:"active"`
}

// Metadata returns the resource type name.
func (r *workflowActivationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_activation"
}

// Schema defines the schema for the resource.
func (r *workflowActivationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the activation state of an n8n workflow. This resource controls whether a workflow is active (running) or inactive. Workflows must have at least one trigger node to be activated.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier (same as workflow_id)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_id": schema.StringAttribute{
				Description: "The ID of the workflow to manage activation for",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the workflow should be active. Note: Workflows must have at least one trigger, poller, or webhook node to be activated.",
				Required:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *workflowActivationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *workflowActivationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan workflowActivationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify the workflow exists
	workflow, err := r.client.GetWorkflow(plan.WorkflowID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.Diagnostics.AddError(
				"Workflow Not Found",
				"The workflow with ID "+plan.WorkflowID.ValueString()+" does not exist. Please ensure the workflow is created before managing its activation state.",
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Reading Workflow",
				"Could not read workflow ID "+plan.WorkflowID.ValueString()+": "+err.Error(),
			)
		}
		return
	}

	// Set the activation state
	if plan.Active.ValueBool() && !workflow.Active {
		// Activate the workflow
		_, err := r.client.ActivateWorkflow(plan.WorkflowID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Activating Workflow",
				"Could not activate workflow: "+err.Error(),
			)
			return
		}
	} else if !plan.Active.ValueBool() && workflow.Active {
		// Deactivate the workflow
		_, err := r.client.DeactivateWorkflow(plan.WorkflowID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Deactivating Workflow",
				"Could not deactivate workflow: "+err.Error(),
			)
			return
		}
	}

	// Set the ID to the workflow ID
	plan.ID = plan.WorkflowID

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowActivationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state workflowActivationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed workflow value from n8n
	workflow, err := r.client.GetWorkflow(state.WorkflowID.ValueString())
	if err != nil {
		// Check if the workflow was deleted outside of Terraform (404 error)
		if strings.Contains(err.Error(), "404") {
			// Remove from state - the workflow is gone
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Workflow",
			"Could not read workflow ID "+state.WorkflowID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update the active state
	state.Active = types.BoolValue(workflow.Active)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowActivationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan workflowActivationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state workflowActivationResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only update if the active state changed
	if plan.Active.ValueBool() != state.Active.ValueBool() {
		if plan.Active.ValueBool() {
			// Activate the workflow
			_, err := r.client.ActivateWorkflow(plan.WorkflowID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Activating Workflow",
					"Could not activate workflow: "+err.Error(),
				)
				return
			}
		} else {
			// Deactivate the workflow
			_, err := r.client.DeactivateWorkflow(plan.WorkflowID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Deactivating Workflow",
					"Could not deactivate workflow: "+err.Error(),
				)
				return
			}
		}
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowActivationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state workflowActivationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// When deleting the activation resource, deactivate the workflow
	// This ensures the workflow is left in an inactive state
	workflow, err := r.client.GetWorkflow(state.WorkflowID.ValueString())
	if err != nil {
		// If workflow doesn't exist, that's fine - nothing to deactivate
		if strings.Contains(err.Error(), "404") {
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Workflow",
			"Could not read workflow ID "+state.WorkflowID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Only deactivate if it's currently active
	if workflow.Active {
		_, err := r.client.DeactivateWorkflow(state.WorkflowID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Deactivating Workflow",
				"Could not deactivate workflow: "+err.Error(),
			)
			return
		}
	}
}

// ImportState imports the resource state.
func (r *workflowActivationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using workflow ID
	// Set both id and workflow_id to the imported value
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workflow_id"), req.ID)...)
}
