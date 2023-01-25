package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/aliases"
	"github.com/FreekingDean/terraform-provider-proxmox/internal/utils"
)

var (
	_ datasource.DataSource = &clusterFirewallAliasDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallAliasDataSource)
}

type clusterFirewallAliasDataSource struct {
	client *aliases.Client
}

func NewClusterFirewallAliasDataSource() datasource.DataSource {
	return &clusterFirewallAliasDataSource{}
}

type clusterFirewallAliasModel struct {
	Name  types.String            `tfsdk:"name"`
	Attrs map[string]types.String `tfsdk:"attrs"`
}

func (d *clusterFirewallAliasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = aliases.New(client)
	}
}

func (d *clusterFirewallAliasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_alias"
}

func (d *clusterFirewallAliasDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"attrs": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *clusterFirewallAliasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallAliasModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterFirewallAlias, err := d.client.Find(
		ctx,
		&aliases.FindRequest{
			Name: state.Name.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewallAlias",
			err.Error(),
		)
		return
	}

	state.Attrs, diags = utils.NormalizeMap(*clusterFirewallAlias)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
