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
	_ datasource.DataSource = &accessGroupsDataSource{}
)

func init() {
	datasources = append(datasources, NewAccessGroupsDataSource)
}

type accessGroupsDataSource struct {
	client *groups.Client
}

func NewAccessGroupsDataSource() datasource.DataSource {
	return &accessGroupsDataSource{}
}

type accessGroupsModel struct {
	AccessGroups []AccessGroups `tfsdk:"access_groups"`
}

type AccessGroups struct {
	Groupid types.String `tfsdk:"groupid"`
	Comment types.String `tfsdk:"comment"`
	Users   types.String `tfsdk:"users"`
}

func (d *accessGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = groups.New(client)
	}
}

func (d *accessGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_groups"
}

func (d *accessGroupsDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"groupid": schema.StringAttribute{
							Computed: true,
						},
						"comment": schema.StringAttribute{
							Computed: true,
						},
						"users": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *accessGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessGroupsModel

	accessGroups, err := d.client.Index(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox AccessGroups",
			err.Error(),
		)
		return
	}

	for _, e := range *accessGroups {
		eState := AccessGroups{}
		eState.Groupid = types.StringValue(e.Groupid)
		if e.Comment != nil {
			eState.Comment = types.StringValue(*e.Comment)
		}
		if e.Users != nil {
			eState.Users = types.StringValue(*e.Users)
		}
		state.AccessGroups = append(state.AccessGroups, eState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
