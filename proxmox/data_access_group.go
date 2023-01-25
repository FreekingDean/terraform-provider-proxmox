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
	_ datasource.DataSource = &accessGroupDataSource{}
)

func init() {
	datasources = append(datasources, NewAccessGroupDataSource)
}

type accessGroupDataSource struct {
	client *groups.Client
}

func NewAccessGroupDataSource() datasource.DataSource {
	return &accessGroupDataSource{}
}

type accessGroupModel struct {
	Groupid types.String   `tfsdk:"groupid"`
	Members []types.String `tfsdk:"members"`
	Comment types.String   `tfsdk:"comment"`
}

func (d *accessGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = groups.New(client)
	}
}

func (d *accessGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_group"
}

func (d *accessGroupDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *accessGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessGroupModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessGroup, err := d.client.Find(
		ctx,
		&groups.FindRequest{
			Groupid: state.Groupid.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox AccessGroup",
			err.Error(),
		)
		return
	}

	for _, e := range accessGroup.Members {
		eState := types.StringValue(e)
		state.Members = append(state.Members, eState)
	}
	if accessGroup.Comment != nil {
		state.Comment = types.StringValue(*accessGroup.Comment)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
