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
	_ datasource.DataSource = &clusterFirewallRuleDataSource{}
)

func init() {
	datasources = append(datasources, NewClusterFirewallRuleDataSource)
}

type clusterFirewallRuleDataSource struct {
	client *rules.Client
}

func NewClusterFirewallRuleDataSource() datasource.DataSource {
	return &clusterFirewallRuleDataSource{}
}

type clusterFirewallRuleModel struct {
	Pos       types.Int64  `tfsdk:"pos"`
	Action    types.String `tfsdk:"action"`
	Pos       types.Int64  `tfsdk:"pos"`
	Type      types.String `tfsdk:"type"`
	Comment   types.String `tfsdk:"comment"`
	Dest      types.String `tfsdk:"dest"`
	Dport     types.String `tfsdk:"dport"`
	Enable    types.Int64  `tfsdk:"enable"`
	IcmpType  types.String `tfsdk:"icmptype"`
	Iface     types.String `tfsdk:"iface"`
	Ipversion types.Int64  `tfsdk:"ipversion"`
	Log       types.String `tfsdk:"log"`
	Macro     types.String `tfsdk:"macro"`
	Proto     types.String `tfsdk:"proto"`
	Source    types.String `tfsdk:"source"`
	Sport     types.String `tfsdk:"sport"`
}

func (d *clusterFirewallRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = rules.New(client)
	}
}

func (d *clusterFirewallRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_firewall_rule"
}

func (d *clusterFirewallRuleDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"pos": schema.Int64Attribute{
				Required: true,
			},
			"action": schema.StringAttribute{
				Computed: true,
			},
			"pos": schema.Int64Attribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"comment": schema.StringAttribute{
				Computed: true,
			},
			"dest": schema.StringAttribute{
				Computed: true,
			},
			"dport": schema.StringAttribute{
				Computed: true,
			},
			"enable": schema.Int64Attribute{
				Computed: true,
			},
			"icmptype": schema.StringAttribute{
				Computed: true,
			},
			"iface": schema.StringAttribute{
				Computed: true,
			},
			"ipversion": schema.Int64Attribute{
				Computed: true,
			},
			"log": schema.StringAttribute{
				Computed: true,
			},
			"macro": schema.StringAttribute{
				Computed: true,
			},
			"proto": schema.StringAttribute{
				Computed: true,
			},
			"source": schema.StringAttribute{
				Computed: true,
			},
			"sport": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *clusterFirewallRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clusterFirewallRuleModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterFirewallRule, err := d.client.Find(
		ctx,
		&rules.FindRequest{
			Pos: state.Pos.ValueInt64(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox ClusterFirewallRule",
			err.Error(),
		)
		return
	}

	state.Action = types.StringValue(clusterFirewallRule.Action)
	state.Pos = types.Int64Value(clusterFirewallRule.Pos)
	state.Type = types.StringValue(clusterFirewallRule.Type)
	if clusterFirewallRule.Comment != nil {
		state.Comment = types.StringValue(*clusterFirewallRule.Comment)
	}
	if clusterFirewallRule.Dest != nil {
		state.Dest = types.StringValue(*clusterFirewallRule.Dest)
	}
	if clusterFirewallRule.Dport != nil {
		state.Dport = types.StringValue(*clusterFirewallRule.Dport)
	}
	if clusterFirewallRule.Enable != nil {
		state.Enable = types.Int64Value(*clusterFirewallRule.Enable)
	}
	if clusterFirewallRule.IcmpType != nil {
		state.IcmpType = types.StringValue(*clusterFirewallRule.IcmpType)
	}
	if clusterFirewallRule.Iface != nil {
		state.Iface = types.StringValue(*clusterFirewallRule.Iface)
	}
	if clusterFirewallRule.Ipversion != nil {
		state.Ipversion = types.Int64Value(*clusterFirewallRule.Ipversion)
	}
	if clusterFirewallRule.Log != nil {
		state.Log = types.StringValue(*clusterFirewallRule.Log)
	}
	if clusterFirewallRule.Macro != nil {
		state.Macro = types.StringValue(*clusterFirewallRule.Macro)
	}
	if clusterFirewallRule.Proto != nil {
		state.Proto = types.StringValue(*clusterFirewallRule.Proto)
	}
	if clusterFirewallRule.Source != nil {
		state.Source = types.StringValue(*clusterFirewallRule.Source)
	}
	if clusterFirewallRule.Sport != nil {
		state.Sport = types.StringValue(*clusterFirewallRule.Sport)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
