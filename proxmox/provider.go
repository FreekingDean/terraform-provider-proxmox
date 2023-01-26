package proxmox

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &proxmoxProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &proxmoxProvider{}
}

// proxmoxProvider is the provider implementation.
type proxmoxProvider struct {
	client *proxmox.Client
}

// Metadata returns the provider type name.
func (p *proxmoxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "proxmox"
}

// Schema defines the provider-level schema for configuration data.
func (p *proxmoxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

type proxmoxProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Configure prepares a Proxmox API client for data sources and resources.
func (p *proxmoxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config proxmoxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Proxmox API Host",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Proxmox API Username",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Proxmox API Password",
			"The provider cannot create the Proxmox API client as there is an unknown configuration value for the Proxmox API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PROXMOX_PASSWORDenvironment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("PROXMOX_HOST")
	username := os.Getenv("PROXMOX_USERNAME")
	password := os.Getenv("PROXMOX_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	client := proxmox.NewClient(host)

	a := access.New(client)
	ticket, err := a.CreateTicket(ctx, &access.CreateTicketRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Proxmox API Client",
			"An unexpected error occurred when creating the Proxmox API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Proxmox Client Error: "+err.Error(),
		)
		return
	}

	client.SetCookie(*ticket.Ticket)
	client.SetCsrf(*ticket.Csrfpreventiontoken)
	p.client = client

	// Make the Proxmox client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *proxmoxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// Resources defines the resources implemented in the provider.
func (p *proxmoxProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		p.resourceFunc(&resourceNodeStorageContent{}),
	}
}

type clientResource interface {
	resource.Resource
	SetClient(c *proxmox.Client)
}

func (p *proxmoxProvider) resourceFunc(r clientResource) func() resource.Resource {
	return func() resource.Resource {
		r.SetClient(p.client)
		return r
	}
}
