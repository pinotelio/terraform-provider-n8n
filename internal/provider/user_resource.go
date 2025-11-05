package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-n8n/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

// NewUserResource is a helper function to simplify the provider implementation.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource is the resource implementation.
type userResource struct {
	client *client.Client
}

// userResourceModel maps the resource schema data.
type userResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	IsOwner   types.Bool   `tfsdk:"is_owner"`
	IsPending types.Bool   `tfsdk:"is_pending"`
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Schema defines the schema for the resource.
func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an n8n user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User identifier",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "Email address of the user (cannot be changed after creation)",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Description: "Role of the user (e.g., 'global:owner', 'global:admin', 'global:member')",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("global:member"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_owner": schema.BoolAttribute{
				Description: "Whether the user is an owner",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_pending": schema.BoolAttribute{
				Description: "Whether the user account is pending activation",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the user was created",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the user was last updated",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new user
	user := &client.User{
		Email: plan.Email.ValueString(),
		Role:  plan.Role.ValueString(),
	}

	createdUser, err := r.client.CreateUser(user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating user",
			"Could not create user, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(createdUser.ID)
	plan.Email = types.StringValue(createdUser.Email)
	plan.Role = types.StringValue(createdUser.GetRole())
	plan.IsOwner = types.BoolValue(createdUser.IsOwner)
	plan.IsPending = types.BoolValue(createdUser.IsPending)
	plan.CreatedAt = types.StringValue(createdUser.CreatedAt)
	plan.UpdatedAt = types.StringValue(createdUser.UpdatedAt)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed user value from n8n
	user, err := r.client.GetUser(state.ID.ValueString())
	if err != nil {
		// Check if the user was deleted outside of Terraform (404 error)
		if strings.Contains(err.Error(), "404") {
			// Remove from state - Terraform will recreate it on next apply
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading n8n User",
			"Could not read n8n user ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Email = types.StringValue(user.Email)
	state.Role = types.StringValue(user.GetRole())
	state.IsOwner = types.BoolValue(user.IsOwner)
	state.IsPending = types.BoolValue(user.IsPending)
	state.CreatedAt = types.StringValue(user.CreatedAt)
	state.UpdatedAt = types.StringValue(user.UpdatedAt)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update existing user
	// Note: Only role can be updated via the n8n API
	user := &client.User{
		Role: plan.Role.ValueString(),
	}

	updatedUser, err := r.client.UpdateUser(plan.ID.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating n8n User",
			"Could not update user, unexpected error: "+err.Error(),
		)
		return
	}

	// Update resource state with refreshed data from API
	plan.Email = types.StringValue(updatedUser.Email)
	plan.Role = types.StringValue(updatedUser.GetRole())
	plan.IsOwner = types.BoolValue(updatedUser.IsOwner)
	plan.IsPending = types.BoolValue(updatedUser.IsPending)
	plan.CreatedAt = types.StringValue(updatedUser.CreatedAt)
	plan.UpdatedAt = types.StringValue(updatedUser.UpdatedAt)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing user
	err := r.client.DeleteUser(state.ID.ValueString())
	if err != nil {
		// Some n8n instances may not support user deletion via API
		// In this case, we log a warning but still remove from state
		resp.Diagnostics.AddWarning(
			"Error Deleting n8n User",
			fmt.Sprintf("Could not delete user %s via API: %s. The user may need to be deleted manually through the n8n UI. The resource will be removed from Terraform state.", state.ID.ValueString(), err.Error()),
		)
		// Don't return - allow state to be removed even if API deletion fails
	}
}

// ImportState imports the resource state.
func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
