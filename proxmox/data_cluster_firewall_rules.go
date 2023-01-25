package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/rules"
)

var (
	_ datasource.DataSource = &clusterFirewallRulesDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallRulesDataSource)
}

type clusterFirewallRulesDataSource struct {
	client *rules.Client
}

func NewClusterFirewallRulesDataSource() datasource.DataSource {
	return &clusterFirewallRulesDataSource{}
}

type clusterFirewallRulesModel struct {
	ClusterFirewallRules []ClusterFirewallRules `tfsdk:"cluster_firewall_rules"`
}

type ClusterFirewallRules struct {
	Pos types.Int64 `tfsdk:"pos"`
}

func (d *clusterFirewallRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = rules.New(client)
	}
}

func (d *clusterFirewallRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_rules"
}

func (d *clusterFirewallRulesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_firewall_rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pos": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *clusterFirewallRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallRulesModel

	clusterFirewallRules, err := d.client.Index(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewallRules",
			err.Error(),
		)
		return
	}

	for _, e := range *clusterFirewallRules {
		eState := ClusterFirewallRules{}
		eState.Pos = types.Int64Value(e.Pos)
		state.ClusterFirewallRules = append(state.ClusterFirewallRules, eState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
