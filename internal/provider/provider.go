package provider

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/managementgroups/armmanagementgroups"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &azurecnProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &azurecnProvider{
			version: version,
		}
	}
}

// azurecnProvider is the provider implementation.
type azurecnProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type azurecnProviderModel struct {
	TenantId     types.String `tfsdk:"tenant_id"`
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// Metadata returns the provider type name.
func (p *azurecnProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azurecnp"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *azurecnProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Optional: true,
			},
			"client_secret": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *azurecnProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config azurecnProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.TenantId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_id"),
			"Unknown Azure API tenant_id",
			"The provider cannot create the Azure API client as there is an unknown configuration value for the Azure API tenant_id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ARM_TENANT_ID environment variable.",
		)
	}

	if config.ClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Azure API tenant_id",
			"The provider cannot create the Azure API client as there is an unknown configuration value for the Azure API client_id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ARM_CLIENT_ID environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Azure API client_secret",
			"The provider cannot create the Azuree API client as there is an unknown configuration value for the Azure API clientSecret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ARM_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	tenantId := os.Getenv("ARM_TENANT_ID")
	clientId := os.Getenv("ARM_CLIENT_ID")
	clientSecret := os.Getenv("ARM_CLIENT_SECRET")

	if !config.TenantId.IsNull() {
		tenantId = config.TenantId.ValueString()
	}

	if !config.ClientId.IsNull() {
		clientId = config.ClientId.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if tenantId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant_id"),
			"Missing Azure API TenantId",
			"The provider cannot create the Azure API client as there is a missing or empty value for the Azure API tenant_id. "+
				"Set the tenant_id value in the configuration or use the ARM_TENANT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing Azure API ClientId",
			"The provider cannot create the Azure API client as there is a missing or empty value for the Azure API client_id. "+
				"Set the client_id value in the configuration or use the ARM_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing Azure API ClientSecret",
			"The provider cannot create the Azure API client as there is a missing or empty value for the Azure API client_secret. "+
				"Set the client_secret value in the configuration or use the ARM_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new HashiCups client using the configuration values
	var credentials, err = azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Azure API Credentials", err.Error())
	}
	managementGroupFactory, err := armmanagementgroups.NewClientFactory(credentials, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Azure API Client factory",
			"An unexpected error occurred when creating the Azure API client factory. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Azure Client Error: "+err.Error(),
		)
		return
	}
	subscrioptionFactory, err := armsubscription.NewClientFactory(credentials, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Azure API Client factory",
			"An unexpected error occurred when creating the Azure API client factory. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Azure Client Error: "+err.Error(),
		)
		return
	}

	var holder = ClientFactoryHolder{
		managementGroupClientFactory: managementGroupFactory,
		subscriptionClientFactory:    subscrioptionFactory,
	}
	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = &holder
	resp.ResourceData = &holder
}

// DataSources defines the data sources implemented in the provider.
func (p *azurecnProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *azurecnProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSubscriptionPoolLeaseResource,
	}
}

type ClientFactoryHolder struct { //TODO there gotta be a better way to move the client factories around
	managementGroupClientFactory *armmanagementgroups.ClientFactory
	subscriptionClientFactory    *armsubscription.ClientFactory
}
