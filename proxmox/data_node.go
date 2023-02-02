package proxmox

import (
	"context"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/nodes/network"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type nodeModel struct {
	Name      types.String `tfsdk:"name"`
	IPAddress types.String `tfsdk:"ip_address"`
}

type dataNode struct {
	//n   *nodes.Client
	net *network.Client
}

func (d *dataNode) SetClient(p *proxmox.Client) {
	d.net = network.New(p)
}

// Metadata returns the data source type name.
func (d *dataNode) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

// Schema defines the schema for the data source.
func (d *dataNode) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"ip_address": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *dataNode) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state nodeModel
	diags := resp.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	networks, err := d.net.Index(ctx, network.IndexRequest{
		Node: state.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get node ",
			"An unexpected error occurred when trying to retreive the "+
				"node networks. "+
				"Proxmox API Error: "+err.Error(),
		)
		return
	}
	for _, net := range networks {
		if t, ok := net["type"]; ok {
			if tStr, ok := t.(string); ok {
				if tStr == "bridge" {
					if ip, ok := net["address"]; ok {
						if ipStr, ok := ip.(string); ok {
							state.IPAddress = types.StringValue(ipStr)
						}
					}
				}
			}
		}
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
