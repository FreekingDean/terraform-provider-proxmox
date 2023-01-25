package proxmox

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/FreekingDean/proxmox-api-go/proxmox"
	"github.com/FreekingDean/proxmox-api-go/proxmox/access/users"
)

var (
	_ datasource.DataSource = &accessUserDataSource{}
)

func init() {
	datasources = append(datasources, NewAccessUserDataSource)
}

type accessUserDataSource struct {
	client *users.Client
}

func NewAccessUserDataSource() datasource.DataSource {
	return &accessUserDataSource{}
}

type accessUserModel struct {
	Userid    types.String            `tfsdk:"userid"`
	Comment   types.String            `tfsdk:"comment"`
	Email     types.String            `tfsdk:"email"`
	Enable    types.Bool              `tfsdk:"enable"`
	Expire    types.Int64             `tfsdk:"expire"`
	Firstname types.String            `tfsdk:"firstname"`
	Groups    []types.String          `tfsdk:"groups"`
	Keys      types.String            `tfsdk:"keys"`
	Lastname  types.String            `tfsdk:"lastname"`
	Tokens    map[string]types.String `tfsdk:"tokens"`
}

func (d *accessUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if client, ok := req.ProviderData.(*proxmox.Client); ok {
		d.client = users.New(client)
	}
}

func (d *accessUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_user"
}

func (d *accessUserDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"userid": schema.StringAttribute{
				Required: true,
			},
			"comment": schema.StringAttribute{
				Computed: true,
			},
			"email": schema.StringAttribute{
				Computed: true,
			},
			"enable": schema.BoolAttribute{
				Computed: true,
			},
			"expire": schema.Int64Attribute{
				Computed: true,
			},
			"firstname": schema.StringAttribute{
				Computed: true,
			},
			"groups": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"keys": schema.StringAttribute{
				Computed: true,
			},
			"lastname": schema.StringAttribute{
				Computed: true,
			},
			"tokens": schema.MapAttribute{
				Computed:    true,
				ElementType: types.String,
			},
		},
	}
}

func (d *accessUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accessUserModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessUser, err := d.client.Find(
		ctx,
		&users.FindRequest{
			Userid: state.Userid.ValueString(),
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Proxmox AccessUser",
			err.Error(),
		)
		return
	}

	if accessUser.Comment != nil {
		state.Comment = types.StringValue(*accessUser.Comment)
	}
	if accessUser.Email != nil {
		state.Email = types.StringValue(*accessUser.Email)
	}
	if accessUser.Enable != nil {
		state.Enable = types.BoolValue(bool(*accessUser.Enable))
	}
	if accessUser.Expire != nil {
		state.Expire = types.Int64Value(*accessUser.Expire)
	}
	if accessUser.Firstname != nil {
		state.Firstname = types.StringValue(*accessUser.Firstname)
	}
	for _, e := range accessUser.Groups {
		eState := types.StringValue(e)
		state.Groups = append(state.Groups, eState)
	}
	if accessUser.Keys != nil {
		state.Keys = types.StringValue(*accessUser.Keys)
	}
	if accessUser.Lastname != nil {
		state.Lastname = types.StringValue(*accessUser.Lastname)
	}
	state.Tokens = map[string]types.StringValue(accessUser.Tokens)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
