package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall/groups"
)

var (
	_ datasource.DataSource = &clusterFirewallGroupDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallGroupDataSource)
}

type clusterFirewallGroupDataSource struct {
	client *groups.Client
}

func NewClusterFirewallGroupDataSource() datasource.DataSource {
	return &clusterFirewallGroupDataSource{}
}

type clusterFirewallGroupModel struct {
	Group                types.String           `tfsdk:"group"`
	ClusterFirewallGroup []ClusterFirewallGroup `tfsdk:"cluster_firewall_group"`
}

type ClusterFirewallGroup struct {
	Pos types.Int64 `tfsdk:"pos"`
}

func (d *clusterFirewallGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = groups.New(client)
	}
}

func (d *clusterFirewallGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_group"
}

func (d *clusterFirewallGroupDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				Required: true,
			},
			"cluster_firewall_group": schema.ListNestedAttribute{
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

func (d *clusterFirewallGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallGroupModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterFirewallGroup, err := d.client.Find(
		ctx,
		&groups.FindRequest{
			Group: state.Group.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewallGroup",
			err.Error(),
		)
		return
	}

	for _, e := range *clusterFirewallGroup {
		eState := ClusterFirewallGroup{}
		eState.Pos = types.Int64Value(e.Pos)
		state.ClusterFirewallGroup = append(state.ClusterFirewallGroup, eState)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
