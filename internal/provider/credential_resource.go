package provider

import (
	"context"
	"encoding/json"
	"fmt"

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
	_ resource.Resource                = &credentialResource{}
	_ resource.ResourceWithConfigure   = &credentialResource{}
	_ resource.ResourceWithImportState = &credentialResource{}
)

// NewCredentialResource is a helper function to simplify the provider implementation.
func NewCredentialResource() resource.Resource {
	return &credentialResource{}
}

// credentialResource is the resource implementation.
type credentialResource struct {
	client *client.Client
}

// credentialResourceModel maps the resource schema data.
type credentialResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
	Data types.String `tfsdk:"data"`
}

// Metadata returns the resource type name.
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}

// Schema defines the schema for the resource.
func (r *credentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an n8n credential.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Credential identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the credential",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the credential (e.g., 'httpBasicAuth', 'slackApi', etc.)",
				Required:    true,
			},
			"data": schema.StringAttribute{
				Description: "JSON string representing the credential data",
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan credentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse JSON string for data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Data.ValueString()), &data); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing data JSON",
			"Could not parse data JSON: "+err.Error(),
		)
		return
	}

	// Create new credential
	credential := &client.Credential{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
		Data: data,
	}

	createdCredential, err := r.client.CreateCredential(credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating credential",
			"Could not create credential, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(createdCredential.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state credentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// n8n API does not support reading credentials for security reasons:
	// - No GET /api/v1/credentials/{id} endpoint (returns 405)
	// - No LIST /api/v1/credentials endpoint available
	//
	// Therefore, we cannot refresh the credential state from the API.
	// We keep the existing state as-is. This means:
	// - Terraform will not detect manual changes to credentials in n8n
	// - The credential data remains in Terraform state
	// - Updates via Terraform will still work (using PATCH)
	// - Deletes via Terraform will still work (using DELETE)
	//
	// This is a common pattern for resources with sensitive data that cannot be read back.

	// Simply return the existing state without making any API calls
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan credentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse JSON string for data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(plan.Data.ValueString()), &data); err != nil {
		resp.Diagnostics.AddError(
			"Error parsing data JSON",
			"Could not parse data JSON: "+err.Error(),
		)
		return
	}

	// Update existing credential
	credential := &client.Credential{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
		Data: data,
	}

	_, err := r.client.UpdateCredential(plan.ID.ValueString(), credential)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating n8n Credential",
			"Could not update credential, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state credentialResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing credential
	err := r.client.DeleteCredential(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting n8n Credential",
			"Could not delete credential, unexpected error: "+err.Error(),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *credentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
