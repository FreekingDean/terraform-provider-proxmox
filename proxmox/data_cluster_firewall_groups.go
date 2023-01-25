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
	_ datasource.DataSource = &clusterFirewallGroupsDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallGroupsDataSource)
}

type clusterFirewallGroupsDataSource struct {
	client *groups.Client
}

func NewClusterFirewallGroupsDataSource() datasource.DataSource {
	return &clusterFirewallGroupsDataSource{}
}

type clusterFirewallGroupsModel struct {
	ClusterFirewallGroups []ClusterFirewallGroups `tfsdk:"cluster_firewall_groups"`
}

type ClusterFirewallGroups struct {
	Digest  types.String `tfsdk:"digest"`
	Group   types.String `tfsdk:"group"`
	Comment types.String `tfsdk:"comment"`
}

func (d *clusterFirewallGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = groups.New(client)
	}
}

func (d *clusterFirewallGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_groups"
}

func (d *clusterFirewallGroupsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_firewall_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"digest": schema.StringAttribute{
							Computed: true,
						},
						"group": schema.StringAttribute{
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

func (d *clusterFirewallGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallGroupsModel

	clusterFirewallGroups, err := d.client.Index(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewallGroups",
			err.Error(),
		)
		return
	}

	for _, e := range *clusterFirewallGroups {
		eState := ClusterFirewallGroups{}
		eState.Digest = types.StringValue(e.Digest)
		eState.Group = types.StringValue(e.Group)
		if e.Comment != nil {
			eState.Comment = types.StringValue(*e.Comment)
		}
		state.ClusterFirewallGroups = append(state.ClusterFirewallGroups, eState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
