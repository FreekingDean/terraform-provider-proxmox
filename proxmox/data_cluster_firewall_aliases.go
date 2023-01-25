package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/aliases"
)

var (
	_ datasource.DataSource = &clusterFirewallAliasesDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallAliasesDataSource)
}

type clusterFirewallAliasesDataSource struct {
	client *aliases.Client
}

func NewClusterFirewallAliasesDataSource() datasource.DataSource {
	return &clusterFirewallAliasesDataSource{}
}

type clusterFirewallAliasesModel struct {
	ClusterFirewallAliases []ClusterFirewallAliases `tfsdk:"cluster_firewall_aliases"`
}

type ClusterFirewallAliases struct {
	Cidr    types.String `tfsdk:"cidr"`
	Digest  types.String `tfsdk:"digest"`
	Name    types.String `tfsdk:"name"`
	Comment types.String `tfsdk:"comment"`
}

func (d *clusterFirewallAliasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = aliases.New(client)
	}
}

func (d *clusterFirewallAliasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_aliases"
}

func (d *clusterFirewallAliasesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_firewall_aliases": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cidr": schema.StringAttribute{
							Computed: true,
						},
						"digest": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"comment": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *clusterFirewallAliasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallAliasesModel

	clusterFirewallAliases, err := d.client.Index(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewallAliases",
			err.Error(),
		)
		return
	}

	for _, e := range *clusterFirewallAliases {
		eState := ClusterFirewallAliases{}
		eState.Cidr = types.StringValue(e.Cidr)
		eState.Digest = types.StringValue(e.Digest)
		eState.Name = types.StringValue(e.Name)
		if e.Comment != nil {
			eState.Comment = types.StringValue(*e.Comment)
		}
		state.ClusterFirewallAliases = append(state.ClusterFirewallAliases, eState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
