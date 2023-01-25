package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/cluster/firewall"
	"github.com/FreekingDean/terraform-provider-proxmox/internal/utils"
)

var (
	_ datasource.DataSource = &clusterFirewallsDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallsDataSource)
}

type clusterFirewallsDataSource struct {
	client *firewall.Client
}

func NewClusterFirewallsDataSource() datasource.DataSource {
	return &clusterFirewallsDataSource{}
}

type clusterFirewallsModel struct {
	ClusterFirewalls []map[string]types.String `tfsdk:"cluster_firewalls"`
}

func (d *clusterFirewallsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = firewall.New(client)
	}
}

func (d *clusterFirewallsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewalls"
}

func (d *clusterFirewallsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_firewalls": schema.ListAttribute{
				Computed: true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
		},
	}
}

func (d *clusterFirewallsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallsModel

	clusterFirewalls, err := d.client.Index(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewalls",
			err.Error(),
		)
		return
	}

	for _, e := range *clusterFirewalls {
		eState, diag := utils.NormalizeMap(*e)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.ClusterFirewalls = append(state.ClusterFirewalls, eState)
	}
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
