package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/groups"
)

var (
	_ datasource.DataSource = &clusterFirewallDataSource{}
)

type clusterFirewallDataSource struct {
	client *groups.Client
}

type clusterFirewallModel struct {
	Groupid types.String   `tfsdk:"groupid"`
	Members []types.String `tfsdk:"members"`
	Comment types.String   `tfsdk:"comment"`
}

func (d *clusterFirewallDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = groups.New(client)
	}
}

func (d *clusterFirewallDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall"
}

func (d *clusterFirewallDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"groupid": schema.StringAttribute{
				Required: true,
			},
			"members": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"comment": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *clusterFirewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterFirewall, err := d.client.Find(
		ctx,
		&groups.FindRequest{
			Groupid: state.Groupid.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewall",
			err.Error(),
		)
		return
	}

	for _, e := range clusterFirewall.Members {
		eState := types.StringValue(e)
		state.Members = append(state.Members, eState)
	}
	if clusterFirewall.Comment != nil {
		state.Comment = types.StringValue(*clusterFirewall.Comment)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
