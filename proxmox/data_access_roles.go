package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/roles"
)

var (
	_ datasource.DataSource = &accessRolesDataSource{}
)

func init() {
	datasources = append(datasources, NewAccessRolesDataSource)
}

type accessRolesDataSource struct {
	client *roles.Client
}

func NewAccessRolesDataSource() datasource.DataSource {
	return &accessRolesDataSource{}
}

type accessRolesModel struct {
	AccessRoles []AccessRoles `tfsdk:"access_roles"`
}

type AccessRoles struct {
	Roleid  types.String `tfsdk:"roleid"`
	Privs   types.String `tfsdk:"privs"`
	Special types.Bool   `tfsdk:"special"`
}

func (d *accessRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = roles.New(client)
	}
}

func (d *accessRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_roles"
}

func (d *accessRolesDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"roleid": schema.StringAttribute{
							Computed: true,
						},
						"privs": schema.StringAttribute{
							Computed: true,
						},
						"special": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *accessRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessRolesModel

	accessRoles, err := d.client.Index(
		ctx,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox AccessRoles",
			err.Error(),
		)
		return
	}

	for _, e := range *accessRoles {
		eState := AccessRoles{}
		eState.Roleid = types.StringValue(e.Roleid)
		if e.Privs != nil {
			eState.Privs = types.StringValue(*e.Privs)
		}
		if e.Special != nil {
			eState.Special = types.BoolValue(bool(*e.Special))
		}
		state.AccessRoles = append(state.AccessRoles, eState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
