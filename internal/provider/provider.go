package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pinotelio/terraform-provider-n8n/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &n8nProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &n8nProvider{
			version: version,
		}
	}
}

// n8nProvider is the provider implementation.
type n8nProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// n8nProviderModel maps provider schema data to a Go type.
type n8nProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	APIKey   types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *n8nProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "n8n"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *n8nProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with n8n self-hosted cluster.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The n8n API endpoint URL. May also be provided via N8N_ENDPOINT environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The n8n API key for authentication. May also be provided via N8N_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a n8n API client for data sources and resources.
func (p *n8nProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config n8nProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown n8n API Endpoint",
			"The provider cannot create the n8n API client as there is an unknown configuration value for the n8n API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the N8N_ENDPOINT environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown n8n API Key",
			"The provider cannot create the n8n API client as there is an unknown configuration value for the n8n API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the N8N_API_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	endpoint := os.Getenv("N8N_ENDPOINT")
	apiKey := os.Getenv("N8N_API_KEY")

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing n8n API Endpoint",
			"The provider cannot create the n8n API client as there is a missing or empty value for the n8n API endpoint. "+
				"Set the endpoint value in the configuration or use the N8N_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing n8n API Key",
			"The provider cannot create the n8n API client as there is a missing or empty value for the n8n API key. "+
				"Set the api_key value in the configuration or use the N8N_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new n8n client using the configuration values
	n8nClient := client.NewClient(endpoint, apiKey)

	// Make the n8n client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = n8nClient
	resp.ResourceData = n8nClient
}

// DataSources defines the data sources implemented in the provider.
func (p *n8nProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewWorkflowDataSource,
		// NewCredentialDataSource is not included because the n8n API does not
		// support reading credentials for security reasons. See CREDENTIAL_LIMITATIONS.md
		NewUserDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *n8nProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWorkflowResource,
		NewWorkflowActivationResource,
		NewCredentialResource,
		NewUserResource,
	}
}
